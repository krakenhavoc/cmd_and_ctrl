package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Imperious Mindbreaker — {1}{U}{U} Creature — Human Wizard 1/4
// (Innistrad: Crimson Vow Commander, #5675):
//
//	"Soulbond (You may pair this creature with another unpaired creature
//	 when either enters. They remain paired for as long as you control
//	 both of them.)
//	 As long as Imperious Mindbreaker is paired with another creature,
//	 each of those creatures has "Whenever this creature attacks, each
//	 opponent mills cards equal to its toughness.""
//
// Soulbond is the keyword's two triggers (CR 702.95a, soulbond.go). The
// second line is a layer-6 grant to both creatures of the pair, so each
// has its own attack trigger and each counts its own toughness: as the
// ability resolves, or as the creature last had it if it has left the
// battlefield by then (CR 608.2h). Every opponent mills, not only the
// player attacked.
//
// No simplification.
const imperiousMindbreakerGrant = "imperious-mindbreaker/mill"

func init() {
	effects.Register(effects.Spec{
		OracleID:        "95157a46-6e0a-4800-93d3-46d2842464db",
		Name:            "Imperious Mindbreaker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordSoulbond},
		Triggered:       effects.Soulbond("Imperious Mindbreaker"),
		Grants: []effects.AbilityGrant{{
			Key: imperiousMindbreakerGrant,
			Triggered: []game.TriggeredAbility{
				effects.WheneverThisAttacks("Imperious Mindbreaker — each opponent mills cards equal to this creature's toughness",
					func(g *game.Game, item *game.StackItem) error {
						return effects.B12EachOpponentMills(g, item, effects.ToughnessNowOrLastKnown(g, item.SourceCardID))
					}),
			},
			Text: "Whenever this creature attacks, each opponent mills cards equal to its toughness.",
		}},
		Static: []game.StaticAbility{
			effects.PairedCreaturesHave("As long as Imperious Mindbreaker is paired with another creature, each of those creatures has \"Whenever this creature attacks, each opponent mills cards equal to its toughness.\"",
				imperiousMindbreakerGrant),
		},
	})
}
