package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Plain Walker — Creature — Kithkin Rogue {2}{W}, 2/3:
//
//		"Plainswalk
//		 Planeswalkerwalk (This creature can't be blocked as long as defending
//		 player controls a planeswalker.)
//		 Whenever Plain Walker deals combat damage to a player or planeswalker,
//		 planeswalk. (If you're playing Planechase, proceed to the next plane.)"
//
//	  - Plainswalk is the printed keyword the block check already enforces
//	    (CR 702.14c).
//	  - Planeswalkerwalk is the evasion its reminder text defines: a block rule
//	    on the Walker that refuses every blocker while the player it is attacking
//	    (or whose planeswalker it is attacking, CR 506.2) controls a planeswalker.
//	    It is read when blocks are declared, live, so a planeswalker that left
//	    before blockers no longer switches it on.
//	  - The trigger fires once per combat damage event dealt to a player or a
//	    planeswalker, and planeswalks its controller (CR 701.31). Outside a
//	    Planechase game there is no planar deck and it does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c8d38d18-2669-494a-89a8-2ead0cc9628a",
		Name:            "Plain Walker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"plainswalk"},
		BlockRules: []game.BlockRule{effects.CantBeBlockedWhile(effects.OnSelf(), func(g *game.Game, _ uuid.UUID, c game.Card) bool {
			defender := g.DefendingPlayerForAttackerForEffect(c.InstanceID)
			if defender == uuid.Nil {
				return false
			}
			for _, p := range g.Battlefield.Cards {
				if p.Controller == defender && p.IsPlaneswalker() {
					return true
				}
			}
			return false
		})},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, effects.ThisDealtCombatDamageToAPlayerOrPlaneswalker,
				"Plain Walker — planeswalk", func(g *game.Game, item *game.StackItem) error {
					if g.Planechase == nil {
						return nil
					}
					return g.PlaneswalkForEffect(item.Controller)
				}),
		},
	})
}
