package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mana_replacements.go — the CR 614 replacement on the MANA PRODUCED:
// "if you tap a permanent for mana, it produces twice as much of that
// mana instead" (Mana Reflection) and the same sentence with three
// (Nyxbloom Ancient). Its own file rather than helpers.go, per the
// convention enters_tapped.go and mill_replacements.go set.
//
// The event is game.RepEventProduceMana (#1222), opened once per
// production and before any of it is in the pool — the way
// RepEventMill is opened once per mill instruction. It carries the
// player the mana is for, the object producing it, whether the
// production is a TAP FOR MANA (CR 106.12a), and the one field a
// replacement rewrites: the COLOURS, one entry per mana.
//
// Colours rather than a count, because the cards say "twice as much of
// THAT mana": a doubled Forest is {G}{G} and a doubled Sol Ring is
// {C}{C}{C}{C}. game.ReplacementEvent.MultiplyMana is the whole
// arithmetic, and a card in this family writes nothing else — changing
// the COLOURS is not something CR 106.12b licenses.
//
// A pipe slot is doubled at the PICK, not at the activation: a Birds of
// Paradise under Mana Reflection is one choice that mints two of the
// chosen colour, because the colour is not known until the choice is
// answered (ADR 0074 §3 fires the triggered mana abilities from the
// same place, for the same reason).

// ManaProducedBecomes is "if <Scope> taps a permanent for mana, it
// produces <Times> times as much of that mana instead" — the whole
// printed family, with the factor as a number.
//
// Times is applied to the amount the event currently carries, NOT to
// the printed one, and that is the rules' own composition: CR 616.1
// applies one replacement and then re-gathers, so a Mana Reflection and
// a Nyxbloom Ancient on one battlefield are ×6. The order between them
// is not put to anybody — a production cannot pause (CR 605.3b), so the
// engine applies the gathered order inline — and it does not matter,
// because multiplication commutes. See ADR 0013 §5ab.
//
// Two copies of the SAME card do not prompt either, and would not even
// if productions could pause: they are two objects contributing one
// declared effect, and #792's identical-window skip applies.
type ManaProducedBecomes struct {
	// Times is the factor. Two for Mana Reflection, three for Nyxbloom
	// Ancient. One is a no-op and is refused by the constructors
	// rather than registered as a replacement that does nothing.
	Times int

	// Scope narrows WHOSE tap this watches. Both printed cards say
	// "if YOU tap", which is ManaTapsByController; the zero value is
	// every player's, so a symmetrical printing is the field left out.
	Scope ManaProducerScope

	Label string
}

// ManaProducerScope narrows whose tap-for-mana a replacement of this
// family watches.
type ManaProducerScope string

const (
	// ManaTapsByAnyone is the zero value: every player's tap for mana,
	// the source's controller included. Nothing prints it yet — it is
	// the honest zero rather than a silent narrowing.
	ManaTapsByAnyone ManaProducerScope = ""

	// ManaTapsByController is "if YOU tap a permanent for mana" — Mana
	// Reflection, Nyxbloom Ancient.
	ManaTapsByController ManaProducerScope = "controller"

	// ManaTapsByOpponents is "if an OPPONENT taps". No printed card,
	// but it is the third of three and costs one arm.
	ManaTapsByOpponents ManaProducerScope = "opponents"
)

// Build turns the description into the ReplacementEffect a Spec
// declares.
func (m ManaProducedBecomes) Build() game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventManaAdded},
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
			if ev.Kind != game.RepEventProduceMana || len(ev.ManaColors) == 0 {
				// Already replaced away, or never a production of
				// anything: "one or more mana" is the printed
				// condition and there is nothing left to multiply.
				return false
			}
			if !ev.ManaFromTap {
				// "If you TAP a permanent for mana" (CR 106.12a). A
				// spell's "Add {B}{B}{B}", an untapped mana ability and
				// a triggered mana ability's own output are all
				// productions, and none of them is a tap — Dark Ritual
				// and Wild Growth's extra {G} are not doubled.
				return false
			}
			return m.scopeMatches(ev.ManaPlayer, src)
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.MultiplyMana(m.Times)
			return nil
		},
		// The EFFECT's controller. It happens to be the affected player
		// for both printed cards ("if YOU tap"), but CR 616.1 reads the
		// affected player off the event, so the two stay separate here
		// the way they do for every other amount replacement.
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: m.Label,
	}
}

// scopeMatches answers "is this a production this card cares about?".
//
// Anyone-else rather than a seat-by-seat opponent check, the posture
// MillBecomes and OpponentsSpell take: the game has no teams, so every
// other player at a Commander table is an opponent, and a source with
// no controller (a fixture, a token mid-construction) matches nobody
// rather than everybody.
func (m ManaProducedBecomes) scopeMatches(tapper uuid.UUID, src *game.Card) bool {
	if src == nil || src.Controller == uuid.Nil {
		return false
	}
	switch m.Scope {
	case ManaTapsByOpponents:
		return tapper != src.Controller
	case ManaTapsByController:
		return tapper == src.Controller
	}
	return true
}

// YouTapForTwiceAsMuchMana is "if you tap a permanent for mana, it
// produces twice as much of that mana instead" — Mana Reflection.
func YouTapForTwiceAsMuchMana(label string) game.ReplacementEffect {
	return ManaProducedBecomes{Times: 2, Scope: ManaTapsByController, Label: label}.Build()
}

// YouTapForThriceAsMuchMana is the same sentence with three — Nyxbloom
// Ancient.
func YouTapForThriceAsMuchMana(label string) game.ReplacementEffect {
	return ManaProducedBecomes{Times: 3, Scope: ManaTapsByController, Label: label}.Build()
}
