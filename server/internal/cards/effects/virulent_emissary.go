package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Virulent Emissary — Creature — Elf Assassin {G}, 1/1 (EDHREC rank
// 3343):
//
//	"Deathtouch
//	 Whenever another creature you control enters, you gain 1 life."
//
// A one-drop Soul Warden with deathtouch. Corpse Knight's condition
// (b13AnotherCreatureYouControlEntered — post-layer type, so an
// animated land counts, and a token's entry counts) and one life.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3d4bec90-7bbc-4385-a2b8-303c7a8d0a0f",
		Name:            "Virulent Emissary",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b13AnotherCreatureYouControlEntered(ev, source, g)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Virulent Emissary — you gain 1 life",
					func(g *game.Game, item *game.StackItem) error {
						return GainLife{Player: item.Controller, Amount: 1}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
