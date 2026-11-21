package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Umara Wizard // Umara Skyfalls — Modal DFC. Front: Creature — Merfolk
// Wizard {4}{U}, 4/3: "Whenever you cast an instant, sorcery, or Wizard
// spell, this creature gains flying until end of turn." Back: Land
// (mdfc_lands.go).
//
// The trigger watches the caster's own casts from the battlefield: the
// spell's card has the Instant or Sorcery type or the Wizard creature
// type. A second qualifying spell is a second trigger that grants the
// same flying again. No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6bc668f4-8fc7-4aaf-891b-277d8328b376",
		Name:         "Umara Wizard",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Or(effects.Instant(), effects.Sorcery(), effects.OfSubtype("Wizard")),
				"Umara Wizard — gains flying until end of turn",
				effects.ThisGainsKeywordUntilEndOfTurn("flying", "Umara Wizard — flying until end of turn")),
		},
	})
}
