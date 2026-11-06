package game

// omen.go — CR 720, Omen cards (#5175).
//
// An omen card is the Adventure's shape with a different exit. It is
// ONE card (CR 720.2c) with two printed faces: a permanent (face 0) and
// an instant or sorcery whose spell type is Omen (face 1). Scryfall
// lays it out as `adventure`, so everything about CASTING it is
// already the adventure card's:
//
//	CR 720.3   the caster chooses the permanent or the Omen. That is
//	           CastableFaces' adventure branch, the one gate the engine
//	           (faceForCastLocked), the view (stampCastableFaces) and the
//	           bot enumerator (legal/cast.go) all walk, so the three
//	           cannot disagree about which halves are offered.
//	CR 720.3a  only the Omen's characteristics are evaluated to see
//	           whether it can be cast, and
//	CR 720.3b  on the stack it has only those: the ADR 0034 face
//	           choice materialises face 1, as it does for an Adventure.
//	CR 720.3c  a copy of an Omen spell is an Omen too, and like every
//	           copy it ceases to exist rather than going anywhere
//	           (CR 707.10a), so nothing below sees it.
//	CR 720.4   everywhere but the stack the card is its permanent half,
//	           mana value included: SetFace(0) as it leaves the stack.
//
// What is the Omen's own is the one rule about where the card goes:
//
//	CR 720.3d  "As an Omen spell resolves, its controller shuffles it
//	           into its owner's library instead of putting it into its
//	           owner's graveyard as it resolves."
//
// That is one more case of the resolution exits (resolution_exits.go),
// beside the Adventure's exile (CR 715.3d) it replaces for this card:
// the same event ("put it into its owner's graveyard as it resolves"),
// so a granted rebound or buyback competes with it under CR 616.1, and
// only on a resolution, so a countered or fizzled Omen is an ordinary
// card in an ordinary graveyard. The exit routes the card to its
// owner's library, and its tail shuffles that library once the route is
// done. The tail runs from every terminal outcome, including one where
// the card went somewhere else (an Omen commander its owner sent to
// the command zone, CR 903.9b), and that is CR 701.24c: "that library
// is shuffled even if … an effect causes all of those objects to be
// moved to another zone."
//
// There is no cast permission and no grant: an Omen that resolved is a
// card in a library, and nothing lets anyone cast it from there.

// omenSubtype is CR 205.3k's spell type that marks an Omen half.
const omenSubtype = "Omen"

// castAsOmen reports whether `c` is an omen card on the stack as its
// OMEN half: the only spell CR 720.3d shuffles away.
//
// Read off the materialised face's type line, the one place the dump
// says it (CR 720.1 gives the frame no other marker, and Scryfall's
// layout is `adventure`). Like the Adventure, a value read off the card
// rather than off the stack item: an omen card cast as its permanent
// half is an ordinary permanent spell.
func castAsOmen(c Card) bool {
	return adventureSpellHalfHasSubtype(c, omenSubtype)
}
