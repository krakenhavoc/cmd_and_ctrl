package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Anzrag's Rampage — Sorcery {3}{R}{R} (Murders at Karlov Manor, #4469):
//
//	"Destroy all artifacts you don't control, then exile the top X cards of
//	 your library, where X is the number of artifacts that were put into
//	 graveyards from the battlefield this turn. You may put a creature card
//	 exiled this way onto the battlefield. It gains haste. Return it to your
//	 hand at the beginning of the next end step."
//
// The sweep goes first and its keyed tail does the rest, so a commander caught
// in it asks its question before X is read. X counts every artifact put into a
// graveyard from the battlefield this turn, whoever's, destroyed by this or
// not, tokens included (the ruling): leaves-the-battlefield records for a
// graveyard, read as they last existed (CR 603.10a). The exile is the top of
// your library; the creature card is an optional pick among those exiled,
// entering under your control, gaining haste until end of turn, and a delayed
// trigger returns it to your hand at the beginning of the next end step if it
// is still the object that entered (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ce690a21-8117-420e-be00-cd6d9923a812",
		Name:         "Anzrag's Rampage",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepArtifacts, How: game.SweepDestroy, OpponentsOnly: true}, Answers: game.AnswerRemove},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DestroyAllMatching{
				Match:    And(Artifact(), Not(YouControl())),
				ThenTail: ctx.ItemTail(anzragsRampageAfterSweep),
			}.Apply(ctx)
		},
	})
}

// artifactsPutIntoGraveyardsFromTheBattlefieldThisTurn counts this turn's
// leaves-the-battlefield events to a graveyard of an object that was an
// artifact as it left.
func artifactsPutIntoGraveyardsFromTheBattlefieldThisTurn(g *game.Game) int {
	n := 0
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneGraveyard {
			continue
		}
		if was, known := ev.WasType("artifact"); known && was {
			n++
		}
	}
	return n
}

var anzragsRampageAfterSweep = CardTail("tail/anzrags-rampage-after-sweep", func(ctx *Context, _ game.TailOutcome) error {
	x := artifactsPutIntoGraveyardsFromTheBattlefieldThisTurn(ctx.Game)
	if x <= 0 {
		return nil
	}
	return ctx.Game.ExileTopKeyedForEffect(ctx.Controller(), ctx.Controller(), x, ctx.Tail(anzragsRampageExiled))
})

// anzragsRampageExiled is "you may put a creature card exiled this way onto
// the battlefield": the pick, over the creature cards that reached exile.
var anzragsRampageExiled = CardTail("tail/anzrags-rampage-exiled", func(ctx *Context, o game.TailOutcome) error {
	var creatures []uuid.UUID
	for _, id := range o.IDs {
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.IsCreature() {
			if z := ctx.Game.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneExile {
				creatures = append(creatures, id)
			}
		}
	}
	if len(creatures) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Anzrag's Rampage — you may put a creature card exiled this way onto the battlefield",
		Cards:    creatures,
		Min:      0,
		Max:      1,
		LiveZone: game.ZoneExile,
		ThenKey:  anzragsRampagePicked,
	})
	return nil
})

var anzragsRampagePicked = game.RegisterChooseCardsThen("choose-cards/anzrags-rampage-creature", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) != 1 {
		return nil
	}
	id := r.Picked[0]
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
		return nil
	}
	return g.PutOntoBattlefieldTogetherKeyedForEffect([]game.BatchEntry{{CardID: id, From: game.ZoneExile}},
		game.ZoneEntryOptions{Controller: r.Chooser},
		game.KeyedTail{Ref: anzragsRampageEntered, Player: r.Chooser, Source: r.Source})
})

// anzragsRampageEntered is "It gains haste. Return it to your hand at the
// beginning of the next end step.", told the permanent that entered.
var anzragsRampageEntered = game.RegisterTail("tail/anzrags-rampage-entered", func(g *game.Game, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	id := o.IDs[0]
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	ctx := keyedContext(g, o.Tail.Player, o.Tail.Source)
	if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"}, Label: "Anzrag's Rampage — haste"}).Apply(ctx); err != nil {
		return err
	}
	g.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
		Controller:   o.Tail.Player,
		SourceCardID: o.Tail.Source,
		Label:        "Anzrag's Rampage — return it to your hand",
		At:           game.StepEnd,
		Cards:        []uuid.UUID{id},
		Body:         zaraReturnToHandBody,
		Params:       game.EffectParams{Object: game.ObjectRef{ID: id, Epoch: c.ObjectEpoch}},
	})
	return nil
})
