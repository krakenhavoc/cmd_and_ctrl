package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Break Through the Line — {1}{R} Enchantment (Fate Reforged):
//
//	"{R}: Target creature with power 2 or less gains haste until end of turn
//	 and can't be blocked this turn."
//
// The target's power is checked as it is chosen and again as the ability
// resolves (CR 608.2b), on the layered board. Both halves apply to the one
// creature.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "555c571e-ccfc-4a8e-8a40-7b1eab0fbbef",
		Name:         "Break Through the Line",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: Target creature with power 2 or less gains haste until end of turn and can't be blocked this turn.",
			Purpose: game.Purpose{Targets: game.ForTargets(effects.CantBeBlockedUntilEOT(0, 0, 0, "haste")), Answers: game.AnswerPump | game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{R}"),
			Targets: effects.TargetCreature("target creature with power 2 or less", effects.PowerLE(2)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				if err := (effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"}, Label: "Break Through the Line — haste"}).Apply(ctx); err != nil {
					return err
				}
				return effects.RestrictUntilEOT{Target: id, Restrictions: game.CantBeBlocked, Label: "Break Through the Line — can't be blocked"}.Apply(ctx)
			},
		}},
	})
}
