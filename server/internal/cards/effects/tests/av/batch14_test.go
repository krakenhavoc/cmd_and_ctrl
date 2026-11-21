package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. Four are rows
// in cycle tables (Blooming Marsh and Spirebluff Canal in
// fastlands.go, the two Guildgates in guildgates.go), so a transposed
// row is invisible until someone plays that exact card. Three of the
// issue's 45 were already on main — Pawn of Ulamog and Mazirek from
// the S21 aristocrats pass, Crux of Fate from the S23 boardwipes —
// pinned here so the table matches the issue.
func TestBatch14CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		effectstest.B14BloomingMarshOracle:        "Blooming Marsh",
		effectstest.B14HangarbackWalkerOracle:     "Hangarback Walker",
		effectstest.B14UtvaraHellkiteOracle:       "Utvara Hellkite",
		effectstest.B14EerieUltimatumOracle:       "Eerie Ultimatum",
		effectstest.B14CreepingBloodsuckerOracle:  "Creeping Bloodsucker",
		effectstest.B14HydroidKrasisOracle:        "Hydroid Krasis",
		effectstest.B14SimicGuildgateOracle:       "Simic Guildgate",
		effectstest.B14DragonmasterOutcastOracle:  "Dragonmaster Outcast",
		effectstest.B14TreeOfTalesOracle:          "Tree of Tales",
		effectstest.B14RebuffTheWickedOracle:      "Rebuff the Wicked",
		effectstest.B14GanaxOracle:                "Ganax, Astral Hunter",
		effectstest.B14CorpsejackMenaceOracle:     "Corpsejack Menace",
		effectstest.B14AdventurersInnOracle:       "Adventurer's Inn",
		effectstest.B14KozilekOracle:              "Kozilek, the Great Distortion",
		effectstest.B14ReturnToDustOracle:         "Return to Dust",
		effectstest.B14SpirebluffCanalOracle:      "Spirebluff Canal",
		effectstest.B14DesynchronizationOracle:    "Desynchronization",
		effectstest.B14ArchfiendOfDepravityOracle: "Archfiend of Depravity",
		effectstest.B14RadiantFountainOracle:      "Radiant Fountain",
		effectstest.B14NullElementalBlastOracle:   "Null Elemental Blast",
		effectstest.B14KiorasFollowerOracle:       "Kiora's Follower",
		effectstest.B14SiegeGangCommanderOracle:   "Siege-Gang Commander",
		effectstest.B14ElectrostaticFieldOracle:   "Electrostatic Field",
		effectstest.B14ScavengingOozeOracle:       "Scavenging Ooze",
		effectstest.B14NurturingPeatlandOracle:    "Nurturing Peatland",
		effectstest.B14ResourcefulDefenseOracle:   "Resourceful Defense",
		effectstest.B14ReplicatingRingOracle:      "Replicating Ring",
		effectstest.B14DesertedTempleOracle:       "Deserted Temple",
		effectstest.B14ConsumingAberrationOracle:  "Consuming Aberration",
		effectstest.B14GolgariGuildgateOracle:     "Golgari Guildgate",
		effectstest.B14UnmarkedGraveOracle:        "Unmarked Grave",
		effectstest.B14CloudOracle:                "Cloud, Midgar Mercenary",
		effectstest.B14RhoxFaithmenderOracle:      "Rhox Faithmender",
		effectstest.B14PawnOfUlamogAlreadyOID:     "Pawn of Ulamog",
		effectstest.B14CruxOfFateAlreadyOID:       "Crux of Fate",
		effectstest.B14MazirekAlreadyOID:          "Mazirek, Kraul Death Priest",
	}
	if len(want) != 36 {
		t.Fatalf("the batch registers 36 cards, the table lists %d", len(want))
	}
	for oracle, name := range want {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("oracle %s registered as %q, want %q", oracle, spec.Name, name)
		}
	}
}

// --- lands -----------------------------------------------------------

// The four cycle rows produce their printed pairs — the loop-leak
// canary every land table carries.
func TestB14LandRowsProduceTheirPrintedColours(t *testing.T) {
	for _, tc := range []struct{ oracle, name, produced string }{
		{effectstest.B14BloomingMarshOracle, "Blooming Marsh", "{B|G}"},
		{effectstest.B14SpirebluffCanalOracle, "Spirebluff Canal", "{U|R}"},
		{effectstest.B14SimicGuildgateOracle, "Simic Guildgate", "{G|U}"},
		{effectstest.B14GolgariGuildgateOracle, "Golgari Guildgate", "{B|G}"},
	} {
		spec, ok := effects.Lookup(tc.oracle)
		if !ok {
			t.Fatalf("%s not registered", tc.name)
		}
		if len(spec.Replacements) != 1 {
			t.Errorf("%s: %d replacements, want 1 (enters tapped …)", tc.name, len(spec.Replacements))
		}
		if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != tc.produced {
			t.Errorf("%s: mana abilities %+v, want one producing %s", tc.name, spec.ManaAbilities, tc.produced)
		}
	}
}

func TestB14BloomingMarshEntersUntappedEarly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLandOnBattlefield(g, me.ID, "Swamp", "Basic Land — Swamp")
	effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	id := effectstest.B12PlayFromHand(t, g, "Blooming Marsh", "Land", effectstest.B14BloomingMarshOracle, game.CastSpellParams{})
	if effectstest.B12Card(t, g, id).Tapped {
		t.Error("with two other lands the fastland enters untapped")
	}
	effectstest.SeedLandOnBattlefield(g, me.ID, "Swamp", "Basic Land — Swamp")
	id2 := effectstest.B12PlayFromHand(t, g, "Spirebluff Canal", "Land", effectstest.B14SpirebluffCanalOracle, game.CastSpellParams{})
	if !effectstest.B12Card(t, g, id2).Tapped {
		t.Error("with four other lands the fastland enters tapped")
	}
}

func TestB14GuildgatesEnterTappedAsAReplacement(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.B12PlayFromHand(t, g, "Simic Guildgate", "Land — Gate", effectstest.B14SimicGuildgateOracle, game.CastSpellParams{})
	if !effectstest.B12Card(t, g, id).Tapped {
		t.Fatal("a Guildgate enters tapped")
	}
	if n := effectstest.TapEventsFor(g, id); n != 0 {
		t.Errorf("a tapped entry is a replacement, not a tap: %d tap events", n)
	}
	if !effectstest.B12Card(t, g, id).HasSubtype("Gate") {
		t.Error("the Gate subtype rides the printed type line")
	}
	effectstest.B08Untap(g, id)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if pick := effectstest.RiderLatestManaPick(g, me.ID); pick == nil || len(pick.ColorOptions) != 2 {
		t.Fatalf("must ask G or U, got %+v", pick)
	}
}

func TestB14TreeOfTalesIsAnArtifactThatTapsForGreen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.B12PlayFromHand(t, g, "Tree of Tales", "Artifact Land", effectstest.B14TreeOfTalesOracle, game.CastSpellParams{})
	c := effectstest.B12Card(t, g, id)
	if c.Tapped {
		t.Error("Tree of Tales enters untapped")
	}
	if !c.IsArtifact() || !c.IsLand() {
		t.Errorf("an artifact land, got %q", c.TypeLine)
	}
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("pool %v, want [G]", got)
	}
}

func TestB14RadiantFountainAndAdventurersInnGainTwoOnEntry(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Radiant Fountain", "Land", effectstest.B14RadiantFountainOracle},
		{"Adventurer's Inn", "Land — Town", effectstest.B14AdventurersInnOracle},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		before := me.Life
		id := effectstest.B12PlayFromHand(t, g, tc.name, tc.typeLine, tc.oracle, game.CastSpellParams{})
		if effectstest.TriggerOnStack(g, id) == nil && len(g.PendingTriggers) == 0 {
			t.Fatalf("%s: the life is a trigger with a response window", tc.name)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if me.Life != before+2 {
			t.Errorf("%s: life %d → %d, want +2", tc.name, before, me.Life)
		}
		if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
			t.Fatalf("%s: ActivateManaAbility: %v", tc.name, err)
		}
		if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "C" {
			t.Errorf("%s: pool %v, want [C]", tc.name, got)
		}
	}
}

func TestB14NurturingPeatlandCostsALifeAndCashesInForACard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.SeedPermanentWithOracle(g, me.ID, "Nurturing Peatland", "Land", effectstest.B14NurturingPeatlandOracle)
	before := me.Life
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if me.Life != before-1 {
		t.Errorf("life %d → %d, want -1", before, me.Life)
	}
	if pick := effectstest.RiderLatestManaPick(g, me.ID); pick == nil || len(pick.ColorOptions) != 2 {
		t.Fatalf("must ask B or G, got %+v", pick)
	}
	effectstest.B10ResolveAllManaPicks(t, g, me.ID, "B")

	effectstest.B08Untap(g, land)
	me.ManaPool.EmptyPool()
	effectstest.B10AddMana(me, "C")
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(land) {
		t.Error("the Peatland is sacrificed as a cost, at announce")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want +1", hand, me.Hand.Size())
	}
}

func TestB14DesertedTempleUntapsTargetLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	temple := effectstest.SeedPermanentWithOracle(g, me.ID, "Deserted Temple", "Land", effectstest.B14DesertedTempleOracle)
	coffers := effectstest.SeedLandOnBattlefield(g, me.ID, "Cabal Coffers", "Land")
	theirs := effectstest.SeedLandOnBattlefield(g, opp.ID, "Swamp", "Basic Land — Swamp")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b05TapOnBattlefield(g, coffers)
	b05TapOnBattlefield(g, theirs)
	b05TapOnBattlefield(g, bear)

	if err := g.ActivateManaAbility(me.ID, temple, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool %v, want [C]", got)
	}
	effectstest.B08Untap(g, temple)
	effectstest.B10AddMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, temple, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err == nil {
		t.Fatal("a creature is not a land")
	}
	if err := g.ActivateCatalogAbility(me.ID, temple, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: coffers}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !effectstest.B12Card(t, g, temple).Tapped {
		t.Error("the Temple taps as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12Card(t, g, coffers).Tapped {
		t.Error("the targeted land should be untapped")
	}
	// Anyone's land is a legal target, as printed.
	effectstest.B08Untap(g, temple)
	effectstest.B10AddMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, temple, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}},
	}); err != nil {
		t.Fatalf("activate on an opponent's land: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12Card(t, g, theirs).Tapped {
		t.Error("an opponent's land is a legal target")
	}
}

// --- Utvara Hellkite / Dragonmaster Outcast --------------------------

func TestB14UtvaraHellkiteMakesADragonPerAttackingDragon(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	hellkite := effectstest.B12Push(g, me.ID, "Utvara Hellkite", "Creature — Dragon", effectstest.B14UtvaraHellkiteOracle, 6, 6)
	other := effectstest.B12Creature(g, me.ID, "Whelp", "Creature — Dragon", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{hellkite, other, bear} {
		if err := g.DeclareAttacker(id, victim.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	// #859: the declaration is announced at its lock-in — the
	// priority wrap inside declare_attackers — so the attack triggers
	// exist only after this.
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	n, dragon := effectstest.B14Tokens(g, me.ID, "Dragon")
	if n != 2 {
		t.Fatalf("two attacking Dragons make two tokens, made %d", n)
	}
	if dragon.Power != 6 || dragon.Toughness != 6 || !dragon.HasSubtype("Dragon") || !dragon.HasColor("R") {
		t.Errorf("the token is a 6/6 red Dragon, got %+v", dragon)
	}
	if !effectstest.HasEffectiveKeyword(t, g, dragon.InstanceID, "flying") {
		t.Error("the token has flying")
	}
	if dragon.AttackingTarget != uuid.Nil {
		t.Error("the token is not attacking")
	}
}

func TestB14DragonmasterOutcastNeedsSixLandsAtTriggerAndResolution(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Dragonmaster Outcast", "Creature — Human Shaman", effectstest.B14DragonmasterOutcastOracle, 1, 1)
	var lands []uuid.UUID
	for i := 0; i < 5; i++ {
		lands = append(lands, effectstest.SeedLandOnBattlefield(g, me.ID, "Mountain", "Basic Land — Mountain"))
	}
	effectstest.B12ToMyNextUpkeep(t, g)
	if len(g.PendingTriggers) != 0 || len(g.StackMeta) != 0 {
		t.Fatal("with five lands the ability must not trigger")
	}
	effectstest.PassPriorityAroundTable(t, g)

	lands = append(lands, effectstest.SeedLandOnBattlefield(g, me.ID, "Mountain", "Basic Land — Mountain"))
	effectstest.B12ToMyNextUpkeep(t, g)
	if len(g.PendingTriggers) == 0 && len(g.StackMeta) == 0 {
		t.Fatal("with six lands the ability triggers")
	}
	// In response, a land leaves: the intervening-if fails on resolution.
	effectstest.B14Kill(g, lands[0])
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.B14Tokens(g, me.ID, "Dragon"); n != 0 {
		t.Fatalf("with five lands at resolution no Dragon is made, made %d", n)
	}

	effectstest.SeedLandOnBattlefield(g, me.ID, "Mountain", "Basic Land — Mountain")
	effectstest.B12ToMyNextUpkeep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	n, dragon := effectstest.B14Tokens(g, me.ID, "Dragon")
	if n != 1 {
		t.Fatalf("with six lands a Dragon is made, made %d", n)
	}
	if dragon.Power != 5 || dragon.Toughness != 5 || !effectstest.HasEffectiveKeyword(t, g, dragon.InstanceID, "flying") {
		t.Errorf("the token is a 5/5 flier, got %d/%d", dragon.Power, dragon.Toughness)
	}
}

// --- Creeping Bloodsucker / Electrostatic Field -----------------------

func TestB14CreepingBloodsuckerDrainsEachOpponentOnUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Creeping Bloodsucker", "Creature — Vampire", effectstest.B14CreepingBloodsuckerOracle, 1, 2)
	mine := me.Life
	theirs := effectstest.LifeOfOpponents(g)
	effectstest.B12ToMyNextUpkeep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	for i, before := range theirs {
		if got := effectstest.LifeOfOpponents(g)[i]; got != before-1 {
			t.Errorf("opponent %d: %d → %d, want -1", i, before, got)
		}
	}
	if me.Life != mine+3 {
		t.Errorf("three opponents hit: life %d → %d, want +3", mine, me.Life)
	}
}

func TestB14ElectrostaticFieldPingsOnInstantsAndSorceriesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	field := effectstest.B12Push(g, me.ID, "Electrostatic Field", "Creature — Wall", effectstest.B14ElectrostaticFieldOracle, 0, 4)
	if !effectstest.HasEffectiveKeyword(t, g, field, "defender") {
		t.Error("the Field has defender")
	}
	before := effectstest.LifeOfOpponents(g)
	effectstest.CastCatalogSpell(t, g, "Divination", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	for i, b := range before {
		if got := effectstest.LifeOfOpponents(g)[i]; got != b-1 {
			t.Errorf("opponent %d: %d → %d, want -1 for a sorcery", i, b, got)
		}
	}
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	for i, b := range before {
		if got := effectstest.LifeOfOpponents(g)[i]; got != b-1 {
			t.Errorf("opponent %d: a creature spell must not trigger it (%d)", i, got)
		}
	}
}

// --- Ganax, Astral Hunter --------------------------------------------

func TestB14GanaxMakesATreasureForHimselfAndEachDragon(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Ganax, Astral Hunter", "Legendary Creature — Dragon", effectstest.B14GanaxOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.B14Tokens(g, me.ID, "Treasure"); n != 1 {
		t.Fatalf("Ganax entering makes a Treasure, made %d", n)
	}
	effectstest.CastCatalogSpell(t, g, "Whelp", "Creature — Dragon", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.B14Tokens(g, me.ID, "Treasure"); n != 2 {
		t.Fatalf("another Dragon makes a Treasure, have %d", n)
	}
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.B14Tokens(g, me.ID, "Treasure"); n != 2 {
		t.Fatalf("a Bear is not a Dragon, have %d", n)
	}
}

// --- Corpsejack Menace / Rhox Faithmender ----------------------------

func TestB14CorpsejackMenaceDoublesPlusOneCountersOnYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Corpsejack Menace", "Creature — Fungus", effectstest.B14CorpsejackMenaceOracle, 4, 4)
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")

	if err := g.AddCounter(mine, "+1/+1", 1); err != nil {
		t.Fatal(err)
	}
	if got := effectstest.B12Counter(t, g, mine, "+1/+1"); got != 2 {
		t.Errorf("one counter on your creature becomes %d, want 2", got)
	}
	if err := g.AddCounter(theirs, "+1/+1", 1); err != nil {
		t.Fatal(err)
	}
	if got := effectstest.B12Counter(t, g, theirs, "+1/+1"); got != 1 {
		t.Errorf("an opponent's creature gets %d, want 1", got)
	}
	if err := g.AddCounter(rock, "+1/+1", 1); err != nil {
		t.Fatal(err)
	}
	if got := effectstest.B12Counter(t, g, rock, "+1/+1"); got != 1 {
		t.Errorf("a noncreature gets %d, want 1", got)
	}
	if err := g.AddCounter(mine, "-1/-1", 1); err != nil {
		t.Fatal(err)
	}
	if got := effectstest.B12Counter(t, g, mine, "-1/-1"); got != 1 {
		t.Errorf("a -1/-1 counter is not doubled: %d", got)
	}
}

func TestB14RhoxFaithmenderDoublesYourLifeGainOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rhox := effectstest.B12Push(g, me.ID, "Rhox Faithmender", "Creature — Rhino Monk", effectstest.B14RhoxFaithmenderOracle, 1, 5)
	if !effectstest.HasEffectiveKeyword(t, g, rhox, "lifelink") {
		t.Error("the Faithmender has lifelink")
	}
	mine, theirs := me.Life, opp.Life
	effectstest.B14Gain(t, g, me.ID, 3)
	if me.Life != mine+6 {
		t.Errorf("gain 3 becomes 6: %d → %d", mine, me.Life)
	}
	effectstest.B14Gain(t, g, me.ID, -3)
	if me.Life != mine+3 {
		t.Errorf("a loss is untouched: %d → %d, want %d", mine, me.Life, mine+3)
	}
	effectstest.B14Gain(t, g, opp.ID, 3)
	if opp.Life != theirs+3 {
		t.Errorf("an opponent's gain is untouched: %d → %d", theirs, opp.Life)
	}
}

// --- Rebuff the Wicked / Null Elemental Blast ------------------------

func TestB14RebuffTheWickedCountersOnlySpellsAimedAtYourPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)

	// A Bolt at your face is not a spell that targets a permanent you control.
	face := effectstest.B10OpponentCastsBolt(t, g, opp, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID})
	if effectstest.LegalCards(g, me.ID, effectstest.B14RebuffTheWickedOracle)[face] {
		t.Error("a spell targeting you, not a permanent, is not a legal target")
	}
	effectstest.PassPriorityAroundTable(t, g)

	bolt := effectstest.B10OpponentCastsBolt(t, g, opp, game.TargetRef{Kind: game.TargetCard, ID: bear})
	if !effectstest.LegalCards(g, me.ID, effectstest.B14RebuffTheWickedOracle)[bolt] {
		t.Fatal("a spell targeting your creature is a legal target")
	}
	effectstest.B14CastAt(t, g, me, "Rebuff the Wicked", "Instant", effectstest.B14RebuffTheWickedOracle, nil,
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bolt}}})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("the Bolt was countered; the Bear lives")
	}
	if effectstest.B12ZoneOf(g, bolt) != game.ZoneGraveyard {
		t.Error("the countered Bolt goes to its owner's graveyard")
	}
}

func TestB14NullElementalBlastHitsMulticoloredOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gold := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Gold Bear", TypeLine: "Creature — Bear", Colors: []string{"B", "G"},
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	mono := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Green Bear", TypeLine: "Creature — Bear", Colors: []string{"G"},
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.CastModal(t, g, "Null Elemental Blast", "Instant", effectstest.B14NullElementalBlastOracle, []int{1},
		[]game.TargetRef{{Kind: game.TargetCard, ID: gold}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(gold) {
		t.Error("the multicolored permanent is destroyed")
	}
	if err := g.CastSpell(me.ID, effectstest.B14HandCard(me, "Null Elemental Blast", "Instant", effectstest.B14NullElementalBlastOracle),
		game.CastSpellParams{Modes: []int{1}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mono}}}); err == nil {
		t.Fatal("a mono-coloured permanent is not a legal target")
	}

	// The counter mode: a gold spell on the stack.
	spell := effectstest.B14CastAt(t, g, opp, "Gold Spell", "Instant", "", []string{"U", "R"}, game.CastSpellParams{})
	effectstest.CastModal(t, g, "Null Elemental Blast", "Instant", effectstest.B14NullElementalBlastOracle, []int{0},
		[]game.TargetRef{{Kind: game.TargetCard, ID: spell}})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, spell) != game.ZoneGraveyard || g.Stack.Contains(spell) {
		t.Error("the multicolored spell is countered")
	}
}

// --- Desynchronization / Return to Dust / Unmarked Grave -------------

func TestB14DesynchronizationBouncesEverythingThatIsNotHistoric(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	legend := effectstest.B12Creature(g, opp.ID, "Legend", "Legendary Creature — Human", 3, 3)
	saga := effectstest.B12Permanent(g, opp.ID, "Saga", "Enchantment — Saga")
	aura := effectstest.B12Permanent(g, opp.ID, "Curse", "Enchantment — Aura Curse")
	land := effectstest.SeedLandOnBattlefield(g, opp.ID, "Swamp", "Basic Land — Swamp")

	effectstest.CastCatalogSpell(t, g, "Desynchronization", "Instant", effectstest.B14DesynchronizationOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{bear, theirs, aura} {
		if effectstest.B12ZoneOf(g, id) != game.ZoneHand {
			t.Errorf("%s should be in its owner's hand, is in %s", id, effectstest.B12ZoneOf(g, id))
		}
	}
	for _, id := range []uuid.UUID{rock, legend, saga, land} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s is historic or a land and stays", id)
		}
	}
}

func TestB14ReturnToDustExilesTwoOnlyInYourMainPhase(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Permanent(g, opp.ID, "Rock A", "Artifact")
	b := effectstest.B12Permanent(g, opp.ID, "Shrine B", "Enchantment")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if _, ok := effectstest.LegalCards(g, me.ID, effectstest.B14ReturnToDustOracle)[bear]; ok {
		t.Error("a creature is not an artifact or enchantment")
	}
	effectstest.CastCatalogSpell(t, g, "Return to Dust", "Instant", effectstest.B14ReturnToDustOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, a) != game.ZoneExile || effectstest.B12ZoneOf(g, b) != game.ZoneExile {
		t.Error("cast in your main phase, both targets are exiled")
	}

	// In combat the second pick is refused at announce, and one target
	// is exiled as usual.
	c := effectstest.B12Permanent(g, opp.ID, "Rock C", "Artifact")
	d := effectstest.B12Permanent(g, opp.ID, "Shrine D", "Enchantment")
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	id := effectstest.HandCardFull(me, "Return to Dust", "Instant", "", effectstest.B14ReturnToDustOracle, []string{"W"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: c}, {Kind: game.TargetCard, ID: d}}}); err == nil {
		t.Fatal("outside your main phase a second target is refused")
	}
	effectstest.B14CastAt(t, g, me, "Return to Dust", "Instant", effectstest.B14ReturnToDustOracle, nil,
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: c}}})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, c) != game.ZoneExile {
		t.Error("the first target is always exiled")
	}
	if !g.Battlefield.Contains(d) {
		t.Error("the unpicked permanent stays")
	}
	if spec, _ := effects.Lookup(effectstest.B14ReturnToDustOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("the second slot is gated by the step now")
	}
}

func TestB14UnmarkedGraveTutorsANonlegendaryCardToTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	legend := effectstest.PushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Legend", TypeLine: "Legendary Creature — Human", Owner: me.ID, Controller: me.ID})
	ogre := effectstest.PushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Ogre", TypeLine: "Creature — Ogre", Owner: me.ID, Controller: me.ID})
	effectstest.CastCatalogSpell(t, g, "Unmarked Grave", "Sorcery", effectstest.B14UnmarkedGraveOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the searcher chooses")
	}
	if effectstest.SearchOptionNamed(g, c, "Legend") != uuid.Nil {
		t.Error("a legendary card is not offered")
	}
	if effectstest.SearchOptionNamed(g, c, "Ogre") == uuid.Nil {
		t.Fatal("the nonlegendary card is offered")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Ogre")
	if effectstest.B12ZoneOf(g, ogre) != game.ZoneGraveyard {
		t.Error("the chosen card goes to the graveyard")
	}
	if effectstest.B12ZoneOf(g, legend) != game.ZoneLibrary {
		t.Error("the legend stays in the library")
	}
}

// --- Eerie Ultimatum -------------------------------------------------

func TestB14EerieUltimatumReturnsOnePermanentPerName(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bearA := effectstest.SeedGraveyardCard(t, g, "Bear", "Creature — Bear", "")
	bearB := effectstest.SeedGraveyardCard(t, g, "Bear", "Creature — Bear", "")
	land := effectstest.SeedGraveyardCard(t, g, "Forest", "Basic Land — Forest", "")
	shrine := effectstest.SeedGraveyardCard(t, g, "Shrine", "Enchantment", "")
	bolt := effectstest.SeedGraveyardCard(t, g, "Bolt", "Instant", "")
	// #2524: no target clause — the spell is cast with none, and the
	// choice is made as it resolves.
	if spec, _ := effects.Lookup(effectstest.B14EerieUltimatumOracle); spec.Targets != nil {
		t.Fatal("Eerie Ultimatum does not target")
	}
	effectstest.CastCatalogSpell(t, g, "Eerie Ultimatum", "Sorcery", effectstest.B14EerieUltimatumOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("the caster is asked on resolution")
	}
	if pick.ChooseMin != 0 || pick.ChooseMax != 4 {
		t.Errorf("bounds %d..%d, want 0..4 (any number of the four permanent cards)", pick.ChooseMin, pick.ChooseMax)
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range pick.ChooseCards {
		offered[id] = true
	}
	if offered[bolt] || !offered[bearA] || !offered[bearB] || !offered[land] || !offered[shrine] {
		t.Errorf("offered %v: want the four permanent cards and not the instant", pick.ChooseCards)
	}
	// Two cards of one name break the set rule and are refused.
	err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{bearA, bearB, land})
	if err == nil {
		t.Fatal("two Bears are refused")
	}
	if effectstest.B12ZoneOf(g, bearA) != game.ZoneGraveyard {
		t.Error("the refused answer moves nothing")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, bearA, land, shrine)
	for _, id := range []uuid.UUID{bearA, land, shrine} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s should have returned", id)
		}
	}
	if effectstest.B12ZoneOf(g, bearB) != game.ZoneGraveyard {
		t.Error("the Bear that was not picked stays")
	}
	if effectstest.B12ZoneOf(g, bolt) != game.ZoneGraveyard {
		t.Error("the instant stays")
	}
}

// "Any number" includes none, and an empty graveyard asks nothing.
func TestB14EerieUltimatumAnyNumberIncludesNone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Eerie Ultimatum", "Sorcery", effectstest.B14EerieUltimatumOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("no permanent card in the graveyard: nothing to ask")
	}
	bear := effectstest.SeedGraveyardCard(t, g, "Bear", "Creature — Bear", "")
	effectstest.CastCatalogSpell(t, g, "Eerie Ultimatum", "Sorcery", effectstest.B14EerieUltimatumOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	if effectstest.B12ZoneOf(g, bear) != game.ZoneGraveyard {
		t.Error("choosing nothing returns nothing")
	}
}

// An opponent's response cannot strand a card: exiling one before
// resolution just shrinks the pool the caster is asked about.
func TestB14EerieUltimatumResponseCannotStrandACard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	keep := effectstest.SeedGraveyardCard(t, g, "Bear", "Creature — Bear", "")
	gone := effectstest.SeedGraveyardCard(t, g, "Shrine", "Enchantment", "")
	effectstest.CastCatalogSpell(t, g, "Eerie Ultimatum", "Sorcery", effectstest.B14EerieUltimatumOracle, nil)
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(gone) })
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil || len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != keep {
		t.Fatalf("the pool is read on resolution: %+v", pick)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, keep)
	if !g.Battlefield.Contains(keep) {
		t.Error("the remaining card returns")
	}
}

// --- Archfiend of Depravity ------------------------------------------

func TestB14ArchfiendOfDepravityLeavesEachOpponentTwoCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Archfiend of Depravity", "Creature — Demon", effectstest.B14ArchfiendOfDepravityOracle, 5, 4)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	var theirs []uuid.UUID
	for i := 0; i < 4; i++ {
		theirs = append(theirs, effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2))
	}

	// Your own end step: nothing.
	effectstest.AdvanceToEndStepOf(t, g, 0)
	if len(g.PendingTriggers) != 0 || len(g.StackMeta) != 0 {
		t.Fatal("the Archfiend does not trigger on its controller's end step")
	}
	effectstest.AdvanceToEndStepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ChooseCardsChoiceFor(g, opp.ID)
	if c == nil {
		t.Fatal("the opponent is asked which creatures to keep")
	}
	if len(c.ChooseCards) != 4 || c.ChooseMin != 0 || c.ChooseMax != 2 {
		t.Errorf("the prompt offers their four creatures, up to two: %d cards, %d..%d", len(c.ChooseCards), c.ChooseMin, c.ChooseMax)
	}
	// Three is more than "up to two": the answer is refused, prompt open.
	if err := g.ResolveChooseCards(c.ID, opp.ID, theirs[:3]); err == nil {
		t.Fatal("keeping three creatures was accepted")
	}
	// The player names the two they KEEP, and the other two are sacrificed.
	effectstest.AnswerChooseCards(t, g, opp.ID, theirs[2], theirs[3])
	effectstest.PassPriorityAroundTable(t, g)
	if n := effects.B14CreaturesControlled(g, opp.ID); n != 2 {
		t.Errorf("the opponent keeps two, has %d", n)
	}
	if !g.Battlefield.Contains(theirs[2]) || !g.Battlefield.Contains(theirs[3]) ||
		g.Battlefield.Contains(theirs[0]) || g.Battlefield.Contains(theirs[1]) {
		t.Error("the two chosen creatures stay and the rest are sacrificed")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("the Archfiend's controller is not asked")
	}
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("no prompt for the controller")
	}
}

// --- Kiora's Follower / Scavenging Ooze / Siege-Gang Commander -------

func TestB14KiorasFollowerUntapsAnotherPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	follower := effectstest.B12Push(g, me.ID, "Kiora's Follower", "Creature — Merfolk", effectstest.B14KiorasFollowerOracle, 2, 2)
	land := effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	b05TapOnBattlefield(g, land)
	if effectstest.LegalCards(g, me.ID, effectstest.B14KiorasFollowerOracle)[follower] {
		t.Error("the Follower cannot target itself")
	}
	if err := g.ActivateCatalogAbility(me.ID, follower, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: land}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !effectstest.B12Card(t, g, follower).Tapped {
		t.Error("the Follower taps as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12Card(t, g, land).Tapped {
		t.Error("the land is untapped")
	}
}

func TestB14ScavengingOozeGrowsOnCreatureCardsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ooze := effectstest.B12Push(g, me.ID, "Scavenging Ooze", "Creature — Ooze", effectstest.B14ScavengingOozeOracle, 2, 2)
	creature := effectstest.Batch01GraveyardCard(opp, "Dead Bear", "Creature — Bear")
	spell := effectstest.Batch01GraveyardCard(opp, "Spent Bolt", "Instant")
	before := me.Life

	effectstest.B10AddMana(me, "G")
	if err := g.ActivateCatalogAbility(me.ID, ooze, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: creature}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, creature) != game.ZoneExile {
		t.Error("the targeted card is exiled")
	}
	if got := effectstest.B12Counter(t, g, ooze, "+1/+1"); got != 1 {
		t.Errorf("a creature card grows the Ooze: %d counters, want 1", got)
	}
	if me.Life != before+1 {
		t.Errorf("life %d → %d, want +1", before, me.Life)
	}

	effectstest.B10AddMana(me, "G")
	if err := g.ActivateCatalogAbility(me.ID, ooze, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: spell}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, spell) != game.ZoneExile {
		t.Error("a noncreature card is still exiled")
	}
	if got := effectstest.B12Counter(t, g, ooze, "+1/+1"); got != 1 {
		t.Errorf("a noncreature card does not grow the Ooze: %d counters", got)
	}
	if me.Life != before+1 {
		t.Errorf("no life for a noncreature: %d", me.Life)
	}
}

func TestB14SiegeGangCommanderMakesGoblinsAndThrowsThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	commander := effectstest.CastCatalogSpell(t, g, "Siege-Gang Commander", "Creature — Goblin", effectstest.B14SiegeGangCommanderOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	n, goblin := effectstest.B14Tokens(g, me.ID, "Goblin")
	if n != 3 {
		t.Fatalf("three Goblins, made %d", n)
	}
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B10AddMana(me, "C", "R")
	if err := g.ActivateCatalogAbility(me.ID, commander, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{bear},
		Targets:      []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err == nil {
		t.Fatal("a Bear is not a Goblin")
	}
	before := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, commander, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{goblin.InstanceID},
		Targets:      []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(goblin.InstanceID) {
		t.Error("the Goblin is sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("opponent %d → %d, want -2", before, opp.Life)
	}
}

// --- Hydroid Krasis / Kozilek / Hangarback Walker --------------------

func TestB14HydroidKrasisPaysOutOnCastAndEntersWithXCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	life, hand := me.Life, me.Hand.Size()
	id := effectstest.CastXSpell(t, g, "Hydroid Krasis", "Creature — Jellyfish Hydra Beast", effectstest.B14HydroidKrasisOracle, "{X}{G}{U}", 5, nil)
	if effectstest.TriggerOnStack(g, id) == nil && len(g.PendingTriggers) == 0 {
		t.Fatal("the cast trigger goes on the stack above the spell")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("X=5 gains 2 life: %d → %d", life, me.Life)
	}
	// The hand lost the Krasis and gained two cards.
	if me.Hand.Size() != hand+2 {
		t.Errorf("X=5 draws 2: hand %d → %d", hand, me.Hand.Size())
	}
	if !g.Battlefield.Contains(id) {
		t.Fatal("the Krasis resolved to the battlefield")
	}
	if got := effectstest.B12Counter(t, g, id, "+1/+1"); got != 5 {
		t.Errorf("enters with X counters: %d, want 5", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, id, "flying") || !effectstest.HasEffectiveKeyword(t, g, id, "trample") {
		t.Error("flying and trample")
	}
}

func TestB14HydroidKrasisPaysOutEvenWhenCountered(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life := me.Life
	id := effectstest.CastXSpell(t, g, "Hydroid Krasis", "Creature — Jellyfish Hydra Beast", effectstest.B14HydroidKrasisOracle, "{X}{G}{U}", 4, nil)
	// The trigger is on the stack above the spell; counter the spell.
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
	effectstest.Batch01OpponentCasts(t, g, opp, "Counterspell", effectstest.B14CounterspellOracle, "{U}{U}",
		[]game.TargetRef{{Kind: game.TargetCard, ID: id}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(id) {
		t.Fatal("the Krasis was countered")
	}
	if !me.Graveyard.Contains(id) {
		t.Fatal("the countered Krasis did not reach its owner's graveyard")
	}
	if me.Life != life+2 {
		t.Errorf("the cast trigger still pays out: %d → %d, want +2", life, me.Life)
	}
}

func TestB14KozilekDrawsUpToSevenOnCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	// Down to three in hand; Kozilek is on the stack, not in it.
	for me.Hand.Size() > 3 {
		if _, err := me.Hand.PopTop(); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.CastCatalogSpell(t, g, "Kozilek, the Great Distortion", "Legendary Creature — Eldrazi", effectstest.B14KozilekOracle, nil)
	if me.Hand.Size() != 3 {
		t.Fatalf("hand %d, want 3 with Kozilek on the stack", me.Hand.Size())
	}
	if len(g.PendingTriggers) == 0 && len(g.StackMeta) < 2 {
		t.Fatal("the cast trigger goes on the stack above the spell")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != 7 {
		t.Errorf("draws the difference: hand %d, want 7", me.Hand.Size())
	}
	if n := len(game.ActivatedAbilitiesForCard(game.Card{OracleID: effectstest.B14KozilekOracle})); n != 1 {
		t.Errorf("the discard-to-counter ability is registered; found %d activated abilities", n)
	}

	// With seven or more in hand the trigger never fires.
	effectstest.FillHandTo(t, g, me, 8)
	before := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Kozilek, the Great Distortion", "Legendary Creature — Eldrazi", effectstest.B14KozilekOracle, nil)
	if len(g.PendingTriggers) != 0 || len(g.StackMeta) != 1 {
		t.Fatal("with seven in hand the intervening-if fails and nothing triggers")
	}
	effectstest.PassPriorityAroundTable(t, g)
	// castCatalogSpell seeds the card into the hand before casting, so
	// the count is unchanged when nothing is drawn.
	if me.Hand.Size() != before {
		t.Errorf("hand %d → %d, want no draw", before, me.Hand.Size())
	}
}

func TestB14HangarbackWalkerGrowsAndLeavesThopters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	walker := effectstest.CastXSpell(t, g, "Hangarback Walker", "Artifact Creature — Construct", effectstest.B14HangarbackWalkerOracle, "{X}{X}", 2, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B12Counter(t, g, walker, "+1/+1"); got != 2 {
		t.Fatalf("enters with X counters: %d, want 2", got)
	}
	effectstest.B10AddMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, walker, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("a Walker that entered this turn is summoning sick")
	}
	effectstest.B10Awake(g, walker)
	if err := g.ActivateCatalogAbility(me.ID, walker, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B12Counter(t, g, walker, "+1/+1"); got != 3 {
		t.Errorf("the tap ability adds a counter: %d, want 3", got)
	}
	effectstest.B14Kill(g, walker)
	effectstest.PassPriorityAroundTable(t, g)
	n, thopter := effectstest.B14Tokens(g, me.ID, "Thopter")
	if n != 3 {
		t.Fatalf("three counters leave three Thopters, left %d", n)
	}
	if !thopter.IsArtifact() || !thopter.IsCreature() || !effectstest.HasEffectiveKeyword(t, g, thopter.InstanceID, "flying") {
		t.Errorf("the Thopter is a flying artifact creature, got %q", thopter.TypeLine)
	}
}

// --- Replicating Ring / Consuming Aberration -------------------------

func TestB14ReplicatingRingReplicatesOnTheEighthNightCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ring := effectstest.B12Push(g, me.ID, "Replicating Ring", "Snow Artifact", effectstest.B14ReplicatingRingOracle, 0, 0)
	if err := g.ActivateManaAbility(me.ID, ring, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if pick := effectstest.RiderLatestManaPick(g, me.ID); pick == nil || len(pick.ColorOptions) != 5 {
		t.Fatalf("any colour, not narrowed: %+v", pick)
	}
	effectstest.B10ResolveAllManaPicks(t, g, me.ID, "U")
	me.ManaPool.EmptyPool()

	effectstest.B12ToMyNextUpkeep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B12Counter(t, g, ring, "night"); got != 1 {
		t.Fatalf("one night counter per upkeep: %d", got)
	}
	if err := g.AddCounter(ring, "night", 6); err != nil {
		t.Fatal(err)
	}
	effectstest.B12ToMyNextUpkeep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B12Counter(t, g, ring, "night"); got != 0 {
		t.Errorf("at eight the counters are removed: %d left", got)
	}
	n, token := effectstest.B14Tokens(g, me.ID, "Replicated Ring")
	if n != 8 {
		t.Fatalf("eight Replicated Rings, made %d", n)
	}
	if !token.IsArtifact() || !effectstest.HasCopySupertype(token, "Snow") {
		t.Errorf("the token is a snow artifact, got %q", token.TypeLine)
	}
	if err := g.ActivateManaAbility(me.ID, token.InstanceID, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("a Replicated Ring taps for mana: %v", err)
	}
	if pick := effectstest.RiderLatestManaPick(g, me.ID); pick == nil || len(pick.ColorOptions) != 5 {
		t.Fatalf("the token offers any colour: %+v", pick)
	}
}

func TestB14ConsumingAberrationSizesToOpponentsGraveyardsAndMillsToLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	aberration := effectstest.B12Push(g, me.ID, "Consuming Aberration", "Creature — Horror", effectstest.B14ConsumingAberrationOracle, 0, 0)
	effectstest.PushGraveyardCardForTest(opp, "Dead A")
	effectstest.PushGraveyardCardForTest(opp, "Dead B")
	effectstest.PushGraveyardCardForTest(me, "My Dead")
	if p := effectstest.EffectivePower(t, g, aberration); p != 2 {
		t.Errorf("two cards in opponents' graveyards: power %d, want 2", p)
	}
	// Library tops: opp reveals Spell, Spell, Land (three milled);
	// other reveals Land at once (one milled).
	effectstest.B10LibraryTop(opp, "Their Land", "Basic Land — Swamp", "", 0, 0)
	effectstest.B10LibraryTop(opp, "Their Spell B", "Instant", "{U}", 0, 0)
	effectstest.B10LibraryTop(opp, "Their Spell A", "Sorcery", "{B}", 0, 0)
	effectstest.B10LibraryTop(other, "Other Land", "Basic Land — Island", "", 0, 0)
	oppLib, otherLib := opp.Library.Size(), other.Library.Size()

	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Library.Size() != oppLib-3 {
		t.Errorf("opp milled %d, want 3 (up to and including the land)", oppLib-opp.Library.Size())
	}
	if other.Library.Size() != otherLib-1 {
		t.Errorf("other milled %d, want 1", otherLib-other.Library.Size())
	}
	// Three opponents: 2 + 3 + 1 + (seat 3's mill) cards.
	want := 0
	for _, p := range g.Seats {
		if p.ID != me.ID {
			want += p.Graveyard.Size()
		}
	}
	if p := effectstest.EffectivePower(t, g, aberration); p != want || effectstest.EffectiveToughness(t, g, aberration) != want {
		t.Errorf("power %d / toughness %d, want %d", p, effectstest.EffectiveToughness(t, g, aberration), want)
	}
}

// --- Resourceful Defense ---------------------------------------------

func TestB14ResourcefulDefenseMovesCountersWhenAPermanentLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Resourceful Defense", "Enchantment", effectstest.B14ResourcefulDefenseOracle, 0, 0)
	grown := effectstest.B12Creature(g, me.ID, "Grown Bear", "Creature — Bear", 2, 2)
	plain := effectstest.B12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	keeper := effectstest.B12Creature(g, me.ID, "Keeper", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := g.AddCounter(grown, "+1/+1", 2); err != nil {
		t.Fatal(err)
	}
	if err := g.AddCounter(grown, "charge", 1); err != nil {
		t.Fatal(err)
	}

	// A permanent with no counters leaving does not trigger.
	effectstest.B14Kill(g, plain)
	if len(g.PendingTriggers) != 0 || effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("no counters, no trigger")
	}
	// An opponent's permanent does not trigger.
	if err := g.AddCounter(theirs, "+1/+1", 1); err != nil {
		t.Fatal(err)
	}
	effectstest.B14Kill(g, theirs)
	if len(g.PendingTriggers) != 0 || effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("an opponent's permanent does not trigger it")
	}

	effectstest.B14Kill(g, grown)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, keeper)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B12Counter(t, g, keeper, "+1/+1"); got != 2 {
		t.Errorf("the +1/+1 counters move: %d, want 2", got)
	}
	if got := effectstest.B12Counter(t, g, keeper, "charge"); got != 1 {
		t.Errorf("every kind moves: charge %d, want 1", got)
	}
}

// CR 704.5q: a +1/+1 and a −1/−1 counter on one permanent cancel as a
// state-based action, and the permanent then HAS no counters. A
// creature that dies after the cancel did not "have counters on it",
// so Resourceful Defense does not trigger (#1497, tier 4-final). The
// old event-log walk still saw the +1/+1 placement and triggered.
func TestB14ResourcefulDefenseIgnoresCountersCancelledByCR7045q(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Resourceful Defense", "Enchantment", effectstest.B14ResourcefulDefenseOracle, 0, 0)
	cancelled := effectstest.B12Creature(g, me.ID, "Cancelled Bear", "Creature — Bear", 2, 2)
	keeper := effectstest.B12Creature(g, me.ID, "Keeper", "Creature — Bear", 2, 2)
	if err := g.AddCounter(cancelled, "+1/+1", 1); err != nil {
		t.Fatal(err)
	}
	if err := g.AddCounter(cancelled, "-1/-1", 1); err != nil {
		t.Fatal(err)
	}
	effectstest.RunStateChecksViaDraw(t, g)
	if p, m := effectstest.B12Counter(t, g, cancelled, "+1/+1"), effectstest.B12Counter(t, g, cancelled, "-1/-1"); p != 0 || m != 0 {
		t.Fatalf("the pair cancels (CR 704.5q): +1/+1 %d, -1/-1 %d", p, m)
	}

	effectstest.B14Kill(g, cancelled)
	if len(g.PendingTriggers) != 0 || effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("a creature whose counters cancelled had none as it died: no trigger")
	}
	if got := effectstest.B12Counter(t, g, keeper, "+1/+1"); got != 0 {
		t.Errorf("no +1/+1 counter moves: %d", got)
	}
	if got := effectstest.B12Counter(t, g, keeper, "-1/-1"); got != 0 {
		t.Errorf("no -1/-1 counter moves: %d", got)
	}
}

// A creature with an uneven pair keeps the remainder after CR 704.5q,
// and Resourceful Defense moves exactly that remainder — what the
// permanent had as it left, not what was ever placed on it.
func TestB14ResourcefulDefenseMovesTheCountersLeftAfterCR7045q(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Resourceful Defense", "Enchantment", effectstest.B14ResourcefulDefenseOracle, 0, 0)
	grown := effectstest.B12Creature(g, me.ID, "Grown Bear", "Creature — Bear", 2, 2)
	keeper := effectstest.B12Creature(g, me.ID, "Keeper", "Creature — Bear", 2, 2)
	if err := g.AddCounter(grown, "+1/+1", 3); err != nil {
		t.Fatal(err)
	}
	if err := g.AddCounter(grown, "-1/-1", 1); err != nil {
		t.Fatal(err)
	}
	effectstest.RunStateChecksViaDraw(t, g)
	if p, m := effectstest.B12Counter(t, g, grown, "+1/+1"), effectstest.B12Counter(t, g, grown, "-1/-1"); p != 2 || m != 0 {
		t.Fatalf("CR 704.5q leaves the remainder: +1/+1 %d, -1/-1 %d", p, m)
	}

	effectstest.B14Kill(g, grown)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, keeper)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B12Counter(t, g, keeper, "+1/+1"); got != 2 {
		t.Errorf("the +1/+1 counters left after the cancel move: %d, want 2", got)
	}
	if got := effectstest.B12Counter(t, g, keeper, "-1/-1"); got != 0 {
		t.Errorf("the cancelled -1/-1 counter does not move: %d", got)
	}
}

func TestB14ResourcefulDefenseMovesTheChosenNumberOfCountersForFiveMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	defense := effectstest.B12Push(g, me.ID, "Resourceful Defense", "Enchantment", effectstest.B14ResourcefulDefenseOracle, 0, 0)
	from := effectstest.B12Creature(g, me.ID, "From", "Creature — Bear", 2, 2)
	to := effectstest.B12Creature(g, me.ID, "To", "Creature — Bear", 2, 2)
	if err := g.AddCounter(from, "+1/+1", 3); err != nil {
		t.Fatal(err)
	}
	effectstest.B10AddMana(me, "C", "C", "C", "C", "W")
	if err := g.ActivateCatalogAbility(me.ID, defense, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: from}, {Kind: game.TargetCard, ID: to}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// "Any number" is the player's: asked as the ability resolves, from
	// none to every counter the first target has.
	c := effectstest.PayAmountPrompt(g, me.ID)
	if c == nil {
		t.Fatal("how many counters to move is asked as the ability resolves")
	}
	if pa := c.PayAmount; pa.Min != 0 || pa.Max != 3 || pa.Goal != 3 {
		t.Fatalf("prompt = %+v, want 0 through 3 with all three suggested", *pa)
	}
	if err := g.ResolvePayAmount(c.ID, me.ID, 2); err != nil {
		t.Fatalf("answer 2: %v", err)
	}
	if got := effectstest.B12Counter(t, g, from, "+1/+1"); got != 1 {
		t.Errorf("one counter stays on the source: %d left", got)
	}
	if got := effectstest.B12Counter(t, g, to, "+1/+1"); got != 2 {
		t.Errorf("the destination receives the two chosen: %d, want 2", got)
	}
}

// --- Cloud, Midgar Mercenary -----------------------------------------

func TestB14CloudTutorsAnEquipmentToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sword := effectstest.PushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Sword", TypeLine: "Artifact — Equipment", Owner: me.ID, Controller: me.ID})
	effectstest.PushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Axe", TypeLine: "Artifact — Equipment", Owner: me.ID, Controller: me.ID})
	effectstest.PushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	effectstest.CastCatalogSpell(t, g, "Cloud, Midgar Mercenary", "Legendary Creature — Human Soldier Mercenary", effectstest.B14CloudOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the searcher chooses")
	}
	if effectstest.SearchOptionNamed(g, c, "Rock") != uuid.Nil {
		t.Error("a non-Equipment artifact is not offered")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Sword")
	if effectstest.B12ZoneOf(g, sword) != game.ZoneHand {
		t.Error("the Equipment goes to hand")
	}
	if spec, _ := effects.Lookup(effectstest.B14CloudOracle); len(spec.Static) != 0 || len(spec.Triggered) != 1 {
		t.Error("the trigger-doubling static is declared missing; only the tutor is live")
	}
}
