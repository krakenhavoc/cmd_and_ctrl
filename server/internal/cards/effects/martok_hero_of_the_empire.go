package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Martok, Hero of the Empire — Legendary Creature — Klingon Warrior
// {2}{W}, 2/2 (Star Trek Commander, #4974):
//
//	"Battlefront — Whenever you attack, put a +1/+1 counter on each of up
//	 to X target creatures you control, where X is the number of opponents
//	 you attacked this turn."
//
// Battlefront is an ability word. "Whenever you attack" is one trigger
// per attack declaration, whatever the attackers attacked: the ability
// waits for the declaration's batch to settle (ThatManyEvents, the "you
// attack with N or more" shape) and is built once. X fixes how many
// targets may be chosen, so it is counted as the ability is put on the
// stack (CR 601.2c by way of CR 603.3d), after every attacker of the
// declaration has been recorded:
// each opponent still in the game that a creature of yours attacked this
// turn counts once (OpponentsYouAttackedThisTurn), and an attack only on
// planeswalkers or battles makes X zero, so nothing is targeted. Each
// target still legal as it resolves gets one counter.
//
// No simplification.
func init() {
	attack := withAtBatchEnd(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
		return attackDeclaredByYou(ev, source.Controller)
	}, "Martok — battlefront: put a +1/+1 counter on each of up to X target creatures you control",
		putPlusOneCounterOnEachLegalTarget), ThatManyEvents())
	attack.TargetsFrom = func(_ game.TriggerContext, source *game.Card, g *game.Game) *game.TargetSpec {
		x := OpponentsYouAttackedThisTurn(g, source.Controller)
		if x <= 0 {
			return nil
		}
		return TargetCreature("up to X target creatures you control", YouControl()).WithCount(0, x)
	}
	Register(Spec{
		OracleID:     "d17c3079-372d-4e4f-8214-12a9893088ff",
		Name:         "Martok, Hero of the Empire",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{attack},
	})
}
