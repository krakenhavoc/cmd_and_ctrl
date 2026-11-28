package cardsgr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gustha's Scepter — Artifact {0} (Alliances, #5634):
//
//	"{T}: Exile a card from your hand face down. You may look at it for as long
//	 as it remains exiled.
//	 {T}: Return a card you own exiled with this artifact to your hand.
//	 When you lose control of this artifact, put all cards exiled with this
//	 artifact into their owner's graveyard."
//
// Three linked abilities over one pile (CR 607.2a), Duplicity's: the exile is
// face down and linked to this object, its controller as the one who may look
// (ExileHiddenLinkedForEffect); the return is a pick over the cards you own that
// this object exiled; and "when you lose control", whether another player takes
// it or it leaves the battlefield, puts the pile into its owners' graveyards.
//
// No simplification.
func init() {
	const name = "Gustha's Scepter"
	effects.Register(effects.Spec{
		OracleID:     "120e9013-b07b-49d9-8b86-f6c26315581d",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}: Exile a card from your hand face down. You may look at it for as long as it remains exiled.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.TapCost(),
				Effect:  oldB815ScepterExile,
			},
			{
				Label:   "{T}: Return a card you own exiled with this artifact to your hand.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.TapCost(),
				Effect:  oldB815ScepterReturn,
			},
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenYouLoseControlOfThis(name+" — put all cards exiled with this artifact into their owner's graveyard", effects.S82OldB761ExiledWithToGraveyards),
			effects.WhenThisLeaves(name+" — put all cards exiled with this artifact into their owner's graveyard", effects.S82OldB761ExiledWithToGraveyards),
		},
	})
}

// oldB815ScepterExilePick exiles the chosen card from the hand, face down and
// linked to the Scepter. An on-disk key.
var oldB815ScepterExilePick = game.RegisterChooseCardsThen("choose-cards/old-b815-gusthas-scepter-exile", func(g *game.Game, r game.CardsChosen) error {
	ref, linked := g.PermanentRefForEffect(r.Source)
	for _, id := range r.Picked {
		if !linked {
			if err := g.ExileCardForEffect(id); err != nil {
				return err
			}
			continue
		}
		if _, err := g.ExileHiddenLinkedForEffect(game.PermissionCardRef(ref), r.Chooser, id); err != nil {
			return err
		}
	}
	return nil
})

func oldB815ScepterExile(g *game.Game, item *game.StackItem) error {
	if !effects.SourceIsStillThisPermanent(g, item) {
		return nil
	}
	p := g.PlayerByIDForEffect(item.Controller)
	if p == nil || p.Hand == nil || p.Hand.Size() == 0 {
		return nil
	}
	ids := make([]uuid.UUID, 0, p.Hand.Size())
	for _, c := range p.Hand.Cards {
		ids = append(ids, c.InstanceID)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Gustha's Scepter — exile a card from your hand face down",
		Cards:    ids,
		Min:      1,
		Max:      1,
		LiveZone: game.ZoneHand,
		ThenKey:  oldB815ScepterExilePick,
	})
	return nil
}

// oldB815ScepterReturnPick returns the chosen card to its owner's hand. An
// on-disk key.
var oldB815ScepterReturnPick = game.RegisterChooseCardsThen("choose-cards/old-b815-gusthas-scepter-return", func(g *game.Game, r game.CardsChosen) error {
	for _, id := range r.Picked {
		if err := effects.NothingIfGone(g.BounceToHandForEffect(id)); err != nil {
			return err
		}
	}
	return nil
})

func oldB815ScepterReturn(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	var ids []uuid.UUID
	for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(ref)) {
		if c.Owner == item.Controller {
			ids = append(ids, c.InstanceID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Gustha's Scepter — return a card you own exiled with this artifact to your hand",
		Cards:    ids,
		Min:      1,
		Max:      1,
		LiveZone: game.ZoneExile,
		ThenKey:  oldB815ScepterReturnPick,
	})
	return nil
}
