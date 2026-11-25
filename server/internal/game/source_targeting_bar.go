package game

import "github.com/google/uuid"

// source_targeting_bar.go — #6196: "can't be the target of" a SPELL OR
// ABILITY WITH A QUALITY, as a static on the permanent.
//
//	Bartel Runeaxe, Tetsuo Umezawa  "can't be the target of Aura spells"
//	Raiding Party                   "can't be the target of white spells
//	                                 or abilities from white sources"
//	Artifact Ward                   "Enchanted creature can't be the
//	                                 target of abilities from artifact
//	                                 sources"
//	Anti-Magic Aura                 "Enchanted creature can't be the
//	                                 target of spells"
//	Lurker                          "can't be the target of spells unless
//	                                 it attacked or blocked this turn"
//
// # Why none of these is hexproof, shroud or protection
//
// Hexproof and shroud are keywords and say nothing about the source's
// characteristics; the by-opponents bar (targeting_bar.go) reads who
// controls the source; the table-wide bar (targeting_bar_table.go) reads
// nothing at all. Protection comes closest — CR 702.16b is exactly
// "can't be targeted by spells with the stated quality and can't be
// targeted by abilities from a source with the stated quality" — but
// protection also prevents damage, forbids blocking and stops enchanting
// (CR 702.16b's other three clauses), and none of these cards says any
// of that. Raiding Party is not protection from white: a white creature
// may still block it, white damage still kills it.
//
// So the quality MATCHER is protection's (ProtectionQuality, read off
// the source's characteristics the way CR 702.16b reads it) and the
// EFFECT is one line in the targeting gate. The reuse is the point: a
// white source, an artifact source, an Aura spell and "everything" are
// all one data value, and nothing here parses a quality of its own.
//
// # Where it sits in canBeTargetedBy
//
// Beside the two existing bars and AHEAD of hexproof, for
// targeting_bar.go's reason: a "may be targeted as though it didn't have
// hexproof" waiver (Nowhere to Run, Detection Tower) waives hexproof and
// nothing else, and must not lift a bar that is not hexproof.
//
// Both halves of targeting legality go through the gate — the CR 601.2c
// announce check and the CR 608.2b resolution re-check — so a spell
// already aimed at Lurker fizzles if the creature's combat record no
// longer lifts the bar.
//
// # What is deliberately not here
//
// Wall of Shadows' "can't be the target of spells that can target only
// Walls or of abilities that can target only Walls" reads the SPELL'S
// TARGET SPEC rather than any characteristic of it, and nothing in the
// engine can answer "could this spell only ever have targeted a Wall":
// a clause's predicate is a closure. It needs a declared flag on the
// target clause itself, which is an owner decision about the targeting
// vocabulary rather than a line here.

// SourceTargetingBar is one "can't be the target of <spells or abilities
// with a quality>" static, as the permanent's characteristic carries it.
//
// Plain data, so the Characteristic stays copyable and a last-known
// record round-trips it (AttackTargetRestriction's rule).
type SourceTargetingBar struct {
	// Label is the clause as printed, for the refusal and the card's
	// chip.
	Label string

	// Quality is the source quality the bar reads, CR 702.16b's matcher
	// (ProtectionQuality): "white", "artifact", the subtype "Aura", or
	// ProtectionQualityEverything for a bar on every source (Anti-Magic
	// Aura's "can't be the target of spells").
	Quality ProtectionQuality

	// Spells and Abilities say which halves the clause bars. "Aura
	// spells" and Anti-Magic Aura's "spells" are Spells alone;
	// "abilities from artifact sources" is Abilities alone; "white
	// spells or abilities from white sources" is both. A bar with
	// neither is inert rather than total: a clause that names no half
	// is a card-file bug and must not bar everything.
	Spells    bool `json:",omitempty"`
	Abilities bool `json:",omitempty"`

	// UnlessAttackedOrBlockedThisTurn lifts the bar once the permanent
	// has attacked or blocked this turn (Lurker). Read live off the
	// turn's records, so it lifts the moment the creature is declared
	// and stays lifted for the rest of the turn — including after combat,
	// when the creature is no longer attacking anything.
	UnlessAttackedOrBlockedThisTurn bool `json:",omitempty"`
}

// barsSource reports whether this bar refuses `src`: the source is the
// half the clause names, and its characteristics carry the quality.
//
// A spell is read as a spell (TargetSource.IsSpell) and everything else
// as an ability, which is what CR 702.16b's two clauses divide the world
// into: the engine's third kind, a triggered ability (#5718), is an
// ability here.
func (b SourceTargetingBar) barsSource(src TargetSource) bool {
	if !b.Spells && !b.Abilities {
		return false
	}
	if src.IsSpell() {
		if !b.Spells {
			return false
		}
	} else if !b.Abilities {
		return false
	}
	return b.Quality.Matches(src.Characteristics())
}

// sourceTargetingBarLocked reports whether a static on `c` bars a spell
// or ability with `src`'s characteristics from targeting it (#6196), and
// the clause that did.
//
// Nil-receiver safe, as scopedTargetingBarLocked is: the exported
// CanBeTargetedBy has no game and asks through a nil one, where Lurker's
// condition cannot be read and the bar therefore applies — the stricter
// reading, and the one a caller with no game has to take.
//
// Caller must hold g.mu (read or write) with fresh layers. Reads only.
func (g *Game) sourceTargetingBarLocked(c *Card, src TargetSource) (string, bool) {
	if c == nil || c.effective == nil {
		return "", false
	}
	for _, b := range c.effective.TargetingBars {
		if !b.barsSource(src) {
			continue
		}
		if b.UnlessAttackedOrBlockedThisTurn && g != nil && g.attackedOrBlockedThisTurnLocked(c) {
			continue
		}
		return b.Label, true
	}
	return "", false
}

// SourceTargetingBarForEffect is sourceTargetingBarLocked for the view
// and for card code: the clause that bars a source with these
// characteristics, or "" when nothing does.
//
// Caller must hold g.mu with fresh layers. Reads only.
func (g *Game) SourceTargetingBarForEffect(c *Card, src TargetSource) string {
	label, _ := g.sourceTargetingBarLocked(c, src)
	return label
}

// attackedOrBlockedThisTurnLocked reports whether the OBJECT `c` is now
// was declared as an attacker or as a blocker this turn (Lurker's
// condition, #6196). The declaration is history, so a creature removed
// from combat still counts, and a creature that left and came back is a
// new object with no combat of its own (CR 400.7).
//
// Caller must hold g.mu.
func (g *Game) attackedOrBlockedThisTurnLocked(c *Card) bool {
	if g == nil || c == nil {
		return false
	}
	if g.AttackedThisTurn(c.InstanceID) {
		return true
	}
	// A creature blocking right now was declared as a blocker this turn.
	// The turn's record of past declarations is the block tally
	// (TurnTally.Blocks); a blocker whose block the engine has since
	// taken back is the one case the live read misses, and it reads
	// STRICTER than printed (the bar stays up), never weaker.
	return c.BlockingTarget != uuid.Nil || len(c.AlsoBlocking) > 0
}
