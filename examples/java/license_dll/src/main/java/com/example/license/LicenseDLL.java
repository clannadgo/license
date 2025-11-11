package com.example.license;

import com.sun.jna.Library;
import com.sun.jna.Native;
import com.sun.jna.Platform;
import com.sun.jna.Pointer;

/**
 * JNA接口定义，用于调用license共享库
 * 支持Windows(.dll)、Linux(.so)和macOS(.dylib)
 */
public interface LicenseDLL extends Library {
    
    // 加载共享库，根据平台自动选择
    LicenseDLL INSTANCE = (LicenseDLL) Native.load(
        getLibraryName(), 
        LicenseDLL.class
    );
    
    /**
     * 根据平台获取库名称
     * @return 库名称
     */
    static String getLibraryName() {
        String libName;
        if (Platform.isWindows()) {
            libName = "license";
        } else if (Platform.isLinux()) {
            libName = "license";
        } else if (Platform.isMac()) {
            libName = "license";
        } else {
            throw new UnsupportedOperationException("Unsupported platform: " + Platform.getOSType());
        }
        return libName;
    }
    
    /**
     * 获取库文件扩展名
     * @return 库文件扩展名
     */
    static String getLibraryExtension() {
        if (Platform.isWindows()) {
            return ".dll";
        } else if (Platform.isLinux()) {
            return ".so";
        } else if (Platform.isMac()) {
            return ".dylib";
        } else {
            throw new UnsupportedOperationException("Unsupported platform: " + Platform.getOSType());
        }
    }
    
    /**
     * 获取完整的库文件名
     * @return 完整的库文件名
     */
    static String getFullLibraryName() {
        return getLibraryName() + getLibraryExtension();
    }
    
    /**
     * 生成机器指纹
     * @return 机器指纹字符串，格式为XXXX-XXXX-XXXX-XXXX
     */
    Pointer GenerateFingerprint();
    
    /**
     * 验证许可证
     * @param publicKeyPath 公钥文件路径
     * @param licenseContent 许可证内容
     * @return 验证结果码
     *         0: 成功
     *         1: 无效的公钥
     *         2: 无效的许可证
     *         3: 许可证已过期
     *         4: 指纹不匹配
     *         5: 内部错误
     */
    int VerifyLicense(String publicKeyPath, String licenseContent);
    
    /**
     * 获取许可证数据（JSON格式）
     * @param publicKeyPath 公钥文件路径
     * @param licenseContent 许可证内容
     * @return 许可证数据的JSON字符串
     */
    Pointer GetLicenseData(String publicKeyPath, String licenseContent);
    
    /**
     * 释放字符串内存
     * @param str 要释放的字符串指针
     */
    void FreeString(Pointer str);
}