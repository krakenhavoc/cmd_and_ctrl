package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Valgavoth, Terror Eater — Legendary Creature — Elder Demon
// {6}{B}{B}{B}, 9/9:
//
//	"Flying, lifelink
//	 Ward—Sacrifice three nonland permanents.
//	 If a card you didn't control would be put into an opponent's
//	 graveyard from anywhere, exile it instead.
//	 During your turn, you may play cards exiled with Valgavoth. If
//	 you cast a spell this way, pay life equal to its mana value
//	 rather than pay its mana cost."
//
// The nine-drop that eats the table's graveyards. Three of the four
// clauses ship.
//
// The WARD is the reason WardSacrificeCost grew a count: every ward
// in the catalog before this one demanded a single permanent, and the
// pick was hardcoded to one. It is one payment (CR 118.4), so a payer
// with only two nonland permanents cannot pay at all — the spell is
// countered with no prompt, rather than taking two and letting the
// removal through.
//
// The REPLACEMENT is the shared GraveyardBecomesExile with both of
// its narrowing clauses set. "Into an OPPONENT's graveyard" is
// OpponentsOnly; "a card YOU DIDN'T CONTROL" is
// NotControlledByYou, and it is not redundant — a creature you stole
// from an opponent goes to its owner's graveyard when it dies, which
// OpponentsOnly alone would exile and printed Valgavoth would not.
// Being stronger than printed is the one direction a simplification
// may never go, so the second clause is a field rather than a comment.
//
// It catches every graveyard arrival since #931 — death, mill,
// discard, a library search that bins the card, surveil — because
// they all push their move through the CR 614 window. The
// battlefield case is the one worth saying out loud: an opponent's
// creature that would DIE is exiled instead, so it never died and its
// own dies-triggers never fire (CR 700.4).
//
// The FOURTH clause is dropped. Playing a card out of exile is a
// permission that belongs to one exiled instance, and the engine
// grants it from a resolving effect; a replacement has no post-move
// hook to register one from, and the permission's filter is a closed
// set of flags with no "exiled by this permanent" member. The cards
// are still exiled, which is most of what Valgavoth is for — they are
// simply gone rather than yours.
func init() {
	Register(Spec{
		OracleID:     "cae3ec72-436d-4086-9dcb-17b3d92ad5c4",
		Name:         "Valgavoth, Terror Eater",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Cards Valgavoth exiles stay exiled — you can't play them.",
		},
		PrintedKeywords: []string{"flying", "lifelink"},
		Triggered: []game.TriggeredAbility{
			Ward(WardSacrificeN(3, "three nonland permanents", Nonland()),
				"Valgavoth, Terror Eater — ward, sacrifice three nonland permanents"),
		},
		Replacements: []game.ReplacementEffect{GraveyardBecomesExile{
			OpponentsOnly:      true,
			NotControlledByYou: true,
			Label:              "Valgavoth, Terror Eater — exile it instead",
		}.Build()},
	})
}
