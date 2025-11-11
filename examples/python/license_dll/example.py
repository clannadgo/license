#!/usr/bin/env python3
"""
Example usage of the License DLL Python SDK.
"""

import os
import sys
import time
from datetime import datetime

# Add the current directory to the path so we can import the license_dll module
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from license_dll import LicenseUtils


def main():
    """
    Main function demonstrating the usage of the License DLL Python SDK.
    """
    print("License DLL Python SDK Example")
    print("=" * 40)
    
    # Initialize the LicenseUtils
    try:
        license_utils = LicenseUtils()
        print("✓ License DLL loaded successfully")
    except Exception as e:
        print(f"✗ Failed to load License DLL: {e}")
        return
    
    # Generate machine fingerprint
    try:
        fingerprint = license_utils.generate_fingerprint()
        print(f"✓ Machine fingerprint: {fingerprint}")
    except Exception as e:
        print(f"✗ Failed to generate fingerprint: {e}")
        return
    
    # Get paths to public key and license files
    project_root = os.path.dirname(os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__)))))
    public_key_path = os.path.join(project_root, "public.pem")
    license_path = os.path.join(project_root, "license.lic")
    
    # Check if files exist
    if not os.path.exists(public_key_path):
        print(f"✗ Public key file not found at: {public_key_path}")
        return
    
    if not os.path.exists(license_path):
        print(f"✗ License file not found at: {license_path}")
        return
    
    # Read license file
    try:
        license_content = license_utils.read_license_file(license_path)
        print(f"✓ License file read successfully from: {license_path}")
    except Exception as e:
        print(f"✗ Failed to read license file: {e}")
        return
    
    # Verify license
    try:
        result = license_utils.verify_license(public_key_path, license_content)
        if result.success:
            print(f"✓ License verification successful: {result.message}")
        else:
            print(f"✗ License verification failed: {result.message}")
            return
    except Exception as e:
        print(f"✗ Failed to verify license: {e}")
        return
    
    # Get license data
    try:
        license_data = license_utils.get_license_data(public_key_path, license_content)
        if license_data:
            print(f"✓ License data retrieved successfully:")
            print(f"  - Customer: {license_data.customer}")
            print(f"  - Issuer: {license_data.issuer}")
            print(f"  - Fingerprint: {license_data.fingerprint}")
            print(f"  - Issued At: {datetime.fromtimestamp(license_data.issued_at)}")
            print(f"  - Expires At: {datetime.fromtimestamp(license_data.expires_at)}")
            
            # Check if license is expired
            is_expired = license_utils.is_license_expired(public_key_path, license_content)
            if is_expired:
                print("  - Status: EXPIRED")
            else:
                print("  - Status: VALID")
        else:
            print("✗ Failed to retrieve license data")
    except Exception as e:
        print(f"✗ Failed to get license data: {e}")
        return
    
    print("\n" + "=" * 40)
    print("Example completed successfully!")


if __name__ == "__main__":
    main()