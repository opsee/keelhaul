drop function batch_upsert_tracking (uuid[]);

create function batch_upsert_tracking(_bast_ids UUID[], _cust_ids UUID[]) returns void as $$
begin
    create temporary table tracking_updates(id UUID, cust_id UUID)
        on commit drop;
    insert into tracking_updates(id, cust_id)
        select * from unnest(_bast_ids, _cust_ids);

    lock table bastion_tracking in exclusive mode;

    update bastion_tracking
        set last_seen = now()
        from tracking_updates
        where tracking_updates.id = bastion_tracking.id;

    insert into bastion_tracking(id, customer_id, last_seen)
        select tracking_updates.id, tracking_updates.cust_id, now()
        from tracking_updates
        left outer join bastion_tracking on (bastion_tracking.id = tracking_updates.id)
        where bastion_tracking.id is null;
end;
$$ language plpgsql;
