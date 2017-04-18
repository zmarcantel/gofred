package gofred

import (
	"testing"
	"time"
)

const (
	CATEGORY_TRADE_BALANCE           = 125
	CATEGORY_STLOUIS_DISTRICT_STATES = 32073
)

func make_client(t *testing.T) Client {
	client, err := NewClient(API_KEY, JSON)
	if err != nil {
		t.Fatalf("could not create client: %v", err)
	}

	return client
}

//==============================================================================
//
// GET: /fred/category
//
//==============================================================================

func TestCategory_TradeBalance(t *testing.T) {
	cat := Category{
		Id:       CATEGORY_TRADE_BALANCE,
		Name:     "Trade Balance",
		ParentId: 13,
	}

	client := make_client(t)
	res, err := client.Category(cat.Id)
	if err != nil {
		t.Fatal(err)
	}

	if res != cat {
		t.Errorf("expected response:\n%+v\ngot:\n%+v", cat, res)
	}
}

func TestCategory_Nonexistant(t *testing.T) {
	client := make_client(t)
	cat, err := client.Category(999999) // as of writing, this returns a 400 response
	if err == nil {
		t.Errorf("expected an error response, got: %+v", cat)
	}
}

//==============================================================================
//
// GET: /fred/category/children
//
//==============================================================================

func TestCategoryChildren_TradeBalance(t *testing.T) {
	client := make_client(t)
	trade, err := client.Category(CATEGORY_TRADE_BALANCE)
	if err != nil {
		t.Fatal(err)
	}

	children, err := client.CategoryChildren(trade.ParentId, time.Unix(0, 0), time.Now())
	if err != nil {
		t.Fatal(err)
	}

	if len(children) == 0 {
		t.Fatalf("got no children in response: %+v", children)
	}

	found := false
	for _, cat := range children {
		if cat.Id == CATEGORY_TRADE_BALANCE {
			found = true
			break
		}
	}
	if found == false {
		t.Errorf("expected to find category ID: %d, in children:\n%+v", CATEGORY_TRADE_BALANCE, children)
	}
}

//==============================================================================
//
// GET: /fred/category/related
//
//==============================================================================

func TestCategoryRelated_TradeBalance(t *testing.T) {
	client := make_client(t)
	related, err := client.RelatedCategories(CATEGORY_STLOUIS_DISTRICT_STATES, time.Unix(0, 0), time.Now())
	if err != nil {
		t.Fatal(err)
	}

	expect_map := map[uint32]Category{
		149: Category{Id: 149, Name: "Arkansas", ParentId: 27281},
		150: Category{Id: 150, Name: "Illinois", ParentId: 27281},
		151: Category{Id: 151, Name: "Indiana", ParentId: 27281},
		152: Category{Id: 152, Name: "Kentucky", ParentId: 27281},
		153: Category{Id: 153, Name: "Mississippi", ParentId: 27281},
		154: Category{Id: 154, Name: "Missouri", ParentId: 27281},
		193: Category{Id: 193, Name: "Tennessee", ParentId: 27281},
	}

	if len(related) != len(expect_map) {
		t.Fatalf("incorrect number of categories in response, expected: %d, got %d",
			len(related), len(expect_map))
	}

	found := 0
	for _, cat := range related {
		expect, exists := expect_map[cat.Id]
		if exists == false {
			t.Errorf("did not find an expected state in district with ID: %d", cat.Id)
			continue
		}

		found += 1
		if cat != expect {
			t.Errorf("expected:\n%+v\ngot:\n%+v", expect, cat)
		}
	}

	if found != len(expect_map) {
		t.Errorf("expected to find all %d states, found: %d", len(expect_map), found)
	}
}
