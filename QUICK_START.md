# Quick Start

## Build

```bash
cd BytomLit
make
``` 

## Run

```bash
# You can modify the config.json
./node config.json
```

## Usage

```bash
curl -X POST 127.0.0.1:9000/<api> -d '<parameter>'
```

JSON-RPC API List
+ `dual-fund`
  - String: fund_asset_id
  - Integer: fund_amount
  - String: peer_id
  - String: peer_asset_id
  - Integer peer_amount
+ `push`
  - String: asset_id
  - Integer: amount
  - String: peer_id
+ `close`: No parameter