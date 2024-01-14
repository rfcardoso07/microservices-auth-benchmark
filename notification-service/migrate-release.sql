CREATE TABLE public.users (
    user_id VARCHAR(255) PRIMARY KEY,
    user_password VARCHAR(255) NOT NULL,
    can_read BOOLEAN NOT NULL,
    can_write BOOLEAN NOT NULL,
    can_delete BOOLEAN NOT NULL
);

CREATE TABLE public.notifications (
    notification_id SERIAL PRIMARY KEY,
    transaction_id INT NOT NULL,
    receiver_id INT NOT NULL,
    amount INT NOT NULL
);