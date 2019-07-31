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
            + util
                * BroadcastMessage
                    - https://github.com/lightningnetwork/lnd/blob/master/server.go#L2188
                * NotifyWhenOnline
                    - https://github.com/lightningnetwork/lnd/blob/master/server.go#L2237
                * NotifyWhenOffline
                    - https://github.com/lightningnetwork/lnd/blob/master/server.go#L2270
                * FindPeer
                    - https://github.com/lightningnetwork/lnd/blob/master/server.go#L2300
        - routing
            + https://github.com/lightningnetwork/lnd/blob/master/routing/router_test.go
            + https://github.com/lightningnetwork/lnd/blob/master/routing/route/route_test.go
        - path finding
            + https://github.com/lightningnetwork/lnd/blob/master/routing/pathfind_test.go
        - peer
            + func TestPeerChannelClosureAcceptFeeResponder(t *testing.T) {
            + func TestPeerChannelClosureAcceptFeeInitiator(t *testing.T) {
            + func TestPeerChannelClosureFeeNegotiationsResponder(t *testing.T) {
            + func TestPeerChannelClosureFeeNegotiationsInitiator(t *testing.T) {
        - penalty
        - payment
            + https://github.com/lightningnetwork/lnd/blob/master/routing/payment_lifecycle.go
        - funding
            + main logic
                * https://github.com/lightningnetwork/lnd/blob/master/fundingmanager.go#L903
                    - TODO
                * fundingmanager_test
            + ???
                * https://github.com/lightningnetwork/lnd/blob/master/fundingmanager.go#L487
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
+ https://github.com/lightningnetwork/lnd/blob/master/rpcserver.go
+ https://github.com/lightningnetwork/lnd/blob/master/lnrpc/

## c2c
+ https://github.com/lightningnetwork/lnd/blob/master/lntest/itest/lnd_test.go

## channel related
### open
+ https://github.com/lightningnetwork/lnd/blob/master/server.go#L3142


## wallets

## resolve
+ https://github.com/lightningnetwork/lnd/blob/master/contractcourt/

## multi-hop claim
+ https://github.com/lightningnetwork/lnd/blob/master/lntest/itest/
