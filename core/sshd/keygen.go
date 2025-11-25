package sshd

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"strings"

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

	// 私钥直接返回原始 PEM 格式，不需要 Base64 编码
	privateKeyStr := string(privateKeyBytes)

	// 公钥直接返回原始格式（ssh-rsa AAAAB3NzaC1yc2E...），不需要 Base64 编码
	// 去除末尾的换行符
	publicKeyStr := strings.TrimSpace(string(publicKeyBytes))

	return []byte(privateKeyStr), []byte(publicKeyStr), nil
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
// returns in the format "ssh-rsa ... roma-auto-gen@roma"
func generatePublicKey(privatekey *rsa.PublicKey) ([]byte, error) {
	publicRsaKey, err := ssh.NewPublicKey(privatekey)
	if err != nil {
		return nil, err
	}

	// 生成公钥，添加注释
	pubKeyBytes := ssh.MarshalAuthorizedKey(publicRsaKey)

	// 移除末尾的换行符，添加注释
	pubKeyStr := strings.TrimSpace(string(pubKeyBytes))
	if !strings.Contains(pubKeyStr, "@") {
		// 如果没有注释，添加默认注释
		pubKeyStr = pubKeyStr + " roma-auto-gen@roma"
	}

	logger.Logger.Info("Public key generated")
	return []byte(pubKeyStr + "\n"), nil
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
