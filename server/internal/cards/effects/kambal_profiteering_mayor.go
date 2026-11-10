package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kambal, Profiteering Mayor — Legendary Creature — Human Advisor {1}{W}{B},
// 2/3 (Outlaws of Thunder Junction, #4468):
//
//	"Whenever one or more tokens your opponents control enter, for each of
//	 them, create a tapped token that's a copy of it. This ability triggers
//	 only once each turn.
//	 Whenever one or more tokens you control enter, each opponent loses 1
//	 life and you gain 1 life."
//
// Both are one trigger per batch of entries (OncePerBatch, CR 603.2c). The
// first reads every token that entered in the trigger's batch under an
// opponent's control, as it last existed (CR 707.2, 608.2h), and makes one
// tapped copy of each under Kambal's controller; the once-each-turn limit
// is the engine's per-object tally. The copies are Kambal's controller's
// tokens, so they trigger the second ability and never the first.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "619e0686-e88c-4238-8364-75395e733533",
		Name:         "Kambal, Profiteering Mayor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggersOnlyOnceEachTurn(OncePerBatch(On(game.EventETB,
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b35TokenEnteredUnderAnOpponent(ev, source, g)
				},
				"Kambal, Profiteering Mayor — for each of them, create a tapped token that's a copy of it",
				kambalCopyThem))),
			OncePerBatch(On(game.EventETB,
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					c, ok := enteredUnderYourControl(ev, source, g, false)
					return ok && c.IsToken()
				},
				"Kambal, Profiteering Mayor — each opponent loses 1 life and you gain 1 life",
				Do(EachOpponentLosesLife{N: 1}, GainLife{Amount: 1}))),
		},
	})
}

// b35TokenEnteredUnderAnOpponent is an ETB of a token under an opponent of
// the source's controller.
func b35TokenEnteredUnderAnOpponent(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventETB || ev.CardID == uuid.Nil || source == nil {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsToken() && c.Controller != source.Controller
}

// kambalCopyThem makes a tapped copy of each token that entered in the
// trigger's batch under an opponent's control.
func kambalCopyThem(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	batch := ctx.Trigger().Event.Batch
	seen := map[uuid.UUID]bool{}
	var refs []game.ObjectRef
	for _, e := range g.Events {
		if e.Batch != batch || e.Kind != game.EventETB || seen[e.CardID] {
			continue
		}
		c, ok := g.LookupCardForEffect(e.CardID)
		if !ok || !c.IsToken() || c.Controller == item.Controller {
			continue
		}
		seen[e.CardID] = true
		if ref, ok := g.PermanentRefForEffect(e.CardID); ok {
			refs = append(refs, ref)
		}
	}
	for _, ref := range refs {
		tmpl, ok := g.TokenCopyTemplateForObjectForEffect(ref)
		if !ok {
			continue
		}
		tmpl.Tapped = true
		if err := createTokenCopies(ctx, item.Controller, tmpl, 1); err != nil {
			return err
		}
	}
	return nil
}
