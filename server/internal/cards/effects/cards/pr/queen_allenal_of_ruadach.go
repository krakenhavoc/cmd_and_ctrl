package cardspr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Queen Allenal of Ruadach — Legendary Creature — Elf Noble {G}{W}{W}, */*
// (Dominaria United):
//
//	"Queen Allenal of Ruadach's power and toughness are each equal to the
//	 number of creatures you control.
//	 If one or more creature tokens would be created under your control,
//	 those tokens plus a 1/1 white Soldier creature token are created
//	 instead."
//
// Crusader of Odric's characteristic-defining ability (layer 7a), the Queen
// counted. The replacement is Chatterfang's additive group (CR 614.1,
// 701.7b), but it applies only to a creation that makes at least one
// CREATURE token (a Treasure alone leaves it cold) and adds exactly ONE
// Soldier however many tokens the instruction makes. The instruction's own
// groups are left alone, so a Treasure made beside a Goblin stays a Treasure.
// Two Queens each add their own Soldier.
//
// No simplification.
func init() {
	isCreatureToken := func(t game.Card) bool { return t.IsCreature() }
	effects.Register(effects.Spec{
		OracleID:     "5f04d4b0-2c75-4f2d-97d5-96220b624b17",
		Name:         "Queen Allenal of Ruadach",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Queen Allenal of Ruadach's power and toughness are each equal to the number of creatures you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := effects.B04CreaturesControlled(g, source.Controller)
				c.Power, c.Toughness = n, n
			},
		}},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventTokenCreated},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return ev.Kind == game.RepEventCreateTokens &&
					ev.TokenController == src.Controller && ev.TokenCount() > 0 &&
					ev.TokenTemplatesMatch(isCreatureToken)
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) error {
				ev.TokenGroups = append(ev.TokenGroups, game.TokenGroup{
					Template: src.TokenThroughText(effects.TokenCard("1/1 white Soldier")), Count: 1,
				})
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Queen Allenal of Ruadach: those tokens plus a 1/1 white Soldier",
		}},
	})
}
