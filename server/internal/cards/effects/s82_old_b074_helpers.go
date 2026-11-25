package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b074_helpers.go — shared bodies for OLD batch 74 (#5506).

// S82OldB074ElfYouControlDied is "an Elf you control dies", the source
// itself included: the creature went to a graveyard from the battlefield
// and was an Elf under the source's controller as it last existed (CR
// 603.10a), so a changeling counts and a creature that was an Elf only
// through a grant does too. "Elf" is a word of the source's text.
func S82OldB074ElfYouControlDied(ev game.Event, source *game.Card, g *game.Game) bool {
	dead, ok := DiedCreature(ev, g)
	return ok && LeftUnderControlOf(ev, dead) == source.Controller && LeftAsSubtype(ev, dead, source.TextSubtype("Elf"))
}

// S82OldB074ElvesYouControl counts the Elves the controller controls, read
// through the layered view, the source itself counted when it is one.
func S82OldB074ElvesYouControl(g *game.Game, controller uuid.UUID, source *game.Card) int {
	elf := source.TextSubtype("Elf")
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.HasSubtype(elf) {
			n++
		}
	}
	return n
}

// S82OldB074ElfCardsInGraveyard counts the Elf cards in the player's
// graveyard (Abomination of Llanowar).
func S82OldB074ElfCardsInGraveyard(g *game.Game, player uuid.UUID, source *game.Card) int {
	elf := source.TextSubtype("Elf")
	return len(GraveyardIDs(g, player, func(c game.Card) bool { return c.HasSubtype(elf) }))
}

// S82OldB074RevealTopTakeAllToHandRestBottom is "Reveal the top N cards of
// your library. Put all <kind> cards revealed this way into your hand and
// the rest on the bottom of your library in any order" (Sylvan Messenger):
// Mulch's reveal-and-take with the rest going to the bottom, the order the
// player's to choose.
func S82OldB074RevealTopTakeAllToHandRestBottom(g *game.Game, player, source uuid.UUID, n int, label string, match CardPredicate) error {
	revealed := g.RevealTopOfLibraryForEffect(player, source, n, label)
	return TakeFromLibraryToHand{
		Player: player,
		Cards:  revealed,
		Match:  match,
		All:    true,
		Label:  label,
		Key:    TakeRestBottomAny,
	}.applyKeyed(NewContext(g, nil), player, source)
}
