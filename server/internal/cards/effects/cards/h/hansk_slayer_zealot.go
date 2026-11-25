package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hansk, Slayer Zealot — Legendary Creature — Human Archer {2}{R}{G}, 4/4
// (Secret Lair Drop):
//
//	"At the beginning of your upkeep, target opponent creates three Walker
//	 tokens. (They're 2/2 black Zombie creatures.)
//	 {T}: Hansk, Slayer Zealot deals 2 damage to target creature.
//	 Whenever a Zombie an opponent controls dies, draw a card."
//
// The Walkers enter under the target opponent's control. The draw trigger
// reads the Zombie's last-known creature types and controller (CR 603.10a);
// "Zombie" is a word of Hansk's text.
func init() {
	upkeep := effects.Targeting(effects.AtYourUpkeep("Hansk, Slayer Zealot — target opponent creates three Walker tokens",
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetPlayer {
					return effects.CreateToken{Controller: t.ID, Template: effects.TokenCard("2/2 black Zombie Walker"), N: 3}.Apply(ctx)
				}
			}
			return nil
		}), effects.TargetPlayer("target opponent", effects.Opponent()))
	effects.Register(effects.Spec{
		OracleID:     "d818f863-17ee-4387-94bf-d7e6e8ab590c",
		Name:         "Hansk, Slayer Zealot",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			upkeep,
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !effects.OldB11OpponentCreatureDied(ev, source, g) {
					return false
				}
				c, _ := g.LookupCardForEffect(ev.CardID)
				return effects.LeftAsSubtype(ev, c, source.TextSubtype("Zombie"))
			}, "Hansk, Slayer Zealot — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Hansk, Slayer Zealot deals 2 damage to target creature.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature"),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
			Effect:  effects.B33DamageChosenTargetFromSource(2),
		}},
	})
}
