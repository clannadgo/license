"""
Flask License Middleware

This middleware integrates license verification into Flask applications,
similar to the Go implementation in examples/go/license_middleware.go.
"""

import os
import sys
from typing import Optional, Dict, Any
from flask import Flask, request, jsonify, g

# Add parent directory to path to allow importing license_dll
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

# Import only the core functionality from license_dll
from license_dll.license_dll import LicenseUtils, LicenseVerificationResult

# Default paths
DEFAULT_PUBLIC_KEY_PATH = os.path.join(os.path.dirname(__file__), "..", "..", "keys", "public.pem")
DEFAULT_LICENSE_PATH = os.path.join(os.path.dirname(__file__), "license.json")

def generate_fingerprint() -> str:
    """
    Generate the machine fingerprint using the Go DLL.
    This ensures complete consistency with the Go implementation.
    
    Returns:
        str: The machine fingerprint
    """
    license_utils = LicenseUtils()
    return license_utils.generate_fingerprint()


class LicenseMiddleware:
    """
    Flask middleware for license verification.
    """
    
    def __init__(
        self,
        app: Optional[Flask] = None,
        public_key_path: str = "./public.pem",
        license_file_path: str = "./license.lic",
        skip_paths: Optional[list] = None
    ):
        """
        Initialize the license middleware.
        
        Args:
            app: Flask application instance
            public_key_path: Path to the public key file
            license_file_path: Path to the license file
            skip_paths: List of paths to skip license verification
        """
        self.public_key_path = public_key_path
        self.license_file_path = license_file_path
        self.skip_paths = skip_paths or ["/license/code", "/health"]
        
        # Initialize LicenseUtils once
        self.license_utils = LicenseUtils()
        
        if app is not None:
            self.init_app(app)
    
    def init_app(self, app: Flask):
        """
        Initialize the middleware with the Flask application.
        
        Args:
            app: Flask application instance
        """
        app.before_request(self._before_request)
    
    def _before_request(self):
        """
        Middleware function that runs before each request.
        """
        # Skip specified paths
        if request.path in self.skip_paths:
            return
        
        try:
            # Read license file
            license_content = self.license_utils.read_license_file(self.license_file_path)
            
            # Verify license
            result = self.license_utils.verify_license(self.public_key_path, license_content)
            if not result.success:
                return jsonify({"error": f"许可证验证失败: {result.message}"}), 401
            
            # Get license data for additional checks
            license_data = self.license_utils.get_license_data(self.public_key_path, license_content)
            if not license_data:
                return jsonify({"error": "无法解析许可证数据"}), 401
            
            # Check if license is expired
            if self.license_utils.is_license_expired(self.public_key_path, license_content):
                return jsonify({
                    "error": "许可证已过期",
                    "valid_until": license_data.expires_at
                }), 401
            
            # Store license info in Flask's g object for later use
            g.license = {
                "customer": license_data.customer,
                "expires_at": license_data.expires_at,
                "fingerprint": license_data.fingerprint
            }
            
        except FileNotFoundError as e:
            return jsonify({"error": f"许可证文件未找到: {str(e)}"}), 401
        except Exception as e:
            return jsonify({"error": f"许可证验证错误: {str(e)}"}), 401


def setup_license_routes(app: Flask, license_file_path: str = "./license.lic"):
    """
    Setup license-related routes for the Flask application.
    
    Args:
        app: Flask application instance
        license_file_path: Path to the license file
    """
    license_utils = LicenseUtils()
    
    @app.route("/license/code")
    def get_license_code():
        """Get the current machine's fingerprint."""
        try:
            fingerprint = license_utils.generate_fingerprint()
            return jsonify({"fingerprint": fingerprint})
        except Exception as e:
            return jsonify({"error": f"无法生成指纹: {str(e)}"}), 500
    
    @app.route("/health")
    def health_check():
        """Health check endpoint."""
        return jsonify({"status": "ok"})
    
    @app.route("/license/status")
    def get_license_status():
        """Get the current license status."""
        try:
            # This endpoint should be protected by the middleware
            # So if we reach here, the license is valid
            return jsonify({"status": "valid", "message": "License is valid"})
        except Exception as e:
            return jsonify({"error": f"无法获取许可证状态: {str(e)}"}), 500


def create_example_app():
    """
    Create an example Flask application with license middleware.
    
    Returns:
        Flask application instance
    """
    app = Flask(__name__)
    app.config['JSON_AS_ASCII'] = False  # Support Chinese characters
    
    # Add license middleware
    LicenseMiddleware(app)
    
    # Setup license routes
    setup_license_routes(app)
    
    # Example protected route
    @app.route("/api/protected")
    def protected_route():
        """Example protected route."""
        license_info = g.license
        return jsonify({
            "message": "This is a protected route",
            "customer": license_info["customer"],
            "license_expires": license_info["expires_at"]
        })
    
    return app


if __name__ == "__main__":
    app = create_example_app()
    app.run(host="0.0.0.0", port=5000, debug=True)