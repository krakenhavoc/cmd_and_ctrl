package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Skilled Animator — Creature — Human Artificer {2}{U}, 1/3 (Jumpstart 2022, #5948):
//
//	"When this creature enters, target artifact you control becomes an
//	 artifact creature with base power and toughness 5/5 for as long as
//	 this creature remains on the battlefield."
//
// Iron Man's become-a-creature scoped effect (type and base P/T, layers 4
// and 7b) with Sower of Temptation's duration: it lasts for as long as the
// Animator remains on the battlefield (CR 611.2b), and never begins if the
// Animator already left (a flickered Animator is a new object, CR 400.7).
// The artifact keeps its other types, so a Vehicle or Equipment stays one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6e51a0b2-e07f-4c9d-b055-dd1dcd2e29d4",
		Name:         "Skilled Animator",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Skilled Animator — target artifact you control becomes a 5/5 artifact creature for as long as this creature remains on the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					d, ok := effects.DurationWhileSourceRemains(ctx, ctx.Source())
					if !ok {
						return nil
					}
					mods := []game.Mod{game.AddTypesMod("Artifact", "Creature")}
					mods = append(mods, game.SetBasePTMods(5, 5)...)
					return effects.ScopedEffectFor{
						Target:   id,
						Mods:     mods,
						Duration: g.PinnedTo(d, id),
						Label:    "Skilled Animator — becomes a 5/5 artifact creature",
					}.Apply(ctx)
				}),
				effects.TargetPermanent("target artifact you control", effects.YouControl(), effects.Artifact())),
		},
	})
}
