package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mount Doom — Legendary Land (EDHREC rank 1200):
//
//	"{T}, Pay 1 life: Add {B} or {R}.
//	 {1}{B}{R}, {T}: Mount Doom deals 1 damage to each opponent.
//	 {5}{B}{R}, {T}, Sacrifice Mount Doom and a legendary artifact:
//	 Choose up to two creatures, then destroy the rest. Activate only
//	 as a sorcery."
//
// Three abilities, all live. The mana is the Horizon Canopy shape
// (life as a COST, printed colours not narrowed). The ping is a
// two-component activated ability with the land itself as the
// source. The wipe's cost is the first in the catalog to sacrifice
// BOTH the source and another permanent — SacrificeThis plus a
// "legendary artifact" sacrifice clause, merged by Plus and validated
// together, so the Ring goes into the fire with the mountain — and
// the destruction is one simultaneous event over every creature not
// chosen.
//
// Sandbox simplification, declared (the Azorius Chancery posture):
// "choose up to two creatures" is a resolution-time choice, not a
// target, and the pick_target prompt is the one picker the engine has
// for choosing among permanents — so the two survivors are picked at
// announce as targets. Two consequences, both weaker than printed:
// opponents see the choice before the ability resolves, and a
// creature with hexproof or shroud cannot be chosen to survive, even
// your own. A third, the same way: if every chosen creature is gone
// by resolution the ability fizzles as a targeted one does (CR
// 608.2b), and the wipe does not happen.
func init() {
	Register(Spec{
		OracleID:     "995c8dac-fd27-468a-abd4-02372cf0c850",
		Name:         "Mount Doom",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The up-to-two creatures that survive the last ability are picked when it's activated rather than on resolution, so opponents can respond to the choice, and a creature with hexproof or shroud can't be picked.",
		},
		ManaAbilities: []ManaAbility{{
			Cost:                    ManaAbilityCost{Tap: true, Life: 1},
			Produced:                "{B|R}",
			Label:                   "{T}, Pay 1 life: Add {B} or {R}",
			IgnoreCommanderIdentity: true,
		}},
		Activated: []ActivatedAbility{
			{
				Label: "{1}{B}{R}, {T}: Mount Doom deals 1 damage to each opponent.",
				Cost:  Plus(ManaCost("{1}{B}{R}"), TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return damageToEachOpponent(g, item, 1)
				},
			},
			{
				Label:        "{5}{B}{R}, {T}, Sacrifice Mount Doom and a legendary artifact: Choose up to two creatures, then destroy the rest.",
				Cost:         Plus(ManaCost("{5}{B}{R}"), TapCost(), SacrificeThis(), b10SacrificeALegendaryArtifact()),
				Targets:      TargetCreature("up to two creatures to keep").WithCount(0, 2),
				SorcerySpeed: true,
				Effect: func(g *game.Game, item *game.StackItem) error {
					var keep []uuid.UUID
					for _, t := range item.Targets {
						if t.Kind == game.TargetCard {
							keep = append(keep, t.ID)
						}
					}
					return b10DestroyAllCreaturesExcept(NewContext(g, item), keep)
				},
			},
		},
	})
}
