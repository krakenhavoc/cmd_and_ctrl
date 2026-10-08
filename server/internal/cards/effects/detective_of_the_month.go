package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Detective of the Month — Creature — Human Detective {2}{U}, 2/3
// (EDHREC rank 6489):
//
//	"Ascend (If you control ten or more permanents, you get the city's
//	 blessing for the rest of the game.)
//	 As long as you have the city's blessing, Detectives you control
//	 can't be blocked.
//	 Whenever you draw your second card each turn, create a 2/2 white
//	 and blue Detective creature token."
//
// The token is a Detective, so once the blessing is earned the card
// feeds its own evasion. "Detectives you control" includes this
// creature, and a token or a changeling counts by its effective
// subtypes (HasSubtype).
//
// The evasion is a CR 509.1b block rule read live as blockers are
// declared: it applies to a Detective attacker whose controller is this
// creature's controller and who has the blessing now; losing the
// Detective of the Month removes it (the rule goes with its source).
// The blessing is never lost once earned (CR 702.131c).
//
// "Your second card each turn" is Thopter Fabricator's tally read: the
// count is bumped before the harvester runs, so the second draw reports
// exactly two and the third reports three.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "10d1d6ca-95f8-48da-8798-37c8aadb1f1f",
		Name:            "Detective of the Month",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordAscend},
		BlockRules: []game.BlockRule{
			CantBeBlockedWhile(yourDetectives, youHaveTheCitysBlessingCast),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b41DrewYourSecondCardThisTurn(ev, source, g)
			}, "Detective of the Month — create a 2/2 Detective",
				Do(CreateToken{Template: TokenCard("2/2 white and blue Detective"), N: 1})),
		},
	})
}

// yourDetectives is the block scope "Detectives you control": the
// attacker is a Detective and is controlled by the rule source's
// controller.
func yourDetectives(_ *game.Game, c, source *game.Card) bool {
	return source != nil && c.Controller == source.Controller && c.HasSubtype("Detective")
}

// youHaveTheCitysBlessingCast reads the blessing for the rule source's
// controller, which is the "caster" the block-rule predicates are asked
// with.
func youHaveTheCitysBlessingCast(g *game.Game, you uuid.UUID, _ game.Card) bool {
	return YouHaveTheCitysBlessing(g, you)
}
