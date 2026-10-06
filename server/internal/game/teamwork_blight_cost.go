package game

import (
	"errors"

	"github.com/google/uuid"
)

// teamwork_blight_cost.go — #1703: two optional additional costs a
// cast pays at CR 601.2h, and the last two components of ADR 0073's
// cost vocabulary that printed cards were waiting on.
//
//	Teamwork N (CR 702.194a): "As an additional cost to cast this
//	spell, you may tap any number of creatures you control with total
//	power N or more." HULK SMASH!, Go Nuts!, Widow's Bite.
//
//	Blight N (CR 701.68a) paid as a cost: "As an additional cost to
//	cast this spell, you may blight 2. (You may put two -1/-1 counters
//	on a creature you control.)" Pyrrhic Strike, Cinder Strike.
//
// Both are components of game.AdditionalCost, declared on an OPTIONAL
// cost (effects.Teamwork / effects.OptionalBlight) and announced by
// index exactly as a kicker is (ADR 0073 §2). What is new is only what
// the caster NAMES while paying: the creatures tapped for teamwork
// (CastSpellParams.TeamworkIDs) and the creature blighted
// (CastSpellParams.BlightIDs). The shapes follow what the engine
// already has, one surface over:
//
//   - Teamwork is crew on a spell. CR 702.194a and CR 702.122a are the
//     same sentence ("tap any number of … creatures you control with
//     total power N or more"), so the validator reads exactly what
//     validateCrewCostLocked reads: distinct, on the battlefield, the
//     caster's, creatures, untapped, and a total EFFECTIVE power
//     (CurrentPower after a layer refresh) of at least N. Overshooting
//     is legal; the printed number is a floor.
//   - It is not the {T} symbol, so summoning sickness is deliberately
//     not checked (CR 302.6 hangs sickness on a creature's own {T} /
//     {Q} abilities; crew's CR 702.122 and convoke's CR 702.51a
//     rulings are the same reading). A creature cast this turn may
//     help pay for HULK SMASH!.
//   - Blight is the counter-placement cost (AbilityCost.AddCounter,
//     #789) aimed at a creature the caster CHOOSES rather than at the
//     source: exactly one creature the caster controls (CR 701.68a),
//     and none at all is a refusal — CR 701.68b says a player who
//     cannot put the counters on a creature they control cannot choose
//     to blight.
//
// And the blight counters go through the CR 614 counter window — see
// blightLocked for which replacements see them and why.

// TeamworkKey and BlightKey are the optional-cost keys of the two
// components. A card's resolution reads them through
// ctx.OptionalCostTimes / the #1655 mode-count conditions; the engine
// itself reads neither — the component fields are what make the cost
// payable, as ManaCost is for a kicker.
const (
	// TeamworkKey is CR 702.194's teamwork. "Cast using teamwork"
	// (CR 702.194b) is "this optional cost was announced".
	TeamworkKey = "teamwork"
	// BlightKey is an optional "you may blight N" additional cost.
	// "If this spell's additional cost was paid" is "this optional cost
	// was announced".
	BlightKey = "blight"
)

// ErrInsufficientTeamwork is returned when the creatures named to pay
// a teamwork cost do not add up to its number (CR 702.194a) —
// including the case where the cost was announced and none were named.
// ErrInsufficientCrew's twin, and distinct from ErrInvalidParam for
// the same reason: the client can say "tap more power".
var ErrInsufficientTeamwork = errors.New("game: the tapped creatures' total power is below the teamwork number")

// planTeamwork is the teamwork number the announced plan demands, or
// zero. effects.Register allows the component only on an optional cost
// that cannot repeat and allows one per card, so there is at most one
// entry — summed anyway, so a plan that somehow carried two would
// demand both rather than the larger.
func planTeamwork(plan []costPayment) int {
	n := 0
	for _, pay := range plan {
		n += pay.cost.Teamwork
	}
	return n
}

// planTapCreatures is the number of untapped creatures the plan taps
// as a cost — one per escalate payment of Collective Effort (CR
// 702.120a). Zero for every other cast.
func planTapCreatures(plan []costPayment) int {
	n := 0
	for _, pay := range plan {
		n += pay.cost.TapCreatures
	}
	return n
}

// TapCreaturesOptionsForEffect is the set of creatures `playerID` could
// tap to pay a "tap an untapped creature you control" cost: the same
// walk teamwork uses (untapped creatures they control). Named apart so
// the view and the enumerator read it by what they ask.
//
// Caller must hold g.mu (read or write).
func (g *Game) TapCreaturesOptionsForEffect(playerID uuid.UUID) []uuid.UUID {
	return g.TeamworkOptionsForEffect(playerID)
}

// validateTapCreaturesLocked checks the creatures named to pay a
// "tap N untapped creatures you control" cost: exactly N, distinct,
// the caster's, creatures, untapped (CR 118.3), and not also named to
// the convoke / waterbend taps. Power is irrelevant. No summoning
// sickness check: this is not the {T} symbol.
func (g *Game) validateTapCreaturesLocked(playerID uuid.UUID, n int, ids, otherTaps []uuid.UUID) error {
	if len(ids) != n {
		return ErrInvalidParam
	}
	tapping := make(map[uuid.UUID]bool, len(otherTaps))
	for _, id := range otherTaps {
		tapping[id] = true
	}
	seen := make(map[uuid.UUID]bool, len(ids))
	for _, id := range ids {
		if seen[id] || tapping[id] {
			return ErrInvalidParam
		}
		seen[id] = true
		c := findBattlefieldCard(g, id)
		if c == nil {
			return ErrCardNotFound
		}
		if c.Controller != playerID {
			return ErrCardCallerMismatch
		}
		if !c.IsCreature() {
			return ErrNotACreature
		}
		if c.Tapped {
			return ErrAlreadyTapped
		}
	}
	return nil
}

// planBlight is the blight N the announced plan demands, or zero. At
// most one entry, for planTeamwork's reason.
func planBlight(plan []costPayment) int {
	n := 0
	for _, pay := range plan {
		n += pay.cost.Blight
	}
	return n
}

// TeamworkOptionsForEffect is the set of creatures that could be
// tapped for a teamwork cost by `playerID` right now, in battlefield
// order: untapped creatures they control. ONE walk, read by the
// validator's predicates, the protocol view's picker and the legal-move
// enumerator, so the three cannot disagree about which creature pays
// (#544).
//
// Not a targeting walk — tapping a creature to pay a cost does not
// target it (CR 601.2h), so a hexproof creature is a legal tap.
//
// Caller must hold g.mu (read or write).
func (g *Game) TeamworkOptionsForEffect(playerID uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller != playerID || c.Tapped || !c.IsCreature() {
			continue
		}
		out = append(out, c.InstanceID)
	}
	return out
}

// TeamworkPowerForEffect is the effective power `id` contributes to a
// teamwork total — the one number the validator, the view and the
// enumerator add up. Zero for a card that is not on the battlefield.
//
// CurrentPower, not the printed power: an anthem and a +1/+1 counter
// both count, exactly as for crew (CR 702.122a). It reads the cached
// effective value the way the view's `power` and the enumerator's
// crewPayment do; the validator refreshes the layers before it adds
// anything up, so the number the engine charges is never stale.
//
// Caller must hold g.mu (read or write).
func (g *Game) TeamworkPowerForEffect(id uuid.UUID) int {
	c := findBattlefieldCard(g, id)
	if c == nil {
		return 0
	}
	return c.CurrentPower()
}

// TeamworkPayableForEffect reports whether `playerID` could pay
// teamwork `n` right now: the untapped creatures they control add up to
// at least n. The question the view asks before offering the toggle and
// the enumerator asks before announcing the cost.
//
// Only POSITIVE powers are added here, because this is "is there a
// subset that clears the bar" and a creature with negative power is
// simply left out of the best subset. The validator adds what the
// caster actually named.
//
// Caller must hold g.mu (read or write).
func (g *Game) TeamworkPayableForEffect(playerID uuid.UUID, n int) bool {
	if n <= 0 {
		return true
	}
	total := 0
	for _, id := range g.TeamworkOptionsForEffect(playerID) {
		if p := g.TeamworkPowerForEffect(id); p > 0 {
			total += p
		}
	}
	return total >= n
}

// validateTeamworkLocked checks the creatures named to pay the plan's
// teamwork cost, without tapping anything (validate everything, then
// pay everything — ADR 0020 §3).
//
//   - No teamwork announced: any named creature is refused, as a crew
//     id on an ability with no crew is.
//   - Teamwork announced: each creature distinct, on the battlefield,
//     controlled by the caster, a creature, untapped (CR 118.3 — a
//     tapped creature cannot be tapped again), and not also named to
//     the convoke / waterbend taps. The effective total must reach N.
//   - No summoning-sickness check: this is not the {T} symbol.
//
// Caller must hold g.mu.
func (g *Game) validateTeamworkLocked(playerID uuid.UUID, plan []costPayment, ids, otherTaps []uuid.UUID) error {
	n := planTeamwork(plan)
	if tapN := planTapCreatures(plan); tapN > 0 {
		return g.validateTapCreaturesLocked(playerID, tapN, ids, otherTaps)
	}
	if n <= 0 {
		if len(ids) > 0 {
			return ErrInvalidParam
		}
		return nil
	}
	if len(ids) == 0 {
		return ErrInsufficientTeamwork
	}
	g.RecomputeLayersIfStaleLocked()
	tapping := make(map[uuid.UUID]bool, len(otherTaps))
	for _, id := range otherTaps {
		tapping[id] = true
	}
	seen := make(map[uuid.UUID]bool, len(ids))
	total := 0
	for _, id := range ids {
		// One creature pays one tap. Naming it twice would let a
		// single 4-power creature pay teamwork 8, and naming it to
		// convoke as well would tap it twice.
		if seen[id] || tapping[id] {
			return ErrInvalidParam
		}
		seen[id] = true
		c := findBattlefieldCard(g, id)
		if c == nil {
			return ErrCardNotFound
		}
		if c.Controller != playerID {
			return ErrCardCallerMismatch
		}
		if !c.IsCreature() {
			return ErrNotACreature
		}
		if c.Tapped {
			return ErrAlreadyTapped
		}
		total += c.CurrentPower()
	}
	if total < n {
		return ErrInsufficientTeamwork
	}
	return nil
}

// payTeamworkLocked taps the validated creatures, one EventTapCard
// each, so "whenever a creature you control becomes tapped" sees every
// one of them. Called after the spell is on the stack (CR 601.2a before
// 601.2h), so a tap-watching trigger goes above it.
//
// Caller must hold g.mu.
func (g *Game) payTeamworkLocked(playerID uuid.UUID, ids []uuid.UUID) {
	for _, id := range ids {
		c := findBattlefieldCard(g, id)
		if c == nil || c.Tapped {
			continue
		}
		c.Tapped = true
		g.EmitEvent(Event{Kind: EventTapCard, Actor: playerID, CardID: id})
	}
}

// BlightOptionsForEffect is the set of creatures `playerID` could
// blight right now, in battlefield order: the creatures they control
// (CR 701.68a). One walk for the validator, the view and the
// enumerator (#544). Not a targeting walk — a hexproof creature can be
// blighted.
//
// A 1/1 IS an option for blight 2. CR 701.68b refuses a blight only
// when the counters cannot be PUT, not when the creature would die of
// them — and "a creature that dies from the counters still pays the
// cost" is the point of the rule, not a corner of it.
//
// Caller must hold g.mu (read or write).
func (g *Game) BlightOptionsForEffect(playerID uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller != playerID || !c.IsCreature() {
			continue
		}
		out = append(out, c.InstanceID)
	}
	return out
}

// validateBlightLocked checks the creature named to pay the plan's
// blight cost: exactly one when blight is announced, none otherwise,
// and that one a creature on the battlefield under the caster's control
// (CR 701.68a). Nothing is placed here.
//
// Caller must hold g.mu.
func (g *Game) validateBlightLocked(playerID uuid.UUID, plan []costPayment, ids []uuid.UUID) error {
	n := planBlight(plan)
	if n <= 0 {
		if len(ids) > 0 {
			return ErrInvalidParam
		}
		return nil
	}
	if len(ids) != 1 {
		return ErrInvalidParam
	}
	c := findBattlefieldCard(g, ids[0])
	if c == nil {
		return ErrCardNotFound
	}
	if c.Controller != playerID {
		return ErrCardCallerMismatch
	}
	if !c.IsCreature() {
		return ErrNotACreature
	}
	return nil
}

// blightLocked puts `n` -1/-1 counters on `creatureID` as a COST
// (CR 701.68a paid at CR 601.2h), placed by `playerID`.
//
// THROUGH the CR 614 counter window, unlike the loyalty cost and
// Devoted Druid's AddCounter, and with the event marked
// CounterFromCost. The distinction the mark carries is CR 614.16's:
//
//	"Some replacement effects apply 'if an effect would … put one or
//	more counters on a permanent.' These replacement effects apply if
//	the effect of a resolving spell or ability … puts a counter on a
//	permanent."
//
// Paying a cost is not the effect of a resolving spell or ability, so
// a replacement worded "if an EFFECT would put" — Doubling Season —
// does NOT double the blight, and reads the mark to say so. A
// replacement that names no effect is a replacement of the EVENT and
// applies to a cost's counters like any other: Vorinclex's "if you
// would put", Winding Constrictor's and Vizier of Remedies' "if one or
// more counters would be put". That is the Devoted Druid + Vizier of
// Remedies ruling, and it is why the window is opened at all rather
// than applyCounterLocked writing the counters directly. (Hardened
// Scales watches +1/+1 counters and never sees a blight.)
//
// A cost cannot pause (CR 601.2h pays the costs as one step), so the
// window settles now: two ordered replacements are applied in gathered
// order rather than asking (the mustSettleNow posture every other
// cost-shaped counter and life payment takes).
//
// The cost is paid whatever the window settles on — a Vizier that
// reduces blight 1 to zero counters still leaves the spell's blight
// paid, and so does a creature that dies of the counters at the next
// state-based check (CR 704.5f). Returns the counters that landed.
//
// Caller must hold g.mu.
func (g *Game) blightLocked(playerID, creatureID uuid.UUID, n int) (int, error) {
	return g.addCounterMustSettleNowLocked(&ReplacementEvent{
		Kind:            RepEventCounter,
		CounterTarget:   creatureID,
		CounterName:     CounterMinusOne,
		CounterDelta:    n,
		CounterPlacer:   playerID,
		CounterFromCost: true,
	})
}

// payBlightLocked pays the plan's blight cost onto the named creature.
// Validated at announce; a creature that has left in between (nothing
// can move it under one lock today) is skipped rather than erroring.
//
// Caller must hold g.mu.
func (g *Game) payBlightLocked(playerID uuid.UUID, plan []costPayment, ids []uuid.UUID) error {
	n := planBlight(plan)
	if n <= 0 || len(ids) != 1 {
		return nil
	}
	if findBattlefieldCard(g, ids[0]) == nil {
		return nil
	}
	_, err := g.blightLocked(playerID, ids[0], n)
	return err
}
