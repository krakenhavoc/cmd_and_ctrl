package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "32e7385e-525e-4925-8cfd-bfea21fcd1bf",
		Name:            name,
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, neoB09SamuraiOrWarriorAttacksAlone,
				name+" — the lone attacker gets +1/+0 until end of turn for each Samurai or Warrior you control",
				asariCaptainPump),
		},
	})
}

func asariCaptainPump(g *game.Game, item *game.StackItem) error {
	attacker, ok := neoB09LoneAttacker(g, item)
	if !ok {
		return nil
	}
	ctx := NewContext(g, item)
	n := len(MatchingBattlefield(ctx, And(YouControl(), neoB09SamuraiOrWarrior(ctx))))
	return BoostUntilEOT{Target: attacker.InstanceID, Power: n, Label: "Asari Captain — +1/+0 for each Samurai or Warrior"}.Apply(ctx)
}
