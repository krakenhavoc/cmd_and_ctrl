package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "fb1c507d-ef94-438e-bee5-29eb48dc8e70",
		Name:            "Zalto, Fire Giant Duke",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return b35SelfWasDealtDamage(ev, source)
			}, ventureLabel("Zalto, Fire Giant Duke"), Venture)),
		},
	})
}
