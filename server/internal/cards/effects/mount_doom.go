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
// "Choose up to two creatures" is made as the ability RESOLVES
// (CR 608.2c), not as targets: one pick over every creature on the
// battlefield with a floor of zero, so opponents cannot answer the
// choice, a creature with hexproof or shroud can be kept, and the wipe
// still happens if the creatures you would have kept are gone.
func init() {
	Register(Spec{
		OracleID:     "995c8dac-fd27-468a-abd4-02372cf0c850",
		Name:         "Mount Doom",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true, Life: 1},
			Produced: "{B|R}",
			Label:    "{T}, Pay 1 life: Add {B} or {R}",
		}},
		Activated: []ActivatedAbility{
			{
				Label:   "{1}{B}{R}, {T}: Mount Doom deals 1 damage to each opponent.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    Plus(ManaCost("{1}{B}{R}"), TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return damageToEachOpponent(g, item, 1)
				},
			},
			{
				Label:        "{5}{B}{R}, {T}, Sacrifice Mount Doom and a legendary artifact: Choose up to two creatures, then destroy the rest.",
				Cost:         Plus(ManaCost("{5}{B}{R}"), TapCost(), SacrificeThis(), b10SacrificeALegendaryArtifact()),
				SorcerySpeed: true,
				Effect: func(g *game.Game, item *game.StackItem) error {
					var creatures []uuid.UUID
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.IsCreature() {
							creatures = append(creatures, c.InstanceID)
						}
					}
					if len(creatures) == 0 {
						return nil
					}
					g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
						Chooser:  item.Controller,
						Source:   item.SourceCardID,
						Question: "Mount Doom — choose up to two creatures; destroy the rest",
						Cards:    creatures,
						Min:      0,
						Max:      2,
						Zone:     game.ZoneBattlefield,
						Then: func(g *game.Game, keep []uuid.UUID) error {
							return b10DestroyAllCreaturesExcept(NewContext(g, item), keep)
						},
					})
					return nil
				},
			},
		},
	})
}
