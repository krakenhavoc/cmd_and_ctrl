package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hedron Blade — Artifact — Equipment {1} (Battle for Zendikar, #5547):
//
//	"Equipped creature gets +1/+1.
//	 Whenever equipped creature becomes blocked by one or more colorless
//	 creatures, it gains deathtouch until end of turn.
//	 Equip {2}"
//
// One trigger per blocked attacker (EventBecomesBlocked), however many
// colorless creatures block it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cf7ca005-410d-40c8-aae4-89f41f81848b",
		Name:         "Hedron Blade",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(1, 1)},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesBlocked, effects.S82OldB314BlockedByColorless,
				"Hedron Blade — equipped creature gains deathtouch until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.GrantKeywordUntilEOT{
						Target:   ctx.Trigger().Event.CardID,
						Keywords: []string{"deathtouch"},
						Label:    "Hedron Blade — deathtouch",
					}.Apply(ctx)
				}),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
