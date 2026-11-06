package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Descendants' Path — Enchantment {2}{G} (Avatar: The Last Airbender
// Eternal, #5075; first printed in Avacyn Restored):
//
//	"At the beginning of your upkeep, reveal the top card of your
//	 library. If it's a creature card that shares a creature type with a
//	 creature you control, you may cast it without paying its mana cost.
//	 If you don't cast it, put it on the bottom of your library."
//
// The reveal is public. "Shares a creature type" is CR 702.73a-aware
// (game.SharesCreatureType): a changeling on either side shares with
// any creature that has a creature type, and land or artifact subtypes
// never count. The creatures you control are read through the layered
// view as the trigger resolves.
//
// The cast is made during the resolution (CR 608.2g) from the top of the
// library, through the resolution_cast prompt over a free, any-timing,
// cast-only permission (Commander Spock's offer). Without paying its
// mana cost, X is 0, and no alternative cost can be paid; additional
// costs still can be, as the rulings say. Once the prompt is answered,
// the card is put on the bottom if it is still in the library, which is
// also what happens at once to a card that is not a creature card, or
// shares no type: "if you don't cast it" covers every card not cast.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ba13cbdb-8043-4c1b-8a85-084c86886ac8",
		Name:         "Descendants' Path",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep(avB30dDescendantsPathLabel, avB30dDescendantsPathReveal),
		},
	})
}

const avB30dDescendantsPathLabel = "Descendants' Path — reveal the top card; you may cast it free if it's a creature card that shares a type"

// avB30dDescendantsPathReveal is the whole trigger.
func avB30dDescendantsPathReveal(g *game.Game, item *game.StackItem) error {
	you, source := item.Controller, item.SourceCardID
	revealed := g.RevealTopOfLibraryForEffect(you, source, 1, avB30dDescendantsPathLabel)
	if len(revealed) == 0 {
		return nil
	}
	top := revealed[0]
	if c, ok := g.LookupCardForEffect(top); ok && c.IsCreature() && avB30dSharesATypeWithYourCreature(g, you, c) {
		perm := anySpell.on(g.FreeCastFromLibraryPermission(you, source, "Descendants' Path"))
		perm.LapseOnPass = game.LapseStaysInHand
		if len(freeCastCandidates(g, &perm, revealed)) > 0 && g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
			Chooser:      you,
			Source:       source,
			Reason:       "Descendants' Path",
			Cards:        revealed,
			Permission:   &perm,
			Keyword:      game.MayCastKeywordFreeCast,
			AcceptLabel:  "Cast it free",
			DeclineLabel: "Put it on the bottom",
			Then:         lookedAtRestToBottom,
			Carry:        revealed,
		}) {
			return nil
		}
	}
	// Not cast: on the bottom.
	return g.PutOnBottomInRandomOrderForEffect(you, game.ZoneLibrary, cardsStillInALibrary(g, revealed))
}

// avB30dSharesATypeWithYourCreature reports whether card shares a
// creature type with a creature `you` control.
func avB30dSharesATypeWithYourCreature(g *game.Game, you uuid.UUID, card game.Card) bool {
	for _, p := range g.BattlefieldCardsForEffect() {
		if p.Controller != you || !p.IsCreature() {
			continue
		}
		if game.SharesCreatureType(&card, &p) {
			return true
		}
	}
	return false
}
