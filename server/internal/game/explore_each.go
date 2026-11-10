package game

import "github.com/google/uuid"

// explore_each.go — CR 701.44d, several permanents told to explore by one
// instruction (#5425): "each Merfolk creature you control explores"
// (Hakbal of the Surging Soul).
//
//	701.44d If multiple permanents are instructed to explore at the same
//	time, the first player in APNAP order who controls (or, in the case of
//	a permanent no longer on the battlefield, last controlled; see rule
//	701.44c) one or more of those permanents chooses one of them and it
//	explores. Then this process is repeated for each remaining instruction
//	to explore.
//
// The choice is made again before every explore, because each one changes
// what the player knows: a nonland card left on top is the card the next
// explore reveals. So the run is a chain, not an order fixed up front: a
// pick prompt over the permanents still to explore, the picked one's whole
// explore (its replacements, its counter, its graveyard question), and
// then the next pick. The permanents still owed ride each explore as
// exploreRest.Queue, plain data on its tails and prompts, so a table
// waiting anywhere in the chain is a restore point.
//
// One controller per call. Every catalogued instruction names one player's
// permanents ("you control"); a card that tells several players' permanents
// to explore is one call per player in APNAP order, each behind the last.

// ExploreEachForEffect has each permanent in `explorers` explore
// (CR 701.44a), all under `controller`, on behalf of `source`, one at a
// time in an order `controller` chooses as it goes (CR 701.44d). A list of
// one is ExploreForEffect and asks nothing.
//
// Caller must hold g.mu.
func (g *Game) ExploreEachForEffect(source uuid.UUID, explorers []ObjectRef, controller uuid.UUID) error {
	return g.exploreEachLocked(source, controller, append([]ObjectRef(nil), explorers...))
}

// exploreEachLocked is one step of the chain: explore the only permanent
// left, or ask which of several explores next.
//
// A permanent that is no longer the object the instruction named cannot be
// offered by a battlefield pick, and it still explores (CR 701.44c), so it
// goes first, with no question.
//
// Caller must hold g.mu.
func (g *Game) exploreEachLocked(source, controller uuid.UUID, left []ObjectRef) error {
	if len(left) == 0 {
		return nil
	}
	next := -1
	present := 0
	for i, ref := range left {
		if c := findBattlefieldCard(g, ref.ID); c == nil || c.ObjectEpoch != ref.Epoch {
			next = i
			break
		}
		present++
	}
	if next < 0 && present == 1 {
		next = 0
	}
	if p := g.playerByIDLocked(controller); next < 0 && (p == nil || p.Eliminated) {
		next = 0
	}
	if next >= 0 {
		return g.exploreOneOfLocked(source, controller, left, next)
	}
	carry := make([]uuid.UUID, 0, len(left)+1)
	carry = append(carry, controller)
	nums := make([]int, 0, len(left))
	for _, ref := range left {
		carry = append(carry, ref.ID)
		nums = append(nums, ref.Epoch)
	}
	asked, err := g.PermanentsPickedThenKeyedForEffect(PermanentPickPrompt{
		Chooser:  controller,
		Source:   source,
		Question: "Explore — choose which of these explores next",
		Of:       []uuid.UUID{controller},
		Candidates: func(*Game, uuid.UUID) ([]uuid.UUID, int, int) {
			return carry[1:], 1, 1
		},
	}, exploreEachPicked, RunData{Carry: carry, Nums: nums})
	_ = asked // nobody asked: the run's continuation has already explored the first.
	return err
}

// exploreOneOfLocked has left[i] explore, with the others owed after it.
//
// Caller must hold g.mu.
func (g *Game) exploreOneOfLocked(source, controller uuid.UUID, left []ObjectRef, i int) error {
	rest := make([]ObjectRef, 0, len(left)-1)
	rest = append(rest, left[:i]...)
	rest = append(rest, left[i+1:]...)
	return g.exploreLocked(source, left[i], controller, exploreRest{Queue: rest})
}

// exploreEachPicked is the answer to "which explores next". Carry is the
// controller, then the permanents still to explore; Nums their epochs. A
// run that asked nobody, or whose pick has since left, explores the first.
// An on-disk key: never renamed, never reused.
var exploreEachPicked = RegisterPromptRunThen("engine/explore-each-pick", func(g *Game, r RunDone) error {
	if len(r.Carry) < 2 || len(r.Nums) != len(r.Carry)-1 {
		return nil
	}
	controller := r.Carry[0]
	left := make([]ObjectRef, len(r.Nums))
	for i := range r.Nums {
		left[i] = ObjectRef{ID: r.Carry[i+1], Epoch: r.Nums[i]}
	}
	pick := 0
	if picked := r.Picks().Cards(); len(picked) > 0 {
		for i, ref := range left {
			if ref.ID == picked[0] {
				pick = i
				break
			}
		}
	}
	return exploreOneOf(g, r.Source, controller, left, pick)
})

// exploreEachNext and exploreOneOf are exploreEachLocked and
// exploreOneOfLocked, reached through variables so the package's
// initialisation order has no cycle (an explore's tails finish an explore,
// and the pick's continuation starts one).
var (
	exploreEachNext func(g *Game, source, controller uuid.UUID, left []ObjectRef) error
	exploreOneOf    func(g *Game, source, controller uuid.UUID, left []ObjectRef, i int) error
)

func init() {
	exploreEachNext = func(g *Game, source, controller uuid.UUID, left []ObjectRef) error {
		return g.exploreEachLocked(source, controller, left)
	}
	exploreOneOf = func(g *Game, source, controller uuid.UUID, left []ObjectRef, i int) error {
		return g.exploreOneOfLocked(source, controller, left, i)
	}
}
