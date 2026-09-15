package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ingenious Artillerist — 3/1 Creature — Human Artificer for {2}{R}:
//
//	"Whenever one or more artifacts you control enter, this creature
//	deals that much damage to each opponent."
//
// Functionally the Fireweaver at a different rate. The "that much"
// clause counts the artifacts in one batch; the engine fires one
// trigger per artifact, so the total across a batch is the same and
// only the stack shape differs. See artifactEnteredUnderYourControl.
func init() {
	Register(Spec{
		OracleID: "752c7723-90f8-4e3a-8266-f251ee0dadd8",
		Name:     "Ingenious Artillerist",
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return artifactEnteredUnderYourControl(ev, source, g)
			}, "Ingenious Artillerist — damage to each opponent", func(g *game.Game, item *game.StackItem) error {
				return damageToEachOpponent(g, item, 1)
			}),
		},
	})
}
