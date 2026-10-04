package protocol

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// renderOpeningRollLogText writes the opening roll's lines (ADR 0121
// §3): "Alice and Carol tied with 17 and roll again", "Carol won the
// opening roll with 20", "Bob rolled for Carol and Dave", "Carol chose
// to take the first turn", "Carol chose Bob to take the first turn".
func renderOpeningRollLogText(e LogEvent, actor, target string) string {
	if e.Kind == LogStartingPlayer {
		if e.TargetSeat != nil && *e.TargetSeat == e.Seat {
			return fmt.Sprintf("%s chose to take the first turn", actor)
		}
		return fmt.Sprintf("%s chose %s to take the first turn", actor, target)
	}
	high := ""
	if len(e.Results) > 0 {
		high = fmt.Sprintf(" with %d", e.Results[0])
	}
	seats := seatList(e.seatNames)
	switch e.Label {
	case game.OpeningRollTie:
		return fmt.Sprintf("%s tied%s and roll again", seats, high)
	case game.OpeningRollRolledFor:
		if e.Seat == NoSeat {
			actor = "The admin"
		}
		return fmt.Sprintf("%s rolled for %s", actor, seats)
	}
	return fmt.Sprintf("%s won the opening roll%s", actor, high)
}

// seatList joins player names as a sentence does: "Alice",
// "Alice and Carol", "Alice, Bob and Carol".
func seatList(names []string) string {
	parts := make([]string, len(names))
	for i, n := range names {
		parts[i] = nameOr(n, "someone")
	}
	switch len(parts) {
	case 0:
		return "nobody"
	case 1:
		return parts[0]
	}
	return strings.Join(parts[:len(parts)-1], ", ") + " and " + parts[len(parts)-1]
}

func renderRandomLogText(e LogEvent, actor, card string) string {
	source := ""
	if e.CardID != "" {
		source = " for " + card
	}
	if e.Kind == LogRoll {
		results := make([]string, len(e.Results))
		for i, n := range e.Results {
			results[i] = strconv.Itoa(n)
		}
		dice := fmt.Sprintf("a d%d", e.Sides)
		if len(e.Results) != 1 {
			dice = fmt.Sprintf("%dd%d", len(e.Results), e.Sides)
		}
		return fmt.Sprintf("%s rolled %s%s: %s", actor, dice, source, strings.Join(results, ", "))
	}
	faces := strings.Join(e.Faces, ", ")
	if e.Call == "" {
		return fmt.Sprintf("%s flipped%s: %s", actor, source, faces)
	}
	outcome := fmt.Sprintf("%d won, %d lost", e.Wins, len(e.Faces)-e.Wins)
	if len(e.Faces) == 1 {
		outcome = "lost"
		if e.Wins == 1 {
			outcome = "won"
		}
	}
	return fmt.Sprintf("%s called %s%s: %s — %s", actor, e.Call, source, faces, outcome)
}
