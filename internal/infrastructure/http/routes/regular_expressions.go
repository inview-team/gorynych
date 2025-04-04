package routes

import "fmt"

const (
	uploadID = "upload_id"
	objectID = "object_id"
)

const regexpUpload = "^[0-9A-F]{16}$"

var patternUploadID = fmt.Sprintf("{%s:%s}", uploadID, regexpUpload)
