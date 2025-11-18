# Python License Middleware

This directory contains Python middleware implementations for license verification in FastAPI and Flask applications, similar to the Go implementation in `examples/go/license_middleware.go`.

## Components

1. **`license_dll.py`** - Python wrapper for the license shared library (DLL/SO/DYLIB) generated from Golang
2. **`fastapi_middleware.py`** - License verification middleware for FastAPI applications
3. **`flask_middleware.py`** - License verification middleware for Flask applications
4. **`example.py`** - Example usage of the license_dll module

## FastAPI Middleware

### Usage

```python
from fastapi import FastAPI
from fastapi_middleware import create_license_middleware, setup_license_api

# Create FastAPI application
app = FastAPI()

# Add license middleware
app.middleware("http")(create_license_middleware(
    public_key_path="./public.pem",
    license_file_path="./license.lic",
    skip_paths=["/license/code", "/health", "/api/public"]
))

# Setup license-related endpoints
setup_license_api(app)

# Add your protected endpoints
@app.get("/api/protected")
async def protected_endpoint(request):
    license_info = request.state.license
    return {
        "message": "This endpoint is protected by license",
        "customer": license_info["customer"]
    }
```

### Key Features
- Automatic license verification for all requests (except skipped paths)
- Checks license validity, expiration, and machine fingerprint
- Provides license code endpoint to get the current machine's fingerprint
- Stores license information in the request state for use in endpoints

## Flask Middleware

### Usage

```python
from flask import Flask
from flask_middleware import LicenseMiddleware, setup_license_routes

# Create Flask application
app = Flask(__name__)
app.config['JSON_AS_ASCII'] = False  # Support Chinese characters

# Add license middleware
LicenseMiddleware(
    app,
    public_key_path="./public.pem",
    license_file_path="./license.lic",
    skip_paths=["/license/code", "/health", "/api/public"]
)

# Setup license-related routes
setup_license_routes(app)

# Add your protected routes
@app.route("/api/protected")
def protected_route():
    from flask import g
    license_info = g.license
    return {
        "message": "This route is protected by license",
        "customer": license_info["customer"]
    }
```

### Key Features
- Automatic license verification for all requests (except skipped paths)
- Checks license validity, expiration, and machine fingerprint
- Provides license code endpoint to get the current machine's fingerprint
- Stores license information in Flask's g object for use in routes

## License DLL Module

The `license_dll` module provides a Python interface to the license shared library generated from Golang. It handles:

- Machine fingerprint generation
- License verification using public key
- License data extraction
- License expiration checking

### Example Usage

```python
from license_dll import LicenseUtils

# Initialize license utilities
license_utils = LicenseUtils()

# Generate machine fingerprint
fingerprint = license_utils.generate_fingerprint()
print(f"Machine fingerprint: {fingerprint}")

# Verify license
result = license_utils.verify_license("./public.pem", license_content)
if result.success:
    print("License is valid")
else:
    print(f"License invalid: {result.message}")

# Get license data
license_data = license_utils.get_license_data("./public.pem", license_content)
if license_data:
    print(f"Customer: {license_data.customer}")
    print(f"Expires at: {license_data.expires_at}")
```

## Requirements

- Python 3.7+
- For FastAPI: `fastapi`, `uvicorn`
- For Flask: `flask`
- The license shared library (DLL/SO/DYLIB) generated from Golang

## Running Examples

### FastAPI Example

```bash
python fastapi_middleware.py
```

### Flask Example

```bash
python flask_middleware.py
```

### License DLL Example

```bash
python license_dll/example.py
```

## Notes

- Make sure the license shared library is available in one of the expected locations
- The public key and license files must be accessible to the application
- The middleware logic is designed to be consistent with the Go implementation in `examples/go/license_middleware.go`
