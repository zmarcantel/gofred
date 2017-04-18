package gofred

import (
	"fmt"
	"net/url"
	"time"
)

type Category struct {
	Id       uint32 `json:"id" xml:"id"`
	Name     string `json:"name" xml:"name"`
	ParentId uint32 `json:"parent_id" xml:"parent_id"`
}

//==============================================================================
//
// GET: /fred/category
//
//==============================================================================

// Holds the data needed to request information on a `Category`.
type CategoryRequest struct {
	BaseRequest
	category uint32
}

// Satisfies the `Request` interface.
func (r CategoryRequest) ToParams() url.Values {
	v := r.BaseRequest.ToParams()
	v.Set("category_id", fmt.Sprint(r.category))
	return v
}

// Response type which _should_ contain only one category.
type CategoryResponse struct {
	Categories []Category `json:"categories" xml:"categories"`
}

//
// Get the `Category` information for the given category.
//
// Asserts there is only one `Category` object in the result, and returns it.
//
func (c Client) Category(category uint32) (Category, Error) {
	cat_req := CategoryRequest{
		BaseRequest: c.base_req,
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
	var result CategoryResponse
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
type CategoryChildrenRequest struct {
	BaseRequest
	category uint32
	start    time.Time
	end      time.Time
}

// Satisfies the `Request` interface.
func (r CategoryChildrenRequest) ToParams() url.Values {
	v := r.BaseRequest.ToParams()
	v.Set("category_id", fmt.Sprint(r.category))
	v.Set("realtime_start", r.start.Format("2006-01-02"))
	v.Set("realtime_end", r.end.Format("2006-01-02"))
	return v
}

// Internal type for parsing the response.
//
// For getter function, only the array is returned.
type CategoryChildrenResponse struct {
	Categories []Category `json:"categories" xml:"categories"`
}

//
// Get the `Category` information for the children of the given category.
//
func (c Client) CategoryChildren(category uint32, start, end time.Time) ([]Category, Error) {
	cat_req := CategoryChildrenRequest{
		BaseRequest: c.base_req,
		category:    category,
		start:       start,
		end:         end,
	}

	req_url := c.base_url
	req_url.RawQuery = cat_req.ToParams().Encode()
	req_url.Path = fmt.Sprintf("%s/category/children", req_url.Path)

	body, err := c.get("category children", req_url.String())
	if err != nil {
		return nil, err.Prefixf("error getting category %d: %v", category)
	}

	var result CategoryChildrenResponse
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
type CategoryRelatedRequest struct {
	BaseRequest
	category uint32
	start    time.Time
	end      time.Time
}

// Satisfies the `Request` interface.
func (r CategoryRelatedRequest) ToParams() url.Values {
	v := r.BaseRequest.ToParams()
	v.Set("category_id", fmt.Sprint(r.category))
	v.Set("realtime_start", r.start.Format("2006-01-02"))
	v.Set("realtime_end", r.end.Format("2006-01-02"))
	return v
}

// Internal type for parsing the response.
//
// For getter function, only the array is returned.
type CategoryRelatedResponse struct {
	Categories []Category `json:"categories" xml:"categories"`
}

//
// Get the `Category` information for the categories related to the given category.
//
func (c Client) RelatedCategories(category uint32, start, end time.Time) ([]Category, Error) {
	cat_req := CategoryRelatedRequest{
		BaseRequest: c.base_req,
		category:    category,
		start:       start,
		end:         end,
	}

	req_url := c.base_url
	req_url.RawQuery = cat_req.ToParams().Encode()
	req_url.Path = fmt.Sprintf("%s/category/related", req_url.Path)

	body, err := c.get("related categories", req_url.String())
	if err != nil {
		return nil, err.Prefixf("error getting categories related to %d", category)
	}

	var result CategoryRelatedResponse
	err = c.unmarshal_body(body, &result)
	return result.Categories, err
}
