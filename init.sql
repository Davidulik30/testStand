create table gateway
(
    gtw_id           integer generated always as identity
        primary key,
    gtw_is_active    boolean default true not null,
    gtw_adapter_id   varchar(30)          not null,
    gtw_params_jsonb jsonb
);

create table channel
(
    chn_id           integer generated always as identity
        primary key,
    chn_name         varchar              not null,
    chn_is_active    boolean default true not null,
    gtw_id           integer              not null
        references gateway,
    chn_params_jsonb jsonb
);

INSERT INTO public.gateway (gtw_adapter_id, gtw_params_jsonb)
VALUES ('paylink', '{"transport": {"base_address": "https://payadap.co/api/v1/"}, "payment_methods": [{"id": "p2pcard", "name": "P2P Card", "required_fields": []}]}'),
 ('auris', '{"transport": {"base_address": "https://changer.icu/api2"}, "payment_methods": [{"id": "p2pcard", "gtw_id": {"KGS": 237, "KZT": 260, "RUB": 86}, "required_fields": []}, {"id": "p2psbp", "gtw_id": {"RUB": 88}, "required_fields": []}]}'),
 ('sequoia', '{"transport": {"base_address": "https://sequoiapay.com"}, "payment_methods": [{"id": "p2pcard", "name": "P2P Card", "required_fields": [{"id": "bank", "name": "Bank", "type": "select", "values": [{"id": "ru_sberbank", "name": "Sberbank"}]}]}]}'),
 ('asupayment', '{"transport": {"base_address": "http://185.230.143.210:8021/api/v1/"}, "payment_methods": [{"id": "p2pcard", "name": "P2P Card", "required_fields": [{"id": "bank", "name": "Bank", "type": "select", "values": [{"id": "ru_sberbank", "name": "Sberbank"}]}]}]}'),
 ('aplex', '{"transport": {"base_address": "http://5.129.242.10:8022/v1/"}, "required_fields": []}'),
 ('nestpay', '{"transport": {"base_address": "http://5.129.242.10:8023/"}, "required_fields": []}');



INSERT INTO public.channel (chn_name, gtw_id, chn_params_jsonb)
VALUES ('paylink', 1, '{"credentials": {"api_key": "m9bb1lelXNK3c198UWb2e41J1EQM", "merch_id": "52161-b615-4211-8638-3998bdb806c0"}, "payment_methods": ["p2pcard"]}'),
 ('auris', 2, '{"credentials": {"api_key": "049050051052-8ECG0WKvwSTWBzHTv2qtzy6MTiCkEoT2QAVAmNoGfQPzm02sPuBtz4nRwCJRs6fmM0WHUNqke6gG8o1MKcVHZGpJScZtOt8SmC3X1PGRx9uMTL3rw-OQP244B0M3HTID6T8PixLROqwaYQHJq6", "shop_id": 1102, "secret_key": "049050051052-jVDNvl8Lm8U4LQ7NPFGP2K03-k7w0o06k"}, "payment_methods": ["p2pcard"]}'),
 ('sequoia', 3, '{"credentials": {"secret_key": "8f5DSbXOolXwhLVmtOd", "callback_secret": "fqLFXw9BQLbW7G83Uf"}, "payment_methods": ["p2pcard"]}'),
 ('asupayment', 4, '{"credentials": {"api_key": "c2RmZHNhYnZmZGFiYWV0dnJ0c3JhZGZoYnN0cmRmdmJzZ2Z4MjM0NTQzZ2czcXZhZQ==",  "secret_key": "superSecretKey228","merchant_id":"32"}, "payment_methods": ["p2pcard"]}'),
 ('aplex', 5, '{"credentials": {"api_key": "6fO5obBS2qM6Q9IWwkhMQ2MD1emVB1aAtHwTAabSD0tBLZ9giDwRtXTJlaiKWh11==","gate_id": "657b6bafa93f477a573a66d0","password":"dev","email":"buyer@dev.alpex.app",  "secret_key": "FH1z1rjdTNQzWHGK7LvoqMLdN3UfZ7eF"}, "payment_methods": ["p2pcard"]}'),
 ('nestpay', 6, '{"credentials":{"clientId":"20001223","currency":"944","storeKey":"stwr3c4v4n9liiryucfhf25520","apiName":"toAdmin","apiPassword":"TEMP000130432"}, "payment_methods": ["p2pcard"]}');

-- auto-generated definition
create table "transaction"
(
    txn_id               bigint                                             not null
        primary key,
    txn_type_id          varchar                                            not null,
    pay_method_id        varchar                                            not null,
    chn_name             varchar,
    gtw_name             varchar,
    gtw_txn_id           varchar,
    txn_amount_src       bigint                                             not null,
    txn_currency_src     varchar                                            not null,
    txn_amount           bigint                                             not null,
    txn_currency         varchar                                            not null,
    txn_status_id        varchar                                            not null,
    txn_updated_at       timestamp with time zone default CURRENT_TIMESTAMP not null,
    txn_created_at       timestamp with time zone default CURRENT_TIMESTAMP not null
)
