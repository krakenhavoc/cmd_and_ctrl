package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zeriam, Golden Wind — Legendary Creature — Griffin {3}{W}, 3/4
// (Dominaria United Commander, S82):
//
//	"Flying
//	 Whenever a Griffin you control deals combat damage to a player,
//	 create a 2/2 white Griffin creature token with flying."
//
// Flying rides PrintedKeywords. The trigger is one per Griffin that
// connects (Zeriam included), read through the effective subtypes so a
// changeling counts; no "one or more" batching because the printed card
// has none.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "fa1dc21d-f7f6-425f-a4ae-2b0f143091e3",
		Name:            "Zeriam, Golden Wind",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b24CombatDamageToPlayerByYourCreatureOfSubtypes(ev, source, g, source.TextSubtype("Griffin"))
			}, "Zeriam, Golden Wind — create a 2/2 Griffin with flying", Do(CreateToken{
				Template: TokenCard("2/2 white Griffin with flying"),
				N:        1,
			})),
		},
	})
}
