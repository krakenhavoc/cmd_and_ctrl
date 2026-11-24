package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tortured Existence — Enchantment {B} (Secret Lair Drop):
//
//	"{B}, Discard a creature card: Return target creature card from your
//	 graveyard to your hand."
//
// Selhoff Entomber's "discard a creature card" cost (named as the ability is
// activated, CR 602.2b) on the graveyard-to-hand return of Midnight
// Scavengers. Targets are chosen before costs are paid (CR 602.2b), so the
// card discarded as the cost is not a legal target of its own ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fa9dfcf7-069c-4441-9ea0-133de47ecada",
		Name:         "Tortured Existence",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}, Discard a creature card: Return target creature card from your graveyard to your hand.",
			Purpose: game.Purpose{Discards: 1},
			Cost:    effects.Plus(effects.ManaCost("{B}"), effects.DiscardCardsMatching(1, "a creature card", effects.MatchCreature)),
			Targets: effects.TargetCardInGraveyard("target creature card from your graveyard", effects.Creature(), effects.YouOwn()),
			Effect:  effects.ReturnTargetedCardToHand,
		}},
	})
}
