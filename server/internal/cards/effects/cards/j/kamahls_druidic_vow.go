package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kamahl's Druidic Vow — Legendary Sorcery {X}{G}{G} (Dominaria, S82 OLD,
// #5530):
//
//	"(You may cast a legendary sorcery only if you control a legendary creature
//	 or planeswalker.)
//	 Look at the top X cards of your library. You may put any number of land
//	 and/or legendary permanent cards with mana value X or less from among them
//	 onto the battlefield. Put the rest into your graveyard."
//
// The cast gate is the legendary-sorcery one (CR 205.4e). The look is private
// to the caster; the pick is any number (including none) of the lands and the
// legendary permanent cards of mana value X or less, and everything not put
// onto the battlefield goes to the graveyard as a put, not a mill (Shigeki's
// PutRestIntoGraveyardTail). X = 0 looks at nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:           "7d012731-f921-47bf-a2da-ce494650cd4f",
		Name:               "Kamahl's Druidic Vow",
		Completeness:       effects.CompletenessFull,
		XMatters:           true,
		CastCondition:      effects.LegendarySorcery(),
		CastConditionLabel: effects.LegendarySorceryLabel,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			x := ctx.X()
			if x <= 0 {
				return nil
			}
			player := ctx.Controller()
			return effects.PutFromLibraryOntoBattlefield{
				Player: player,
				Cards:  ctx.Game.LookAtTopOfLibraryForEffect(player, x),
				Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
					return (c.IsLand() || c.IsLegendary()) && c.ManaValue() <= x
				},
				Max:      0,
				Optional: true,
				Label:    "Kamahl's Druidic Vow — put any number of land and/or legendary permanent cards with mana value X or less onto the battlefield",
				ThenTail: effects.PutRestIntoGraveyardTail,
			}.Apply(ctx)
		},
	})
}
