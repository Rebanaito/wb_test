package main

import (
	"os"
	"testing"
)

func testNumError(t *testing.T, want error, file string) {
	data, _ := os.ReadFile(file)
	orderBlank, _ := jsonParser(data)
	valid, err := validOrder(orderBlank)
	if valid {
		t.Fatalf("Order should not be valid")
	} else {
		ok := false
		for i := range err {
			if err[i] == want {
				ok = true
				break
			}
		}
		if !ok {
			t.Fatalf("Should return '%s'", want)
		}
	}
}

func TestMultipleErrors(t *testing.T) {
	testNumError(t, ErrInvalidOrderUid, "tests/invalid_order_uid.txt")
	testNumError(t, ErrInvalidTrackNumber, "tests/invalid_track_number.txt")
	testNumError(t, ErrInvalidPaymentAmount, "tests/invalid_payment_amount.txt")
	testNumError(t, ErrInvalidPaymentDeliveryCost, "tests/invalid_payment_delivery_cost.txt")
	testNumError(t, ErrInvalidPaymentGoodsTotal, "tests/invalid_payment_goods_total.txt")
	testNumError(t, ErrInvalidItemPrice, "tests/invalid_item_price.txt")
	testNumError(t, ErrInvalidItemSale, "tests/invalid_item_sale.txt")
	testNumError(t, ErrInvalidItemTotal, "tests/invalid_item_total.txt")
	testNumError(t, ErrInvalidTotal, "tests/invalid_total.txt")
	testNumError(t, ErrInvalidPaymentCustomFee, "tests/invalid_payment_custom_fee.txt")
}

func testError(t *testing.T, want error, file string) {
	data, _ := os.ReadFile(file)
	orderBlank, _ := jsonParser(data)
	valid, err := validOrder(orderBlank)
	if valid {
		t.Fatalf("Order should not be valid")
	} else if len(err) != 1 {
		t.Fatalf("Should return 1 %s", want)
	} else if err[0] != want {
		t.Fatalf("Should return 1 '%s', instead have: '%s'", want, err[0].Error())
	}
}

func TestSingleErrors(t *testing.T) {
	testError(t, ErrMissingOrderData, "tests/missing_order_data.txt")
	testError(t, ErrMissingItemData, "tests/missing_item_data.txt")
	testError(t, ErrInvalidEntry, "tests/invalid_entry.txt")
	testError(t, ErrInvalidLocale, "tests/invalid_locale.txt")
	testError(t, ErrInvalidCustomerId, "tests/invalid_customer_id.txt")
	testError(t, ErrInvalidDeliveryService, "tests/invalid_delivery_service.txt")
	testError(t, ErrInvalidShardKey, "tests/invalid_shardkey.txt")
	testError(t, ErrInvalidSmId, "tests/invalid_sm_id.txt")
	testError(t, ErrInvalidDate, "tests/invalid_date.txt")
	testError(t, ErrInvalidOofShard, "tests/invalid_oof_shard.txt")
	testError(t, ErrInvalidDeliveryName, "tests/invalid_delivery_name.txt")
	testError(t, ErrInvalidDeliveryPhone, "tests/invalid_delivery_phone.txt")
	testError(t, ErrInvalidDeliveryZip, "tests/invalid_delivery_zip.txt")
	testError(t, ErrInvalidDeliveryCity, "tests/invalid_delivery_city.txt")
	testError(t, ErrInvalidDeliveryAddress, "tests/invalid_delivery_address.txt")
	testError(t, ErrInvalidDeliveryRegion, "tests/invalid_delivery_region.txt")
	testError(t, ErrInvalidDeliveryEmail, "tests/invalid_delivery_email.txt")
	testError(t, ErrInvalidPaymentTransaction, "tests/invalid_payment_transaction.txt")
	testError(t, ErrInvalidPaymentCurrency, "tests/invalid_payment_currency.txt")
	testError(t, ErrInvalidPaymentProvider, "tests/invalid_payment_provider.txt")
	testError(t, ErrInvalidPaymentDt, "tests/invalid_payment_dt.txt")
	testError(t, ErrInvalidPaymentBank, "tests/invalid_payment_bank.txt")
	testError(t, ErrInvalidItemChrtId, "tests/invalid_item_chrt_id.txt")
	testError(t, ErrInvalidItemTrackNumber, "tests/invalid_item_track_number.txt")
	testError(t, ErrInvalidItemRid, "tests/invalid_item_rid.txt")
	testError(t, ErrInvalidItemName, "tests/invalid_item_name.txt")
	testError(t, ErrInvalidItemSize, "tests/invalid_item_size.txt")
	testError(t, ErrInvalidItemNmId, "tests/invalid_item_nm_id.txt")
	testError(t, ErrInvalidItemBrand, "tests/invalid_item_brand.txt")
	testError(t, ErrInvalidItemStatus, "tests/invalid_item_status.txt")
	testError(t, ErrInvalidMath, "tests/invalid_math.txt")
}
