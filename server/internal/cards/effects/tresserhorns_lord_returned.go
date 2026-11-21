package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tresserhorn's Lord, Returned — Legendary Creature — Zombie Noble
// {2}{U}{B}{R}, 10/4 (Mystery Booster Commander Edition):
//
//	"Trample, indestructible
//	 When Tresserhorn's Lord enters, sacrifice him unless you sacrifice
//	 three creatures, pay 3 life, and have target opponent draw three
//	 cards.
//	 Whenever you sacrifice a creature, draw a card."
//
// The enter trigger targets an opponent as it goes on the stack, so it is
// removed if that opponent is gone (CR 608.2b) and the Lord stays. As it
// resolves, the controller chooses between sacrificing the Lord and the
// whole payment; the payment is offered only when they control three
// creatures they can sacrifice (the Lord may be one of them) and have 3
// life to pay (CR 119.4), and is made together (sacrifice three, pay 3
// life, the opponent draws three). The Lord's own sacrifice and the three
// sacrifices each trigger its draw ability. A Lord that has already left,
// or come back as a new object, is not sacrificed (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "11cc638c-b1c5-4b73-8eac-00cac6ae32b6",
		Name:            "Tresserhorn's Lord, Returned",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample", "indestructible"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Tresserhorn's Lord, Returned — sacrifice him unless you sacrifice three creatures, pay 3 life, and have target opponent draw three cards",
				tresserhornETB), TargetPlayer("target opponent", Opponent())),
			On(game.EventSacrifice, jmpB1YouSacrificedACreature, "Tresserhorn's Lord, Returned — draw a card",
				Do(DrawCards{N: 1})),
		},
	})
}

const (
	tresserhornSacrificeLabel = "Sacrifice Tresserhorn's Lord"
	tresserhornPayLabel       = "Sacrifice three creatures, pay 3 life, and have the target opponent draw three cards"
)

// tresserhornSacrificeable lists the creatures the controller may sacrifice.
func tresserhornSacrificeable(g *game.Game, controller uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, id := range permanentsControlledByMatching(g, controller, Creature()) {
		if !g.CantBeSacrificedForEffect(id) {
			out = append(out, id)
		}
	}
	return out
}

func tresserhornETB(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if !onBattlefield(g, item.SourceCardID) || sourceIsNewObject(g, item) {
		return nil
	}
	p := g.PlayerByIDForEffect(item.Controller)
	if p == nil || p.Life < 3 || len(tresserhornSacrificeable(g, item.Controller)) < 3 || len(ctx.LegalTargets()) == 0 {
		return tresserhornSacrificeSelf(ctx)
	}
	return PickOption{
		Question: "Tresserhorn's Lord, Returned — sacrifice him, or sacrifice three creatures, pay 3 life, and have target opponent draw three cards",
		Options:  []game.ChoiceOption{{Label: tresserhornSacrificeLabel}, {Label: tresserhornPayLabel}},
		ThenKey:  tresserhornChose,
		WithItem: true,
	}.Apply(ctx)
}

func tresserhornSacrificeSelf(ctx *Context) error {
	ctx.Game.SacrificeAllForEffect(ctx.Source(), []uuid.UUID{ctx.Source()})
	return nil
}

// tresserhornChose is the keyed answer to the unless-cost choice. An
// on-disk key: never renamed, never reused.
var tresserhornChose = OptionPickThen("option-pick/tresserhorns-lord-returned-choice", func(ctx *Context, r game.OptionPicked) error {
	if r.Option == nil || r.Option.Label != tresserhornPayLabel {
		return tresserhornSacrificeSelf(ctx)
	}
	var opp uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			opp = t.ID
		}
	}
	cands := tresserhornSacrificeable(ctx.Game, ctx.Controller())
	p := ctx.Game.PlayerByIDForEffect(ctx.Controller())
	if opp == uuid.Nil || p == nil || p.Life < 3 || len(cands) < 3 {
		return tresserhornSacrificeSelf(ctx)
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    ctx.Controller(),
		FromPlayer: ctx.Controller(),
		Source:     ctx.Source(),
		Question:   "Tresserhorn's Lord, Returned — sacrifice three creatures",
		Cards:      cands,
		Min:        3,
		Max:        3,
		ThenKey:    tresserhornSacrificed,
		LiveZone:   game.ZoneBattlefield,
		Carry:      []uuid.UUID{opp},
	})
	return nil
})

// tresserhornSacrificed pays the rest of the cost: the three creatures,
// 3 life, and the opponent's three cards. An on-disk key: never renamed,
// never reused.
var tresserhornSacrificed = game.RegisterChooseCardsThen("choose-cards/tresserhorns-lord-returned-sacrifice", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 || len(r.Carry) == 0 {
		return nil
	}
	ctx := NewContext(g, &game.StackItem{
		Kind:         game.StackItemSpell,
		Controller:   r.Chooser,
		Owner:        r.Chooser,
		SourceCardID: r.Source,
	})
	g.SacrificeAllForEffect(r.Source, r.Picked)
	if err := g.PayLifeForEffect(r.Source, r.Chooser, 3); err != nil {
		return err
	}
	return DrawCards{Player: r.Carry[0], N: 3}.Apply(ctx)
})
