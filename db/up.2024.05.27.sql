drop view if exists sources_view;
drop table if exists keys;

create table if not exists apps (
    id uuid not null primary key default uuid_generate_v4(),
    user_id uuid not null references users(id) on delete cascade,
    name varchar(64) not null,
    key varchar(32) not null unique,
    created_at timestamp not null default now(),
    updated_at timestamp not null default now()
);
create trigger apps_update_updated_at_trigger before update on apps for each row execute procedure update_updated_at();

alter table sources add column app_id uuid references apps(id) on delete cascade;

create table if not exists presets (
    id uuid not null primary key default uuid_generate_v4(),
    app_id uuid not null references apps(id) on delete cascade,
    name varchar(255) not null,
    version bigint not null default 0,
    preset jsonb not null default '{}',
    created_at timestamp not null default now(),
    updated_at timestamp not null default now()
);
create trigger presets_update_updated_at_trigger before update on presets for each row execute procedure update_updated_at();

create or replace function presets_increment_version() returns trigger as $$
begin
    new.version = new.version + 1;
    return new;
end;
$$ language plpgsql;
create trigger presets_increment_version_trigger before update on presets for each row execute procedure presets_increment_version();

create or replace view sources_view as
select
    apps.key,
    sources.app_id,
    sources.sid,
    sources.config
from
    sources
join
    apps on sources.app_id = apps.id;

