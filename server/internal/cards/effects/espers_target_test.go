package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// espers_target_test.go — Espers to Magicite's reflexive trigger
// TARGETS the creature card (CR 603.12, CR 115.1; #3846): the pick is
// made as the trigger goes on the stack, is limited to the cards the
// spell exiled, and is re-checked as the trigger resolves (CR 608.2b).

func espersTokenCopies(g *game.Game, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && c.IsToken() {
			n++
		}
	}
	return n
}

// castEspers casts Espers to Magicite and settles the spell, leaving the
// reflexive trigger's pick_target prompt open.
func castEspers(t *testing.T, g *game.Game) *game.PendingChoice {
	t.Helper()
	castCatalogSpell(t, g, "Espers to Magicite", "Instant", deck3655EspersOracle, nil)
	passPriorityAroundTable(t, g)
	return pickTargetPrompt(t, g)
}

func TestEspersTargetIsLimitedToTheCardsExiledThisWay(t *testing.T) {
	g := newCatalogGame(t)
	me, oppA, oppB := g.Seats[0], g.Seats[1], g.Seats[2]
	a := deck3655GraveyardCreature(oppA, "Alpha Bear", 2)
	b := deck3655GraveyardCreature(oppB, "Beta Bear", 3)
	// A creature card that was already in exile before the spell, and one
	// in my own graveyard, are not "exiled this way".
	earlier := uuid.New()
	g.Exile.PushTop(game.Card{InstanceID: earlier, Name: "Earlier Bear", TypeLine: "Creature — Bear", Power: 4, Toughness: 4, Owner: oppA.ID, Controller: oppA.ID})
	deck3655GraveyardCreature(me, "My Bear", 1)

	prompt := castEspers(t, g)
	if prompt.PickTargetMin != 0 || prompt.PickTargetMax != 1 {
		t.Errorf("clause count = %d..%d, want up to one (0..1)", prompt.PickTargetMin, prompt.PickTargetMax)
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range prompt.PickTargetCards {
		offered[id] = true
	}
	if len(offered) != 2 || !offered[a] || !offered[b] {
		t.Errorf("offered %v, want exactly the two creature cards exiled this way", prompt.PickTargetCards)
	}

	// An illegal pick is refused at the pick (CR 601.2c's trigger analogue).
	if err := g.ResolvePickTarget(prompt.ID, prompt.Chooser, game.TargetRef{Kind: game.TargetCard, ID: earlier}); err == nil {
		t.Fatal("a creature card exiled before the spell was accepted as a target")
	}
	answerPickTarget(t, g, b)
	passPriorityAroundTable(t, g)
	if espersTokenCopies(g, "Beta Bear") != 1 || espersTokenCopies(g, "Alpha Bear") != 0 {
		t.Errorf("copies: Beta %d, Alpha %d; want only the targeted Beta Bear",
			espersTokenCopies(g, "Beta Bear"), espersTokenCopies(g, "Alpha Bear"))
	}
}

// The target is on the stack with the trigger, so the table can respond
// to it: a target that left exile before resolution is no longer legal
// (CR 608.2b) and no copy is made.
func TestEspersTargetLeavingExileInResponseFizzlesTheCopy(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	dragon := deck3655GraveyardCreature(opp, "Elder Dragon", 5)

	castEspers(t, g)
	answerPickTarget(t, g, dragon)
	if stackFullyEmpty(g) {
		t.Fatal("the reflexive trigger is not on the stack: the table has no window to respond")
	}
	if _, err := g.Exile.Remove(dragon); err != nil {
		t.Fatalf("take the card out of exile in response: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := espersTokenCopies(g, "Elder Dragon"); n != 0 {
		t.Errorf("%d token copies of a card that left exile before the trigger resolved, want 0", n)
	}
}

// "Up to one": choosing nothing is legal and makes no token.
func TestEspersMayChooseNoTarget(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	deck3655GraveyardCreature(opp, "Elder Dragon", 5)

	prompt := castEspers(t, g)
	if err := g.ResolvePickTargets(prompt.ID, prompt.Chooser, nil); err != nil {
		t.Fatalf("declining the up-to-one target: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := espersTokenCopies(g, "Elder Dragon"); n != 0 {
		t.Errorf("%d token copies with no target chosen, want 0", n)
	}
}

// The trigger is a restore point: a snapshot taken with the targeted
// trigger on the stack re-derives its clause WITH the payload set, so the
// resolution re-check still admits the exiled card and still refuses one
// that left.
func TestEspersTargetedTriggerSurvivesASnapshotRestore(t *testing.T) {
	for _, tc := range []struct {
		name   string
		leaves bool
		want   int
	}{
		{"target still exiled", false, 1},
		{"target left exile", true, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			opp := g.Seats[1]
			dragon := deck3655GraveyardCreature(opp, "Elder Dragon", 5)

			castEspers(t, g)
			answerPickTarget(t, g, dragon)
			restored, err := g.CaptureSnapshot().Restore()
			if err != nil {
				t.Fatalf("restore: %v", err)
			}
			if tc.leaves {
				if _, err := restored.Exile.Remove(dragon); err != nil {
					t.Fatalf("remove from exile: %v", err)
				}
			}
			passPriorityAroundTable(t, restored)
			if n := espersTokenCopies(restored, "Elder Dragon"); n != tc.want {
				t.Errorf("token copies = %d, want %d", n, tc.want)
			}
		})
	}
}
