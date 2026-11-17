package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hml_5635_cards_test.go — a batch of Homelands cards (#5635). Every
// test plays the real catalog entry through the engine.

const (
	hmlAnHavvaTownshipOracle    = "40ae17be-9998-4ee4-9d95-82a08895405f"
	hmlAysenAbbeyOracle         = "6ff85e73-bf7a-4a9c-80ef-6ce76656fab7"
	hmlCastleSengirOracle       = "c7f0251a-9341-4ff2-8b15-31c06eb4f2e7"
	hmlKoskunKeepOracle         = "184c5a0d-7654-4421-86e4-7f04bcf49494"
	hmlWizardsSchoolOracle      = "51635f96-af1d-4b33-9118-432495aaf07f"
	hmlAbbeyGargoylesOracle     = "c208b959-d0e4-4a9a-8255-2c7cc7596767"
	hmlAmbushPartyOracle        = "454ed336-362d-46d9-96d5-8fcaa8a1c543"
	hmlAnabaBodyguardOracle     = "8c64db66-879e-46ad-9d11-8a95b6cf5d19"
	hmlCemeteryGateOracle       = "443f8dd1-333d-42c4-a286-302a9496209a"
	hmlDeathSpeakersOracle      = "4a1db129-cc7c-410a-a41e-d633e30cfdad"
	hmlDwarvenTraderOracle      = "b40c5f3e-7080-4828-9215-79727c7a911f"
	hmlEbonyRhinoOracle         = "3d2b5201-c2bf-4753-89ad-543795740216"
	hmlNarwhalOracle            = "3552eea9-aee9-4555-b3cb-8e4fc5b728cc"
	hmlSeaSpriteOracle          = "b318d94e-314f-4df8-a178-b5ddd756e1c4"
	hmlAbbeyMatronOracle        = "62e3f285-886c-414e-b4ff-403a7c01c23a"
	hmlMesaFalconOracle         = "bc4c5f33-0f3d-4815-9974-fb12ac725d08"
	hmlAnabaShamanOracle        = "bc416c17-e1f0-4344-8c21-6184cac10205"
	hmlGrandmotherSengirOracle  = "4f92804f-a231-4406-9a16-f4e102517b06"
	hmlChandlerOracle           = "520abf49-7631-4e7f-889e-547ccc8c3b95"
	hmlJovenOracle              = "9e767d44-feff-449a-bed9-0866c7bce846"
	hmlAysenBureaucratsOracle   = "786b2028-cf24-4f00-b1d6-05b94e78d617"
	hmlLeapingLizardOracle      = "23d42beb-5293-4f12-a732-3531a0fa0ca0"
	hmlEronTheRelentlessOracle  = "cabcc2b7-0cdb-4ec2-b6a3-47644335abff"
	hmlAnabaSpiritCrafterOracle = "e9bee79c-144f-41e4-a53d-17df8381f691"
	hmlAysenHighwayOracle       = "415bc26d-60fb-41ac-8006-7dd861671c14"
	hmlFeastOfTheUnicornOracle  = "274d89b8-1e59-4992-9299-dc793b7f6752"
	hmlShrinkOracle             = "252330b1-67cf-4d9e-a413-917ba61e731f"
)

// hmlSetColors gives a test permanent printed colours, for the
// predicates and protections that read them.
func hmlSetColors(g *game.Game, id uuid.UUID, colors ...string) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Colors = colors
			}
		}
	})
}

func hmlTapped(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.Tapped
}

// --- every card is in the catalog ----------------------------------------

func TestHml5635EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		hmlAnHavvaTownshipOracle, hmlAysenAbbeyOracle, hmlCastleSengirOracle,
		hmlKoskunKeepOracle, hmlWizardsSchoolOracle, hmlAbbeyGargoylesOracle,
		hmlAmbushPartyOracle, hmlAnabaBodyguardOracle, hmlCemeteryGateOracle,
		hmlDeathSpeakersOracle, hmlDwarvenTraderOracle, hmlEbonyRhinoOracle,
		hmlNarwhalOracle, hmlSeaSpriteOracle, hmlAbbeyMatronOracle,
		hmlMesaFalconOracle, hmlAnabaShamanOracle, hmlGrandmotherSengirOracle,
		hmlChandlerOracle, hmlJovenOracle, hmlAysenBureaucratsOracle,
		hmlLeapingLizardOracle, hmlEronTheRelentlessOracle, hmlAnabaSpiritCrafterOracle,
		hmlAysenHighwayOracle, hmlFeastOfTheUnicornOracle, hmlShrinkOracle,
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

// --- the keyword creatures -----------------------------------------------

func TestHml5635KeywordCreatures(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Abbey Gargoyles", "Creature — Gargoyle", hmlAbbeyGargoylesOracle, []string{"flying", "protection from red"}},
		{"Ambush Party", "Creature — Human Rogue", hmlAmbushPartyOracle, []string{"first strike", "haste"}},
		{"Anaba Bodyguard", "Creature — Minotaur", hmlAnabaBodyguardOracle, []string{"first strike"}},
		{"Cemetery Gate", "Creature — Wall", hmlCemeteryGateOracle, []string{"defender", "protection from black"}},
		{"Death Speakers", "Creature — Human Cleric", hmlDeathSpeakersOracle, []string{"protection from black"}},
		{"Ebony Rhino", "Artifact Creature — Rhino", hmlEbonyRhinoOracle, []string{"trample"}},
		{"Narwhal", "Creature — Whale", hmlNarwhalOracle, []string{"first strike", "protection from red"}},
		{"Sea Sprite", "Creature — Faerie", hmlSeaSpriteOracle, []string{"flying", "protection from red"}},
		{"Mesa Falcon", "Creature — Bird", hmlMesaFalconOracle, []string{"flying"}},
		{"Eron the Relentless", "Legendary Creature — Human Rogue", hmlEronTheRelentlessOracle, []string{"haste"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := b12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 1, 1)
			abilities := effectiveAbilities(t, g, id)
			for _, kw := range tc.keywords {
				if !hasAbility(abilities, kw) {
					t.Errorf("%s has no %s", tc.name, kw)
				}
			}
		})
	}
}

// Dwarven Trader is vanilla: no keyword and no ability rows.
func TestHml5635DwarvenTraderIsVanilla(t *testing.T) {
	spec, ok := Lookup(hmlDwarvenTraderOracle)
	if !ok {
		t.Fatal("Dwarven Trader is not in the catalog")
	}
	if len(spec.PrintedKeywords) != 0 || len(spec.Activated) != 0 || len(spec.Static) != 0 || len(spec.Triggered) != 0 {
		t.Errorf("Dwarven Trader declares rules: %+v", spec)
	}
}

// Protection from red: a red spell can't target Sea Sprite.
func TestHml5635SeaSpriteCantBeTargetedByRed(t *testing.T) {
	g := newCatalogGame(t)
	sprite := b12Push(g, g.Seats[1].ID, "Sea Sprite", "Creature — Faerie", hmlSeaSpriteOracle, 1, 1)
	hmlSetColors(g, sprite, "U")
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	bolt := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant", OracleID: lightningBoltOracle,
		ManaCost: "{R}", Colors: []string{"R"}, Owner: me.ID, Controller: me.ID,
	})
	me.ManaPool.AddMana(game.ManaToken{Color: "R"})
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Targets: cardTarget(sprite)}); err == nil {
		t.Error("Lightning Bolt targeted a creature with protection from red")
	}
}

// --- the land cycle --------------------------------------------------------

func TestHml5635TieredLandsAddWhatTheyPrint(t *testing.T) {
	cases := []struct {
		name, oracle, one, a, b string
	}{
		{"An-Havva Township", hmlAnHavvaTownshipOracle, "G", "R", "W"},
		{"Aysen Abbey", hmlAysenAbbeyOracle, "W", "G", "U"},
		{"Castle Sengir", hmlCastleSengirOracle, "B", "U", "R"},
		{"Koskun Keep", hmlKoskunKeepOracle, "R", "B", "G"},
		{"Wizards' School", hmlWizardsSchoolOracle, "U", "W", "B"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// {T}: Add {C}.
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := top100ActiveSeat(g)
			land := pushCatalogPermanent(g, me.ID, tc.name, "Land", tc.oracle, false)
			if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
				t.Fatal(err)
			}
			if n := me.ManaPool.CountColor("C"); n != 1 || me.ManaPool.Total() != 1 {
				t.Errorf("{T} row: pool holds %d colorless of %d, want 1 of 1", n, me.ManaPool.Total())
			}

			// {1}, {T}: Add the one colour.
			g = newCatalogGame(t)
			advanceToMain(t, g)
			me = top100ActiveSeat(g)
			land = pushCatalogPermanent(g, me.ID, tc.name, "Land", tc.oracle, false)
			apaMana(me, "C")
			if err := g.ActivateManaAbility(me.ID, land, 1, game.ManaAbilityParams{}); err != nil {
				t.Fatal(err)
			}
			if n := me.ManaPool.CountColor(tc.one); n != 1 || me.ManaPool.Total() != 1 {
				t.Errorf("{1} row: pool holds %d {%s} of %d, want 1 of 1", n, tc.one, me.ManaPool.Total())
			}
			if !hmlTapped(t, g, land) {
				t.Error("{1} row did not tap the land")
			}

			// {2}, {T}: Add either of the pair, and not a colour outside it.
			for _, pick := range []string{tc.a, tc.b} {
				g = newCatalogGame(t)
				advanceToMain(t, g)
				me = top100ActiveSeat(g)
				land = pushCatalogPermanent(g, me.ID, tc.name, "Land", tc.oracle, false)
				apaMana(me, "C", "C")
				if err := g.ActivateManaAbility(me.ID, land, 2, game.ManaAbilityParams{Colors: []string{pick}}); err != nil {
					t.Fatal(err)
				}
				if n := me.ManaPool.CountColor(pick); n != 1 || me.ManaPool.Total() != 1 {
					t.Errorf("{2} row picking %s: pool holds %d of %d, want 1 of 1", pick, n, me.ManaPool.Total())
				}
			}
			g = newCatalogGame(t)
			advanceToMain(t, g)
			me = top100ActiveSeat(g)
			land = pushCatalogPermanent(g, me.ID, tc.name, "Land", tc.oracle, false)
			apaMana(me, "C", "C")
			if err := g.ActivateManaAbility(me.ID, land, 2, game.ManaAbilityParams{Colors: []string{tc.one}}); err == nil {
				t.Errorf("{2} row produced {%s}, outside its pair", tc.one)
			}
		})
	}
}

// The {1} row can't be paid from an empty pool, and the land stays untapped.
func TestHml5635TieredLandNeedsTheGenericMana(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	land := pushCatalogPermanent(g, me.ID, "Castle Sengir", "Land", hmlCastleSengirOracle, false)
	if err := g.ActivateManaAbility(me.ID, land, 1, game.ManaAbilityParams{}); err == nil {
		t.Error("the {1} row activated with nothing to pay {1}")
	}
	if hmlTapped(t, g, land) {
		t.Error("the refused activation tapped the land")
	}
}

// --- the self pumps ----------------------------------------------------------

func TestHml5635SelfPumps(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
		mana                   []string
		power, toughness       int
		tapped                 bool
	}{
		{"Abbey Matron", "Creature — Human Cleric", hmlAbbeyMatronOracle, []string{"W"}, 1, 6, true},
		{"Mesa Falcon", "Creature — Bird", hmlMesaFalconOracle, []string{"C", "W"}, 1, 4, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := top100ActiveSeat(g)
			src := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 1, 3)
			apaMana(me, tc.mana...)
			apaActivate(t, g, me, src, 0, game.ActivateAbilityParams{})
			if got := hmlTapped(t, g, src); got != tc.tapped {
				t.Errorf("tapped = %v, want %v", got, tc.tapped)
			}
			if p, tt := ptOf(t, g, src); p != tc.power || tt != tc.toughness {
				t.Errorf("after the pump %d/%d, want %d/%d", p, tt, tc.power, tc.toughness)
			}
			advanceToMainOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
			if p, tt := ptOf(t, g, src); p != 1 || tt != 3 {
				t.Errorf("next turn %d/%d, want 1/3", p, tt)
			}
		})
	}
}

// The Matron's {T} keeps it off the turn it arrives; the Falcon's ability
// has no {T} and works at once.
func TestHml5635SummoningSickness(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	matron := pushCatalogPermanent(g, me.ID, "Abbey Matron", "Creature — Human Cleric", hmlAbbeyMatronOracle, true)
	apaMana(me, "W")
	if err := g.ActivateCatalogAbility(me.ID, matron, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("a summoning-sick Abbey Matron paid {T}")
	}
	falcon := pushCatalogPermanent(g, me.ID, "Mesa Falcon", "Creature — Bird", hmlMesaFalconOracle, true)
	me.ManaPool.EmptyPool()
	apaMana(me, "C", "W")
	apaActivate(t, g, me, falcon, 0, game.ActivateAbilityParams{})
}

// Leaping Lizard loses a point of toughness each time, and the third
// activation in a turn kills it.
func TestHml5635LeapingLizardFliesAndShrinks(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	lizard := b12Push(g, me.ID, "Leaping Lizard", "Creature — Lizard", hmlLeapingLizardOracle, 2, 3)
	if hasEffectiveKeyword(t, g, lizard, "flying") {
		t.Fatal("flying before the ability")
	}
	apaMana(me, "C", "G")
	apaActivate(t, g, me, lizard, 0, game.ActivateAbilityParams{})
	if !hasEffectiveKeyword(t, g, lizard, "flying") {
		t.Error("no flying after the ability")
	}
	if p, tt := ptOf(t, g, lizard); p != 2 || tt != 2 {
		t.Errorf("after one activation %d/%d, want 2/2", p, tt)
	}
	apaMana(me, "C", "G")
	apaActivate(t, g, me, lizard, 0, game.ActivateAbilityParams{})
	if p, tt := ptOf(t, g, lizard); p != 2 || tt != 1 {
		t.Errorf("after two activations %d/%d, want 2/1", p, tt)
	}
	apaMana(me, "C", "G")
	apaActivate(t, g, me, lizard, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(lizard) {
		t.Error("the Lizard survived at 0 toughness")
	}
}

func TestHml5635LeapingLizardFlyingEndsAtCleanup(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	lizard := b12Push(g, me.ID, "Leaping Lizard", "Creature — Lizard", hmlLeapingLizardOracle, 2, 3)
	apaMana(me, "C", "G")
	apaActivate(t, g, me, lizard, 0, game.ActivateAbilityParams{})
	advanceToMainOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	if hasEffectiveKeyword(t, g, lizard, "flying") {
		t.Error("flying outlasted the turn")
	}
	if p, tt := ptOf(t, g, lizard); p != 2 || tt != 3 {
		t.Errorf("next turn %d/%d, want 2/3", p, tt)
	}
}

// --- Eron the Relentless -----------------------------------------------------

func TestHml5635EronRegenerates(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	eron := b12Push(g, me.ID, "Eron the Relentless", "Legendary Creature — Human Rogue", hmlEronTheRelentlessOracle, 5, 2)
	apaMana(me, "R", "R", "R")
	apaActivate(t, g, me, eron, 0, game.ActivateAbilityParams{})
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(eron); err != nil {
			t.Fatal(err)
		}
	})
	if !g.Battlefield.Contains(eron) {
		t.Fatal("the shielded Eron was destroyed")
	}
	if !hmlTapped(t, g, eron) {
		t.Error("regeneration did not tap Eron")
	}
	// The shield is used up: a second destruction this turn sticks.
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(eron); err != nil {
			t.Fatal(err)
		}
	})
	if g.Battlefield.Contains(eron) {
		t.Error("one shield saved Eron twice")
	}
}

// --- the targeted abilities ------------------------------------------------

func TestHml5635AnabaShamanPingsAnyTarget(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	shaman := pushCatalogPermanent(g, me.ID, "Anaba Shaman", "Creature — Minotaur Shaman", hmlAnabaShamanOracle, false)
	life := opp.Life
	apaMana(me, "R")
	apaActivate(t, g, me, shaman, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	if opp.Life != life-1 {
		t.Errorf("opponent life %d -> %d, want -1", life, opp.Life)
	}
	if !hmlTapped(t, g, shaman) {
		t.Error("the Shaman was not tapped to pay the cost")
	}
}

func TestHml5635GrandmotherSengirShrinksATarget(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	sengir := pushCatalogPermanent(g, me.ID, "Grandmother Sengir", "Legendary Creature — Human Wizard", hmlGrandmotherSengirOracle, false)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	apaMana(me, "C", "B")
	apaActivate(t, g, me, sengir, 0, game.ActivateAbilityParams{Targets: cardTarget(bear)})
	if p, tt := ptOf(t, g, bear); p != 1 || tt != 1 {
		t.Errorf("target %d/%d, want 1/1", p, tt)
	}
	if !hmlTapped(t, g, sengir) {
		t.Error("Grandmother Sengir was not tapped to pay the cost")
	}
}

func TestHml5635ChandlerDestroysOnlyArtifactCreatures(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	chandler := pushCatalogPermanent(g, me.ID, "Chandler", "Legendary Creature — Human Rogue", hmlChandlerOracle, false)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	golem := b12Creature(g, opp.ID, "Golem", "Artifact Creature — Golem", 3, 3)
	for _, wrong := range []uuid.UUID{bear, rock} {
		apaMana(me, "R", "R", "R")
		if err := g.ActivateCatalogAbility(me.ID, chandler, 0, game.ActivateAbilityParams{Strict: true, Targets: cardTarget(wrong)}); err == nil {
			t.Fatal("Chandler targeted a permanent that is not an artifact creature")
		}
		me.ManaPool.EmptyPool()
	}
	apaMana(me, "R", "R", "R")
	apaActivate(t, g, me, chandler, 0, game.ActivateAbilityParams{Targets: cardTarget(golem)})
	if g.Battlefield.Contains(golem) {
		t.Error("the artifact creature survived Chandler")
	}
}

func TestHml5635JovenDestroysOnlyNoncreatureArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	joven := pushCatalogPermanent(g, me.ID, "Joven", "Legendary Creature — Human Rogue", hmlJovenOracle, false)
	golem := b12Creature(g, opp.ID, "Golem", "Artifact Creature — Golem", 3, 3)
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	apaMana(me, "R", "R", "R")
	if err := g.ActivateCatalogAbility(me.ID, joven, 0, game.ActivateAbilityParams{Strict: true, Targets: cardTarget(golem)}); err == nil {
		t.Fatal("Joven targeted an artifact creature")
	}
	me.ManaPool.EmptyPool()
	apaMana(me, "R", "R", "R")
	apaActivate(t, g, me, joven, 0, game.ActivateAbilityParams{Targets: cardTarget(rock)})
	if g.Battlefield.Contains(rock) {
		t.Error("the noncreature artifact survived Joven")
	}
}

func TestHml5635AysenBureaucratsTapsOnlyASmallCreature(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	clerks := pushCatalogPermanent(g, me.ID, "Aysen Bureaucrats", "Creature — Human Advisor", hmlAysenBureaucratsOracle, false)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	ogre := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	if err := g.ActivateCatalogAbility(me.ID, clerks, 0, game.ActivateAbilityParams{Strict: true, Targets: cardTarget(ogre)}); err == nil {
		t.Fatal("the Bureaucrats targeted a creature with power 3")
	}
	apaActivate(t, g, me, clerks, 0, game.ActivateAbilityParams{Targets: cardTarget(bear)})
	if !hmlTapped(t, g, bear) {
		t.Error("the 2-power target was not tapped")
	}
	if !hmlTapped(t, g, clerks) {
		t.Error("the Bureaucrats were not tapped to pay the cost")
	}
}

func TestHml5635ShrinkTakesFivePower(t *testing.T) {
	g := newCatalogGame(t)
	bear := b12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 6, 6)
	castCatalogSpell(t, g, "Shrink", "Instant", hmlShrinkOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if p, tt := ptOf(t, g, bear); p != 1 || tt != 6 {
		t.Errorf("target %d/%d, want 1/6", p, tt)
	}
	advanceToMainOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	if p, tt := ptOf(t, g, bear); p != 6 || tt != 6 {
		t.Errorf("next turn %d/%d, want 6/6", p, tt)
	}
}

// --- the statics ---------------------------------------------------------------

func TestHml5635AnabaSpiritCrafterPumpsEveryMinotaur(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	crafter := b12Push(g, me.ID, "Anaba Spirit Crafter", "Creature — Minotaur Shaman", hmlAnabaSpiritCrafterOracle, 1, 3)
	theirs := b12Creature(g, opp.ID, "Their Minotaur", "Creature — Minotaur Warrior", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if p, tt := ptOf(t, g, crafter); p != 2 || tt != 3 {
		t.Errorf("the Spirit Crafter is %d/%d, want 2/3", p, tt)
	}
	if p, tt := ptOf(t, g, theirs); p != 3 || tt != 2 {
		t.Errorf("an opponent's Minotaur is %d/%d, want 3/2", p, tt)
	}
	if p, tt := ptOf(t, g, bear); p != 2 || tt != 2 {
		t.Errorf("a Bear is %d/%d, want 2/2", p, tt)
	}
}

func TestHml5635AysenHighwayGivesWhiteCreaturesPlainswalk(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Aysen Highway", "Enchantment", hmlAysenHighwayOracle, 0, 0)
	knight := b12Creature(g, opp.ID, "Knight", "Creature — Human Knight", 2, 2)
	hmlSetColors(g, knight, "W")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	hmlSetColors(g, bear, "G")
	if !hasEffectiveKeyword(t, g, knight, "plainswalk") {
		t.Error("a white creature has no plainswalk")
	}
	if hasEffectiveKeyword(t, g, bear, "plainswalk") {
		t.Error("a green creature has plainswalk")
	}
}

func TestHml5635FeastOfTheUnicornPumpsItsHost(t *testing.T) {
	g := newCatalogGame(t)
	bear := b12Creature(g, g.Seats[0].ID, "Bear", "Creature — Bear", 2, 2)
	auraCast(t, g, "Feast of the Unicorn", hmlFeastOfTheUnicornOracle, bear)
	if p, tt := ptOf(t, g, bear); p != 6 || tt != 2 {
		t.Errorf("enchanted %d/%d, want 6/2", p, tt)
	}
}
