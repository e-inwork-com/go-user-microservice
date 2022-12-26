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
5. Create a user in the User API with CURL command line:  
    ```
    curl -d '{"email":"jon@doe.com", "password":"pa55word", "first_name": "Jon", "last_name": "Doe"}' -H "Content-Type: application/json" -X POST http://localhost:4000/service/users
    ```
6. Login to the User API:
   ```
   curl -d '{"email":"jon@doe.com", "password":"pa55word"}' -H "Content-Type: application/json" -X POST http://localhost:4000/service/users/authentication
   ```
7. You will get a token from the response login, and set it as a token variable for example like below:
   ```
   token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjhhY2NkNTUzLWIwZTgtNDYxNC1iOTY0LTA5MTYyODhkMmExOCIsImV4cCI6MTY3MjUyMTQ1M30.S-G5gGvetOrdQTLOw46SmEv-odQZ5cqqA1KtQm0XaL4
   ```
8. Get a record of the current user with the User API:
   ```
   curl -H "Authorization: Bearer $token" -X GET http://localhost:4000/service/users/me
   ```
9. Good luck!
