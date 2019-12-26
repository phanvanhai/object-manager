# object-manager

## Running in docker

```
docker pull phanvanhai/docker-object-manager:1.1.0

docker run -it -d -p 5000:5000 --name object-manager phanvanhai/docker-object-manager:1.1.0
```


## Install and Deploy

To fetch the code and compile the web-based UI:

Using Git:
```
cd $GOPATH/src
git clone http://github.com/phanvanhai/object-manager.git github.com/edgexfoundry/object-manager
cd $GOPATH/src/github.com/edgexfoundry/object-manager
make build
```

To start the application :

```
make run
```

To rebuild after making changes to source:

```
make clean
make build
```
To test the web-based UI:

```
make test
```