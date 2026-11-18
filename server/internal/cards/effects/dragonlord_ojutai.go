package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dragonlord Ojutai — Legendary Creature — Elder Dragon {3}{W}{U}, 5/4:
//
//	"Flying
//	 Dragonlord Ojutai has hexproof as long as it's untapped.
//	 Whenever Dragonlord Ojutai deals combat damage to a player, look at
//	 the top three cards of your library. Put one of them into your hand
//	 and the rest on the bottom of your library in any order."
//
// The hexproof is a layer 6 grant to itself that applies only while it is
// untapped, re-read on every layer pass (Illusion Spinners), so attacking
// without vigilance removes it at once. The dig is Impulse's, over three
// cards (LookAtTopThenTakeOneRestOnBottom).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "654ff673-4630-4068-9bcf-e88a7877eb58",
		Name:            "Dragonlord Ojutai",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			func() game.StaticAbility {
				ab := KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.InstanceID == source.InstanceID && !target.Tapped
				}, "hexproof")
				ab.Label = "Dragonlord Ojutai has hexproof as long as it's untapped."
				return ab
			}(),
		},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Dragonlord Ojutai — look at the top three cards of your library; put one into your hand and the rest on the bottom in any order",
				LookAtTopThenTakeOneRestOnBottom(3, "Dragonlord Ojutai — put one of them into your hand")),
		},
	})
}
