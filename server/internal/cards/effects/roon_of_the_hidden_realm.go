package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Roon of the Hidden Realm — Legendary Creature — Rhino Soldier
// {2}{G}{W}{U}, 4/4:
//
//	"Vigilance, trample
//	 {2}, {T}: Exile another target creature. Return that card to the
//	 battlefield under its owner's control at the beginning of the next
//	 end step."
//
// The ability is Angel of Condemnation's first one at {2}: the creature
// is exiled as the ability resolves and a CR 603.7 delayed trigger
// returns it at the next end step (exileTargetsThenScheduleReturn), so
// an ability activated during an end step waits for the next one. The
// card comes back as a new object under its OWNER's control (CR 400.7),
// so blinking a creature stolen from an opponent hands it back, and a
// token exiled this way is gone (CR 111.8). "Another" is the
// object-level exclusion (Another), so Roon is never offered itself.
//
// Roon is a creature, so the {T} cost waits out summoning sickness
// (CR 302.6). Vigilance means attacking does not spend the tap.
//
// No simplification. Roon deck request (#4128).
func init() {
	Register(Spec{
		OracleID:        "fd336830-4a11-42b8-9fc7-d7526f569124",
		Name:            "Roon of the Hidden Realm",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance", "trample"},
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}: Exile another target creature. Return that card to the battlefield under its owner's control at the beginning of the next end step.",
			Cost:    Plus(ManaCost("{2}"), TapCost()),
			Targets: Another(TargetCreature("another target creature")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return exileTargetsThenScheduleReturn(NewContext(g, item),
					"Roon of the Hidden Realm — return the exiled creature")
			},
		}},
	})
}
