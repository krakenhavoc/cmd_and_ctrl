package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hornet Queen — Creature — Insect {4}{G}{G}{G}, 2/2 (EDHREC rank
// 1696):
//
//	"Flying, deathtouch
//	 When this creature enters, create four 1/1 green Insect creature
//	 tokens with flying and deathtouch."
//
// Seven mana for five deathtouch flyers — the green deck's wall
// against any attacker at the table. The tokens carry their two
// keywords on the template (a token has no oracle ID for the
// catalog to key on), which is the same surface the Queen's own
// printed keywords reach through the layer engine.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3b1f8108-6911-49e9-8f78-f950bb58cb6c",
		Name:            "Hornet Queen",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "deathtouch"},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: b06SelfETB,
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Hornet Queen — create four 1/1 Insects with flying and deathtouch",
					func(g *game.Game, item *game.StackItem) error {
						return CreateToken{Controller: item.Controller, Template: b15GreenInsectToken(), N: 4}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
