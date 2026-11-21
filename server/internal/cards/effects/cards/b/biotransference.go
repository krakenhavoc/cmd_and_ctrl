package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Biotransference — Enchantment {2}{B}{B} (Warhammer 40,000 Commander, S82 GAP
// batch 196, #5948):
//
//	"Creatures you control are artifacts in addition to their other types. The
//	 same is true for creature spells you control and creature cards you own that
//	 aren't on the battlefield.
//	 Whenever you cast an artifact spell, you lose 1 life and create a 2/2 black
//	 Necron Warrior artifact creature token."
//
// Leyline of Transformation's three layer-4 statics (creatures you control,
// creature spells you control on the stack, creature cards you own off the
// battlefield) adding the artifact type once. Because creature spells are
// artifact spells while they are on the stack, every creature spell you cast
// triggers the second ability, as the card's rulings say. The cast trigger loses
// the life first, then makes the token (a step each, CR 608.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5443557e-440f-4fdd-a29b-7740d7cec085",
		Name:         "Biotransference",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			{
				Label:     "Creatures you control are artifacts in addition to their other types.",
				Layer:     game.Layer4Type,
				AppliesTo: effects.TribeFilter{YoursOnly: true}.Matches,
				Apply:     effects.AddTheArtifactTypeOnce,
			},
			{
				Label:        "Creature spells you control are artifacts in addition to their other types.",
				Layer:        game.Layer4Type,
				AffectsZones: effects.SpellsOnTheStack(),
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.Controller == source.Controller && target.IsCreature()
				},
				Apply: effects.AddTheArtifactTypeOnce,
			},
			{
				Label:        "Creature cards you own that aren't on the battlefield are artifacts in addition to their other types.",
				Layer:        game.Layer4Type,
				AffectsZones: effects.CardsOffTheBattlefield(),
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.Owner == source.Controller && target.IsCreature()
				},
				Apply: effects.AddTheArtifactTypeOnce,
			},
		},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Artifact(), "Biotransference — you lose 1 life and create a 2/2 black Necron Warrior artifact creature token",
				effects.Do(effects.YouLoseLife(1), effects.CreateToken{Template: effects.TokenCard("2/2 black Necron Warrior artifact"), N: 1})),
		},
	})
}
