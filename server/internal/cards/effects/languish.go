package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Languish — Sorcery {2}{B}{B}:
//
//	"All creatures get -4/-4 until end of turn."
//
// Toxic Deluge with the X fixed at four and no life to pay. Same
// mechanism and the same reasons to want it over a wrath: a -4/-4 is
// not destruction, so indestructible, regeneration and totem armor
// are all irrelevant, and the creatures that survive spend the rest
// of the turn as much smaller creatures.
//
// The four-toughness line is the point: Languish is the sweeper a
// deck plays when its own threats are 5/5s.
//
// CR 611.2c — the affected set is snapshotted at resolution, so a
// creature that arrives later this turn is a full-size creature.
func init() {
	Register(Spec{
		OracleID: "ef1a83f2-6707-41a2-b5ed-861c8e45ae07",
		Name:     "Languish",
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return BoostUntilEOT{
				Match:     Creature(),
				Power:     -4,
				Toughness: -4,
				Label:     "Languish — -4/-4",
			}.Apply(ctx)
		},
	})
}
