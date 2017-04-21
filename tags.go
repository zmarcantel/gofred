package gofred

type Tag struct {
	Name        string   `json:"name" xml:"name,attr"`
	GroupId     TagId    `json:"group_id" xml:"group_id,attr"`
	Notes       string   `json:"notes" xml:"notes,attr"`
	Created     DateTime `json:"created" xml:"created,attr"`
	Popularity  uint     `json:"popularity" xml:"popularity,attr"`
	SeriesCount uint     `json:"series_count" xml:"series_count,attr"`
}
