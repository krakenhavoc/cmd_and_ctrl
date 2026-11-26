package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b413_test.go — Gatecrash, batch 413 (#5565): every card played
// through the engine.

const (
	oldB413ScatterArc  = "424683a5-d239-400d-af2b-9c00d1d32d7f"
	oldB413Broker      = "4993e373-4b76-4582-a789-6a3123e7ca89"
	oldB413Confession  = "49b84c50-ca0a-4947-8225-ae41aa08e1f1"
	oldB413Primordial  = "52b88034-bf0f-4d29-b776-40ce421f1107"
	oldB413Tilling     = "c6b915b3-b504-42fb-8059-62a3e8e224e1"
	oldB413SignalClans = "fd68ae3c-52cd-4b20-b015-bd31378d972e"
)

func TestOldB413AllRegisteredFull(t *testing.T) {
	for id, name := range map[string]string{
		oldB413ScatterArc:  "Scatter Arc",
		oldB413Broker:      "Bane Alley Broker",
		oldB413Confession:  "Coerced Confession",
		oldB413Primordial:  "Sylvan Primordial",
		oldB413Tilling:     "Frenzied Tilling",
		oldB413SignalClans: "Signal the Clans",
	} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s registration: %+v", name, spec)
		}
	}
}

// --- Scatter Arc ---------------------------------------------------------------

func TestOldB413ScatterArcCountersANoncreatureSpellAndDrawsACard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	them := g.Seats[1]
	effectstest.Msc18Library(them, 6)
	effectstest.ToMainForCost(t, g)
	victim := effectstest.CastCatalogSpell(t, g, "Their Sorcery", "Sorcery", "", nil)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	hand := them.Hand.Size()
	effectstest.CastFromSeat(t, g, them, "Scatter Arc", "Instant", oldB413ScatterArc, effectstest.CardTargets(victim))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Stack.Contains(victim) {
		t.Error("the sorcery is still on the stack")
	}
	if !g.Seats[g.Turn.ActiveSeat].Graveyard.Contains(victim) {
		t.Error("the countered sorcery is not in its owner's graveyard")
	}
	// The cast took Scatter Arc out of the hand it had been pushed into.
	if got := them.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d (Scatter Arc drew a card)", got, hand+1)
	}
}

func TestOldB413ScatterArcRefusesACreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	them := g.Seats[1]
	effectstest.ToMainForCost(t, g)
	victim := effectstest.CastCatalogSpell(t, g, "Their Bear", "Creature — Bear", "", nil)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	id := uuid.New()
	them.Hand.PushTop(game.Card{InstanceID: id, Name: "Scatter Arc", TypeLine: "Instant",
		OracleID: oldB413ScatterArc, Owner: them.ID, Controller: them.ID})
	if err := g.CastSpell(them.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardTargets(victim)}); err == nil {
		t.Error("Scatter Arc targeted a creature spell")
	}
}

// --- Coerced Confession -----------------------------------------------------------

func TestOldB413CoercedConfessionDrawsForEachCreatureMilled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.Msc18Library(me, 8)
	effectstest.Msc18Library(opp, 8)
	for _, c := range []struct{ name, typ string }{
		{"Mill Land", "Land"}, {"Mill Bear B", "Creature — Bear"}, {"Mill Sorcery", "Sorcery"}, {"Mill Bear A", "Creature — Bear"},
	} {
		effectstest.SeedLibraryTop(opp, c.name, c.typ, "")
	}
	effectstest.ToMainForCost(t, g)
	handBefore, libBefore := me.Hand.Size(), opp.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Coerced Confession", "Sorcery", oldB413Confession,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Library.Size(); got != libBefore-4 {
		t.Errorf("opponent's library %d, want %d", got, libBefore-4)
	}
	if got := opp.Graveyard.Size(); got != 4 {
		t.Errorf("opponent's graveyard %d, want 4", got)
	}
	// The cast itself removes the spell from the hand it was pushed into.
	if got := me.Hand.Size(); got != handBefore+2 {
		t.Errorf("my hand %d, want %d (two creature cards milled)", got, handBefore+2)
	}
}

// --- Frenzied Tilling ---------------------------------------------------------------

func TestOldB413FrenziedTillingDestroysALandAndFetchesATappedBasic(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := effectstest.PushCatalogPermanent(g, opp.ID, "Their Island", "Basic Land — Island", "", false)
	basic := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.CastCatalogSpell(t, g, "Frenzied Tilling", "Sorcery", oldB413Tilling, effectstest.CardTargets(land))
	effectstest.PassPriorityAroundTable(t, g)
	open := effectstest.ChkB1OpenSearch(g)
	if open == nil {
		t.Fatal("no library search opened")
	}
	if err := g.ResolveSearchLibrary(open.ID, open.Chooser, []uuid.UUID{basic}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Error("the targeted land is still on the battlefield")
	}
	c := effectstest.FindBattlefieldCardForTest(g, basic)
	if c == nil || !c.Tapped {
		t.Errorf("the basic land is not on the battlefield tapped: %+v", c)
	}
}

func TestOldB413FrenziedTillingNeedsALandTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.SeedBear(g, opp.ID)
	if err := effectstest.CastCatalogSpellErr(t, g, "Frenzied Tilling", "Sorcery", oldB413Tilling, effectstest.CardTargets(bear)); err == nil {
		t.Error("Frenzied Tilling targeted a creature")
	}
}

// --- Signal the Clans -------------------------------------------------------------------

func oldB413Creatures(me *game.Player, names ...string) []uuid.UUID {
	var ids []uuid.UUID
	for _, n := range names {
		ids = append(ids, effectstest.PushLibraryCardForTest(me, game.Card{Name: n, TypeLine: "Creature — Bear"}))
	}
	return ids
}

func oldB413SignalSearch(t *testing.T, g *game.Game, ids []uuid.UUID) {
	t.Helper()
	effectstest.CastCatalogSpell(t, g, "Signal the Clans", "Instant", oldB413SignalClans, nil)
	effectstest.PassPriorityAroundTable(t, g)
	open := effectstest.ChkB1OpenSearch(g)
	if open == nil {
		t.Fatal("no library search opened")
	}
	if err := g.ResolveSearchLibrary(open.ID, open.Chooser, ids); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB413SignalTheClansPutsOneOfThreeDifferentlyNamedCreaturesInHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := oldB413Creatures(me, "Alpha", "Beta", "Gamma")
	hand := me.Hand.Size()
	oldB413SignalSearch(t, g, ids)
	inHand := 0
	for _, id := range ids {
		if me.Hand.Contains(id) {
			inHand++
		} else if !me.Library.Contains(id) {
			t.Errorf("a revealed card is neither in hand nor in the library")
		}
	}
	if inHand != 1 {
		t.Errorf("%d of the three revealed creatures reached the hand, want exactly 1", inHand)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
}

func TestOldB413SignalTheClansTakesNothingWhenTwoShareAName(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := oldB413Creatures(me, "Alpha", "Alpha", "Gamma")
	oldB413SignalSearch(t, g, ids)
	for _, id := range ids {
		if me.Hand.Contains(id) {
			t.Error("a creature reached the hand though two of the three share a name")
		}
		if !me.Library.Contains(id) {
			t.Error("a revealed creature is not back in the library")
		}
	}
}

// --- Sylvan Primordial ---------------------------------------------------------------------

func TestOldB413SylvanPrimordialDestroysOnePermanentPerOpponentAndFetchesThatManyForests(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	var rocks []uuid.UUID
	for _, p := range g.Seats {
		if p.ID != me.ID {
			rocks = append(rocks, effectstest.PushCatalogPermanent(g, p.ID, "Rock", "Artifact", "", false))
		}
	}
	mine := effectstest.PushCatalogPermanent(g, me.ID, "My Rock", "Artifact", "", false)
	var forests []uuid.UUID
	for i := 0; i < 5; i++ {
		forests = append(forests, effectstest.PushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"}))
	}
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"})
	effectstest.CastCatalogSpell(t, g, "Sylvan Primordial", "Creature — Avatar", oldB413Primordial, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range rocks {
		effectstest.PickCard(t, g, me.ID, id)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range rocks {
		if g.Battlefield.Contains(id) {
			t.Error("an opponent's noncreature permanent survived")
		}
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("my own permanent was destroyed")
	}
	open := effectstest.ChkB1OpenSearch(g)
	if open == nil {
		t.Fatal("no library search opened")
	}
	if err := g.ResolveSearchLibrary(open.ID, open.Chooser, forests[:len(rocks)]); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for i, id := range forests {
		c := effectstest.FindBattlefieldCardForTest(g, id)
		if i < len(rocks) && (c == nil || !c.Tapped) {
			t.Errorf("forest %d did not enter tapped", i)
		}
		if i >= len(rocks) && c != nil {
			t.Errorf("forest %d entered but only %d were destroyed", i, len(rocks))
		}
	}
}

// --- Bane Alley Broker -----------------------------------------------------------------------

func TestOldB413BaneAlleyBrokerDrawsExilesFaceDownAndReturnsTheCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.Msc18Library(me, 6)
	effectstest.AdvanceToMain(t, g)
	broker := effectstest.PushCatalogPermanent(g, me.ID, "Bane Alley Broker", "Creature — Human Rogue", oldB413Broker, false)
	keep := effectstest.HandCardForTest(me, "Kept Card", "Sorcery", "")
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, broker, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("draw/exile: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsPrompt(t, g)
	if err := g.ResolveChooseCards(pick.ID, pick.Chooser, []uuid.UUID{keep}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d, want %d (drew one, exiled one)", got, hand)
	}
	var linked *game.Card
	for i := range g.Exile.Cards {
		if g.Exile.Cards[i].InstanceID == keep {
			linked = &g.Exile.Cards[i]
		}
	}
	if linked == nil {
		t.Fatal("the chosen card is not in exile")
	}
	if linked.ExiledWith.ID != broker {
		t.Errorf("the exiled card is linked to %v, want the Broker", linked.ExiledWith.ID)
	}
	if !linked.IsKnownTo(me.ID) {
		t.Error("the Broker's controller may not look at the card")
	}
	// Return it: {U}{B}, {T}.
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, broker).Tapped = false })
	effectstest.B06AddMana(me, "U", "B")
	if err := g.ActivateCatalogAbility(me.ID, broker, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("return: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick = effectstest.ChooseCardsPrompt(t, g)
	if err := g.ResolveChooseCards(pick.ID, pick.Chooser, []uuid.UUID{keep}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(keep) {
		t.Error("the card did not return to its owner's hand")
	}
}
