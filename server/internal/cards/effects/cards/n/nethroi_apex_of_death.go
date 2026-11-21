package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nethroi, Apex of Death — Legendary Creature — Cat Nightmare Beast
// {2}{W}{B}{G}, 5/5 (Ikoria: Lair of Behemoths; reprinted in the S82
// sets' Commander decks, #5908):
//
//	"Mutate {4}{G/W}{B}{B} (If you cast this spell for its mutate cost,
//	 put it over or under target non-Human creature you own. They mutate
//	 into the creature on top plus all abilities from under it.)
//	 Deathtouch, lifelink
//	 Whenever this creature mutates, return any number of target creature
//	 cards with total power 10 or less from your graveyard to the
//	 battlefield."
//
// Mutate is the engine's (ADR 0223): Everquill Phoenix's shape. The
// merged creature has every component's abilities (CR 702.140e), so
// deathtouch, lifelink and the trigger work from underneath too, and the
// trigger fires on the merge whichever card ended on top (CR 702.140d).
// Cast for its mana cost, or with its mutate target gone, nothing merges
// and nothing triggers.
//
// The trigger's clause is Reunion of the House's: any number of target
// creature cards in your graveyard with a total power of 10 or less
// (game.TargetTotalPower), power read in the graveyard. Whatever is
// still legal as it resolves (CR 608.2b) returns together, under its
// owner's (your) control.
//
// No simplification.
func init() {
	returns := effects.TargetCardInGraveyard("any number of target creature cards with total power 10 or less from your graveyard",
		effects.Creature(), effects.YouOwn()).WithCount(0, 0).WithTotal(effects.CmdrB13TotalPowerAtMost(10))
	effects.Register(effects.Spec{
		OracleID:         "66bef5a7-79b7-4a9c-98dc-8670f867f04c",
		Name:             "Nethroi, Apex of Death",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"deathtouch", "lifelink"},
		AlternativeCosts: []game.AlternativeCost{effects.Mutate("{4}{G/W}{B}{B}")},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(
				effects.On(game.EventMutate, effects.ThisCreatureMutates,
					"Nethroi, Apex of Death — return any number of target creature cards with total power 10 or less from your graveyard to the battlefield",
					func(g *game.Game, item *game.StackItem) error {
						return effects.ReturnLegalGraveyardTargetsToBattlefield(effects.NewContext(g, item))
					}),
				returns), effects.ForTargets(effects.Reanimates(0))),
		},
	})
}
