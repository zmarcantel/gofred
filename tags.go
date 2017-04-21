package gofred

type Tag struct {
	Name        string   `json:"name" xml:"name"`
	GroupId     TagId    `json:"group_id" xml:"group_id"`
	Notes       string   `json:"notes" xml:"notes"`
	Created     DateTime `json:"created" xml:"created"`
	Popularity  uint     `json:"popularity" xml:"popularity"`
	SeriesCount uint     `json:"series_count" xml:"series_count"`
}
