package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Eiganjo Exemplar — Enchantment Creature — Human Samurai {1}{W}, 2/1
// (Kamigawa: Neon Dynasty):
//
//	"Whenever a Samurai or Warrior you control attacks alone, it gets +1/+1
//	 until end of turn."
//
// Agents of S.H.I.E.L.D.'s pump on the lone attacker, which is the
// creature the event names; one that has left gets nothing.
//
// No simplification.
func init() {
	const name = "Eiganjo Exemplar"
	effects.Register(effects.Spec{
		OracleID:     "8e0f4e22-0677-4246-a4c5-6fcc74cb54a3",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, effects.NeoB09SamuraiOrWarriorAttacksAlone,
				name+" — the lone attacker gets +1/+1 until end of turn", eiganjoExemplarPump),
		},
	})
}

func eiganjoExemplarPump(g *game.Game, item *game.StackItem) error {
	attacker, ok := effects.NeoB09LoneAttacker(g, item)
	if !ok {
		return nil
	}
	return effects.BoostUntilEOT{Target: attacker.InstanceID, Power: 1, Toughness: 1, Label: "Eiganjo Exemplar — +1/+1"}.Apply(effects.NewContext(g, item))
}
