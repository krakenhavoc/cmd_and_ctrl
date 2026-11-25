package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b102_cards_test.go — S82 OLD batch 102 (#5510, Jumpstart): activated
// abilities, life gain and loss. Each test plays the real catalog entry.

const (
	oldB102Entrancer = "04cc5292-a485-4bb8-a143-9364229d480f"
	oldB102Merfolk   = "19ff1424-6505-4702-86ad-2d7518b584f6"
	oldB102Brute     = "3835cb81-b6e4-4be8-9d49-8307ddc2e1a2"
	oldB102Bat       = "542887e9-1032-4396-9cbb-aa3c11ee43a0"
	oldB102Familiar  = "12eeabe1-a02c-42a3-aa87-6774c4de7093"
	oldB102Pilgrim   = "264fafb7-3bd3-4bf1-81b2-2e436c9ca8d0"
	oldB102TakeHeart = "2c5e30c3-0b9f-4d3e-af44-be7fcd34df25"
	oldB102Feeder    = "f71bbaf2-02ea-4fc0-9a39-49d6689b36cd"
	oldB102Path      = "fc04bfa9-a8f9-478d-8164-34b1e6c85417"
)

func TestOldB102EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB102Entrancer: "Vedalken Entrancer",
		oldB102Merfolk:   "Wishful Merfolk",
		oldB102Brute:     "Chained Brute",
		oldB102Bat:       "Blighted Bat",
		oldB102Familiar:  "Malakir Familiar",
		oldB102Pilgrim:   "Stone Haven Pilgrim",
		oldB102TakeHeart: "Take Heart",
		oldB102Feeder:    "Nocturnal Feeder",
		oldB102Path:      "Path of Bravery",
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

func TestOldB102VedalkenEntrancerMillsTwoFromTheChosenPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ent := effectstest.B12Push(g, me.ID, "Vedalken Entrancer", "Creature — Vedalken Wizard", oldB102Entrancer, 1, 4)
	effectstest.SeedLibrary(opp, "One", "Two", "Three")
	before, gy := opp.Library.Size(), opp.Graveyard.Size()
	effectstest.FloatMana(t, g, me, "{U}")
	effectstest.Pr7Activate(t, g, me.ID, ent, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}})
	if got := before - opp.Library.Size(); got != 2 {
		t.Errorf("milled %d, want 2", got)
	}
	if got := opp.Graveyard.Size() - gy; got != 2 {
		t.Errorf("graveyard grew by %d, want 2", got)
	}
}

func TestOldB102WishfulMerfolkLosesDefenderAndBecomesAHumanUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	m := effectstest.B12Push(g, me.ID, "Wishful Merfolk", "Creature — Merfolk", oldB102Merfolk, 3, 2)
	if !effectstest.HasEffectiveKeyword(t, g, m, "defender") {
		t.Fatal("the Merfolk has no defender")
	}
	effectstest.FloatMana(t, g, me, "{U}{U}")
	effectstest.Pr7Activate(t, g, me.ID, m, 0, game.ActivateAbilityParams{})
	if effectstest.HasEffectiveKeyword(t, g, m, "defender") {
		t.Error("it still has defender")
	}
	if st := effectstest.EffectiveOf(t, g, m).Subtypes; len(st) != 1 || st[0] != "Human" {
		t.Errorf("subtypes %v, want [Human]", st)
	}
	seat := g.Turn.ActiveSeat
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.AdvanceToStepOf(t, g, (seat+1)%len(g.Seats), game.StepPrecombatMain)
	if !effectstest.HasEffectiveKeyword(t, g, m, "defender") {
		t.Error("defender did not return after the turn")
	}
	if st := effectstest.EffectiveOf(t, g, m).Subtypes; len(st) != 1 || st[0] != "Merfolk" {
		t.Errorf("subtypes %v after the turn, want [Merfolk]", st)
	}
}

func TestOldB102ChainedBruteUntapsOnlyForAnotherCreatureOnYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	brute := effectstest.B12Push(g, me.ID, "Chained Brute", "Creature — Devil", oldB102Brute, 4, 3)
	fodder := effectstest.SeedCreature(g, "Fodder", me.ID)
	effectstest.TapForTest(t, g, brute)
	effectstest.FloatMana(t, g, me, "{R}")
	if err := g.ActivateCatalogAbility(me.ID, brute, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{brute}}); err == nil {
		t.Fatal("the Brute sacrificed itself")
	}
	effectstest.Pr7Activate(t, g, me.ID, brute, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}})
	if g.Battlefield.Contains(fodder) {
		t.Error("the other creature was not sacrificed")
	}
	if tapped, _ := effectstest.BattlefieldCardTapped(g, brute); tapped {
		t.Error("the Brute did not untap")
	}
}

func TestOldB102ChainedBruteDoesNotUntapInYourUntapStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	brute := effectstest.B12Push(g, me.ID, "Chained Brute", "Creature — Devil", oldB102Brute, 4, 3)
	effectstest.TapForTest(t, g, brute)
	seat := g.Turn.ActiveSeat
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.AdvanceToStepOf(t, g, seat, game.StepPrecombatMain)
	if !effectstest.B16Tapped(t, g, brute) {
		t.Error("the Brute untapped during its controller's untap step")
	}
}

func TestOldB102BlightedBatGainsHasteUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	bat := effectstest.B12Push(g, me.ID, "Blighted Bat", "Creature — Zombie Bat", oldB102Bat, 2, 1)
	if effectstest.HasEffectiveKeyword(t, g, bat, "haste") {
		t.Fatal("the Bat has haste already")
	}
	effectstest.FloatMana(t, g, me, "{B}")
	effectstest.Pr7Activate(t, g, me.ID, bat, 0, game.ActivateAbilityParams{})
	if !effectstest.HasEffectiveKeyword(t, g, bat, "haste") {
		t.Error("the Bat did not gain haste")
	}
}

func TestOldB102MalakirFamiliarGrowsOnceForEachLifeGain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	f := effectstest.B12Push(g, me.ID, "Malakir Familiar", "Creature — Bat", oldB102Familiar, 2, 1)
	effectstest.E2ChangeLife(g, opp.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, f); p != 2 {
		t.Errorf("an opponent's life gain pumped it: power %d", p)
	}
	effectstest.E2ChangeLife(g, me.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, f); p != 3 {
		t.Errorf("power %d after one life gain, want 3", p)
	}
	effectstest.E2ChangeLife(g, me.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, f); p != 4 {
		t.Errorf("power %d after two life gains, want 4", p)
	}
}

func TestOldB102StoneHavenPilgrimNeedsAnArtifactOrEnchantmentToGrow(t *testing.T) {
	for _, tc := range []struct {
		name, typ string
		grows     bool
	}{{"none", "", false}, {"artifact", "Artifact", true}, {"enchantment", "Enchantment", true}, {"land", "Basic Land — Forest", false}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me := g.Seats[g.Turn.ActiveSeat]
			opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			p := effectstest.B12Push(g, me.ID, "Stone Haven Pilgrim", "Creature — Kor Cleric", oldB102Pilgrim, 2, 2)
			if tc.typ != "" {
				effectstest.PushPermanent(g, me.ID, game.Card{Name: "Thing", TypeLine: tc.typ, Owner: me.ID, Controller: me.ID})
			}
			effectstest.DeclareAttack(t, g, opp.ID, p)
			effectstest.PassPriorityAroundTable(t, g)
			pw := effectstest.EffectivePower(t, g, p)
			ll := effectstest.HasEffectiveKeyword(t, g, p, "lifelink")
			if tc.grows && (pw != 3 || !ll) {
				t.Errorf("power %d lifelink %v, want 3 and lifelink", pw, ll)
			}
			if !tc.grows && (pw != 2 || ll) {
				t.Errorf("power %d lifelink %v, want 2 and none", pw, ll)
			}
		})
	}
}

func TestOldB102TakeHeartPumpsAndGainsLifeForEachAttackerYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	a := effectstest.B12Push(g, me.ID, "Attacker A", "Creature — Bear", "", 2, 2)
	b := effectstest.B12Push(g, me.ID, "Attacker B", "Creature — Bear", "", 2, 2)
	idle := effectstest.B12Push(g, me.ID, "Idle", "Creature — Bear", "", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	card := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: card, Name: "Take Heart", TypeLine: "Instant", OracleID: oldB102TakeHeart, Owner: me.ID, Controller: me.ID})
	life := me.Life
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Targets: effectstest.CardRefs(idle)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, idle); p != 4 {
		t.Errorf("target power %d, want 4", p)
	}
	if got := me.Life - life; got != 2 {
		t.Errorf("gained %d life, want 2 (two attackers)", got)
	}
}

func TestOldB102NocturnalFeederDrainsTwoWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	f := effectstest.B12Push(g, me.ID, "Nocturnal Feeder", "Creature — Vampire Rogue", oldB102Feeder, 2, 1)
	mine, theirs := me.Life, make([]int, len(g.Seats))
	for i, s := range g.Seats {
		theirs[i] = s.Life
	}
	effectstest.KillCreature(t, g, me.ID, f)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Life - mine; got != 2 {
		t.Errorf("gained %d, want 2", got)
	}
	for i, s := range g.Seats {
		if s.ID != me.ID && theirs[i]-s.Life != 2 {
			t.Errorf("opponent %d lost %d, want 2", i, theirs[i]-s.Life)
		}
	}
}

func TestOldB102PathOfBraveryPumpsAtStartingLifeAndGainsLifeWhenYouAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.PushCatalogPermanent(g, me.ID, "Path of Bravery", "Enchantment", oldB102Path, false)
	a := effectstest.B12Push(g, me.ID, "Attacker A", "Creature — Bear", "", 2, 2)
	b := effectstest.B12Push(g, me.ID, "Attacker B", "Creature — Bear", "", 2, 2)
	start := me.Life
	if p := effectstest.EffectivePower(t, g, a); p != 3 {
		t.Errorf("power %d at the starting life total, want 3", p)
	}
	effectstest.E2ChangeLife(g, me.ID, -1)
	if p := effectstest.EffectivePower(t, g, a); p != 2 {
		t.Errorf("power %d below the starting life total, want 2", p)
	}
	effectstest.E2ChangeLife(g, me.ID, 1)
	effectstest.DeclareAttack(t, g, opp.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Life - start; got != 2 {
		t.Errorf("gained %d life for two attackers, want 2 (one trigger)", got)
	}
}
