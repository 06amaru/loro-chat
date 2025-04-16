-- +goose Up
-- +goose StatementBegin

CREATE TABLE public.chats (
	id int8 GENERATED BY DEFAULT AS IDENTITY( INCREMENT BY 1 MINVALUE 1 MAXVALUE 9223372036854775807 START 1 CACHE 1 NO CYCLE) NOT NULL,
	"type" varchar NOT NULL,
	deleted bool DEFAULT false NOT NULL,
	CONSTRAINT chats_pkey PRIMARY KEY (id)
);

CREATE TABLE public.users (
	id int8 GENERATED BY DEFAULT AS IDENTITY( INCREMENT BY 1 MINVALUE 1 MAXVALUE 9223372036854775807 START 1 CACHE 1 NO CYCLE) NOT NULL,
	username varchar NOT NULL,
	created_at timestamptz NOT NULL,
	"password" varchar NOT NULL,
	private_key bytea NOT NULL,
	public_key bytea NOT NULL,
	CONSTRAINT users_pkey PRIMARY KEY (id)
);
CREATE UNIQUE INDEX users_username_key ON public.users USING btree (username);

CREATE TABLE public.chat_members (
	chat_id int8 NOT NULL,
	user_id int8 NOT NULL,
	CONSTRAINT chat_members_pkey PRIMARY KEY (chat_id, user_id),
	CONSTRAINT chat_members_chat_id FOREIGN KEY (chat_id) REFERENCES public.chats(id) ON DELETE CASCADE,
	CONSTRAINT chat_members_user_id FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE
);

CREATE TABLE public.messages (
	id int8 GENERATED BY DEFAULT AS IDENTITY( INCREMENT BY 1 MINVALUE 1 MAXVALUE 9223372036854775807 START 1 CACHE 1 NO CYCLE) NOT NULL,
	body varchar NOT NULL,
	created_at timestamptz NOT NULL,
	user_messages int8 NULL,
	CONSTRAINT messages_pkey PRIMARY KEY (id),
	CONSTRAINT messages_users_messages FOREIGN KEY (user_messages) REFERENCES public.users(id) ON DELETE SET NULL
);

CREATE TABLE public.chat_messages (
	chat_id int8 NOT NULL,
	message_id int8 NOT NULL,
	CONSTRAINT chat_messages_pkey PRIMARY KEY (chat_id, message_id),
	CONSTRAINT chat_messages_chat_id FOREIGN KEY (chat_id) REFERENCES public.chats(id) ON DELETE CASCADE,
	CONSTRAINT chat_messages_message_id FOREIGN KEY (message_id) REFERENCES public.messages(id) ON DELETE CASCADE
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE public.chat_messages;
DROP TABLE public.messages;
DROP TABLE public.chat_members;
DROP TABLE public.users;
DROP TABLE public.chats;

-- +goose StatementEnd
