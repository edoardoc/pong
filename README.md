# pong
a video transmission framework with websockets

# first time run only, db creation                                
``` 
> docker-compose run api app createDb                               
																	
Creating broadcaster_api_run ... done                             
2022/10/18 08:45:01 creating database...                          
2022/10/18 08:45:01 db is alive                                   
selectedChannelResult:  &{ObjectID("634e678db7aa91c6715224b3")}   
```	
# from broadcaster folder                                         
## START API                                                       
```
docker-compose up -d mongostorage
```

## START API                                                       
```
docker-compose up -d api
```

## START antenna                                                   
```
docker-compose up -d antenna                                      
```																	


## START TV                                                        
```
docker build tv -t tv
docker run --name tv -t -d -p 80:80 tv
```
http://46.101.142.193/

advance one channel
curl http://46.101.142.193:8090/channel/next

previous channel
curl http://46.101.142.193:8090/channel/previous

Docker version 20.10.7, build f0df350
docker-compose version 1.27.4, build 40524192

## THIS IS A WORK IN PROGRESS
