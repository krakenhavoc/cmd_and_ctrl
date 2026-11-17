package effects

import (
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// all_5634_cards_test.go — a batch of Alliances cards (#5634). Every
// test plays the real catalog entry through the engine.

const (
	alliStormCrowOracle            = "000d5588-5a4c-434e-988d-396632ade42c"
	alliDeadlyInsectOracle         = "98273b0d-2b41-486b-9498-30a692c03982"
	alliElvishRangerOracle         = "092f2c34-1f75-4998-b219-2cf1ca73656d"
	alliAesthirGliderOracle        = "99dcde0b-4f4d-4d7a-89aa-8a89c256e38f"
	alliStormShamanOracle          = "65e07c0f-4eff-4b38-9785-63684d67c0d8"
	alliYavimayaAncientsOracle     = "0e91a873-a680-4ca2-b073-26aa39376d77"
	alliWildAesthirOracle          = "dbdf076c-18ce-45e3-8f70-0a32391ea4b6"
	alliEnslavedScoutOracle        = "f39a3798-1ff0-4d1b-933f-bfcc0f265168"
	alliGorillaChieftainOracle     = "4137ee86-bbf0-49dc-87c5-b48a384baac4"
	alliLimDulsHighGuardOracle     = "b0873be6-0b77-4ca4-8c45-a1207b184771"
	alliSoldeviSteamBeastOracle    = "935728dd-ad97-47a9-97ca-068af7960d07"
	alliAgentOfStromgaldOracle     = "099fd4b3-92fd-4b5f-bb4f-eefa6f09700c"
	alliSchoolOfTheUnseenOracle    = "5028dfe8-c505-4643-b493-760b1f19d47f"
	alliPillageOracle              = "0b137853-7cb9-424b-8285-12938991eafb"
	alliReprisalOracle             = "fae90808-bafa-44d8-972a-17ded87576ee"
	alliExileOracle                = "8c91d50b-bf58-4d06-b3b7-4b6db867bfce"
	alliFeastOrFamineOracle        = "485f710d-5bc2-4c8f-b7e9-eed6caa261a6"
	alliFeveredStrengthOracle      = "52512d7a-9b1c-4475-80aa-c95f4efa1015"
	alliBurnoutOracle              = "90865918-2f9a-4ab6-a881-e86f61a527cc"
	alliCarrierPigeonsOracle       = "2d30a634-43c6-4e78-ba10-7ef7ee17ac40"
	alliStenchOfDecayOracle        = "4d4fbb16-0421-4e9c-af9e-e2c12a11c4ed"
	alliHailStormOracle            = "73eacb1f-3d35-4ead-bead-f85e6cfc17ab"
	alliKaysaOracle                = "0fd18a9b-c112-4e48-8fd7-e53fe4500943"
	alliJuniperOrderAdvocateOracle = "ded2fc21-1a69-48f8-865a-467e21b2758b"
	alliPhyrexianBoonOracle        = "a18e6468-b8a2-4681-b8fa-223ab5f3061a"
	alliVisceridArmorOracle        = "87b4cbec-3521-4bdb-bf01-64ddb526b655"
	alliNobleSteedsOracle          = "d6005e47-d545-4c16-b54a-bfe1dde61a98"
	alliUnlikelyAllianceOracle     = "f7ac18bc-ea63-4808-b709-92c1e69cf541"
	alliShieldSphereOracle         = "df9ec9b0-be22-4136-923b-11a71ca4b6c2"
	alliSoldierOfFortuneOracle     = "8b2367b3-3a8d-442f-8d58-6e4c03fa5a46"
	alliYavimayaAntsOracle         = "3f5bbfaa-cbd7-47af-9aa2-abb49d672734"
)

// alliSetColors gives a test permanent printed colours.
func alliSetColors(g *game.Game, id uuid.UUID, colors ...string) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Colors = colors
			}
		}
	})
}

func alliTapped(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.Tapped
}

// alliCastInPlace casts an instant from `caster`'s hand at the current
// step, for the cards that read combat state.
func alliCastInPlace(t *testing.T, g *game.Game, caster uuid.UUID, name, oracle string, targets []game.TargetRef) error {
	t.Helper()
	for _, p := range g.Seats {
		if p == nil || p.ID != caster {
			continue
		}
		id := uuid.New()
		p.Hand.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: "Instant", OracleID: oracle,
			Owner: caster, Controller: caster,
		})
		return g.CastSpell(caster, id, game.CastSpellParams{Targets: targets})
	}
	t.Fatalf("no seat %s", caster)
	return nil
}

// alliAnswerPlayerPick answers the open pick_target prompt with the first
// player it offers, and returns that player.
func alliAnswerPlayerPick(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c == nil || c.Kind != game.PendingChoicePickTarget || len(c.PickTargetPlayers) == 0 {
			continue
		}
		pick := c.PickTargetPlayers[0]
		if err := g.ResolvePickTarget(c.ID, c.Chooser, game.TargetRef{Kind: game.TargetPlayer, ID: pick}); err != nil {
			t.Fatalf("ResolvePickTarget: %v", err)
		}
		return pick
	}
	t.Fatal("no pick_target prompt offering a player")
	return uuid.Nil
}

// --- every card is in the catalog ----------------------------------------

func TestAll5634EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		alliStormCrowOracle, alliDeadlyInsectOracle, alliElvishRangerOracle,
		alliAesthirGliderOracle, alliStormShamanOracle, alliYavimayaAncientsOracle,
		alliWildAesthirOracle, alliEnslavedScoutOracle, alliGorillaChieftainOracle,
		alliLimDulsHighGuardOracle, alliSoldeviSteamBeastOracle, alliAgentOfStromgaldOracle,
		alliSchoolOfTheUnseenOracle, alliPillageOracle, alliReprisalOracle,
		alliExileOracle, alliFeastOrFamineOracle, alliFeveredStrengthOracle,
		alliBurnoutOracle, alliCarrierPigeonsOracle, alliStenchOfDecayOracle,
		alliHailStormOracle, alliKaysaOracle, alliJuniperOrderAdvocateOracle,
		alliPhyrexianBoonOracle, alliVisceridArmorOracle, alliNobleSteedsOracle,
		alliUnlikelyAllianceOracle, alliShieldSphereOracle, alliSoldierOfFortuneOracle,
		alliYavimayaAntsOracle,
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

// --- keyword bodies ----------------------------------------------------------

func TestAll5634KeywordCreatures(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Storm Crow", "Creature — Bird", alliStormCrowOracle, []string{"flying"}},
		{"Deadly Insect", "Creature — Insect", alliDeadlyInsectOracle, []string{"shroud"}},
		{"Aesthir Glider", "Artifact Creature — Bird Construct", alliAesthirGliderOracle, []string{"flying"}},
		{"Wild Aesthir", "Creature — Bird", alliWildAesthirOracle, []string{"flying", "first strike"}},
		{"Lim-Dûl's High Guard", "Creature — Skeleton", alliLimDulsHighGuardOracle, []string{"first strike"}},
		{"Carrier Pigeons", "Creature — Bird", alliCarrierPigeonsOracle, []string{"flying"}},
		{"Shield Sphere", "Artifact Creature — Wall", alliShieldSphereOracle, []string{"defender"}},
		{"Yavimaya Ants", "Creature — Insect", alliYavimayaAntsOracle, []string{"trample", "haste"}},
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

func TestAll5634ElvishRangerIsVanilla(t *testing.T) {
	spec, ok := Lookup(alliElvishRangerOracle)
	if !ok {
		t.Fatal("Elvish Ranger is not in the catalog")
	}
	if len(spec.PrintedKeywords) != 0 || len(spec.Activated) != 0 || len(spec.Static) != 0 || len(spec.Triggered) != 0 {
		t.Errorf("Elvish Ranger declares rules: %+v", spec)
	}
}

func TestAll5634AesthirGliderCantBlock(t *testing.T) {
	g := newCatalogGame(t)
	glider := b12Push(g, g.Seats[0].ID, "Aesthir Glider", "Artifact Creature — Bird Construct", alliAesthirGliderOracle, 2, 1)
	if auraRestrictions(t, g, glider)&game.CantBlock == 0 {
		t.Error("Aesthir Glider can block")
	}
	if auraRestrictions(t, g, glider)&game.CantAttack != 0 {
		t.Error("Aesthir Glider can't attack")
	}
}

// Shroud: not even its controller's spell can target Deadly Insect.
func TestAll5634DeadlyInsectCantBeTargeted(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	insect := b12Push(g, me.ID, "Deadly Insect", "Creature — Insect", alliDeadlyInsectOracle, 6, 1)
	bolt := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant", OracleID: lightningBoltOracle,
		ManaCost: "{R}", Colors: []string{"R"}, Owner: me.ID, Controller: me.ID,
	})
	me.ManaPool.AddMana(game.ManaToken{Color: "R"})
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Targets: cardTarget(insect)}); err == nil {
		t.Error("Lightning Bolt targeted a creature with shroud")
	}
}

// --- the self pumps ------------------------------------------------------------

func TestAll5634SelfPumps(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
		power, toughness       int
		mana                   []string
		wantP, wantT           int
	}{
		{"Storm Shaman", "Creature — Human Cleric Shaman", alliStormShamanOracle, 0, 4, []string{"R"}, 1, 4},
		{"Yavimaya Ancients", "Creature — Treefolk", alliYavimayaAncientsOracle, 2, 7, []string{"G"}, 3, 5},
		{"Wild Aesthir", "Creature — Bird", alliWildAesthirOracle, 1, 1, []string{"W", "W"}, 3, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := top100ActiveSeat(g)
			src := pushCatalogPermanent(g, me.ID, tc.name, tc.typeLine, tc.oracle, true)
			g.WithWriteLock(func() {
				c := findBattlefieldCardForTest(g, src)
				c.Power, c.Toughness = tc.power, tc.toughness
			})
			apaMana(me, tc.mana...)
			apaActivate(t, g, me, src, 0, game.ActivateAbilityParams{})
			if p, tt := ptOf(t, g, src); p != tc.wantP || tt != tc.wantT {
				t.Errorf("after the pump %d/%d, want %d/%d", p, tt, tc.wantP, tc.wantT)
			}
			if alliTapped(t, g, src) {
				t.Error("the ability tapped its source")
			}
			advanceToMainOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
			if p, tt := ptOf(t, g, src); p != tc.power || tt != tc.toughness {
				t.Errorf("next turn %d/%d, want %d/%d", p, tt, tc.power, tc.toughness)
			}
		})
	}
}

// Storm Shaman's ability has no limit; Wild Aesthir's works once a turn.
func TestAll5634WildAesthirPumpsOnlyOnceATurn(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	shaman := b12Push(g, me.ID, "Storm Shaman", "Creature — Human Cleric Shaman", alliStormShamanOracle, 0, 4)
	for range 3 {
		apaMana(me, "R")
		apaActivate(t, g, me, shaman, 0, game.ActivateAbilityParams{})
	}
	if p, _ := ptOf(t, g, shaman); p != 3 {
		t.Errorf("Storm Shaman after three activations has power %d, want 3", p)
	}

	aesthir := b12Push(g, me.ID, "Wild Aesthir", "Creature — Bird", alliWildAesthirOracle, 1, 1)
	apaMana(me, "W", "W")
	apaActivate(t, g, me, aesthir, 0, game.ActivateAbilityParams{})
	apaMana(me, "W", "W")
	if err := g.ActivateCatalogAbility(me.ID, aesthir, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("Wild Aesthir pumped twice in one turn")
	}
	me.ManaPool.EmptyPool()
	if p, tt := ptOf(t, g, aesthir); p != 3 || tt != 1 {
		t.Errorf("Wild Aesthir %d/%d, want 3/1", p, tt)
	}
}

// Four activations leave Yavimaya Ancients at -1 toughness, and it dies.
func TestAll5634YavimayaAncientsDiesOfItsOwnAbility(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	ancients := b12Push(g, me.ID, "Yavimaya Ancients", "Creature — Treefolk", alliYavimayaAncientsOracle, 2, 7)
	for range 3 {
		apaMana(me, "G")
		apaActivate(t, g, me, ancients, 0, game.ActivateAbilityParams{})
	}
	if p, tt := ptOf(t, g, ancients); p != 5 || tt != 1 {
		t.Fatalf("after three activations %d/%d, want 5/1", p, tt)
	}
	apaMana(me, "G")
	apaActivate(t, g, me, ancients, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(ancients) {
		t.Error("Yavimaya Ancients survived at -1 toughness")
	}
}

func TestAll5634EnslavedScoutGainsMountainwalkUntilEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	scout := b12Push(g, me.ID, "Enslaved Scout", "Creature — Goblin Scout", alliEnslavedScoutOracle, 2, 2)
	if hasEffectiveKeyword(t, g, scout, "mountainwalk") {
		t.Fatal("mountainwalk before the ability")
	}
	apaMana(me, "C", "C")
	apaActivate(t, g, me, scout, 0, game.ActivateAbilityParams{})
	if !hasEffectiveKeyword(t, g, scout, "mountainwalk") {
		t.Error("no mountainwalk after the ability")
	}
	advanceToMainOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	if hasEffectiveKeyword(t, g, scout, "mountainwalk") {
		t.Error("mountainwalk outlasted the turn")
	}
}

// --- the regenerators ----------------------------------------------------------

func TestAll5634RegenerateThisCreature(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
		row                    int
		mana                   []string
	}{
		{"Gorilla Chieftain", "Creature — Ape", alliGorillaChieftainOracle, 0, []string{"C", "G"}},
		{"Lim-Dûl's High Guard", "Creature — Skeleton", alliLimDulsHighGuardOracle, 0, []string{"C", "B"}},
		{"Soldevi Steam Beast", "Artifact Creature — Beast", alliSoldeviSteamBeastOracle, 0, []string{"C", "C"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := top100ActiveSeat(g)
			id := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 3, 3)
			apaMana(me, tc.mana...)
			apaActivate(t, g, me, id, tc.row, game.ActivateAbilityParams{})
			g.WithWriteLock(func() {
				if err := g.DestroyPermanentForEffect(id); err != nil {
					t.Fatal(err)
				}
			})
			if !g.Battlefield.Contains(id) {
				t.Fatal("the shielded creature was destroyed")
			}
			if !alliTapped(t, g, id) {
				t.Error("regeneration did not tap it")
			}
			// The Steam Beast's tap trigger may be waiting; settle it.
			if len(g.PendingChoices) > 0 {
				alliAnswerPlayerPick(t, g)
			}
			passPriorityAroundTable(t, g)
			g.WithWriteLock(func() {
				if err := g.DestroyPermanentForEffect(id); err != nil {
					t.Fatal(err)
				}
			})
			if g.Battlefield.Contains(id) {
				t.Error("one shield saved it twice")
			}
		})
	}
}

// Soldevi Steam Beast: becoming tapped gives the targeted opponent 2 life.
func TestAll5634SoldeviSteamBeastTapGivesAnOpponentLife(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	beast := b12Push(g, me.ID, "Soldevi Steam Beast", "Artifact Creature — Beast", alliSoldeviSteamBeastOracle, 4, 2)
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	if err := g.TapCard(beast, true); err != nil {
		t.Fatal(err)
	}
	picked := alliAnswerPlayerPick(t, g)
	if picked == me.ID {
		t.Fatal("the trigger offered its controller as an opponent")
	}
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		want := lives[p.ID]
		if p.ID == picked {
			want += 2
		}
		if p.Life != want {
			t.Errorf("%s life %d -> %d, want %d", p.Name, lives[p.ID], p.Life, want)
		}
	}
}

// --- mana ------------------------------------------------------------------------

func TestAll5634AgentOfStromgaldTurnsRedIntoBlack(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	agent := pushCatalogPermanent(g, me.ID, "Agent of Stromgald", "Creature — Human Knight", alliAgentOfStromgaldOracle, true)
	if err := g.ActivateManaAbility(me.ID, agent, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("the ability activated with no {R} to pay")
	}
	apaMana(me, "R", "R")
	for range 2 {
		if err := g.ActivateManaAbility(me.ID, agent, 0, game.ManaAbilityParams{}); err != nil {
			t.Fatal(err)
		}
	}
	if b := me.ManaPool.CountColor("B"); b != 2 || me.ManaPool.Total() != 2 {
		t.Errorf("pool holds %d {B} of %d, want 2 of 2", b, me.ManaPool.Total())
	}
	if alliTapped(t, g, agent) {
		t.Error("the ability tapped the Agent")
	}
}

func TestAll5634SchoolOfTheUnseenAddsWhatItPrints(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	land := pushCatalogPermanent(g, me.ID, "School of the Unseen", "Land", alliSchoolOfTheUnseenOracle, false)
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	if n := me.ManaPool.CountColor("C"); n != 1 || me.ManaPool.Total() != 1 {
		t.Errorf("{T} row: pool holds %d colorless of %d, want 1 of 1", n, me.ManaPool.Total())
	}

	for _, color := range []string{"W", "U", "B", "R", "G"} {
		g = newCatalogGame(t)
		advanceToMain(t, g)
		me = top100ActiveSeat(g)
		land = pushCatalogPermanent(g, me.ID, "School of the Unseen", "Land", alliSchoolOfTheUnseenOracle, false)
		if err := g.ActivateManaAbility(me.ID, land, 1, game.ManaAbilityParams{Colors: []string{color}}); err == nil {
			t.Fatal("the {2} row activated with nothing to pay {2}")
		}
		apaMana(me, "C", "C")
		if err := g.ActivateManaAbility(me.ID, land, 1, game.ManaAbilityParams{Colors: []string{color}}); err != nil {
			t.Fatal(err)
		}
		if n := me.ManaPool.CountColor(color); n != 1 || me.ManaPool.Total() != 1 {
			t.Errorf("{2} row picking %s: pool holds %d of %d, want 1 of 1", color, n, me.ManaPool.Total())
		}
		if !alliTapped(t, g, land) {
			t.Error("the {2} row did not tap the land")
		}
	}
}

// --- removal -----------------------------------------------------------------------

func TestAll5634PillageDestroysAnArtifactOrALandThroughRegeneration(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := castCatalogSpellErr(t, g, "Pillage", "Sorcery", alliPillageOracle, cardTarget(bear)); err == nil {
		t.Fatal("Pillage targeted a creature")
	}
	for _, typeLine := range []string{"Artifact", "Land"} {
		g := newCatalogGame(t)
		target := b12Permanent(g, g.Seats[1].ID, "Target", typeLine)
		p1Regenerate(t, g, target)
		castCatalogSpell(t, g, "Pillage", "Sorcery", alliPillageOracle, cardTarget(target))
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(target) {
			t.Errorf("the shielded %s survived Pillage", typeLine)
		}
	}
}

func TestAll5634ReprisalNeedsPowerFour(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	small := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	if err := castCatalogSpellErr(t, g, "Reprisal", "Instant", alliReprisalOracle, cardTarget(small)); err == nil {
		t.Fatal("Reprisal targeted a creature with power 3")
	}
	g = newCatalogGame(t)
	big := b12Creature(g, g.Seats[1].ID, "Giant", "Creature — Giant", 4, 4)
	p1Regenerate(t, g, big)
	castCatalogSpell(t, g, "Reprisal", "Instant", alliReprisalOracle, cardTarget(big))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("the shielded power-4 creature survived Reprisal")
	}
}

func TestAll5634ExileTakesANonwhiteAttackerAndGainsItsToughness(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	raider := b12Creature(g, me.ID, "Raider", "Creature — Orc", 2, 5)
	alliSetColors(g, raider, "R")
	knight := b12Creature(g, me.ID, "Knight", "Creature — Knight", 2, 2)
	alliSetColors(g, knight, "W")
	homebody := b12Creature(g, me.ID, "Homebody", "Creature — Bear", 2, 2)
	alliSetColors(g, homebody, "G")
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, a := range []uuid.UUID{raider, knight} {
		if err := g.DeclareAttacker(a, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
	for _, wrong := range []uuid.UUID{knight, homebody} {
		if err := alliCastInPlace(t, g, me.ID, "Exile", alliExileOracle, cardTarget(wrong)); err == nil {
			t.Fatal("Exile targeted a white attacker or a creature that is not attacking")
		}
	}
	life := me.Life
	if err := alliCastInPlace(t, g, me.ID, "Exile", alliExileOracle, cardTarget(raider)); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(raider) {
		t.Error("the attacker was not exiled")
	}
	if me.Life != life+5 {
		t.Errorf("life %d -> %d, want +5 (its toughness)", life, me.Life)
	}
}

func TestAll5634FeastOrFamine(t *testing.T) {
	// The first mode makes a 2/2 black Zombie.
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := g.Battlefield.Size()
	castModal(t, g, "Feast or Famine", "Instant", alliFeastOrFamineOracle, []int{0}, nil)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Size() != before+1 {
		t.Fatalf("battlefield %d -> %d, want one token", before, g.Battlefield.Size())
	}
	zombie := g.Battlefield.Cards[len(g.Battlefield.Cards)-1]
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Zombie" {
			zombie = c
		}
	}
	if zombie.Name != "Zombie" || zombie.Controller != me.ID || zombie.Power != 2 || zombie.Toughness != 2 ||
		!slices.Contains(zombie.Colors, "B") {
		t.Errorf("token %s %d/%d %v under %s, want my 2/2 black Zombie", zombie.Name, zombie.Power, zombie.Toughness, zombie.Colors, zombie.Controller)
	}

	// The second destroys a nonartifact, nonblack creature through a shield.
	g = newCatalogGame(t)
	opp := g.Seats[1]
	shade := b12Creature(g, opp.ID, "Shade", "Creature — Shade", 2, 2)
	alliSetColors(g, shade, "B")
	golem := b12Creature(g, opp.ID, "Golem", "Artifact Creature — Golem", 2, 2)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	alliSetColors(g, bear, "G")
	for _, wrong := range []uuid.UUID{shade, golem} {
		active := g.Seats[g.Turn.ActiveSeat]
		id := uuid.New()
		active.Hand.PushTop(game.Card{InstanceID: id, Name: "Feast or Famine", TypeLine: "Instant",
			OracleID: alliFeastOrFamineOracle, Owner: active.ID, Controller: active.ID})
		if err := g.CastSpell(active.ID, id, game.CastSpellParams{Modes: []int{1}, Targets: cardTarget(wrong)}); err == nil {
			t.Fatal("Feast or Famine targeted a black or an artifact creature")
		}
	}
	p1Regenerate(t, g, bear)
	castModal(t, g, "Feast or Famine", "Instant", alliFeastOrFamineOracle, []int{1}, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the shielded green creature survived")
	}
}

// --- the delayed draws --------------------------------------------------------------

// alliDrawsAtTheNextUpkeep asserts `p` draws one card at the next upkeep,
// and not before.
func alliDrawsAtTheNextUpkeep(t *testing.T, g *game.Game, p *game.Player) {
	t.Helper()
	before := p.Hand.Size()
	next := (g.Turn.ActiveSeat + 1) % len(g.Seats)
	advanceToUpkeepOf(t, g, next)
	if p.Hand.Size() != before {
		t.Fatalf("drew before the next upkeep's trigger resolved: hand %d -> %d", before, p.Hand.Size())
	}
	passPriorityAroundTable(t, g)
	if p.Hand.Size() != before+1 {
		t.Errorf("hand %d -> %d at the next upkeep, want +1", before, p.Hand.Size())
	}
}

func TestAll5634FeveredStrengthPumpsAndDrawsLater(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "A", "B", "C")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Fevered Strength", "Instant", alliFeveredStrengthOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if p, tt := ptOf(t, g, bear); p != 4 || tt != 2 {
		t.Errorf("target %d/%d, want 4/2", p, tt)
	}
	alliDrawsAtTheNextUpkeep(t, g, me)
	if p, tt := ptOf(t, g, bear); p != 2 || tt != 2 {
		t.Errorf("next turn %d/%d, want 2/2", p, tt)
	}
}

func TestAll5634CarrierPigeonsDrawsAtTheNextUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "A", "B", "C")
	castCatalogSpell(t, g, "Carrier Pigeons", "Creature — Bird", alliCarrierPigeonsOracle, nil)
	passPriorityAroundTable(t, g)
	alliDrawsAtTheNextUpkeep(t, g, me)
}

// alliOpponentCastsBolt has seat 1 cast a Lightning Bolt at seat 0 whose
// colours are `colors`, and returns it.
func alliOpponentCastsBolt(t *testing.T, g *game.Game, colors ...string) uuid.UUID {
	t.Helper()
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	id := uuid.New()
	opp.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Lightning Bolt", TypeLine: "Instant", OracleID: lightningBoltOracle,
		ManaCost: "{R}", Colors: colors, Owner: opp.ID, Controller: opp.ID,
	})
	if err := g.CastSpell(opp.ID, id, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}); err != nil {
		t.Fatalf("opponent casts Lightning Bolt: %v", err)
	}
	return id
}

func TestAll5634BurnoutCountersOnlyABlueInstant(t *testing.T) {
	// A red instant is a legal target and is not countered; Burnout still draws.
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "A", "B", "C")
	life := me.Life
	red := alliOpponentCastsBolt(t, g, "R")
	castCatalogSpell(t, g, "Burnout", "Instant", alliBurnoutOracle, cardTarget(red))
	passPriorityAroundTable(t, g)
	if me.Life != life-3 {
		t.Errorf("the red Bolt was countered: life %d -> %d", life, me.Life)
	}
	alliDrawsAtTheNextUpkeep(t, g, me)

	// A blue one is countered.
	g = newCatalogGame(t)
	me = g.Seats[0]
	seedLibrary(me, "A", "B", "C")
	life = me.Life
	blue := alliOpponentCastsBolt(t, g, "U")
	castCatalogSpell(t, g, "Burnout", "Instant", alliBurnoutOracle, cardTarget(blue))
	passPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("the blue Bolt resolved: life %d -> %d", life, me.Life)
	}
	alliDrawsAtTheNextUpkeep(t, g, me)
}

func TestAll5634BurnoutCantTargetASorcery(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Blue Sorcery", TypeLine: "Sorcery", ManaCost: "{U}", Colors: []string{"U"},
		Owner: me.ID, Controller: me.ID,
	})
	me.ManaPool.AddMana(game.ManaToken{Color: "U"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell Blue Sorcery: %v", err)
	}
	if err := castCatalogSpellErr(t, g, "Burnout", "Instant", alliBurnoutOracle, cardTarget(id)); err == nil {
		t.Error("Burnout targeted a sorcery spell")
	}
}

// --- the mass effects -----------------------------------------------------------------

func TestAll5634StenchOfDecayShrinksNonartifactCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	elf := b12Creature(g, opp.ID, "Elf", "Creature — Elf", 1, 1)
	golem := b12Creature(g, opp.ID, "Golem", "Artifact Creature — Golem", 2, 2)
	castCatalogSpell(t, g, "Stench of Decay", "Instant", alliStenchOfDecayOracle, nil)
	passPriorityAroundTable(t, g)
	if p, tt := ptOf(t, g, bear); p != 1 || tt != 1 {
		t.Errorf("Bear %d/%d, want 1/1", p, tt)
	}
	if g.Battlefield.Contains(elf) {
		t.Error("the 1/1 survived")
	}
	if p, tt := ptOf(t, g, golem); p != 2 || tt != 2 {
		t.Errorf("artifact creature %d/%d, want 2/2", p, tt)
	}
	advanceToMainOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	if p, tt := ptOf(t, g, bear); p != 2 || tt != 2 {
		t.Errorf("Bear next turn %d/%d, want 2/2", p, tt)
	}
}

func TestAll5634HailStormHitsAttackersAndYourSide(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := b12Creature(g, me.ID, "Attacker", "Creature — Ape", 4, 4)
	homebody := b12Creature(g, me.ID, "Homebody", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	life := me.Life
	if err := alliCastInPlace(t, g, me.ID, "Hail Storm", alliHailStormOracle, nil); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	damage := func(id uuid.UUID) int {
		c, ok := battlefieldCard(g, id)
		if !ok {
			t.Fatalf("%s left the battlefield", id)
		}
		return c.DamageMarked
	}
	if d := damage(attacker); d != 3 {
		t.Errorf("my attacker was dealt %d, want 3 (2 as an attacker, 1 as mine)", d)
	}
	if d := damage(homebody); d != 1 {
		t.Errorf("my creature at home was dealt %d, want 1", d)
	}
	if d := damage(theirs); d != 0 {
		t.Errorf("an opponent's creature at home was dealt %d, want 0", d)
	}
	if me.Life != life-1 {
		t.Errorf("my life %d -> %d, want -1", life, me.Life)
	}
}

// --- the statics ----------------------------------------------------------------------

func TestAll5634KaysaPumpsYourGreenCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kaysa := b12Push(g, me.ID, "Kaysa", "Legendary Creature — Elf Druid", alliKaysaOracle, 2, 3)
	alliSetColors(g, kaysa, "G")
	elf := b12Creature(g, me.ID, "Elf", "Creature — Elf", 1, 1)
	alliSetColors(g, elf, "G")
	knight := b12Creature(g, me.ID, "Knight", "Creature — Knight", 2, 2)
	alliSetColors(g, knight, "W")
	theirs := b12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1)
	alliSetColors(g, theirs, "G")
	for _, tc := range []struct {
		id   uuid.UUID
		p, t int
		what string
	}{
		{kaysa, 3, 4, "Kaysa"}, {elf, 2, 2, "my green creature"},
		{knight, 2, 2, "my white creature"}, {theirs, 1, 1, "an opponent's green creature"},
	} {
		if p, tt := ptOf(t, g, tc.id); p != tc.p || tt != tc.t {
			t.Errorf("%s is %d/%d, want %d/%d", tc.what, p, tt, tc.p, tc.t)
		}
	}
}

func TestAll5634JuniperOrderAdvocateWorksOnlyUntapped(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advocate := b12Push(g, me.ID, "Juniper Order Advocate", "Creature — Human Knight", alliJuniperOrderAdvocateOracle, 1, 2)
	alliSetColors(g, advocate, "W")
	elf := b12Creature(g, me.ID, "Elf", "Creature — Elf", 1, 1)
	alliSetColors(g, elf, "G")
	theirs := b12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1)
	alliSetColors(g, theirs, "G")
	if p, tt := ptOf(t, g, elf); p != 2 || tt != 2 {
		t.Errorf("untapped: my green creature %d/%d, want 2/2", p, tt)
	}
	if p, tt := ptOf(t, g, theirs); p != 1 || tt != 1 {
		t.Errorf("an opponent's green creature %d/%d, want 1/1", p, tt)
	}
	if p, tt := ptOf(t, g, advocate); p != 1 || tt != 2 {
		t.Errorf("the white Advocate %d/%d, want 1/2", p, tt)
	}
	if err := g.TapCard(advocate, true); err != nil {
		t.Fatal(err)
	}
	if p, tt := ptOf(t, g, elf); p != 1 || tt != 1 {
		t.Errorf("tapped: my green creature %d/%d, want 1/1", p, tt)
	}
	if err := g.TapCard(advocate, false); err != nil {
		t.Fatal(err)
	}
	if p, tt := ptOf(t, g, elf); p != 2 || tt != 2 {
		t.Errorf("untapped again: my green creature %d/%d, want 2/2", p, tt)
	}

	// An attack taps the Advocate too.
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(advocate, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	if !alliTapped(t, g, advocate) {
		t.Fatal("the attack did not tap the Advocate")
	}
	if p, tt := ptOf(t, g, elf); p != 1 || tt != 1 {
		t.Errorf("attacking: my green creature %d/%d, want 1/1", p, tt)
	}
}

func TestAll5634PhyrexianBoonDependsOnTheHostsColor(t *testing.T) {
	g := newCatalogGame(t)
	shade := b12Creature(g, g.Seats[0].ID, "Shade", "Creature — Shade", 2, 2)
	alliSetColors(g, shade, "B")
	auraCast(t, g, "Phyrexian Boon", alliPhyrexianBoonOracle, shade)
	if p, tt := ptOf(t, g, shade); p != 4 || tt != 3 {
		t.Errorf("black host %d/%d, want 4/3", p, tt)
	}

	g = newCatalogGame(t)
	bear := b12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 2, 3)
	alliSetColors(g, bear, "G")
	auraCast(t, g, "Phyrexian Boon", alliPhyrexianBoonOracle, bear)
	if p, tt := ptOf(t, g, bear); p != 1 || tt != 1 {
		t.Errorf("green host %d/%d, want 1/1", p, tt)
	}
}

func TestAll5634VisceridArmorPumpsAndGoesBackToHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	armor := auraCast(t, g, "Viscerid Armor", alliVisceridArmorOracle, bear)
	if p, tt := ptOf(t, g, bear); p != 3 || tt != 3 {
		t.Errorf("enchanted %d/%d, want 3/3", p, tt)
	}
	apaMana(me, "C", "U")
	apaActivate(t, g, me, armor, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(armor) {
		t.Fatal("the Aura is still on the battlefield")
	}
	inHand := false
	for _, c := range me.Hand.Cards {
		if c.InstanceID == armor || c.OracleID == alliVisceridArmorOracle {
			inHand = true
		}
	}
	if !inHand {
		t.Error("the Aura is not in its owner's hand")
	}
	if p, tt := ptOf(t, g, bear); p != 2 || tt != 2 {
		t.Errorf("unenchanted %d/%d, want 2/2", p, tt)
	}
}

func TestAll5634NobleSteedsGivesFirstStrike(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	steeds := b12Push(g, me.ID, "Noble Steeds", "Enchantment", alliNobleSteedsOracle, 0, 0)
	bear := b12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 2, 2)
	apaMana(me, "C", "W")
	apaActivate(t, g, me, steeds, 0, game.ActivateAbilityParams{Targets: cardTarget(bear)})
	if !hasEffectiveKeyword(t, g, bear, "first strike") {
		t.Error("the target has no first strike")
	}
	advanceToMainOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	if hasEffectiveKeyword(t, g, bear, "first strike") {
		t.Error("first strike outlasted the turn")
	}
}

func TestAll5634UnlikelyAllianceSkipsCombatants(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	alliance := b12Push(g, me.ID, "Unlikely Alliance", "Enchantment", alliUnlikelyAllianceOracle, 0, 0)
	attacker := b12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	homebody := b12Creature(g, me.ID, "Homebody", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	apaMana(me, "C", "W")
	if err := g.ActivateCatalogAbility(me.ID, alliance, 0, game.ActivateAbilityParams{Strict: true, Targets: cardTarget(attacker)}); err == nil {
		t.Fatal("Unlikely Alliance targeted an attacking creature")
	}
	me.ManaPool.EmptyPool()
	apaMana(me, "C", "W")
	apaActivate(t, g, me, alliance, 0, game.ActivateAbilityParams{Targets: cardTarget(homebody)})
	if p, tt := ptOf(t, g, homebody); p != 2 || tt != 4 {
		t.Errorf("target %d/%d, want 2/4", p, tt)
	}
}

// --- the triggered and tapped bodies ------------------------------------------------------

func TestAll5634ShieldSphereShrinksWhenItBlocks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := b12Creature(g, me.ID, "Attacker", "Creature — Bear", 0, 2)
	sphere := b12Push(g, opp.ID, "Shield Sphere", "Artifact Creature — Wall", alliShieldSphereOracle, 0, 6)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(sphere, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	lockInBlocks(t, g)
	passPriorityAroundTable(t, g)
	if p, tt := ptOf(t, g, sphere); p != 0 || tt != 5 {
		t.Errorf("after one block %d/%d, want 0/5", p, tt)
	}
}

func TestAll5634SoldierOfFortuneShufflesATargetLibrary(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	names := make([]string, 20)
	for i := range names {
		names[i] = string(rune('A' + i))
	}
	seedLibrary(opp, names...)
	order := func() []uuid.UUID {
		out := make([]uuid.UUID, 0, opp.Library.Size())
		for _, c := range opp.Library.Cards {
			out = append(out, c.InstanceID)
		}
		return out
	}
	before := order()
	soldier := pushCatalogPermanent(g, me.ID, "Soldier of Fortune", "Creature — Human Mercenary", alliSoldierOfFortuneOracle, false)
	apaMana(me, "R")
	apaActivate(t, g, me, soldier, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	if !alliTapped(t, g, soldier) {
		t.Error("the Soldier was not tapped to pay the cost")
	}
	after := order()
	if len(after) != len(before) {
		t.Fatalf("library size %d -> %d", len(before), len(after))
	}
	if slices.Equal(before, after) {
		t.Error("the target library is in the same order after the shuffle")
	}
}

// Yavimaya Ants: the first upkeep asks {G}{G}, the second {G}{G}{G}{G},
// and an unpaid upkeep sacrifices it.
func TestAll5634YavimayaAntsCumulativeUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	ants := b12Push(g, owner.ID, "Yavimaya Ants", "Creature — Insect", alliYavimayaAntsOracle, 5, 1)
	for turn := 1; turn <= 2; turn++ {
		if turn > 1 {
			if _, err := g.AdvanceStep(); err != nil {
				t.Fatal(err)
			}
		}
		advanceToUpkeepOf(t, g, 1)
		passPriorityAroundTable(t, g)
		if got := ageCountersOn(t, g, ants); got != turn {
			t.Fatalf("turn %d: age counters %d, want %d", turn, got, turn)
		}
		choice := remoraUpkeepPrompt(t, g, owner.ID)
		parsed, err := game.ParseCost(choice.PayCost)
		if err != nil {
			t.Fatal(err)
		}
		if len(parsed.Required) != 2*turn || parsed.Generic != 0 || strings.Count(choice.PayCost, "{G}") != 2*turn {
			t.Fatalf("turn %d: cost %q, want %d green", turn, choice.PayCost, 2*turn)
		}
		if turn == 2 {
			if err := g.ResolvePayUnless(choice.ID, owner.ID, false); err != nil {
				t.Fatal(err)
			}
			break
		}
		for range 2 * turn {
			owner.ManaPool.AddMana(game.ManaToken{Color: "G"})
		}
		if err := g.ResolvePayUnless(choice.ID, owner.ID, true); err != nil {
			t.Fatal(err)
		}
		if !g.Battlefield.Contains(ants) {
			t.Fatal("a paid upkeep sacrificed the Ants")
		}
	}
	if g.Battlefield.Contains(ants) {
		t.Error("the Ants survived an unpaid upkeep")
	}
}
