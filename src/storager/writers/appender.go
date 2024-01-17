package writers

import (
	"bytes"
	"fmt"
	"go.beyondstorage.io/v5/types"
	"io/fs"
)

type appender struct {
	appender     types.Appender
	appendObject *types.Object
	path         string
}

func NewAppender(appenderStorage types.Appender, path string) Writer {
	return &appender{
		appender: appenderStorage,
		path:     path,
	}
}

func (f *appender) Close() error {
	if f.appendObject == nil {
		return nil
	}
	return f.appender.CommitAppend(f.appendObject)
}

func (f *appender) Write(p []byte) (int, error) {
	if f.appendObject == nil {
		obj, err := f.appender.CreateAppend(f.path)
		if err != nil {
			return 0, err
		}

		f.appendObject = obj
	}

	buf := bytes.NewBuffer(p)
	n := int64(len(p))
	n64, err := f.appender.WriteAppend(f.appendObject, buf, n)

	return int(n64), err
}

func (f *appender) Stat() (fs.FileInfo, error) {
	return nil, fmt.Errorf("appender writer does cannot return Stat()")
}
