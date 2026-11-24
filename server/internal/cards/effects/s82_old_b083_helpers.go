package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b083_helpers.go — helpers for OLD batch 083 (#5507), Commander
// Legends token makers.

// S82OldB083ILostLifeLastTurn is "you lost life last turn" for the item's
// controller (First Response): the previous turn, whoever's it was. The fact
// is stamped as the turn begins and no step changes it; it is declared as a
// board read so the trigger keeps a footprint.
var S82OldB083ILostLifeLastTurn = CondBoard{
	Name: "you lost life last turn",
	Test: func(ctx *Context) bool { return ctx.Game.LostLifeLastTurnForEffect(ctx.Controller()) },
}

// S82OldB083ILostLifeLastTurnWhen is the When for the same condition as the
// trigger fires.
func S82OldB083ILostLifeLastTurnWhen(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return g.LostLifeLastTurnForEffect(source.Controller)
}

// S82OldB083YouControlACommander is "if you control a commander" (Biowaste
// Blob), the CR 603.4 condition read as the trigger fires and again as it
// resolves: a commander permanent the controller controls.
var S82OldB083YouControlACommander = CondBoard{
	Name: "you control a commander",
	Test: func(ctx *Context) bool { return ControlsACommander(ctx.Game, ctx.Controller()) },
}

// S82OldB083TappedTokensForEachCreatureDied is "create a tapped <Template>
// token for each creature that died this turn" (Nevinyrral, Urborg Tyrant):
// the table-wide tally of creatures that died this turn, anyone's, tokens
// included, read as the ability resolves.
type S82OldB083TappedTokensForEachCreatureDied struct {
	Template game.Card
}

func (s S82OldB083TappedTokensForEachCreatureDied) Apply(ctx *Context) error {
	n := B11CreaturesDiedThisTurn(ctx.Game)
	if n <= 0 {
		return nil
	}
	_, err := ctx.Game.CreateTokensForEffect(ctx.Controller(), ctx.TextToken(s.Template), n, game.TokenEntryOptions{Tapped: true})
	return err
}

func (s S82OldB083TappedTokensForEachCreatureDied) footprints() ([]game.FootprintStep, bool) {
	return s82OldB083CountedTokenFootprints(s.Template), true
}

// s82OldB083CountedTokenFootprints is the footprint of a token step whose count
// is read off the board: a board read, then the token step.
func s82OldB083CountedTokenFootprints(tmpl game.Card) []game.FootprintStep {
	key, manual, keywords := game.FootprintTokenOf(tmpl)
	return []game.FootprintStep{
		CondStep(game.FootprintCondBoard, 0),
		{Kind: game.FootprintCreateToken, N: 1, TokenKey: key, TokenManual: manual, TokenKeywords: keywords},
	}
}

// S82OldB083FewestCreaturesEachMakes is "each player who controls the fewest
// creatures creates a <Template> token" (Gor Muldrak, Amphinologist): every
// player still in the game whose creature count equals the lowest count at the
// table (ties all make one, a player with none counts as zero), counted as the
// ability resolves, each player the token's controller.
type S82OldB083FewestCreaturesEachMakes struct {
	Template game.Card
}

func (s S82OldB083FewestCreaturesEachMakes) Apply(ctx *Context) error {
	counts := map[uuid.UUID]int{}
	var order []uuid.UUID
	for _, p := range ctx.Game.Seats {
		if p == nil || p.Eliminated {
			continue
		}
		counts[p.ID] = 0
		order = append(order, p.ID)
	}
	if len(order) == 0 {
		return nil
	}
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if _, ok := counts[c.Controller]; ok && c.IsCreature() {
			counts[c.Controller]++
		}
	}
	fewest := counts[order[0]]
	for _, id := range order {
		fewest = min(fewest, counts[id])
	}
	for _, id := range order {
		if counts[id] != fewest {
			continue
		}
		if err := (CreateToken{Controller: id, Template: s.Template, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (s S82OldB083FewestCreaturesEachMakes) footprints() ([]game.FootprintStep, bool) {
	return s82OldB083CountedTokenFootprints(s.Template), true
}

// S82OldB083ExileCardsEqualToOpponents is "Exile cards from your graveyard
// equal to the number of opponents you have" as a CR 602 ability's cost
// (Araumi of the Dead Tide): the count is read as the ability is announced.
func S82OldB083ExileCardsEqualToOpponents(label string) game.AbilityCost {
	return game.AbilityCost{ExileCards: &game.ExileCost{Label: label, From: game.ZoneGraveyard, CountOpponents: true}}
}

// S82OldB083HorrorToken is "an X/X colorless Horror artifact creature
// token" sized by `x` (Profane Transfusion): a 0/0 placeholder set to the
// difference as it is created, as Stitcher Geralf's Zombie is.
func S82OldB083HorrorToken(ctx *Context, x int) error {
	tok := ctx.TextToken(TokenCard("0/0 colorless Horror artifact"))
	tok.Power, tok.Toughness = max(x, 0), max(x, 0)
	return ctx.Game.CreateTokenForEffect(ctx.Controller(), tok, 1)
}
