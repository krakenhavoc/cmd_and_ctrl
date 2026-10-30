package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "c077f939-73a9-455f-8de5-385162121b94",
		Name:         "Voltaic Whip",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: voltaicWhipGrant,
			Triggered: []game.TriggeredAbility{
				On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return attackDeclared(ev, source) && game.AttackedAlone(g)
				}, "Voltaic Whip — you draw a card and lose 1 life", Do(DrawCards{N: 1}, GainLife{Amount: -1})),
			},
			Text: "Whenever this creature attacks alone, you draw a card and lose 1 life.",
		}},
		Static: []game.StaticAbility{
			PumpAttached(2, 0),
			GrantAbilitiesToAttached(voltaicWhipGrant),
		},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}
