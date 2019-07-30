# Reading LND

## chain sync
+ https://github.com/lightningnetwork/lnd/blob/master/lnd.go#L475-L519

## peer connections
Set up the core server which will listen for incoming peer connections.

+ https://github.com/lightningnetwork/lnd/blob/master/lnd.go#L419
    * -> https://github.com/lightningnetwork/lnd/blob/master/server.go#L291

## RPC server
Initialize, and register our implementation of the gRPC interface exported by the rpcServer.

+ https://github.com/lightningnetwork/lnd/blob/master/lnd.go#L454


## channel related


## wallets

