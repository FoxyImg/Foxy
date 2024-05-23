package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"log"
)

func VerifySignature(key string, sig string, data string) bool {
	hmackDaddy := hmac.New(sha256.New, []byte(key))
	hmackDaddy.Write([]byte(data))
	hexSig := hex.EncodeToString(hmackDaddy.Sum(nil))

	log.Println("Signature:", sig, "Hex:", hexSig)

	return sig == hexSig
}
