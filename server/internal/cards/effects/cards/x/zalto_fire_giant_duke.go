package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zalto, Fire Giant Duke — Legendary Creature — Giant Barbarian
// {3}{R}{R}, 7/3:
//
//	"Trample
//	 Whenever Zalto is dealt damage, venture into the dungeon."
//
// Damage dealt to Zalto at the same time by several sources is one
// event batch, and it triggers once (OncePerBatch); first-strike damage
// and regular damage are two batches, and two triggers.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fb1c507d-ef94-438e-bee5-29eb48dc8e70",
		Name:            "Zalto, Fire Giant Duke",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.B35SelfWasDealtDamage(ev, source)
			}, effects.VentureLabel("Zalto, Fire Giant Duke"), effects.Venture)),
		},
	})
}
