package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brazen Blademaster — Creature — Orc Pirate {2}{R}, 2/3 (The Lost
// Caverns of Ixalan, #4470):
//
//	"Whenever this creature attacks while you control two or more
//	 artifacts, it gets +2/+1 until end of turn."
//
// Moon-Boy, Dino Rider's shape: "while" is part of the trigger event, so
// the artifact count is made once, as the attack is declared, on the
// post-layer type line (an animated artifact still counts). The trigger
// is not re-checked on resolution, and the bonus lands if the
// Blademaster is still the same object then (CR 611.2c, ending at
// CR 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "43c92618-0aae-41f8-aa39-393e5d6b5ff7",
		Name:         "Brazen Blademaster",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attackDeclared(ev, source) && b03ArtifactsControlled(g, source.Controller) >= 2
			}, "Brazen Blademaster — gets +2/+1 until end of turn", thisGetsUntilEndOfTurn(2, 1, "Brazen Blademaster")),
		},
	})
}
