# TODO

And also try `grep` with "TODO:"

+ fix coinparam
    * fields
        - do we need to add merkle tx status? 
    * bytom param
        - rename to bytom?
+ do we need to worry about bytom's curve
+ link wallet
    + fix wallit
    * fix resync
    * walkthrough
    * merkle
    * LitNode.AutoReconnect()
    + `bech32.Encode("ln"` can be a good start point
- lit-af
    + addresses
    + Type: 0   Sync Height: 0  FeeRate: 80 Utxo: 0 WitConf: 0 Channel: 0
+ failure loading exchange rates: open rates.json: no such file or directory
+ payment channel walkthrough in `qln`
    * open a channel: `fund.go`
    * send payment: `pushpull.go`
    * close a channel: `close.go` and `break.go`
    * htlc: `htlc.go`
    * multi-hop: `multihop.go`
