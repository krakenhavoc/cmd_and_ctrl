package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Voltaic Whip — Artifact — Equipment {2}{B} (Marvel Super Heroes
// Commander, #4455):
//
//	"Equipped creature gets +2/+0 and has "Whenever this creature attacks
//	 alone, you draw a card and lose 1 life."
//	 Equip {2}"
//
// Bilbo's Ring's draw-and-lose-1 as a trigger the equipped creature has
// (ADR 0093, Conformer Shuriken's shape). "Alone" is game.AttackedAlone,
// read when the attack trigger is harvested, after every attacker of the
// declaration is stamped (Luke Cage); the attacker must be the equipped
// creature itself, and "you" is its controller.
//
// No simplification.
const voltaicWhipGrant = "voltaic-whip/attack-alone-draw"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "c077f939-73a9-455f-8de5-385162121b94",
		Name:         "Voltaic Whip",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: voltaicWhipGrant,
			Triggered: []game.TriggeredAbility{
				effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return effects.AttackDeclared(ev, source) && game.AttackedAlone(g)
				}, "Voltaic Whip — you draw a card and lose 1 life", effects.Do(effects.DrawCards{N: 1}, effects.GainLife{Amount: -1})),
			},
			Text: "Whenever this creature attacks alone, you draw a card and lose 1 life.",
		}},
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 0),
			effects.GrantAbilitiesToAttached(voltaicWhipGrant),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
