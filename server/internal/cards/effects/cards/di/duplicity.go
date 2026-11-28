package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Duplicity — Enchantment {3}{U}{U} (Tempest, S82 OLD, #5628):
//
//	"When this enchantment enters, exile the top five cards of your library
//	 face down.
//	 At the beginning of your upkeep, you may exile all cards from your hand
//	 face down. If you do, put all other cards you own exiled with this
//	 enchantment into your hand.
//	 At the beginning of your end step, discard a card.
//	 When you lose control of this enchantment, put all cards exiled with this
//	 enchantment into their owner's graveyard."
//
// Four linked abilities over one pile. Every exile is face down with nobody
// allowed to look (CR 406.3) and linked to this object (CR 607.2a), the way
// Asmodeus the Archfiend's is. The upkeep "other cards" are read before the hand
// goes, so the hand just exiled stays; only cards you own come back. The end
// step discard is the controller's pick. "You lose control" is both ways it
// happens, another player taking the enchantment and its leaving the
// battlefield, as Khârn the Betrayer's is, and the cards are found through the
// object's last-known record.
//
// No simplification.
func init() {
	const name = "Duplicity"
	effects.Register(effects.Spec{
		OracleID:     "01072a45-ab1c-4d19-b493-76cf4aae351e",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(name+" — exile the top five cards of your library face down", effects.S82OldB761ExileTopFiveFaceDownLinked),
			effects.Optional(effects.AtYourUpkeep(name+" — you may exile your hand face down; if you do, put the other exiled cards into your hand", effects.S82OldB761DuplicityUpkeep),
				"Exile all cards from your hand face down?"),
			effects.AtYourEndStep(name+" — discard a card", effects.S82OldB761DiscardACard),
			effects.WhenYouLoseControlOfThis(name+" — put all cards exiled with this enchantment into their owner's graveyard", effects.S82OldB761ExiledWithToGraveyards),
			effects.WhenThisLeaves(name+" — put all cards exiled with this enchantment into their owner's graveyard", effects.S82OldB761ExiledWithToGraveyards),
		},
	})
}
