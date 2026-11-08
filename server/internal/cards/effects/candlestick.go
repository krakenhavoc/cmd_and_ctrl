package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const candlestickSurveil = "candlestick/attack-surveil"

// Candlestick — Artifact — Clue Equipment {U} (Murders at Karlov Manor,
// #4469):
//
//	"Equipped creature gets +1/+1 and has 'Whenever this creature
//	 attacks, surveil 2.'
//	 {2}, Sacrifice this Equipment: Draw a card.
//	 Equip {2}"
//
// The attack trigger is GRANTED to the creature (ADR 0093), so its
// controller surveils, as Thornbite Staff's grants work. The draw is
// printed on the card, and sacrificing it is sacrificing a Clue for
// every card that watches that.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2f7b46c1-3d82-4fd9-9882-589315cab47b",
		Name:         "Candlestick",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: candlestickSurveil,
			Triggered: []game.TriggeredAbility{
				WheneverThisAttacks("Candlestick — surveil 2", Do(Surveil{N: 2})),
			},
			Text: "Whenever this creature attacks, surveil 2.",
		}},
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			GrantAbilitiesToAttached(candlestickSurveil),
		},
		Activated: []ActivatedAbility{
			{
				Label:   "{2}, Sacrifice this Equipment: Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
				Cost:    Plus(ManaCost("{2}"), SacrificeThis()),
				Effect:  Do(DrawCards{N: 1}),
			},
			EquipAbility("{2}"),
		},
	})
}
