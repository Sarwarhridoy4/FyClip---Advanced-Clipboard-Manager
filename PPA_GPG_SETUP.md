# PPA GPG Signing Setup Guide

The PPA upload failed because the `.changes` file needs to be signed with a GPG key.

## Step 1: Generate GPG Key

If you don't have a GPG key:

```bash
# Generate a new GPG key
gpg --full-generate-key

# Select:
# - Key type: RSA and RSA
# - Key size: 4096
# - Expiration: 0 (never expires) or 2y (2 years)
# - Name: Sarwar Hossain
# - Email: sarwarhridoy4@gmail.com
```

## Step 2: Get Your GPG Key ID

```bash
# List your GPG keys
gpg --list-keys

# Find your key ID (looks like: 1234567890ABCDEF)
# It's the long hex string after "pub"
```

## Step 3: Export Your GPG Public Key

```bash
# Export your public key
gpg --armor --export sarwarhridoy4@gmail.com > public.key

# Upload to Ubuntu keyserver
gpg --keyserver keyserver.ubuntu.com --send-keys YOUR_KEY_ID
```

## Step 4: Configure Launchpad (Detailed Guide)

### What is Launchpad?

Launchpad is Ubuntu's platform for hosting packages and PPAs. You need to:
1. Create a Launchpad account
2. Link your GPG key to your account
3. Create a PPA repository

### Step 4.1: Create Launchpad Account

1. Go to https://launchpad.net
2. Click "Log in / Register"
3. Click "Create account"
4. Fill in:
   - **Display name**: Sarwar Hossain
   - **Email**: sarwarhridoy4@gmail.com
   - **Username**: sarwar-hossain (or your preferred username)
5. Click "Create account"
6. Check your email and verify your account

### Step 4.2: Get Your GPG Fingerprint

```bash
# Get your GPG fingerprint
gpg --fingerprint sarwarhridoy4@gmail.com

# Output will look like:
# pub   rsa4096 2026-04-01 [SC]
#       2938 364E E8A4 5FA8 3DBF  C1C1 921F C7DC B240 6B07
# uid           [ultimate] Sarwar Hossain <sarwarhridoy4@gmail.com>
# sub   rsa4096 2026-04-01 [E]

# The fingerprint is the long hex string:
# 2938364EE8A45FA83DBFC1C1921FC7DCB2406B07
```

### Step 4.3: Add GPG Key to Launchpad

1. Log in to https://launchpad.net
2. Go to your profile: https://launchpad.net/~YOUR_USERNAME
3. Click "Change details" (top right)
4. Scroll down to "OpenPGP keys"
5. Click "Import Key"
6. Paste your GPG fingerprint (the long hex string)
7. Click "Import"
8. Check your email for verification link
9. Click the verification link to confirm

### Step 4.4: Create PPA

1. Go to https://launchpad.net/~YOUR_USERNAME
2. Click "Create a new PPA"
3. Fill in:
   - **Name**: fyclip
   - **Display name**: FyClip Clipboard Manager
   - **Description**: Advanced Clipboard Manager for Linux
4. Click "Create"

### Step 4.5: Verify PPA Setup

1. Go to https://launchpad.net/~YOUR_USERNAME/+archive/fyclip
2. You should see your PPA page
3. Note the PPA address: `ppa:sarwar-hossain/fyclip`

## Step 5: Sign the .changes File

```bash
# Sign the .changes file
debsign -k YOUR_KEY_ID dist/fyclip_2.2.0-1_source.changes

# Or sign with your email
debsign -k sarwarhridoy4@gmail.com dist/fyclip_2.2.0-1_source.changes
```

## Step 6: Upload to PPA

```bash
# Upload signed package
dput ppa:sarwar-hossain/fyclip dist/fyclip_2.2.0-1_source.changes
```

## Alternative: Build with Signing

```bash
# Build and sign in one step
dpkg-buildpackage -kYOUR_KEY_ID -S -sa

# Or with email
dpkg-buildpackage -ksarwarhridoy4@gmail.com -S -sa
```

## Quick Setup Script

```bash
#!/bin/bash
# Setup GPG for PPA uploads

# Check if GPG key exists
if ! gpg --list-keys sarwarhridoy4@gmail.com > /dev/null 2>&1; then
    echo "Generating GPG key..."
    gpg --full-generate-key
fi

# Get key ID
KEY_ID=$(gpg --list-keys --keyid-format LONG sarwarhridoy4@gmail.com | grep pub | awk '{print $2}' | cut -d'/' -f2)

echo "Your GPG Key ID: $KEY_ID"

# Export public key
gpg --armor --export sarwarhridoy4@gmail.com > public.key

echo "Public key exported to public.key"
echo ""
echo "Next steps:"
echo "1. Upload public.key to https://keyserver.ubuntu.com"
echo "2. Add key fingerprint to Launchpad"
echo "3. Sign .changes file: debsign -k $KEY_ID dist/fyclip_2.2.0-1_source.changes"
echo "4. Upload to PPA: dput ppa:sarwar-hossain/fyclip dist/fyclip_2.2.0-1_source.changes"
```

## Troubleshooting

### GPG Key Not Found

```bash
# List all keys
gpg --list-keys

# If no keys, generate one
gpg --full-generate-key
```

### Signing Failed

```bash
# Check if key is available
gpg --list-keys sarwarhridoy4@gmail.com

# Try signing with key ID
debsign -k YOUR_KEY_ID dist/fyclip_2.2.0-1_source.changes
```

### Upload Failed

```bash
# Check if .changes file is signed
cat dist/fyclip_2.2.0-1_source.changes | grep -A 5 "-----BEGIN PGP"

# If not signed, sign it first
debsign -k YOUR_KEY_ID dist/fyclip_2.2.0-1_source.changes
```

## Your PPA

**PPA URL**: https://launchpad.net/~sarwar-hossain/+archive/ubuntu/fyclip

**Installation Instructions**:
```bash
sudo add-apt-repository ppa:sarwar-hossain/fyclip
sudo apt-get update
sudo apt-get install fyclip
```

## Resources

- [GPG Guide](https://www.gnupg.org/gph/en/manual.html)
- [Launchpad GPG Setup](https://help.launchpad.net/YourAccount/ImportingYourPGPKey)
- [Debian Signing](https://www.debian.org/doc/manuals/maint-guide/build.en.html#signing)
- [Launchpad Help](https://help.launchpad.net)

## Contact

For questions:
- **Maintainer**: Sarwar Hossain <sarwarhridoy4@gmail.com>
- **Upstream**: https://github.com/Sarwarhridoy4/FyClip---Advanced-Clipboard-Manager

## License

FyClip is licensed under the MIT License. See [Licence](Licence) for details.
