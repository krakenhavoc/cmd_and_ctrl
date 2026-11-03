package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Unbound Flourishing — Enchantment {2}{G} (Secrets of Strixhaven Commander,
// #4834):
//
//	"Whenever you cast a permanent spell with a mana cost that contains {X},
//	 double the value of X.
//	 Whenever you cast an instant or sorcery spell or activate an ability, if
//	 that spell's mana cost or that ability's activation cost contains {X},
//	 copy that spell or ability. You may choose new targets for the copy."
//
// Two triggered abilities.
//
// The first watches the controller's casts of permanent spells (not an
// instant or sorcery) whose own printed mana cost has an {X}: an {X} in an
// additional or alternative cost does not count (the rulings). It doubles the
// X announced for the spell while it is on the stack (DoubleSpellXForEffect,
// CR 107.3a), so its mana value on the stack (CR 202.3e), the counters it
// enters with and the X its enters-the-battlefield abilities read (CR 107.3m)
// all follow, and the caster pays nothing more. X of 0 stays 0. A spell
// already countered has no value left to double. When it triggers together
// with another cast trigger that reads X, the controller orders the two, so
// they pick whether to double before or after.
//
// The second is one ability with two trigger events (OnAny): the
// controller casts an instant or sorcery spell, or activates an ability. Its
// CR 603.4 intervening "if" is checked as it triggers and again as it
// resolves (InterveningIf): the spell's printed mana cost contains {X}, or the
// activation announced an {X} mana symbol in its cost (Event.XCost). A -X
// loyalty cost, or an X that is only in the ability's text, is not that. A
// mana ability uses no stack and is never copied. The copy is created on the
// stack, not cast or activated (CR 707.10), with the same X, modes and
// additional-cost facts as the original, and may take new targets (CR
// 707.10c). A copy is made even if the original has been countered by the
// time the trigger resolves (last known information, CR 608.2h), and it
// resolves first. Two Flourishings give two copies.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4db030f5-d7c4-4aa0-ae2e-12943e2683d7",
		Name:         "Unbound Flourishing",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, socB14YouCastAPermanentSpellWithX,
				"Unbound Flourishing — double the value of X",
				Do(Instruction(socB14DoubleTheSpellsX))),
			InterveningIf(socB14TheCostContainsX,
				OnAny([]game.EventKind{game.EventCast, game.EventActivateAbility}, socB14YouCastOrActivateAnXThing,
					"Unbound Flourishing — copy that spell or ability",
					Do(Instruction(socB14CopyThatSpellOrAbility)))),
		},
	})
}
