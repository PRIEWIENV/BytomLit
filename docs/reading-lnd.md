# Reading LND

## chain sync
+ https://github.com/lightningnetwork/lnd/blob/master/lnd.go#L475-L519

## peer connections
Set up the core server which will listen for incoming peer connections.

+ https://github.com/lightningnetwork/lnd/blob/master/lnd.go#L419
    * -> https://github.com/lightningnetwork/lnd/blob/master/server.go#L291
        - connect to peer
            + https://github.com/lightningnetwork/lnd/blob/master/server.go#L1976
            + bootstrapping
                * https://github.com/lightningnetwork/lnd/blob/master/server.go#L1299-L1306
                    - -> https://github.com/lightningnetwork/lnd/blob/master/server.go#L1568-L1847
                        + -> https://github.com/lightningnetwork/lnd/blob/master/server.go#L1718
                            * -> https://github.com/lightningnetwork/lnd/blob/master/server.go#L3084
        - routing
        - path finding
        - penalty
        - payment
        - funding
        - gossip???
        - callback for closing conn
            + https://github.com/lightningnetwork/lnd/blob/master/server.go#L413-L434
                * disconn peer
                    - https://github.com/lightningnetwork/lnd/blob/master/server.go#L3104
        - NAT
            + discover
                * https://github.com/lightningnetwork/lnd/blob/master/server.go#L470
            + set up port forwarding
                * https://github.com/lightningnetwork/lnd/blob/master/server.go#L523
                    - -> https://github.com/lightningnetwork/lnd/blob/master/server.go#L1389

## RPC server
Initialize, and register the gRPC server.

+ https://github.com/lightningnetwork/lnd/blob/master/lnd.go#L454


## channel related
### open
+ https://github.com/lightningnetwork/lnd/blob/master/server.go#L3142


## wallets

