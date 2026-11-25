package cardsso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Specter's Shriek — Sorcery {B} (Throne of Eldraine, S82 old sets):
//
//	"Target opponent reveals their hand. You may choose a nonland card from
//	 it. If you do, that player exiles that card. If a nonblack card is
//	 exiled this way, exile a card from your hand."
//
// Revealing Eye's optional revealed-hand pick (ADR 0116), exiling the card
// (the exile is not a discard) with the whole table seeing the hand. "If you
// do" is the keyed continuation, run once the card is in exile: it reads the
// exiled card's colours (a colourless card is nonblack) and, for a nonblack
// one, asks the caster to exile a card from their own hand. An empty hand
// ends it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8c9164b5-10ed-4cad-af33-7d4b5d3d6bf2",
		Name:         "Specter's Shriek",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ChooseFromRevealedHand{
				Player:   effects.TargetedPlayer(ctx),
				Filter:   effects.Nonland(),
				Label:    "nonland card",
				Exile:    true,
				Optional: true,
				Then:     specterShriekIfYouDo,
			}.Apply(ctx)
		},
	})
}

// specterShriekIfYouDo is "If a nonblack card is exiled this way, exile a
// card from your hand". An on-disk key: never renamed, never reused.
var specterShriekIfYouDo = effects.RevealedPickThen("revealed-pick/specters-shriek-exile-from-hand",
	func(ctx *effects.Context, pick game.RevealedPick) error {
		g := ctx.Game
		nonblack := false
		for _, chosen := range pick.Chosen {
			if !effects.OfColor("B")(g, pick.Chooser, chosen) {
				nonblack = true
			}
		}
		if !nonblack {
			return nil
		}
		me := g.PlayerByIDForEffect(pick.Chooser)
		if me == nil || me.Hand == nil || me.Hand.Size() == 0 {
			return nil
		}
		cards := make([]uuid.UUID, 0, me.Hand.Size())
		for _, c := range me.Hand.Cards {
			cards = append(cards, c.InstanceID)
		}
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:    pick.Chooser,
			FromPlayer: pick.Chooser,
			Source:     pick.Source,
			Question:   "Specter's Shriek — exile a card from your hand",
			Cards:      cards,
			Min:        1,
			Max:        1,
			ThenKey:    specterShriekExile,
			LiveZone:   game.ZoneHand,
		})
		return nil
	})

// specterShriekExile exiles the picked card from the caster's hand. An
// on-disk key: never renamed, never reused.
var specterShriekExile = game.RegisterChooseCardsThen("choose-cards/specters-shriek-exile-from-hand",
	func(g *game.Game, r game.CardsChosen) error {
		for _, id := range r.Picked {
			if err := g.ExileCardForEffect(id); err != nil {
				return err
			}
		}
		return nil
	})
