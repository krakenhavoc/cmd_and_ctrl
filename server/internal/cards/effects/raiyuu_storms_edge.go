package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Raiyuu, Storm's Edge — Legendary Creature — Human Samurai {2}{R}{W}, 3/3
// (Kamigawa: Neon Dynasty):
//
//	"First strike
//	 Whenever a Samurai or Warrior you control attacks alone, untap it. If
//	 it's the first combat phase of the turn, there is an additional combat
//	 phase after this phase."
//
// The lone attacker is untapped (nothing if it has left); the extra combat
// is the shared ExtraCombatAfterThisPhase, asked as the trigger resolves
// (an instruction's own "if", not an intervening one).
//
// No simplification.
func init() {
	const name = "Raiyuu, Storm's Edge"
	Register(Spec{
		OracleID:        "9028f726-50c7-49f3-8386-e584415dcba2",
		Name:            name,
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, neoB09SamuraiOrWarriorAttacksAlone,
				name+" — untap it; if it's the first combat phase, an additional combat phase follows",
				raiyuuUntapAndExtraCombat),
		},
	})
}

func raiyuuUntapAndExtraCombat(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if attacker, ok := neoB09LoneAttacker(g, item); ok {
		if err := (UntapTarget{Target: attacker.InstanceID}).Apply(ctx); err != nil {
			return err
		}
	}
	if !IsFirstCombatPhase(g) {
		return nil
	}
	return ExtraCombatAfterThisPhase().Apply(ctx)
}
