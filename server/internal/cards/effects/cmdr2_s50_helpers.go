package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_s50_helpers.go — helpers private to Commander wave 2, batch 50
// (#5368): Duskmourn enchantments, horrors and hideaway.

// cmdr2S50SpellOfAnyCreatureType passes on a spell with any of the given
// creature types, read as the spell sits on the stack (a changeling or a
// kindred spell counts, CR 702.73a / CR 308.3). A spell with two of the
// types is still one spell. The types are words of the source's text
// (CR 612.2).
func cmdr2S50SpellOfAnyCreatureType(subtypes ...string) CostPredicate {
	return func(q game.CostQuery) bool {
		for _, s := range subtypes {
			if q.Card.HasSubtype(q.Source.TextSubtype(s)) {
				return true
			}
		}
		return false
	}
}

// cmdr2S50ManaShortfall is how many mana a spell on the stack cost less
// than its mana value: the difference between its mana value (X included,
// CR 202.3e) and the mana spent to cast it, or zero when it cost at least
// that much (or its cost cannot be read). Neither number can change once
// the spell is cast (CR 601.2h).
func cmdr2S50ManaShortfall(g *game.Game, spellID uuid.UUID) int {
	mv, ok := g.LastKnownSpellManaValueForEffect(spellID)
	if !ok {
		return 0
	}
	if d := mv - g.StackItemPaidForEffect(spellID).ManaSpentCount(); d > 0 {
		return d
	}
	return 0
}

// cmdr2S50ShareAllCardTypes reports whether two cards have exactly the same
// card types, as printed: "two cards that share all their card types"
// (Demonic Covenant). A Kindred Enchantment shares all its types with
// another Kindred Enchantment, not with an Enchantment alone.
func cmdr2S50ShareAllCardTypes(a, b game.Card) bool {
	_, ta, _ := game.ParseTypeLine(a.TypeLine)
	_, tb, _ := game.ParseTypeLine(b.TypeLine)
	if len(ta) != len(tb) {
		return false
	}
	seen := map[string]bool{}
	for _, t := range ta {
		seen[t] = true
	}
	for _, t := range tb {
		if !seen[t] {
			return false
		}
	}
	return true
}

// cmdr2S50PlayHiddenCard is "you may play" ONE named card the source's
// hideaway exiled: grantHiddenPlay's two shapes (a land is played now with
// the turn's land drop; a spell is a {0} grant cast right after, ADR 0091)
// over a card the caller picked, for a source that hid more than one card
// (Evercoat Ursine's two hideaway abilities). The exile is the link
// (CR 607.2a), so a card that is no longer hidden by that object is ignored.
func cmdr2S50PlayHiddenCard(g *game.Game, player uuid.UUID, source game.PermissionCardRef, id uuid.UUID, label string) {
	stillHidden := false
	for _, h := range g.HiddenCardsForEffect(source) {
		if h == id {
			stillHidden = true
		}
	}
	if !stillHidden {
		return
	}
	if c, ok := g.LookupCardForEffect(id); ok && c.IsLand() {
		if !g.CanPlayLandDuringResolutionForEffect(player, id) {
			return
		}
		_ = g.QueueMayCastKeyedForEffect(game.MayCastPrompt{
			Chooser:  player,
			Source:   source.ID,
			Card:     id,
			Question: label + " — play " + c.Name + "?",
		}, hideawayPlayLand, nil)
		return
	}
	g.GrantCastPermissionOverCardForEffect(id, game.CastPermission{
		Player: player,
		Zone:   game.ZoneExile,
		Cost:   "{0}",
		Timing: game.TimingFlash,
		Source: source.ID,
		Label:  label,
	})
}
