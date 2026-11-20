package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kabira Outrider — Creature — Human Warrior {3}{W}, 3/3 (Zendikar
// Rising, #5508):
//
//	"When this creature enters, target creature gets +1/+1 until end of
//	 turn for each creature in your party. (Your party consists of up to
//	 one each of Cleric, Rogue, Warrior, and Wizard.)"
//
// The target is chosen as the trigger goes on the stack and re-checked as
// it resolves (CR 608.2b). The size is the party (CR 700.8,
// finB9PartySize) counted at resolution (CR 700.8a), the Outrider itself
// being the Warrior if nothing else is, and it is fixed from then on (CR
// 611.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "69f9a2fc-4873-4d7f-8405-3fb9a2091de5",
		Name:         "Kabira Outrider",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Kabira Outrider — target creature gets +1/+1 until end of turn for each creature in your party",
					Do(Instruction(func(ctx *Context) error {
						id, ok := b16FirstLegalTargetCard(ctx)
						if x := finB9PartySize(ctx.Game, ctx.Controller()); ok && x > 0 {
							return BoostUntilEOT{Target: id, Power: x, Toughness: x, Label: "Kabira Outrider — +1/+1 per party member"}.Apply(ctx)
						}
						return nil
					}))),
				TargetCreature("target creature"),
			),
		},
	})
}
