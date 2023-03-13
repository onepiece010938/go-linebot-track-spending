-- +goose Up
CREATE TABLE "items" (
  "id" BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  "user_id" bigint NOT NULL,
  "name" varchar(255) NOT NULL,
  "balance_limit" integer NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "records" (
  "id" BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  "item_id" bigint NOT NULL,
  "name" varchar(255) NOT NULL,
  "price" integer NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "users" (
  "id" BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  "lineid" varchar(100) NOT NULL,
  "name" varchar(255) NOT NULL,
  "month_limit" integer NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now())
);

ALTER TABLE "items" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");

ALTER TABLE "records" ADD FOREIGN KEY ("item_id") REFERENCES "items" ("id");
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- +goose Down
DROP TABLE records;
DROP TABLE items;
DROP TABLE users;
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
