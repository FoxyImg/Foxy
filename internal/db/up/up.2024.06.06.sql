alter table sources drop column if exists user_id;

alter table apps add column if not exists admin_key varchar(64)  unique;
update apps set admin_key = uuid_generate_v5(uuid_generate_v4(), 'admin key') where admin_key is null;

create table if not exists migrations (
    name varchar(255) not null primary key unique,
    date_created timestamp not null default now()
);
