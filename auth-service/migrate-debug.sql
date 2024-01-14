CREATE TABLE public.users (
    user_id VARCHAR(255) PRIMARY KEY,
    user_password VARCHAR(255) NOT NULL,
    can_read BOOLEAN NOT NULL,
    can_write BOOLEAN NOT NULL,
    can_delete BOOLEAN NOT NULL
);

INSERT INTO public.users (user_id, user_password, can_read, can_write, can_delete) VALUES
    ('john', '12345', TRUE, TRUE, TRUE),
    ('jane', '23456', TRUE, FALSE, FALSE),
    ('bob', '34567', FALSE, FALSE, FALSE),
    ('paul', '45678', TRUE, FALSE, TRUE);