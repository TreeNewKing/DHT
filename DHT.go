package main

import (
	"crypto/sha1"
	"fmt"
	"math/rand"
	"time"
)

const (
	BUCKET_SIZE = 3
	BUCKET_NUM  = 160
)

type Peer struct {
	ID     [20]byte
	Bucket [BUCKET_NUM][]*Peer
}

func NewPeer(id [20]byte) *Peer {
	p := &Peer{ID: id}
	for i := 0; i < BUCKET_NUM; i++ {
		p.Bucket[i] = make([]*Peer, 0, BUCKET_SIZE)
	}
	return p
}

func (p *Peer) FindNode(target [20]byte) [BUCKET_SIZE]*Peer {
	var ret [BUCKET_SIZE]*Peer
	var dist [BUCKET_NUM][20]byte
	for i := 0; i < BUCKET_NUM; i++ {
		for j := 0; j < len(p.Bucket[i]); j++ {
			dist[i] = xor(p.Bucket[i][j].ID, target)
		}
	}
	for i := 0; i < BUCKET_SIZE; i++ {
		minIndex := 0
		for j := 1; j < BUCKET_NUM; j++ {
			if byteCompare(dist[j][:], dist[minIndex][:]) < 0 {
				minIndex = j
			}
		}
		ret[i] = p.Bucket[minIndex][0]
		dist[minIndex] = [20]byte{255}
	}
	return ret
}

func (p *Peer) SetValue(key, value []byte) bool {
	hash := sha1.Sum(key)
	if byteCompare(hash[:], p.ID[:]) == 0 {
		return false
	}
	for i := 0; i < BUCKET_NUM; i++ {
		for j := 0; j < len(p.Bucket[i]); j++ {
			if byteCompare(p.Bucket[i][j].ID[:], p.ID[:]) == 0 {
				continue
			}
			if byteCompare(hash[:], p.Bucket[i][j].ID[:]) == 0 {
				return true
			}
		}
	}
	for i := 0; i < BUCKET_NUM; i++ {
		dist := xor(p.ID, hash)
		if dist[i] == 0 {
			p.Bucket[i] = append(p.Bucket[i], &Peer{ID: hash})
			return true
		}
	}
	closePeers := p.FindNode(hash)
	for i := 0; i < BUCKET_SIZE; i++ {
		if byteCompare(closePeers[i].ID[:], p.ID[:]) == 0 {
			continue
		}
		if closePeers[i].SetValue(key, value) {
			return true
		}
	}
	return false
}

func (p *Peer) GetValue(key [20]byte) []byte {
	for i := 0; i < BUCKET_NUM; i++ {
		for j := 0; j < len(p.Bucket[i]); j++ {
			if byteCompare(p.Bucket[i][j].ID[:], p.ID[:]) == 0 {
				continue
			}
			if byteCompare(key[:], p.Bucket[i][j].ID[:]) == 0 {
				return p.Bucket[i][j].GetValue(key)
			}
		}
	}
	closePeers := p.FindNode(key)
	for i := 0; i < BUCKET_SIZE; i++ {
		if byteCompare(closePeers[i].ID[:], p.ID[:]) == 0 {
			continue
		}
		value := closePeers[i].GetValue(key)
		if value != nil {
			sum := sha1.Sum(value)
			if byteCompare(sum[:], key[:]) == 0 {
				return value
			}
		}
	}
	return nil
}

func byteCompare(a []byte, b []byte) int {
	for i := 0; i < 20; i++ {
		if a[i] < b[i] {
			return -1
		} else if a[i] > b[i] {
			return 1
		}
	}
	return 0
}

func xor(a, b [20]byte) [20]byte {
	var ret [20]byte
	for i := 0; i < 20; i++ {
		ret[i] = a[i] ^ b[i]
	}
	return ret
}

func main() {
	rand.Seed(time.Now().UnixNano())

	peers := make([]*Peer, BUCKET_NUM)
	for i := 0; i < BUCKET_NUM; i++ {
		id := sha1.Sum([]byte(fmt.Sprintf("peer-%d", i)))
		peers[i] = NewPeer(id)
	}

	for i := 0; i < 200; i++ {
		key := sha1.Sum([]byte(fmt.Sprintf("key-%d", i)))
		value := make([]byte, rand.Intn(1024))
		rand.Read(value)
		peerIndex := rand.Intn(BUCKET_NUM)
		if peers[peerIndex].SetValue(key[:], value) {
			fmt.Printf("Peer %d set value for key %x\n", peerIndex, key)
		}
	}

	keys := make([][20]byte, 100)
	for i := 0; i < 100; i++ {
		keys[i] = sha1.Sum([]byte(fmt.Sprintf("key-%d", rand.Intn(200))))
	}

	for i := 0; i < 100; i++ {
		key := keys[i]
		peerIndex := rand.Intn(BUCKET_NUM)
		value := peers[peerIndex].GetValue(key)
		if value != nil {
			fmt.Printf("Peer %d get value for key %x: %x\n", peerIndex, key, value)
		}
	}
}
