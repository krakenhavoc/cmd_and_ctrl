package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vibrance — Creature — Elemental Incarnation {3}{R/G}{R/G}, 4/4
// (Lorwyn Eclipsed, #4458):
//
//	"When this creature enters, if {R}{R} was spent to cast it, this
//	 creature deals 3 damage to any target.
//	 When this creature enters, if {G}{G} was spent to cast it, search
//	 your library for a land card, reveal it, put it into your hand,
//	 then shuffle. You gain 2 life.
//	 Evoke {R/G}{R/G}"
//
// Deceit's cycle-mate: two enters triggers behind intervening ifs
// (CR 603.4) on the mana spent (deceitSpent). Vibrance deals the damage
// even if it has left the battlefield by then (an evoked one sacrificed
// first): the source is read as it last existed (CR 113.7a). The search
// is Expedition Map's (any land card, revealed), and the life is gained
// after it, whether or not a land was found.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "52867cdd-8bf2-4cae-bb8d-0f16a18a455b",
		Name:           "Vibrance",
		ReadsManaSpent: game.ReadsOwnPayment,
		Completeness:   CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{
			Evoke("{R/G}{R/G}"),
		},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(
				On(game.EventETB, AllOf(Self, deceitSpent("R")),
					"Vibrance — 3 damage to any target",
					sourceDealsDamageToEachLegalTarget(3)),
				TargetAny()), ForTargets(DamageToTarget(0, 3))),
			On(game.EventETB, AllOf(Self, deceitSpent("G")),
				"Vibrance — search your library for a land card, put it into your hand, then you gain 2 life",
				Do(
					Instruction(func(ctx *Context) error {
						return SearchLibrary{
							Player:    ctx.Controller(),
							Predicate: b03IsLandCard,
							Dest:      game.ZoneHand,
							Limit:     1,
							Reveal:    true,
							Shuffle:   true,
							Reason:    "Vibrance — a land card, to your hand",
						}.Apply(ctx)
					}),
					GainLife{Amount: 2},
				)),
		},
	})
}
