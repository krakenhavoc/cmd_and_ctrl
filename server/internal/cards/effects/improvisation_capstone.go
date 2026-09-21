package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Improvisation Capstone — Sorcery — Lesson {5}{R}{R}:
//
//	"Exile cards from the top of your library until you exile cards
//	 with total mana value 4 or greater. You may cast any number of
//	 spells from among them without paying their mana costs.
//	 Paradigm (Then exile this spell. After you first resolve a spell
//	 with this name, you may cast a copy of it from exile without
//	 paying its mana cost at the beginning of each of your first main
//	 phases.)"
//
// The body ships whole, in two halves.
//
// # The depth is a running total, not a count
//
// MillToZone{To: ZoneExile} with an Until predicate is exactly the
// printed sentence: Until ends the run after the first card it
// answers true for, so a closure that adds each card's mana value and
// asks "am I at 4 yet" stops on the card that carries the total over
// the line, and that card is exiled too. N is left at 0, which with
// Until set means "no limit but the library" — a library that cannot
// reach 4 exiles what it has, because "until" is satisfied by
// exhaustion, and nobody loses for it (only a draw from an empty
// library does, CR 104.3c).
//
// The IDs come back through the continuation rather than being
// guessed at from a pre-count of the library. That is the difference
// CR 400.7 cares about: a card a replacement sent somewhere else —
// a commander taking the command zone, "if a card would be put into
// exile, instead …" — never reached exile, so it is not one of the
// cards the second sentence is about, and it is not granted.
//
// THE RUNNING TOTAL DOES NOT GET THAT RIGHT, and the second caveat
// below says so (#1158). `Until` is answered in millPlanLocked
// against the cards that come OFF the library, before the CR 614
// window has said where any of them went, which is what lets the plan
// be a flat list of IDs the batch body can proceed around. So the
// diverted commander's mana value still counts toward the 4 and can
// end the run one card short — the same limitation Helm of Obedience
// declares in its own words, on the same line of the same helper. The
// SECOND sentence stays right either way, because it reads the landed
// list; it is only the FIRST sentence's arithmetic that is generous.
//
// # The permission is per-object, and exactly what is printed
//
// Each exiled card gets its own ADR 0066 grant, stamped against that
// card OBJECT at its current epoch. That is what "any number of
// spells from among them" means: the grants are independent, casting
// one does not touch the others, and a card that leaves exile and
// comes back is a new object with nothing granted.
//
// Three fields carry the printed clause and no more than it:
//
//	Cost "{0}"   "without paying their mana costs". Empty would mean
//	             "pay the printed cost", which is the opposite.
//	CastOnly     "cast any number of SPELLS" — a land among them is
//	             stranded, because playing a land is not casting
//	             (CR 305.1, 116.2a).
//	Indefinite   the card states no window, and CR 611.2a is that an
//	             effect with no stated duration lasts indefinitely. It
//	             would be an invention to stamp an end-of-turn expiry
//	             the card does not print. The grant is still swept the
//	             moment its object is no longer in exile wearing the
//	             epoch it was granted at, so nothing here can leave a
//	             card castable forever.
//
// No Timing is granted, so ordinary timing rules still apply on top:
// the card says you MAY cast, not when, and a sorcery among them
// waits for a main phase.
//
// There is no "grant instead of an inline cast" deviation to declare
// here, the one Malcolm, Alluring Scoundrel and the madness cards
// carry. Those cards print "cast it" and the engine offers a grant
// instead; this card prints the permission itself, so a grant IS the
// printed effect.
//
// # What is missing
//
// Paradigm, for the reasons written out on Echocasting Symposium: no
// self-exile-instead-of-graveyard, no per-name resolved-one-yet
// record, and — the part with no primitive at all — a recurring offer
// to cast a COPY of the card out of exile at each of your first main
// phases. Every cast permission in the engine opens the CARD, never a
// copy of it. New seam row, opened by this batch in
// docs/engine-seams.md.
//
// Shipping without it is strictly weaker: one exile-and-cast and the
// card is done.
func init() {
	Register(Spec{
		OracleID:     "fd4f0315-567a-4cc5-bc7e-88a9d2cee910",
		Name:         "Improvisation Capstone",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Paradigm isn't implemented — the spell goes to your graveyard and never offers you the repeating copies it promises.",
			"If a card the run turns up never reaches exile — a commander whose owner takes the command zone instead (CR 903.9) — its mana value still counts toward the total of 4 and can end the run early, and it isn't one of the cards you may cast.",
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			controller, source := item.Controller, item.SourceCardID
			total := 0
			return MillToZone{
				Player: controller,
				To:     game.ZoneExile,
				Until: func(c game.Card) bool {
					total += c.ManaValue()
					return total >= improvisationCapstoneThreshold
				},
				Then: func(ctx *Context, exiled []uuid.UUID) error {
					improvisationCapstoneGrantCasts(ctx.Game, controller, source, exiled)
					return nil
				},
			}.Apply(ctx)
		},
	})
}

// improvisationCapstoneThreshold is the printed "total mana value 4
// or greater".
const improvisationCapstoneThreshold = 4

// improvisationCapstoneGrantCasts is the second sentence: one grant
// per card that actually reached exile, each naming that one object.
//
// Captures nothing but IDs, so an undo replays it against the
// restored game.
func improvisationCapstoneGrantCasts(g *game.Game, controller, source uuid.UUID, exiled []uuid.UUID) {
	for _, id := range exiled {
		g.GrantCastPermissionOverCardForEffect(id, game.CastPermission{
			Player: controller,
			// Named rather than filled in from wherever the card
			// landed: a card that is not in exile is not one of
			// "them", and the grant should be refused rather than
			// follow it.
			Zone:     game.ZoneExile,
			Cost:     "{0}",
			CastOnly: true,
			Duration: game.IndefiniteDuration(),
			Source:   source,
			Label:    "Improvisation Capstone — cast it without paying its mana cost",
		})
	}
}
