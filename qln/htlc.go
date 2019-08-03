package qln

import (
	"bytes"
	"fmt"
	"time"

	"github.com/mit-dci/lit/btcutil/chaincfg/chainhash"
	"github.com/mit-dci/lit/btcutil/txscript"
	"github.com/mit-dci/lit/consts"
	"github.com/mit-dci/lit/crypto/fastsha256"
	"github.com/mit-dci/lit/lnutil"
	"github.com/mit-dci/lit/logging"
	"github.com/mit-dci/lit/portxo"
	"github.com/mit-dci/lit/sig64"
	"github.com/mit-dci/lit/wire"
)

func (nd *LitNode) OfferHTLC(qc *Qchan, amt uint32, RHash [32]byte, locktime uint32, data [32]byte) error {
	logging.Infof("starting HTLC offer")

	if qc.State.Failed {
		return fmt.Errorf("cannot offer HTLC, channel failed")
	}

	if amt >= consts.MaxSendAmt {
		return fmt.Errorf("max send 1G sat (1073741823)")
	}
	if amt == 0 {
		return fmt.Errorf("have to send non-zero amount")
	}

	// see if channel is busy
	// lock this channel
	cts := false
	for !cts {
		qc.ChanMtx.Lock()
		select {
		case <-qc.ClearToSend:
			cts = true
		default:
			qc.ChanMtx.Unlock()
		}
	}
	// ClearToSend is now empty

	// reload from disk here, after unlock
	err := nd.ReloadQchanState(qc)
	if err != nil {
		// don't clear to send here; something is wrong with the channel
		nd.FailChannel(qc)
		qc.ChanMtx.Unlock()
		return err
	}

	// check that channel is confirmed, if non-test coin
	wal, ok := nd.SubWallet[qc.Coin()]
	if !ok {
		qc.ClearToSend <- true
		qc.ChanMtx.Unlock()
		return fmt.Errorf("Not connected to coin type %d\n", qc.Coin())
	}

	if !wal.Params().TestCoin && qc.Height < 100 {
		qc.ClearToSend <- true
		qc.ChanMtx.Unlock()
		return fmt.Errorf(
			"height %d; must wait min 1 conf for non-test coin\n", qc.Height)
	}

	myAmt, _ := qc.GetChannelBalances()
	myAmt -= qc.State.Fee + int64(amt)

	// check if this push would lower my balance below minBal
	if myAmt < consts.MinOutput {
		qc.ClearToSend <- true
		qc.ChanMtx.Unlock()
		return fmt.Errorf("want to push %s but %s available after %s fee and %s consts.MinOutput",
			lnutil.SatoshiColor(int64(amt)),
			lnutil.SatoshiColor(qc.State.MyAmt-qc.State.Fee-consts.MinOutput),
			lnutil.SatoshiColor(qc.State.Fee),
			lnutil.SatoshiColor(consts.MinOutput))
	}

	// if we got here, but channel is not in rest state, try to fix it.
	if qc.State.Delta != 0 || qc.State.InProgHTLC != nil {
		nd.FailChannel(qc)
		qc.ChanMtx.Unlock()
		return fmt.Errorf("channel not in rest state")
	}

	qc.State.Data = data

	qc.State.InProgHTLC = new(HTLC)
	qc.State.InProgHTLC.Idx = qc.State.HTLCIdx
	qc.State.InProgHTLC.Incoming = false
	qc.State.InProgHTLC.Amt = int64(amt)
	qc.State.InProgHTLC.RHash = RHash
	qc.State.InProgHTLC.Locktime = locktime
	qc.State.InProgHTLC.TheirHTLCBase = qc.State.NextHTLCBase

	qc.State.InProgHTLC.KeyGen.Depth = 5
	qc.State.InProgHTLC.KeyGen.Step[0] = 44 | 1<<31
	qc.State.InProgHTLC.KeyGen.Step[1] = qc.Coin() | 1<<31
	qc.State.InProgHTLC.KeyGen.Step[2] = UseHTLCBase
	qc.State.InProgHTLC.KeyGen.Step[3] = qc.State.HTLCIdx | 1<<31
	qc.State.InProgHTLC.KeyGen.Step[4] = qc.Idx() | 1<<31

	qc.State.InProgHTLC.MyHTLCBase, _ = nd.GetUsePub(qc.State.InProgHTLC.KeyGen,
		UseHTLCBase)

	// save to db with ONLY InProgHTLC changed
	err = nd.SaveQchanState(qc)
	if err != nil {
		// don't clear to send here; something is wrong with the channel
		nd.FailChannel(qc)
		qc.ChanMtx.Unlock()
		return err
	}

	logging.Infof("OfferHTLC: Sending HashSig")

	err = nd.SendHashSig(qc)
	if err != nil {
		nd.FailChannel(qc)
		qc.ChanMtx.Unlock()
		return err
	}

	logging.Info("OfferHTLC: Done: sent HashSig")

	logging.Info("got pre CTS...")
	qc.ChanMtx.Unlock()

	timeout := time.NewTimer(time.Second * consts.ChannelTimeout)

	cts = false
	for !cts {
		qc.ChanMtx.Lock()
		select {
		case <-qc.ClearToSend:
			cts = true
		case <-timeout.C:
			nd.FailChannel(qc)
			qc.ChanMtx.Unlock()
			return fmt.Errorf("channel failed: operation timed out")
		default:
			qc.ChanMtx.Unlock()
		}
	}

	logging.Info("got post CTS...")
	// since we cleared with that statement, fill it again before returning
	qc.ClearToSend <- true
	qc.ChanMtx.Unlock()

	return nil
}

func (nd *LitNode) SendHashSig(q *Qchan) error {
	q.State.StateIdx++

	q.State.MyAmt -= int64(q.State.InProgHTLC.Amt)

	q.State.ElkPoint = q.State.NextElkPoint
	q.State.NextElkPoint = q.State.N2ElkPoint

	// make the signature to send over
	commitmentSig, HTLCSigs, err := nd.SignState(q)
	if err != nil {
		return err
	}

	q.State.NextHTLCBase = q.State.N2HTLCBase

	outMsg := lnutil.NewHashSigMsg(q.Peer(), q.Op, q.State.InProgHTLC.Amt, q.State.InProgHTLC.Locktime, q.State.InProgHTLC.RHash, commitmentSig, HTLCSigs, q.State.Data)

	logging.Infof("Sending HashSig with %d HTLC sigs", len(HTLCSigs))

	nd.tmpSendLitMsg(outMsg)

	return nil
}

func (nd *LitNode) HashSigHandler(msg lnutil.HashSigMsg, qc *Qchan) error {
	logging.Infof("Got HashSig: %v", msg)

	var collision bool

	// we should be clear to send when we get a hashSig
	select {
	case <-qc.ClearToSend:
	// keep going, normal
	default:
		// collision
		collision = true
	}

	logging.Infof("COLLISION is (%t)\n", collision)

	// load state from disk
	err := nd.ReloadQchanState(qc)
	if err != nil {
		nd.FailChannel(qc)
		return fmt.Errorf("HashSigHandler ReloadQchan err %s", err.Error())
	}

	// TODO we should send a response that the channel is closed.
	// or offer to double spend with a cooperative close?
	// or update the remote node on closed channel status when connecting
	// TODO should disallow 'break' command when connected to the other node
	// or merge 'break' and 'close' UI so that it breaks when it can't
	// connect, and closes when it can.
	if qc.CloseData.Closed {
		return fmt.Errorf("HashSigHandler err: %d, %d is closed.",
			qc.Peer(), qc.Idx())
	}

	inProgHTLC := qc.State.InProgHTLC

	htlcIdx := qc.State.HTLCIdx

	clearingIdxs := make([]uint32, 0)
	for _, h := range qc.State.HTLCs {
		if h.Clearing {
			clearingIdxs = append(clearingIdxs, h.Idx)
		}
	}

	// If we are colliding
	if collision {
		if qc.State.InProgHTLC != nil {
			// HashSig-HashSig collision
			// Set the Idx to the InProg one first - to allow signature
			// verification. Correct later
			htlcIdx = qc.State.InProgHTLC.Idx
		} else if len(clearingIdxs) > 0 {
			// HashSig-PreimageSig collision
			// Remove the clearing state for signature verification and
			// add back afterwards.
			for _, idx := range clearingIdxs {
				qh := &qc.State.HTLCs[idx]
				qh.Clearing = false
			}
			qc.State.CollidingHashPreimage = true
		} else {
			// We are colliding with DeltaSig
			qc.State.CollidingHashDelta = true
		}
	}

	incomingHTLC := new(HTLC)
	incomingHTLC.Idx = htlcIdx
	incomingHTLC.Incoming = true
	incomingHTLC.Amt = int64(msg.Amt)
	incomingHTLC.RHash = msg.RHash
	incomingHTLC.Locktime = msg.Locktime
	incomingHTLC.TheirHTLCBase = qc.State.NextHTLCBase

	incomingHTLC.KeyGen.Depth = 5
	incomingHTLC.KeyGen.Step[0] = 44 | 1<<31
	incomingHTLC.KeyGen.Step[1] = qc.Coin() | 1<<31
	incomingHTLC.KeyGen.Step[2] = UseHTLCBase
	incomingHTLC.KeyGen.Step[3] = htlcIdx | 1<<31
	incomingHTLC.KeyGen.Step[4] = qc.Idx() | 1<<31

	incomingHTLC.MyHTLCBase, _ = nd.GetUsePub(incomingHTLC.KeyGen,
		UseHTLCBase)

	// In order to check the incoming HTLC sigs, put it as the in progress one.
	// We'll set the record straight later.
	qc.State.InProgHTLC = incomingHTLC

	// they have to actually send you money
	if msg.Amt < consts.MinOutput {
		nd.FailChannel(qc)
		return fmt.Errorf("HashSigHandler err: HTLC amount %d less than minOutput", msg.Amt)
	}

	_, theirAmt := qc.GetChannelBalances()
	theirAmt -= int64(msg.Amt)

	// check if this push is takes them below minimum output size
	if theirAmt < consts.MinOutput {
		nd.FailChannel(qc)
		return fmt.Errorf(
			"making HTLC of size %s reduces them too low; counterparty bal %s fee %s consts.MinOutput %s",
			lnutil.SatoshiColor(int64(msg.Amt)),
			lnutil.SatoshiColor(theirAmt),
			lnutil.SatoshiColor(qc.State.Fee),
			lnutil.SatoshiColor(consts.MinOutput))
	}

	// update to the next state to verify
	qc.State.StateIdx++

	logging.Infof("Got message %x", msg.Data)
	qc.State.Data = msg.Data

	// verify sig for the next state. only save if this works
	curElk := qc.State.ElkPoint
	qc.State.ElkPoint = qc.State.NextElkPoint

	// TODO: There are more signatures required
	err = qc.VerifySigs(msg.CommitmentSignature, msg.HTLCSigs)
	if err != nil {
		nd.FailChannel(qc)
		return fmt.Errorf("HashSigHandler err %s", err.Error())
	}
	qc.State.ElkPoint = curElk

	// After verification of signatures, add back the clearing state in case
	// of HashSig-PreimageSig collisions
	for _, idx := range clearingIdxs {
		qh := &qc.State.HTLCs[idx]
		qh.Clearing = true
	}

	// (seems odd, but everything so far we still do in case of collision, so
	// only check here.  If it's a collision, set, save, send gapSigRev

	// save channel with new state, new sig, and positive delta set
	// and maybe collision; still haven't checked
	err = nd.SaveQchanState(qc)
	if err != nil {
		nd.FailChannel(qc)
		return fmt.Errorf("HashSigHandler SaveQchanState err %s", err.Error())
	}

	// If we are colliding Hashsig-Hashsig, determine who has what place in the
	// HTLC structure
	if collision && inProgHTLC != nil {
		curIdx := qc.State.InProgHTLC.Idx
		nextIdx := qc.State.HTLCIdx + 1

		if bytes.Compare(qc.State.MyNextHTLCBase[:], qc.State.NextHTLCBase[:]) > 0 {
			qc.State.CollidingHTLC = inProgHTLC
			qc.State.InProgHTLC = incomingHTLC
		} else {
			qc.State.CollidingHTLC = incomingHTLC
			qc.State.InProgHTLC = inProgHTLC
		}
		qc.State.InProgHTLC.Idx = curIdx
		qc.State.CollidingHTLC.Idx = nextIdx
		qc.State.CollidingHTLC.TheirHTLCBase = qc.State.N2HTLCBase
		qc.State.CollidingHTLC.KeyGen.Depth = 5
		qc.State.CollidingHTLC.KeyGen.Step[0] = 44 | 1<<31
		qc.State.CollidingHTLC.KeyGen.Step[1] = qc.Coin() | 1<<31
		qc.State.CollidingHTLC.KeyGen.Step[2] = UseHTLCBase
		qc.State.CollidingHTLC.KeyGen.Step[3] = qc.State.CollidingHTLC.Idx | 1<<31
		qc.State.CollidingHTLC.KeyGen.Step[4] = qc.Idx() | 1<<31

		qc.State.CollidingHTLC.MyHTLCBase, _ = nd.GetUsePub(qc.State.CollidingHTLC.KeyGen,
			UseHTLCBase)
	}

	var kg portxo.KeyGen
	kg.Depth = 5
	kg.Step[0] = 44 | 1<<31
	kg.Step[1] = qc.Coin() | 1<<31
	kg.Step[2] = UseHTLCBase
	kg.Step[3] = qc.State.HTLCIdx + 2 | 1<<31
	kg.Step[4] = qc.Idx() | 1<<31

	qc.State.MyNextHTLCBase = qc.State.MyN2HTLCBase
	qc.State.MyN2HTLCBase, err = nd.GetUsePub(kg, UseHTLCBase)
	if err != nil {
		nd.FailChannel(qc)
		return err
	}

	// save channel with new HTLCBases
	err = nd.SaveQchanState(qc)
	if err != nil {
		nd.FailChannel(qc)
		return fmt.Errorf("HashSigHandler SaveQchanState err %s", err.Error())
	}

	if qc.State.Collision != 0 || qc.State.CollidingHTLC != nil || qc.State.CollidingHashPreimage || qc.State.CollidingHashDelta {
		err = nd.SendGapSigRev(qc)
		if err != nil {
			nd.FailChannel(qc)
			return fmt.Errorf("HashSigHandler SendGapSigRev err %s", err.Error())
		}
	} else { // saved to db, now proceed to create & sign their tx
		err = nd.SendSigRev(qc)
		if err != nil {
			nd.FailChannel(qc)
			return fmt.Errorf("HashSigHandler SendSigRev err %s", err.Error())
		}
	}
	return nil
}

func (nd *LitNode) ClearHTLC(qc *Qchan, R [16]byte, HTLCIdx uint32, data [32]byte) error {
	if qc.State.Failed {
		return fmt.Errorf("cannot clear, channel failed")
	}

	// see if channel is busy
	// lock this channel
	cts := false
	for !cts {
		qc.ChanMtx.Lock()
		select {
		case <-qc.ClearToSend:
			cts = true
		default:
			qc.ChanMtx.Unlock()
		}
	}
	// ClearToSend is now empty

	// reload from disk here, after unlock
	err := nd.ReloadQchanState(qc)
	if err != nil {
		// don't clear to send here; something is wrong with the channel
		nd.FailChannel(qc)
		qc.ChanMtx.Unlock()
		return err
	}

	// check that channel is confirmed, if non-test coin
	wal, ok := nd.SubWallet[qc.Coin()]
	if !ok {
		qc.ClearToSend <- true
		qc.ChanMtx.Unlock()
		return fmt.Errorf("Not connected to coin type %d\n", qc.Coin())
	}

	if !wal.Params().TestCoin && qc.Height < 100 {
		qc.ClearToSend <- true
		qc.ChanMtx.Unlock()
		return fmt.Errorf(
			"height %d; must wait min 1 conf for non-test coin\n", qc.Height)
	}

	if int(HTLCIdx) >= len(qc.State.HTLCs) {
		qc.ClearToSend <- true
		qc.ChanMtx.Unlock()
		return fmt.Errorf("HTLC idx %d out of range", HTLCIdx)
	}

	if qc.State.HTLCs[HTLCIdx].Cleared {
		qc.ClearToSend <- true
		qc.ChanMtx.Unlock()
		return fmt.Errorf("HTLC %d already cleared", HTLCIdx)
	}

	var timeout bool
	if R == [16]byte{} {
		if int32(qc.State.HTLCs[HTLCIdx].Locktime) > wal.CurrentHeight() {
			qc.ClearToSend <- true
			qc.ChanMtx.Unlock()
			return fmt.Errorf("Cannot timeout HTLC because locktime %d has not expired. Height: %d", qc.State.HTLCs[HTLCIdx].Locktime, wal.CurrentHeight())
		}

		timeout = true
	}

	if !timeout {
		RHash := fastsha256.Sum256(R[:])
		if qc.State.HTLCs[HTLCIdx].RHash != RHash {
			qc.ClearToSend <- true
			qc.ChanMtx.Unlock()
			return fmt.Errorf("Preimage does not hash to expected value. Expected %x got %x", qc.State.HTLCs[HTLCIdx].RHash, RHash)
		}
	}

	// if we got here, but channel is not in rest state, try to fix it.
	if qc.State.Delta != 0 || qc.State.InProgHTLC != nil {
		nd.FailChannel(qc)
		qc.ChanMtx.Unlock()
		return fmt.Errorf("channel not in rest state")
	}

	qc.State.HTLCs[HTLCIdx].Clearing = true
	qc.State.HTLCs[HTLCIdx].R = R
	qc.State.Data = data

	// save to db with ONLY Clearing & R changed
	err = nd.SaveQchanState(qc)
	if err != nil {
		// don't clear to send here; something is wrong with the channel
		nd.FailChannel(qc)
		qc.ChanMtx.Unlock()
		return err
	}

	logging.Info("ClearHTLC: Sending PreimageSig")

	err = nd.SendPreimageSig(qc, HTLCIdx)
	if err != nil {
		nd.FailChannel(qc)
		qc.ChanMtx.Unlock()
		return err
	}

	logging.Info("got pre CTS...")
	qc.ChanMtx.Unlock()

	timeoutTimer := time.NewTimer(time.Second * consts.ChannelTimeout)

	cts = false
	for !cts {
		qc.ChanMtx.Lock()
		select {
		case <-qc.ClearToSend:
			cts = true
		case <-timeoutTimer.C:
			nd.FailChannel(qc)
			qc.ChanMtx.Unlock()
			return fmt.Errorf("channel failed: operation timed out")
		default:
			qc.ChanMtx.Unlock()
		}
	}

	logging.Info("got post CTS...")
	// since we cleared with that statement, fill it again before returning
	qc.ClearToSend <- true
	qc.ChanMtx.Unlock()

	return nil
}

func (nd *LitNode) SendPreimageSig(q *Qchan, Idx uint32) error {
	q.State.StateIdx++

	if q.State.HTLCs[Idx].Incoming != (q.State.HTLCs[Idx].R == [16]byte{}) {
		q.State.MyAmt += q.State.HTLCs[Idx].Amt
	}

	q.State.ElkPoint = q.State.NextElkPoint
	q.State.NextElkPoint = q.State.N2ElkPoint

	// make the signature to send over
	commitmentSig, HTLCSigs, err := nd.SignState(q)
	if err != nil {
		return err
	}

	outMsg := lnutil.NewPreimageSigMsg(q.Peer(), q.Op, Idx, q.State.HTLCs[Idx].R, commitmentSig, HTLCSigs, q.State.Data)

	logging.Infof("Sending PreimageSig with %d HTLC sigs", len(HTLCSigs))

	nd.tmpSendLitMsg(outMsg)

	return nil
}

func (nd *LitNode) PreimageSigHandler(msg lnutil.PreimageSigMsg, qc *Qchan) error {
	logging.Infof("Got PreimageSig: %v", msg)

	var collision bool

	// we should be clear to send when we get a preimageSig
	select {
	case <-qc.ClearToSend:
	// keep going, normal
	default:
		// collision
		collision = true
	}

	logging.Infof("COLLISION is (%t)\n", collision)

	// load state from disk
	err := nd.ReloadQchanState(qc)
	if err != nil {
		nd.FailChannel(qc)
		return fmt.Errorf("PreimageSigHandler ReloadQchan err %s", err.Error())
	}

	wal, ok := nd.SubWallet[qc.Coin()]
	if !ok {
		return fmt.Errorf("Not connected to coin type %d\n", qc.Coin())
	}

	// TODO we should send a response that the channel is closed.
	// or offer to double spend with a cooperative close?
	// or update the remote node on closed channel status when connecting
	// TODO should disallow 'break' command when connected to the other node
	// or merge 'break' and 'close' UI so that it breaks when it can't
	// connect, and closes when it can.
	if qc.CloseData.Closed {
		return fmt.Errorf("PreimageSigHandler err: %d, %d is closed.",
			qc.Peer(), qc.Idx())
	}

	clearingIdxs := make([]uint32, 0)
	for _, h := range qc.State.HTLCs {
		if h.Clearing {
			clearingIdxs = append(clearingIdxs, h.Idx)
		}
	}

	if qc.State.Delta > 0 {
		logging.Errorf(
			"PreimageSigHandler err: chan %d delta %d, expect rev, send empty rev",
			qc.Idx(), qc.State.Delta)

		return nd.SendREV(qc)
	}

	if int(msg.Idx) >= len(qc.State.HTLCs) {
		return fmt.Errorf("HTLC Idx %d out of range", msg.Idx)
	}

	if qc.State.HTLCs[msg.Idx].Cleared {
		return fmt.Errorf("HTLC %d already cleared", msg.Idx)
	}

	var timeout bool
	if msg.R == [16]byte{} {
		if int32(qc.State.HTLCs[msg.Idx].Locktime) > wal.CurrentHeight() {
			return fmt.Errorf("Cannot timeout HTLC because locktime %d has not expired. Height: %d", qc.State.HTLCs[msg.Idx].Locktime, wal.CurrentHeight())
		}

		timeout = true
	}

	RHash := fastsha256.Sum256(msg.R[:])
	if !timeout {
		if qc.State.HTLCs[msg.Idx].RHash != RHash {
			return fmt.Errorf("Preimage does not hash to expected value. Expected %x got %x", qc.State.HTLCs[msg.Idx].RHash, RHash)
		}
	}

	go func() {
		txids, err := nd.ClaimHTLC(msg.R)
		if err != nil {
			logging.Errorf("error claiming HTLCs: %s", err.Error())
		}

		if len(txids) == 0 {
			logging.Infof("found no other HTLCs to claim with R: %x, RHash: %x", msg.R, RHash)
		}

		for _, id := range txids {
			logging.Infof("claimed HTLC with txid: %x", id)
		}
	}()

	inProgHTLC := qc.State.InProgHTLC
	if collision {
		if inProgHTLC != nil {
			// PreimageSig-HashSig collision. Temporarily remove inprog HTLC for
			// verifying the signature, then do a GapSigRev
			qc.State.InProgHTLC = nil
			qc.State.CollidingHashPreimage = true
		} else if len(clearingIdxs) > 0 {
			// PreimageSig-PreimageSig collision.
			// Remove the clearing state for signature verification and
			// add back afterwards.
			for _, idx := range clearingIdxs {
				qh := &qc.State.HTLCs[idx]
				qh.Clearing = false
			}
			qc.State.CollidingPreimages = true
		} else {
			// PreimageSig-DeltaSig collision. Figure out later.
			qc.State.CollidingPreimageDelta = true
		}
	}

	// update to the next state to verify
	qc.State.StateIdx++

	logging.Infof("Got message %x", msg.Data)
	qc.State.Data = msg.Data

	h := &qc.State.HTLCs[msg.Idx]

	h.Clearing = true
	h.R = msg.R

	if h.Incoming != timeout {
		qc.State.MyAmt += h.Amt
	}

	// verify sig for the next state. only save if this works

	stashElk := qc.State.ElkPoint
	qc.State.ElkPoint = qc.State.NextElkPoint
	// TODO: There are more signatures required
	err = qc.VerifySigs(msg.CommitmentSignature, msg.HTLCSigs)
	if err != nil {
		nd.FailChannel(qc)
		return fmt.Errorf("PreimageSigHandler err %s", err.Error())
	}
	qc.State.ElkPoint = stashElk

	qc.State.InProgHTLC = inProgHTLC

	// After verification of signatures, add back the clearing state in case
	// of PreimageSig-PreimageSig collisions
	for _, idx := range clearingIdxs {
		qh := &qc.State.HTLCs[idx]
		qh.Clearing = true
	}

	if qc.State.CollidingHashPreimage {
		var kg portxo.KeyGen
		kg.Depth = 5
		kg.Step[0] = 44 | 1<<31
		kg.Step[1] = qc.Coin() | 1<<31
		kg.Step[2] = UseHTLCBase
		kg.Step[3] = qc.State.HTLCIdx + 2 | 1<<31
		kg.Step[4] = qc.Idx() | 1<<31

		qc.State.MyNextHTLCBase = qc.State.MyN2HTLCBase
		qc.State.MyN2HTLCBase, err = nd.GetUsePub(kg,
			UseHTLCBase)

		if err != nil {
			nd.FailChannel(qc)
			return err
		}
	}

	// (seems odd, but everything so far we still do in case of collision, so
	// only check here.  If it's a collision, set, save, send gapSigRev

	// save channel with new state, new sig, and positive delta set
	// and maybe collision; still haven't checked
	err = nd.SaveQchanState(qc)
	if err != nil {
		nd.FailChannel(qc)
		return fmt.Errorf("PreimageSigHandler SaveQchanState err %s", err.Error())
	}

	if qc.State.Collision != 0 || qc.State.CollidingHashPreimage || qc.State.CollidingPreimages || qc.State.CollidingPreimageDelta {
		err = nd.SendGapSigRev(qc)
		if err != nil {
			nd.FailChannel(qc)
			return fmt.Errorf("PreimageSigHandler SendGapSigRev err %s", err.Error())
		}
	} else { // saved to db, now proceed to create & sign their tx
		err = nd.SendSigRev(qc)
		if err != nil {
			nd.FailChannel(qc)
			return fmt.Errorf("PreimageSigHandler SendSigRev err %s", err.Error())
		}
	}
	return nil
}

func (nd *LitNode) SetHTLCClearedOnChain(q *Qchan, h HTLC) error {
	q.ChanMtx.Lock()
	err := nd.ReloadQchanState(q)
	if err != nil {
		logging.Errorf("Error reloading qchan state: %s", err.Error())
		q.ChanMtx.Unlock()
		return err
	}
	qh := &q.State.HTLCs[h.Idx]
	qh.ClearedOnChain = true
	err = nd.SaveQchanState(q)
	if err != nil {
		logging.Errorf("Error saving qchan state: %s", err.Error())
		q.ChanMtx.Unlock()
		return err
	}
	q.ChanMtx.Unlock()

	return nil
}

// ClaimHTLC will claim an HTLC on-chain output from a broken channel using
// the given preimage. Returns the TXIDs of the claim transactions
func (nd *LitNode) ClaimHTLC(R [16]byte) ([][32]byte, error) {
	txids := make([][32]byte, 0)
	RHash := fastsha256.Sum256(R[:])
	htlcs, channels, err := nd.FindHTLCsByHash(RHash)
	if err != nil {
		return nil, err
	}
	for i, h := range htlcs {

		// Outgoing HTLCs should not be claimed using the preimage, but
		// using the timeout. So only claim incoming ones in this routine
		if h.Incoming && !h.Cleared {
			q := channels[i]
			if q.CloseData.Closed {
				copy(h.R[:], R[:])
				tx, err := nd.ClaimHTLCOnChain(q, h)
				if err != nil {
					logging.Errorf("Error claiming HTLC: %s", err.Error())
					continue
				}
				nd.SetHTLCClearedOnChain(q, h)
				txids = append(txids, tx.TxHash())
			} else {
				// For off-chain we need to fetch the channel from the node
				// otherwise we're talking to a different instance of the channel
				nd.RemoteMtx.Lock()
				peer, ok := nd.RemoteCons[q.Peer()]

				nd.RemoteMtx.Unlock()
				if !ok {
					logging.Errorf("Couldn't find peer %d in RemoteCons", q.Peer())
					continue
				}
				qc, ok := peer.QCs[q.Idx()]
				if !ok {
					logging.Errorf("Couldn't find channel %d in peer.QCs", q.Idx())
					continue
				}

				logging.Infof("Cleaing HTLC from channel [%d] idx [%d]\n", q.Idx(), h.Idx)
				err = nd.ClearHTLC(qc, R, h.Idx, [32]byte{})
				if err != nil {
					logging.Errorf("failed to clear HTLC: %s", err.Error())
					continue
				}
			}

			nd.MultihopMutex.Lock()
			defer nd.MultihopMutex.Unlock()
			for idx, mu := range nd.InProgMultihop {
				if bytes.Equal(mu.HHash[:], RHash[:]) && !mu.Succeeded {
					nd.InProgMultihop[idx].Succeeded = true
					nd.InProgMultihop[idx].PreImage = R
					err = nd.SaveMultihopPayment(nd.InProgMultihop[idx])
					if err != nil {
						return txids, err
					}
				}
			}
		}
	}
	return txids, nil
}

func (nd *LitNode) ClaimHTLCTimeouts(coinType uint32, height int32) ([][32]byte, error) {
	txids := make([][32]byte, 0)
	htlcs, channels, err := nd.FindHTLCsByTimeoutHeight(coinType, height)
	if err != nil {
		return nil, err
	}
	if len(htlcs) > 0 {
		logging.Infof("Found [%d] HTLC Outpoints that have timed out\n", len(htlcs))
		for i, h := range htlcs {
			if !h.Incoming { // only for timed out HTLCs!
				q := channels[i]
				if q.CloseData.Closed {
					tx, err := nd.ClaimHTLCOnChain(q, h)
					if err != nil {
						logging.Errorf("Error claiming HTLC: %s", err.Error())
						continue
					}
					nd.SetHTLCClearedOnChain(q, h)
					txids = append(txids, tx.TxHash())
				} else {
					// For off-chain we need to fetch the channel from the node
					// otherwise we're talking to a different instance of the channel
					nd.RemoteMtx.Lock()
					peer, ok := nd.RemoteCons[q.Peer()]

					nd.RemoteMtx.Unlock()
					if !ok {
						return nil, fmt.Errorf("Couldn't find peer %d in RemoteCons", q.Peer())
					}
					qc, ok := peer.QCs[q.Idx()]
					if !ok {
						return nil, fmt.Errorf("Couldn't find channel %d in peer.QCs", q.Idx())
					}
					logging.Infof("Timing out HTLC from channel [%d] idx [%d]\n", q.Idx(), h.Idx)
					err = nd.ClearHTLC(qc, [16]byte{}, h.Idx, [32]byte{})
					if err != nil {
						logging.Errorf("error clearing HTLC: %s", err.Error())
						continue
					}
				}
			}
		}
	}
	return txids, nil
}

func (nd *LitNode) FindHTLCsByTimeoutHeight(coinType uint32, height int32) ([]HTLC, []*Qchan, error) {
	htlcs := make([]HTLC, 0)
	channels := make([]*Qchan, 0)
	qc, err := nd.GetAllQchans()
	if err != nil {
		return nil, nil, err
	}
	for _, q := range qc {
		err := nd.ReloadQchanState(q)
		if err != nil {
			return nil, nil, err
		}
		if q.Coin() == coinType {
			for _, h := range q.State.HTLCs {
				if !h.Incoming && !h.Cleared {
					if height >= int32(h.Locktime) {
						htlcs = append(htlcs, h)
						channels = append(channels, q)
					} else {
						logging.Infof("Ignoring HTLC in chan [%d] idx [%d] - expires at block [%d] (now: %d)", q.Idx(), h.Idx, h.Locktime, height)
					}
				}
			}
		}
	}
	return htlcs, channels, nil
}

func (nd *LitNode) FindHTLCsByHash(hash [32]byte) ([]HTLC, []*Qchan, error) {
	htlcs := make([]HTLC, 0)
	channels := make([]*Qchan, 0)
	qc, err := nd.GetAllQchans()
	if err != nil {
		return nil, nil, err
	}
	for _, q := range qc {
		for _, h := range q.State.HTLCs {
			if bytes.Equal(h.RHash[:], hash[:]) {
				htlcs = append(htlcs, h)
				channels = append(channels, q)
			}
		}
	}
	return htlcs, channels, nil
}

func (nd *LitNode) GetHTLC(op *wire.OutPoint) (HTLC, *Qchan, error) {
	var empty HTLC
	qc, err := nd.GetAllQchans()
	if err != nil {
		return empty, nil, err
	}
	for _, q := range qc {
		tx, _, _, err := q.BuildStateTxs(false)
		if err != nil {
			return empty, nil, err
		}
		for _, h := range q.State.HTLCs {
			txid := tx.TxHash()
			_, i, err := GetHTLCOut(q, h, tx, false)
			if err != nil {
				return empty, nil, err
			}
			hashOp := wire.NewOutPoint(&txid, i)
			if lnutil.OutPointsEqual(*op, *hashOp) {
				return h, q, nil
			}
		}
	}
	return empty, nil, nil
}

func GetHTLCOut(q *Qchan, h HTLC, tx *wire.MsgTx, mine bool) (*wire.TxOut, uint32, error) {
	for i, out := range tx.TxOut {
		htlcOut, err := q.GenHTLCOut(h, mine)
		if err != nil {
			return nil, 0, err
		}
		if bytes.Compare(out.PkScript, htlcOut.PkScript) == 0 {
			return out, uint32(i), nil
		}
	}

	return nil, 0, fmt.Errorf("Could not find HTLC output with desired PkScript")
}

func (q *Qchan) GetCloseTxs() (*wire.MsgTx, []*wire.MsgTx, bool, error) {
	for i, h := range q.State.HTLCs {
		if !h.Cleared && h.Clearing {
			q.State.HTLCs[i].Clearing = false
		}
	}

	q.State.InProgHTLC = nil
	q.State.CollidingHTLC = nil

	mine := true
	stateTx, htlcSpends, _, err := q.BuildStateTxs(mine)
	if err != nil {
		return nil, nil, false, err
	}

	stateTxID := stateTx.TxHash()
	if !q.CloseData.CloseTxid.IsEqual(&stateTxID) {
		mine = false
		stateTx, htlcSpends, _, err = q.BuildStateTxs(mine)
		if err != nil {
			return nil, nil, false, err
		}

		stateTxID = stateTx.TxHash()

		if !q.CloseData.CloseTxid.IsEqual(&stateTxID) {
			return nil, nil, false, fmt.Errorf("Could not find/regenerate proper close TX")
		}
	}
	return stateTx, htlcSpends, mine, nil
}

func (nd *LitNode) ClaimHTLCOnChain(q *Qchan, h HTLC) (*wire.MsgTx, error) {
	wal, ok := nd.SubWallet[q.Coin()]
	if !ok {
		return nil, fmt.Errorf("Unable to find wallet for cointype [%d]", q.Coin())
	}

	if !h.Incoming {
		unlockHeight := int32(h.Locktime)
		if wal.CurrentHeight() < unlockHeight {
			err := fmt.Errorf("Trying to claim timeout before timelock expires - wait until height %d", unlockHeight)
			logging.Error(err.Error())
			return nil, err
		}

	}

	stateTx, htlcSpends, mine, err := q.GetCloseTxs()
	if err != nil {
		return nil, err
	}
	stateTxID := stateTx.TxHash()

	htlcTxo, i, err := GetHTLCOut(q, h, stateTx, mine)
	if err != nil {
		return nil, err
	}

	curElk, err := q.ElkSnd.AtIndex(q.State.StateIdx)
	if err != nil {
		return nil, err
	}
	elkScalar := lnutil.ElkScalar(curElk)
	HTLCPrivBase, err := wal.GetPriv(h.KeyGen)
	if err != nil {
		return nil, err
	}

	HTLCPriv := lnutil.CombinePrivKeyWithBytes(HTLCPrivBase, elkScalar[:])
	tx := wire.NewMsgTx()
	op := wire.NewOutPoint(&stateTxID, i)
	if mine {
		for _, s := range htlcSpends {
			if lnutil.OutPointsEqual(*op, s.TxIn[0].PreviousOutPoint) {
				tx = s
				break
			}
		}
	} else {
		// They broke.
		// Claim back to my wallet, i only need my sig & preimage (for success)
		// or just my sig when timed out.

		tx.Version = 2
		tx.LockTime = 0

		in := wire.NewTxIn(op, nil, nil)
		in.Sequence = 0

		tx.AddTxIn(in)

		pkh, err := wal.NewAdr()
		if err != nil {
			return nil, err
		}

		if !h.Incoming {
			tx.LockTime = h.Locktime
		}

		tx.AddTxOut(wire.NewTxOut(h.Amt-q.State.Fee, lnutil.DirectWPKHScriptFromPKH(pkh)))
	}
	hc := txscript.NewTxSigHashes(tx)

	HTLCScript, err := q.GenHTLCScript(h, mine)
	if err != nil {
		return nil, err
	}

	HTLCparsed, err := txscript.ParseScript(HTLCScript)
	if err != nil {
		return nil, err
	}

	spendHTLCHash := txscript.CalcWitnessSignatureHash(
		HTLCparsed, hc, txscript.SigHashAll, tx, 0, htlcTxo.Value)

	logging.Infof("Signing HTLC Hash [%x] with key [%x]\n", spendHTLCHash, HTLCPriv.PubKey().SerializeCompressed())
	mySig, err := HTLCPriv.Sign(spendHTLCHash)
	if err != nil {
		return nil, err
	}

	myBigSig := append(mySig.Serialize(), byte(txscript.SigHashAll))
	theirBigSig := sig64.SigDecompress(h.Sig)
	theirBigSig = append(theirBigSig, byte(txscript.SigHashAll))

	if mine {
		tx.TxIn[0].Witness = make([][]byte, 5)
		tx.TxIn[0].Witness[0] = nil
		tx.TxIn[0].Witness[1] = theirBigSig
		tx.TxIn[0].Witness[2] = myBigSig
		if h.Incoming {
			tx.TxIn[0].Witness[3] = h.R[:]
		} else {
			tx.TxIn[0].Witness[3] = nil
		}
		tx.TxIn[0].Witness[4] = HTLCScript
	} else {
		tx.TxIn[0].Witness = make([][]byte, 3)
		tx.TxIn[0].Witness[0] = myBigSig

		if h.Incoming {
			tx.TxIn[0].Witness[1] = h.R[:]
		} else {
			tx.TxIn[0].Witness[1] = nil
		}

		tx.TxIn[0].Witness[2] = HTLCScript
	}

	logging.Debug("Claiming HTLC on-chain. TX:")
	lnutil.PrintTx(tx)

	wal.StopWatchingThis(*op)
	wal.DirectSendTx(tx)

	if mine {
		// TODO: Refactor this into a function shared with close.go's GetCloseTxos
		// Store the timeout utxo into the wallit
		comNum := GetStateIdxFromTx(stateTx, q.GetChanHint(true))

		theirElkPoint, err := q.ElkPoint(false, comNum)
		if err != nil {
			return nil, err
		}

		// build script to store in porTxo, make pubkeys
		timeoutPub := lnutil.AddPubsEZ(q.MyHAKDBase, theirElkPoint)
		revokePub := lnutil.CombinePubs(q.TheirHAKDBase, theirElkPoint)

		script := lnutil.CommitScript(revokePub, timeoutPub, q.Delay)
		// script check.  redundant / just in case
		genSH := fastsha256.Sum256(script)
		if !bytes.Equal(genSH[:], tx.TxOut[0].PkScript[2:34]) {
			logging.Warnf("got different observed and generated SH scripts.\n")
			logging.Warnf("in %s:%d, see %x\n", tx.TxHash().String(), 0, tx.TxOut[0].PkScript)
			logging.Warnf("generated %x \n", genSH)
			logging.Warnf("revokable pub %x\ntimeout pub %x\n", revokePub, timeoutPub)
			return tx, nil
		}

		// create the ScriptHash, timeout portxo.
		var shTxo portxo.PorTxo // create new utxo and copy into it
		// use txidx's elkrem as it may not be most recent
		elk, err := q.ElkSnd.AtIndex(comNum)
		if err != nil {
			return nil, err
		}
		// keypath is the same, except for use
		shTxo.KeyGen = q.KeyGen
		shTxo.Op.Hash = tx.TxHash()
		shTxo.Op.Index = 0
		shTxo.Height = q.CloseData.CloseHeight
		shTxo.KeyGen.Step[2] = UseChannelHAKDBase

		elkpoint := lnutil.ElkPointFromHash(elk)
		addhash := chainhash.DoubleHashH(append(elkpoint[:], q.MyHAKDBase[:]...))

		shTxo.PrivKey = addhash

		shTxo.Mode = portxo.TxoP2WSHComp
		shTxo.Value = tx.TxOut[0].Value
		shTxo.Seq = uint32(q.Delay)
		shTxo.PreSigStack = make([][]byte, 1) // timeout has one presig item
		shTxo.PreSigStack[0] = nil            // and that item is a nil (timeout)

		shTxo.PkScript = script

		wal.ExportUtxo(&shTxo)
	}

	return tx, nil
}
