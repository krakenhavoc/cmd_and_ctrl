package cardsso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Squirming Emergence — Sorcery {1}{B}{G} (The Lost Caverns of Ixalan,
// #4470):
//
//	"Fathomless descent — Return to the battlefield target nonland
//	 permanent card in your graveyard with mana value less than or equal
//	 to the number of permanent cards in your graveyard."
//
// The mana-value cap is a live target predicate (Betor's shape,
// manaValueWithinLifeSwing) over the descend count
// (permanentCardsInGraveyard), so the card being targeted counts itself.
// It is checked as the target is chosen and again as the spell resolves
// (CR 608.2b): a card that no longer fits returns nothing. The return is
// the shared reanimate body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e468f1fe-d99c-4034-a94f-91c41a129092",
		Name:         "Squirming Emergence",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetCardInGraveyard("target nonland permanent card in your graveyard with mana value less than or equal to the number of permanent cards in your graveyard",
			effects.Nonland(), effects.Permanent(), effects.YouOwn(), manaValueAtMostPermanentCardsInYourGraveyard),
		Purpose: effects.ForTargets(effects.Reanimates(0)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			effects.ReanimateSingleTarget(ctx, item.Controller)
			return nil
		},
	})
}

// manaValueAtMostPermanentCardsInYourGraveyard is "with mana value less
// than or equal to the number of permanent cards in your graveyard", for
// the caster's graveyard.
func manaValueAtMostPermanentCardsInYourGraveyard(g *game.Game, caster uuid.UUID, c game.Card) bool {
	mv, ok := g.ManaValueForEffect(c)
	return ok && mv <= effects.PermanentCardsInGraveyard(g, caster)
}
