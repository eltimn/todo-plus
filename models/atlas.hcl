table "users" {
  schema = schema.main
  column "id" {
    null = true
    type = integer
  }
  column "email" {
    null = false
    type = text
  }
  column "full_name" {
    null = false
    type = text
  }
  column "username" {
    null = false
    type = text
  }
  column "password" {
    null = false
    type = text
  }
  primary_key {
    columns = [column.id]
  }
  index "idx_users_email" {
    unique  = true
    columns = [column.email]
  }
  index "idx_users_username" {
    unique  = true
    columns = [column.username]
  }
}
table "todos" {
  schema = schema.main
  column "id" {
    null = true
    type = integer
  }
  column "user_id" {
    null = false
    type = integer
  }
  column "plain_text" {
    null = false
    type = text
  }
  column "rich_text" {
    null = false
    type = text
  }
  column "is_completed" {
    null = false
    type = boolean
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "0" {
    columns     = [column.user_id]
    ref_columns = [table.users.column.id]
    on_update   = CASCADE
    on_delete   = RESTRICT
  }
}
schema "main" {
}
