// Copyright 2023 - MinIO, Inc. All rights reserved.
// Use of this source code is governed by the AGPLv3
// license that can be found in the LICENSE file.

package edge_test

import (
	"bytes"
	"context"
	"os"
	"os/signal"
	"testing"

	"github.com/minio/kes/kv"
)

type SetupFunc func(context.Context, kv.Store[string, []byte]) error

var createTests = []struct {
	Args       map[string][]byte
	Setup      SetupFunc
	ShouldFail bool
}{
	{ // 0
		Args: map[string][]byte{"my-key": []byte("my-value")},
	},
	{ // 1
		Args: map[string][]byte{"my-key": []byte("my-value")},
		Setup: func(ctx context.Context, s kv.Store[string, []byte]) error {
			return s.Create(ctx, "my-key", []byte(""))
		},
		ShouldFail: true,
	},
}

func testCreate(ctx context.Context, store kv.Store[string, []byte], t *testing.T) {
	defer clean(ctx, store, t)
	for i, test := range createTests {
		if test.Setup != nil {
			if err := test.Setup(ctx, store); err != nil {
				t.Fatalf("Test %d: failed to setup: %v", i, err)
			}
		}

		for key, value := range test.Args {
			err := store.Create(ctx, key, value)
			if err != nil && !test.ShouldFail {
				t.Errorf("Test %d: failed to create key '%s': %v", i, key, err)
			}
			if err == nil && test.ShouldFail {
				t.Errorf("Test %d: creating key '%s' should have failed: %v", i, key, err)
			}
		}
		clean(ctx, store, t)
	}
}

var setTests = []struct {
	Args       map[string][]byte
	Setup      SetupFunc
	ShouldFail bool
}{
	{ // 0
		Args: map[string][]byte{"my-key": []byte("my-value")},
	},
	{ // 1
		Args: map[string][]byte{"my-key": []byte("my-value")},
		Setup: func(ctx context.Context, s kv.Store[string, []byte]) error {
			return s.Create(ctx, "my-key", []byte(""))
		},
		ShouldFail: true,
	},
}

func testSet(ctx context.Context, store kv.Store[string, []byte], t *testing.T) {
	defer clean(ctx, store, t)
	for i, test := range setTests {
		if test.Setup != nil {
			if err := test.Setup(ctx, store); err != nil {
				t.Fatalf("Test %d: failed to setup: %v", i, err)
			}
		}

		for key, value := range test.Args {
			err := store.Create(ctx, key, value)
			if err != nil && !test.ShouldFail {
				t.Errorf("Test %d: failed to set key '%s': %v", i, key, err)
			}
			if err == nil && test.ShouldFail {
				t.Errorf("Test %d: setting key '%s' should have failed: %v", i, key, err)
			}
		}
		clean(ctx, store, t)
	}
}

var getTests = []struct {
	Args       map[string][]byte
	Setup      SetupFunc
	ShouldFail bool
}{
	{ // 0
		Args: map[string][]byte{"my-key": []byte("my-value")},
		Setup: func(ctx context.Context, s kv.Store[string, []byte]) error {
			return s.Create(ctx, "my-key", []byte("my-value"))
		},
	},
	{ // 1
		Args:       map[string][]byte{"my-key": []byte("my-value")},
		ShouldFail: true,
	},
	{ // 1
		Args: map[string][]byte{"my-key": []byte("my-value")},
		Setup: func(ctx context.Context, s kv.Store[string, []byte]) error {
			return s.Create(ctx, "my-key", []byte("my-value2"))
		},
		ShouldFail: true,
	},
}

func testGet(ctx context.Context, store kv.Store[string, []byte], t *testing.T) {
	defer clean(ctx, store, t)
	for i, test := range getTests {
		if test.Setup != nil {
			if err := test.Setup(ctx, store); err != nil {
				t.Fatalf("Test %d: failed to setup: %v", i, err)
			}
		}

		for key, value := range test.Args {
			v, err := store.Get(ctx, key)
			if !test.ShouldFail {
				if err != nil {
					t.Errorf("Test %d: failed to get key '%s': %v", i, key, err)
				}
				if !bytes.Equal(v, value) {
					t.Errorf("Test %d: failed to get key: got '%s' - want '%s'", i, string(v), string(value))
				}
			}
			if test.ShouldFail && err == nil && bytes.Equal(v, value) {
				t.Errorf("Test %d: getting key '%s' should have failed: %v", i, key, err)
			}
		}
		clean(ctx, store, t)
	}
}

func testStatus(ctx context.Context, store kv.Store[string, []byte], t *testing.T) {
	if _, err := store.Status(ctx); err != nil {
		t.Fatalf("Failed to fetch status: %v", err)
	}
}

var osCtx, _ = signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)

func testingContext(t *testing.T) (context.Context, context.CancelFunc) {
	d, ok := t.Deadline()
	if !ok {
		return osCtx, func() {}
	}
	return context.WithDeadline(osCtx, d)
}

func clean(ctx context.Context, store kv.Store[string, []byte], t *testing.T) {
	iter, err := store.List(ctx)
	if err != nil {
		t.Fatalf("Cleanup: failed to list keys: %v", err)
	}
	defer iter.Close()

	var names []string
	for next, ok := iter.Next(); ok; next, ok = iter.Next() {
		names = append(names, next)
	}
	for _, name := range names {
		if err = store.Delete(ctx, name); err != nil {
			t.Errorf("Cleanup: failed to delete '%s': %v", name, err)
		}
	}
	if err = iter.Close(); err != nil {
		t.Errorf("Cleanup: failed to close iter: %v", err)
	}
}
