package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Eternal Scourge — {3} Creature — Eldrazi Horror 3/3 (Eldritch Moon, S82
// OLD, #5541):
//
//	"You may cast this card from exile.
//	 When this creature becomes the target of a spell or ability an opponent
//	 controls, exile this creature."
//
// The first line is the card's own text (Khan, Vengeful Heart's
// CastsItselfFromExile): an ordinary cast at the card's own timing and cost.
// The trigger is Altanak's watch on EventBecomesTarget, held to once per spell
// or ability however many times it names the Scourge, and exiles the Scourge
// as it resolves if it is still the same object (CR 400.7), so the spell or
// ability that targeted it fizzles, and the Scourge can be cast again from
// exile.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:             "c8b0436e-16d2-4d48-a894-454d69a3743b",
		Name:                 "Eternal Scourge",
		Completeness:         effects.CompletenessFull,
		CastableZones:        []game.ZoneKind{game.ZoneExile},
		CastsItselfFromExile: true,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesTarget, effects.B24SelfTargetedByAnOpponentOnce, "Eternal Scourge — exile this creature",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.ExileTarget{Target: ctx.Source()}.Apply(ctx)
				}),
		},
	})
}
