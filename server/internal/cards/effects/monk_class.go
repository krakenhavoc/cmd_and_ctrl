package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Monk Class — {W}{U} Enchantment — Class (Adventures in the Forgotten Realms):
//
//	"(Gain the next level as a sorcery to add its ability.)
//	 The second spell you cast each turn costs {1} less to cast.
//	 {W}{U}: Level 2
//	 When this Class becomes level 2, return up to one target nonland permanent
//	 to its owner's hand.
//	 {1}{W}{U}: Level 3
//	 At the beginning of your upkeep, exile the top card of your library. For as
//	 long as it remains exiled, it has "You may cast this card from exile as long
//	 as you've cast another spell this turn.""
//
// Level 1 is Raging Battle Mouse's discount (the second spell of the turn, {1}
// off generic mana), ungated (CR 716.3). Level 2 is Disruptor of Currents' "up to
// one target nonland permanent" bounce on BecomesLevel, gated at level 2. Level 3
// is an upkeep trigger gated at level 3 that exiles the top card of the
// controller's library with a cast permission for as long as it remains exiled
// (WhileExiled): the permission is the granted ability, so it follows the card in
// exile and is gone once the card leaves. It is "cast", so a land is stranded; its
// condition is CondCastASpellThisTurn, dormant (not lost) until the holder has
// cast a spell that turn, and the cast itself keeps its normal timing and cost. The
// holder is the controller at the moment of the exile, which is the card's owner.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "65b28ec7-1099-4f9d-8d57-2ffd8a5b0e91",
		Name:         "Monk Class",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(1, "The second spell you cast each turn costs {1} less to cast.",
				YourSpell(), secondSpellYouCastThisTurn()),
		},
		Activated: []ActivatedAbility{
			LevelUp(2, ManaCost("{W}{U}")),
			LevelUp(3, ManaCost("{1}{W}{U}")),
		},
		Triggered: []game.TriggeredAbility{
			func() game.TriggeredAbility {
				t := BecomesLevel(2, "Monk Class — level 2: return up to one target nonland permanent to its owner's hand",
					func(g *game.Game, item *game.StackItem) error {
						return bounceTheTarget(item, NewContext(g, item))
					})
				t.Targets = TargetPermanent("up to one target nonland permanent", Nonland()).WithCount(0, 1)
				return t
			}(),
			AtLevel(3, AtYourUpkeep("Monk Class — exile the top card of your library; you may cast it while you've cast another spell this turn",
				func(g *game.Game, item *game.StackItem) error {
					return ExileTopWithPermission{
						From: item.Controller, GrantTo: item.Controller, N: 1,
						CastOnly:    true,
						WhileExiled: true,
						Condition:   &game.PermissionCondition{Key: game.CondCastASpellThisTurn},
					}.Apply(NewContext(g, item))
				})),
		},
	})
}
