"""
License DLL Python SDK

Python SDK for calling the license.dll generated from Golang.
"""

from .license_dll import LicenseDLL, LicenseUtils, LicenseVerificationResult, LicenseData

__version__ = "1.0.0"
__all__ = ["LicenseDLL", "LicenseUtils", "LicenseVerificationResult", "LicenseData"]