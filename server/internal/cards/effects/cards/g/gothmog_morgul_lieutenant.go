package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gothmog, Morgul Lieutenant — Legendary Creature — Human Soldier {3}{B}, 3/3
// (The Lord of the Rings: Tales of Middle-earth):
//
//	"When Gothmog enters, amass Orcs 1.
//	 Creature tokens you control have deathtouch."
//
// Aven Eternal's amass (the Army is found or made, CR 701.47) and a layer-6
// keyword grant over the creature tokens its controller controls, the
// Army token it makes included, read live.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f3817f63-7bde-46c5-93cb-b4756f5f821e",
		Name:         "Gothmog, Morgul Lieutenant",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Amass: 1},
		Static: []game.StaticAbility{effects.KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.IsToken() && target.IsCreature() && target.Controller == source.Controller
		}, "deathtouch")},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Gothmog, Morgul Lieutenant — amass Orcs 1", effects.Do(effects.Amass{Subtype: "Orc", N: 1})),
		},
	})
}
