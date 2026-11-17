package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tranquil Frillback — Creature — Dinosaur {2}{G}, 3/3 (March of the Machine:
// The Aftermath, #5476):
//
//	"When this creature enters, you may pay {G} up to three times. When you
//	 pay this cost one or more times, choose up to that many —
//	 • Destroy target artifact or enchantment.
//	 • Exile target player's graveyard.
//	 • You gain 4 life."
//
// Hawkeye, Master Marksman's sentence on an enters trigger with {G}: the
// controller picks how many times to pay as the trigger resolves, from 0 to
// the least of three and the mana they can make (CR 608.2d), and pays that
// much through the ordinary "you may pay" prompt. Paying one or more times
// makes ONE reflexive trigger (CR 603.12a), whose bullets are chosen as it
// goes on the stack (CR 603.3c): "up to that many", so none removes it
// (CR 700.2b), each bullet picks its own target then (CR 603.3d) and runs in
// printed order (CR 608.2c), and no bullet is chosen twice (CR 700.2d). The
// number of payments is fixed when the trigger is created (Params.Amount).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e87a4461-f3d1-443f-ad4c-dba91a993809",
		Name:         "Tranquil Frillback",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Tranquil Frillback — you may pay {G} up to three times; when you do, choose up to that many", tranquilFrillbackEnters),
		},
	})
}

// tranquilFrillbackEnters is the enters trigger's body: pay {G} up to three
// times.
func tranquilFrillbackEnters(g *game.Game, item *game.StackItem) error {
	return MayPayX{
		Symbol: "{G}",
		Max:    3,
		Label:  "Tranquil Frillback",
		Buys:   "choose up to X modes",
		Goal:   XAsHighAsYouCan,
		Unit:   game.PayAmountOther,
		Then:   tranquilFrillbackPaid,
	}.Apply(NewContext(g, item))
}

// tranquilFrillbackPaid is the paid half, keyed: the number of payments rides
// the prompt.
var tranquilFrillbackPaid = PayUnlessBranches("pay-unless/tranquil-frillback-pay-g",
	func(ctx *Context, r game.PayUnlessAnswer) error {
		if r.Number <= 0 {
			return nil
		}
		return ReflexiveTrigger{
			Label:  "Tranquil Frillback — choose up to " + strconv.Itoa(r.Number),
			Body:   tranquilFrillbackBody,
			Params: game.EffectParams{Amount: r.Number},
		}.Apply(ctx)
	}, nil)

// tranquilFrillbackBody is the reflexive trigger: its bullets are its mode
// clause, its count the payments made.
var tranquilFrillbackBody = game.ReflexiveModalBody("tranquil-frillback/modes", tranquilFrillbackModes)

func tranquilFrillbackModes(_ uuid.UUID, p game.EffectParams) *game.ModeSpec {
	return ChooseN("Choose up to "+strconv.Itoa(p.Amount), 0, p.Amount,
		ModeDoing("Destroy target artifact or enchantment.",
			TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
			DestroyTheModesTarget),
		ModeDoing("Exile target player's graveyard.", TargetPlayer("target player"), exileTheModesTargetPlayersGraveyard),
		ModeDoing("You gain 4 life.", nil, func(item *game.StackItem, ctx *Context, _ int) error {
			return GainLife{Player: item.Controller, Amount: 4}.Apply(ctx)
		}),
	)
}
