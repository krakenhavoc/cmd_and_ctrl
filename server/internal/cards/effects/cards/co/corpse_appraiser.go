package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Corpse Appraiser — Creature — Vampire Rogue {U}{B}{R}, 3/3 (Streets of
// New Capenna):
//
//	"When this creature enters, exile up to one target creature card from a
//	 graveyard. If a card is put into exile this way, look at the top
//	 three cards of your library, then put one of those cards into your
//	 hand and the rest into your graveyard."
//
// "Up to one" target: with none chosen nothing is exiled and the rider
// does not happen. The rider is gated on the exile having happened (the
// keyed tail runs only for a card that moved), so a target that left the
// graveyard in response gives nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bd3cdd9b-aa67-4bf5-80c8-53574d2c221e",
		Name:         "Corpse Appraiser",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Corpse Appraiser — exile up to one target creature card from a graveyard",
				corpseAppraiserExile),
				effects.TargetCardInGraveyard("up to one target creature card from a graveyard", effects.Creature()).WithCount(0, 1)),
		},
	})
}

func corpseAppraiserExile(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return effects.ExileTarget{Target: t.ID, ThenTail: ctx.ItemTail(corpseAppraiserLook)}.Apply(ctx)
	}
	return nil
}

var corpseAppraiserLook = effects.CardTail("corpse-appraiser/look-at-top-three", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	player := ctx.Controller()
	return effects.TakeFromLibraryToHand{
		Player: player,
		Cards:  ctx.Game.LookAtTopOfLibraryForEffect(player, 3),
		Max:    1,
		Label:  "Corpse Appraiser — put one of them into your hand and the rest into your graveyard",
		Key:    corpseAppraiserTake,
	}.Apply(ctx)
})

var corpseAppraiserTake = effects.RegisterTake("choose-cards/corpse-appraiser-take",
	effects.TakeDef{Then: effects.TakeRestIntoGraveyard})
