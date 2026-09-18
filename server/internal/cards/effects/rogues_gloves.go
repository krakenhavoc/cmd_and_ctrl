package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rogue's Gloves — Artifact — Equipment {2} (EDHREC rank 4102):
//
//	"Whenever equipped creature deals combat damage to a player, you
//	 may draw a card.
//	 Equip {2}"
//
// The colourless Curiosity: two to cast, two to equip, and every
// connection is a card. It is in the batch because a deck with an
// unblockable creature and no colour to spare runs it, and because
// it is the Equipment half of a trigger shape the Auras already
// have — the same condition Curiosity reads, scoped by attachment
// instead of enchantment.
//
// Three differences from Curiosity, all printed:
//
//   - COMBAT damage only. A Gloved creature that pings with an
//     activated ability draws nothing.
//   - Any player, not only an opponent. A goaded creature forced to
//     swing at the Gloves' own controller still draws — the printed
//     text says "a player" and does not narrow it.
//   - The draw is the EQUIPMENT controller's (CR 109.5), so Gloves
//     you have stolen draw for you.
//
// "YOU MAY" IS A REAL PROMPT — TriggeredAbility.OptionalPrompt — so
// the one case where you decline, an empty library, is answerable
// rather than lethal.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "308d7868-49f9-47a3-a7b1-4b0332d610f1",
		Name:         "Rogue's Gloves",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attachedCreatureDealtCombatDamageToPlayer(ev, source, g)
			}, "Rogue's Gloves — draw a card", Do(DrawCards{N: 1})), "Rogue's Gloves — draw a card?"),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{2}"),
		},
	})
}
