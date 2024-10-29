// Copyright 2024 - MinIO, Inc. All rights reserved.
// Use of this source code is governed by the AGPLv3
// license that can be found in the LICENSE file.ackage fs

package efs

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// test all operations
func TestCreateListGetKEK(t *testing.T) {
	tmp := t.TempDir()
	ctx := context.Background()

	// create master key
	masterKey := "passwordpasswordpasswordpassword"
	masterKeyCipher := "AES256"
	masterKeyPath := filepath.Join(tmp, "master-key")
	if err := os.WriteFile(masterKeyPath, []byte(masterKey), 0o644); err != nil {
		t.Fatalf("Failed to write master key into test temp dir")
	}

	// init keystore
	keystore, err := NewStore(masterKeyPath, masterKeyCipher, tmp)
	if err != nil {
		t.Fatalf("Failed to init keystore: %v", err)
	}

	// get name
	name := keystore.String()
	if !strings.Contains(name, "Encrypted Filesystem") {
		t.Fatalf("Unexpected keystore name: %s", name)
	}

	// get status
	_, err = keystore.Status(ctx)
	if err != nil {
		t.Fatalf("Failed to get keystore status: %v", err)
	}

	// empty kek list
	kekList, _, err := keystore.List(ctx, "test", 10)
	if err != nil {
		t.Fatalf("Failed to list keystore: %v", err)
	}
	if len(kekList) != 0 {
		t.Fatalf("Unexpected kek list entries, expected empty list: %d entries", len(kekList))
	}

	// create kek
	kekName := "test-kek"
	kekPlaintext := "my-plaintext-kek"
	err = keystore.Create(ctx, kekName, []byte(kekPlaintext))
	if err != nil {
		t.Fatalf("Unable to create kek: %v", err)
	}

	// list new kek
	kekList, _, err = keystore.List(ctx, "test", 10)
	if err != nil {
		t.Fatalf("Failed to list keystore: %v", err)
	}
	if len(kekList) != 1 {
		t.Fatalf("Unexpected kek list entries, expected list with one entry: %d entries", len(kekList))
	}
	if kekList[0] != kekName {
		t.Fatalf("Unexpected kek list entry: %s", kekList[0])
	}

	// read kek
	decryptetdKek, err := keystore.Get(ctx, kekName)
	if err != nil {
		t.Fatalf("Failed to read kek: %v", err)
	}
	if !bytes.Equal(decryptetdKek, []byte(kekPlaintext)) {
		t.Fatalf("Failed to decrypt kek: %s vs. %s", string(decryptetdKek), kekPlaintext)
	}

	// delete kek
	err = keystore.Delete(ctx, kekName)
	if err != nil {
		t.Fatalf("Failed to delete kek: %v", err)
	}

	// empty kek list
	kekList, _, err = keystore.List(ctx, "test", 10)
	if err != nil {
		t.Fatalf("Failed to list keystore: %v", err)
	}
	if len(kekList) != 0 {
		t.Fatalf("Unexpected kek list entries, expected empty list: %d entries", len(kekList))
	}

	// close keystore
	if err := keystore.Close(); err != nil {
		t.Fatalf("Failed to close keystore: %v", err)
	}
}

// ensure backward compatibility: read a known encrypted kek
func TestGetEnryptedKEK(t *testing.T) {
	tmp := t.TempDir()
	ctx := context.Background()

	// create master key
	masterKey := "passwordpasswordpasswordpassword"
	masterKeyCipher := "AES256"
	masterKeyPath := filepath.Join(tmp, "master-key")
	if err := os.WriteFile(masterKeyPath, []byte(masterKey), 0o644); err != nil {
		t.Fatalf("Failed to write master key into test temp dir")
	}

	// init keystore
	keystore, err := NewStore(masterKeyPath, masterKeyCipher, tmp)
	if err != nil {
		t.Fatalf("Failed to init keystore: %v", err)
	}

	// empty kek list
	kekList, _, err := keystore.List(ctx, "test", 10)
	if err != nil {
		t.Fatalf("Failed to list keystore: %v", err)
	}
	if len(kekList) != 0 {
		t.Fatalf("Unexpected kek list entries, expected empty list: %d entries", len(kekList))
	}

	// write encrypted kek to disk
	kekName := "test-kek"
	kekPlaintext := "my-plaintext-kek"
	encrypetdKek, err := base64.StdEncoding.DecodeString("Eu4t1j1T8CuLjgxqZoCBXguh6DJ+Jg4oZyhPUE6CNsgeGGZ3UhxQ0Eozh1A0THfsx/EK9rc97V2RTg5U")
	if err != nil {
		t.Fatalf("Failed to decode encrypted kek base64: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, kekName), encrypetdKek, 0o644); err != nil {
		t.Fatalf("Failed to write encrypted key into test temp dir")
	}

	// read kek
	decryptetdKek, err := keystore.Get(ctx, kekName)
	if err != nil {
		t.Fatalf("Failed to read kek: %v", err)
	}
	if !bytes.Equal(decryptetdKek, []byte(kekPlaintext)) {
		t.Fatalf("Failed to decrypt kek: %s vs. %s", string(decryptetdKek), kekPlaintext)
	}
}

// basic test to ensure kek was not written in plaintext to disk
func TestEncryptsFile(t *testing.T) {
	tmp := t.TempDir()
	ctx := context.Background()

	// create master key
	masterKey := "passwordpasswordpasswordpassword"
	masterKeyCipher := "AES256"
	masterKeyPath := filepath.Join(tmp, "master-key")
	if err := os.WriteFile(masterKeyPath, []byte(masterKey), 0o644); err != nil {
		t.Fatalf("Failed to write master key into test temp dir")
	}

	// init keystore
	keystore, err := NewStore(masterKeyPath, masterKeyCipher, tmp)
	if err != nil {
		t.Fatalf("Failed to init keystore: %v", err)
	}

	// create kek
	kekName := "test-kek"
	kekPlaintext := "my-plaintext-kek"
	err = keystore.Create(ctx, kekName, []byte(kekPlaintext))
	if err != nil {
		t.Fatalf("Unable to create kek: %v", err)
	}

	// ensure file on disk does not contain plaintext kek
	fileContent, err := os.ReadFile(filepath.Join(tmp, kekName))
	if err != nil {
		t.Fatalf("Failed to read kek file: %v", err)
	}
	if bytes.Equal(fileContent, []byte(kekPlaintext)) {
		t.Fatalf("Content of kek file not encrypted")
	}
}

// test key context gets validated
func TestKEKEncryptionContext(t *testing.T) {
	tmp := t.TempDir()
	ctx := context.Background()

	// create master key
	masterKey := "passwordpasswordpasswordpassword"
	masterKeyCipher := "AES256"
	masterKeyPath := filepath.Join(tmp, "master-key")
	if err := os.WriteFile(masterKeyPath, []byte(masterKey), 0o644); err != nil {
		t.Fatalf("Failed to write master key into test temp dir")
	}

	// init keystore
	keystore, err := NewStore(masterKeyPath, masterKeyCipher, tmp)
	if err != nil {
		t.Fatalf("Failed to init keystore: %v", err)
	}

	// create kek
	kekName := "test-kek"
	kekPlaintext := "my-plaintext-kek"
	err = keystore.Create(ctx, kekName, []byte(kekPlaintext))
	if err != nil {
		t.Fatalf("Unable to create kek: %v", err)
	}

	// copy kek
	otherKekName := "other-kek"
	fileContent, err := os.ReadFile(filepath.Join(tmp, kekName))
	if err != nil {
		t.Fatalf("Failed to read encrypted kek file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, otherKekName), fileContent, 0o644); err != nil {
		t.Fatalf("Failed to write encrypted kek into new file: %v", err)
	}

	// read other kek
	_, err = keystore.Get(ctx, otherKekName)
	if err == nil || !strings.Contains(fmt.Sprint(err), "ciphertext is not authentic") {
		t.Fatalf("Expected get to fail with ciphertext is not authentic")
	}
}

// test keystore init fails if master key is missing
func TestMissingMasterKey(t *testing.T) {
	tmp := t.TempDir()

	// missing master key
	masterKeyCipher := "AES256"
	masterKeyPath := filepath.Join(tmp, "master-key")

	// init keystore fails
	_, err := NewStore(masterKeyPath, masterKeyCipher, tmp)
	if err == nil {
		t.Fatalf("Expected init to fail on missing master key")
	}
}

// test keystore init fails if master key has unknown length
func TestInvalidMasterKeyLengthToShort(t *testing.T) {
	tmp := t.TempDir()

	// create master key with invalid length
	masterKey := "veryshortkey"
	masterKeyCipher := "AES256"
	masterKeyPath := filepath.Join(tmp, "master-key")
	if err := os.WriteFile(masterKeyPath, []byte(masterKey), 0o644); err != nil {
		t.Fatalf("Failed to write master key into test temp dir")
	}

	// init keystore
	_, err := NewStore(masterKeyPath, masterKeyCipher, tmp)
	if err == nil {
		t.Fatalf("Expected init to fail on invalid master key length")
	}
}

// test keystore init fails if master key has unknown length
func TestInvalidMasterKeyLengthToLarge(t *testing.T) {
	tmp := t.TempDir()

	// create master key with invalid length
	masterKey := "verylongverylongverylongverylongverylongverylongverylongverylong"
	masterKeyCipher := "AES256"
	masterKeyPath := filepath.Join(tmp, "master-key")
	if err := os.WriteFile(masterKeyPath, []byte(masterKey), 0o644); err != nil {
		t.Fatalf("Failed to write master key into test temp dir")
	}

	// init keystore
	_, err := NewStore(masterKeyPath, masterKeyCipher, tmp)
	if err == nil {
		t.Fatalf("Expected init to fail on invalid master key length")
	}
}

// test keystore init fails on unknown cipher
func TestUnknownMasterKeyCipher(t *testing.T) {
	tmp := t.TempDir()

	// create master key with unknown cipher
	masterKey := "passwordpasswordpasswordpassword"
	masterKeyCipher := "UNKNOWN"
	masterKeyPath := filepath.Join(tmp, "master-key")
	if err := os.WriteFile(masterKeyPath, []byte(masterKey), 0o644); err != nil {
		t.Fatalf("Failed to write master key into test temp dir")
	}

	// init keystore fails
	_, err := NewStore(masterKeyPath, masterKeyCipher, tmp)
	if err == nil {
		t.Fatalf("Expected init to fail on unknown master key cipher")
	}
}
