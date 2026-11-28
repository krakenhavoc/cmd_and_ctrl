package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nature's Wrath — Enchantment {4}{G}{G} (Alliances, S82 OLD batch 820):
//
//	"At the beginning of your upkeep, sacrifice this enchantment unless you
//	 pay {G}.
//	 Whenever a player puts an Island or blue permanent onto the battlefield,
//	 that player sacrifices an Island or blue permanent of their choice.
//	 Whenever a player puts a Swamp or black permanent onto the battlefield,
//	 that player sacrifices a Swamp or black permanent of their choice."
//
// The player who puts it there is its controller as it enters; the
// sacrifice is their own choice among every such permanent they control, the
// new one included, so a lone Island is sacrificed at once.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5a647f28-6e26-4262-8593-d582e0a691f7",
		Name:         "Nature's Wrath",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Nature's Wrath — sacrifice this enchantment unless you pay {G}",
				func(g *game.Game, item *game.StackItem) error {
					return effects.UpkeepPayUnless{
						Chooser:  item.Controller,
						Cost:     "{G}",
						Question: "Nature's Wrath — pay {G} or sacrifice it?",
						Else:     effects.UpkeepSacrificeSource,
					}.Apply(effects.NewContext(g, item))
				}),
			effects.OldB820PermanentEntersMatching("Nature's Wrath — that player sacrifices an Island or blue permanent of their choice",
				"an Island or blue permanent", effects.OldB820IslandOrBlue),
			effects.OldB820PermanentEntersMatching("Nature's Wrath — that player sacrifices a Swamp or black permanent of their choice",
				"a Swamp or black permanent", effects.OldB820SwampOrBlack),
		},
	})
}
