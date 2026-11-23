package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thopter Spy Network — Enchantment, {2}{U}{U} (EDHREC rank 936):
//
//	"At the beginning of your upkeep, if you control an artifact,
//	 create a 1/1 colorless Thopter artifact creature token with
//	 flying.
//	 Whenever one or more artifact creatures you control deal combat
//	 damage to a player, draw a card."
//
// The artifact deck's engine: a Thopter every upkeep as long as any
// artifact is around (the first Thopter keeps it going by itself),
// and a card whenever the fliers connect.
//
// The upkeep half is an intervening "if" (CR 603.4): "if you control
// an artifact" is checked when the trigger would go on the stack and
// again as it resolves, so an artifact removed in response stops the
// Thopter.
//
// The combat half is "one or more … to a player", so it is one draw
// per PLAYER the artifact creatures connect with (CR 603.2c, #784):
// the engine emits one damage event per creature, and a second one
// naming the same player is declined as a later event of the same
// batch, while a second PLAYER hit is its own occurrence and its own
// card (see docs/adding-cards.md). Without the batch half the card would
// ship STRONGER than printed (#259); without the player half,
// weaker.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "49be65fd-3755-410d-b0dc-2e5861ea2552",
		Name:         "Thopter Spy Network",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginUpkeep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && effects.B03ArtifactsControlled(g, source.Controller) > 0
			}, "Thopter Spy Network — create a 1/1 Thopter", func(g *game.Game, item *game.StackItem) error {
				if effects.B03ArtifactsControlled(g, item.Controller) == 0 {
					return nil
				}
				return effects.CreateToken{Controller: item.Controller, Template: effects.TokenCard("1/1 colorless Thopter artifact with flying"), N: 1}.Apply(effects.NewContext(g, item))
			}),
			effects.WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(effects.Artifact(),
				"Thopter Spy Network — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
