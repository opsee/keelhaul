SET statement_timeout = 0;
SET lock_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;

CREATE EXTENSION "uuid-ossp";

CREATE FUNCTION update_time() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
      BEGIN
      NEW.updated_at := CURRENT_TIMESTAMP;
      RETURN NEW;
      END;
      $$;


SET default_tablespace = '';
SET default_with_oids = false;

create type bastion_state as enum ('new', 'launching', 'failed', 'active', 'disabled', 'deleted');

create table bastions (
    id UUID primary key default uuid_generate_v1mc(),
    customer_id UUID not null,
    user_id int not null,
    stack_id character varying(256),
    image_id character varying(256),
    instance_type character varying(24) not null,
    vpc_id character varying(24) not null,
    subnet_id character varying(24) not null,
    state bastion_state not null default 'new',
    password_hash character varying(60) NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);

create index idx_bastions_customers on bastions (customer_id);
create trigger update_bastions before update on bastions for each row execute procedure update_time();

create table regions (
  customer_id UUID not null,
  region character varying(24) not null,
  data jsonb not null,
  primary key (customer_id, region),
  created_at timestamp with time zone DEFAULT now() NOT NULL,
  updated_at timestamp with time zone DEFAULT now() NOT NULL
);

create index idx_regions_customers on regions (customer_id);
create trigger update_regions before update on regions for each row execute procedure update_time();
