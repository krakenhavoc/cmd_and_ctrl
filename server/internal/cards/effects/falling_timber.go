package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Falling Timber — Instant {2}{G}:
//
//	"Kicker—Sacrifice a land. (You may sacrifice a land in addition to any other costs as you cast this spell.)
//	 Prevent all combat damage target creature would deal this turn. If this spell was kicked, prevent all combat damage another target creature would deal this turn."
//
// ADR 0108 §7 (#1904, Delivery PR 7b): kicked, the spell's target
// statement is two clauses, the second Distinct ("another"), chosen only
// when the kicker is paid (its ruling; WhenPaid, CR 601.2b–c). Each
// target gets a combat-damage shield with itself as the source, pinned as
// the spell resolves (CR 400.7) — one record per source, which nothing a
// player can see tells from one effect, since an event has one source.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "4ea19457-97ef-4ac8-b67c-41be1109ca73",
		Name:         "Falling Timber",
		Completeness: CompletenessFull,
		OptionalCosts: []game.AdditionalCost{WhenPaid(KickerSacrifice("a land", Land()),
			Clauses(
				TargetCreature("target creature"),
				Distinct(TargetCreature("another target creature")),
			))},
		Targets:   TargetCreature("target creature"),
		OnResolve: shieldAgainstTargetsSpell(true),
	})
}
