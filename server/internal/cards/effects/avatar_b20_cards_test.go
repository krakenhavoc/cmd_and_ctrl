package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avatar_b20_cards_test.go — Avatar: The Last Airbender, batches 20 and 21
// (#4459). Each test plays the real catalog entry and asserts what its text
// says.

const (
	avB20PlanetariumOracle   = "f89d093d-c815-4df7-a348-2005aa227586"
	avB20BoomerangOracle     = "3c032e74-69c6-48ca-95d4-51aa3d894c14"
	avB20EngineRoomOracle    = "01ed9139-293d-4477-a145-3c6fe59266c7"
	avB20RealmOfKohOracle    = "bb9ce416-eef1-49e8-89a0-2b6837505070"
	avB20RumbleArenaOracle   = "83f2470e-2114-4cd9-ab1a-21b76e9cdf63"
	avB20SecretTunnelOracle  = "632e2979-d88a-482e-9bb8-57b683c5310f"
	avB20WhiteLotusOracle    = "2cdbfda3-98fc-4108-b551-c7049168924e"
	avB20AzulaOracle         = "7c9e351d-2e6c-4050-9c3d-b15bb9255a4d"
	avB20DaiLiAgentsOracle   = "8c05e716-b6eb-4f95-866b-7e43a898ebe2"
	avB20OzaiOracle          = "7472af8e-2e93-43b2-bc32-a3741f6d7502"
	avB20LessonSpellTypeLine = "Sorcery — Lesson"
)

// avB20CyclingDuals is the ten-card dual cycle: oracle id, name, colours.
var avB20CyclingDuals = []struct{ oracle, name, a, b string }{
	{"01ed9139-293d-4477-a145-3c6fe59266c7", "Airship Engine Room", "U", "R"},
	{"6caabc45-4644-4941-a2f3-f04ee7b50a89", "Boiling Rock Prison", "B", "R"},
	{"b56bb8cb-457e-4b85-9bf9-5e891e79f1dd", "Foggy Bottom Swamp", "B", "G"},
	{"da81b78b-89fd-4fc1-9395-b5b01b96be70", "Kyoshi Village", "G", "W"},
	{"66208129-20fb-4cb4-8f00-8d4574c9c653", "Meditation Pools", "G", "U"},
	{"dfd2c57a-4557-4df1-8f6f-da2cbd317f12", "Misty Palms Oasis", "W", "B"},
	{"3fcd1432-66d0-40e1-85ef-b57a6fa94b7c", "North Pole Gates", "W", "U"},
	{"6f9a4bb2-5959-4a5c-83af-c2729e6673df", "Omashu City", "R", "G"},
	{"f715f701-a735-42ab-b31a-8e1bd04ac5ff", "Serpent's Pass", "U", "B"},
	{"d188a397-c7d9-4f16-bc21-d9e1b5ffc951", "Sun-Blessed Peak", "R", "W"},
}

func TestAvB20EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		avB20PlanetariumOracle:  "Planetarium of Wan Shi Tong",
		avB20BoomerangOracle:    "Trusty Boomerang",
		avB20RealmOfKohOracle:   "Realm of Koh",
		avB20RumbleArenaOracle:  "Rumble Arena",
		avB20SecretTunnelOracle: "Secret Tunnel",
		avB20WhiteLotusOracle:   "White Lotus Hideout",
		avB20AzulaOracle:        "Azula, Cunning Usurper",
		avB20DaiLiAgentsOracle:  "Dai Li Agents",
		avB20OzaiOracle:         "Ozai, the Phoenix King",
	}
	for _, l := range avB20CyclingDuals {
		want[l.oracle] = l.name
	}
	if len(want) != 19 {
		t.Fatalf("the batch ships 19 cards, the table lists %d", len(want))
	}
	for oracle, name := range want {
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

// --- fixtures ----------------------------------------------------------------

// avB20Creature puts a vanilla creature onto the battlefield.
func avB20Creature(g *game.Game, owner uuid.UUID, name, typeLine string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// avB20LibraryTop puts a card on top of p's library.
func avB20LibraryTop(g *game.Game, p *game.Player, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	g.WithWriteLock(func() { p.Library.PushTop(c) })
	return c.InstanceID
}

// avB20AnswerTwoClauses answers a two-clause trigger's target prompts, one
// clause at a time.
func avB20AnswerTwoClauses(t *testing.T, g *game.Game, chooser, first, second uuid.UUID) {
	t.Helper()
	pick := latestPickTarget(g, chooser)
	if pick == nil {
		t.Fatal("no target prompt for the first clause")
	}
	if err := g.ResolvePickTarget(pick.ID, chooser, game.TargetRef{Kind: game.TargetCard, ID: first}); err != nil {
		t.Fatalf("pick the first land: %v", err)
	}
	pick = latestPickTarget(g, chooser)
	if pick == nil {
		t.Fatal("no target prompt for the second clause")
	}
	if err := g.ResolvePickTargets(pick.ID, chooser, []game.TargetRef{{Kind: game.TargetCard, ID: second, Slot: 1}}); err != nil {
		t.Fatalf("pick the second land: %v", err)
	}
}

// --- the dual cycle ----------------------------------------------------------

func TestAvB20CyclingDualsEnterTappedAndTapForTheirTwoColours(t *testing.T) {
	for _, l := range avB20CyclingDuals {
		spec, _ := Lookup(l.oracle)
		if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{"+l.a+"|"+l.b+"}" {
			t.Errorf("%s: mana abilities %+v, want one producing {%s} or {%s}", l.name, spec.ManaAbilities, l.a, l.b)
		}
		g := newCatalogGame(t)
		id := playLandFromHand(t, g, l.name, l.oracle)
		top100AssertEnteredTapped(t, g, id, l.name)
	}
}

func TestAvB20CyclingDualSacrificesForFourToDrawACard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	land := pushCatalogPermanent(g, me.ID, "Airship Engine Room", "Land", avB20EngineRoomOracle, false)
	avB20LibraryTop(g, me, game.Card{Name: "Library Card", TypeLine: "Sorcery"})
	b09AddColorless(me, 3)
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated for three mana; the cost is {4}")
	}
	b09AddColorless(me, 1)
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !inGraveyard(me, land) {
		t.Fatal("the land is sacrificed as a cost")
	}
	if me.Hand.Size() != hand {
		t.Fatal("the draw must wait for the ability to resolve")
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn", hand, me.Hand.Size())
	}
}

// --- Realm of Koh ------------------------------------------------------------

func TestAvB20RealmOfKohEntersTappedWithoutABasic(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLandOnBattlefield(g, me.ID, "Command Tower", "Land")
	tapped := playLandFromHand(t, g, "Realm of Koh", avB20RealmOfKohOracle)
	top100AssertEnteredTapped(t, g, tapped, "Realm of Koh without a basic")

	seedLandOnBattlefield(g, me.ID, "Swamp", "Basic Land — Swamp")
	untapped := playLandFromHand(t, g, "Realm of Koh", avB20RealmOfKohOracle)
	top100AssertEnteredUntapped(t, g, untapped, "Realm of Koh with a basic Swamp")
}

func TestAvB20RealmOfKohMakesASpiritThatOnlySpiritsCanBlock(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	realm := pushCatalogPermanent(g, me.ID, "Realm of Koh", "Land", avB20RealmOfKohOracle, false)
	b09AddColorless(me, 3)
	if err := g.ActivateCatalogAbility(me.ID, realm, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without {B}")
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "B"})
	if err := g.ActivateCatalogAbility(me.ID, realm, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	var spirit *game.Card
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Name == "Spirit" && c.IsToken() && c.Controller == me.ID {
			spirit = c
		}
	}
	if spirit == nil {
		t.Fatal("no Spirit token")
	}
	if spirit.Power != 1 || spirit.Toughness != 1 || len(spirit.Colors) != 0 || !spirit.HasSubtype("Spirit") {
		t.Errorf("token = %d/%d colours %v type %q, want a 1/1 colourless Spirit", spirit.Power, spirit.Toughness, spirit.Colors, spirit.TypeLine)
	}
	want := kurukSpiritToken()
	if spirit.OracleID != want.OracleID {
		t.Errorf("the token is %q, want Avatar Kuruk's printed Spirit %q", spirit.OracleID, want.OracleID)
	}
}

// --- Rumble Arena ------------------------------------------------------------

func TestAvB20RumbleArenaScriesOneAndHasVigilance(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	avB20LibraryTop(g, me, game.Card{Name: "Library Card", TypeLine: "Sorcery"})
	arena := playLandFromHand(t, g, "Rumble Arena", avB20RumbleArenaOracle)
	top100AssertEnteredUntapped(t, g, arena, "Rumble Arena")
	passPriorityAroundTable(t, g)
	c := scryChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 1 {
		t.Fatalf("Rumble Arena entering asks scry 1, got %+v", c)
	}
	answerScryKeepAll(t, g, me.ID)
	if !hasEffectiveKeyword(t, g, arena, "vigilance") {
		t.Error("Rumble Arena has vigilance")
	}
	spec, _ := Lookup(avB20RumbleArenaOracle)
	if len(spec.ManaAbilities) != 2 || spec.ManaAbilities[0].Produced != "{C}" ||
		spec.ManaAbilities[1].Cost.Mana != "{1}" || spec.ManaAbilities[1].Produced != "{W|U|B|R|G}" {
		t.Errorf("mana abilities %+v, want {T}: {C} and {1}, {T}: any colour", spec.ManaAbilities)
	}
}

// --- White Lotus Hideout ------------------------------------------------------

// avB20HideoutRestrictedBlue taps White Lotus Hideout's restricted ability
// for blue.
func avB20HideoutRestrictedBlue(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	hideout := pushCatalogPermanent(g, me.ID, "White Lotus Hideout", "Land", avB20WhiteLotusOracle, false)
	if err := g.ActivateManaAbility(me.ID, hideout, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := riderLatestManaPick(g, me.ID)
	if pick == nil {
		t.Fatal("the restricted ability asks for a colour")
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "U"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
}

func TestAvB20WhiteLotusHideoutManaCastsALesson(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	avB20HideoutRestrictedBlue(t, g, me)
	if _, err := castFromHandErr(g, me, "A Lesson", avB20LessonSpellTypeLine, "{U}", "", game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the Hideout's mana pays for a Lesson: %v", err)
	}
}

func TestAvB20WhiteLotusHideoutManaCastsAShrine(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	avB20HideoutRestrictedBlue(t, g, me)
	if _, err := castFromHandErr(g, me, "A Shrine", "Enchantment — Shrine", "{U}", "", game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the Hideout's mana pays for a Shrine: %v", err)
	}
}

func TestAvB20WhiteLotusHideoutManaRefusesOtherSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	avB20HideoutRestrictedBlue(t, g, me)
	if _, err := castFromHandErr(g, me, "A Sorcery", "Sorcery", "{U}", "", game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("the Hideout's restricted mana paid for a spell that is neither a Lesson nor a Shrine")
	}
}

// --- Secret Tunnel -----------------------------------------------------------

func TestAvB20SecretTunnelMakesTwoCreaturesThatShareATypeUnblockable(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	tunnel := pushCatalogPermanent(g, me.ID, "Secret Tunnel", "Land — Cave", avB20SecretTunnelOracle, false)
	assertRestrictions(t, g, tunnel, game.CantBeBlocked)
	aang := avB20Creature(g, me.ID, "Aang", "Creature — Human Avatar Ally", 2, 2)
	momo := avB20Creature(g, me.ID, "Momo", "Creature — Lemur Bat Ally", 1, 1)
	b09AddColorless(me, 4)
	if err := g.ActivateCatalogAbility(me.ID, tunnel, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: aang, Slot: 0}, {Kind: game.TargetCard, ID: momo, Slot: 1},
	}}); err != nil {
		t.Fatalf("activate on two Allies: %v", err)
	}
	passPriorityAroundTable(t, g)
	assertRestrictions(t, g, aang, game.CantBeBlocked)
	assertRestrictions(t, g, momo, game.CantBeBlocked)
}

func TestAvB20SecretTunnelRefusesCreaturesThatShareNoType(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	tunnel := pushCatalogPermanent(g, me.ID, "Secret Tunnel", "Land — Cave", avB20SecretTunnelOracle, false)
	aang := avB20Creature(g, me.ID, "Aang", "Creature — Human Avatar Ally", 2, 2)
	bear := avB20Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b09AddColorless(me, 4)
	if err := g.ActivateCatalogAbility(me.ID, tunnel, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: aang, Slot: 0}, {Kind: game.TargetCard, ID: bear, Slot: 1},
	}}); err == nil {
		t.Fatal("activated on an Ally and a Bear, which share no creature type")
	}
}

func TestAvB20SecretTunnelKeepsTheSecondCreatureWhenTheFirstLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	tunnel := pushCatalogPermanent(g, me.ID, "Secret Tunnel", "Land — Cave", avB20SecretTunnelOracle, false)
	sokka := avB20Creature(g, me.ID, "Sokka", "Creature — Human Warrior Ally", 2, 2)
	suki := avB20Creature(g, me.ID, "Suki", "Creature — Human Warrior Ally", 2, 2)
	b09AddColorless(me, 4)
	if err := g.ActivateCatalogAbility(me.ID, tunnel, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: sokka, Slot: 0}, {Kind: game.TargetCard, ID: suki, Slot: 1},
	}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(sokka); err != nil {
			t.Fatalf("destroy the first target: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	assertRestrictions(t, g, suki, game.CantBeBlocked)
}

// --- Trusty Boomerang --------------------------------------------------------

func TestAvB20TrustyBoomerangTapsATargetAndComesBackToHand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sokka := b16Creature(g, me.ID, "Sokka", "Creature — Human Warrior Ally", 2, 2, "W")
	boomerang := b21Push(g, me.ID, "Trusty Boomerang", "Artifact — Equipment", avB20BoomerangOracle, 0, 0)
	target := b16Creature(g, opp.ID, "Rhino", "Creature — Rhino", 4, 4, "R")
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}")
	equipTo(t, g, me.ID, boomerang, sokka)
	idx, ref := grantedActivatedIndex(t, g, sokka)
	if idx < 0 {
		t.Fatal("the equipped creature has the Boomerang's ability")
	}
	if err := g.ActivateCatalogAbility(me.ID, sokka, idx, game.ActivateAbilityParams{
		Ref: ref, Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
	}); err == nil {
		t.Fatal("activated with no mana; the cost is {1}")
	}
	floatMana(t, g, me, "{C}")
	b16Activate(t, g, me.ID, sokka, idx, game.ActivateAbilityParams{
		Ref: ref, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
	})
	passPriorityAroundTable(t, g)
	if !b16Tapped(t, g, sokka) {
		t.Error("the equipped creature taps to throw")
	}
	if !b16Tapped(t, g, target) {
		t.Error("the target creature is tapped")
	}
	if !me.Hand.Contains(boomerang) {
		t.Error("Trusty Boomerang returns to its owner's hand")
	}
}

// --- Dai Li Agents -----------------------------------------------------------

func TestAvB20DaiLiAgentsEarthbendsTwiceThenDrainsPerCounteredCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	forest := seedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	swamp := seedLandOnBattlefield(g, me.ID, "Swamp", "Basic Land — Swamp")
	floatForTest(g, me, "BGCCC")
	agents := castFromHandForTest(t, g, me, "Dai Li Agents", "Creature — Human Soldier", "{3}{B}{G}", avB20DaiLiAgentsOracle,
		game.CastSpellParams{Strict: true})
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(agents) {
		t.Fatal("Dai Li Agents did not resolve")
	}
	avB20AnswerTwoClauses(t, g, me.ID, forest, swamp)
	passPriorityAroundTable(t, g)
	for _, land := range []uuid.UUID{forest, swamp} {
		if got := countersOn(g, land, "+1/+1"); got != 1 {
			t.Errorf("an earthbent land has %d +1/+1 counters, want 1", got)
		}
		if !hasEffectiveKeyword(t, g, land, "haste") {
			t.Error("an earthbent land has haste")
		}
	}

	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == agents {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	oppBefore, meBefore := opp.Life, me.Life
	declareAttack(t, g, opp.ID, agents)
	passPriorityAroundTable(t, g)
	if opp.Life != oppBefore-2 || me.Life != meBefore+2 {
		t.Errorf("drain: opponent %d → %d, me %d → %d; want X = 2 (the two earthbent lands)", oppBefore, opp.Life, meBefore, me.Life)
	}
}

func TestAvB20DaiLiAgentsMayEarthbendTheSameLandTwice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	forest := seedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	floatForTest(g, me, "BGCCC")
	castFromHandForTest(t, g, me, "Dai Li Agents", "Creature — Human Soldier", "{3}{B}{G}", avB20DaiLiAgentsOracle,
		game.CastSpellParams{Strict: true})
	passPriorityAroundTable(t, g)
	avB20AnswerTwoClauses(t, g, me.ID, forest, forest)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, forest, "+1/+1"); got != 2 {
		t.Errorf("the land earthbent twice has %d +1/+1 counters, want 2", got)
	}
}

// --- Planetarium of Wan Shi Tong ---------------------------------------------

func TestAvB20PlanetariumCastsTheTopCardFreeOnceEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	planetarium := pushCatalogPermanent(g, me.ID, "Planetarium of Wan Shi Tong", "Legendary Artifact", avB20PlanetariumOracle, false)
	second := avB20LibraryTop(g, me, game.Card{Name: "Second Wurm", TypeLine: "Creature — Wurm", ManaCost: "{6}{G}", Power: 6, Toughness: 6})
	wurm := avB20LibraryTop(g, me, game.Card{Name: "Big Wurm", TypeLine: "Creature — Wurm", ManaCost: "{7}{G}", Power: 7, Toughness: 7})
	avB20LibraryTop(g, me, game.Card{Name: "Filler A", TypeLine: "Sorcery", ManaCost: "{R}"})

	b09AddColorless(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, planetarium, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate the scry: %v", err)
	}
	passPriorityUntilChoice(t, g)
	c := scryChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 2 {
		t.Fatalf("{1}, {T}: scry 2, got %+v", c)
	}
	// Bottom the filler, keep the Big Wurm on top.
	var filler uuid.UUID
	for _, id := range c.ScryCards {
		if id != wurm {
			filler = id
		}
	}
	if err := g.ResolveScry(c.ID, me.ID, []uuid.UUID{filler}, []uuid.UUID{wurm}); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	passPriorityUntilChoice(t, g)
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, wurm)
	if err := g.CastSpell(me.ID, wurm, game.CastSpellParams{FromZone: "library"}); err != nil {
		t.Fatalf("free cast of the top card: %v", err)
	}
	castInsideTheResolution(t, g, wurm)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(wurm) {
		t.Fatal("the free spell resolves")
	}

	// A second scry this turn does not trigger it again.
	untapForTest(g, planetarium)
	b09AddColorless(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, planetarium, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate the scry again: %v", err)
	}
	passPriorityUntilChoice(t, g)
	answerScryKeepAll(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if resolutionCastOfferFor(g, me.ID) != nil {
		t.Error("the Planetarium offered a second free cast in the same turn")
	}
	if libraryIndex(me, second) < 0 {
		t.Error("the second card left the library")
	}
}

func TestAvB20PlanetariumDeclinedLeavesTheTurnsUse(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	planetarium := pushCatalogPermanent(g, me.ID, "Planetarium of Wan Shi Tong", "Legendary Artifact", avB20PlanetariumOracle, false)
	avB20LibraryTop(g, me, game.Card{Name: "Filler", TypeLine: "Sorcery", ManaCost: "{R}"})
	top := avB20LibraryTop(g, me, game.Card{Name: "Cheap Sorcery", TypeLine: "Sorcery", ManaCost: "{1}{R}"})
	for round := 0; round < 2; round++ {
		untapForTest(g, planetarium)
		b09AddColorless(me, 1)
		if err := g.ActivateCatalogAbility(me.ID, planetarium, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("activate the scry: %v", err)
		}
		passPriorityUntilChoice(t, g)
		answerScryKeepAll(t, g, me.ID)
		passPriorityUntilChoice(t, g)
		mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, top)
		declineResolutionCast(t, g, me.ID)
		if libraryIndex(me, top) != len(me.Library.Cards)-1 {
			t.Fatalf("round %d: a declined card stays on top of the library", round)
		}
	}
}

// --- Azula, Cunning Usurper --------------------------------------------------

type avB20AzulaTable struct {
	g          *game.Game
	me, opp    *game.Player
	azula      uuid.UUID
	rhino      uuid.UUID
	instant    uuid.UUID
	keptToken  uuid.UUID
	keptLand   uuid.UUID
	exiledFrom game.PermissionCardRef
}

// avB20AzulaEnters casts Azula and answers her enters trigger: the opponent
// exiles their Rhino, then the instant in their graveyard.
func avB20AzulaEnters(t *testing.T) avB20AzulaTable {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	a := avB20AzulaTable{g: g, me: me, opp: opp}
	a.rhino = avB20Creature(g, opp.ID, "Rhino", "Creature — Rhino", 4, 4)
	a.keptToken = pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Soldier", TypeLine: "Token Creature — Soldier",
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
	a.instant = pushGraveyardTyped(opp, "Shock", "Instant")
	g.WithWriteLock(func() {
		for i := range opp.Graveyard.Cards {
			if opp.Graveyard.Cards[i].InstanceID == a.instant {
				opp.Graveyard.Cards[i].ManaCost = "{R}"
			}
		}
	})
	a.keptLand = pushGraveyardTyped(opp, "Mountain", "Basic Land — Mountain")
	advanceToMain(t, g)
	floatForTest(g, me, "UBBCC")
	a.azula = castFromHandForTest(t, g, me, "Azula, Cunning Usurper", "Legendary Creature — Human Noble Rogue", "{2}{U}{B}{B}",
		avB20AzulaOracle, game.CastSpellParams{Strict: true})
	passPriorityAroundTable(t, g)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityUntilChoice(t, g)

	pick := chooseCardsChoiceFor(g, opp.ID)
	if pick == nil {
		t.Fatal("the opponent is asked to exile a creature")
	}
	if hasID(pick.ChooseCards, a.keptToken) || !hasID(pick.ChooseCards, a.rhino) {
		t.Fatalf("creature options %v: the nontoken Rhino, never the token", pick.ChooseCards)
	}
	if err := g.ResolveChooseCards(pick.ID, opp.ID, []uuid.UUID{a.rhino}); err != nil {
		t.Fatalf("exile the Rhino: %v", err)
	}
	pick = chooseCardsChoiceFor(g, opp.ID)
	if pick == nil {
		t.Fatal("then the opponent is asked to exile a card from their graveyard")
	}
	if hasID(pick.ChooseCards, a.keptLand) || !hasID(pick.ChooseCards, a.instant) {
		t.Fatalf("graveyard options %v: the nonland card, never the land", pick.ChooseCards)
	}
	if err := g.ResolveChooseCards(pick.ID, opp.ID, []uuid.UUID{a.instant}); err != nil {
		t.Fatalf("exile the instant: %v", err)
	}
	g.ReadSnapshot(func() {
		c, _ := g.LookupCardForEffect(a.azula)
		a.exiledFrom = game.PermissionCardRef{ID: c.InstanceID, Epoch: c.ObjectEpoch}
	})
	return a
}

func TestAvB20AzulaExilesACreatureThenAGraveyardCardLinkedToHer(t *testing.T) {
	a := avB20AzulaEnters(t)
	for _, id := range []uuid.UUID{a.rhino, a.instant} {
		if got := cardInExile(t, a.g, id).ExiledWith; got != a.exiledFrom {
			t.Errorf("ExiledWith = %+v, want this Azula %+v", got, a.exiledFrom)
		}
	}
	if !a.g.Battlefield.Contains(a.keptToken) || !inGraveyard(a.opp, a.keptLand) {
		t.Error("the token and the land were not eligible and stay where they are")
	}
}

func TestAvB20AzulaCastsTheExiledCardsWithAnyManaAsThoughTheyHadFlash(t *testing.T) {
	a := avB20AzulaEnters(t)
	g, me := a.g, a.me
	// The creature, cast in combat (flash) with blue mana (any type).
	advanceTo(t, g, game.StepBeginCombat)
	floatForTest(g, me, "UUUU")
	if err := g.CastSpell(me.ID, a.rhino, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("cast the exiled creature in combat with blue mana: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c, ok := battlefieldCard(g, a.rhino); !ok || c.Controller != me.ID {
		t.Fatal("the exiled creature resolves under Azula's controller")
	}
	floatForTest(g, me, "U")
	if err := g.CastSpell(me.ID, a.instant, game.CastSpellParams{FromZone: "exile", Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: a.opp.ID}}}); err != nil {
		t.Fatalf("cast the exiled instant with blue mana: %v", err)
	}
}

func TestAvB20AzulaRefusesTheCastOffYourTurn(t *testing.T) {
	a := avB20AzulaEnters(t)
	g, me := a.g, a.me
	if !enumeratedFor(g, me.ID, a.rhino) {
		t.Fatal("non-vacuity: during your turn the exiled creature is castable")
	}
	for g.Seats[g.Turn.ActiveSeat].ID == me.ID {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	advanceTo(t, g, game.StepPrecombatMain)
	floatForTest(g, me, "UUUU")
	if err := g.CastSpell(me.ID, a.rhino, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Fatal("cast a card exiled with Azula during an opponent's turn")
	}
}

// --- Ozai, the Phoenix King --------------------------------------------------

func TestAvB20OzaiTurnsLostManaRedAndKeepsIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Ozai, the Phoenix King", "Legendary Creature — Human Noble", avB20OzaiOracle, false)
	mkAdd(t, g, me, "{G}{U}{R}", game.AddManaOptions{})
	mkAdvance(t, g)
	mkWantPool(t, me, map[string]int{"R": 3})
	mkAdvance(t, g)
	mkWantPool(t, me, map[string]int{"R": 3})
}

func TestAvB20OzaiHasFlyingAndIndestructibleWithSixUnspentMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	ozai := pushCatalogPermanent(g, me.ID, "Ozai, the Phoenix King", "Legendary Creature — Human Noble", avB20OzaiOracle, false)
	for _, kw := range []string{"trample", "haste"} {
		if !hasEffectiveKeyword(t, g, ozai, kw) {
			t.Errorf("Ozai has %s", kw)
		}
	}
	mkAdd(t, g, me, "{R}{R}{R}{R}{R}", game.AddManaOptions{})
	if hasEffectiveKeyword(t, g, ozai, "flying") || hasEffectiveKeyword(t, g, ozai, "indestructible") {
		t.Fatal("five unspent mana is not six")
	}
	mkAdd(t, g, me, "{B}", game.AddManaOptions{})
	for _, kw := range []string{"flying", "indestructible"} {
		if !hasEffectiveKeyword(t, g, ozai, kw) {
			t.Errorf("with six unspent mana Ozai has %s", kw)
		}
	}
}

func TestAvB20OzaiFirebendsFour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ozai := pushCatalogPermanent(g, me.ID, "Ozai, the Phoenix King", "Legendary Creature — Human Noble", avB20OzaiOracle, false)
	declareAttack(t, g, opp.ID, ozai)
	passPriorityAroundTable(t, g)
	if got := me.ManaPool.CountColor("R"); got != 4 {
		t.Errorf("firebending 4 added %d red, want 4", got)
	}
}
