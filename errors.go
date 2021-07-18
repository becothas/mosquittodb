package mosquittodb

import "errors"

var (
	ErrBadMagic                         = errors.New("bad magic")
	ErrBadChunkID                       = errors.New("wrong chunk id")
	ErrUnexpectedConfigurationChunkSize = errors.New("unexpected ConfigurationChunk size")
)
