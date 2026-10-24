package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tersa Lightshatter — Legendary Creature — Orc Wizard {2}{R}, 3/3:
//
//	"Haste
//	 When Tersa Lightshatter enters, discard up to two cards, then draw
//	 that many cards.
//	 Whenever Tersa Lightshatter attacks, if there are seven or more
//	 cards in your graveyard, exile a card at random from your
//	 graveyard. You may play that card this turn."
//
// The ETB is Obsessive Astronomer's run (b39MayDiscardThenDraw): the
// count drawn is the count actually discarded.
//
// The attack trigger has an intervening "if" (CR 603.4): it is checked
// when the creature attacks and again as the ability resolves, so a
// graveyard that has been exiled below seven in response fizzles it.
// The card is drawn from the engine's random stream, exiled, and given
// a play (not cast-only) permission for the rest of the turn, so a
// land can be played from exile as well as a spell cast. Timing still
// applies on top of the grant.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9918eb7b-054e-4b1c-afc7-1163ab642d50",
		Name:            "Tersa Lightshatter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Tersa Lightshatter — discard up to two cards, then draw that many",
				func(g *game.Game, item *game.StackItem) error {
					return b39MayDiscardThenDraw(2, false,
						"Tersa Lightshatter — discard up to two cards, then draw that many",
						drawAsManyAsDiscarded)(NewContext(g, item))
				}),
			On(game.EventAttack,
				AllOf(ThisAttacked, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return tersaSevenInGraveyard(g, source.Controller)
				}),
				"Tersa Lightshatter — exile a card at random from your graveyard; you may play it this turn",
				tersaExileRandom),
		},
	})
}

func tersaSevenInGraveyard(g *game.Game, controller uuid.UUID) bool {
	p := g.PlayerByIDForEffect(controller)
	return p != nil && p.Graveyard != nil && len(p.Graveyard.Cards) >= 7
}

func tersaExileRandom(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	controller := ctx.Controller()
	if !tersaSevenInGraveyard(g, controller) {
		return nil
	}
	pick := g.ChooseAtRandomForEffect(randomDraw(ctx), graveyardIDs(g, controller, nil), 1)
	if len(pick) == 0 {
		return nil
	}
	return g.ExileCardWithPermissionForEffect(pick[0], game.CastPermission{
		Player: controller,
		Label:  "Tersa Lightshatter — play it this turn",
	})
}
