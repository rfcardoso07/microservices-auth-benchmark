CREATE TABLE public.users (
    user_id PRIMARY KEY,
    user_password VARCHAR(255) NOT NULL,
    can_read BOOLEAN NOT NULL,
    can_write BOOLEAN NOT NULL,
    can_delete BOOLEAN NOT NULL
);

INSERT INTO public.users (user_id, user_password, can_read, can_write, can_delete) VALUES
    ('john', '12345', 1, 1, 1),
    ('jane', '23456', 1, 0, 0),
    ('bob', '34567', 0, 0, 0),
    ('paul', '45678', 1, 0, 1);