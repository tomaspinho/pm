package util

import (
	"fmt"
	"time"
)

// RelativeTime returns a human-readable relative time string.
// Shows seconds for times under 1 minute.
func RelativeTime(t time.Time) string {
	seconds := int(time.Since(t).Seconds())

	if seconds < 0 {
		return "just now"
	}
	if seconds < 10 {
		return "just now"
	}
	if seconds < 60 {
		return fmt.Sprintf("%d seconds ago", seconds)
	}

	minutes := seconds / 60
	if minutes < 60 {
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	}

	hours := minutes / 60
	if hours < 24 {
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}

	days := hours / 24
	if days == 1 {
		return "yesterday"
	}
	if days < 30 {
		return fmt.Sprintf("%d days ago", days)
	}

	months := days / 30
	if months < 12 {
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	}

	years := months / 12
	if years == 1 {
		return "1 year ago"
	}
	return fmt.Sprintf("%d years ago", years)
}

// FormatAbsoluteTime formats a timestamp for tooltip display.
func FormatAbsoluteTime(t time.Time) string {
	return t.Format("January 2, 2006 at 3:04 PM MST")
}
