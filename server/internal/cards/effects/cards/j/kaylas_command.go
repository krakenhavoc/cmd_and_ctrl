package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kayla's Command — Sorcery {1}{W}{W} (The Brothers' War, #5482):
//
//	"Choose two —
//	 • Create a 2/2 colorless Construct artifact creature token.
//	 • Put a +1/+1 counter on a creature you control. It gains double
//	   strike until end of turn.
//	 • Search your library for a basic Plains card, reveal it, put it
//	   into your hand, then shuffle.
//	 • You gain 2 life and scry 2."
//
// Brigid's Command's "choose two" template (ChooseN 2..2): each chosen
// bullet has its own target group, checked again as the spell resolves
// (CR 608.2b), and the bullets run in printed order (CR 608.2c). The
// counter bullet's "it" is the one target: the counter goes on, then the
// creature gains double strike until end of turn, both only while it is
// still a legal target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3d937cd9-b1e1-406d-9e29-3b9d3ceea3be",
		Name:         "Kayla's Command",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseN("Choose two", 2, 2,
			effects.ModeWithPurpose(effects.ModeDoing("Create a 2/2 colorless Construct artifact creature token.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.CreateToken{Template: effects.TokenCard("2/2 colorless Construct artifact"), N: 1}.Apply(ctx)
				}), game.Purpose{Tokens: 1, Answers: game.AnswerMakesBlocker}),
			effects.ModeWithPurpose(effects.ModeDoing("Put a +1/+1 counter on a creature you control. It gains double strike until end of turn.",
				effects.TargetCreature("target creature you control", effects.YouControl()),
				kaylasCommandCounterAndDoubleStrike),
				effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "double strike"))),
			effects.ModeWithPurpose(effects.ModeDoing("Search your library for a basic Plains card, reveal it, put it into your hand, then shuffle.", nil,
				func(item *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.SearchLibrary{
						Player:    item.Controller,
						Predicate: effects.IsBasicLandWithSubtype("plains"),
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Reason:    "Kayla's Command — a basic Plains card",
						Source:    item.SourceCardID,
					}.Apply(ctx)
				}), game.Purpose{Tutors: 1}),
			effects.ModeWithPurpose(effects.ModeDoing("You gain 2 life and scry 2.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					if err := (effects.GainLife{Amount: 2}).Apply(ctx); err != nil {
						return err
					}
					return effects.Scry{N: 2}.Apply(ctx)
				}), game.Purpose{Scry: 2}),
		),
	})
}

// kaylasCommandCounterAndDoubleStrike is the second bullet's body.
func kaylasCommandCounterAndDoubleStrike(_ *game.StackItem, ctx *effects.Context, occ int) error {
	t, ok := effects.ModeTarget(ctx, occ)
	if !ok || t.Kind != game.TargetCard {
		return nil
	}
	if err := (effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
		return err
	}
	return effects.GrantKeywordUntilEOT{
		Target:   t.ID,
		Keywords: []string{"double strike"},
		Label:    "Kayla's Command — gains double strike",
	}.Apply(ctx)
}
