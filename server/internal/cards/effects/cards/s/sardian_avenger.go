package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sardian Avenger — Creature — Goblin Warrior {1}{R}, 1/1 (The Brothers' War
// Commander, S82):
//
//	"First strike, trample
//	 Whenever this creature attacks, it gets +X/+0 until end of turn, where X
//	 is the number of artifacts your opponents control.
//	 Whenever an artifact an opponent controls is put into a graveyard from
//	 the battlefield, this creature deals 1 damage to that player."
//
// X is counted as the attack trigger resolves. The second trigger reads the
// artifact as it last existed and its controller as it left (CR 603.10a),
// tokens included; the damage goes to that player, from the Avenger as it
// last existed if it has gone (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6db90f86-205a-4e0b-944b-74742bc4e59d",
		Name:            "Sardian Avenger",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike", "trample"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Sardian Avenger — +X/+0 until end of turn, X the number of artifacts your opponents control",
				func(g *game.Game, item *game.StackItem) error {
					if !effects.OnBattlefield(g, item.SourceCardID) {
						return nil
					}
					n := effects.HobB7ArtifactsByOpponents(g, item.Controller)
					return effects.BoostUntilEOT{Target: item.SourceCardID, Power: n,
						Label: "Sardian Avenger — +X/+0"}.Apply(effects.NewContext(g, item))
				}),
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, controller, ok := effects.S82BroB08ArtifactLeft(ev, g)
				return ok && controller != source.Controller
			}, "Sardian Avenger — 1 damage to that player", func(g *game.Game, item *game.StackItem) error {
				if item.Trigger == nil {
					return nil
				}
				player, known := item.Trigger.Event.LeftUnderControlOf()
				if !known {
					return nil
				}
				return effects.DealDamage{Source: item.SourceCardID, Target: player, Amount: 1}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
