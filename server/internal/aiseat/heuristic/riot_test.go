package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// riot_test.go is ADR 0109 §10 decision 6: riot takes haste on the
// bot's own turn while the creature could still attack, and the counter
// otherwise; unleash takes the counter unless the creature would be the
// bot's only untapped blocker on an opponent's turn.

const (
	riotChoiceID    = "00000000-0000-4000-8000-000000001556"
	unleashChoiceID = "00000000-0000-4000-8000-000000002556"
)

func entryAnswerMoves(t *testing.T, id string) []legal.Move {
	t.Helper()
	return []legal.Move{
		choiceMove(t, 0, id, "counter", map[string]any{"apply": true}),
		choiceMove(t, 0, id, "no counter", map[string]any{"apply": false}),
	}
}

func withEntryChoice(id, kind, keyword string) viewOpt {
	return withChoice(protocol.PendingChoiceView{
		ID:           id,
		Kind:         kind,
		Chooser:      seatID(0).String(),
		FromPlayer:   seatID(0).String(),
		Count:        1,
		Source:       cardID(9),
		EntryKeyword: keyword,
	})
}

func riotDecision(t *testing.T, active int, step string, entering protocol.CardView) string {
	t.Helper()
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
		withTurn(3, active, step), withStack(entering), withEntryChoice(riotChoiceID, "entry_riot", "riot"))
	in := input(0, v, entryAnswerMoves(t, riotChoiceID)...)
	return chose(t, in, decide(t, heuristic.New(), in))
}

func TestRiotTakesHasteWhenItCanAttackNow(t *testing.T) {
	goblin := creature(cardID(9), 0, "Zhur-Taa Goblin", 2, 2)
	if got := riotDecision(t, 0, "precombat_main", goblin); got != "no counter" {
		t.Errorf("own precombat main: chose %q, want haste", got)
	}
}

func TestRiotTakesTheCounterWhenItCannotAttack(t *testing.T) {
	goblin := creature(cardID(9), 0, "Zhur-Taa Goblin", 2, 2)
	for _, tc := range []struct {
		name   string
		active int
		step   string
		card   protocol.CardView
	}{
		{"after combat", 0, "postcombat_main", goblin},
		{"after attackers are declared", 0, "declare_blockers", goblin},
		{"an opponent's turn", 1, "precombat_main", goblin},
		{"a defender", 0, "precombat_main", creature(cardID(9), 0, "Wall", 0, 4, func(c *protocol.CardView) {
			c.Abilities = []string{"defender", "riot"}
		})},
	} {
		if got := riotDecision(t, tc.active, tc.step, tc.card); got != "counter" {
			t.Errorf("%s: chose %q, want the counter", tc.name, got)
		}
	}
}

func unleashDecision(t *testing.T, active int, board ...protocol.CardView) string {
	t.Helper()
	ogre := creature(cardID(9), 0, "Hellhole Flailer", 3, 2)
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
		withTurn(3, active, "precombat_main"), withStack(ogre), withBattlefield(board...),
		withEntryChoice(unleashChoiceID, "optional_replacement", "unleash"))
	in := input(0, v, entryAnswerMoves(t, unleashChoiceID)...)
	return chose(t, in, decide(t, heuristic.New(), in))
}

func TestUnleashTakesTheCounter(t *testing.T) {
	if got := unleashDecision(t, 0); got != "counter" {
		t.Errorf("own turn, empty board: chose %q, want the counter", got)
	}
	blocker := creature(cardID(1), 0, "Bear", 2, 2)
	if got := unleashDecision(t, 1, blocker); got != "counter" {
		t.Errorf("an opponent's turn with another blocker: chose %q, want the counter", got)
	}
}

func TestUnleashKeepsTheOnlyBlocker(t *testing.T) {
	tapped := creature(cardID(1), 0, "Bear", 2, 2, func(c *protocol.CardView) { c.Tapped = true })
	if got := unleashDecision(t, 1, tapped); got != "no counter" {
		t.Errorf("an opponent's turn, no other untapped creature: chose %q, want no counter", got)
	}
}
