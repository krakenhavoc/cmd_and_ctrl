package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deadly Designs — Enchantment {1}{B}:
//
//	"{2}: Put a plot counter on this enchantment. Any player may
//	 activate this ability.
//	 When there are five or more plot counters on this enchantment,
//	 sacrifice it. If you do, destroy up to two target creatures."
//
// ADR 0107 §1 (#1858), the first of the two cards the state-trigger row
// was filed with:
//
//   - The {2} row is any-player (ADR 0106 PR 6, CR 602.2): the activator
//     pays, and the counter is the effect.
//   - The rest is a CR 603.8 state trigger. It triggers once when the
//     fifth counter lands, whoever put it there, and a sixth counter
//     while it waits does not trigger it again. Its controller — the
//     enchantment's, not the activator's — chooses up to two target
//     creatures as it goes on the stack (CR 603.3d), and on resolution
//     sacrifices the enchantment and, only if that happened, destroys
//     the targets that are still legal (CR 608.2b).
//
// No purpose for the bot: no Purpose field says "brings a
// two-creature wipe closer", so a bot never pays for another player's
// counter.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "293c6f8f-9e6b-4267-a15d-1d5ce95fb5b4",
		Name:         "Deadly Designs",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     "{2}: Put a plot counter on this enchantment. Any player may activate this ability.",
			Cost:      game.AbilityCost{Mana: "{2}"},
			AnyPlayer: true,
			Effect:    putACounterOnThis("plot"),
		}},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisHasAtLeast("plot", 5, "Deadly Designs — sacrifice it and destroy up to two target creatures",
					SacrificeThisThen(func(ctx *Context) error { return destroyEachLegalTarget(ctx.Item, ctx) })),
				TargetCreature("up to two target creatures").WithCount(0, 2)),
		},
	})
}
