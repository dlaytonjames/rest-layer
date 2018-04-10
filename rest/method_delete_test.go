package rest_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/rs/rest-layer-mem"
	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/schema"
	"github.com/rs/rest-layer/schema/query"
)

func TestDeleteList(t *testing.T) {
	sharedInit := func() *requestTestVars {
		s := mem.NewHandler()
		s.Insert(context.Background(), []*resource.Item{
			{ID: "1", Payload: map[string]interface{}{"id": "1", "foo": "odd"}},
			{ID: "2", Payload: map[string]interface{}{"id": "2", "foo": "even"}},
			{ID: "3", Payload: map[string]interface{}{"id": "3", "foo": "odd"}},
			{ID: "4", Payload: map[string]interface{}{"id": "4", "foo": "even"}},
			{ID: "5", Payload: map[string]interface{}{"id": "5", "foo": "odd"}},
		})

		idx := resource.NewIndex()
		idx.Bind("foo", schema.Schema{
			Fields: schema.Fields{
				"id":  {Sortable: true, Filterable: true},
				"foo": {Filterable: true},
			},
		}, s, resource.Conf{AllowedModes: resource.ReadWrite})

		return &requestTestVars{
			Index:   idx,
			Storers: map[string]resource.Storer{"foo": s},
		}
	}
	checkFooIDs := func(ids ...interface{}) requestCheckerFunc {
		return func(t *testing.T, vars *requestTestVars) {
			s := vars.Storers["foo"]
			items, err := s.Find(context.Background(), &query.Query{Sort: query.Sort{{Name: "id", Reversed: false}}})
			if err != nil {
				t.Errorf("s.Find failed: %s", err)
			}
			if el, al := len(ids), len(items.Items); el != al {
				t.Errorf("Expected resource 'foo' to contain %d items, got %d", el, al)
				return
			}
			for i, eid := range ids {
				if aid := items.Items[i].ID; eid != aid {
					el := len(ids)
					t.Errorf("Expected item %d/%d to have ID %q, got ID %q", i+1, el, eid, aid)
				}
			}
		}
	}

	tests := map[string]requestTest{
		`clearAll`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("DELETE", "/foo", nil)
			},
			ResponseCode:   http.StatusNoContent,
			ResponseBody:   ``,
			ResponseHeader: http.Header{"X-Total": []string{"5"}},
			ExtraTest:      checkFooIDs(),
		},
		`limit=2`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("DELETE", `/foo?limit=2`, nil)
			},
			ResponseCode:   http.StatusNoContent,
			ResponseBody:   ``,
			ResponseHeader: http.Header{"X-Total": []string{"2"}},
			ExtraTest:      checkFooIDs("3", "4", "5"),
		},
		`limit=2,skip=1`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("DELETE", `/foo?limit=2&skip=1`, nil)
			},
			ResponseCode:   http.StatusNoContent,
			ResponseBody:   ``,
			ResponseHeader: http.Header{"X-Total": []string{"2"}},
			ExtraTest:      checkFooIDs("1", "4", "5"),
		},
		`filter=invalid`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("DELETE", `/foo?filter=invalid`, nil)
			},
			ResponseCode: http.StatusUnprocessableEntity,
			ResponseBody: `{
				"code": 422,
				"message": "URL parameters contain error(s)",
				"issues": {
					"filter": ["char 0: expected '{' got 'i'"]
				}}`,
			ExtraTest: checkFooIDs("1", "2", "3", "4", "5"),
		},
		`filter={foo:"even"}`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("DELETE", `/foo?filter={foo:"even"}`, nil)
			},
			ResponseCode:   http.StatusNoContent,
			ResponseBody:   ``,
			ResponseHeader: http.Header{"X-Total": []string{"2"}},
			ExtraTest:      checkFooIDs("1", "3", "5"),
		},
		`filter={foo:"odd"}`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("DELETE", `/foo?filter={foo:"odd"}`, nil)
			},
			ResponseCode:   http.StatusNoContent,
			ResponseBody:   ``,
			ResponseHeader: http.Header{"X-Total": []string{"3"}},
			ExtraTest:      checkFooIDs("2", "4"),
		},
		`filter={foo:"odd"},limit=2`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("DELETE", `/foo?filter={foo:"odd"}&limit=2`, nil)
			},
			ResponseCode:   http.StatusNoContent,
			ResponseBody:   ``,
			ResponseHeader: http.Header{"X-Total": []string{"2"}},
			ExtraTest:      checkFooIDs("2", "4", "5"),
		},
		`filter={foo:"odd"},limit=2,skip=1`: {
			Init: sharedInit,
			NewRequest: func() (*http.Request, error) {
				return http.NewRequest("DELETE", `/foo?filter={foo:"odd"}&limit=2&skip=1`, nil)
			},
			ResponseCode:   http.StatusNoContent,
			ResponseBody:   ``,
			ResponseHeader: http.Header{"X-Total": []string{"2"}},
			ExtraTest:      checkFooIDs("1", "2", "4"),
		},
	}

	for n, tc := range tests {
		tc := tc // capture range variable
		t.Run(n, tc.Test)
	}
}

/*
func TestHandlerDeleteList(t *testing.T) {
	s := mem.NewHandler()
	s.Insert(context.TODO(), []*resource.Item{
		{ID: "1", Payload: map[string]interface{}{}},
		{ID: "2", Payload: map[string]interface{}{}},
		{ID: "3", Payload: map[string]interface{}{}},
		{ID: "4", Payload: map[string]interface{}{}},
		{ID: "5", Payload: map[string]interface{}{}},
	})
	index := resource.NewIndex()
	test := index.Bind("test", schema.Schema{}, s, resource.DefaultConf)
	r, _ := http.NewRequest("DELETE", "/test", bytes.NewBufferString("{}"))
	rm := &RouteMatch{
		ResourcePath: []*ResourcePathComponent{
			&ResourcePathComponent{
				Name:     "test",
				Resource: test,
			},
		},
	}
	status, headers, body := listDelete(context.TODO(), r, rm)
	assert.Equal(t, http.StatusNoContent, status)
	assert.Equal(t, http.Header{"X-Total": []string{"5"}}, headers)
	assert.Nil(t, body)

	l, err := s.Find(context.TODO(), &query.Query{})
	assert.NoError(t, err)
	assert.Len(t, l.Items, 0)
}

func TestHandlerDeleteListFilter(t *testing.T) {
	s := mem.NewHandler()
	s.Insert(context.TODO(), []*resource.Item{
		{ID: "1", Payload: map[string]interface{}{"foo": "bar"}},
		{ID: "2", Payload: map[string]interface{}{"foo": "bar"}},
		{ID: "3", Payload: map[string]interface{}{"foo": "baz"}},
		{ID: "4", Payload: map[string]interface{}{"foo": "baz"}},
		{ID: "5", Payload: map[string]interface{}{}},
	})
	index := resource.NewIndex()
	test := index.Bind("test", schema.Schema{
		Fields: schema.Fields{"foo": {Filterable: true}},
	}, s, resource.DefaultConf)
	r, _ := http.NewRequest("DELETE", "/test", bytes.NewBufferString("{}"))
	rm := &RouteMatch{
		ResourcePath: []*ResourcePathComponent{
			&ResourcePathComponent{
				Name:     "test",
				Resource: test,
			},
		},
		Params: url.Values{
			"filter": []string{`{"foo": "bar"}`},
		},
	}
	status, headers, body := listDelete(context.TODO(), r, rm)
	assert.Equal(t, http.StatusNoContent, status)
	assert.Equal(t, http.Header{"X-Total": []string{"2"}}, headers)
	assert.Nil(t, body)

	l, err := s.Find(context.TODO(), &query.Query{})
	assert.NoError(t, err)
	if assert.Len(t, l.Items, 3) {
		assert.Equal(t, "3", l.Items[0].ID)
		assert.Equal(t, "4", l.Items[1].ID)
		assert.Equal(t, "5", l.Items[2].ID)
	}
}

func TestHandlerDeleteListInvalidFilter(t *testing.T) {
	index := resource.NewIndex()
	test := index.Bind("test", schema.Schema{}, nil, resource.DefaultConf)
	r, _ := http.NewRequest("DELETE", "/test", bytes.NewBufferString("{}"))
	rm := &RouteMatch{
		ResourcePath: []*ResourcePathComponent{
			&ResourcePathComponent{
				Name:     "test",
				Resource: test,
			},
		},
		Params: url.Values{
			"filter": []string{"invalid"},
		},
	}
	status, headers, body := listDelete(context.TODO(), r, rm)
	assert.Equal(t, 422, status)
	assert.Nil(t, headers)
	if assert.IsType(t, body, &Error{}) {
		err := body.(*Error)
		assert.Equal(t, 422, err.Code)
		assert.Equal(t, "Invalid `filter` parameter: char 0: expected '{' got 'i'", err.Message)
	}
}

func TestHandlerDeleteListNoStorage(t *testing.T) {
	index := resource.NewIndex()
	test := index.Bind("test", schema.Schema{}, nil, resource.DefaultConf)
	r, _ := http.NewRequest("DELETE", "/test", bytes.NewBufferString("{}"))
	rm := &RouteMatch{
		ResourcePath: []*ResourcePathComponent{
			&ResourcePathComponent{
				Name:     "test",
				Resource: test,
			},
		},
	}
	status, headers, body := listDelete(context.TODO(), r, rm)
	assert.Equal(t, http.StatusNotImplemented, status)
	assert.Nil(t, headers)
	if assert.IsType(t, body, &Error{}) {
		err := body.(*Error)
		assert.Equal(t, http.StatusNotImplemented, err.Code)
		assert.Equal(t, "No Storage Defined", err.Message)
	}
}
*/
