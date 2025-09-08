# High Cardinality Solutions for S3 Metrics

## 🚨 **The Problem: 100 Buckets + Thousands of Events**

With 100 buckets and thousands of events, the "Event Processing Latency by Bucket" panel will face:

### **Performance Issues:**

- **Slow Loading**: 30+ seconds to load
- **Memory Overload**: Browser crashes
- **Visual Chaos**: Unreadable with 100+ values
- **Prometheus Strain**: High cardinality impacts performance

### **Cardinality Math:**

```
100 buckets × 3 event types × 2 instances = 600 time series
+ Additional labels (prefix, subtype, etc.) = 1000+ time series
= Cardinality explosion!
```

## 🔧 **Solution Strategies**

### **1. Top-K Filtering (Recommended)**

```promql
# Show only top 10 buckets with highest latency
topk(10, avg(s3_event_latency_seconds) by (bucket))
```

**Benefits:**

- ✅ Fast loading
- ✅ Shows most important data
- ✅ Prevents overload
- ✅ Easy to read

### **2. Table Visualization**

```promql
# Show as sortable table
topk(20, avg(s3_event_latency_seconds) by (bucket))
```

**Benefits:**

- ✅ Handles many rows
- ✅ Sortable by value
- ✅ Searchable
- ✅ Exportable

### **3. Statistical Aggregation**

```promql
# Show overall statistics
avg(s3_event_latency_seconds)  # Average across all buckets
max(s3_event_latency_seconds)  # Maximum latency
quantile(0.95, s3_event_latency_seconds)  # 95th percentile
```

**Benefits:**

- ✅ Single values
- ✅ Fast performance
- ✅ High-level insights
- ✅ No cardinality issues

### **4. Bucket Grouping**

```promql
# Group by environment
avg(s3_event_latency_seconds) by (bucket) and on(bucket)
label_replace(s3_event_latency_seconds, "bucket", "prod", "bucket", "prod.*")
```

**Benefits:**

- ✅ Reduces cardinality
- ✅ Logical grouping
- ✅ Environment-based analysis
- ✅ Scalable

## 📊 **Dashboard Solutions Created**

### **1. High Cardinality Dashboard**

- **File**: `grafana-dashboard-high-cardinality.json`
- **Features**: Multiple strategies for handling 100+ buckets
- **Panels**: Top-K, tables, statistics, filtering

### **2. Panel Types for Different Scenarios**

| Scenario              | Panel Type     | Query                                                 | Max Buckets |
| --------------------- | -------------- | ----------------------------------------------------- | ----------- |
| **Overview**          | Stat           | `avg(s3_event_latency_seconds)`                       | All         |
| **Top Performers**    | Gauge          | `topk(10, avg(s3_event_latency_seconds) by (bucket))` | 10          |
| **Detailed Analysis** | Table          | `topk(20, avg(s3_event_latency_seconds) by (bucket))` | 20          |
| **Trends**            | Time Series    | `topk(5, avg(s3_event_latency_seconds) by (bucket))`  | 5           |
| **Alerts**            | Status History | `s3_event_latency_seconds > 60`                       | Variable    |

## 🎯 **Recommended Approach**

### **For 100+ Buckets:**

1. **Primary Dashboard**: Use statistical aggregation

   ```promql
   avg(s3_event_latency_seconds)  # Overall average
   max(s3_event_latency_seconds)  # Worst case
   quantile(0.95, s3_event_latency_seconds)  # 95th percentile
   ```

2. **Drill-down Dashboard**: Use top-K filtering

   ```promql
   topk(20, avg(s3_event_latency_seconds) by (bucket))
   ```

3. **Alert Dashboard**: Use threshold filtering
   ```promql
   s3_event_latency_seconds > 60  # High latency buckets
   ```

### **For Different Use Cases:**

| Use Case              | Solution      | Query Example                                         |
| --------------------- | ------------- | ----------------------------------------------------- |
| **Executive Summary** | Overall stats | `avg(s3_event_latency_seconds)`                       |
| **Operations**        | Top 10 worst  | `topk(10, avg(s3_event_latency_seconds) by (bucket))` |
| **Detailed Analysis** | Table view    | `topk(20, avg(s3_event_latency_seconds) by (bucket))` |
| **Trends**            | Time series   | `topk(5, avg(s3_event_latency_seconds) by (bucket))`  |
| **Alerts**            | Threshold     | `s3_event_latency_seconds > 60`                       |

## ⚡ **Performance Optimizations**

### **1. Query Optimization**

```promql
# ❌ Bad: High cardinality
avg(s3_event_latency_seconds) by (bucket, event, prefix)

# ✅ Good: Reduced cardinality
topk(10, avg(s3_event_latency_seconds) by (bucket))
```

### **2. Time Range Optimization**

```promql
# ❌ Bad: Long time range
avg(s3_event_latency_seconds) by (bucket)  # Last 7 days

# ✅ Good: Shorter time range
avg(s3_event_latency_seconds) by (bucket)  # Last 1 hour
```

### **3. Refresh Rate Optimization**

```json
// ❌ Bad: Too frequent
"refresh": "5s"

// ✅ Good: Reasonable frequency
"refresh": "30s"
```

## 🔍 **Monitoring Cardinality**

### **Check Current Cardinality**

```promql
# Count unique label combinations
count by (__name__)({__name__=~"s3_.*"})
```

### **Monitor Cardinality Growth**

```promql
# Track cardinality over time
count by (__name__)({__name__=~"s3_.*"})
```

### **Set Cardinality Alerts**

```yaml
# In Prometheus rules
- alert: HighCardinality
  expr: count by (__name__)({__name__=~"s3_.*"}) > 1000
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "High cardinality detected"
```

## 🚀 **Implementation Steps**

### **Step 1: Import High Cardinality Dashboard**

1. Import `grafana-dashboard-high-cardinality.json`
2. Test with your current data
3. Verify performance improvements

### **Step 2: Choose Strategy Based on Needs**

- **< 20 buckets**: Use original gauge panel
- **20-50 buckets**: Use table visualization
- **50+ buckets**: Use top-K filtering
- **100+ buckets**: Use statistical aggregation

### **Step 3: Configure Alerts**

```yaml
# High latency alert
- alert: HighLatency
  expr: avg(s3_event_latency_seconds) by (bucket) > 60
  for: 5m
  labels:
    severity: warning
```

### **Step 4: Monitor Performance**

- Check dashboard load times
- Monitor Prometheus performance
- Adjust refresh rates as needed

## 📈 **Expected Results**

### **Before (100 buckets):**

- ❌ 30+ second load time
- ❌ Browser crashes
- ❌ Unreadable visualization
- ❌ High memory usage

### **After (with solutions):**

- ✅ < 5 second load time
- ✅ Stable performance
- ✅ Clear visualization
- ✅ Low memory usage

## 🎯 **Best Practices**

### **1. Start Conservative**

- Begin with `topk(10)` or `topk(20)`
- Increase gradually if needed
- Monitor performance

### **2. Use Appropriate Visualizations**

- **Gauges**: < 10 values
- **Tables**: 10-50 values
- **Time Series**: < 10 series
- **Statistics**: Any number

### **3. Implement Layered Dashboards**

- **Executive**: High-level stats
- **Operations**: Top performers
- **Detailed**: Drill-down analysis
- **Alerts**: Threshold monitoring

### **4. Regular Monitoring**

- Check cardinality growth
- Monitor performance metrics
- Adjust strategies as needed
- Set up alerts for issues

## 📞 **Troubleshooting**

### **Still Slow Performance?**

1. Reduce `topk()` value (try `topk(5)`)
2. Shorten time range
3. Increase refresh interval
4. Use statistical aggregation

### **Missing Important Data?**

1. Increase `topk()` value
2. Use table visualization
3. Implement filtering
4. Create drill-down dashboards

### **Cardinality Still Growing?**

1. Review metric labels
2. Implement label filtering
3. Consider metric redesign
4. Use aggregation strategies

The high cardinality dashboard provides multiple strategies to handle 100+ buckets efficiently while maintaining useful insights!
