create or replace view sources_view as
select
    keys.key,
    sources.user_id,
    sources.sid,
    sources.config
from
    sources
join
    keys
on sources.user_id = keys.user_id;