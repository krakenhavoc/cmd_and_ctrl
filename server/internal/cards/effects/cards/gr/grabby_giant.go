package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
const GrabbyGiantOracleID = "e6f59807-eaf6-4889-8c89-2ac915afefff"

func init() {
	effects.Register(effects.Spec{
		OracleID:        GrabbyGiantOracleID,
		Name:            "Grabby Giant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}, Sacrifice an artifact or land: Draw a card.",
			Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
			Cost: effects.Plus(effects.ManaCost("{2}{R}"),
				game.AbilityCost{SacrificeOther: effects.SacrificeSpec("an artifact or land", effects.Or(effects.Artifact(), effects.Land()))}),
			Effect: effects.Do(effects.DrawCards{N: 1}),
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     GrabbyGiantOracleID + "#1",
		Name:         "That's Mine",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.ResolveSteps(effects.CreateToken{Template: effects.TreasureToken(), N: 1}),
	})
}
