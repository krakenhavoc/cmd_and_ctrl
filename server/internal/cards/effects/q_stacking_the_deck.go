package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Q, Stacking the Deck — {2}{R}{R} Legendary Creature — Q 4/4 (Star Trek,
// #4973):
//
//	"When Q enters, shuffle your library, then exile cards from the top of
//	 your library until you exile a nonland card. That nonland card gains
//	 'Exile a card from your hand: Create a copy of this card. You may cast
//	 the copy. Activate only during your turn and only if this card is
//	 exiled.'"
//
// The trigger is the shared exile-until-nonland walk (MillToZone with
// UntilNonland, as Kang Prime's is); its continuation gives the nonland
// card, if it is still in exile, the bundle below through
// game.GrantAbilityToExiledCardForEffect (ADR 0219). The grant lives on
// the exiled card itself and ends when that card leaves exile (CR 400.7);
// lands exiled on the way get nothing.
//
// The granted ability is an exile-zone bundle: "only if this card is
// exiled" is its Zones, "only during your turn" its Condition, and the
// cost the ordinary exile-from-hand clause. The effect is CopyAndCast at
// the printed mana cost: the copy is made in exile and may be cast during
// the resolution (CR 707.12); one not cast ceases to exist (CR 704.5e),
// and a permanent copy resolves as a token (CR 707.10b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6866f0a4-5fe6-40d9-bbfc-6f116f4e415d",
		Name:         "Q, Stacking the Deck",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:  qStackingTheDeckGrant,
			Zone: game.ZoneExile,
			Activated: []ActivatedAbility{{
				Label:     "Exile a card from your hand: Create a copy of this card. You may cast the copy.",
				Cost:      game.AbilityCost{ExileCards: ExileACardFromHand()},
				Zones:     []game.ZoneKind{game.ZoneExile},
				Condition: DuringYourTurn(),
				// What it answers is whatever the exiled card does, so the
				// row declares the union of every answer a card can give
				// (ADR 0142's ruling for a conditional row, as Isochron
				// Scepter's).
				Purpose: game.Purpose{Answers: game.AnswerProtect | game.AnswerSurvive | game.AnswerPump |
					game.AnswerPrevent | game.AnswerRemove | game.AnswerSacOutlet | game.AnswerRestrict |
					game.AnswerGraveyard | game.AnswerCombatGrant | game.AnswerAnimate | game.AnswerMakesBlocker},
				Effect: qStackingTheDeckCopy,
			}},
			Text: "Exile a card from your hand: Create a copy of this card. You may cast the copy. Activate only during your turn and only if this card is exiled.",
		}},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Q, Stacking the Deck — shuffle, exile until a nonland card, it gains a copy ability",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (ShuffleLibrary{}).Apply(ctx); err != nil {
						return err
					}
					return MillToZone{
						Player:   ctx.Controller(),
						To:       game.ZoneExile,
						Until:    UntilNonland,
						ThenTail: ctx.Tail(qStackingTheDeckGained),
					}.Apply(ctx)
				}),
		},
	})
}

// qStackingTheDeckGrant is the bundle the nonland card gains. An on-disk
// identity: an exiled card in a snapshot names it.
const qStackingTheDeckGrant = "q-stacking-the-deck/copy-from-exile"

// qStackingTheDeckGained is the clause after the walk: the nonland card
// that ended it gains the bundle while it is still in exile.
var qStackingTheDeckGained = CardTail("q-stacking-the-deck/gains-copy-ability", func(ctx *Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || c.IsLand() {
			continue
		}
		ctx.Game.GrantAbilityToExiledCardForEffect(id, ctx.Source(), ctx.Controller(), qStackingTheDeckGrant)
	}
	return nil
})

// qStackingTheDeckCopy is the granted ability's effect: "Create a copy of
// this card. You may cast the copy." — at its printed cost.
func qStackingTheDeckCopy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return CopyAndCast{Cards: []uuid.UUID{ctx.Source()}}.Apply(ctx)
}
