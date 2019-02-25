Iost explorer backend
======


### Requirements 

1. golang 1.10+
2. mongodb



### How to deploy

1. clone code to $GOPATH/src/github.com/iost-official/iost-api
2. add config file cp backend/config/config.json.sample backend/config/config.json
    update the ```rpcHost```, ```mongodb``` config if need
3. Run blockchain sync task: 
```bash
cd task
make
nohup ./iost-api-task &
```
4. Run REST api service
```bash
cd $GOPATH/src/github.com/iost-official/iost-api 
make 
nohup ./iost-api&
```


