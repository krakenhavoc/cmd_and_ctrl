package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Specter of Mortality — Creature — Specter {3}{B}{B}, 4/4 (Wilds of
// Eldraine, #4471):
//
//	"Flying
//	 When this creature enters, you may exile one or more creature cards
//	 from your graveyard. When you do, each other creature gets -X/-X
//	 until end of turn, where X is the number of cards exiled this way."
//
// The "may" is a choose-cards prompt over your graveyard's creature cards
// with a floor of zero; picking none ends it, and nothing is exiled.
// "When you do" is a reflexive trigger (CR 603.12) raised once at least
// one card really reached exile, carrying that count as its X, so the
// sweep is on the stack and can be answered; it shrinks every creature
// but the Specter, read as it resolves, and fixes X when it was raised.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "caff40b1-28d1-4ef6-90d1-51a1dc06cbcd",
		Name:            "Specter of Mortality",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		// -X/-X to every other creature, X the cards exiled: a partial
		// minus sweep whose amount is its X (ADR 0126).
		Purpose: game.Purpose{Sweep: game.Sweep{
			Matches: game.SweepCreatures, How: game.SweepMinus, AmountIsX: true, Partial: true}},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Specter of Mortality — you may exile one or more creature cards from your graveyard; when you do, each other creature gets -X/-X until end of turn",
				specterOfMortalityEnters),
		},
	})
}

func specterOfMortalityEnters(g *game.Game, item *game.StackItem) error {
	p := g.PlayerByIDForEffect(item.Controller)
	if p == nil || p.Graveyard == nil {
		return nil
	}
	var offer []uuid.UUID
	for _, c := range p.Graveyard.Cards {
		if c.IsCreature() {
			offer = append(offer, c.InstanceID)
		}
	}
	if len(offer) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    item.Controller,
		FromPlayer: item.Controller,
		Source:     item.SourceCardID,
		Question:   "Specter of Mortality — you may exile one or more creature cards from your graveyard",
		Cards:      offer,
		Min:        0,
		Max:        len(offer),
		LiveZone:   game.ZoneGraveyard,
		ThenKey:    specterOfMortalityExiled,
		Asked:      game.NewAskedItem(item),
	})
	return nil
}

// specterOfMortalityExiled exiles the picked cards and, when at least one
// reached exile, raises the reflexive sweep with that count. An on-disk
// key: never renamed, never reused.
var specterOfMortalityExiled = game.RegisterChooseCardsThen("choose-cards/specter-of-mortality-exile", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 || r.Asked == nil {
		return nil
	}
	ctx := NewContext(g, r.Asked)
	return g.ExileCardsKeyedForEffect(r.Picked, ctx.ItemTail(specterOfMortalityExiledTail))
})

// specterOfMortalityExiledTail raises the reflexive sweep once the exile has
// settled, with the number of cards that really reached exile. An on-disk key.
var specterOfMortalityExiledTail = CardTail("specter-of-mortality/after-exile", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return ReflexiveTrigger{
		Label:  "Specter of Mortality — each other creature gets -X/-X until end of turn, where X is the number of cards exiled this way",
		Body:   specterOfMortalitySweepBody,
		Params: game.EffectParams{Amount: len(o.IDs)},
	}.Apply(ctx)
})

// specterOfMortalitySweepBody is the reflexive trigger: -X/-X to each
// creature other than the Specter.
var specterOfMortalitySweepBody = game.ReflexiveBody("specter-of-mortality/sweep", func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	return BoostUntilEOT{
		Match:     And(Creature(), OtherThan(item.SourceCardID)),
		Power:     -p.Amount,
		Toughness: -p.Amount,
		Label:     "Specter of Mortality — -X/-X",
	}.Apply(NewContext(g, item))
}, nil)
