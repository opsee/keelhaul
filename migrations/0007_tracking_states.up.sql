alter type bastion_status rename to __bastion_status;

create type bastion_status as enum ('active', 'inactive', 'launching', 'disabled', 'deleted', 'failed_launch');

alter table bastion_tracking rename column status to _status;

alter table bastion_tracking add status bastion_status not null default 'active';

update bastion_tracking set status = _status::text::bastion_status;

alter table bastion_tracking drop column _status;

drop type __bastion_status;
