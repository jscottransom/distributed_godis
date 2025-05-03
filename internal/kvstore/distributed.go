package kvstore

import (
	"bytes"
	"crypto/tls"
	"encoding/gob"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb/v2"
	mdb "github.com/hashicorp/raft-mdb"
	"github.com/jscottransom/distributed_godis/api"
	kmap "github.com/jscottransom/distributed_godis/internal/keymap"
	"github.com/jscottransom/distributed_godis/internal/kvstore/store"
	"google.golang.org/protobuf/proto"
)

type DistributedKVStore struct {
	config   Config
	kvstore  *KVstore
	raft	 *raft.Raft
}

func NewDistributedKVStore(dataDir, storeName string) (*DistributedKVStore, error,) {

	kv := &DistributedKVStore{
		config: config,
	}

	if err := kv.setupKV(dataDir, storeName); err != nil {
		return nil, err
	}

	if err := kv.setupRaft(dataDir); err != nil {
		return nil, err
	}

	return kv, nil
}

func (kv *DistributedKVStore) setupKV(dataDir, storeName string) error {
	var err error
	kv.kvstore, err = NewKVstore(dataDir,storeName)

	return err
}

func (kv *DistributedKVStore) setupRaft(dataDir string) error {
	fsm := &fsm{kv: kv.kvstore}

	raftLogDir := filepath.Join(dataDir, "raftLog")
	if err := os.MkdirAll(raftLogDir, 0755); err != nil {
		return err
	}

	logStore, err := mdb.NewMDBStore(raftLogDir)
	if err != nil {
		return err
	}

	stableStore, err := raftboltdb.NewBoltStore(
		filepath.Join(dataDir, "raft", "stable"),
	)
	if err != nil {
		return err
	}
	
	retain := 1
	snapshotStore, err := raft.NewFileSnapshotStore(
		filepath.Join(dataDir, "raft"), 
		retain,
		os.Stderr,
	)

	if err != nil {
		return err 
	}

	maxPool := 5
	timeout := 10 * time.Second
	transport := raft.NewNetworkTransport(
			kv.config.Raft.Streamlayer, 
			maxPool, 
			timeout, 
			os.Stderr,
	)

	config := raft.DefaultConfig()
	config.LocalID = kv.config.Raft.LocalID
	if kv.config.Raft.HearbeatTimeout != 0 {
		config.HeartbeatTimeout = kv.config.Raft.HeartbeatTimeout
	}
	if kv.config.Raft.ElectionTimeout != 0 {
		config.ElectionTimeout = kv.config.Raft.ElectionTimeout
	}
	if kv.config.Raft.LeaderLeaseTimeout != 0 {
		config.LeaderLeaseTimeout = kv.config.Raft.LeaderLeaseTimeout
	}
	if kv.config.Raft.CommitTimeout != 0 {
		config.CommitTimeout = kv.config.Raft.CommitTimeout
	}

	kv.raft, err = raft.NewRaft(
					config,
					fsm,
					logStore,
					stableStore,
					snapshotStore,
					transport,
	)
	if err != nil {
		return err
	}

	hasState, err := raft.HasExistingState(
			logStore,
			stableStore,
			snapshotStore,
	)
	if err != nil {
		return err
	}
	if kv.config.Raft.Bootstrap && !hasState {
		config := raft.Configuration{
			Servers: []raft.Server{{
				ID: config.LocalID,
				Address: transport.LocalAddr(),
			}},
		}
		err = kv.raft.BootstrapCluster(config).Error()
	}
	return err

}

func (kv *DistributedKVStore) SetKey(record *store.Record) (string, error) {

	res, err := kv.apply(SetRequestType,
						&api.SetRequest{Key: record.Key,
										Value: record.Value},)
	
	if err != nil {
		return 
	}

	return res.(*api.SetResponse).Response, nil

}

func (kv *DistributedKVStore) apply(reqType RequestType, req proto.Message) (
		  interface{},
		  error,
	) {

		var buf bytes.Buffer
		_, err := buf.Write([]byte{byte(reqType)})
		if err != nil {
			return nil, err
		}
		b, err := proto.Marshal(req)
		if err != nil {
			return nil, err 
		}
		_, err = buf.Write(b)
		if err != nil {
			return nil, err
		}
		timeout := 10 * time.Second
		future := kv.raft.Apply(buf.Bytes(), timeout)
		if future.Error() != nil {
			return nil, future.Error()
		}

		res := future.Response()
		if err, ok := res.(error); ok {
			return nil, err
		}

		return res, nil 

}

func (kv *DistributedKVStore) GetKey(key string) ([]byte, error) {
	return kv.kvstore.Get(key)
}

var _ raft.FSM = (*fsm)(nil)
type fsm struct {
	kvstore *KVstore
}

type RequestType uint8
const (
	SetRequestType RequestType = 0
)

func (kv *fsm) Apply(record *raft.Log) interface{} {
	buf := record.data
	reqType := RequestType(buf[0])
	switch reqType {
	case SetRequestType:
		return kv.applySetKey(buf[1:])
	}
	return nil
}

func (kv *fsm) applySetKey(b []byte) interface{} {
	var req api.SetRequest
	err := proto.Unmarshal(b, &req)
	if err != nil {
		return err
	}

	// Conver the setRequest to a reecord type

	record := &store.Record{
		Key: req.Key,
		Value: req.Value,
	}

	err := kv.kvstore.Set(record)
	if err != nil {
		return err
	}

	return &api.SetResponse{Response: "OK"}

}


type KVPair struct {
	Key string
	Value []byte
}

func (f *fsm) Snapshot() (raft.FSMSnapshot, error) {
	f.kvstore.Mu.Lock()
	defer f.kvstore.Mu.Unlock()

	// Get the list of keys for the store
	keylist := []string{}

	// Iterate through the list of keys
	for k := range f.kvstore.Keymap.Map {
		keylist = append(keylist, k)
	}

	
	kvPairs := []KVPair{}
	
	for key := range keylist {
		val, err := s.Config.Store.Get(key)
		if err != nil {
			fmt.Printf("Unable to get key: %s", req.Key)
			return nil, err
		}

		kvPair := &KVPair{
			Key: key,
			Value: val,
		}

		kvPairs = append(kvPairs, kvPair)
		
	}

	// Need to serialize state objects since they must implement io.Reader

	return &snapshot{
		keymap: f.kvstore.Keymap.Map,
		keylist: keylist,
		kvPairs: kvPairs,
	}

}

var _ raft.FSMSnapshot = (*snapshot)(nil)


type snapshot struct {
	keymap *kmap.SafeMap
	keylist []string
	kvPairs []KVPair
}


func (s *snapshot) Persist(sink raft.SnapshotSink) error {
	
	// Copy the multiple elements to the chosen sink
	
	// TODO: gob encoding is faster, get this working with json marshaling first
	mapData, err := json.Marshal(s.keymap)
	if err != nil {
		_ = sink.Cancel()
		return err
	}

	mapDataReader := bytes.NewReader(mapData)
	if _, err := io.Copy(sink, mapDataReader); err != nil {
		_ = sink.Cancel()
		return err
	}


	keyData, err := json.Marshal(s.keylist)
	if err != nil {
		_ = sink.Cancel()
		return err
	}

	keyDataReader := bytes.NewReader(keyData)
	if _, err := io.Copy(sink, keyDataReader); err != nil {
		_ = sink.Cancel()
		return err
	}

	kvData, err := json.Marshal(s.kvPairs)
	if err != nil {
		_ = sink.Cancel()
		return err
	}

	kvDataReader := bytes.NewReader(kvData)
	if _, err := io.Copy(sink, kvDataReader); err != nil {
		_ = sink.Cancel()
		return err
	}

}

func (s *snapshot) Release() {}

func (f *fsm) Restore(r io.ReadCloser) error {

	var buf bytes.Buffer 
	_, err := io.copy(&buf, r)
	if err != nil {
		log.Fatal(err)
	}

	


}
