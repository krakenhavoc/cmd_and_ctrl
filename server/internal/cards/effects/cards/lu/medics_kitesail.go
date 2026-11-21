package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Medic's Kitesail — Artifact — Equipment {2} (Reality Fracture):
//
//	"Equipped creature gets +1/+0 and has flying and "Whenever this
//	 creature attacks, you gain 1 life."
//	 Equip {2}"
//
// The life-gain trigger is the EQUIPPED creature's (ADR 0093), so it is
// controlled by whoever controls the creature, and moving the Kitesail
// moves it.
//
// No simplification.
const medicsKitesailGrant = "medics-kitesail/gain-life"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "eacfe895-5106-437c-b8c1-2ff03c107744",
		Name:         "Medic's Kitesail",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: medicsKitesailGrant,
			Triggered: []game.TriggeredAbility{
				effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return effects.AttackDeclared(ev, source)
				}, "Medic's Kitesail — you gain 1 life", func(g *game.Game, item *game.StackItem) error {
					return effects.GainLife{Player: item.Controller, Amount: 1}.Apply(effects.NewContext(g, item))
				}),
			},
			Text: "Whenever this creature attacks, you gain 1 life.",
		}},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 0),
			effects.GrantToAttached("flying"),
			effects.GrantAbilitiesToAttached(medicsKitesailGrant),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
