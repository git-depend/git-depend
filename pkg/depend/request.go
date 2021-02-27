package depend

import "encoding/json"

// Repository holds all the information about the merge which is taking place.
// Dependencies will only contain the URL and Branch.
type Request struct {
	Url    string     `json:"Url"`
	Branch string     `json:"Branch"`
	Author string     `json:"Author,omitempty"`
	Email  string     `json:"Email,omitempty"`
	Date   string     `json:"Date,omitempty"`
	Deps   *[]Request `json:"Deps,omitempty"`
}

// NewRequest creates the Request struct from the given fields
func NewRequest(url string, branch string, author string, email string, date string, deps *[]Request) *Request {
	return &Request{
		url,
		branch,
		author,
		email,
		date,
		deps,
	}
}

// GetJson from struct.
// Tab intends for prettier formating.
func (req *Request) GetJson() ([]byte, error) {
	return json.MarshalIndent(req, "", "\t")
}

// UpdateFromJson will unmarshall into this struct.
func (req *Request) UpdateFromJson(data []byte) error {
	return json.Unmarshal(data, req)
}
