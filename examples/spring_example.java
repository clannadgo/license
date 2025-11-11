package com.example.license;

import java.io.BufferedReader;
import java.io.FileReader;
import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.security.KeyFactory;
import java.security.PublicKey;
import java.security.interfaces.RSAPublicKey;
import java.security.spec.*;
import java.time.Instant;
import java.util.Base32;
import java.util.Map;

import com.nimbusds.jose.JWSObject;
import com.nimbusds.jose.crypto.RSASSAVerifier;

/**
 <dependency>
     <groupId>com.nimbusds</groupId>
     <artifactId>nimbus-jose-jwt</artifactId>
     <version>9.36</version>
 </dependency>
 */

public class LicenseUtils {

    // 读取 PEM 公钥
    public static RSAPublicKey loadPublicKey(String pemPath) throws Exception {
        StringBuilder sb = new StringBuilder();
        try (BufferedReader br = new BufferedReader(new FileReader(pemPath))) {
            String line;
            while ((line = br.readLine()) != null) {
                if (line.contains("BEGIN") || line.contains("END")) continue;
                sb.append(line.trim());
            }
        }
        byte[] keyBytes = java.util.Base64.getDecoder().decode(sb.toString());
        try {
            // PKCS#8
            X509EncodedKeySpec spec = new X509EncodedKeySpec(keyBytes);
            KeyFactory kf = KeyFactory.getInstance("RSA");
            return (RSAPublicKey) kf.generatePublic(spec);
        } catch (Exception e) {
            // try PKCS#1
            RSAPublicKeySpec spec = decodeRSAPublicKeyPKCS1(keyBytes);
            KeyFactory kf = KeyFactory.getInstance("RSA");
            return (RSAPublicKey) kf.generatePublic(spec);
        }
    }

    // PKCS#1 -> RSAPublicKeySpec
    private static RSAPublicKeySpec decodeRSAPublicKeyPKCS1(byte[] keyBytes) throws Exception {
        // ASN.1 parsing (simplified, using Java built-in)
        java.io.DataInputStream dis = new java.io.DataInputStream(new java.io.ByteArrayInputStream(keyBytes));
        if (dis.readByte() != 0x30) throw new IllegalArgumentException("Invalid PKCS#1 format");
        // skip length
        dis.readByte();
        // modulus
        dis.readByte(); int modLen = dis.readUnsignedByte();
        byte[] modulus = new byte[modLen];
        dis.readFully(modulus);
        java.math.BigInteger n = new java.math.BigInteger(1, modulus);
        // exponent
        dis.readByte(); int expLen = dis.readUnsignedByte();
        byte[] exponent = new byte[expLen];
        dis.readFully(exponent);
        java.math.BigInteger e = new java.math.BigInteger(1, exponent);
        return new RSAPublicKeySpec(n, e);
    }

    // 验证 JWS license
    public static Map<String, Object> verifyLicense(String jwsCompact, RSAPublicKey pub) throws Exception {
        JWSObject jws = JWSObject.parse(jwsCompact);
        RSASSAVerifier verifier = new RSASSAVerifier(pub);
        if (!jws.verify(verifier)) {
            throw new IllegalArgumentException("invalid license signature");
        }
        String payload = jws.getPayload().toString();
        Map<String,Object> claims = new com.fasterxml.jackson.databind.ObjectMapper().readValue(payload, Map.class);

        // check expiration
        Object expObj = claims.get("exp");
        if (expObj != null) {
            long exp = Long.parseLong(expObj.toString());
            long now = Instant.now().getEpochSecond();
            if (now > exp) {
                throw new IllegalArgumentException("license expired");
            }
        }

        return claims;
    }

    // 机器码 XXXX-XXXX-XXXX-XXXX -> hex
    public static String decodeFingerprintToHex(String code) {
        String s = code.replace("-", "").replace(" ", "").toUpperCase();
        if (s.length() != 16) throw new IllegalArgumentException("invalid fingerprint length");
        Base32 b32 = new Base32();
        byte[] bytes = b32.decode(s);
        StringBuilder sb = new StringBuilder();
        for (byte b : bytes) {
            sb.append(String.format("%02x", b));
        }
        return sb.toString();
    }

    // 校验 fingerprint
    public static void checkFingerprint(String localActivationCode, String licenseFingerprintHex) {
        String localHex = decodeActivationCodeToHex(localActivationCode);
        if (!localHex.equalsIgnoreCase(licenseFingerprintHex)) {
            throw new IllegalArgumentException("fingerprint mismatch");
        }
    }
}
