package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dungeons_forgotten_realms_test.go pins the three Adventures in the
// Forgotten Realms dungeons (ADR 0203 amendment of 2026-10-19, #4100):
// a plain venture offers all three and never the Undercity, and one walk
// per dungeon checks every room's ability against its printed text.

// frVenture has `player` venture into the dungeon (the card-side
// primitive) and settles, answering the dungeon and room prompts with
// `picks` in order: a dungeon key for "choose a dungeon", a room key for
// "choose the next room", a game.TargetRef for a room's target. It
// fails the test if a pick was never asked for.
func FrVenture(t *testing.T, g *game.Game, player uuid.UUID, picks ...any) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := (effects.VentureIntoTheDungeon{Player: player}).Apply(&effects.Context{Game: g}); err != nil {
			t.Errorf("VentureIntoTheDungeon: %v", err)
		}
	})
	if left := FrSettle(t, g, picks...); len(left) > 0 {
		t.Fatalf("picks %v were never asked for", left)
	}
}

// frSettle resolves the stack, answering trigger orders, scries (keep
// the order) and searches (find nothing) itself and the venture and
// target prompts from `picks`. It stops at any other prompt and returns
// the picks it never used.
func FrSettle(t *testing.T, g *game.Game, picks ...any) []any {
	t.Helper()
	for i := 0; i < 128; i++ {
		var open *game.PendingChoice
		for _, c := range g.PendingChoices {
			if c != nil {
				open = c
				break
			}
		}
		if open != nil {
			var err error
			switch open.Kind {
			case game.PendingChoiceTriggerOrder:
				err = g.ResolveTriggerOrder(open.ID, open.Chooser, append([]uuid.UUID(nil), open.TriggerOrderIDs...))
			case game.PendingChoiceSearchLibrary:
				err = g.ResolveSearchLibrary(open.ID, open.Chooser, nil)
			case game.PendingChoiceScry:
				err = g.ResolveScry(open.ID, open.Chooser, nil, append([]uuid.UUID(nil), open.ScryCards...))
			case game.PendingChoiceOptionPick:
				if len(picks) == 0 {
					return picks
				}
				key, ok := picks[0].(string)
				if !ok {
					return picks
				}
				idx := -1
				for j, o := range open.PickOptions {
					if (o.Room != "" && o.Room == key) || (o.Room == "" && o.Dungeon == key) {
						idx = j
					}
				}
				if idx < 0 {
					t.Fatalf("%q not offered: %+v", key, open.PickOptions)
				}
				err = g.ResolveOptionPick(open.ID, open.Chooser, idx)
				picks = picks[1:]
			case game.PendingChoicePickTarget:
				if len(picks) == 0 {
					return picks
				}
				ref, ok := picks[0].(game.TargetRef)
				if !ok {
					return picks
				}
				err = g.ResolvePickTargets(open.ID, open.Chooser, []game.TargetRef{ref})
				picks = picks[1:]
			default:
				return picks
			}
			if err != nil {
				t.Fatalf("answering %s: %v", open.Kind, err)
			}
			continue
		}
		if StackFullyEmpty(g) {
			return picks
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority iter %d: %v", i, err)
		}
	}
	t.Fatal("the stack did not settle")
	return picks
}

// frTokens counts `controller`'s tokens named `name`.
func FrTokens(g *game.Game, controller uuid.UUID, name string) int {
	return MonTokens(g, controller, name)
}

// frCard targets a card.
func FrCard(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetCard, ID: id} }

// frPushTop puts a card on top of `p`'s library.
func FrPushTop(p *game.Player, name, typeLine, manaCost string) uuid.UUID {
	c := game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: manaCost, Owner: p.ID, Controller: p.ID}
	p.Library.PushTop(c)
	return c.InstanceID
}
