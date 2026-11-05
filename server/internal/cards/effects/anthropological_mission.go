package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Anthropological Mission — Enchantment {2}{U} (Star Trek):
//
//	Whenever you cast a noncreature spell, put a quest counter on this
//	enchantment. Draw a card. Then discard a card unless this enchantment
//	has three or more quest counters on it.
//
// The count is read after the counter is placed, so the third trigger's
// own counter already spares its discard. An enchantment that left before
// the trigger resolved gets no counter and is read by its last-known
// counters (CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "535c0280-fdb5-4ff8-8608-c2f98b8a7ffc",
		Name:         "Anthropological Mission",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Noncreature(),
				"Anthropological Mission — put a quest counter on it, draw a card, then discard a card unless it has three or more quest counters",
				anthropologicalMissionResolve),
		},
	})
}

func anthropologicalMissionResolve(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (AddCounter{Target: item.SourceCardID, Kind: "quest", N: 1}).Apply(ctx); err != nil {
		return err
	}
	if trekB5CountersOnThis(g, item.SourceCardID, "quest") >= 3 {
		return DrawCards{N: 1}.Apply(ctx)
	}
	return lootOne(g, item, 1)
}
