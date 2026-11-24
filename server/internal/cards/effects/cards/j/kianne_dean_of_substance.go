package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kianne, Dean of Substance // Imbraham, Dean of Theory — a modal
// double-faced card (Strixhaven: School of Mages, #5502).
//
//	Kianne, Dean of Substance   {2}{G}   2/2
//	Legendary Creature — Elf Druid
//	  {T}: Exile the top card of your library. If it's a land card, put it
//	  into your hand. Otherwise, put a study counter on it.
//	  {4}{G}: Create a 0/0 green and blue Fractal creature token. Put a
//	  +1/+1 counter on it for each different mana value among nonland cards
//	  you own in exile with study counters on them.
//
//	Imbraham, Dean of Theory    {2}{U}{U}   3/3
//	Legendary Creature — Bird Wizard
//	  Flying
//	  {X}{U}{U}, {T}: Exile the top X cards of your library and put a study
//	  counter on each of them. Then you may put a card you own in exile with
//	  a study counter on it into your hand.
//
// The exile is the plain face-up exile (Spark of Creativity's
// ExileTopKeyedForEffect) and what happens to the cards runs in its keyed
// tail, so a replacement prompt on the way is a restore point. A land goes to
// the hand from exile; any other card stays and takes a study counter (Kianne
// only; Imbraham counters every card it exiled, lands included). The counter
// kind "study" is the one Lattice Library already uses.
//
// The Fractal counts the distinct mana values (CR 202.3) of the nonland cards
// in exile that the controller OWNS with a study counter, whoever exiled them
// and whichever of the two faces put them there, read as the ability resolves
// (CR 608.2h); a card with no mana cost has value 0 and counts as one. A
// face-down card has no characteristics and is not counted. With none, the
// Fractal is a 0/0 and dies to CR 704.5f.
//
// Imbraham's "you may" is an optional pick of zero or one card from every card
// the player owns in exile with a study counter on it, the ones just exiled
// included. X = 0 exiles nothing and still offers the pick, so the ability is
// a real move at X = 0 and the card takes an xMattersAllowlist entry.
//
// No simplification.
func init() {
	const oracle = "9eab65ee-e8be-4fa6-9634-fb03582bd745"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Kianne, Dean of Substance",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Label:   "{T}: Exile the top card of your library. If it's a land card, put it into your hand. Otherwise, put a study counter on it.",
				Cost:    effects.TapCost(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return g.ExileTopKeyedForEffect(item.Controller, item.Controller, 1, ctx.ItemTail(kianneExiled))
				},
			},
			{
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Label:   "{4}{G}: Create a 0/0 green and blue Fractal creature token. Put a +1/+1 counter on it for each different mana value among nonland cards you own in exile with study counters on them.",
				Cost:    effects.ManaCost("{4}{G}"),
				Effect: effects.Do(effects.SocB8FractalWithCountersSteps("kianne/fractal", func(ctx *effects.Context) int {
					return effects.OldB41DifferentManaValuesAmongStudyNonlands(ctx.Game, ctx.Controller())
				})...),
			},
		},
	})
	effects.Register(effects.Spec{
		OracleID:        oracle + "#1",
		Name:            "Imbraham, Dean of Theory",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Label:   "{X}{U}{U}, {T}: Exile the top X cards of your library and put a study counter on each of them. Then you may put a card you own in exile with a study counter on it into your hand.",
			Cost:    effects.Plus(effects.ManaCost("{X}{U}{U}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				tail := ctx.ItemTail(imbrahamExiled)
				if ctx.X() <= 0 {
					return tail.Run(g, nil)
				}
				return g.ExileTopKeyedForEffect(item.Controller, item.Controller, ctx.X(), tail)
			},
		}},
	})
}

// kianneExiled is Kianne's tap ability once the card is exiled: a land goes to
// its owner's hand, anything else takes a study counter. An on-disk key.
var kianneExiled = effects.CardTail("old-b041/kianne-exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok {
			continue
		}
		if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
			continue
		}
		if c.IsLand() {
			if _, err := ctx.Game.ReturnFromAnyZoneToHandForEffect(ctx.Controller(), ctx.Source(), id); err != nil {
				return err
			}
			continue
		}
		if err := ctx.Game.AddCounterForEffect(id, effects.OldB41StudyCounter, 1); err != nil {
			return err
		}
	}
	return nil
})

// imbrahamExiled is Imbraham's ability once the top X cards are exiled: a study
// counter on each, then the optional pick of one card with a study counter.
// An on-disk key.
var imbrahamExiled = effects.CardTail("old-b041/imbraham-exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	g := ctx.Game
	for _, id := range o.IDs {
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
			continue
		}
		if err := g.AddCounterForEffect(id, effects.OldB41StudyCounter, 1); err != nil {
			return err
		}
	}
	var cands []uuid.UUID
	for _, c := range effects.OldB41StudyCardsOwnedInExile(g, ctx.Controller()) {
		cands = append(cands, c.InstanceID)
	}
	if len(cands) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Imbraham, Dean of Theory — you may put a card you own in exile with a study counter on it into your hand",
		Cards:    cands,
		Min:      0,
		Max:      1,
		LiveZone: game.ZoneExile,
		ThenKey:  imbrahamPicked,
	})
	return nil
})

// imbrahamPicked puts the chosen card into its owner's hand. An on-disk key.
var imbrahamPicked = game.RegisterChooseCardsThen("choose-cards/old-b041-imbraham-to-hand", func(g *game.Game, r game.CardsChosen) error {
	for _, id := range r.Picked {
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
			continue
		}
		if _, err := g.ReturnFromAnyZoneToHandForEffect(r.Chooser, r.Source, id); err != nil {
			return err
		}
	}
	return nil
})
