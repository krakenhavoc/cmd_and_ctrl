package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Impulsive Pilferer — 1/1 Creature — Goblin Pirate for {R}:
//
//	"When this creature dies, create a Treasure token."
//
// A one-drop that turns into mana when it trades or gets sacrificed
// — the ramp half of a deck that wants artifacts entering.
//
// Encore {3}{R} is the shared Encore constructor (encore.go): from the
// graveyard, exile this card, and each opponent gets a hasty token copy
// that must attack that opponent and is sacrificed at the next end step.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7d9fc9e7-d80b-49c3-871c-ed25b3059ae8",
		Name:         "Impulsive Pilferer",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Encore("{3}{R}")},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Impulsive Pilferer — create a Treasure", Do(CreateToken{
				Template: TreasureToken(),
				N:        1,
			})),
		},
	})
}
