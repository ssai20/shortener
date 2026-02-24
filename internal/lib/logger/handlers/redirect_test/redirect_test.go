package redirect_test

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/assert/v2"
	"github.com/stretchr/testify/require"
	"net/http/httptest"
	"testing"
	"url-shortender/internal/lib/api"
	"url-shortender/internal/lib/logger/handlers/redirect"

	"url-shortender/internal/lib/logger/handlers/redirect/mocks"
	"url-shortender/internal/lib/logger/handlers/slogdiscard"
)

func TestSaveHandler(t *testing.T) {
	cases := []struct {
		name      string
		alias     string
		url       string
		respError string
		mockError error
	}{
		{
			name:  "Success",
			alias: "test_alias",
			url:   "https://yandex.ru",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			urlGetterMock := mocks.NewURLGetter(t)

			if tc.respError == "" || tc.mockError != nil {
				urlGetterMock.On("GetURL", tc.alias).
					Return(tc.url, tc.mockError).
					Once()
			}

			r := chi.NewRouter()
			r.Get("/{alias}", redirect.New(slogdiscard.NewDiscardLogger(), urlGetterMock))
			ts := httptest.NewServer(r)
			defer ts.Close()

			redirectedToURL, err := api.GetRedirect(ts.URL + "/" + tc.alias)
			require.NoError(t, err)
			assert.Equal(t, tc.url, redirectedToURL)

		})

	}

}
