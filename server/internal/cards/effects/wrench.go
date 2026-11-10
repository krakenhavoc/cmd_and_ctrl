package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const wrenchTapGrant = "wrench/tap-target-creature"

// Wrench — Artifact — Clue Equipment {W} (Murders at Karlov Manor,
// #4469):
//
//	"Equipped creature gets +1/+1 and has vigilance and "{3}, {T}: Tap
//	 target creature."
//	 {2}, Sacrifice this Equipment: Draw a card.
//	 Equip {2}"
//
// Candlestick's shape: the anthem and vigilance are statics on the
// attached creature, the quoted ability is an ADR 0093 grant (the
// creature's own activated ability, so it taps to pay and summoning
// sickness applies, CR 302.6). The draw is printed on the card, so
// sacrificing it is sacrificing a Clue.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fdbac0a7-18c0-48a5-a174-0d4d1473fd9c",
		Name:         "Wrench",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: wrenchTapGrant,
			Activated: []ActivatedAbility{{
				Label:   "{3}, {T}: Tap target creature.",
				Cost:    Plus(ManaCost("{3}"), TapCost()),
				Targets: TargetCreature("target creature"),
				Effect:  tapChosenPermanent,
			}},
			Text: "{3}, {T}: Tap target creature.",
		}},
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			GrantToAttached("vigilance"),
			GrantAbilitiesToAttached(wrenchTapGrant),
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
