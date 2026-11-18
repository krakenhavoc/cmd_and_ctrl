package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nebelgast Herald — Creature — Spirit {2}{U}, 2/1:
//
//	"Flash
//	 Flying
//	 Whenever this creature or another Spirit you control enters, tap target
//	 creature an opponent controls."
//
// One trigger that watches every entry under your control: the Herald's own
// and any other Spirit's, judged on the entering card's layered types. The
// target is chosen as the ability goes on the stack (CR 603.3d) and
// re-checked as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "cf336e4c-a0d3-43aa-ad23-b98effb2b751",
		Name:            "Nebelgast Herald",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.HasSubtype(source.TextSubtype("Spirit"))
			}, "Nebelgast Herald — tap target creature an opponent controls",
				b36TapChosenCreature),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}
