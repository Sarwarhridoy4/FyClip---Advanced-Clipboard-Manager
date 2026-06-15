# Submitting FyClip to Official Debian APT Repository

This guide explains how to submit FyClip to the official Debian APT repository.

## Overview

To submit to the official Debian APT repository, you need to:

1. **Create a Salsa account** (Debian's GitLab instance)
2. **Push your source code** to Salsa
3. **Request sponsorship** from a Debian Developer
4. **Package review** and approval
5. **Package enters Debian unstable**
6. **Automatic migration to testing**
7. **Included in next stable release**

## Ubuntu PPA (Available Now)

While working on official Debian submission, you can use the Ubuntu PPA:

**PPA URL**: https://launchpad.net/~sarwar-hossain/+archive/ubuntu/fyclip

**Installation**:
```bash
sudo add-apt-repository ppa:sarwar-hossain/fyclip
sudo apt-get update
sudo apt-get install fyclip
```

This provides immediate access to FyClip for Ubuntu users while the official Debian package is being reviewed.

## Step 1: Create Salsa Account

1. Go to https://salsa.debian.org
2. Sign up using your Debian credentials
3. Set up your profile and GPG key

## Step 2: Create Salsa Repository

1. Log in to https://salsa.debian.org
2. Create a new project:
   - **Project name**: `fyclip`
   - **Visibility**: Public
   - **Initialize repository**: No

3. Clone the repository:
   ```bash
   git clone git@salsa.debian.org:YOUR_USERNAME/fyclip.git
   cd fyclip
   ```

## Step 3: Import Source Code

### Option A: Import from GitHub

```bash
# Add GitHub as upstream
git remote add upstream https://github.com/Sarwarhridoy4/FyClip---Advanced-Clipboard-Manager.git

# Fetch upstream
git fetch upstream

# Merge upstream/main
git merge upstream/main --allow-unrelated-histories

# Push to Salsa
git push origin main
```

### Option B: Manual Import

```bash
# Download source tarball
wget https://github.com/Sarwarhridoy4/FyClip---Advanced-Clipboard-Manager/archive/refs/tags/v2.2.2.tar.gz

# Extract
tar -xzf v2.2.2.tar.gz
cd FyClip---Advanced-Clipboard-Manager-2.2.2

# Initialize git
git init
git add .
git commit -m "Initial import of FyClip 2.2.2"

# Add Salsa remote
git remote add origin git@salsa.debian.org:YOUR_USERNAME/fyclip.git

# Push to Salsa
git push -u origin main
```

## Step 4: Verify Package Quality

Before requesting sponsorship, ensure your package meets Debian standards:

```bash
# Install lintian
sudo apt-get install lintian

# Build the package
./build-deb.sh

# Run lintian checks
lintian -i -I --show-overrides dist/fyclip_2.2.2-1_amd64.changes

# Review and fix any issues
```

### Common Issues to Fix

1. **debian-watch-does-not-check-gpg-signature**
   - Add GPG signature verification to `debian/watch`

2. **package-uses-old-debhelper-compat-level**
   - Update `debian/control` to use latest compat level

3. **binary-without-manpage**
   - Create a manpage for `fyclip` (already done: `debian/fyclip.1`)

4. **description-synopsis-starts-with-article**
   - Fix package description in `debian/control`

## Step 5: Request Sponsorship

### Option A: Debian Mentors (Recommended)

1. Subscribe to debian-mentors mailing list:
   ```bash
   # Send email to debian-mentors-request@lists.debian.org
   # Subject: subscribe
   ```

2. Send sponsorship request email:
   ```
   To: debian-mentors@lists.debian.org
   Subject: RFS: fyclip -- Advanced Clipboard Manager for Linux
   
   Dear mentors,
   
   I am seeking a sponsor for my package "fyclip".
   
   Package: fyclip
   Version: 2.2.2-1
   Description: Advanced Clipboard Manager for Linux
   
   Salsa repository: https://salsa.debian.org/YOUR_USERNAME/fyclip
   
   The package builds successfully and passes lintian checks.
   
   Key features:
   - Clipboard history with search functionality
   - Quick panel for fast access
   - Sensitive content detection
   - Backup and restore capabilities
   - Customizable exclusions
   - System tray integration
   
   Thank you for your consideration.
   
   Best regards,
   Your Name
   ```

### Option B: Direct Sponsorship

1. Find a Debian Developer willing to sponsor:
   - Check https://www.debian.org/devel/developers
   - Contact developers in your area
   - Attend Debian events

2. Provide them with:
   - Salsa repository access
   - Package documentation
   - Build instructions

### Option C: Debian Package Tracker

1. Create a package page on https://tracker.debian.org
2. Upload your package
3. Wait for review

## Step 6: Package Review Process

The review process typically includes:

1. **Initial Review**: Sponsor reviews package structure
2. **Lintian Check**: Automated quality checks
3. **Build Test**: Package builds on Debian infrastructure
4. **Manual Review**: Sponsor reviews code and packaging
5. **Approval**: Package accepted into Debian

## Step 7: After Approval

Once approved:

1. **Package enters unstable**: Available in Debian unstable
   ```bash
   # Add unstable repository
   echo "deb http://deb.debian.org/debian/ unstable main" | sudo tee /etc/apt/sources.list.d/unstable.list
   
   # Update package list
   sudo apt-get update
   
   # Install FyClip
   sudo apt-get install fyclip
   ```

2. **Migration to testing**: After 5-10 days without issues
   ```bash
   # Add testing repository
   echo "deb http://deb.debian.org/debian/ testing main" | sudo tee /etc/apt/sources.list.d/testing.list
   
   # Update package list
   sudo apt-get update
   
   # Install FyClip
   sudo apt-get install fyclip
   ```

3. **Stable release**: Included in next Debian stable release
   ```bash
   # Available in Debian stable (e.g., Debian 13)
   sudo apt-get install fyclip
   ```

## Timeline

- **Week 1-2**: Create Salsa repository, push code
- **Week 2-3**: Request sponsorship, wait for sponsor
- **Week 3-6**: Package review and fixes
- **Week 6-8**: Package approved, enters unstable
- **Week 8-10**: Migration to testing
- **6-12 months**: Included in next stable release

## Useful Resources

- [Debian New Maintainer's Guide](https://www.debian.org/doc/manuals/maint-guide/)
- [Debian Policy Manual](https://www.debian.org/doc/debian-policy/)
- [Debian Packaging Tutorial](https://www.debian.org/doc/manuals/packaging-tutorial/)
- [Lintian](https://lintian.debian.org/)
- [Debian Mentors](https://mentors.debian.net/)
- [Debian Salsa](https://salsa.debian.org/)
- [Debian Developers](https://www.debian.org/devel/developers)

## Troubleshooting

### Build Failures

```bash
# Clean and rebuild
dpkg-buildpackage -us -uc -b -tc

# Check build dependencies
dpkg-checkbuilddeps
```

### Lintian Errors

```bash
# Get detailed explanation
lintian -i -I --show-overrides dist/fyclip_2.2.2-1_amd64.changes

# Fix issues and rebuild
```

### Git Issues

```bash
# Reset if needed
git reset --hard HEAD

# Force push (use with caution)
git push -f origin main
```

## Contact

For questions about this package:
- **Maintainer**: Sarwar Hossain <sarwarhridoy4@gmail.com>
- **Upstream**: https://github.com/Sarwarhridoy4/FyClip---Advanced-Clipboard-Manager
- **Salsa**: https://salsa.debian.org/YOUR_USERNAME/fyclip

## License

FyClip is licensed under the MIT License. See [Licence](Licence) for details.
