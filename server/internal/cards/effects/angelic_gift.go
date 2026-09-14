package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Angelic Gift — Enchantment — Aura for {1}{W} (EDHREC rank 3628):
//
//	"Enchant creature
//	 When this Aura enters, draw a card.
//	 Enchanted creature has flying."
//
// The cantrip Aura, and the reason it is worth a slot when Rancor is
// strictly bigger: the card replaces itself the moment it resolves,
// so the two-for-one that makes Auras bad never happens even if the
// creature dies in response to the next removal spell.
//
// Flying is the evasion that makes the rest of a Voltron board work —
// it is what gets a Loxodon Warhammer's trample damage to a player at
// all — and it is honoured by the block-legality check.
//
// THE ETB IS A TRIGGER, NOT AN OnETB HOOK, because that is what the
// card prints: "when this Aura enters" uses the stack, can be
// responded to, and resolves separately from the Aura itself. It
// fires on the Aura's own entry (b06SelfETB), which is the same
// self-ETB predicate every other catalog enters-trigger uses.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e5e04968-d9b7-4bd5-b826-be9502360cd3",
		Name:         "Angelic Gift",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static: []game.StaticAbility{
			GrantToAttached("flying"),
		},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: b06SelfETB,
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Angelic Gift — draw a card",
					func(g *game.Game, item *game.StackItem) error {
						return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
