package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// designations.go — the card-facing half of ADR 0071. A DESIGNATION
// is a marker a permanent has on the battlefield that switches some
// of its own printed abilities on: a Class's level (CR 716.2), a
// Case being solved (CR 719.3), a station threshold's charge counters
// (CR 721.2), a Room's unlocked door (CR 709.5, not built).
//
// The engine side is server/internal/game/designations.go: one field
// (`ActiveWhen`) on every printed-ability slot, one predicate, and one
// place it is evaluated. What a card file writes is the gate and the
// constructors below.
//
//	Activated: []ActivatedAbility{
//	    LevelUp(2, ManaCost("{2}{U}")),
//	    LevelUp(3, ManaCost("{4}{U}")),
//	},
//	Triggered: []game.TriggeredAbility{
//	    AtLevel(2, On(game.EventClassLevel, …)),   // "When this Class becomes level 2, …"
//	    AtLevel(3, On(game.EventDrawCard, …)),     // the level-3 line
//	},
//
// The gate is declarative and says NOTHING about when it is checked.
// A card file never writes "if the Class is level 3" inside an
// AppliesTo or an Apply: the ability printed on the level-3 line does
// not exist below level 3, which is a stronger statement and the one
// CR 716.2a actually makes.

// --- gates -----------------------------------------------------------

// Level is the CR 716.2 gate: this ability exists while the Class is
// level n or greater. "Or greater" is the rule, not a convenience — a
// level-3 Class keeps everything its level-2 line printed.
func Level(n int) game.Designation { return game.ClassLevel(n) }

// Solved is the CR 719.3 gate: this ability exists while the Case is
// solved. A solved Case stays solved for as long as it is on the
// battlefield, so nothing that reads this has to worry about it going
// back off.
func Solved() game.Designation { return game.CaseSolved() }

// Harnessed is the CR 701.64 / 702.186b gate: this ability exists
// while the permanent is harnessed — "∞ — [Ability]" (#1321). Once
// harnessed a permanent stays harnessed for as long as it is on the
// battlefield (CR 701.64b), so nothing that reads this has to worry
// about it going back off.
func Harnessed() game.Designation { return game.Harnessed() }

// AtLevel stamps a CR 716.2 level gate onto a triggered ability built
// with any of the ordinary trigger constructors, so the level line
// reads as one thing:
//
//	AtLevel(3, On(game.EventDrawCard, ByYou, "…", effect))
func AtLevel(n int, t game.TriggeredAbility) game.TriggeredAbility {
	t.ActiveWhen = Level(n)
	return t
}

// WhenSolved stamps a CR 719.3 gate onto a triggered ability — the
// "Solved — Whenever you cast an instant or sorcery spell, draw a
// card" line.
func WhenSolved(t game.TriggeredAbility) game.TriggeredAbility {
	t.ActiveWhen = Solved()
	return t
}

// --- Class (CR 716) --------------------------------------------------

// LevelUp is the "[cost]: Level N" activated ability every Class
// prints once per level above the first (CR 716.2a).
//
// One constructor rather than a hand-written ability per card,
// because three rules ride along and a card file that wrote them out
// could get any of them wrong:
//
//	CR 716.2d  sorcery speed
//	CR 716.2e  "activate only if this Class is level N-1" — so the
//	           levels are climbed in order and none is skipped
//	CR 716.2b  the effect is "this Class's level becomes N", not
//	           "+1 level", so two copies of the ability on the stack
//	           cannot overshoot
//
// The cost is an ORDINARY AbilityCost, so everything that cost
// component grew since — counter costs (#958), X, life — works here
// with no arm of its own, and the level-up is announced, paid, put on
// the stack and responded to exactly like any other activation.
//
// It carries no ActiveWhen. "{3}{U}: Level 2" is printed on the Class
// from the moment it enters and is always THERE; what limits it is
// the CR 602.1b activation instruction, which is a Condition and
// which the client already greys with condition_unmet. A gate would
// make the line vanish instead, which is neither what the card looks
// like nor what the rule says.
func LevelUp(level int, cost game.AbilityCost) ActivatedAbility {
	label := "Level " + strconv.Itoa(level)
	if cost.Mana != "" {
		label = cost.Mana + ": " + label
	}
	return ActivatedAbility{
		Label:        label,
		Cost:         cost,
		SorcerySpeed: true,
		Condition: func(g *game.Game, _ uuid.UUID, source uuid.UUID) bool {
			return g.ClassLevelFor(source) == level-1
		},
		Effect: func(g *game.Game, item *game.StackItem) error {
			return g.SetClassLevelForEffect(item.SourceCardID, level)
		},
	}
}

// BecomesLevel is the "When this Class becomes level N, …" trigger
// (Wizard Class, Caretaker's Talent). It watches the level event the
// level-up ability emits, and it is gated at level N like every other
// ability printed on that line — so it cannot fire for a Class that
// somehow reached N by another road and then dropped back.
func BecomesLevel(n int, label string, effect func(g *game.Game, item *game.StackItem) error) game.TriggeredAbility {
	return game.TriggeredAbility{
		ActiveWhen: Level(n),
		Watches:    []game.EventKind{game.EventClassLevel},
		Key:        label,
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			return ev.CardID == source.InstanceID && ev.Amount == n
		},
		Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
			return game.NewTriggeredItem(source, label, effect)
		},
	}
}

// --- Case (CR 719) ---------------------------------------------------

// ToSolve is the "To solve — [condition]" clause every Case prints
// (CR 719.3a). It is a triggered ability with an intervening if:
//
//	"At the beginning of your end step, if [condition],
//	 this Case becomes solved."
//
// so the condition is checked TWICE — once when the trigger would go
// on the stack, and again as it resolves (CR 603.4). A Case whose
// condition stopped being true while the trigger sat on the stack does
// not solve, and this constructor is what makes that true for every
// Case without any card file writing it out.
//
// `condition` runs under g.mu (write mode at resolution, and from the
// harvest) — read-only, *ForEffect accessors only, exactly the
// contract ActivatedAbility.Condition has. `label` is the stack
// overlay's line; name the card.
//
// An already-solved Case never triggers again: solving is idempotent
// (CR 719.3b), and skipping the trigger keeps the stack clean rather
// than putting an ability there that will do nothing.
func ToSolve(label string, condition func(g *game.Game, controller, source uuid.UUID) bool) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches: []game.EventKind{game.EventBeginEndStep},
		Key:     label,
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
			if source == nil || source.Solved || ev.Actor != source.Controller {
				return false
			}
			return condition(g, source.Controller, source.InstanceID)
		},
		Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
			return game.NewTriggeredItem(source, label, func(g *game.Game, item *game.StackItem) error {
				// CR 603.4: the intervening if is checked again on
				// resolution. Nothing happens if it is false — the
				// ability is simply removed from the stack.
				if !condition(g, item.Controller, item.SourceCardID) {
					return nil
				}
				return g.SolveCaseForEffect(item.SourceCardID)
			})
		},
	}
}

// --- Harness (CR 701.64) ----------------------------------------------

// Harness is CR 701.64a's "Harness [this permanent]" activated
// ability: "If this permanent isn't harnessed, it becomes harnessed."
//
// It carries no ActiveWhen and no Condition: CR 701.64a is worded as
// an EFFECT ("if … isn't … it becomes …"), not a legality
// restriction, so nothing stops a permanent already harnessed from
// having the ability activated again — it simply does nothing the
// second time, because HarnessForEffect is idempotent. A Condition
// that refused the second activation would make the ability
// unactivatable rather than a no-op, which is a stronger and wrong
// restriction no printed Harness ability states.
//
// `label` is the printed line verbatim ("{5}{W}, {T}: Harness The
// Mind Stone."), because the cost shapes vary card to card exactly as
// LevelUp's do not, and there is only one card to generalise from so
// far.
func Harness(label string, cost game.AbilityCost) ActivatedAbility {
	return ActivatedAbility{
		Label: label,
		Cost:  cost,
		Effect: func(g *game.Game, item *game.StackItem) error {
			return g.HarnessForEffect(item.SourceCardID)
		},
	}
}

// specDesignations is every designation gate a Spec declares, across
// all five gateable slots. One walk, so a guard (Register's door
// check, and the catalog-wide test beside it) cannot cover four slots
// and forget the fifth.
//
// #1314 added GatedCastPermissions: a standing cast/play permission
// can now be gated exactly like a static, a trigger, an activated
// ability or a cost modifier (Fortune Teller's Talent's level-2 line),
// and the door guard has to see it or a card could ship a door-gated
// permission that silently never opens. Plain CastPermissions carries
// no gate at all (game.CastPermission has none — see
// game.CastPermissionGate's doc comment for why the gate lives on a
// separate wrapper), so it is not walked here.
func specDesignations(spec Spec) []game.Designation {
	out := make([]game.Designation, 0, len(spec.Static)+len(spec.Triggered)+len(spec.Activated)+len(spec.CostModifiers)+len(spec.GatedCastPermissions))
	for _, a := range spec.Static {
		out = append(out, a.ActiveWhen)
	}
	for _, t := range spec.Triggered {
		out = append(out, t.ActiveWhen)
	}
	for _, a := range spec.Activated {
		out = append(out, a.ActiveWhen)
	}
	for _, m := range spec.CostModifiers {
		out = append(out, m.ActiveWhen)
	}
	for _, p := range spec.GatedCastPermissions {
		out = append(out, p.ActiveWhen)
	}
	return out
}
