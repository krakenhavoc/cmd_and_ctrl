package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// clb_b1_cards_test.go — the first batch of Commander Legends: Battle for
// Baldur's Gate cards (#5490). Each test plays the real catalog entry and
// asserts what its text says.

const (
	clbB1ArmorOracle      = "64de050c-7167-4ebd-8584-f064d8be3302"
	clbB1BlurOracle       = "00b36996-43c6-42a5-892c-c7c8864cf973"
	clbB1CloakOracle      = "ecd28e2a-e172-4f4c-8416-b9a0d97ee497"
	clbB1SchematicOracle  = "5b0dcf09-df51-4ad3-848e-c0ffb25824f9"
	clbB1RaphaelOracle    = "23bd166b-d766-4d03-8016-6708481d2195"
	clbB1ShadowheartOracl = "180919a5-67b2-473a-9b90-a828bdc2477c"
	clbB1StoneskinOracle  = "6d1d53e3-60ee-40a1-89a5-0648760cd83f"
	clbB1ThrakkusOracle   = "f7cd7ede-51c1-4117-99b5-484d0d80c95f"
	clbB1MyconidOracle    = "127f5b7e-79bc-4fe0-b9d4-1a267a6a4d46"
	clbB1TempleOracle     = "f48f69c0-cca8-4f44-84cc-71e7cbbb9208"
)

func TestClbB1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		clbB1ArmorOracle:      "Armor of Shadows",
		clbB1BlurOracle:       "Blur",
		clbB1CloakOracle:      "Cloak of the Bat",
		clbB1SchematicOracle:  "Nimblewright Schematic",
		clbB1RaphaelOracle:    "Raphael, Fiendish Savior",
		clbB1ShadowheartOracl: "Shadowheart, Dark Justiciar",
		clbB1StoneskinOracle:  "Stoneskin",
		clbB1ThrakkusOracle:   "Thrakkus the Butcher",
		clbB1MyconidOracle:    "Undercellar Myconid",
		clbB1TempleOracle:     "Your Temple Is Under Attack",
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

// clbB1BattlefieldNamed is the one battlefield card with this name.
func clbB1BattlefieldNamed(t *testing.T, g *game.Game, name string) uuid.UUID {
	t.Helper()
	ids := battlefieldIDsNamed(g, name)
	if len(ids) != 1 {
		t.Fatalf("%d permanents named %s on the battlefield, want 1", len(ids), name)
	}
	return ids[0]
}

// --- Armor of Shadows ---------------------------------------------------------

func TestClbB1ArmorOfShadowsPumpsAndProtectsUntilEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	_, opp, _ := blbB1Seats(g)
	theirs := blbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	castCatalogSpell(t, g, "Armor of Shadows", "Instant", clbB1ArmorOracle, cardTarget(theirs))
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, theirs); got != 3 {
		t.Errorf("power %d, want 3: any creature may be the target", got)
	}
	if !hasEffectiveKeyword(t, g, theirs, "indestructible") {
		t.Error("the target should be indestructible")
	}
	endTurn(t, g)
	if got := effectivePower(t, g, theirs); got != 2 {
		t.Errorf("power %d after the turn, want 2", got)
	}
	if hasEffectiveKeyword(t, g, theirs, "indestructible") {
		t.Error("indestructible should end with the turn")
	}
}

// --- Blur ---------------------------------------------------------------------

func TestClbB1BlurFlickersToTheOwnerAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	seedLibrary(me, "A", "B")
	theirs := blbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	if err := castCatalogSpellErr(t, g, "Blur", "Instant", clbB1BlurOracle, cardTarget(theirs)); err == nil {
		t.Fatal("Blur targeted a creature an opponent controls")
	}
	borrowed := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Borrowed Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: me.ID})
	addCountersForTest(g, borrowed, game.CounterPlusOne, 1)
	castCatalogSpell(t, g, "Blur", "Instant", clbB1BlurOracle, cardTarget(borrowed))
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	back, ok := battlefieldCard(g, clbB1BattlefieldNamed(t, g, "Borrowed Bear"))
	if !ok || back.Controller != opp.ID {
		t.Error("the creature should return under its owner's control")
	}
	if back.Counters[game.CounterPlusOne] != 0 {
		t.Error("the returned creature is a new object and keeps no counters")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d: Blur draws a card", got, hand+1)
	}
}

// --- Cloak of the Bat ---------------------------------------------------------

func TestClbB1CloakOfTheBatGivesFlyingAndHaste(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	advanceToMain(t, g)
	cloak := pushCatalogPermanent(g, me.ID, "Cloak of the Bat", "Artifact — Equipment", clbB1CloakOracle, false)
	bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	if hasEffectiveKeyword(t, g, bear, "flying") {
		t.Fatal("the bear flies before it is equipped")
	}
	fillPoolColored(me, "C", 2)
	b16Activate(t, g, me.ID, cloak, 0, game.ActivateAbilityParams{Targets: cardTarget(bear)})
	passPriorityAroundTable(t, g)
	for _, kw := range []string{"flying", "haste"} {
		if !hasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("the equipped creature lacks %s", kw)
		}
	}
	if got := effectivePower(t, g, bear); got != 2 {
		t.Errorf("power %d, want 2: the Cloak gives no pump", got)
	}
}

// --- Nimblewright Schematic ---------------------------------------------------

func TestClbB1NimblewrightSchematicMakesAConstructOnEntryAndDeath(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	castCatalogSpell(t, g, "Nimblewright Schematic", "Artifact", clbB1SchematicOracle, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Construct"); n != 1 {
		t.Fatalf("%d Constructs after it entered, want 1", n)
	}
	destroyNow(t, g, clbB1BattlefieldNamed(t, g, "Nimblewright Schematic"))
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Construct"); n != 2 {
		t.Errorf("%d Constructs after it was destroyed, want 2", n)
	}
}

func TestClbB1NimblewrightSchematicExiledMakesNoConstruct(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	castCatalogSpell(t, g, "Nimblewright Schematic", "Artifact", clbB1SchematicOracle, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	exileCreature(t, g, clbB1BattlefieldNamed(t, g, "Nimblewright Schematic"))
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Construct"); n != 1 {
		t.Errorf("%d Constructs, want 1: exile is not the graveyard", n)
	}
}

// --- Raphael, Fiendish Savior -------------------------------------------------

func TestClbB1RaphaelPumpsOtherFiendsYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	raphael := blbB1Push(g, me.ID, "Raphael, Fiendish Savior", "Legendary Creature — Devil Noble", clbB1RaphaelOracle, 4, 4)
	imp := blbB1Push(g, me.ID, "Imp", "Creature — Imp", "", 1, 1)
	tiefling := blbB1Push(g, me.ID, "Tiefling", "Creature — Tiefling Warlock", "", 2, 2)
	human := blbB1Push(g, me.ID, "Human", "Creature — Human", "", 2, 2)
	theirDemon := blbB1Push(g, opp.ID, "Their Demon", "Creature — Demon", "", 5, 5)
	for id, want := range map[uuid.UUID]int{raphael: 4, imp: 2, tiefling: 3, human: 2, theirDemon: 5} {
		if got := effectivePower(t, g, id); got != want {
			t.Errorf("power %d, want %d", got, want)
		}
	}
	if !hasEffectiveKeyword(t, g, imp, "lifelink") || !hasEffectiveKeyword(t, g, tiefling, "lifelink") {
		t.Error("the Imp and the Tiefling should have lifelink")
	}
	for _, id := range []uuid.UUID{raphael, human, theirDemon} {
		if hasEffectiveKeyword(t, g, id, "lifelink") {
			t.Error("only other fiends you control have lifelink")
		}
	}
}

func TestClbB1RaphaelMakesADevilIfACreatureCardHitYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	blbB1Push(g, me.ID, "Raphael, Fiendish Savior", "Legendary Creature — Devil Noble", clbB1RaphaelOracle, 4, 4)
	victim := blbB1Push(g, me.ID, "Victim", "Creature — Human", "", 2, 2)
	advanceToMain(t, g)
	killCreature(t, g, me.ID, victim)
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Devil"); n != 1 {
		t.Errorf("%d Devils, want 1", n)
	}
}

func TestClbB1RaphaelMakesNoDevilWithoutACreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	raphael := blbB1Push(g, me.ID, "Raphael, Fiendish Savior", "Legendary Creature — Devil Noble", clbB1RaphaelOracle, 4, 4)
	theirs := blbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	advanceToMain(t, g)
	killCreature(t, g, opp.ID, theirs)
	advanceToEndStep(t, g)
	if n := triggersOnStackFrom(g, raphael); n != 0 {
		t.Errorf("%d triggers from Raphael: a card in an opponent's graveyard is not in yours", n)
	}
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Devil"); n != 0 {
		t.Errorf("%d Devils, want 0", n)
	}
}

// --- Shadowheart, Dark Justiciar ----------------------------------------------

func TestClbB1ShadowheartDrawsTheSacrificedPower(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	seedLibrary(me, "A", "B", "C", "D", "E", "F", "G", "H")
	advanceToMain(t, g)
	shadowheart := blbB1Push(g, me.ID, "Shadowheart, Dark Justiciar", "Legendary Creature — Human Elf Cleric", clbB1ShadowheartOracl, 3, 4)
	// A 4/4 with two +1/+1 counters: power 6 as it last existed.
	big := pushCounterCreature(g, me.ID, "Big", game.CounterPlusOne, 2)

	fillPoolColored(me, "B", 2)
	if err := g.ActivateCatalogAbility(me.ID, shadowheart, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{shadowheart}}); err == nil {
		t.Fatal("Shadowheart sacrificed herself: the cost is another creature")
	}
	hand := me.Hand.Size()
	b16Activate(t, g, me.ID, shadowheart, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{big}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Fatal("the creature was not sacrificed")
	}
	if got := me.Hand.Size(); got != hand+6 {
		t.Errorf("hand %d, want %d (the sacrificed creature's power)", got, hand+6)
	}
}

// --- Stoneskin ----------------------------------------------------------------

func TestClbB1StoneskinGivesPlusZeroPlusTen(t *testing.T) {
	g := newCatalogGame(t)
	_, opp, _ := blbB1Seats(g)
	theirs := blbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	castCatalogSpell(t, g, "Stoneskin", "Enchantment — Aura", clbB1StoneskinOracle, cardTarget(theirs))
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, theirs); p != 2 || tough != 12 {
		t.Errorf("%d/%d, want 2/12", p, tough)
	}
	spec, _ := Lookup(clbB1StoneskinOracle)
	if !hasFold(spec.PrintedKeywords, "flash") {
		t.Error("Stoneskin has flash")
	}
}

// --- Thrakkus the Butcher -----------------------------------------------------

func TestClbB1ThrakkusDoublesYourDragonsWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	thrakkus := blbB1Push(g, me.ID, "Thrakkus the Butcher", "Legendary Creature — Dragon Peasant", clbB1ThrakkusOracle, 3, 4)
	whelp := blbB1Push(g, me.ID, "Whelp", "Creature — Dragon", "", 2, 2)
	bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirDragon := blbB1Push(g, opp.ID, "Their Dragon", "Creature — Dragon", "", 3, 3)
	declareAttack(t, g, opp.ID, thrakkus)
	passPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]int{thrakkus: 6, whelp: 4, bear: 2, theirDragon: 3} {
		if got := effectivePower(t, g, id); got != want {
			t.Errorf("power %d, want %d", got, want)
		}
	}
	if !hasEffectiveKeyword(t, g, thrakkus, "trample") {
		t.Error("Thrakkus has trample")
	}
	endTurn(t, g)
	if got := effectivePower(t, g, whelp); got != 2 {
		t.Errorf("the Whelp's power is %d after the turn, want 2", got)
	}
}

// --- Undercellar Myconid ------------------------------------------------------

func TestClbB1UndercellarMyconidMakesASaprolingOnEntryAndDeath(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	myconid := blbGLCastCreature(t, g, "Undercellar Myconid", "Creature — Fungus", clbB1MyconidOracle, 1, 2, nil)
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Saproling"); n != 1 {
		t.Fatalf("%d Saprolings after it entered, want 1", n)
	}
	if ids := battlefieldIDsNamed(g, "Undercellar Myconid"); len(ids) == 1 {
		myconid = ids[0]
	}
	destroyNow(t, g, myconid)
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Saproling"); n != 2 {
		t.Errorf("%d Saprolings after it died, want 2", n)
	}
	spec, _ := Lookup(clbB1MyconidOracle)
	if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{W|U|B|R|G}" || !spec.ManaAbilities[0].Cost.Tap {
		t.Errorf("mana abilities %+v, want one {T}: any color", spec.ManaAbilities)
	}
}

// --- Your Temple Is Under Attack ----------------------------------------------

func TestClbB1TemplePrayForProtectionShieldsYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	mine := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := blbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	castModal(t, g, "Your Temple Is Under Attack", "Instant", clbB1TempleOracle, []int{0}, nil)
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, mine, "indestructible") {
		t.Error("your creature should be indestructible")
	}
	if hasEffectiveKeyword(t, g, theirs, "indestructible") {
		t.Error("an opponent's creature gained indestructible")
	}
}

func TestClbB1TempleStrikeADealDrawsTwoEach(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := blbB1Seats(g)
	seedLibrary(me, "A", "B", "C")
	seedLibrary(opp, "D", "E", "F")
	seedLibrary(other, "G", "H", "I")
	castModal(t, g, "Your Temple Is Under Attack", "Instant", clbB1TempleOracle, []int{1}, []game.TargetRef{playerTarget(opp.ID)})
	mine, theirs, others := me.Hand.Size(), opp.Hand.Size(), other.Hand.Size()
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != mine+2 {
		t.Errorf("my hand %d, want %d", got, mine+2)
	}
	if got := opp.Hand.Size(); got != theirs+2 {
		t.Errorf("the target's hand %d, want %d", got, theirs+2)
	}
	if got := other.Hand.Size(); got != others {
		t.Errorf("a player who was not targeted drew: %d, want %d", got, others)
	}
}
