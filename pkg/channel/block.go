package channel

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

type downloadBlockFn func(ctx context.Context, fileCid string, filePath string) error

type BlockChannel struct {
	ctx   context.Context
	ps    *pubsub.PubSub
	topic *pubsub.Topic
	sub   *pubsub.Subscription

	selfID        peer.ID
	downloadBlock downloadBlockFn
}

type Blocks struct {
	Items map[string]string `json:"items"`
}

func NewBlockChannel(ctx context.Context, ps *pubsub.PubSub, selfID peer.ID, downloadBlock downloadBlockFn) (*BlockChannel, error) {
	topic, err := ps.Join("blocks")
	if err != nil {
		return nil, err
	}

	sub, err := topic.Subscribe()
	if err != nil {
		return nil, err
	}

	bc := &BlockChannel{
		ctx:           ctx,
		topic:         topic,
		sub:           sub,
		ps:            ps,
		selfID:        selfID,
		downloadBlock: downloadBlock,
	}

	go bc.ReadBlocks()

	return bc, nil

}

func (bc *BlockChannel) BroadcastBlocks(blocks map[string]string) error {
	data, err := json.Marshal(Blocks{Items: blocks})
	if err != nil {
		panic(err)
	}

	return bc.topic.Publish(bc.ctx, data)
}

func (bc *BlockChannel) ReadBlocks() {
	fmt.Println("Waiting for blocks...")

	go func() {
		for {
			time.Sleep(time.Second * 5)
			peerList := bc.ps.ListPeers("blocks")

			fmt.Println("Found", len(peerList), "peers")

			for _, peer := range peerList {
				fmt.Println(peer.Pretty())
			}
		}

	}()
	for {
		msg, err := bc.sub.Next(bc.ctx)
		if err != nil {
			return
		}

		if msg.ReceivedFrom == bc.selfID {
			fmt.Println("Self message", msg.ReceivedFrom.Pretty())
			continue
		}

		fmt.Println(msg.ReceivedFrom)

		var blocksMsg Blocks
		err = json.Unmarshal(msg.Data, &blocksMsg)
		if err != nil {
			continue
		}

		for cid, filepath := range blocksMsg.Items {
			bc.downloadBlock(bc.ctx, cid, filepath)
			fmt.Println(cid, filepath)
		}
	}
}
