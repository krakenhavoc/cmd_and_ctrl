package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Web-Shooters — {1}{W} Artifact — Equipment (Marvel's Spider-Man, #4460):
//
//	"Equipped creature gets +1/+1 and has reach and "Whenever this creature
//	 attacks, tap target creature an opponent controls."
//	 Equip {2}"
//
// The granted trigger is the equipped creature's own (ADR 0093, Conformer
// Shuriken's shape) with a plain target of a creature an opponent controls.
//
// No simplification.
const webShootersGrant = "web-shooters/tap-on-attack"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "ec528198-a38a-4ef7-9804-78a25c94ecc1",
		Name:         "Web-Shooters",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: webShootersGrant,
			Triggered: []game.TriggeredAbility{
				effects.Targeting(
					effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
						return effects.AttackDeclared(ev, source)
					}, "Web-Shooters — tap target creature an opponent controls",
						effects.OldB319TapTheTargetCreature),
					effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
			},
			Text: "Whenever this creature attacks, tap target creature an opponent controls.",
		}},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 1),
			effects.GrantToAttached("reach"),
			effects.GrantAbilitiesToAttached(webShootersGrant),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
