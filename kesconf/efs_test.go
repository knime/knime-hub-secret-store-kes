// Copyright 2024 - MinIO, Inc. All rights reserved.
// Use of this source code is governed by the AGPLv3
// license that can be found in the LICENSE file.

package kesconf_test

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/minio/kes/kesconf"
)

var EncryptedFSPath = flag.String("efs.path", "", "Path used for EncryptedFS tests")

func TestEncryptedFS(t *testing.T) {
	if *EncryptedFSPath == "" {
		t.Skip("EncryptedFS tests disabled. Use -efs.path=<path> to enable them")
	}

	masterKey := "passwordpasswordpasswordpassword"
	masterKeyPath := filepath.Join(*EncryptedFSPath, "test-master-key")
	masterKeyCipher := "AES256"
	if err := os.WriteFile(masterKeyPath, []byte(masterKey), 0o644); err != nil {
		t.Fatalf("Failed to write master key into test dir")
	}

	config := kesconf.EncryptedFSKeyStore{
		MasterKeyPath:   masterKeyPath,
		MasterKeyCipher: masterKeyCipher,
		Path:            *EncryptedFSPath,
	}

	ctx, cancel := testingContext(t)
	defer cancel()

	store, err := config.Connect(ctx)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Create", func(t *testing.T) { testCreate(ctx, store, t, RandString(ranStringLength)) })
	t.Run("Get", func(t *testing.T) { testGet(ctx, store, t, RandString(ranStringLength)) })
	t.Run("Status", func(t *testing.T) { testStatus(ctx, store, t) })
}
