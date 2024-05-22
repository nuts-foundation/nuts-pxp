-- +goose ENVSUB ON
-- +goose Up
create table data
(
    id varchar(255) not null primary key,
    scope      varchar(255) not null,
    verifier   varchar(370) not null,
    client     varchar(370) not null,
    auth_input  $TEXT_TYPE
);
create index idx_data on data (scope, verifier, client);

-- +goose Down
drop index idx_data;
drop table data;
