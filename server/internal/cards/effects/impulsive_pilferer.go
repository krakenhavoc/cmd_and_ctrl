package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Impulsive Pilferer — 1/1 Creature — Goblin Pirate for {R}:
//
//	"When this creature dies, create a Treasure token."
//
// A one-drop that turns into mana when it trades or gets sacrificed
// — the ramp half of a deck that wants artifacts entering.
//
// Encore {3}{R} isn't modelled. It is not a cast: it is an activated
// ability that works from the graveyard (CR 702.141a), and
// activation from a non-battlefield zone is an open engine seam
// (docs/engine-seams.md).
func init() {
	Register(Spec{
		OracleID:     "7d9fc9e7-d80b-49c3-871c-ed25b3059ae8",
		Name:         "Impulsive Pilferer",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Encore isn't implemented — you can't pay {3}{R} to bring it back from your graveyard for token copies."},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Impulsive Pilferer — create a Treasure", Do(CreateToken{
				Template: TreasureToken(),
				N:        1,
			})),
		},
	})
}
