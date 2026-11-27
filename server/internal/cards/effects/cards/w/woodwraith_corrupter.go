package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Woodwraith Corrupter — Creature — Elemental Horror {3}{B}{B}{G}, 3/6
// (Ravnica: City of Guilds, S82 OLD batch 574, #5600):
//
//	"{1}{B}{G}, {T}: Target Forest becomes a 4/4 black and green Elemental
//	 Horror creature. It's still a land."
//
// Roaring Earth's animation, for good (CR 611.2a): the Forest gains the
// Creature type and the Elemental Horror creature types (layer 4), becomes
// black and green (layer 5) and is set to 4/4 (layer 7b), keeping its land
// type and its mana ability. The effect is pinned to that Forest, so it ends
// when the Forest leaves the battlefield (CR 400.7). A Forest that arrived
// this turn can't attack (CR 302.6) like any creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "faddaa9d-3f70-49a2-81e6-22b16cf8e9fe",
		Name:         "Woodwraith Corrupter",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}{G}, {T}: Target Forest becomes a 4/4 black and green Elemental Horror creature. It's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    effects.Plus(effects.ManaCost("{1}{B}{G}"), effects.TapCost()),
			Targets: effects.TargetPermanent("target Forest", effects.Subtype("Forest")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				land, ok := g.LookupCardForEffect(id)
				if !ok || !effects.OnBattlefield(g, id) {
					return nil
				}
				mods := []game.Mod{
					game.AddTypesMod("Creature"),
					game.AddSubtypesMod(ctx.TextSubtypes("Elemental", "Horror")...),
					game.SetColorsMod(ctx.TextColors("B", "G")...),
				}
				mods = append(mods, game.SetBasePTMods(4, 4)...)
				g.RegisterScopedEffectForEffect(ctx.Source(),
					[]game.AffectedObject{game.PinObject(id, land.EnteredBattlefieldAt)}, mods,
					g.PinnedTo(game.IndefiniteDuration(), id),
					"Woodwraith Corrupter — becomes a 4/4 black and green Elemental Horror creature that's still a land")
				return nil
			},
		}},
	})
}
