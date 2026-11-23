package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// target_ability_cards_test.go — #1211's card proofs: the cards that
// could not be written while a target clause could only name a card,
// and the two whose "or ability" half shipped as a caveat.
//
// One board shape serves nearly all of them: an opponent's triggered
// or activated ability on the stack, an answer cast on top of it, and
// the question of whether the ability still resolves.

const (
	StifleOracle    = "b3b00911-ece7-4484-bc36-f211ce72b6cc"
	TalesEndOracle  = "8ebb6fe0-3ed1-4cc9-bcb1-5317de199efc"
	VoidslimeOracle = "c4851768-e210-4aaf-935e-715942e198f9"
)

// taAnnounce puts a triggered ability on the stack for `controller`,
// with an Effect that flips `fired` when it resolves — which is how
// every test below tells "countered" from "resolved" without needing
// a real card behind the trigger.
func TaAnnounce(t *testing.T, g *game.Game, controller, source uuid.UUID, label string,
	targets []game.TargetRef, fired *bool) uuid.UUID {
	t.Helper()
	if err := g.AnnounceTrigger(controller, source, game.AbilityParams{
		Label:   label,
		Targets: targets,
	}); err != nil {
		t.Fatalf("AnnounceTrigger: %v", err)
	}
	id := AcItemLabelled(g, label)
	if id == uuid.Nil {
		t.Fatalf("the announced trigger %q is not on the stack", label)
	}
	if fired != nil {
		g.WithWriteLock(func() {
			g.StackMeta[id].Effect = func(_ *game.Game, _ *game.StackItem) error {
				*fired = true
				return nil
			}
		})
	}
	return id
}

// taLegalTargets is the set a SPELL's own target clause offers right
// now, as a membership map.
func TaLegalTargets(g *game.Game, chooser uuid.UUID, oracle string) map[uuid.UUID]bool {
	spec, ok := effects.Lookup(oracle)
	if !ok {
		return nil
	}
	return TaLegalForSpec(g, chooser, spec.Targets)
}

func TaLegalForSpec(g *game.Game, chooser uuid.UUID, spec *game.TargetSpec) map[uuid.UUID]bool {
	out := map[uuid.UUID]bool{}
	g.WithWriteLock(func() {
		for _, id := range g.LegalTargetsForEffect(game.SourceChooser(chooser), spec).Cards {
			out[id] = true
		}
	})
	return out
}
