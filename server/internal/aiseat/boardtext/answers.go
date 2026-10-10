package boardtext

import (
	"encoding/json"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// AnswersPrefix starts a move's answers note: "answers: protect".
const AnswersPrefix = "answers: "

// MoveAnswers is what an activation's row declares it answers (ADR 0142
// decision 6): the wire names from `activated_abilities[].purpose.answers`
// on the permanent the move activates, nil for any other move or for a
// row that declares none. Read from the view alone, so the bot's prompt
// and the MCP seat show what a human sees in the stop.
func MoveAnswers(v *protocol.GameView, m legal.Move) []string {
	if m.Kind != legal.KindActivate || v == nil {
		return nil
	}
	var p struct {
		Source string `json:"source_card_id"`
		Index  int    `json:"ability_index"`
	}
	if json.Unmarshal(m.Params, &p) != nil || p.Source == "" {
		return nil
	}
	for i := range v.Battlefield.Cards {
		c := &v.Battlefield.Cards[i]
		if c.InstanceID != p.Source {
			continue
		}
		for j := range c.ActivatedAbilities {
			if r := &c.ActivatedAbilities[j]; r.Index == p.Index && r.Purpose != nil && r.Purpose.Answers != nil {
				return *r.Purpose.Answers
			}
		}
		return nil
	}
	return nil
}

// AnswersNote is MoveAnswers as the text a move line carries,
// "answers: protect, pump", or "" when the row declares none.
func AnswersNote(v *protocol.GameView, m legal.Move) string {
	a := MoveAnswers(v, m)
	if len(a) == 0 {
		return ""
	}
	return AnswersPrefix + strings.Join(a, ", ")
}
