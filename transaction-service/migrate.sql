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
    sender_id INT REFERENCES public.customers(customer_id),
    receiver_id INT REFERENCES public.customers(customer_id),
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