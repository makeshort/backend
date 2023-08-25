# Rest API Server for URL Shortener App

## Authorization:

Authorization is performed by the `AccessToken` in `Authorization` header. Access token issues for 30 minutes, and refreshs by `RefreshToken` in cookies. RefreshToken issues for 30 days. On logout refresh token adds to blacklist, and access token will never updated with this refresh token.


## Data structures:

#### User:

| Field    | Type   | Description          |
|:---------|:-------|:---------------------|
| id       | string | The ID of user       |
| username | string | The username of user |
| email    | string | The email of user    |

#### URL:

| Field     | Type   | Description            |
|:----------|:-------|:-----------------------|
| id        | string | The ID of url          |
| alias     | string | The short alias of url |
| url       | string | The original url       |
| redirects | int    | The redirects counter  |

#### Token pair:

| Field         | Type   | Description       |
|:--------------|:-------|:------------------|
| access_token  | string | The access token  |
| refresh_token | string | The refresh token |



## Endpoints:

#### **POST** `/api/auth/session` - login (create a session)

**Body:**

| Field    | Type   | Required |
|:---------|:-------|:---------|
| email    | string | Yes      |
| password | string | Yes      |

**Success response:** `200 OK`

| Field         | Type   |
|:--------------|:-------|
| access_token  | string |
| refresh_token | string |

**Possible errors:**

| Code | Description                                                                      |
|:-----|:---------------------------------------------------------------------------------|
| 400  | Bad request. Missing required fields. User with this credentials already exists. |

---

#### **DELETE** `/api/auth/session` - logout (close a session): 

**Success response:** `200 OK`

**Possible errors:**

| Code | Description  |
|:-----|:-------------|
| 401  | Unauthorized |

---

#### **POST** `/api/auth/signup` - registration (create user)

**Body:**

| Field    | Type   | Required |
|:---------|:-------|:---------|
| email    | string | Yes      |
| username | string | Yes      |
| password | string | Yes      | 

**Success response:** `201 Created`

| Field    | Type   |
|:---------|:-------|
| email    | string |
| username | string |


**Possible errors:**

| Code | Description                                     |
|:-----|:------------------------------------------------|
| 400  | Bad request. Missing required fields            |
| 409  | User with this email or username already exists |

---

#### **POST** `/api/auth/refresh` - refresh tokens

**Body:**

| Field    | Type   | Required |
|:---------|:-------|:---------|
| token    | string | Yes      |

**Success response:** `200 OK`

| Field         | Type   |
|:--------------|:-------|
| access_token  | string |
| refresh_token | string |


**Possible errors:**

| Code | Description           |
|:-----|:----------------------|
| 403  | Invalid refresh token |

#### **DELETE** `/api/user/me` - delete me

**Success response:** `200 OK`

**Possible errors:**

| Code | Description                             |
|:-----|:----------------------------------------|
| 400  | Bad request. User not found in database |
| 401  | Unauthorized                            |

#### **GET** `/api/user/me/urls` - get my URLs

**Success response:** `200 OK`

Array of URL entities:

| Field     | Type   |
|:----------|:-------|
| url       | string |
| alias     | string |
| redirects | int    |

**Possible errors:**

| Code | Description  |
|:-----|:-------------|
| 401  | Unauthorized |

---

#### **POST** `/api/url` - create URL

**Request body:**

| Field | Type   | Required |
|:------|:-------|:---------|
| url   | string | Yes      |
| alias | string | No       |

**Success response:** `201 Created`

| Field | Type   |
|:------|:-------|
| url   | string |
| alias | string |

**Possible errors:**

| Code | Description                          |
|:-----|:-------------------------------------|
| 400  | Bad request. Missing required fields |
| 401  | Unauthorized                         |
| 409  | URL with this alias already exists   |

---

#### **DELETE** `/api/url/:alias` - delete URL

**Success response:** `200 OK`

**Possible errors:**

| Code | Description                              |
|:-----|:-----------------------------------------|
| 401  | Unauthorized                             |
| 403  | Forbidden. You are not owner of this URL |
| 404  | URL to delete not found                  |
