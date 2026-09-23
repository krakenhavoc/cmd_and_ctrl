package game

import "github.com/google/uuid"

// storm.go — the engine half of CR 702.40, which is one read and one
// announcement. See ADR 0086.
//
// Storm is unusual among keywords in how little engine it needs. The
// copies are CR 707.10 and already exist (spell_copy.go); the trigger
// is CR 603 and already exists (triggers.go, FromStack, the shape
// cascade uses); the catalog declaration is one constructor
// (cards/effects/storm.go). What was missing was the NUMBER — "each
// other spell that was cast before it this turn" — and that is
// TurnTally.Casts plus the function below.
//
// The catalog side is deliberately thin for the same reason cascade's
// is: a card with printed storm should be its printed effect and one
// line, and every decision about what the count means belongs here,
// where the rules citations are, rather than in five card files.

// StormCountForEffect is CR 702.40a's number, taken as the storm
// trigger RESOLVES: how many other spells were cast this turn, by any
// player, before `spellID` was cast.
//
// It also emits EventStorm, which is the line the public log narrates
// ("Grapeshot — storm count 3"). A read with an announcement attached,
// deliberately and in that order: the number and the telling of the
// number are one moment, and splitting them into two calls is how a
// future caller ends up copying N times and announcing M. Every
// caller wants both; there is no reading of this that is not about to
// make copies.
//
// `caster` is the player the copies belong to (CR 707.10b), which for
// storm is always the player who cast the spell — the trigger's own
// controller. It is passed rather than looked up because by the time
// the trigger resolves the spell may be in a graveyard, and the
// trigger has carried the right answer since it was built.
//
// Zero is an ordinary answer, not a failure: the turn's first spell
// has a storm count of zero and makes no copies. It is also what a
// spell this turn's casts do not name answers — see
// SpellsCastBeforeThisTurn.
//
// Caller must hold g.mu.
func (g *Game) StormCountForEffect(caster, spellID uuid.UUID) int {
	n := g.SpellsCastBeforeThisTurn(spellID)
	g.EmitEvent(Event{
		Kind:   EventStorm,
		Actor:  caster,
		Source: spellID,
		CardID: spellID,
		Amount: n,
	})
	return n
}
