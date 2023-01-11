create table "resident_alerts" (
  "id" varchar(25) Primary KEY,
  "resident_id" varchar(25) not null,
  "site_id" varchar(25),
  "name" varchar(300),
  "phone_number" varchar(30),
  "address" varchar(250),
  "status" int not null default 1,
  "attr" jsonb NOT NULL DEFAULT '{}',
  "time_logged" timestamp not null default LOCALTIMESTAMP,
  "time_responded" timestamp not null default LOCALTIMESTAMP
);

alter table public.gate_pass add column plate_number varchar(50) not null default '';

alter table public.resident add column push_token varchar(200) default '';

alter table public.user add column push_token varchar(200) not null default '';

alter table public.user add column address varchar(200) default '';

alter table public.gate_pass alter column plate_number drop not null;
