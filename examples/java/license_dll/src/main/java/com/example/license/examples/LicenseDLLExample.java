package com.example.license.examples;

import com.example.license.LicenseUtils;
import com.example.license.LicenseUtils.LicenseData;
import com.example.license.LicenseUtils.LicenseVerificationResult;

import java.io.File;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Paths;

/**
 * Java调用DLL示例
 */
public class LicenseDLLExample {
    
    public static void main(String[] args) {
        // 设置DLL路径（如果不在系统路径中）
        // System.setProperty("jna.library.path", "path/to/dll");
        
        try {
            // 示例1: 生成机器指纹
            System.out.println("=== 生成机器指纹 ===");
            String fingerprint = LicenseUtils.generateFingerprint();
            System.out.println("机器指纹: " + fingerprint);
            
            // 示例2: 验证许可证
            System.out.println("\n=== 验证许可证 ===");
            
            // 公钥文件路径（假设在项目根目录）
            String publicKeyPath = "public.pem";
            
            // 许可证内容（从文件读取或直接提供）
            String licenseContent = readLicenseFromFile("license.lic");
            
            if (licenseContent != null) {
                // 验证许可证
                LicenseVerificationResult result = LicenseUtils.verifyLicense(publicKeyPath, licenseContent);
                System.out.println("验证结果: " + result);
                
                // 如果验证成功，获取许可证数据
                if (result.isSuccess()) {
                    System.out.println("\n=== 许可证数据 ===");
                    LicenseData licenseData = LicenseUtils.getLicenseData(publicKeyPath, licenseContent);
                    System.out.println("许可证数据: " + licenseData);
                    
                    // 检查许可证是否即将过期
                    long currentTime = System.currentTimeMillis() / 1000;
                    long expiresAt = licenseData.getExpiresAt();
                    long daysUntilExpiry = (expiresAt - currentTime) / (24 * 60 * 60);
                    
                    if (daysUntilExpiry < 0) {
                        System.out.println("许可证已过期");
                    } else if (daysUntilExpiry < 7) {
                        System.out.println("许可证将在 " + daysUntilExpiry + " 天后过期");
                    } else {
                        System.out.println("许可证有效，剩余 " + daysUntilExpiry + " 天");
                    }
                }
            } else {
                System.out.println("无法读取许可证文件");
            }
            
        } catch (Exception e) {
            System.err.println("发生错误: " + e.getMessage());
            e.printStackTrace();
        }
    }
    
    /**
     * 从文件读取许可证内容
     * @param filePath 文件路径
     * @return 许可证内容，如果读取失败返回null
     */
    private static String readLicenseFromFile(String filePath) {
        try {
            File file = new File(filePath);
            if (!file.exists()) {
                System.out.println("许可证文件不存在: " + filePath);
                return null;
            }
            
            return new String(Files.readAllBytes(Paths.get(filePath)));
        } catch (IOException e) {
            System.err.println("读取许可证文件失败: " + e.getMessage());
            return null;
        }
    }
}