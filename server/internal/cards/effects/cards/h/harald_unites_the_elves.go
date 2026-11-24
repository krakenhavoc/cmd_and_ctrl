package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Harald Unites the Elves — Enchantment — Saga {2}{B}{G} (Kaldheim Promos,
// #5505):
//
//	"I — Mill three cards. You may put an Elf or Tyvar card from your
//	 graveyard onto the battlefield.
//	 II — Put a +1/+1 counter on each Elf you control.
//	 III — Whenever an Elf you control attacks this turn, target creature an
//	 opponent controls gets -1/-1 until end of turn."
//
// Chapter I mills, then offers every Elf or Tyvar card in your graveyard (not
// only the milled ones, and a changeling is an Elf); the pick may be none.
// Chapter II counts the Elves as it resolves. Chapter III is a repeating
// delayed trigger (CR 603.7b) on any Elf its controller attacks with this
// turn, one trigger per attacking Elf (CR 603.2c), the target chosen as each
// instance goes on the stack and rechecked as it resolves.
//
// No simplification.
func init() {
	const name = "Harald Unites the Elves"
	effects.Register(effects.Spec{
		OracleID:     "d0cc7e79-69dc-4f9d-9eed-1848f8ddb1be",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel(name, 1, "mill three cards; you may put an Elf or Tyvar card from your graveyard onto the battlefield"),
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.MillToZone{N: 3, ThenTail: ctx.ItemTail(haraldUnitesTheElvesMilled)}.Apply(ctx)
				}),
			effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, "put a +1/+1 counter on each Elf you control"),
				effects.AvB30aPutCounterOnEachOfSubtypeYouControl("Elf")),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, "whenever an Elf you control attacks this turn, target creature an opponent controls gets -1/-1 until end of turn"),
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.DelayedOnEvent{
						Label:      name + " — target creature an opponent controls gets -1/-1 until end of turn",
						On:         []game.EventKind{game.EventAttack},
						Condition:  effects.GapB41AssassinYouControlAttacks,
						CondParams: game.EffectParams{Name: ctx.TextSubtype("Elf")},
						Body:       haraldUnitesTheElvesShrink,
						Repeats:    true,
					}.Apply(ctx)
				}),
		},
	})
}

// haraldUnitesTheElvesMilled offers the Elf and Tyvar cards in the graveyard
// once the mill is done. An on-disk key.
var haraldUnitesTheElvesMilled = effects.CardTail("harald-unites-the-elves/after-mill", func(ctx *effects.Context, _ game.TailOutcome) error {
	p := ctx.Game.PlayerByIDForEffect(ctx.Controller())
	if p == nil {
		return nil
	}
	elf := ctx.TextSubtype("Elf")
	var ids []uuid.UUID
	for _, c := range p.Graveyard.Cards {
		if c.HasSubtype(elf) || c.HasSubtype("Tyvar") {
			ids = append(ids, c.InstanceID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Harald Unites the Elves — you may put an Elf or Tyvar card from your graveyard onto the battlefield",
		Cards:    ids,
		Min:      0,
		Max:      1,
		LiveZone: game.ZoneGraveyard,
		ThenKey:  haraldUnitesTheElvesPick,
		Asked:    game.NewAskedItem(ctx.Item),
	})
	return nil
})

// haraldUnitesTheElvesPick puts the chosen card onto the battlefield. An
// on-disk key.
var haraldUnitesTheElvesPick = effects.AskedPick("choose-cards/harald-unites-the-elves-reanimate", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return effects.ReturnFromGraveyard{Target: r.Picked[0], Dest: game.ZoneBattlefield, Controller: ctx.Controller()}.Apply(ctx)
})

// haraldUnitesTheElvesShrink is chapter III's delayed trigger. An on-disk key.
var haraldUnitesTheElvesShrink = game.TargetedDelayedBody("harald-unites-the-elves/shrink",
	effects.SimpleBody(func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind == game.TargetCard {
				return effects.UntilEndOfTurn(ctx, t.ID, nil, "Harald Unites the Elves — -1/-1", game.ModifyPTMod(-1, -1))
			}
		}
		return nil
	}),
	effects.ConstTargets(func() *game.TargetSpec {
		return effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())
	}))
