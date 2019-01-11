# smor-serve

Server for social media app, with merkle tree post list data structure. 

Posts are stored chronologically in a list implemented as a merkle tree. 

This is an experiment following the hypothesis that social media data should be mutable, and the way that decentralized social networks handle data generally focus too much on permanence. Putting social media data on a PoW blockchain is computational overkill. You do not need global consensus and strict ordering, which a blockchain provides at great expense, for posts intended for only a limited circle of acquaintances that will not be relevant for long. Putting social media data in an append-only log, which does not have global consensus but is still ordered and immutable, is better, but is still too permanent. The pros and cons of our tree implementation is described below, but overall it is better suited for representing mutable social data. 

Smor = social media object representation

## Building

```go
go get .
```


## API Routes

#### GET `/feed/:user`
#### POST `/feed/:user`

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

#### GET `/user/new`
#### POST `/user/:username`

```
{
  "username": "arcalinea",
  "created_at": 1531813366,
  "pubkey": ""
}
```

#### GET `/post/:user/:timestamp`
#### DELETE `/post/:user/:timestamp`


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

Posts are sorted chronologically by timestamp in a merkle tree customized for storing and transferring data in a decentralized network. 

Intermediate nodes store links to other nodes, which may contain other intermediate nodes or posts. 

When a node containing posts reaches the limit of number of posts it can store, it splits its posts into two child nodes and instead stores the hashes of the child nodes. 

### Pros and Cons

Storing data in a tree, rather than an append-only data structure, offers the following advantages:

- It is easier to edit and delete data in a tree than in an append-only log. 

- Looking up old posts is faster, _O(log(N))_.

- You can do arbitrary insertions for posts with any timestamp, which is useful for importing your historical social data from other networks. 

- Range queries, to retrieve posts from x to y time, are easy to do. 

- Users can provide the signed root hash to others to indicate if there have been state changes and validate current state. (public/private keys are not yet implemented, but the idea is to verify that posts belong in a state tree with a signed merkle proof)

Possible disadvantages: 

- Allowing edits and arbitrary insertions also means you could easily pretend you posted something earlier than you did, or edit an earlier post to say something else, and it will look valid unless someone has a signed copy of a contradicting post you made. If you store your data with a third-party service or create your posts through an app, they could prevent you from fictionalizing history, like Facebook or Twitter does. But in a purely decentralized implementation with no append-only log, there is no objective ordering of events, so your history is under your control.
