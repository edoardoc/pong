## Cloud Television Data Layer for jam.gg

MongoDb is used as a database, in its own docker container

## Startup
The project docker definition is in `docker-compose.yml`

### first step
```
cd broadcaster/
docker-compose up mongostorage
```
this will create the database, together with the `jamRS` that is needed for the streaming capability

`mngdata/ ` is the mongo db folder 


tested with:
```
Docker version 20.10.12, build e91ed57
Docker Compose version v2.2.3
```

### known problems
> - one 
> - two 
