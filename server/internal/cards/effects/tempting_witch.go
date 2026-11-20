package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tempting Witch — Creature — Human Warlock {2}{B}, 1/3 (Throne of Eldraine, #5516):
//
//	"When this creature enters, create a Food token.
//	 {2}, {T}, Sacrifice a Food: Target player loses 3 life."
//
// Sweettooth Witch with a {T} in the cost and 3 life.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "82284fe4-ecb3-4b27-bb03-27290ad15b17",
		Name:         "Tempting Witch",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Tempting Witch — create a Food token", Do(CreateToken{Template: FoodToken(), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}, Sacrifice a Food: Target player loses 3 life.",
			Cost:    Plus(ManaCost("{2}"), TapCost(), game.AbilityCost{SacrificeOther: sacrificeSpec("a Food", HasSubtype("Food"))}),
			Targets: TargetPlayer("target player"),
			Effect:  Do(TargetPlayerLosesLife{N: 3}),
		}},
	})
}
