#!/bin/bash

# Script to manually release Helm chart to GitHub Pages
# This can be used for initial setup or manual releases

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
CHARTS_DIR="${REPO_ROOT}/helm"
PACKAGE_DIR="${REPO_ROOT}/.helm-packages"
PAGES_DIR="${REPO_ROOT}/.helm-pages"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    if ! command -v helm &> /dev/null; then
        log_error "Helm is not installed"
        exit 1
    fi
    
    if ! command -v git &> /dev/null; then
        log_error "Git is not installed"
        exit 1
    fi
    
    log_info "Prerequisites check passed"
}

# Clean up previous builds
cleanup() {
    log_info "Cleaning up previous builds..."
    rm -rf "${PACKAGE_DIR}" "${PAGES_DIR}"
    mkdir -p "${PACKAGE_DIR}" "${PAGES_DIR}"
}

# Package Helm charts
package_charts() {
    log_info "Packaging Helm charts..."
    
    cd "${REPO_ROOT}"
    
    for chart_dir in "${CHARTS_DIR}"/*/; do
        if [[ -f "${chart_dir}/Chart.yaml" ]]; then
            chart_name=$(basename "${chart_dir}")
            log_info "Packaging chart: ${chart_name}"
            
            # Update dependencies
            helm dependency update "${chart_dir}"
            
            # Package the chart
            helm package "${chart_dir}" --destination "${PACKAGE_DIR}"
        fi
    done
}

# Generate index.yaml
generate_index() {
    log_info "Generating Helm repository index..."
    
    cd "${PACKAGE_DIR}"
    
    # Check if there's an existing index from gh-pages branch
    if git show-branch gh-pages &> /dev/null; then
        log_info "Fetching existing index from gh-pages branch..."
        git show gh-pages:index.yaml > "${PAGES_DIR}/index.yaml" 2>/dev/null || true
    fi
    
    # Generate new index
    if [[ -f "${PAGES_DIR}/index.yaml" ]]; then
        helm repo index . --url https://codebyrupinder.github.io/s3-event-exporter/ --merge "${PAGES_DIR}/index.yaml"
    else
        helm repo index . --url https://codebyrupinder.github.io/s3-event-exporter/
    fi
    
    # Copy index and packages to pages directory
    cp index.yaml "${PAGES_DIR}/"
    cp *.tgz "${PAGES_DIR}/" 2>/dev/null || log_warn "No chart packages found"
}

# Commit and push to gh-pages
publish_to_github_pages() {
    log_info "Publishing to GitHub Pages..."
    
    cd "${REPO_ROOT}"
    
    # Check if gh-pages branch exists
    if ! git show-branch gh-pages &> /dev/null; then
        log_info "Creating gh-pages branch..."
        git checkout --orphan gh-pages
        git rm -rf .
        echo "# Helm Repository" > README.md
        git add README.md
        git commit -m "Initial gh-pages commit"
        git push -u origin gh-pages
        git checkout main
    fi
    
    # Switch to gh-pages branch
    git checkout gh-pages
    
    # Copy files
    cp "${PAGES_DIR}"/* . 2>/dev/null || log_warn "No files to copy"
    
    # Add and commit
    git add .
    
    if git diff --staged --quiet; then
        log_info "No changes to publish"
    else
        log_info "Committing changes..."
        git commit -m "Release Helm charts $(date -u +%Y%m%d-%H%M%S)"
        git push origin gh-pages
        log_info "Successfully published to GitHub Pages"
    fi
    
    # Return to main branch
    git checkout main
}

# Validate the repository
validate_repository() {
    log_info "Validating Helm repository..."
    
    # Add the repository temporarily
    helm repo add s3-event-exporter-test https://codebyrupinder.github.io/s3-event-exporter/ --force-update
    
    # Search for charts
    helm search repo s3-event-exporter-test/
    
    # Remove test repository
    helm repo remove s3-event-exporter-test
    
    log_info "Repository validation completed"
}

# Main execution
main() {
    log_info "Starting Helm chart release process..."
    
    check_prerequisites
    cleanup
    package_charts
    generate_index
    
    if [[ "${1}" == "--publish" ]]; then
        publish_to_github_pages
        
        # Wait a bit for GitHub Pages to update
        log_info "Waiting for GitHub Pages to update..."
        sleep 30
        
        validate_repository
    else
        log_info "Chart packaging completed. Use --publish flag to publish to GitHub Pages"
        log_info "Generated files are in: ${PAGES_DIR}"
    fi
    
    log_info "Release process completed successfully!"
}

# Help message
show_help() {
    cat << EOF
Helm Chart Release Script

Usage: $0 [OPTIONS]

OPTIONS:
    --publish    Publish the charts to GitHub Pages
    --help       Show this help message

Examples:
    $0                  # Package charts only
    $0 --publish        # Package and publish to GitHub Pages

EOF
}

# Parse arguments
case "${1:-}" in
    --help|-h)
        show_help
        exit 0
        ;;
    --publish)
        main --publish
        ;;
    "")
        main
        ;;
    *)
        log_error "Unknown option: $1"
        show_help
        exit 1
        ;;
esac
