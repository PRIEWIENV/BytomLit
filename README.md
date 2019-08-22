# An Anti-Wormhole Payment Channel Network on Bytom

A payment channel network daemon based on Bytom using Path Hashed Time Locked Contract (BytomLit) instead of Hashed Time Locked Contract (HTLC), which is robust against wormhole attack.

## Quick Start

### Build

```bash
cd BytomLit
make
``` 

### Run

```bash
# You can modify the config.json
./node config.json
```

### Usage

```bash
curl -X POST 127.0.0.1:9000/<api> -d '<parameter>'
```

JSON-RPC API List
+ `dual-fund`: Open a channel between `A` and `B`.
	- `inputs`: *JSON array [aInput, bInput, gasInput]*. Input from `A`, input from `B` and input for tx gas, respectively.
  	- `program`: *string*. Control program of the input.
  	- `source_id`: *string*.
  	- `source_pos`: *integer*.
  	- `asset_id`: *string*.
  	- `amount`: *integer*. Amount of the asset to be fund to open a channel.
  	- `arguments`: *string*.
	- `aPub`: *string*. Pubkey from `A` to lock the funding tx in the output.
	- `bPub`: *string*. Pubkey from `B` to lock the funding tx in the output.
+ `push`: Pay a certain amount of asset from `A` to `B` or from `B` to `A`.
  - `amount`: *integer*. The amount of the asset to be paid. Remark: `amount > 0: A -> B; amount < 0: B -> A`.
+ `close`: Close the established channel.
  - `receipt`: *string*. Close the established channel by submit receipt to Bytom chain.

## Demo

```bash
# Send dual-funding transaction to Bytom chain. The response is supposed to be
# a tx id.
curl -X POST 127.0.0.1:9000/dual-fund -d '{
  "inputs": [
    {
      "source_id":"dcdd21f5775d8205519204a9e8380632ba6d255f5d9f83640f7f00fe0414c942",
      "source_pos": 2,
      "program": "0014f077b8a83998adfa8df7c529e8643cfebce2dff8",
      "asset_id": "f08f0da2b982fdc7aab517de724be5e5eed1c49330826501c88a261ae9cb0edf",
      "amount": 10000000000000000
    },
    {
      "source_id": "dcdd21f5775d8205519204a9e8380632ba6d255f5d9f83640f7f00fe0414c942",
      "source_pos": 1,
      "program": "001472e49786aea9ae75a5ec4543259b6d10c2c4f57d",
      "asset_id": "f08f0da2b982fdc7aab517de724be5e5eed1c49330826501c88a261ae9cb0edf",
      "amount": 10000000000000000
    },
    {
      "source_id": "d9b60b8b3d1e3d249b0efaefddd50a2bcc846f5462a2903c228d3c39b6dfcdf3",
      "source_pos": 0,
      "program": "001472e49786aea9ae75a5ec4543259b6d10c2c4f57d",
      "asset_id": "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
      "amount": 41250000000
    }
  ],
  "pubkey_a": "b7e5e40c0de6d4cd0048968f047f1ed05215e04e03b7ce22f92ade9ff0791c5d",
  "pubkey_b": "343132656a747d98a40488fcd68670f6723abb1f29dfaba36a3b6af18c6360d4"
}'

	// DEBUG
	// aInputKeys := []key{
	// 	key{
	// 		XPub: "e0446ee8c0f0d559d6eeaeddf5b676ff89de4cdd0477f69f2662269f5e4a6e43d7e6aefa9547408839bc6796f9d22e6865c796d652027a966f1da46a34b94e78",
	// 		DerivPath: []string{
	// 			"2c000000",
	// 			"99000000",
	// 			"01000000",
	// 			"00000000",
	// 			"02000000",
	// 		},
	// 	},
	// }
	// bInputKeys := []key{
	// 	key{
	// 		XPub: "e0446ee8c0f0d559d6eeaeddf5b676ff89de4cdd0477f69f2662269f5e4a6e43d7e6aefa9547408839bc6796f9d22e6865c796d652027a966f1da46a34b94e78",
	// 		DerivPath: []string{
	// 			"2c000000",
	// 			"99000000",
	// 			"01000000",
	// 			"00000000",
	// 			"01000000",
	// 		},
	// 	},
	// }
	// gasInputKeys := []key{
	// 	key{
	// 		XPub: "e0446ee8c0f0d559d6eeaeddf5b676ff89de4cdd0477f69f2662269f5e4a6e43d7e6aefa9547408839bc6796f9d22e6865c796d652027a966f1da46a34b94e78",
	// 		DerivPath: []string{
	// 			"2c000000",
	// 			"99000000",
	// 			"01000000",
	// 			"00000000",
	// 			"01000000",
	// 		},
	// 	},
	// }
	// DEBUG: ============

# Send several BTLs (a user-defined asset for test) to another address. The
# response is supposed to be a signed raw transaction
curl -X POST 127.0.0.1:9000/push -d '{"amount": 0}'

aPub := "b7e5e40c0de6d4cd0048968f047f1ed05215e04e03b7ce22f92ade9ff0791c5d"
bPub := "343132656a747d98a40488fcd68670f6723abb1f29dfaba36a3b6af18c6360d4"
fundSourceID, gasSourceID := btmBc.Hash{}, btmBc.Hash{}
fundSourceID.UnmarshalText([]byte("7423542dade2528182812b199eafedc8cb013f04dcf62ddae0c4ef207bfd4e8a"))
gasSourceID.UnmarshalText([]byte("409fa556dad4ab99f1cedf78656b9221231aa70f06cb531e4df068127598582e"))
fundInput := inputType{
	SourceID: fundSourceID,
	SourcePos: 1,
	Program: "5a20343132656a747d98a40488fcd68670f6723abb1f29dfaba36a3b6af18c6360d420b7e5e40c0de6d4cd0048968f047f1ed05215e04e03b7ce22f92ade9ff0791c5d7424537a641b000000537a547a526bae547a547a526c7cad63240000007bcd9f697b7cae7cac00c0",
	AssetID: "f08f0da2b982fdc7aab517de724be5e5eed1c49330826501c88a261ae9cb0edf",
	Amount: 20000000000000000,
}
gasInput := inputType{
	SourceID: gasSourceID,
	SourcePos: 0,
	Program: "001472e49786aea9ae75a5ec4543259b6d10c2c4f57d",
	AssetID: "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
	Amount: 41250000000,
}

# Send the close-channel transaction (latest signed raw transaction got from
# "/push") to Bytom chain. The response is supposed to be a tx id.
curl -X POST 127.0.0.1:9000/close -d '{"receipt":"0701000201b90101b6017423542dade2528182812b199eafedc8cb013f04dcf62ddae0c4ef207bfd4e8af08f0da2b982fdc7aab517de724be5e5eed1c49330826501c88a261ae9cb0edf808084fea6dee11101016b5a20343132656a747d98a40488fcd68670f6723abb1f29dfaba36a3b6af18c6360d420b7e5e40c0de6d4cd0048968f047f1ed05215e04e03b7ce22f92ade9ff0791c5d7424537a641b000000537a547a526bae547a547a526c7cad63240000007bcd9f697b7cae7cac00c0c5010440fd083f7923f88d5a3d427e6519d149573d34fcdbf18d583ecd26d5ac2dba198b2cd5a455140dea12746a80df80daf6312173941fe4d4d28aadeb72549f04140240f3fae7a1734cb144c75e5d370ffcf42c746ce9008d0121551751da06d0ebbb30d0886fbf0966b713572350afb10b537a4585252cc2065f9b8dcbd939e2f1c10c2018cd420713da2b5075f5282dc0ab8abd32e0ad0ec611ddc936d866e04297310120d1a80162ad4c529000196b1c44d8bcb07b045190779648a1441e31d086d2e71d0161015f409fa556dad4ab99f1cedf78656b9221231aa70f06cb531e4df068127598582effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff8099c4d59901000116001472e49786aea9ae75a5ec4543259b6d10c2c4f57d6302400903027dc48f4352d08169be7cf7d44e6cf5e2f373d9f666bbd4729ee52c6751c69d52dc868992dae1900a814adf3fcbd41a87a64b4a9c677f93119cf7f59c0020d1a80162ad4c529000196b1c44d8bcb07b045190779648a1441e31d086d2e71d020140f08f0da2b982fdc7aab517de724be5e5eed1c49330826501c88a261ae9cb0edf808084fea6dee11101160014a796b852f5db234d4450f80260e5640faf3808ce00013effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff80ece1d0990101160014a796b852f5db234d4450f80260e5640faf3808ce00"}'
```