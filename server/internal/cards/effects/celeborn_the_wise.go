package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Celeborn the Wise — Legendary Creature — Elf Noble {3}{G}, 3/3:
//
//	"Whenever you attack with one or more Elves, scry 1.
//	 Whenever you scry, Celeborn gets +1/+1 until end of turn for each
//	 card looked at while scrying this way."
//
// Arboreal Alliance's "one or more Elves" attack trigger (one event per
// attacker, collapsed by OncePerBatch) and Galadriel of Lothlorien's
// "whenever you scry" watch. The pump counts the cards the scry looked
// at, which is the size the action actually had (Event.LookedAt: a
// library shorter than the count clamps it, a Crystal Ball adds to it),
// not the cards it moved; a scry that looked at nothing gives nothing.
// Celeborn's own scry 1 triggers the second ability, so attacking with an
// Elf makes it 4/4 for the turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "64fa0416-16b7-4128-ac6c-f30373074a1e",
		Name:         "Celeborn the Wise",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, hobB6AttackWithAnElf, "Celeborn the Wise — scry 1", Do(Scry{N: 1}))),
			On(game.EventScry, ByYou, "Celeborn the Wise — +1/+1 until end of turn for each card looked at",
				func(g *game.Game, item *game.StackItem) error {
					n := item.Trigger.Event.LookedAt
					if n <= 0 {
						return nil
					}
					return BoostUntilEOT{Target: item.SourceCardID, Power: n, Toughness: n, Label: "Celeborn the Wise — +1/+1 for each card looked at"}.Apply(NewContext(g, item))
				}),
		},
	})
}
