package cardsb

import (
	"fmt"
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bend or Break — Sorcery {3}{R} (Invasion, #5616):
//
//	"Each player separates all nontoken lands they control into two piles. For
//	 each player, one of their piles is chosen by one of their opponents of
//	 their choice. Destroy all lands in the chosen piles. Tap all lands in the
//	 other piles."
//
// Three steps, all prompts keyed so a table waiting on one is a restore
// point. (1) Every player separates their nontoken lands into two piles
// before anything else happens (SeparateIntoPiles, CR 700.3; a pile can be
// empty, CR 700.3d). (2) For each player in turn order, that player names one
// of their opponents (asked only when there is more than one), and that
// opponent chooses one of the two piles. (3) With every choice made, the lands
// in the chosen piles are destroyed together and the lands in the other piles
// are tapped. A player with no nontoken lands has nothing to separate and
// nothing is chosen for them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4362efe4-339d-40e1-bd72-3cc9e41bf93f",
		Name:         "Bend or Break",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.SeparateIntoPiles{Piles: 2, Then: bendOrBreakSplit}.Apply(ctx)
		},
	})
}

// bendState is the whole of a Bend or Break in progress. Carry (state.ids):
// the players in order, then each player's pile 0 and pile 1. Nums
// (state.nums): the player count, the player being asked, each player's two
// pile sizes, then each player's decision (0 not yet, 1 or 2 the chosen pile,
// 3 nobody chose).
type bendState struct {
	players []uuid.UUID
	piles   [][2][]uuid.UUID
	at      int
	chosen  []int
}

func (s bendState) encode() ([]uuid.UUID, []int) {
	n := len(s.players)
	ids := append([]uuid.UUID(nil), s.players...)
	nums := []int{n, s.at}
	for _, p := range s.piles {
		ids = append(ids, p[0]...)
		ids = append(ids, p[1]...)
		nums = append(nums, len(p[0]), len(p[1]))
	}
	return ids, append(nums, s.chosen...)
}

func decodeBendState(ids []uuid.UUID, nums []int) (bendState, bool) {
	if len(nums) < 2 {
		return bendState{}, false
	}
	n := nums[0]
	if n < 0 || len(ids) < n || len(nums) != 2+3*n {
		return bendState{}, false
	}
	s := bendState{players: ids[:n:n], at: nums[1], chosen: append([]int(nil), nums[2+2*n:]...)}
	rest := ids[n:]
	for i := 0; i < n; i++ {
		a, b := nums[2+2*i], nums[3+2*i]
		if a < 0 || b < 0 || len(rest) < a+b {
			return bendState{}, false
		}
		s.piles = append(s.piles, [2][]uuid.UUID{rest[:a:a], rest[a : a+b : a+b]})
		rest = rest[a+b:]
	}
	return s, len(rest) == 0
}

// bendOrBreakSplit starts the choices once every player has separated.
// An on-disk key: never renamed, never reused.
var bendOrBreakSplit = effects.RegisterPilesThenOfNontokenLands("number/bend-or-break-split", func(ctx *effects.Context, all []effects.PlayerPiles) error {
	s := bendState{chosen: make([]int, len(all))}
	for _, pp := range all {
		s.players = append(s.players, pp.Player)
		var two [2][]uuid.UUID
		if len(pp.Piles) > 0 {
			two[0] = pp.Piles[0]
		}
		if len(pp.Piles) > 1 {
			two[1] = pp.Piles[1]
		}
		s.piles = append(s.piles, two)
	}
	return bendRunner.run(ctx, s)
})

// bendOpponentPicked is the answer to "choose an opponent". An on-disk key:
// never renamed, never reused.
var bendOpponentPicked = effects.OptionPickThen("option-pick/bend-or-break-opponent", func(ctx *effects.Context, r game.OptionPicked) error {
	s, ok := decodeBendState(r.Carry, r.Nums)
	if !ok {
		return fmt.Errorf("%w: a Bend or Break with no state", game.ErrInvalidParam)
	}
	if r.Option == nil || r.Option.Player == uuid.Nil {
		s.chosen[s.at] = 3
		s.at++
		return bendRunner.run(ctx, s)
	}
	return bendAskPile(ctx, s, r.Option.Player)
})

// bendPilePicked is the answer to "choose a pile". An on-disk key: never
// renamed, never reused.
var bendPilePicked = effects.OptionPickThen("option-pick/bend-or-break-pile", func(ctx *effects.Context, r game.OptionPicked) error {
	s, ok := decodeBendState(r.Carry, r.Nums)
	if !ok || s.at >= len(s.players) {
		return fmt.Errorf("%w: a Bend or Break with no state", game.ErrInvalidParam)
	}
	s.chosen[s.at] = 3
	if r.Index == 0 || r.Index == 1 {
		s.chosen[s.at] = r.Index + 1
	}
	s.at++
	return bendRunner.run(ctx, s)
})

// bendRunner is how the answers go on without the package-level keys
// referring to the step that refers to them (an initialisation cycle).
type bendStepper interface {
	run(ctx *effects.Context, s bendState) error
}

type bendSteps struct{}

func (bendSteps) run(ctx *effects.Context, s bendState) error { return bendRun(ctx, s) }

var bendRunner bendStepper = bendSteps{}

// bendRun asks for the next player's opponent and pile, or finishes.
func bendRun(ctx *effects.Context, s bendState) error {
	for s.at < len(s.players) {
		p := s.players[s.at]
		if len(s.piles[s.at][0])+len(s.piles[s.at][1]) == 0 {
			s.chosen[s.at] = 3
			s.at++
			continue
		}
		opps := effects.Opponents.Seats(ctx, p)
		switch len(opps) {
		case 0:
			s.chosen[s.at] = 3
			s.at++
			continue
		case 1:
			return bendAskPile(ctx, s, opps[0])
		}
		var options []game.ChoiceOption
		for _, o := range opps {
			label := "Opponent"
			if pl := ctx.Game.PlayerByIDForEffect(o); pl != nil && pl.Name != "" {
				label = pl.Name
			}
			options = append(options, game.ChoiceOption{Label: label, Player: o})
		}
		ids, nums := s.encode()
		return effects.PickOption{
			Player:   p,
			Question: "Bend or Break — choose an opponent to choose one of your piles",
			Options:  options,
			ThenKey:  bendOpponentPicked,
			Carry:    ids,
			Nums:     nums,
			WithItem: true,
		}.Apply(ctx)
	}
	return bendFinish(ctx, s)
}

// bendAskPile has `chooser` choose one of player s.at's two piles.
func bendAskPile(ctx *effects.Context, s bendState, chooser uuid.UUID) error {
	owner := s.players[s.at]
	label := func(i int) string {
		n := len(s.piles[s.at][i])
		switch n {
		case 0:
			return fmt.Sprintf("Choose pile %d (empty)", i+1)
		case 1:
			return fmt.Sprintf("Choose pile %d (1 land)", i+1)
		}
		return fmt.Sprintf("Choose pile %d (%d lands)", i+1, n)
	}
	ids, nums := s.encode()
	return effects.PickOption{
		Player:   chooser,
		From:     owner,
		Question: "Bend or Break — choose one of the two piles of lands: its lands are destroyed and the other pile's are tapped",
		Options: []game.ChoiceOption{
			{Label: label(0), Cards: s.piles[s.at][0]},
			{Label: label(1), Cards: s.piles[s.at][1]},
		},
		ThenKey:  bendPilePicked,
		Carry:    ids,
		Nums:     nums,
		WithItem: true,
	}.Apply(ctx)
}

// bendFinish destroys the lands in the chosen piles together, then taps the
// lands in the other piles.
func bendFinish(ctx *effects.Context, s bendState) error {
	var doomed, tapped []uuid.UUID
	for i, p := range s.piles {
		switch s.chosen[i] {
		case 1:
			doomed = append(doomed, p[0]...)
			tapped = append(tapped, p[1]...)
		case 2:
			doomed = append(doomed, p[1]...)
			tapped = append(tapped, p[0]...)
		}
	}
	if len(doomed) > 0 {
		err := effects.DestroyAllMatching{Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
			return c.IsLand() && slices.Contains(doomed, c.InstanceID)
		}}.Apply(ctx)
		if err != nil {
			return err
		}
	}
	for _, id := range tapped {
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.IsLand() && !c.Tapped {
			if err := ctx.Game.TapTargetForEffect(id); err != nil {
				return err
			}
		}
	}
	return nil
}
