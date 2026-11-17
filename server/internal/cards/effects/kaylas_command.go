package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "3d937cd9-b1e1-406d-9e29-3b9d3ceea3be",
		Name:         "Kayla's Command",
		Completeness: CompletenessFull,
		Modes: ChooseN("Choose two", 2, 2,
			ModeWithPurpose(ModeDoing("Create a 2/2 colorless Construct artifact creature token.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return CreateToken{Template: TokenCard("2/2 colorless Construct artifact"), N: 1}.Apply(ctx)
				}), game.Purpose{Tokens: 1, Answers: game.AnswerMakesBlocker}),
			ModeWithPurpose(ModeDoing("Put a +1/+1 counter on a creature you control. It gains double strike until end of turn.",
				TargetCreature("target creature you control", YouControl()),
				kaylasCommandCounterAndDoubleStrike),
				ForTargets(BenefitUntilEOT(0, 0, 0, "double strike"))),
			ModeWithPurpose(ModeDoing("Search your library for a basic Plains card, reveal it, put it into your hand, then shuffle.", nil,
				func(item *game.StackItem, ctx *Context, _ int) error {
					return SearchLibrary{
						Player:    item.Controller,
						Predicate: IsBasicLandWithSubtype("plains"),
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Reason:    "Kayla's Command — a basic Plains card",
						Source:    item.SourceCardID,
					}.Apply(ctx)
				}), game.Purpose{Tutors: 1}),
			ModeWithPurpose(ModeDoing("You gain 2 life and scry 2.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					if err := (GainLife{Amount: 2}).Apply(ctx); err != nil {
						return err
					}
					return Scry{N: 2}.Apply(ctx)
				}), game.Purpose{Scry: 2}),
		),
	})
}

// kaylasCommandCounterAndDoubleStrike is the second bullet's body.
func kaylasCommandCounterAndDoubleStrike(_ *game.StackItem, ctx *Context, occ int) error {
	t, ok := ModeTarget(ctx, occ)
	if !ok || t.Kind != game.TargetCard {
		return nil
	}
	if err := (AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
		return err
	}
	return GrantKeywordUntilEOT{
		Target:   t.ID,
		Keywords: []string{"double strike"},
		Label:    "Kayla's Command — gains double strike",
	}.Apply(ctx)
}
