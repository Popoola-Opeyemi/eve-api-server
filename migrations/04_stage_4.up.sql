create view "old_units_residents" as 
select 
  r.id,
  r.first_name,
  r.last_name,
  (r.first_name || ' ' || r.last_name) as "name",
  r.email,
  r.phone,
  r.attr,
  r.type,
  rs.date_start,
  rs.date_exit,
  rs.previous_unit_id,
    concat(
    s.name, ', ',
    (case when u.attr->>'unit_number' is not null then u.attr->>'unit_number'||', ' else '' end), u.label || ''
  ) as "unit",

  rs.site_id
  from resident as r
  join residency as rs on rs.id = r.residency_id
  left join "site" as a on a.id = rs.site_id
  left join "unit" as u on u.id = rs.previous_unit_id
  left join "street" as s on s.id = u.street_id
  where rs.active_status = 0;

CREATE TABLE "new_resident_registrations" (
  "id" varchar(25) PRIMARY KEY,
  "site_id" varchar(25) not null references site(id),
  "site_name" varchar(150) not null default '',
  "email" varchar(250) UNIQUE NOT NULL,
  "first_name" varchar(150) NOT NULL,
  "last_name" varchar(150) NOT NULL,
  "name" varchar(300) not null, 
  "phone" varchar(150) NOT NULL,
  "address" varchar(150) NOT NULL,
  "attr" jsonb not null default '{}',
  "date_registered" timestamp not null default LOCALTIMESTAMP
);

create view "payment_details" as 
select 
p.id,
p.site_id,
p.amount,
p.reference_id,
p.dues,
(case 
when p.pay_mode = 0 then 'manual'
when p.pay_mode = 2 then 'bank transaction' 
when p.pay_mode = 3 then 'online' end) as pay_mode,
t.id as "transaction_id",
r.email,
(case when p.pay_mode = 1 then p.attr->>'provider_name' else '' end) as "provider",
cast (p.attr->>'transaction_id' as varchar(500)) as "provider_transaction_id",
cast (p.attr->>'reference_id' as varchar(500)) as "provider_reference_id",
concat(r.first_name, ' ', r.last_name) as "name",
p.date_trx as "date"
from payment as p
left join "resident" as r on r.id = p.resident_id
left join "transaction" as t on t.payment_id = p.id 
;

create table "payment_providers" (
  "id" varchar(25) PRIMARY KEY,
  "name" varchar(50) not null,
  "value" int not null
);

insert into "payment_providers" ("id", "name", "value") values 
('c17hv9ocvcqi2p2jvusg', 'manual', 0),
('c17hv9ocvcqi2p2jvut0', 'paystack', 1),
('c17hv9ocvcqi2p2jvutg', 'remita', 2),
('c17hv9ocvcqi2p2jvuu0', 'flutterwave', 3);

create table "payment_log" (
  "id" varchar(25) PRIMARY KEY,
  "site_id" varchar(25) not null,
  "operation" varchar(25) not null,
  "amount" numeric(15,2) not null default 0,
  "initiated_by" jsonb not null default '{}',
  "initiated_for" jsonb not null default '{}',
  "narration" text not null default '',
  "date" timestamp not null default LOCALTIMESTAMP
);

create table "payment_pending" (
  "id" varchar(25) PRIMARY KEY,
  "site_id" varchar(25),
  "resident_id" varchar(25) NOT NULL,
  "resident_name" varchar(25) NOT NULL,
  "date_trx" timestamp NOT NULL DEFAULT localtimestamp,
  "narration" text NOT NULL DEFAULT '',
  "amount" numeric(15,2) NOT NULL,
  "pay_mode" int not null default 0,
  "dues" jsonb NOT NULL DEFAULT '[]',
  "attr" jsonb NOT NULL DEFAULT '{}'
);


create view "reporting_residents" as
select
split_part(u.label::varchar , ', ', 2) as label,
concat(u.attr->>'unit_number', ' ', u.street) as unit,
(case when r.id is not null then 'occupied' else 'vacant' end) as occupied_status,
r.id,
r.first_name,
r.last_name,
r.email,
r.phone,
(case when r.type = 1 then 'Primary resident' 
else 'Secondary resident' end ) as resident_type,
rs.active_status,
u.attr,
u.street,
u.unit_type,
u.site_id
from unit_list as u
left join residency as rs on rs.unit_id = u.id
left join resident as r on r.residency_id = rs.id
left join resident_account_status as ras on ras.id = r.id
where rs.id is not null 
;


create view "reporting_payments" as 
select 
split_part(u.label::varchar , ', ', 2) as label,
u.attr,
u.street,
u.unit_type,
p.amount,
p.date_trx,
p.site_id,
t.id as transaction_id,
concat(r.first_name, ' ', r.last_name) as name,
p.dues,
(case 
when p.pay_mode = 0 then 'manual'
when p.pay_mode = 2 then 'bank transaction' 
when p.pay_mode = 3 then 'online' end) as pay_mode
from payment as p
left join resident as r on r.id = p.resident_id
left join residency as rs on rs.id = r.residency_id
left join unit_list as u on u.id = rs.unit_id
left join transaction as t on t.payment_id = p.id
;


create view "reporting_bill" as  
select 
b.site_id,
ut.label as unit_label,
b.date_created,
b.name as bill_name,
b.total
from bill as b
left join unit_type as ut on ut.id = b.unit_type
;

create view "reporting_unit" as 
select
u.site_id,
ut.label as unit_label,
u.label,
s.name as street_name,
u.attr->>'unit_number' as unit_no
from unit as u
left join street as s on s.id = u.street_id
left join unit_type as ut on ut.id = u.type
;

create view "reporting_invoice" as 
select
i.id, i.site_id, i.resident_id,
i.first_name, i.last_name, i.address,
i.month, i.year, i.date_created,
i.bill_id, i.unit_type,
i.description, i.amount,
i.dues,
lpad(i.invoice_number::varchar, 8, '0') as invoice_number,
concat(i.first_name, ' ', i.last_name) as resident
from invoice_list as i
left join unit_type as ut on ut.id = i.unit_type
;

CREATE SEQUENCE payment_sequence;
alter table payment add column reference_id varchar(20) NOT NULL DEFAULT lpad(nextval('payment_sequence')::text,8,'0');

alter table public.user add column support_account boolean default false;
insert into "user_type" ("id", "label") values (7, 'support');

alter table public.site add column site_code varchar(50) not null default ''