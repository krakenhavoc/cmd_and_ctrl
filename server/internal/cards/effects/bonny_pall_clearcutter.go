package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bonny Pall, Clearcutter — Legendary Creature — Giant Scout {3}{G}{U}{U},
// 6/5 (Outlaws of Thunder Junction, #4468):
//
//	"Reach
//	 When Bonny Pall enters, create Beau, a legendary blue Ox creature
//	 token with 'Beau's power and toughness are each equal to the number of
//	 lands you control.'
//	 Whenever you attack, draw a card, then you may put a land card from
//	 your hand or graveyard onto the battlefield."
//
// "Whenever you attack" is one trigger per declaration (OncePerBatch). The
// land pick opens only after the draw has settled (DrawThen, CR 608.2c), so
// the drawn card is a candidate, over the lands in the controller's hand and
// graveyard together; the chosen card enters from whichever zone holds it
// when the answer arrives, and one that has left both is skipped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "fcb811a8-c6e6-4bc4-b64b-8963a7183669",
		Name:            "Bonny Pall, Clearcutter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Purpose:         game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Bonny Pall, Clearcutter — create Beau, a legendary blue Ox", Do(CreateToken{Template: BeauToken(), N: 1})),
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller)
			}, "Bonny Pall, Clearcutter — draw a card, then you may put a land card from your hand or graveyard onto the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					return g.DrawNThenForEffect(item.Controller, 1,
						game.DrawThen{Ref: bonnyPallPutLand, Player: item.Controller, Source: item.SourceCardID})
				})),
		},
	})
}

// bonnyPallPutLand is the rest of the trigger once the draw has settled:
// the land pick. An on-disk key.
var bonnyPallPutLand = game.RegisterDrawThen("bonny-pall-clearcutter/put-land", func(g *game.Game, d game.DrawThen) error {
	p := g.PlayerByIDForEffect(d.Player)
	if p == nil {
		return nil
	}
	var ids []uuid.UUID
	for _, z := range []*game.Zone{p.Hand, p.Graveyard} {
		if z == nil {
			continue
		}
		for _, c := range z.Cards {
			if c.IsLand() && !c.IsToken() {
				ids = append(ids, c.InstanceID)
			}
		}
	}
	if len(ids) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  d.Player,
		Source:   d.Source,
		Question: "Bonny Pall, Clearcutter — you may put a land card from your hand or graveyard onto the battlefield",
		Cards:    ids,
		Min:      0,
		Max:      1,
		ThenKey:  bonnyPallLandPicked,
	})
	return nil
})

// bonnyPallLandPicked puts the picked land onto the battlefield from the
// zone that holds it now. An on-disk key: never renamed, never reused.
var bonnyPallLandPicked = game.RegisterChooseCardsThen("choose-cards/bonny-pall-land-from-hand-or-graveyard", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	id := r.Picked[0]
	z := g.FindCardZoneForEffect(id)
	if z == nil {
		return nil
	}
	switch z.Kind {
	case game.ZoneHand:
		return g.PutFromHandOntoBattlefieldKeyedForEffect(id, game.HandEntryOptions{Controller: r.Chooser}, game.KeyedTail{})
	case game.ZoneGraveyard:
		return g.ReturnFromGraveyardToBattlefieldKeyedForEffect(id, r.Chooser, false, game.KeyedTail{})
	}
	return nil
})
