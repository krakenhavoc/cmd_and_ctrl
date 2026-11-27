package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Yuan Shao's Infantry — Creature — Human Soldier {3}{R}, 2/2
// (Portal Three Kingdoms, #5622):
//
//	"Whenever this creature attacks alone, this creature can't be blocked this combat."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8ad91015-db7f-43ea-97c9-bf2bf5656858",
		Name:         "Yuan Shao's Infantry",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.AttackDeclared(ev, source) && game.AttackedAlone(g)
			}, yuanShaosInfantryLabel, func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				d, ok := effects.DurationUntilEndOfCombat(ctx)
				if !ok || !effects.OnBattlefield(g, item.SourceCardID) {
					return nil
				}
				return effects.ScopedEffectFor{
					Target:   item.SourceCardID,
					Mods:     []game.Mod{game.AddRestrictionsMod(game.CantBeBlocked)},
					Duration: d,
					Label:    yuanShaosInfantryLabel,
				}.Apply(ctx)
			}),
		},
	})
}

const yuanShaosInfantryLabel = "Yuan Shao's Infantry — it can't be blocked this combat"
