# Debian Salsa Submission Guide for FyClip

This guide explains how to submit FyClip to Debian Salsa for official Debian package review.

## Prerequisites

1. **Debian Salsa Account**: Create an account at https://salsa.debian.org
2. **Debian Developer/Maintainer**: You need to be a Debian Developer or have a sponsor
3. **GPG Key**: Set up a GPG key for signing packages
4. **Git**: Familiarity with Git and Debian packaging

## Step 1: Create Salsa Repository

1. Log in to https://salsa.debian.org
2. Create a new project:
   - **Project name**: `fyclip`
   - **Visibility**: Public
   - **Initialize repository**: No (we'll push existing code)

3. Clone the repository:
   ```bash
   git clone git@salsa.debian.org:YOUR_USERNAME/fyclip.git
   cd fyclip
   ```

## Step 2: Prepare Source for Salsa

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

## Step 3: Verify Debian Packaging

Before submitting, verify your Debian packaging:

```bash
# Install lintian
sudo apt-get install lintian

# Build the package
./build-deb.sh

# Run lintian checks
lintian -i -I --show-overrides dist/fyclip_2.2.2-1_amd64.changes

# Review any warnings or errors
```

### Common Lintian Issues

1. **debian-watch-does-not-check-gpg-signature**
   - Add GPG signature verification to `debian/watch`

2. **package-uses-old-debhelper-compat-level**
   - Update `debian/control` to use latest compat level

3. **binary-without-manpage**
   - Create a manpage for `fyclip`

4. **description-synopsis-starts-with-article**
   - Fix package description in `debian/control`

## Step 4: Create Manpage (Optional but Recommended)

Create `debian/fyclip.1`:

```bash
cat > debian/fyclip.1 << 'EOF'
.TH FYCLIP 1 "April 2026" "FyClip 2.2.2" "User Commands"
.SH NAME
fyclip \- Advanced Clipboard Manager for Linux
.SH SYNOPSIS
.B fyclip
.RI [ options ]
.SH DESCRIPTION
FyClip is a powerful, feature-rich clipboard manager built with Go and Fyne.
.PP
Key features:
.IP \(bu 2
Clipboard History: Automatically saves text, images, HTML, and files
.IP \(bu 2
Pin Items: Keep important items at the top
.IP \(bu 2
Enhanced Search: Regex, case-sensitive, and fuzzy matching
.IP \(bu 2
Image Support: Preview and save clipboard images
.IP \(bu 2
Encrypted Storage: AES-256-GCM encryption at rest
.IP \(bu 2
Snippets: Save and expand text templates
.IP \(bu 2
AutoStart: Launch on system startup
.IP \(bu 2
Theme Support: Light, Dark, and System theme modes
.IP \(bu 2
Auto Update: Check for and install updates from GitHub
.SH OPTIONS
.TP
.B \-\-version
Display version information
.TP
.B \-\-help
Display help information
.SH FILES
.TP
.I ~/.config/fyclip/
User configuration directory
.TP
.I ~/.local/share/fyclip/
User data directory
.SH AUTHORS
Sarwar Hossain <sarwarhridoy4@gmail.com>
.SH SEE ALSO
.BR xclip (1),
.BR xsel (1),
.BR wl-clipboard (1)
EOF
```

Update `debian/rules` to install the manpage:

```makefile
override_dh_auto_install:
	# ... existing installation code ...
	
	# Install manpage
	install -d $(CURDIR)/debian/fyclip/usr/share/man/man1
	install -m 0644 debian/fyclip.1 $(CURDIR)/debian/fyclip/usr/share/man/man1/
```

## Step 5: Update debian/control

Add manpage to package description:

```control
Description: Advanced Clipboard Manager for Linux
 FyClip is a powerful, feature-rich clipboard manager built with Go and Fyne.
 .
 Key features:
  * Clipboard History: Automatically saves text, images, HTML, and files
  * Pin Items: Keep important items at the top
  * Enhanced Search: Regex, case-sensitive, and fuzzy matching
  * Image Support: Preview and save clipboard images
  * Encrypted Storage: AES-256-GCM encryption at rest
  * Snippets: Save and expand text templates
  * AutoStart: Launch on system startup
  * Theme Support: Light, Dark, and System theme modes
  * Auto Update: Check for and install updates from GitHub
  * Quick Panel: Fast access to clipboard history
  * Sensitive Content Detection: Automatically detect and hide sensitive data
  * Backup and Restore: Export and import clipboard data
  * Customizable Exclusions: Exclude specific applications from monitoring
 .
 This package also includes a manpage (fyclip.1).
```

## Step 6: Commit and Push to Salsa

```bash
# Add all files
git add .

# Commit
git commit -m "Initial Debian packaging for FyClip 2.2.2"

# Push to Salsa
git push origin main
```

## Step 7: Request Sponsorship

### Option A: Debian Mentors

1. Subscribe to debian-mentors mailing list
2. Send an email with:
   - Package name and version
   - Salsa repository URL
   - Lintian output
   - Description of the package

### Option B: Direct Sponsorship

1. Find a Debian Developer willing to sponsor
2. Provide them with:
   - Salsa repository access
   - Package documentation
   - Build instructions

### Option C: Debian Package Tracker

1. Create a package page on https://tracker.debian.org
2. Upload your package
3. Wait for review

## Step 8: Review Process

The review process typically includes:

1. **Initial Review**: Sponsor reviews package structure
2. **Lintian Check**: Automated quality checks
3. **Build Test**: Package builds on Debian infrastructure
4. **Manual Review**: Sponsor reviews code and packaging
5. **Approval**: Package accepted into Debian

## Step 9: After Approval

Once approved:

1. **Package enters unstable**: Available in Debian unstable
2. **Migration to testing**: After 5-10 days without issues
3. **Stable release**: Included in next Debian stable release

## Useful Resources

- [Debian New Maintainer's Guide](https://www.debian.org/doc/manuals/maint-guide/)
- [Debian Policy Manual](https://www.debian.org/doc/debian-policy/)
- [Debian Packaging Tutorial](https://www.debian.org/doc/manuals/packaging-tutorial/)
- [Lintian](https://lintian.debian.org/)
- [Debian Mentors](https://mentors.debian.net/)
- [Debian Salsa Documentation](https://salsa.debian.org/help/)

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
