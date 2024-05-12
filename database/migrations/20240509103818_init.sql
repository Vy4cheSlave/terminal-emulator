-- +goose Up
-- +goose StatementBegin
create table if not exists commands (id serial primary key, command text not null, is_error boolean not null, log text not null);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists commands;
-- +goose StatementEnd
