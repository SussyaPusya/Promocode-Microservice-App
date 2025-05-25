create table if not exists promo
(
    id               uuid primary key not null ,
    company_id       varchar not null ,
    description      text,
    image_url        varchar,
    active_from      timestamptz ,
    active_until     timestamptz,
    created_at       timestamptz,
    mode             varchar check (mode in ('COMMON', 'UNIQUE')),
    target_age_from  int,
    target_age_until int,
    target_country   varchar,
    target_categories varchar[],
    check (target_age_from <= target_age_until),
    check (active_from <= active_until)
);

create table if not exists promo_code
(
    id           uuid primary key,
    promo_id     uuid references promo (id) on delete cascade,
    code         varchar,
    activations  int,
    max_count    int
);

