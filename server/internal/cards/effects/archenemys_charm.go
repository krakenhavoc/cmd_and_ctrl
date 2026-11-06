package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Archenemy's Charm — Instant {B}{B}{B} (Edge of Eternities, #4461):
//
//	"Choose one —
//	 • Exile target creature or planeswalker.
//	 • Return one or two target creature and/or planeswalker cards from
//	   your graveyard to your hand.
//	 • Put two +1/+1 counters on target creature you control. It gains
//	   lifelink until end of turn."
//
// Three bullets, each with its own clause (ModeDoing, ADR 0065 §3). The
// first is Dominion Assault Force's exile bullet (ExileTheModesTarget).
// The second is one clause with a count of one to two over creature and
// planeswalker cards you own in your graveyard; each pick still there
// as the Charm resolves goes to hand, and one that left in response is
// skipped (CR 608.2b). The third puts the counters on first and then
// grants lifelink to the same creature (Beetle, Legacy Criminal's
// shape), both skipped if it is no longer a legal target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "edcb0248-e7c9-4df9-8dba-982048654fd2",
		Name:         "Archenemy's Charm",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeWithPurpose(ModeDoing("Exile target creature or planeswalker.",
				TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
				ExileTheModesTarget), game.Purpose{Answers: game.AnswerRemove}),
			ModeWithPurpose(ModeDoing("Return one or two target creature and/or planeswalker cards from your graveyard to your hand.",
				TargetCardInGraveyard("one or two target creature and/or planeswalker cards from your graveyard",
					YouOwn(), Or(Creature(), Planeswalker())).WithCount(1, 2),
				eoeABCharmReturn), game.Purpose{Answers: game.AnswerValue}),
			ModeWithPurpose(ModeDoing("Put two +1/+1 counters on target creature you control. It gains lifelink until end of turn.",
				TargetCreature("target creature you control", YouControl()),
				eoeABCharmCounters), game.Purpose{Answers: game.AnswerPump | game.AnswerCombatGrant}),
		),
	})
}

// eoeABCharmReturn returns each still-legal graveyard pick of the bullet
// to its owner's hand.
func eoeABCharmReturn(_ *game.StackItem, ctx *Context, occ int) error {
	for _, t := range ctx.ModeTargets(occ) {
		if t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
			continue
		}
		if err := (ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// eoeABCharmCounters puts two +1/+1 counters on the bullet's creature and
// gives it lifelink until end of turn.
func eoeABCharmCounters(_ *game.StackItem, ctx *Context, occ int) error {
	t, ok := ModeTarget(ctx, occ)
	if !ok || t.Kind != game.TargetCard {
		return nil
	}
	if err := (AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 2}).Apply(ctx); err != nil {
		return err
	}
	return GrantKeywordUntilEOT{
		Target:   t.ID,
		Keywords: []string{"lifelink"},
		Label:    "Archenemy's Charm — lifelink",
	}.Apply(ctx)
}
