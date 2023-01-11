-- deprecated
CREATE TABLE "invoice_master" (
  "id" varchar(25) PRIMARY KEY,
  "site_id" varchar(25),
  "month" integer not null,
  "year" integer not null,
  -- expected to contain an array of {bill_name, bill_id, unit_type_label, unit_type}
  "bills" jsonb NOT NULL DEFAULT '[]',
  "date_created" timestamp NOT NULL DEFAULT localtimestamp,
  "description" text NOT NULL DEFAULT '',
  "attr" jsonb NOT NULL DEFAULT '{}'
);

CREATE TABLE "invoice" (
  "id" varchar(25) PRIMARY KEY,
  "site_id" varchar(25),
  "invoice_number" bigserial not null,
  "resident_id" varchar(25) NOT NULL,
  "first_name" varchar(150) NOT NULL,
  "last_name" varchar(150) NOT NULL,
  "address" text NOT NULL DEFAULT '',
  "month" integer not null,
  "year" integer not null,
  "date_created" timestamp NOT NULL DEFAULT localtimestamp,
  "bill_id" varchar(25) NOT NULL REFERENCES "bill"("id"),
  "unit_type" integer NOT NULL,
  "description" text NOT NULL DEFAULT '',
  "amount" numeric(15,2) NOT NULL,
  -- array of {due_id, amount, name}
  "dues" jsonb NOT NULL DEFAULT '[]'
);

DROP TABLE IF EXISTS "transaction";
CREATE TABLE "transaction" (
  "id" varchar(25) PRIMARY KEY,
  "site_id" varchar(25) NOT NULL REFERENCES "site"("id"),
  "resident_id" varchar(25) NOT NULL REFERENCES "resident"("id"),
  -- 1: credit(payment) , 2: debit(invoice/bill)
  "type" integer NOT NULL DEFAULT 0,
  "date_trx" timestamp NOT NULL,
  "date_created" timestamp NOT NULL DEFAULT localtimestamp,
  "invoice_id" varchar(25),
  "payment_id" varchar(25),
  "due_id" varchar(25),
  "amount" numeric(15,2) NOT NULL
);

DROP TABLE IF EXISTS "visitor";
CREATE TABLE "visitor" (
  "id" varchar(25) PRIMARY KEY,
  "site_id" varchar(25),
  "date_created" timestamp NOT NULL DEFAULT localtimestamp,
  "date_arrival" date NOT NULL DEFAULT current_date,
  "name" varchar(150) NOT NULL DEFAULT '',
  "vehicle_number" varchar(20) NOT NULL DEFAULT '',
  "arrival_time" varchar(6) NOT NULL DEFAULT '00:00',
  "departure_time" varchar(6) NOT NULL DEFAULT '00:00',

  -- 1: registered by resident, 2: registered_by security
  "registration_type" int NOT NULL DEFAULT 0,
  -- 1: in, 2:out
  "status" int NOT NULL DEFAULT 0,
  "resident_id" varchar(25) REFERENCES "resident"("id"),
  "unit_id" varchar(25) REFERENCES "unit"("id"),
  "security_id" varchar(25),

  "attr" jsonb NOT NULL DEFAULT '{}'
);


CREATE TABLE "payment" (
  "id" varchar(25) PRIMARY KEY,
  "site_id" varchar(25),
  "resident_id" varchar(25) NOT NULL,
  "date_trx" timestamp NOT NULL DEFAULT localtimestamp,
  "narration" text NOT NULL DEFAULT '',
  "amount" numeric(15,2) NOT NULL,
  "pay_mode" int not null default 0,
  "metadata" jsonb not null default '{}',
  -- which dues were paid for:
  -- array of {id, amount, name}
  "dues" jsonb NOT NULL DEFAULT '[]',
  "attr" jsonb NOT NULL DEFAULT '{}'
);

DELETE FROM "bill";
ALTER TABLE "bill" ADD COLUMN IF NOT EXISTS "status" INTEGER NOT NULL DEFAULT 1;


CREATE TABLE "bill_generate" (
  "id" varchar(25) PRIMARY KEY,
  "site_id" varchar(25),
  "bill_id" varchar(25) NOT NULL REFERENCES "bill"("id"),
  "date_created" timestamp NOT NULL DEFAULT localtimestamp,
  "month" integer not null,
  "year" integer not null,
  "user_id" varchar(25) not null REFERENCES "user"("id"),

  UNIQUE ("site_id", "bill_id", "year", "month")
);

CREATE TABLE "task_queue" (
  "id"    bigserial primary key,
  "site_id" varchar(25),
  "type"  int not null,
  "date_created" timestamp not null DEFAULT LOCALTIMESTAMP,
  "data"  jsonb not null default '{}',
  "status"  int not null default 0
);

CREATE TABLE "registration" (
  "id" VARCHAR(25) primary key,
  "date_created" timestamp not null default LOCALTIMESTAMP,
  "data" jsonb not null default '{}',
  "status" int not null default 0
);

CREATE VIEW "visitor_list" as
select
  v.*,
  concat(u.first_name, ' ', u.last_name) as "security",
  concat(r.first_name, ' ', r.last_name) as "resident"

from
  "visitor" as v
left join "user" as u
  on u.id = v.security_id and u.site_id = v.site_id
left join "residency" as rs 
  on rs.unit_id = v.unit_id and rs.site_id = v.site_id
left join "resident" as r 
  on r.id = v.resident_id
;

CREATE VIEW "resident_list" as
select
  r.id,
  rs.site_id,
  (r.first_name || ' ' || r.last_name) as "name",
  r.email,
  r.phone,
  r.attr,
  r.type,
  rs.unit_id,
  r.primary_id,
  r.residency_id,
  u.type as unit_type
from "resident" as r
left join "residency" as rs on rs.id = r.residency_id 
left join unit as u on u.id = rs.unit_id
where rs.active_status = 1
;

CREATE VIEW "resident_family_list" as
select
  r.id,
  rs.site_id,
  (r.first_name || ' ' || r.last_name) as "name",
  r.email,
  r.phone,
  r.attr,
  r.type,
  rs.unit_id,
  r.residency_id,
  (case
   	when r.primary_id is null then r.id
    else r.primary_id
   end) as primary_id
from "resident" as r
left join "residency" as rs on rs.id = r.residency_id
;

CREATE VIEW "secondary_resident_list" as
select
  r.id,
  rs.site_id,
  (r.first_name || ' ' || r.last_name) as "name",
  r.email,
  r.phone,
  r.attr,
  rs.unit_id,
  r.residency_id,
  r.primary_id
from "resident" as r
left join "residency" as rs on rs.id = r.residency_id
where type = 2;

CREATE VIEW "security_resident_list" as
select
  rs.unit_id,
  r.residency_id,
  r.id,
  rs.site_id,
  (r.attr->>'title' || ' ' || r.first_name || ' ' || r.last_name) as "name",
  r.email,
  r.phone,
  r.type,
  concat(
    (case when u.attr->>'unit_number' is not null then u.attr->>'unit_number'||', ' else '' end)
    , s.name, ', '||u.label
  ) as "unit"

from "resident" as r
left join "residency" as rs on rs.id = r.residency_id
left join "unit" as u on u.id = rs.unit_id
left join "street" as s on s.id = u.street_id
where rs.active_status = 1
order by r.last_name
;

create view "bill_detail_list" as
with "items" as (
select
  i.site_id, i.bill_id, sum(i.amount) as total,
  jsonb_agg(
    jsonb_build_object(
      'name', d.name,
      'due_id', i.due_id,
      'amount', i.amount
    )
  )  as items
from
  "bill_item" as i
join
  due as d on d.id = i.due_id
group by
  i.site_id, i.bill_id
)
select
  b.id, b.site_id, b.unit_type, b.date_created, b.name, b.note,
  b.status,
  ut.label as unit_type_name,
  (case when i.total is null then 0.00 else i.total end) as total,
  i.items
from "bill" as b
left join "unit_type" as ut on ut.id = b.unit_type
left join items as i on i.bill_id = b.id and i.site_id = b.site_id
;

-- CREATE VIEW "invoice_master_list" as
-- select
--   i.*,
--   (select count(j.id) from invoice as j where j.invoice_master_id = i.id) as residents

-- from
--   invoice_master as i
-- ;


CREATE VIEW "invoice_list" AS
with inv_trx as (
  select
  	t.invoice_id,
    json_agg(
        json_build_object('due_id', t.due_id, 'amount', t.amount, 'due', d.name)
    ) as dues

  	from
  		transaction as t
  		left outer join due as d on d.id = t.due_id

 	where
  		t.type = 2

  	group by
  		t.invoice_id

)
select
  i.id, i.site_id, i.invoice_number, i.resident_id,
  i.first_name, i.last_name, i.address,
  i.month, i.year, i.date_created,
  i.bill_id, i.unit_type,
  i.description, i.amount,
  t.dues,
  concat(i.first_name, ' ', i.last_name) as resident,
  u.label as unit_type_label

from
  invoice as i
  left join unit_type as u
    on u.id = i.unit_type

  left join inv_trx as t
    on t.invoice_id = i.id
;


CREATE VIEW "payment_list" AS
select
  p.id,
  p.site_id,
  p.resident_id,
  p.date_trx,
  p.narration,
  p.amount,
  p.pay_mode,
  p.dues,
  p.attr,
  ut.label as unit_type_label,
  p.reference_id,
  r.first_name,
  r.last_name,
  rl.unit_type,
  concat(r.first_name, ' ', r.last_name) as "resident"

from
  payment as p
left join "resident" as r
  on r.id = p.resident_id
left join "resident_list" as rl
  on rl.id = r.id
left join "unit_type" as ut
  on ut.id = rl.unit_type
;

CREATE VIEW "payment_details_list" AS
select
  t.*,
  r.first_name,
  r.last_name,
  concat(r.first_name, ' ', r.last_name) as "resident",
  d.name as "due",
  i.invoice_number,
  i.month,
  i.year

from
  transaction as t
left join "resident" as r
  on r.id = t.resident_id
left join "due" as d
  on d.id = t.due_id
left join "invoice" as i
  on i.id = t.invoice_id
where
  t.type = 2
;

CREATE VIEW "resident_billing_summary" as
with "summary" as (
select
  t.resident_id,
  sum(case when t.type = 2 then t.amount else 0 end) as invoices,
  sum(case when t.type = 1 then t.amount else 0 end) as payments,
  sum(t.amount) as balance

from
  transaction as t
group by
  t.resident_id
)
select
  r.id,
  rs.site_id,
  r.attr->>'title' as "title",
  r.first_name,
  r.last_name,
  r.residency_id,
  rs.unit_id,
  (case when s.invoices is null then 0 else s.invoices end) as invoices,
  (case when s.payments is null then 0 else s.payments end) as payments,
  (case when s.balance is null then 0 else s.balance end) as balance

from
  resident as r
left join "summary" as s
  on s.resident_id = r.id
left join "residency" as rs
  on rs.id = r.residency_id
where
  r.type = 1
;


CREATE VIEW "invoice_summary" as
with "summary" as (
select
  t.invoice_id,
  sum(case when t.type = 2 then t.amount else 0 end) as invoices,
  sum(case when t.type = 1 then t.amount else 0 end) as payments,
  sum(t.amount) as balance

from
  transaction as t
group by
  t.invoice_id
)
select
  i.id,
  i.site_id,
  i.month,
  i.year,
  i.invoice_number,
  i.resident_id,
  r.first_name,
  r.last_name,
  r.residency_id,
  r.attr->>'title' as "title",
  rs.unit_id,
  (case when s.invoices is null then 0 else s.invoices end) as invoices,
  (case when s.payments is null then 0 else s.payments end) as payments,
  (case when s.balance is null then 0 else s.balance end) as balance

from
  "invoice" as i
left join "resident" as r 
  on r.id = i.resident_id
left join "residency" as rs
  on rs.id = r.residency_id
left join
  "summary" as s on s.invoice_id = i.id
;

DROP VIEW if EXISTS "bill_list";
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
  b.status,
  ut.label as unit_type_name,
  (case when i.total is null then 0.00 else i.total end) as total
from "bill" as b
left join "unit_type" as ut on ut.id = b.unit_type
left join items as i on i.bill_id = b.id and i.site_id = b.site_id
;

DROP VIEW "bill_item_list";
CREATE VIEW "bill_item_list" as
SELECT
  b.*,
  d.name as due
from
  bill_item as b
left JOIN
  due as d on d.id = b.due_id
;

CREATE VIEW "resident_account_status" AS
with "account" as (
SELECT
  resident_id,
  sum(t.amount) as balance,
  sum(
    case
      when t.type = 2 then t.amount*-1
      else 0
    end
  ) as invoices,
  sum(
    case
      when t.type = 1 then t.amount
      else 0
    end
  ) as payments
from
  TRANSACTION as t
group by
  resident_id
)
SELECT
  r.id,
  r.first_name,
  r.last_name,
  r.residency_id,
  concat(r.first_name, ' ', r.last_name) as "name",
  r.email,
  r.phone,
  r.attr,
  rs.site_id,
  rs.unit_id,
  u.type as unit_type_id,
  ut.label as unit_type,
  case
    when a.balance is not null then a.balance
    else 0
  end as balance,
  case
    when a.invoices is not null then a.invoices
    else 0
  end as invoices,
  case
    when a.payments is not null then a.payments
    else 0
  end as payments
from
  resident as r
left join residency as rs
  on rs.id = r.residency_id
left join
  account as a on a.resident_id = r.id
left join
  unit as u on u.id = rs.unit_id
left join
  unit_type as ut on ut.id = u.type
where
  r.type = 1
;

CREATE VIEW "resident_due_status" AS
with "account" as (
SELECT
  resident_id,
  due_id,
  sum(t.amount) as balance,
  sum(
    case
      when t.type = 2 then t.amount*-1
      else 0
    end
  ) as invoices,
  sum(
    case
      when t.type = 1 then t.amount
      else 0
    end
  ) as payments

from
  TRANSACTION as t

group by
  t.resident_id, t.due_id
)

SELECT
  r.id,
  rs.site_id,
  r.first_name,
  r.last_name,
  r.residency_id,
  concat(r.first_name, ' ', r.last_name) as "name",
  r.attr,
  rs.unit_id,
  u.type as unit_type,
  d.name as due,
  a.due_id,
  case
    when a.balance is not null then a.balance
    else 0
  end as balance,
  case
    when a.invoices is not null then a.invoices
    else 0
  end as invoices,
  case
    when a.payments is not null then a.payments
    else 0
  end as payments
from
  resident as r
left join residency as rs
  on rs.id = r.residency_id
left join
  account as a on a.resident_id = r.id
left join
  unit as u on u.id = rs.unit_id
left join
  due as d on d.id = a.due_id
where
  r.type = 1
;

CREATE VIEW account_history as
select
  	tr.resident_id, invoice_id as document_id, sum(tr.amount) as amount, date_trunc('second', tr.date_trx) as date_trx, 
    tr.type, concat('0000000', cast(inv.invoice_number as varchar) ) as invoice_number
  from transaction as tr
  left join invoice as inv on inv.id = tr.invoice_id
  where type = 2
  group by tr.resident_id, tr.invoice_id, date_trunc('second', tr.date_trx), tr.type, inv.invoice_number
 
union
  select
  	tr.resident_id, tr.payment_id as document_id, sum(tr.amount) as amount, date_trunc('second', tr.date_trx) as date_trx, 
    tr.type, p.reference_id as invoice_number
  from transaction as tr
  left join payment as p on p.id = tr.payment_id
  where type = 1
  group by tr.resident_id, tr.payment_id, date_trunc('second', tr.date_trx), tr.type, invoice_number
order by  date_trx
;

CREATE TABLE content (
  "id" varchar(25) PRIMARY KEY,
  "site_id" varchar(25),
  "data" jsonb NOT NULL DEFAULT '{}'
);
