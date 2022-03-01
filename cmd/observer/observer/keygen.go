package observer

import (
	"context"
	"crypto/ecdsa"
	"github.com/ledgerwatch/erigon/crypto"
	"github.com/ledgerwatch/erigon/p2p/enode"
	"github.com/ledgerwatch/log/v3"
	"runtime"
	"time"
)

func keygen(parentContext context.Context, targetKey *ecdsa.PublicKey, timeout time.Duration, logger log.Logger) []*ecdsa.PublicKey {
	ctx, cancel := context.WithTimeout(parentContext, timeout)
	defer cancel()

	targetID := enode.PubkeyToIDV4(targetKey)
	cpus := runtime.GOMAXPROCS(-1)

	type result struct {
		key      *ecdsa.PublicKey
		distance int
	}

	generatedKeys := make(chan result, cpus)

	for i := 0; i < cpus; i++ {
		go func() {
			for ctx.Err() == nil {
				keyPair, err := crypto.GenerateKey()
				if err != nil {
					logger.Error("keygen has failed to generate a key", "err", err)
					break
				}

				key := &keyPair.PublicKey
				id := enode.PubkeyToIDV4(key)
				distance := enode.LogDist(targetID, id)

				select {
				case generatedKeys <- result{key, distance}:
				case <-ctx.Done():
					break
				}
			}
		}()
	}

	keysAtDist := make(map[int]*ecdsa.PublicKey)

	for ctx.Err() == nil {
		select {
		case res := <-generatedKeys:
			keysAtDist[res.distance] = res.key
		case <-ctx.Done():
			break
		}
	}

	keys := make([]*ecdsa.PublicKey, 0, len(keysAtDist))
	for _, key := range keysAtDist {
		keys = append(keys, key)
	}

	return keys
}
