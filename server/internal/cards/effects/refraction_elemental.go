package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Refraction Elemental — Creature — Elemental 4/4, the BACK FACE of
// Invasion of Karsus (oracle 2d125ef5…, face 1):
//
//	"Ward—Pay 2 life.
//	 Whenever you cast a spell, this creature deals 2 damage to each
//	 opponent."
//
// Registered under "<oracle_id>#1" — see deluge_of_the_dead.go for
// why the composite key exists and mdfc_lands.go for the sixty cards
// that opened the keyspace.
//
// "Whenever you CAST a spell" is a cast trigger on an ordinary
// permanent, so it watches EventCast with no type filter at all: any
// spell, including a land-drop's absence (playing a land is not
// casting, CR 305.1, so it correctly does not fire) and including
// the Elemental's own controller casting the SECOND half of a
// battle they later defeat. There is no "any spell you cast" helper
// in batchNN_helpers.go and writing one for a single card would hide
// two lines behind a name — the filter is exactly "the caster is my
// controller".
//
// SANDBOX SIMPLIFICATION, weaker than printed: Ward—Pay 2 life is
// not implemented, so the Elemental can be targeted for free.
// effects/ward.go ships MANA wards only — PendingChoicePayUnless
// parses its cost with game.ParseCost, and a life payment is not a
// mana cost. Sedgemoor Witch and Vein Ripper are behind the same
// seam; the file names WardCost as where a life-or-sacrifice ward
// lands. Omitting the ward is the weaker direction (#259): the
// creature is easier to answer than printed, never harder.
func init() {
	Register(Spec{
		OracleID:     invasionOfKarsusOracleID + "#1",
		Name:         "Refraction Elemental",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Ward—Pay 2 life isn't implemented, so opponents can target this creature without paying anything."},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, ByYou, "Refraction Elemental — 2 damage to each opponent", func(g *game.Game, item *game.StackItem) error {
				return damageToEachOpponent(g, item, 2)
			}),
		},
	})
}
