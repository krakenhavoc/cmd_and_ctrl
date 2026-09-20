package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mister Negative — Legendary Creature — Human Villain {5}{W}{B},
// 5/5 :
//
//	"Vigilance, lifelink.
//	 Darkforce Inversion — When Mister Negative enters, you may
//	 exchange life totals with target opponent. If you lost life
//	 this way, draw that many cards."
//
// The exchange is the card, and the second sentence is what makes it
// castable when you are AHEAD: hand an opponent your comfortable 38
// and take their 12, and the 26 life you just lost is 26 cards. With
// lifelink on a 5/5 the total you traded away starts climbing back
// the turn after.
//
// "If you lost life this way" is the APPLIED amount, which is why the
// controller's half runs through ChangePlayerLifeThenForEffect and
// the draw reads its continuation rather than subtracting the two
// totals: a replacement can change what actually happened (Rhox
// Faithmender on the gaining side, a "your life total can't change"
// effect on either), and arithmetic on the raw totals would draw
// cards for life that never moved. A swap that gains you life, or
// moves nothing at all, draws nothing.
//
// The trigger goes on the stack (CR 603) and is a "you may" (CR
// 603.4), so declining is a line and so is answering the trigger by
// removing the target in response — one target, no legal target, the
// trigger does nothing and no cards are drawn.
//
// Vigilance and lifelink are printed keywords and ride
// PrintedKeywords; the engine generates the layer-6 grant.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c8da3262-826d-4ea8-bb22-06374bebefe5",
		Name:            "Mister Negative",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance", "lifelink"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				Optional(
					WhenThisEnters("Mister Negative — exchange life totals with target opponent",
						misterNegativeInversion),
					"Darkforce Inversion — exchange life totals with target opponent?"),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}

// misterNegativeInversion is Darkforce Inversion: swap with the
// chosen opponent, then draw for the life the controller actually
// lost. `applied` is signed the way the event is, so a loss is
// negative and a gain draws nothing.
func misterNegativeInversion(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	victim, ok := firstLegalPlayerTarget(ctx)
	if !ok {
		return nil
	}
	return exchangeLifeTotalsThen(ctx, item.Controller, victim, func(g *game.Game, applied int) error {
		if applied >= 0 {
			return nil
		}
		return g.DrawNForEffect(item.Controller, -applied)
	})
}
