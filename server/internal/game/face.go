package game

import "github.com/google/uuid"

// face.go — the multi-face card model (ADR 0034).
//
// Scryfall ships two structurally different "multi-face" shapes and
// conflating them is why this looked like one problem:
//
//	transform / modal_dfc   top-level mana_cost, colors and image_uris
//	                        are all NULL; the real data lives on each
//	                        card_faces entry.
//	adventure / split /     top-level mana_cost is the JOINED string
//	prepare                 "{4}{U} // {1}{U}", which ParseCost rejects.
//
// Either way the top-level type_line is the concatenation
// ("Sorcery // Land"), and game.Card used to copy exactly those
// top-level fields. The consequences, all silent:
//
//   - 1,503 byName keys imported with ManaCost "" — ParseCost("")
//     succeeds and yields the zero cost, so every transform and modal
//     DFC in the game was castable for FREE.
//   - Card.IsLand() is a substring scan over the whole type line, so
//     "Sorcery // Land" matched "land" and CastSpell's land branch
//     put Sea Gate Restoration straight onto the battlefield, before
//     the cost gate ever ran (#289, #265).
//   - ParseTypeLine whitespace-tokenises the line, so every DFC
//     carried the literal types "//" and "—" — 501 oracle IDs whose
//     subtype matching was quietly wrong.
//   - Aang, Swift Savior imported with no cost and no colours, which
//     is what blocked #343 / #325.
//
// The fix is representational: Card grows the printed faces and an
// index naming which one is up, and the existing flat printed fields
// become the MATERIALISATION of the active face rather than a copy of
// Scryfall's top-level record. Everything that already reads
// c.TypeLine / c.ManaCost / c.Power keeps compiling and starts being
// right, because "the characteristics of the face that is currently
// up" is exactly what CR 712.8 says a double-faced permanent has.

// Face is one printed side of a multi-face card, in engine terms.
// Values arrive already parsed — the deck importer does the Scryfall
// string work once, at import, so nothing downstream has to know that
// power is a string on the wire.
type Face struct {
	// Name is the face's own printed name ("Sea Gate Restoration"),
	// never the composite "A // B".
	Name string

	// TypeLine is the FACE's type line — "Sorcery", never
	// "Sorcery // Land". This is the field that makes IsLand(),
	// ParseTypeLine and every client-side type check correct
	// without any of them changing.
	TypeLine string

	// ManaCost is the face's printed cost. Scryfall puts a real
	// cost here for every face-carrying layout, including the ones
	// whose top-level cost is null.
	ManaCost string

	// Colors is the face's colours, resolved at import in the order
	// colors → color_indicator → derived from mana_cost, because
	// the two Scryfall shapes populate different ones (a transform
	// back face has `colors` but no cost; a split face has a cost
	// but null `colors`).
	Colors []string

	// Power and Toughness are the face's printed stats, already
	// parsed. Zero for non-creatures and for non-numeric values
	// ("*"), matching the top-level import's posture.
	Power     int
	Toughness int

	// VariableToughness is Card.VariableToughness for this face: the
	// printed toughness is not a number the engine can use — "*", or
	// missing from a creature face's record — and Toughness is its 0
	// stand-in. SetFace materialises it onto the card.
	VariableToughness bool

	// StartingLoyalty is the face's printed loyalty (CR 306.5b).
	// Zero for every non-planeswalker face — which is the point:
	// Nissa, Vastwood Seer's FRONT face is a creature and must not
	// inherit the back face's loyalty, as the old whole-card
	// fallback made it do.
	StartingLoyalty int

	// StartingDefense is the face's printed defense (CR 310.4).
	// Zero for every non-battle face — and for a battle it lives on
	// the face rather than at the top level, because that is where
	// Scryfall puts it: every printed battle is a `transform` card
	// whose top-level `defense` is null and whose FRONT face carries
	// the number. A top-level-only read would have given every
	// battle in the game zero defense. Added in S27.
	StartingDefense int

	// OracleText is the face's rules text. Not read by the engine
	// (game.Card has never carried oracle text) but carried for the
	// wire so the client's hover overlay can show the back face
	// without a Scryfall round-trip.
	OracleText string
}

// Multi-face layouts, as Scryfall spells them. Only the ones the
// engine branches on are named; every other layout string is carried
// verbatim on Card.Layout and treated as single-faced by
// faceCastable / faceOnResolve.
const (
	// LayoutModalDFC is a modal double-faced card: the two faces are
	// INDEPENDENTLY playable and the choice is made at announce
	// (CR 712.12). 100 oracle IDs, 60 of them with a land back.
	LayoutModalDFC = "modal_dfc"

	// LayoutTransform is an ordinary double-faced card. It is always
	// cast as its front face (CR 712.11); the back face is reached
	// only by an effect that transforms the permanent in place.
	// 401 oracle IDs.
	LayoutTransform = "transform"

	// LayoutAdventure is a creature with an instant/sorcery half
	// (CR 715). The face choice picks which SPELL is cast, but the
	// permanent that ends up on the battlefield is always face 0 —
	// the Adventure half is an instant or sorcery and never becomes
	// one. See adventure.go for the rest of the lifecycle: the
	// Adventure spell exiles as it resolves and its owner may cast
	// the creature from exile for as long as it stays there.
	LayoutAdventure = "adventure"

	// LayoutSplit and LayoutPrepare carry a joined top-level cost.
	// The spine makes them cost their LEFT half instead of being
	// free. For a split card that is still a simplification (fusing);
	// for a preparation card it is the rule — CR 722.3, the card is
	// only ever cast as its permanent half, and its prepare spell
	// (face 1) is cast as a COPY out of exile while the permanent is
	// prepared (CR 722.3c, ADR 0090, prepare.go).
	LayoutSplit   = "split"
	LayoutPrepare = "prepare"
)

// SetFace switches the card to face i and re-materialises the flat
// printed fields from it. It is the ONLY writer of ActiveFace: the
// invariant that Name / TypeLine / ManaCost / Colors / Power /
// Toughness / StartingLoyalty / StartingDefense equal
// Faces[ActiveFace] is maintained here and nowhere else.
//
// A no-op for single-faced cards (Faces == nil), which is every one
// of the ~33,000 ordinary oracle IDs — they keep today's behaviour to
// the byte. An out-of-range index is clamped to face 0 rather than
// panicking: the index reaches here from a client-supplied cast
// parameter, and refusing the cast is the action layer's job (see
// validateFaceChoice), not this setter's.
//
// Callers switching a face on the BATTLEFIELD must follow with a
// layer-staleness mark — a face change is a printed-value change and
// the cached Effective() characteristic is stale the moment it lands.
func (c *Card) SetFace(i int) {
	if len(c.Faces) == 0 {
		return
	}
	if i < 0 || i >= len(c.Faces) {
		i = 0
	}
	f := c.Faces[i]
	c.ActiveFace = i
	c.Name = f.Name
	c.TypeLine = f.TypeLine
	c.ManaCost = f.ManaCost
	c.Colors = append([]string(nil), f.Colors...)
	c.Power = f.Power
	c.Toughness = f.Toughness
	c.VariableToughness = f.VariableToughness
	c.StartingLoyalty = f.StartingLoyalty
	c.StartingDefense = f.StartingDefense
}

// ColorsInManaCost returns the unique WUBRG letters in a printed
// mana-cost string, hybrid symbols contributing both halves.
//
// Exported wrapper over printedColorsFromCost so the deck importer
// can derive a face's colours from its cost — the last resort when a
// face carries neither `colors` nor a `color_indicator`, which is
// every adventure and split face.
func ColorsInManaCost(cost string) []string {
	return printedColorsFromCost(cost)
}

// IsMultiFace reports whether the card has more than one printed
// face the engine knows about.
func (c Card) IsMultiFace() bool { return len(c.Faces) > 1 }

// FaceCount returns the number of printed faces, or 1 for a
// single-faced card. Never zero, so callers can index-bound against
// it without a special case.
func (c Card) FaceCount() int {
	if len(c.Faces) == 0 {
		return 1
	}
	return len(c.Faces)
}

// CastableFaces returns the face indices a player may choose between
// when casting or playing this card from hand.
//
// This is where the per-layout semantics of "choose a face" live:
//
//	modal_dfc          both — the faces are independently playable
//	                   (CR 712.11b), and this is the whole reason
//	                   the picker exists.
//	adventure          both — CR 715.3 lets the caster choose whether
//	                   to cast the creature or the Adventure, and the
//	                   choice is made wherever the cast is legal from.
//	                   Which HALF a granted cast opens is the
//	                   permission's business, not the card's:
//	                   CR 715.4's exile grant names the creature face
//	                   and an impulse grant over the same card names
//	                   none, so a zone rule here would be wrong for
//	                   one of them.
//	transform          front only (CR 712.11). The back is reached by
//	                   transforming the permanent, not by casting it.
//	split              front only for now — split needs fusing.
//	                   Deferred.
//	prepare            front only, and that is CR 722.3 rather than
//	                   a deferral: the prepare spell is never cast
//	                   from the card. Its copy in exile is, through
//	                   the permission prepare.go derives, which names
//	                   face 1 (ADR 0090).
//	anything else      front only.
func (c Card) CastableFaces() []int {
	if len(c.Faces) < 2 {
		return []int{0}
	}
	if c.Layout == LayoutModalDFC || c.Layout == LayoutAdventure {
		out := make([]int, len(c.Faces))
		for i := range c.Faces {
			out[i] = i
		}
		return out
	}
	return []int{0}
}

// CastableFacesUnder is CastableFaces narrowed by a permission that
// names faces: the halves a cast of this card BY THIS PLAYER may
// actually choose between, right now.
//
// Two rules, and they are faceForCastLocked's, read from the other
// end. The card's own layout decides when the grant says nothing —
// every cast from hand, the graveyard and the command zone — and a
// grant that NAMES faces opens those and no other: a defeated Siege's
// back, CR 715.4's "cast the creature from exile".
//
// The legal-move enumerator and the view both walk this (legal/cast.go,
// protocol.stampCastableFaces) rather than each spelling the pair out,
// because a face one offers and the other does not is either a bot
// move the announce path refuses or a picker row with nothing behind
// it — and until #992 the view did not walk faces at all, which is
// the second of those.
func CastableFacesUnder(c Card, perm *CastPermission, player uuid.UUID) []int {
	if faces, ok := perm.GrantsFaces(player); ok {
		return faces
	}
	return c.CastableFaces()
}

// faceCastable reports whether face i is a legal announce-time choice
// for this card. Face 0 is always legal — that is the single-faced
// case and the default for every parameter that arrives unset.
func faceCastable(c Card, i int) bool {
	if i == 0 {
		return true
	}
	for _, f := range c.CastableFaces() {
		if f == i {
			return true
		}
	}
	return false
}

// faceForCastLocked settles which face a cast announces, given the
// face the caller asked for and any granted cast permission covering
// the card (S32; ADR 0066 made the permission the one model).
// Returns the face and whether the cast is allowed at all.
//
// `grant` is whatever CastPermissionForLocked answered, so it is
// already this player's and already live — which is why #945 could
// take the turn out of the signature rather than re-deriving the
// window here.
//
// Two rules, and the split between them is the whole seam:
//
//  1. NO grant names a face — every cast before S32, and every cast
//     from hand, the graveyard or the command zone today. The card's
//     own CastableFaces decides, exactly as it always has, and a face
//     it does not offer is ErrInvalidFace.
//
//  2. A grant NAMES faces. Then those are the only ones this
//     permission opens; when it names exactly ONE — which is every
//     grant anything declares today — that face is also the ANSWER
//     rather than a thing to check the request against, and the
//     caller's `want` is ignored.
//
// The second half of (2) is the deliberate part, and it is the one
// place this departs from ADR 0034's "reject, never clamp" rule.
// That rule exists because a modal DFC OFFERS a choice and silently
// casting the wrong half of a choice is the worst available failure.
// A single-face grant offers no choice: there is exactly one legal
// cast of that card by that player, so there is nothing to mis-pick
// and nothing a stricter reading would protect. Rejecting instead
// would mean every caller — the client's exile pile, the zone
// browser, a future enumerator entry — had to re-derive the one
// possible answer and spell it back, and each of them forgetting is a
// cast that fails for no reason a player can see. A grant naming
// SEVERAL faces is back to being a choice, so it narrows `want`
// instead of answering for it.
func faceForCastLocked(c Card, want int, grant *CastPermission, playerID uuid.UUID) (int, bool) {
	if faces, ok := grant.GrantsFaces(playerID); ok {
		// A grant for a face the card does not have is REFUSED, not
		// clamped. SetFace clamps, by design, so a card is never left
		// incoherent — but here the clamp would land on face 0, and
		// face 0 of a card whose back face was granted is the half
		// the grant exists to exclude. A defeated Siege whose back
		// face never imported would become a free cast of the battle
		// itself. Refusing costs the player a card they were owed;
		// clamping hands them one they were not.
		if len(faces) == 1 {
			face := faces[0]
			if face < 0 || face >= c.FaceCount() {
				return 0, false
			}
			return face, true
		}
		for _, face := range faces {
			if face == want && want >= 0 && want < c.FaceCount() {
				return want, true
			}
		}
		return 0, false
	}
	if !faceCastable(c, want) {
		return 0, false
	}
	return want, true
}

// faceOnResolve returns the face the PERMANENT keeps, given the face
// that was cast.
//
// An MDFC keeps what you chose: cast Sea Gate Restoration's back and
// a land called Sea Gate, Reborn is what reaches the battlefield, and
// the front face never returns.
//
// A `transform` card keeps it too, and that is the S32 half of the
// battle seam. The rule it was written to enforce — "a transform card
// always enters front-up (CR 712.11)" — is really a rule about CASTING
// and it is already enforced where it belongs, by CastableFaces
// refusing to offer the back. So face 1 can only ever arrive here
// through an effect that said "cast it TRANSFORMED", and for such a
// cast CR 712.11 does not apply: the object that was put on the stack
// was the back face and the permanent it resolves into is the back
// face. Returning 0 here was what made the Siege seam unfixable from
// the grant side alone — the cast would have announced Refraction
// Elemental and resolved into Invasion of Karsus, a battle
// re-entering the battlefield, which is worse than not casting it.
//
// For every ordinary transform cast castFace is 0, so this is a no-op
// for all 401 transform oracle IDs until a grant opens a back face.
//
// Everything else resolves as its front face — an adventure's
// creature half is what becomes a permanent no matter which half was
// cast (and its spell half never reaches this branch at all, being an
// instant or sorcery).
func faceOnResolve(layout string, castFace int) int {
	switch layout {
	case LayoutModalDFC, LayoutTransform:
		return castFace
	}
	return 0
}

// setFaceInZoneLocked re-materialises the chosen face on the card as
// it sits in a zone, rather than on a local copy.
//
// CastSpell works from a value copy of the card, which is exactly
// what makes one SetFace call cover the whole announce sequence — but
// a copy is invisible to anything that looks the card up by ID. Two
// paths do: the CR 614 replacement pipeline, which resolves the
// entering card through LookupCardForEffect to find its own
// self-replacement (that is how an MDFC land back finds its
// "enters tapped unless you pay 3 life" clause), and the entry-choice
// resume, which re-enters after the prompt with only the card ID.
// Both read the card in its SOURCE zone, so the face has to be
// written there before the pipeline runs.
//
// Returns the face that was previously active so a failed cast can
// put it back.
//
// Caller must hold g.mu.
func setFaceInZoneLocked(z *Zone, cardID uuid.UUID, face int) int {
	if z == nil {
		return 0
	}
	for i := range z.Cards {
		if z.Cards[i].InstanceID != cardID {
			continue
		}
		was := z.Cards[i].ActiveFace
		z.Cards[i].SetFace(face)
		return was
	}
	return 0
}
