package model

type MultipartUpload struct {
	ID       string
	Bucket   string
	Partials []string
}

func NewMultiPartUpload(id string, bucket string) *MultipartUpload {
	return &MultipartUpload{
		ID:       id,
		Bucket:   bucket,
		Partials: make([]string, 0),
	}
}

func (u *MultipartUpload) AddPartial(tag string) {
	u.Partials = append(u.Partials, tag)
}
