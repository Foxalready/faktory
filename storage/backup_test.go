package storage

import (
	"os"
	"testing"

	"github.com/mperham/faktory/util"
	"github.com/stretchr/testify/assert"
)

func init() {
	DefaultPath = "/tmp"
	os.Mkdir("/tmp", os.FileMode(os.ModeDir|0755))
}

func TestBackupAndRestore(t *testing.T) {
	t.Parallel()

	defer os.RemoveAll("/tmp/backup.db")
	// open db
	db, err := Open("rocksdb", "backup.db")
	assert.NoError(t, err)

	// put elements
	q, err := db.GetQueue("default")
	assert.NoError(t, err)
	q.Push([]byte("f"))
	q.Push([]byte("fo"))
	assert.Equal(t, int64(2), q.Size())

	rs := db.Retries()
	rs.AddElement(util.Nows(), "foobar", []byte("thepayload"))
	assert.Equal(t, int64(1), rs.Size())

	count := 0
	db.EachBackup(func(element BackupInfo) {
		count += 1
	})
	assert.Equal(t, 0, count)

	// take backup
	err = db.Backup()
	assert.NoError(t, err)
	count = 0
	db.EachBackup(func(element BackupInfo) {
		count += 1
	})
	assert.Equal(t, 1, count)

	// put more elements
	q.Push([]byte("foo"))
	assert.Equal(t, int64(3), q.Size())

	// restore from backup
	err = db.RestoreFromLatest()
	assert.NoError(t, err)

	db, err = Open("rocksdb", "backup.db")
	assert.NoError(t, err)

	// verify elements
	q, err = db.GetQueue("default")
	assert.NoError(t, err)
	assert.Equal(t, int64(2), q.Size())

	elm, err := q.Pop()
	assert.NoError(t, err)
	assert.Equal(t, []byte("f"), elm)

	elm, err = q.Pop()
	assert.NoError(t, err)
	assert.Equal(t, []byte("fo"), elm)

	elm, err = q.Pop()
	assert.NoError(t, err)
	assert.Nil(t, elm)

	assert.Equal(t, int64(1), db.Retries().Size())
}