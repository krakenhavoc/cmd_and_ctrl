package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sanguine Spy — Creature — Vampire Rogue {2}{B}, 2/3 (Streets of New Capenna):
//
//	"Menace, lifelink
//	 {1}, Sacrifice another creature: Surveil 1.
//	 At the beginning of your end step, if there are five or more mana values
//	 among cards in your graveyard, you may pay 2 life. If you do, draw a card."
//
// The activated ability is a mana cost plus "sacrifice another creature" (the
// source is never a legal payment). The end-step trigger carries its
// intervening "if" (CR 603.4): checked as the step begins and again as the
// trigger resolves (InterveningIf), and the life is paid as the trigger
// resolves (Bringer of the Black Dawn's MayChoice with LifeCost), the draw only
// if it was paid.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "03af4b10-900e-4013-bd5d-dd8d132d50fb",
		Name:            "Sanguine Spy",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace", "lifelink"},
		Activated: []ActivatedAbility{{
			Label:  "{1}, Sacrifice another creature: Surveil 1.",
			Cost:   Plus(ManaCost("{1}"), SacrificeAnotherN(1, "another creature", Creature())),
			Effect: Do(Surveil{N: 1}),
		}},
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondSNCFiveManaValues{},
				On(game.EventBeginEndStep, AllOf(ByYou, sncFiveMVsWhen),
					"Sanguine Spy — you may pay 2 life; if you do, draw a card",
					func(g *game.Game, item *game.StackItem) error {
						return MayChoice{
							Question: "Sanguine Spy — pay 2 life to draw a card?",
							LifeCost: 2,
							Then:     sanguineSpyPaid,
						}.Apply(NewContext(g, item))
					})),
		},
	})
}

// sanguineSpyPaid is the "yes": pay 2 life, and only if that was possible,
// draw a card. Keyed.
var sanguineSpyPaid = MayYes("confirm/sanguine-spy-pay-life-draw", func(ctx *Context, _ game.ConfirmAnswer) error {
	p := ctx.Game.PlayerByIDForEffect(ctx.Controller())
	if p == nil || p.Eliminated || p.Life < 2 {
		return nil
	}
	if err := ctx.Game.PayLifeForEffect(ctx.Source(), ctx.Controller(), 2); err != nil {
		return err
	}
	return DrawCards{N: 1}.Apply(ctx)
})
