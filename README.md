* Given a file called operations that contains several lines describing operations in json format:
```shell
$ cat operations
[
  {
    "account": {
      "active-card": true,
      "available-limit": 100
    }
  },
  {
    "transaction": {
      "merchant": "Burger King",
      "amount": 20,
      "time": "2019-02-13T10:00:00.000Z"
    }
  },
  {
    "transaction": {
      "merchant": "Habbib's",
      "amount": 90,
      "time": "2019-02-13T11:00:00.000Z"
    }
  },
  {
    "transaction": {
      "merchant": "McDonald's",
      "amount": 30,
      "time": "2019-02-13T12:00:00.000Z"
    }
  }
]
```

* The application should be able to receive the file content through stdin , and for each processed operation return an output according to the business rules:
```shell
$ authorize < operations
[
  {
    "account": {
      "active-card": true,
      "available-limit": 100
    },
    "violations": []
  },
  {
    "account": {
      "active-card": true,
      "available-limit": 80
    },
    "violations": []
  },
  {
    "account": {
      "active-card": true,
      "available-limit": 80
    },
    "violations": [
      "insufficientlimit"
    ]
  },
  {
    "account": {
      "active-card": true,
      "available-limit": 50
    },
    "violations": []
  }
]
```