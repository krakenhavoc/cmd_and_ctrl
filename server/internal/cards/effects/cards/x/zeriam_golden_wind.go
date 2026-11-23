package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "fa1dc21d-f7f6-425f-a4ae-2b0f143091e3",
		Name:            "Zeriam, Golden Wind",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B24CombatDamageToPlayerByYourCreatureOfSubtypes(ev, source, g, source.TextSubtype("Griffin"))
			}, "Zeriam, Golden Wind — create a 2/2 Griffin with flying", effects.Do(effects.CreateToken{
				Template: effects.TokenCard("2/2 white Griffin with flying"),
				N:        1,
			})),
		},
	})
}
