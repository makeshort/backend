# Rest API Server for URL Shortener App

### Authorization:

Authorization is performed by the `SessionID` parameter in the request header

### Endpoints:

---

### Log in (create a session): **POST /api/session**

**Body:**

| Field    | Type   | Required |
|:---------|:-------|:---------|
| email    | string | Yes      |
| password | string | Yes      |

**Success response:** `201 Created`

| Field      | Type   |
|:-----------|:-------|
| session_id | string |

**Possible errors:**

| Code | Description                                                                      |
|:-----|:---------------------------------------------------------------------------------|
| 400  | Bad request. Missing required fields. User with this credentials already exists. |

---

### Log out (close a session): **DELETE /api/session**

**Success response:** `200 OK`

**Possible errors:**

| Code | Description  |
|:-----|:-------------|
| 401  | Unauthorized |

---

### Create URL: **POST /api/url**

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

### Delete URL: **DELETE /api/url/:alias**

**Success response:** `200 OK`

**Possible errors:**

| Code | Description                              |
|:-----|:-----------------------------------------|
| 403  | Forbidden. You are not owner of this URL |
| 401  | Unauthorized                             |

---

### Create user: **POST /api/user**

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

| Code | Description                          |
|:-----|:-------------------------------------|
| 400  | Bad request. Missing required fields |

### Dleete me: **DELETE /api/user/me**

**Success response:** `200 OK`

**Possible errors:**

| Code | Description                             |
|:-----|:----------------------------------------|
| 400  | Bad request. User not found in database |
| 401  | Unauthorized                            |

### Get my URLs: **GET /api/user/me/urls**

**Success response:** `200 OK`

**Possible errors:**

| Code | Description  |
|:-----|:-------------|
| 401  | Unauthorized |