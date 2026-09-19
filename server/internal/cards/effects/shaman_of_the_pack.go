package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shaman of the Pack — Creature — Elf Shaman {1}{B}{G}, 3/2 (EDHREC
// rank 4515):
//
//	"When this creature enters, target opponent loses life equal to
//	 the number of Elves you control."
//
// The Golgari Elf deck's reach. An Elfball board that cannot get
// through a wall of blockers does not need to: the Shaman turns the
// board into a drain, and a deck that can recur or blink it turns
// one opponent's life total into a countdown.
//
// The count includes the Shaman, because the Shaman is an Elf and the
// card says "the number of Elves you control" rather than "other
// Elves" — so it is never worse than 1. It is taken when the trigger
// RESOLVES, not when it goes on the stack, so an Elf killed in
// response really does shrink the drain (CR 608.2). Post-layer
// subtypes, so a changeling counts.
//
// Life LOSS, not damage: no prevention shield stops it and no
// lifelink triggers off it (CR 119.3).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "39c7d2cc-4fd1-4f1a-be1a-a6ae56d2356d",
		Name:         "Shaman of the Pack",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: b06SelfETB,
			Targets:   TargetPlayer("target opponent", Opponent()),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Shaman of the Pack — target opponent loses life for each Elf you control",
					func(g *game.Game, item *game.StackItem) error {
						return b43TargetPlayerLoses(g, item,
							b43CreaturesOfSubtypeControlled(g, item.Controller, "Elf"))
					})
			},
		}},
	})
}
