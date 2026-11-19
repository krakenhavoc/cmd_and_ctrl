package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gornog, the Red Reaper — Legendary Creature — Minotaur Warrior {2}{R},
// 2/3 (Foundations Jumpstart):
//
//	"Haste
//	 Cowards can't block Warriors.
//	 Whenever one or more Warriors you control attack a player, target
//	 creature that player controls becomes a Coward.
//	 Attacking Warriors you control get +X/+0, where X is the number of
//	 Cowards your opponents control."
//
// Kargan Intimidator's block rule, with the becoming made permanent: the
// subtype is added with no duration (CR 611.2a), pinned to that object
// (CR 400.7), and it is a word of the card's text (CR 612.2). The attack
// trigger is Attuma's "attack a player" guard, once per player attacked
// (CR 603.2c), and its clause names a creature THAT player controls. The
// last line is a layer 7c boost over your attacking Warriors, counted live.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "adfd6c2d-a7f8-4f5c-bc31-d65c8150ed49",
		Name:            "Gornog, the Red Reaper",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		BlockRules: []game.BlockRule{
			CantBlockAttackers(OnMatching(OfCreatureType("Coward")), OnMatching(OfCreatureType("Warrior")),
				"Cowards can't block Warriors"),
		},
		Triggered: []game.TriggeredAbility{
			OncePerBatchPerPlayer(Targeting(
				On(game.EventAttack, s82BnsB51WarriorAttackAPlayer,
					"Gornog, the Red Reaper — target creature that player controls becomes a Coward",
					s82BnsB51BecomeCoward),
				ControlledByThatPlayer(TargetCreature("target creature that player controls")))),
		},
		Static: []game.StaticAbility{{
			Label:                    "Attacking Warriors you control get +X/+0, where X is the number of Cowards your opponents control.",
			Layer:                    game.Layer7PT,
			SubLayer:                 game.SubLayer7C_Modify,
			DependsOnAttackingStatus: true,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller &&
					target.AttackingTarget != uuid.Nil && target.HasSubtype(source.TextSubtype("Warrior"))
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power += s82BnsB51CowardsOpponentsControl(g, source)
			},
		}},
	})
}
