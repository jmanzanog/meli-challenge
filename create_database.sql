create table if not exists item
(
	id varchar not null
		constraint item_pk
			primary key,
	title varchar,
	category_id varchar,
	price bigint default 0,
	start_time time,
	stop_time time
);

alter table item owner to postgres;

create table if not exists child
(
	id varchar,
	stop_time time,
	parent_id varchar
		constraint child_item_id_fk
			references item
);

alter table child owner to postgres;

