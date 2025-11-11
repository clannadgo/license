package com.example.license;

import com.sun.jna.Pointer;
import com.sun.jna.Native;
import com.fasterxml.jackson.databind.ObjectMapper;

/**
 * 许可证工具类，提供高级API
 * 支持Windows(.dll)、Linux(.so)和macOS(.dylib)
 */
public class LicenseUtils {
    
    private static final ObjectMapper objectMapper = new ObjectMapper();
    
    /**
     * 初始化共享库，确保库文件可以被正确加载
     * @param libraryPath 库文件路径，如果为null则使用系统默认路径
     * @return 是否初始化成功
     */
    public static boolean initialize(String libraryPath) {
        try {
            if (libraryPath != null && !libraryPath.isEmpty()) {
                // 从指定路径加载库
                System.load(libraryPath);
            } else {
                // 从系统路径加载库
                System.loadLibrary(LicenseDLL.getLibraryName());
            }
            return true;
        } catch (UnsatisfiedLinkError e) {
            System.err.println("Failed to load native library: " + e.getMessage());
            return false;
        }
    }
    
    /**
     * 获取当前平台信息
     * @return 平台信息字符串
     */
    public static String getPlatformInfo() {
        String os = System.getProperty("os.name").toLowerCase();
        String arch = System.getProperty("os.arch").toLowerCase();
        String libExt = LicenseDLL.getLibraryExtension();
        String libName = LicenseDLL.getFullLibraryName();
        
        return String.format("OS: %s, Arch: %s, Library: %s", os, arch, libName);
    }
    
    /**
     * 生成机器指纹
     * @return 机器指纹字符串，格式为XXXX-XXXX-XXXX-XXXX
     */
    public static String generateFingerprint() {
        Pointer ptr = LicenseDLL.INSTANCE.GenerateFingerprint();
        if (ptr == null) {
            throw new RuntimeException("Failed to generate fingerprint");
        }
        
        try {
            return ptr.getString(0);
        } finally {
            LicenseDLL.INSTANCE.FreeString(ptr);
        }
    }
    
    /**
     * 验证许可证
     * @param publicKeyPath 公钥文件路径
     * @param licenseContent 许可证内容
     * @return 验证结果
     */
    public static LicenseVerificationResult verifyLicense(String publicKeyPath, String licenseContent) {
        int code = LicenseDLL.INSTANCE.VerifyLicense(publicKeyPath, licenseContent);
        
        LicenseVerificationResult result = new LicenseVerificationResult();
        result.setCode(code);
        
        switch (code) {
            case 0:
                result.setSuccess(true);
                result.setMessage("验证成功");
                break;
            case 1:
                result.setSuccess(false);
                result.setMessage("无效的公钥");
                break;
            case 2:
                result.setSuccess(false);
                result.setMessage("无效的许可证");
                break;
            case 3:
                result.setSuccess(false);
                result.setMessage("许可证已过期");
                break;
            case 4:
                result.setSuccess(false);
                result.setMessage("指纹不匹配");
                break;
            case 5:
                result.setSuccess(false);
                result.setMessage("内部错误");
                break;
            default:
                result.setSuccess(false);
                result.setMessage("未知错误");
                break;
        }
        
        return result;
    }
    
    /**
     * 获取许可证数据
     * @param publicKeyPath 公钥文件路径
     * @param licenseContent 许可证内容
     * @return 许可证数据对象
     */
    public static LicenseData getLicenseData(String publicKeyPath, String licenseContent) {
        Pointer ptr = LicenseDLL.INSTANCE.GetLicenseData(publicKeyPath, licenseContent);
        if (ptr == null) {
            throw new RuntimeException("Failed to get license data");
        }
        
        try {
            String jsonStr = ptr.getString(0);
            return objectMapper.readValue(jsonStr, LicenseData.class);
        } catch (Exception e) {
            throw new RuntimeException("Failed to parse license data", e);
        } finally {
            LicenseDLL.INSTANCE.FreeString(ptr);
        }
    }
    
    /**
     * 许可证验证结果
     */
    public static class LicenseVerificationResult {
        private int code;
        private boolean success;
        private String message;
        
        public int getCode() {
            return code;
        }
        
        public void setCode(int code) {
            this.code = code;
        }
        
        public boolean isSuccess() {
            return success;
        }
        
        public void setSuccess(boolean success) {
            this.success = success;
        }
        
        public String getMessage() {
            return message;
        }
        
        public void setMessage(String message) {
            this.message = message;
        }
        
        @Override
        public String toString() {
            return "LicenseVerificationResult{" +
                    "code=" + code +
                    ", success=" + success +
                    ", message='" + message + '\'' +
                    '}';
        }
    }
    
    /**
     * 许可证数据
     */
    public static class LicenseData {
        private String issuer;
        private String customer;
        private String fingerprint;
        private long issuedAt;
        private long expiresAt;
        private String error;
        
        public String getIssuer() {
            return issuer;
        }
        
        public void setIssuer(String issuer) {
            this.issuer = issuer;
        }
        
        public String getCustomer() {
            return customer;
        }
        
        public void setCustomer(String customer) {
            this.customer = customer;
        }
        
        public String getFingerprint() {
            return fingerprint;
        }
        
        public void setFingerprint(String fingerprint) {
            this.fingerprint = fingerprint;
        }
        
        public long getIssuedAt() {
            return issuedAt;
        }
        
        public void setIssuedAt(long issuedAt) {
            this.issuedAt = issuedAt;
        }
        
        public long getExpiresAt() {
            return expiresAt;
        }
        
        public void setExpiresAt(long expiresAt) {
            this.expiresAt = expiresAt;
        }
        
        public String getError() {
            return error;
        }
        
        public void setError(String error) {
            this.error = error;
        }
        
        @Override
        public String toString() {
            return "LicenseData{" +
                    "issuer='" + issuer + '\'' +
                    ", customer='" + customer + '\'' +
                    ", fingerprint='" + fingerprint + '\'' +
                    ", issuedAt=" + issuedAt +
                    ", expiresAt=" + expiresAt +
                    ", error='" + error + '\'' +
                    '}';
        }
    }
}