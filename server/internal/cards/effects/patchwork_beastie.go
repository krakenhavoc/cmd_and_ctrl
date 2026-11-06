package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Patchwork Beastie — Artifact Creature — Beast {G}, 3/3 (Duskmourn:
// House of Horror, #4466):
//
//	"Delirium — This creature can't attack or block unless there are four
//	 or more card types among cards in your graveyard.
//	 At the beginning of your upkeep, you may mill a card. (You may put the
//	 top card of your library into your graveyard.)"
//
// Delirium is an ability word (CR 207.2c). The restriction is The
// Lion-Turtle's shape: a layer 6 static on itself, on while its
// controller's graveyard holds fewer than four card types
// (b16CardTypesInGraveyard, Demonic Counsel's delirium count), setting CantAttack and CantBlock, read live as
// attackers and blockers are declared (CR 508.1c, 509.1b). The upkeep
// mill is a "you may" asked as the trigger resolves (CR 603.5): the way
// the Beastie turns itself on.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6cf995e9-115a-4659-bace-9fbe75898b48",
		Name:         "Patchwork Beastie",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label: "Delirium — This creature can't attack or block unless there are four or more card types among cards in your graveyard.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return selfOnly(target, g, source) && b16CardTypesInGraveyard(g, source.Controller) < 4
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Restrictions |= game.CantAttack | game.CantBlock
			},
		}},
		Triggered: []game.TriggeredAbility{
			Optional(AtYourUpkeep("Patchwork Beastie — mill a card", Do(MillCards{N: 1})),
				"Patchwork Beastie — mill a card?"),
		},
	})
}
