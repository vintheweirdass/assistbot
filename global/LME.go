package global

// bsco database. not affiliated in any way to Lenovo

type LMEBscoResult struct {
	Name    string `json:"name"`
	Text    bool   `json:"text"`
	Graphic bool   `json:"graphic"`
	Version int    `json:"version"`
	Outside bool   `json:"boolean"`
	Id      int    `json:"id"`
}
type LMESpecResult struct {
	Id         string `json:"Id"`
	Guid       string `json:"Guid"`
	Brand      string `json:"Brand"`
	Name       string `json:"Name"`
	Image      string `json:"Image"`
	Serial     string `json:"Serial,omitempty"`
	Type       string `json:"Type"`
	ParentID   string `json:"ParentID"`
	Popularity string `json:"Popularity"`
	FullGuid   string `json:"FullGuid"`
}

var BscoDB []*LMEBscoResult = nil
