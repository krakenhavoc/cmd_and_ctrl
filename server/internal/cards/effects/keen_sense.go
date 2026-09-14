package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Keen Sense — Enchantment — Aura for {G} (EDHREC rank 2857):
//
//	"Enchant creature
//	 Whenever enchanted creature deals damage to an opponent, you may
//	 draw a card."
//
// Curiosity in green, word for word, and in the batch for the same
// reason two Signets are: a deck that wants the effect wants all the
// copies of it its colours allow, and a green Voltron deck cannot
// play the blue one.
//
// Everything Curiosity's file records applies here — any damage, an
// opponent only, "you" being the Aura's controller, and the "may"
// being a real prompt rather than an assumed yes.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8d47b78b-11cc-4351-91d6-c891eb63dd98",
		Name:         "Keen Sense",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attachedCreatureDealtDamageToOpponent(ev, source, g)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Keen Sense — draw a card",
					func(g *game.Game, item *game.StackItem) error {
						return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
					})
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{
				Question: "Keen Sense — draw a card?",
			},
		}},
	})
}
