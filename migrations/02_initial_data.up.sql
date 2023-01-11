insert into "site" ("id", "name", "subdomain", "status", "date_registered", "platform")
    values('bkuthn8cijlcn5lv6hqg', 'platform manager', 'manager', 1, '2019-08-29 15:47:00', true);

insert into "user_type" ("id", "label")
    values
        (1, 'services'),
        (2, 'security'),
        (3, 'resident'),
        (4, 'official'),
        (5, 'admin'),
        (6, 'platform')
;

-- password: letmein
insert into "user" ("id", "site_id", "status", "role", "type", "first_name", "last_name", "email", "password", "phone", "is_site_user")
   values ('bkuthj0cijlcd6ohmilg', 'bkuthn8cijlcn5lv6hqg', 1, 4, 6, 'default', 'user', 'defa@user.com', '$2a$14$ZajP6Lt64hzwBd8gMyZDB.6dok6.mGKOJHC3X3o.D1rxVsxiLBaNW', '08092341166', true);

insert into "unit_type" ("id", "label")
    values
        (1, 'flat'),
        (2, 'duplex'),
        (3, 'bungalow'),
        (4, 'boys quarters')
;
