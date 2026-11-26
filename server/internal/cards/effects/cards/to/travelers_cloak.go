package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Traveler's Cloak — Enchantment — Aura {2}{U} (Conspiracy, S82 OLD batch
// 362, #5557):
//
//	"Enchant creature
//	 As this Aura enters, choose a land type.
//	 When this Aura enters, draw a card.
//	 Enchanted creature has landwalk of the chosen type."
//
// The choice is the CR 614.12 anchor word over every CR 205.3i land type
// (Convincing Mirage's shape); one layer-6 grant per type, each gated on that
// type being the chosen one, gives the creature "<type>walk". The five basic
// types and Desert are the engine's own landwalk tokens; the others are the
// plain parameterised "<land type>walk" token (game.LandwalkToken). The draw
// is the Aura's enters trigger.
//
// No simplification.
func init() {
	statics := make([]game.StaticAbility, 0, len(game.LandTypes))
	for _, t := range game.LandTypes {
		statics = append(statics, effects.StaticWhenChosen(t, effects.GrantToAttached(game.LandwalkToken(t))))
	}
	effects.Register(effects.Spec{
		OracleID:     "5241182d-e2ca-4b31-809f-fbff2822dbf6",
		Name:         "Traveler's Cloak",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit, Draws: 1},
		AsEnters:     effects.ChooseOptionAsEnters("Traveler's Cloak", game.LandTypes...),
		Static:       statics,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Traveler's Cloak — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
