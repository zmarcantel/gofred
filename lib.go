package gofred

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	// Base API endpoint URL all requests go through
	API_URL = "https://api.stlouisfed.org/fred"

	DATE_FORMAT = "2006-01-02"
	TIME_FORMAT = "2006-01-02 15:04:05-07"
)

//==============================================================================
// client config
//==============================================================================

// Enum type that allows selecting the response format.
type ResponseFormat uint8

const (
	// Get responses in JSON format
	JSON ResponseFormat = iota
	// Get responses in XML format
	XML ResponseFormat = iota + 1
)

// Get the string representation of the response format as it should be used in a URL param.
func (f ResponseFormat) String() string {
	switch f {
	case JSON:
		return "json"
	case XML:
		return "xml"
	default:
		return "ERROR" // TODO
	}
}

//==============================================================================
// requests
//==============================================================================

type ApiKey string

func (k ApiKey) String() string {
	return "REDACTED"
}

// Interface to a generic request to the Fred API.
type Request interface {
	ToParams() url.Values
	MergeParams(url.Values)
}

// Minimal shared request objects.
//
// A `baseRequest` object is held in the client and copied for every request
// generated through the library. Once copied, the request-specific parameters
// are filled in.
type baseRequest struct {
	fmt     ResponseFormat
	api_key ApiKey
}

// Satisfies the `Request` interface, generating a `url.Values` set with the API key and format.
func (r baseRequest) ToParams() url.Values {
	v := url.Values{}
	r.MergeParams(v)
	return v
}

func (r baseRequest) MergeParams(v url.Values) {
	v.Set("api_key", string(r.api_key))
	v.Set("file_type", r.fmt.String())
}

type DatedRequest struct {
	Start Date `json:"realtime_start" xml:"realtime_start"`
	End   Date `json:"realtime_end" xml:"realtime_end"`
}

func (r DatedRequest) ToParams() url.Values {
	v := url.Values{}
	r.MergeParams(v)
	return v
}

func (r DatedRequest) MergeParams(v url.Values) {
	if time.Time(r.Start).IsZero() == false {
		v.Set("realtime_start", time.Time(r.Start).Format(DATE_FORMAT))
	}
	if time.Time(r.End).IsZero() == false {
		v.Set("realtime_end", time.Time(r.End).Format(DATE_FORMAT))
	}
}

// Embedded struct for requests with offset/limit params
type PagedRequest struct {
	Limit  uint // 1 - 1000
	Offset uint // 0 - 999
}

func (r PagedRequest) ToParams() url.Values {
	v := url.Values{}
	r.MergeParams(v)
	return v
}

func (r PagedRequest) MergeParams(v url.Values) {
	if r.Limit > 0 {
		adj := r.Limit % 1000 // limit is max 1000
		if adj == 0 {
			adj = 1
		}
		v.Set("limit", fmt.Sprint(adj))
	}
	if r.Offset > 0 {
		v.Set("offset", fmt.Sprint(r.Offset%999)) // max offset = max limit - 1
	}
}

// Embedded struct for requests with order params
type OrderedRequest struct {
	Order OrderType
	Sort  SortType
}

func (r OrderedRequest) ToParams() url.Values {
	v := url.Values{}
	r.MergeParams(v)
	return v
}

func (r OrderedRequest) MergeParams(v url.Values) {
	if len(r.Order) > 0 {
		v.Set("order_by", fmt.Sprint(r.Order))
	}
	if len(r.Sort) > 0 {
		v.Set("sort_order", fmt.Sprint(r.Sort))
	}
}

// Embedded struct for requests with filter params
type FilteredRequest struct {
	Variable FilterType
	Value    string // TODO: ToString interface exist?
}

func (r FilteredRequest) ToParams() url.Values {
	v := url.Values{}
	r.MergeParams(v)
	return v
}

func (r FilteredRequest) MergeParams(v url.Values) {
	if len(r.Variable) > 0 {
		v.Set("filter_variable", string(r.Variable))
	}
	if len(r.Value) > 0 {
		v.Set("filter_value", r.Value)
	}
}

// Embedded struct for requests with tag params
type TaggedRequest struct {
	Tags    []string
	Exclude []string
}

func (r TaggedRequest) ToParams() url.Values {
	v := url.Values{}
	r.MergeParams(v)
	return v
}

func (r TaggedRequest) MergeParams(v url.Values) {
	if len(r.Tags) > 0 {
		v.Set("tag_names", strings.Join(r.Tags, ";"))
	}
	if len(r.Exclude) > 0 {
		v.Set("exclude_tag_name", strings.Join(r.Exclude, ";"))
	}
}

//==============================================================================
// API internal types
//==============================================================================

type SeasonalAdjustment bool

const (
	Adjusted   SeasonalAdjustment = true
	Unadjusted SeasonalAdjustment = false
)

func ParseSeasonalAdjustment(str string) (SeasonalAdjustment, error) {
	switch str {
	case "Not Seasonally Adjusted":
		return Unadjusted, nil

	case "Seasonally Adjusted":
		return Adjusted, nil
	case "Seasonally Adjusted Annual Rate":
		return Adjusted, nil
	}

	return Unadjusted, fmt.Errorf("unknown seasonal adjustment string: %s", str)
}

func (a SeasonalAdjustment) String() string {
	if a {
		return "\"Seasonally Adjusted\""
	}

	return "\"Not Seasonally Adjusted\""
}

func (a *SeasonalAdjustment) UnmarshalJSON(input []byte) error {
	var as_str string
	err := json.Unmarshal(input, &as_str)
	if err != nil {
		return err
	}

	*a, err = ParseSeasonalAdjustment(as_str)

	return err
}

func (a SeasonalAdjustment) MarshalJSON() ([]byte, error) {
	return []byte(a.String()), nil
}
func (a *SeasonalAdjustment) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return d.DecodeElement(a, &start)
}
func (a *SeasonalAdjustment) UnmarshalXMLAttr(attr xml.Attr) error {
	var err error
	*a, err = ParseSeasonalAdjustment(attr.Value)
	return err
}

type Date time.Time

func (d *Date) UnmarshalJSON(input []byte) error {
	var as_str string
	err := json.Unmarshal(input, &as_str)
	if err != nil {
		return err
	}

	as_time, err := time.Parse(DATE_FORMAT, as_str)
	*d = Date(as_time)
	return err
}

func (d Date) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(d).Format(DATE_FORMAT))
}

func (d *Date) UnmarshalXMLAttr(attr xml.Attr) error {
	as_time, err := time.Parse(DATE_FORMAT, attr.Value)
	*d = Date(as_time)
	return err
}

type DateTime time.Time

func (d *DateTime) UnmarshalJSON(input []byte) error {
	var as_str string
	if err := json.Unmarshal(input, &as_str); err != nil {
		return err
	}

	as_time, err := time.Parse(TIME_FORMAT, as_str)
	*d = DateTime(as_time)
	return err
}

func (d DateTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(d).Format(TIME_FORMAT))
}

func (d *DateTime) UnmarshalXMLAttr(attr xml.Attr) error {
	as_time, err := time.Parse(TIME_FORMAT, attr.Value)
	*d = DateTime(as_time)
	return err
}

type Frequency uint8

const (
	Daily = iota
	Weekly
	Biweekly
	Monthly
	Quarterly
	Semiannual
	Annual

	WeeklyEndingFriday
	WeeklyEndingThursday
	WeeklyEndingWednesday
	WeeklyEndingTuesday
	WeeklyEndingMonday
	WeeklyEndingSunday
	WeeklyEndingSaturday
	BiweeklyEndingWednesday
	BiweeklyEndingMonday

	UnknownFrequency
)

func FrequencyFromString(str string) (Frequency, error) {
	switch str {
	case "d", "Daily", "Daily, 7-Day",
		"Daily, Close":
		return Daily, nil
	case "w", "Weekly":
		return Weekly, nil
	case "bw", "Biweekly":
		return Biweekly, nil
	case "m", "Monthly", "Monthly, End of Month":
		return Monthly, nil
	case "q", "Quarterly", "Quarterly, End of Quarter", "Quarterly, End of Period":
		return Quarterly, nil
	case "sa", "Semiannual":
		return Semiannual, nil
	case "a", "Annual",
		"Annual, Fiscal Year", "Annual, As of February", "Annual, End of Year":
		return Annual, nil

	case "wef", "Weekly, Ending Friday":
		return WeeklyEndingFriday, nil
	case "weth", "Weekly, Ending Thursday":
		return WeeklyEndingThursday, nil
	case "wew", "Weekly, Ending Wednesday":
		return WeeklyEndingWednesday, nil
	case "wetu", "Weekly, Ending Tuesday":
		return WeeklyEndingTuesday, nil
	case "wem", "Weekly, Ending Monday":
		return WeeklyEndingMonday, nil
	case "wesu", "Weekly, Ending Sunday":
		return WeeklyEndingSunday, nil
	case "wesa", "Weekly, Ending Saturday":
		return WeeklyEndingSaturday, nil
	case "bwew", "Biweekly, Ending Wednesday":
		return BiweeklyEndingWednesday, nil
	case "bwem", "Biweekly, Ending Monday":
		return BiweeklyEndingMonday, nil

	case "Not Applicable":
		return UnknownFrequency, nil
	}

	return UnknownFrequency, fmt.Errorf("unknown frequency format: %s", str)
}

func (f *Frequency) UnmarshalJSON(input []byte) error {
	*f = UnknownFrequency

	var as_str string
	err := json.Unmarshal(input, &as_str)
	if err != nil {
		return err
	}

	*f, err = FrequencyFromString(as_str)
	return err
}

func (f Frequency) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", f.LongString())), nil
}

func (f *Frequency) UnmarshalXMLAttr(attr xml.Attr) error {
	var err error
	*f, err = FrequencyFromString(attr.Value)
	return err
}

func (f Frequency) String() string {
	switch f {
	case Daily:
		return "d"
	case Weekly:
		return "w"
	case Biweekly:
		return "bw"
	case Monthly:
		return "m"
	case Quarterly:
		return "q"
	case Semiannual:
		return "sa"
	case Annual:
		return "a"
	case WeeklyEndingFriday:
		return "wef"
	case WeeklyEndingThursday:
		return "weth"
	case WeeklyEndingWednesday:
		return "wew"
	case WeeklyEndingTuesday:
		return "wetu"
	case WeeklyEndingMonday:
		return "wem"
	case WeeklyEndingSunday:
		return "wesu"
	case WeeklyEndingSaturday:
		return "wesa"
	case BiweeklyEndingWednesday:
		return "bwew"
	case BiweeklyEndingMonday:
		return "bwem"
	}

	return "unknown frequency"
}

func (f Frequency) LongString() string {
	switch f {
	case Daily:
		return "Daily"
	case Weekly:
		return "Weekly"
	case Biweekly:
		return "Biweekly"
	case Monthly:
		return "Monthly"
	case Quarterly:
		return "Quarterly"
	case Semiannual:
		return "Semiannual"
	case Annual:
		return "Annual"
	case WeeklyEndingFriday:
		return "Weekly, Ending Friday"
	case WeeklyEndingThursday:
		return "Weekly, Ending Thursday"
	case WeeklyEndingWednesday:
		return "Weekly, Ending Wednesday"
	case WeeklyEndingTuesday:
		return "Weekly, Ending Tuesday"
	case WeeklyEndingMonday:
		return "Weekly, Ending Monday"
	case WeeklyEndingSunday:
		return "Weekly, Ending Sunday"
	case WeeklyEndingSaturday:
		return "Weekly, Ending Saturday"
	case BiweeklyEndingWednesday:
		return "Biweekly, Ending Wednesday"
	case BiweeklyEndingMonday:
		return "Biweekly, Ending Monday"
	}

	return "unknown frequency"
}

type DataPoint struct {
	Date  Date    `json:"date" xml:"date"`
	Value float64 `json:"value" xml:"value"`
	Valid bool
}

func (d *DataPoint) UnmarshalJSON(input []byte) error {
	d.Valid = false
	var as_map map[string]string
	if err := json.Unmarshal(input, &as_map); err != nil {
		return err
	}

	date, exists := as_map["date"]
	if !exists {
		return fmt.Errorf("no date in datapoint")
	}

	value, exists := as_map["value"]
	if !exists {
		return fmt.Errorf("no value in datapoint")
	}
	if value == "." {
		return nil // TODO: can we sentinel with error rather than filtering
	}

	as_time, err := time.Parse(DATE_FORMAT, date)
	if err != nil {
		return err
	}
	d.Date = Date(as_time)

	d.Value, err = strconv.ParseFloat(value, 64)
	if err != nil {
		fmt.Println(string(input))
		return fmt.Errorf("could not parse '%s': %v", value, err)
	}

	d.Valid = true
	return nil
}

// ordering
type OrderType string

const (
	OrderId                 OrderType = "series_id"
	OrderGroupId            OrderType = "group_id"
	OrderTitle              OrderType = "title"
	OrderName               OrderType = "name"
	OrderCreated            OrderType = "created"
	OrderSeriesCount        OrderType = "series_count"
	OrderUnits              OrderType = "units"
	OrderFrequency          OrderType = "frequency"
	OrderSeasonalAdjustment OrderType = "seasonal_adjustment"
	OrderStart              OrderType = "realtime_start"
	OrderEnd                OrderType = "realtime_end"
	OrderLastUpdated        OrderType = "last_updated"
	OrderObservationStart   OrderType = "observation_start"
	OrderObservationEnd     OrderType = "observation_end"
	OrderPopularity         OrderType = "popularity"
)

// filter
type FilterType string

const (
	FilterFrequency        FilterType = "frequency"
	FilterUnits            FilterType = "units"
	FilterSeasonalAdjusted FilterType = "seasonal_adjustment"
	FilterAll              FilterType = "all"
	FilterRegional         FilterType = "regional"
	FilterMacro            FilterType = "macro"
)

// sort
type SortType string

const (
	SortAscending  SortType = "asc"
	SortDescending SortType = "desc"
)

// tags
type TagId string

const (
	TagNone               TagId = ""
	TagFrequency          TagId = "Frequency"
	TagGeneral            TagId = "General or Concept"
	TagGeography          TagId = "Geography"
	TagGeographyType      TagId = "Geography Type"
	TagRelease            TagId = "Release"
	TagSeasonalAdjustment TagId = "Seasonal Adjustment"
	TagSource             TagId = "Source"
)

func TagIdFromString(str string) TagId {
	switch str {
	case "freq":
		return TagFrequency
	case "gen":
		return TagGeneral
	case "geo":
		return TagGeography
	case "geot":
		return TagGeographyType
	case "rls":
		return TagRelease
	case "seas":
		return TagSeasonalAdjustment
	case "src":
		return TagSource
	}

	return TagNone
}

func (t *TagId) UnmarshalJSON(input []byte) error {
	var as_str string
	if err := json.Unmarshal(input, &as_str); err != nil {
		return err
	}

	*t = TagIdFromString(as_str)

	if *t == TagNone {
		return fmt.Errorf("unknown tag id '%s'", as_str)
	}
	return nil
}

func (t *TagId) UnmarshalXMLAttr(attr xml.Attr) error {
	*t = TagIdFromString(attr.Value)
	return nil
}

func (t TagId) String() string {
	switch t {
	case TagFrequency:
		return "freq"
	case TagGeneral:
		return "gen"
	case TagGeography:
		return "geo"
	case TagGeographyType:
		return "geot"
	case TagRelease:
		return "rls"
	case TagSeasonalAdjustment:
		return "seas"
	case TagSource:
		return "src"
	}

	return "unknown tag id"
}

func (t TagId) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

// units

type UnitType uint

const (
	UnitLinear                                   UnitType = 0
	UnitChange                                   UnitType = 1
	UnitChangeFromYearAgo                        UnitType = 2
	UnitPercentChange                            UnitType = 3
	UnitPercentChangeFromYearAgo                 UnitType = 4
	UnitCompoundedAnnualRateOfChange             UnitType = 5
	UnitContinuouslyCompoundedRateOfChange       UnitType = 6
	UnitContinuouslyCompoundedAnnualRateOfChange UnitType = 7
	UnitNaturalLog                               UnitType = 8
)

func (u UnitType) String() string {
	switch u {
	case UnitLinear:
		return "Linear"
	case UnitChange:
		return "Change"
	case UnitChangeFromYearAgo:
		return "Change from Year Ago"
	case UnitPercentChange:
		return "Percent Change"
	case UnitPercentChangeFromYearAgo:
		return "Percent Change from Year Ago"
	case UnitCompoundedAnnualRateOfChange:
		return "Compounded Annual Rate of Change"
	case UnitContinuouslyCompoundedRateOfChange:
		return "Continuously Compounded Rate of Change"
	case UnitContinuouslyCompoundedAnnualRateOfChange:
		return "Continuously Compounded Annual Rate of Change"
	case UnitNaturalLog:
		return "Natural Log"
	}

	return "unknwon unit"
}

func UnitTypeFromString(str string) (UnitType, error) {
	switch str {
	case "lin":
		return UnitLinear, nil
	case "chg":
		return UnitChange, nil
	case "ch1":
		return UnitChangeFromYearAgo, nil
	case "pch":
		return UnitPercentChange, nil
	case "pc1":
		return UnitPercentChangeFromYearAgo, nil
	case "pca":
		return UnitCompoundedAnnualRateOfChange, nil
	case "cch":
		return UnitContinuouslyCompoundedRateOfChange, nil
	case "cca":
		return UnitContinuouslyCompoundedAnnualRateOfChange, nil
	case "log":
		return UnitNaturalLog, nil
	}

	return UnitLinear, fmt.Errorf("unknown unit: %s", str)
}

func (u *UnitType) UnmarshalJSON(input []byte) error {
	var as_str string
	err := json.Unmarshal(input, &as_str)
	if err != nil {
		return err
	}

	*u, err = UnitTypeFromString(as_str)
	return err
}

func (u *UnitType) UnmarshalXMLAttr(attr xml.Attr) error {
	var err error
	*u, err = UnitTypeFromString(attr.Value)
	return err
}

func (u UnitType) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.String())
}

//==============================================================================
// errors
//==============================================================================

// Error type describing the various internal and API errors.
type ErrorType uint32

const (
	ParseError            ErrorType = 0 // internal errors
	HTTPError             ErrorType = 1
	ReadError             ErrorType = 2
	UnexpectedCount       ErrorType = 3
	UnknownResponseFormat ErrorType = 4
	NotFound              ErrorType = 404 // HTTP errors
	Invalid               ErrorType = 400
	UnknownError          ErrorType = 999 // misc
)

type Error interface {
	error
	Type() ErrorType
	Prefix(string) Error
	Prefixf(string, ...interface{}) Error
}

type APIError struct {
	ty  ErrorType
	msg string
}

func (e *APIError) Error() string   { return e.msg }
func (e *APIError) Type() ErrorType { return e.ty }
func (e *APIError) Prefix(p string) Error {
	e.msg = fmt.Sprintf("%s %v", p, e.msg)
	return e
}
func (e *APIError) Prefixf(f string, args ...interface{}) Error {
	e.msg = fmt.Sprintf("%s %v", fmt.Sprintf(f, args...), e.msg)
	return e
}

// Generic error response type.
//
// If a non-success return code is returned, this type is expected to be parseable.
type baseError struct {
	Message string `json:"error_message" xml:"error_message"`
	Code    uint32 `json:"error_code" xml:"error_code"`
}

//==============================================================================
// client
//==============================================================================

// Main interface to the API.
//
// Requires specifying the API key and response format for all future requests
// through this client.
type Client struct {
	base_req baseRequest
	base_url url.URL
}

// Create a new client with the given API key and response format.
func NewClient(key string, format ResponseFormat) (Client, error) {
	if len(key) != 32 {
		return Client{}, fmt.Errorf("api key is invalid length")
	}

	api_url, err := url.Parse(API_URL)
	if err != nil {
		return Client{}, err
	}

	return Client{
		base_req: baseRequest{
			fmt:     format,
			api_key: ApiKey(key),
		},
		base_url: *api_url,
	}, nil
}

// Unmarshals the byte slice into the target interface based on the internal
// response format given when the client was created.
func (c Client) unmarshal_body(body []byte, into interface{}) Error {
	switch c.base_req.fmt {
	case JSON:
		err := json.Unmarshal(body, into)
		if err != nil {
			return &APIError{
				ty:  ParseError,
				msg: fmt.Sprintf("failed to parse json response: %v", err),
			}
		}
	case XML:
		err := xml.Unmarshal(body, into)
		if err != nil {
			return &APIError{
				ty:  ParseError,
				msg: fmt.Sprintf("failed to parse xml response: %v", err),
			}
		}
	default:
		return &APIError{
			ty:  UnknownResponseFormat,
			msg: fmt.Sprintf("unknown request/response type: %v", c.base_req.fmt),
		}
	}

	return nil
}

// Parses the byte slice as a `baseError` depending on the response format.
func (c Client) get_error(body []byte) (baseError, Error) {
	var result baseError
	switch c.base_req.fmt {
	case JSON:
		err := json.Unmarshal(body, &result)
		if err != nil {
			return baseError{}, &APIError{
				ty:  ParseError,
				msg: fmt.Sprintf("failed to parse json error response: %v", err),
			}
		}
	case XML:
		err := xml.Unmarshal(body, &result)
		if err != nil {
			return baseError{}, &APIError{
				ty:  ParseError,
				msg: fmt.Sprintf("failed to parse xml error response: %v", err),
			}
		}
	default:
		return baseError{}, &APIError{
			ty:  UnknownResponseFormat,
			msg: fmt.Sprintf("unknown request/response type: %v", c.base_req.fmt),
		}
	}

	return result, nil
}

// Wrapper around `http.Get()` which checks status codes and proxies back either a
// valid response or a parsed/generated error.
func (c Client) get(desc, req_url string) ([]byte, Error) {
	res, err := http.Get(req_url)
	if err != nil {
		return nil, &APIError{ty: HTTPError, msg: err.Error()}
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, &APIError{ty: ReadError, msg: err.Error()}
	}

	// we need this a few times
	failed_to_parse := &APIError{
		ty:  ParseError,
		msg: fmt.Sprintf("failed to parse %s error response: %v", desc, err),
	}

	// catch early errors
	switch res.StatusCode {
	case 200:
		// do nothing

	// not found (endpoint, seems to not be returned by API)
	case 404:
		return nil, &APIError{
			ty:  NotFound,
			msg: fmt.Sprintf("could not find %s: %d", desc, res.StatusCode),
		}

	// invalid request
	case 400:
		req_err, err := c.get_error(body)
		if err != nil {
			return nil, failed_to_parse
		}
		return nil, &APIError{
			ty:  Invalid,
			msg: fmt.Sprintf("invalid %s request: %s", desc, req_err.Message),
		}

	// anything else
	default:
		req_err, err := c.get_error(body)
		if err != nil {
			return nil, failed_to_parse
		}
		return nil, &APIError{
			ty:  UnknownError,
			msg: fmt.Sprintf("could not get %s (%d): %v", desc, req_err.Code, req_err.Message),
		}
	}

	return body, nil
}
