package gofred

import (
	"testing"
	"time"
)

const (
	SERIES_GNP_ANNUAL     = "GNPCA"
	SERIES_EXCHANGE_JP_US = "EXJPUS"
)

//==============================================================================
//
// GET: /fred/series
//
//==============================================================================

func TestSeries_AnnualGNP(t *testing.T) {
	client := make_client(t)

	req := SeriesRequest{
		Series: SERIES_GNP_ANNUAL,
	}
	res, err := client.Series(req)
	if err != nil {
		t.Fatal(err)
	}

	expect_title := "Real Gross National Product"
	if res.Title != expect_title {
		t.Errorf("expected title:\n%+v\ngot:\n%+v", expect_title, res.Title)
	}

	expect_obvs_start, form_err := time.Parse(DATE_FORMAT, "1929-01-01")
	if form_err != nil {
		t.Fatalf("could not create expected start date: %v", form_err)
	}
	if time.Time(res.ObservationStart) != expect_obvs_start {
		t.Errorf("incorrect observation start time: expected %v, got %v",
			expect_obvs_start, time.Time(res.ObservationStart))
	}

	if res.SeasonallyAdjusted {
		t.Errorf("GNP should not be seasonally adjusted")
	}

	if res.Frequency != Annual {
		t.Errorf("expect GNP to be reported annually, got: %v", res.Frequency)
	}
}

func TestSeries_Nonexistant(t *testing.T) {
	client := make_client(t)
	series, err := client.Series(NewSeriesRequest("ABCD"))
	if err == nil {
		t.Fatalf("expected an error response, got: %+v", series)
	}
	if err.Type() != Invalid {
		t.Errorf("expected type:", Invalid, ", got:", err.Type())
	}
}

//==============================================================================
//
// GET: /fred/series/categories
//
//==============================================================================

func TestSeriesCategories_EchangeJpUs(t *testing.T) {
	client := make_client(t)

	req := SeriesRequest{
		Series: SERIES_EXCHANGE_JP_US,
	}
	res, err := client.CategoriesForSeries(req)
	if err != nil {
		t.Fatal(err)
	}

	if len(res) != 2 {
		t.Fatalf("expected %d categories, found %d: %+v", 2, len(res), res)
	}

	expect_monthly := Category{
		Id:       95,
		Name:     "Monthly Rates",
		ParentId: 15,
	}
	expect_japan := Category{
		Id:       275,
		Name:     "Japan",
		ParentId: 158,
	}
	for _, c := range res {
		if c != expect_monthly && c != expect_japan {
			t.Errorf("expected one of:\n%+v\nor\n%+v\nbut got:\n%+v\n", expect_monthly, expect_japan, c)
		}
	}
}
