alter table users
    add column otp_secret text;

alter table users
    add column otp_auth_url text;