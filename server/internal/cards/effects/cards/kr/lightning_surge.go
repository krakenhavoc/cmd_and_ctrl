package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lightning Surge — Sorcery {3}{R}{R}:
//
//	"Lightning Surge deals 4 damage to any target.
//	 Threshold — If there are seven or more cards in your graveyard,
//	 instead Lightning Surge deals 6 damage to that permanent or player
//	 and the damage can't be prevented.
//	 Flashback {5}{R}{R}"
//
// Threshold is read as the spell resolves, for both the amount and the
// spell's own "can't be prevented" (SpellThreshold, ADR 0107 §5). Cast
// with flashback, the card is on the stack, not in the graveyard, so it
// does not count itself — the printed rule.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "dc55e69e-e1b8-4129-902c-c71bcb952418",
		Name:                       "Lightning Surge",
		Completeness:               effects.CompletenessFull,
		SpellDamageCantBePrevented: effects.SpellThreshold(),
		CastableZones:              []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts:           []game.AlternativeCost{effects.Flashback("{5}{R}{R}")},
		Targets:                    effects.TargetAny(),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			amount := 4
			if effects.SpellThreshold()(ctx.Game, item) {
				amount = 6
			}
			return effects.DamageToFirstTargetStep(amount)(item, ctx)
		},
	})
}
