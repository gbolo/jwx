package examples_test

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"log"

	"github.com/cloudflare/circl/sign/ed25519"
	"github.com/lestrrat-go/jwx/internal/base64"
	"github.com/lestrrat-go/jwx/internal/json"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jws"
	"github.com/pkg/errors"
)

type CirclEdDSASignerVerifier struct{}

func NewCirclEdDSASigner() (jws.Signer, error) {
	return &CirclEdDSASignerVerifier{}, nil
}

func NewCirclEdDSAVerifier() (jws.Verifier, error) {
	return &CirclEdDSASignerVerifier{}, nil
}

func (s CirclEdDSASignerVerifier) Algorithm() jwa.SignatureAlgorithm {
	return jwa.EdDSA
}

func (s CirclEdDSASignerVerifier) Sign(payload []byte, keyif interface{}) ([]byte, error) {
	switch key := keyif.(type) {
	case ed25519.PrivateKey:
		return ed25519.Sign(key, payload), nil
	default:
		return nil, errors.Errorf(`invalid key type %T`, keyif)
	}
}

func (s CirclEdDSASignerVerifier) Verify(payload []byte, signature []byte, keyif interface{}) error {
	switch key := keyif.(type) {
	case ed25519.PublicKey:
		if ed25519.Verify(key, payload, signature) {
			return nil
		}
		return errors.New(`failed to verify EdDSA signature`)
	default:
		return errors.Errorf(`invalid key type %T`, keyif)
	}
}

func ExampleJWS_ExternalSignerVerifier() {
	// This example shows how to register external jws.Signer / jws.Verifier for
	// a given algorithm.
	jws.RegisterSigner(jwa.EdDSA, jws.SignerFactoryFn(NewCirclEdDSASigner))
	jws.RegisterVerifier(jwa.EdDSA, jws.VerifierFactoryFn(NewCirclEdDSAVerifier))

	pubkey, privkey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		fmt.Printf(`failed to generate keys: %s`, err)
		return
	}

	const payload = "Lorem Ipsum"
	signed, err := jws.Sign([]byte(payload), jwa.EdDSA, privkey)
	if err != nil {
		fmt.Printf(`failed to generate signed message: %s`, err)
		return
	}

	verified, err := jws.Verify(signed, jwa.EdDSA, pubkey)
	if err != nil {
		fmt.Printf(`failed to verify signed message: %s`, err)
		return
	}

	if string(verified) != payload {
		fmt.Printf(`got invalid payload: %s`, verified)
		return
	}

	// OUTPUT:
}

func ExampleJWS_Sign() {
	privkey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Printf("failed to create private key: %s", err)
		return
	}

	buf, err := jws.Sign([]byte("Lorem ipsum"), jwa.RS256, privkey)
	if err != nil {
		log.Printf("failed to sign payload: %s", err)
		return
	}

	verified, err := jws.Verify(buf, jwa.RS256, &privkey.PublicKey)
	if err != nil {
		log.Printf("failed to verify JWS message: %s", err)
		return
	}

	// Do something with `verified` ....
	_ = verified

	// OUTPUT:
}

func ExampleJWS_Message() {
	const payload = `eyJpc3MiOiJqb2UiLA0KICJleHAiOjEzMDA4MTkzODAsDQogImh0dHA6Ly9leGFtcGxlLmNvbS9pc19yb290Ijp0cnVlfQ`
	const encodedSig1 = `cC4hiUPoj9Eetdgtv3hF80EGrhuB__dzERat0XF9g2VtQgr9PJbu3XOiZj5RZmh7AAuHIm4Bh-0Qc_lF5YKt_O8W2Fp5jujGbds9uJdbF9CUAr7t1dnZcAcQjbKBYNX4BAynRFdiuB--f_nZLgrnbyTyWzO75vRK5h6xBArLIARNPvkSjtQBMHlb1L07Qe7K0GarZRmB_eSN9383LcOLn6_dO--xi12jzDwusC-eOkHWEsqtFZESc6BfI7noOPqvhJ1phCnvWh6IeYI2w9QOYEUipUTI8np6LbgGY9Fs98rqVt5AXLIhWkWywlVmtVrBp0igcN_IoypGlUPQGe77Rw`
	const encodedSig2 = "DtEhU3ljbEg8L38VWAfUAqOyKAM6-Xx-F4GawxaepmXFCgfTjDxw5djxLa8ISlSApmWQxfKTUJqPP3-Kg6NU1Q"

	decodedPayload, err := base64.DecodeString(payload)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	decodedSig1, err := base64.DecodeString(encodedSig1)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	decodedSig2, err := base64.DecodeString(encodedSig2)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	public1 := jws.NewHeaders()
	_ = public1.Set(jws.AlgorithmKey, jwa.RS256)
	protected1 := jws.NewHeaders()
	_ = protected1.Set(jws.KeyIDKey, "2010-12-29")

	public2 := jws.NewHeaders()
	_ = public2.Set(jws.AlgorithmKey, jwa.ES256)
	protected2 := jws.NewHeaders()
	_ = protected2.Set(jws.KeyIDKey, "e9bc097a-ce51-4036-9562-d2ade882db0d")

	// Construct a message. DO NOT use values that are base64 encoded
	m := jws.NewMessage().
		SetPayload(decodedPayload).
		AppendSignature(
			jws.NewSignature().
				SetSignature(decodedSig1).
				SetProtectedHeaders(public1).
				SetPublicHeaders(protected1),
		).
		AppendSignature(
			jws.NewSignature().
				SetSignature(decodedSig2).
				SetProtectedHeaders(public2).
				SetPublicHeaders(protected2),
		)

	buf, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	fmt.Printf("%s", buf)
	// OUTPUT:
	// {
	//   "payload": "eyJpc3MiOiJqb2UiLA0KICJleHAiOjEzMDA4MTkzODAsDQogImh0dHA6Ly9leGFtcGxlLmNvbS9pc19yb290Ijp0cnVlfQ",
	//   "signatures": [
	//     {
	//       "header": {
	//         "kid": "2010-12-29"
	//       },
	//       "protected": "eyJhbGciOiJSUzI1NiJ9",
	//       "signature": "cC4hiUPoj9Eetdgtv3hF80EGrhuB__dzERat0XF9g2VtQgr9PJbu3XOiZj5RZmh7AAuHIm4Bh-0Qc_lF5YKt_O8W2Fp5jujGbds9uJdbF9CUAr7t1dnZcAcQjbKBYNX4BAynRFdiuB--f_nZLgrnbyTyWzO75vRK5h6xBArLIARNPvkSjtQBMHlb1L07Qe7K0GarZRmB_eSN9383LcOLn6_dO--xi12jzDwusC-eOkHWEsqtFZESc6BfI7noOPqvhJ1phCnvWh6IeYI2w9QOYEUipUTI8np6LbgGY9Fs98rqVt5AXLIhWkWywlVmtVrBp0igcN_IoypGlUPQGe77Rw"
	//     },
	//     {
	//       "header": {
	//         "kid": "e9bc097a-ce51-4036-9562-d2ade882db0d"
	//       },
	//       "protected": "eyJhbGciOiJFUzI1NiJ9",
	//       "signature": "DtEhU3ljbEg8L38VWAfUAqOyKAM6-Xx-F4GawxaepmXFCgfTjDxw5djxLa8ISlSApmWQxfKTUJqPP3-Kg6NU1Q"
	//     }
	//   ]
	// }
}
