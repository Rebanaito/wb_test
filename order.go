package main

import "github.com/jackc/pgx/v5"

// This function checks that all the fields are present (json message didn't have any missing fields),
// that all numeric values are valid and add up, and that string values are not empty (except for two
// fields that are empty in the provided model)
func validOrder(order OrderBlank) (bool, []error) {
	var errors []error
	if order.Order_uid == nil || order.Track_number == nil || order.Entry == nil || order.Locale == nil ||
		order.Customer_id == nil || order.Internal_signature == nil || order.Delivery_service == nil ||
		order.Shardkey == nil || order.Oof_shard == nil || order.Sm_id == nil || order.Delivery.Name == nil || order.Delivery.Phone == nil ||
		order.Delivery.Zip == nil || order.Delivery.City == nil || order.Delivery.Address == nil ||
		order.Delivery.Region == nil || order.Delivery.Email == nil || order.Payment.Transaction == nil ||
		order.Payment.Currency == nil || order.Payment.Provider == nil || order.Payment.Bank == nil ||
		order.Payment.Request_id == nil || order.Payment.Amount == nil || order.Payment.Payment_dt == nil ||
		order.Payment.Custom_fee == nil || order.Payment.Delivery_cost == nil || order.Payment.Goods_total == nil ||
		len(order.Items) == 0 {
		errors = append(errors, ErrMissingOrderData)
		return false, errors
	}
	if order.Date_created.IsZero() {
		errors = append(errors, ErrInvalidDate)
	}
	if *order.Order_uid == "" {
		errors = append(errors, ErrInvalidOrderUid)
	}
	if *order.Track_number == "" {
		errors = append(errors, ErrInvalidTrackNumber)
	}
	if *order.Entry == "" {
		errors = append(errors, ErrInvalidEntry)
	}
	if *order.Locale == "" {
		errors = append(errors, ErrInvalidLocale)
	}
	if *order.Customer_id == "" {
		errors = append(errors, ErrInvalidCustomerId)
	}
	if *order.Delivery_service == "" {
		errors = append(errors, ErrInvalidDeliveryService)
	}
	if *order.Shardkey == "" {
		errors = append(errors, ErrInvalidShardKey)
	}
	if *order.Sm_id < 0 {
		errors = append(errors, ErrInvalidSmId)
	}
	if *order.Oof_shard == "" {
		errors = append(errors, ErrInvalidOofShard)
	}
	if *order.Delivery.Name == "" {
		errors = append(errors, ErrInvalidDeliveryName)
	}
	if *order.Delivery.Phone == "" {
		errors = append(errors, ErrInvalidDeliveryPhone)
	}
	if *order.Delivery.Zip == "" {
		errors = append(errors, ErrInvalidDeliveryZip)
	}
	if *order.Delivery.City == "" {
		errors = append(errors, ErrInvalidDeliveryCity)
	}
	if *order.Delivery.Address == "" {
		errors = append(errors, ErrInvalidDeliveryAddress)
	}
	if *order.Delivery.Region == "" {
		errors = append(errors, ErrInvalidDeliveryRegion)
	}
	if *order.Delivery.Email == "" {
		errors = append(errors, ErrInvalidDeliveryEmail)
	}
	if *order.Payment.Transaction == "" || *order.Payment.Transaction != *order.Order_uid {
		errors = append(errors, ErrInvalidPaymentTransaction)
	}
	if *order.Payment.Currency == "" {
		errors = append(errors, ErrInvalidPaymentCurrency)
	}
	if *order.Payment.Provider == "" {
		errors = append(errors, ErrInvalidPaymentProvider)
	}
	if *order.Payment.Amount < 0 {
		errors = append(errors, ErrInvalidPaymentAmount)
	}
	if *order.Payment.Payment_dt < 0 {
		errors = append(errors, ErrInvalidPaymentDt)
	}
	if *order.Payment.Bank == "" {
		errors = append(errors, ErrInvalidPaymentBank)
	}
	if *order.Payment.Delivery_cost < 0 {
		errors = append(errors, ErrInvalidPaymentDeliveryCost)
	}
	if *order.Payment.Goods_total < 0 {
		errors = append(errors, ErrInvalidPaymentGoodsTotal)
	}
	if *order.Payment.Custom_fee < 0 {
		errors = append(errors, ErrInvalidPaymentCustomFee)
	}
	total := 0
	for _, item := range order.Items {
		if item.Chrt_id == nil || item.Track_number == nil || item.Price == nil || item.Rid == nil || item.Name == nil ||
			item.Sale == nil || item.Size == nil || item.Total_price == nil || item.Nm_id == nil || item.Brand == nil || item.Status == nil {
			errors = append(errors, ErrMissingItemData)
			return false, errors
		}
		if *item.Chrt_id < 0 {
			errors = append(errors, ErrInvalidItemChrtId)
		}
		if *item.Track_number == "" || *item.Track_number != *order.Track_number {
			errors = append(errors, ErrInvalidItemTrackNumber)
		}
		if *item.Price < 0 {
			errors = append(errors, ErrInvalidItemPrice)
		}
		if *item.Rid == "" {
			errors = append(errors, ErrInvalidItemRid)
		}
		if *item.Name == "" {
			errors = append(errors, ErrInvalidItemName)
		}
		if *item.Sale < 0 || *item.Sale > 100 {
			errors = append(errors, ErrInvalidItemSale)
		}
		if *item.Size == "" {
			errors = append(errors, ErrInvalidItemSize)
		}
		if *item.Total_price < 0 {
			errors = append(errors, ErrInvalidItemPrice)
		}
		if *item.Nm_id < 0 {
			errors = append(errors, ErrInvalidItemNmId)
		}
		if *item.Brand == "" {
			errors = append(errors, ErrInvalidItemBrand)
		}
		if *item.Status < 0 {
			errors = append(errors, ErrInvalidItemStatus)
		}
		if (*item.Price*(100-*item.Sale))/100 != *item.Total_price {
			errors = append(errors, ErrInvalidMath)
		}
		total += *item.Total_price
	}
	if total != *order.Payment.Goods_total {
		errors = append(errors, ErrInvalidItemTotal)
	}
	if total+*order.Payment.Delivery_cost+*order.Payment.Custom_fee != *order.Payment.Amount {
		errors = append(errors, ErrInvalidTotal)
	}
	return len(errors) == 0, errors
}

// This function simply transfers the information from the Blank struct we used for validation
// into the new Order struct to be used for storing
func createOrder(blank OrderBlank) Order {
	var new Order
	new.Customer_id = *blank.Customer_id
	new.Date_created = blank.Date_created
	new.Delivery_service = *blank.Delivery_service
	new.Entry = *blank.Entry
	new.Internal_signature = *blank.Internal_signature
	new.Locale = *blank.Locale
	new.Oof_shard = *blank.Oof_shard
	new.Order_uid = *blank.Order_uid
	new.Shardkey = *blank.Shardkey
	new.Sm_id = *blank.Sm_id
	new.Track_number = *blank.Track_number
	new.Delivery.Order_uid = *blank.Order_uid
	new.Delivery.Address = *blank.Delivery.Address
	new.Delivery.City = *blank.Delivery.City
	new.Delivery.Email = *blank.Delivery.Email
	new.Delivery.Name = *blank.Delivery.Name
	new.Delivery.Phone = *blank.Delivery.Phone
	new.Delivery.Region = *blank.Delivery.Region
	new.Delivery.Zip = *blank.Delivery.Zip
	new.Payment.Amount = *blank.Payment.Amount
	new.Payment.Bank = *blank.Payment.Bank
	new.Payment.Currency = *blank.Payment.Currency
	new.Payment.Custom_fee = *blank.Payment.Custom_fee
	new.Payment.Delivery_cost = *blank.Payment.Delivery_cost
	new.Payment.Goods_total = *blank.Payment.Goods_total
	new.Payment.Payment_dt = *blank.Payment.Payment_dt
	new.Payment.Provider = *blank.Payment.Provider
	new.Payment.Request_id = *blank.Payment.Request_id
	new.Payment.Transaction = *blank.Payment.Transaction
	new.Items = make([]Item, len(blank.Items))
	for i := range blank.Items {
		new.Items[i].Order_uid = *blank.Order_uid
		new.Items[i].Brand = *blank.Items[i].Brand
		new.Items[i].Chrt_id = *blank.Items[i].Chrt_id
		new.Items[i].Name = *blank.Items[i].Name
		new.Items[i].Nm_id = *blank.Items[i].Nm_id
		new.Items[i].Price = *blank.Items[i].Price
		new.Items[i].Rid = *blank.Items[i].Rid
		new.Items[i].Sale = *blank.Items[i].Sale
		new.Items[i].Size = *blank.Items[i].Size
		new.Items[i].Status = *blank.Items[i].Status
		new.Items[i].Total_price = *blank.Items[i].Total_price
		new.Items[i].Track_number = *blank.Items[i].Track_number
	}
	return new
}

// This function uses the Sanitize method from the pgx package on all string values
func sanitizeOrder(order Order) {
	order.Customer_id = pgx.Identifier{order.Customer_id}.Sanitize()
	order.Delivery_service = pgx.Identifier{order.Delivery_service}.Sanitize()
	order.Entry = pgx.Identifier{order.Entry}.Sanitize()
	order.Internal_signature = pgx.Identifier{order.Internal_signature}.Sanitize()
	order.Locale = pgx.Identifier{order.Locale}.Sanitize()
	order.Oof_shard = pgx.Identifier{order.Oof_shard}.Sanitize()
	order.Order_uid = pgx.Identifier{order.Order_uid}.Sanitize()
	order.Shardkey = pgx.Identifier{order.Shardkey}.Sanitize()
	order.Track_number = pgx.Identifier{order.Track_number}.Sanitize()
	order.Delivery.Address = pgx.Identifier{order.Delivery.Address}.Sanitize()
	order.Delivery.City = pgx.Identifier{order.Delivery.City}.Sanitize()
	order.Delivery.Email = pgx.Identifier{order.Delivery.Email}.Sanitize()
	order.Delivery.Name = pgx.Identifier{order.Delivery.Name}.Sanitize()
	order.Delivery.Phone = pgx.Identifier{order.Delivery.Phone}.Sanitize()
	order.Delivery.Region = pgx.Identifier{order.Delivery.Region}.Sanitize()
	order.Delivery.Zip = pgx.Identifier{order.Delivery.Zip}.Sanitize()
	order.Payment.Bank = pgx.Identifier{order.Payment.Bank}.Sanitize()
	order.Payment.Currency = pgx.Identifier{order.Payment.Currency}.Sanitize()
	order.Payment.Provider = pgx.Identifier{order.Payment.Provider}.Sanitize()
	order.Payment.Request_id = pgx.Identifier{order.Payment.Request_id}.Sanitize()
	order.Payment.Transaction = pgx.Identifier{order.Payment.Transaction}.Sanitize()
	for _, item := range order.Items {
		item.Brand = pgx.Identifier{item.Brand}.Sanitize()
		item.Name = pgx.Identifier{item.Name}.Sanitize()
		item.Rid = pgx.Identifier{item.Rid}.Sanitize()
		item.Size = pgx.Identifier{item.Size}.Sanitize()
		item.Track_number = pgx.Identifier{item.Track_number}.Sanitize()
	}
}
