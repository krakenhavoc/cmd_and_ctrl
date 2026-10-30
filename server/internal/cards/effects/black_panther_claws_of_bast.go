package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Black Panther, Claws of Bast — Legendary Creature — Human Warrior Hero {2}{W},
// 3/3 (Marvel Super Heroes Commander, #4455):
//
//	"Lifelink
//	 Whenever a creature you control attacks alone, put a +1/+1 counter on
//	 it."
//
// Black Widow, Double Agent's trigger (mshB1YouAttackAlone): "alone" is
// game.AttackedAlone, read when the attack trigger is harvested, and the
// creature that gets the counter is the one that attacked, read off the event.
// A creature removed in response gets nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0a41221a-0cce-4aae-989a-24b70acca8db",
		Name:            "Black Panther, Claws of Bast",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, mshB1YouAttackAlone,
				"Black Panther, Claws of Bast — put a +1/+1 counter on the lone attacker",
				func(g *game.Game, item *game.StackItem) error {
					attacker := item.Trigger.Event.CardID
					if !onBattlefield(g, attacker) {
						return nil
					}
					return AddCounter{Target: attacker, Kind: game.CounterPlusOne, N: 1}.Apply(NewContext(g, item))
				}),
		},
	})
}
