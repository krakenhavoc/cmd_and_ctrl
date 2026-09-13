package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch14_test.go — card-level coverage for the card-coverage
// roadmap's batch 14 (#307, `edhrec_rank` 1525–1625): the "no new
// machinery" group. One test per observable behaviour, driven
// through a real cast, activation, land play or attack. Helpers from
// the earlier batch test files are reused by name; new ones are
// b14-prefixed.

const (
	b14BloomingMarshOracle        = "66fa2326-1b5d-41fb-b919-83bf9f383577"
	b14HangarbackWalkerOracle     = "dde55256-5259-44e7-a267-fca45a7f0d04"
	b14UtvaraHellkiteOracle       = "05a6a571-643e-429e-8e5c-1c3f8b0dc746"
	b14EerieUltimatumOracle       = "5674f6ae-ed5d-441e-a534-b5dd415165fd"
	b14CreepingBloodsuckerOracle  = "0bdddaf9-579a-4588-b5b3-faa189d4bdcc"
	b14HydroidKrasisOracle        = "6bd872b2-5c40-4e11-9a7f-0136a51b0642"
	b14SimicGuildgateOracle       = "e8705df9-6439-4930-91b6-229f818559af"
	b14DragonmasterOutcastOracle  = "b6fb79c3-cd32-4045-8177-e52841eea65b"
	b14TreeOfTalesOracle          = "8b4aa971-b919-4750-8388-33d4f42c9280"
	b14RebuffTheWickedOracle      = "f5822e53-ddec-4c77-bcf4-091e35c1e731"
	b14GanaxOracle                = "a3112971-3af4-46e8-b2d6-a8759b39d0d1"
	b14CorpsejackMenaceOracle     = "ca0cc02b-b106-4eca-9388-d4b48dd3be49"
	b14AdventurersInnOracle       = "232bd88c-ecdb-43dd-b34a-d381cb3bedf2"
	b14KozilekOracle              = "4c1c1537-e519-4e2f-9bc2-d34b289d4487"
	b14ReturnToDustOracle         = "3029df1d-d02a-4fed-8ab4-000a2096f823"
	b14SpirebluffCanalOracle      = "eb0d8093-5f93-4b25-9384-08f9731bfb28"
	b14DesynchronizationOracle    = "dab64d0f-1246-4a4d-9a79-6db2ca0c8882"
	b14ArchfiendOfDepravityOracle = "af247e2f-b271-4f5b-ab98-4579d2c17c21"
	b14RadiantFountainOracle      = "6db442e5-fbcc-4456-a4c5-bea1aee3fc8e"
	b14NullElementalBlastOracle   = "ec7699e3-ce56-46e4-9e4d-5ed2bd5fca83"
	b14KiorasFollowerOracle       = "22c044ad-77d7-4c93-953d-e2daa9686ff7"
	b14SiegeGangCommanderOracle   = "ddc7f59a-bbb1-4ba1-82c8-6813fd191940"
	b14ElectrostaticFieldOracle   = "2fa94a07-c932-4f85-b6e0-a97d2b29eb52"
	b14ScavengingOozeOracle       = "1ff25f67-36a7-4cfa-a2b1-2135b5b6fb67"
	b14NurturingPeatlandOracle    = "8ed932ff-986c-4592-ad70-53b3fac80d69"
	b14ResourcefulDefenseOracle   = "83e78565-e61f-4bbc-b834-f47941f7e3ec"
	b14ReplicatingRingOracle      = "1ff00f5b-4bf9-4724-8bb5-6b9a9eb0ec7f"
	b14DesertedTempleOracle       = "9f12bf9a-6e1a-4377-b4af-e8cabd3ee58a"
	b14ConsumingAberrationOracle  = "9b55fb72-237d-4935-b645-8ebc6eb4140e"
	b14GolgariGuildgateOracle     = "fa2da325-6859-45bb-b185-35526b01bcc1"
	b14UnmarkedGraveOracle        = "134dfb35-0c32-4143-aa63-7701f406b59e"
	b14CloudOracle                = "33d2584b-bf29-4c22-bd45-14ba2fb98c0e"
	b14RhoxFaithmenderOracle      = "2abe9303-d498-4aad-b6b2-8b5064bd2ffd"
	b14PawnOfUlamogAlreadyOID     = "9bcaf141-1f1f-491f-aced-13dc093b9e2c"
	b14CruxOfFateAlreadyOID       = "52a0dae4-2a95-487e-acd4-eabdb2d031e2"
	b14MazirekAlreadyOID          = "e0420f2c-d578-421e-ae75-e7dc5f70661a"
	b14CounterspellOracle         = "cc187110-1148-4090-bbb8-e205694a39f5"
)

// b14CastAt casts a card from `caster`'s hand at the current step,
// without walking the cursor — for an instant cast outside a main
// phase (Return to Dust's clause) or by a non-active player.
func b14CastAt(t *testing.T, g *game.Game, caster *game.Player, name, typeLine, oracle string, colors []string, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	id := uuid.New()
	caster.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, Colors: colors,
		Owner: caster.ID, Controller: caster.ID,
	})
	if err := g.CastSpell(caster.ID, id, params); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// b14Tokens counts the battlefield tokens with a name under a
// controller and returns the first one.
func b14Tokens(g *game.Game, controller uuid.UUID, name string) (int, game.Card) {
	n := 0
	var first game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && c.Controller == controller && IsToken(c) {
			if n == 0 {
				first = c
			}
			n++
		}
	}
	return n, first
}

// b14Kill destroys a battlefield permanent through the engine's own
// destroy path, so its LKI and dies-triggers are real.
func b14Kill(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id) })
}

// b14Gain changes a player's life through the public mutator, so a
// replacement watching life sees it.
func b14Gain(t *testing.T, g *game.Game, player uuid.UUID, delta int) {
	t.Helper()
	if _, err := g.ChangePlayerLife(player, delta); err != nil {
		t.Fatalf("ChangePlayerLife: %v", err)
	}
}

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
		b14BloomingMarshOracle:        "Blooming Marsh",
		b14HangarbackWalkerOracle:     "Hangarback Walker",
		b14UtvaraHellkiteOracle:       "Utvara Hellkite",
		b14EerieUltimatumOracle:       "Eerie Ultimatum",
		b14CreepingBloodsuckerOracle:  "Creeping Bloodsucker",
		b14HydroidKrasisOracle:        "Hydroid Krasis",
		b14SimicGuildgateOracle:       "Simic Guildgate",
		b14DragonmasterOutcastOracle:  "Dragonmaster Outcast",
		b14TreeOfTalesOracle:          "Tree of Tales",
		b14RebuffTheWickedOracle:      "Rebuff the Wicked",
		b14GanaxOracle:                "Ganax, Astral Hunter",
		b14CorpsejackMenaceOracle:     "Corpsejack Menace",
		b14AdventurersInnOracle:       "Adventurer's Inn",
		b14KozilekOracle:              "Kozilek, the Great Distortion",
		b14ReturnToDustOracle:         "Return to Dust",
		b14SpirebluffCanalOracle:      "Spirebluff Canal",
		b14DesynchronizationOracle:    "Desynchronization",
		b14ArchfiendOfDepravityOracle: "Archfiend of Depravity",
		b14RadiantFountainOracle:      "Radiant Fountain",
		b14NullElementalBlastOracle:   "Null Elemental Blast",
		b14KiorasFollowerOracle:       "Kiora's Follower",
		b14SiegeGangCommanderOracle:   "Siege-Gang Commander",
		b14ElectrostaticFieldOracle:   "Electrostatic Field",
		b14ScavengingOozeOracle:       "Scavenging Ooze",
		b14NurturingPeatlandOracle:    "Nurturing Peatland",
		b14ResourcefulDefenseOracle:   "Resourceful Defense",
		b14ReplicatingRingOracle:      "Replicating Ring",
		b14DesertedTempleOracle:       "Deserted Temple",
		b14ConsumingAberrationOracle:  "Consuming Aberration",
		b14GolgariGuildgateOracle:     "Golgari Guildgate",
		b14UnmarkedGraveOracle:        "Unmarked Grave",
		b14CloudOracle:                "Cloud, Midgar Mercenary",
		b14RhoxFaithmenderOracle:      "Rhox Faithmender",
		b14PawnOfUlamogAlreadyOID:     "Pawn of Ulamog",
		b14CruxOfFateAlreadyOID:       "Crux of Fate",
		b14MazirekAlreadyOID:          "Mazirek, Kraul Death Priest",
	}
	if len(want) != 36 {
		t.Fatalf("the batch registers 36 cards, the table lists %d", len(want))
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
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
		{b14BloomingMarshOracle, "Blooming Marsh", "{B|G}"},
		{b14SpirebluffCanalOracle, "Spirebluff Canal", "{U|R}"},
		{b14SimicGuildgateOracle, "Simic Guildgate", "{G|U}"},
		{b14GolgariGuildgateOracle, "Golgari Guildgate", "{B|G}"},
	} {
		spec, ok := Lookup(tc.oracle)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLandOnBattlefield(g, me.ID, "Swamp", "Basic Land — Swamp")
	seedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	id := b12PlayFromHand(t, g, "Blooming Marsh", "Land", b14BloomingMarshOracle, game.CastSpellParams{})
	if b12Card(t, g, id).Tapped {
		t.Error("with two other lands the fastland enters untapped")
	}
	seedLandOnBattlefield(g, me.ID, "Swamp", "Basic Land — Swamp")
	id2 := b12PlayFromHand(t, g, "Spirebluff Canal", "Land", b14SpirebluffCanalOracle, game.CastSpellParams{})
	if !b12Card(t, g, id2).Tapped {
		t.Error("with four other lands the fastland enters tapped")
	}
}

func TestB14GuildgatesEnterTappedAsAReplacement(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := b12PlayFromHand(t, g, "Simic Guildgate", "Land — Gate", b14SimicGuildgateOracle, game.CastSpellParams{})
	if !b12Card(t, g, id).Tapped {
		t.Fatal("a Guildgate enters tapped")
	}
	if n := tapEventsFor(g, id); n != 0 {
		t.Errorf("a tapped entry is a replacement, not a tap: %d tap events", n)
	}
	if !b12Card(t, g, id).HasSubtype("Gate") {
		t.Error("the Gate subtype rides the printed type line")
	}
	b08Untap(g, id)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if pick := riderLatestManaPick(g, me.ID); pick == nil || len(pick.ColorOptions) != 2 {
		t.Fatalf("must ask G or U, got %+v", pick)
	}
}

func TestB14TreeOfTalesIsAnArtifactThatTapsForGreen(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := b12PlayFromHand(t, g, "Tree of Tales", "Artifact Land", b14TreeOfTalesOracle, game.CastSpellParams{})
	c := b12Card(t, g, id)
	if c.Tapped {
		t.Error("Tree of Tales enters untapped")
	}
	if !c.IsArtifact() || !c.IsLand() {
		t.Errorf("an artifact land, got %q", c.TypeLine)
	}
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("pool %v, want [G]", got)
	}
}

func TestB14RadiantFountainAndAdventurersInnGainTwoOnEntry(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Radiant Fountain", "Land", b14RadiantFountainOracle},
		{"Adventurer's Inn", "Land — Town", b14AdventurersInnOracle},
	} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		before := me.Life
		id := b12PlayFromHand(t, g, tc.name, tc.typeLine, tc.oracle, game.CastSpellParams{})
		if triggerOnStack(g, id) == nil && len(g.PendingTriggers) == 0 {
			t.Fatalf("%s: the life is a trigger with a response window", tc.name)
		}
		passPriorityAroundTable(t, g)
		if me.Life != before+2 {
			t.Errorf("%s: life %d → %d, want +2", tc.name, before, me.Life)
		}
		if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
			t.Fatalf("%s: ActivateManaAbility: %v", tc.name, err)
		}
		if got := batch01PoolColors(me); len(got) != 1 || got[0] != "C" {
			t.Errorf("%s: pool %v, want [C]", tc.name, got)
		}
	}
}

func TestB14NurturingPeatlandCostsALifeAndCashesInForACard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := seedPermanentWithOracle(g, me.ID, "Nurturing Peatland", "Land", b14NurturingPeatlandOracle)
	before := me.Life
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if me.Life != before-1 {
		t.Errorf("life %d → %d, want -1", before, me.Life)
	}
	if pick := riderLatestManaPick(g, me.ID); pick == nil || len(pick.ColorOptions) != 2 {
		t.Fatalf("must ask B or G, got %+v", pick)
	}
	b10ResolveAllManaPicks(t, g, me.ID, "B")

	b08Untap(g, land)
	me.ManaPool.EmptyPool()
	b10AddMana(me, "C")
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(land) {
		t.Error("the Peatland is sacrificed as a cost, at announce")
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want +1", hand, me.Hand.Size())
	}
}

func TestB14DesertedTempleUntapsTargetLand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	temple := seedPermanentWithOracle(g, me.ID, "Deserted Temple", "Land", b14DesertedTempleOracle)
	coffers := seedLandOnBattlefield(g, me.ID, "Cabal Coffers", "Land")
	theirs := seedLandOnBattlefield(g, opp.ID, "Swamp", "Basic Land — Swamp")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b05TapOnBattlefield(g, coffers)
	b05TapOnBattlefield(g, theirs)
	b05TapOnBattlefield(g, bear)

	if err := g.ActivateManaAbility(me.ID, temple, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool %v, want [C]", got)
	}
	b08Untap(g, temple)
	b10AddMana(me, "C")
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
	if !b12Card(t, g, temple).Tapped {
		t.Error("the Temple taps as a cost")
	}
	passPriorityAroundTable(t, g)
	if b12Card(t, g, coffers).Tapped {
		t.Error("the targeted land should be untapped")
	}
	// Anyone's land is a legal target, as printed.
	b08Untap(g, temple)
	b10AddMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, temple, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}},
	}); err != nil {
		t.Fatalf("activate on an opponent's land: %v", err)
	}
	passPriorityAroundTable(t, g)
	if b12Card(t, g, theirs).Tapped {
		t.Error("an opponent's land is a legal target")
	}
}

// --- Utvara Hellkite / Dragonmaster Outcast --------------------------

func TestB14UtvaraHellkiteMakesADragonPerAttackingDragon(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	hellkite := b12Push(g, me.ID, "Utvara Hellkite", "Creature — Dragon", b14UtvaraHellkiteOracle, 6, 6)
	other := b12Creature(g, me.ID, "Whelp", "Creature — Dragon", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{hellkite, other, bear} {
		if err := g.DeclareAttacker(id, victim.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	n, dragon := b14Tokens(g, me.ID, "Dragon")
	if n != 2 {
		t.Fatalf("two attacking Dragons make two tokens, made %d", n)
	}
	if dragon.Power != 6 || dragon.Toughness != 6 || !dragon.HasSubtype("Dragon") || !dragon.HasColor("R") {
		t.Errorf("the token is a 6/6 red Dragon, got %+v", dragon)
	}
	if !hasEffectiveKeyword(t, g, dragon.InstanceID, "flying") {
		t.Error("the token has flying")
	}
	if dragon.AttackingTarget != uuid.Nil {
		t.Error("the token is not attacking")
	}
}

func TestB14DragonmasterOutcastNeedsSixLandsAtTriggerAndResolution(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Dragonmaster Outcast", "Creature — Human Shaman", b14DragonmasterOutcastOracle, 1, 1)
	var lands []uuid.UUID
	for i := 0; i < 5; i++ {
		lands = append(lands, seedLandOnBattlefield(g, me.ID, "Mountain", "Basic Land — Mountain"))
	}
	b12ToMyNextUpkeep(t, g)
	if len(g.PendingTriggers) != 0 || len(g.StackMeta) != 0 {
		t.Fatal("with five lands the ability must not trigger")
	}
	passPriorityAroundTable(t, g)

	lands = append(lands, seedLandOnBattlefield(g, me.ID, "Mountain", "Basic Land — Mountain"))
	b12ToMyNextUpkeep(t, g)
	if len(g.PendingTriggers) == 0 && len(g.StackMeta) == 0 {
		t.Fatal("with six lands the ability triggers")
	}
	// In response, a land leaves: the intervening-if fails on resolution.
	b14Kill(g, lands[0])
	passPriorityAroundTable(t, g)
	if n, _ := b14Tokens(g, me.ID, "Dragon"); n != 0 {
		t.Fatalf("with five lands at resolution no Dragon is made, made %d", n)
	}

	seedLandOnBattlefield(g, me.ID, "Mountain", "Basic Land — Mountain")
	b12ToMyNextUpkeep(t, g)
	passPriorityAroundTable(t, g)
	n, dragon := b14Tokens(g, me.ID, "Dragon")
	if n != 1 {
		t.Fatalf("with six lands a Dragon is made, made %d", n)
	}
	if dragon.Power != 5 || dragon.Toughness != 5 || !hasEffectiveKeyword(t, g, dragon.InstanceID, "flying") {
		t.Errorf("the token is a 5/5 flier, got %d/%d", dragon.Power, dragon.Toughness)
	}
}

// --- Creeping Bloodsucker / Electrostatic Field -----------------------

func TestB14CreepingBloodsuckerDrainsEachOpponentOnUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Creeping Bloodsucker", "Creature — Vampire", b14CreepingBloodsuckerOracle, 1, 2)
	mine := me.Life
	theirs := lifeOfOpponents(g)
	b12ToMyNextUpkeep(t, g)
	passPriorityAroundTable(t, g)
	for i, before := range theirs {
		if got := lifeOfOpponents(g)[i]; got != before-1 {
			t.Errorf("opponent %d: %d → %d, want -1", i, before, got)
		}
	}
	if me.Life != mine+3 {
		t.Errorf("three opponents hit: life %d → %d, want +3", mine, me.Life)
	}
}

func TestB14ElectrostaticFieldPingsOnInstantsAndSorceriesOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	field := b12Push(g, me.ID, "Electrostatic Field", "Creature — Wall", b14ElectrostaticFieldOracle, 0, 4)
	if !hasEffectiveKeyword(t, g, field, "defender") {
		t.Error("the Field has defender")
	}
	before := lifeOfOpponents(g)
	castCatalogSpell(t, g, "Divination", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	for i, b := range before {
		if got := lifeOfOpponents(g)[i]; got != b-1 {
			t.Errorf("opponent %d: %d → %d, want -1 for a sorcery", i, b, got)
		}
	}
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	for i, b := range before {
		if got := lifeOfOpponents(g)[i]; got != b-1 {
			t.Errorf("opponent %d: a creature spell must not trigger it (%d)", i, got)
		}
	}
}

// --- Ganax, Astral Hunter --------------------------------------------

func TestB14GanaxMakesATreasureForHimselfAndEachDragon(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Ganax, Astral Hunter", "Legendary Creature — Dragon", b14GanaxOracle, nil)
	passPriorityAroundTable(t, g)
	if n, _ := b14Tokens(g, me.ID, "Treasure"); n != 1 {
		t.Fatalf("Ganax entering makes a Treasure, made %d", n)
	}
	castCatalogSpell(t, g, "Whelp", "Creature — Dragon", "", nil)
	passPriorityAroundTable(t, g)
	if n, _ := b14Tokens(g, me.ID, "Treasure"); n != 2 {
		t.Fatalf("another Dragon makes a Treasure, have %d", n)
	}
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if n, _ := b14Tokens(g, me.ID, "Treasure"); n != 2 {
		t.Fatalf("a Bear is not a Dragon, have %d", n)
	}
}

// --- Corpsejack Menace / Rhox Faithmender ----------------------------

func TestB14CorpsejackMenaceDoublesPlusOneCountersOnYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Corpsejack Menace", "Creature — Fungus", b14CorpsejackMenaceOracle, 4, 4)
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")

	if err := g.AddCounter(mine, "+1/+1", 1); err != nil {
		t.Fatal(err)
	}
	if got := b12Counter(t, g, mine, "+1/+1"); got != 2 {
		t.Errorf("one counter on your creature becomes %d, want 2", got)
	}
	if err := g.AddCounter(theirs, "+1/+1", 1); err != nil {
		t.Fatal(err)
	}
	if got := b12Counter(t, g, theirs, "+1/+1"); got != 1 {
		t.Errorf("an opponent's creature gets %d, want 1", got)
	}
	if err := g.AddCounter(rock, "+1/+1", 1); err != nil {
		t.Fatal(err)
	}
	if got := b12Counter(t, g, rock, "+1/+1"); got != 1 {
		t.Errorf("a noncreature gets %d, want 1", got)
	}
	if err := g.AddCounter(mine, "-1/-1", 1); err != nil {
		t.Fatal(err)
	}
	if got := b12Counter(t, g, mine, "-1/-1"); got != 1 {
		t.Errorf("a -1/-1 counter is not doubled: %d", got)
	}
}

func TestB14RhoxFaithmenderDoublesYourLifeGainOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rhox := b12Push(g, me.ID, "Rhox Faithmender", "Creature — Rhino Monk", b14RhoxFaithmenderOracle, 1, 5)
	if !hasEffectiveKeyword(t, g, rhox, "lifelink") {
		t.Error("the Faithmender has lifelink")
	}
	mine, theirs := me.Life, opp.Life
	b14Gain(t, g, me.ID, 3)
	if me.Life != mine+6 {
		t.Errorf("gain 3 becomes 6: %d → %d", mine, me.Life)
	}
	b14Gain(t, g, me.ID, -3)
	if me.Life != mine+3 {
		t.Errorf("a loss is untouched: %d → %d, want %d", mine, me.Life, mine+3)
	}
	b14Gain(t, g, opp.ID, 3)
	if opp.Life != theirs+3 {
		t.Errorf("an opponent's gain is untouched: %d → %d", theirs, opp.Life)
	}
}

// --- Rebuff the Wicked / Null Elemental Blast ------------------------

func TestB14RebuffTheWickedCountersOnlySpellsAimedAtYourPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)

	// A Bolt at your face is not a spell that targets a permanent you control.
	face := b10OpponentCastsBolt(t, g, opp, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID})
	if legalCards(g, me.ID, b14RebuffTheWickedOracle)[face] {
		t.Error("a spell targeting you, not a permanent, is not a legal target")
	}
	passPriorityAroundTable(t, g)

	bolt := b10OpponentCastsBolt(t, g, opp, game.TargetRef{Kind: game.TargetCard, ID: bear})
	if !legalCards(g, me.ID, b14RebuffTheWickedOracle)[bolt] {
		t.Fatal("a spell targeting your creature is a legal target")
	}
	b14CastAt(t, g, me, "Rebuff the Wicked", "Instant", b14RebuffTheWickedOracle, nil,
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bolt}}})
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("the Bolt was countered; the Bear lives")
	}
	if b12ZoneOf(g, bolt) != game.ZoneGraveyard {
		t.Error("the countered Bolt goes to its owner's graveyard")
	}
}

func TestB14NullElementalBlastHitsMulticoloredOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gold := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Gold Bear", TypeLine: "Creature — Bear", Colors: []string{"B", "G"},
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	mono := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Green Bear", TypeLine: "Creature — Bear", Colors: []string{"G"},
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	castModal(t, g, "Null Elemental Blast", "Instant", b14NullElementalBlastOracle, []int{1},
		[]game.TargetRef{{Kind: game.TargetCard, ID: gold}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(gold) {
		t.Error("the multicolored permanent is destroyed")
	}
	if err := g.CastSpell(me.ID, b14HandCard(me, "Null Elemental Blast", "Instant", b14NullElementalBlastOracle),
		game.CastSpellParams{Modes: []int{1}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mono}}}); err == nil {
		t.Fatal("a mono-coloured permanent is not a legal target")
	}

	// The counter mode: a gold spell on the stack.
	spell := b14CastAt(t, g, opp, "Gold Spell", "Instant", "", []string{"U", "R"}, game.CastSpellParams{})
	castModal(t, g, "Null Elemental Blast", "Instant", b14NullElementalBlastOracle, []int{0},
		[]game.TargetRef{{Kind: game.TargetCard, ID: spell}})
	passPriorityAroundTable(t, g)
	if b12ZoneOf(g, spell) != game.ZoneGraveyard || g.Stack.Contains(spell) {
		t.Error("the multicolored spell is countered")
	}
}

// b14HandCard puts a card into a player's hand and returns its ID.
func b14HandCard(p *game.Player, name, typeLine, oracle string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, Owner: p.ID, Controller: p.ID,
	})
	return id
}

// --- Desynchronization / Return to Dust / Unmarked Grave -------------

func TestB14DesynchronizationBouncesEverythingThatIsNotHistoric(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	legend := b12Creature(g, opp.ID, "Legend", "Legendary Creature — Human", 3, 3)
	saga := b12Permanent(g, opp.ID, "Saga", "Enchantment — Saga")
	aura := b12Permanent(g, opp.ID, "Curse", "Enchantment — Aura Curse")
	land := seedLandOnBattlefield(g, opp.ID, "Swamp", "Basic Land — Swamp")

	castCatalogSpell(t, g, "Desynchronization", "Instant", b14DesynchronizationOracle, nil)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{bear, theirs, aura} {
		if b12ZoneOf(g, id) != game.ZoneHand {
			t.Errorf("%s should be in its owner's hand, is in %s", id, b12ZoneOf(g, id))
		}
	}
	for _, id := range []uuid.UUID{rock, legend, saga, land} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s is historic or a land and stays", id)
		}
	}
}

func TestB14ReturnToDustExilesTwoOnlyInYourMainPhase(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Permanent(g, opp.ID, "Rock A", "Artifact")
	b := b12Permanent(g, opp.ID, "Shrine B", "Enchantment")
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if _, ok := legalCards(g, me.ID, b14ReturnToDustOracle)[bear]; ok {
		t.Error("a creature is not an artifact or enchantment")
	}
	castCatalogSpell(t, g, "Return to Dust", "Instant", b14ReturnToDustOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}})
	passPriorityAroundTable(t, g)
	if b12ZoneOf(g, a) != game.ZoneExile || b12ZoneOf(g, b) != game.ZoneExile {
		t.Error("cast in your main phase, both targets are exiled")
	}

	// In combat, only the first goes.
	c := b12Permanent(g, opp.ID, "Rock C", "Artifact")
	d := b12Permanent(g, opp.ID, "Shrine D", "Enchantment")
	advanceTo(t, g, game.StepDeclareAttackers)
	b14CastAt(t, g, me, "Return to Dust", "Instant", b14ReturnToDustOracle, nil,
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: c}, {Kind: game.TargetCard, ID: d}}})
	passPriorityAroundTable(t, g)
	if b12ZoneOf(g, c) != game.ZoneExile {
		t.Error("the first target is always exiled")
	}
	if !g.Battlefield.Contains(d) {
		t.Error("outside your main phase the second target is not exiled")
	}
}

func TestB14UnmarkedGraveTutorsANonlegendaryCardToTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	legend := pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Legend", TypeLine: "Legendary Creature — Human", Owner: me.ID, Controller: me.ID})
	ogre := pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Ogre", TypeLine: "Creature — Ogre", Owner: me.ID, Controller: me.ID})
	castCatalogSpell(t, g, "Unmarked Grave", "Sorcery", b14UnmarkedGraveOracle, nil)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the searcher chooses")
	}
	if searchOptionNamed(g, c, "Legend") != uuid.Nil {
		t.Error("a legendary card is not offered")
	}
	if searchOptionNamed(g, c, "Ogre") == uuid.Nil {
		t.Fatal("the nonlegendary card is offered")
	}
	answerSearchNamed(t, g, me.ID, "Ogre")
	if b12ZoneOf(g, ogre) != game.ZoneGraveyard {
		t.Error("the chosen card goes to the graveyard")
	}
	if b12ZoneOf(g, legend) != game.ZoneLibrary {
		t.Error("the legend stays in the library")
	}
}

// --- Eerie Ultimatum -------------------------------------------------

func TestB14EerieUltimatumReturnsOnePermanentPerName(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bearA := seedGraveyardCard(t, g, "Bear", "Creature — Bear", "")
	bearB := seedGraveyardCard(t, g, "Bear", "Creature — Bear", "")
	land := seedGraveyardCard(t, g, "Forest", "Basic Land — Forest", "")
	shrine := seedGraveyardCard(t, g, "Shrine", "Enchantment", "")
	bolt := seedGraveyardCard(t, g, "Bolt", "Instant", "")
	if legalCards(g, me.ID, b14EerieUltimatumOracle)[bolt] {
		t.Error("an instant is not a permanent card")
	}
	castCatalogSpell(t, g, "Eerie Ultimatum", "Sorcery", b14EerieUltimatumOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: bearA}, {Kind: game.TargetCard, ID: bearB},
		{Kind: game.TargetCard, ID: land}, {Kind: game.TargetCard, ID: shrine},
	})
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{bearA, land, shrine} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s should have returned", id)
		}
	}
	if b12ZoneOf(g, bearB) != game.ZoneGraveyard {
		t.Error("the second Bear shares a name and stays")
	}
	if b12ZoneOf(g, bolt) != game.ZoneGraveyard {
		t.Error("the instant stays")
	}
	// "Any number" includes none: castable with nothing chosen.
	castCatalogSpell(t, g, "Eerie Ultimatum", "Sorcery", b14EerieUltimatumOracle, nil)
	passPriorityAroundTable(t, g)
}

// --- Archfiend of Depravity ------------------------------------------

func TestB14ArchfiendOfDepravityLeavesEachOpponentTwoCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Archfiend of Depravity", "Creature — Demon", b14ArchfiendOfDepravityOracle, 5, 4)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	var theirs []uuid.UUID
	for i := 0; i < 4; i++ {
		theirs = append(theirs, b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2))
	}

	// Your own end step: nothing.
	advanceToEndStepOf(t, g, 0)
	if len(g.PendingTriggers) != 0 || len(g.StackMeta) != 0 {
		t.Fatal("the Archfiend does not trigger on its controller's end step")
	}
	advanceToEndStepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	c := sacrificeChoiceFor(g, opp.ID)
	if c == nil {
		t.Fatal("the opponent is asked to sacrifice")
	}
	if len(c.SacrificeOptions) != 4 {
		t.Errorf("the prompt offers their four creatures, offered %d", len(c.SacrificeOptions))
	}
	answerSacrifice(t, g, opp.ID, theirs[0])
	answerSacrifice(t, g, opp.ID, theirs[1])
	if sacrificeChoiceFor(g, opp.ID) != nil {
		t.Error("with two creatures left there is nothing more to sacrifice")
	}
	if n := b14CreaturesControlled(g, opp.ID); n != 2 {
		t.Errorf("the opponent keeps two, has %d", n)
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("the Archfiend's controller is not asked")
	}
	if sacrificeChoiceFor(g, me.ID) != nil {
		t.Error("no prompt for the controller")
	}
}

// --- Kiora's Follower / Scavenging Ooze / Siege-Gang Commander -------

func TestB14KiorasFollowerUntapsAnotherPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	follower := b12Push(g, me.ID, "Kiora's Follower", "Creature — Merfolk", b14KiorasFollowerOracle, 2, 2)
	land := seedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	b05TapOnBattlefield(g, land)
	if legalCards(g, me.ID, b14KiorasFollowerOracle)[follower] {
		t.Error("the Follower cannot target itself")
	}
	if err := g.ActivateCatalogAbility(me.ID, follower, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: land}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !b12Card(t, g, follower).Tapped {
		t.Error("the Follower taps as a cost")
	}
	passPriorityAroundTable(t, g)
	if b12Card(t, g, land).Tapped {
		t.Error("the land is untapped")
	}
}

func TestB14ScavengingOozeGrowsOnCreatureCardsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ooze := b12Push(g, me.ID, "Scavenging Ooze", "Creature — Ooze", b14ScavengingOozeOracle, 2, 2)
	creature := batch01GraveyardCard(opp, "Dead Bear", "Creature — Bear")
	spell := batch01GraveyardCard(opp, "Spent Bolt", "Instant")
	before := me.Life

	b10AddMana(me, "G")
	if err := g.ActivateCatalogAbility(me.ID, ooze, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: creature}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if b12ZoneOf(g, creature) != game.ZoneExile {
		t.Error("the targeted card is exiled")
	}
	if got := b12Counter(t, g, ooze, "+1/+1"); got != 1 {
		t.Errorf("a creature card grows the Ooze: %d counters, want 1", got)
	}
	if me.Life != before+1 {
		t.Errorf("life %d → %d, want +1", before, me.Life)
	}

	b10AddMana(me, "G")
	if err := g.ActivateCatalogAbility(me.ID, ooze, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: spell}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if b12ZoneOf(g, spell) != game.ZoneExile {
		t.Error("a noncreature card is still exiled")
	}
	if got := b12Counter(t, g, ooze, "+1/+1"); got != 1 {
		t.Errorf("a noncreature card does not grow the Ooze: %d counters", got)
	}
	if me.Life != before+1 {
		t.Errorf("no life for a noncreature: %d", me.Life)
	}
}

func TestB14SiegeGangCommanderMakesGoblinsAndThrowsThem(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	commander := castCatalogSpell(t, g, "Siege-Gang Commander", "Creature — Goblin", b14SiegeGangCommanderOracle, nil)
	passPriorityAroundTable(t, g)
	n, goblin := b14Tokens(g, me.ID, "Goblin")
	if n != 3 {
		t.Fatalf("three Goblins, made %d", n)
	}
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b10AddMana(me, "C", "R")
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
	passPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("opponent %d → %d, want -2", before, opp.Life)
	}
}

// --- Hydroid Krasis / Kozilek / Hangarback Walker --------------------

func TestB14HydroidKrasisPaysOutOnCastAndEntersWithXCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	life, hand := me.Life, me.Hand.Size()
	id := castXSpell(t, g, "Hydroid Krasis", "Creature — Jellyfish Hydra Beast", b14HydroidKrasisOracle, "{X}{G}{U}", 5, nil)
	if triggerOnStack(g, id) == nil && len(g.PendingTriggers) == 0 {
		t.Fatal("the cast trigger goes on the stack above the spell")
	}
	passPriorityAroundTable(t, g)
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
	if got := b12Counter(t, g, id, "+1/+1"); got != 5 {
		t.Errorf("enters with X counters: %d, want 5", got)
	}
	if !hasEffectiveKeyword(t, g, id, "flying") || !hasEffectiveKeyword(t, g, id, "trample") {
		t.Error("flying and trample")
	}
}

func TestB14HydroidKrasisPaysOutEvenWhenCountered(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life := me.Life
	id := castXSpell(t, g, "Hydroid Krasis", "Creature — Jellyfish Hydra Beast", b14HydroidKrasisOracle, "{X}{G}{U}", 4, nil)
	// The trigger is on the stack above the spell; counter the spell.
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
	batch01OpponentCasts(t, g, opp, "Counterspell", b14CounterspellOracle, "{U}{U}",
		[]game.TargetRef{{Kind: game.TargetCard, ID: id}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(id) {
		t.Fatal("the Krasis was countered")
	}
	if me.Life != life+2 {
		t.Errorf("the cast trigger still pays out: %d → %d, want +2", life, me.Life)
	}
}

func TestB14KozilekDrawsUpToSevenOnCast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	// Down to three in hand; Kozilek is on the stack, not in it.
	for me.Hand.Size() > 3 {
		if _, err := me.Hand.PopTop(); err != nil {
			t.Fatal(err)
		}
	}
	castCatalogSpell(t, g, "Kozilek, the Great Distortion", "Legendary Creature — Eldrazi", b14KozilekOracle, nil)
	if me.Hand.Size() != 3 {
		t.Fatalf("hand %d, want 3 with Kozilek on the stack", me.Hand.Size())
	}
	if len(g.PendingTriggers) == 0 && len(g.StackMeta) < 2 {
		t.Fatal("the cast trigger goes on the stack above the spell")
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != 7 {
		t.Errorf("draws the difference: hand %d, want 7", me.Hand.Size())
	}
	if n := len(game.ActivatedAbilitiesForCard(game.Card{OracleID: b14KozilekOracle})); n != 0 {
		t.Errorf("the counter ability is declared missing; found %d activated abilities", n)
	}

	// With seven or more in hand the trigger never fires.
	fillHandTo(t, g, me, 8)
	before := me.Hand.Size()
	castCatalogSpell(t, g, "Kozilek, the Great Distortion", "Legendary Creature — Eldrazi", b14KozilekOracle, nil)
	if len(g.PendingTriggers) != 0 || len(g.StackMeta) != 1 {
		t.Fatal("with seven in hand the intervening-if fails and nothing triggers")
	}
	passPriorityAroundTable(t, g)
	// castCatalogSpell seeds the card into the hand before casting, so
	// the count is unchanged when nothing is drawn.
	if me.Hand.Size() != before {
		t.Errorf("hand %d → %d, want no draw", before, me.Hand.Size())
	}
}

func TestB14HangarbackWalkerGrowsAndLeavesThopters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	walker := castXSpell(t, g, "Hangarback Walker", "Artifact Creature — Construct", b14HangarbackWalkerOracle, "{X}{X}", 2, nil)
	passPriorityAroundTable(t, g)
	if got := b12Counter(t, g, walker, "+1/+1"); got != 2 {
		t.Fatalf("enters with X counters: %d, want 2", got)
	}
	b10AddMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, walker, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("a Walker that entered this turn is summoning sick")
	}
	b10Awake(g, walker)
	if err := g.ActivateCatalogAbility(me.ID, walker, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := b12Counter(t, g, walker, "+1/+1"); got != 3 {
		t.Errorf("the tap ability adds a counter: %d, want 3", got)
	}
	b14Kill(g, walker)
	passPriorityAroundTable(t, g)
	n, thopter := b14Tokens(g, me.ID, "Thopter")
	if n != 3 {
		t.Fatalf("three counters leave three Thopters, left %d", n)
	}
	if !thopter.IsArtifact() || !thopter.IsCreature() || !hasEffectiveKeyword(t, g, thopter.InstanceID, "flying") {
		t.Errorf("the Thopter is a flying artifact creature, got %q", thopter.TypeLine)
	}
}

// --- Replicating Ring / Consuming Aberration -------------------------

func TestB14ReplicatingRingReplicatesOnTheEighthNightCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ring := b12Push(g, me.ID, "Replicating Ring", "Snow Artifact", b14ReplicatingRingOracle, 0, 0)
	if err := g.ActivateManaAbility(me.ID, ring, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if pick := riderLatestManaPick(g, me.ID); pick == nil || len(pick.ColorOptions) != 5 {
		t.Fatalf("any colour, not narrowed: %+v", pick)
	}
	b10ResolveAllManaPicks(t, g, me.ID, "U")
	me.ManaPool.EmptyPool()

	b12ToMyNextUpkeep(t, g)
	passPriorityAroundTable(t, g)
	if got := b12Counter(t, g, ring, "night"); got != 1 {
		t.Fatalf("one night counter per upkeep: %d", got)
	}
	if err := g.AddCounter(ring, "night", 6); err != nil {
		t.Fatal(err)
	}
	b12ToMyNextUpkeep(t, g)
	passPriorityAroundTable(t, g)
	if got := b12Counter(t, g, ring, "night"); got != 0 {
		t.Errorf("at eight the counters are removed: %d left", got)
	}
	n, token := b14Tokens(g, me.ID, "Replicated Ring")
	if n != 8 {
		t.Fatalf("eight Replicated Rings, made %d", n)
	}
	if !token.IsArtifact() || !hasCopySupertype(token, "Snow") {
		t.Errorf("the token is a snow artifact, got %q", token.TypeLine)
	}
	if err := g.ActivateManaAbility(me.ID, token.InstanceID, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("a Replicated Ring taps for mana: %v", err)
	}
	if pick := riderLatestManaPick(g, me.ID); pick == nil || len(pick.ColorOptions) != 5 {
		t.Fatalf("the token offers any colour: %+v", pick)
	}
}

func TestB14ConsumingAberrationSizesToOpponentsGraveyardsAndMillsToLands(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	aberration := b12Push(g, me.ID, "Consuming Aberration", "Creature — Horror", b14ConsumingAberrationOracle, 0, 0)
	pushGraveyardCardForTest(opp, "Dead A")
	pushGraveyardCardForTest(opp, "Dead B")
	pushGraveyardCardForTest(me, "My Dead")
	if p := effectivePower(t, g, aberration); p != 2 {
		t.Errorf("two cards in opponents' graveyards: power %d, want 2", p)
	}
	// Library tops: opp reveals Spell, Spell, Land (three milled);
	// other reveals Land at once (one milled).
	b10LibraryTop(opp, "Their Land", "Basic Land — Swamp", "", 0, 0)
	b10LibraryTop(opp, "Their Spell B", "Instant", "{U}", 0, 0)
	b10LibraryTop(opp, "Their Spell A", "Sorcery", "{B}", 0, 0)
	b10LibraryTop(other, "Other Land", "Basic Land — Island", "", 0, 0)
	oppLib, otherLib := opp.Library.Size(), other.Library.Size()

	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
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
	if p := effectivePower(t, g, aberration); p != want || effectiveToughness(t, g, aberration) != want {
		t.Errorf("power %d / toughness %d, want %d", p, effectiveToughness(t, g, aberration), want)
	}
}

// --- Resourceful Defense ---------------------------------------------

func TestB14ResourcefulDefenseMovesCountersWhenAPermanentLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Resourceful Defense", "Enchantment", b14ResourcefulDefenseOracle, 0, 0)
	grown := b12Creature(g, me.ID, "Grown Bear", "Creature — Bear", 2, 2)
	plain := b12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	keeper := b12Creature(g, me.ID, "Keeper", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := g.AddCounter(grown, "+1/+1", 2); err != nil {
		t.Fatal(err)
	}
	if err := g.AddCounter(grown, "charge", 1); err != nil {
		t.Fatal(err)
	}

	// A permanent with no counters leaving does not trigger.
	b14Kill(g, plain)
	if len(g.PendingTriggers) != 0 || latestPickTarget(g, me.ID) != nil {
		t.Fatal("no counters, no trigger")
	}
	// An opponent's permanent does not trigger.
	if err := g.AddCounter(theirs, "+1/+1", 1); err != nil {
		t.Fatal(err)
	}
	b14Kill(g, theirs)
	if len(g.PendingTriggers) != 0 || latestPickTarget(g, me.ID) != nil {
		t.Fatal("an opponent's permanent does not trigger it")
	}

	b14Kill(g, grown)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, keeper)
	passPriorityAroundTable(t, g)
	if got := b12Counter(t, g, keeper, "+1/+1"); got != 2 {
		t.Errorf("the +1/+1 counters move: %d, want 2", got)
	}
	if got := b12Counter(t, g, keeper, "charge"); got != 1 {
		t.Errorf("every kind moves: charge %d, want 1", got)
	}
}

func TestB14ResourcefulDefenseMovesAllCountersForFiveMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	defense := b12Push(g, me.ID, "Resourceful Defense", "Enchantment", b14ResourcefulDefenseOracle, 0, 0)
	from := b12Creature(g, me.ID, "From", "Creature — Bear", 2, 2)
	to := b12Creature(g, me.ID, "To", "Creature — Bear", 2, 2)
	if err := g.AddCounter(from, "+1/+1", 3); err != nil {
		t.Fatal(err)
	}
	b10AddMana(me, "C", "C", "C", "C", "W")
	if err := g.ActivateCatalogAbility(me.ID, defense, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: from}, {Kind: game.TargetCard, ID: to}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := b12Counter(t, g, from, "+1/+1"); got != 0 {
		t.Errorf("the source is emptied: %d left", got)
	}
	if got := b12Counter(t, g, to, "+1/+1"); got != 3 {
		t.Errorf("the destination receives them: %d, want 3", got)
	}
}

// --- Cloud, Midgar Mercenary -----------------------------------------

func TestB14CloudTutorsAnEquipmentToHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sword := pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Sword", TypeLine: "Artifact — Equipment", Owner: me.ID, Controller: me.ID})
	pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Axe", TypeLine: "Artifact — Equipment", Owner: me.ID, Controller: me.ID})
	pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	castCatalogSpell(t, g, "Cloud, Midgar Mercenary", "Legendary Creature — Human Soldier Mercenary", b14CloudOracle, nil)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the searcher chooses")
	}
	if searchOptionNamed(g, c, "Rock") != uuid.Nil {
		t.Error("a non-Equipment artifact is not offered")
	}
	answerSearchNamed(t, g, me.ID, "Sword")
	if b12ZoneOf(g, sword) != game.ZoneHand {
		t.Error("the Equipment goes to hand")
	}
	if spec, _ := Lookup(b14CloudOracle); len(spec.Static) != 0 || len(spec.Triggered) != 1 {
		t.Error("the trigger-doubling static is declared missing; only the tutor is live")
	}
}
