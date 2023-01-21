# [e-inwork.com](https://e-inwork.com)

## Golang User Microservice
Settings up the Docker Environment & run this microservice:
1. Install Docker
    - https://docs.docker.com/get-docker/
2. Git clone this repository to your localhost, and from the terminal, run the below command:
   ```
   git clone git@github.com:e-inwork-com/go-user-service
   ```
3. Change the active folder to `go-user-service`:
   ```
   cd go-user-service
   ```
4. Run the Docker Compose local:
   ```
   docker-compose -f docker-compose.local.yml up -d
   ```
5. Continue to the next step after the `migrate-local` status is `exited (0)`:
   ```
   docker-compose -f docker-compose.local.yml ps
   ```
6. Create a user in the User API with the CURL command line:
    ```
    curl -d '{"email":"jon@doe.com", "password":"pa55word", "first_name": "Jon", "last_name": "Doe"}' -H "Content-Type: application/json" -X POST http://localhost:8000/service/users
    ```
7. Login to the User API:
   ```
   curl -d '{"email":"jon@doe.com", "password":"pa55word"}' -H "Content-Type: application/json" -X POST http://localhost:8000/service/users/authentication
   ```
8. You will get a token from the response login and set it as a `token` variable for an example like the below:
   ```
   token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjhhY2NkNTUzLWIwZTgtNDYxNC1iOTY0LTA5MTYyODhkMmExOCIsImV4cCI6MTY3MjUyMTQ1M30.S-G5gGvetOrdQTLOw46SmEv-odQZ5cqqA1KtQm0XaL4
   ```
9. Get a record of the current user with the User API:
   ```
   curl -H "Authorization: Bearer $token" -X GET http://localhost:8000/service/users/me
   ```
10. Run unit testing (required Golang Version: 1.19.4):
    ```
    # From folder "go-team-service", run:
    go mod tidy
    go test -v -run TestRoutes ./api
    ```
11. Run end to end testing (required Golang Version: 1.19.4):
    ```
    # Down the Docker Compose local if you run it on No. 4
    docker-compose -f docker-compose.local.yml down
    # Run the Docker Compose test
    docker-compose -f docker-compose.test.yml up -d
    # Check the status "curl-test" and "migrate-test", and wait until status "exited (0)", run bellow command to check it
    docker-compose -f docker-compose.test.yml ps
    # Run end to end tesing
    go test -v -run TestE2E ./api
    ```
12. This application will create a folder `local` on the current directory. You can delete the `local` folder if you want to run from the start and run `docker system prune -a` if necessary.
13. Good luck!
