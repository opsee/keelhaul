SET statement_timeout = 0;
SET lock_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET client_min_messages = warning;

SET default_tablespace = '';

create type bastion_status as enum ('active', 'inactive');

create table bastion_tracking (
    id UUID primary key not null,
    status bastion_status not null default 'active',
    last_seen timestamp with time zone DEFAULT to_timestamp(0) NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);

create index idx_tracking_status on bastion_tracking (status);
create trigger update_tracking_status before update on bastion_tracking for each row execute procedure update_time();

create function batch_upsert_tracking(UUID[]) returns void as $$
begin
    create temporary table tracking_updates(id UUID)
        on commit drop;
    insert into tracking_updates(id)
        select * from unnest($1);
    lock table bastion_tracking in exclusive mode;

    update bastion_tracking
        set last_seen = now()
        from tracking_updates
        where tracking_updates.id = bastion_tracking.id;

    insert into bastion_tracking(id, last_seen)
        select tracking_updates.id, now()
        from tracking_updates
        left outer join bastion_tracking on (bastion_tracking.id = tracking_updates.id)
        where bastion_tracking.id is null;
end;
$$ language plpgsql;
