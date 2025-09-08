#!/bin/bash

# Base URL for the exporter
EXPORTER_URL="http://localhost:8087"

# Function to send event
send_event() {
    local event="$1"
    curl -X POST "${EXPORTER_URL}/events" \
        -H "Content-Type: application/json" \
        -d "$event"
    echo -e "\nSent event: $event\n"
    sleep 2  # Wait between events
}

# 1. Normal object creation
send_event '{
    "Records": [{
        "eventVersion": "2.1",
        "eventTime": "2025-08-13T10:00:00.000Z",
        "eventName": "ObjectCreated:Put",
        "userIdentity": {"principalId": "EXAMPLE"},
        "requestParameters": {"sourceIPAddress": "192.168.1.1"},
        "s3": {
            "bucket": {"name": "test-bucket"},
            "object": {"key": "folder1/test.txt", "size": 1024}
        }
    }]
}'

# 2. System-initiated deletion (anomaly type: system_delete)
send_event '{
    "Records": [{
        "eventVersion": "2.1",
        "eventTime": "2025-08-13T10:01:00.000Z",
        "eventName": "ObjectRemoved:Delete",
        "userIdentity": {"principalId": "s3.amazonaws.com"},
        "requestParameters": {"sourceIPAddress": "s3.amazonaws.com"},
        "s3": {
            "bucket": {"name": "test-bucket"},
            "object": {"key": "folder1/expired.txt"}
        }
    }]
}'

# 3. Delete marker creation (anomaly type: delete_marker_created)
send_event '{
    "Records": [{
        "eventVersion": "2.1",
        "eventTime": "2025-08-13T10:02:00.000Z",
        "eventName": "ObjectRemoved:DeleteMarkerCreated",
        "userIdentity": {"principalId": "EXAMPLE"},
        "requestParameters": {"sourceIPAddress": "192.168.1.1"},
        "s3": {
            "bucket": {"name": "test-bucket"},
            "object": {"key": "folder2/versioned.txt"}
        }
    }]
}'

# 4. Manual deletion (anomaly type: manual_delete)
send_event '{
    "Records": [{
        "eventVersion": "2.1",
        "eventTime": "2025-08-13T10:03:00.000Z",
        "eventName": "ObjectRemoved:Delete",
        "userIdentity": {"principalId": "EXAMPLE"},
        "requestParameters": {"sourceIPAddress": "192.168.1.1"},
        "s3": {
            "bucket": {"name": "test-bucket"},
            "object": {"key": "folder1/manual-delete.txt"}
        }
    }]
}'

# 5. Multiple files in root (for prefix testing)
send_event '{
    "Records": [{
        "eventVersion": "2.1",
        "eventTime": "2025-08-13T10:04:00.000Z",
        "eventName": "ObjectCreated:Put",
        "userIdentity": {"principalId": "EXAMPLE"},
        "requestParameters": {"sourceIPAddress": "192.168.1.1"},
        "s3": {
            "bucket": {"name": "test-bucket"},
            "object": {"key": "root-file.txt", "size": 2048}
        }
    }]
}'

# 6. Lifecycle expiration
send_event '{
    "Records": [{
        "eventVersion": "2.1",
        "eventTime": "2025-08-13T10:05:00.000Z",
        "eventName": "ObjectRemoved:Delete",
        "userIdentity": {"principalId": "s3.amazonaws.com"},
        "requestParameters": {"sourceIPAddress": "s3.amazonaws.com"},
        "s3": {
            "bucket": {"name": "test-bucket"},
            "object": {"key": "folder3/expired-file.txt"}
        },
        "reason": "Lifecycle Expiration"
    }]
}'

echo "All test events have been sent!"
