package cardsn

import (
	"slices"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nahiri's Resolve — Enchantment {3}{R}{W} (March of the Machine: The
// Aftermath, #5476):
//
//	"Creatures you control get +1/+0 and have haste.
//	 At the beginning of your end step, exile any number of nontoken
//	 artifacts and/or creatures you control. Return those cards to the
//	 battlefield under their owner's control at the beginning of your next
//	 upkeep."
//
// The first line is a layer 7c bonus and a layer 6 haste grant, read live.
// The end-step trigger asks for any number (none included) of the nontoken
// artifacts and creatures you control as it resolves (CR 608.2d, nothing is
// targeted), exiles them together, and schedules ONE delayed trigger for the
// beginning of YOUR next upkeep over exactly the cards that reached exile
// (CR 603.7); it returns them under their owners' control, the same body as
// every flicker. A card that left exile, or a token, is not returned.
//
// No simplification.
func init() {
	mine := func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.IsCreature() && target.Controller == source.Controller
	}
	effects.Register(effects.Spec{
		OracleID:     "12bd2051-297b-4bab-ac64-6eb26a42e3e2",
		Name:         "Nahiri's Resolve",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			{
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: mine,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power++
				},
			},
			effects.KeywordGrant(mine, "haste"),
		},
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("Nahiri's Resolve — exile any number of nontoken artifacts and/or creatures you control; return them at the beginning of your next upkeep",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					cands := nahirisResolveCandidates(g, ctx.Controller())
					if len(cands) == 0 {
						return nil
					}
					g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
						Chooser:  ctx.Controller(),
						Source:   ctx.Source(),
						Question: "Nahiri's Resolve — exile any number of nontoken artifacts and/or creatures you control",
						Cards:    cands,
						Min:      0,
						Max:      len(cands),
						ThenKey:  nahirisResolvePick,
						LiveZone: game.ZoneBattlefield,
						Asked:    game.NewAskedItem(ctx.Item),
					})
					return nil
				}),
		},
	})
}

// nahirisResolveCandidates are the nontoken artifacts and creatures `player`
// controls.
func nahirisResolveCandidates(g *game.Game, player uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && !c.IsToken() && (c.IsCreature() || c.IsArtifact()) {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// nahirisResolvePick exiles the chosen cards that are still nontoken
// artifacts or creatures its controller controls, then schedules the return.
var nahirisResolvePick = effects.AskedPick("choose-cards/nahiris-resolve-exile", func(ctx *effects.Context, r game.CardsChosen) error {
	live := nahirisResolveCandidates(ctx.Game, ctx.Controller())
	var ids []uuid.UUID
	for _, id := range r.Picked {
		if slices.Contains(live, id) && !slices.Contains(ids, id) {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	return ctx.Game.ExileCardsKeyedForEffect(ids, effects.WithWord(ctx.ItemTail(nahirisResolveSchedule),
		effects.ReturnExiledToOwnersBody.Key()+"|Nahiri's Resolve — return the exiled cards to the battlefield"))
})

// nahirisResolveSchedule schedules the return of what reached exile at the
// beginning of the controller's next upkeep. The body's key and the label
// ride the Word.
var nahirisResolveSchedule = effects.CardTail("nahiris-resolve/schedule-return-at-next-upkeep", func(ctx *effects.Context, o game.TailOutcome) error {
	bodyKey, label, _ := strings.Cut(o.Tail.Word, "|")
	body, ok := game.BodyRefByKey(bodyKey)
	if !ok || len(o.IDs) == 0 {
		return nil
	}
	return effects.ScheduleDelayedTrigger{
		At:                 game.StepUpkeep,
		ControllerTurnOnly: true,
		Label:              label,
		Cards:              o.IDs,
		Body:               body,
	}.Apply(ctx)
})
