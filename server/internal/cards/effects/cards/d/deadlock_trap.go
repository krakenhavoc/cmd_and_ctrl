package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "726c911d-6543-4400-a8fa-b8a5c9f0c15d",
		Name:         "Deadlock Trap",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Energy: 2},
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersYouGetEnergy("Deadlock Trap", 2),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Pay {E}: Tap target creature or planeswalker. Its activated abilities can't be activated this turn.",
			Cost:    effects.Plus(effects.TapCost(), effects.PayEnergy(1)),
			Targets: effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
			Effect:  deadlockTrapLock,
		}},
	})
}

// deadlockTrapLock taps the still-legal target and locks its activated
// abilities until end of turn (CR 608.2b: nothing if it left).
func deadlockTrapLock(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	target := effects.FirstLegalBattlefieldTarget(ctx)
	if target == uuid.Nil {
		return nil
	}
	if err := (effects.TapTarget{Target: target}).Apply(ctx); err != nil {
		return err
	}
	return effects.ScopedEffectFor{
		Target:   target,
		Mods:     []game.Mod{game.AddRestrictionsMod(game.CantActivate | game.CantActivateMana)},
		Duration: effects.DurationUntilEndOfTurn(ctx),
		Label:    "Deadlock Trap — its activated abilities can't be activated this turn",
	}.Apply(ctx)
}
