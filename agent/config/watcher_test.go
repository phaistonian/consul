package config

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/consul/sdk/testutil"
	"github.com/stretchr/testify/require"
)

func TestNewWatcher(t *testing.T) {
	w, err := New(func(event *WatcherEvent) error {
		return nil
	})
	defer func() {
		_ = w.Close()
	}()
	require.NoError(t, err)
	require.NotNil(t, w)
}

func TestWatcherAddRemoveExist(t *testing.T) {
	watcherCh := make(chan *WatcherEvent)
	w, err := New(func(event *WatcherEvent) error {
		watcherCh <- event
		return nil
	})
	require.NoError(t, err)
	defer func() {
		_ = w.Close()
	}()

	filepath := createTempConfigFile(t, "temp_config1")
	filepath2 := createTempConfigFile(t, "temp_config2")
	filepath3 := createTempConfigFile(t, "temp_config3")

	err = w.Add(filepath)
	require.NoError(t, err)

	h, ok := w.configFiles[filepath]
	require.True(t, ok)
	require.NotEqual(t, 0, h.iNode)

	require.NoError(t, err)
	w.Start()
	err = os.Rename(filepath2, filepath)
	require.NoError(t, err)
	require.NoError(t, assertEvent(filepath, watcherCh))
	// make sure we consume all events
	assertEvent(filepath, watcherCh)

	// wait for file to be added back
	time.Sleep(w.reconcileTimeout + 50*time.Millisecond)
	err = w.Remove(filepath)
	require.NoError(t, err)
	time.Sleep(w.reconcileTimeout + 50*time.Millisecond)
	_, ok = w.configFiles[filepath]
	require.False(t, ok)

	err = os.Rename(filepath3, filepath)
	require.NoError(t, err)
	require.Error(t, assertEvent(filepath, watcherCh), "timedout waiting for event")
}

func TestWatcherAddNotExist(t *testing.T) {
	w, err := New(func(event *WatcherEvent) error {
		return nil
	})
	defer func() {
		_ = w.Close()
	}()
	require.NoError(t, err)
	file := testutil.TempFile(t, "temp_config")
	filename := file.Name() + randomString(16)
	err = w.Add(filename)
	require.True(t, os.IsNotExist(err))
	_, ok := w.configFiles[filename]
	require.False(t, ok)
}

func TestEventWatcherWrite(t *testing.T) {
	watcherCh := make(chan *WatcherEvent)
	w, err := New(func(event *WatcherEvent) error {
		watcherCh <- event
		return nil
	})
	defer func() {
		_ = w.Close()
	}()
	require.NoError(t, err)
	file := testutil.TempFile(t, "temp_config")
	_, err = file.WriteString("test config")
	require.NoError(t, err)
	err = file.Sync()
	require.NoError(t, err)

	err = w.Add(file.Name())
	require.NoError(t, err)
	w.Start()
	_, err = file.WriteString("test config 2")
	require.NoError(t, err)
	err = file.Sync()
	require.NoError(t, err)
	require.Error(t, assertEvent(file.Name(), watcherCh), "timedout waiting for event")
}

func TestEventWatcherRead(t *testing.T) {
	watcherCh := make(chan *WatcherEvent)
	w, err := New(func(event *WatcherEvent) error {
		watcherCh <- event
		return nil
	})
	require.NoError(t, err)
	defer func() {
		_ = w.Close()
	}()

	filepath := createTempConfigFile(t, "temp_config1")

	err = w.Add(filepath)
	require.NoError(t, err)
	w.Start()
	_, err = os.ReadFile(filepath)
	require.NoError(t, err)
	require.Error(t, assertEvent(filepath, watcherCh), "timedout waiting for event")
}

func TestEventWatcherChmod(t *testing.T) {
	watcherCh := make(chan *WatcherEvent)
	w, err := New(func(event *WatcherEvent) error {
		watcherCh <- event
		return nil
	})
	defer func() {
		_ = w.Close()
	}()
	require.NoError(t, err)
	file := testutil.TempFile(t, "temp_config")
	require.NoError(t, err)
	defer func() {
		err := file.Close()
		require.NoError(t, err)
	}()
	_, err = file.WriteString("test config")
	require.NoError(t, err)
	err = file.Sync()
	require.NoError(t, err)

	err = w.Add(file.Name())
	require.NoError(t, err)
	w.Start()
	file.Chmod(0777)
	require.NoError(t, err)
	require.Error(t, assertEvent(file.Name(), watcherCh), "timedout waiting for event")
}

func TestEventWatcherRemoveCreate(t *testing.T) {
	watcherCh := make(chan *WatcherEvent)
	w, err := New(func(event *WatcherEvent) error {
		watcherCh <- event
		return nil
	})
	defer func() {
		_ = w.Close()
	}()
	require.NoError(t, err)
	filepath := createTempConfigFile(t, "temp_config1")

	err = w.Add(filepath)
	require.NoError(t, err)
	w.reconcileTimeout = 20 * time.Millisecond
	w.Start()
	err = os.Remove(filepath)
	require.NoError(t, err)
	time.Sleep(w.reconcileTimeout + 50*time.Millisecond)
	recreated, err := os.Create(filepath)
	require.NoError(t, err)
	_, err = recreated.WriteString("config 2")
	require.NoError(t, err)
	err = recreated.Sync()
	require.NoError(t, err)
	time.Sleep(w.reconcileTimeout + 50*time.Millisecond)
	// this an event coming from the reconcile loop
	require.NoError(t, assertEvent(filepath, watcherCh))
	iNode, err := w.getFileId(recreated.Name())
	require.NoError(t, err)
	require.Equal(t, iNode, w.configFiles[recreated.Name()].iNode)
}

func TestEventWatcherMove(t *testing.T) {
	watcherCh := make(chan *WatcherEvent)
	w, err := New(func(event *WatcherEvent) error {
		watcherCh <- event
		return nil
	})
	defer func() {
		_ = w.Close()
	}()
	w.reconcileTimeout = 20 * time.Millisecond
	require.NoError(t, err)
	filepath := createTempConfigFile(t, "temp_config1")
	err = w.Add(filepath)
	require.NoError(t, err)
	w.Start()
	for i := 0; i < 100; i++ {
		filepath2 := createTempConfigFile(t, "temp_config2")
		err = os.Rename(filepath2, filepath)
		require.NoError(t, err)
		require.NoError(t, assertEvent(filepath, watcherCh))

		// wait for the reconcile to update the file with right inode
		//in case the Add was to quick and the new file was not yet created.
		time.Sleep(w.reconcileTimeout + 10*time.Millisecond)
		iNode, err := w.getFileId(filepath)
		require.NoError(t, err)
		require.Equal(t, iNode, w.configFiles[filepath].iNode)

	}
}

func TestEventReconcileMove(t *testing.T) {
	watcherCh := make(chan *WatcherEvent)
	w, err := New(func(event *WatcherEvent) error {
		watcherCh <- event
		return nil
	})
	defer func() {
		_ = w.Close()
	}()
	require.NoError(t, err)
	filepath := createTempConfigFile(t, "temp_config1")

	filepath2 := createTempConfigFile(t, "temp_config2")

	err = w.Add(filepath)
	require.NoError(t, err)
	w.reconcileTimeout = 20 * time.Millisecond
	// remove the file from the internal watcher to only trigger the reconcile
	err = w.watcher.Remove(filepath)
	require.NoError(t, err)
	w.Start()
	err = os.Rename(filepath2, filepath)
	require.NoError(t, err)
	require.NoError(t, assertEvent(filepath, watcherCh))
	iNode, err := w.getFileId(filepath)
	require.NoError(t, err)
	require.Equal(t, iNode, w.configFiles[filepath].iNode)
}

func assertEvent(name string, watcherCh chan *WatcherEvent) error {
	timeout := time.After(1000 * time.Millisecond)
	select {
	case ev := <-watcherCh:
		if ev.Filename != name {
			return fmt.Errorf("filename do not match")
		}
		return nil
	case <-timeout:
		return fmt.Errorf("timedout waiting for event")
	}
}

func createTempConfigFile(t *testing.T, filename string) string {
	file := testutil.TempFile(t, filename)
	defer func() {
		err := file.Close()
		require.NoError(t, err)
	}()
	_, err := file.WriteString("test config")
	require.NoError(t, err)
	err = file.Sync()
	require.NoError(t, err)
	return file.Name()
}