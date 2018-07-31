package main

import (
	"fmt"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-ipfs-blockstore"
	cbor "github.com/ipfs/go-ipld-cbor"
)

const postsPerNode = 4

type MerkleList struct {
	bs blockstore.Blockstore
	root *MerkleListNode
}

type MerkleListNode struct {
	Posts    []*cid.Cid
	Children []*childLink
	Depth    int
}

type childLink struct {
	// Beg is the lowest timestamp on any post in this child
	Beg uint64
	// End is the highest timestamp on any post in this child
	End uint64
	// Node is the hash link to the child object
	Node *cid.Cid
}

func LoadMerkleList(bs blockstore.Blockstore, c *cid.Cid) (*MerkleList, error) {
	if c == nil {
		return &MerkleList{
			bs: bs,
			root: nil,
		}, nil
	}
	node, err := getNode(bs, c)
	if err != nil {
		return nil, err
	}
	return &MerkleList{
		bs: bs,
		root: node,
	}, nil
}

func NewPostsNode(posts []*cid.Cid) *MerkleListNode {
	return &MerkleListNode{
		Depth: 0,
		Posts: posts,
	}
}

func NewChildrenNode(children []*childLink, depth int) *MerkleListNode {
	return &MerkleListNode{
		Depth: depth,
		Children: children,
	}
}

func (mln *MerkleListNode) getChildLink(bs blockstore.Blockstore) (*childLink, error) {
	cid, err := putNode(bs, mln)
	if err != nil {
		return nil, err
	}
	if len(mln.Posts) > 0 {
		begNode, err := getPost(bs, mln.Posts[0])
		if err != nil {
			return nil, err
		}
		endNode, err := getPost(bs, mln.Posts[len(mln.Posts) - 1])
		if err != nil {
			return nil, err
		}
		child := &childLink {
			Beg: begNode.CreatedAt,
			End: endNode.CreatedAt,
			Node: cid,
		}
		return child, nil
	} else if len(mln.Children) > 0 {
		child := &childLink {
			Beg: mln.Children[0].Beg,
			End: mln.Children[len(mln.Children) - 1].End,
			Node: cid,
		}
		return child, nil
	} else {
		panic("no posts or children on node, why")
	}
}

// InsertPost inserts the given Smor in order into the merklelist
func (ml *MerkleList) InsertPost(p *Smor) error {
	c, err := putPost(ml.bs, p)
	if err != nil {
		return err
	}

	if ml.root == nil {
		// base case of an empty tree, just make a new node with the thing in it
		ml.root = &MerkleListNode{
			Posts: []*cid.Cid{c},
		}

		return nil
	}

	// pass it off to the recursive function (also pass our 'blockstore' so it can persist state)
	extra, err := ml.root.insertPost(ml.bs, p.CreatedAt, c)
	if err != nil {
		return err
	}

	if extra != nil {
		ml.splitNode(extra)
	}

	return nil
}

func (ml *MerkleList) splitNode(mln *MerkleListNode) error {
		rootCl, err := ml.root.getChildLink(ml.bs)
		if err != nil {
			return err
		}
		
		mlnCl, err := mln.getChildLink(ml.bs)
		if err != nil {
			return err
		}
		
		children := []*childLink{
			rootCl,
			mlnCl,
		}
		
		ml.root = NewChildrenNode(children, ml.root.Depth + 1)
		
		return nil
}

// Inserting a post into the leaf node it belongs in, base case of insertPost()
func (mln *MerkleListNode) insertIntoLeaf(bs blockstore.Blockstore, time uint64, c *cid.Cid) (*MerkleListNode, error) {
			// iterate from the end to the front, we expect most 'inserts' to be 'append'
			var i int
			for i = len(mln.Posts) - 1; i >= 0; i-- {
				sm, err := mln.getPostByIndex(bs, i)
				if err != nil {
					return nil, err
				}

				if time >= sm.CreatedAt {
					// insert it here!

					// snippet below from golang slice tricks
					mln.Posts = append(mln.Posts[:i+1], append([]*cid.Cid{c}, mln.Posts[i+1:]...)...)
					break
				}
			}

			if i == -1 {
				// if we make it here, our post occurs before every other post, insert it to the front
				mln.Posts = append([]*cid.Cid{c}, mln.Posts...)
			}

			// now check if we need to split
			if len(mln.Posts) > postsPerNode {
				fmt.Println("Splitting node...")
				/* split this node into two
				Go from:
				  [ ............... ]
				To:
				  [ X X ]
				    | |--------|
				    |          |
				  [ .......]  [ .........]
				*/

				extra := mln.Posts[postsPerNode:]
				mln.Posts = mln.Posts[:postsPerNode]

				if len(extra) > 1 {
					panic("don't handle this case yet")
				}

				return NewPostsNode(extra), nil
			}

			return nil, nil
}

func (mln *MerkleListNode) insertPost(bs blockstore.Blockstore, time uint64, c *cid.Cid) (*MerkleListNode, error) {
	// Base case, no child nodes, insert in this node
	if mln.Depth == 0 {
		return mln.insertIntoLeaf(bs, time, c)
	}

	// recursive case, find the child it belongs in
	for i := len(mln.Children) - 1; i >= 0; i-- {
		if time >= mln.Children[i].Beg || i == 0 {
			var extra *MerkleListNode
			err := mln.mutateChild(bs, i, func(cmlnn *MerkleListNode) error {
				// inserting
				ex, err := cmlnn.insertPost(bs, time, c)
				if err != nil {
					return err
				}

				extra = ex
				return nil
			})
			if err != nil {
				return nil, err
			}

			if extra != nil {
				// if we're at end of the array, append
				cl, err := extra.getChildLink(bs)
				if err != nil {
					return nil, err
				}
				if i == len(mln.Children) - 1 {
					mln.Children = append(mln.Children, cl)
				} else {
					mln.Children = append(mln.Children[:i+1], append([]*childLink{cl}, mln.Children[i+1:]...)...)
				}
				if len(mln.Children) > postsPerNode {
					// Splitting child node
					extra := mln.Children[postsPerNode:]
					mln.Children = mln.Children[:postsPerNode]

					if len(extra) > 1 {
						panic("don't handle this case yet")
					}

					return NewChildrenNode(extra, mln.Depth), nil
				}
			}
			return nil, nil
		}
	}
	panic("shouldnt ever get here...")
}

// mutateChild loads the given child from its hash, applys the given function
// to it, then rehashes it and updates the link in the children array
func (mln *MerkleListNode) mutateChild(bs blockstore.Blockstore, i int, mutateFunc func(*MerkleListNode) error) error {
	ch := mln.Children[i]
	blk, err := bs.Get(ch.Node)
	if err != nil {
		return err
	}

	var childNode MerkleListNode
	if err := cbor.DecodeInto(blk.RawData(), &childNode); err != nil {
		return err
	}

	if err := mutateFunc(&childNode); err != nil {
		return err
	}

	chl, err := childNode.getChildLink(bs)
	if err != nil {
		return err
	}

	mln.Children[i] = chl
	return nil
}
