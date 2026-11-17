package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bright-Palm, Soul Awakener — Legendary Creature — Fox Shaman {1}{R}{G}{W},
// 4/3 (March of the Machine Commander, #5477):
//
//	"Backup 1 (When this creature enters, put a +1/+1 counter on target
//	 creature. If that's another creature, it gains the following ability
//	 until end of turn.)
//	 Whenever this creature attacks, double the number of +1/+1 counters on
//	 target creature. That creature can't be blocked by creatures with
//	 power 2 or less this turn."
//
// Backup (CR 702.165) grants the attack trigger, a bundle in Spec.Grants
// that Bright-Palm also has itself. The doubling puts as many +1/+1
// counters on the target as it has as the ability resolves (CR 701.10e);
// the evasion is the scoped block rule of Doorman (CantBeBlockedThisTurnByPower)
// on the same target, its power bound read live as blockers are declared.
//
// No simplification.
const brightPalmGrant = "bright-palm-soul-awakener/attack-double"

func init() {
	Register(Spec{
		OracleID:     "f74888ee-5086-49a6-ae74-fd9894f46332",
		Name:         "Bright-Palm, Soul Awakener",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:       brightPalmGrant,
			Triggered: []game.TriggeredAbility{brightPalmAttack()},
			Text:      "Whenever this creature attacks, double the number of +1/+1 counters on target creature. That creature can't be blocked by creatures with power 2 or less this turn.",
		}},
		Triggered: []game.TriggeredAbility{
			Backup("Bright-Palm, Soul Awakener", 1, brightPalmGrant),
			brightPalmAttack(),
		},
	})
}

func brightPalmAttack() game.TriggeredAbility {
	return Targeting(
		On(game.EventAttack, ThisAttacked,
			"Bright-Palm, Soul Awakener — double the +1/+1 counters on target creature; it can't be blocked by power 2 or less",
			func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					if c, ok := g.LookupCardForEffect(t.ID); ok {
						if n := c.Counters[game.CounterPlusOne]; n > 0 {
							if err := (AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: n}).Apply(ctx); err != nil {
								return err
							}
						}
					}
					return CantBeBlockedThisTurnByPower{Target: t.ID, AtMost: 2,
						Label: "Bright-Palm, Soul Awakener — can't be blocked by creatures with power 2 or less"}.Apply(ctx)
				}
				return nil
			}),
		TargetCreature("target creature"))
}
