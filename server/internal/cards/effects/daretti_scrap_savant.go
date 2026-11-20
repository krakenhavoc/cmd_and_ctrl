package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Daretti, Scrap Savant — Legendary Planeswalker — Daretti {3}{R}, loyalty
// 3 (Commander 2014, #5553):
//
//	"+2: Discard up to two cards, then draw that many cards.
//	 −2: Sacrifice an artifact. If you do, return target artifact card from
//	 your graveyard to the battlefield.
//	 −10: You get an emblem with "Whenever an artifact is put into your
//	 graveyard from the battlefield, return that card to the battlefield at
//	 the beginning of the next end step."
//	 Daretti, Scrap Savant can be your commander."
//
// The +2 is Greasewrench Goblin's rummage (b39MayDiscardThenDraw): zero to
// two cards chosen as it resolves, then that many drawn. The −2 targets the
// artifact card as it is activated, so the artifact sacrificed as it resolves
// can't be it; the sacrifice is not optional, and only once one has really
// been sacrificed is the target returned (Daretti, Rocketeer Engineer's
// sacrifice-then-return run). The emblem watches every artifact of yours that
// goes from the battlefield to your graveyard and schedules its return for
// the next end step; a token has ceased to exist by then (CR 111.7), and a
// card that has left the graveyard is a new object (CR 400.7), so neither
// comes back. "Can be your commander" is
// deck construction (CR 903.3a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ea9f4de3-fc35-435a-8dff-a5657d14a9bd",
		Name:            "Daretti, Scrap Savant",
		Completeness:    CompletenessFull,
		StartingLoyalty: 3,
		Emblem: &EmblemSpec{
			Label: "Daretti, Scrap Savant emblem",
			Text:  "Whenever an artifact is put into your graveyard from the battlefield, return that card to the battlefield at the beginning of the next end step.",
			Triggered: []game.TriggeredAbility{
				On(game.EventLTB, darettiScrapSavantArtifactToYourGraveyard,
					"Daretti, Scrap Savant emblem — return that card to the battlefield at the beginning of the next end step",
					darettiScrapSavantScheduleReturn),
			},
		},
		Activated: []ActivatedAbility{
			{
				Label: "+2: Discard up to two cards, then draw that many cards.",
				Cost:  LoyaltyCost(2),
				Effect: Do(Instruction(b39MayDiscardThenDraw(2, false,
					"Daretti, Scrap Savant — discard up to two cards, then draw that many", drawAsManyAsDiscarded))),
			},
			{
				Label:   "−2: Sacrifice an artifact. If you do, return target artifact card from your graveyard to the battlefield.",
				Cost:    LoyaltyCost(-2),
				Targets: TargetCardInGraveyard("target artifact card from your graveyard", YouOwn(), Artifact()),
				Effect:  Do(Instruction(darettiSacrificeThenReturn)),
			},
			{
				Label:  "−10: You get an emblem with \"Whenever an artifact is put into your graveyard from the battlefield, return that card to the battlefield at the beginning of the next end step.\"",
				Cost:   LoyaltyCost(-10),
				Effect: Do(CreateEmblem{}),
			},
		},
	})
}

// darettiScrapSavantArtifactToYourGraveyard is the emblem's "an artifact is
// put into your graveyard from the battlefield": an artifact that left the
// battlefield for a graveyard, owned by the emblem's owner.
func darettiScrapSavantArtifactToYourGraveyard(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !b23ArtifactPutIntoGraveyardFromBattlefield(ev, g) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.Owner == source.Controller
}

// darettiScrapSavantScheduleReturn schedules the return of the card the
// trigger names, as the object it now is in the graveyard.
func darettiScrapSavantScheduleReturn(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	id := item.Trigger.Event.CardID
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	return ScheduleDelayedTrigger{
		Label:  "Daretti, Scrap Savant emblem — return that card to the battlefield",
		Cards:  []uuid.UUID{id},
		Body:   darettiScrapSavantReturnBody,
		Params: game.EffectParams{Object: game.ObjectRef{ID: id, Epoch: c.ObjectEpoch}},
	}.Apply(NewContext(g, item))
}

// darettiScrapSavantReturnBody returns the card to the battlefield under the
// delayed trigger's controller's control, if it is still the same object in
// a graveyard.
var darettiScrapSavantReturnBody = game.DelayedBody("delayed/daretti-scrap-savant-return",
	func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
		if !objectStillInAGraveyard(g, p.Object) {
			return nil
		}
		return ReturnFromGraveyard{Target: p.Object.ID, Dest: game.ZoneBattlefield, Controller: item.Controller}.Apply(NewContext(g, item))
	})
