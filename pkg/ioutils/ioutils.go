package ioutils

import "io"

type InterceptWriter struct {
    Writer io.Writer
    Hook   func(bytes []byte)
}	

func (iw *InterceptWriter) Write(bytes []byte) (n int, err error) {
    if iw.Hook != nil {
        iw.Hook(bytes)
    }
    return iw.Writer.Write(bytes)
}