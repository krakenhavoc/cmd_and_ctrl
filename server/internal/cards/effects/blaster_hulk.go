package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blaster Hulk — Artifact Creature — Pirate {6}{R}{R}, 8/8:
//
//	"This spell costs {1} less to cast for each {E} (energy counter)
//	 you've paid or lost this turn.
//	 Haste
//	 Whenever this creature attacks, you get {E}{E}, then you may pay
//	 eight {E}. When you do, this creature deals 8 damage divided as you
//	 choose among up to eight targets."
//
// ADR 0129 §6 (#1995). The discount reads the turn tally of energy paid
// or lost (PlayerTurnTally.EnergyPaidOrLost) at CR 601.2f, through the
// one pricer, like Dargo's; it spends generic mana only, so the Hulk
// never costs less than {R}{R}. The attack trigger gets the energy, then
// asks the pay_unless prompt with an energy payment (CR 118.12); "when
// you do" is a reflexive trigger (CR 603.12) whose targets and division
// are announced as it goes on the stack (CR 601.2d through CR 603.3d):
// up to eight targets, each assigned at least 1 of the 8.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4b8ca7fd-3c8e-44fe-ab0a-0ddc2f2b047c",
		Name:            "Blaster Hulk",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Purpose:         game.Purpose{Energy: 2},
		SelfCostModifiers: []game.CostModifier{
			CostsLessForEachEnergyPaidOrLost(1, "This spell costs {1} less to cast for each {E} you've paid or lost this turn"),
		},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Blaster Hulk — you get {E}{E}, then you may pay eight {E}",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (GetEnergy{N: 2}).Apply(ctx); err != nil {
						return err
					}
					return MayPayEnergy{
						N:        8,
						Question: "Blaster Hulk — pay eight {E} to deal 8 damage divided among up to eight targets?",
						OnPay: func(ctx *Context) error {
							return WhenYouDo("Blaster Hulk — 8 damage divided as you choose among up to eight targets",
								blasterHulkDamageBody).Apply(ctx)
						},
					}.Apply(ctx)
				}),
		},
	})
}

// blasterHulkDamage is the reflexive body: each target its announced
// share of the 8, from the Hulk (its last-known information if it has
// left, CR 608.2h).
func blasterHulkDamage(g *game.Game, item *game.StackItem) error {
	return DealDividedDamage(NewContext(g, item))
}
