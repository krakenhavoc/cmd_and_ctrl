package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Combustion Man — Legendary Creature — Human Assassin {3}{R}{R}, 4/6
// (Avatar: The Last Airbender, #4459):
//
//	"Whenever Combustion Man attacks, destroy target permanent unless its
//	 controller has Combustion Man deal damage to them equal to his
//	 power."
//
// The target's controller, read as the trigger resolves, is the one
// player asked (a pay-unless with a damage payment, ADR 0150 §3). The
// damage is his power then, or his last known power if he has left
// (CR 608.2h), and he deals it through the ordinary damage path, so
// prevention applies and taking it is still the payment. Declining
// destroys the target if it is still the same permanent (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "533b107c-fd66-45ef-b74e-e6bd9b75ea8f",
		Name:         "Combustion Man",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverThisAttacks("Combustion Man — destroy target permanent unless its controller has him deal damage to them equal to his power",
				combustionManAsk),
				TargetPermanent("target permanent")),
		},
	})
}

// combustionManAsk asks the target's controller.
func combustionManAsk(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		c, ok := g.LookupCardForEffect(t.ID)
		if !ok {
			return nil
		}
		n := max(sourcePowerNow(ctx), 0)
		return g.QueuePayUnlessPromptKeyedForEffect(game.PayUnlessKeyedPrompt{
			Chooser:  c.Controller,
			Source:   ctx.Source(),
			Cost:     "Have Combustion Man deal " + strconv.Itoa(n) + " damage to you",
			Question: "Combustion Man — take " + strconv.Itoa(n) + " damage, or " + c.Name + " is destroyed?",
			Action:   &game.PayAction{Kind: game.PayActionDamage, Count: n},
			Then:     combustionManDestroy,
			Carry:    []uuid.UUID{ctx.Controller()},
			Asked:    game.NewAskedItem(item),
		})
	}
	return nil
}

// combustionManDestroy is "destroy target permanent" when its controller
// declines the damage.
var combustionManDestroy = PayUnlessElse("pay-unless/combustion-man-destroy",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		for _, t := range ctx.LegalTargets() {
			if t.Kind == game.TargetCard {
				return DestroyTarget{Target: t.ID}.Apply(ctx)
			}
		}
		return nil
	})
