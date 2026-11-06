package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fanatical Offering — Instant {1}{B} (The Lost Caverns of Ixalan, #4470):
//
//	"As an additional cost to cast this spell, sacrifice an artifact or
//	 creature.
//	 Draw two cards and create a Map token."
//
// Deadly Dispute with a Map in place of the Treasure. The sacrifice is an
// additional cost (CR 601.2f), paid as the spell is cast (CR 601.2h), so
// it can't be cast with nothing to sacrifice and the sacrificed permanent
// stays gone if the spell is countered. The Map is the catalog's token
// (CR 111.10s), with its own sorcery-speed explore ability; it waits for
// any prompt the draw asks (CR 608.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "3e5402bf-ce7b-4c23-9333-0acdcf349caf",
		Name:           "Fanatical Offering",
		Purpose:        game.Purpose{Draws: 2, Tokens: 1, Answers: game.AnswerSacOutlet},
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeCost("an artifact or creature", Or(Artifact(), Creature())),
		OnResolve: ResolveSteps(
			DrawCards{N: 2},
			CreateToken{Template: MapToken(), N: 1},
		),
	})
}
