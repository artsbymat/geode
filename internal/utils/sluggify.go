package utils

import (
	"strings"
)

func PathToSlug(path string) string {
	if path == "" {
		return ""
	}

	segments := strings.Split(path, "/")

	for i, seg := range segments {
		seg = strings.ReplaceAll(seg, " ", "-")
		seg = strings.ReplaceAll(seg, "&", "-and-")
		seg = strings.ReplaceAll(seg, "%", "-percent")
		seg = strings.ReplaceAll(seg, "?", "")
		seg = strings.ReplaceAll(seg, "#", "")

		segments[i] = seg
	}

	out := strings.Join(segments, "/")

	out = strings.TrimSuffix(out, "/")

	return out
}
