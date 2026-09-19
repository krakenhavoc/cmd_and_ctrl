package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ingenious Artillerist — 3/1 Creature — Human Artificer for {2}{R}:
//
//	"Whenever one or more artifacts you control enter, this creature
//	deals that much damage to each opponent."
//
// Functionally the Fireweaver at a different rate — except this one
// prints the batched "one or more ... enter" wording (CR 603.3) and
// Fireweaver doesn't. The engine has no way to read "how many
// artifacts entered in this batch" (see
// artifactEnteredUnderYourControl); it fires one trigger per
// artifact instead of one trigger for the whole batch, dealing 1
// damage each time. The running total comes out the same as printed
// for an ordinary game — N artifacts still deal N total damage.
//
// It stops being the same the moment something scales PER INSTANCE
// of damage rather than per point: Torbran, Thane of Red Fell adds
// +2 to every red damage event a source you control causes. Printed,
// two Treasures entering at once is one 2-damage event plus Torbran
// once (4 total); here it is two separate 1-damage events, each
// getting Torbran's +2 (6 total). That is stronger than printed —
// the #259 direction — so this ships as a caveat rather than Full,
// audited under #1112.
func init() {
	Register(Spec{
		OracleID:     "752c7723-90f8-4e3a-8266-f251ee0dadd8",
		Name:         "Ingenious Artillerist",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"When two or more of your artifacts enter at the same time, this deals the damage as that many separate hits instead of one combined hit, so a per-hit damage booster like Torbran, Thane of Red Fell adds its bonus more than once.",
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return artifactEnteredUnderYourControl(ev, source, g)
			}, "Ingenious Artillerist — damage to each opponent", func(g *game.Game, item *game.StackItem) error {
				return damageToEachOpponent(g, item, 1)
			}),
		},
	})
}
