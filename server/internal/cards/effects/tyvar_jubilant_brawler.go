package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tyvar, Jubilant Brawler — Legendary Planeswalker — Tyvar {1}{B}{G},
// starting loyalty 3 (Phyrexia: All Will Be One, #5479):
//
//	"You may activate abilities of creatures you control as though those
//	 creatures had haste.
//	 +1: Untap up to one target creature.
//	 −2: Mill three cards, then you may return a creature card with mana
//	     value 2 or less from your graveyard to the battlefield."
//
// The static is Shang-Chi's ActivateAsThoughHaste (CR 302.6's {T} bar is
// lifted for abilities, not for attacking). The −2's return is chosen once
// the mill has finished, from the cards in the graveyard then (CR 608.2c),
// so a creature just milled is a choice.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:              "848c2b8f-8086-4047-8ad0-24b701dc4d0d",
		Name:                  "Tyvar, Jubilant Brawler",
		Completeness:          CompletenessFull,
		StartingLoyalty:       3,
		ActivateAsThoughHaste: true,
		Activated: []ActivatedAbility{
			{
				Label:   "+1: Untap up to one target creature.",
				Cost:    LoyaltyCost(1),
				Targets: TargetCreature("up to one target creature").WithCount(0, 1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					targets := ctx.LegalTargets()
					if len(targets) == 0 || targets[0].Kind != game.TargetCard {
						return nil
					}
					return UntapTarget{Target: targets[0].ID}.Apply(ctx)
				},
			},
			{
				Label: "−2: Mill three cards, then you may return a creature card with mana value 2 or less from your graveyard to the battlefield.",
				Cost:  LoyaltyCost(-2),
				Effect: Do(
					Instruction(func(ctx *Context) error {
						return millEachPlayerInTurn(ctx, []uuid.UUID{ctx.Controller()}, constantMill(3), game.KeyedTail{})
					}),
					ReturnChosenFromGraveyard{
						Question: "Tyvar, Jubilant Brawler — you may return a creature card with mana value 2 or less from your graveyard to the battlefield",
						Match:    func(c game.Card) bool { return c.IsCreature() && c.ManaValue() <= 2 },
						Max:      1,
					},
				),
			},
		},
	})
}
