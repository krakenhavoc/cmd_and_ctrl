package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Venser, Shaper Savant — Legendary Creature — Human Wizard {2}{U}{U},
// 2/2 (EDHREC rank 1967):
//
//	"Flash
//	 When Venser enters, return target spell or permanent to its
//	 owner's hand."
//
// The flash Man-o'-War that also answers a spell. The permanent
// half is whole: a targeted ETB, any permanent, back to its owner's
// hand — a token bounced this way ceases to exist, and a commander
// goes where its owner chooses.
//
// DECLARED SIMPLIFICATION: the SPELL half is not implemented — the
// target clause offers permanents only. "Return target spell to its
// owner's hand" is a counter with a destination other than the
// graveyard (Remand's shape), and under the effect lock the engine
// exposes only CounterTargetForEffect, which routes to the
// graveyard; the destination-taking counter (counterSpellLocked with
// a ZoneRef) is unexported, and a plain hand move off the stack
// would leave the item's StackMeta behind. Until that seam opens,
// Venser cannot be pointed at a spell — weaker than printed, never
// stronger.
func init() {
	Register(Spec{
		OracleID:        "0f41cefc-d6ff-4db7-ba35-502b7e081de1",
		Name:            "Venser, Shaper Savant",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"Venser can only return a permanent to its owner's hand — returning a spell on the stack isn't implemented."},
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: b06SelfETB,
			Targets:   TargetPermanent("target permanent"),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Venser, Shaper Savant — return target permanent to its owner's hand",
					func(g *game.Game, item *game.StackItem) error {
						if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
							return nil
						}
						return BounceToHand{Target: item.Targets[0].ID}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
