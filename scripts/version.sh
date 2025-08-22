#!/bin/bash

# Version Management Script for s3-event-exporter
# Usage: ./scripts/version.sh [patch|minor|major|prerelease] [prerelease-type]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Helper functions
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
    exit 1
}

# Check if we're in a git repository
check_git_repo() {
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        log_error "Not in a git repository"
    fi
}

# Check if working directory is clean
check_clean_working_dir() {
    if [[ -n $(git status --porcelain) ]]; then
        log_error "Working directory is not clean. Please commit or stash your changes."
    fi
}

# Get current version from git tags
get_current_version() {
    git tag --sort=-version:refname | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+' | head -1 || echo "v0.0.0"
}

# Parse semantic version
parse_version() {
    local version=$1
    if [[ $version =~ ^v([0-9]+)\.([0-9]+)\.([0-9]+)(-(.+))?$ ]]; then
        MAJOR=${BASH_REMATCH[1]}
        MINOR=${BASH_REMATCH[2]}
        PATCH=${BASH_REMATCH[3]}
        PRERELEASE=${BASH_REMATCH[5]}
    else
        log_error "Invalid version format: $version"
    fi
}

# Increment version based on type
increment_version() {
    local increment_type=$1
    local prerelease_type=${2:-"alpha"}
    
    case $increment_type in
        major)
            ((MAJOR++))
            MINOR=0
            PATCH=0
            PRERELEASE=""
            ;;
        minor)
            ((MINOR++))
            PATCH=0
            PRERELEASE=""
            ;;
        patch)
            ((PATCH++))
            PRERELEASE=""
            ;;
        prerelease)
            if [[ -n $PRERELEASE ]]; then
                # If already a prerelease, increment the number
                if [[ $PRERELEASE =~ ^(.+)\.([0-9]+)$ ]]; then
                    local pre_name=${BASH_REMATCH[1]}
                    local pre_num=${BASH_REMATCH[2]}
                    ((pre_num++))
                    PRERELEASE="${pre_name}.${pre_num}"
                else
                    PRERELEASE="${PRERELEASE}.1"
                fi
            else
                # Create new prerelease
                ((PATCH++))
                PRERELEASE="${prerelease_type}.1"
            fi
            ;;
        *)
            log_error "Invalid increment type: $increment_type. Use: patch, minor, major, or prerelease"
            ;;
    esac
}

# Generate new version string
generate_version_string() {
    if [[ -n $PRERELEASE ]]; then
        echo "v${MAJOR}.${MINOR}.${PATCH}-${PRERELEASE}"
    else
        echo "v${MAJOR}.${MINOR}.${PATCH}"
    fi
}

# Update version in files
update_version_files() {
    local new_version=$1
    
    # Update main.go with version info (if version variables exist)
    if grep -q "var Version" cmd/main.go; then
        sed -i.bak "s/var Version = .*/var Version = \"$new_version\"/" cmd/main.go
        rm cmd/main.go.bak
        log_info "Updated version in cmd/main.go"
    fi
    
    # Update Dockerfile labels (if exists)
    if [[ -f Dockerfile ]]; then
        if grep -q "LABEL version=" Dockerfile; then
            sed -i.bak "s/LABEL version=.*/LABEL version=\"$new_version\"/" Dockerfile
            rm Dockerfile.bak
            log_info "Updated version in Dockerfile"
        fi
    fi
    
    # Update Helm chart (if exists)
    if [[ -f charts/s3-event-exporter/Chart.yaml ]]; then
        sed -i.bak "s/^version:.*/version: ${new_version#v}/" charts/s3-event-exporter/Chart.yaml
        sed -i.bak "s/^appVersion:.*/appVersion: \"$new_version\"/" charts/s3-event-exporter/Chart.yaml
        rm charts/s3-event-exporter/Chart.yaml.bak
        log_info "Updated version in Helm chart"
    fi
}

# Create changelog entry
create_changelog_entry() {
    local new_version=$1
    local current_date=$(date '+%Y-%m-%d')
    
    # Create or update CHANGELOG.md
    if [[ ! -f CHANGELOG.md ]]; then
        cat > CHANGELOG.md << EOF
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [$new_version] - $current_date

### Added
- Initial release

EOF
    else
        # Add new entry after "## [Unreleased]"
        sed -i.bak "/## \[Unreleased\]/a\\
\\
## [$new_version] - $current_date\\
\\
### Added\\
- [Add your changes here]\\
\\
### Changed\\
- [Add your changes here]\\
\\
### Fixed\\
- [Add your changes here]\\
" CHANGELOG.md
        rm CHANGELOG.md.bak
    fi
    
    log_info "Updated CHANGELOG.md with new version entry"
    log_warning "Please update CHANGELOG.md with your actual changes before committing"
}

# Main function
main() {
    local increment_type=${1:-"patch"}
    local prerelease_type=${2:-"alpha"}
    
    log_info "🏷️  Version Management for s3-event-exporter"
    echo
    
    # Checks
    check_git_repo
    check_clean_working_dir
    
    # Get current version
    local current_version=$(get_current_version)
    log_info "Current version: $current_version"
    
    # Parse and increment version
    parse_version "$current_version"
    increment_version "$increment_type" "$prerelease_type"
    
    local new_version=$(generate_version_string)
    log_info "New version: $new_version"
    
    # Confirmation
    echo
    read -p "Do you want to create version $new_version? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "Version creation cancelled"
        exit 0
    fi
    
    # Update files
    log_info "Updating version files..."
    update_version_files "$new_version"
    
    # Create changelog entry
    create_changelog_entry "$new_version"
    
    # Git operations
    log_info "Creating git commit and tag..."
    git add .
    git commit -m "chore: bump version to $new_version"
    git tag -a "$new_version" -m "Release $new_version"
    
    log_success "Version $new_version created successfully!"
    echo
    log_info "Next steps:"
    echo "1. Review and update CHANGELOG.md with actual changes"
    echo "2. Run: git push origin main --tags"
    echo "3. The release pipeline will automatically trigger"
    echo
    log_info "To push now:"
    echo "git push origin main --tags"
}

# Help function
show_help() {
    cat << EOF
Version Management Script for s3-event-exporter

Usage: $0 [increment_type] [prerelease_type]

Increment Types:
  patch       Increment patch version (1.0.0 -> 1.0.1)
  minor       Increment minor version (1.0.0 -> 1.1.0)
  major       Increment major version (1.0.0 -> 2.0.0)
  prerelease  Create or increment prerelease (1.0.0 -> 1.0.1-alpha.1)

Prerelease Types (used with 'prerelease'):
  alpha       Alpha release (default)
  beta        Beta release
  rc          Release candidate

Examples:
  $0 patch                    # 1.0.0 -> 1.0.1
  $0 minor                    # 1.0.0 -> 1.1.0
  $0 major                    # 1.0.0 -> 2.0.0
  $0 prerelease alpha         # 1.0.0 -> 1.0.1-alpha.1
  $0 prerelease beta          # 1.0.1-alpha.1 -> 1.0.1-beta.1
  $0 prerelease rc            # 1.0.1-beta.1 -> 1.0.1-rc.1

EOF
}

# Handle command line arguments
case "${1:-}" in
    -h|--help|help)
        show_help
        exit 0
        ;;
    *)
        main "$@"
        ;;
esac
