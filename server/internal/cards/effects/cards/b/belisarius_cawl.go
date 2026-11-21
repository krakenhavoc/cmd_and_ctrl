package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Belisarius Cawl — Legendary Artifact Creature — Human {2}{W}{U}, 2/4
// (Warhammer 40,000 Commander, S82 GAP batch 195, #5948):
//
//	"Ultima Founding — {T}, Tap two untapped artifacts you control: Create a 2/2
//	 white Astartes Warrior creature token with vigilance.
//	 Master of Machines — {T}, Tap X untapped creatures you control: Look at the
//	 top X cards of your library. You may reveal an artifact card from among them
//	 and put it into your hand. Put the rest on the bottom of your library in a
//	 random order."
//
// Both abilities' {T} is the Cawl's own, so he is summoning-sick the turn he
// arrives (CR 302.6), and the permanents tapped for the rest of the cost are
// others (the Cawl, already tapped by his {T}, cannot pay twice). The first is
// Tenured Tethermage's fixed tap-two cost; the second is Apothecary White's
// variable "Tap X untapped creatures" (the announced X is the width of the
// payment), with Horn of the Mark's look-and-take, which reveals the artifact
// taken and puts the rest on the bottom in a random order. The first ability
// makes a creature at instant speed; the second answers nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6b336228-92c0-4f14-a479-1f97aaeae018",
		Name:         "Belisarius Cawl",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}, Tap two untapped artifacts you control: Create a 2/2 white Astartes Warrior creature token with vigilance.",
				Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
				Cost: effects.Plus(effects.TapCost(), game.AbilityCost{TapOthers: &game.TapOthersCost{
					Count:  2,
					Filter: effects.TargetPermanent("two untapped artifacts you control", effects.Artifact()),
					Label:  "two untapped artifacts you control",
				}}),
				Effect: effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 white Astartes Warrior with vigilance"), N: 1}),
			},
			{
				Label:   "{T}, Tap X untapped creatures you control: Look at the top X cards of your library. You may reveal an artifact card from among them and put it into your hand. Put the rest on the bottom of your library in a random order.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.TapCost(), effects.TapXUntapped("X untapped creatures you control", effects.Creature())),
				Effect:  effects.GapB195MasterOfMachines,
			},
		},
	})
}
