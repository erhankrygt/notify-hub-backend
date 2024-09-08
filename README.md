# Assessment Project

## Prepare for Local Development

- Create Environment File for Configuration

> copy  ```.env``` file from ```.env.example```


### If you use Docker for local Development

Example Dockerfile in main project folder


Example docker-compose.yaml in main project folder


If you create the following contents open the project main folder and run :

```shell
docker-compose up --build
```
Then run with following command:


### Without Docker

If you create ```.env``` file open the project main folder and run the following command :
```shell
sh -ac '. ./.env.example; go run cmd/main.go'
```
## Folder Structure:

- cmd folder for main function file
- configs folder for env variable retrieve functions

## Endpoints:

Fetch Sent Messages

- Retrieve a list of sent messages from the server.

```shell
curl --location 'http://localhost:9090/fetch-sent-messages'
```

Switch Auto-Send Mode

- Toggle the auto-send mode of messages on or off.

```shell
curl --location --request POST 'http://localhost:9090/switch-auto-send' \
--header 'accept: application/json'
```

- Service running every 2 minutes name' CronSendMessage

## Client:

Since the curl given in the case did not work, You must change HOOK_CLIENT_URL amd
HOOK_CLIENT_SECRET with active values.

 environment:
      - POSTGRES_DSN=postgres://user:password@db:5432/mydb?sslmode=disable
      - REDIS_ADDRESS=redis:6379
      - REDIS_PASSWORD=redispassword
      - REDIS_DB=0
      - SERVICE_ENVIRONMENT=dev
      - HTTP_SERVER_PORT=:9090
      - HOOK_CLIENT_URL=https://webhook.site/eb8a1637-0cfb-422c-adb3-8efcbd00443d
      - HOOK_CLIENT_SECRET=INS.me1x9uMcyYGlhKKQVPoc.bO3j9aZwRTOcA2Ywo


### Swagger

Swagger Document Url

```shell
http://localhost:9090/docs
```

