# Object Created Subtypes Guide

## Overview

The S3 metrics adaptor tracks different types of object creation operations through the `subtype` label in the `s3_event_total` metric. This guide explains what each subtype means and how to interpret them in your dashboard.

## Object Created Subtypes

### 📤 **Put**

- **What it is**: Standard single-part upload
- **When it occurs**: When you upload a file directly to S3
- **Use cases**: Small to medium files, direct uploads
- **Example**: `aws s3 cp file.txt s3://bucket/`

### 📋 **Copy**

- **What it is**: Copy operation within S3
- **When it occurs**: When you copy an object from one location to another
- **Use cases**: Moving files, creating backups, data migration
- **Example**: `aws s3 cp s3://bucket/source/ s3://bucket/dest/`

### 📤 **Post (Multipart Upload)**

- **What it is**: Multipart upload operations
- **When it occurs**: When uploading large files in chunks
- **Use cases**: Large files (>5GB), resumable uploads
- **Example**: Large file uploads via AWS CLI or SDK

### 📤 **PostCopy**

- **What it is**: Multipart copy operation
- **When it occurs**: When copying large files using multipart
- **Use cases**: Large file migrations, cross-region copies
- **Example**: `aws s3 cp` with large files

### 🔄 **CompleteMultipartUpload**

- **What it is**: Final step of multipart upload
- **When it occurs**: When multipart upload is completed
- **Use cases**: Finishing large file uploads
- **Example**: Finalizing a multipart upload

## Dashboard Panels Added

### 1. **Object Created Subtypes Pie Chart**

- **Query**: `sum(increase(s3_event_total{bucket=~"$bucket", event="Object Created"}[$__range])) by (subtype)`
- **Shows**: Distribution of subtypes as percentages
- **Use**: Quick overview of operation types

### 2. **Object Created Subtypes Detail Table**

- **Query**: `sum(increase(s3_event_total{bucket=~"$bucket", event="Object Created"}[$__range])) by (subtype, bucket)`
- **Shows**: Exact counts for each subtype
- **Use**: Detailed analysis and comparison

### 3. **Object Created Subtypes Over Time**

- **Query**: `sum(rate(s3_event_total{bucket=~"$bucket", event="Object Created"}[5m])) by (bucket, subtype)`
- **Shows**: Rate of each subtype over time
- **Use**: Trend analysis and performance monitoring

## Interpreting the Data

### High Put Operations

- **Indicates**: Normal file uploads
- **Action**: Monitor for unusual spikes
- **Health**: Generally good

### High Copy Operations

- **Indicates**: Data movement or backup activities
- **Action**: Check if expected (backups, migrations)
- **Health**: Monitor for unexpected copies

### High Post Operations

- **Indicates**: Large file uploads
- **Action**: Monitor storage usage and costs
- **Health**: Normal for large file workflows

### High PostCopy Operations

- **Indicates**: Large file migrations
- **Action**: Monitor for data transfer costs
- **Health**: Check if migrations are expected

## Common Patterns

### Normal Operations

```
Put: 80-90%
Copy: 5-10%
Post: 5-10%
PostCopy: 0-5%
```

### Backup Operations

```
Put: 60-70%
Copy: 20-30%
Post: 5-10%
PostCopy: 0-5%
```

### Large File Workflows

```
Put: 40-50%
Copy: 10-20%
Post: 20-30%
PostCopy: 10-20%
```

## Monitoring Recommendations

### Alerts to Set Up

1. **High Copy Rate**: Unusual data movement
2. **High Post Rate**: Large file uploads (monitor costs)
3. **Zero Put Operations**: Potential upload issues
4. **High PostCopy Rate**: Large migrations (monitor costs)

### Key Metrics to Watch

- **Put/Copy Ratio**: Should be stable
- **Post Rate**: Monitor for cost implications
- **Subtype Distribution**: Should match expected patterns

## Troubleshooting

### No Subtype Data

- Check if `subtype` label is being populated
- Verify S3 events are being parsed correctly
- Check time range selection

### Unexpected Subtypes

- Review S3 event parsing logic
- Check for custom S3 operations
- Verify event source configuration

### Performance Issues

- Use `topk()` to limit results
- Filter by specific subtypes if needed
- Consider shorter time ranges

## Query Examples

### Get All Subtypes

```promql
sum(increase(s3_event_total{event="Object Created"}[1h])) by (subtype)
```

### Get Specific Subtype

```promql
sum(increase(s3_event_total{event="Object Created", subtype="Put"}[1h])) by (bucket)
```

### Get Subtype Rate

```promql
sum(rate(s3_event_total{event="Object Created"}[5m])) by (subtype)
```

### Get Top Subtypes

```promql
topk(5, sum(increase(s3_event_total{event="Object Created"}[1h])) by (subtype))
```

## Summary

The Object Created subtypes provide detailed insights into how your S3 objects are being created. This helps with:

- **Cost Optimization**: Understanding which operations are most expensive
- **Performance Monitoring**: Tracking upload patterns
- **Security Analysis**: Identifying unusual data movement
- **Capacity Planning**: Understanding storage growth patterns

The new dashboard panels give you complete visibility into your S3 object creation patterns!
