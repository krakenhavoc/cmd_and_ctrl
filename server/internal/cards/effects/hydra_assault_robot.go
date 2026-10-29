package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// HYDRA Assault Robot — {1}{R} Artifact Creature — Robot Villain 1/3 (Marvel
// Super Heroes, #4455):
//
//	"Whenever another Villain and/or artifact you control enters, this
//	 creature deals 1 damage to target opponent."
//
// "Villain and/or artifact" is one trigger event: a permanent that is both
// (an artifact Villain) triggers it once, not twice. The opponent is chosen
// as the trigger goes on the stack and the damage is dealt by the Robot.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a2558476-5506-4a09-93fc-b8ab33c98985",
		Name:         "HYDRA Assault Robot",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					c, ok := enteredUnderYourControl(ev, source, g, true)
					return ok && (c.IsArtifact() || c.HasSubtype(source.TextSubtype("Villain")))
				}, "HYDRA Assault Robot — 1 damage to target opponent", b36DesertDualDamage),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}
