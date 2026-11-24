package testsu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const wandOfWonderOracle = "8f1123dd-19b7-402b-8ecd-ab62494c9fd5"

func TestWandOfWonderIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(wandOfWonderOracle)
	if !ok || spec.Name != "Wand of Wonder" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Wand of Wonder spec = %+v, registered %v", spec, ok)
	}
}

// wandBoard is a table on which Wand of Wonder's roll shows `natural`: the die
// is the first draw from the Wand's keyed stream, so the table is built under
// each of 256 RNG keys until the roll the ability actually made is the wanted
// one. Each opponent's library is `libs[seat]` top first (a Sorcery per entry
// named "Sorcery ..." and a Land per "Land ..."); the walk is run and the
// resolution left open on its first prompt.
type wandTable struct {
	g    *game.Game
	me   *game.Player
	ids  map[string]uuid.UUID
	opps []*game.Player
}

func wandRun(t *testing.T, natural int, libs [3][]string) wandTable {
	t.Helper()
	for k := 0; k < 256; k++ {
		g := effectstest.NewCatalogGame(t)
		g.SetRNGKeyForTest(effectstest.RandomModesKey(byte(k)))
		me := g.Seats[0]
		ids := map[string]uuid.UUID{}
		for i, lib := range libs {
			opp := g.Seats[i+1]
			opp.Library.Cards = nil
			for j := len(lib) - 1; j >= 0; j-- {
				typeLine := "Creature — Bear"
				switch {
				case len(lib[j]) >= 7 && lib[j][:7] == "Sorcery":
					typeLine = "Sorcery"
				case len(lib[j]) >= 7 && lib[j][:7] == "Instant":
					typeLine = "Instant"
				case len(lib[j]) >= 4 && lib[j][:4] == "Land":
					typeLine = "Basic Land — Forest"
				}
				ids[lib[j]] = effectstest.PushLibraryTyped(opp, lib[j], typeLine)
			}
		}
		wand := effectstest.PushCatalogPermanent(g, me.ID, "Wand of Wonder", "Artifact", wandOfWonderOracle, false)
		effectstest.AdvanceToMain(t, g)
		effectstest.B06AddMana(me, "C", "C", "C", "C")
		if err := g.ActivateCatalogAbility(me.ID, wand, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		rolls := effectstest.RollEventsOf(g)
		if len(rolls) == 1 && rolls[0].Amount == natural {
			return wandTable{g: g, me: me, ids: ids, opps: g.Seats[1:4]}
		}
	}
	t.Fatalf("no RNG key makes the Wand roll a %d", natural)
	return wandTable{}
}

// wandCast casts the named exiled card through the open resolution_cast offer.
func wandCast(t *testing.T, w wandTable, name string) {
	t.Helper()
	var offer *game.PendingChoice
	for _, c := range w.g.PendingChoices {
		if effectstest.IsKeywordCastPrompt(c) && c.Chooser == w.me.ID {
			offer = c
			break
		}
	}
	if offer == nil {
		t.Fatalf("no cast offer open for %s", name)
	}
	id := w.ids[name]
	if err := w.g.CastForResolutionChoice(w.me.ID, offer.ID, id, game.CastSpellParams{FromZone: string(effectstest.OfferedZone(w.g, w.me.ID, id))}); err != nil {
		t.Fatalf("casting %s: %v", name, err)
	}
}

func (w wandTable) zone(name string) game.ZoneKind { return effectstest.B12ZoneOf(w.g, w.ids[name]) }

func TestWandOfWonderLowRollCastsOneAndPutsTheRestBackShuffled(t *testing.T) {
	w := wandRun(t, 5, [3][]string{
		{"Land A", "Bear A", "Sorcery A", "Bear B"},
		{"Sorcery B", "Bear C"},
		{"Land B", "Instant A"},
	})
	// Each opponent exiled down to their first instant or sorcery; everything
	// before it went back into the library, the card itself stayed exiled.
	for _, name := range []string{"Sorcery A", "Sorcery B", "Instant A"} {
		if z := w.zone(name); z != game.ZoneExile {
			t.Errorf("%s is in %v, want exile", name, z)
		}
	}
	for _, name := range []string{"Land A", "Bear A", "Land B"} {
		if z := w.zone(name); z != game.ZoneLibrary {
			t.Errorf("%s is in %v, want back in the library", name, z)
		}
	}
	for _, name := range []string{"Bear B", "Bear C"} {
		if z := w.zone(name); z != game.ZoneLibrary {
			t.Errorf("%s was never reached but is in %v", name, z)
		}
	}
	// X is one: one cast, then the offer closes and the other two stay exiled.
	if got := effectstest.CountMayCastPrompts(w.g); got == 0 {
		t.Fatal("no free-cast offer")
	}
	wandCast(t, w, "Sorcery A")
	effectstest.PassPriorityAroundTable(t, w.g)
	if z := w.zone("Sorcery A"); z != game.ZoneGraveyard {
		t.Errorf("the cast Sorcery A is in %v, want its owner's graveyard after resolving", z)
	}
	if n := effectstest.CountMayCastPrompts(w.g); n != 0 {
		t.Errorf("%d offers still open after the one cast X = 1 allows", n)
	}
	for _, name := range []string{"Sorcery B", "Instant A"} {
		if z := w.zone(name); z != game.ZoneExile {
			t.Errorf("%s is in %v, want still exiled", name, z)
		}
	}
}

func TestWandOfWonderMiddleRollCastsTwoAndTwentyCastsThree(t *testing.T) {
	for _, tc := range []struct {
		roll, casts int
	}{{15, 2}, {20, 3}} {
		w := wandRun(t, tc.roll, [3][]string{{"Sorcery A"}, {"Instant A"}, {"Sorcery B"}})
		made := 0
		for _, name := range []string{"Sorcery A", "Instant A", "Sorcery B"} {
			if effectstest.CountMayCastPrompts(w.g) == 0 {
				break
			}
			wandCast(t, w, name)
			effectstest.PassPriorityAroundTable(t, w.g)
			made++
		}
		if made != tc.casts {
			t.Errorf("roll %d: %d casts made, want %d", tc.roll, made, tc.casts)
		}
	}
}

func TestWandOfWonderALibraryWithNoInstantOrSorcerySuffersAFullExileAndComesBack(t *testing.T) {
	w := wandRun(t, 10, [3][]string{{"Land A", "Bear A"}, {"Land B"}, {"Bear B", "Land C"}})
	for _, name := range []string{"Land A", "Bear A", "Land B", "Bear B", "Land C"} {
		if z := w.zone(name); z != game.ZoneLibrary {
			t.Errorf("%s is in %v, want back in the library", name, z)
		}
	}
	if n := effectstest.CountMayCastPrompts(w.g); n != 0 {
		t.Errorf("%d cast offers with nothing to cast", n)
	}
	if n := w.opps[0].Library.Size(); n != 2 {
		t.Errorf("first opponent's library has %d cards, want 2", n)
	}
}
