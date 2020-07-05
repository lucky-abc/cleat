package record

import (
	"github.com/syndtr/goleveldb/leveldb"
	"os"
	"strconv"
)

type RecordPoint struct {
	key    string
	offset uint64
	db     *leveldb.DB
}

func NewCheckpoint(dbpath string) (*RecordPoint, error) {
	_, err := os.Stat(dbpath)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(dbpath, 0777)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	db, err := leveldb.OpenFile(dbpath, nil)
	if err != nil {
		return nil, err
	}
	ck := &RecordPoint{}
	ck.db = db
	return ck, nil
}

func (ck *RecordPoint) SetCheckpoint(key string, offset uint64) {
	ck.db.Put([]byte(key), []byte(strconv.FormatUint(offset, 10)), nil)
}

func (ck *RecordPoint) DelCheckpoint(key string) {
	ck.db.Delete([]byte(key), nil)
}

func (ck *RecordPoint) GetCheckpoint(key string) (uint64, error) {
	val, err := ck.db.Get([]byte(key), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return 0, nil
		}
		return 0, err
	}
	v, err := strconv.ParseUint(string(val), 10, 54)
	if err != nil {
		return 0, err
	}
	return v, nil
}

func (ck *RecordPoint) Close() {
	if ck.db != nil {
		ck.db.Close()
	}
}
