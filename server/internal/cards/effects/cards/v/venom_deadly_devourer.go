package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Venom, Deadly Devourer — Legendary Creature — Symbiote Villain
// {2}{B}{G}, 4/4 (Marvel's Spider-Man, #4460):
//
//	"Vigilance, menace
//	 {3}: Exile target creature card from a graveyard. When you do, put X
//	 +1/+1 counters on target Symbiote, where X is the exiled card's
//	 toughness."
//
// The exile is the activated ability's effect; "when you do" is a CR
// 603.12 reflexive trigger queued by the exile's keyed tail, so it only
// happens if the card actually reached exile (a replacement that sends it
// elsewhere, or a target that left, makes no trigger). The Symbiote is
// chosen as the trigger goes on the stack (CR 603.3d) and X is the exiled
// card's toughness as it sits in exile, carried as Params.Amount.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f064fc04-28fb-423f-a4f3-bb567c71168f",
		Name:            "Venom, Deadly Devourer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance", "menace"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}: Exile target creature card from a graveyard. When you do, put X +1/+1 counters on target Symbiote, where X is the exiled card's toughness.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerGraveyard},
			Cost:    effects.ManaCost("{3}"),
			Targets: effects.TargetCardInGraveyard("target creature card from a graveyard", effects.Creature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					return effects.ExileTarget{Target: t.ID, ThenTail: ctx.ItemTail(venomDeadlyDevourerExiled)}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}

var venomDeadlyDevourerExiled = effects.CardTail("exile-then/venom-deadly-devourer-counters", func(ctx *effects.Context, o game.TailOutcome) error {
	if !effects.OneWentThisWay(o) {
		return nil
	}
	exiled, ok := ctx.Game.LookupCardForEffect(o.IDs[0])
	if !ok {
		return nil
	}
	return effects.ReflexiveTrigger{
		Label:  "Venom, Deadly Devourer — put X +1/+1 counters on target Symbiote",
		Body:   venomDeadlyDevourerBody,
		Params: game.EffectParams{Amount: exiled.CurrentToughness()},
	}.Apply(ctx)
})

var venomDeadlyDevourerBody = game.ReflexiveBody("venom-deadly-devourer/counters", venomDeadlyDevourerCounters,
	effects.ConstTargets(func() *game.TargetSpec {
		return effects.TargetPermanent("target Symbiote", effects.Subtype("Symbiote"))
	}))

func venomDeadlyDevourerCounters(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	if p.Amount <= 0 {
		return nil
	}
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: p.Amount}.Apply(ctx)
		}
	}
	return nil
}
