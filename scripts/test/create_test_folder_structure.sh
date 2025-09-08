#!/bin/bash

# Script to create a comprehensive folder structure in S3 for testing path labeling
# This will create various organizational structures to test different path labeling strategies

BUCKET="s3event-test-s3"

echo "🏗️  Creating comprehensive S3 folder structure for path labeling testing..."
echo "Bucket: $BUCKET"
echo ""

# Function to create a file in S3
create_file() {
    local path="$1"
    local content="$2"
    local file_type="$3"
    
    echo "Creating: $path"
    echo "$content" | aws s3 cp - "s3://$BUCKET/$path" --content-type "$file_type" --metadata "test-file=true,created-by=path-labeling-test"
    sleep 1  # Small delay to ensure events are processed
}

# 1. Finance Department Structure (Department/Year/Quarter/Reports)
echo "📊 Creating Finance Department Structure..."
create_file "finance/2025/q1/reports/monthly-summary.pdf" "Q1 2025 Monthly Financial Summary" "application/pdf"
create_file "finance/2025/q1/reports/quarterly-budget.xlsx" "Q1 2025 Budget Report" "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
create_file "finance/2025/q1/reports/expense-analysis.csv" "Q1 2025 Expense Analysis" "text/csv"
create_file "finance/2025/q1/reports/audit-trail.log" "Q1 2025 Audit Trail" "text/plain"

create_file "finance/2025/q2/reports/monthly-summary.pdf" "Q2 2025 Monthly Financial Summary" "application/pdf"
create_file "finance/2025/q2/reports/quarterly-budget.xlsx" "Q2 2025 Budget Report" "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
create_file "finance/2025/q2/reports/expense-analysis.csv" "Q2 2025 Expense Analysis" "text/csv"

create_file "finance/2025/q3/reports/monthly-summary.pdf" "Q3 2025 Monthly Financial Summary" "application/pdf"
create_file "finance/2025/q3/reports/quarterly-budget.xlsx" "Q3 2025 Budget Report" "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"

create_file "finance/2025/q4/reports/monthly-summary.pdf" "Q4 2025 Monthly Financial Summary" "application/pdf"
create_file "finance/2025/q4/reports/quarterly-budget.xlsx" "Q4 2025 Budget Report" "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"

# 2. HR Department Structure (Department/Year/Type/Documents)
echo "👥 Creating HR Department Structure..."
create_file "hr/2025/employees/employee-database.json" "2025 Employee Database" "application/json"
create_file "hr/2025/employees/payroll-data.csv" "2025 Payroll Data" "text/csv"
create_file "hr/2025/employees/performance-reviews.pdf" "2025 Performance Reviews" "application/pdf"

create_file "hr/2025/recruitment/job-postings.json" "2025 Job Postings" "application/json"
create_file "hr/2025/recruitment/candidate-resumes.pdf" "2025 Candidate Resumes" "application/pdf"
create_file "hr/2025/recruitment/interview-notes.txt" "2025 Interview Notes" "text/plain"

create_file "hr/2025/policies/employee-handbook.pdf" "2025 Employee Handbook" "application/pdf"
create_file "hr/2025/policies/benefits-guide.pdf" "2025 Benefits Guide" "application/pdf"

# 3. IT Department Structure (Department/Project/Environment/Configs)
echo "💻 Creating IT Department Structure..."
create_file "it/2025/infrastructure/production/config.yaml" "Production Infrastructure Config" "application/x-yaml"
create_file "it/2025/infrastructure/production/nginx.conf" "Production Nginx Config" "text/plain"
create_file "it/2025/infrastructure/production/database.sql" "Production Database Schema" "application/sql"

create_file "it/2025/infrastructure/staging/config.yaml" "Staging Infrastructure Config" "application/x-yaml"
create_file "it/2025/infrastructure/staging/nginx.conf" "Staging Nginx Config" "text/plain"

create_file "it/2025/infrastructure/development/config.yaml" "Development Infrastructure Config" "application/x-yaml"
create_file "it/2025/infrastructure/development/docker-compose.yml" "Development Docker Compose" "application/x-yaml"

create_file "it/2025/applications/web-app/frontend/build.js" "Web App Frontend Build" "application/javascript"
create_file "it/2025/applications/web-app/backend/api.py" "Web App Backend API" "text/x-python"
create_file "it/2025/applications/web-app/database/schema.sql" "Web App Database Schema" "application/sql"

# 4. Marketing Department Structure (Department/Campaign/Type/Assets)
echo "📢 Creating Marketing Department Structure..."
create_file "marketing/2025/campaigns/summer-sale/banner.png" "Summer Sale Banner" "image/png"
create_file "marketing/2025/campaigns/summer-sale/email-template.html" "Summer Sale Email Template" "text/html"
create_file "marketing/2025/campaigns/summer-sale/analytics.json" "Summer Sale Analytics" "application/json"

create_file "marketing/2025/campaigns/black-friday/banner.png" "Black Friday Banner" "image/png"
create_file "marketing/2025/campaigns/black-friday/email-template.html" "Black Friday Email Template" "text/html"
create_file "marketing/2025/campaigns/black-friday/analytics.json" "Black Friday Analytics" "application/json"

create_file "marketing/2025/content/blog-posts/article-1.md" "Blog Post Article 1" "text/markdown"
create_file "marketing/2025/content/blog-posts/article-2.md" "Blog Post Article 2" "text/markdown"
create_file "marketing/2025/content/social-media/instagram-post.jpg" "Instagram Post" "image/jpeg"

# 5. Operations Department Structure (Department/Process/Type/Logs)
echo "⚙️  Creating Operations Department Structure..."
create_file "operations/2025/monitoring/application-logs/app.log" "Application Log" "text/plain"
create_file "operations/2025/monitoring/application-logs/error.log" "Error Log" "text/plain"
create_file "operations/2025/monitoring/application-logs/access.log" "Access Log" "text/plain"

create_file "operations/2025/monitoring/system-metrics/cpu-usage.json" "CPU Usage Metrics" "application/json"
create_file "operations/2025/monitoring/system-metrics/memory-usage.json" "Memory Usage Metrics" "application/json"
create_file "operations/2025/monitoring/system-metrics/disk-usage.json" "Disk Usage Metrics" "application/json"

create_file "operations/2025/backups/daily/database-backup.sql.gz" "Daily Database Backup" "application/gzip"
create_file "operations/2025/backups/weekly/full-backup.tar.gz" "Weekly Full Backup" "application/gzip"
create_file "operations/2025/backups/monthly/archive-backup.tar.gz" "Monthly Archive Backup" "application/gzip"

# 6. Legal Department Structure (Department/Case/Type/Documents)
echo "⚖️  Creating Legal Department Structure..."
create_file "legal/2025/contracts/employment/contract-template.pdf" "Employment Contract Template" "application/pdf"
create_file "legal/2025/contracts/employment/nda-template.pdf" "NDA Template" "application/pdf"
create_file "legal/2025/contracts/vendor/service-agreement.pdf" "Vendor Service Agreement" "application/pdf"

create_file "legal/2025/compliance/gdpr/privacy-policy.pdf" "GDPR Privacy Policy" "application/pdf"
create_file "legal/2025/compliance/gdpr/data-processing-agreement.pdf" "Data Processing Agreement" "application/pdf"
create_file "legal/2025/compliance/sox/audit-report.pdf" "SOX Audit Report" "application/pdf"

# 7. Sales Department Structure (Department/Region/Quarter/Reports)
echo "💰 Creating Sales Department Structure..."
create_file "sales/2025/north-america/q1/revenue-report.xlsx" "North America Q1 Revenue" "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
create_file "sales/2025/north-america/q1/leads.csv" "North America Q1 Leads" "text/csv"
create_file "sales/2025/north-america/q1/forecast.json" "North America Q1 Forecast" "application/json"

create_file "sales/2025/europe/q1/revenue-report.xlsx" "Europe Q1 Revenue" "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
create_file "sales/2025/europe/q1/leads.csv" "Europe Q1 Leads" "text/csv"

create_file "sales/2025/asia/q1/revenue-report.xlsx" "Asia Q1 Revenue" "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
create_file "sales/2025/asia/q1/leads.csv" "Asia Q1 Leads" "text/csv"

# 8. Research Department Structure (Department/Project/Phase/Data)
echo "🔬 Creating Research Department Structure..."
create_file "research/2025/ai-project/phase-1/experiment-data.json" "AI Project Phase 1 Data" "application/json"
create_file "research/2025/ai-project/phase-1/results.pdf" "AI Project Phase 1 Results" "application/pdf"
create_file "research/2025/ai-project/phase-1/code/model.py" "AI Project Model Code" "text/x-python"

create_file "research/2025/ai-project/phase-2/experiment-data.json" "AI Project Phase 2 Data" "application/json"
create_file "research/2025/ai-project/phase-2/results.pdf" "AI Project Phase 2 Results" "application/pdf"

create_file "research/2025/blockchain-project/phase-1/whitepaper.pdf" "Blockchain Whitepaper" "application/pdf"
create_file "research/2025/blockchain-project/phase-1/prototype.js" "Blockchain Prototype" "application/javascript"

echo ""
echo "✅ Folder structure creation completed!"
echo ""
echo "📁 Created folder structure:"
echo "   📊 finance/2025/q{1,2,3,4}/reports/ - Financial reports by quarter"
echo "   👥 hr/2025/{employees,recruitment,policies}/ - HR documents by type"
echo "   💻 it/2025/{infrastructure,applications}/ - IT resources by project"
echo "   📢 marketing/2025/{campaigns,content}/ - Marketing materials by campaign"
echo "   ⚙️  operations/2025/{monitoring,backups}/ - Operations data by process"
echo "   ⚖️  legal/2025/{contracts,compliance}/ - Legal documents by type"
echo "   💰 sales/2025/{north-america,europe,asia}/q1/ - Sales data by region"
echo "   🔬 research/2025/{ai-project,blockchain-project}/ - Research data by project"
echo ""
echo "🎯 Path Labeling Test Scenarios:"
echo "   • Department strategy: Will extract 'department', 'year', 'quarter', 'type' labels"
echo "   • Generic strategy: Will extract 'path_1', 'path_2', 'path_3', 'path_4' labels"
echo "   • Custom strategy: Can be configured with specific segment names"
echo ""
echo "📊 Check your metrics at: http://localhost:8087/metrics"
echo "🔍 Look for metrics with path labels like:"
echo "   • s3_event_total{department=\"finance\", year=\"2025\", quarter=\"q1\"}"
echo "   • s3_events_hierarchical_path_total{department=\"hr\", year=\"2025\", type=\"employees\"}"
echo ""
