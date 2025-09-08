# Troubleshooting Missing Metrics in Grafana

## Problem: Missing Data in Specific Panels

You're not seeing data in:

- **Event Processing Latency by Bucket**
- **Source IP Activity**
- **Anomalies by Type**

## 🔍 **Step 1: Use Diagnostic Dashboard**

I've created a diagnostic dashboard (`grafana-dashboard-diagnostic.json`) to help identify the issue:

1. **Import the diagnostic dashboard** into Grafana
2. **Check the "Available S3 Metrics" panel** - this will show all S3 metrics being collected
3. **Look at the specific metric panels** to see if data exists

## 🔧 **Step 2: Common Issues & Solutions**

### Issue 1: Latency Metrics Not Populated

**Problem**: `s3_event_latency_seconds` metric is empty

**Possible Causes**:

- Latency calculation is disabled in config
- No events are being processed
- Metric is not being updated

**Solutions**:

```yaml
# Check config.yaml
metrics:
  types:
    latency: true # Make sure this is enabled
```

**Alternative Query**:

```promql
# Try this query instead
avg(s3_event_latency_seconds) by (bucket)
```

### Issue 2: Source IP Metrics Not Populated

**Problem**: `s3_event_ip_total` metric is empty

**Possible Causes**:

- IP tracking is disabled
- S3 events don't contain source IP information
- Events are coming from system (no IP)

**Solutions**:

```yaml
# Check config.yaml
metrics:
  types:
    ipTotal: true # Make sure this is enabled
```

**Check Raw Data**:

```promql
# Check if any IP data exists
s3_event_ip_total
```

### Issue 3: Anomaly Metrics Not Populated

**Problem**: `s3_event_anomaly_total` metric is empty

**Possible Causes**:

- Anomaly detection is disabled
- No anomalous events are occurring
- Anomaly detection logic is too strict

**Solutions**:

```yaml
# Check config.yaml
metrics:
  types:
    anomalyDetection: true # Make sure this is enabled
```

**Check Raw Data**:

```promql
# Check if any anomaly data exists
s3_event_anomaly_total
```

## 🛠️ **Step 3: Quick Fixes**

### Fix 1: Update Latency Query

The latency query was using histogram syntax. I've fixed it:

**Before**:

```promql
avg(rate(s3_event_latency_seconds_sum{bucket=~"$bucket"}[5m]) / rate(s3_event_latency_seconds_count{bucket=~"$bucket"}[5m])) by (bucket)
```

**After**:

```promql
avg(s3_event_latency_seconds{bucket=~"$bucket"}) by (bucket)
```

### Fix 2: Check Metric Availability

Run these queries in Grafana to check if metrics exist:

```promql
# Check if latency metric exists
s3_event_latency_seconds

# Check if IP metric exists
s3_event_ip_total

# Check if anomaly metric exists
s3_event_anomaly_total
```

### Fix 3: Verify Configuration

Check your `config.yaml`:

```yaml
metrics:
  enabled: true
  types:
    latency: true # For latency metrics
    ipTotal: true # For IP tracking
    anomalyDetection: true # For anomaly detection
```

## 🔍 **Step 4: Debugging Steps**

### 1. Check if S3 Events are Being Processed

```promql
# This should show data
sum(s3_event_total)
```

### 2. Check if Specific Metrics are Enabled

```promql
# Check all S3 metrics
count by (__name__)({__name__=~"s3_.*"})
```

### 3. Check Raw Event Data

```promql
# See what events are being processed
s3_event_total
```

### 4. Check Time Range

- Make sure you're looking at the right time range
- Check if data exists in the selected time period
- Try a longer time range (last 24 hours)

## 🚨 **Step 5: Common Scenarios**

### Scenario 1: No Data at All

**Cause**: S3 metrics adaptor not running or not receiving events
**Solution**:

1. Check if application is running
2. Verify SQS queue has messages
3. Check application logs

### Scenario 2: Some Metrics Work, Others Don't

**Cause**: Specific metrics disabled in configuration
**Solution**: Enable missing metrics in `config.yaml`

### Scenario 3: Data Exists but Panels Show Nothing

**Cause**: Incorrect query syntax
**Solution**: Use the diagnostic dashboard to verify correct metric names

### Scenario 4: Data Exists but Wrong Time Range

**Cause**: Time range doesn't include when data was collected
**Solution**: Adjust time range in Grafana

## 📊 **Step 6: Alternative Queries**

If the standard queries don't work, try these alternatives:

### For Latency:

```promql
# Simple average
avg(s3_event_latency_seconds)

# By bucket
avg(s3_event_latency_seconds) by (bucket)

# Current value
s3_event_latency_seconds
```

### For Source IP:

```promql
# All IP data
s3_event_ip_total

# By IP
sum(s3_event_ip_total) by (ip)

# Rate
sum(rate(s3_event_ip_total[5m])) by (ip)
```

### For Anomalies:

```promql
# All anomalies
s3_event_anomaly_total

# By type
sum(s3_event_anomaly_total) by (type)

# Rate
sum(rate(s3_event_anomaly_total[5m])) by (type)
```

## 🎯 **Step 7: Verification**

After applying fixes:

1. **Import the diagnostic dashboard**
2. **Check if metrics appear in the diagnostic panels**
3. **Verify the main dashboard shows data**
4. **Test with different time ranges**

## 📞 **Still Having Issues?**

If you're still not seeing data:

1. **Check application logs** for errors
2. **Verify S3 events are being received** (check SQS queue)
3. **Confirm metrics are enabled** in configuration
4. **Use the diagnostic dashboard** to identify missing metrics
5. **Check Prometheus** directly at `/metrics` endpoint

The diagnostic dashboard will help identify exactly which metrics are missing and guide you to the solution!
