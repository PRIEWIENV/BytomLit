package qln

import (
	"encoding/json"
	"sync"

	"github.com/mit-dci/lit/elkrem"
	"github.com/mit-dci/lit/portxo"

	"github.com/getlantern/deepcopy"
)

/*
note that sigs are truncated and don't have the sighash type byte at the end.

their rev hash can be derived from the elkrem sender
and the stateidx.  hash160(elkremsend(sIdx)[:16])

TODO Does it make sense to move this comment somewhere else?  It's not really
relevant here anymore.
*/

// ChanData is a struct used as a surrogate for going between Qchan and the JSON
// representation stored on disk.  There's an annoying layer in between that
// make it all work out with hopefully little friction.
type ChanData struct {
	Txo       portxo.PorTxo `json:"txo"`
	CloseData QCloseData    `json:"closedata"`

	TheirPub       [33]byte `json:"rpub"`
	TheirRefundPub [33]byte `json:"rrefpub"`
	TheirHAKDBase  [33]byte `json:"rhakdbase"`

	State *StatCom `json:"state"`

	LastUpdate uint64 `json:"updateunix"`
}

// NewQchanFromChanData creates a new qchan from a chandata.
func (nd *LitNode) NewQchanFromChanData(data *ChanData) (*Qchan, error) {

	sc := new(StatCom)
	deepcopy.Copy(sc, data.State)

	// I don't know if these errors are supposed to be ignored but that's what
	// other code does so I'm just copying that.
	mp, _ := nd.GetUsePub(data.Txo.KeyGen, UseChannelFund)
	mrp, _ := nd.GetUsePub(data.Txo.KeyGen, UseChannelRefund)
	mhb, _ := nd.GetUsePub(data.Txo.KeyGen, UseChannelHAKDBase)
	elkroot, _ := nd.GetElkremRoot(data.Txo.KeyGen)

	qc := &Qchan{
		PorTxo:    data.Txo,
		CloseData: data.CloseData,

		MyPub:          mp,
		TheirPub:       data.TheirPub,
		MyRefundPub:    mrp,
		TheirRefundPub: data.TheirRefundPub,
		MyHAKDBase:     mhb,
		TheirHAKDBase:  data.TheirHAKDBase,

		ElkSnd: elkrem.NewElkremSender(elkroot),
		ElkRcv: elkrem.NewElkremReceiver(),

		Delay: 5, // This is defined to just be 5.

		State: sc,

		ClearToSend: make(chan bool, 1),
		ChanMtx:     sync.Mutex{},

		LastUpdate: data.LastUpdate,
	}

	// I think this might fix the problem?
	go (func() {
		qc.ClearToSend <- true
	})()

	return qc, nil

}

// NewChanDataFromQchan extracts the chandata from the qchan.
func (nd *LitNode) NewChanDataFromQchan(qc *Qchan) *ChanData {

	// We have to make copies of some thing because it's weird.
	sc := new(StatCom)
	deepcopy.Copy(sc, qc.State)

	cd := &ChanData{
		Txo:            qc.PorTxo,
		CloseData:      qc.CloseData,
		TheirPub:       qc.TheirPub,
		TheirRefundPub: qc.TheirRefundPub,
		TheirHAKDBase:  qc.TheirHAKDBase,
		State:          sc,
		LastUpdate:     qc.LastUpdate,
	}

	return cd

}

// ApplyChanDataToQchan applies the chandata to the qchan without destroying it.
func (nd *LitNode) ApplyChanDataToQchan(data *ChanData, qc *Qchan) error {

	fake, err := nd.NewQchanFromChanData(data)
	if err != nil {
		return err
	}

	// now just copy them over.  this is bad but it should work well enough.
	qc.PorTxo = fake.PorTxo
	qc.CloseData = fake.CloseData
	qc.MyPub = fake.MyPub
	qc.TheirPub = fake.TheirPub
	qc.MyRefundPub = fake.MyRefundPub
	qc.TheirRefundPub = fake.TheirRefundPub
	qc.MyHAKDBase = fake.MyHAKDBase
	qc.TheirHAKDBase = fake.TheirHAKDBase
	qc.WatchRefundAdr = fake.WatchRefundAdr
	qc.ElkSnd = fake.ElkSnd
	qc.ElkRcv = fake.ElkRcv
	qc.Delay = fake.Delay
	qc.State = fake.State
	qc.LastUpdate = fake.LastUpdate

	return nil

}

func (nd *LitNode) QchanSerializeToBytes(qc *Qchan) []byte {
	cd := nd.NewChanDataFromQchan(qc)
	data, _ := json.Marshal(cd)
	return data
}

func (nd *LitNode) QchanDeserializeFromBytes(buf []byte) (*Qchan, error) {

	var cd ChanData
	err := json.Unmarshal(buf, &cd)
	if err != nil {
		return nil, err
	}

	qc, err := nd.NewQchanFromChanData(&cd)
	if err != nil {
		return nil, err
	}

	return qc, nil

}

func (nd *LitNode) QchanUpdateFromBytes(qc *Qchan, buf []byte) error {

	var err error
	var cd ChanData
	err = json.Unmarshal(buf, &cd)
	if err != nil {
		return err
	}

	err = nd.ApplyChanDataToQchan(&cd, qc)
	if err != nil {
		return err
	}

	return nil

}
