package main

import (
	"fmt"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-ipfs-blockstore"
)

// Iterates through nodes to retrieve a post

func getByTimestampClosure(out **Smor, timestamp uint64) func(*Smor) {
	return func(s *Smor) {
		if s.CreatedAt == timestamp {
			*out = s
		}
	}
}

func isInRange(timestamp uint64, beg uint64, end uint64) bool {
	return timestamp >= beg && timestamp <= end // extracting to avoid off-by-one errors
}

func searchPostsByTimestamp(bs blockstore.Blockstore, posts []*cid.Cid, timestamp uint64) (*Smor, error) {
	for i := range posts {
		sm, err := getPost(bs, posts[i])
		if err != nil {
			return nil, err
		}
		if sm.CreatedAt == timestamp {
			return sm, nil
		}
	}
	return nil, fmt.Errorf("Smor not found in posts")
}

func searchChildrenByTimestamp(bs blockstore.Blockstore, children []*childLink, timestamp uint64) (*Smor, error) {
	for i := range children {
		if isInRange(timestamp, children[i].Beg, children[i].End) {
			node, err := getNode(bs, children[i].Node)
			if err != nil {
				return nil, err
			}
			sm, err := node.searchNodeByTimestamp(bs, timestamp)
			if err != nil {
				return nil, err
			}
			return sm, nil
		}
	}
	return nil, fmt.Errorf("Smor not found in children")
}

func (mln *MerkleListNode) searchNodeByTimestamp(bs blockstore.Blockstore, timestamp uint64) (*Smor, error) {
	if len(mln.Posts) > 0 {
		return searchPostsByTimestamp(bs, mln.Posts, timestamp)
	} else if len(mln.Children) > 0 {
		return searchChildrenByTimestamp(bs, mln.Children, timestamp)
	} else {
		panic("merkle node has no posts or children")
	}
}

func (ml *MerkleList) RetrievePost(timestamp uint64) (*Smor, error) {
	if len(ml.root.Posts) != 0 {
		// has to be in posts on root node
		return searchPostsByTimestamp(ml.bs, ml.root.Posts, timestamp)
	} else if len(ml.root.Children) > 0 {
		children := ml.root.Children
		if isInRange(timestamp, children[0].Beg, children[len(children) - 1].End) {
			// iterate through child nodes of root node
			return searchChildrenByTimestamp(ml.bs, children, timestamp)
		} else {
			// It's beyond range of root's children, so search from last child
			node, err := getNode(ml.bs, children[len(children) - 1].Node)
			if err != nil {
				return nil, err
			}
			sm, err := node.searchNodeByTimestamp(ml.bs, timestamp)
			if err != nil {
				return nil, err
			}
			return sm, nil
		}
	}
	panic("Shouldn't get here in RetrievePost")
}