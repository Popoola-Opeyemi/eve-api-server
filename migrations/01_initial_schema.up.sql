CREATE TABLE "site" (
  "id" varchar(25) PRIMARY KEY,
  "subdomain" varchar(25) UNIQUE,
  "name" varchar(200),
  "status" integer NOT NULL,
  "date_registered" timestamp,
  "attr" jsonb NOT NULL DEFAULT '{}',
  "platform" bool not null default false
);

CREATE TABLE "user" (
  "id" varchar(25) PRIMARY KEY,
  "site_id" varchar(25),
  "status" integer NOT NULL DEFAULT 0,
  "first_name" varchar(150) NOT NULL,
  "last_name" varchar(150) NOT NULL,
  "email" varchar(250) NOT NULL,
  "password" varchar(150) NOT NULL,
  "phone" varchar(150) NOT NULL,
  "attr" jsonb NOT NULL DEFAULT '{}',
  "role" integer NOT NULL DEFAULT 0,
  "type" integer NOT NULL,
  "is_site_user" boolean NOT NULL DEFAULT false,

  constraint uq_user_site_id_email unique ("site_id", "email")
);

CREATE TABLE "user_type" (
  "id" integer PRIMARY KEY,
  "label" varchar(25) UNIQUE NOT NULL
);

CREATE TABLE "street" (
  "id" varchar(25) PRIMARY KEY,
  "site_id" varchar(25),
  "name" varchar(200) NOT NULL
);

CREATE TABLE "unit_type" (
  "id" integer PRIMARY KEY,
  "label" varchar(25) UNIQUE NOT NULL
);

CREATE TABLE "unit" (
  "id" varchar(25) PRIMARY KEY,
  "site_id" varchar(25),
  "type" integer,
  "street_id" varchar(25),
  "label" varchar(100) NOT NULL,
  "attr" jsonb NOT NULL DEFAULT '{"unit_number":""}'
);

create table "residency" (
  "id" varchar(25) PRIMARY KEY,
  "site_id" varchar(25) references site(id),
  "unit_id" varchar(25) references unit(id),
  "previous_unit_id" varchar(25),
  "active_status" int not null default 1,
  "date_start" date NOT NULL DEFAULT current_date,
  "date_exit" date
);

create table "resident" (
  "id" varchar(25) PRIMARY KEY,
  "can_login" bool NOT NULL,
  "first_name" varchar(150) NOT NULL,
  "last_name" varchar(150) NOT NULL,
  "email" varchar(250) UNIQUE NOT NULL,
  "password" varchar(150) NOT NULL,
  "phone" varchar(150) NOT NULL,
  "attr" jsonb NOT NULL DEFAULT '{}',
  "type" integer NOT NULL DEFAULT 0,
  "status" int NOT NULL DEFAULT 1,
  "primary_id" varchar(25),
  "residency_id" varchar(25) not null references residency(id) ON DELETE CASCADE
);

CREATE TABLE "due" (
  "id" varchar(25) PRIMARY KEY,
  "site_id" varchar(25),
  "date_created" timestamp NOT NULL DEFAULT localtimestamp,
  "name" varchar(250) NOT NULL,
  "description" text NOT NULL,
  "amount" numeric(15,2) NOT NULL default 0.00,
  "status" int NOT NULL DEFAULT 0,
  "attr" jsonb NOT NULL DEFAULT '{}'
);

CREATE UNIQUE INDEX ix_due_name on "due" ("site_id", "name");

CREATE TABLE "bill" (
  "id" varchar(25) PRIMARY KEY,
  "site_id" varchar(25),
  "unit_type" integer,
  "date_created" timestamp NOT NULL DEFAULT localtimestamp,
  "name" varchar(250) NOT NULL,
  "note" text NOT NULL,
  "total" numeric(15,2) NOT NULL,
  "attr" jsonb NOT NULL DEFAULT '{}'
);

-- note that in code id == due_id
CREATE TABLE "bill_item" (
  "id" varchar(25),
  "site_id" varchar(25) REFERENCES "site"("id"),
  "bill_id" varchar(25) REFERENCES "bill"("id") ON DELETE CASCADE,
  "due_id" varchar(25)  REFERENCES "due"("id"),
  "amount" numeric(15,2) NOT NULL default 0.00,
  PRIMARY KEY ("site_id", "id", "bill_id")
);

CREATE UNIQUE INDEX uq_bill_item_bill_due on "bill_item"("site_id", "bill_id", "due_id");

CREATE TABLE "transaction" (
  "id" varchar(25) PRIMARY KEY,
  "site_id" varchar(25),
  "resident_id" varchar(25) NOT NULL DEFAULT '',
  "type" integer NOT NULL DEFAULT 0,
  "date_trx" timestamp NOT NULL,
  "bill_id" varchar(25) REFERENCES "bill"("id"),
  "due_id" varchar(25),
  "description" text NOT NULL DEFAULT '',
  "amount" numeric(15,2) NOT NULL,
  "attr" jsonb NOT NULL DEFAULT '{}'
);

CREATE TABLE "notice_board" (
  "id" varchar(25) PRIMARY KEY,
  "site_id" varchar(25),
  "date_created" timestamp NOT NULL DEFAULT localtimestamp,
  "date_expiry" timestamp NOT NULL DEFAULT localtimestamp,
  "title" varchar(200) NOT NULL,
  "message" text NOT NULL,
  "attr" jsonb NOT NULL DEFAULT '{}'
);

CREATE TABLE "gate_pass" (
  "id" varchar(25) PRIMARY KEY,
  "site_id" varchar(25),
  "resident_id" varchar(25),
  "date_created" timestamp NOT NULL,
  "token" varchar(25) NOT NULL,
  -- 1: in, 2:out, 3:in&out
  "type" int NOT NULL DEFAULT 0,
  -- 1: in, 2:out, 3:in&out
  "status" int NOT NULL DEFAULT 0,
  "attr" jsonb NOT NULL DEFAULT '{}'
);

CREATE TABLE "visitor" (
  "id" varchar(25) PRIMARY KEY,
  "site_id" varchar(25),
  "gatepass_id" varchar(25),
  "date_created" timestamp NOT NULL,
  -- 1: registered by resident, 2: registered_by security
  "registration_type" int NOT NULL DEFAULT 0,
  -- 1: in, 2:out
  "status" int NOT NULL DEFAULT 0,
  "resident_id" varchar(25),
  "security_id" varchar(25),

  "attr" jsonb NOT NULL DEFAULT '{}'
);

CREATE VIEW "association_view" as
select
  s.*,
  u.id as admin_id,
  u.first_name as admin_first_name,
  u.last_name as admin_last_name,
  u.email as admin_email,
  u.phone as admin_phone,
  u.status as admin_status,
  (select count(*) from user where site_id=s.id and type = 7) as support_active,
  (select count(*) from resident_account_status where site_id=s.id and balance > 0) as stats_paid_residents,
  (select count(*) from resident_account_status where site_id=s.id and balance < 0) as stats_indebt_residents,
  (select count(*) from resident_family_list where type = 2 and site_id=s.id) as stats_secondary_residents,
  (select count(*) from resident_family_list where type = 1 and site_id=s.id) as stats_primary_residents,
  (select count(*) from resident_family_list as rf left join residency as rs on rs.id = rf.residency_id where rs.site_id = s.id and rf.type = 1 and rs.unit_id is not null) as units_occupied,
  (select count(*) from resident_family_list where unit_id is not null and s.id = resident_family_list.site_id ) as stats_active_residents,
  (select count(*) from resident_family_list where site_id=s.id) as stats_residents,
  (select count(*) from "user" where site_id=s.id) as stats_users,
  (select count(*) from "unit" where site_id=s.id) as stats_units,
  (select count(*) from "street" where site_id=s.id) as stats_streets
 from site as s
 left join "user" as u on u.site_id = s.id and u.is_site_user = true
 
 where platform = false;

CREATE VIEW "association_list" as
select
  "id",
  "subdomain",
  "name",
  "status",
  "date_registered"
 from site
 where "platform" = false;

CREATE VIEW "resident_view" as
select
  rs.id as residency_id,
  rs.site_id,
  rs.unit_id,
  rs.date_start,
  rs.date_exit,
  rs.active_status,
  r.id,
  r.can_login,
  (r.first_name || ' ' || r.last_name) as "name",
  r.first_name,
  r.last_name,
  r.email,
  r.phone,
  r.attr,
  r.type,
  r.status,
  ut.label as unit_type,

  concat(
    s.name, ', ',
    (case when u.attr->>'unit_number' is not null then u.attr->>'unit_number'||', ' else '' end), u.label || ''
  ) as "unit",

  concat(
    (case when u.attr->>'unit_number' is not null then u.attr->>'unit_number'||', ' else '' end)
    , s.name
  ) as "unit_name",

  s.name as "street_name",
  a.name as "association",
  a.subdomain
from "resident" as r
join "residency" as rs on rs.id = r.residency_id
left join "site" as a on a.id = rs.site_id
left join "unit" as u on u.id = rs.unit_id
left join "unit_type" as ut on ut.id = u.type
left join "street" as s on s.id = u.street_id
where rs.unit_id != '' and rs.date_exit is null and rs.active_status = 1;

CREATE VIEW "unit_street_view" as
select
  u.id,
  u.site_id,
  concat(
    (case when u.attr->>'unit_number' is not null then u.attr->>'unit_number'||', ' else '' end)
    , s.name, ', '||u.label
  ) as label
from "unit" as u
left join "street" as s on s.id = u.street_id;


CREATE VIEW "bill_list" AS
with "items" as (
  select
    i.site_id, i.bill_id, sum(i.amount) as total
  from
    "bill_item" as i
  group by
    i.site_id, i.bill_id
)
select
  b.id, b.site_id, b.unit_type, b.date_created, b.name, b.note,
  ut.label as unit_type_name,
  (case when i.total is null then 0.00 else i.total end) as total
from "bill" as b
left join "unit_type" as ut on ut.id = b.unit_type
left join items as i on i.bill_id = b.id and i.site_id = b.site_id;


CREATE VIEW "bill_item_list" as
select
  i.id, i.site_id, i.bill_id, i.due_id,
  i.amount
from "bill_item" as i;


CREATE VIEW "unit_list" as
select
  u.id, u.site_id, u.type, u.street_id, u.attr,
  case
    when u.attr->>'unit_number' is not null then concat(u.attr->>'unit_number', ', ', u.label)
    else u.label
  end as label,
  s.name as "street",
  ut.label as "unit_type",
  (case 
  when rs.active_status = 0 or rs.active_status is null then 'vacant' 
  when rs.active_status = 1 then 'occupied' end) as occupied_status,
  cast(regexp_replace(attr->>'unit_number', '[^0-9]', '') as int) as olbl

from "unit" as u
join "street" as s
  on s.id = u.street_id
join "unit_type" as ut
  on ut.id = u.type
left join "residency" as rs
  on u.id = rs.unit_id
order by
  olbl, s.name
;

CREATE VIEW "available_units_list" as
select
  u.id,
  u.site_id,
  concat(
    (case when u.attr->>'unit_number' is not null then u.attr->>'unit_number'||', ' else '' end),
    u.label , ', '||s.name
  ) as label,
  s.name as "street",
  cast(regexp_replace(u.attr->>'unit_number', '[^0-9]', '') as int) as olbl

from "unit" as u
left join "street" as s on s.id = u.street_id
left join "residency" as rs on rs.unit_id = u.id and rs.site_id = u.site_id
left join "resident" as r on r.residency_id = rs.id
where
  r.id is null and rs.unit_id is null;


CREATE VIEW "gate_pass_list" as
select
  g.*,
  g.attr->>'visitor' as "visitor",
  g.attr->>'plate_number' as "plate_number",
  concat(r.first_name, ' ', r.last_name) as "resident",
  r.type as resident_type,
  r.residency_id
from
  gate_pass as g
left join "resident" as r on r.id = g.resident_id
;

CREATE VIEW "active_notice" as
select
  *
from
  notice_board
where
  cast(localtimestamp as date) <= (date_expiry::date)
;

CREATE VIEW "expired_notice" as
select
  *
from
  notice_board
where
  cast(localtimestamp as date) > (date_expiry::date)
;


ALTER TABLE "user" ADD FOREIGN KEY ("site_id") REFERENCES "site" ("id") ON DELETE CASCADE;

ALTER TABLE "user" ADD FOREIGN KEY ("type") REFERENCES "user_type" ("id");

ALTER TABLE "street" ADD FOREIGN KEY ("site_id") REFERENCES "site" ("id");

ALTER TABLE "unit" ADD FOREIGN KEY ("site_id") REFERENCES "site" ("id");

ALTER TABLE "unit" ADD FOREIGN KEY ("type") REFERENCES "unit_type" ("id");

ALTER TABLE "unit" ADD FOREIGN KEY ("street_id") REFERENCES "street" ("id");

ALTER TABLE "due" ADD FOREIGN KEY ("site_id") REFERENCES "site" ("id");

ALTER TABLE "bill" ADD FOREIGN KEY ("site_id") REFERENCES "site" ("id");

ALTER TABLE "bill" ADD FOREIGN KEY ("unit_type") REFERENCES "unit_type" ("id");

ALTER TABLE "notice_board" ADD FOREIGN KEY ("site_id") REFERENCES "site" ("id");

ALTER TABLE "gate_pass" ADD FOREIGN KEY ("site_id") REFERENCES "site" ("id");

ALTER TABLE "visitor" ADD FOREIGN KEY ("site_id") REFERENCES "site" ("id");

ALTER TABLE "visitor" ADD FOREIGN KEY ("gatepass_id") REFERENCES "gate_pass" ("id");
