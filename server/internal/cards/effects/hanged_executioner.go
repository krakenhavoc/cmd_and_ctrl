package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hanged Executioner — Creature — Spirit {2}{W}, 1/1 (EDHREC rank 6740):
//
//	"Flying
//	 When this creature enters, create a 1/1 white Spirit creature
//	 token with flying.
//	 {3}{W}, Exile this creature: Exile target creature."
//
// Two flyers for three, and later a Swords to Plowshares that costs
// the body. The exile ability is #1404's cost: ExileThis() paid from
// the BATTLEFIELD, so the Executioner leaves the battlefield at
// announce — it does not die, so a "whenever a creature you control
// dies" payoff never sees it, while a "leaves the battlefield" watcher
// always does. The target is chosen before the cost is paid (CR 601.2c
// before 601.2h), so naming the Executioner itself is legal and the
// ability then has nothing left to exile; the rules allow it and so
// does this.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6ff4ff67-ad08-447f-a112-1a071c1474a4",
		Name:            "Hanged Executioner",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Hanged Executioner — create a 1/1 white Spirit with flying",
				Do(CreateToken{Template: b28WhiteSpiritFlyingToken(), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}{W}, Exile this creature: Exile target creature.",
			Cost:    Plus(ManaCost("{3}{W}"), ExileThis()),
			Targets: TargetCreature("target creature"),
			Effect:  exileFirstLegalCardTarget,
		}},
	})
}
