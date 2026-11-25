package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Awakener Druid — Creature — Human Druid {2}{G}, 1/1 (Jumpstart, #5510):
//
//	"When this creature enters, target Forest becomes a 4/5 green Treefolk
//	 creature for as long as this creature remains on the battlefield. It's
//	 still a land."
//
// Skilled Animator's scoped become-a-creature effect (type and subtype in
// layer 4, colour in layer 5, base power and toughness in layer 7b) with
// Sower of Temptation's duration (CR 611.2b): it lasts for as long as the
// Druid remains on the battlefield and never begins if the Druid already
// left. The Forest keeps its land types and abilities.
//
// No simplification.
func init() {
	const label = "Awakener Druid — target Forest becomes a 4/5 green Treefolk creature for as long as this creature remains on the battlefield"
	effects.Register(effects.Spec{
		OracleID:     "7e8f5b34-fd16-4307-9104-17b584b25b14",
		Name:         "Awakener Druid",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters(label,
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
					mods := []game.Mod{game.AddTypesMod("Creature"), game.AddSubtypesMod(ctx.TextSubtype("Treefolk")), game.SetColorsMod(ctx.TextColor("G"))}
					mods = append(mods, game.SetBasePTMods(4, 5)...)
					return effects.ScopedEffectFor{
						Target:   id,
						Mods:     mods,
						Duration: g.PinnedTo(d, id),
						Label:    "Awakener Druid — becomes a 4/5 green Treefolk creature",
					}.Apply(ctx)
				}),
				effects.TargetPermanent("target Forest", effects.Land(), effects.HasSubtype("Forest"))),
		},
	})
}
