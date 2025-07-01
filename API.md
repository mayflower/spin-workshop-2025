# Image Repository API Documentation

This document provides a concise reference for all HTTP endpoints in the Image Repository service.

## Overview

The Image Repository is a microservice-based system for storing, transforming, and serving images. It consists of several components:

- **Originals Service**: Manages original image uploads and retrieval
- **Derived Service**: Handles transformed image retrieval and caching
- **Transformer Service**: Performs image transformations
- **Init Service**: Initializes the database schema
- **Maintenance Service**: Handles cleanup of old derived images

## Endpoints

### Originals Service

#### `GET /originals`

Returns metadata for all original images.

**Response:**
- `200 OK`: JSON object containing an array of image metadata
  ```json
  {
    "originals": [
      {
        "name": "example",
        "created": 1656789012,
        "length": 12345
      }
    ]
  }
  ```

#### `GET /originals/meta`

Alias for `GET /originals`.

#### `GET /originals/meta/:name`

Returns metadata for a specific original image.

**Parameters:**
- `name`: Image identifier

**Response:**
- `200 OK`: JSON object containing image metadata
  ```json
  {
    "name": "example",
    "created": 1656789012,
    "length": 12345
  }
  ```
- `404 Not Found`: Image not found
  ```json
  {
    "reason": "Not found"
  }
  ```

#### `POST /originals/:name`

Uploads a new original image.

**Parameters:**
- `name`: Image identifier (must contain only letters, digits, and `.`, `-`, `_`)

**Request Body:**
- Binary image data (PNG or JPEG)

**Response:**
- `200 OK`: JSON object containing image metadata
  ```json
  {
    "name": "example",
    "created": 1656789012,
    "length": 12345
  }
  ```
- `400 Bad Request`: Invalid name or image format
  ```json
  {
    "reason": "invalid name: only letters, digits, and . - _ are allowed"
  }
  ```
- `409 Conflict`: Image with this name already exists
  ```json
  {
    "reason": "duplicate"
  }
  ```

#### `GET /originals/:name`

Retrieves an original image.

**Parameters:**
- `name`: Image identifier

**Response:**
- `200 OK`: Binary image data with `Content-Type: image/png`
- `404 Not Found`: Image not found

#### `DELETE /originals/:name`

Deletes an original image and all its derived versions.

**Parameters:**
- `name`: Image identifier

**Response:**
- `200 OK`: Empty JSON object
  ```json
  {}
  ```
- `404 Not Found`: Image not found
  ```json
  {
    "reason": "Not found"
  }
  ```

### Derived Service

#### `GET /derived/:name`

Retrieves a transformed version of an image. If the transformed version doesn't exist, it will be created and cached.

**Parameters:**
- `name`: Original image identifier

**Query Parameters:**
- `width` (optional): Desired width (1-50000)
- `height` (optional): Desired height (1-50000)
- `type` (optional): Output format (`image/png` or `image/jpeg`, default: `image/png`)

**Response:**
- `200 OK`: Binary image data with appropriate `Content-Type`
- `400 Bad Request`: Invalid parameters
  ```json
  {
    "reason": "must be a positive integer between 1 and 50000"
  }
  ```
- `404 Not Found`: Original image not found
- `500 Internal Server Error`: Error during transformation

### Transformer Service

#### `POST /transform`

Transforms an image according to the specified parameters.

**Query Parameters:**
- `width` (optional): Desired width (positive integer)
- `height` (optional): Desired height (positive integer)
- `type` (optional): Output format (`image/png` or `image/jpeg`, default: `image/png`)

**Request Body:**
- Binary image data (PNG or JPEG)

**Response:**
- `200 OK`: Transformed binary image data with appropriate `Content-Type`
- `400 Bad Request`: Invalid parameters or image format
  ```json
  {
    "reason": "Invalid width parameter: value must be positive"
  }
  ```

### Init Service

#### `GET /init`

Initializes the database schema.

**Response:**
- `200 OK`: Text message "init complete"

## Error Responses

All error responses follow this format:
```json
{
  "reason": "Error message describing the issue"
}
```

## Notes

- If only one dimension (width or height) is specified, the other will be calculated to maintain the aspect ratio.
- Derived images are cached for faster retrieval.
- Derived images that haven't been accessed for a configurable period will be automatically deleted by the maintenance service.
- The maximum allowed image dimension is 50000 pixels.