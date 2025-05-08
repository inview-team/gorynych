package service

import "errors"

var (
	ErrNoAvailableBuckets = errors.New("no available buckets")
	ErrStorageExists      = errors.New("storage with this id already exists")
	ErrStorageNotFound    = errors.New("storage not found")
	ErrBucketNotFound     = errors.New("bucket not found")
	ErrUploadNotFound     = errors.New("upload not found")
	ErrUploadBig          = errors.New("upload is too big")
	ErrWrongOffset        = errors.New("wrong offset")
)

var (
	ErrObjectNotFound = errors.New("object not found")
)

// Service accounts errors
var (
	ErrNoAvailableAccounts = errors.New("no available accounts")
	ErrAccountNotFound     = errors.New("account not found")
)

var (
	ErrTaskNotFound = errors.New("task not found")
)
