package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b541_cards_test.go — Shadowmoor, S82 OLD batch 541 (#5591).
// Every test plays the real catalog entry through the engine.

const (
	oldB541PutAway     = "1bc9a89c-7831-4c08-8934-896286d63e7c"
	oldB541Demigod     = "602ce676-380d-4c3c-b97a-9234ccdc2904"
	oldB541HeapDoll    = "aa7f0f15-bcbb-4af5-ad31-ff215dd81e71"
	oldB541Slinking    = "6d86dcaa-2295-4f4d-ada7-216f96138d81"
	oldB541Cinderbones = "8e3c2f44-c4ca-4555-80c6-316e62c338dd"
	oldB541Brute       = "cb69e07c-c831-4992-8089-3a3a1278c45b"
	oldB541Lockjaw     = "fadb0d89-0345-4edd-9546-dd4664e751f7"
	oldB541Madblind    = "0ee0b090-3f1e-49d6-bcad-91e0cf1d12ae"
	oldB541Moonring    = "cf620c66-7db1-4db8-ae56-ee4bc2f77d74"
	oldB541Manaforge   = "3c9535a2-7995-42de-a56f-0a445f2dd697"
	oldB541Rosheen     = "892d318e-34bd-4fca-a2a1-0c3534c68068"
)

func TestOldB541EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB541PutAway: "Put Away", oldB541Demigod: "Demigod of Revenge", oldB541HeapDoll: "Heap Doll",
		oldB541Slinking: "Slinking Giant", oldB541Cinderbones: "Cinderbones", oldB541Brute: "Inescapable Brute",
		oldB541Lockjaw: "Lockjaw Snapper", oldB541Madblind: "Madblind Mountain", oldB541Moonring: "Moonring Island",
		oldB541Manaforge: "Manaforge Cinder", oldB541Rosheen: "Rosheen Meanderer",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB541WitherCardsHaveWither(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for _, c := range []struct{ name, tl, oracle string }{
		{"Cinderbones", "Creature — Elemental Skeleton", oldB541Cinderbones},
		{"Slinking Giant", "Creature — Giant Rogue", oldB541Slinking},
		{"Inescapable Brute", "Creature — Giant Warrior", oldB541Brute},
		{"Lockjaw Snapper", "Artifact Creature — Scarecrow", oldB541Lockjaw},
	} {
		id := effectstest.B12Push(g, me.ID, c.name, c.tl, c.oracle, 2, 2)
		if !effectstest.P1g2HasKeyword(g, id, "wither") {
			t.Errorf("%s has no wither", c.name)
		}
	}
}

func TestOldB541CinderbonesRegenerates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	bones := effectstest.PushCatalogPermanent(g, me.ID, "Cinderbones", "Creature — Elemental Skeleton", oldB541Cinderbones, false)
	effectstest.B06AddMana(me, "B", "B")
	effectstest.B16Activate(t, g, me.ID, bones, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(bones); err != nil {
			t.Fatal(err)
		}
	})
	if !g.Battlefield.Contains(bones) {
		t.Fatal("the regenerated Cinderbones was destroyed")
	}
	if !effectstest.HmlTapped(t, g, bones) {
		t.Error("regeneration did not tap Cinderbones")
	}
}

func TestOldB541CinderbonesWithoutAShieldDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bones := effectstest.PushCatalogPermanent(g, me.ID, "Cinderbones", "Creature — Elemental Skeleton", oldB541Cinderbones, false)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(bones); err != nil {
			t.Fatal(err)
		}
	})
	if g.Battlefield.Contains(bones) {
		t.Fatal("Cinderbones survived with no regeneration shield")
	}
}

func TestOldB541SlinkingGiantShrinksOncePerBlockOrBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	giant := effectstest.B12Push(g, me.ID, "Slinking Giant", "Creature — Giant Rogue", oldB541Slinking, 4, 4)
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 1, 1)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Bear", 1, 1)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventBlock, Actor: me.ID, CardID: giant, Target: a, Amount: 1})
		g.EmitEvent(game.Event{Kind: game.EventBlock, Actor: me.ID, CardID: giant, Target: b, Amount: 2})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, giant); p != 1 || tough != 4 {
		t.Fatalf("after blocking two attackers the Giant is %d/%d, want 1/4: blocks triggers once", p, tough)
	}

	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	giant = effectstest.B12Push(g, me.ID, "Slinking Giant", "Creature — Giant Rogue", oldB541Slinking, 4, 4)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventBecomesBlocked, Actor: opp.ID, Source: giant, CardID: giant, Target: giant})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, giant); p != 1 || tough != 4 {
		t.Errorf("after becoming blocked the Giant is %d/%d, want 1/4", p, tough)
	}
}

func TestOldB541InescapableBruteMustBeBlocked(t *testing.T) {
	spec, ok := effects.Lookup(oldB541Brute)
	if !ok || len(spec.Static) != 1 {
		t.Fatalf("registered=%v with %d statics, want the one must-be-blocked static", ok, len(spec.Static))
	}
}

func TestOldB541LockjawSnapperSpreadsMinusCountersWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	snapper := effectstest.PushCatalogPermanent(g, me.ID, "Lockjaw Snapper", "Artifact Creature — Scarecrow", oldB541Lockjaw, false)
	marked := effectstest.B12Creature(g, opp.ID, "Marked", "Creature — Bear", 4, 4)
	clean := effectstest.B12Creature(g, opp.ID, "Clean", "Creature — Bear", 4, 4)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(marked, game.CounterMinusOne, 1); err != nil {
			t.Fatal(err)
		}
		if err := g.DestroyPermanentForEffect(snapper); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.Pr8Counters(g, marked, game.CounterMinusOne); got != 2 {
		t.Errorf("the marked creature has %d -1/-1 counters, want 2", got)
	}
	if got := effectstest.Pr8Counters(g, clean, game.CounterMinusOne); got != 0 {
		t.Errorf("the clean creature has %d -1/-1 counters, want 0", got)
	}
}

func TestOldB541HeapDollExilesATargetCardFromAGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	doll := effectstest.PushCatalogPermanent(g, me.ID, "Heap Doll", "Artifact Creature — Scarecrow", oldB541HeapDoll, false)
	card := effectstest.PushGraveyardCardForTest(opp, "Doomed Card")
	effectstest.B16Activate(t, g, me.ID, doll, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(card)})
	if g.Battlefield.Contains(doll) {
		t.Fatal("the Heap Doll was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range opp.Graveyard.Cards {
		if c.InstanceID == card {
			t.Fatal("the targeted card is still in the graveyard")
		}
	}
}

func TestOldB541ManaforgeCinderMakesManaThreeTimesEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	cinder := effectstest.PushCatalogPermanent(g, me.ID, "Manaforge Cinder", "Creature — Elemental Shaman", oldB541Manaforge, false)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.FillPoolColored(me, "G", 4)
	if err := g.ActivateManaAbility(me.ID, cinder, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err == nil {
		t.Fatal("the Cinder made {G}; it adds {B} or {R}")
	}
	for i, color := range []string{"B", "R", "B"} {
		if err := g.ActivateManaAbility(me.ID, cinder, 0, game.ManaAbilityParams{Colors: []string{color}}); err != nil {
			t.Fatalf("activation %d: %v", i+1, err)
		}
	}
	if err := g.ActivateManaAbility(me.ID, cinder, 0, game.ManaAbilityParams{Colors: []string{"R"}}); err == nil {
		t.Fatal("a fourth activation this turn was allowed")
	}
}

func TestOldB541RosheenMeandererAddsFourRestrictedColorless(t *testing.T) {
	g, me, _ := effectstest.FlipMain(t)
	id := effectstest.PushCatalogPermanent(g, me.ID, "Rosheen Meanderer", "Legendary Creature — Giant Shaman", oldB541Rosheen, false)
	effectstest.Unsick(g, id)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	n := 0
	for _, tok := range me.ManaPool {
		if tok.Color != "C" {
			continue
		}
		n += tok.N()
		if len(tok.Restrictions) != 1 || tok.Restrictions[0] != game.ManaRestrictXCost {
			t.Errorf("a {C} token has restrictions %v, want only costs that contain {X}", tok.Restrictions)
		}
	}
	if n != 4 {
		t.Fatalf("pool has %d {C}, want 4", n)
	}
}

func TestOldB541MadblindMountainShufflesOnlyWithTwoRedPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := effectstest.PushCatalogPermanent(g, me.ID, "Madblind Mountain", "Land — Mountain", oldB541Madblind, false)
	effectstest.AcPermanent(g, me.ID, "Red Knight", "Creature — Knight", 2, 2, "R")
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "R")
	effectstest.AcRefused(t, g, me.ID, land, 0, game.ActivateAbilityParams{})
	effectstest.AcPermanent(g, me.ID, "Red Archer", "Creature — Archer", 1, 1, "R")
	effectstest.B16Activate(t, g, me.ID, land, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HmlTapped(t, g, land) {
		t.Error("the land did not tap as a cost")
	}
}

func TestOldB541MoonringIslandLooksOnlyWithTwoBluePermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	land := effectstest.PushCatalogPermanent(g, me.ID, "Moonring Island", "Land — Island", oldB541Moonring, false)
	effectstest.AcPermanent(g, me.ID, "Blue Knight", "Creature — Knight", 2, 2, "U")
	effectstest.PushLibraryCardForTest(opp, game.Card{Name: "Top Card", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "U")
	effectstest.AcRefused(t, g, me.ID, land, 0, game.ActivateAbilityParams{Targets: effectstest.PlayerRefs(opp.ID)})
	effectstest.AcPermanent(g, me.ID, "Blue Mage", "Creature — Wizard", 1, 1, "U")
	effectstest.B16Activate(t, g, me.ID, land, 0, game.ActivateAbilityParams{Targets: effectstest.PlayerRefs(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	g.ReadSnapshot(func() {
		if n := len(opp.Library.Cards); n == 0 || !opp.Library.Cards[n-1].IsKnownTo(me.ID) {
			t.Error("the controller did not learn the top card")
		}
	})
	if !effectstest.HmlTapped(t, g, land) {
		t.Error("the land did not tap as a cost")
	}
}

func TestOldB541PutAwayCountersAndMayShuffleACardBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	spell := effectstest.CastFromHandForTest(t, g, opp, "Their Spell", "Instant", "", "", game.CastSpellParams{})
	back := effectstest.PushGraveyardCardForTest(me, "Old Card")
	effectstest.ApaMana(me, "U", "U", "C", "C")
	effectstest.CastFromHandForTest(t, g, me, "Put Away", "Instant", "{2}{U}{U}", oldB541PutAway, game.CastSpellParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: spell, Slot: 0}, {Kind: game.TargetCard, ID: back, Slot: 1}},
	})
	effectstest.PassPriorityUntilChoice(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.P7Countered(g, spell) {
		t.Error("the targeted spell was not countered")
	}
	for _, c := range me.Graveyard.Cards {
		if c.InstanceID == back {
			t.Error("the card is still in the graveyard after saying yes")
		}
	}
	if !oldB541LibraryHas(me, back) {
		t.Error("the card did not go into the library")
	}
}

func TestOldB541PutAwayDecliningLeavesTheGraveyardAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	spell := effectstest.CastFromHandForTest(t, g, opp, "Their Spell", "Instant", "", "", game.CastSpellParams{})
	back := effectstest.PushGraveyardCardForTest(me, "Old Card")
	effectstest.ApaMana(me, "U", "U", "C", "C")
	effectstest.CastFromHandForTest(t, g, me, "Put Away", "Instant", "{2}{U}{U}", oldB541PutAway, game.CastSpellParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: spell, Slot: 0}, {Kind: game.TargetCard, ID: back, Slot: 1}},
	})
	effectstest.PassPriorityUntilChoice(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if oldB541LibraryHas(me, back) {
		t.Error("the card was shuffled in after declining")
	}
}

func oldB541LibraryHas(p *game.Player, id uuid.UUID) bool {
	for _, c := range p.Library.Cards {
		if c.InstanceID == id {
			return true
		}
	}
	return false
}

func TestOldB541DemigodOfRevengeReturnsItsNamesakesWhenCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	one := effectstest.PushGraveyardCardForTest(me, "Demigod of Revenge")
	two := effectstest.PushGraveyardCardForTest(me, "Demigod of Revenge")
	other := effectstest.PushGraveyardCardForTest(me, "Some Other Card")
	effectstest.CastFromHandForTest(t, g, me, "Demigod of Revenge", "Creature — Spirit Avatar", "", oldB541Demigod, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{one, two} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("graveyard Demigod %s did not return to the battlefield", id)
		}
	}
	if g.Battlefield.Contains(other) {
		t.Error("a card with another name returned")
	}
}
