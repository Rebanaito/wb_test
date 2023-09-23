CREATE TABLE orders (
	order_uid VARCHAR(200) NOT NULL PRIMARY KEY,
  	track_number VARCHAR(200) NOT NULL,
  	entry VARCHAR(200) NOT NULL,
	locale VARCHAR(200) NOT NULL,
	internal_signature VARCHAR(200) NOT NULL,
	customer_id VARCHAR(200) NOT NULL,
	delivery_service VARCHAR(200) NOT NULL,
	shardkey VARCHAR(200) NOT NULL,
	sm_id INT NOT NULL,
	date_created TIMESTAMP NOT NULL,
	oof_shard VARCHAR(200)
);

CREATE TABLE deliveries (
	order_uid VARCHAR(200) REFERENCES orders (order_uid),
	name VARCHAR(200) NOT NULL,
    phone VARCHAR(200) NOT NULL,
    zip VARCHAR(200) NOT NULL,
    city VARCHAR(200) NOT NULL,
    address VARCHAR(200) NOT NULL,
    region VARCHAR(200) NOT NULL,
    email VARCHAR(200) NOT NULL,
	UNIQUE(order_uid)
);

CREATE TABLE payments (
	transaction VARCHAR(200) REFERENCES orders (order_uid),
    request_id VARCHAR(200) NOT NULL,
    currency VARCHAR(200) NOT NULL,
    provider VARCHAR(200) NOT NULL,
    amount INT NOT NULL,
    payment_dt INT NOT NULL,
    bank VARCHAR(200) NOT NULL,
    delivery_cost INT NOT NULL,
    goods_total INT NOT NULL,
    custom_fee INT NOT NULL,
	UNIQUE(transaction)
);

CREATE TABLE items (
	order_uid VARCHAR(200) REFERENCES orders (order_uid),
	chrt_id INT NOT NULL,
    track_number VARCHAR(200) NOT NULL,
    price INT NOT NULL,
    rid VARCHAR(200) NOT NULL,
    name VARCHAR(200) NOT NULL,
    sale INT NOT NULL,
    size VARCHAR(200) NOT NULL,
    total_price INT NOT NULL,
    nm_id INT NOT NULL,
    brand VARCHAR(200) NOT NULL,
    status INT NOT NULL
);