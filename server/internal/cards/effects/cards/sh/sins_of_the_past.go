package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sins of the Past — Sorcery {4}{B}{B} (Ravnica: City of Guilds):
//
//	"Until end of turn, you may cast target instant or sorcery card from your
//	 graveyard without paying its mana cost. If that spell would be put into
//	 your graveyard, exile it instead. Exile Sins of the Past."
//
// Nicol Bolas, God-Pharaoh's shape for a card in a graveyard: a cast-only
// permission over the targeted card, free ({0}, so X is 0 and additional costs
// are still paid), that lasts the whole turn (CR 611.2a) and leaves the card's
// own timing alone, so a sorcery taken this way is cast in a later main phase
// of an empty stack, not at instant speed. Its ExileInsteadOfGraveyard is the
// printed replacement (#3863): the spell is exiled whether it resolves, is
// countered or fizzles. The target is checked again as the spell resolves
// (CR 608.2b); a card that left the graveyard is granted nothing. Sins then
// exiles itself as its last instruction, and the engine does not also put it
// in the graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9059a940-ac28-4322-a11b-af1d107b2edd",
		Name:         "Sins of the Past",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetCardInGraveyard("target instant or sorcery card from your graveyard",
			effects.YouOwn(), effects.Or(effects.Instant(), effects.Sorcery())),
		Purpose: game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			g := ctx.Game
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				z := g.FindCardZoneForEffect(t.ID)
				c, ok := g.LookupCardForEffect(t.ID)
				if !ok || z == nil || z.Kind != game.ZoneGraveyard {
					break
				}
				g.GrantCastPermissionToCardsForEffect(game.CastPermission{
					Player:                  item.Controller,
					Zone:                    game.ZoneGraveyard,
					Duration:                g.UntilEndOfTurnDuration(),
					Cost:                    "{0}",
					AltCostKey:              game.AltCostKeyGrantedCost,
					CastOnly:                true,
					ExileInsteadOfGraveyard: true,
					Source:                  item.SourceCardID,
					SourceName:              "Sins of the Past",
					Label:                   "Sins of the Past — cast " + c.Name + " without paying its mana cost",
				}, []game.Card{c})
				break
			}
			return g.ExileCardForEffect(item.SourceCardID)
		},
	})
}
