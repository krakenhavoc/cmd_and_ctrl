package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Falcon, Joaquin Torres — Legendary Creature — Bird Human Hero {2}{W},
// 2/2 (Marvel Super Heroes Commander, #4455):
//
//	"Flying, lifelink
//	 Battalion — Whenever Falcon and at least two other creatures attack,
//	 put a +1/+1 counter on him and scry 1."
//
// Battalion is the same group test as Boros Strike-Captain's: every
// attacker is staged before the declaration is announced, so Falcon's own
// attack event already sees the other two.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "46399b09-1b5c-4c28-9134-cc2ed573b6aa",
		Name:            "Falcon, Joaquin Torres",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attackDeclared(ev, source) && attackersYouControlOtherThan(g, source.Controller, source.InstanceID) >= 2
			}, "Falcon, Joaquin Torres — battalion: put a +1/+1 counter on him and scry 1",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1}, Scry{N: 1})),
		},
	})
}
