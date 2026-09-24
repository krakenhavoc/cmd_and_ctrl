package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cast_restriction.go — S42, ADR 0073 §7: constructors for
// Spec.CastRestrictions ("players can't cast …", CR 101.2) and for
// Spec.CastCondition (a spell's own "cast this only if …").
//
// One constructor per printed SHAPE rather than a generic builder,
// for the reason the alternative-cost keywords have one: the shape
// carries the clause the client is shown along with the predicate,
// and a card file that wrote the predicate and forgot the label ships
// a refusal nobody can explain. Register refuses both halves apart.
//
// The predicate runs under g.mu inside the cast path. Read-only:
// *ForEffect accessors and plain field reads, never a locking
// mutator.

// EachPlayerMaxSpellsPerTurn is Rule of Law, Eidolon of Rhetoric,
// Archon of Emeria: "Each player can't cast more than N spells each
// turn." Counted off the engine's own per-turn tally
// (Game.CastTallyFor), which is bumped after the cast succeeds — so
// at announce a player who has already cast N has a tally of exactly
// N and this is the (N+1)th.
//
// "Each player" includes the permanent's controller: Rule of Law is
// symmetrical, and a card that meant otherwise says "your opponents".
func EachPlayerMaxSpellsPerTurn(n int, label string) game.CastRestriction {
	return game.CastRestriction{
		Label: label,
		Forbids: func(q game.CastQuery) bool {
			return q.Game.CastTallyFor(q.Controller).Total >= n
		},
	}
}

// EachPlayerMaxNoncreatureSpellsPerTurn is Deafening Silence: "Each
// player can't cast more than N noncreature spells each turn." The
// engine's tally already counts the noncreature half, so this is the
// same read one field over.
//
// The spell being announced is judged by its OWN type line, not by
// the tally — a creature spell is never refused by this clause
// however many noncreature spells came before it.
func EachPlayerMaxNoncreatureSpellsPerTurn(n int, label string) game.CastRestriction {
	return game.CastRestriction{
		Label: label,
		Forbids: func(q game.CastQuery) bool {
			if q.Card.IsCreature() {
				return false
			}
			return q.Game.CastTallyFor(q.Controller).Noncreature >= n
		},
	}
}

// PlayersCantCastFrom is Grafdigger's Cage's second sentence:
// "Players can't cast spells from graveyards or libraries."
//
// A zone predicate and nothing else — it does not care who is casting
// or what. The first sentence ("creature cards in graveyards and
// libraries can't enter the battlefield") is a different seam and a
// card declaring this must say so in its caveats.
func PlayersCantCastFrom(label string, zones ...game.ZoneKind) game.CastRestriction {
	banned := make(map[game.ZoneKind]bool, len(zones))
	for _, z := range zones {
		banned[z] = true
	}
	return game.CastRestriction{
		Label: label,
		Forbids: func(q game.CastQuery) bool {
			return banned[q.FromZone]
		},
	}
}

// YouCantCastUnless is Rakdos, Lord of Riots: "You can't cast
// creature spells unless an opponent lost life this turn."
//
// "You" is the permanent's own CONTROLLER (q.Source.Controller), not
// the player casting — which is the whole reason CastQuery carries
// both. A Rakdos an opponent has stolen restricts THEM, and that is
// the printed behaviour.
//
// `match` narrows which of that player's spells the clause covers
// (creature spells, here); nil covers every spell. `unless` is the
// condition that LIFTS the ban, so the predicate reads the way the
// card does.
func YouCantCastUnless(label string, match CardPredicate, unless func(g *game.Game, controller uuid.UUID) bool) game.CastRestriction {
	return game.CastRestriction{
		Label: label,
		Forbids: func(q game.CastQuery) bool {
			if q.Source.Controller != q.Controller {
				return false
			}
			if !matchCastCard(q.Game, match, q.Controller, q.Card) {
				return false
			}
			return unless == nil || !unless(q.Game, q.Controller)
		},
	}
}

// OpponentsCantCast is the Archon-of-Emeria / Dragonlord-Dromoka
// half: a restriction that binds every player EXCEPT the source's
// controller. Same narrowing as YouCantCastUnless, opposite seat.
func OpponentsCantCast(label string, match CardPredicate) game.CastRestriction {
	return game.CastRestriction{
		Label: label,
		Forbids: func(q game.CastQuery) bool {
			if q.Source.Controller == q.Controller {
				return false
			}
			return matchCastCard(q.Game, match, q.Controller, q.Card)
		},
	}
}

// matchCastCard runs a card predicate against a SPELL being announced
// rather than against a permanent on the battlefield. The spell is
// still in its source zone at announce, so the predicate sees printed
// characteristics — which is what every restriction in the catalog
// asks about ("creature spells", "noncreature spells").
//
// The predicate's `caster` is the player casting, so a clause could
// say "creature spells you control"; none does today.
func matchCastCard(g *game.Game, pred CardPredicate, caster uuid.UUID, card game.Card) bool {
	if pred == nil {
		return true
	}
	return pred(g, caster, card)
}

// LegendarySorcery is CR 205.4e: "You may cast a legendary sorcery
// only if you control a legendary creature or planeswalker." Urza's
// Ruinous Blast, Jaya's Immolating Inferno, The Antiquities War's
// siblings.
//
// Checked at announce and never at resolution: a legendary creature
// that dies while the sorcery is on the stack does not counter it
// (CR 205.4e restricts the cast, and nothing re-checks it).
func LegendarySorcery() func(g *game.Game, controller uuid.UUID, card game.Card) bool {
	return func(g *game.Game, controller uuid.UUID, _ game.Card) bool {
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller != controller || !c.IsLegendary() {
				continue
			}
			if c.IsCreature() || c.IsPlaneswalker() {
				return true
			}
		}
		return false
	}
}

// LegendarySorceryLabel is the clause CR 205.4e prints in the rules
// text box of every legendary sorcery, verbatim, so ten card files
// cannot spell it ten ways.
const LegendarySorceryLabel = "Cast this spell only if you control a legendary creature or planeswalker."

// OpponentsCantCastDuringYourTurn is Dragonlord Dromoka's third
// clause: "Your opponents can't cast spells during your turn."
//
// OpponentsCantCast with the turn added. Both halves are read off the
// SOURCE's controller: they are the "your" in "your opponents" and
// the "your" in "your turn", so a Dromoka an opponent has stolen
// locks the table out on THEIR turn instead — which is the printed
// card and the same reason YouCantCastUnless reads q.Source.
//
// The active seat is read off g.Turn directly rather than through
// g.ActivePlayer(), which takes an RLock: this predicate runs inside
// the cast path, which already holds g.mu for write, and a second
// acquisition there deadlocks.
func OpponentsCantCastDuringYourTurn(label string, match CardPredicate) game.CastRestriction {
	return game.CastRestriction{
		Label: label,
		Forbids: func(q game.CastQuery) bool {
			if q.Source.Controller == q.Controller {
				return false
			}
			if !isActivePlayer(q.Game, q.Source.Controller) {
				return false
			}
			return matchCastCard(q.Game, match, q.Controller, q.Card)
		},
	}
}
