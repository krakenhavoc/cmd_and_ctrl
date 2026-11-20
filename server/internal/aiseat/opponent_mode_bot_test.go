package aiseat_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// opponent_mode_bot_test.go — #5771, CR 700.2e, through the real card,
// the real wire, the real enumerator and the real heuristic: a bot casts
// Fatal Lore naming no mode, and a bot opponent chooses its mode, priced
// as the caster's gain, from the bullets the wire carries to it alone.

const botFatalLoreOracle = "aeb85786-16fc-45d8-9383-60054a878c40"

func fatalLoreInHand(g *game.Game, me *game.Player) uuid.UUID {
	c := game.Card{InstanceID: uuid.New(), Name: "Fatal Lore", OracleID: botFatalLoreOracle,
		TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID}
	c.AddKnowersAll(seatIDs(g))
	me.Hand.PushTop(c)
	return c.InstanceID
}

func TestHeuristicChoosesTheModeOfAnOpponentsFatalLore(t *testing.T) {
	g := newSettledTable(t, 5771)
	advanceToStep(t, g, game.StepPrecombatMain)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := otherSeat(g, me)
	me.Hand.Cards = nil
	mine := putCreature(g, me, "Grizzly Bears", 2, 2)
	lore := fatalLoreInHand(g, me)

	// The caster's enumerator offers the cast with no mode and no target:
	// the mode is the opponent's (CR 700.2e).
	var offered bool
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Type == "cast_spell" && strings.Contains(string(m.Params), lore.String()) {
			offered = true
			if strings.Contains(string(m.Params), `"modes"`) || strings.Contains(string(m.Params), `"targets"`) {
				t.Fatalf("the caster's move names a mode or target: %s", m.Params)
			}
		}
	}
	if !offered {
		t.Fatal("the caster was not offered Fatal Lore")
	}
	if err := g.CastSpell(me.ID, lore, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	var pc *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceModePick {
			pc = c
		}
	}
	if pc == nil || pc.Chooser != opp.ID {
		t.Fatalf("the mode_pick = %+v, want one owed by the only opponent", pc)
	}
	if owesChoice(g, me.ID) {
		t.Fatal("the caster owes a choice while the opponent chooses the mode")
	}

	// The wire tells the chooser whose spell it is and what each bullet
	// does: the destroy bullet reaches the chooser's creatures, not the
	// caster's.
	view, ok := bulletsFor(g, opp.ID, pc.ID)
	if !ok || view.ModeChosenFor != me.ID.String() || view.ModeSpellName != "Fatal Lore" {
		t.Fatalf("the chooser's prompt: for %q spell %q", view.ModeChosenFor, view.ModeSpellName)
	}
	if len(view.ModeBullets) != 2 || view.ModeBullets[1].LegalTargets == nil {
		t.Fatalf("the chooser's prompt carries %d bullets", len(view.ModeBullets))
	}
	if slices.Contains(view.ModeBullets[1].LegalTargets.Cards, mine.String()) {
		t.Fatal("the destroy bullet reaches the caster's own creature; it is \"that player controls\"")
	}
	if theirs, _ := bulletsFor(g, me.ID, pc.ID); theirs.ModeBullets != nil {
		t.Error("the caster's copy of the prompt carries the chooser's bullets")
	}

	// With nothing of its own to lose, the bot denies the caster the draw.
	if got := decidedModes(t, g, opp, heuristic.New()); !slices.Equal(got, []int{1}) {
		t.Fatalf("the bot chose modes %v, want the destroy bullet [1]", got)
	}
	driveChoices(t, g, heuristic.New(), opp.ID, 2)
	item := g.StackMeta[lore]
	if item == nil || !slices.Equal(item.Modes, []int{1}) || item.ModeChooser != opp.ID {
		t.Fatalf("the spell's item = %+v, want the destroy bullet chosen by the bot", item)
	}
}
