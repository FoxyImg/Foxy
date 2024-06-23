package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net/url"
	"slices"
	"strings"
)

func VerifySignature(key string, sig string, data string) bool {
	hmackDaddy := hmac.New(sha256.New, []byte(key))
	hmackDaddy.Write([]byte(data))
	hexSig := hex.EncodeToString(hmackDaddy.Sum(nil))

	log.Println("Signature:", sig, "Hex:", hexSig)

	return sig == hexSig
}

func VerifySignatureFromQuery(secret string, sourceKey string, query url.Values) bool {
	sig := query.Get("s")

	decodedParams := []string{}

	for key, value := range query {
		if key == "s" || key == "_" || key == "showpreset" {
			continue
		}

		decodedParams = append(decodedParams, key+"="+strings.Join(value, ","))
	}

	slices.Sort(decodedParams)
	sortedQuery := sourceKey + "?" + strings.Join(decodedParams, "&")

	log.Println("sorted: " + sortedQuery)

	return VerifySignature(secret, sig, sortedQuery)
}
