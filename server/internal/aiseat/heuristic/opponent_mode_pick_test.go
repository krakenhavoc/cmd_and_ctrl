package heuristic_test

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// opponent_mode_pick_test.go — #5771, CR 700.2e: the bot chooses the
// mode of somebody else's "An opponent chooses one" spell. The mode is
// the bot's, the spell and its targets the caster's, so the bot takes the
// bullet that does the caster the least good.

// opponentModeInput is seat 0 owing the mode of seat 1's spell over
// `bullets` (a choose-one), on the board `opts` builds.
func opponentModeInput(t *testing.T, bullets []bullet, opts ...viewOpt) aiseat.Input {
	t.Helper()
	ch := protocol.PendingChoiceView{
		ID: modePickChoice, Kind: "mode_pick", Chooser: seatID(0).String(),
		Reason: "B's Fatal Lore: choose one", ModeMin: 1, ModeMax: 1,
		ModeChosenFor: seatID(1).String(), ModeSpellName: "Fatal Lore",
	}
	var moves []legal.Move
	for _, b := range bullets {
		ch.ModeIndexes = append(ch.ModeIndexes, b.index)
		ch.ModeOptions = append(ch.ModeOptions, b.view.Label)
		ch.ModeBullets = append(ch.ModeBullets, b.view)
	}
	v := newView([]protocol.PlayerView{newSeat(0, withHandCount(3)), newSeat(1), newSeat(2)},
		append(opts, withChoice(ch))...)
	for _, b := range bullets {
		moves = append(moves, choiceMove(t, 0, modePickChoice, b.view.Label, map[string]any{"modes": []int{b.index}}))
	}
	return input(0, v, moves...)
}

// Fatal Lore's two bullets as the wire carries them to the chooser: the
// caster's draw, and the caster's destroy of up to two creatures the
// chooser controls.
func fatalLoreBullets(mine ...string) []bullet {
	destroy := protocol.ModeOptionView{
		Label:        "You destroy up to two target creatures that player controls. They can't be regenerated. That player draws up to three cards.",
		LegalTargets: &protocol.LegalTargetsView{Min: 0, Max: 2, Cards: mine},
	}
	return []bullet{
		{0, untargeted("You draw three cards.", &protocol.PurposeView{Draws: 3})},
		{1, destroy},
	}
}

// With nothing of the bot's to destroy, the destroy bullet costs the
// bot nothing and the draw gives the caster three cards: the bot takes
// the destroy (and draws up to three itself).
func TestBotChoosingAnOpponentsModeDeniesTheDrawWhenItHasNoCreature(t *testing.T) {
	in := opponentModeInput(t, fatalLoreBullets())
	if got := pickWith(t, heuristic.New(), in); !strings.HasPrefix(got, "You destroy") {
		t.Errorf("chose %q, want the bullet that gives the caster nothing", got)
	}
}

// With two big creatures on the line, losing them costs the bot more than
// the caster's three cards: the bot gives up the draw instead.
func TestBotChoosingAnOpponentsModeSparesItsBestCreatures(t *testing.T) {
	a, b := cardID(31), cardID(32)
	board := withBattlefield(creature(a, 0, "Craterhoof Behemoth", 5, 5), creature(b, 0, "Craterhoof Behemoth", 5, 5))
	in := opponentModeInput(t, fatalLoreBullets(a, b), board)
	if got := pickWith(t, heuristic.New(), in); got != "You draw three cards." {
		t.Errorf("chose %q, want the draw rather than two of its best creatures", got)
	}
}

// The bot's own trigger is still priced for the bot: ModeChosenFor names
// the bot itself only on nobody's prompt, and an empty one is its own.
func TestBotChoosingItsOwnModeIsUnchanged(t *testing.T) {
	in := opponentModeInput(t, fatalLoreBullets())
	for i := range in.View.PendingChoices {
		in.View.PendingChoices[i].ModeChosenFor = ""
	}
	if got := pickWith(t, heuristic.New(), in); got != "You draw three cards." {
		t.Errorf("chose %q for its own modal choice, want the draw", got)
	}
}
