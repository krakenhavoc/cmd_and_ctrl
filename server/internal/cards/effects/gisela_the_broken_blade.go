package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gisela, the Broken Blade — Legendary Creature — Angel Horror {2}{W}{W},
// 4/3 (Innistrad Remastered):
//
//	"Flying, first strike, lifelink
//	 At the beginning of your end step, if you both own and control Gisela
//	 and a creature named Bruna, the Fading Light, exile them, then meld
//	 them into Brisela, Voice of Nightmares."
//
// An intervening "if" (CR 603.4), checked as the end step begins and again
// as the ability resolves (MeldWith), the way Titania's upkeep meld is.
func init() {
	Register(Spec{
		OracleID:        "f3e23d5e-bd88-4e7c-a3fb-db2a8cb05b22",
		Name:            "Gisela, the Broken Blade",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "first strike", "lifelink"},
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondBoard{
				Name: "you both own and control Gisela and a creature named Bruna, the Fading Light",
				Test: func(ctx *Context) bool {
					src, ok := ctx.Game.LookupCardForEffect(ctx.Source())
					return ok && YouOwnAndControlThisAndANamed("Bruna, the Fading Light", "creature")(ctx.Game, &src)
				},
			}, AtYourEndStep("Gisela, the Broken Blade — exile Gisela and Bruna, then meld them into Brisela, Voice of Nightmares",
				Do(MeldWith{Partner: "Bruna, the Fading Light", PartnerType: "creature"}))),
		},
	})
}
