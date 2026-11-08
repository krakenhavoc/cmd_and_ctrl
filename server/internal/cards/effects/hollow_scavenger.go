package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hollow Scavenger // Bakery Raid — Creature — Wolf {2}{G}, 3/2 //
// Sorcery — Adventure {G} (Wilds of Eldraine, #4471):
//
//	Hollow Scavenger — "{1}, Sacrifice a Food: This creature gets +2/+2
//	 until end of turn. Activate only once each turn."
//	Bakery Raid — "Create a Food token. (Then exile this card. You may
//	 cast the creature later from exile.)"
//
// Both faces register under the Adventure keyspace (the bare oracle ID
// and "#1"). The pump sacrifices any Food as a cost (Greta, Sweettooth
// Scourge's cost) and is Mindful Biomancer's once-each-turn +2/+2: the
// per-object activation tally (OncePerTurnActivation, CR 602.1b) refuses
// a second activation this turn, and the boost lands only on the
// Scavenger the ability came from (CR 400.7). Bakery Raid makes the
// catalog's Food token (CR 111.10b); the Adventure lifecycle is the
// engine's (CR 715.3d).
//
// No simplification.
const hollowScavengerOracleID = "4fad3816-41af-4fa4-85b4-ba37614ec848"

func init() {
	const label = "{1}, Sacrifice a Food: This creature gets +2/+2 until end of turn. Activate only once each turn."
	Register(Spec{
		OracleID:     hollowScavengerOracleID,
		Name:         "Hollow Scavenger",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   label,
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 2, Toughness: 2}},
			Cost: Plus(ManaCost("{1}"),
				game.AbilityCost{SacrificeOther: sacrificeSpec("a Food", HasSubtype("Food"))}),
			Condition: OncePerTurnActivation(label),
			Effect:    thisGetsUntilEndOfTurn(2, 2, "Hollow Scavenger — +2/+2 until end of turn"),
		}},
	})
	Register(Spec{
		OracleID:     hollowScavengerOracleID + "#1",
		Name:         "Bakery Raid",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		OnResolve:    ResolveSteps(CreateToken{Template: FoodToken(), N: 1}),
	})
}
