package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ominous Roost — Enchantment {2}{U} (Innistrad: Double Feature):
//
//	"When this enchantment enters and whenever you cast a spell from your
//	 graveyard, create a 1/1 blue Bird creature token with flying and 'This
//	 token can block only creatures with flying.'"
//
// One printed ability with two conditions, so two declarations sharing one
// body (Autarch Mammoth). The cast condition reads the zone the spell was
// cast from and that the card is the caster's own (a card in a graveyard is
// in its owner's, CR 404.1), as Devoted Grafkeeper's does. The Bird's block
// restriction is on the token's template (ADR 0083).
//
// No simplification.
func init() {
	const label = "Ominous Roost — create a 1/1 blue Bird with flying that can block only creatures with flying"
	makeBird := effects.Do(effects.CreateToken{Template: effects.S82MidB11BirdToken(), N: 1})
	effects.Register(effects.Spec{
		OracleID:     "745fc118-ffab-45c6-b51a-7bef2f6931e8",
		Name:         "Ominous Roost",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(label, makeBird),
			effects.On(game.EventCast, effects.YouCastASpellFromYourGraveyard, label, makeBird),
		},
	})
}
