# Sally API Gateway

Sally is an api gateway which allows you to configure proxies for your various application services.
Currently, we support only REST protocols, but in a couple of weeks would be shipping support for GRPC protocol and a web client for ease of navigation.

### Getting Started

1. Clone the repository

```
  $ git clone https://github.com/Oluwatunmise-olat/Sally-API-Gateway.git
  $ cd Sally-API-Gateway
```

2. Start the application
   Note: It is assumed you have `golang` installed already, preferably from version `1.23.x`. See the [documentation](https://go.dev/dl/) to install golang

```
  $ PORT=XXXX go run main.go
```

3. Upload your config file (Yaml)
   Upload via terminal

```
  $ curl -X POST -F "config_file=@path/to/your-config-file" http://localhost:$PORT/gw-upload

  Or using postman

    - Open Postman and create a new POST request
    - Set request URL to http://localhost:$PORT/gw-upload
    - In the request body, select the form-data option
    - Add a field named `config_file` and select the config file you want to upload
    - Click the "Send" button to make the request
```

4. Make Calls to your services as defined in your config file

### Understanding Config File Format and Structure

The configuration file supported should be in a `.yaml` extension.
It follows the [Open API Specification](https://swagger.io/specification/) version 3.0.0 standard to define configs.

You find a sample configuration file in `sample.yaml`.

### Explanation of configuration file

Note: The order of route matters. i.e Order your more specific endpoint paths before the more generic paths.

- openapi: 3.0.0: This indicates the OpenAPI version used for the configuration file.
- info: The info section provides metadata about the API, such as its title and version.
- paths: The paths section defines API routes and their associated methods (e.g., GET, POST). This is a major section in the configuration file.
- tags: The tags section allows grouping related API paths for better organization, ease of defining a single target url and documentation. i.e
  Say i have a service for handling `payments` related actions e.g `transfer`, `fetching transaction details`, I can group them under a tag like this

```

**tags**:
  - name: Payment
  - description: Payment Service
  - x-target: Service to proxy request to.
```

This helps prevent `x-target` repetition and allows for ease of update in a single section per http method and per route as they can be associated with the defined tag as follows

```
  /users:
    get:
      summary: Get a list of users
      responses:
        '200':
      description: Successful response
      x-tag: Users
      x-target: https://httpbin.org/get
    options:
      summary: Random
      responses:
        '200':
      description: Successful response
      x-tag: Users
      x-target: https://httpbin.org/get
```

**x-tag and x-target**: These are custom extensions that specify the tag associated with a path and the target URL where the request will be forwarded.

**x-target**: Service to proxy request to. There is an order of precedence with this parameter. If you have a route method say

```
  /users:
    get:
      x-target: https://httpbin.org/get
      x-tag: Users

```

And you specify the `x-target` and also specify the `x-tag` which already has an `x-target`
defined as it is required there, the `x-target` defined in the above block takes more weight and is used
over the `x-target` defined in the `Tags` section.

**parameters**: This section specifies the request parameters, such as path parameters and query parameters.
This section is very important especially when you have routes such as `/users/:id`, as this is what we use to resolve those path parameters. Below, you find a way to properly define the parameters

```
/users/{id}:
  get:
    summary: Get a user by ID
    parameters:
      - name: id
        in: path
        required: true
        schema:
          type: integer
    responses:
      '200':
        description: Successful response
      '404':
        description: User not found
    x-target: https://httpbin.org/status/{id}
```

**responses (Optional)**: The responses section defines the expected responses for a given path and HTTP method. This would be useful for the
web client.

### More on paths

It can be daunting to have to define routes as seen below

```
paths:
  /users/:
    get:
      summary: xyz
      x-target: https://httpbin.org/users
    options:
      x-target: https://httpbin.org/users
      summary: xyz
    post:
      x-target: https://httpbin.org/users
      summary: xyz
    delete:
      x-target: https://httpbin.org/users/
      summary: xyz
  /users/profile:
    get:
      summary: xyz
      x-target: https://httpbin.org/users/profile
    options:
    x-target: https://httpbin.org/users
```

You get the point, we have a way to prevent this by using a path pattern

```
paths:
  /{users+}:
    get:
      summary: xyz
      x-target: https://httpbin.org/users
    options:
      x-target: https://httpbin.org/users
      summary: xyz
    post:
      x-target: https://httpbin.org/users
      summary: xyz
```

This way when you hit `/users/path0` or `/users/path1/profile` we handle the process of proxing respectively as
`https://httpbin.org/users/path0` and `https://httpbin.org/users/path1/profile`.

#### v0 TODO

- [x] Config File Upload and Validation
- [x] Base proxy in place
- [x] Pass query strings and parameters and validations
- [x] Regex matcher
- [x] Prevent path overrides
- [x] Testing

#### v1 TODO

- [ ] Validate request header and request body
- [x] Add Logging integration
- [ ] Ip white listing/ black listing
- [ ] Versioned rollbacks
- [ ] Allow multiple config files
- [ ] Grpc Support

#### Docker
