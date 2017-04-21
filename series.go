package gofred

import (
	"fmt"
	"net/url"
	"time"
)

type Series struct {
	Id                 string             `json:"id" xml:"id,attr"`
	Start              Date               `json:"realtime_start" xml:"realtime_start,attr"`
	End                Date               `json:"realtime_end" xml:"realtime_end,attr"`
	Title              string             `json:"title" xml:"title,attr"`
	ObservationStart   Date               `json:"observation_start" xml:"observation_start,attr"`
	ObservationEnd     Date               `json:"observation_end" xml:"observation_end,attr"`
	Frequency          Frequency          `json:"frequency" xml:"frequency,attr"`
	Units              string             `json:"units" xml:"units,attr"`
	UnitsShort         string             `json:"units_short" xml:"units_short,attr"`
	SeasonallyAdjusted SeasonalAdjustment `json:"seasonal_adjustment" xml:"seasonal_adjustment,attr"`
	LastUpdate         DateTime           `json:"last_updated" xml:"last_updated,attr"`
	Popularity         uint16             `json:"popularity" xml:"popularity,attr"` // TODO: type check
	Notes              string             `json:"notes" xml:"notes,attr"`
}

type SeriesSearchType string

const (
	SearchTypeNone SeriesSearchType = "NONE"
	SearchFullText SeriesSearchType = "full_text"
	SearchSeriesId SeriesSearchType = "series_id"
)

//==============================================================================
//
// GET: /fred/series
//
//==============================================================================

// Holds the data needed to request information on a `Series`.
type SeriesRequest struct {
	baseRequest
	DatedRequest
	Series string
}

func NewSeriesRequest(series string) SeriesRequest {
	return SeriesRequest{
		Series: series,
	}
}

// Satisfies the `Request` interface.
func (r SeriesRequest) ToParams() url.Values {
	v := r.baseRequest.ToParams()
	r.DatedRequest.MergeParams(v)
	v.Set("series_id", r.Series)
	return v
}

// Response type which _should_ contain only one category.
type seriesResponse struct {
	Start  Date     `json:"realtime_start" xml:"realtime_start,attr"`
	End    Date     `json:"realtime_end" xml:"realtime_end,attr"`
	Series []Series `json:"seriess" xml:"seriess,attr"`
}

//
// Get the `Series` information for the given series ID.
//
// Asserts there is only one `Series` object in the result, and returns it.
//
func (c Client) Series(req SeriesRequest) (Series, Error) { // TODO: add a SeriesById(string) for simplicity
	req.baseRequest = c.base_req

	req_url := c.base_url
	req_url.RawQuery = req.ToParams().Encode()
	req_url.Path = fmt.Sprintf("%s/series", req_url.Path)

	body, err := c.get("series", req_url.String())
	if err != nil {
		return Series{}, err.Prefixf("error getting series %s: %v", req.Series, err)
	}

	// parse the correct format
	var result seriesResponse
	err = c.unmarshal_body(body, &result)
	if err != nil {
		return Series{}, err.Prefixf("could not get series %s: %v", req.Series, err)
	}

	// pull out the singular category
	switch len(result.Series) {
	case 0:
		return Series{}, &APIError{
			ty:  UnexpectedCount,
			msg: fmt.Sprintf("received an empty series list"),
		}
	case 1:
		return result.Series[0], nil
	default:
		return Series{}, &APIError{
			ty:  UnexpectedCount,
			msg: fmt.Sprintf("expected only a single series, received %d", len(result.Series)),
		}
	}
}

//==============================================================================
//
// GET: /fred/series/categories
//
//==============================================================================

// Response type which _should_ contain only one category.
type seriesCategoriesResponse struct {
	Categories []Category `json:"categories" xml:"categories,attr"`
}

func (c Client) CategoriesForSeries(req SeriesRequest) ([]Category, Error) {
	req.baseRequest = c.base_req

	req_url := c.base_url
	req_url.RawQuery = req.ToParams().Encode()
	req_url.Path = fmt.Sprintf("%s/series/categories", req_url.Path)

	body, err := c.get("series", req_url.String())
	if err != nil {
		return nil, err.Prefixf("error getting series' categories %s: %v", req.Series, err)
	}

	// parse the correct format
	var result seriesCategoriesResponse
	err = c.unmarshal_body(body, &result)
	if err != nil {
		return nil, err.Prefixf("could not get series' categories %s: %v", req.Series, err)
	}

	return result.Categories, nil
}

//==============================================================================
//
// GET: /fred/series/observations
//
//==============================================================================

type SeriesObservationsRequest struct {
	baseRequest
	DatedRequest
	PagedRequest

	Series           string
	ObservationStart time.Time
	ObservationEnd   time.Time
}

func NewSeriesObservationsRequest(series string, start, end time.Time) SeriesObservationsRequest {
	return SeriesObservationsRequest{
		Series:           series,
		ObservationStart: start,
		ObservationEnd:   end,
	}
}

// Satisfies the `Request` interface.
func (r SeriesObservationsRequest) ToParams() url.Values {
	v := r.baseRequest.ToParams()
	r.DatedRequest.MergeParams(v)
	r.PagedRequest.MergeParams(v)

	v.Set("series_id", r.Series)
	if r.ObservationStart.IsZero() == false {
		v.Set("observation_start", r.ObservationStart.Format(DATE_FORMAT))
	}
	if r.ObservationEnd.IsZero() == false {
		v.Set("observation_end", r.ObservationEnd.Format(DATE_FORMAT))
	}

	return v
}

// Response type which _should_ contain only one category.
type SeriesObservationsResponse struct {
	Start            Date      `json:"realtime_start" xml:"realtime_start,attr"`
	End              Date      `json:"realtime_end" xml:"realtime_end,attr"`
	ObservationStart Date      `json:"observation_start" xml:"observation_start,attr"`
	ObservationEnd   Date      `json:"observation_end" xml:"observation_end,attr"`
	Order            OrderType `json:"order_by" xml:"order_by,attr"`
	Sort             SortType  `json:"sort_order" xml:"sort_order,attr"`
	Count            uint      `json:"count" xml:"count,attr"`
	Offset           uint      `json:"offset" xml:"offset,attr"`
	Limit            uint      `json:"limit" xml:"limit,attr"`

	Units        UnitType    `json:"units" xml:"units,attr"`
	Observations []DataPoint `json:"observations" xml:"observations,attr"`
}

func (c Client) SeriesObservations(req SeriesObservationsRequest) (SeriesObservationsResponse, Error) {
	req.baseRequest = c.base_req

	req_url := c.base_url
	req_url.RawQuery = req.ToParams().Encode()
	req_url.Path = fmt.Sprintf("%s/series/observations", req_url.Path)

	body, err := c.get("series", req_url.String())
	if err != nil {
		return SeriesObservationsResponse{}, err.Prefixf("error getting series %s: %v",
			req.Series, err)
	}

	// parse the correct format
	var result SeriesObservationsResponse
	err = c.unmarshal_body(body, &result)
	if err != nil {
		return SeriesObservationsResponse{}, err.Prefixf("could not get series observations %s: %v",
			req.Series, err)
	}

	return result, err
}

//==============================================================================
//
// GET: /fred/series/search
//
//==============================================================================

type SeriesSearchRequest struct {
	baseRequest
	DatedRequest
	PagedRequest
	OrderedRequest
	FilteredRequest
	TaggedRequest

	Search     string
	SearchType SeriesSearchType
}

func NewSeriesSearchRequest(text string, ty SeriesSearchType) SeriesSearchRequest {
	return SeriesSearchRequest{
		Search:     text,
		SearchType: ty,
	}
}

func (r SeriesSearchRequest) ToParams() url.Values {
	v := r.baseRequest.ToParams()
	r.DatedRequest.MergeParams(v)
	r.PagedRequest.MergeParams(v)
	r.OrderedRequest.MergeParams(v)
	r.FilteredRequest.MergeParams(v)
	r.TaggedRequest.MergeParams(v)

	v.Set("search_text", r.Search)
	if r.SearchType != SearchTypeNone {
		v.Set("search_type", string(r.SearchType))
	}

	return v
}

type SeriesSearchResponse struct {
	Start  Date      `json:"realtime_start" xml:"realtime_start,attr"`
	End    Date      `json:"realtime_end" xml:"realtime_end,attr"`
	Order  OrderType `json:"order_by" xml:"order_by,attr"`
	Sort   SortType  `json:"sort_order" xml:"sort_order,attr"`
	Count  uint      `json:"count" xml:"count,attr"`
	Offset uint      `json:"offset" xml:"offset,attr"`
	Limit  uint      `json:"limit" xml:"limit,attr"`
	Series []Series  `json:"seriess" xml:"seriess,attr"`
}

func (c Client) SeriesSearch(req SeriesSearchRequest) (SeriesSearchResponse, Error) {
	req.baseRequest = c.base_req

	req_url := c.base_url
	req_url.RawQuery = req.ToParams().Encode()
	req_url.Path = fmt.Sprintf("%s/series/search", req_url.Path)

	body, err := c.get("series", req_url.String())
	if err != nil {
		return SeriesSearchResponse{}, err.Prefixf("error searching series '%s'", req.Search)
	}

	// parse the correct format
	var result SeriesSearchResponse
	err = c.unmarshal_body(body, &result)
	if err != nil {
		return SeriesSearchResponse{}, err.Prefixf("could not search series '%s'", req.Search)
	}

	return result, err
}

//==============================================================================
//
// GET: /fred/series/search/tags
//
//==============================================================================

type SeriesSearchTagsRequest struct {
	baseRequest
	DatedRequest
	TaggedRequest
	PagedRequest
	OrderedRequest

	TagGroup     TagId
	SeriesSearch string
	TagSearch    string
}

func NewSeriesSearchTagsRequest(series string, tags ...string) SeriesSearchTagsRequest {
	return SeriesSearchTagsRequest{
		SeriesSearch: series,
		TaggedRequest: TaggedRequest{
			Tags: tags,
		},
	}
}

func (r SeriesSearchTagsRequest) ToParams() url.Values {
	v := r.baseRequest.ToParams()
	r.DatedRequest.MergeParams(v)
	r.TaggedRequest.MergeParams(v)
	r.PagedRequest.MergeParams(v)
	r.OrderedRequest.MergeParams(v)

	v.Set("series_search_text", r.SeriesSearch)
	if len(r.TagSearch) > 0 {
		v.Set("tag_search_text", string(r.TagSearch))
	}
	if r.TagGroup != TagNone {
		v.Set("tag_group_id", string(r.TagGroup))
	}

	return v
}

type SeriesSearchTagsResponse struct {
	Start  Date      `json:"realtime_start" xml:"realtime_start,attr"`
	End    Date      `json:"realtime_end" xml:"realtime_end,attr"`
	Order  OrderType `json:"order_by" xml:"order_by,attr"`
	Sort   SortType  `json:"sort_order" xml:"sort_order,attr"`
	Count  uint      `json:"count" xml:"count,attr"`
	Offset uint      `json:"offset" xml:"offset,attr"`
	Limit  uint      `json:"limit" xml:"limit,attr"`
	Tags   []Tag     `json:"tags" xml:"tags,attr"`
}

func (c Client) SeriesSearchTags(req SeriesSearchTagsRequest) (SeriesSearchTagsResponse, Error) {
	req.baseRequest = c.base_req

	req_url := c.base_url
	req_url.RawQuery = req.ToParams().Encode()
	req_url.Path = fmt.Sprintf("%s/series/search/tags", req_url.Path)

	body, err := c.get("series", req_url.String())
	if err != nil {
		return SeriesSearchTagsResponse{}, err.Prefixf("error searching series tags '%s'", req.SeriesSearch)
	}

	// parse the correct format
	var result SeriesSearchTagsResponse
	err = c.unmarshal_body(body, &result)
	if err != nil {
		return SeriesSearchTagsResponse{}, err.Prefixf("could not search series tags '%s'", req.SeriesSearch)
	}

	return result, err
}

//==============================================================================
//
// GET: /fred/series/search/related_tags
//
//==============================================================================

func (c Client) SeriesSearchRelatedTags(req SeriesSearchTagsRequest) (SeriesSearchTagsResponse, Error) {
	req.baseRequest = c.base_req

	req_url := c.base_url
	req_url.RawQuery = req.ToParams().Encode()
	req_url.Path = fmt.Sprintf("%s/series/search/related_tags", req_url.Path)

	var result SeriesSearchTagsResponse

	body, err := c.get("series", req_url.String())
	if err != nil {
		return result, err.Prefixf("error searching series related tags '%s'", req.SeriesSearch)
	}

	// parse the correct format
	err = c.unmarshal_body(body, &result)
	if err != nil {
		return result, err.Prefixf("could not search series related tags '%s'", req.SeriesSearch)
	}

	return result, err
}
