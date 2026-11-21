package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Alesha, Who Smiles at Death — Legendary Creature — Human Warrior
// {2}{R}, 3/2:
//
//	"First strike
//	 Whenever Alesha attacks, you may pay {W/B}{W/B}. If you do, return
//	 target creature card with power 2 or less from your graveyard to the
//	 battlefield tapped and attacking."
//
// The target is chosen as the trigger goes on the stack; the payment is
// offered on resolution and carries the card as data. The card must
// still be in the graveyard when the payment is made (putAttacking
// checks the zone), and it enters tapped and attacking without having
// been declared as an attacker (CR 508.4), the player choosing what it
// attacks.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6969a7e2-6866-4001-a139-24b3be13deae",
		Name:            "Alesha, Who Smiles at Death",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverThisAttacks(
				"Alesha, Who Smiles at Death — you may pay {W/B}{W/B} to return target creature card with power 2 or less from your graveyard to the battlefield tapped and attacking",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					ids := effects.LegalTargetCardIDs(ctx)
					if len(ids) == 0 {
						return nil
					}
					return effects.MayPay{
						Chooser:  ctx.Controller(),
						Cost:     "{W/B}{W/B}",
						Question: "Alesha, Who Smiles at Death — pay {W/B}{W/B} to return the target creature card to the battlefield tapped and attacking?",
						Then:     aleshaSmilesPaid,
						Carry:    []uuid.UUID{ids[0]},
					}.Apply(ctx)
				}),
				effects.TargetCardInGraveyard("target creature card with power 2 or less from your graveyard",
					effects.Creature(), effects.YouOwn(), effects.PowerLE(2))),
		},
	})
}

var aleshaSmilesPaid = effects.PayUnlessBranches("pay-unless/alesha-who-smiles-at-death-return",
	func(ctx *effects.Context, r game.PayUnlessAnswer) error {
		if len(r.Carry) != 1 {
			return nil
		}
		return effects.PutAttackingYourChoice(ctx, r.Carry[0], game.ZoneGraveyard, "Alesha, Who Smiles at Death")
	}, nil)
