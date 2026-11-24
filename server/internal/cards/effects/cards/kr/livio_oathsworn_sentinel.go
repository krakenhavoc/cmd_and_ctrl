package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Livio, Oathsworn Sentinel — Legendary Creature — Human Knight {1}{W}, 2/2
// (Commander Legends, #5507):
//
//	"{1}{W}: Choose another target creature. Its controller may exile it with
//	 an aegis counter on it.
//	 {2}{W}, {T}: Return all exiled cards with aegis counters on them to the
//	 battlefield under their owners' control.
//	 Partner"
//
// The first ability targets another creature (any controller's); as it
// resolves its CONTROLLER, not Livio's, is asked whether to exile it, and a
// yes exiles it and puts an aegis counter on the card once it really is in
// exile (a replacement that sent it elsewhere gets none, and a token is gone).
// The second returns, as one entry, every card in any exile with an aegis
// counter, each under its owner's control, whoever exiled it. Partner is a
// deck-construction rule.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "53c6f2cf-ed3e-4c0c-ad42-56a6ce1622b6",
		Name:         "Livio, Oathsworn Sentinel",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}{W}: Choose another target creature. Its controller may exile it with an aegis counter on it.",
				Purpose: game.Purpose{Answers: game.AnswerProtect},
				Cost:    effects.ManaCost("{1}{W}"),
				Targets: effects.Another(effects.TargetCreature("another target creature")),
				Effect:  livioOfferExile,
			},
			{
				Label:   "{2}{W}, {T}: Return all exiled cards with aegis counters on them to the battlefield under their owners' control.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{2}{W}"), effects.TapCost()),
				Effect:  livioReturnAegis,
			},
		},
	})
}

const livioAegisCounter = "aegis"

// livioOfferExile asks the chosen creature's controller whether to exile it.
func livioOfferExile(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id := effects.FirstLegalCardTarget(ctx)
	if id == uuid.Nil {
		return nil
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	return effects.MayChoice{
		Player:   c.Controller,
		Question: "Livio, Oathsworn Sentinel — exile " + c.Name + " with an aegis counter on it?",
		YesLabel: "Exile it",
		NoLabel:  "Keep it",
		Then:     livioExile,
		Carry:    []uuid.UUID{id},
	}.Apply(ctx)
}

// livioExile is the "yes": exile the creature, then put the counter on the
// card that reached exile. An on-disk key: never renamed, never reused.
var livioExile = effects.MayYes("confirm/livio-oathsworn-sentinel-exile", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	if len(r.Carry) == 0 {
		return nil
	}
	if z := ctx.Game.FindCardZoneForEffect(r.Carry[0]); z == nil || z.Kind != game.ZoneBattlefield {
		return nil
	}
	return effects.ExileTarget{Target: r.Carry[0], ThenTail: ctx.ItemTail(livioExiled)}.Apply(ctx)
})

// livioExiled puts the aegis counter on each card that reached exile. An
// on-disk key.
var livioExiled = effects.CardTail("livio-oathsworn-sentinel/aegis-counter", func(ctx *effects.Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
			continue
		}
		if err := ctx.Game.AddCounterForEffect(id, livioAegisCounter, 1); err != nil {
			return err
		}
	}
	return nil
})

// livioReturnAegis returns every exiled card with an aegis counter at once.
func livioReturnAegis(g *game.Game, item *game.StackItem) error {
	var ids []uuid.UUID
	if g.Exile != nil {
		for _, c := range g.Exile.Cards {
			if c.Counters[livioAegisCounter] > 0 {
				ids = append(ids, c.InstanceID)
			}
		}
	}
	if len(ids) == 0 {
		return nil
	}
	return effects.ReturnFromExileTogether{Targets: ids}.Apply(effects.NewContext(g, item))
}
