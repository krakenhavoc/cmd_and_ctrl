package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Marchesa's Smuggler — Creature — Human Rogue {U}{R}, 1/1 (Conspiracy, S82
// OLD batch 363, #5557):
//
//	"Dethrone (Whenever this creature attacks the player with the most life
//	 or tied for most life, put a +1/+1 counter on it.)
//	 {1}{U}{R}: Target creature you control gains haste until end of turn and
//	 can't be blocked this turn."
//
// Dethrone is the shared GapB201Dethrone trigger (CR 702.105a). The ability
// pins its target at resolution (CR 611.2c): a haste grant and Rogue's
// Passage's can't-be-blocked restriction, both until end of turn; a target
// that left is skipped (CR 608.2b). Blocks already declared are not undone
// (CR 509.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c3e11b61-896e-4a6b-9b33-286313d4a5e2",
		Name:         "Marchesa's Smuggler",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.GapB201Dethrone("Marchesa's Smuggler — dethrone: put a +1/+1 counter on it"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{U}{R}: Target creature you control gains haste until end of turn and can't be blocked this turn.",
			Purpose: effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 0, 0, "haste")),
			Cost:    effects.ManaCost("{1}{U}{R}"),
			Targets: effects.TargetCreature("target creature you control", effects.YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				if err := (effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"}, Label: "Marchesa's Smuggler — haste"}).Apply(ctx); err != nil {
					return err
				}
				return effects.RestrictUntilEOT{Target: id, Restrictions: game.CantBeBlocked, Label: "Marchesa's Smuggler — can't be blocked"}.Apply(ctx)
			},
		}},
	})
}
