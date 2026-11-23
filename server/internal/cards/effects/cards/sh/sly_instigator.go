package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sly Instigator — Creature — Human Wizard {3}{U}, 2/4 (Commander Legends:
// Battle for Baldur's Gate, #5948):
//
//	"{U}, {T}: Until your next turn, target creature an opponent controls
//	 can't be blocked. Goad that creature."
//
// The restriction is Mu Yanling's until-your-next-turn scoped effect (CR
// 611.2b) carrying the cannot-be-blocked bit; the goad is the engine's own,
// which also ends as its controller's next turn begins (CR 701.15a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8c2dd8dc-8623-4518-b3fd-dcb9f647fa98",
		Name:         "Sly Instigator",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}, {T}: Until your next turn, target creature an opponent controls can't be blocked. Goad that creature.",
			Cost:    effects.Plus(effects.ManaCost("{U}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				if err := (effects.ScopedEffectFor{
					Target:   id,
					Mods:     []game.Mod{game.AddRestrictionsMod(game.CantBeBlocked)},
					Duration: effects.DurationUntilYourNextTurn(ctx, item.Controller),
					Label:    "Sly Instigator — can't be blocked until your next turn",
				}).Apply(ctx); err != nil {
					return err
				}
				return effects.GoadTarget(g, item)
			},
		}},
	})
}
