package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hero of Oxid Ridge — Creature — Human Knight {2}{R}{R}, 4/2 (Mirrodin
// Besieged, S82 OLD b467, #5576):
//
//	"Haste
//	 Battle cry (Whenever this creature attacks, each other attacking creature
//	 gets +1/+0 until end of turn.)
//	 Whenever this creature attacks, creatures with power 1 or less can't block
//	 this turn."
//
// Three printed abilities: haste, and two separate attack triggers, the battle
// cry (the shared BattleCry row) and the can't-block rider, so the controller
// orders them (CR 603.3b). "Can't block" changes no characteristic: the set is
// every creature, anyone's, whose power is 1 or less, read live as blockers
// are declared (CR 611.2c; Mournwillow's query), so a creature pumped above 1
// by then may block and one shrunk to 1 may not.
//
// No simplification.
func init() {
	const name = "Hero of Oxid Ridge"
	effects.Register(effects.Spec{
		OracleID:        "f18799c4-ec54-4435-be51-03ea25809289",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.BattleCry(name),
			effects.WheneverThisAttacks(name+" — creatures with power 1 or less can't block this turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.CantBlockThisTurnMatching(false,
						name+" — creatures with power 1 or less can't block this turn",
						game.PermanentQuery{Types: []string{"creature"}}.WithPowerAtMost(1)).Apply(effects.NewContext(g, item))
				}),
		},
	})
}
