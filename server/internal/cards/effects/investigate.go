package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// investigate.go — the card-facing half of the CR 701.16 keyword action
// "investigate" (#5425). The engine half is game/investigate.go.
//
//	701.16a "Investigate" means "Create a Clue token." See rule 111.10f.
//
// A card that prints the word "investigate" uses Investigate, so a
// "whenever you investigate" ability (Erdwal Illuminator) sees it. A card
// that prints "create a Clue token" without the word keeps
// CreateToken{Template: ClueToken()}: that is not an investigate.

func init() {
	game.RegisterInvestigateToken(func() game.Card { return ruleDefinedToken(ClueToken()) })
}

// Investigate is "investigate" (N of 1), "investigate twice" (N of 2) and
// "investigate X times": Player, or the effect's controller when it is
// nil, takes the action N times, each a Clue token and then
// game.EventInvestigate.
type Investigate struct {
	Player uuid.UUID
	N      int
}

func (s Investigate) Apply(ctx *Context) error {
	player := s.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	return ctx.Game.InvestigateForEffect(player, ctx.Source(), s.N)
}

func (s Investigate) footprint() (game.FootprintStep, bool) {
	key, manual, keywords := game.FootprintTokenOf(ClueToken())
	return game.FootprintStep{
		Kind: game.FootprintInvestigate, N: s.N, TokenKey: key, TokenManual: manual, TokenKeywords: keywords,
	}, s.Player == uuid.Nil && s.N > 0
}

// InvestigateForEach is "investigate once for each <thing>", counted as
// the item resolves (Wojek Investigator): CreateTokensForEach's contract
// for Count, and its footprint with the investigate step in place of the
// token step.
type InvestigateForEach struct {
	// Name says what it counts, for a reader of the card file.
	Name  string
	Count func(ctx *Context) int
}

func (s InvestigateForEach) Apply(ctx *Context) error {
	if s.Count == nil {
		return nil
	}
	return Investigate{N: s.Count(ctx)}.Apply(ctx)
}

func (s InvestigateForEach) footprints() ([]game.FootprintStep, bool) {
	if s.Count == nil {
		return nil, false
	}
	step, _ := Investigate{N: 1}.footprint()
	return []game.FootprintStep{condStep(game.FootprintCondBoard, 0), step}, true
}

// WheneverYouInvestigateForTheFirstTimeEachTurn is that trigger (Erdwal
// Illuminator): the turn tally already counts the investigate the ability
// is asked about (game.InvestigatedThisTurn), so the turn's first is the
// one that reads 1. The tally is the player's, so a permanent that enters
// after their first investigate of the turn sees no first one.
func WheneverYouInvestigateForTheFirstTimeEachTurn(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventInvestigate, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return ev.Actor == source.Controller && g.InvestigatedThisTurn(source.Controller) == 1
	}, label, effect)
}
