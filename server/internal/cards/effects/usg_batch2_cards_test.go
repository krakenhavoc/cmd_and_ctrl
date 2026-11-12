package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// usg_batch2_cards_test.go — Urza's Saga, batch 2 (A–C, #5624). Each test
// plays the real catalog entry and asserts what its text says.

const (
	usgB2AbsoluteGraceOracle    = "41785f7b-3d02-4b7f-86ff-b381f626142e"
	usgB2AbsoluteLawOracle      = "f4aa8391-34d5-4ad2-a981-9aa08f8229db"
	usgB2AcidicSoilOracle       = "4ccc1ac5-db90-4fc9-9795-7afaf7466c3e"
	usgB2AngelicChorusOracle    = "e66b12ff-45da-46dc-9148-bd8ea9b7fde0"
	usgB2ArcaneLabOracle        = "f5e3c26c-c28c-455f-a682-2597e4b92d10"
	usgB2ArgothianElderOracle   = "b57f06e4-da97-4473-9fbf-f66e9129502a"
	usgB2BarrinOracle           = "8c6d6684-1943-4af3-98da-951604193911"
	usgB2BefoulOracle           = "c573c59a-5a79-4fe1-a6d8-5e0153b5c059"
	usgB2BloodVassalOracle      = "f792f40d-1d37-4854-98da-e4020c6a44d4"
	usgB2BravadoOracle          = "69b5af09-b3e6-4abf-b62e-9199c4f43ba7"
	usgB2BrilliantHaloOracle    = "8f68f6d5-1953-4ccd-8af0-7d0a69139556"
	usgB2ClawsOfGixOracle       = "c4d384d7-f294-4b2d-9971-a4689c150255"
	usgB2CloakOfMistsOracle     = "13358b1d-0a35-48ed-a7fb-70359fa3998f"
	usgB2CrosswindsOracle       = "68d7fafd-15f4-46fe-8233-09c3c873d1a1"
	usgB2CrystalChimesOracle    = "fbff7452-1da4-4cd0-8247-a625e4b4da2b"
	usgB2AngelicChorusToughness = 4
)

func TestUsgB2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		usgB2AbsoluteGraceOracle:  "Absolute Grace",
		usgB2AbsoluteLawOracle:    "Absolute Law",
		usgB2AcidicSoilOracle:     "Acidic Soil",
		usgB2AngelicChorusOracle:  "Angelic Chorus",
		usgB2ArcaneLabOracle:      "Arcane Laboratory",
		usgB2ArgothianElderOracle: "Argothian Elder",
		usgB2BarrinOracle:         "Barrin, Master Wizard",
		usgB2BefoulOracle:         "Befoul",
		usgB2BloodVassalOracle:    "Blood Vassal",
		usgB2BravadoOracle:        "Bravado",
		usgB2BrilliantHaloOracle:  "Brilliant Halo",
		usgB2ClawsOfGixOracle:     "Claws of Gix",
		usgB2CloakOfMistsOracle:   "Cloak of Mists",
		usgB2CrosswindsOracle:     "Crosswinds",
		usgB2CrystalChimesOracle:  "Crystal Chimes",
	} {
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

// usgB2Activate floats `mana` for `who`, then activates ability 0 of `id`
// with strict payment, paying any sacrifice with `sac`.
func usgB2Activate(t *testing.T, g *game.Game, who *game.Player, id uuid.UUID, mana string, targets []game.TargetRef, sac ...uuid.UUID) error {
	t.Helper()
	floatForTest(g, who, mana)
	return g.ActivateCatalogAbility(who.ID, id, 0, game.ActivateAbilityParams{Targets: targets, SacrificeIDs: sac, Strict: true})
}

// --- Absolute Grace, Absolute Law -------------------------------------------------

func TestUsgB2AbsolutesGiveEveryCreatureProtection(t *testing.T) {
	for _, tc := range []struct{ name, oracle, keyword, other string }{
		{"Absolute Grace", usgB2AbsoluteGraceOracle, "protection from black", "protection from red"},
		{"Absolute Law", usgB2AbsoluteLawOracle, "protection from red", "protection from black"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
			theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
			rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
			if hasKeywordOf(g, mine, tc.keyword) {
				t.Fatalf("control: the Bear has %s before %s", tc.keyword, tc.name)
			}
			seedEnchantment(g, tc.oracle, tc.name, "Enchantment", me.ID)
			for _, id := range []uuid.UUID{mine, theirs} {
				if !hasKeywordOf(g, id, tc.keyword) {
					t.Errorf("a creature lacks %s", tc.keyword)
				}
				if hasKeywordOf(g, id, tc.other) {
					t.Errorf("a creature gained %s too", tc.other)
				}
			}
			if hasKeywordOf(g, rock, tc.keyword) {
				t.Errorf("a noncreature artifact gained %s", tc.keyword)
			}
		})
	}
}

// --- Acidic Soil -------------------------------------------------------------------

func TestUsgB2AcidicSoilDamagesEachPlayerByTheirLands(t *testing.T) {
	g := newCatalogGame(t)
	// Seat i controls i+1 lands, the caster included; a creature is not
	// counted.
	lands := map[uuid.UUID]int{}
	for i, p := range g.Seats {
		lands[p.ID] = i + 1
		for range i + 1 {
			b12Permanent(g, p.ID, "Forest", "Basic Land — Forest")
		}
	}
	b12Creature(g, g.Seats[1].ID, "Not a Land", "Creature — Bear", 2, 2)
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = p.Life
	}
	castCatalogSpell(t, g, "Acidic Soil", "Sorcery", usgB2AcidicSoilOracle, nil)
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if want := before[p.ID] - lands[p.ID]; p.Life != want {
			t.Errorf("%s (%d lands): life %d, want %d", p.Name, lands[p.ID], p.Life, want)
		}
	}
}

// --- Angelic Chorus ----------------------------------------------------------------

func TestUsgB2AngelicChorusGainsTheEnteringCreaturesToughness(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Angelic Chorus", "Enchantment", usgB2AngelicChorusOracle, false)
	life := me.Life
	dftABCast(t, g, me, game.Card{Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: usgB2AngelicChorusToughness})
	if me.Life != life+usgB2AngelicChorusToughness {
		t.Errorf("life %d → %d, want +%d for the creature's toughness", life, me.Life, usgB2AngelicChorusToughness)
	}
	life = me.Life
	b24CastNoncreature(t, g, "Plain Anthem", "Enchantment", "{1}{W}")
	passPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("a noncreature permanent gained life: %d → %d", life, me.Life)
	}
}

// --- Arcane Laboratory -------------------------------------------------------------

func TestUsgB2ArcaneLaboratoryRefusesTheSecondSpell(t *testing.T) {
	g := newCatalogGame(t)
	// An opponent's Laboratory: "each player" binds the caster too.
	seedEnchantment(g, usgB2ArcaneLabOracle, "Arcane Laboratory", "Enchantment", g.Seats[1].ID)
	if err := castCatalogSpellErr(t, g, "Acidic Soil", "Sorcery", usgB2AcidicSoilOracle, nil); err != nil {
		t.Fatalf("the first spell of the turn was refused: %v", err)
	}
	passPriorityAroundTable(t, g)
	err := castCatalogSpellErr(t, g, "Acidic Soil", "Sorcery", usgB2AcidicSoilOracle, nil)
	assertCantCast(t, err, "more than one spell")
}

// --- Argothian Elder ---------------------------------------------------------------

func TestUsgB2ArgothianElderUntapsTwoTargetLands(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	elder := pushCatalogPermanent(g, me.ID, "Argothian Elder", "Creature — Elf Druid", usgB2ArgothianElderOracle, false)
	a := b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	b := b12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	for i := range g.Battlefield.Cards {
		if id := g.Battlefield.Cards[i].InstanceID; id == a || id == b || id == bear {
			g.Battlefield.Cards[i].Tapped = true
		}
	}
	advanceToMain(t, g)
	if err := usgB2Activate(t, g, me, elder, "", cardRefs(a, bear)); err == nil {
		t.Fatal("targeted a creature as a land")
	}
	if err := usgB2Activate(t, g, me, elder, "", cardRefs(a)); err == nil {
		t.Fatal("activated with one target; it takes exactly two")
	}
	if err := usgB2Activate(t, g, me, elder, "", cardRefs(a, b)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if b13Tapped(t, g, a) || b13Tapped(t, g, b) {
		t.Error("a target land is still tapped")
	}
	if !b13Tapped(t, g, elder) {
		t.Error("Argothian Elder did not tap")
	}
}

// --- Barrin, Master Wizard ---------------------------------------------------------

func TestUsgB2BarrinSacrificesAPermanentToBounceACreature(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	barrin := pushCatalogPermanent(g, me.ID, "Barrin, Master Wizard", "Legendary Creature — Human Wizard", usgB2BarrinOracle, true)
	rock := b12Permanent(g, me.ID, "My Rock", "Artifact")
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := usgB2Activate(t, g, me, barrin, "", cardRefs(bear), rock); err == nil {
		t.Fatal("activated without paying {2}")
	}
	if err := usgB2Activate(t, g, me, barrin, "CC", cardRefs(bear)); err == nil {
		t.Fatal("activated without paying the sacrifice")
	}
	// The {2} floated above is still in the pool.
	if err := usgB2Activate(t, g, me, barrin, "", cardRefs(bear), rock); err != nil {
		t.Fatalf("activate (summoning sick, no {T} in the cost): %v", err)
	}
	if g.Battlefield.Contains(rock) || !me.Graveyard.Contains(rock) {
		t.Error("the sacrificed artifact is still on the battlefield")
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opp.Hand.Contains(bear) {
		t.Error("the target creature did not return to its owner's hand")
	}
}

// --- Befoul -------------------------------------------------------------------------

func TestUsgB2BefoulDestroysALandOrANonblackCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	black := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Black Knight", TypeLine: "Creature — Human Knight",
		ManaCost: "{B}{B}", Colors: []string{"B"}, Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	if err := castCatalogSpellErr(t, g, "Befoul", "Sorcery", usgB2BefoulOracle, cardRefs(black)); err == nil {
		t.Fatal("Befoul targeted a black creature")
	}
	for _, tc := range []struct {
		name string
		seed func(g *game.Game) uuid.UUID
	}{
		{"a nonblack creature", func(g *game.Game) uuid.UUID {
			return b12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
		}},
		{"a land", func(g *game.Game) uuid.UUID { return b12Permanent(g, g.Seats[1].ID, "Swamp", "Basic Land — Swamp") }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := tc.seed(g)
			castCatalogSpell(t, g, "Befoul", "Sorcery", usgB2BefoulOracle, cardRefs(id))
			passPriorityAroundTable(t, g)
			if g.Battlefield.Contains(id) || !g.Seats[1].Graveyard.Contains(id) {
				t.Errorf("%s was not destroyed", tc.name)
			}
		})
	}
}

// --- Blood Vassal -------------------------------------------------------------------

func TestUsgB2BloodVassalSacrificesForTwoBlack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	vassal := pushCatalogPermanent(g, me.ID, "Blood Vassal", "Creature — Thrull", usgB2BloodVassalOracle, true)
	if err := g.ActivateManaAbility(me.ID, vassal, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility (summoning sick, no {T} in the cost): %v", err)
	}
	if got := poolColors(me); len(got) != 2 || got[0] != "B" || got[1] != "B" {
		t.Errorf("pool %v, want B B", got)
	}
	if g.Battlefield.Contains(vassal) || !me.Graveyard.Contains(vassal) {
		t.Error("the Thrull was not sacrificed")
	}
}

// --- Bravado ------------------------------------------------------------------------

func TestUsgB2BravadoCountsTheOtherCreaturesYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "Ally A", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "Ally B", "Creature — Bear", 2, 2)
	b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	b12Permanent(g, me.ID, "My Rock", "Artifact")
	auraCast(t, g, "Bravado", usgB2BravadoOracle, bear)
	if p, tough := effectivePTOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the enchanted Bear is %d/%d, want 4/4 (two other creatures you control)", p, tough)
	}
	b12Creature(g, me.ID, "Ally C", "Creature — Bear", 2, 2)
	if p, tough := effectivePTOf(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("after a third ally the Bear is %d/%d, want 5/5", p, tough)
	}
}

// --- Brilliant Halo -------------------------------------------------------------------

func TestUsgB2BrilliantHaloPumpsAndReturnsToHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	halo := auraCast(t, g, "Brilliant Halo", usgB2BrilliantHaloOracle, bear)
	if p, tough := effectivePTOf(t, g, bear); p != 3 || tough != 4 {
		t.Errorf("the enchanted Bear is %d/%d, want 3/4", p, tough)
	}
	// The public mover runs the state checks, and the Aura falling off its
	// departed host is one (CR 704.5m), as in the Rancor test.
	if err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneBattlefield},
		game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID},
		bear,
	); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(halo) || me.Graveyard.Contains(halo) {
		t.Fatal("the Halo is still on the battlefield or in the graveyard")
	}
	if !me.Hand.Contains(halo) {
		t.Error("the Halo did not return to its owner's hand")
	}
}

func TestUsgB2BrilliantHaloStaysPutWhenBounced(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	halo := auraCast(t, g, "Brilliant Halo", usgB2BrilliantHaloOracle, bear)
	if err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneBattlefield},
		game.ZoneRef{Kind: game.ZoneExile},
		halo,
	); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Contains(halo) {
		t.Error("an exiled Halo came back to hand; only the graveyard returns it")
	}
}

// --- Claws of Gix -------------------------------------------------------------------------

func TestUsgB2ClawsOfGixSacrificesAnyPermanentForOneLife(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	claws := pushCatalogPermanent(g, me.ID, "Claws of Gix", "Artifact", usgB2ClawsOfGixOracle, false)
	land := b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	theirs := b12Permanent(g, g.Seats[1].ID, "Their Forest", "Basic Land — Forest")
	if err := usgB2Activate(t, g, me, claws, "C", nil, theirs); err == nil {
		t.Fatal("sacrificed an opponent's permanent")
	}
	life := me.Life
	if err := usgB2Activate(t, g, me, claws, "", nil, land); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(land) {
		t.Error("the sacrificed land is still on the battlefield")
	}
	passPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d → %d, want +1", life, me.Life)
	}
	if !g.Battlefield.Contains(claws) {
		t.Error("Claws of Gix is not part of its own cost")
	}
}

// --- Cloak of Mists -------------------------------------------------------------------------

func TestUsgB2CloakOfMistsMakesTheHostUnblockable(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	other := b12Creature(g, me.ID, "Other Bear", "Creature — Bear", 2, 2)
	auraCast(t, g, "Cloak of Mists", usgB2CloakOfMistsOracle, bear)
	if auraRestrictions(t, g, bear)&game.CantBeBlocked == 0 {
		t.Error("the enchanted creature can be blocked")
	}
	if auraRestrictions(t, g, other)&game.CantBeBlocked != 0 {
		t.Error("an unenchanted creature can't be blocked")
	}
}

// --- Crosswinds --------------------------------------------------------------------------------

func TestUsgB2CrosswindsShrinksEveryFlier(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushKeywordCreature(g, me.ID, "My Bird", "flying")
	theirs := pushKeywordCreature(g, opp.ID, "Their Bird", "flying")
	ground := pushKeywordCreature(g, opp.ID, "Their Bear")
	seedEnchantment(g, usgB2CrosswindsOracle, "Crosswinds", "Enchantment", me.ID)
	for _, id := range []uuid.UUID{mine, theirs} {
		if p, tough := effectivePTOf(t, g, id); p != 0 || tough != 2 {
			t.Errorf("a flier is %d/%d, want 0/2", p, tough)
		}
	}
	if p, tough := effectivePTOf(t, g, ground); p != 2 || tough != 2 {
		t.Errorf("a creature without flying is %d/%d, want 2/2", p, tough)
	}
}

// --- Crystal Chimes ------------------------------------------------------------------------------

func TestUsgB2CrystalChimesReturnsEveryEnchantmentCard(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	chimes := pushCatalogPermanent(g, me.ID, "Crystal Chimes", "Artifact", usgB2CrystalChimesOracle, false)
	grave := func(name, typeLine string) uuid.UUID {
		id := uuid.New()
		me.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: me.ID, Controller: me.ID})
		return id
	}
	anthem := grave("Anthem", "Enchantment")
	aura := grave("Aura", "Enchantment — Aura")
	bear := grave("Bear", "Creature — Bear")
	theirs := uuid.New()
	g.Seats[1].Graveyard.PushTop(game.Card{InstanceID: theirs, Name: "Their Anthem", TypeLine: "Enchantment", Owner: g.Seats[1].ID, Controller: g.Seats[1].ID})
	if err := usgB2Activate(t, g, me, chimes, "CC", nil); err == nil {
		t.Fatal("activated for {2}; it costs {3}")
	}
	if err := usgB2Activate(t, g, me, chimes, "C", nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(chimes) {
		t.Error("Crystal Chimes was not sacrificed")
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{anthem, aura} {
		if !me.Hand.Contains(id) {
			t.Error("an enchantment card did not return to hand")
		}
	}
	if !me.Graveyard.Contains(bear) {
		t.Error("a creature card left the graveyard")
	}
	if !g.Seats[1].Graveyard.Contains(theirs) {
		t.Error("an opponent's enchantment card left their graveyard")
	}
}
