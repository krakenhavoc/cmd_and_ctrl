package testscm

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	cmdr2RamunapRuinsOracle = "d0d35864-1edc-4af1-9b89-3d7e94908011"
	cmdr2ShefetDunesOracle  = "8305715e-f711-47d6-8efe-d0efe4ced418"
	cmdr2HashepOasisOracle  = "eab70fff-6a9f-4f9f-89a2-b6910c199e46"
)

func cmdr2B13Desert(g *game.Game, owner uuid.UUID, name, oracle string) uuid.UUID {
	return effectstest.B12Push(g, owner, name, "Land — Desert", oracle, 0, 0)
}

// The three life-cost Deserts: {T}: {C}, {T} + 1 life: their colour.
func TestCmdr2B13LifeDesertsMakeColourlessAndPayLifeForColour(t *testing.T) {
	for _, tc := range []struct{ name, oracle, colour string }{
		{"Ramunap Ruins", cmdr2RamunapRuinsOracle, "R"},
		{"Shefet Dunes", cmdr2ShefetDunesOracle, "W"},
		{"Hashep Oasis", cmdr2HashepOasisOracle, "G"},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		land := cmdr2B13Desert(g, me.ID, tc.name, tc.oracle)
		before := me.Life
		if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
			t.Fatalf("%s: colourless: %v", tc.name, err)
		}
		if me.Life != before {
			t.Errorf("%s: the colourless ability costs no life", tc.name)
		}
		if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "C" {
			t.Errorf("%s: pool %v, want [C]", tc.name, got)
		}
		effectstest.B08Untap(g, land)
		if err := g.ActivateManaAbility(me.ID, land, 1, game.ManaAbilityParams{}); err != nil {
			t.Fatalf("%s: coloured: %v", tc.name, err)
		}
		if me.Life != before-1 {
			t.Errorf("%s: life %d → %d, want -1", tc.name, before, me.Life)
		}
		pool := effectstest.Batch01PoolColors(me)
		if len(pool) != 2 || pool[1] != tc.colour && pool[0] != tc.colour {
			t.Errorf("%s: pool %v lacks %s", tc.name, pool, tc.colour)
		}
	}
}

func TestCmdr2B13RamunapRuinsDamagesEachOpponentBySacrificingADesert(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ruins := cmdr2B13Desert(g, me.ID, "Ramunap Ruins", cmdr2RamunapRuinsOracle)
	other := cmdr2B13Desert(g, me.ID, "Desert", "cmdr2-b13-test-desert")
	effectstest.AdvanceToMain(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "R"}, game.ManaToken{Color: "R"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	before := map[uuid.UUID]int{}
	for _, s := range g.Seats[1:] {
		before[s.ID] = s.Life
	}
	if err := g.ActivateCatalogAbility(me.ID, ruins, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{other}}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.Seats) < 2 {
		t.Fatal("need opponents")
	}
	for _, s := range g.Seats[1:] {
		if s.Life != before[s.ID]-2 {
			t.Errorf("opponent life %d → %d, want -2", before[s.ID], s.Life)
		}
	}
	if g.Battlefield.Contains(other) || !g.Battlefield.Contains(ruins) {
		t.Error("the chosen Desert should have been sacrificed, not the Ruins")
	}
}

const (
	cmdr2DunesOfTheDeadOracle = "c761f71c-785c-4533-a2b7-2da3667688b8"
	cmdr2HostileDesertOracle  = "41459587-7509-404e-bd7d-fb8831dee789"
	cmdr2CactusPreserveOracle = "8da29533-f389-4bc2-ab9b-b469f893a362"
	cmdr2NeedleSpiresOracle   = "e7bb8160-0a4b-4e46-b196-7a19fb388d8e"
	cmdr2FaerieConclaveOracle = "0c25f6b1-8fb3-4406-9605-0282d2dbbcec"
	cmdr2LoamspeakerOracle    = "08377424-619e-48ae-bc3c-fb199cd77cf9"
)

func TestCmdr2B13ShefetDunesPumpsYourCreaturesOnlyAtSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	dunes := cmdr2B13Desert(g, me.ID, "Shefet Dunes", cmdr2ShefetDunesOracle)
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	if err := effectstest.Cmdr2B13Activate(t, g, me, dunes, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{dunes}}, "C", "C", "W", "W"); err != nil {
		t.Fatal(err)
	}
	if g.Battlefield.Contains(dunes) {
		t.Error("the Dunes sacrificed themselves as the Desert")
	}
	if p, tt := effectstest.EffectivePTOf(t, g, mine); p != 3 || tt != 3 {
		t.Errorf("my Bear %d/%d, want 3/3", p, tt)
	}
	if p, tt := effectstest.EffectivePTOf(t, g, theirs); p != 3 || tt != 3 {
		t.Errorf("their Ogre %d/%d, want unchanged 3/3", p, tt)
	}
	// Not at instant speed: a second Desert, in the declare-attackers step.
	again := cmdr2B13Desert(g, me.ID, "Shefet Dunes", cmdr2ShefetDunesOracle)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	me.ManaPool.EmptyPool()
	for _, c := range []string{"C", "C", "W", "W"} {
		effectstest.FillPoolColored(me, c, 1)
	}
	if err := g.ActivateCatalogAbility(me.ID, again, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{again}}); err == nil {
		t.Error("activated Shefet Dunes outside a main phase")
	}
}

func TestCmdr2B13HashepOasisPumpsTheTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	oasis := cmdr2B13Desert(g, me.ID, "Hashep Oasis", cmdr2HashepOasisOracle)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	params := game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{oasis},
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}
	if err := effectstest.Cmdr2B13Activate(t, g, me, oasis, params, "C", "G", "G"); err != nil {
		t.Fatal(err)
	}
	if p, tt := effectstest.EffectivePTOf(t, g, bear); p != 5 || tt != 5 {
		t.Errorf("Bear %d/%d, want 5/5", p, tt)
	}
	if g.Battlefield.Contains(oasis) {
		t.Error("the Oasis should be sacrificed")
	}
}

func TestCmdr2B13DunesOfTheDeadMakeAZombieWhenTheyDie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dunes := cmdr2B13Desert(g, me.ID, "Dunes of the Dead", cmdr2DunesOfTheDeadOracle)
	if err := g.ActivateManaAbility(me.ID, dunes, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for colourless: %v", err)
	}
	if effectstest.FindBattlefieldByName(g, "Zombie") != uuid.Nil {
		t.Fatal("a Zombie before the Dunes died")
	}
	effectstest.B15Destroy(t, g, dunes)
	z := effectstest.FindBattlefieldByName(g, "Zombie")
	if z == uuid.Nil {
		t.Fatal("no Zombie token")
	}
	if p, tt := effectstest.EffectivePTOf(t, g, z); p != 2 || tt != 2 {
		t.Errorf("Zombie %d/%d, want 2/2", p, tt)
	}
}

func TestCmdr2B13DunesOfTheDeadDoNotTriggerWhenBounced(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dunes := cmdr2B13Desert(g, me.ID, "Dunes of the Dead", cmdr2DunesOfTheDeadOracle)
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(dunes) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldByName(g, "Zombie") != uuid.Nil {
		t.Error("a bounce is not a graveyard move")
	}
}

func TestCmdr2B13HostileDesertExilesALandCardToBecomeAThreeFourElemental(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	desert := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Hostile Desert", TypeLine: "Land — Desert", OracleID: cmdr2HostileDesertOracle,
		Owner: me.ID, Controller: me.ID,
	})
	spell := uuid.New()
	land := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: spell, Name: "Shock", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
		me.Graveyard.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	})
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	if err := effectstest.Cmdr2B13Activate(t, g, me, desert, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{spell}}, "C", "C"); err == nil {
		t.Fatal("exiled a non-land card to pay the cost")
	}
	if err := effectstest.Cmdr2B13Activate(t, g, me, desert, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{land}}, "C", "C"); err != nil {
		t.Fatal(err)
	}
	if me.Graveyard.Contains(land) {
		t.Error("the land card should be exiled")
	}
	c := effectstest.S58p6View(t, g, desert)
	if !c.IsCreature() || !c.IsLand() || c.CurrentPower() != 3 || c.CurrentToughness() != 4 {
		t.Errorf("creature=%v land=%v %d/%d, want a 3/4 creature land", c.IsCreature(), c.IsLand(), c.CurrentPower(), c.CurrentToughness())
	}
	if !slices.Contains(c.Effective().Subtypes, "Elemental") {
		t.Errorf("subtypes %v, want Elemental", c.Effective().Subtypes)
	}
	if len(c.EffectiveColors()) != 0 {
		t.Errorf("colours %v, want colourless", c.EffectiveColors())
	}
}

func TestCmdr2B13NeedleSpiresAndFaerieConclaveAnimate(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, sub, kw string
		power, tough          int
		cost                  []string
		colours               []string
	}{
		{"Needle Spires", cmdr2NeedleSpiresOracle, "Elemental", "double strike", 2, 1, []string{"C", "C", "R", "W"}, []string{"R", "W"}},
		{"Faerie Conclave", cmdr2FaerieConclaveOracle, "Faerie", "flying", 2, 1, []string{"C", "U"}, []string{"U"}},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		land := effectstest.S58p6Push(g, me.ID, tc.name, "Land", tc.oracle, "", 0, 0)
		effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
		if effectstest.S58p6View(t, g, land).IsCreature() {
			t.Fatalf("%s: an idle land is a creature", tc.name)
		}
		if err := effectstest.Cmdr2B13Activate(t, g, me, land, game.ActivateAbilityParams{}, tc.cost...); err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		c := effectstest.S58p6View(t, g, land)
		if !c.IsCreature() || !c.IsLand() || c.CurrentPower() != tc.power || c.CurrentToughness() != tc.tough {
			t.Errorf("%s: creature=%v land=%v %d/%d", tc.name, c.IsCreature(), c.IsLand(), c.CurrentPower(), c.CurrentToughness())
		}
		if !slices.Contains(c.Effective().Subtypes, tc.sub) {
			t.Errorf("%s: subtypes %v, want %s", tc.name, c.Effective().Subtypes, tc.sub)
		}
		if !slices.Equal(c.EffectiveColors(), tc.colours) {
			t.Errorf("%s: colours %v, want %v", tc.name, c.EffectiveColors(), tc.colours)
		}
		effectstest.AssertEffectiveKeyword(t, g, land, tc.kw)
		if len(game.ManaAbilitiesForCard(c)) == 0 {
			t.Errorf("%s lost its mana ability", tc.name)
		}
	}
}

func TestCmdr2B13NeedleSpiresAndConclaveEnterTapped(t *testing.T) {
	for _, tc := range []struct{ name, oracle string }{
		{"Needle Spires", cmdr2NeedleSpiresOracle},
		{"Faerie Conclave", cmdr2FaerieConclaveOracle},
		{"Cactus Preserve", cmdr2CactusPreserveOracle},
	} {
		g := effectstest.NewCatalogGame(t)
		id := effectstest.B12PlayFromHand(t, g, tc.name, "Land", tc.oracle, game.CastSpellParams{})
		if !effectstest.IsTapped(g, id) {
			t.Errorf("%s entered untapped", tc.name)
		}
	}
}

func TestCmdr2B13CactusPreserveIsAsBigAsTheGreatestCommander(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() {
		me.Command.PushTop(game.Card{
			InstanceID: uuid.New(), Name: "Cmdr One", TypeLine: "Legendary Creature — Elf", ManaCost: "{2}{G}",
			Owner: me.ID, Controller: me.ID, IsCommander: true,
		})
		me.Command.PushTop(game.Card{
			InstanceID: uuid.New(), Name: "Cmdr Two", TypeLine: "Legendary Creature — Elf", ManaCost: "{4}{G}{G}",
			Owner: me.ID, Controller: me.ID, IsCommander: true,
		})
	})
	land := effectstest.S58p6Push(g, me.ID, "Cactus Preserve", "Land — Desert", cmdr2CactusPreserveOracle, "", 0, 0)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	if err := effectstest.Cmdr2B13Activate(t, g, me, land, game.ActivateAbilityParams{}, "C", "C", "C"); err != nil {
		t.Fatal(err)
	}
	c := effectstest.S58p6View(t, g, land)
	if !c.IsCreature() || c.CurrentPower() != 6 || c.CurrentToughness() != 6 {
		t.Errorf("creature=%v %d/%d, want 6/6", c.IsCreature(), c.CurrentPower(), c.CurrentToughness())
	}
	if !slices.Equal(c.EffectiveColors(), []string{"G"}) || !slices.Contains(c.Effective().Subtypes, "Plant") {
		t.Errorf("colours %v subtypes %v, want green Plant", c.EffectiveColors(), c.Effective().Subtypes)
	}
	effectstest.AssertEffectiveKeyword(t, g, land, "reach")
}

func TestCmdr2B13CactusPreserveMakesAnyTypeAnotherLandCouldProduce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	preserve := effectstest.SeedPermanentWithOracle(g, me.ID, "Cactus Preserve", "Land — Desert", cmdr2CactusPreserveOracle)
	effectstest.SeedManaLand(g, me.ID, "Mountain", "Basic Land — Mountain", "R")
	effectstest.SeedManaLand(g, me.ID, "Wastes", "Basic Land — Wastes", "C")
	effectstest.SeedManaLand(g, g.Seats[1].ID, "Forest", "Basic Land — Forest", "G")
	if err := g.ActivateManaAbility(me.ID, preserve, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := effectstest.ManaPickFor(g, me.ID)
	if pick == nil {
		t.Fatal("no colour pick queued")
	}
	got := map[string]bool{}
	for _, o := range pick.ColorOptions {
		got[o] = true
	}
	if len(got) != 2 || !got["R"] || !got["C"] {
		t.Errorf("options = %v, want exactly {R,C}: your lands only, colourless included", pick.ColorOptions)
	}
}

func TestCmdr2B13LoamspeakerAnimatesATargetLandAtSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	speaker := effectstest.B12Creature(g, me.ID, "Llanowar Loamspeaker", "Creature — Elf Druid", 1, 3)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardByID(g, speaker).OracleID = cmdr2LoamspeakerOracle })
	forest := effectstest.S58p6Push(g, me.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	theirs := effectstest.S58p6Push(g, opp.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	if err := g.ActivateCatalogAbility(me.ID, speaker, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}},
	}); err == nil {
		t.Error("targeted a land an opponent controls")
	}
	if err := g.ActivateCatalogAbility(me.ID, speaker, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: forest}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.S58p6View(t, g, forest)
	if !c.IsCreature() || !c.IsLand() || c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("creature=%v land=%v %d/%d, want a 3/3 creature land", c.IsCreature(), c.IsLand(), c.CurrentPower(), c.CurrentToughness())
	}
	if !slices.Contains(c.Effective().Subtypes, "Elemental") || !slices.Contains(c.Effective().Subtypes, "Forest") {
		t.Errorf("subtypes %v, want Elemental and still a Forest", c.Effective().Subtypes)
	}
	effectstest.AssertEffectiveKeyword(t, g, forest, "haste")
}

const (
	cmdr2SkolaOracle      = "0e3246ef-e0b4-49bc-a715-b69c4b6be461"
	cmdr2TurntimberOracle = "23d7d8ba-5237-4ffc-bd5f-b07ff9ec73af"
	cmdr2BehemothOracle   = "ee4b7aca-31dd-4369-8e85-b300892b9b41"
)

func cmdr2B13Mill(t *testing.T, g *game.Game, me *game.Player, name, typeLine string) {
	t.Helper()
	g.WithWriteLock(func() {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine, Owner: me.ID, Controller: me.ID})
	})
	g.WithWriteLock(func() {
		if err := g.MillNForEffect(me.ID, 1); err != nil {
			t.Errorf("mill: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
}

func TestCmdr2B13SkolaGrovedancerGainsLifeForEachLandCardMilledAndNotOthers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Skola Grovedancer", "Creature — Merfolk Druid", cmdr2SkolaOracle, 0, 3)
	before := me.Life
	cmdr2B13Mill(t, g, me, "Shock", "Instant")
	if me.Life != before {
		t.Errorf("life %d → %d for a nonland card", before, me.Life)
	}
	cmdr2B13Mill(t, g, me, "Forest", "Basic Land — Forest")
	if me.Life != before+1 {
		t.Errorf("life %d → %d, want +1 for a land card", before, me.Life)
	}
}

func TestCmdr2B13TurntimberSowerMakesAPlantForLandCardsAndReturnsALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	sower := effectstest.B12Creature(g, me.ID, "Turntimber Sower", "Creature — Elf Druid", 2, 1)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardByID(g, sower).OracleID = cmdr2TurntimberOracle })
	cmdr2B13Mill(t, g, me, "Forest", "Basic Land — Forest")
	if effectstest.FindBattlefieldByName(g, "Plant") == uuid.Nil {
		t.Fatal("no Plant token for a land card in the graveyard")
	}
	land := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: land, Name: "Island", TypeLine: "Basic Land — Island", Owner: me.ID, Controller: me.ID})
	})
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 1, 1)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 1, 1)
	c := effectstest.B12Creature(g, me.ID, "C", "Creature — Bear", 1, 1)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	params := game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{a, b, c},
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: land}},
	}
	if err := effectstest.Cmdr2B13Activate(t, g, me, sower, params, "G"); err != nil {
		t.Fatal(err)
	}
	if !me.Hand.Contains(land) {
		t.Error("the land card should be back in hand")
	}
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) || g.Battlefield.Contains(c) {
		t.Error("three creatures should have been sacrificed")
	}
}

func TestCmdr2B13PerennialBehemothPlaysLandsFromGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Perennial Behemoth", "Artifact Creature — Elemental", cmdr2BehemothOracle, 5, 3)
	land := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	})
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{FromZone: string(game.ZoneGraveyard)}); err != nil {
		t.Fatalf("PlayLand from graveyard: %v", err)
	}
	if !g.Battlefield.Contains(land) {
		t.Error("the land should be on the battlefield")
	}
}
