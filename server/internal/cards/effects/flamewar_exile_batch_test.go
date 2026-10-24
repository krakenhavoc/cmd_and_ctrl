package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #4245 (CR 616.1): a prompt on the first of the three cards Flamewar's back
// face exiles must not leave the other two in the library. All three are
// exiled face down once it is answered, the marking step marks exactly those
// three and not an older exiled card of mine, and Flamewar converts back.
func TestFlamewarExilesAllItsCardsAfterAPromptOnTheFirst(t *testing.T) {
	g, me, opp := flipMain(t)
	id := transformerConverted(t, g, flamewarRow(), me)
	var top []uuid.UUID
	older := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.Cards = nil
		for i := 0; i < 3; i++ {
			top = append(top, me.Library.Cards[len(me.Library.Cards)-1-i].InstanceID)
		}
		for _, label := range []string{"first order", "second order"} {
			g.RegisterReplacementForTest(game.ReplacementEffect{
				Watches: []game.EventKind{game.EventZoneMove},
				AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) bool {
					return ev.Kind == game.RepEventMove && ev.CardID == top[0] && ev.NewZone == game.ZoneExile
				},
				Replace: func(*game.ReplacementEvent, *game.Game, *game.Card) error { return nil },
				Label:   label,
			})
		}
		g.Exile.PushTop(game.Card{InstanceID: older, Name: "Older Exile", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	})
	libBefore := me.Library.Size()

	emitCombatDamage(g, id, opp.ID, 3)
	tfSettle(t, g)

	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != game.PendingChoiceReplacementOrder {
		t.Fatalf("pending choices %+v, want the one ordering prompt", g.PendingChoices)
	}
	if got := intelCards(g, me.ID); len(got) != 0 {
		t.Fatalf("%d cards marked while the prompt was open", len(got))
	}
	prompt := g.PendingChoices[0]
	if err := g.ResolveReplacementOrder(prompt.ID, prompt.Chooser, prompt.ReplacementEffectIDs); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
	tfSettle(t, g)

	if me.Library.Size() != libBefore-3 {
		t.Fatalf("library %d -> %d, want exactly 3 exiled", libBefore, me.Library.Size())
	}
	marked := map[uuid.UUID]bool{}
	for _, m := range intelCards(g, me.ID) {
		marked[m] = true
	}
	if len(marked) != 3 || marked[older] {
		t.Fatalf("marked %v, want the three exiled cards and not the older one", marked)
	}
	for _, c := range top {
		if !marked[c] {
			t.Errorf("exiled card %s has no intel counter", c)
		}
		if zone, _ := cardWhere(g, me, c); zone != "exile" {
			t.Errorf("exiled card %s is in %s", c, zone)
		}
	}
	for _, e := range g.Exile.Cards {
		if marked[e.InstanceID] && !e.FaceDown {
			t.Errorf("%s is exiled face up", e.Name)
		}
	}
	wantFace(t, g, id, 0, "Flamewar, Brash Veteran")
}
