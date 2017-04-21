package gofred

import (
	"fmt"
	"net/url"
	"time"
)

type Category struct {
	Id       uint   `json:"id" xml:"id"`
	Name     string `json:"name" xml:"name"`
	ParentId uint   `json:"parent_id" xml:"parent_id"`
}

//==============================================================================
//
// GET: /fred/category
//
//==============================================================================

// Holds the data needed to request information on a `Category`.
type categoryRequest struct {
	baseRequest
	category uint
}

// Satisfies the `Request` interface.
func (r categoryRequest) ToParams() url.Values {
	v := r.baseRequest.ToParams()
	v.Set("category_id", fmt.Sprint(r.category))
	return v
}

// Response type which _should_ contain only one category.
type categoryResponse struct {
	Categories []Category `json:"categories" xml:"categories"`
}

//
// Get the `Category` information for the given category.
//
// Asserts there is only one `Category` object in the result, and returns it.
//
func (c Client) Category(category uint) (Category, Error) {
	cat_req := categoryRequest{
		baseRequest: c.base_req,
		category:    category,
	}

	req_url := c.base_url
	req_url.RawQuery = cat_req.ToParams().Encode()
	req_url.Path = fmt.Sprintf("%s/category", req_url.Path)

	body, err := c.get("category", req_url.String())
	if err != nil {
		return Category{}, err.Prefixf("error getting category %d: %v", category, err)
	}

	// parse the correct format
	var result categoryResponse
	parse_err := c.unmarshal_body(body, &result)
	if parse_err != nil {
		return Category{}, err.Prefixf("could not get category %d: %v", category, err)
	}

	// pull out the singular category
	switch len(result.Categories) {
	case 0:
		return Category{}, &APIError{
			ty:  UnexpectedCount,
			msg: fmt.Sprintf("received an empty category list"),
		}
	case 1:
		return result.Categories[0], nil
	default:
		return Category{}, &APIError{
			ty:  UnexpectedCount,
			msg: fmt.Sprintf("expected only a single category, received %d", len(result.Categories)),
		}
	}
}

//==============================================================================
//
// GET: /fred/category/children
//
//==============================================================================

// Holds the data needed to request `Category` information for the children of another category.
type categoryChildrenRequest struct {
	baseRequest
	DatedRequest
	category uint
}

// Satisfies the `Request` interface.
func (r categoryChildrenRequest) ToParams() url.Values {
	v := r.baseRequest.ToParams()
	r.DatedRequest.MergeParams(v)
	v.Set("category_id", fmt.Sprint(r.category))
	return v
}

// Internal type for parsing the response.
//
// For getter function, only the array is returned.
type categoryChildrenResponse struct {
	Categories []Category `json:"categories" xml:"categories"`
}

//
// Get the `Category` information for the children of the given category.
//
func (c Client) CategoryChildren(category uint, start, end time.Time) ([]Category, Error) {
	cat_req := categoryChildrenRequest{
		baseRequest: c.base_req,
		DatedRequest: DatedRequest{
			Start: Date(start),
			End:   Date(end),
		},
		category: category,
	}

	req_url := c.base_url
	req_url.RawQuery = cat_req.ToParams().Encode()
	req_url.Path = fmt.Sprintf("%s/category/children", req_url.Path)

	body, err := c.get("category children", req_url.String())
	if err != nil {
		return nil, err.Prefixf("error getting category %d: %v", category)
	}

	var result categoryChildrenResponse
	err = c.unmarshal_body(body, &result)
	return result.Categories, err
}

//==============================================================================
//
// GET: /fred/category/related
//
//==============================================================================

// Holds the data needed to request `Category` information for the categories
// related to a given category.
type categoryRelatedRequest struct {
	baseRequest
	DatedRequest
	category uint
}

// Satisfies the `Request` interface.
func (r categoryRelatedRequest) ToParams() url.Values {
	v := r.baseRequest.ToParams()
	r.DatedRequest.MergeParams(v)
	v.Set("category_id", fmt.Sprint(r.category))
	return v
}

// Internal type for parsing the response.
//
// For getter function, only the array is returned.
type categoryRelatedResponse struct {
	Categories []Category `json:"categories" xml:"categories"`
}

//
// Get the `Category` information for the categories related to the given category.
//
func (c Client) RelatedCategories(category uint, start, end time.Time) ([]Category, Error) {
	cat_req := categoryRelatedRequest{
		baseRequest: c.base_req,
		DatedRequest: DatedRequest{
			Start: Date(start),
			End:   Date(end),
		},
		category: category,
	}

	req_url := c.base_url
	req_url.RawQuery = cat_req.ToParams().Encode()
	req_url.Path = fmt.Sprintf("%s/category/related", req_url.Path)

	body, err := c.get("related categories", req_url.String())
	if err != nil {
		return nil, err.Prefixf("error getting categories related to %d", category)
	}

	var result categoryRelatedResponse
	err = c.unmarshal_body(body, &result)
	return result.Categories, err
}

//==============================================================================
//
// GET: /fred/category/series
//
//==============================================================================

// Holds the data needed to request the `Series` information for the category.
type CategorySeriesRequest struct {
	baseRequest
	DatedRequest
	PagedRequest
	OrderedRequest
	FilteredRequest
	TaggedRequest

	Category uint
}

func NewCategorySeriesRequest(category uint) CategorySeriesRequest {
	return CategorySeriesRequest{
		Category: category,
	}
}

// Satisfies the `Request` interface.
func (r CategorySeriesRequest) ToParams() url.Values {
	v := r.baseRequest.ToParams()
	r.DatedRequest.MergeParams(v)
	r.PagedRequest.MergeParams(v)
	r.OrderedRequest.MergeParams(v)
	r.FilteredRequest.MergeParams(v)
	r.TaggedRequest.MergeParams(v)

	v.Set("category_id", fmt.Sprint(r.Category))

	return v
}

type CategorySeriesResponse struct {
	Start  Date      `json:"realtime_start" xml:"realtime_start"` // "2013-08-13",
	End    Date      `json:"realtime_end" xml:"realtime_end"`     // "2013-08-13",
	Order  OrderType `json:"order_by" xml:"order_by"`
	Sort   SortType  `json:"sort_order" xml:"sort_order"`
	Count  uint      `json:"count" xml:"count"`
	Offset uint      `json:"offset" xml:"offset"`
	Limit  uint      `json:"limit" xml:"limit"`
	Series []Series  `json:"seriess" xml:"seriess"` // TODO: example has seriess, is typo?
}

//
// Get the `Category` information for the categories related to the given category.
//
func (c Client) SeriesInCategory(req CategorySeriesRequest) (CategorySeriesResponse, Error) {
	req.baseRequest = c.base_req

	req_url := c.base_url
	req_url.RawQuery = req.ToParams().Encode()
	req_url.Path = fmt.Sprintf("%s/category/series", req_url.Path)

	body, err := c.get("series in category", req_url.String())
	if err != nil {
		return CategorySeriesResponse{}, err.Prefixf("error getting categories related to %d", req.Category)
	}

	var result CategorySeriesResponse
	err = c.unmarshal_body(body, &result)
	return result, err
}

//==============================================================================
//
// GET: /fred/category/tags
//
//==============================================================================

// Holds the data needed to request the `Series` information for the category.
type CategoryTagsRequest struct {
	baseRequest
	DatedRequest
	PagedRequest
	OrderedRequest

	Category   uint
	TagGroupId TagId
	Search     string
}

func NewCategoryTagsRequest(category uint, tag_id TagId, search string) CategoryTagsRequest {
	return CategoryTagsRequest{
		Category:   category,
		TagGroupId: tag_id,
		Search:     search,
	}
}

// Satisfies the `Request` interface.
func (r CategoryTagsRequest) ToParams() url.Values {
	v := r.baseRequest.ToParams()
	r.DatedRequest.MergeParams(v)
	r.PagedRequest.MergeParams(v)
	r.OrderedRequest.MergeParams(v)

	v.Set("category_id", fmt.Sprint(r.Category))

	if r.TagGroupId != TagNone {
		v.Set("tag_group_id", r.TagGroupId.String())
	}
	if len(r.Search) > 0 {
		v.Set("search_text", r.Search)
	}

	return v
}

type CategoryTagsResponse struct {
	Start  Date      `json:"realtime_start" xml:"realtime_start"`
	End    Date      `json:"realtime_end" xml:"realtime_end"`
	Order  OrderType `json:"order_by" xml:"order_by"`
	Sort   SortType  `json:"sort_order" xml:"sort_order"`
	Count  uint      `json:"count" xml:"count"`
	Offset uint      `json:"offset" xml:"offset"`
	Limit  uint      `json:"limit" xml:"limit"`
	Tags   []Tag     `json:"tags" xml:"tags"`
}

//
// Get the `Category` information for the categories related to the given category.
//
func (c Client) CategoryTags(req CategoryTagsRequest) (CategoryTagsResponse, Error) {
	req.baseRequest = c.base_req

	req_url := c.base_url
	req_url.RawQuery = req.ToParams().Encode()
	req_url.Path = fmt.Sprintf("%s/category/tags", req_url.Path)

	body, err := c.get("series in category", req_url.String())
	if err != nil {
		return CategoryTagsResponse{}, err.Prefixf("error getting category tags")
	}

	var result CategoryTagsResponse
	err = c.unmarshal_body(body, &result)
	return result, err
}

//==============================================================================
//
// GET: /fred/category/related_tags
//
//==============================================================================

// Holds the data needed to request the `Series` information for the category.
type CategoryRelatedTagsRequest struct {
	baseRequest
	DatedRequest
	TaggedRequest
	PagedRequest
	OrderedRequest

	Category   uint
	TagGroupId TagId
	Search     string
}

func NewCategoryRelatedTagsRequest(category uint, tags ...string) CategoryRelatedTagsRequest {
	return CategoryRelatedTagsRequest{
		TaggedRequest: TaggedRequest{
			Tags: tags,
		},
		Category: category,
	}
}

// Satisfies the `Request` interface.
func (r CategoryRelatedTagsRequest) ToParams() url.Values {
	v := r.baseRequest.ToParams()
	r.DatedRequest.MergeParams(v)
	r.TaggedRequest.MergeParams(v)
	r.PagedRequest.MergeParams(v)
	r.OrderedRequest.MergeParams(v)

	v.Set("category_id", fmt.Sprint(r.Category))

	if r.TagGroupId != TagNone {
		v.Set("tag_group_id", r.TagGroupId.String())
	}
	if len(r.Search) > 0 {
		v.Set("search_text", r.Search)
	}

	return v
}

type CategoryRelatedTagsResponse struct {
	Start  Date      `json:"realtime_start" xml:"realtime_start"`
	End    Date      `json:"realtime_end" xml:"realtime_end"`
	Order  OrderType `json:"order_by" xml:"order_by"`
	Sort   SortType  `json:"sort_order" xml:"sort_order"`
	Count  uint      `json:"count" xml:"count"`
	Offset uint      `json:"offset" xml:"offset"`
	Limit  uint      `json:"limit" xml:"limit"`
	Tags   []Tag     `json:"tags" xml:"tags"`
}

//
// Get the `Category` information for the categories related to the given category.
//
func (c Client) CategoryRelatedTags(req CategoryRelatedTagsRequest) (CategoryRelatedTagsResponse, Error) {
	req.baseRequest = c.base_req

	req_url := c.base_url
	req_url.RawQuery = req.ToParams().Encode()
	req_url.Path = fmt.Sprintf("%s/category/related_tags", req_url.Path)

	body, err := c.get("series in category", req_url.String())
	if err != nil {
		return CategoryRelatedTagsResponse{}, err.Prefixf("error getting category tags")
	}

	var result CategoryRelatedTagsResponse
	err = c.unmarshal_body(body, &result)
	return result, err
}
