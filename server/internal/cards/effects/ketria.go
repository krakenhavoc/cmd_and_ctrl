package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ketria — Plane — Ikoria (Commander Masters, #4870):
//
//	"When you planeswalk to Ketria and at the beginning of your upkeep, put
//	 your choice of a vigilance, menace, or trample counter on target
//	 creature you control.
//	 Whenever chaos ensues, exile cards from the top of your library until
//	 you exile a nonland permanent card. Put that card onto the battlefield
//	 or into your hand."
//
// A plane (ADR 0216 §6). The first ability is one ability with two trigger
// conditions that targets a creature the planar controller controls; the
// counter is chosen as it resolves (the ruling) and is a keyword counter (CR
// 122.1b). The chaos ability (CR 311.7) exiles from the top of the planar
// controller's library until a nonland permanent card is exiled; the cards
// exiled before it stay exiled (the ruling), and the planar controller puts
// that card onto the battlefield or into their hand.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f3b27908-ca10-4793-ab8e-6639c73cfadd",
		Name:         "Ketria",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenYouPlaneswalkToThisAndAtYourUpkeep("Ketria — put your choice of a vigilance, menace, or trample counter on target creature you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					t, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return PickOption{
						Question: "Ketria — put a vigilance, menace, or trample counter on the creature",
						Options: []game.ChoiceOption{
							{Label: "Vigilance counter"}, {Label: "Menace counter"}, {Label: "Trample counter"},
						},
						ThenKey: ketriaCounterChosen,
						Carry:   []uuid.UUID{t},
					}.Apply(ctx)
				}), TargetCreature("target creature you control", YouControl())),
			WheneverChaosEnsues("Ketria — exile cards from the top of your library until you exile a nonland permanent card; put it onto the battlefield or into your hand",
				func(g *game.Game, item *game.StackItem) error {
					return g.MillToZoneKeyedForEffect(item.Controller, 0, game.ZoneExile, UntilNonlandPermanent, NewContext(g, item).ItemTail(ketriaExiled))
				}),
		},
	})
}

// ketriaCounterChosen puts the chosen keyword counter on the creature the
// ability targeted. An on-disk key: never renamed, never reused.
var ketriaCounterChosen = OptionPickThen("option-pick/ketria-counter", func(ctx *Context, r game.OptionPicked) error {
	if len(r.Carry) == 0 || r.Index < 0 || r.Index > 2 {
		return nil
	}
	kind := []string{game.CounterVigilance, game.CounterMenace, game.CounterTrample}[r.Index]
	return AddCounter{Target: r.Carry[0], Kind: kind, N: 1}.Apply(ctx)
})

// ketriaExiled is the run's "then": the nonland permanent card it stopped
// on (the last card exiled) goes onto the battlefield or into the hand, at
// the planar controller's choice. An on-disk key.
var ketriaExiled = CardTail("ketria/exiled-until-a-nonland-permanent", func(ctx *Context, o game.TailOutcome) error {
	var hit uuid.UUID
	for _, id := range o.IDs {
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && !c.IsLand() && isPermanentCard(c) {
			hit = id
		}
	}
	if hit == uuid.Nil {
		return nil
	}
	return PickOption{
		Question: "Ketria — put the exiled card onto the battlefield or into your hand",
		Options:  []game.ChoiceOption{{Label: "Onto the battlefield"}, {Label: "Into your hand"}},
		ThenKey:  ketriaPlaced,
		Carry:    []uuid.UUID{hit},
	}.Apply(ctx)
})

// ketriaPlaced puts the exiled card where the planar controller chose. An
// on-disk key.
var ketriaPlaced = OptionPickThen("option-pick/ketria-placement", func(ctx *Context, r game.OptionPicked) error {
	if len(r.Carry) == 0 {
		return nil
	}
	if z := ctx.Game.FindCardZoneForEffect(r.Carry[0]); z == nil || z.Kind != game.ZoneExile {
		return nil
	}
	if r.Index == 1 {
		return ctx.Game.BounceToHandForEffect(r.Carry[0])
	}
	return ReturnFromExile{Target: r.Carry[0], Controller: ctx.Controller()}.Apply(ctx)
})
