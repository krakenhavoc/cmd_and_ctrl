package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tandem Lookout — {2}{U} Creature — Human Scout 2/1 (Shadows over Innistrad
// via Duel Decks: Blessed vs. Cursed, S82 OLD b294):
//
//	"Soulbond (You may pair this creature with another unpaired creature when
//	 either enters. They remain paired for as long as you control both of
//	 them.)
//	 As long as Tandem Lookout is paired with another creature, each of those
//	 creatures has "Whenever this creature deals damage to an opponent, draw a
//	 card.""
//
// Soulbond is the keyword's two triggers (CR 702.95a); the second line is a
// layer-6 grant of a triggered ability to both creatures of the pair
// (Imperious Mindbreaker's shape), so each draws for its own damage, combat
// or not.
//
// No simplification.
func init() {
	const grant = "tandem-lookout/draw"
	effects.Register(effects.Spec{
		OracleID:        "1576f031-57ca-4894-b461-8ea57b6cefdc",
		Name:            "Tandem Lookout",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordSoulbond},
		Triggered:       effects.Soulbond("Tandem Lookout"),
		Grants: []effects.AbilityGrant{{
			Key: grant,
			Triggered: []game.TriggeredAbility{
				effects.WheneverThisDealsDamageToAnOpponent("Tandem Lookout — draw a card", effects.Do(effects.DrawCards{N: 1})),
			},
			Text: "Whenever this creature deals damage to an opponent, draw a card.",
		}},
		Static: []game.StaticAbility{
			effects.PairedCreaturesHave("As long as Tandem Lookout is paired with another creature, each of those creatures has \"Whenever this creature deals damage to an opponent, draw a card.\"", grant),
		},
	})
}
