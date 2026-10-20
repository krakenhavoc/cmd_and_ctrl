package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Dominion Bracelet — Legendary Artifact — Equipment {2}:
//
//	"Equipped creature gets +1/+1 and has "{15}, Exile The Dominion
//	 Bracelet: You control target opponent during their next turn. This
//	 ability costs {X} less to activate, where X is this creature's
//	 power. Activate only as a sorcery." (You see all cards that player
//	 could see and make all decisions for them.)
//	 Equip {1}"
//
// The quoted ability is the equipped creature's (a layer-6 grant, ADR
// 0093), so two of its words point away from the Bracelet:
//
//   - "Exile The Dominion Bracelet" is the row's grantor
//     (ExileTheGrantor, #4000), paid at announce, so the control
//     survives the Bracelet being gone and the creature keeps nothing.
//   - "this creature's power" is the equipped creature's, read as the
//     cost is determined (CR 601.2f via 602.2b): the ability's own cost
//     clause is bound to the ability's source, which is the creature.
//
// The control is the plain turn-long one Mindslaver gives (ADR 0167,
// no extra turn). Sorcery timing on a granted row is the recipient
// controller's (activation_timing_grant.go).
//
// No simplification.
const dominionBraceletGrant = "the-dominion-bracelet/control"

func init() {
	Register(Spec{
		OracleID:     "c017f8ae-9103-4965-a07a-f259f44a157d",
		Name:         "The Dominion Bracelet",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: dominionBraceletGrant,
			Activated: []ActivatedAbility{{
				Label:        "{15}, Exile The Dominion Bracelet: You control target opponent during their next turn.",
				Cost:         Plus(ManaCost("{15}"), ExileTheGrantor()),
				Targets:      TargetPlayer("target opponent", Opponent()),
				SorcerySpeed: true,
				CostModifiers: []game.CostModifier{
					CostsLessEach(func(q game.CostQuery) int {
						return max(0, q.Card.PowerForComparison())
					}, "This ability costs {X} less to activate, where X is this creature's power."),
				},
				Effect: controlTargetPlayerNextTurn("The Dominion Bracelet"),
			}},
			Text: "{15}, Exile The Dominion Bracelet: You control target opponent during their next turn. This ability costs {X} less to activate, where X is this creature's power. Activate only as a sorcery.",
		}},
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			GrantAbilitiesToAttached(dominionBraceletGrant),
		},
		Activated: []ActivatedAbility{EquipAbility("{1}")},
	})
}
