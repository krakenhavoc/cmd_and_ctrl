package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Primal Witchstalker — Creature — Wolf {1}{B}{G}, 2/1:
//
//	"Menace
//	 When this creature enters, mill four cards. When you do, return
//	 target land card from your graveyard to the battlefield tapped."
//
// The second sentence is a CR 603.12 reflexive trigger created when the
// mill happens, so the land is chosen as that trigger goes on the stack
// and a land the mill just put there is a legal pick. An empty library
// mills nothing, so there is nothing to "do" and no reflexive trigger.
// With no land card in the graveyard the reflexive trigger has no legal
// target and is removed (CR 603.3d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c21fafaf-4510-4752-a501-606ae68e6699",
		Name:            "Primal Witchstalker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Primal Witchstalker — mill four cards", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				p := g.PlayerByIDForEffect(item.Controller)
				if p == nil || p.Library == nil || p.Library.Size() == 0 {
					return nil
				}
				if err := (MillCards{Player: item.Controller, N: 4}).Apply(ctx); err != nil {
					return err
				}
				return WhenYouDo("Primal Witchstalker — return target land card from your graveyard to the battlefield tapped",
					primalWitchstalkerLandBody).Apply(ctx)
			}),
		},
	})
}

// primalWitchstalkerLandBody is the reflexive trigger: the chosen land
// card comes back tapped under its owner's control.
var primalWitchstalkerLandBody = game.ReflexiveBody("primal-witchstalker/land",
	func(g *game.Game, item *game.StackItem, _ game.EffectParams) error {
		ctx := NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			return ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield, Tapped: true}.Apply(ctx)
		}
		return nil
	},
	constTargets(func() *game.TargetSpec {
		return TargetCardInGraveyard("target land card from your graveyard", Land(), YouOwn())
	}))
