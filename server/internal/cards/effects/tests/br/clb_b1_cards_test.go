package testsbr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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

// clbB1BattlefieldNamed is the one battlefield card with this name.
func clbB1BattlefieldNamed(t *testing.T, g *game.Game, name string) uuid.UUID {
	t.Helper()
	ids := effectstest.BattlefieldIDsNamed(g, name)
	if len(ids) != 1 {
		t.Fatalf("%d permanents named %s on the battlefield, want 1", len(ids), name)
	}
	return ids[0]
}

// --- Armor of Shadows ---------------------------------------------------------

func TestClbB1ArmorOfShadowsPumpsAndProtectsUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp, _ := effectstest.BlbB1Seats(g)
	theirs := effectstest.BlbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Armor of Shadows", "Instant", clbB1ArmorOracle, effectstest.CardTarget(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, theirs); got != 3 {
		t.Errorf("power %d, want 3: any creature may be the target", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, theirs, "indestructible") {
		t.Error("the target should be indestructible")
	}
	effectstest.EndTurn(t, g)
	if got := effectstest.EffectivePower(t, g, theirs); got != 2 {
		t.Errorf("power %d after the turn, want 2", got)
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "indestructible") {
		t.Error("indestructible should end with the turn")
	}
}

// --- Blur ---------------------------------------------------------------------

func TestClbB1BlurFlickersToTheOwnerAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	effectstest.SeedLibrary(me, "A", "B")
	theirs := effectstest.BlbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Blur", "Instant", clbB1BlurOracle, effectstest.CardTarget(theirs)); err == nil {
		t.Fatal("Blur targeted a creature an opponent controls")
	}
	borrowed := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Borrowed Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: me.ID})
	effectstest.AddCountersForTest(g, borrowed, game.CounterPlusOne, 1)
	effectstest.CastCatalogSpell(t, g, "Blur", "Instant", clbB1BlurOracle, effectstest.CardTarget(borrowed))
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	back, ok := effectstest.BattlefieldCard(g, clbB1BattlefieldNamed(t, g, "Borrowed Bear"))
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
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	effectstest.AdvanceToMain(t, g)
	cloak := effectstest.PushCatalogPermanent(g, me.ID, "Cloak of the Bat", "Artifact — Equipment", clbB1CloakOracle, false)
	bear := effectstest.BlbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	if effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Fatal("the bear flies before it is equipped")
	}
	effectstest.FillPoolColored(me, "C", 2)
	effectstest.B16Activate(t, g, me.ID, cloak, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	for _, kw := range []string{"flying", "haste"} {
		if !effectstest.HasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("the equipped creature lacks %s", kw)
		}
	}
	if got := effectstest.EffectivePower(t, g, bear); got != 2 {
		t.Errorf("power %d, want 2: the Cloak gives no pump", got)
	}
}

// --- Nimblewright Schematic ---------------------------------------------------

func TestClbB1NimblewrightSchematicMakesAConstructOnEntryAndDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	effectstest.CastCatalogSpell(t, g, "Nimblewright Schematic", "Artifact", clbB1SchematicOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Construct"); n != 1 {
		t.Fatalf("%d Constructs after it entered, want 1", n)
	}
	effectstest.DestroyNow(t, g, clbB1BattlefieldNamed(t, g, "Nimblewright Schematic"))
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Construct"); n != 2 {
		t.Errorf("%d Constructs after it was destroyed, want 2", n)
	}
}

func TestClbB1NimblewrightSchematicExiledMakesNoConstruct(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	effectstest.CastCatalogSpell(t, g, "Nimblewright Schematic", "Artifact", clbB1SchematicOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.ExileCreature(t, g, clbB1BattlefieldNamed(t, g, "Nimblewright Schematic"))
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Construct"); n != 1 {
		t.Errorf("%d Constructs, want 1: exile is not the graveyard", n)
	}
}

// --- Raphael, Fiendish Savior -------------------------------------------------

func TestClbB1RaphaelPumpsOtherFiendsYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	raphael := effectstest.BlbB1Push(g, me.ID, "Raphael, Fiendish Savior", "Legendary Creature — Devil Noble", clbB1RaphaelOracle, 4, 4)
	imp := effectstest.BlbB1Push(g, me.ID, "Imp", "Creature — Imp", "", 1, 1)
	tiefling := effectstest.BlbB1Push(g, me.ID, "Tiefling", "Creature — Tiefling Warlock", "", 2, 2)
	human := effectstest.BlbB1Push(g, me.ID, "Human", "Creature — Human", "", 2, 2)
	theirDemon := effectstest.BlbB1Push(g, opp.ID, "Their Demon", "Creature — Demon", "", 5, 5)
	for id, want := range map[uuid.UUID]int{raphael: 4, imp: 2, tiefling: 3, human: 2, theirDemon: 5} {
		if got := effectstest.EffectivePower(t, g, id); got != want {
			t.Errorf("power %d, want %d", got, want)
		}
	}
	if !effectstest.HasEffectiveKeyword(t, g, imp, "lifelink") || !effectstest.HasEffectiveKeyword(t, g, tiefling, "lifelink") {
		t.Error("the Imp and the Tiefling should have lifelink")
	}
	for _, id := range []uuid.UUID{raphael, human, theirDemon} {
		if effectstest.HasEffectiveKeyword(t, g, id, "lifelink") {
			t.Error("only other fiends you control have lifelink")
		}
	}
}

func TestClbB1RaphaelMakesADevilIfACreatureCardHitYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	effectstest.BlbB1Push(g, me.ID, "Raphael, Fiendish Savior", "Legendary Creature — Devil Noble", clbB1RaphaelOracle, 4, 4)
	victim := effectstest.BlbB1Push(g, me.ID, "Victim", "Creature — Human", "", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.KillCreature(t, g, me.ID, victim)
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Devil"); n != 1 {
		t.Errorf("%d Devils, want 1", n)
	}
}

func TestClbB1RaphaelMakesNoDevilWithoutACreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	raphael := effectstest.BlbB1Push(g, me.ID, "Raphael, Fiendish Savior", "Legendary Creature — Devil Noble", clbB1RaphaelOracle, 4, 4)
	theirs := effectstest.BlbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.KillCreature(t, g, opp.ID, theirs)
	effectstest.AdvanceToEndStep(t, g)
	if n := effectstest.TriggersOnStackFrom(g, raphael); n != 0 {
		t.Errorf("%d triggers from Raphael: a card in an opponent's graveyard is not in yours", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Devil"); n != 0 {
		t.Errorf("%d Devils, want 0", n)
	}
}

// --- Shadowheart, Dark Justiciar ----------------------------------------------

func TestClbB1ShadowheartDrawsTheSacrificedPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E", "F", "G", "H")
	effectstest.AdvanceToMain(t, g)
	shadowheart := effectstest.BlbB1Push(g, me.ID, "Shadowheart, Dark Justiciar", "Legendary Creature — Human Elf Cleric", clbB1ShadowheartOracl, 3, 4)
	// A 4/4 with two +1/+1 counters: power 6 as it last existed.
	big := effectstest.PushCounterCreature(g, me.ID, "Big", game.CounterPlusOne, 2)

	effectstest.FillPoolColored(me, "B", 2)
	if err := g.ActivateCatalogAbility(me.ID, shadowheart, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{shadowheart}}); err == nil {
		t.Fatal("Shadowheart sacrificed herself: the cost is another creature")
	}
	hand := me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, shadowheart, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{big}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Fatal("the creature was not sacrificed")
	}
	if got := me.Hand.Size(); got != hand+6 {
		t.Errorf("hand %d, want %d (the sacrificed creature's power)", got, hand+6)
	}
}

// --- Stoneskin ----------------------------------------------------------------

func TestClbB1StoneskinGivesPlusZeroPlusTen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp, _ := effectstest.BlbB1Seats(g)
	theirs := effectstest.BlbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Stoneskin", "Enchantment — Aura", clbB1StoneskinOracle, effectstest.CardTarget(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, theirs); p != 2 || tough != 12 {
		t.Errorf("%d/%d, want 2/12", p, tough)
	}
	spec, _ := effects.Lookup(clbB1StoneskinOracle)
	if !effects.HasFold(spec.PrintedKeywords, "flash") {
		t.Error("Stoneskin has flash")
	}
}

// --- Thrakkus the Butcher -----------------------------------------------------

func TestClbB1ThrakkusDoublesYourDragonsWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	thrakkus := effectstest.BlbB1Push(g, me.ID, "Thrakkus the Butcher", "Legendary Creature — Dragon Peasant", clbB1ThrakkusOracle, 3, 4)
	whelp := effectstest.BlbB1Push(g, me.ID, "Whelp", "Creature — Dragon", "", 2, 2)
	bear := effectstest.BlbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirDragon := effectstest.BlbB1Push(g, opp.ID, "Their Dragon", "Creature — Dragon", "", 3, 3)
	effectstest.DeclareAttack(t, g, opp.ID, thrakkus)
	effectstest.PassPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]int{thrakkus: 6, whelp: 4, bear: 2, theirDragon: 3} {
		if got := effectstest.EffectivePower(t, g, id); got != want {
			t.Errorf("power %d, want %d", got, want)
		}
	}
	if !effectstest.HasEffectiveKeyword(t, g, thrakkus, "trample") {
		t.Error("Thrakkus has trample")
	}
	effectstest.EndTurn(t, g)
	if got := effectstest.EffectivePower(t, g, whelp); got != 2 {
		t.Errorf("the Whelp's power is %d after the turn, want 2", got)
	}
}

// --- Undercellar Myconid ------------------------------------------------------

func TestClbB1UndercellarMyconidMakesASaprolingOnEntryAndDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	myconid := effectstest.BlbGLCastCreature(t, g, "Undercellar Myconid", "Creature — Fungus", clbB1MyconidOracle, 1, 2, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Saproling"); n != 1 {
		t.Fatalf("%d Saprolings after it entered, want 1", n)
	}
	if ids := effectstest.BattlefieldIDsNamed(g, "Undercellar Myconid"); len(ids) == 1 {
		myconid = ids[0]
	}
	effectstest.DestroyNow(t, g, myconid)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Saproling"); n != 2 {
		t.Errorf("%d Saprolings after it died, want 2", n)
	}
	spec, _ := effects.Lookup(clbB1MyconidOracle)
	if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{W|U|B|R|G}" || !spec.ManaAbilities[0].Cost.Tap {
		t.Errorf("mana abilities %+v, want one {T}: any color", spec.ManaAbilities)
	}
}

// --- Your Temple Is Under Attack ----------------------------------------------

func TestClbB1TemplePrayForProtectionShieldsYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	mine := effectstest.BlbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := effectstest.BlbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	effectstest.CastModal(t, g, "Your Temple Is Under Attack", "Instant", clbB1TempleOracle, []int{0}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, mine, "indestructible") {
		t.Error("your creature should be indestructible")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "indestructible") {
		t.Error("an opponent's creature gained indestructible")
	}
}

func TestClbB1TempleStrikeADealDrawsTwoEach(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := effectstest.BlbB1Seats(g)
	effectstest.SeedLibrary(me, "A", "B", "C")
	effectstest.SeedLibrary(opp, "D", "E", "F")
	effectstest.SeedLibrary(other, "G", "H", "I")
	effectstest.CastModal(t, g, "Your Temple Is Under Attack", "Instant", clbB1TempleOracle, []int{1}, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	mine, theirs, others := me.Hand.Size(), opp.Hand.Size(), other.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
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
