package game

import "github.com/google/uuid"

// damage_as_though.go — "damage is dealt as though its source had wither"
// (Everlasting Torment) and "damage is dealt to you as though its source
// had infect" (Phyrexian Unlife), read at one gate (ADR 0108 §10, #1889).
//
// THE RULES.
//
//   - CR 120.3: what damage does depends on "the characteristics of the
//     damage's source". CR 120.3b: damage to a player by a source with
//     infect gives that player poison counters; CR 120.3d: damage to a
//     creature by a source with wither and/or infect puts -1/-1 counters
//     on it. Both are put or given by the source's controller.
//   - CR 702.80a, 702.90b, 702.90c: wither and infect are those results.
//     CR 702.80c and 702.90e: they work from any zone.
//   - CR 609.4: an "as though" effect "applies only to the stated effect
//     … For all other purposes, treat the game normally." So the source
//     gains no ability. A "whenever a creature with infect deals damage"
//     trigger does not see it, and the view does not badge the source.
//   - CR 613.11: a rules effect, read where the rule it changes is used —
//     as the damage is processed into its results (CR 120.4c), after the
//     CR 614 window has settled how much damage there is (CR 120.4b).
//
// WHERE IT IS READ. The source's own infect, wither and toxic are
// snapshotted onto the damage tail as the event is opened
// (damageTail.result, ADR 0056 Decision 2). This gate ORs a battlefield
// static's Wither or Infect into that snapshot as the damage lands, in the
// two landing functions (applyResolvedDamageToPlayerLocked and
// applyResolvedDamageToPermanentLocked) and nowhere else. So:
//
//   - prevention and every other damage replacement apply first, to the
//     damage, as they always have: a prevented point gives no counter, a
//     doubled hit gives twice the counters, a charged shield takes its
//     charge off the damage before any counter is counted. Under "damage
//     can't be prevented" (ADR 0107 §5, Everlasting Torment's own second
//     line) nothing is prevented, so every point is a counter;
//   - the counters go through the CR 614 counter window as a source's own
//     infect counters do (placeDamageResultCountersLocked), placed by the
//     source's controller;
//   - the tail's snapshot is never written, so a paused or staged event
//     asks again as it lands, against the board then (CR 611.3a).
//
// PHYREXIAN UNLIFE'S CONDITION. "As long as you have 0 or less life" is
// read ONCE PER DAMAGE INSTANCE (owner decision 4), against the life
// total the player had as the instance began. Its ruling: "Phyrexian
// Unlife won't affect damage that reduces your life total from a positive
// number to 0 or less. … The next time you're dealt damage, it will be
// dealt as though its source had infect." Combat damage from several
// attackers is one instance (CR 510.2), so all of it is life loss even
// when the first attacker's damage alone takes the player to 0.
// nextDamageInstanceLocked records every seat's life as it hands out an
// ID; an event with no instance (a hand-built one) reads the live total.

// DamageAsThoughStatic is one printed "damage is dealt as though its
// source had …" static on a permanent. Catalog data, never stored: the
// gate reads it off the battlefield on every landing.
type DamageAsThoughStatic struct {
	// Label is the clause as printed.
	Label string
	// Wither and Infect are what the damage is dealt as though its
	// source had. At least one is set (the effects package refuses a
	// static with neither).
	Wither bool
	Infect bool
	// ToYou narrows the static to damage dealt to its controller, the
	// player: "all damage is dealt to you". False is all damage, to
	// anything.
	ToYou bool
	// WhileAtOrBelowZeroLife is "as long as you have 0 or less life",
	// about the static's controller, read against the life total that
	// player had as the damage instance began.
	WhileAtOrBelowZeroLife bool
}

// CatalogDamageAsThough returns the "dealt as though its source had"
// statics a permanent with the given catalog key has. carddef.go sets it
// from CardDef.DamageAsThough; a game-package test may stub it.
var CatalogDamageAsThough func(key string) []DamageAsThoughStatic

// damageAsThoughStaticsOf is the statics a permanent has right now: its
// catalog entry's, keyed by CatalogAbilityKey so one that has lost its
// abilities has none (CR 613.1f).
func damageAsThoughStaticsOf(c Card) []DamageAsThoughStatic {
	if CatalogDamageAsThough == nil {
		return nil
	}
	key := CatalogAbilityKey(c)
	if key == "" {
		return nil
	}
	return CatalogDamageAsThough(key)
}

// damageResultAsThoughLocked is the result a settled damage event lands
// with: the source's own (the tail's snapshot), with every battlefield
// static that covers the event ORed in. A phased-out permanent is not in
// the battlefield slice (CR 702.26b), and a departed player's permanents
// left with them (CR 800.4a).
//
// Caller must hold g.mu. Reads only.
func (g *Game) damageResultAsThoughLocked(ev *ReplacementEvent, t *damageTail) DamageResultSource {
	res := t.result
	if CatalogDamageAsThough == nil || g.Battlefield == nil || ev == nil || res.Infect {
		// A source with infect already has every result either static
		// could add (CR 120.3b, 120.3d).
		return res
	}
	toPlayer := t.kind == damageTailPlayer
	for i := range g.Battlefield.Cards {
		src := &g.Battlefield.Cards[i]
		for _, s := range damageAsThoughStaticsOf(*src) {
			if s.ToYou && (!toPlayer || ev.DamageTarget != src.Controller) {
				continue
			}
			if s.WhileAtOrBelowZeroLife && !g.atOrBelowZeroLifeAsInstanceBeganLocked(ev.DamageInstance, src.Controller) {
				continue
			}
			res.Wither = res.Wither || s.Wither
			res.Infect = res.Infect || s.Infect
		}
	}
	return res
}

// maxDamageInstanceLives bounds how many instances' life records are
// kept. An instance's events all land while it is among the newest few:
// a combat damage step has no explicit end, and a paused event resumes
// before anything else can deal damage. A record that has aged out reads
// the live total, which is what a restored game does too.
const maxDamageInstanceLives = 16

// damageInstanceLife is every seat's life total as a damage instance
// began. Immutable once recorded, so a clone may share it.
type damageInstanceLife struct {
	inst  DamageInstance
	lives []seatLife
}

// seatLife is one player's life total at a moment.
type seatLife struct {
	player uuid.UUID
	life   int
}

// recordDamageInstanceLifeLocked records every seat's life total as
// `inst` begins. Copy on write: the new slice never shares a backing
// array with one a clone may hold.
//
// Caller must hold g.mu (write).
func (g *Game) recordDamageInstanceLifeLocked(inst DamageInstance) {
	rec := damageInstanceLife{inst: inst, lives: make([]seatLife, 0, len(g.Seats))}
	for _, p := range g.Seats {
		if p != nil {
			rec.lives = append(rec.lives, seatLife{player: p.ID, life: p.Life})
		}
	}
	keep := g.damageInstanceLives
	if len(keep) >= maxDamageInstanceLives {
		keep = keep[len(keep)-maxDamageInstanceLives+1:]
	}
	next := make([]damageInstanceLife, 0, len(keep)+1)
	next = append(next, keep...)
	g.damageInstanceLives = append(next, rec)
}

// atOrBelowZeroLifeAsInstanceBeganLocked reports whether `player` had 0
// or less life as `inst` began, or has now when the instance is zero or
// its record is gone (a restored game: the record is transient, like the
// stamp it is keyed on). A player who is not seated has no life total and
// answers false.
//
// Caller must hold g.mu.
func (g *Game) atOrBelowZeroLifeAsInstanceBeganLocked(inst DamageInstance, player uuid.UUID) bool {
	if inst != 0 {
		for i := len(g.damageInstanceLives) - 1; i >= 0; i-- {
			rec := g.damageInstanceLives[i]
			if rec.inst != inst {
				continue
			}
			for _, sl := range rec.lives {
				if sl.player == player {
					return sl.life <= 0
				}
			}
			break
		}
	}
	p := g.playerByIDLocked(player)
	return p != nil && p.Life <= 0
}
