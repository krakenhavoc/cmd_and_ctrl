package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	cmdr2RamunapRuinsOracle = "d0d35864-1edc-4af1-9b89-3d7e94908011"
	cmdr2ShefetDunesOracle  = "8305715e-f711-47d6-8efe-d0efe4ced418"
	cmdr2HashepOasisOracle  = "eab70fff-6a9f-4f9f-89a2-b6910c199e46"
)

func cmdr2B13Desert(g *game.Game, owner uuid.UUID, name, oracle string) uuid.UUID {
	return b12Push(g, owner, name, "Land — Desert", oracle, 0, 0)
}

// The three life-cost Deserts: {T}: {C}, {T} + 1 life: their colour.
func TestCmdr2B13LifeDesertsMakeColourlessAndPayLifeForColour(t *testing.T) {
	for _, tc := range []struct{ name, oracle, colour string }{
		{"Ramunap Ruins", cmdr2RamunapRuinsOracle, "R"},
		{"Shefet Dunes", cmdr2ShefetDunesOracle, "W"},
		{"Hashep Oasis", cmdr2HashepOasisOracle, "G"},
	} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		land := cmdr2B13Desert(g, me.ID, tc.name, tc.oracle)
		before := me.Life
		if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
			t.Fatalf("%s: colourless: %v", tc.name, err)
		}
		if me.Life != before {
			t.Errorf("%s: the colourless ability costs no life", tc.name)
		}
		if got := batch01PoolColors(me); len(got) != 1 || got[0] != "C" {
			t.Errorf("%s: pool %v, want [C]", tc.name, got)
		}
		b08Untap(g, land)
		if err := g.ActivateManaAbility(me.ID, land, 1, game.ManaAbilityParams{}); err != nil {
			t.Fatalf("%s: coloured: %v", tc.name, err)
		}
		if me.Life != before-1 {
			t.Errorf("%s: life %d → %d, want -1", tc.name, before, me.Life)
		}
		pool := batch01PoolColors(me)
		if len(pool) != 2 || pool[1] != tc.colour && pool[0] != tc.colour {
			t.Errorf("%s: pool %v lacks %s", tc.name, pool, tc.colour)
		}
	}
}

func TestCmdr2B13RamunapRuinsDamagesEachOpponentBySacrificingADesert(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ruins := cmdr2B13Desert(g, me.ID, "Ramunap Ruins", cmdr2RamunapRuinsOracle)
	other := cmdr2B13Desert(g, me.ID, "Desert", "cmdr2-b13-test-desert")
	advanceToMain(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "R"}, game.ManaToken{Color: "R"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	before := map[uuid.UUID]int{}
	for _, s := range g.Seats[1:] {
		before[s.ID] = s.Life
	}
	if err := g.ActivateCatalogAbility(me.ID, ruins, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{other}}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g)
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

func cmdr2B13Activate(t *testing.T, g *game.Game, me *game.Player, src uuid.UUID, params game.ActivateAbilityParams, colours ...string) error {
	t.Helper()
	me.ManaPool.EmptyPool()
	for _, c := range colours {
		fillPoolColored(me, c, 1)
	}
	if err := g.ActivateCatalogAbility(me.ID, src, 0, params); err != nil {
		return err
	}
	passPriorityAroundTable(t, g)
	return nil
}

func TestCmdr2B13ShefetDunesPumpsYourCreaturesOnlyAtSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	dunes := cmdr2B13Desert(g, me.ID, "Shefet Dunes", cmdr2ShefetDunesOracle)
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	advanceTo(t, g, game.StepPrecombatMain)
	if err := cmdr2B13Activate(t, g, me, dunes, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{dunes}}, "C", "C", "W", "W"); err != nil {
		t.Fatal(err)
	}
	if g.Battlefield.Contains(dunes) {
		t.Error("the Dunes sacrificed themselves as the Desert")
	}
	if p, tt := effectivePTOf(t, g, mine); p != 3 || tt != 3 {
		t.Errorf("my Bear %d/%d, want 3/3", p, tt)
	}
	if p, tt := effectivePTOf(t, g, theirs); p != 3 || tt != 3 {
		t.Errorf("their Ogre %d/%d, want unchanged 3/3", p, tt)
	}
	// Not at instant speed: a second Desert, in the declare-attackers step.
	again := cmdr2B13Desert(g, me.ID, "Shefet Dunes", cmdr2ShefetDunesOracle)
	advanceTo(t, g, game.StepDeclareAttackers)
	me.ManaPool.EmptyPool()
	for _, c := range []string{"C", "C", "W", "W"} {
		fillPoolColored(me, c, 1)
	}
	if err := g.ActivateCatalogAbility(me.ID, again, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{again}}); err == nil {
		t.Error("activated Shefet Dunes outside a main phase")
	}
}

func TestCmdr2B13HashepOasisPumpsTheTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	oasis := cmdr2B13Desert(g, me.ID, "Hashep Oasis", cmdr2HashepOasisOracle)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepPrecombatMain)
	params := game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{oasis},
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}
	if err := cmdr2B13Activate(t, g, me, oasis, params, "C", "G", "G"); err != nil {
		t.Fatal(err)
	}
	if p, tt := effectivePTOf(t, g, bear); p != 5 || tt != 5 {
		t.Errorf("Bear %d/%d, want 5/5", p, tt)
	}
	if g.Battlefield.Contains(oasis) {
		t.Error("the Oasis should be sacrificed")
	}
}

func TestCmdr2B13DunesOfTheDeadMakeAZombieWhenTheyDie(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dunes := cmdr2B13Desert(g, me.ID, "Dunes of the Dead", cmdr2DunesOfTheDeadOracle)
	if err := g.ActivateManaAbility(me.ID, dunes, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for colourless: %v", err)
	}
	if findBattlefieldByName(g, "Zombie") != uuid.Nil {
		t.Fatal("a Zombie before the Dunes died")
	}
	b15Destroy(t, g, dunes)
	z := findBattlefieldByName(g, "Zombie")
	if z == uuid.Nil {
		t.Fatal("no Zombie token")
	}
	if p, tt := effectivePTOf(t, g, z); p != 2 || tt != 2 {
		t.Errorf("Zombie %d/%d, want 2/2", p, tt)
	}
}

func TestCmdr2B13DunesOfTheDeadDoNotTriggerWhenBounced(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dunes := cmdr2B13Desert(g, me.ID, "Dunes of the Dead", cmdr2DunesOfTheDeadOracle)
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(dunes) })
	passPriorityAroundTable(t, g)
	if findBattlefieldByName(g, "Zombie") != uuid.Nil {
		t.Error("a bounce is not a graveyard move")
	}
}

func TestCmdr2B13HostileDesertExilesALandCardToBecomeAThreeFourElemental(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	desert := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Hostile Desert", TypeLine: "Land — Desert", OracleID: cmdr2HostileDesertOracle,
		Owner: me.ID, Controller: me.ID,
	})
	spell := uuid.New()
	land := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: spell, Name: "Shock", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
		me.Graveyard.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	})
	advanceTo(t, g, game.StepPrecombatMain)
	if err := cmdr2B13Activate(t, g, me, desert, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{spell}}, "C", "C"); err == nil {
		t.Fatal("exiled a non-land card to pay the cost")
	}
	if err := cmdr2B13Activate(t, g, me, desert, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{land}}, "C", "C"); err != nil {
		t.Fatal(err)
	}
	if me.Graveyard.Contains(land) {
		t.Error("the land card should be exiled")
	}
	c := s58p6View(t, g, desert)
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
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		land := s58p6Push(g, me.ID, tc.name, "Land", tc.oracle, "", 0, 0)
		advanceTo(t, g, game.StepPrecombatMain)
		if s58p6View(t, g, land).IsCreature() {
			t.Fatalf("%s: an idle land is a creature", tc.name)
		}
		if err := cmdr2B13Activate(t, g, me, land, game.ActivateAbilityParams{}, tc.cost...); err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		c := s58p6View(t, g, land)
		if !c.IsCreature() || !c.IsLand() || c.CurrentPower() != tc.power || c.CurrentToughness() != tc.tough {
			t.Errorf("%s: creature=%v land=%v %d/%d", tc.name, c.IsCreature(), c.IsLand(), c.CurrentPower(), c.CurrentToughness())
		}
		if !slices.Contains(c.Effective().Subtypes, tc.sub) {
			t.Errorf("%s: subtypes %v, want %s", tc.name, c.Effective().Subtypes, tc.sub)
		}
		if !slices.Equal(c.EffectiveColors(), tc.colours) {
			t.Errorf("%s: colours %v, want %v", tc.name, c.EffectiveColors(), tc.colours)
		}
		assertEffectiveKeyword(t, g, land, tc.kw)
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
		g := newCatalogGame(t)
		id := b12PlayFromHand(t, g, tc.name, "Land", tc.oracle, game.CastSpellParams{})
		if !isTapped(g, id) {
			t.Errorf("%s entered untapped", tc.name)
		}
	}
}

func TestCmdr2B13CactusPreserveIsAsBigAsTheGreatestCommander(t *testing.T) {
	g := newCatalogGame(t)
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
	land := s58p6Push(g, me.ID, "Cactus Preserve", "Land — Desert", cmdr2CactusPreserveOracle, "", 0, 0)
	advanceTo(t, g, game.StepPrecombatMain)
	if err := cmdr2B13Activate(t, g, me, land, game.ActivateAbilityParams{}, "C", "C", "C"); err != nil {
		t.Fatal(err)
	}
	c := s58p6View(t, g, land)
	if !c.IsCreature() || c.CurrentPower() != 6 || c.CurrentToughness() != 6 {
		t.Errorf("creature=%v %d/%d, want 6/6", c.IsCreature(), c.CurrentPower(), c.CurrentToughness())
	}
	if !slices.Equal(c.EffectiveColors(), []string{"G"}) || !slices.Contains(c.Effective().Subtypes, "Plant") {
		t.Errorf("colours %v subtypes %v, want green Plant", c.EffectiveColors(), c.Effective().Subtypes)
	}
	assertEffectiveKeyword(t, g, land, "reach")
}

func TestCmdr2B13CactusPreserveMakesAnyTypeAnotherLandCouldProduce(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	preserve := seedPermanentWithOracle(g, me.ID, "Cactus Preserve", "Land — Desert", cmdr2CactusPreserveOracle)
	seedManaLand(g, me.ID, "Mountain", "Basic Land — Mountain", "R")
	seedManaLand(g, me.ID, "Wastes", "Basic Land — Wastes", "C")
	seedManaLand(g, g.Seats[1].ID, "Forest", "Basic Land — Forest", "G")
	if err := g.ActivateManaAbility(me.ID, preserve, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := manaPickFor(g, me.ID)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	speaker := b12Creature(g, me.ID, "Llanowar Loamspeaker", "Creature — Elf Druid", 1, 3)
	g.WithWriteLock(func() { findBattlefieldCardByID(g, speaker).OracleID = cmdr2LoamspeakerOracle })
	forest := s58p6Push(g, me.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	theirs := s58p6Push(g, opp.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	advanceTo(t, g, game.StepPrecombatMain)
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
	passPriorityAroundTable(t, g)
	c := s58p6View(t, g, forest)
	if !c.IsCreature() || !c.IsLand() || c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("creature=%v land=%v %d/%d, want a 3/3 creature land", c.IsCreature(), c.IsLand(), c.CurrentPower(), c.CurrentToughness())
	}
	if !slices.Contains(c.Effective().Subtypes, "Elemental") || !slices.Contains(c.Effective().Subtypes, "Forest") {
		t.Errorf("subtypes %v, want Elemental and still a Forest", c.Effective().Subtypes)
	}
	assertEffectiveKeyword(t, g, forest, "haste")
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
	if err := g.MillNForEffect(me.ID, 1); err != nil {
		t.Fatalf("mill: %v", err)
	}
	passPriorityAroundTable(t, g)
}

func TestCmdr2B13SkolaGrovedancerGainsLifeForEachLandCardMilledAndNotOthers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Skola Grovedancer", "Creature — Merfolk Druid", cmdr2SkolaOracle, 0, 3)
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	sower := b12Creature(g, me.ID, "Turntimber Sower", "Creature — Elf Druid", 2, 1)
	g.WithWriteLock(func() { findBattlefieldCardByID(g, sower).OracleID = cmdr2TurntimberOracle })
	cmdr2B13Mill(t, g, me, "Forest", "Basic Land — Forest")
	if findBattlefieldByName(g, "Plant") == uuid.Nil {
		t.Fatal("no Plant token for a land card in the graveyard")
	}
	land := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: land, Name: "Island", TypeLine: "Basic Land — Island", Owner: me.ID, Controller: me.ID})
	})
	a := b12Creature(g, me.ID, "A", "Creature — Bear", 1, 1)
	b := b12Creature(g, me.ID, "B", "Creature — Bear", 1, 1)
	c := b12Creature(g, me.ID, "C", "Creature — Bear", 1, 1)
	advanceTo(t, g, game.StepPrecombatMain)
	params := game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{a, b, c},
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: land}},
	}
	if err := cmdr2B13Activate(t, g, me, sower, params, "G"); err != nil {
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Perennial Behemoth", "Artifact Creature — Elemental", cmdr2BehemothOracle, 5, 3)
	land := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	})
	advanceTo(t, g, game.StepPrecombatMain)
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{FromZone: string(game.ZoneGraveyard)}); err != nil {
		t.Fatalf("PlayLand from graveyard: %v", err)
	}
	if !g.Battlefield.Contains(land) {
		t.Error("the land should be on the battlefield")
	}
}
