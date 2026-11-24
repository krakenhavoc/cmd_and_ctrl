package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Pandorica — Legendary Artifact {2}{W}:
//
//	"You may choose not to untap The Pandorica during your untap step.
//	 {1}{W}, {T}: Untap another target nonland permanent, then it phases
//	 out. It can't phase in for as long as The Pandorica remains tapped.
//	 When The Pandorica becomes untapped or leaves the battlefield, that
//	 permanent phases in. Activate only as a sorcery."
//
// The opt-out is Rust Tick's (CR 502.3's determination, asked during the
// untap step).
//
// The activation untaps the target, then phases it out with everything
// attached to it (CR 702.26g). The #6095 bar
// (game.CantPhaseInWhileTappedForEffect) keeps it out through every untap
// step, its controller's included (CR 502.1), for as long as The Pandorica
// is on the battlefield as the same object and tapped. It never starts if
// The Pandorica was untapped in response (CR 611.2b), and the permanent is
// then on the ordinary clock.
//
// "When The Pandorica becomes untapped or leaves the battlefield, that
// permanent phases in" is a delayed triggered ability the resolution
// creates (CR 603.7), waiting on either event for as long as the game
// lasts and firing once: on the first untap of The Pandorica as the object
// that made it, or as it leaves the battlefield. It goes on the stack
// (CR 603.7c), and as it resolves the permanent phases in if it is still
// phased out as the same object (phasing keeps the object, CR 702.26d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "94d5c904-6504-4df2-b242-daf88a988475",
		Name:         "The Pandorica",
		Completeness: effects.CompletenessFull,
		UntapOptOuts: []game.UntapOptOut{
			effects.MayChooseNotToUntapSelf("The Pandorica — you may choose not to untap The Pandorica"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "{1}{W}, {T}: Untap another target nonland permanent, then it phases out. It can't phase in for as long as The Pandorica remains tapped. When The Pandorica becomes untapped or leaves the battlefield, that permanent phases in. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{1}{W}"), effects.TapCost()),
			Targets:      effects.Another(effects.TargetPermanent("another target nonland permanent", effects.Nonland())),
			SorcerySpeed: true,
			Effect:       thePandoricaActivate,
		}},
	})
}

const thePandoricaPhaseInLabel = "The Pandorica — that permanent phases in"

// thePandoricaPhaseInBody is the delayed trigger's body: the permanent its
// Params name phases in if it is still phased out as that object.
var thePandoricaPhaseInBody = game.DelayedBody("the-pandorica/that-permanent-phases-in", func(g *game.Game, _ *game.StackItem, p game.EffectParams) error {
	for _, c := range g.PhasedOutCardsForEffect() {
		if c.InstanceID == p.Object.ID && c.ObjectEpoch == p.Object.Epoch {
			g.PhaseInForEffect(c.InstanceID)
			return nil
		}
	}
	return nil
})

// thePandoricaUntappedOrLeft is the delayed trigger's event: The Pandorica,
// as the object that made the trigger, becoming untapped or leaving the
// battlefield.
var thePandoricaUntappedOrLeft = game.DelayedCondition("the-pandorica/untapped-or-left", func(ev game.Event, dt *game.DelayedTrigger, g *game.Game, _ game.EffectParams) bool {
	if ev.CardID != dt.SourceObject.ID {
		return false
	}
	switch ev.Kind {
	case game.EventLTB:
		return true
	case game.EventUntapCard:
		c, ok := g.LookupCardForEffect(ev.CardID)
		return ok && (dt.SourceObject.Epoch == 0 || c.ObjectEpoch == dt.SourceObject.Epoch)
	}
	return false
})

// thePandoricaActivate is the activated ability's effect.
func thePandoricaActivate(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := (effects.UntapTarget{Target: id}).Apply(ctx); err != nil {
		return err
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	if err := g.PhaseOutForEffect(item.SourceCardID, id); err != nil {
		return err
	}
	if !g.IsPhasedOutForEffect(id) {
		return nil
	}
	if !ctx.IsNewSourceObject(item.SourceCardID) {
		g.CantPhaseInWhileTappedForEffect(item.SourceCardID, id)
	}
	d := game.IndefiniteDuration()
	g.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
		Controller:   item.Controller,
		SourceCardID: item.SourceCardID,
		Label:        thePandoricaPhaseInLabel,
		On:           []game.EventKind{game.EventUntapCard, game.EventLTB},
		Condition:    thePandoricaUntappedOrLeft,
		Body:         thePandoricaPhaseInBody,
		Params:       game.EffectParams{Object: game.ObjectRef{ID: id, Epoch: c.ObjectEpoch}},
		Duration:     &d,
	})
	return nil
}
