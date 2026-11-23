package effectstest

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// answerAnchorWord answers the open as-enters prompt with `word`
// through the wire resolver — the prompt is half of what is tested.
func AnswerAnchorWord(t *testing.T, g *game.Game, chooser uuid.UUID, word string) {
	t.Helper()
	c := PendingOfKind(g, game.PendingChoiceOptionPick)
	if c == nil {
		t.Fatalf("no as-enters option prompt is open: %+v", g.PendingChoices)
	}
	for i, opt := range c.PickOptions {
		if opt.Label == word {
			if err := g.ResolveOptionPick(c.ID, chooser, i); err != nil {
				t.Fatalf("ResolveOptionPick: %v", err)
			}
			return
		}
	}
	t.Fatalf("%q is not offered by %v", word, OptionWords(c))
}

func OptionWords(c *game.PendingChoice) []string {
	out := make([]string, 0, len(c.PickOptions))
	for _, opt := range c.PickOptions {
		out = append(out, opt.Label)
	}
	return out
}

// castSiege casts the named Siege from the active seat's hand, lets it
// resolve, and answers its prompt with `word`.
func CastSiege(t *testing.T, g *game.Game, name, oracle, word string) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := CastCatalogSpell(t, g, name, "Enchantment", oracle, nil)
	PassPriorityAroundTable(t, g)
	AnswerAnchorWord(t, g, active.ID, word)
	if got := g.ChosenOptionOf(id); got != word {
		t.Fatalf("%s chose %q, want %q", name, got, word)
	}
	return id
}

// siegeTriggerKeys is the stack labels of the triggers the permanent
// has RIGHT NOW — the list ADR 0071's gate filters.
func SiegeTriggerKeys(g *game.Game, id uuid.UUID) []string {
	var out []string
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID != id {
				continue
			}
			for _, tr := range game.TriggersForCard(c) {
				out = append(out, tr.Key)
			}
		}
	})
	return out
}

func KeysMention(keys []string, sub string) bool {
	for _, k := range keys {
		if strings.Contains(k, sub) {
			return true
		}
	}
	return false
}

// plusOneCounters is the +1/+1 counter count on a battlefield card.
func PlusOneCounters(g *game.Game, id uuid.UUID) int {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return c.Counters["+1/+1"]
		}
	}
	return -1
}

// battlefieldTapped reports whether the battlefield card is tapped.
func BattlefieldTapped(g *game.Game, id uuid.UUID) bool {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return c.Tapped
		}
	}
	return false
}

// emitETBForTest announces a pushed permanent's entry the way the
// entry pipeline would, so an "enters" trigger sees it.
func EmitETBForTest(g *game.Game, controller, id uuid.UUID) {
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, Actor: controller, CardID: id})
	})
}
