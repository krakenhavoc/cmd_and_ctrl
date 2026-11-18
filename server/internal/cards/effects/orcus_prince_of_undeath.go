package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Orcus, Prince of Undeath — Legendary Creature — Demon {X}{2}{B}{R},
// 5/3 (Adventures in the Forgotten Realms, #5898):
//
//	"Flying, trample
//	 When Orcus enters, choose one —
//	 • Each other creature gets -X/-X until end of turn. You lose X life.
//	 • Return up to X target creature cards with total mana value X or
//	   less from your graveyard to the battlefield. They gain haste until
//	   end of turn."
//
// X in the enters trigger is the X Orcus was cast with (CR 107.3m),
// stamped on the trigger by the engine (XFromCast); an Orcus put onto
// the battlefield without being cast has X = 0. The second bullet's
// clause is counted "up to X" and capped at a total mana value of X, both
// read from that X as the trigger is put on the stack (#5898,
// game.bindTriggerModeStepsCastX and TotalManaValueAtMostX); at X = 0 it
// takes no target. As it resolves the picks still legal are re-summed
// (CR 608.2b): when they no longer fit, none is returned. The cards that
// do come back enter together and gain haste.
//
// The first bullet shrinks every creature but Orcus, read as it
// resolves, then you lose X life.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9d02b5e1-bd29-45ad-b11d-f2cd21c48979",
		Name:            "Orcus, Prince of Undeath",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		XMatters:        true,
		Triggered:       []game.TriggeredAbility{orcusEnters()},
	})
}

// orcusEnters is the modal enters trigger.
func orcusEnters() game.TriggeredAbility {
	returns := TargetCardInGraveyard("up to X target creature cards with total mana value X or less from your graveyard",
		Creature(), YouOwn()).WithCount(0, 0).WithTotal(TotalManaValueAtMostX())
	returns.CountFromX = true
	returns.UpToX = true

	enters := WhenThisEnters("Orcus, Prince of Undeath — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	enters.XFromCast = true
	enters.Modes = ChooseOne(
		ModeWithPurpose(ModeDoing("Each other creature gets -X/-X until end of turn. You lose X life.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				x := ctx.X()
				if x <= 0 {
					return nil
				}
				if err := (BoostUntilEOT{
					Match:     And(Creature(), OtherThan(item.SourceCardID)),
					Power:     -x,
					Toughness: -x,
					Label:     "Orcus, Prince of Undeath — -X/-X",
				}).Apply(ctx); err != nil {
					return err
				}
				return GainLife{Player: item.Controller, Amount: -x}.Apply(ctx)
			}),
			game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, AmountIsX: true, Partial: true}}),
		ModeWithPurpose(ModeDoing("Return up to X target creature cards with total mana value X or less from your graveyard to the battlefield. They gain haste until end of turn.", returns,
			func(_ *game.StackItem, ctx *Context, occ int) error {
				var ids []uuid.UUID
				for _, t := range ctx.ModeTargets(occ) {
					if t.Kind == game.TargetCard && ctx.IsTargetLegal(t) {
						ids = append(ids, t.ID)
					}
				}
				if len(ids) == 0 {
					return nil
				}
				return ReturnFromGraveyardTogether{Targets: ids, ThenTail: orcusHasteTail}.Apply(ctx)
			}),
			ForTargets(Reanimates(0))),
	)
	return enters
}

// orcusHasteTail is "they gain haste until end of turn" over the cards
// the return settled with, so a card that failed to come back gains
// nothing. An on-disk key: never renamed, never reused.
var orcusHasteTail = TogetherTail("orcus-prince-of-undeath/gain-haste-until-eot", func(ctx *Context, entered []uuid.UUID, _ int) error {
	for _, id := range entered {
		if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"}, Label: "Orcus, Prince of Undeath — haste"}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
