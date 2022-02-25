package wallets

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"

	"github.com/mr-tron/base58"
	"golang.org/x/crypto/ripemd160"
)

const (
	ChecksumLength = 4
	Version        = byte(0x00)
)

type Wallet struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey  []byte
}

func NewWallet() (*Wallet, error) {
	curve := elliptic.P256()

	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, err
	}

	pub := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)

	return &Wallet{
		PrivateKey: private,
		PublicKey:  pub,
	}, nil
}

func (w *Wallet) Address() []byte {
	pubHash := w.PublicKeyHash()

	versionedHash := append([]byte{Version}, pubHash...)
	checksum := Checksum(versionedHash)

	fullHash := append(versionedHash, checksum...)

	return []byte(base58.Encode(fullHash))
}

func (w *Wallet) PublicKeyHash() []byte {
	pubHash := sha256.Sum256(w.PublicKey)

	hasher := ripemd160.New()
	_, err := hasher.Write(pubHash[:])
	if err != nil {
		panic(err)
	}

	return hasher.Sum(nil)
}

func Checksum(payload []byte) []byte {
	firstHash := sha256.Sum256(payload)
	secondHash := sha256.Sum256(firstHash[:])

	return secondHash[:ChecksumLength]
}
