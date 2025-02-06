# Calling Admin API GraphQL in Golang

## Step 1: Set up your environment

You need to have `go` in your `PATH`.
It is assumed that you know how to install Golang on your machine.

## Step 2: Git clone this repo

```
git clone https://github.com/louischan-oursky/authgear_example_admin_api_golang.git
```

## Step 3: Change directory to the just cloned repository

```
cd authgear_example_admin_api_golang
```

## Step 4: Make changes to main.go

In main.go, there are some places marked with `TODO: ` where you need to make changes.

## Step 5: Run it

```
make
```

If you run it in a project without any users, you should see this output

```
{
  "data": {
    "users": {
      "edges": []
    }
  }
}
{
  "data": {
    "createUser": {
      "user": {
        "id": "VXNlcjo4ZTNlNjc4ZC0yMGRkLTQ3YWUtOGM4MS0wNjk3N2JmMGIwOTA",
        "standardAttributes": {
          "email": "user@example.com",
          "email_verified": false,
          "updated_at": 1738839536
        }
      }
    }
  }
}
```

Running it again, you will see

```
{
  "data": {
    "users": {
      "edges": [
        {
          "node": {
            "id": "VXNlcjo4ZTNlNjc4ZC0yMGRkLTQ3YWUtOGM4MS0wNjk3N2JmMGIwOTA",
            "standardAttributes": {
              "email": "user@example.com",
              "email_verified": false,
              "updated_at": 1738839536
            }
          }
        }
      ]
    }
  }
}
panic: {"errors":[{"extensions":{"errorName":"Invalid","info":{"IdentityTypeExisting":"login_id","IdentityTypeIncoming":"login_id","LoginIDTypeExisting":"email","LoginIDTypeIncoming":"email","cause":{"kind":"DuplicatedIdentity"}},"reason":"InvariantViolated"},"locations":[{"column":4,"line":2}],"message":"identity already exists","path":["createUser"]}]}

goroutine 1 [running]:
main.main()
        /Users/yourname/authgear_example_admin_api_golang/main.go:204 +0x1f0
exit status 2
make: *** [Makefile:3: run] Error 1
```

Because the example always tries to create a user with `user@example.com`, and obviously, it cannot be done twice.
