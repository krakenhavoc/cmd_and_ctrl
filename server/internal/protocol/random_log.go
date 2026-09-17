package protocol

import (
	"fmt"
	"strconv"
	"strings"
)

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
