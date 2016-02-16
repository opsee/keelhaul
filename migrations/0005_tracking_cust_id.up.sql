delete from bastion_tracking;
alter table bastion_tracking add column customer_id UUID not null;
