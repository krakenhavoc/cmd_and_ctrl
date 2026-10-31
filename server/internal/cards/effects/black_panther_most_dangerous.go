package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Black Panther, Most Dangerous — {2}{W} Legendary Creature — Human
// Warrior Hero 3/3 (Marvel Super Heroes Commander, #4455):
//
//	"Whenever Black Panther is dealt damage, he deals that much damage
//	 to any other target.
//	 Power-up — {5}{W}{W}: Put two +1/+1 counters on Black Panther.
//	 Other creatures you control get +2/+2 until end of turn.
//	 (Activate each power-up ability only once. Reduce the cost by his
//	 mana cost if he entered this turn.)"
//
// The trigger is Brash Taunter's reflect (b35SelfWasDealtDamage, the
// batch's total via ThatMuchDamage, damagedCreatureReflectsChosenAmount)
// aimed at any target but the Panther himself (Another). The power-up
// is Brave Brawler's counters plus a +2/+2 pump on the other creatures
// the controller has as it resolves (CR 611.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7ba593fd-93db-452f-ad80-7f0f37afb09d",
		Name:         "Black Panther, Most Dangerous",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return b35SelfWasDealtDamage(ev, source)
			},
			Targets:    Another(TargetAny()),
			Key:        "Black Panther — he deals that much damage to any other target",
			AtBatchEnd: ThatMuchDamage(),
			Effect:     damagedCreatureReflectsChosenAmount,
		}},
		Activated: []ActivatedAbility{
			powerUpAnswering("{5}{W}{W}: Put two +1/+1 counters on Black Panther. Other creatures you control get +2/+2 until end of turn.",
				ManaCost("{5}{W}{W}"), blackPantherPowerUp),
		},
	})
}

func blackPantherPowerUp(g *game.Game, item *game.StackItem) error {
	if err := plusOneCountersOnThis(2)(g, item); err != nil {
		return err
	}
	return BoostUntilEOT{
		Match:     And(Creature(), YouControl(), NotSelf(item.SourceCardID)),
		Power:     2,
		Toughness: 2,
		Label:     "Black Panther — other creatures you control get +2/+2",
	}.Apply(NewContext(g, item))
}
