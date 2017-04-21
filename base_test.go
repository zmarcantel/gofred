package gofred

import "testing"

func make_client_type(t *testing.T, format ResponseFormat) Client {
	client, err := NewClient(API_KEY, format)
	if err != nil {
		t.Fatalf("could not create client: %v", err)
	}

	return client
}

func make_client(t *testing.T) Client {
	return make_client_type(t, JSON)
}

func mux_filetypes(t *testing.T, test func(Client)) {
	js_client := make_client_type(t, JSON)
	xml_client := make_client_type(t, XML)

	test(js_client)
	test(xml_client)
}
