package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gideon Jura — Legendary Planeswalker — Gideon {3}{W}{W}, starting
// loyalty 6:
//
//	"+2: During target opponent's next turn, creatures that player
//	     controls attack Gideon Jura if able.
//	 −2: Destroy target tapped creature.
//	 0: Until end of turn, Gideon Jura becomes a 6/6 Human Soldier
//	    creature that's still a planeswalker. Prevent all damage that
//	    would be dealt to him this turn."
//
// THE +2 is PlayersCreaturesAttackThisIfAble (attack_this_permanent.go,
// #2567): a CR 508.1d requirement on "creatures that player controls",
// read live (CR 611.2c, so a creature that player gains before their
// turn must attack too), obeyed only by attacking this Gideon as the
// object he was when the ability was activated, and binding only in the
// target opponent's next turn — every combat of it. Another opponent's
// creatures are never asked. A creature that cannot attack him for free
// (tapped, summoning-sick, "can't attack", an attack tax, Gideon gone)
// may attack anything else or stay home: the rulings.
//
// THE −2 destroys the target if it is still a tapped creature as the
// ability resolves (CR 608.2b). THE 0 is animateGideon
// (gideon_animate.go; #2046, ADR 0032 amendment of 2026-10-07).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7340dbcb-4b26-4038-be8c-a89e76294221",
		Name:            "Gideon Jura",
		Completeness:    CompletenessFull,
		StartingLoyalty: 6,
		Activated: []ActivatedAbility{
			{
				Label:   "+2: During target opponent's next turn, creatures that player controls attack Gideon Jura if able.",
				Cost:    LoyaltyCost(2),
				Targets: TargetPlayer("target opponent", Opponent()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetPlayer {
							continue
						}
						return PlayersCreaturesAttackThisIfAble{
							Player: t.ID,
							Label:  "Gideon Jura — creatures that player controls attack Gideon Jura if able",
						}.Apply(ctx)
					}
					return nil
				},
			},
			{
				Label:   "−2: Destroy target tapped creature.",
				Cost:    LoyaltyCost(-2),
				Targets: TargetCreature("target tapped creature", tappedPermanent()),
				Effect:  destroyFirstLegalTarget,
			},
			{
				Label: "0: Until end of turn, Gideon Jura becomes a 6/6 Human Soldier creature that's still a planeswalker. Prevent all damage that would be dealt to him this turn.",
				Cost:  LoyaltyCost(0),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return animateGideon(g, item, gideonAnimation{
						Label:     "Gideon Jura — a 6/6 Human Soldier creature until end of turn",
						Subtypes:  []string{"Human", "Soldier"},
						Power:     6,
						Toughness: 6,
					})
				},
			},
		},
	})
}
