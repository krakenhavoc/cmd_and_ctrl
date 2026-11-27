package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Back from the Brink — {4}{U}{U} Enchantment (Innistrad, S82 OLD b447, #5572,
// seam #6270):
//
//	"Exile a creature card from your graveyard and pay its mana cost: Create a
//	 token that's a copy of that card. Activate only as a sorcery."
//
// The cost is ExileFromGraveyardAndPayItsManaCost: the activator names the
// card as the ability is activated (CR 602.2b) and its mana cost is the mana
// they pay (CR 601.2h). The token copies the card as it sits in exile, whose
// copiable values are the printed ones (CR 707.2). X in the card's cost is 0
// and a card with no mana cost costs nothing (CR 202.1b, 107.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "704aaf92-590f-46d3-8b39-4ab84e1c766d",
		Name:         "Back from the Brink",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "Exile a creature card from your graveyard and pay its mana cost: Create a token that's a copy of that card. Activate only as a sorcery.",
			Cost:         effects.ExileFromGraveyardAndPayItsManaCost("a creature card", func(c game.Card) bool { return c.IsCreature() }),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, id := range ctx.Exiled() {
					return effects.CreateTokenCopy{Copy: id, N: 1}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
