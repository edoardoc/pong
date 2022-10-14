## Cloud Television Data Layer for jam.gg

MongoDb is used as a database, in its own docker container

## Startup
The project docker definition is in `docker-compose.yml`
everything should start with a simple

`docker-compose up`

which will bring up a database container (that in turns  will create a local `mngdata/` folder for the storage)

tested with:
```
Docker version 20.10.12, build e91ed57
Docker Compose version v2.2.3
```

### known problems
> - one 
> - two 
