package effects

// Spirit of the Labyrinth — Enchantment Creature — Spirit {1}{W}, 3/1:
//
//	"Each player can't draw more than one card each turn."
//
// #3927. Narset's static reaching every player, its controller included
// (Spec.DrawLimits, CR 121.2b): the active player's draw-step draw uses
// their one draw, and a draw made before the Spirit entered still counts
// (the 2014-02-01 rulings). A draw it stops can't be replaced.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1463795b-ec0c-44d6-ae1a-55f78d9843ec",
		Name:         "Spirit of the Labyrinth",
		Completeness: CompletenessFull,
		DrawLimits:   EachPlayerCantDrawMoreThanOneCardEachTurn(),
	})
}
