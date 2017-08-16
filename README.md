# Please see branch *dev* for the latest code.

This project manages subscriptions to changes in cryptocurrency prices. Upon the set change percentage by the user,
a SMS message will be sent. You can find the associated frontend for this project here. https://github.com/jgpasch/ng-noti

Steps to run:

## 1. Clone this repo
## 2. Start your postgres server
## 3. cd into the main folder
## 4. open main.go and change the parameters to the Initialize function.
    1. postgres user
    2. db name
    3. password (if needed)
## 5. ensure that a db by that name exists.

## 6. enter the psql command line and create the subscription table
    CREATE TABLE IF NOT EXISTS subs
    (
      id SERIAL PRIMARY KEY,
      token VARCHAR(30) NOT NULL,
      percent NUMERIC(10,2) NOT NULL DEFAULT 0,
      minval NUMERIC(10,2) NOT NULL DEFAULT 0,
      maxval NUMERIC(10,2) NOT NULL DEFAULT 0,
      minmaxchange NUMERIC(10,2) NOT NULL DEFAULT 0,
      owner TEXT NOT NULL,
      active BOOLEAN DEFAULT FALSE
    )

## 7. ensure the user table exists
    CREATE TABLE IF NOT EXISTS users
    (
      id SERIAL PRIMARY KEY,
      email TEXT NOT NULL,
      password TEXT NOT NULL,
      number TEXT DEFAULT 0,
      request_id TEXT DEFAULT 0
    )

## 8. Change nexmo details to use your own account.
    1. Create a file in the folder *main* called config.json
    2. create a json object with three items
      1. port - port on which the server will run
      2. nexmo_api_key: from your nexmo account
      3. nexmo_secret: from your nexmo account

## 9. execute command ```go run !(*_test).go```
