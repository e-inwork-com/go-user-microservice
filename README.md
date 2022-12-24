# [e-inwork.com](https://e-inwork.com)

## Gettings Started
### Settings Up Docker Environment & Run Application
1. Install Docker
    - https://docs.docker.com/get-docker/
2. Git clone this repository to your localhost, and from the terminal run below command:
   ```
   git clone git@github.com:e-inwork-com/go-user-service
   ```
3. Change the active folder to `go-user-service`:
   ```
   cd go-user-service
   ```
4. Run Docker Compose:
   ```
   docker-compose up -d
   ```
5. Go to the terminal of the Docker PostgreSQL Container:
   ```
   docker-compose exec db /bin/sh
   ```
6. Login to PostgreSQL:
   ```
   psql --username postgres
   ```
7. Create User Table in PostgreSQL database:
   ```
   CREATE TABLE IF NOT EXISTS users (
       id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
       created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
       email text UNIQUE NOT NULL,
       password_hash bytea NOT NULL,
       first_name char varying(100) NOT NULL,
       last_name char varying(100) NOT NULL,
       activated bool NOT NULL DEFAULT false,
       version integer NOT NULL DEFAULT 1
   );
   ```
8. Exit from PostgreSQL:  
   ```
   exit
   ```
9. Exit again from the Docker PostgreSQL Container:  
   ```
   exit
   ```
10. Create a user in the User API with CURL command line:  
    ```
    curl -d '{"email":"jon@doe.com", "password":"pa55word", "first_name": "Jon", "last_name": "Doe"}' -H "Content-Type: application/json" -X POST http://localhost:4000/api/users
    ```
11. Login to the User API:
    ```
    curl -d '{"email":"jon@doe.com", "password":"pa55word"}' -H "Content-Type: application/json" -X POST http://localhost:4000/api/authentication
    ```
12. You will get a token from the response login, and set it as a token variable for example like below:
    ```
    token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjhhY2NkNTUzLWIwZTgtNDYxNC1iOTY0LTA5MTYyODhkMmExOCIsImV4cCI6MTY3MjUyMTQ1M30.S-G5gGvetOrdQTLOw46SmEv-odQZ5cqqA1KtQm0XaL4
    ```
13. Get a record of the current user with the User API:
    ```
    curl -H "Authorization: Bearer $token" -X GET http://localhost:4000/api/users/me
    ```
14. Good luck!
