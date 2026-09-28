package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Inscription of Ruin — {2}{B} Sorcery:
//
//	"Kicker {2}{B}{B}
//	 Choose one. If this spell was kicked, choose any number instead.
//	 • Target opponent discards two cards.
//	 • Return target creature card with mana value 2 or less from your
//	   graveyard to the battlefield.
//	 • Destroy target creature with mana value 3 or less."
//
// The proof card for #1655's announcement-driven mode count: the
// count reads the kicker the caster announces WITH the modes (CR
// 601.2b), so an unkicked cast may take exactly one bullet and a
// kicked one any number of them — but still at least one.
// AnyNumberIf(WasKicked) says it; the cast gate, the hand card's
// `modes.if_optional_paid` and the bot's enumerator all read the same
// game.ModeBoundsForEffect.
//
// Each bullet reads its own target group (ModeTarget), in printed
// order (CR 608.2c). The discard is the target opponent's choice, as
// Mind Rot's is. No simplification.
func init() {
	Register(Spec{
		OracleID:     "760e8561-4ec6-4594-ba5d-f79cb9f25fd0",
		Name:         "Inscription of Ruin",
		Completeness: CompletenessFull,
		OptionalCosts: []game.AdditionalCost{
			Kicker("{2}{B}{B}"),
		},
		Modes: ChooseOne(
			ModeDoing("Target opponent discards two cards.",
				TargetPlayer("target opponent", Opponent()),
				inscriptionOfRuinDiscard),
			ModeDoing("Return target creature card with mana value 2 or less from your graveyard to the battlefield.",
				TargetCardInGraveyard("target creature card with mana value 2 or less from your graveyard",
					YouOwn(), Creature(), ManaValueLE(2)),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield}.Apply(ctx)
				}),
			ModeDoing("Destroy target creature with mana value 3 or less.",
				TargetCreature("target creature with mana value 3 or less", ManaValueLE(3)),
				DestroyTheModesTarget),
		).AnyNumberIf(WasKicked),
	})
}

// inscriptionOfRuinDiscard is the first bullet. Its own function, not
// a closure beside the other two, so nothing printed after it reads a
// value from before the prompted discard: the other bullets touch
// your graveyard and the battlefield, and neither depends on which
// cards the opponent picks.
func inscriptionOfRuinDiscard(item *game.StackItem, ctx *Context, occ int) error {
	t, ok := ModeTarget(ctx, occ)
	if !ok || t.Kind != game.TargetPlayer {
		return nil
	}
	ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
		Player:   t.ID,
		Source:   item.SourceCardID,
		N:        2,
		Question: "Inscription of Ruin — discard two cards",
	})
	return nil
}
