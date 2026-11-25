package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Adeliz, the Cinder Wind — Legendary Creature — Human Wizard {1}{U}{R}, 2/2
// (Dominaria, S82 OLD, #5530):
//
//	"Flying, haste
//	 Whenever you cast an instant or sorcery spell, Wizards you control get
//	 +1/+1 until end of turn."
//
// A one-shot effect over the Wizards its controller controls as the trigger
// resolves (CR 611.2c): Adeliz herself is one, a Wizard that arrives later
// that turn is not pumped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "850eb3c3-029b-49b0-91a9-6daeb1b3a9e8",
		Name:            "Adeliz, the Cinder Wind",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Or(effects.Instant(), effects.Sorcery()), "Adeliz, the Cinder Wind — Wizards you control get +1/+1 until end of turn",
				effects.Do(effects.BoostUntilEOT{
					Match:     effects.And(effects.OfCreatureType("Wizard"), effects.YouControl()),
					Power:     1,
					Toughness: 1,
					Label:     "Adeliz, the Cinder Wind — Wizards get +1/+1 until end of turn",
				})),
		},
	})
}
