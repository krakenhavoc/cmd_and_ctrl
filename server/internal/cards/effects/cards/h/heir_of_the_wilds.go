package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Heir of the Wilds — Khans of Tarkir (S82 OLD b347, #5554):
//
//	Creature — Human Warrior {1}{G}, 2/2: Deathtouch. Ferocious — Whenever this
//	creature attacks, if you control a creature with power 4 or greater, this
//	creature gets +1/+1 until end of turn.
//
// The "if" is an intervening-if (CR 603.4): checked when the attack is
// declared and again as the trigger resolves. Ferocious is read through YouControlPowerFourOrGreater (current power, layered).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "35cb88aa-6ed9-4a2d-9bbc-5937e1df4f4a",
		Name:            "Heir of the Wilds",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.CardID == source.InstanceID && effects.YouControlPowerFourOrGreater(g, source.Controller)
			}, "Heir of the Wilds — ferocious: this creature gets +1/+1 until end of turn", func(g *game.Game, item *game.StackItem) error {
				if !effects.YouControlPowerFourOrGreater(g, item.Controller) {
					return nil
				}
				return effects.BoostUntilEOT{Target: item.SourceCardID, Power: 1, Toughness: 1, Label: "Heir of the Wilds — +1/+1"}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
