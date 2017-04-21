package gofred

import (
	"fmt"
	"net/url"
)

type Series struct {
	Id                 string             `json:"id" xml:"id"`
	Start              Date               `json:"realtime_start" xml:"realtime_start"`
	End                Date               `json:"realtime_end" xml:"realtime_end"`
	Title              string             `json:"title" xml:"title"`
	ObservationStart   Date               `json:"observation_start" xml:"observation_start"`
	ObservationEnd     Date               `json:"observation_end" xml:"observation_end"`
	Frequency          Frequency          `json:"frequency" xml:"frequency"`
	Units              string             `json:"units" xml:"units"`
	UnitsShort         string             `json:"units_short" xml:"units_short"`
	SeasonallyAdjusted SeasonalAdjustment `json:"seasonal_adjustment" xml:"seasonal_adjustment"`
	LastUpdate         DateTime           `json:"last_updated" xml:"last_updated"`
	Popularity         uint16             `json:"popularity" xml:"popularity"` // TODO: type check
	Notes              string             `json:"notes" xml:"notes"`
}

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
	Start  Date     `json:"realtime_start" xml:"realtime_start"`
	End    Date     `json:"realtime_end" xml:"realtime_end"`
	Series []Series `json:"seriess" xml:"seriess"`
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
	Categories []Category `json:"categories" xml:"categories"`
}

//
// Get the `Series` information for the given series ID.
//
// Asserts there is only one `Series` object in the result, and returns it.
//
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
