## Code Challenge: Authorizer

### Requirements
* Docker

### Architecture
* Why Go?
  * General purpose language
  * Good learning curve and strongly typed
  * Easy build into binary for many arq & OS 
  * Testing is easy enough
  * Many built-in package available
  * Fast for cli purpose
  * Great concurrence support
* Why not?
  * Lack of generics (lots of workarounds)
  * Not native FP language
  * Hard to manage non-structured json
  * Not built-in Map functions implementation (Map, Filter, Reduce, ...)

### Makefile usage
```shell
# Default actions are build & test
docker run --rm -v "$PWD":/usr/src/myapp -w /usr/src/myapp golang:1.17 make

# build
docker run --rm -v "$PWD":/usr/src/myapp -w /usr/src/myapp golang:1.17 make build

# compile
docker run --rm -v "$PWD":/usr/src/myapp -w /usr/src/myapp golang:1.17 make compile

# test
docker run --rm -v "$PWD":/usr/src/myapp -w /usr/src/myapp golang:1.17 make test

# run
docker run --rm -v "$PWD":/usr/src/myapp -w /usr/src/myapp golang:1.17 make run
```

###Examples
Given a file called operations that contains several lines describing operations in json format:
```shell
cat operations
{"account": {"active-card": true, "available-limit": 100}}
{"transaction": {"merchant": "Burger King", "amount": 20, "time": "2019-02-13T10:00:00.000Z"}}
{"transaction": {"merchant": "Habbib's", "amount": 90, "time": "2019-02-13T11:00:00.000Z"}}
{"transaction": {"merchant": "McDonald's", "amount": 30, "time": "2019-02-13T12:00:00.000Z"}}
```

The application should be able to receive the file content through stdin , and for each processed operation return an output according to the business rules:
```shell
authorize < operations
{"account": {"active-card": true, "available-limit": 100}, "violations": []}
{"account": {"active-card": true, "available-limit": 80}, "violations": []}
{"account": {"active-card": true, "available-limit": 80}, "violations": ["insufficientlimit"]}
{"account": {"active-card": true, "available-limit": 50}, "violations": []}
```

### TODO
* Add linter
* Add more examples

### Info
* How to convert JSON string: https://www.sohamkamani.com/golang/json/ 

### More...
* This is my first Golang approach, don't be so hard with me