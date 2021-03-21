package datasources

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/hairyhenderson/gomplate/v3/internal/config"
	"github.com/hairyhenderson/gomplate/v3/vault"
	"github.com/pkg/errors"
)

type vaultRequester struct {
}

func (r *vaultRequester) Request(ctx context.Context, u *url.URL, header http.Header) (resp *Response, err error) {
	vc := config.VaultClientFromContext(ctx)
	if vc == nil {
		vc, err = vault.New(u)
		if err != nil {
			return nil, err
		}
	}

	err = vc.Login()
	if err != nil {
		return nil, err
	}

	p := u.Path
	if p == "" && u.Opaque != "" {
		p = u.Opaque
	}

	q := u.Query()
	params := make(map[string]interface{}, len(q))
	for k, v := range q {
		params[k] = strings.Join(v, " ")
	}

	resp = &Response{}
	hint := jsonMimetype

	var data []byte
	switch {
	case len(params) > 0:
		data, err = vc.Write(p, params)
	case strings.HasSuffix(p, "/"):
		hint = jsonArrayMimetype
		data, err = vc.List(p)
	default:
		data, err = vc.Read(p)
	}
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, errors.Errorf("no value found for path %s", p)
	}

	resp.ContentLength = int64(len(data))
	resp.ContentType, err = mimeType(u, hint)
	if err != nil {
		return nil, err
	}

	resp.Body = ioutil.NopCloser(bytes.NewReader(data))

	return resp, nil
}