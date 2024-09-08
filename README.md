# AuthAPI

This module implements the services including:
1) Registering a user. (In memory implementation)
2) Login a user with his credentials and returning back a jwt token.
3) Refreshing the user token and invalidating the old one.
4) Revoking the user token.(In memory implementation)
5) Returning back the protected resources on valid token request.

### Starting the application:

```shell
go run main.go
```


### Use cases

**PLEASE RUN THESE COMMANDS IN BASH(EXAMPLE, GIT BASH), IF YOU ARE 0N WINDOWS.**

1) Using SignIn when user does not exist.
```shell
curl --location 'localhost:8080/api/auth/signin' --header 'Content-Type: application/json' --data '{"email":"test@example.com","password":"password123"}' -c cookies.txt -i
```

Output: 
<pre>
HTTP/1.1 401 Unauthorized
Content-Type: application/json
Date: Sat, 07 Sep 2024 14:39:05 GMT
Content-Length: 29

{"message":"user not found"}
</pre>


2) Creating new user.

```shell
curl --location 'localhost:8080/api/auth/signup' --header 'Content-Type: application/json' --data '{"email":"test@example.com","password":"password123"}' -i
```
Output: 
<pre>
HTTP/1.1 201 Created
Content-Type: application/json
Date: Sat, 07 Sep 2024 15:17:51 GMT
Content-Length: 43

{"message":"User registered successfully"}
</pre>

3) Trying login with wrong password:

```shell
curl --location 'localhost:8080/api/auth/signin' --header 'Content-Type: application/json' --data '{"email":"test@example.com","password":"ppassword123"}' -c cookies.txt -i
```

Output: 
<pre>
HTTP/1.1 401 Unauthorized
Content-Type: application/json
Date: Sat, 07 Sep 2024 15:19:19 GMT
Content-Length: 31

{"message":"invalid password"}
</pre>

4) Trying login with non-existing user:

```shell
curl --location 'localhost:8080/api/auth/signin' --header 'Content-Type: application/json' --data '{"email":"test2@example.com","password":"password123"}' -c cookies.txt -i
```

Output:
<pre>
HTTP/1.1 401 Unauthorized
Content-Type: application/json
Date: Sat, 07 Sep 2024 15:30:15 GMT
Content-Length: 29

{"message":"user not found"}
</pre>

5) Trying login with existing user credentials:

```shell
curl --location 'localhost:8080/api/auth/signin' --header 'Content-Type: application/json' --data '{"email":"test@example.com","password":"password123"}' -c cookies.txt -i
```

Output:
<pre>
HTTP/1.1 200 OK
Content-Type: application/json
Set-Cookie: token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InRlc3RAZXhhbXBsZS5jb20iLCJleHAiOjE3MjU3MjQwMDZ9.lbD7RoJiRSK0K1_eHoeM3hU0y0Q99JfIPGaoHO_FCHY; Path=/api; Expires=Sat, 07 Sep 2024 15:46:46 GMT; HttpOnly
Date: Sat, 07 Sep 2024 15:31:46 GMT
Content-Length: 37

{"message":"Logged in successfully"}
</pre>

A file named cookies.txt created holding the token value.

Save current token in a variable (we will require it later)

```shell
export TOKEN=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InRlc3RAZXhhbXBsZS5jb20iLCJleHAiOjE3MjU3MjQwMDZ9.lbD7RoJiRSK0K1_eHoeM3hU0y0Q99JfIPGaoHO_FCHY 
```


6) Accessing protected resource with this token:

```shell
curl --location --request GET 'localhost:8080/api/protected' -i -b cookies.txt
```

Output: 
<pre>
HTTP/1.1 200 OK
Content-Type: application/json
Date: Sat, 07 Sep 2024 15:33:43 GMT
Content-Length: 70

{"message":"Welcome test@example.com! This is a protected resource."}
</pre>

7) Refreshing token with valid token:
```shell
curl --location --request POST 'localhost:8080/api/auth/refresh' -b cookies.txt -c cookies.txt -i
```
Output:
<pre>
HTTP/1.1 200 OK
Content-Type: application/json
Set-Cookie: token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InRlc3RAZXhhbXBsZS5jb20iLCJleHAiOjE3MjU3MjUzMjN9.JFVbSKqHMbT5K_-65EhUxfuTIdvrMqeZOABU1TPClHM; Path=/api; Expires=Sat, 07 Sep 2024 16:08:43 GMT; HttpOnly
Date: Sat, 07 Sep 2024 15:53:43 GMT
Content-Length: 43

{"message":"Token refreshed successfully"}
</pre>

Save this refreshed token in a variable(we will require it later):
```shell
export REFRESHEDTOKEN=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InRlc3RAZXhhbXBsZS5jb20iLCJleHAiOjE3MjU3MjUzMjN9.JFVbSKqHMbT5K_-65EhUxfuTIdvrMqeZOABU1TPClHM
```
8) Try accessing protected resource with old token:

```shell
curl --location --request GET 'localhost:8080/api/protected' -i -b "token=$TOKEN"
```
Output: 
<pre>
HTTP/1.1 401 Unauthorized
Content-Type: application/json
Date: Sat, 07 Sep 2024 15:59:07 GMT
Content-Length: 37

{"message":"token has been revoked"}
</pre>

9) Try accessing protected resource with NEWTOKEN:

```shell
curl --location --request GET 'localhost:8080/api/protected' -i -b cookies.txt
```
Output: 
<pre>
HTTP/1.1 200 OK
Content-Type: application/json
Date: Sat, 07 Sep 2024 16:00:06 GMT
Content-Length: 70

{"message":"Welcome test@example.com! This is a protected resource."}
</pre>

10) Revoking NEWTOKEN:
```shell    
curl --location --request POST 'localhost:8080/api/auth/revoke' -b cookies.txt -c cookies.txt -i
```
Output: 
<pre>
HTTP/1.1 200 OK
Content-Type: application/json
Set-Cookie: token=; Path=/api; Expires=Sat, 07 Sep 2024 15:01:54 GMT; HttpOnly
Date: Sat, 07 Sep 2024 16:01:54 GMT
Content-Length: 38

{"message":"Logged out successfully"}
</pre>

11) Accessing protected resource with the revoked token:
```shell
curl --location --request GET 'localhost:8080/api/protected' -i -b "token=$REFRESHEDTOKEN"
```
Output:
<pre>
HTTP/1.1 401 Unauthorized
Content-Type: application/json
Date: Sat, 07 Sep 2024 16:05:23 GMT
Content-Length: 37

{"message":"token has been revoked"}
</pre>

12) Revoking token If token already revoked,
```shell
curl --location --request POST 'localhost:8080/api/auth/revoke' -b cookies.txt -c cookies.txt -i
```
Output: 
<pre>
0HTTP/1.1 401 Unauthorized
Content-Type: application/json
Date: Sat, 07 Sep 2024 16:06:29 GMT
Content-Length: 29

{"message":"No token found"}
</pre>

<!---
Note that we can change the response text based on usecases, depending upon sufficient information we need to reveal.
-->

### Resources:
[Postman Collections](https://drive.google.com/file/d/1u0K-0yWBJomyPIO5E0otFgms84xtEtBI/view)
[Demo Video](https://youtu.be/pUp7jIwZha4)

### Way Forward:

- In Database implementation of saving users and invalidated tokens.
- Authentication and Authorization part can be extracted out in a different microservice. Thus, can be exposed as gateway to our backend.
- Latest implementation of jwt-go library can be utilized.  

![](img/mygiphy.gif)