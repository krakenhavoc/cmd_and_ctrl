package game

import (
	"slices"

	"github.com/google/uuid"
)

// random_target_announce.go — #4051: "target … chosen at random" on an
// ACTIVATED ability (Goblin Test Pilot: "{T}: This creature deals 2
// damage to any target chosen at random").
//
// An activation's targets are announced with it (CR 602.2b, then
// CR 601.2c), and the ruling puts the draw there: "The target is chosen
// at random as part of activating the ability. Players can respond to
// the ability knowing what the target of the ability is." So the
// activator's payload names no target for a random clause (the client's
// walk skips it, the bot's enumerator names none), and the announcement
// draws it: uniformly over the clause's legal set ("all possible legal
// targets (including creatures and players) must have an equal chance
// of being chosen"), through the same keyed "pick" stream the trigger
// walk draws from (drawStepTargetLocked, #1815), so an undo replays it.
// With no legal target the activation is refused, as any required
// clause's is (CR 601.2c).
//
// The draw is carried on the announcement (ActivateAbilityParams.
// randomTargets), as a random discard is (ADR 0109 §7): a run parked on
// a CR 903.9 question and made again aims at the same target rather
// than drawing a new one. It is logged once the ability is on the stack.

// StepChosenAtRandom reports whether a step is a "target … chosen at
// random" clause, which the engine draws and the controller's own
// announcement leaves out (the client's walk and the bot's enumerator
// ask the same question).
func StepChosenAtRandom(s AnnouncedClause) bool {
	return s.Clause.AtRandom
}

// announceRandomTargetsLocked fills an announcement's random clauses.
// `targets` are the refs the payload named, slotted; a ref aimed at a
// random clause is refused (ErrInvalidParam), since no payload may
// choose a random target. `preset` is the draw an earlier run of the
// same announcement made, reused as it is; nil draws afresh. Returns
// the full list in step order (the order validateAnnouncedTargetsLocked
// requires), and the refs drawn, for the announcement to carry.
//
// Caller must hold g.mu.
func (g *Game) announceRandomTargetsLocked(src TargetSource, d RandomDraw, steps []AnnouncedClause, targets, preset []TargetRef) (all, drawn []TargetRef, err error) {
	if !slices.ContainsFunc(steps, StepChosenAtRandom) {
		if len(preset) > 0 {
			return nil, nil, ErrInvalidParam
		}
		return targets, nil, nil
	}
	var unslotted []TargetRef
	for _, t := range targets {
		if t.Kind != TargetPlayer && t.Kind != TargetCard {
			unslotted = append(unslotted, t)
			continue
		}
		if i := stepIndexOf(steps, t.Mode, t.Slot); i < 0 || steps[i].Clause.AtRandom {
			return nil, nil, ErrInvalidParam
		}
	}
	all = make([]TargetRef, 0, len(targets)+1)
	for i, s := range steps {
		if !s.Clause.AtRandom {
			for _, t := range targets {
				if (t.Kind == TargetPlayer || t.Kind == TargetCard) && t.Mode == s.Mode && t.Slot == s.Slot {
					all = append(all, t)
				}
			}
			continue
		}
		ref, ok := presetRandomTarget(preset, s.Mode, s.Slot)
		if !ok {
			lt := g.randomStepLegalLocked(src, steps, i, all)
			if len(lt.Players) == 0 && len(lt.Cards) == 0 {
				return nil, nil, ErrIllegalTarget
			}
			ref = g.drawStepTargetLocked(d, s, lt)
		}
		all = append(all, ref)
		drawn = append(drawn, ref)
	}
	return append(all, unslotted...), drawn, nil
}

// presetRandomTarget is the earlier run's draw for (mode, slot).
func presetRandomTarget(preset []TargetRef, mode, slot int) (TargetRef, bool) {
	for _, r := range preset {
		if r.Mode == mode && r.Slot == slot {
			return r, true
		}
	}
	return TargetRef{}, false
}

// randomStepLegalLocked is the legal set a random step draws from,
// given the picks made for the earlier steps: the clause's own legal
// set, less what a Distinct clause may not repeat.
//
// Caller must hold g.mu.
func (g *Game) randomStepLegalLocked(src TargetSource, steps []AnnouncedClause, idx int, picked []TargetRef) LegalTargets {
	lt := g.LegalTargetsForAnnouncedClauseForEffect(src, steps, idx, append([]TargetRef(nil), picked...))
	return withoutPicked(lt, picked, steps[idx].Clause.Distinct)
}

// RandomStepFillableForEffect reports whether step idx, a random
// clause, has anything to draw given the picks made so far: the
// question the bot's enumerator asks of a clause nobody picks (#4051),
// the same gate the announcement refuses on.
//
// Caller must hold g.mu.
func (g *Game) RandomStepFillableForEffect(src TargetSource, steps []AnnouncedClause, idx int, picked []TargetRef) bool {
	lt := g.randomStepLegalLocked(src, steps, idx, picked)
	return len(lt.Players) > 0 || len(lt.Cards) > 0
}

// emitAnnouncedRandomTargetsLocked logs each target an activation's
// announcement drew, once the ability is on the stack.
//
// Caller must hold g.mu.
func (g *Game) emitAnnouncedRandomTargetsLocked(actor, source uuid.UUID, steps []AnnouncedClause, drawn []TargetRef) {
	for _, ref := range drawn {
		if i := stepIndexOf(steps, ref.Mode, ref.Slot); i >= 0 {
			g.emitRandomTargetLocked(actor, source, steps[i], ref)
		}
	}
}
