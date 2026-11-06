package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mechan Assembler — Artifact Creature — Robot Artificer {4}{U}, 4/4
// (Edge of Eternities, #4461):
//
//	"Whenever another artifact you control enters, create a 2/2 colorless
//	 Robot artifact creature token. This ability triggers only once each
//	 turn."
//
// Machinesmith Automaton's watch on another artifact entering under your
// control (artifactEnteredUnderYourControl, the Assembler itself left
// out), with Exemplar of Light's once-a-turn gate: the per-turn trigger
// record (b11TriggeredThisTurn) keeps every later artifact from
// triggering it once the ability has gone on the stack this turn, even
// if that first trigger was countered. The Robot it makes is an artifact
// too, but it enters after the ability has triggered, so it can't
// trigger it again that turn. The token is Biomechan Engineer's Robot.
//
// No simplification.
func init() {
	const label = "Mechan Assembler — create a 2/2 colorless Robot artifact creature token"
	Register(Spec{
		OracleID:     "2c2ef1ed-2f93-48c5-bf86-e46318599aeb",
		Name:         "Mechan Assembler",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return artifactEnteredUnderYourControl(ev, source, g) && !b11TriggeredThisTurn(g, source.InstanceID, label)
			}, label, Do(CreateToken{Template: TokenCard("2/2 colorless Robot artifact"), N: 1})),
		},
	})
}
