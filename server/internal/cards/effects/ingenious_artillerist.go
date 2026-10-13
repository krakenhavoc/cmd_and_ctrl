package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ingenious Artillerist — 3/1 Creature — Human Artificer for {2}{R}:
//
//	"Whenever one or more artifacts you control enter, this creature
//	deals that much damage to each opponent."
//
// Functionally the Fireweaver at a different rate, and it prints the
// batched "one or more ... enter" wording (CR 603.2c) that Fireweaver
// doesn't. The engine emits one EventETB per permanent, so the trigger
// waits for the entry to finish (AtBatchEnd, ThatManyEvents), counts the
// artifacts that entered under your control in that batch, and deals the
// count to each opponent as ONE damage event. That is what makes a
// per-instance damage booster right: Torbran, Thane of Red Fell adds its
// +2 once to a 2-damage hit from two Treasures, not once per Treasure.
func init() {
	Register(Spec{
		OracleID:     "752c7723-90f8-4e3a-8266-f251ee0dadd8",
		Name:         "Ingenious Artillerist",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			withAtBatchEnd(On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return artifactEnteredUnderYourControl(ev, source, g)
			}, "Ingenious Artillerist — that much damage to each opponent", func(g *game.Game, item *game.StackItem) error {
				return damageToEachOpponent(g, item, settledEventAmount(item))
			}), ThatManyEvents()),
		},
	})
}
