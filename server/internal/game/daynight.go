package game

// daynight.go — day and night (CR 731) and the daybound / nightbound
// keyword pair (CR 702.145). ADR 0132.
//
// The designation belongs to the GAME, not to a player or a permanent:
// the game starts with neither (CR 731.1), the first "it becomes day" or
// "it becomes night" gives it one, and from then on it has exactly one,
// for good. So the state is one string on Game, plus the one number the
// turn-based check needs and that nothing else keeps (the previous
// turn's spell count, see recordPrevTurnSpellsLocked).
//
// Four rules make up the whole mechanic, each with a home below:
//
//   - CR 731.2 / 502.2 / 703.4b: as the second part of the untap step,
//     day becomes night if the previous turn's active player cast no
//     spells, and night becomes day if they cast two or more. Neither
//     stays neither. dayNightTurnCheckLocked, called from
//     performUntapStepLocked straight after phasing.
//   - CR 731.1a: "day becomes night" is the game losing one designation
//     and gaining the other. "It becomes night" while it is already
//     night does nothing, and neither-to-night is a gain, not a change
//     of one into the other: no "whenever day becomes night" trigger
//     sees it. becomeDayNightLocked emits EventDayNightChanged on every
//     real change and carries which kind in Amount.
//   - CR 702.145b / 702.145e: a permanent with daybound or nightbound
//     can't transform except due to that ability. TransformPermanent-
//     ForEffect refuses one; the day/night transform goes around the
//     refusal through transformInPlaceLocked.
//   - CR 702.145b-g: daybound makes it day if it is neither (d), every
//     front-face-up daybound permanent transforms when it is night (c),
//     every back-face-up nightbound permanent when it is day (f), and a
//     daybound permanent that enters at night enters transformed (b).
//     settleDayNightLocked is the "any time" half; the entry half is
//     dayNightListener.

import "github.com/google/uuid"

// DayNightDesignation is the game's day/night designation. The zero
// value is "neither day nor night", which is how every game starts.
type DayNightDesignation string

const (
	// DesignationNeither is the start of every game (CR 731.1).
	DesignationNeither DayNightDesignation = ""
	DesignationDay     DayNightDesignation = "day"
	DesignationNight   DayNightDesignation = "night"
)

// KeywordDaybound and KeywordNightbound are the canonical tokens for
// CR 702.145. They are named because three packages read them and a
// typo would leave a werewolf stuck on one face.
const (
	KeywordDaybound   = "daybound"
	KeywordNightbound = "nightbound"
)

// DayNightState is everything the game remembers about day and night.
// Plain data, carried by Clone, RestoreFrom and the snapshot.
type DayNightState struct {
	// Designation is the current designation; empty before the first
	// "it becomes day / night".
	Designation DayNightDesignation `json:"designation,omitempty"`
	// PrevTurnSpells is how many spells the previous turn's active
	// player cast during that turn, and PrevTurnKnown says there was a
	// previous turn at all. CR 502.2 asks "did the previous turn's
	// active player cast no spells" at the start of the NEXT turn, and
	// by then SpellsCastThisTurn has been reset, so the count is copied
	// out as the turn ends. Unknown (the first turn of the game) means
	// the check has nothing to read and does not happen.
	PrevTurnSpells int  `json:"prevTurnSpells,omitempty"`
	PrevTurnKnown  bool `json:"prevTurnKnown,omitempty"`
}

// snapshotDayNight is the snapshot's pointer form: nil for the zero
// state, so a game that never had a designation writes no key at all
// and every frozen fixture re-encodes byte for byte (the OpeningRoll
// precedent). The state is a flat value, so the copy is the clone.
func snapshotDayNight(s DayNightState) *DayNightState {
	if s == (DayNightState{}) {
		return nil
	}
	return &s
}

// valueOrZero is the restore half: a missing key is the zero state.
func (s *DayNightState) valueOrZero() DayNightState {
	if s == nil {
		return DayNightState{}
	}
	return *s
}

// EventDayNightChanged — the game gained or changed its day/night
// designation (CR 731.1). `Label` is the NEW designation ("day" or
// "night"). `Amount` is 1 when it was already day or night and has
// flipped to the other (CR 731.1a: "day becomes night", "night becomes
// day") and 0 when the game had neither and has just gained one. Only a
// flip is what "whenever day becomes night or night becomes day" waits
// for, so use DayNightFlipped rather than reading the fields.
//
// Emitted by becomeDayNightLocked only, and only on a real change, so a
// second "it becomes day" is not an event. Added by #2561 (ADR 0132).
const EventDayNightChanged EventKind = "day_night_changed"

// DayNightFlipped reports whether ev is "day becomes night or night
// becomes day": a change of the designation from one to the other, not
// the first one the game ever gains.
func DayNightFlipped(ev Event) bool {
	return ev.Kind == EventDayNightChanged && ev.Amount == 1
}

// DayNightDesignation returns the game's current designation. Caller
// must hold g.mu.
func (g *Game) DayNightDesignation() DayNightDesignation { return g.DayNight.Designation }

// IsDay and IsNight are the conditions "if it's day" and "if it's
// night" ask. Neither is true before the game has a designation. Caller
// must hold g.mu.
func (g *Game) IsDay() bool   { return g.DayNight.Designation == DesignationDay }
func (g *Game) IsNight() bool { return g.DayNight.Designation == DesignationNight }

// BecomeDayForEffect is "it becomes day". Already day: nothing happens
// (CR 731.1). Caller must hold g.mu in write mode (the effect_api.go
// convention).
func (g *Game) BecomeDayForEffect() { g.becomeDayNightLocked(DesignationDay) }

// BecomeNightForEffect is "it becomes night".
func (g *Game) BecomeNightForEffect() { g.becomeDayNightLocked(DesignationNight) }

// BecomeDayIfNeitherForEffect is "if it's neither day nor night, it
// becomes day" — the clause on The Celestus, Brimstone Vandal and every
// other card that sets the clock going as it enters. A game that
// already has a designation is left alone, which is what the "if" is
// for: it must not turn night into day.
func (g *Game) BecomeDayIfNeitherForEffect() {
	if g.DayNight.Designation == DesignationNeither {
		g.becomeDayNightLocked(DesignationDay)
	}
}

// ToggleDayNightForEffect is The Celestus's "if it's night, it becomes
// day. Otherwise, it becomes night": night goes to day, and day or
// neither goes to night.
func (g *Game) ToggleDayNightForEffect() {
	if g.IsNight() {
		g.becomeDayNightLocked(DesignationDay)
		return
	}
	g.becomeDayNightLocked(DesignationNight)
}

// becomeDayNightLocked is the one write to the designation. It
// announces the change, then flips whatever daybound / nightbound
// permanents the new designation turns over. Announcing first is what
// reads naturally in the log ("it becomes night", then the werewolves
// transform) and costs nothing: the triggers the event queues resolve
// later, so none of them sees a half-turned board.
//
// Caller must hold g.mu in write mode.
func (g *Game) becomeDayNightLocked(to DayNightDesignation) {
	from := g.DayNight.Designation
	if to == DesignationNeither || from == to {
		return
	}
	g.DayNight.Designation = to
	flipped := 0
	if from != DesignationNeither {
		flipped = 1
	}
	g.EmitEvent(Event{
		Kind:   EventDayNightChanged,
		Label:  string(to),
		Amount: flipped,
	})
	g.settleDayNightLocked()
}

// recordPrevTurnSpellsLocked copies the ending turn's active player's
// spell count out before the rotation replaces the turn and the tally
// (CR 502.2 reads it one turn later). Runs from beginNextTurnLocked
// beside recordLastTurnAttacksLocked. Copies are not cast (CR 707.10),
// so CastTally.Total is exactly the number the rule counts.
//
// Caller must hold g.mu.
func (g *Game) recordPrevTurnSpellsLocked() {
	seat := g.Turn.ActiveSeat
	if seat < 0 || seat >= len(g.Seats) || g.Seats[seat] == nil {
		return
	}
	g.DayNight.PrevTurnSpells = g.CastTallyFor(g.Seats[seat].ID).Total
	g.DayNight.PrevTurnKnown = true
}

// dayNightTurnCheckLocked is CR 731.2 / 502.2: the check that changes
// the designation on its own, as the second turn-based action of the
// untap step, after phasing and before the permanents untap. No stack.
//
// Day goes to night when the previous turn's active player cast no
// spells; night goes to day when they cast two or more; one spell
// changes nothing either way. Neither stays neither (CR 731.2c). The
// first turn has no previous turn, so there is nothing to read.
//
// Caller must hold g.mu in write mode.
func (g *Game) dayNightTurnCheckLocked() {
	if !g.DayNight.PrevTurnKnown {
		return
	}
	switch g.DayNight.Designation {
	case DesignationDay:
		if g.DayNight.PrevTurnSpells == 0 {
			g.becomeDayNightLocked(DesignationNight)
		}
	case DesignationNight:
		if g.DayNight.PrevTurnSpells >= 2 {
			g.becomeDayNightLocked(DesignationDay)
		}
	}
}

// settleDayNightLocked is the "any time" half of CR 702.145: the
// clauses that hold whenever the board and the designation line up the
// wrong way, run at every moment either of them can have moved.
//
//   - CR 702.145d: a permanent with daybound on the battlefield and
//     neither day nor night: it becomes day.
//   - CR 702.145g: failing that, a permanent with nightbound and no
//     daybound permanent anywhere on the battlefield: it becomes night.
//   - CR 702.145c: at night every front-face-up daybound permanent
//     transforms. CR 702.145f: by day every back-face-up nightbound one
//     does. "This happens immediately and isn't a state-based action",
//     so it is a plain loop here, not a check in the state-based pass.
//
// A permanent that isn't represented by a double-faced card can't
// transform (CR 712.9), which CanTransform already says; a Clone of a
// werewolf stays what it is.
//
// Caller must hold g.mu in write mode. Idempotent.
func (g *Game) settleDayNightLocked() {
	if g.Battlefield == nil {
		return
	}
	if g.DayNight.Designation == DesignationNeither {
		daybound, nightbound := false, false
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if HasKeyword(c, KeywordDaybound) {
				daybound = true
				break
			}
			if HasKeyword(c, KeywordNightbound) {
				nightbound = true
			}
		}
		switch {
		case daybound:
			g.becomeDayNightLocked(DesignationDay)
		case nightbound:
			g.becomeDayNightLocked(DesignationNight)
		}
		// becomeDayNightLocked settled the board itself.
		return
	}
	night := g.DayNight.Designation == DesignationNight
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		frontDaybound := c.ActiveFace == 0 && HasKeyword(c, KeywordDaybound)
		backNightbound := c.ActiveFace == 1 && HasKeyword(c, KeywordNightbound)
		if (night && frontDaybound || !night && backNightbound) && CanTransform(*c) {
			g.transformInPlaceLocked(c)
		}
	}
}

// hasDayNightKeyword reports whether the permanent has daybound or
// nightbound, the two abilities that take its transforming away from
// everything else (CR 702.145b, 702.145e).
func hasDayNightKeyword(c *Card) bool {
	return HasKeyword(c, KeywordDaybound) || HasKeyword(c, KeywordNightbound)
}

// dayNightListener is the entry half of CR 702.145b: "If it is night
// and this permanent is represented by a double-faced card, it enters
// transformed". It is a Listener on the zone move rather than a
// replacement in the CR 614 pipeline because the pipeline's job there
// is to let an entering permanent be changed BEFORE it arrives, and
// what this changes is only which face is up: no counters, no tapped
// status, nothing a player is asked about. The move event is emitted
// after the card is on the battlefield and BEFORE its ETB event, so
// "when this enters" triggers are harvested off the face it actually
// entered as, and nothing ever sees it transform (it never did: CR
// 712.18 does not apply, and no EventTransform is emitted).
//
// It also settles the board after any entry or phase-in, which is when
// a daybound permanent first appears (CR 702.145d).
type dayNightListener struct{}

// OnEvent runs under g.mu in write mode.
func (dayNightListener) OnEvent(g *Game, ev Event) {
	switch ev.Kind {
	case EventZoneMove:
		if ev.NewZone != ZoneBattlefield {
			return
		}
		g.dayNightEntryLocked(ev.CardID)
		g.settleDayNightLocked()
	case EventPhaseIn:
		g.settleDayNightLocked()
	}
}

// dayNightEntryLocked turns a daybound permanent that has just entered
// at night over to its back face, without a transform.
func (g *Game) dayNightEntryLocked(cardID uuid.UUID) {
	if !g.IsNight() {
		return
	}
	card := findBattlefieldCard(g, cardID)
	if card == nil || card.ActiveFace != 0 || !HasKeyword(card, KeywordDaybound) || !CanTransform(*card) {
		return
	}
	card.SetFace(1)
	card.effective = nil
	g.layerVersion.Add(1)
}
