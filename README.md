# eth scan service
 - Scan block from n to latest and then store block include transaction info to db.  
 - Subscribe for new block event and then store block include transaction info to db.

# api service
 - Api service provide api to query blocks info and transaction info.

## run
This command will run two containers locally , database and service server.
```
make run 
```


## clean
To clean running containers
```
make clean
```


## config
- Docker compose
Config can set in /devenv/docker-compose.yml  
environment section 

- Local go run
Config can set in /localenv/localrc


#### rpc , ws endpoint
configure rpc ws endpoint for different chain or provider
```
ex:
      JSON_RPC_ENDPOINT: https://mainnet.infura.io/v3/49c81384a9ed44f1bcdb04c5efbc776f
      WS_ENDPOINT: wss://mainnet.infura.io/ws/v3/49c81384a9ed44f1bcdb04c5efbc776f
```

#### worker num 
param : SCAN_WORK_NUM (uint32)
- Configure scan worker num to adjust speed of scan block process.
- Note: default rpc/ws endpoint is free trial . Too many worker num may exceed ratelimit of infura.

#### worker num
param : WRITE_TRANSACTION_WORK_NUM (uint32)
- Configure transaction worker num to adjust speed of scan block process.
- Note: default rpc/ws endpoint is free trial . Too many worker num may exceed ratelimit of infura.


#### fetch block from N
param : SYNC_BLOCK_FROM_N (uint64)
- Configure this number to tell service fetch block from which block number.



#### stable block num
param : CONFIRMED_BLOCK_NUM (uint64)
- There are some fork situation happened commonly .
- We usually give a number to assume pass through such count blocks , this block define as stable one.


## API 
### Get Transaction Info
[Get] /transaction/:txHash
```
ex:
curl --location --request GET 'http://localhost:8080/transaction/0xd276699999cb630c2667dd240496c7237cd2218e16e1a1d47299ae86a14427a2'
```

### Get Block Info
[Get] /blocks/:id
```
ex:
curl --location --request GET 'http://localhost:8080/blocks/16413972'
```

### List Latest n Blocks
[Get] /blocks?limit=n
```
ex:
curl --location --request GET 'http://localhost:8080/blocks?limit=2'
```

### DB

![alt text](https://github.com/ryanCool/ethService/blob/master/docs/blocks_db.png)

![alt text](https://github.com/ryanCool/ethService/blob/master/docs/transaction_db.png)

### Workflow
![alt text](https://github.com/ryanCool/ethService/blob/master/docs/workflow.png)