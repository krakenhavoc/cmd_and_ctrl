package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_seam_6169_cards_test.go — #6169: a draw knows the cycling ability
// that made it (ReplacementEvent.DrawCycled), and Unpredictable Cyclone
// replaces that draw with an exile-until-shared-type and a free cast.
// CR 614.1a: "instead" is a replacement effect. CR 702.29a: cycling is
// "[Cost], Discard this card: Draw a card."

const (
	seam6169CycloneOracle = "5b67a944-ab0b-4155-8bc0-becb1b38b3bb"
	seam6169CensorOracle  = "3a42558c-7432-4c05-9599-95343eb09be1"
	seam6169SteppeOracle  = "8dec6fcf-1254-4b1b-ba23-7a3e492a7241"
)

func TestSeam6169CycloneIsFull(t *testing.T) {
	spec, ok := effects.Lookup(seam6169CycloneOracle)
	if !ok || spec.Name != "Unpredictable Cyclone" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Unpredictable Cyclone: %+v %v", spec, ok)
	}
}

// seam6169Setup puts the Cyclone on the battlefield, the named cycling card
// in the active player's hand and the given cards on top of the library
// (the first is the top card).
func seam6169Setup(t *testing.T, name, typeLine, oracle string, top ...game.Card) (*game.Game, *game.Player, uuid.UUID, []uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedPermanentWithOracle(g, me.ID, "Unpredictable Cyclone", "Enchantment", seam6169CycloneOracle)
	cycler := uuid.New()
	var ids []uuid.UUID
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: cycler, Name: name, TypeLine: typeLine, OracleID: oracle, Owner: me.ID, Controller: me.ID})
		for i := len(top) - 1; i >= 0; i-- {
			c := top[i]
			c.InstanceID, c.Owner, c.Controller = uuid.New(), me.ID, me.ID
			me.Library.PushTop(c)
			ids = append([]uuid.UUID{c.InstanceID}, ids...)
		}
	})
	return g, me, cycler, ids
}

// Cycling an instant exiles until an instant, offers it free, and puts the
// rest on the bottom; no card is drawn.
func TestUnpredictableCycloneReplacesAnInstantsCyclingDraw(t *testing.T) {
	g, me, censor, lib := seam6169Setup(t, "Censor", "Instant", seam6169CensorOracle,
		game.Card{Name: "Top Sorcery", TypeLine: "Sorcery", ManaCost: "{2}{R}"},
		game.Card{Name: "Top Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2},
		game.Card{Name: "Free Gain", TypeLine: "Instant", ManaCost: "{4}{W}"},
		game.Card{Name: "Under", TypeLine: "Instant", ManaCost: "{U}"},
	)
	drawsBefore := effectstest.CountEvents(g, game.EventDrawCard)
	handBefore := len(me.Hand.Cards)
	effectstest.FillPoolColored(me, "U", 1)
	if err := g.ActivateCatalogAbility(me.ID, censor, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("cycle: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountEvents(g, game.EventDrawCard); got != drawsBefore {
		t.Fatalf("%d draws, want the draw replaced", got-drawsBefore)
	}
	if len(me.Hand.Cards) != handBefore-1 {
		t.Errorf("hand %d → %d, want one fewer (the cycled card, no draw)", handBefore, len(me.Hand.Cards))
	}
	hit := lib[2]
	if !g.Exile.Contains(hit) {
		t.Fatalf("the instant was not exiled: %+v", g.PendingChoices)
	}
	offer := effectstest.ResolutionCastFor(g, me.ID)
	if offer == nil {
		t.Fatalf("no free cast offered: %+v", g.PendingChoices)
	}
	if err := g.CastForResolutionChoice(me.ID, offer.ID, hit, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("cast the instant free: %v", err)
	}
	if !g.Stack.Contains(hit) {
		t.Error("the free instant is not on the stack")
	}
	bottom := map[uuid.UUID]bool{me.Library.Cards[0].InstanceID: true, me.Library.Cards[1].InstanceID: true}
	for _, id := range lib[:2] {
		if !bottom[id] {
			t.Errorf("card %s is not on the bottom of the library", id)
		}
	}
	if top, _ := me.Library.Top(); top.InstanceID != lib[3] {
		t.Error("the unexiled card is no longer on top")
	}
}

// Cycling a land is not replaced: the card is drawn.
func TestUnpredictableCycloneLeavesALandsCyclingDraw(t *testing.T) {
	g, me, steppe, lib := seam6169Setup(t, "Secluded Steppe", "Land", seam6169SteppeOracle,
		game.Card{Name: "Drawn", TypeLine: "Instant", ManaCost: "{U}"},
	)
	drawsBefore := effectstest.CountEvents(g, game.EventDrawCard)
	effectstest.FillPoolColored(me, "W", 1)
	if err := g.ActivateCatalogAbility(me.ID, steppe, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("cycle: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(lib[0]) || effectstest.CountEvents(g, game.EventDrawCard) != drawsBefore+1 {
		t.Error("a land's cycling draw was replaced")
	}
	if g.Exile.Size() != 0 {
		t.Error("cards were exiled for a land's cycling")
	}
}

// A draw that no cycling ability makes is not replaced.
func TestUnpredictableCycloneLeavesOtherDraws(t *testing.T) {
	g, me, _, lib := seam6169Setup(t, "Censor", "Instant", seam6169CensorOracle,
		game.Card{Name: "Drawn", TypeLine: "Instant", ManaCost: "{U}"},
	)
	g.WithWriteLock(func() {
		if err := g.DrawNForEffect(me.ID, 1); err != nil {
			t.Fatal(err)
		}
	})
	if !me.Hand.Contains(lib[0]) || g.Exile.Size() != 0 {
		t.Error("an ordinary draw was replaced")
	}
}
