CREATE TABLE public.customers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL
);

CREATE TABLE public.accounts (
    account_id SERIAL PRIMARY KEY,
    customer_id INT REFERENCES public.customers(id),
    balance INT DEFAULT 0
);

INSERT INTO public.customers (name, email) VALUES
    ('John Doe', 'john.doe@example.com'),
    ('Jane Smith', 'jane.smith@example.com'),
    ('Bob Johnson', 'bob.johnson@example.com');