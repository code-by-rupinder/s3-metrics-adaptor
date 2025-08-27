package metrics

import (
	"testing"
)

func TestGetPrefix(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{
			name:     "key with prefix",
			key:      "folder/file.txt",
			expected: "folder",
		},
		{
			name:     "key without prefix",
			key:      "file.txt",
			expected: "/",
		},
		{
			name:     "nested folders",
			key:      "folder/subfolder/file.txt",
			expected: "folder",
		},
		{
			name:     "empty key",
			key:      "",
			expected: "/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getPrefix(tt.key)
			if result != tt.expected {
				t.Errorf("getPrefix(%q) = %q, want %q", tt.key, result, tt.expected)
			}
		})
	}
}

func TestGetFileInfo(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected fileInfo
	}{
		{
			name: "file with extension",
			key:  "folder/test.txt",
			expected: fileInfo{
				extension: ".txt",
				fileType:  "file",
			},
		},
		{
			name: "file without extension",
			key:  "folder/testfile",
			expected: fileInfo{
				extension: "none",
				fileType:  "file",
			},
		},
		{
			name: "directory",
			key:  "folder/subfolder/",
			expected: fileInfo{
				extension: "none",
				fileType:  "directory",
			},
		},
		{
			name: "file with multiple dots",
			key:  "folder/test.backup.tar.gz",
			expected: fileInfo{
				extension: ".gz",
				fileType:  "file",
			},
		},
		{
			name: "hidden file with extension",
			key:  "folder/.config.json",
			expected: fileInfo{
				extension: ".json",
				fileType:  "file",
			},
		},
		{
			name: "hidden file without extension",
			key:  "folder/.gitignore",
			expected: fileInfo{
				extension: "none",
				fileType:  "file",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getFileInfo(tt.key)
			if result.extension != tt.expected.extension || result.fileType != tt.expected.fileType {
				t.Errorf("getFileInfo(%q) = {extension: %q, fileType: %q}, want {extension: %q, fileType: %q}",
					tt.key, result.extension, result.fileType, tt.expected.extension, tt.expected.fileType)
			}
		})
	}
}

func TestGetPrefixAtDepth(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		depth    int
		expected string
	}{
		{
			name:     "depth 1",
			key:      testNestedPath,
			depth:    1,
			expected: "folder1",
		},
		{
			name:     "depth 2",
			key:      testNestedPath,
			depth:    2,
			expected: "folder1/subfolder",
		},
		{
			name:     "depth 3 with shorter path",
			key:      testNestedPath,
			depth:    3,
			expected: testNestedPath,
		},
		{
			name:     "depth 2 with shorter path",
			key:      "folder1/file.txt",
			depth:    2,
			expected: "folder1/file.txt",
		},
		{
			name:     "depth 0",
			key:      testNestedPath,
			depth:    0,
			expected: "",
		},
		{
			name:     "negative depth",
			key:      testNestedPath,
			depth:    -1,
			expected: "",
		},
		{
			name:     "empty key",
			key:      "",
			depth:    2,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getPrefixAtDepth(tt.key, tt.depth)
			if result != tt.expected {
				t.Errorf("getPrefixAtDepth(%q, %d) = %q, want %q", tt.key, tt.depth, result, tt.expected)
			}
		})
	}
}
