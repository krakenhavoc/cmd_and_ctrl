package cardsgr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Grave Sifter — Creature — Elemental Beast {5}{G} 5/7 (Commander 2014):
//
//	"When this creature enters, each player chooses a creature type and
//	 returns any number of cards of that type from their graveyard to
//	 their hand."
//
// Each player in turn order from the controller (CR 101.4) names a creature
// type, then picks any number of their own graveyard's cards of it
// (Card.HasSubtype, so a changeling matches every type) and returns them.
// A player whose graveyard holds none skips the pick. The chain is one
// registered tail and one registered pick continuation, so every pause is a
// restore point.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "07a26388-773e-4614-a3de-f133c6694d59",
		Name:         "Grave Sifter",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerGraveyard},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Grave Sifter — each player chooses a creature type and returns any number of cards of that type from their graveyard to their hand",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return graveSifterAsk(ctx, effects.SeatsFromController(g, item.Controller))
				}),
		},
	})
}

func graveSifterAsk(ctx *effects.Context, seats []uuid.UUID) error {
	if len(seats) == 0 {
		return nil
	}
	then := ctx.ItemTail(graveSifterTypeChosen)
	then.Carry = append([]uuid.UUID(nil), seats...)
	if !effects.ChooseCreatureTypeTail(ctx.Game, seats[0], ctx.Source(), "Grave Sifter — choose a creature type", then) {
		return graveSifterAsk(ctx, seats[1:])
	}
	return nil
}

var graveSifterTypeChosen game.TailRef
var graveSifterReturn game.ChooseCardsThen

func init() {
	graveSifterTypeChosen = effects.CardTail("then/grave-sifter-type-chosen", graveSifterPickType)
	graveSifterReturn = game.RegisterChooseCardsThen("choose-cards/grave-sifter-return", graveSifterReturnPicked)
}

func graveSifterPickType(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	seat, rest := o.Tail.Carry[0], o.Tail.Carry[1:]
	p := ctx.Game.PlayerByIDForEffect(seat)
	if o.Tail.Word == "" || p == nil || p.Graveyard == nil {
		return graveSifterAsk(ctx, rest)
	}
	var cards []uuid.UUID
	for _, c := range p.Graveyard.Cards {
		if c.HasSubtype(o.Tail.Word) {
			cards = append(cards, c.InstanceID)
		}
	}
	if len(cards) == 0 {
		return graveSifterAsk(ctx, rest)
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:   seat,
		Source:    ctx.Source(),
		Question:  "Grave Sifter — return any number of " + o.Tail.Word + " cards from your graveyard to your hand",
		Cards:     cards,
		Min:       0,
		Max:       len(cards),
		ThenKey:   graveSifterReturn,
		Carry:     append([]uuid.UUID(nil), rest...),
		LiveZone:  game.ZoneGraveyard,
		Asked:     game.NewAskedItem(ctx.Item),
		CarryWord: o.Tail.Word,
	})
	return nil
}

func graveSifterReturnPicked(g *game.Game, r game.CardsChosen) error {
	if r.Asked == nil {
		return nil
	}
	ctx := effects.NewContext(g, r.Asked)
	for _, id := range r.Picked {
		if err := (effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
			return err
		}
	}
	return graveSifterAsk(ctx, r.Carry)
}
