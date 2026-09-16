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
// "Ward—Pay 2 life" is effects/ward.go's life cost: the spell's
// controller gets a pay-2-life-or-it-is-countered prompt, and a
// controller below 2 life cannot pay (CR 119.4), so the spell is
// countered outright. It shipped without the ward until the life cost
// existed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     invasionOfKarsusOracleID + "#1",
		Name:         "Refraction Elemental",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Ward(WardLife(2), "Refraction Elemental — ward, pay 2 life"),
			On(game.EventCast, ByYou, "Refraction Elemental — 2 damage to each opponent", func(g *game.Game, item *game.StackItem) error {
				return damageToEachOpponent(g, item, 2)
			}),
		},
	})
}
