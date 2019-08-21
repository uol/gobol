package hashing_test

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uol/gobol/hashing"
)

// TestSHA256 - tests the sha256 implementation
func TestSHA256(t *testing.T) {

	results, err := hashing.GenerateSHA256("hello world!")
	if !assert.NoError(t, err, "error generating sha256") {
		return
	}

	assert.Equal(t, "7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9", hex.EncodeToString(results), "unexpected sha256")
}

// TestSHA1 - tests the sha1 implementation
func TestSHA1(t *testing.T) {

	results, err := hashing.GenerateSHA1("hello world!")
	if !assert.NoError(t, err, "error generating sha1") {
		return
	}

	assert.Equal(t, "430ce34d020724ed75a196dfc2ad67c77772d169", hex.EncodeToString(results), "unexpected sha1")
}

// TestCRC32 - tests the crc32 implementation
func TestCRC32(t *testing.T) {

	results, err := hashing.GenerateCRC32("hello world!")
	if !assert.NoError(t, err, "error generating crc32") {
		return
	}

	assert.Equal(t, "03b4c26d", hex.EncodeToString(results), "unexpected crc32")
}

// TestMD5 - tests the md5 implementation
func TestMD5(t *testing.T) {

	results, err := hashing.GenerateMD5("hello world!")
	if !assert.NoError(t, err, "error generating md5") {
		return
	}

	assert.Equal(t, "fc3ff98e8c6a0d3087d515c0473f8677", hex.EncodeToString(results), "unexpected md5")
}
