package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Warren Warleader — Creature — Rabbit Knight {2}{W}{W}, 4/4
// (Bloomburrow, #4467):
//
//	"Offspring {2} (You may pay an additional {2} as you cast this
//	 spell. If you do, when this creature enters, create a 1/1 token
//	 copy of it.)
//	 Whenever you attack, choose one —
//	 • Create a 1/1 white Rabbit creature token that's tapped and
//	   attacking.
//	 • Attacking creatures you control get +1/+1 until end of turn."
//
// Offspring is the shared seam (offspring.go). "Whenever you attack" is
// one trigger per attack declaration of yours (OncePerBatch over
// attackDeclaredByYou, Boosted Sloop's), modal as it goes on the stack
// (CR 603.3c). The Rabbit is Hero of Bladehold's token: you choose what
// it attacks (the ruling), the first declared attacker's choice offered
// first, and it was never declared, so nothing that reads "whenever a
// creature attacks" sees it (CR 508.4). The pump is locked to the
// creatures attacking as it resolves (CR 611.2c).
//
// No simplification.
func init() {
	trigger := OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
		return attackDeclaredByYou(ev, source.Controller)
	}, "Warren Warleader — you attacked: choose one", func(*game.Game, *game.StackItem) error { return nil }))
	trigger.Modes = ChooseOne(
		ModeDoing("Create a 1/1 white Rabbit creature token that's tapped and attacking.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return CreateTokensAttackingYourChoice{
					Template: TokenCard("1/1 white Rabbit"),
					N:        1,
					Tapped:   true,
					Prefer:   triggerAttackTarget(item),
					Label:    "Warren Warleader",
				}.Apply(ctx)
			}),
		ModeDoing("Attacking creatures you control get +1/+1 until end of turn.", nil,
			func(_ *game.StackItem, ctx *Context, _ int) error {
				return BoostUntilEOT{
					Match:     And(AttackingCreature(), YouControl()),
					Power:     1,
					Toughness: 1,
					Label:     "Warren Warleader — attacking creatures +1/+1",
				}.Apply(ctx)
			}),
	)
	Register(Spec{
		OracleID:      "a7e7059e-83f8-4124-9cc0-64ba150b5d96",
		Name:          "Warren Warleader",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Offspring("{2}")},
		Triggered: []game.TriggeredAbility{
			OffspringToken("Warren Warleader"),
			trigger,
		},
	})
}
