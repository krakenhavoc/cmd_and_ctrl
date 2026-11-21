package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Foretold Soldier — Creature — Alien Zombie Soldier {2}{G}{G}, 6/6 (Doctor Who):
//
//	"This creature must be blocked if able.
//	 This creature can't be blocked by more than one creature.
//	 Whenever this creature deals damage, exile it face down. It becomes foretold.
//	 Foretell {1}{G} (During your turn, you may pay {2} and exile this card from
//	 your hand face down. Cast it on a later turn for its foretell cost.)"
//
// The block requirement is Canopy Stalker's and the single-blocker limit
// Battlefront Krushok's (MaxBlockers on itself). Foretell is the engine's special
// action. The trigger fires once per damage event the creature deals, combat or
// not, and exiles the permanent face down as a foretold card (CR 702.143b): a new
// seam, game.ForetellPermanentForEffect, files it in its owner's exile with the
// foretell permission at {1}{G} on a LATER turn. A creature that has already left
// the battlefield (it died in the same combat-damage step) is not exiled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "02a45719-ee4a-49c4-96ec-a140cd6e1547",
		Name:           "The Foretold Soldier",
		Completeness:   effects.CompletenessFull,
		Static:         []game.StaticAbility{effects.MustBeBlocked()},
		BlockRules:     []game.BlockRule{effects.MaxBlockers(effects.OnSelf(), 1)},
		SpecialActions: []game.SpecialAction{effects.Foretell("{1}{G}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventDealDamage && ev.Amount > 0 && ev.Source == source.InstanceID
			}, "The Foretold Soldier — exile it face down; it becomes foretold",
				func(g *game.Game, item *game.StackItem) error {
					if !effects.SourceIsStillThisPermanent(g, item) {
						return nil
					}
					return g.ForetellPermanentForEffect(item.SourceCardID, "{1}{G}")
				}),
		},
	})
}
