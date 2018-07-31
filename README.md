# smor-serve

In progress

## Building

```go
go get .
```

## API Routes

get `/feed/:user`
post `/feed/:user`

```
[{
  "type": "tweet",
  "source": "twitter.com/tweetid",
  "author": "arcalinea",
  "created_at": 1532813359,
  "data": {"text": "A big statue in the desert" },
  "response_to": "",
  "signature": "" 
}]
```

get `/user/new`
post `/user/:username`

```
{
  "username": "arcalinea",
  "created_at": 1531813366,
  "pubkey": ""
}
```

get `/post/:user/:timestamp`
delete `/post/:user/:timestamp`

## Example Usage

Post feed data:
```
curl -X POST -d @data.json http://localhost:7777/feed/username
```

Get feed data:

```
curl http://localhost:7777/feed/username
```

Create a new user: 

```
curl -X POST -d @user.json http://localhost:7777/user/new
```

Get user: 

```
curl http://localhost:7777/user/username
```

## Data structure 

Posts are stored in a list implemented as a tree, referenced by hash. 

When a node containing posts reaches the limit of number of posts it can store, it splits its posts into two child nodes and instead stores the hashes of the child nodes. 

```
Go from:
  [ ............... ]
To:
  [ X X ]
    | |--------|
    |          |
  [ .......]  [ .........]
```