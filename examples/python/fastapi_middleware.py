# FastAPI License Middleware Implementation
# Based on Go implementation in license_middleware.go
# This middleware provides license validation for FastAPI applications

from fastapi import FastAPI, Request, HTTPException, Depends
from fastapi.responses import JSONResponse
import os
import sys
import time
from typing import Optional, Callable, Dict, Any

# Add parent directory to path to allow importing license_dll
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

# Import only the core functionality from license_dll
from license_dll import LicenseUtils, LicenseVerificationResult

# Default paths - these should be configured based on your application's needs
DEFAULT_PUBLIC_KEY_PATH = os.path.join(os.path.dirname(__file__), "..", "..", "keys", "public.pem")
DEFAULT_LICENSE_PATH = os.path.join(os.path.dirname(__file__), "license.json")

def generate_fingerprint() -> str:
    """
    Generate the machine fingerprint using the LicenseUtils from Go DLL.
    This ensures complete consistency with the Go implementation.
    
    Returns:
        str: The machine fingerprint
    """
    license_utils = LicenseUtils()
    return license_utils.generate_fingerprint()


def create_license_middleware(
    public_key_path: str = "./public.pem",
    license_file_path: str = "./license.lic",
    skip_paths: Optional[list] = None
):
    """
    Create a FastAPI license middleware.
    
    Args:
        public_key_path: Path to the public key file
        license_file_path: Path to the license file
        skip_paths: List of paths to skip license verification
        
    Returns:
        A FastAPI middleware function
    """
    if skip_paths is None:
        skip_paths = ["/license/code", "/health"]
    
    # Initialize LicenseUtils once
    license_utils = LicenseUtils()
    
    async def license_middleware(request: Request, call_next):
        # Skip specified paths
        if request.url.path in skip_paths:
            return await call_next(request)
        
        try:
            # Read license file
            license_content = license_utils.read_license_file(license_file_path)
            
            # 使用合并后的get_license_data函数进行验证和获取数据
            license_data = license_utils.get_license_data(public_key_path, license_content)
            if not license_data:
                return JSONResponse(
                    status_code=401,
                    content={"error": "无法获取许可证数据"}
                )
            
            # 检查验证结果
            if isinstance(license_data, dict) and "valid" in license_data:
                # 新的数据结构：包含valid和data字段
                if not license_data.get("valid", False):
                    error_msg = license_data.get("error", "许可证验证失败")
                    return JSONResponse(
                        status_code=401,
                        content={"error": f"许可证验证失败: {error_msg}"}
                    )
                
                # 验证成功，获取许可证数据
                license_info = license_data.get("data", {})
            else:
                # 旧的数据结构：直接返回LicenseData对象
                license_info = license_data
            
            # 检查许可证是否过期
            if license_utils.is_license_expired(public_key_path, license_content):
                return JSONResponse(
                    status_code=401,
                    content={
                        "error": "许可证已过期",
                        "valid_until": license_info.expires_at if hasattr(license_info, 'expires_at') else "unknown"
                    }
                )
            
            # Store license info in request state for later use
            request.state.license = {
                "customer": license_info.customer if hasattr(license_info, 'customer') else "",
                "expires_at": license_info.expires_at if hasattr(license_info, 'expires_at') else 0,
                "fingerprint": license_info.fingerprint if hasattr(license_info, 'fingerprint') else ""
            }
            
            # Continue to the next handler
            response = await call_next(request)
            return response
            
        except FileNotFoundError as e:
            return JSONResponse(
                status_code=401,
                content={"error": f"许可证文件未找到: {str(e)}"}
            )
        except Exception as e:
            return JSONResponse(
                status_code=401,
                content={"error": f"许可证验证错误: {str(e)}"}
            )
    
    return license_middleware


def setup_license_api(app: FastAPI, license_file_path: str = "./license.lic"):
    """
    Setup license-related API endpoints.
    
    Args:
        app: FastAPI application instance
        license_file_path: Path to the license file
    """
    license_utils = LicenseUtils()
    
    @app.get("/license/code")
    async def get_license_code():
        """Get the current machine's fingerprint."""
        try:
            fingerprint = license_utils.generate_fingerprint()
            return {"fingerprint": fingerprint}
        except Exception as e:
            raise HTTPException(status_code=500, detail=f"无法生成指纹: {str(e)}")
    
    @app.get("/health")
    async def health_check():
        """Health check endpoint."""
        return {"status": "ok"}
    
    @app.get("/license/status")
    async def get_license_status():
        """Get the current license status."""
        try:
            # This endpoint should be protected by the middleware
            # So if we reach here, the license is valid
            return {"status": "valid", "message": "License is valid"}
        except Exception as e:
            raise HTTPException(status_code=500, detail=f"无法获取许可证状态: {str(e)}")


def create_example_app():
    """
    Create an example FastAPI application with license middleware.
    
    Returns:
        FastAPI application instance
    """
    app = FastAPI(title="License Protected API", version="1.0.0")
    
    # Add license middleware
    app.middleware("http")(create_license_middleware())
    
    # Setup license endpoints
    setup_license_api(app)
    
    # Example protected endpoint
    @app.get("/api/protected")
    async def protected_endpoint(request: Request):
        """Example protected endpoint."""
        license_info = request.state.license
        return {
            "message": "This is a protected endpoint",
            "customer": license_info["customer"],
            "license_expires": license_info["expires_at"]
        }
    
    return app


if __name__ == "__main__":
    import uvicorn
    app = create_example_app()
    uvicorn.run(app, host="0.0.0.0", port=8000)