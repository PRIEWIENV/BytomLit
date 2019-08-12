# HOWTO

+ `make` 
+ `./node config.json` 
+ .
    ```
    curl -X POST 127.0.0.1:9000/build-tx -d '{
        "outputs":[
            {
                "program":"00148c9d063ff74ee6d9ffa88d83aeb038068366c4c4",
                "asset_id":"ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
                "amount":100000000
            },
            {
                "program":"00148c9d063ff74ee6d9ffa88d83aeb038068366c4c4",
                "asset_id":"ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
                "amount":200000000
            }
        ]
    }'
    ```