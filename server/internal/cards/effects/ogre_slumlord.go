package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ogre Slumlord — Creature — Ogre Rogue {3}{B}{B}, 3/3 (EDHREC rank
// 1963):
//
//	"Whenever another nontoken creature dies, you may create a 1/1
//	 black Rat creature token.
//	 Rats you control have deathtouch."
//
// The aristocrats deck's Rat factory. The trigger is Harvester of
// Souls' condition — any player's nontoken creature, its own death
// excluded — with the "may" as a real prompt; the Rats are plain
// 1/1s that get deathtouch from the Slumlord's grant, the
// TribalKeywordGrant shape, so a Rat that outlives the Slumlord is a
// 1/1 and nothing more, and a Rat from any other source (Marrow-Gnawer,
// a Pack Rat) gets it too, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0a5e3748-2e58-4e53-9653-8af4e21cf223",
		Name:         "Ogre Slumlord",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b17AnotherNontokenCreatureDied(ev, source, g)
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Ogre Slumlord — create a 1/1 black Rat?"},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Ogre Slumlord — create a Rat",
					func(g *game.Game, item *game.StackItem) error {
						return CreateToken{
							Controller: item.Controller,
							Template:   b18BlackRatToken(),
							N:          1,
						}.Apply(NewContext(g, item))
					})
			},
		}},
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Tribes: []string{"Rat"}, YoursOnly: true}, "deathtouch"),
		},
	})
}
