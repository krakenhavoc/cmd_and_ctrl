package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fireglass Mentor — Creature — Lizard Warlock {B}{R}, 2/1 (Bloomburrow,
// #4467):
//
//	"At the beginning of your second main phase, if an opponent lost life
//	 this turn, exile the top two cards of your library. Choose one of
//	 them. Until end of turn, you may play that card."
//
// The trigger is Michelangelo's second-main-phase shape with an
// intervening if (CR 603.4): an opponent's life lost this turn, asked as
// the phase begins (a gain-and-lose net still counts, the ruling) and again
// as the ability resolves (avB34CondAnOpponentLostLife). The exile is the plain
// top-of-library exile (ExileTopKeyedForEffect), whose registered tail
// then asks the controller to choose one of the cards still in exile; that
// card alone gets a play permission until end of turn (a land is played,
// spending the land drop), and the other stays exiled with none. A library
// of one card exiles it and chooses it; an empty one does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b173245e-2009-48ff-ac42-5f28d12aec79",
		Name:         "Fireglass Mentor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			InterveningIf(avB34CondAnOpponentLostLife{},
				On(game.EventStepBegan, AllOf(YourSecondMainPhaseBegan, fireglassOpponentLostLife),
					"Fireglass Mentor — exile the top two cards of your library; choose one of them; until end of turn you may play that card",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						return g.ExileTopKeyedForEffect(ctx.Controller(), ctx.Controller(), 2, ctx.ItemTail(fireglassChoose))
					})),
		},
	})
}

// fireglassOpponentLostLife is the trigger-time half of the intervening if.
func fireglassOpponentLostLife(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return opponentLifeLostThisTurn(g, source.Controller) > 0
}

// fireglassChoose is "choose one of them": a pick among the cards that
// landed in exile. An on-disk key.
var fireglassChoose = CardTail("fireglass-mentor/choose-exiled", func(ctx *Context, o game.TailOutcome) error {
	g, you := ctx.Game, ctx.Controller()
	var candidates []uuid.UUID
	for _, id := range o.IDs {
		if z := g.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneExile {
			candidates = append(candidates, id)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    you,
		FromPlayer: you,
		Source:     ctx.Source(),
		Question:   "Fireglass Mentor — choose one of the exiled cards; until end of turn you may play it",
		Cards:      candidates,
		Min:        1,
		Max:        1,
		LiveZone:   game.ZoneExile,
		ThenKey:    fireglassChosen,
		Asked:      game.NewAskedItem(ctx.Item),
	})
	return nil
})

// fireglassChosen grants the play permission over the chosen card. An
// on-disk key.
var fireglassChosen = askedPick("choose-cards/fireglass-mentor-play", func(ctx *Context, r game.CardsChosen) error {
	for _, id := range r.Picked {
		ctx.Game.GrantCastPermissionOverCardForEffect(id, game.CastPermission{
			Player: ctx.Controller(),
			Label:  "Fireglass Mentor — you may play that card until end of turn",
		})
	}
	return nil
})
