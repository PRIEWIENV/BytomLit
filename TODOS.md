# TODO

And also try `grep` with "TODO:"

+ fix coinparam
    * fields
        - do we need to add merkle tx status? 
    * bytom param
        - rename to bytom?
+ do we need to worry about bytom's curve
+ link wallet
    * fix resync
    + __sync bytom height?__
        * fix SPVCon.Connect
        * fix wallit
        * fix SPVCon.incomingMessageHandler
    + `bech32.Encode("ln"` can be a good start point
+ failure loading exchange rates: open rates.json: no such file or directory