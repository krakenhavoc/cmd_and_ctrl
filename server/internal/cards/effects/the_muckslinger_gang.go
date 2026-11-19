package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Muckslinger Gang — Legendary Creature — Goblin Lizard Beast
// {1}{B}{R}{G}, 3/3 (Mystery Booster Commander Edition):
//
//	"Each other creature you control gets +3/+3 as long as it's not
//	 attacking or blocking.
//	 Creature spells you control and creature cards you own in any zone
//	 other than the battlefield or the stack get +3/+3."
//
// Three layer 7c statics. The first is Intrepid Ace's attacking-status
// reading over the other creatures the controller controls. The other two
// reach past the battlefield through the zone layer step (AffectsZones,
// Maskwood Nexus' shape): creature spells the controller controls, and
// creature cards the controller owns in a hand, library, graveyard, exile
// or the command zone. Printed "in any zone other than the battlefield or
// the stack" is read with the spells on top, as the oracle names them
// separately, so a creature spell gets the bonus once. Anything the Gang
// does not reach is untouched: an opponent's creatures, the Gang itself,
// and an attacking or blocking creature on the battlefield.
//
// No simplification.
func init() {
	bonus := func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
		c.Power += 3
		c.Toughness += 3
	}
	Register(Spec{
		OracleID:     "56cacdcf-2f11-4969-9629-446068a15369",
		Name:         "The Muckslinger Gang",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			{
				Label:                    "Each other creature you control gets +3/+3 as long as it's not attacking or blocking.",
				Layer:                    game.Layer7PT,
				SubLayer:                 game.SubLayer7C_Modify,
				DependsOnAttackingStatus: true,
				AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
					return OtherCreaturesYouControl(target, g, source) &&
						target.AttackingTarget == uuid.Nil && target.BlockingTarget == uuid.Nil
				},
				Apply: bonus,
			},
			{
				Label:        "Creature spells you control get +3/+3.",
				Layer:        game.Layer7PT,
				SubLayer:     game.SubLayer7C_Modify,
				AffectsZones: SpellsOnTheStack(),
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.Controller == source.Controller && target.IsCreature()
				},
				Apply: bonus,
			},
			{
				Label:        "Creature cards you own in any zone other than the battlefield or the stack get +3/+3.",
				Layer:        game.Layer7PT,
				SubLayer:     game.SubLayer7C_Modify,
				AffectsZones: CardsOffTheBattlefield(),
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.Owner == source.Controller && target.IsCreature()
				},
				Apply: bonus,
			},
		},
	})
}
