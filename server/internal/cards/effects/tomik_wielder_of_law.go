package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tomik, Wielder of Law — Legendary Creature — Human Advisor {1}{W}{B},
// 2/4 (Secrets of Strixhaven Commander, #4834):
//
//	"Affinity for planeswalkers (This spell costs {1} less to cast for each
//	 planeswalker you control.)
//	 Flying, vigilance
//	 Whenever an opponent attacks with creatures, if two or more of those
//	 creatures are attacking you and/or planeswalkers you control, that
//	 opponent loses 3 life and you draw a card."
//
// Affinity is the CR 702.41a reduction over the planeswalkers its controller
// has. The trigger is one occurrence per attack declaration however many
// creatures swing (OncePerBatch), with its CR 603.4 intervening "if" checked
// as it triggers (On) and again as it resolves (InterveningIf): the opponent
// must still have two or more attackers at you and/or your planeswalkers. A
// creature attacking a battle, even one you protect, is not counted. "That
// opponent" is the player who declared the attack, so it works against a
// table of several opponents; the life is lost (not damage), then you draw.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2ef0961a-0543-4371-8c2a-333881b735e5",
		Name:            "Tomik, Wielder of Law",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		SelfCostModifiers: []game.CostModifier{
			AffinityFor("Affinity for planeswalkers", Planeswalker()),
		},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(InterveningIf(socB14StillAttackedTwice,
				On(game.EventAttack, socB14OpponentAttackedYouTwice,
					"Tomik, Wielder of Law — that opponent loses 3 life and you draw a card",
					Do(Instruction(socB14ThatOpponentLoses3), DrawCards{N: 1})))),
		},
	})
}
