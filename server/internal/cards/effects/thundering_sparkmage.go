package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thundering Sparkmage — Creature — Human Wizard {3}{R}, 2/2 (Zendikar
// Rising, #5508):
//
//	"When this creature enters, it deals X damage to target creature or
//	 planeswalker, where X is the number of creatures in your party.
//	 (Your party consists of up to one each of Cleric, Rogue, Warrior,
//	 and Wizard.)"
//
// The target is chosen as the trigger goes on the stack and re-checked as
// it resolves (CR 608.2b). X is the party (CR 700.8, finB9PartySize)
// counted at resolution (CR 700.8a), the Sparkmage itself being the
// Wizard if nothing else is. The Sparkmage is the damage's source.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a786e495-ae3a-4e15-ab32-661d524e2b2d",
		Name:         "Thundering Sparkmage",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Thundering Sparkmage — it deals X damage to target creature or planeswalker, X the creatures in your party",
					Do(Instruction(func(ctx *Context) error {
						id, ok := b16FirstLegalTargetCard(ctx)
						if x := finB9PartySize(ctx.Game, ctx.Controller()); ok && x > 0 {
							return DealDamage{Source: ctx.Source(), Target: id, Amount: x}.Apply(ctx)
						}
						return nil
					}))),
				TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
			),
		},
	})
}
