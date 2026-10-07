package middleware

import "strings"

// SafeLogPath removes capability-style values from routes before they reach
// application logs. Payment authorities and tracking codes are secrets.
func SafeLogPath(path string) string {
	parts := strings.Split(path, "/")
	for i := 0; i < len(parts)-1; i++ {
		switch parts[i] {
		case "payments", "tracking":
			parts[i+1] = ":redacted"
		}
	}
	return strings.Join(parts, "/")
}
