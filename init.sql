drop table if exists todos;
create table todos
(
    title text    not null,
    done  boolean not null
);