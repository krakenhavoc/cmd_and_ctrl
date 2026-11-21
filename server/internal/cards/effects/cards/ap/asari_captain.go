package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Asari Captain — Creature — Human Samurai {3}{R}{W}, 4/3 (Kamigawa: Neon
// Dynasty):
//
//	"Haste
//	 Whenever a Samurai or Warrior you control attacks alone, it gets +1/+0
//	 until end of turn for each Samurai or Warrior you control."
//
// The count is taken as the trigger resolves (CR 611.2c), over every
// Samurai or Warrior its controller controls, the attacker and the Captain
// included.
//
// No simplification.
func init() {
	const name = "Asari Captain"
	effects.Register(effects.Spec{
		OracleID:        "32e7385e-525e-4925-8cfd-bfea21fcd1bf",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, effects.NeoB09SamuraiOrWarriorAttacksAlone,
				name+" — the lone attacker gets +1/+0 until end of turn for each Samurai or Warrior you control",
				asariCaptainPump),
		},
	})
}

func asariCaptainPump(g *game.Game, item *game.StackItem) error {
	attacker, ok := effects.NeoB09LoneAttacker(g, item)
	if !ok {
		return nil
	}
	ctx := effects.NewContext(g, item)
	n := len(effects.MatchingBattlefield(ctx, effects.And(effects.YouControl(), effects.NeoB09SamuraiOrWarrior(ctx))))
	return effects.BoostUntilEOT{Target: attacker.InstanceID, Power: n, Label: "Asari Captain — +1/+0 for each Samurai or Warrior"}.Apply(ctx)
}
