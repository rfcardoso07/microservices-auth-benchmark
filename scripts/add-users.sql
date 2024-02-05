INSERT INTO public.users (user_id, user_password, can_read, can_write, can_delete) VALUES
    ('john', '12345', TRUE, TRUE, TRUE),
    ('jane', '23456', TRUE, FALSE, FALSE),
    ('bob', '34567', FALSE, FALSE, FALSE),
    ('paul', '45678', TRUE, FALSE, TRUE),
    ('alice', '56789', TRUE, TRUE, FALSE);

-- Generate random user data and insert into the users table
DO $$
DECLARE
    i INTEGER := 1;
    user_id_prefix VARCHAR(10) := 'user';
    user_password_length INTEGER := 8;
BEGIN
    FOR i IN 1..10000 LOOP
        INSERT INTO public.users (user_id, user_password, can_read, can_write, can_delete)
        VALUES (
            user_id_prefix || i,
            md5(random()::text), -- Generating a random password using MD5 hash
            CASE WHEN random() < 0.5 THEN TRUE ELSE FALSE END,
            CASE WHEN random() < 0.5 THEN TRUE ELSE FALSE END,
            CASE WHEN random() < 0.5 THEN TRUE ELSE FALSE END
        );
    END LOOP;
END $$;