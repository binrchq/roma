package sshd

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"io/ioutil"

	"binrc.com/roma/core/utils/logger"
	"golang.org/x/crypto/ssh"
)

// GenKey gen ssh private and public key
// func GenKey(keyFilePath string) (string, string, error) {
// 	savePrivateFileTo := utils.FilePath(keyFilePath)
// 	savePublicFileTo := fmt.Sprintf("%s.pub", savePrivateFileTo)
// 	bitSize := 4096

// 	privateKey, err := generatePrivateKey(bitSize)
// 	if err != nil {
// 		logger.Logger.Error(err.Error())
// 		return "", "", err
// 	}

// 	publicKeyBytes, err := generatePublicKey(&privateKey.PublicKey)
// 	if err != nil {
// 		logger.Logger.Error(err.Error())
// 		return "", "", err
// 	}

// 	privateKeyBytes := encodePrivateKeyToPEM(privateKey)

// 	err = writeKeyToFile(privateKeyBytes, savePrivateFileTo)
// 	if err != nil {
// 		logger.Logger.Error(err.Error())
// 		return "", "", err
// 	}

// 	err = writeKeyToFile([]byte(publicKeyBytes), savePublicFileTo)
// 	if err != nil {
// 		logger.Logger.Error(err.Error())
// 		return "", "", err
// 	}

// 	return savePrivateFileTo, savePublicFileTo, nil
// }

func GenKey() ([]byte, []byte, error) {
	bitSize := 4096

	privateKey, err := generatePrivateKey(bitSize)
	if err != nil {
		logger.Logger.Error(err.Error())
		return nil, nil, err
	}

	publicKeyBytes, err := generatePublicKey(&privateKey.PublicKey)
	if err != nil {
		logger.Logger.Error(err.Error())
		return nil, nil, err
	}

	privateKeyBytes := encodePrivateKeyToPEM(privateKey)

	// 将私钥和公钥转换为Base64编码的字符串，方便存储到数据库
	privateKeyBase64 := base64.StdEncoding.EncodeToString(privateKeyBytes)
	publicKeyBase64 := base64.StdEncoding.EncodeToString([]byte(publicKeyBytes))

	return []byte(privateKeyBase64), []byte(publicKeyBase64), nil
}

// generatePrivateKey creates a RSA Private Key of specified byte size
func generatePrivateKey(bitSize int) (*rsa.PrivateKey, error) {
	// Private Key generation
	privateKey, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return nil, err
	}

	// Validate Private Key
	err = privateKey.Validate()
	if err != nil {
		return nil, err
	}

	logger.Logger.Info("Private Key generated")
	return privateKey, nil
}

// encodePrivateKeyToPEM encodes Private Key from RSA to PEM format
func encodePrivateKeyToPEM(privateKey *rsa.PrivateKey) []byte {
	// Get ASN.1 DER format
	privDER := x509.MarshalPKCS1PrivateKey(privateKey)

	// pem.Block
	privBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privDER,
	}

	// Private key in PEM format
	privatePEM := pem.EncodeToMemory(&privBlock)

	return privatePEM
}

// generatePublicKey take a rsa.PublicKey and return bytes suitable for writing to .pub file
// returns in the format "ssh-rsa ..."
func generatePublicKey(privatekey *rsa.PublicKey) ([]byte, error) {
	publicRsaKey, err := ssh.NewPublicKey(privatekey)
	if err != nil {
		return nil, err
	}

	pubKeyBytes := ssh.MarshalAuthorizedKey(publicRsaKey)

	logger.Logger.Info("Public key generated")
	return pubKeyBytes, nil
}

// writePemToFile writes keys to a file
func writeKeyToFile(keyBytes []byte, saveFileTo string) error {
	err := ioutil.WriteFile(saveFileTo, keyBytes, 0600)
	if err != nil {
		return err
	}

	logger.Logger.Info("Key saved to: %s", saveFileTo)
	return nil
}
