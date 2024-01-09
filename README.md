# microservices-auth-benchmark

--------------------------------------------------------------------------------------------------

Account Service <> Create Account

Account Service <> Delete Account

Account Service <> Delete Accounts By Customer

Account Service <> Add To Balance

Account Service <> Subtract From Balance

Account Service <> Get Account

Account Service <> Get Accounts By Customer

--------------------------------------------------------------------------------------------------

Customer Service <> Create Customer
    --> Account Service <> Create Account

Customer Service <> Delete Customer
    --> Account Service <> Delete Accounts By Customer

Customer Service <> Get Customer

--------------------------------------------------------------------------------------------------

Balance Service <> Get Balance By Customer
    --> Account Service <> Get Accounts By Customer

Balance Service <> Get Balance History

--------------------------------------------------------------------------------------------------

Notification Service <> Notify
    --> Account Service <> Get Account
    --> Customer Service <> Get Customer

Notification Service <> Get Notification

--------------------------------------------------------------------------------------------------

Transaction Service <> Transfer Amount
    --> Account Service <> Add To Balance
    --> Account Service <> Subtract From Balance

Transaction Service <> Transfer Amount And Notify
    --> Account Service <> Add To Balance
    --> Account Service <> Subtract From Balance
    --> Notification Service <> Notify
        --> Account Service <> Get Account
        --> Customer Service <> Get Customer

Transaction Service <> Get Transaction

--------------------------------------------------------------------------------------------------