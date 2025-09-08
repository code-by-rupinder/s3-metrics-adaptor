#!/bin/bash

# Script to upload actual files to the existing S3 folder structure for path labeling testing
# This will create real files with content to test the path labeling feature

BUCKET="s3event-test-s3"

echo "📁 Uploading files to existing S3 folder structure for path labeling testing..."
echo "Bucket: $BUCKET"
echo ""

# Function to create and upload a file
upload_file() {
    local path="$1"
    local content="$2"
    local file_type="$3"
    local local_file="/tmp/$(basename "$path")"
    
    echo "Uploading: $path"
    echo "$content" > "$local_file"
    aws s3 cp "$local_file" "s3://$BUCKET/$path" --content-type "$file_type" --metadata "test-file=true,created-by=path-labeling-test,upload-time=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
    rm -f "$local_file"
    sleep 1  # Small delay to ensure events are processed
}

# 1. Finance Department Files
echo "📊 Uploading Finance Department Files..."
upload_file "finance/2025/q1/reports/monthly-summary.pdf" "Q1 2025 Monthly Financial Summary Report\n\nRevenue: $2.5M\nExpenses: $1.8M\nNet Profit: $700K\n\nKey Metrics:\n- Customer Acquisition: 1,200\n- Revenue Growth: 15%\n- Cost Reduction: 8%" "application/pdf"

upload_file "finance/2025/q1/reports/quarterly-budget.xlsx" "Q1 2025 Budget Allocation\n\nDepartment,Allocated,Spent,Remaining\nMarketing,500000,450000,50000\nEngineering,800000,750000,50000\nSales,600000,580000,20000\nOperations,400000,380000,20000" "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"

upload_file "finance/2025/q1/reports/expense-analysis.csv" "Date,Category,Amount,Description\n2025-01-15,Office Supplies,2500.00,Stationery and equipment\n2025-01-20,Travel,15000.00,Client meetings\n2025-01-25,Software,5000.00,License renewals\n2025-02-10,Marketing,25000.00,Digital advertising" "text/csv"

upload_file "finance/2025/q1/reports/audit-trail.log" "2025-01-01 09:00:00 [INFO] Financial audit started\n2025-01-01 09:15:00 [INFO] Revenue records verified\n2025-01-01 09:30:00 [INFO] Expense records verified\n2025-01-01 09:45:00 [INFO] Budget compliance checked\n2025-01-01 10:00:00 [INFO] Audit completed successfully" "text/plain"

# Q2 Files
upload_file "finance/2025/q2/reports/monthly-summary.pdf" "Q2 2025 Monthly Financial Summary Report\n\nRevenue: $2.8M\nExpenses: $1.9M\nNet Profit: $900K\n\nKey Metrics:\n- Customer Acquisition: 1,500\n- Revenue Growth: 18%\n- Cost Reduction: 10%" "application/pdf"

upload_file "finance/2025/q2/reports/quarterly-budget.xlsx" "Q2 2025 Budget Allocation\n\nDepartment,Allocated,Spent,Remaining\nMarketing,550000,520000,30000\nEngineering,850000,800000,50000\nSales,650000,620000,30000\nOperations,420000,400000,20000" "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"

# 2. HR Department Files
echo "👥 Uploading HR Department Files..."
upload_file "hr/2025/employees/employee-database.json" '{"employees": [{"id": 1, "name": "John Doe", "department": "Engineering", "salary": 95000, "hire_date": "2023-01-15"}, {"id": 2, "name": "Jane Smith", "department": "Marketing", "salary": 75000, "hire_date": "2023-03-20"}, {"id": 3, "name": "Bob Johnson", "department": "Sales", "salary": 80000, "hire_date": "2023-06-10"}]}' "application/json"

upload_file "hr/2025/employees/payroll-data.csv" "Employee ID,Name,Department,Salary,Pay Period,Net Pay\n1,John Doe,Engineering,95000,2025-01,71250\n2,Jane Smith,Marketing,75000,2025-01,56250\n3,Bob Johnson,Sales,80000,2025-01,60000" "text/csv"

upload_file "hr/2025/recruitment/job-postings.json" '{"postings": [{"title": "Senior Software Engineer", "department": "Engineering", "location": "Remote", "salary_range": "120000-150000", "posted_date": "2025-01-15"}, {"title": "Marketing Manager", "department": "Marketing", "location": "New York", "salary_range": "80000-100000", "posted_date": "2025-01-20"}]}' "application/json"

# 3. IT Department Files
echo "💻 Uploading IT Department Files..."
upload_file "it/2025/infrastructure/production/config.yaml" "api:\n  host: api.production.company.com\n  port: 443\n  ssl: true\n  rate_limit: 1000\n\ndatabase:\n  host: db-prod.company.com\n  port: 5432\n  ssl: true\n  max_connections: 100\n\nmonitoring:\n  enabled: true\n  metrics_endpoint: /metrics\n  health_check: /health" "application/x-yaml"

upload_file "it/2025/infrastructure/production/nginx.conf" "server {\n    listen 80;\n    server_name api.production.company.com;\n    \n    location / {\n        proxy_pass http://backend;\n        proxy_set_header Host $host;\n        proxy_set_header X-Real-IP $remote_addr;\n    }\n    \n    location /health {\n        access_log off;\n        return 200 'healthy';\n    }\n}" "text/plain"

upload_file "it/2025/applications/web-app/frontend/build.js" "// Web App Frontend Build Configuration\nconst webpack = require('webpack');\nconst path = require('path');\n\nmodule.exports = {\n  entry: './src/index.js',\n  output: {\n    path: path.resolve(__dirname, 'dist'),\n    filename: 'bundle.js'\n  },\n  module: {\n    rules: [\n      {\n        test: /\\.js$/,\n        exclude: /node_modules/,\n        use: 'babel-loader'\n      }\n    ]\n  }\n};" "application/javascript"

# 4. Marketing Department Files
echo "📢 Uploading Marketing Department Files..."
upload_file "marketing/2025/campaigns/summer-sale/email-template.html" "<!DOCTYPE html>\n<html>\n<head>\n    <title>Summer Sale 2025</title>\n</head>\n<body>\n    <h1>Summer Sale - Up to 50% Off!</h1>\n    <p>Don't miss our biggest sale of the year!</p>\n    <a href='https://company.com/summer-sale'>Shop Now</a>\n</body>\n</html>" "text/html"

upload_file "marketing/2025/campaigns/summer-sale/analytics.json" '{"campaign": "summer-sale-2025", "metrics": {"impressions": 150000, "clicks": 15000, "conversions": 1500, "revenue": 75000, "ctr": 0.10, "conversion_rate": 0.10}}' "application/json"

upload_file "marketing/2025/content/blog-posts/article-1.md" "# The Future of E-commerce\n\nE-commerce is rapidly evolving with new technologies...\n\n## Key Trends\n- AI-powered personalization\n- Voice commerce\n- Social shopping\n- Sustainability focus\n\n## Conclusion\nThese trends will shape the future of online retail." "text/markdown"

# 5. Operations Department Files
echo "⚙️ Uploading Operations Department Files..."
upload_file "operations/2025/monitoring/application-logs/app.log" "2025-01-15 10:00:00 [INFO] Application started\n2025-01-15 10:01:00 [INFO] Database connection established\n2025-01-15 10:02:00 [INFO] Cache initialized\n2025-01-15 10:03:00 [INFO] API server listening on port 8080\n2025-01-15 10:04:00 [INFO] Health check endpoint ready" "text/plain"

upload_file "operations/2025/monitoring/system-metrics/cpu-usage.json" '{"timestamp": "2025-01-15T10:00:00Z", "metrics": {"cpu_usage_percent": 45.2, "memory_usage_percent": 67.8, "disk_usage_percent": 23.1, "network_in_mbps": 125.5, "network_out_mbps": 89.3}}' "application/json"

upload_file "operations/2025/backups/daily/database-backup.sql.gz" "CREATE TABLE users (id INT PRIMARY KEY, name VARCHAR(100), email VARCHAR(100));\nINSERT INTO users VALUES (1, 'John Doe', 'john@company.com');\nINSERT INTO users VALUES (2, 'Jane Smith', 'jane@company.com');" "application/gzip"

# 6. Legal Department Files
echo "⚖️ Uploading Legal Department Files..."
upload_file "legal/2025/contracts/employment/contract-template.pdf" "EMPLOYMENT CONTRACT TEMPLATE\n\nThis Employment Contract (\"Agreement\") is entered into between [Company Name] (\"Company\") and [Employee Name] (\"Employee\")...\n\n1. POSITION AND DUTIES\n2. COMPENSATION\n3. BENEFITS\n4. TERMINATION\n5. CONFIDENTIALITY" "application/pdf"

upload_file "legal/2025/compliance/gdpr/privacy-policy.pdf" "PRIVACY POLICY\n\n1. INFORMATION WE COLLECT\n2. HOW WE USE YOUR INFORMATION\n3. DATA SHARING\n4. DATA SECURITY\n5. YOUR RIGHTS\n6. CONTACT INFORMATION" "application/pdf"

# 7. Sales Department Files
echo "💰 Uploading Sales Department Files..."
upload_file "sales/2025/north-america/q1/revenue-report.xlsx" "Region,Quarter,Revenue,Target,Performance\nNorth America,Q1 2025,2500000,2200000,113.6%\nNorth America,Q1 2025,2500000,2200000,113.6%" "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"

upload_file "sales/2025/north-america/q1/leads.csv" "Lead ID,Company,Contact,Email,Status,Value\nL001,Acme Corp,John Smith,john@acme.com,Qualified,50000\nL002,Tech Inc,Jane Doe,jane@tech.com,Prospect,75000\nL003,Global Ltd,Bob Wilson,bob@global.com,Closed Won,100000" "text/csv"

upload_file "sales/2025/europe/q1/revenue-report.xlsx" "Region,Quarter,Revenue,Target,Performance\nEurope,Q1 2025,1800000,2000000,90.0%" "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"

# 8. Research Department Files
echo "🔬 Uploading Research Department Files..."
upload_file "research/2025/ai-project/phase-1/experiment-data.json" '{"experiment": "ai-model-training", "phase": 1, "results": {"accuracy": 0.92, "precision": 0.89, "recall": 0.91, "f1_score": 0.90}, "parameters": {"learning_rate": 0.001, "batch_size": 32, "epochs": 100}}' "application/json"

upload_file "research/2025/ai-project/phase-1/code/model.py" "import tensorflow as tf\nfrom tensorflow import keras\n\nclass AIModel:\n    def __init__(self):\n        self.model = keras.Sequential([\n            keras.layers.Dense(128, activation='relu'),\n            keras.layers.Dropout(0.2),\n            keras.layers.Dense(64, activation='relu'),\n            keras.layers.Dense(1, activation='sigmoid')\n        ])\n    \n    def train(self, x_train, y_train):\n        self.model.compile(optimizer='adam', loss='binary_crossentropy')\n        self.model.fit(x_train, y_train, epochs=100)\n\nif __name__ == '__main__':\n    model = AIModel()\n    print('AI Model initialized successfully')" "text/x-python"

echo ""
echo "✅ File upload completed!"
echo ""
echo "📊 Files uploaded to test path labeling:"
echo "   📊 finance/2025/q{1,2}/reports/ - Financial reports with real data"
echo "   👥 hr/2025/{employees,recruitment}/ - HR data with JSON and CSV"
echo "   💻 it/2025/{infrastructure,applications}/ - IT configs and code"
echo "   📢 marketing/2025/{campaigns,content}/ - Marketing materials and analytics"
echo "   ⚙️  operations/2025/{monitoring,backups}/ - Operations logs and metrics"
echo "   ⚖️  legal/2025/{contracts,compliance}/ - Legal documents"
echo "   💰 sales/2025/{north-america,europe}/q1/ - Sales reports by region"
echo "   🔬 research/2025/ai-project/phase-1/ - Research data and code"
echo ""
echo "🎯 Path Labeling Test Results:"
echo "   • Department strategy will extract: department, year, quarter, type"
echo "   • Example: finance/2025/q1/reports → department='finance', year='2025', quarter='q1', type='reports'"
echo "   • Example: hr/2025/employees → department='hr', year='2025', type='employees'"
echo ""
echo "📊 Check your metrics at: http://localhost:8087/metrics"
echo "🔍 Look for path-labeled metrics like:"
echo "   • s3_event_total{department=\"finance\", year=\"2025\", quarter=\"q1\"}"
echo "   • s3_events_hierarchical_path_total{department=\"hr\", year=\"2025\", type=\"employees\"}"
echo "   • s3_bucket_extension_files_total{department=\"it\", year=\"2025\", type=\"infrastructure\", extension=\"yaml\"}"
echo ""
