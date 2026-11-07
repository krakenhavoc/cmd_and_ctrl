package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sage of the Fang — Creature — Human Druid {2}{G}, 2/2 (Tarkir:
// Dragonstorm, #4463):
//
//	"When this creature enters, put a +1/+1 counter on target creature.
//	 Renew — {3}{G}, Exile this card from your graveyard: Put a +1/+1
//	 counter on target creature, then double the number of +1/+1 counters
//	 on that creature. Activate only as a sorcery."
//
// The enters trigger is Ivy Lane Denizen's body (one counter on the
// chosen creature, if still legal). Renew is renew.go, and its body is
// Invigorating Surge's: one counter, then as many more as the creature now
// has, read after the first counter lands, so a creature with none ends on
// two. A target that has left by resolution is illegal and nothing happens
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d6eb2bfc-1e86-4e36-9d48-12de16955b98",
		Name:         "Sage of the Fang",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Sage of the Fang — put a +1/+1 counter on target creature", plusOneCounterOnChosenTargets),
				TargetCreature("target creature")),
		},
		Activated: []ActivatedAbility{
			Renew("{3}{G}", "Put a +1/+1 counter on target creature, then double the number of +1/+1 counters on that creature.",
				TargetCreature("target creature"), func(g *game.Game, item *game.StackItem) error {
					return b34PutCounterThenDoubleCounters(NewContext(g, item))
				}),
		},
	})
}
