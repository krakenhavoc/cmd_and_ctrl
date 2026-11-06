package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Grabby Giant // That's Mine — Creature — Giant {3}{R}, 4/3 // Instant —
// Adventure {1}{R} (Wilds of Eldraine, #4471):
//
//	Grabby Giant — "Reach
//	 {2}{R}, Sacrifice an artifact or land: Draw a card."
//	That's Mine — "Create a Treasure token. (Then exile this card. You may
//	 cast the creature later from exile.)"
//
// Both faces register under the Adventure keyspace (the bare oracle ID
// and "#1"). The draw is Marwyn, the Clearcutter's sacrifice cost without
// the {T}, so it can be activated while summoning sick and more than once
// a turn; the sacrifice may be the Treasure That's Mine made. That's Mine
// is Craft with Pride's Treasure (CR 111.10a); the Adventure lifecycle is
// the engine's (CR 715.3d).
//
// No simplification.
const grabbyGiantOracleID = "e6f59807-eaf6-4889-8c89-2ac915afefff"

func init() {
	Register(Spec{
		OracleID:        grabbyGiantOracleID,
		Name:            "Grabby Giant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated: []ActivatedAbility{{
			Label:   "{2}{R}, Sacrifice an artifact or land: Draw a card.",
			Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
			Cost: Plus(ManaCost("{2}{R}"),
				game.AbilityCost{SacrificeOther: sacrificeSpec("an artifact or land", Or(Artifact(), Land()))}),
			Effect: Do(DrawCards{N: 1}),
		}},
	})
	Register(Spec{
		OracleID:     grabbyGiantOracleID + "#1",
		Name:         "That's Mine",
		Completeness: CompletenessFull,
		OnResolve:    ResolveSteps(CreateToken{Template: TreasureToken(), N: 1}),
	})
}
