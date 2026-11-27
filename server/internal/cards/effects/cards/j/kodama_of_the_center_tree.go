package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kodama of the Center Tree — Legendary Creature — Spirit {4}{G}, */*
// (Betrayers of Kamigawa, S82 OLD, #5602):
//
//	"Kodama of the Center Tree's power and toughness are each equal to the
//	 number of Spirits you control.
//	 Kodama of the Center Tree has soulshift X, where X is the number of
//	 Spirits you control. (When this creature dies, you may return target
//	 Spirit card with mana value X or less from your graveyard to your
//	 hand.)"
//
// The first sentence is Battle Squadron's characteristic-defining ability
// (layer 7a, CR 604.3), the Kodama itself counted. Soulshift X is Soulshift
// N (soulshift.go) with N read from the board as the target is chosen, so
// the Kodama, already in the graveyard, is not among the Spirits counted
// (CR 702.46a). The "you may" is the CR 603.5 optional prompt; a graveyard
// with no Spirit card of mana value X or less puts nothing on the stack.
//
// No simplification.
func init() {
	const name = "Kodama of the Center Tree"
	effects.Register(effects.Spec{
		OracleID:     "88188cc1-3211-4bf8-a591-613b66192695",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Kodama of the Center Tree's power and toughness are each equal to the number of Spirits you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := effects.OldB601CountSubtypeWordYouControl(g, source.Controller, source.TextSubtype("Spirit"))
				c.Power, c.Toughness = n, n
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.Optional(effects.WhenThisDies(name+" — soulshift X", effects.ReturnFirstLegalGraveyardTargetToHand),
					name+" — return a Spirit card with mana value X or less from your graveyard to your hand?"),
				effects.TargetCardInGraveyard("target Spirit card with mana value X or less from your graveyard, where X is the number of Spirits you control",
					effects.YouOwn(), effects.HasSubtype("Spirit"),
					func(g *game.Game, caster uuid.UUID, c game.Card) bool {
						mv, ok := g.ManaValueForEffect(c)
						return ok && mv <= effects.OldB601SoulshiftX(g, caster)
					})),
		},
	})
}
