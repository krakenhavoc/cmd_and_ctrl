package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gemcutter Buccaneer — 1/3 Creature — Orc Pirate Artificer for
// {3}{R}:
//
//	"Whenever this creature or another Pirate you control enters,
//	 create a tapped Treasure token.
//	 Treasures you control are Equipment in addition to their other
//	 types and have 'Equipped creature gets +2/+0,' equip Pirate {1},
//	 and equip {3}."
//
// The first half ships: a Treasure for every Pirate, including its
// own arrival, which is why the trigger doesn't say "another".
//
// **The second half is not modelled.** Making Treasures into
// Equipment needs the attachment layer (S24) plus a Layer 4 type-add
// and a granted activated ability — three pieces the engine doesn't
// have. That is half the card, and a real one: in paper this turns
// a pile of Treasures into a pile of +2/+0. Declared here rather
// than silently dropped.
func init() {
	Register(Spec{
		OracleID:     "68e45c07-96c5-4f87-a816-d9fa4f119740",
		Name:         "Gemcutter Buccaneer",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The second ability is missing — your Treasures don't become Equipment granting +2/+0, so there's nothing to equip."},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && (c.InstanceID == source.InstanceID || isPirate(c))
			}, "Gemcutter Buccaneer — create a tapped Treasure", Do(CreateToken{
				Template: tappedTreasureToken(),
				N:        1,
			})),
		},
	})
}
