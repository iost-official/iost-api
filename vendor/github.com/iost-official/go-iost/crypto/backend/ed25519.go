package backend

import (
	"crypto/rand"

	"github.com/iost-official/go-iost/ilog"
	"golang.org/x/crypto/ed25519"
)

// Ed25519 is the ed25519 crypto algorithm
type Ed25519 struct{}

// Sign will signature the message with seckey by ed25519
func (b *Ed25519) Sign(message []byte, seckey []byte) []byte {
	return ed25519.Sign(seckey, message)
}

// Verify will verify the message with pubkey and sig by ed25519
func (b *Ed25519) Verify(message []byte, pubkey []byte, sig []byte) bool {
	return ed25519.Verify(pubkey, message, sig)
}

// GetPubkey will get the public key of the secret key by ed25519
func (b *Ed25519) GetPubkey(seckey []byte) []byte {
	pubkey, ok := ed25519.PrivateKey(seckey).Public().(ed25519.PublicKey)
	if !ok {
		ilog.Errorf("Failed to assert ed25519.PublicKey")
		return nil
	}
	return pubkey
}

// GenSeckey will generate the secret key by ed25519
func (b *Ed25519) GenSeckey() []byte {
	seed := make([]byte, 32)
	_, err := rand.Read(seed)
	if err != nil {
		ilog.Errorf("Failed to random seed, %v", err)
		return nil
	}
	return ed25519.NewKeyFromSeed(seed)
}
