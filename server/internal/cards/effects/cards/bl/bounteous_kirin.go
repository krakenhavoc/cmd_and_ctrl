package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bounteous Kirin — Legendary Creature — Kirin Spirit {5}{G}{G}, 4/4 (Saviors
// of Kamigawa, #5601):
//
//	"Flying
//	 Whenever you cast a Spirit or Arcane spell, you may gain life equal to
//	 that spell's mana value."
//
// The ki cycle's cast trigger (Spirit or Arcane, read on the spell's type
// line), made a "you may" (CR 603.5). The amount is the spell's mana value as
// the trigger resolves, with the X it was announced with while it is still on
// the stack (CR 608.2h).
//
// No simplification.
func init() {
	const label = "Bounteous Kirin — you may gain life equal to that spell's mana value"
	effects.Register(effects.Spec{
		OracleID:        "31d572b8-2c5e-4819-b2a6-e7ef90872d9e",
		Name:            "Bounteous Kirin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.WheneverYouCast(effects.ChkB1SpiritOrArcane(), label, effects.OldB586GainLifeEqualToThatSpellsManaValue),
				"Bounteous Kirin — gain life equal to that spell's mana value?"),
		},
	})
}
