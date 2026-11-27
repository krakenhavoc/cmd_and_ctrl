package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Verdant Touch — Sorcery {1}{G} (Stronghold, S82 OLD b749, #5627):
//
//	"Buyback {3} (You may pay an additional {3} as you cast this spell. If you
//	 do, put this card into your hand as it resolves.)
//	 Target land becomes a 2/2 creature that's still a land. (This effect
//	 lasts indefinitely.)"
//
// Allay's buyback (CR 702.27a) on a scoped effect with no duration, so it lasts
// indefinitely (CR 611.2a): the creature type in layer 4 and the base 2/2 in
// layer 7b, with the land keeping its land type and abilities. It has no
// creature subtype and no colour of its own.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "68845200-5346-473d-becd-27315879443d",
		Name:          "Verdant Touch",
		Completeness:  effects.CompletenessFull,
		Targets:       effects.TargetPermanent("target land", effects.Land()),
		OptionalCosts: []game.AdditionalCost{effects.Buyback("{3}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				mods := append([]game.Mod{game.AddTypesMod("Creature")}, game.SetBasePTMods(2, 2)...)
				return effects.ScopedEffectFor{
					Target:   t.ID,
					Mods:     mods,
					Duration: game.IndefiniteDuration(),
					Label:    "Verdant Touch — becomes a 2/2 creature that's still a land",
				}.Apply(ctx)
			}
			return nil
		},
	})
}
