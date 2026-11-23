package testsg

import (
	"errors"
	"reflect"
	"sort"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
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

// --- Purphoros, God of the Forge -----------------------------------------

func TestPurphorosGodOfTheForgeIsNotACreatureUnderFiveDevotion(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	p := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Purphoros, God of the Forge",
		TypeLine: "Legendary Enchantment Creature — God", OracleID: effectstest.Gl1bPurphorosOracle,
		ManaCost: "{3}{R}", Power: 6, Toughness: 5, Owner: me.ID, Controller: me.ID,
	})
	if effectstest.ContainsStr(effectstest.EffectiveTypes(t, g, p), "Creature") {
		t.Error("below five devotion Purphoros must not be a creature")
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, p, "indestructible") {
		t.Error("Purphoros is always indestructible")
	}
	for i := 0; i < 2; i++ {
		effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Red Permanent", TypeLine: "Artifact",
			ManaCost: "{R}{R}", Owner: me.ID, Controller: me.ID,
		})
	}
	if !effectstest.ContainsStr(effectstest.EffectiveTypes(t, g, p), "Creature") {
		t.Error("at five devotion Purphoros is a creature")
	}
}

func TestPurphorosGodOfTheForgeDamagesEachOpponentForAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Purphoros, God of the Forge",
		"Legendary Enchantment Creature — God", effectstest.Gl1bPurphorosOracle, false)
	before := effectstest.LifeOfOpponents(g)
	effectstest.EnterFromHand(t, g, me.ID, "Bear", "Creature — Bear", "")
	effectstest.PassPriorityAroundTable(t, g)
	for i, life := range before {
		if got := g.Seats[i+1].Life; got != life-2 {
			t.Errorf("opponent %d life %d -> %d, want -2", i+1, life, got)
		}
	}
	// An opponent's creature entering does nothing.
	before = effectstest.LifeOfOpponents(g)
	effectstest.EnterFromHand(t, g, g.Seats[1].ID, "Their Bear", "Creature — Bear", "")
	effectstest.PassPriorityAroundTable(t, g)
	for i, life := range before {
		if got := g.Seats[i+1].Life; got != life {
			t.Errorf("opponent %d lost life to an opposing creature: %d -> %d", i+1, life, got)
		}
	}
}

func TestPurphorosGodOfTheForgePumpsYourCreaturesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	p := effectstest.PushCatalogPermanent(g, me.ID, "Purphoros, God of the Forge",
		"Legendary Enchantment Creature — God", effectstest.Gl1bPurphorosOracle, false)
	mine := effectstest.PushVanillaCreature(g, me.ID, "Mine", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Theirs", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	gl1bGivePool(me, "R", "R", "R")
	if err := g.ActivateCatalogAbility(me.ID, p, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, mine); got != 3 {
		t.Errorf("my creature power = %d, want 3", got)
	}
	if got := effectstest.EffectivePower(t, g, theirs); got != 2 {
		t.Errorf("their creature power = %d, want 2", got)
	}
}

// --- Crucible of Worlds / Ancient Greenwarden -----------------------------

func TestCrucibleOfWorldsPlaysLandsFromTheGraveyardButNotSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.PushGraveyardCardTyped(me, "Dead Forest", "Basic Land — Forest")
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{FromZone: string(game.ZoneGraveyard)}); err == nil {
		t.Fatal("a graveyard land play was allowed with no Crucible out")
	}
	effectstest.PushCatalogPermanent(g, me.ID, "Crucible of Worlds", "Artifact", effectstest.Gl1bCrucibleOracle, false)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{FromZone: string(game.ZoneGraveyard)}); err != nil {
		t.Fatalf("play the land from the graveyard: %v", err)
	}
	if !g.Battlefield.Contains(land) {
		t.Error("the land never reached the battlefield")
	}
	if got := g.LandsPlayedThisTurnFor(me.ID); got != 1 {
		t.Errorf("land plays this turn = %d, want 1 (the land drop is spent)", got)
	}
	bolt := effectstest.PushGraveyardCardTyped(me, "Some Sorcery", "Sorcery")
	if effectstest.GrantedPermissionOn(g, me.ID, bolt, game.ZoneGraveyard).Granted() {
		t.Error("Crucible must not open graveyard spell casts")
	}
}

func TestAncientGreenwardenPlaysGraveyardLandsAndDoublesLandfall(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Ancient Greenwarden", "Creature — Elemental", effectstest.Gl1bGreenwardenOracle, false)
	effectstest.PushCatalogPermanent(g, me.ID, "Ruin Crab", "Creature — Crab", effectstest.Gl1bRuinCrabOracle, false)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)

	land := effectstest.PushGraveyardCardTyped(me, "Dead Forest", "Basic Land — Forest")
	if !effectstest.GrantedPermissionOn(g, me.ID, land, game.ZoneGraveyard).Granted() {
		t.Fatal("Greenwarden should open the graveyard land play")
	}
	before := opp.Library.Size()
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{FromZone: string(game.ZoneGraveyard)}); err != nil {
		t.Fatalf("play the land from the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// Ruin Crab mills three per land entering; the Greenwarden doubles it.
	if got := before - opp.Library.Size(); got != 6 {
		t.Errorf("Ruin Crab milled %d, want 6 (doubled landfall)", got)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, effectstest.Gl1bNamed(g, me.ID, "Ancient Greenwarden")[0].InstanceID, "reach") {
		t.Error("Greenwarden has reach")
	}
}

// --- Sculpting Steel ---------------------------------------------------------

func TestSculptingSteelEntersAsACopyOfAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	rock := effectstest.S58Rock(g, opp.ID, "Their Rock", "Artifact", "{2}")
	bear := effectstest.PushTribalCreature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	id := effectstest.CastCatalogSpell(t, g, "Sculpting Steel", "Artifact", effectstest.Gl1bSculptingSteelOracle, nil)
	for i := 0; i < 8 && effectstest.CopyPrompt(g) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	pc := effectstest.CopyPrompt(g)
	if pc == nil {
		t.Fatal("no copy prompt")
	}
	if !effectstest.HasID(pc.CopyOptions, rock) || effectstest.HasID(pc.CopyOptions, bear) {
		t.Errorf("CopyOptions = %v, want the artifact and not the creature", pc.CopyOptions)
	}
	effectstest.ResolveWithCopyChoice(t, g, rock)
	c := effectstest.LayeredCard(t, g, id)
	if c.Name != "Their Rock" || !c.IsArtifact() || c.IsEnchantment() {
		t.Errorf("copy is %q %q, want a plain Their Rock artifact", c.Name, c.TypeLine)
	}
}

func TestSculptingSteelDeclinedStaysSculptingSteel(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.S58Rock(g, g.Seats[1].ID, "Their Rock", "Artifact", "{2}")
	id := effectstest.CastCatalogSpell(t, g, "Sculpting Steel", "Artifact", effectstest.Gl1bSculptingSteelOracle, nil)
	effectstest.ResolveWithCopyChoice(t, g, uuid.Nil)
	if c := effectstest.LayeredCard(t, g, id); c.Name != "Sculpting Steel" {
		t.Errorf("declined copy is %q, want Sculpting Steel", c.Name)
	}
}

// --- Drannith Magistrate -----------------------------------------------------

func TestDrannithMagistrateRefusesAnOpponentsGraveyardCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.SeedEnchantment(g, effectstest.Gl1bDrannithOracle, "Drannith Magistrate", "Creature — Human Wizard", opp.ID)
	id, err := effectstest.CastFlashbackFromGraveyard(t, g, effectstest.FaithlessLootingOracle)
	effectstest.AssertCantCast(t, err, "anywhere other than their hands")
	if !g.Seats[0].Graveyard.Contains(id) {
		t.Error("a refused flashback did not leave the card in the graveyard")
	}
}

func TestDrannithMagistrateAllowsHandCastsAndItsControllersOwnZones(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.SeedEnchantment(g, effectstest.Gl1bDrannithOracle, "Drannith Magistrate", "Creature — Human Wizard", opp.ID)
	// A cast from hand is untouched.
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})

	// The Magistrate's own controller is exempt.
	g2 := effectstest.NewCatalogGame(t)
	effectstest.SeedEnchantment(g2, effectstest.Gl1bDrannithOracle, "Drannith Magistrate", "Creature — Human Wizard", g2.Seats[0].ID)
	if _, err := effectstest.CastFlashbackFromGraveyard(t, g2, effectstest.FaithlessLootingOracle); err != nil {
		t.Errorf("the Magistrate's controller was refused a flashback: %v", err)
	}
}

// --- the four lands -------------------------------------------------------------

func TestFerrousLakeFiltersOneManaIntoBlueRed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.SeedPermanentWithOracle(g, me.ID, "Ferrous Lake", "Land", effectstest.Gl1bFerrousLakeOracle)
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err == nil {
		t.Fatal("Ferrous Lake made mana with nothing in the pool to pay {1}")
	}
	if c, _ := effectstest.BattlefieldCard(g, land); c.Tapped {
		t.Error("a refused activation tapped the land")
	}
	gl1bGivePool(me, "G")
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := gl1bSorted(effectstest.PoolColors(me)); !reflect.DeepEqual(got, []string{"R", "U"}) {
		t.Errorf("pool = %v, want {U}{R}", got)
	}
}

func TestOverflowingBasinFiltersOneManaIntoGreenBlue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.SeedPermanentWithOracle(g, me.ID, "Overflowing Basin", "Land", effectstest.Gl1bOverflowingBasin)
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err == nil {
		t.Fatal("Overflowing Basin made mana with nothing in the pool to pay {1}")
	}
	gl1bGivePool(me, "R")
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := gl1bSorted(effectstest.PoolColors(me)); !reflect.DeepEqual(got, []string{"G", "U"}) {
		t.Errorf("pool = %v, want {G}{U}", got)
	}
}

func TestElegantParlorEntersTappedAndSurveils(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "Top Card", "Second")
	id := effectstest.PlayLandFromHand(t, g, "Elegant Parlor", effectstest.Gl1bElegantParlorOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if c, ok := effectstest.BattlefieldCard(g, id); !ok || !c.Tapped {
		t.Errorf("Elegant Parlor should enter tapped (found %v, tapped %v)", ok, c.Tapped)
	}
	c := effectstest.SurveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no surveil prompt")
	}
	if err := g.ResolveSurveil(c.ID, me.ID, c.ScryCards, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	if got := effectstest.GraveyardNames(me); len(got) != 1 || got[0] != "Top Card" {
		t.Errorf("graveyard = %v, want the surveilled Top Card", got)
	}
}

func TestThornspireVergeSecondColourNeedsAMountainOrAForest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	verge := effectstest.SeedPermanentWithOracle(g, me.ID, "Thornspire Verge", "Land", effectstest.Gl1bThornspireVergeOracle)
	effectstest.SeedManaLand(g, me.ID, "Island", "Basic Land — Island", "U")

	if err := g.ActivateManaAbility(me.ID, verge, 1, game.ManaAbilityParams{}); !errors.Is(err, game.ErrConditionNotMet) {
		t.Fatalf("gated {G} with no Mountain or Forest: err = %v, want ErrConditionNotMet", err)
	}
	if err := g.ActivateManaAbility(me.ID, verge, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("unconditional {R}: %v", err)
	}
	if got := effectstest.PoolColors(me); !reflect.DeepEqual(got, []string{"R"}) {
		t.Errorf("pool = %v, want {R}", got)
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	verge2 := effectstest.SeedPermanentWithOracle(g2, me2.ID, "Thornspire Verge", "Land", effectstest.Gl1bThornspireVergeOracle)
	effectstest.SeedManaLand(g2, me2.ID, "Forest", "Basic Land — Forest", "G")
	if err := g2.ActivateManaAbility(me2.ID, verge2, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("gated {G} beside a Forest: %v", err)
	}
	if got := effectstest.PoolColors(me2); !reflect.DeepEqual(got, []string{"G"}) {
		t.Errorf("pool = %v, want {G}", got)
	}
}

// --- Black Market ------------------------------------------------------------

func TestBlackMarketCountsEveryCreatureDeathAndAddsBlackAtYourMain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	market := effectstest.PushCatalogPermanent(g, me.ID, "Black Market", "Enchantment", effectstest.Gl1bBlackMarketOracle, false)
	mine := effectstest.PushVanillaCreature(g, me.ID, "Mine", 1, 1)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Theirs", 1, 1)
	effectstest.KillCreature(t, g, me.ID, mine)
	effectstest.KillCreature(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.BattlefieldCardPtr(t, g, market).Counters["charge"]; got != 2 {
		t.Fatalf("charge counters = %d, want 2 (both creatures counted)", got)
	}
	// A creature that is exiled does not die.
	exiled := effectstest.PushVanillaCreature(g, me.ID, "Exiled", 1, 1)
	effectstest.ExileCreature(t, g, exiled)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.BattlefieldCardPtr(t, g, market).Counters["charge"]; got != 2 {
		t.Errorf("charge counters after an exile = %d, want 2", got)
	}

	// The next of my precombat mains adds {B}{B}.
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	black := 0
	for _, c := range effectstest.PoolColors(me) {
		if c == "B" {
			black++
		}
	}
	if black != 2 {
		t.Errorf("pool has %d black, want 2", black)
	}
}

func TestHarmonicProdigyDoublesAShamansTriggerButNotItsOwnProwess(t *testing.T) {
	// Baseline: one Young Pyromancer, one token per instant.
	base := effectstest.NewCatalogGame(t)
	effectstest.PushCatalogPermanent(base, base.Seats[0].ID, "Young Pyromancer", "Creature — Human Shaman", effectstest.B07YoungPyromancerOracle, false)
	effectstest.CastCatalogSpell(t, base, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: base.Seats[1].ID}})
	effectstest.PassPriorityAroundTable(t, base)
	if got := effectstest.Gl1bPyromancerTokens(base, base.Seats[0].ID); got != 1 {
		t.Fatalf("baseline Elementals = %d, want 1", got)
	}

	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	prodigy := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Harmonic Prodigy", TypeLine: "Creature — Human Wizard",
		OracleID: effectstest.Gl1bProdigyOracle, Power: 1, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	effectstest.PushCatalogPermanent(g, me.ID, "Young Pyromancer", "Creature — Human Shaman", effectstest.B07YoungPyromancerOracle, false)
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.SettleProwess(t, g)
	if got := effectstest.Gl1bPyromancerTokens(g, me.ID); got != 2 {
		t.Errorf("Elementals with Harmonic Prodigy = %d, want 2", got)
	}
	// "Another Wizard": the Prodigy's own prowess trigger fires once.
	if got := effectstest.EffectivePower(t, g, prodigy); got != 2 {
		t.Errorf("Prodigy power after one noncreature spell = %d, want 2 (prowess not doubled)", got)
	}
}

func TestHarmonicProdigyDoesNotDoubleAnOpponentsShaman(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Harmonic Prodigy", "Creature — Human Wizard", effectstest.Gl1bProdigyOracle, false)
	effectstest.PushCatalogPermanent(g, opp.ID, "Young Pyromancer", "Creature — Human Shaman", effectstest.B07YoungPyromancerOracle, false)
	// The opponent's instant triggers THEIR Pyromancer, which the
	// Prodigy's controller does not control.
	g.WithWriteLock(func() { g.Turn.ActiveSeat = 1 })
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.Gl1bPyromancerTokens(g, opp.ID); got != 1 {
		t.Errorf("opponent Elementals = %d, want 1", got)
	}
}

// --- Shared Animosity -----------------------------------------------------------

func TestSharedAnimosityCountsOtherAttackersSharingAType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Shared Animosity", "Enchantment", effectstest.Gl1bSharedAnimosityOracle, false)
	g1 := effectstest.PushTribalCreature(g, me.ID, "Goblin A", "Creature — Goblin Warrior", 2, 2)
	g2 := effectstest.PushTribalCreature(g, me.ID, "Goblin B", "Creature — Goblin Shaman", 2, 2)
	elf := effectstest.PushTribalCreature(g, me.ID, "Elf", "Creature — Elf", 2, 2)

	effectstest.DeclareAttack(t, g, opp.ID, g1, g2, elf)
	effectstest.PassPriorityAroundTable(t, g)

	// Each Goblin shares a type with exactly one other attacker; the
	// Elf shares with none.
	if got := effectstest.EffectivePower(t, g, g1); got != 3 {
		t.Errorf("Goblin A power = %d, want 3", got)
	}
	if got := effectstest.EffectivePower(t, g, g2); got != 3 {
		t.Errorf("Goblin B power = %d, want 3", got)
	}
	if got := effectstest.EffectivePower(t, g, elf); got != 2 {
		t.Errorf("Elf power = %d, want 2", got)
	}
}

func TestSharedAnimosityIgnoresNonAttackersAndLoneAttackers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Shared Animosity", "Enchantment", effectstest.Gl1bSharedAnimosityOracle, false)
	a := effectstest.PushTribalCreature(g, me.ID, "Goblin A", "Creature — Goblin", 2, 2)
	effectstest.PushTribalCreature(g, me.ID, "Goblin Home", "Creature — Goblin", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, a)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, a); got != 2 {
		t.Errorf("a lone attacker's power = %d, want 2 (a Goblin that stayed home does not count)", got)
	}
}

// --- Hellkite Tyrant -------------------------------------------------------------

func TestHellkiteTyrantStealsEveryArtifactTheDamagedPlayerControls(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	tyrant := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Hellkite Tyrant", TypeLine: "Creature — Dragon",
		OracleID: effectstest.Gl1bHellkiteTyrantOracle, Power: 6, Toughness: 5, Owner: me.ID, Controller: me.ID,
	})
	rock := effectstest.S58Rock(g, opp.ID, "Rock", "Artifact", "{2}")
	bot := effectstest.S58Rock(g, opp.ID, "Bot", "Artifact Creature — Construct", "{3}")
	spared := effectstest.S58Rock(g, other.ID, "Other Rock", "Artifact", "{2}")
	land := effectstest.SeedManaLand(g, opp.ID, "Island", "Basic Land — Island", "U")

	effectstest.AttackWith(t, g, opp.ID, tyrant)
	effectstest.PassPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{rock, bot} {
		c, ok := effectstest.BattlefieldCard(g, id)
		if !ok || c.Controller != me.ID {
			t.Errorf("artifact %v not under my control (found %v)", id, ok)
		}
	}
	if c, _ := effectstest.BattlefieldCard(g, spared); c.Controller != other.ID {
		t.Error("an artifact of a player I did not damage was taken")
	}
	if c, _ := effectstest.BattlefieldCard(g, land); c.Controller != opp.ID {
		t.Error("a nonartifact land was taken")
	}
}

func gl1bTyrantWithArtifacts(t *testing.T, n int) *game.Game {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Hellkite Tyrant", TypeLine: "Creature — Dragon",
		OracleID: effectstest.Gl1bHellkiteTyrantOracle, Power: 6, Toughness: 5, Owner: me.ID, Controller: me.ID,
	})
	for i := 0; i < n; i++ {
		effectstest.S58Rock(g, me.ID, "Rock", "Artifact", "{1}")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Loyal Apprentice", TypeLine: "Creature — Human Artificer",
		OracleID: effectstest.Gl1bLoyalApprenticeOracle, Power: 2, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(effectstest.Gl1bNamed(g, me.ID, "Thopter")); got != 0 {
		t.Fatalf("Thopters without a commander = %d, want 0", got)
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	effectstest.PushBattlefieldCardWithTimestamp(g2, game.Card{
		InstanceID: uuid.New(), Name: "Loyal Apprentice", TypeLine: "Creature — Human Artificer",
		OracleID: effectstest.Gl1bLoyalApprenticeOracle, Power: 2, Toughness: 1, Owner: me2.ID, Controller: me2.ID,
	})
	effectstest.PushBattlefieldCardWithTimestamp(g2, game.Card{
		InstanceID: uuid.New(), Name: "My Commander", TypeLine: "Legendary Creature — Bear",
		Power: 2, Toughness: 2, Owner: me2.ID, Controller: me2.ID, IsCommander: true,
	})
	effectstest.AdvanceTo(t, g2, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g2)
	thopters := effectstest.Gl1bNamed(g2, me2.ID, "Thopter")
	if len(thopters) != 1 {
		t.Fatalf("Thopters with a commander = %d, want 1", len(thopters))
	}
	if !effectstest.EffectiveAbilitiesContain(t, g2, thopters[0].InstanceID, "flying") ||
		!effectstest.EffectiveAbilitiesContain(t, g2, thopters[0].InstanceID, "haste") {
		t.Error("the Thopter should have flying and, this turn, haste")
	}
}

// --- Sylvan Safekeeper -----------------------------------------------------------------

func TestSylvanSafekeeperSacrificesALandForShroud(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	keeper := effectstest.PushCatalogPermanent(g, me.ID, "Sylvan Safekeeper", "Creature — Human Wizard", effectstest.Gl1bSafekeeperOracle, false)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	land := effectstest.SeedManaLand(g, me.ID, "Forest", "Basic Land — Forest", "G")
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)

	if err := g.ActivateCatalogAbility(me.ID, keeper, 0, game.ActivateAbilityParams{
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
		SacrificeIDs: []uuid.UUID{land},
	}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Error("the land was not sacrificed")
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "shroud") {
		t.Error("the Bear should have shroud")
	}
}

func TestSylvanSafekeeperCantBePaidWithoutALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	keeper := effectstest.PushCatalogPermanent(g, me.ID, "Sylvan Safekeeper", "Creature — Human Wizard", effectstest.Gl1bSafekeeperOracle, false)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	if err := g.ActivateCatalogAbility(me.ID, keeper, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err == nil {
		t.Error("activated with no land to sacrifice")
	}
}
