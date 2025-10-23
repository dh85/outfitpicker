# Winget Package Manifests

This directory contains the winget package manifests for outfitpicker.

## Steps to Submit to Winget

1. **Get SHA256 hashes** after creating the v2.0.0 release:
   ```bash
   # For x64
   curl -sL https://github.com/dh85/outfitpicker/releases/download/v2.0.0/outfitpicker_Windows_x86_64.zip | sha256sum
   
   # For x86
   curl -sL https://github.com/dh85/outfitpicker/releases/download/v2.0.0/outfitpicker_Windows_i386.zip | sha256sum
   ```

2. **Update the installer manifest** with actual SHA256 hashes

3. **Fork** the `microsoft/winget-pkgs` repository

4. **Copy** the manifest files to:
   ```
   microsoft/winget-pkgs/manifests/d/dh85/outfitpicker/2.0.0/
   ```

5. **Submit PR** to `microsoft/winget-pkgs`

6. **Wait** for Microsoft validation (1-3 days)

## Test Locally

```powershell
# Validate manifests
winget validate winget/manifests/d/dh85/outfitpicker/2.0.0/

# Test installation (after updating SHA256 hashes)
winget install --manifest winget/manifests/d/dh85/outfitpicker/2.0.0/dh85.outfitpicker.yaml
```

## After Approval

Users can install with:
```powershell
winget install dh85.outfitpicker
```