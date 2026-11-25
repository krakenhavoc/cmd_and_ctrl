package game

import (
	"fmt"

	"github.com/google/uuid"
)

// flagbearer.go — #6236, Coalition Flag and Standard Bearer: "While an
// opponent is choosing targets as part of casting a spell they control or
// activating an ability they control, that player must choose at least one
// Flagbearer on the battlefield if able."
//
// CR 601.2c (and CR 602.2b for an activated ability): "If any effects say
// that an object or player must be chosen as a target, the player chooses
// targets so that they obey the maximum possible number of such effects
// without violating any rules or effects that say that an object or player
// can't be chosen as a target." The rule is a static of the permanent that
// prints it (Characteristic.FlagbearerTargetRule, written by a layer-6 self
// static, so it goes with the permanent's abilities, CR 613.1f), and it
// binds each OPPONENT of that permanent's controller. It is judged on the
// whole announcement: a spell or ability whose chosen targets include no
// Flagbearer on the battlefield is refused when one of its target steps
// could legally have chosen one. Every such rule is obeyed by the same one
// Flagbearer, so how many there are does not matter.
//
// Not judged: a triggered ability's targets (the rule names casting and
// activating), a step whose targets an opponent or a random draw chooses
// (the chooser is not the bound player), and a spell's new targets (CR
// 115.7 is not "choosing targets as part of casting"). Data, not a func:
// the bit is on the characteristic, and the reader walks the battlefield.

// ErrMustTargetFlagbearer is the refusal of an announcement that chose no
// Flagbearer while one could have been chosen. It wraps ErrIllegalTarget.
var ErrMustTargetFlagbearer = fmt.Errorf("%w: you must choose at least one Flagbearer as a target if able", ErrIllegalTarget)

// flagbearerRuleBindsLocked reports whether a permanent with the rule is on
// the battlefield under the control of an opponent of `chooser`.
//
// Caller must hold g.mu with fresh layers.
func (g *Game) flagbearerRuleBindsLocked(chooser uuid.UUID) bool {
	if g.Battlefield == nil || chooser == uuid.Nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.effective != nil && c.effective.FlagbearerTargetRule && c.Controller != chooser {
			return true
		}
	}
	return false
}

// isFlagbearerLocked reports whether `id` names a Flagbearer on the
// battlefield (the creature type, read through the layered view).
//
// Caller must hold g.mu with fresh layers.
func (g *Game) isFlagbearerLocked(id uuid.UUID) bool {
	c := findBattlefieldCard(g, id)
	return c != nil && c.HasSubtype("Flagbearer")
}

// checkFlagbearerTargetsLocked is the announce-time gate: nil when the
// rule does not bind the chooser, when a chosen target is a Flagbearer,
// or when no step the chooser picks for could have chosen one; otherwise
// ErrMustTargetFlagbearer.
//
// Caller must hold g.mu with fresh layers (it reads, never recomputes,
// so the enumerator's read lock is enough).
func (g *Game) checkFlagbearerTargetsLocked(src TargetSource, steps []AnnouncedClause, targets []TargetRef) error {
	if len(steps) == 0 {
		return nil
	}
	if !g.flagbearerRuleBindsLocked(src.Controller) {
		return nil
	}
	for _, t := range targets {
		if t.Kind == TargetCard && g.isFlagbearerLocked(t.ID) {
			return nil
		}
	}
	for i := range steps {
		if StepChosenByOpponent(steps[i]) || StepChosenAtRandom(steps[i]) {
			continue
		}
		lt := g.legalTargetsLocked(src, &steps[i].Clause)
		for _, id := range lt.Cards {
			if g.isFlagbearerLocked(id) {
				return ErrMustTargetFlagbearer
			}
		}
	}
	return nil
}

// FlagbearerTargetsAllowedForEffect is checkFlagbearerTargetsLocked as a
// yes or no, for the legal-move enumerator, so a bot is never offered an
// announcement the engine refuses. Caller must hold g.mu.
func (g *Game) FlagbearerTargetsAllowedForEffect(src TargetSource, steps []AnnouncedClause, targets []TargetRef) bool {
	return g.checkFlagbearerTargetsLocked(src, steps, targets) == nil
}
