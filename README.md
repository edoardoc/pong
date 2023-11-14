# pong
a video transmission framework with websockets: 
- a broadcaster which gets a transmission data from a mongo db
- an antenna which tunes in from the web to the server via websockets and keeps showing the thing coming from the antenna
- the channel information is constantly watched by the broadcaster, so that anytime I want to change channel I just need to write something else in a tabl.. er collection


# first time run only, db creation                                
``` 
> docker-compose run api app createDb                               
																	
Creating broadcaster_api_run ... done                             
2022/10/18 08:45:01 creating database...                          
2022/10/18 08:45:01 db is alive                                   
selectedChannelResult:  &{ObjectID("634e678db7aa91c6715224b3")}   
```	
# from the broadcaster folder                                         
## START MONGO                                                       
```
docker-compose up -d mongostorage
```

## START API                                                       
```
docker-compose up -d api
```

## START the antenna                                                   
```
docker-compose up -d antenna                                      
```																	


## START the TV                                                        
```
docker build tv -t tv
docker run --name tv -t -d -p 80:80 tv
```
This will show the transmission
http://ip.address/

advance one channel
curl http://ip.address:8090/channel/next

previous channel
curl http://ip.address:8090/channel/previous

Docker version 20.10.7, build f0df350
docker-compose version 1.27.4, build 40524192

## THIS IS A WORK IN PROGRESS

## TODO
some buffering is needed
