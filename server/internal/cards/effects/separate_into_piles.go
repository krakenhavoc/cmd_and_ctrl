package effects

import (
	"fmt"
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// separate_into_piles.go — "Starting with you, each player separates all
// permanents they control into three piles. Then each player chooses one
// of their piles at random and sacrifices those permanents." (Whims of the
// Fates, #6248, CR 700.3).
//
// PileSplit is two piles and two players (one splits, one picks). Here each
// player splits their OWN permanents into N piles, so a split is N-1 keyed
// choose_cards picks to that player: the first pile from all their
// permanents, the next from what is left, and the last pile is whatever no
// pick took. A pile can be empty (CR 700.3d), so every pick may take none.
// "Starting with you": the players split one after another in turn order
// from the resolving effect's controller (CR 101.4), every split made
// before anything happens to any pile.
//
// No new prompt kind: the reveal_pick prompt is a choose_cards over a
// frozen list the client and the bot already answer, and permanents are
// public. The whole split rides each prompt as data (Carry: the players in
// order, then every permanent already put in a pile; Nums: how many
// players, the player and the pile being asked, then each placed
// permanent's pile), so a table waiting on a split is a restore point. The
// card's PilesThen runs once every player has split.

// PlayerPiles is one player's piles, in the order they were made; the last
// is the permanents no pick took.
type PlayerPiles struct {
	Player uuid.UUID
	Piles  [][]uuid.UUID
}

// SeparateIntoPiles has each player, starting with the controller and in
// turn order, separate all permanents they control into Piles piles, then
// runs Then.
type SeparateIntoPiles struct {
	Piles int
	Then  PilesThen
}

// PilesThen is a SeparateIntoPiles' registered rest of the card.
type PilesThen struct{ then game.NumberThen }

// Nums layout of every pile prompt, before the per-card piles.
const (
	pilesPlayers = iota
	pilesPlayerAt
	pilesPileAt
	pilesCount
	pilesHeader
)

// RegisterPilesThen registers what happens once every player has
// separated their permanents (an on-disk key, ledgered "number <key>":
// never renamed, never reused). `body` is handed each player's piles, in
// the order the players split, against the resolving item. Call it once,
// from a package-level var.
func RegisterPilesThen(key string, body func(ctx *Context, piles []PlayerPiles) error) PilesThen {
	return PilesThen{then: NumberThen(key, func(ctx *Context, r game.NumberAnswer) error {
		piles, ok := decodePiles(r.Carry, r.Nums)
		if !ok {
			return fmt.Errorf("%w: a pile split with no state", game.ErrInvalidParam)
		}
		return body(ctx, piles)
	})}
}

func (s SeparateIntoPiles) Apply(ctx *Context) error {
	if s.Then.then.Key() == "" {
		return fmt.Errorf("%w: SeparateIntoPiles with no Then", game.ErrInvalidParam)
	}
	if s.Piles < 2 {
		return fmt.Errorf("%w: SeparateIntoPiles of %d piles", game.ErrInvalidParam, s.Piles)
	}
	// #6175: a spell that groups permanents into piles notes itself once.
	NotePileOrGuess(ctx)
	players := playersInTurnOrderFrom(ctx.Game, ctx.Controller())
	return pilesStep(ctx, []string{s.Then.then.Key()}, players, nil,
		[]int{len(players), 0, 0, s.Piles}, nil)
}

// playersInTurnOrderFrom is every player still in the game, `from` first
// and then in turn order (CR 101.4).
func playersInTurnOrderFrom(g *game.Game, from uuid.UUID) []uuid.UUID {
	at := -1
	for i, p := range g.Seats {
		if p != nil && p.ID == from {
			at = i
		}
	}
	if at < 0 {
		return nil
	}
	var out []uuid.UUID
	for k := 0; k < len(g.Seats); k++ {
		p := g.Seats[g.SeatInTurnOrderForEffect(at, k)]
		if p != nil && !p.Eliminated {
			out = append(out, p.ID)
		}
	}
	return out
}

// pilesStep asks the current player for their next pile, or puts the rest
// of their permanents in their last pile and moves to the next player, or
// runs the card's Then once every player has split.
func pilesStep(ctx *Context, words []string, players, placed []uuid.UUID, head, piles []int) error {
	g := ctx.Game
	n, at, pile, count := head[pilesPlayers], head[pilesPlayerAt], head[pilesPileAt], head[pilesCount]
	for at < n {
		who := players[at]
		var left []uuid.UUID
		for _, id := range PermanentsControlledByPlayer(g, who) {
			if !slices.Contains(placed, id) {
				left = append(left, id)
			}
		}
		if pile < count-1 && len(left) > 0 {
			nums := append([]int{n, at, pile, count}, piles...)
			carry := append(append([]uuid.UUID{ctx.Controller()}, players...), placed...)
			queued, err := g.RevealPickThenForEffect(game.RevealPickPrompt{
				Chooser: who,
				Owner:   who,
				Source:  ctx.Source(),
				Question: fmt.Sprintf("%s — separate the permanents you control into %d piles: choose pile %d of %d (the last pile is the rest)",
					pilesCardName(ctx), count, pile+1, count),
				Cards:   left,
				Min:     0,
				Max:     len(left),
				ThenKey: pilesPicked,
				Carry:   carry,
				Asked:   game.NewAskedItem(ctx.Item),
				Nums:    nums,
				Word:    words[0],
			})
			if err != nil || queued > 0 {
				return err
			}
			// A chooser who has left queues nothing: their permanents are
			// in no pile and the next player splits.
			at, pile = at+1, 0
			continue
		}
		for _, id := range left {
			placed = append(placed, id)
			piles = append(piles, count-1+count*at)
		}
		at, pile = at+1, 0
	}
	then, ok := game.NumberThenOf(words[0])
	if !ok {
		return fmt.Errorf("%w: piles branch %q", game.ErrUnknownEffectKey, words[0])
	}
	return g.RunNumberThenForEffect(game.NumberBranch{
		Then:  then,
		Carry: append(append([]uuid.UUID{ctx.Controller()}, players...), placed...),
		Nums:  append([]int{n, n, 0, count}, piles...),
		Asked: game.NewAskedItem(ctx.Item),
	}, ctx.Controller(), ctx.Source(), 0)
}

func pilesCardName(ctx *Context) string {
	if c, ok := ctx.Game.LookupCardForEffect(ctx.Source()); ok && c.Name != "" {
		return c.Name
	}
	return "Piles"
}

// decodePiles reads the piles back out of the prompt's data.
func decodePiles(carry []uuid.UUID, nums []int) ([]PlayerPiles, bool) {
	if len(nums) < pilesHeader {
		return nil, false
	}
	n, count := nums[pilesPlayers], nums[pilesCount]
	if n < 0 || n > len(carry) || count < 1 || len(carry)-n != len(nums)-pilesHeader {
		return nil, false
	}
	players, placed, piles := carry[:n], carry[n:], nums[pilesHeader:]
	out := make([]PlayerPiles, n)
	at := make(map[uuid.UUID]int, n)
	for i, p := range players {
		out[i] = PlayerPiles{Player: p, Piles: make([][]uuid.UUID, count)}
		at[p] = i
	}
	// Each placed permanent belongs to the player whose split placed it;
	// the split walks the players in order, so a run of placements is
	// one player's until the next player's begin. The player is recorded
	// in the pile number: pile + count*player.
	for j, id := range placed {
		v := piles[j]
		who, pile := v/count, v%count
		if who < 0 || who >= n {
			return nil, false
		}
		out[who].Piles[pile] = append(out[who].Piles[pile], id)
	}
	return out, true
}

// pilesAnswer puts the permanents a player picked in the pile they were
// asked for, and asks the next pile.
func pilesAnswer(g *game.Game, r game.CardsChosen) error {
	ctx, carry := AskedContext(g, r.Asked, r.Chooser, r.Source, r.Carry)
	nums := r.Nums
	if len(nums) < pilesHeader {
		return fmt.Errorf("%w: a pile split with no state", game.ErrInvalidParam)
	}
	n := nums[pilesPlayers]
	if n < 0 || n > len(carry) {
		return fmt.Errorf("%w: a pile split with no state", game.ErrInvalidParam)
	}
	players, placed := carry[:n:n], append([]uuid.UUID(nil), carry[n:]...)
	head := append([]int(nil), nums[:pilesHeader]...)
	piles := append([]int(nil), nums[pilesHeader:]...)
	at, pile, count := head[pilesPlayerAt], head[pilesPileAt], head[pilesCount]
	picked, _ := RevealPartition(r.Candidates, r.Picked)
	for _, id := range picked {
		placed = append(placed, id)
		piles = append(piles, pile+count*at)
	}
	head[pilesPileAt] = pile + 1
	return pilesStep(ctx, []string{r.Word}, players, placed, head, piles)
}

// pilesPicked is the keyed answer to every pile prompt. Registered in init,
// not a var initialiser, because the step refers to it and it refers to
// the step.
var pilesPicked game.ChooseCardsThen

func init() {
	pilesPicked = game.RegisterChooseCardsThen("choose-cards/separate-into-piles", pilesAnswer)
}

// SacrificeARandomPileEach is Whims of the Fates' rest: "Then each player
// chooses one of their piles at random and sacrifices those permanents."
// Each player's pile is drawn from the seeded pick stream (an undo rewinds
// it), every pile chosen before anything is sacrificed, and the chosen
// piles are sacrificed together, one event (CR 701.21a).
var SacrificeARandomPileEach = RegisterPilesThen("number/sacrifice-a-random-pile-each", func(ctx *Context, all []PlayerPiles) error {
	var doomed []uuid.UUID
	for _, pp := range all {
		i := ctx.Game.ChooseIndexAtRandomForEffect(game.RandomDraw{Player: pp.Player, Source: ctx.Source()}, len(pp.Piles))
		if i < 0 {
			continue
		}
		doomed = append(doomed, pp.Piles[i]...)
	}
	if len(doomed) > 0 {
		ctx.Game.SacrificeAllForEffect(ctx.Source(), doomed)
	}
	return nil
})
