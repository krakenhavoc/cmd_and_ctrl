package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Katara, Bending Prodigy — Legendary Creature — Human Warrior Ally
// {2}{U}, 2/3:
//
//	"At the beginning of your end step, if Katara is tapped, put a
//	 +1/+1 counter on her.
//	 Waterbend {6}: Draw a card. (While paying a waterbend cost, you can
//	 tap your artifacts and creatures to help. Each one pays for {1}.)"
//
// The end-step trigger is an intervening "if" (CR 603.4): it triggers
// only while she is tapped and does nothing if she has untapped (or left)
// by the time it resolves (avatarB5CondThisIsTapped). The waterbend is
// Avatar Kuruk's cost: {6} of mana, any of it paid by tapping untapped
// artifacts and creatures you control, Katara herself included
// (CR 701.67a). Tapping her to waterbend on your turn is what sets up the
// counter.
//
// No simplifications. Avatar: The Last Airbender (#4459).
func init() {
	Register(Spec{
		OracleID:     "55adde23-c818-40fe-9bba-123de93568a5",
		Name:         "Katara, Bending Prodigy",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginEndStep, AllOf(ByYou, thisPermanentIsTapped),
				"Katara, Bending Prodigy — if she is tapped, put a +1/+1 counter on her",
				Do(If{Cond: avatarB5CondThisIsTapped{}, Then: []Applier{
					CounterOnThis{Kind: game.CounterPlusOne, N: 1},
				}})),
		},
		Activated: []ActivatedAbility{{
			Label:   "Waterbend {6}: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Cost:    WaterbendCost("{6}"),
			Effect:  Do(DrawCards{N: 1}),
		}},
	})
}
