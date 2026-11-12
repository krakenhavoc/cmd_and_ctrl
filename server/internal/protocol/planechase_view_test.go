package protocol

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// planechase_view_test.go — ADR 0216 §9: GameView.planechase. Absent on
// a plain table; on a Planechase table it names the planar controller,
// counts the deck and shows the face-up plane to every viewer, and
// never puts a card of the planar deck on the wire.

type viewPlanes struct{ cards []game.Card }

func (v viewPlanes) PlanarCards(game.PlanarKinds) []game.Card { return v.cards }

func planechaseViewGame(t *testing.T, on bool) *game.Game {
	t.Helper()
	var planes []game.Card
	for i := range 20 {
		c := game.NewCard(fmt.Sprintf("View Plane %02d", i), uuid.Nil)
		c.OracleID = uuid.NewSHA1(uuid.NameSpaceOID, []byte(c.Name)).String()
		c.TypeLine = "Plane — Viewworld"
		c.Layout = "planar"
		planes = append(planes, c)
	}
	prev := game.PlanarCards
	game.PlanarCards = viewPlanes{cards: planes}
	t.Cleanup(func() { game.PlanarCards = prev })

	g := game.NewGame()
	for i := range 2 {
		deck := []game.Card{game.NewCommander(fmt.Sprintf("Commander %d", i), uuid.Nil)}
		for j := range 20 {
			deck = append(deck, game.NewCard(fmt.Sprintf("Filler %d-%d", i, j), uuid.Nil))
		}
		if _, err := g.AddPlayer(fmt.Sprintf("Seat%d", i), deck); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.UpdateSettings(uuid.Nil, game.SettingsPatch{Planechase: &on}); err != nil {
		t.Fatal(err)
	}
	if err := g.StartWithSource(nil); err != nil {
		t.Fatal(err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatal(err)
		}
	}
	return g
}

func TestThePlanechaseViewIsPublicAndHidesTheDeck(t *testing.T) {
	g := planechaseViewGame(t, true)
	up := g.PlanarFaceUpCards()
	if len(up) != 1 {
		t.Fatalf("%d face-up planes", len(up))
	}
	for _, viewer := range []string{g.Seats[0].ID.String(), g.Seats[1].ID.String(), SpectatorViewerID, ""} {
		v := ViewOfGameFor(g, viewer)
		pc := v.Planechase
		if pc == nil {
			t.Fatalf("viewer %q: no planechase view", viewer)
		}
		if pc.PlanarController != g.Planechase.Controller.String() || pc.DeckCount != 19 {
			t.Errorf("viewer %q: controller %q, deck_count %d", viewer, pc.PlanarController, pc.DeckCount)
		}
		if len(pc.FaceUp) != 1 || pc.FaceUp[0].Name != up[0].Name || pc.FaceUp[0].Layout != "planar" {
			t.Errorf("viewer %q: face_up %+v", viewer, pc.FaceUp)
		}
		if !v.Settings.Planechase {
			t.Errorf("viewer %q: settings.planechase is off", viewer)
		}
		raw, _ := json.Marshal(v)
		for _, c := range g.PlanarDeckCards() {
			if strings.Contains(string(raw), c.Name) {
				t.Errorf("viewer %q: the wire carries %s from the planar deck", viewer, c.Name)
			}
		}
	}
}

func TestAPlainTableHasNoPlanechaseView(t *testing.T) {
	g := planechaseViewGame(t, false)
	raw, _ := json.Marshal(ViewOfGameFor(g, g.Seats[0].ID.String()))
	if strings.Contains(string(raw), `"planechase"`) {
		t.Error("a plain table's view carries a planechase key")
	}
}

// ADR 0216 §§4, 9: the planar die's count and cost ride the view, and
// its roll and a planeswalk are log lines with no number.
func TestThePlanarDieOnTheWireAndInTheLog(t *testing.T) {
	g := planechaseViewGame(t, true)
	v := ViewOfGameFor(g, g.Seats[0].ID.String())
	if v.Planechase.DieRolls != 0 || v.Planechase.DieCost != "{0}" {
		t.Errorf("die_rolls %d, die_cost %q; want 0 and {0}", v.Planechase.DieRolls, v.Planechase.DieCost)
	}
	roll := renderRandomLogText(LogEvent{Kind: LogRoll, PlanarFace: game.PlanarFacePlaneswalk}, "Alice", "")
	if roll != "Alice rolled the planar die: the Planeswalker symbol" {
		t.Errorf("roll line %q", roll)
	}
	ev := game.Event{Kind: game.EventRollDie, Actor: g.Seats[0].ID, PlanarFace: game.PlanarFaceChaos, Sides: 6}
	raw, _ := json.Marshal(ev)
	if strings.Contains(string(raw), `"amount"`) {
		t.Errorf("a planar roll event carries a number: %s", raw)
	}
}

// #5648: an effect's "chaos ensues" (CR 311.7) is a log line of its own,
// naming the card whose effect it was.
func TestChaosEnsuingFromAnEffectIsALogLine(t *testing.T) {
	g := planechaseViewGame(t, true)
	g.WithWriteLock(func() { g.ChaosEnsuesForEffect(g.Seats[0].ID, uuid.Nil) })
	v := ViewOfGameFor(g, g.Seats[1].ID.String())
	lines := 0
	for _, e := range v.Log {
		if e.Kind != LogChaosEnsues {
			continue
		}
		lines++
		if e.Text != "Chaos ensues" {
			t.Errorf("text %q", e.Text)
		}
	}
	if lines != 1 {
		t.Errorf("%d chaos_ensues log lines, want 1", lines)
	}
	named := renderLogText(LogEvent{Kind: LogChaosEnsues, Label: "Path of the Animist"}, "", "")
	if named != "Chaos ensues (Path of the Animist)" {
		t.Errorf("named line %q", named)
	}
}
