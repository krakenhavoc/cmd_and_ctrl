package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tilonalli's Summoner — Creature — Human Shaman {1}{R}, 1/1:
//
//	"Ascend (If you control ten or more permanents, you get the city's
//	 blessing for the rest of the game.)
//	 Whenever this creature attacks, you may pay {X}{R}. If you do,
//	 create X 1/1 red Elemental creature tokens that are tapped and
//	 attacking. At the beginning of the next end step, exile those
//	 tokens unless you have the city's blessing."
//
// X is chosen and {X}{R} paid as the trigger resolves (CR 608.2d,
// 118.12), through MayPayX (#2727): a number from 0 to the most you can
// pay, then the ordinary "you may pay {3}{R}" prompt, held in the
// declare attackers step so the tokens join this combat. The tokens are
// put onto the battlefield attacking (CR 506.3c), so they fire no attack
// triggers, and you choose what each attacks (CR 508.4) when there is
// more than one player, planeswalker or battle to attack
// (CreateTokensAttackingYourChoice); a bot sends them where the
// Summoner went. "Those tokens" are the ones this creation made,
// doubled ones included, and the city's blessing is read as the end
// step trigger resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "830d3450-18e7-4bac-b330-3906ef5e4514",
		Name:            "Tilonalli's Summoner",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordAscend},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Tilonalli's Summoner — pay {X}{R} for X attacking Elementals", tilonallisSummonerAttacks),
		},
	})
}

const tilonallisSummonerName = "Tilonalli's Summoner"

// tilonallisSummonerPaid is the paid half, keyed (#4289): X rides the
// prompt, and the creature it attacked is read off the asking trigger's
// event, which the prompt carries.
var tilonallisSummonerPaid = PayUnlessBranches("pay-unless/tilonallis-summoner-attacking-elementals",
	func(ctx *Context, r game.PayUnlessAnswer) error {
		attacked := uuid.Nil
		if ctx.Item != nil && ctx.Item.Trigger != nil {
			attacked = ctx.Item.Trigger.Event.Target
		}
		return CreateTokensAttackingYourChoice{
			Template: TokenCard("1/1 red Elemental"),
			N:        r.Number,
			Tapped:   true,
			Prefer:   attacked,
			Label:    tilonallisSummonerName,
			Then: func(ctx *Context, created []uuid.UUID) error {
				return ExileUnlessCitysBlessing(ctx, tilonallisSummonerName+" — exile the Elementals unless you have the city's blessing", created)
			},
		}.Apply(ctx)
	}, nil)

func tilonallisSummonerAttacks(g *game.Game, item *game.StackItem) error {
	return MayPayX{
		Cost:       "{X}{R}",
		Label:      tilonallisSummonerName,
		Buys:       "create X tapped and attacking 1/1 red Elementals",
		Goal:       XAsHighAsYouCan,
		InThisStep: true,
		Then:       tilonallisSummonerPaid,
	}.Apply(NewContext(g, item))
}
