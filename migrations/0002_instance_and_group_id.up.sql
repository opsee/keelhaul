alter table bastions add column instance_id character varying(24);
alter table bastions add column group_id character varying(24);

create index idx_bastions_instance_id on bastions (instance_id);
