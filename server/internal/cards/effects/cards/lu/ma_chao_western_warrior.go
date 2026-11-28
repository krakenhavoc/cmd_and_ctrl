package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ma Chao, Western Warrior — Legendary Creature — Human Soldier Warrior {3}{R}{R}, 3/3
// (Portal Three Kingdoms, #5622):
//
//	"Horsemanship (This creature can't be blocked except by creatures with horsemanship.)"
//	"Whenever Ma Chao attacks alone, it can't be blocked this combat."
//
// Horsemanship is an engine keyword carried as printed data: the block
// check refuses any blocker without horsemanship. It doesn't stop this
// creature blocking.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "023e26d8-b67d-43dd-89b1-8905b1183094",
		Name:            "Ma Chao, Western Warrior",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"horsemanship"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.AttackDeclared(ev, source) && game.AttackedAlone(g)
			}, machaoLabel, func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				d, ok := effects.DurationUntilEndOfCombat(ctx)
				if !ok || !effects.OnBattlefield(g, item.SourceCardID) {
					return nil
				}
				return effects.ScopedEffectFor{
					Target:   item.SourceCardID,
					Mods:     []game.Mod{game.AddRestrictionsMod(game.CantBeBlocked)},
					Duration: d,
					Label:    machaoLabel,
				}.Apply(ctx)
			}),
		},
	})
}

const machaoLabel = "Ma Chao — it can't be blocked this combat"
