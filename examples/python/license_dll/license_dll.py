"""
License DLL Python SDK

Python SDK for calling the license shared library generated from Golang.
Supports Windows (.dll), Linux (.so) and macOS (.dylib).
"""

import ctypes
import os
import sys
import platform
from typing import Optional, Dict, Any, Tuple
import json


class LicenseDLL:
    """
    Python wrapper for the license shared library generated from Golang.
    Supports Windows (.dll), Linux (.so) and macOS (.dylib).
    """
    
    def __init__(self, library_path: Optional[str] = None):
        """
        Initialize the LicenseDLL wrapper.
        
        Args:
            library_path: Path to the license shared library file. 
                         If None, will try to find it based on the platform.
        """
        if library_path is None:
            # Try to find the library in the current directory
            current_dir = os.path.dirname(os.path.abspath(__file__))
            library_name = self._get_library_name()
            library_path = os.path.join(current_dir, library_name)
            
            # If not found, try the parent directory
            if not os.path.exists(library_path):
                parent_dir = os.path.dirname(current_dir)
                library_path = os.path.join(parent_dir, library_name)
                
            # If still not found, try the examples/dll directory
            if not os.path.exists(library_path):
                project_root = os.path.dirname(os.path.dirname(os.path.dirname(parent_dir)))
                library_path = os.path.join(project_root, "examples", "dll", library_name)
        
        if not os.path.exists(library_path):
            raise FileNotFoundError(f"License library not found at {library_path}")
        
        self.dll = ctypes.CDLL(library_path)
        self._setup_function_prototypes()
    
    def _get_library_name(self) -> str:
        """
        Get the library name based on the platform.
        
        Returns:
            Library name with appropriate extension.
        """
        system = platform.system().lower()
        if system == "windows":
            return "license.dll"
        elif system == "linux":
            return "license.so"
        elif system == "darwin":  # macOS
            return "license.dylib"
        else:
            raise RuntimeError(f"Unsupported platform: {system}")
    
    def get_platform_info(self) -> str:
        """
        Get information about the current platform.
        
        Returns:
            Platform information string.
        """
        system = platform.system()
        machine = platform.machine()
        library_name = self._get_library_name()
        
        return f"OS: {system}, Arch: {machine}, Library: {library_name}"
    
    def _setup_function_prototypes(self):
        """Set up function prototypes for the library functions."""
        # GenerateFingerprint
        self.dll.GenerateFingerprint.argtypes = []
        self.dll.GenerateFingerprint.restype = ctypes.c_char_p
        
        # VerifyLicense
        self.dll.VerifyLicense.argtypes = [ctypes.c_char_p, ctypes.c_char_p]
        self.dll.VerifyLicense.restype = ctypes.c_int
        
        # GetLicenseData
        self.dll.GetLicenseData.argtypes = [ctypes.c_char_p, ctypes.c_char_p]
        self.dll.GetLicenseData.restype = ctypes.c_char_p
        
        # FreeString
        self.dll.FreeString.argtypes = [ctypes.c_char_p]
        self.dll.FreeString.restype = None
    
    def generate_fingerprint(self) -> str:
        """
        Generate machine fingerprint.
        
        Returns:
            Machine fingerprint as a string.
        """
        result = self.dll.GenerateFingerprint()
        if result is None:
            raise RuntimeError("Failed to generate fingerprint")
        
        fingerprint = result.decode('utf-8')
        self.dll.FreeString(result)
        return fingerprint
    
    def verify_license(self, public_key_path: str, license_content: str) -> Tuple[int, str]:
        """
        Verify license.
        
        Args:
            public_key_path: Path to the public key file.
            license_content: License content as a string.
            
        Returns:
            Tuple of (result_code, message):
            - result_code: 0 for success, non-zero for failure
            - message: Result message
        """
        public_key_path_bytes = public_key_path.encode('utf-8')
        license_content_bytes = license_content.encode('utf-8')
        
        result_code = self.dll.VerifyLicense(public_key_path_bytes, license_content_bytes)
        
        # Map result codes to messages
        messages = {
            0: "Success",
            1: "Invalid public key",
            2: "Invalid license",
            3: "License expired",
            4: "Fingerprint mismatch",
            5: "Internal error"
        }
        
        message = messages.get(result_code, f"Unknown error code: {result_code}")
        return result_code, message
    
    def get_license_data(self, public_key_path: str, license_content: str) -> Optional[Dict[str, Any]]:
        """
        Get license data.
        
        Args:
            public_key_path: Path to the public key file.
            license_content: License content as a string.
            
        Returns:
            License data as a dictionary, or None if failed.
        """
        public_key_path_bytes = public_key_path.encode('utf-8')
        license_content_bytes = license_content.encode('utf-8')
        
        result = self.dll.GetLicenseData(public_key_path_bytes, license_content_bytes)
        if result is None:
            return None
        
        try:
            license_data_json = result.decode('utf-8')
            license_data = json.loads(license_data_json)
            return license_data
        except (UnicodeDecodeError, json.JSONDecodeError):
            return None
        finally:
            self.dll.FreeString(result)


class LicenseVerificationResult:
    """
    Result of license verification.
    """
    
    def __init__(self, result_code: int, message: str):
        self.result_code = result_code
        self.message = message
        self.success = result_code == 0
    
    def __str__(self):
        return f"LicenseVerificationResult(success={self.success}, code={self.result_code}, message='{self.message}')"


class LicenseData:
    """
    License data parsed from the license file.
    """
    
    def __init__(self, data: Dict[str, Any]):
        self.customer = data.get("customer", "")
        self.issuer = data.get("issuer", "")
        self.fingerprint = data.get("fingerprint", "")
        self.expires_at = data.get("expires_at", 0)
        self.issued_at = data.get("issued_at", 0)
        self.raw_data = data
    
    def __str__(self):
        return f"LicenseData(customer='{self.customer}', issuer='{self.issuer}', fingerprint='{self.fingerprint}', expires_at={self.expires_at})"


class LicenseUtils:
    """
    Utility class for license operations.
    Supports Windows (.dll), Linux (.so) and macOS (.dylib).
    """
    
    def __init__(self, library_path: Optional[str] = None):
        self.license_dll = LicenseDLL(library_path)
    
    def get_platform_info(self) -> str:
        """
        Get information about the current platform.
        
        Returns:
            Platform information string.
        """
        return self.license_dll.get_platform_info()
    
    def generate_fingerprint(self) -> str:
        """
        Generate machine fingerprint.
        
        Returns:
            Machine fingerprint as a string.
        """
        return self.license_dll.generate_fingerprint()
    
    def verify_license(self, public_key_path: str, license_content: str) -> LicenseVerificationResult:
        """
        Verify license.
        
        Args:
            public_key_path: Path to the public key file.
            license_content: License content as a string.
            
        Returns:
            LicenseVerificationResult object.
        """
        result_code, message = self.license_dll.verify_license(public_key_path, license_content)
        return LicenseVerificationResult(result_code, message)
    
    def get_license_data(self, public_key_path: str, license_content: str) -> Optional[LicenseData]:
        """
        Get license data.
        
        Args:
            public_key_path: Path to the public key file.
            license_content: License content as a string.
            
        Returns:
            LicenseData object or None if failed.
        """
        data_dict = self.license_dll.get_license_data(public_key_path, license_content)
        if data_dict is None:
            return None
        
        return LicenseData(data_dict)
    
    def read_license_file(self, license_path: str) -> str:
        """
        Read license file content.
        
        Args:
            license_path: Path to the license file.
            
        Returns:
            License file content as a string.
        """
        with open(license_path, 'r', encoding='utf-8') as f:
            return f.read()
    
    def is_license_expired(self, public_key_path: str, license_content: str) -> bool:
        """
        Check if license is expired.
        
        Args:
            public_key_path: Path to the public key file.
            license_content: License content as a string.
            
        Returns:
            True if license is expired, False otherwise.
        """
        license_data = self.get_license_data(public_key_path, license_content)
        if license_data is None:
            return True
        
        import time
        return time.time() > license_data.expires_at