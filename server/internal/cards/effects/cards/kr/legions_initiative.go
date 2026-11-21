package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Legion's Initiative — Enchantment {R}{W} (S82 GAP, #5948):
//
//	"Red creatures you control get +1/+0.
//	 White creatures you control get +0/+1.
//	 {R}{W}, Exile this enchantment: Exile all creatures you control. At the
//	 beginning of the next combat, return those cards to the battlefield under
//	 their owner's control and those creatures gain haste until end of turn."
//
// The exile is one mass event; its continuation schedules ONE delayed trigger
// at the next beginning of combat (CR 603.7) over the cards that reached exile.
// The return is one entry (CR 603.6a), and each card comes back as a new object
// (CR 400.7), so the haste is granted to the new objects the entry reports.
// A token is exiled and ceases to exist, so it does not return (CR 111.8).
func init() {
	legionsInitiativeSchedule = effects.MassTail("legions-initiative/schedule-return-at-next-combat",
		func(ctx *effects.Context, swept []game.Card, _ int) error {
			ids := make([]uuid.UUID, 0, len(swept))
			for _, c := range swept {
				ids = append(ids, c.InstanceID)
			}
			if len(ids) == 0 {
				return nil
			}
			return effects.ScheduleDelayedTrigger{
				At:    game.StepBeginCombat,
				Label: "Legion's Initiative — return the exiled creatures; they gain haste",
				Cards: ids,
				Body:  legionsInitiativeReturn,
			}.Apply(ctx)
		})
	effects.Register(effects.Spec{
		OracleID:     "7b23bceb-1b82-4c9c-b7c1-8e8ef4643fb0",
		Name:         "Legion's Initiative",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.B16Anthem(legionsInitiativeCreaturesOfColor("R"), 1, 0),
			effects.B16Anthem(legionsInitiativeCreaturesOfColor("W"), 0, 1),
		},
		Activated: []effects.ActivatedAbility{{
			Label: "{R}{W}, Exile this enchantment: Exile all creatures you control. At the beginning of the next combat, return those cards to the battlefield under their owner's control and those creatures gain haste until end of turn.",
			// ADR 0126 §6: exiles only the creatures its controller controls, and they come back.
			Purpose: game.Purpose{Answers: game.AnswerSurvive, Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepExile, Partial: true}},
			Cost:    effects.Plus(effects.ManaCost("{R}{W}"), effects.ExileThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.ExileAllMatching{
					Match:    effects.And(effects.Creature(), effects.YouControl()),
					ThenTail: ctx.ItemTail(legionsInitiativeSchedule),
				}.Apply(ctx)
			},
		}},
	})
}

var legionsInitiativeSchedule game.TailRef

// legionsInitiativeReturn is the delayed trigger's body: the cards still in
// exile come back together under their owners' control, and the entry tells
// the haste tail which new objects arrived.
var legionsInitiativeReturn = game.SimpleDelayedBody("legions-initiative/return-with-haste", func(g *game.Game, item *game.StackItem) error {
	var ids []uuid.UUID
	for _, t := range item.Targets {
		if t.Kind == game.TargetCard && t.ID != uuid.Nil {
			ids = append(ids, t.ID)
		}
	}
	return effects.ReturnFromExileTogether{Targets: ids, ThenTail: legionsInitiativeHaste}.Apply(effects.NewContext(g, item))
})

var legionsInitiativeHaste = effects.CardTail("legions-initiative/haste-for-the-returned", func(ctx *effects.Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		if !effects.OnBattlefield(ctx.Game, id) {
			continue
		}
		if err := (effects.GrantKeywordUntilEOT{
			Target: id, Keywords: []string{"haste"}, Label: "Legion's Initiative — haste until end of turn",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})

// legionsInitiativeCreaturesOfColor is "<color> creatures you control": the
// colour is a word of the card's text (CR 612.2), read as the creature is now.
func legionsInitiativeCreaturesOfColor(color string) func(target *game.Card, g *game.Game, source *game.Card) bool {
	return func(target *game.Card, _ *game.Game, source *game.Card) bool {
		if !target.IsCreature() || target.Controller != source.Controller {
			return false
		}
		want := source.TextColor(color)
		for _, c := range target.EffectiveColors() {
			if c == want {
				return true
			}
		}
		return false
	}
}
