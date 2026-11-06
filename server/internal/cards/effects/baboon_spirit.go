package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Baboon Spirit — Creature — Monkey Spirit {2}{U}, 2/4 (Avatar: The Last
// Airbender Eternal, #5075):
//
//	"Whenever another nontoken Spirit you control enters, create a 1/1
//	 colorless Spirit creature token with "This token can't block or be
//	 blocked by non-Spirit creatures."
//	 {3}{U}: Exile another target creature you control. Return it to the
//	 battlefield under its owner's control at the beginning of the next
//	 end step."
//
// The token is Avatar Kuruk's printed Spirit (kurukSpiritToken), block
// rule included. "Spirit" is read through the layered view, so a
// changeling counts and a token Spirit does not trigger it.
//
// The activated ability is Angel of Condemnation's delayed blink
// (exileTargetsThenScheduleReturn) narrowed to another creature you
// control: the card returns under its owner's control as a new object at
// the next end step (CR 400.7, CR 603.7); a token exiled this way does
// not return.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "eb036e02-f43e-46a2-90f1-931579467c9b",
		Name:         "Baboon Spirit",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, avB30dAnotherNontokenSpiritYouControlEntered,
				"Baboon Spirit — create a 1/1 colorless Spirit",
				Do(CreateToken{Template: kurukSpiritToken(), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}{U}: Exile another target creature you control. Return it to the battlefield under its owner's control at the beginning of the next end step.",
			Cost:    ManaCost("{3}{U}"),
			Targets: Another(TargetCreature("another target creature you control", YouControl())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return exileTargetsThenScheduleReturn(NewContext(g, item), "Baboon Spirit — return the exiled creature")
			},
		}},
	})
}

// avB30dAnotherNontokenSpiritYouControlEntered is "whenever another
// nontoken Spirit you control enters". Spirit is a creature type, so a
// permanent that has it is a creature or a Kindred permanent; either
// counts. The type is a word of the source's text (CR 612.2).
func avB30dAnotherNontokenSpiritYouControlEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, true)
	return ok && !IsToken(c) && c.HasSubtype(source.TextSubtype("Spirit"))
}
