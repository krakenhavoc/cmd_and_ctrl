package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Monument to Perfection — {2} Artifact (ONE):
//
//	"{3}, {T}: Search your library for a basic, Sphere, or Locus land
//	 card, reveal it, put it into your hand, then shuffle.
//	 {3}: This artifact becomes a 9/9 Phyrexian Construct artifact
//	 creature, loses all abilities, and gains indestructible and toxic 9.
//	 Activate only if there are nine or more lands with different names
//	 among the basic, Sphere, and Locus lands you control."
//
// The tutor is a SearchLibrary over a land that is basic or has the land
// type Sphere or Locus. The second ability has no duration, so its effect
// lasts as long as the object does (CR 611.2a): a layer 4 type change, a
// layer 7b 9/9 and a layer 6 removal of every ability that hands back
// indestructible and toxic 9 as the same effect's grant (CR 613.1f), pinned
// to the artifact. The gate is read as the ability is activated and is not
// rechecked on resolution (it's a restriction on activation, CR 602.5b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "df4bb104-d939-434f-8a1c-e89beda27c90",
		Name:         "Monument to Perfection",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{3}, {T}: Search your library for a basic, Sphere, or Locus land card, reveal it, put it into your hand, then shuffle.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player:    item.Controller,
						Predicate: effects.S82OneB18IsBasicSphereOrLocusLand,
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Reason:    "Monument to Perfection — a basic, Sphere, or Locus land card",
					}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label:     "{3}: This artifact becomes a 9/9 Phyrexian Construct artifact creature, loses all abilities, and gains indestructible and toxic 9. Activate only if there are nine or more lands with different names among the basic, Sphere, and Locus lands you control.",
				Purpose:   game.Purpose{Answers: game.AnswerAnimate},
				Cost:      effects.ManaCost("{3}"),
				Condition: effects.S82OneB18NineDifferentlyNamedLands,
				Effect: func(g *game.Game, item *game.StackItem) error {
					id := item.SourceCardID
					if !effects.OnBattlefield(g, id) {
						return nil
					}
					return effects.ScopedEffectFor{
						Target: id,
						Mods: []game.Mod{
							game.SetTypesMod([]string{"Artifact", "Creature"}, []string{"Phyrexian", "Construct"}),
							game.SetBasePowerMod(9),
							game.SetBaseToughnessMod(9),
							game.LoseAllAbilitiesMod("indestructible", "toxic 9"),
						},
						Duration: g.PinnedTo(game.IndefiniteDuration(), id),
						Label:    "Monument to Perfection — becomes a 9/9 Phyrexian Construct artifact creature with indestructible and toxic 9",
					}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}
