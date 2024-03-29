package main

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type Product struct {
	Name string
	MinPrice float64
	Shipping string
}
var host = "http://localhost:8080"
var prodId = "1_7581887"

var procedures = []fetchFunction{
	fetchName,
	fetchShipping,
	fetchPrice,
}

func main() {
	p := Product{}

	start := time.Now()
	loadDataFromApi(&p)
	duration := time.Since(start)

	printResults(duration, p)
}

func loadDataFromApi(p *Product) {
	for _, f := range procedures {
		func() {
			checkErr(f(p))
		}()
	}
}

func printResults(d time.Duration, p Product) {
	fmt.Printf("\n\nfull duration: %s", d.String())
	fmt.Println("\nfetched product data:")
	fmt.Printf("\n%#v\n", p)
}

func checkErr(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

type fetchFunction func(p *Product) error

func fetchName(p *Product) error {
	url := host + "/products/" + prodId

	type productNameResponse struct {
		Name string
	}
	pnr := productNameResponse{}
	err := fillModel(url, &pnr)
	if err != nil {
		return err
	}
	p.Name = pnr.Name

	return nil
}

func fetchPrice(p *Product) error {
	url := host + "/products/" + prodId + "/price"

	type productPriceResponse struct {
		Retail struct {
			From struct {
				Value float64
			}
		}
	}
	ppr := productPriceResponse{}
	err := fillModel(url, &ppr)
	if err != nil {
		return err
	}
	p.MinPrice = ppr.Retail.From.Value

	return nil
}

func fetchShipping(p *Product) error {
	url := host + "/products/" + prodId + "/shippings"

	type productShippingResponse struct {
		Embedded struct {
			Items []struct {
				Date struct {
					From string
				}
			}
		} `json:"_embedded"`
	}
	psr := productShippingResponse{}
	err := fillModel(url, &psr)
	if err != nil {
		return err
	}
	if len(psr.Embedded.Items) == 0 {
		return errors.New("there were no shipping elements")
	}

	p.Shipping = psr.Embedded.Items[0].Date.From
	return nil
}

// fills
func fillModel(address string, responseModel interface{}) error {
	start := time.Now()

	response, err := http.Get(address)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return errors.New("request failed to " + address)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, &responseModel)
	if err != nil {
		return err
	}

	duration := time.Since(start)
	fmt.Printf("\n->    %s fetch time: %s\n", address, duration.String())

	return nil
}


