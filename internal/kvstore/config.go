package kvstore

import (
	"github.com/hashicorp/raft"
)

type Config struct {
	Raft struct {
		raft.Config
		Streamlayer *Streamlayer
		Bootstrap 	bool
	}
}