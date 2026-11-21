package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Star Whale — Creature — Alien Whale {6}{U}{U}, 8/8 (Doctor Who, #5948):
//
//	"Flying, vigilance
//	 Other creatures you control have ward {2}.
//	 Suspend 6—{1}{U} (Rather than cast this card from your hand, you may
//	 pay {1}{U} and exile it with six time counters on it. At the beginning of
//	 your upkeep, remove a time counter. When the last is removed, you may
//	 cast it without paying its mana cost. It has haste.)"
//
// Flying and vigilance are printed keywords; the ward grant is WardGranted
// over the other creatures its controller has (Long River Lurker's shape);
// suspend is the keyword's special action (Suspend).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3e84d719-bb76-43a4-b61a-fee30c3aac68",
		Name:            "Star Whale",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		SpecialActions:  []game.SpecialAction{Suspend(6, "{1}{U}")},
		Triggered: []game.TriggeredAbility{
			WardGranted(WardMana("{2}"), "Star Whale — other creatures you control have ward {2}",
				func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.InstanceID != source.InstanceID && target.Controller == source.Controller && target.IsCreature()
				}),
		},
	})
}
