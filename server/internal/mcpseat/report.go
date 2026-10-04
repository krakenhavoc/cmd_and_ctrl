package mcpseat

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// stats is §9's game-end report: what the seat measured about its own
// game. The binary cannot see the model's token count; the client
// reports that.
type stats struct {
	windows    int            // decision windows seen
	absorbed   map[string]int // answered automatically, by rule
	shown      int            // shown to the model
	acts       int            // act calls that reached the server
	rejections int            // ... and were refused
	autoErrors int            // automatic answers the server refused

	truncatedWindows int // windows whose wire list was cut
	fullListRequests int // legal_moves_request frames sent
	largestMoves     int // the largest move list shown, in moves
	largestMoveBytes int // ... and in bytes

	toolBytes int // bytes returned in tool results

	decisionTimes []time.Duration // decision opening to its accepted act
}

func newStats() stats { return stats{absorbed: map[string]int{}} }

func (s *stats) noteMovesShown(n, bytes int) {
	if n > s.largestMoves {
		s.largestMoves = n
	}
	if bytes > s.largestMoveBytes {
		s.largestMoveBytes = bytes
	}
}

// summary is the one-line count since the last decision.
func absorbedSummary(m map[string]int) string {
	if len(m) == 0 {
		return ""
	}
	keys := make([]string, 0, len(m))
	total := 0
	for k, n := range m {
		keys = append(keys, k)
		total += n
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s %d", k, m[k]))
	}
	return fmt.Sprintf("answered automatically since your last decision: %d (%s)", total, strings.Join(parts, ", "))
}

// report renders the stats.
func (s *stats) report() string {
	var b strings.Builder
	b.WriteString("SEAT REPORT (ADR 0122 §9)\n")
	total := 0
	for _, n := range s.absorbed {
		total += n
	}
	keys := make([]string, 0, len(s.absorbed))
	for k := range s.absorbed {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s %d", k, s.absorbed[k]))
	}
	fmt.Fprintf(&b, "  windows seen: %d; answered automatically: %d", s.windows, total)
	if len(parts) > 0 {
		fmt.Fprintf(&b, " (%s)", strings.Join(parts, ", "))
	}
	fmt.Fprintf(&b, "; shown to the model: %d\n", s.shown)
	fmt.Fprintf(&b, "  act calls: %d; rejected: %d; automatic answers refused: %d\n", s.acts, s.rejections, s.autoErrors)
	fmt.Fprintf(&b, "  truncated windows: %d; full-list requests: %d; largest move list shown: %d moves, %d bytes\n",
		s.truncatedWindows, s.fullListRequests, s.largestMoves, s.largestMoveBytes)
	fmt.Fprintf(&b, "  bytes returned in tool results: %d (about %d tokens)\n", s.toolBytes, s.toolBytes/4)
	if n := len(s.decisionTimes); n > 0 {
		ds := append([]time.Duration(nil), s.decisionTimes...)
		sort.Slice(ds, func(i, j int) bool { return ds[i] < ds[j] })
		var sum time.Duration
		for _, d := range ds {
			sum += d
		}
		fmt.Fprintf(&b, "  decision to act: %d decisions, median %s, mean %s, max %s\n",
			n, ds[n/2].Round(100*time.Millisecond), (sum / time.Duration(n)).Round(100*time.Millisecond),
			ds[n-1].Round(100*time.Millisecond))
	}
	return b.String()
}

// oneLine is the report flattened for the stderr log.
func (s *stats) oneLine() string {
	return strings.Join(strings.Fields(strings.ReplaceAll(s.report(), "\n", "; ")), " ")
}
