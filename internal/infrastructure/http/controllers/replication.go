package controllers

type ReplicateInput struct {
	SourceStorage Storage `json:"source_storage"`
	TargetStorage Storage `json:"target_storage"`
}

type Storage struct {
	AccountID string `json:"account_id"`
	Bucket    string `json:"bucket"`
}
