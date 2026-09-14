package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Curiosity — Enchantment — Aura for {U} (EDHREC rank 585):
//
//	"Enchant creature
//	 Whenever enchanted creature deals damage to an opponent, you may
//	 draw a card."
//
// One blue mana for a card every turn the creature connects, and the
// half of the Aura catalog that is card advantage rather than stats.
//
// THE CONDITION IS WIDER THAN THE SWORDS', in three ways the card
// prints and attachedCreatureDealtDamageToOpponent implements:
//
//   - ANY damage, not just combat damage. A pinger wearing Curiosity
//     draws on every activation.
//   - An OPPONENT, not any player. A goaded creature forced to attack
//     its own controller draws nothing.
//   - "Opponent" is measured against the AURA's controller (CR 109.5),
//     because that is who "you may draw a card" refers to.
//
// "YOU MAY" IS A REAL PROMPT. TriggeredAbility.OptionalPrompt queues
// the yes/no to the Aura's controller and drops the trigger on a no,
// so the one case where you would decline — an empty library — is
// answerable rather than lethal. Nothing here is assumed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "223fa044-d387-4884-bf4e-75f1b61c6a46",
		Name:         "Curiosity",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attachedCreatureDealtDamageToOpponent(ev, source, g)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Curiosity — draw a card",
					func(g *game.Game, item *game.StackItem) error {
						return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
					})
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{
				Question: "Curiosity — draw a card?",
			},
		}},
	})
}
