INSERT INTO public.users (user_id, user_password, can_read, can_write, can_delete) VALUES
    ('john', '12345', TRUE, TRUE, TRUE),
    ('jane', '23456', TRUE, FALSE, FALSE),
    ('bob', '34567', FALSE, FALSE, FALSE),
    ('paul', '45678', TRUE, FALSE, TRUE),
    ('alice', '56789', TRUE, TRUE, FALSE);