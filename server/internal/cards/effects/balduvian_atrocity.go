package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Balduvian Atrocity — Creature — Phyrexian Berserker {2}{B}, 2/3 (Dominaria
// United):
//
//	"Kicker {R}
//	 Menace
//	 When this creature enters, if it was kicked, return target creature
//	 card with mana value 3 or less from your graveyard to the battlefield.
//	 It gains haste. Sacrifice it at the beginning of the next end step."
//
// Sun-Blessed Healer's kicked enters trigger (an intervening "if", CR 603.4,
// read off the permanent's carried kicker record) with a creature target of
// mana value 3 or less. The returned creature gains haste and a delayed
// trigger sacrifices that object at the next end step; a target gone by
// resolution makes the whole ability do nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0dcd13ab-37f9-4d66-8227-1e72d920f297",
		Name:            "Balduvian Atrocity",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		OptionalCosts:   []game.AdditionalCost{Kicker("{R}")},
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventETB, AllOf(Self, ThisKickedAtLeast(1)),
					"Balduvian Atrocity — kicked, return target creature card with mana value 3 or less from your graveyard to the battlefield; it gains haste; sacrifice it at the beginning of the next end step",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						return ReturnFromGraveyardTogether{
							Targets:    legalTargetCardIDs(ctx),
							Controller: ctx.Controller(),
							ThenTail:   dmuB07HasteSacrificeTail,
						}.Apply(ctx)
					}),
				TargetCardInGraveyard("target creature card with mana value 3 or less from your graveyard",
					YouOwn(), Creature(), ManaValueLE(3))),
		},
	})
}
