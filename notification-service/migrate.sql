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

CREATE TABLE public.customers (
    customer_id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL
);

CREATE TABLE public.accounts (
    account_id SERIAL PRIMARY KEY,
    customer_id INT REFERENCES public.customers(customer_id),
    balance INT DEFAULT 0
);

CREATE TABLE public.transactions (
    transaction_id SERIAL PRIMARY KEY,
    sender_id INT REFERENCES public.accounts(account_id),
    receiver_id INT REFERENCES public.accounts(account_id),
    amount INT NOT NULL
);

CREATE TABLE public.notifications (
    notification_id SERIAL PRIMARY KEY,
    transaction_id INT REFERENCES public.transactions(transaction_id),
    receiver_id INT REFERENCES public.accounts(account_id),
    amount INT NOT NULL
);

INSERT INTO public.customers (name, email) VALUES
    ('John Doe', 'john.doe@example.com'),
    ('Jane Smith', 'jane.smith@example.com'),
    ('Bob Johnson', 'bob.johnson@example.com');

INSERT INTO public.accounts (customer_id, balance) VALUES
    (1, 0),
    (2, 0),
    (3, 0);

INSERT INTO public.transactions (sender_id, receiver_id, amount) VALUES
    (1, 2, 10),
    (2, 3, 20),
    (3, 1, 30);