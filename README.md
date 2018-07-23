# smor-serve

## Building

```go
go get .
```

## Usage

Post some data:
```
curl -X POST -d @data.json http://localhost:7777/feed/username
```

Get feed data:

```
curl http://localhost:7777/feed/username
```
