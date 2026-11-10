package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thoughtweft Imbuer — Creature — Kithkin Advisor {3}{W}, 0/5 (Lorwyn
// Eclipsed, #4458):
//
//	"Whenever a creature you control attacks alone, it gets +X/+X until
//	 end of turn, where X is the number of Kithkin you control."
//
// Black Widow, Double Agent's shape: "alone" is game.AttackedAlone read
// when the attack trigger is harvested, and "it" is the attacker named by
// the event, so a creature removed in response gets nothing. X is counted
// once, as the ability resolves (CR 608.2h): the Kithkin its controller
// controls then, the Imbuer included, "Kithkin" read as a word of the
// card's text and a changeling counting as one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9dee8c0b-71e0-4a94-b411-78b29b599847",
		Name:         "Thoughtweft Imbuer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, mshB1YouAttackAlone,
				"Thoughtweft Imbuer — the lone attacker gets +X/+X, X the number of Kithkin you control",
				func(g *game.Game, item *game.StackItem) error {
					attacker := item.Trigger.Event.CardID
					if !onBattlefield(g, attacker) {
						return nil
					}
					ctx := NewContext(g, item)
					word := ctx.TextSubtype("Kithkin")
					x := 0
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == item.Controller && c.HasSubtype(word) {
							x++
						}
					}
					return BoostUntilEOT{Target: attacker, Power: x, Toughness: x,
						Label: "Thoughtweft Imbuer — +X/+X"}.Apply(ctx)
				}),
		},
	})
}
