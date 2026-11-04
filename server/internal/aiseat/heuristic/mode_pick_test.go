package heuristic_test

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// mode_pick_test.go — #4674: a modal trigger's mode_pick is answered by
// what its bullets do. Before it every answer scored the same, so on a
// "choose up to" prompt the bot took the first offered, which is none.
//
// The bullets below are shaped as the wire carries them for the four
// cards the issue names (each bullet's purpose, legal targets and
// harmful flag); the engine-level test that the wire really carries them
// is aiseat's mode_pick_bot_test.go.

const modePickChoice = "choice-mode-pick"

// bullet is one offered bullet: its ModeSpec index and its view.
type bullet struct {
	index int
	view  protocol.ModeOptionView
}

func targeting(label string, harmful bool, players []int, cards ...string) protocol.ModeOptionView {
	lt := &protocol.LegalTargetsView{Min: 1, Max: 1, Harmful: harmful, Cards: cards}
	for _, p := range players {
		lt.Players = append(lt.Players, seatID(p).String())
	}
	return protocol.ModeOptionView{Label: label, LegalTargets: lt}
}

func untargeted(label string, purpose *protocol.PurposeView) protocol.ModeOptionView {
	return protocol.ModeOptionView{Label: label, Purpose: purpose}
}

// selections lists every answer to a non-repeatable prompt of `n`
// offered bullets with bounds lo..hi, in the enumerator's order
// (game.ModePickSelections): none first on an "up to", then each
// combination in index order.
func selections(idx []int, lo, hi int) [][]int {
	var out [][]int
	if lo == 0 {
		out = append(out, []int{})
	}
	var rec func(start int, cur []int)
	rec = func(start int, cur []int) {
		if len(cur) >= max(lo, 1) && len(cur) <= hi {
			out = append(out, slices.Clone(cur))
		}
		if len(cur) == hi {
			return
		}
		for i := start; i < len(idx); i++ {
			rec(i+1, append(cur, idx[i]))
		}
	}
	rec(0, nil)
	return out
}

// modePickInput is seat 0 owing a mode_pick over `bullets`, with bounds
// lo..hi, on the board `opts` builds.
func modePickInput(t *testing.T, lo, hi int, bullets []bullet, opts ...viewOpt) aiseat.Input {
	t.Helper()
	ch := protocol.PendingChoiceView{
		ID: modePickChoice, Kind: "mode_pick", Chooser: seatID(0).String(),
		Reason: "Choose up to one", ModeMin: lo, ModeMax: hi,
	}
	labels := map[int]string{}
	var idx []int
	for _, b := range bullets {
		ch.ModeIndexes = append(ch.ModeIndexes, b.index)
		ch.ModeOptions = append(ch.ModeOptions, b.view.Label)
		ch.ModeBullets = append(ch.ModeBullets, b.view)
		labels[b.index] = b.view.Label
		idx = append(idx, b.index)
	}
	v := newView([]protocol.PlayerView{newSeat(0, withHandCount(3)), newSeat(1), newSeat(2)},
		append(opts, withChoice(ch))...)
	var moves []legal.Move
	for _, sel := range selections(idx, lo, hi) {
		names := make([]string, 0, len(sel))
		for _, i := range sel {
			names = append(names, labels[i])
		}
		label := "none"
		if len(names) > 0 {
			label = strings.Join(names, " + ")
		}
		moves = append(moves, choiceMove(t, 0, modePickChoice, label, map[string]any{"modes": sel}))
	}
	return input(0, v, moves...)
}

func pickWith(t *testing.T, p *heuristic.Policy, in aiseat.Input) string {
	t.Helper()
	return chose(t, in, decide(t, p, in))
}

func noModePicks() *heuristic.Policy {
	cfg := heuristic.DefaultConfig()
	cfg.PriceModePicks = false
	return heuristic.NewWithConfig(cfg)
}

// Ertai Resurrected: "choose up to one — counter target spell, activated
// ability, or triggered ability; destroy another target creature or
// planeswalker." Its controller draws a card either way.
func ertaiBullets(counter []string, destroy []string) []bullet {
	var out []bullet
	if len(counter) > 0 {
		out = append(out, bullet{0, targeting("Counter target spell, activated ability, or triggered ability. Its controller draws a card.", false, nil, counter...)})
	}
	if len(destroy) > 0 {
		out = append(out, bullet{1, targeting("Destroy another target creature or planeswalker. Its controller draws a card.", true, nil, destroy...)})
	}
	return out
}

func TestErtaiDestroysAnOpponentsCreature(t *testing.T) {
	theirs, mine := cardID(1), cardID(2)
	board := withBattlefield(creature(theirs, 1, "Craterhoof Behemoth", 5, 5), creature(mine, 0, "Grizzly Bears", 2, 2))
	in := modePickInput(t, 0, 1, ertaiBullets(nil, []string{theirs, mine}), board)
	if got := pickWith(t, heuristic.New(), in); !strings.HasPrefix(got, "Destroy") {
		t.Errorf("Ertai with an opponent's creature to destroy chose %q, want the destroy bullet", got)
	}
	// #4674's bug, kept runnable: with the pricing off the bot takes none.
	if got := pickWith(t, noModePicks(), in); got != "none" {
		t.Errorf("PriceModePicks off chose %q, want the first answer (none)", got)
	}
}

func TestErtaiChoosesNoneWhenOnlyItsOwnCreatureCanBeDestroyed(t *testing.T) {
	mine := cardID(2)
	board := withBattlefield(creature(mine, 0, "Grizzly Bears", 2, 2))
	in := modePickInput(t, 0, 1, ertaiBullets(nil, []string{mine}), board)
	if got := pickWith(t, heuristic.New(), in); got != "none" {
		t.Errorf("Ertai whose only target is its own creature chose %q, want none", got)
	}
}

func TestErtaiCountersAnOpponentsSpellButNotItsOwn(t *testing.T) {
	theirs, mine := cardID(3), cardID(4)
	in := modePickInput(t, 0, 1, ertaiBullets([]string{theirs}, nil), withStack(spell(theirs, 1, "Demonic Tutor", "{1}{B}")))
	if got := pickWith(t, heuristic.New(), in); !strings.HasPrefix(got, "Counter") {
		t.Errorf("Ertai over an opponent's spell chose %q, want the counter", got)
	}
	in = modePickInput(t, 0, 1, ertaiBullets([]string{mine}, nil), withStack(spell(mine, 0, "Demonic Tutor", "{1}{B}")))
	if got := pickWith(t, heuristic.New(), in); got != "none" {
		t.Errorf("Ertai over only its own spell chose %q, want none", got)
	}
}

// Hullbreaker Horror: "choose up to one — return target spell you don't
// control to its owner's hand; return target nonland permanent to its
// owner's hand."
func TestHullbreakerBouncesWhatHurtsTheBotAndSparesItsOwn(t *testing.T) {
	spellID, theirs, mine := cardID(5), cardID(6), cardID(7)
	both := []bullet{
		{0, targeting("Return target spell you don't control to its owner's hand.", false, nil, spellID)},
		{1, targeting("Return target nonland permanent to its owner's hand.", true, nil, theirs, mine)},
	}
	in := modePickInput(t, 0, 1, both,
		withStack(spell(spellID, 1, "Demonic Tutor", "{1}{B}")),
		withBattlefield(creature(theirs, 1, "Craterhoof Behemoth", 5, 5), creature(mine, 0, "Grizzly Bears", 2, 2)))
	if got := pickWith(t, heuristic.New(), in); got == "none" {
		t.Errorf("Hullbreaker with an opponent's spell and creature to bounce chose none")
	}
	// No spell of an opponent's on the stack (that bullet is not offered)
	// and nothing but its own permanent to bounce.
	own := []bullet{{1, targeting("Return target nonland permanent to its owner's hand.", true, nil, mine)}}
	in = modePickInput(t, 0, 1, own, withBattlefield(creature(mine, 0, "Grizzly Bears", 2, 2)))
	if got := pickWith(t, heuristic.New(), in); got != "none" {
		t.Errorf("Hullbreaker with only its own permanent to bounce chose %q, want none", got)
	}
}

// Ravager Wurm: "choose up to one — this creature fights target creature
// you don't control; destroy target land with an activated ability that
// isn't a mana ability."
func TestRavagerWurmFightsAnOpponentsCreatureAndSparesItsOwnLand(t *testing.T) {
	theirs, myLand := cardID(8), cardID(9)
	fight := bullet{0, targeting("This creature fights target creature you don't control.", true, nil, theirs)}
	destroyLand := bullet{1, targeting("Destroy target land with an activated ability that isn't a mana ability.", true, nil, myLand)}
	board := withBattlefield(creature(theirs, 1, "Llanowar Elves", 1, 1), land(myLand, 0))
	in := modePickInput(t, 0, 1, []bullet{fight, destroyLand}, board)
	if got := pickWith(t, heuristic.New(), in); !strings.HasPrefix(got, "This creature fights") {
		t.Errorf("Ravager Wurm chose %q, want the fight", got)
	}
	in = modePickInput(t, 0, 1, []bullet{destroyLand}, withBattlefield(land(myLand, 0)))
	if got := pickWith(t, heuristic.New(), in); got != "none" {
		t.Errorf("Ravager Wurm whose only target is its own land chose %q, want none", got)
	}
}

// The Ruinous Wrecking Crew: "choose up to X — discard a card, then draw
// a card; target opponent loses 2 life; destroy target token; each player
// sacrifices a creature of their choice."
func crewBullets(tokens ...string) []bullet {
	out := []bullet{
		{0, untargeted("Discard a card, then draw a card.", &protocol.PurposeView{Draws: 1, Discards: 1})},
		{1, targeting("Target opponent loses 2 life.", false, []int{1, 2})},
	}
	if len(tokens) > 0 {
		out = append(out, bullet{2, targeting("Destroy target token.", true, nil, tokens...)})
	}
	return append(out, bullet{3, untargeted("Each player sacrifices a creature of their choice.", nil)})
}

func TestWreckingCrewTakesEveryBulletThatHelpsUpToX(t *testing.T) {
	myToken := cardID(10)
	board := withBattlefield(creature(myToken, 0, "Soldier", 1, 1, func(c *protocol.CardView) { c.IsToken = true }))
	for x, want := range map[int]string{
		1: "Target opponent loses 2 life.",
		2: "Discard a card, then draw a card. + Target opponent loses 2 life.",
		4: "Discard a card, then draw a card. + Target opponent loses 2 life.",
	} {
		t.Run(fmt.Sprintf("X=%d", x), func(t *testing.T) {
			in := modePickInput(t, 0, x, crewBullets(myToken), board)
			if got := pickWith(t, heuristic.New(), in); got != want {
				t.Errorf("X = %d: chose %q, want %q", x, got, want)
			}
		})
	}
}

func TestWreckingCrewChoosesNoneWhenEveryBulletOnOfferHurts(t *testing.T) {
	myToken := cardID(10)
	board := withBattlefield(creature(myToken, 0, "Soldier", 1, 1, func(c *protocol.CardView) { c.IsToken = true }))
	bad := []bullet{
		{2, targeting("Destroy target token.", true, nil, myToken)},
		{3, untargeted("Each player sacrifices a creature of their choice.", nil)},
	}
	in := modePickInput(t, 0, 2, bad, board)
	if got := pickWith(t, heuristic.New(), in); got != "none" {
		t.Errorf("chose %q with only its own token to destroy and an unpriced edict, want none", got)
	}
}

// A "choose one" (no empty answer) takes the best bullet, not the first.
func TestChooseOneTakesTheBestBulletNotTheFirst(t *testing.T) {
	theirs, mine := cardID(11), cardID(12)
	bullets := []bullet{
		{0, targeting("Destroy target creature.", true, nil, mine)},
		{1, targeting("Destroy target creature an opponent controls.", true, nil, theirs)},
	}
	board := withBattlefield(creature(theirs, 1, "Craterhoof Behemoth", 5, 5), creature(mine, 0, "Grizzly Bears", 2, 2))
	in := modePickInput(t, 1, 1, bullets, board)
	if got := pickWith(t, heuristic.New(), in); got != "Destroy target creature an opponent controls." {
		t.Errorf("chose %q, want the bullet aimed at the opponent", got)
	}
}

// A prompt that does not carry its bullets (one restored with no
// continuation) keeps the enumerator's first answer.
func TestAModePickWithoutBulletsKeepsTheFirstAnswer(t *testing.T) {
	in := modePickInput(t, 0, 1, ertaiBullets(nil, []string{cardID(1)}),
		withBattlefield(creature(cardID(1), 1, "Craterhoof Behemoth", 5, 5)))
	in.View.PendingChoices[0].ModeBullets = nil
	if got := pickWith(t, heuristic.New(), in); got != "none" {
		t.Errorf("chose %q on a prompt with no bullets, want the first answer", got)
	}
}

// A clause that declares no direction and can pick the bot's own says
// nothing about which way it cuts, so it is priced at nothing: Charming
// Prince's "choose one" keeps the enumerator's first bullet rather than
// blinking a creature for the target's sake, as before #4674.
func TestAnUndeclaredClauseOverTheBotsOwnIsNotPriced(t *testing.T) {
	mine := cardID(13)
	bullets := []bullet{
		{0, untargeted("Scry 2.", nil)},
		{1, untargeted("You gain 3 life.", nil)},
		{2, targeting("Exile another target creature you own. Return it to the battlefield under your control at the beginning of the next end step.", false, nil, mine)},
	}
	in := modePickInput(t, 1, 1, bullets, withBattlefield(creature(mine, 0, "Grizzly Bears", 2, 2)))
	if got := pickWith(t, heuristic.New(), in); got != "Scry 2." {
		t.Errorf("Charming Prince chose %q, want the first bullet", got)
	}
}
