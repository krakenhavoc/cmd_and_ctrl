package effects

import (
	"errors"
	"reflect"
	"sort"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// gap_list_1b_test.go — the second slice of tracker #293's "Gap list
// 1/2": one or more tests per card.

const (
	gl1bPurphorosOracle       = "4fdbbec2-e921-4b63-958d-f9ba1e417197"
	gl1bCrucibleOracle        = "33c722cf-b4bf-431f-aefd-ee96241a7fbf"
	gl1bSculptingSteelOracle  = "6c85271c-c711-49b0-a72e-9e576c33714d"
	gl1bDrannithOracle        = "aadd10d0-6dd0-4bdc-8d93-ff08e29a5863"
	gl1bFerrousLakeOracle     = "62c15af0-40e1-407d-b056-7a3d909e3fdb"
	gl1bOverflowingBasin      = "5ac8e01c-b0a7-4855-a122-1cd26b07c4a5"
	gl1bElegantParlorOracle   = "9ea747cf-5d04-4aa7-bdc3-8145860cd1ba"
	gl1bThornspireVergeOracle = "e861bc08-4f0b-4d22-9b85-9d20227fd5b4"
	gl1bBlackMarketOracle     = "21338b71-37f5-4121-9b84-565025ebcd17"
	gl1bGreenwardenOracle     = "3bcf090c-e890-4a9f-a8aa-6079e4ec9947"
	gl1bProdigyOracle         = "2e2ace5b-4018-43af-8e72-ebafec1a7739"
	gl1bSharedAnimosityOracle = "a27445db-33f2-4571-98b5-83206b797484"
	gl1bHellkiteTyrantOracle  = "d9b066ff-9519-415c-ae17-bfea703c9889"
	gl1bLoyalApprenticeOracle = "7f268b15-ac92-4e98-821b-78d15d9285d9"
	gl1bSafekeeperOracle      = "bddf8f4a-3149-4dd6-a9e5-7747e7e45a1c"
	gl1bRuinCrabOracle        = "8afc00d4-a1c6-4329-af2c-a7f58a0c33e7"
)

func gl1bGivePool(p *game.Player, colors ...string) {
	for _, c := range colors {
		p.ManaPool.AddMana(game.ManaToken{Color: c})
	}
}

func gl1bSorted(in []string) []string {
	out := append([]string(nil), in...)
	sort.Strings(out)
	return out
}

func gl1bNamed(g *game.Game, owner uuid.UUID, name string) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && c.Controller == owner {
			out = append(out, c)
		}
	}
	return out
}

// --- Purphoros, God of the Forge -----------------------------------------

func TestPurphorosGodOfTheForgeIsNotACreatureUnderFiveDevotion(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	p := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Purphoros, God of the Forge",
		TypeLine: "Legendary Enchantment Creature — God", OracleID: gl1bPurphorosOracle,
		ManaCost: "{3}{R}", Power: 6, Toughness: 5, Owner: me.ID, Controller: me.ID,
	})
	if containsStr(effectiveTypes(t, g, p), "Creature") {
		t.Error("below five devotion Purphoros must not be a creature")
	}
	if !effectiveAbilitiesContain(t, g, p, "indestructible") {
		t.Error("Purphoros is always indestructible")
	}
	for i := 0; i < 2; i++ {
		pushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Red Permanent", TypeLine: "Artifact",
			ManaCost: "{R}{R}", Owner: me.ID, Controller: me.ID,
		})
	}
	if !containsStr(effectiveTypes(t, g, p), "Creature") {
		t.Error("at five devotion Purphoros is a creature")
	}
}

func TestPurphorosGodOfTheForgeDamagesEachOpponentForAnotherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Purphoros, God of the Forge",
		"Legendary Enchantment Creature — God", gl1bPurphorosOracle, false)
	before := lifeOfOpponents(g)
	enterFromHand(t, g, me.ID, "Bear", "Creature — Bear", "")
	passPriorityAroundTable(t, g)
	for i, life := range before {
		if got := g.Seats[i+1].Life; got != life-2 {
			t.Errorf("opponent %d life %d -> %d, want -2", i+1, life, got)
		}
	}
	// An opponent's creature entering does nothing.
	before = lifeOfOpponents(g)
	enterFromHand(t, g, g.Seats[1].ID, "Their Bear", "Creature — Bear", "")
	passPriorityAroundTable(t, g)
	for i, life := range before {
		if got := g.Seats[i+1].Life; got != life {
			t.Errorf("opponent %d lost life to an opposing creature: %d -> %d", i+1, life, got)
		}
	}
}

func TestPurphorosGodOfTheForgePumpsYourCreaturesOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	p := pushCatalogPermanent(g, me.ID, "Purphoros, God of the Forge",
		"Legendary Enchantment Creature — God", gl1bPurphorosOracle, false)
	mine := pushVanillaCreature(g, me.ID, "Mine", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Theirs", 2, 2)
	advanceTo(t, g, game.StepPrecombatMain)
	gl1bGivePool(me, "R", "R", "R")
	if err := g.ActivateCatalogAbility(me.ID, p, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, mine); got != 3 {
		t.Errorf("my creature power = %d, want 3", got)
	}
	if got := effectivePower(t, g, theirs); got != 2 {
		t.Errorf("their creature power = %d, want 2", got)
	}
}

// --- Crucible of Worlds / Ancient Greenwarden -----------------------------

func TestCrucibleOfWorldsPlaysLandsFromTheGraveyardButNotSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := pushGraveyardCardTyped(me, "Dead Forest", "Basic Land — Forest")
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{FromZone: string(game.ZoneGraveyard)}); err == nil {
		t.Fatal("a graveyard land play was allowed with no Crucible out")
	}
	pushCatalogPermanent(g, me.ID, "Crucible of Worlds", "Artifact", gl1bCrucibleOracle, false)
	advanceTo(t, g, game.StepPrecombatMain)
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{FromZone: string(game.ZoneGraveyard)}); err != nil {
		t.Fatalf("play the land from the graveyard: %v", err)
	}
	if !g.Battlefield.Contains(land) {
		t.Error("the land never reached the battlefield")
	}
	if got := g.LandsPlayedThisTurnFor(me.ID); got != 1 {
		t.Errorf("land plays this turn = %d, want 1 (the land drop is spent)", got)
	}
	bolt := pushGraveyardCardTyped(me, "Some Sorcery", "Sorcery")
	if grantedPermissionOn(g, me.ID, bolt, game.ZoneGraveyard).Granted() {
		t.Error("Crucible must not open graveyard spell casts")
	}
}

func TestAncientGreenwardenPlaysGraveyardLandsAndDoublesLandfall(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Ancient Greenwarden", "Creature — Elemental", gl1bGreenwardenOracle, false)
	pushCatalogPermanent(g, me.ID, "Ruin Crab", "Creature — Crab", gl1bRuinCrabOracle, false)
	advanceTo(t, g, game.StepPrecombatMain)

	land := pushGraveyardCardTyped(me, "Dead Forest", "Basic Land — Forest")
	if !grantedPermissionOn(g, me.ID, land, game.ZoneGraveyard).Granted() {
		t.Fatal("Greenwarden should open the graveyard land play")
	}
	before := opp.Library.Size()
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{FromZone: string(game.ZoneGraveyard)}); err != nil {
		t.Fatalf("play the land from the graveyard: %v", err)
	}
	passPriorityAroundTable(t, g)
	// Ruin Crab mills three per land entering; the Greenwarden doubles it.
	if got := before - opp.Library.Size(); got != 6 {
		t.Errorf("Ruin Crab milled %d, want 6 (doubled landfall)", got)
	}
	if !effectiveAbilitiesContain(t, g, gl1bNamed(g, me.ID, "Ancient Greenwarden")[0].InstanceID, "reach") {
		t.Error("Greenwarden has reach")
	}
}

// --- Sculpting Steel ---------------------------------------------------------

func TestSculptingSteelEntersAsACopyOfAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	rock := s58Rock(g, opp.ID, "Their Rock", "Artifact", "{2}")
	bear := pushTribalCreature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	id := castCatalogSpell(t, g, "Sculpting Steel", "Artifact", gl1bSculptingSteelOracle, nil)
	for i := 0; i < 8 && copyPrompt(g) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	pc := copyPrompt(g)
	if pc == nil {
		t.Fatal("no copy prompt")
	}
	if !hasID(pc.CopyOptions, rock) || hasID(pc.CopyOptions, bear) {
		t.Errorf("CopyOptions = %v, want the artifact and not the creature", pc.CopyOptions)
	}
	resolveWithCopyChoice(t, g, rock)
	c := layeredCard(t, g, id)
	if c.Name != "Their Rock" || !c.IsArtifact() || c.IsEnchantment() {
		t.Errorf("copy is %q %q, want a plain Their Rock artifact", c.Name, c.TypeLine)
	}
}

func TestSculptingSteelDeclinedStaysSculptingSteel(t *testing.T) {
	g := newCatalogGame(t)
	s58Rock(g, g.Seats[1].ID, "Their Rock", "Artifact", "{2}")
	id := castCatalogSpell(t, g, "Sculpting Steel", "Artifact", gl1bSculptingSteelOracle, nil)
	resolveWithCopyChoice(t, g, uuid.Nil)
	if c := layeredCard(t, g, id); c.Name != "Sculpting Steel" {
		t.Errorf("declined copy is %q, want Sculpting Steel", c.Name)
	}
}

// --- Drannith Magistrate -----------------------------------------------------

func TestDrannithMagistrateRefusesAnOpponentsGraveyardCast(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	seedEnchantment(g, gl1bDrannithOracle, "Drannith Magistrate", "Creature — Human Wizard", opp.ID)
	id, err := castFlashbackFromGraveyard(t, g, faithlessLootingOracle)
	assertCantCast(t, err, "anywhere other than their hands")
	if !g.Seats[0].Graveyard.Contains(id) {
		t.Error("a refused flashback did not leave the card in the graveyard")
	}
}

func TestDrannithMagistrateAllowsHandCastsAndItsControllersOwnZones(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	seedEnchantment(g, gl1bDrannithOracle, "Drannith Magistrate", "Creature — Human Wizard", opp.ID)
	// A cast from hand is untouched.
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})

	// The Magistrate's own controller is exempt.
	g2 := newCatalogGame(t)
	seedEnchantment(g2, gl1bDrannithOracle, "Drannith Magistrate", "Creature — Human Wizard", g2.Seats[0].ID)
	if _, err := castFlashbackFromGraveyard(t, g2, faithlessLootingOracle); err != nil {
		t.Errorf("the Magistrate's controller was refused a flashback: %v", err)
	}
}

// --- the four lands -------------------------------------------------------------

func TestFerrousLakeFiltersOneManaIntoBlueRed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := seedPermanentWithOracle(g, me.ID, "Ferrous Lake", "Land", gl1bFerrousLakeOracle)
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err == nil {
		t.Fatal("Ferrous Lake made mana with nothing in the pool to pay {1}")
	}
	if c, _ := battlefieldCard(g, land); c.Tapped {
		t.Error("a refused activation tapped the land")
	}
	gl1bGivePool(me, "G")
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := gl1bSorted(poolColors(me)); !reflect.DeepEqual(got, []string{"R", "U"}) {
		t.Errorf("pool = %v, want {U}{R}", got)
	}
}

func TestOverflowingBasinFiltersOneManaIntoGreenBlue(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := seedPermanentWithOracle(g, me.ID, "Overflowing Basin", "Land", gl1bOverflowingBasin)
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err == nil {
		t.Fatal("Overflowing Basin made mana with nothing in the pool to pay {1}")
	}
	gl1bGivePool(me, "R")
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := gl1bSorted(poolColors(me)); !reflect.DeepEqual(got, []string{"G", "U"}) {
		t.Errorf("pool = %v, want {G}{U}", got)
	}
}

func TestElegantParlorEntersTappedAndSurveils(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "Top Card", "Second")
	id := playLandFromHand(t, g, "Elegant Parlor", gl1bElegantParlorOracle)
	passPriorityAroundTable(t, g)
	if c, ok := battlefieldCard(g, id); !ok || !c.Tapped {
		t.Errorf("Elegant Parlor should enter tapped (found %v, tapped %v)", ok, c.Tapped)
	}
	c := surveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no surveil prompt")
	}
	if err := g.ResolveSurveil(c.ID, me.ID, c.ScryCards, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	if got := graveyardNames(me); len(got) != 1 || got[0] != "Top Card" {
		t.Errorf("graveyard = %v, want the surveilled Top Card", got)
	}
}

func TestThornspireVergeSecondColourNeedsAMountainOrAForest(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	verge := seedPermanentWithOracle(g, me.ID, "Thornspire Verge", "Land", gl1bThornspireVergeOracle)
	seedManaLand(g, me.ID, "Island", "Basic Land — Island", "U")

	if err := g.ActivateManaAbility(me.ID, verge, 1, game.ManaAbilityParams{}); !errors.Is(err, game.ErrConditionNotMet) {
		t.Fatalf("gated {G} with no Mountain or Forest: err = %v, want ErrConditionNotMet", err)
	}
	if err := g.ActivateManaAbility(me.ID, verge, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("unconditional {R}: %v", err)
	}
	if got := poolColors(me); !reflect.DeepEqual(got, []string{"R"}) {
		t.Errorf("pool = %v, want {R}", got)
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	verge2 := seedPermanentWithOracle(g2, me2.ID, "Thornspire Verge", "Land", gl1bThornspireVergeOracle)
	seedManaLand(g2, me2.ID, "Forest", "Basic Land — Forest", "G")
	if err := g2.ActivateManaAbility(me2.ID, verge2, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("gated {G} beside a Forest: %v", err)
	}
	if got := poolColors(me2); !reflect.DeepEqual(got, []string{"G"}) {
		t.Errorf("pool = %v, want {G}", got)
	}
}

// --- Black Market ------------------------------------------------------------

func TestBlackMarketCountsEveryCreatureDeathAndAddsBlackAtYourMain(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	market := pushCatalogPermanent(g, me.ID, "Black Market", "Enchantment", gl1bBlackMarketOracle, false)
	mine := pushVanillaCreature(g, me.ID, "Mine", 1, 1)
	theirs := pushVanillaCreature(g, opp.ID, "Theirs", 1, 1)
	killCreature(t, g, me.ID, mine)
	killCreature(t, g, opp.ID, theirs)
	passPriorityAroundTable(t, g)
	if got := battlefieldCardPtr(t, g, market).Counters["charge"]; got != 2 {
		t.Fatalf("charge counters = %d, want 2 (both creatures counted)", got)
	}
	// A creature that is exiled does not die.
	exiled := pushVanillaCreature(g, me.ID, "Exiled", 1, 1)
	exileCreature(t, g, exiled)
	passPriorityAroundTable(t, g)
	if got := battlefieldCardPtr(t, g, market).Counters["charge"]; got != 2 {
		t.Errorf("charge counters after an exile = %d, want 2", got)
	}

	// The next of my precombat mains adds {B}{B}.
	advanceToPrecombatMainOf(t, g, 0)
	passPriorityAroundTable(t, g)
	black := 0
	for _, c := range poolColors(me) {
		if c == "B" {
			black++
		}
	}
	if black != 2 {
		t.Errorf("pool has %d black, want 2", black)
	}
}

// --- Harmonic Prodigy ------------------------------------------------------------

func gl1bPyromancerTokens(g *game.Game, owner uuid.UUID) int {
	return len(gl1bNamed(g, owner, "Elemental"))
}

func TestHarmonicProdigyDoublesAShamansTriggerButNotItsOwnProwess(t *testing.T) {
	// Baseline: one Young Pyromancer, one token per instant.
	base := newCatalogGame(t)
	pushCatalogPermanent(base, base.Seats[0].ID, "Young Pyromancer", "Creature — Human Shaman", b07YoungPyromancerOracle, false)
	castCatalogSpell(t, base, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: base.Seats[1].ID}})
	passPriorityAroundTable(t, base)
	if got := gl1bPyromancerTokens(base, base.Seats[0].ID); got != 1 {
		t.Fatalf("baseline Elementals = %d, want 1", got)
	}

	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	prodigy := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Harmonic Prodigy", TypeLine: "Creature — Human Wizard",
		OracleID: gl1bProdigyOracle, Power: 1, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	pushCatalogPermanent(g, me.ID, "Young Pyromancer", "Creature — Human Shaman", b07YoungPyromancerOracle, false)
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	settleProwess(t, g)
	if got := gl1bPyromancerTokens(g, me.ID); got != 2 {
		t.Errorf("Elementals with Harmonic Prodigy = %d, want 2", got)
	}
	// "Another Wizard": the Prodigy's own prowess trigger fires once.
	if got := effectivePower(t, g, prodigy); got != 2 {
		t.Errorf("Prodigy power after one noncreature spell = %d, want 2 (prowess not doubled)", got)
	}
}

func TestHarmonicProdigyDoesNotDoubleAnOpponentsShaman(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Harmonic Prodigy", "Creature — Human Wizard", gl1bProdigyOracle, false)
	pushCatalogPermanent(g, opp.ID, "Young Pyromancer", "Creature — Human Shaman", b07YoungPyromancerOracle, false)
	// The opponent's instant triggers THEIR Pyromancer, which the
	// Prodigy's controller does not control.
	g.WithWriteLock(func() { g.Turn.ActiveSeat = 1 })
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	passPriorityAroundTable(t, g)
	if got := gl1bPyromancerTokens(g, opp.ID); got != 1 {
		t.Errorf("opponent Elementals = %d, want 1", got)
	}
}

// --- Shared Animosity -----------------------------------------------------------

func TestSharedAnimosityCountsOtherAttackersSharingAType(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Shared Animosity", "Enchantment", gl1bSharedAnimosityOracle, false)
	g1 := pushTribalCreature(g, me.ID, "Goblin A", "Creature — Goblin Warrior", 2, 2)
	g2 := pushTribalCreature(g, me.ID, "Goblin B", "Creature — Goblin Shaman", 2, 2)
	elf := pushTribalCreature(g, me.ID, "Elf", "Creature — Elf", 2, 2)

	declareAttack(t, g, opp.ID, g1, g2, elf)
	passPriorityAroundTable(t, g)

	// Each Goblin shares a type with exactly one other attacker; the
	// Elf shares with none.
	if got := effectivePower(t, g, g1); got != 3 {
		t.Errorf("Goblin A power = %d, want 3", got)
	}
	if got := effectivePower(t, g, g2); got != 3 {
		t.Errorf("Goblin B power = %d, want 3", got)
	}
	if got := effectivePower(t, g, elf); got != 2 {
		t.Errorf("Elf power = %d, want 2", got)
	}
}

func TestSharedAnimosityIgnoresNonAttackersAndLoneAttackers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Shared Animosity", "Enchantment", gl1bSharedAnimosityOracle, false)
	a := pushTribalCreature(g, me.ID, "Goblin A", "Creature — Goblin", 2, 2)
	pushTribalCreature(g, me.ID, "Goblin Home", "Creature — Goblin", 2, 2)
	declareAttack(t, g, opp.ID, a)
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, a); got != 2 {
		t.Errorf("a lone attacker's power = %d, want 2 (a Goblin that stayed home does not count)", got)
	}
}

// --- Hellkite Tyrant -------------------------------------------------------------

func TestHellkiteTyrantStealsEveryArtifactTheDamagedPlayerControls(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	tyrant := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Hellkite Tyrant", TypeLine: "Creature — Dragon",
		OracleID: gl1bHellkiteTyrantOracle, Power: 6, Toughness: 5, Owner: me.ID, Controller: me.ID,
	})
	rock := s58Rock(g, opp.ID, "Rock", "Artifact", "{2}")
	bot := s58Rock(g, opp.ID, "Bot", "Artifact Creature — Construct", "{3}")
	spared := s58Rock(g, other.ID, "Other Rock", "Artifact", "{2}")
	land := seedManaLand(g, opp.ID, "Island", "Basic Land — Island", "U")

	attackWith(t, g, opp.ID, tyrant)
	passPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{rock, bot} {
		c, ok := battlefieldCard(g, id)
		if !ok || c.Controller != me.ID {
			t.Errorf("artifact %v not under my control (found %v)", id, ok)
		}
	}
	if c, _ := battlefieldCard(g, spared); c.Controller != other.ID {
		t.Error("an artifact of a player I did not damage was taken")
	}
	if c, _ := battlefieldCard(g, land); c.Controller != opp.ID {
		t.Error("a nonartifact land was taken")
	}
}

func gl1bTyrantWithArtifacts(t *testing.T, n int) *game.Game {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[1]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Hellkite Tyrant", TypeLine: "Creature — Dragon",
		OracleID: gl1bHellkiteTyrantOracle, Power: 6, Toughness: 5, Owner: me.ID, Controller: me.ID,
	})
	for i := 0; i < n; i++ {
		s58Rock(g, me.ID, "Rock", "Artifact", "{1}")
	}
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	return g
}

func TestHellkiteTyrantWinsAtUpkeepWithTwentyArtifacts(t *testing.T) {
	g := gl1bTyrantWithArtifacts(t, 20)
	if g.State != game.StateEnded || g.Outcome == nil || g.Outcome.Winner != g.Seats[1].ID {
		t.Fatalf("state %s outcome %+v, want seat 1 to win", g.State, g.Outcome)
	}
}

func TestHellkiteTyrantDoesNotWinWithNineteenArtifacts(t *testing.T) {
	g := gl1bTyrantWithArtifacts(t, 19)
	if g.State != game.StateActive {
		t.Fatalf("19 artifacts ended the game: %+v", g.Outcome)
	}
}

// --- Loyal Apprentice ---------------------------------------------------------------

func TestLoyalApprenticeMakesAHastyThopterOnlyWithYourCommander(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Loyal Apprentice", TypeLine: "Creature — Human Artificer",
		OracleID: gl1bLoyalApprenticeOracle, Power: 2, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if got := len(gl1bNamed(g, me.ID, "Thopter")); got != 0 {
		t.Fatalf("Thopters without a commander = %d, want 0", got)
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	pushBattlefieldCardWithTimestamp(g2, game.Card{
		InstanceID: uuid.New(), Name: "Loyal Apprentice", TypeLine: "Creature — Human Artificer",
		OracleID: gl1bLoyalApprenticeOracle, Power: 2, Toughness: 1, Owner: me2.ID, Controller: me2.ID,
	})
	pushBattlefieldCardWithTimestamp(g2, game.Card{
		InstanceID: uuid.New(), Name: "My Commander", TypeLine: "Legendary Creature — Bear",
		Power: 2, Toughness: 2, Owner: me2.ID, Controller: me2.ID, IsCommander: true,
	})
	advanceTo(t, g2, game.StepBeginCombat)
	passPriorityAroundTable(t, g2)
	thopters := gl1bNamed(g2, me2.ID, "Thopter")
	if len(thopters) != 1 {
		t.Fatalf("Thopters with a commander = %d, want 1", len(thopters))
	}
	if !effectiveAbilitiesContain(t, g2, thopters[0].InstanceID, "flying") ||
		!effectiveAbilitiesContain(t, g2, thopters[0].InstanceID, "haste") {
		t.Error("the Thopter should have flying and, this turn, haste")
	}
}

// --- Sylvan Safekeeper -----------------------------------------------------------------

func TestSylvanSafekeeperSacrificesALandForShroud(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	keeper := pushCatalogPermanent(g, me.ID, "Sylvan Safekeeper", "Creature — Human Wizard", gl1bSafekeeperOracle, false)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	land := seedManaLand(g, me.ID, "Forest", "Basic Land — Forest", "G")
	advanceTo(t, g, game.StepPrecombatMain)

	if err := g.ActivateCatalogAbility(me.ID, keeper, 0, game.ActivateAbilityParams{
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
		SacrificeIDs: []uuid.UUID{land},
	}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Error("the land was not sacrificed")
	}
	if !hasEffectiveKeyword(t, g, bear, "shroud") {
		t.Error("the Bear should have shroud")
	}
}

func TestSylvanSafekeeperCantBePaidWithoutALand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	keeper := pushCatalogPermanent(g, me.ID, "Sylvan Safekeeper", "Creature — Human Wizard", gl1bSafekeeperOracle, false)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	advanceTo(t, g, game.StepPrecombatMain)
	if err := g.ActivateCatalogAbility(me.ID, keeper, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err == nil {
		t.Error("activated with no land to sacrifice")
	}
}
