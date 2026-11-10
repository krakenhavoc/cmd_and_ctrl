package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wingmantle Chaplain — Creature — Human Cleric {3}{W}, 0/3 (Foundations
// Commander):
//
//	"Defender
//	 When this creature enters, create a 1/1 white Bird creature token
//	 with flying for each creature with defender you control.
//	 Whenever another creature you control with defender enters, create a
//	 1/1 white Bird creature token with flying."
//
// The count is taken as the first ability resolves, the Chaplain itself
// included (the printed ruling). The second ability triggers once for
// each OTHER defender that enters, the Chaplain's own entry excluded.
// A creature has defender as it enters, read through the layers.
//
// No simplification.
func init() {
	bird := TokenCard("1/1 white Bird with flying")
	Register(Spec{
		OracleID:        "99e24041-910e-4ae2-a55b-b938d752eeff",
		Name:            "Wingmantle Chaplain",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Wingmantle Chaplain — create a Bird for each creature with defender you control",
				func(g *game.Game, item *game.StackItem) error {
					return CreateToken{Template: bird, N: cmdrB17DefendersControlled(g, item.Controller)}.Apply(NewContext(g, item))
				}),
			On(game.EventETB, func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
				if !AnotherCreatureEnteredUnderYourControl(ev, source, ch, g) {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && game.HasKeyword(&c, "defender")
			}, "Wingmantle Chaplain — create a 1/1 white Bird creature token with flying",
				Do(CreateToken{Template: bird, N: 1})),
		},
	})
}
