package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Deadlock Trap — Artifact {3}:
//
//	"This artifact enters tapped.
//	 When this artifact enters, you get {E}{E} (two energy counters).
//	 {T}, Pay {E}: Tap target creature or planeswalker. Its activated
//	 abilities can't be activated this turn."
//
// ADR 0129 §2 (#1995): "Pay {E}" is the energy cost component. "Its
// activated abilities can't be activated this turn" is one record pinned
// to the permanent until the cleanup step, as Dovin Baan's +1 does it:
// CantActivate and CantActivateMana, because a mana ability is an
// activated ability (CR 605.1a) and the card does not spare them. A
// permanent that leaves the battlefield is a new object the record no
// longer follows (CR 400.7).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "726c911d-6543-4400-a8fa-b8a5c9f0c15d",
		Name:         "Deadlock Trap",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Energy: 2},
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouGetEnergy("Deadlock Trap", 2),
		},
		Activated: []ActivatedAbility{{
			Label:   "{T}, Pay {E}: Tap target creature or planeswalker. Its activated abilities can't be activated this turn.",
			Cost:    Plus(TapCost(), PayEnergy(1)),
			Targets: TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
			Effect:  deadlockTrapLock,
		}},
	})
}

// deadlockTrapLock taps the still-legal target and locks its activated
// abilities until end of turn (CR 608.2b: nothing if it left).
func deadlockTrapLock(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target := FirstLegalBattlefieldTarget(ctx)
	if target == uuid.Nil {
		return nil
	}
	if err := (TapTarget{Target: target}).Apply(ctx); err != nil {
		return err
	}
	return ScopedEffectFor{
		Target:   target,
		Mods:     []game.Mod{game.AddRestrictionsMod(game.CantActivate | game.CantActivateMana)},
		Duration: DurationUntilEndOfTurn(ctx),
		Label:    "Deadlock Trap — its activated abilities can't be activated this turn",
	}.Apply(ctx)
}
