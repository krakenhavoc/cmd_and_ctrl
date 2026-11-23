package cardsb

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bag of Devouring — Artifact {B} (Adventures in the Forgotten Realms
// Commander, S82):
//
//	"Whenever you sacrifice another nontoken artifact or creature, exile it.
//	 {2}, {T}, Sacrifice another artifact or creature: Draw a card.
//	 {3}, {T}, Sacrifice this artifact: Roll a d10. Return up to X cards
//	 from among cards exiled with this artifact to their owners' hands,
//	 where X is the result."
//
// The first and third abilities are linked (CR 607.2a). The trigger fires on
// the sacrifice event while the permanent is still on the battlefield (so a
// token and a permanent that is neither an artifact nor a creature are
// declined) and exiles the card from the graveyard it went to, linked to this
// Bag object (ExileTarget.Linked, as Colfenor's Urn). The sacrifice ability's
// own cost is a sacrifice, so it feeds the trigger too. The third ability
// sacrifices the Bag as a cost, so the link is read from the Bag's last-known
// object (a Bag that came back is a new object with no claim, CR 400.7); the d10
// goes through the one dice door and the player then picks up to X of the
// cards still in exile and linked, which go to their owners' hands.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4443345d-a292-4394-97dd-0e6f4d9aedf7",
		Name:         "Bag of Devouring",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventSacrifice, bagOfDevouringSacrificed,
				"Bag of Devouring — exile the sacrificed permanent", bagOfDevouringExile),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{2}, {T}, Sacrifice another artifact or creature: Draw a card.",
				Purpose: game.Purpose{Draws: 1, Answers: game.AnswerSacOutlet},
				Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.SacrificeAnotherN(1, "another artifact or creature", effects.Or(effects.Artifact(), effects.Creature()))),
				Effect:  effects.Do(effects.DrawCards{N: 1}),
			},
			{
				Label:   "{3}, {T}, Sacrifice this artifact: Roll a d10. Return up to X cards from among cards exiled with this artifact to their owners' hands.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost(), effects.SacrificeThis()),
				Effect:  effects.Do(effects.RollDice{Sides: 10, N: 1, Then: bagOfDevouringRolled}),
			},
		},
	})
}

// bagOfDevouringSacrificed is "whenever you sacrifice another nontoken
// artifact or creature": the event fires while the permanent is still on the
// battlefield, so its types are read as it was.
func bagOfDevouringSacrificed(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !effects.CmdrB9YouSacrificedAnotherNontoken(ev, source, game.Characteristic{}, g) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && (c.IsArtifact() || c.IsCreature())
}

// bagOfDevouringExile exiles the sacrificed card from the graveyard, linked
// to this Bag; a card that is no longer there is left alone.
func bagOfDevouringExile(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	id := item.Trigger.Event.CardID
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return effects.ExileTarget{Target: id, Linked: true}.Apply(effects.NewContext(g, item))
}

// bagOfDevouringRolled asks which of the exiled cards come back, after the
// d10. An on-disk key.
var bagOfDevouringRolled = effects.CardTail("dice/bag-of-devouring-rolled", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.Rolls) == 0 {
		return nil
	}
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	var ids []uuid.UUID
	for _, c := range ctx.Game.ExiledWithForEffect(game.PermissionCardRef(ref)) {
		ids = append(ids, c.InstanceID)
	}
	if len(ids) == 0 {
		return nil
	}
	x := o.Rolls[0]
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Bag of Devouring — return up to " + strconv.Itoa(x) + " cards exiled with it to their owners' hands",
		Cards:    ids,
		Min:      0,
		Max:      x,
		LiveZone: game.ZoneExile,
		ThenKey:  bagOfDevouringReturned,
	})
	return nil
})

// bagOfDevouringReturned returns each picked card still in exile to its
// owner's hand. An on-disk key.
var bagOfDevouringReturned = game.RegisterChooseCardsThen("choose-cards/bag-of-devouring-return", func(g *game.Game, r game.CardsChosen) error {
	ctx := effects.KeyedContext(g, r.Chooser, r.Source)
	for _, id := range r.Picked {
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
			continue
		}
		if err := (effects.BounceToHand{Target: id}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
