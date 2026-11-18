package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_vow_b10_cards_test.go — Innistrad: Crimson Vow batch 10 (#5495):
// graveyard value, Zombies, self-mill and recursion. Each test plays the real
// catalog entry through the engine.

const (
	vowB10ArchghoulOracle   = "575ffeb4-6d12-4f69-bcc9-15db325daac1"
	vowB10LancerOracle      = "200bbf1d-835d-4d61-a6ab-a0b2abfbed79"
	vowB10SkaabOracle       = "e11b2b6e-3819-4dcb-b64e-2dfe6a6b049d"
	vowB10EntomberOracle    = "3b97794a-3aaa-45a6-9cad-f1b768615b6d"
	vowB10SpecimenOracle    = "5f4ad578-f2c4-47f6-8080-8fbc17383f6f"
	vowB10DyingOracle       = "2b719057-1bc0-4c3a-97c3-8377844570dd"
	vowB10ThrongOracle      = "7e484372-a052-4df6-a6b2-2c6accccabef"
	vowB10SwarmOracle       = "5ea5d171-a7b8-4269-9405-d2f4ccf0cbcb"
	vowB10WurmOracle        = "db4f67a9-6b47-4abe-9367-718da9e0609e"
	vowB10FearOracle        = "5187c0b2-1996-42ea-bd77-3031b3add596"
	vowB10ScatteredOracle   = "7594d892-1589-4643-a72a-2cfb2730a83d"
	vowB10WitnessOracle     = "fc7f00ad-7cc1-4be4-9924-16531e4fb829"
	vowB10RetrieveOracle    = "157a4cba-b2e6-4c42-b7dc-62db19625c9c"
	vowB10ZombieTypeLine    = "Creature — Zombie"
	vowB10CreatureTypeLine  = "Creature — Bear"
	vowB10ArtifactTypeLine  = "Artifact"
	vowB10SorceryTypeLine   = "Sorcery"
	vowB10GraveyardCreature = "Creature — Bear"
)

func TestVowB10EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		vowB10ArchghoulOracle: "Archghoul of Thraben",
		vowB10LancerOracle:    "Cobbled Lancer",
		vowB10SkaabOracle:     "Skywarp Skaab",
		vowB10EntomberOracle:  "Selhoff Entomber",
		vowB10SpecimenOracle:  "Persistent Specimen",
		vowB10DyingOracle:     "Dying to Serve",
		vowB10ThrongOracle:    "Wretched Throng",
		vowB10SwarmOracle:     "Screaming Swarm",
		vowB10WurmOracle:      "Bramble Wurm",
		vowB10FearOracle:      "Fear of Death",
		vowB10ScatteredOracle: "Scattered Thoughts",
		vowB10WitnessOracle:   "Witness the Future",
		vowB10RetrieveOracle:  "Retrieve",
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// vowB10Cast puts the card in `me`'s hand, floats `pool` and casts it under
// the strict mana gate.
func vowB10Cast(t *testing.T, g *game.Game, me *game.Player, card game.Card, pool string, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	toMainForCost(t, g)
	card.InstanceID = uuid.New()
	card.Owner, card.Controller = me.ID, me.ID
	me.Hand.PushTop(card)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	floatForTest(g, me, pool)
	params.Strict = true
	return card.InstanceID, g.CastSpell(me.ID, card.InstanceID, params)
}

func vowB10GraveyardCreatures(p *game.Player, n int) []uuid.UUID {
	ids := make([]uuid.UUID, n)
	for i := range ids {
		ids[i] = pushCatalogGraveyardCard(p, "Dead Bear", vowB10CreatureTypeLine, "", 2, 2)
	}
	return ids
}

// --- Persistent Specimen ------------------------------------------------------

func TestVowB10PersistentSpecimenReturnsTappedFromTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	id, _ := activateFromGraveyard(t, g, "Persistent Specimen", "Creature — Skeleton", vowB10SpecimenOracle, 1, 1, "{C}{C}{B}", game.ActivateAbilityParams{Strict: true})
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCardByID(g, id)
	if !ok {
		t.Fatal("Persistent Specimen is not on the battlefield")
	}
	if !c.Tapped {
		t.Error("Persistent Specimen must enter tapped")
	}
}

func TestVowB10PersistentSpecimenCannotBePaidForWithTooLittleMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMainForCost(t, g)
	id := pushCatalogGraveyardCard(me, "Persistent Specimen", "Creature — Skeleton", vowB10SpecimenOracle, 1, 1)
	floatForTest(g, me, "CB")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated for {1}{B} when {2}{B} is owed")
	}
}

func TestVowB10PersistentSpecimenDoesNothingOnceExiled(t *testing.T) {
	g := newCatalogGame(t)
	id, me := activateFromGraveyard(t, g, "Persistent Specimen", "Creature — Skeleton", vowB10SpecimenOracle, 1, 1, "{C}{C}{B}", game.ActivateAbilityParams{Strict: true})
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneExile}, id); err != nil {
		t.Fatalf("exile in response: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(id) {
		t.Error("an exiled Specimen returned")
	}
}

// --- Bramble Wurm -------------------------------------------------------------

func TestVowB10BrambleWurmGainsFiveWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := me.Life
	if _, err := vowB10Cast(t, g, me, game.Card{Name: "Bramble Wurm", TypeLine: "Creature — Wurm", ManaCost: "{6}{G}", OracleID: vowB10WurmOracle, Power: 7, Toughness: 6}, "CCCCCCG", game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != before+5 {
		t.Errorf("life %d, want %d", me.Life, before+5)
	}
}

func TestVowB10BrambleWurmExilesItselfFromTheGraveyardToGainFive(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := me.Life
	id, _ := activateFromGraveyard(t, g, "Bramble Wurm", "Creature — Wurm", vowB10WurmOracle, 7, 6, "{C}{C}{G}", game.ActivateAbilityParams{Strict: true})
	if zoneOf(g, id) != game.ZoneExile {
		t.Errorf("Bramble Wurm is in %q, want exile (a cost)", zoneOf(g, id))
	}
	passPriorityAroundTable(t, g)
	if me.Life != before+5 {
		t.Errorf("life %d, want %d", me.Life, before+5)
	}
}

// --- Selhoff Entomber ---------------------------------------------------------

func TestVowB10SelhoffEntomberDiscardsACreatureCardToDraw(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[0]
	ent := b12Push(g, me.ID, "Selhoff Entomber", "Creature — Zombie", vowB10EntomberOracle, 1, 3)
	spell := handCardForTest(me, "Lightning Bolt", "Instant", "")
	bear := handCardForTest(me, "Hand Bear", vowB10CreatureTypeLine, "")
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, ent, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{spell}}); err == nil {
		t.Fatal("discarded a noncreature card")
	}
	if err := g.ActivateCatalogAbility(me.ID, ent, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if zoneOf(g, bear) != game.ZoneGraveyard {
		t.Errorf("the discarded creature is in %q", zoneOf(g, bear))
	}
	if me.Hand.Size() != hand {
		t.Errorf("hand %d, want %d (one discarded, one drawn)", me.Hand.Size(), hand)
	}
	if c, _ := battlefieldCardByID(g, ent); !c.Tapped {
		t.Error("the Entomber was not tapped for its cost")
	}
}

// --- Fear of Death ------------------------------------------------------------

func TestVowB10FearOfDeathMillsTwoAndShrinksByTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushVanillaCreature(g, opp.ID, "Victim", 6, 6)
	vowB10GraveyardCreatures(me, 1)
	spmB17Top(me, game.Card{Name: "Milled A", TypeLine: vowB10SorceryTypeLine})
	spmB17Top(me, game.Card{Name: "Milled B", TypeLine: vowB10SorceryTypeLine})
	castCatalogSpell(t, g, "Fear of Death", "Enchantment — Aura", vowB10FearOracle, cardRefs(victim))
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if got := me.Graveyard.Size(); got != 3 {
		t.Fatalf("graveyard holds %d, want 3 (one seeded, two milled)", got)
	}
	if got := apaLive(g, victim).CurrentPower(); got != 3 {
		t.Errorf("power %d, want 3 (-3/-0)", got)
	}
	if got := apaLive(g, victim).CurrentToughness(); got != 6 {
		t.Errorf("toughness %d, want 6", got)
	}
	// Your graveyard only, read live.
	pushCatalogGraveyardCard(opp, "Their Card", vowB10CreatureTypeLine, "", 1, 1)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if got := apaLive(g, victim).CurrentPower(); got != 3 {
		t.Errorf("power %d after an opponent's graveyard grew, want 3", got)
	}
	spmB17Top(me, game.Card{Name: "Milled C", TypeLine: vowB10SorceryTypeLine})
	g.WithWriteLock(func() {
		if _, err := g.MillToZoneForEffect(me.ID, 1, game.ZoneGraveyard); err != nil {
			t.Fatalf("mill: %v", err)
		}
	})
	if got := apaLive(g, victim).CurrentPower(); got != 2 {
		t.Errorf("power %d after your graveyard grew, want 2", got)
	}
}

// --- Cobbled Lancer -----------------------------------------------------------

func TestVowB10CobbledLancerNeedsAGraveyardCreatureToExile(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	card := game.Card{Name: "Cobbled Lancer", TypeLine: "Creature — Zombie Horse", ManaCost: "{U}", OracleID: vowB10LancerOracle, Power: 3, Toughness: 3}
	if _, err := vowB10Cast(t, g, me, card, "U", game.CastSpellParams{}); err == nil {
		t.Fatal("cast with no creature card in the graveyard")
	}
	ids := vowB10GraveyardCreatures(me, 1)
	artifact := pushCatalogGraveyardCard(me, "Dead Rock", vowB10ArtifactTypeLine, "", 0, 0)
	if _, err := vowB10Cast(t, g, me, card, "U", game.CastSpellParams{ExileIDs: []uuid.UUID{artifact}}); err == nil {
		t.Fatal("exiled a noncreature card to pay")
	}
	lancer, err := vowB10Cast(t, g, me, card, "U", game.CastSpellParams{ExileIDs: ids})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	if zoneOf(g, ids[0]) != game.ZoneExile {
		t.Errorf("the creature card is in %q, want exile", zoneOf(g, ids[0]))
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(lancer) {
		t.Error("Cobbled Lancer did not resolve")
	}
}

func TestVowB10CobbledLancerExilesItselfFromTheGraveyardToDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hand := me.Hand.Size()
	id, _ := activateFromGraveyard(t, g, "Cobbled Lancer", "Creature — Zombie Horse", vowB10LancerOracle, 3, 3, "{C}{C}{C}{U}", game.ActivateAbilityParams{Strict: true})
	if zoneOf(g, id) != game.ZoneExile {
		t.Errorf("Lancer is in %q, want exile", zoneOf(g, id))
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
}
