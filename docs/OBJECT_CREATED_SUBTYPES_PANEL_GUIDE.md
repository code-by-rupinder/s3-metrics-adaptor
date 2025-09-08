# Object Created Subtypes Panel Guide

## 🎯 **Panel Overview**

The "Object Created Subtypes" panel provides individual gauges for each S3 Object Created subtype, allowing you to monitor different types of object creation operations in real-time.

## 📊 **Panel Features**

### **Visualization Type**: Gauge

- **Purpose**: Shows current values for each subtype
- **Layout**: Horizontal arrangement with individual gauges
- **Repeat**: Per bucket (shows separate panel for each selected bucket)

### **Subtypes Monitored**:

| Subtype              | Description                | Use Case                            |
| -------------------- | -------------------------- | ----------------------------------- |
| **Put**              | Direct object uploads      | File uploads, API calls             |
| **Copy**             | Object copy operations     | Data replication, backups           |
| **Post**             | POST-based uploads         | Web form uploads, multipart uploads |
| **Multipart Upload** | CompleteMultipartUpload    | Large file uploads                  |
| **PostCopy**         | POST-based copy operations | Web-based data copying              |

## 🔧 **Configuration**

### **Thresholds**:

- **Green**: 0-99 events (Normal activity)
- **Yellow**: 100-499 events (Elevated activity)
- **Orange**: 500-999 events (High activity)
- **Red**: 1000+ events (Very high activity)

### **Time Range**:

- Uses dashboard time range (`$__range`)
- Shows total events in selected time period

### **Bucket Filtering**:

- Respects bucket variable selection
- Shows separate panel for each bucket
- Supports "All" selection for multiple buckets

## 📈 **PromQL Queries**

### **Put Operations**:

```promql
sum(increase(s3_event_total{bucket=~"$bucket", event="Object Created", subtype="Put"}[$__range]))
```

### **Copy Operations**:

```promql
sum(increase(s3_event_total{bucket=~"$bucket", event="Object Created", subtype="Copy"}[$__range]))
```

### **Post Operations**:

```promql
sum(increase(s3_event_total{bucket=~"$bucket", event="Object Created", subtype="Post"}[$__range]))
```

### **Multipart Upload Operations**:

```promql
sum(increase(s3_event_total{bucket=~"$bucket", event="Object Created", subtype="CompleteMultipartUpload"}[$__range]))
```

### **PostCopy Operations**:

```promql
sum(increase(s3_event_total{bucket=~"$bucket", event="Object Created", subtype="PostCopy"}[$__range]))
```

## 🎨 **Visual Design**

### **Gauge Configuration**:

- **Orientation**: Auto (adjusts based on available space)
- **Min Size**: 75x75 pixels
- **Threshold Labels**: Enabled
- **Threshold Markers**: Enabled
- **Unit**: Short format (K, M, B for thousands, millions, billions)

### **Color Scheme**:

- **Green**: Normal activity levels
- **Yellow**: Elevated activity (attention needed)
- **Orange**: High activity (monitoring required)
- **Red**: Very high activity (investigation needed)

## 📊 **Use Cases**

### **1. Upload Pattern Analysis**

- **Put**: Direct file uploads
- **Post**: Web form uploads
- **Multipart Upload**: Large file uploads

### **2. Data Replication Monitoring**

- **Copy**: Cross-region replication
- **PostCopy**: Web-based data copying

### **3. Performance Monitoring**

- **High Put**: Normal file uploads
- **High Copy**: Data replication activity
- **High Multipart**: Large file processing

### **4. Anomaly Detection**

- **Unexpected spikes** in any subtype
- **Missing activity** in expected subtypes
- **Unusual patterns** across subtypes

## 🔍 **Interpretation Guide**

### **Normal Patterns**:

- **Put**: High (normal file uploads)
- **Copy**: Low to medium (replication)
- **Post**: Low (web uploads)
- **Multipart**: Low to medium (large files)
- **PostCopy**: Very low (rare)

### **Alert Conditions**:

- **Red Put**: Excessive uploads (possible attack)
- **Red Copy**: High replication (backup activity)
- **Red Post**: Web upload spike (form submissions)
- **Red Multipart**: Large file processing (data migration)

### **Investigation Triggers**:

- **Sudden spikes** in any subtype
- **Missing activity** in expected subtypes
- **Unusual ratios** between subtypes

## ⚙️ **Customization Options**

### **Threshold Adjustment**:

```json
"thresholds": {
  "mode": "absolute",
  "steps": [
    {"color": "green"},
    {"color": "yellow", "value": 50},    // Lower threshold
    {"color": "orange", "value": 200},   // Medium threshold
    {"color": "red", "value": 500}       // High threshold
  ]
}
```

### **Time Range Modification**:

```promql
# Last 1 hour
sum(increase(s3_event_total{bucket=~"$bucket", event="Object Created", subtype="Put"}[1h]))

# Last 24 hours
sum(increase(s3_event_total{bucket=~"$bucket", event="Object Created", subtype="Put"}[24h]))
```

### **Additional Subtypes**:

```promql
# Add new subtype
sum(increase(s3_event_total{bucket=~"$bucket", event="Object Created", subtype="NewSubtype"}[$__range]))
```

## 🚨 **Alerting Recommendations**

### **High Activity Alerts**:

```yaml
# High Put activity
- alert: HighPutActivity
  expr: sum(increase(s3_event_total{event="Object Created", subtype="Put"}[5m])) > 100
  for: 2m
  labels:
    severity: warning
  annotations:
    summary: "High Put activity detected"

# High Copy activity
- alert: HighCopyActivity
  expr: sum(increase(s3_event_total{event="Object Created", subtype="Copy"}[5m])) > 50
  for: 2m
  labels:
    severity: warning
  annotations:
    summary: "High Copy activity detected"
```

### **Anomaly Detection**:

```yaml
# Unexpected Post activity
- alert: UnexpectedPostActivity
  expr: sum(increase(s3_event_total{event="Object Created", subtype="Post"}[5m])) > 10
  for: 1m
  labels:
    severity: critical
  annotations:
    summary: "Unexpected Post activity detected"
```

## 📋 **Best Practices**

### **1. Monitoring Strategy**:

- **Set appropriate thresholds** based on your environment
- **Monitor trends** over time, not just absolute values
- **Compare across buckets** to identify patterns

### **2. Alert Configuration**:

- **Use different thresholds** for different environments
- **Set up escalation** for critical alerts
- **Include context** in alert messages

### **3. Dashboard Organization**:

- **Group related panels** together
- **Use consistent color schemes** across panels
- **Provide clear titles** and descriptions

### **4. Performance Optimization**:

- **Limit time ranges** for better performance
- **Use appropriate refresh rates** (30s-1m)
- **Filter by bucket** when possible

## 🔧 **Troubleshooting**

### **No Data Showing**:

1. **Check bucket selection** in variables
2. **Verify time range** is appropriate
3. **Confirm metrics are being generated**
4. **Check Prometheus data source**

### **Incorrect Values**:

1. **Verify metric labels** match expected format
2. **Check time range** calculation
3. **Confirm bucket filtering** is working
4. **Validate PromQL syntax**

### **Performance Issues**:

1. **Reduce time range** if too long
2. **Filter by specific buckets** instead of "All"
3. **Increase refresh interval** if too frequent
4. **Check Prometheus performance**

## 📚 **Related Documentation**

- [S3 Event Types Guide](S3_EVENT_TYPES_GUIDE.md)
- [Grafana Dashboard Guide](GRAFANA_DASHBOARD_GUIDE.md)
- [High Cardinality Solutions](HIGH_CARDINALITY_SOLUTIONS.md)
- [Object Created Subtypes Guide](OBJECT_CREATED_SUBTYPES_GUIDE.md)

The Object Created Subtypes panel provides detailed insights into different types of object creation operations, helping you monitor and analyze S3 activity patterns effectively!
