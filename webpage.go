package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"sync"
)

// Handler function for HTTP requests. All incoming requests are directed to the same search page.
// Once the user submits an order_uid to search for, the searchResults function is executed
func handler(rw http.ResponseWriter, r *http.Request, cache map[string]Order, mu *sync.Mutex, tmpl *template.Template) {
	switch r.URL.Path {
	case "/":
		tmpl.Execute(rw, nil)
	case "/search":
		searchResults(rw, r, cache, mu)
	}
}

// This function takes the order_uid from the HTTP request and tries to find matching order in the cache map.
// If successful, it returns the same page but now with tables for data
func searchResults(rw http.ResponseWriter, r *http.Request, cache map[string]Order, mu *sync.Mutex) {
	search := r.FormValue("order_uid")
	mu.Lock()
	value, ok := cache[search]
	mu.Unlock()
	if ok {
		rw.WriteHeader(http.StatusOK)
		timePretty := datePretty(value.Date_created)
		data := fmt.Sprintf(`<html>
		<style>
			table, th, td {
				border: 1px solid black;
				text-align: center;
				padding: 5px;
			}
		</style>
		<form action="/search">
			<label for="order_uid">Order ID: </label>
			<input type="text" id="order_uid" name="order_uid" value="">
			<input type="submit" value="Search"><br>
		</form>
		<h3>Order ID: %s</h3><br>
							<h4>Order details:</h4>
							<table>
								<tr>
									<th>track number</th>
									<th>entry</th>
									<th>locale</th>
									<th>internal signature</th>
									<th>customer id</th>
									<th>delivery service</th>
									<th>shard key</th>
									<th>sm id</th>
									<th>date created</th>
									<th>oof shard</th>
								</tr>
								<tr>
									<td>%s</td>
									<td>%s</td>
									<td>%s</td>
									<td>%s</td>
									<td>%s</td>
									<td>%s</td>
									<td>%s</td>
									<td>%d</td>
									<td>%s</td>
									<td>%s</td>
								</tr>
							</table>
							<h4>Delivery details:</h4>
							<table>
								<tr>
									<th>name</th>
									<th>phone</th>
									<th>zip</th>
									<th>city</th>
									<th>address</th>
									<th>region</th>
									<th>e-mail</th>
								</tr>
								<tr>
									<td>%s</td>
									<td>%s</td>
									<td>%s</td>
									<td>%s</td>
									<td>%s</td>
									<td>%s</td>
									<td>%s</td>
								</tr>
							</table>
							<h4>Payment details:</h4>
							<table>
								<tr>
									<th>transaction</th>
									<th>request id</th>
									<th>currency</th>
									<th>provider</th>
									<th>amount</th>
									<th>payment dt</th>
									<th>bank</th>
									<th>delivery cost</th>
									<th>goods total</th>
									<th>custom fee</th>
								</tr>
								<tr>
									<td>%s</td>
									<td>%s</td>
									<td>%s</td>
									<td>%s</td>
									<td>%d</td>
									<td>%d</td>
									<td>%s</td>
									<td>%d</td>
									<td>%d</td>
									<td>%d</td>
								</tr>
							</table>
							<h4>Items:</h4>
							<table><tr>
						<th>chrt id</th>
						<th>track number</th>
						<th>price</th>
						<th>rid</th>
						<th>name</th>
						<th>sale</th>
						<th>size</th>
						<th>total price</th>
						<th>nm id</th>
						<th>brand</th>
						<th>status</th>
					</tr>`, value.Order_uid, value.Track_number, value.Entry, value.Locale,
			value.Internal_signature, value.Customer_id, value.Delivery_service,
			value.Shardkey, value.Sm_id, timePretty, value.Oof_shard, value.Delivery.Name,
			value.Delivery.Phone, value.Delivery.Zip, value.Delivery.City, value.Delivery.Address,
			value.Delivery.Region, value.Delivery.Email, value.Payment.Transaction, value.Payment.Request_id,
			value.Payment.Currency, value.Payment.Provider, value.Payment.Amount, value.Payment.Payment_dt,
			value.Payment.Bank, value.Payment.Delivery_cost, value.Payment.Goods_total, value.Payment.Custom_fee)
		for _, item := range value.Items {
			new := fmt.Sprintf(`<tr>
						<td>%d</td>
						<td>%s</td>
						<td>%d</td>
						<td>%s</td>
						<td>%s</td>
						<td>%d</td>
						<td>%s</td>
						<td>%d</td>
						<td>%d</td>
						<td>%s</td>
						<td>%d</td>
					</tr>`, item.Chrt_id, item.Track_number, item.Price, item.Rid, item.Name,
				item.Sale, item.Size, item.Total_price, item.Nm_id, item.Brand, item.Status)
			data = string(append([]byte(data), []byte(new)...))
		}
		data = string(append([]byte(data), []byte("</table></html>")...))
		tmpl, _ := template.New("searchResults").Parse(data)
		err := tmpl.Execute(rw, nil)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error executing a template", err)
			return
		}
	} else {
		rw.WriteHeader(http.StatusNotFound)
		data := `<html>
		<style>
			table, th, td {
				border: 1px solid black;
			}
		</style>
		<form action="/search">
			<label for="order_uid">Order ID: </label>
			<input type="text" id="order_uid" name="order_uid" value="">
			<input type="submit" value="Search"><br>
		</form>
		<h3>Couldn't find an order with this ID.</h3></html>`
		tmpl, _ := template.New("searchResults").Parse(data)
		err := tmpl.Execute(rw, nil)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error executing a template", err)
			return
		}
	}
}
