package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Su-Chi Cave Guard — Artifact Creature — Construct {8}, 8/8 (The
// Brothers' War):
//
//	"Vigilance
//	 Ward {4}
//	 When this creature dies, add eight {C}. Until end of turn, you don't
//	 lose this mana as steps and phases end."
//
// Ward is the engine's ward trigger (Aboleth Spawn's); the mana is
// Branch of Vitu-Ghazi's keep-until-end-of-turn rider on {C}x8, added by
// the trigger's controller whether or not the Cave Guard is still around
// to be looked at (CR 608.2h).
func init() {
	Register(Spec{
		OracleID:        "28264920-9d14-4661-83b8-56f950fbff39",
		Name:            "Su-Chi Cave Guard",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{4}"), "Su-Chi Cave Guard — ward {4}"),
			WhenThisDies("Su-Chi Cave Guard — add eight {C}; you don't lose it as steps and phases end",
				Do(AddMana{
					Produced: "{C}{C}{C}{C}{C}{C}{C}{C}",
					Riders:   []game.ManaSpendRider{KeepManaUntilEndOfTurn()},
				})),
		},
	})
}
