create extension if not exists "uuid-ossp";

create or replace function update_updated_at() returns trigger as $$
begin
    new.updated_at = now();
    return new;
end;
$$ language plpgsql;

create table if not exists users (
   id uuid not null primary key default uuid_generate_v4(),
   name varchar(255) not null,
   email varchar(255) not null,
   created_at timestamp not null default now(),
   updated_at timestamp not null default now()
);
create trigger users_update_updated_at_trigger before update on users for each row execute procedure update_updated_at();

create table if not exists sources (
    id uuid not null primary key default uuid_generate_v4(),
    sid varchar(32) not null unique,
    user_id uuid not null references users(id) on delete cascade,
    name varchar(255) not null,
    type varchar(32) not null default 's3',
    config jsonb not null default '{}',
    created_at timestamp not null default now(),
    updated_at timestamp not null default now()
);
create trigger sources_update_updated_at_trigger before update on sources for each row execute procedure update_updated_at();

create table if not exists keys (
  id uuid not null primary key default uuid_generate_v4(),
  user_id uuid not null references users(id) on delete cascade,
  key varchar(255) not null,
  created_at timestamp not null default now(),
  updated_at timestamp not null default now()
);
create trigger keys_update_updated_at_trigger before update on keys for each row execute procedure update_updated_at();


