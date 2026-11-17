package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cyclopean Snare — Artifact {2} (Ravnica: City of Guilds, #5600):
//
//	"{3}, {T}: Tap target creature, then return this artifact to its
//	 owner's hand."
//
// Trusty Boomerang's throw on the artifact itself: the target is tapped,
// then the Snare goes back to its owner's hand. A target gone by
// resolution makes the ability do nothing at all (CR 608.2b), so the Snare
// stays. A Snare that has already left the battlefield, or come back as a
// new object, is not returned.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6f5a86bf-caa7-48c9-a59f-b13fb483a9d4",
		Name:         "Cyclopean Snare",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{3}, {T}: Tap target creature, then return this artifact to its owner's hand.",
			Cost:    Plus(ManaCost("{3}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Effect:  cyclopeanSnareTapThenReturn,
		}},
	})
}

// cyclopeanSnareTapThenReturn taps the target, then bounces the Snare.
func cyclopeanSnareTapThenReturn(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target := FirstLegalBattlefieldTarget(ctx)
	if target == uuid.Nil {
		return nil
	}
	if err := (TapTarget{Target: target}).Apply(ctx); err != nil {
		return err
	}
	return b28ReturnSelfToOwnersHand(g, item)
}
