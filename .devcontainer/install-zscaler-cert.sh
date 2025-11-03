#!/bin/bash
# Auto-install Zscaler certificates if needed

if [ -f "/tmp/zscaler-certs/zscaler-root.crt" ]; then
    echo "Installing Zscaler certificate from mounted volume..."
    sudo cp /tmp/zscaler-certs/zscaler-root.crt /usr/local/share/ca-certificates/
    sudo update-ca-certificates
else
    # Check if Zscaler is intercepting HTTPS
    ISSUER=$(echo | openssl s_client -connect google.com:443 -servername google.com 2>/dev/null | openssl x509 -noout -issuer 2>/dev/null | grep -i zscaler)
    
    if [ -n "$ISSUER" ]; then
        echo "Zscaler detected, extracting certificate..."
        echo | openssl s_client -connect google.com:443 -servername google.com -showcerts 2>/dev/null | sed -n '/-----BEGIN CERTIFICATE-----/,/-----END CERTIFICATE-----/p' > /tmp/zscaler-chain.pem
        openssl crl2pkcs7 -nocrl -certfile /tmp/zscaler-chain.pem | openssl pkcs7 -print_certs | awk '/-----BEGIN CERTIFICATE-----/{cert++} cert==3' > /tmp/zscaler-root.crt
        sudo cp /tmp/zscaler-root.crt /usr/local/share/ca-certificates/zscaler-root.crt
        sudo update-ca-certificates
    else
        echo "No Zscaler detected, skipping certificate installation"
    fi
fi