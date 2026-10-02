package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skeleton Ship — Legendary Creature {3}{U}{B}, 0/3:
//
//	"When you control no Islands, sacrifice Skeleton Ship.
//	 {T}: Put a -1/-1 counter on target creature."
//
// ADR 0107 §1 (#1858). The sacrifice is a CR 603.8 state trigger over
// the Islands its controller controls. The tap ability is an ordinary
// targeted activated ability; the counter is placed on resolution if the
// target is still legal (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8c85887d-5935-46ec-a075-dc9f5133bb96",
		Name:         "Skeleton Ship",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenYouControlNo(QuerySubtype("Island"), "Skeleton Ship — sacrifice it", SacrificeThisIfStillOnBattlefield),
		},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Put a -1/-1 counter on target creature.",
			Cost:    TapCost(),
			Targets: TargetCreature("target creature"),
			Effect:  putACounterOnTheTarget(game.CounterMinusOne),
		}},
	})
}
