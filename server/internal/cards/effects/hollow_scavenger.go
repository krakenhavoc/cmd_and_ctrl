package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hollow Scavenger // Bakery Raid — Creature — Wolf {2}{G}, 3/2 //
// Sorcery — Adventure {G} (Wilds of Eldraine, #4471):
//
//	Hollow Scavenger — "{1}, Sacrifice a Food: This creature gets +2/+2
//	 until end of turn. Activate only once each turn."
//	Bakery Raid — "Create a Food token."
//
// The pump is Bosco's "Sacrifice a Food" cost (any Food artifact you
// control) with the once-each-turn limit counted per object (CR 602.1b).
// The Adventure lifecycle is the engine's (CR 715.3d).
//
// No simplification.
const hollowScavengerOracleID = "4fad3816-41af-4fa4-85b4-ba37614ec848"

func init() {
	label := "{1}, Sacrifice a Food: This creature gets +2/+2 until end of turn. Activate only once each turn."
	Register(Spec{
		OracleID:     hollowScavengerOracleID,
		Name:         "Hollow Scavenger",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerPump},
			Cost:      Plus(ManaCost("{1}"), game.AbilityCost{SacrificeOther: sacrificeSpec("a Food", HasSubtype("Food"))}),
			Condition: OncePerTurnActivation(label),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				return thisGetsUntilEndOfTurn(2, 2, "Hollow Scavenger — +2/+2 until end of turn")(g, item)
			},
		}},
	})
	Register(Spec{
		OracleID:     hollowScavengerOracleID + "#1",
		Name:         "Bakery Raid",
		Completeness: CompletenessFull,
		OnResolve:    ResolveSteps(CreateToken{Template: FoodToken(), N: 1}),
	})
}
