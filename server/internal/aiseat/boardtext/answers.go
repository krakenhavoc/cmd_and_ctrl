package boardtext

import (
	"encoding/json"
	"strings"

	"github.com/google/uuid"

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

// ActingAsNote is "as Bob" for a move made for another seat: one whose
// Player is a seat the viewer controls (CR 723, ADR 0167 §7: the turn of
// a player the seat is playing, or Word of Command's forced play), and
// "" for the viewer's own moves. Read from the view alone.
func ActingAsNote(v *protocol.GameView, m legal.Move, me string) string {
	if v == nil || m.Player == uuid.Nil || m.Player.String() == me {
		return ""
	}
	for i := range v.Seats {
		if v.Seats[i].ID == m.Player.String() {
			return "as " + SeatLabel(&v.Seats[i], me)
		}
	}
	return "as another player"
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

// RecommendedText is the note RecommendedNote returns.
const RecommendedText = "recommended"

// RecommendedNote is "recommended" for the answer the engine recommends
// to the prompt a move answers, and "" for every other move: a
// "turn_skip_order" question's "use my skip" (ADR 0188 Decision 12,
// PendingChoiceView.TurnSkipOrder.Recommended). Read from the view
// alone, so the bot's prompt and the MCP seat mark the answer the dock
// preselects.
func RecommendedNote(v *protocol.GameView, m legal.Move) string {
	if m.Kind != legal.KindChoice || v == nil {
		return ""
	}
	var p struct {
		ChoiceID string `json:"choice_id"`
		Apply    *bool  `json:"apply"`
	}
	if json.Unmarshal(m.Params, &p) != nil || p.ChoiceID == "" || p.Apply == nil {
		return ""
	}
	for i := range v.PendingChoices {
		ch := &v.PendingChoices[i]
		if ch.ID != p.ChoiceID {
			continue
		}
		// Recommended is the answer's index: 0 is apply true.
		if o := ch.TurnSkipOrder; o != nil && *p.Apply == (o.Recommended == 0) {
			return RecommendedText
		}
		return ""
	}
	return ""
}

// AnswersCouncilVote reports whether a move casts a ballot in a vote
// (CR 701.38, ADR 0146): a choice answering an option_pick that carries
// `council_vote`. Read from the view alone.
//
// The enumerator marks the first option of every option_pick
// AlwaysLegal, because a mandatory vote needs one answer the bot and the
// MCP seat can fall back on (#3669). On a vote that flag means "the
// escape hatch", not "this is the safe default": every ballot is equally
// legal, so a text seat must not be told that one of them is special.
func AnswersCouncilVote(v *protocol.GameView, m legal.Move) bool {
	if m.Kind != legal.KindChoice || v == nil {
		return false
	}
	var p struct {
		ChoiceID string `json:"choice_id"`
	}
	if json.Unmarshal(m.Params, &p) != nil || p.ChoiceID == "" {
		return false
	}
	for i := range v.PendingChoices {
		if ch := &v.PendingChoices[i]; ch.ID == p.ChoiceID {
			return ch.CouncilVote != nil
		}
	}
	return false
}
