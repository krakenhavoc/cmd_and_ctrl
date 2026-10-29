package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	supperForSpidersOracle = "4d891515-39da-492e-ac19-1aa524245449"
	supperForSpidersFood   = "supper-for-spiders/food"
)

// Supper for Spiders — Instant {1}{B} (The Hobbit, #4454, #4506):
//
//	"Put onto the battlefield under your control all creature cards in
//	 your opponents' graveyards that were put there from the battlefield
//	 this turn. They are Food artifacts with '{2}, {T}, Sacrifice this
//	 artifact: You gain 3 life.' (They lose all other types and
//	 subtypes.)"
//
//   - The set is read once as the spell resolves, from every opponent's
//     graveyard, by each card's newest arrival there
//     (putIntoGraveyardFromBattlefieldThisTurn, Twilight Shepherd's and
//     Gerrard's Hourglass Pendant's reader): a milled or discarded card
//     is not on it, nor one that died last turn, nor one that left and
//     came back (CR 400.7). "Creature card" is the card as it is in the
//     graveyard. A token is not a card (CR 108.2).
//   - The cards enter together, under your control, as one simultaneous
//     entry (CR 603.6a).
//   - "They are Food artifacts with …" is ADR 0210's entry overwrite: each
//     card is a Food artifact from the moment it lands (CR 614.1c), so it
//     is never a creature on the battlefield and a "whenever a creature
//     enters" trigger does not see it, and the entry's replacements judge
//     it as a Food (CR 614.12). The types are SET (CR 205.1a): Artifact,
//     subtype Food, every creature type gone; supertypes stay (CR 205.4b).
//     The Food ability is a granted bundle (ADR 0093); the card keeps its
//     own abilities. The overwrite lasts for as long as that permanent is
//     on the battlefield; a copy of it copies the printed creature
//     (CR 707.2).
//
// No simplification.

func init() {
	Register(Spec{
		OracleID:     supperForSpidersOracle,
		Name:         "Supper for Spiders",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: supperForSpidersFood,
			Activated: []ActivatedAbility{{
				Label:   "{2}, {T}, Sacrifice this artifact: You gain 3 life.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    Plus(ManaCost("{2}"), TapCost(), SacrificeThis()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return GainLife{Player: item.Controller, Amount: 3}.Apply(NewContext(g, item))
				},
			}},
			Text: "{2}, {T}, Sacrifice this artifact: You gain 3 life.",
		}},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			var ids []uuid.UUID
			for _, opp := range ctx.Opponents() {
				ids = append(ids, graveyardCardIDs(ctx, opp, func(c game.Card) bool {
					return !c.IsToken() && c.IsCreature() && putIntoGraveyardFromBattlefieldThisTurn(ctx.Game, c.InstanceID)
				})...)
			}
			if len(ids) == 0 {
				return nil
			}
			return ReturnFromGraveyardTogether{
				Targets:    ids,
				Controller: item.Controller,
				EntersAs: &game.EntersAs{
					Types:    []string{"Artifact"},
					Subtypes: []string{"Food"},
					Grants:   []string{supperForSpidersFood},
					Label:    "Supper for Spiders — a Food artifact",
				},
			}.Apply(ctx)
		},
	})
}
