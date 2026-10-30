package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Heroes' Hangout — {R} Sorcery:
//
//	"Choose one —
//	 • Date Night — Exile the top two cards of your library. Choose one of
//	   them. Until the end of your next turn, you may play that card.
//	 • Patrol Night — One or two target creatures each get +1/+0 and gain
//	   first strike until end of turn."
//
// Date Night exiles both cards (face up), then asks the controller which one
// it chooses; only the chosen card gets the permission to be played (a land
// is played, a spell cast) until the end of the controller's NEXT turn (ADR
// 0063), and the other stays exiled with nothing. The chain is data: the
// exile's tail asks the pick (a card still in exile) and the pick's
// continuation grants the window. Patrol Night's two statements are applied
// to every target still legal at resolution.
//
// No simplification. Marvel Super Heroes (#4455).
func init() {
	// Patrol Night helps the creatures it targets: a +1/+0 and first strike
	// until end of turn (ADR 0126).
	patrol := ModeDoing("Patrol Night — One or two target creatures each get +1/+0 and gain first strike until end of turn.",
		TargetCreature("one or two target creatures").WithCount(1, 2),
		func(_ *game.StackItem, ctx *Context, occ int) error {
			for _, t := range ctx.ModeTargets(occ) {
				if t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
					continue
				}
				if err := (BoostUntilEOT{Target: t.ID, Power: 1, Label: "Heroes' Hangout — Patrol Night: +1/+0"}).Apply(ctx); err != nil {
					return err
				}
				if err := (GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"first strike"},
					Label: "Heroes' Hangout — Patrol Night: first strike"}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		})
	patrol.Purpose = ForTargets(BenefitUntilEOT(0, 1, 0, "first strike"))
	Register(Spec{
		OracleID:     "43306a78-85b5-4df4-a018-75dd75aa8997",
		Name:         "Heroes' Hangout",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Date Night — Exile the top two cards of your library. Choose one of them. Until the end of your next turn, you may play that card.",
				nil, func(item *game.StackItem, ctx *Context, _ int) error {
					ids := ctx.Game.LookAtTopOfLibraryForEffect(item.Controller, 2)
					if len(ids) == 0 {
						return nil
					}
					return ctx.Game.ExileCardsKeyedForEffect(ids, ctx.ItemTail(msc18DateNightChoose))
				}),
			patrol,
		),
	})
}

// msc18DateNightChoose is "choose one of them" over the cards that reached
// exile: the controller picks one, and msc18DateNightGrant gives it the
// window. An on-disk key.
var msc18DateNightChoose = CardTail("then/heroes-hangout-choose", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Heroes' Hangout — choose one of them (you may play it until the end of your next turn)",
		Cards:    o.IDs,
		Min:      1,
		Max:      1,
		LiveZone: game.ZoneExile,
		ThenKey:  msc18DateNightGrant,
	})
	return nil
})

// msc18DateNightGrant gives the chosen card its window: the chooser may play
// it until the end of their next turn. An on-disk key.
var msc18DateNightGrant = game.RegisterChooseCardsThen("choose-cards/heroes-hangout-play", func(g *game.Game, r game.CardsChosen) error {
	for _, id := range r.Picked {
		g.GrantCastPermissionOverCardForEffect(id, game.CastPermission{
			Player:   r.Chooser,
			Zone:     game.ZoneExile,
			Duration: g.UntilEndOfYourNextTurnDuration(r.Chooser),
			Label:    "Heroes' Hangout — play it until the end of your next turn",
		})
	}
	return nil
})
