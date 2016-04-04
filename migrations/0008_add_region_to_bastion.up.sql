alter table bastions add column region character varying(24) default '';
update bastions set region = (select region from regions where bastions.customer_id = regions.customer_id order by created_at desc limit 1);
alter table bastions alter column region set not null;
