package opencensus

import (
	"go.opencensus.io/tag"
	"net/http"
	"bytes"
	"strings"
	"fmt"
	"encoding/json"
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

func getTenant (r *http.Request) string {
	tenant := strings.Trim(strings.Split(strings.SplitAfter(r.URL.Path, "tenants/")[1], "/")[0], "")
	if len(tenant) > 0 {
		return tenant
	}
	return ""
}

func getProduct (r *http.Request) string {
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
	fmt.Println("[Getting Product Id]: "+cr.ProductId)
	if len(cr.ProductId) > 0 {
		return cr.ProductId
	}
	fmt.Println("Product Not Found In Body, checking path...")
	product := strings.Trim(strings.Split(strings.SplitAfter(r.URL.Path, "products/")[1], "/")[0], "")
	fmt.Println("Product from path: "+product)
	return product
}