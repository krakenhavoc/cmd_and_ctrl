package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Prismari Apprentice — Creature — Human Shaman {U}{R}, 2/2 (Strixhaven:
// School of Mages, #5502):
//
//	"Magecraft — Whenever you cast or copy an instant or sorcery spell, this
//	 creature can't be blocked this turn. If that spell has mana value 5 or
//	 greater, put a +1/+1 counter on this creature."
//
// Both halves land on the Apprentice only while it is still the permanent
// the trigger came from (CR 400.7). The mana value is read off the spell
// that triggered it as the ability resolves, with an announced X counted
// (CR 202.3b), and a copy has its original's mana value (CR 707.10).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0d60ffba-9904-442e-8d4a-a6b1b50a38eb",
		Name:         "Prismari Apprentice",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Magecraft("Prismari Apprentice — can't be blocked this turn; if that spell has mana value 5 or greater, put a +1/+1 counter on it (magecraft)",
				func(g *game.Game, item *game.StackItem) error {
					if !effects.SourceIsStillThisPermanent(g, item) {
						return nil
					}
					ctx := effects.NewContext(g, item)
					if err := (effects.RestrictUntilEOT{Target: item.SourceCardID, Restrictions: game.CantBeBlocked,
						Label: "Prismari Apprentice — can't be blocked this turn"}).Apply(ctx); err != nil {
						return err
					}
					if effects.TriggeringSpellManaValue(g, item) < 5 {
						return nil
					}
					return effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
				}),
		},
	})
}
