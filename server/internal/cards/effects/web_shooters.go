package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "ec528198-a38a-4ef7-9804-78a25c94ecc1",
		Name:         "Web-Shooters",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: webShootersGrant,
			Triggered: []game.TriggeredAbility{
				Targeting(
					On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
						return attackDeclared(ev, source)
					}, "Web-Shooters — tap target creature an opponent controls",
						func(g *game.Game, item *game.StackItem) error {
							ctx := NewContext(g, item)
							target, ok := b16FirstLegalTargetCard(ctx)
							if !ok {
								return nil
							}
							return TapTarget{Target: target}.Apply(ctx)
						}),
					TargetCreature("target creature an opponent controls", OpponentControls())),
			},
			Text: "Whenever this creature attacks, tap target creature an opponent controls.",
		}},
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			GrantToAttached("reach"),
			GrantAbilitiesToAttached(webShootersGrant),
		},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}
