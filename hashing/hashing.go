package hashing

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"hash/crc32"
	"reflect"
)

/**
* Has hashing functions that produce byte array hashes.
* @author rnojiri
**/

// GenerateByteArray - generates a new byte array based on the given parameters
func GenerateByteArray(parameters interface{}) ([]byte, error) {

	result, err := getByteArray(reflect.ValueOf(parameters))
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GenerateSHA256 - generates a sha256 hash based on the specified parameters
func GenerateSHA256(parameters interface{}) ([]byte, error) {

	byteArray, err := GenerateByteArray(parameters)
	if err != nil {
		return nil, err
	}

	h := sha256.New()

	_, err = h.Write(byteArray)
	if err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

// GenerateCRC32 - generates a sha256 hash based on the specified parameters
func GenerateCRC32(parameters interface{}) ([]byte, error) {

	byteArray, err := GenerateByteArray(parameters)
	if err != nil {
		return nil, err
	}

	h := crc32.NewIEEE()

	_, err = h.Write(byteArray)
	if err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

// GenerateMD5 - generates a md5 hash based on the specified parameters
func GenerateMD5(parameters interface{}) ([]byte, error) {

	byteArray, err := GenerateByteArray(parameters)
	if err != nil {
		return nil, err
	}

	h := md5.New()

	_, err = h.Write(byteArray)
	if err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

// GenerateSHA1 - generates a sha1 hash based on the specified parameters
func GenerateSHA1(parameters interface{}) ([]byte, error) {

	byteArray, err := GenerateByteArray(parameters)
	if err != nil {
		return nil, err
	}

	h := sha1.New()

	_, err = h.Write(byteArray)
	if err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}
