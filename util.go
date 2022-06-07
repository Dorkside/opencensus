package opencensus

import (
	"go.opencensus.io/tag"
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"fmt"
	"io/ioutil"
)

type CCRequest struct {
	ProductId string
}

func appendIfMissing(slice []tag.Key, i tag.Key) []tag.Key {
	for _, ele := range slice {
		if ele == i {
			return slice
		}
	}
	return append(slice, i)
}

func GetTenant (r *http.Request) string {
	tenant := strings.Trim(strings.Split(strings.SplitAfter(r.URL.Path, "tenants/")[1], "/")[0], "")
	if len(tenant) > 0 {
		return tenant
	}
	return ""
}

func GetProduct (r *http.Request) string {
	b, err := ioutil.ReadAll(r.Body)
	r.Body = ioutil.NopCloser(bytes.NewBuffer(b))
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}
	
	var cr CCRequest
	err = json.Unmarshal(b, &cr)
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}
	if len(cr.ProductId) > 0 {
		return cr.ProductId
	}
	product := strings.Trim(strings.Split(strings.SplitAfter(r.URL.Path, "products/")[1], "/")[0], "")
	return product
}