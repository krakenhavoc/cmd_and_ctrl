package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mill_me_mad_2_test.go — S58 PR 11, the Mill Me Mad deck's graveyard
// engines. #2065.

const (
	hedgeShredderOracle         = "f22665cd-4710-48df-83c1-3c9081097cb5"
	hellsCaretakerOracle        = "2f0d797e-d897-453d-92b6-a90e1a548dc5"
	stillnessInMotionOracle     = "d73913ea-44d0-408f-9f8b-8e91843f2826"
	avatarDestinyOracle         = "ecb81888-589b-4486-a444-e0c26384854e"
	beifongsBountyHuntersOracle = "8561b088-56b3-4732-82e9-8da1e3affd55"
	defilingDaemogothOracle     = "d3f7095a-b287-4858-9d08-242cf18e87cb"
	illicitMasqueradeOracle     = "fa2a275a-a28b-45cd-b8ca-a93247ff59bc"
	moonlitMeditationOracle     = "df702b0c-e011-497f-ae29-9876efac4a4c"
	cabalRitualOracle           = "5b5bf1fa-6502-4790-b66b-f0f8504ebc7c"
)

func TestMillMeMad2CardsAreRegistered(t *testing.T) {
	for oracle, name := range map[string]string{
		hedgeShredderOracle:         "Hedge Shredder",
		hellsCaretakerOracle:        "Hell's Caretaker",
		stillnessInMotionOracle:     "Stillness in Motion",
		avatarDestinyOracle:         "Avatar Destiny",
		beifongsBountyHuntersOracle: "Beifong's Bounty Hunters",
		defilingDaemogothOracle:     "Defiling Daemogoth",
		illicitMasqueradeOracle:     "Illicit Masquerade",
		moonlitMeditationOracle:     "Moonlit Meditation",
		cabalRitualOracle:           "Cabal Ritual",
	} {
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

// mm2Creature parks a creature that has been in play a while.
func mm2Creature(g *game.Game, owner uuid.UUID, name string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: testCreatureTypeLine,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// mm2Library replaces a player's library with the given cards, the
// first one on top.
func mm2Library(p *game.Player, cards ...game.Card) []uuid.UUID {
	p.Library.Cards = nil
	ids := make([]uuid.UUID, 0, len(cards))
	for _, c := range cards {
		if c.InstanceID == uuid.Nil {
			c.InstanceID = uuid.New()
		}
		c.Owner, c.Controller = p.ID, p.ID
		p.Library.PushBottom(c)
		ids = append(ids, c.InstanceID)
	}
	return ids
}

func mm2Land(name string) game.Card {
	return game.Card{Name: name, TypeLine: "Basic Land — Forest"}
}

func mm2Spell(name string) game.Card {
	return game.Card{Name: name, TypeLine: "Sorcery"}
}

func mm2CreatureCard(name string, power int) game.Card {
	return game.Card{Name: name, TypeLine: testCreatureTypeLine, Power: power, Toughness: power}
}

func mm2Zone(g *game.Game, id uuid.UUID) game.ZoneKind {
	var kind game.ZoneKind
	g.ReadSnapshot(func() {
		if z := g.FindCardZoneForEffect(id); z != nil {
			kind = z.Kind
		}
	})
	return kind
}

// mm2Settle passes priority once, then until the stack is empty or a
// prompt waits.
func mm2Settle(t *testing.T, g *game.Game) {
	t.Helper()
	passPriorityAroundTable(t, g)
	for i := 0; i < 16 && g.Stack.Size() > 0 && len(g.PendingChoices) == 0; i++ {
		passPriorityAroundTable(t, g)
	}
}

// --- Cabal Ritual -----------------------------------------------------

func TestCabalRitualAddsThreeWithoutThresholdAndFiveWithIt(t *testing.T) {
	for _, tc := range []struct {
		yard, want int
	}{{6, 3}, {7, 5}} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		for i := 0; i < tc.yard; i++ {
			pushGraveyardCardForTest(me, "filler")
		}
		castCatalogSpell(t, g, "Cabal Ritual", "Instant", cabalRitualOracle, nil)
		passPriorityAroundTable(t, g)
		got := batch01PoolColors(me)
		if len(got) != tc.want {
			t.Errorf("%d cards in graveyard: pool %v, want %d black", tc.yard, got, tc.want)
		}
		for _, c := range got {
			if c != "B" {
				t.Errorf("%d cards in graveyard: pool %v, want only black", tc.yard, got)
				break
			}
		}
	}
}

// --- Hell's Caretaker -------------------------------------------------

func TestHellsCaretakerReanimatesDuringYourUpkeepOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	caretaker := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Hell's Caretaker", OracleID: hellsCaretakerOracle,
		TypeLine: "Creature — Horror", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	fodder := mm2Creature(g, me.ID, "Fodder", 1, 1)
	dead := pushGraveyardCardForTest(me, "Big Thing")
	params := game.ActivateAbilityParams{
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: dead}},
		SacrificeIDs: []uuid.UUID{fodder},
	}

	// The harness parks on seat 0's draw step: its own turn, wrong step.
	acRefused(t, g, me.ID, caretaker, 0, params)
	advanceToUpkeepOf(t, g, 1)
	acRefused(t, g, me.ID, caretaker, 0, params)
	_ = opp

	advanceToUpkeepOf(t, g, 0)
	if err := g.ActivateCatalogAbility(me.ID, caretaker, 0, params); err != nil {
		t.Fatalf("ActivateCatalogAbility in your upkeep: %v", err)
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrifice was not paid at announce")
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dead) {
		t.Error("the creature card did not return to the battlefield")
	}
}

// --- Defiling Daemogoth -----------------------------------------------

func TestDefilingDaemogothDrainsTheLifeYouGainedThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp1, opp2 := g.Seats[0], g.Seats[1], g.Seats[2]
	pushPermanentForTest(g, me.ID, "Defiling Daemogoth", defilingDaemogothOracle, "Creature — Demon")
	advanceToMainOf(t, g, 0)
	g.WithWriteLock(func() {
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3)
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 2)
	})
	start1, start2 := opp1.Life, opp2.Life
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if opp1.Life != start1-5 || opp2.Life != start2-5 {
		t.Errorf("each opponent should lose 5: %d -> %d, %d -> %d", start1, opp1.Life, start2, opp2.Life)
	}
}

func TestDefilingDaemogothGainsOneForEachCreatureThatConnects(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	daemogoth := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Defiling Daemogoth", OracleID: defilingDaemogothOracle,
		TypeLine: "Creature — Demon", Power: 5, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	bear := mm2Creature(g, me.ID, "Bear", 2, 2)
	start := me.Life
	g.WithWriteLock(func() {
		for _, src := range []uuid.UUID{daemogoth, bear} {
			g.EmitEvent(game.Event{Kind: game.EventDealDamage, Actor: me.ID, Source: src, Target: opp.ID, Amount: 2, Combat: true})
		}
		// Noncombat damage is not the trigger.
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Actor: me.ID, Source: bear, Target: opp.ID, Amount: 2})
	})
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if me.Life != start+2 {
		t.Errorf("life %d -> %d, want +2 (two creatures connected in combat)", start, me.Life)
	}
}

// --- Beifong's Bounty Hunters -----------------------------------------

func TestBeifongsBountyHuntersEarthbendsTheDeadCreaturesPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushPermanentForTest(g, me.ID, "Beifong's Bounty Hunters", beifongsBountyHuntersOracle, "Creature — Human Mercenary")
	land := pushEarthbendLand(g, me.ID, "Forest", "Basic Land — Forest")
	victim := mm2Creature(g, me.ID, "Ogre", 3, 3)

	killCreature(t, g, me.ID, victim)
	if p := latestPickTarget(g, me.ID); p != nil {
		pickCard(t, g, me.ID, land)
	}
	passPriorityAroundTable(t, g)

	power, toughness, creatureLand, hasty := earthbentBody(t, g, land)
	if power != 3 || toughness != 3 || !creatureLand || !hasty {
		t.Errorf("earthbent land = %d/%d creature land %v haste %v, want a 3/3 hasty land creature", power, toughness, creatureLand, hasty)
	}
}

func TestBeifongsBountyHuntersIgnoresALandCreatureDying(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushPermanentForTest(g, me.ID, "Beifong's Bounty Hunters", beifongsBountyHuntersOracle, "Creature — Human Mercenary")
	pushEarthbendLand(g, me.ID, "Forest", "Basic Land — Forest")
	manland := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Dryad Arbor", TypeLine: "Land Creature — Forest Dryad",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	killCreature(t, g, me.ID, manland)
	if p := latestPickTarget(g, me.ID); p != nil {
		t.Fatal("a land creature dying asked for an earthbend target")
	}
	if triggerOnStackFrom(g, "Beifong's Bounty Hunters") {
		t.Fatal("a land creature dying triggered the Hunters")
	}
}

// --- Illicit Masquerade -----------------------------------------------

func TestIllicitMasqueradeExilesTheImpostorAndReturnsAnotherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	impostor := mm2Creature(g, me.ID, "Impostor", 2, 2)
	other := mm2Creature(g, me.ID, "Other", 2, 2)
	castCatalogSpell(t, g, "Illicit Masquerade", "Enchantment", illicitMasqueradeOracle, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{impostor, other} {
		if c, ok := findBattlefieldByID(g, id); !ok || c.Counters["impostor"] != 1 {
			t.Fatalf("%s should carry one impostor counter: %+v", id, c.Counters)
		}
	}

	back := pushGraveyardCardForTest(me, "Old Friend")
	killCreature(t, g, me.ID, impostor)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the return")
	}
	for _, id := range p.PickTargetCards {
		if id == impostor {
			t.Error("the dying impostor was offered as the 'other' creature card")
		}
	}
	pickCard(t, g, me.ID, back)
	passPriorityAroundTable(t, g)

	if mm2Zone(g, impostor) != game.ZoneExile {
		t.Errorf("the impostor is in %s, want exile", mm2Zone(g, impostor))
	}
	if !g.Battlefield.Contains(back) {
		t.Error("the other creature card did not return")
	}
}

// "Up to one": with no other creature card to return, the trigger still
// resolves and still exiles the impostor (CR 603.3d does not remove it).
func TestIllicitMasqueradeExilesEvenWithNothingToReturn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushPermanentForTest(g, me.ID, "Illicit Masquerade", illicitMasqueradeOracle, "Enchantment")
	impostor := mm2Creature(g, me.ID, "Impostor", 2, 2)
	g.WithWriteLock(func() {
		if err := g.AddCounterByForEffect(me.ID, impostor, "impostor", 1); err != nil {
			t.Fatalf("AddCounter: %v", err)
		}
	})
	killCreature(t, g, me.ID, impostor)
	mm2Settle(t, g)
	if mm2Zone(g, impostor) != game.ZoneExile {
		t.Errorf("the impostor is in %s, want exile", mm2Zone(g, impostor))
	}
}

func TestIllicitMasqueradeIgnoresCreaturesWithoutTheCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushPermanentForTest(g, me.ID, "Illicit Masquerade", illicitMasqueradeOracle, "Enchantment")
	plain := mm2Creature(g, me.ID, "Plain", 2, 2)
	pushGraveyardCardForTest(me, "Old Friend")
	killCreature(t, g, me.ID, plain)
	if latestPickTarget(g, me.ID) != nil || triggerOnStackFrom(g, "Illicit Masquerade") {
		t.Fatal("a creature with no impostor counter triggered the Masquerade")
	}
}

// --- Avatar Destiny ---------------------------------------------------

func TestAvatarDestinyGrowsAndMakesAnAvatar(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := auraBear(g, me.ID)
	pushGraveyardCardForTest(me, "Dead One")
	pushGraveyardCardForTest(me, "Dead Two")
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Not a creature", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	auraCast(t, g, "Avatar Destiny", avatarDestinyOracle, bear)

	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	c := findBattlefieldCardByID(g, bear)
	if c == nil {
		t.Fatal("the Bear is gone")
	}
	if c.CurrentPower() != 4 || c.CurrentToughness() != 4 {
		t.Errorf("enchanted Bear is %d/%d, want 4/4 (two creature cards)", c.CurrentPower(), c.CurrentToughness())
	}
	if subs := c.Effective().Subtypes; !eotHasType(subs, "Avatar") || !eotHasType(subs, "Bear") {
		t.Errorf("subtypes %v, want Avatar in addition to Bear", c.Effective().Subtypes)
	}
}

func TestAvatarDestinyMillsByPowerAndReturnsItselfAndAMilledCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := auraBear(g, me.ID)
	pushGraveyardCardForTest(me, "Dead One")
	lib := mm2Library(me, mm2Spell("Top"), mm2CreatureCard("Milled Beast", 5), mm2Land("Forest"), mm2Spell("Bottom"), mm2Spell("Below"))
	aura := auraCast(t, g, "Avatar Destiny", avatarDestinyOracle, bear)

	// The Bear dies a 3/3 (one creature card in the graveyard), so the
	// trigger mills three.
	killCreature(t, g, me.ID, bear)
	mm2Settle(t, g)
	prompt := pendingOfKind(g, game.PendingChoiceChooseCards)
	if prompt == nil {
		t.Fatalf("no choice of a milled creature card: %+v", g.PendingChoices)
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{lib[1]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	mm2Settle(t, g)

	if mm2Zone(g, lib[0]) != game.ZoneGraveyard || mm2Zone(g, lib[2]) != game.ZoneGraveyard {
		t.Error("the top three cards were not milled")
	}
	if mm2Zone(g, lib[3]) != game.ZoneLibrary {
		t.Error("more than three cards were milled")
	}
	if !g.Battlefield.Contains(lib[1]) {
		t.Error("the chosen milled creature card did not return to the battlefield")
	}
	if mm2Zone(g, aura) != game.ZoneHand {
		t.Errorf("Avatar Destiny is in %s, want its owner's hand", mm2Zone(g, aura))
	}
}

// --- Stillness in Motion ----------------------------------------------

func TestStillnessInMotionMillsThreeAndKeepsItselfWhileCardsRemain(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	stillness := pushPermanentForTest(g, me.ID, "Stillness in Motion", stillnessInMotionOracle, "Enchantment")
	lib := mm2Library(me, mm2Spell("A"), mm2Spell("B"), mm2Spell("C"), mm2Spell("D"))
	advanceToUpkeepOf(t, g, 0)
	mm2Settle(t, g)
	for _, id := range lib[:3] {
		if mm2Zone(g, id) != game.ZoneGraveyard {
			t.Errorf("%s was not milled", id)
		}
	}
	if !g.Battlefield.Contains(stillness) {
		t.Error("Stillness in Motion left with a card still in the library")
	}
}

func TestStillnessInMotionRestacksFiveWhenTheLibraryRunsOut(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	stillness := pushPermanentForTest(g, me.ID, "Stillness in Motion", stillnessInMotionOracle, "Enchantment")
	mm2Library(me, mm2Spell("A"), mm2Spell("B"))
	for i := 0; i < 6; i++ {
		pushGraveyardCardForTest(me, "old")
	}
	advanceToUpkeepOf(t, g, 0)
	mm2Settle(t, g)

	pick := pendingOfKind(g, game.PendingChoiceChooseCards)
	if pick == nil {
		t.Fatalf("no choice of five graveyard cards: %+v", g.PendingChoices)
	}
	var five []uuid.UUID
	for _, c := range me.Graveyard.Cards {
		if len(five) < 5 {
			five = append(five, c.InstanceID)
		}
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, five); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	order := pendingOfKind(g, game.PendingChoicePutInLibrary)
	if order == nil {
		t.Fatalf("no order prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, nil, five); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if mm2Zone(g, stillness) != game.ZoneExile {
		t.Errorf("Stillness in Motion is in %s, want exile", mm2Zone(g, stillness))
	}
	if got := me.Library.Size(); got != 5 {
		t.Fatalf("library has %d cards, want 5", got)
	}
	if top, err := me.Library.Top(); err != nil || top.InstanceID != five[0] {
		t.Errorf("the first card in the chosen order is not on top")
	}
	if got := me.Graveyard.Size(); got != 3 {
		t.Errorf("graveyard has %d cards, want 3 (eight, less the five)", got)
	}
}

// --- Hedge Shredder ---------------------------------------------------

func TestHedgeShredderPutsMilledLandsOntoTheBattlefieldTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushPermanentForTest(g, me.ID, "Hedge Shredder", hedgeShredderOracle, "Artifact — Vehicle")
	lib := mm2Library(me, mm2Land("Forest"), mm2Spell("Spell"), mm2Land("Swamp"), mm2Spell("Kept"))
	g.WithWriteLock(func() {
		if _, err := g.MillToZoneForEffect(me.ID, 3, game.ZoneGraveyard); err != nil {
			t.Fatalf("mill: %v", err)
		}
	})
	mm2Settle(t, g)
	for _, id := range []uuid.UUID{lib[0], lib[2]} {
		c, ok := findBattlefieldByID(g, id)
		if !ok {
			t.Errorf("milled land %s is not on the battlefield", id)
			continue
		}
		if !c.Tapped {
			t.Errorf("milled land %s entered untapped", id)
		}
	}
	if mm2Zone(g, lib[1]) != game.ZoneGraveyard {
		t.Error("the milled nonland card left the graveyard")
	}
	if mm2Zone(g, lib[3]) != game.ZoneLibrary {
		t.Error("the fourth card was milled")
	}
}

func TestHedgeShredderIgnoresAnOpponentsMilledLands(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushPermanentForTest(g, me.ID, "Hedge Shredder", hedgeShredderOracle, "Artifact — Vehicle")
	lib := mm2Library(opp, mm2Land("Forest"))
	g.WithWriteLock(func() {
		if _, err := g.MillToZoneForEffect(opp.ID, 1, game.ZoneGraveyard); err != nil {
			t.Fatalf("mill: %v", err)
		}
	})
	mm2Settle(t, g)
	if mm2Zone(g, lib[0]) != game.ZoneGraveyard {
		t.Error("an opponent's milled land moved")
	}
}

// --- Moonlit Meditation -----------------------------------------------

func mm2Moonlit(t *testing.T, g *game.Game, me uuid.UUID) uuid.UUID {
	t.Helper()
	host := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Host Golem", TypeLine: "Artifact Creature — Golem",
		Power: 3, Toughness: 3, Owner: me, Controller: me,
	})
	auraCast(t, g, "Moonlit Meditation", moonlitMeditationOracle, host)
	return host
}

func mm2AnswerOptional(t *testing.T, g *game.Game, me uuid.UUID, yes bool) {
	t.Helper()
	p := pendingOfKind(g, game.PendingChoiceOptionalReplacement)
	if p == nil {
		t.Fatalf("no Moonlit Meditation prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveOptionalReplacement(p.ID, me, yes); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
}

func TestMoonlitMeditationTurnsTheFirstCreationIntoCopies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0].ID
	mm2Moonlit(t, g, me)

	makeTokens(t, g, me, TreasureToken(), 2)
	mm2AnswerOptional(t, g, me, true)
	if got := onBattlefieldNamed(g, "Host Golem"); got != 3 {
		t.Errorf("Host Golems = %d, want the original and two copies", got)
	}
	if got := onBattlefieldNamed(g, "Treasure"); got != 0 {
		t.Errorf("Treasures = %d, want none", got)
	}

	// The second creation this turn is not the first.
	makeTokens(t, g, me, TreasureToken(), 1)
	if p := pendingOfKind(g, game.PendingChoiceOptionalReplacement); p != nil {
		t.Fatal("the second creation of the turn was offered the swap")
	}
	if got := onBattlefieldNamed(g, "Treasure"); got != 1 {
		t.Errorf("Treasures = %d, want 1", got)
	}
}

func TestMoonlitMeditationDecliningUsesUpTheTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0].ID
	mm2Moonlit(t, g, me)

	makeTokens(t, g, me, TreasureToken(), 1)
	mm2AnswerOptional(t, g, me, false)
	makeTokens(t, g, me, TreasureToken(), 1)
	if p := pendingOfKind(g, game.PendingChoiceOptionalReplacement); p != nil {
		t.Fatal("declining did not use up the turn's first time")
	}
	if got := onBattlefieldNamed(g, "Treasure"); got != 2 {
		t.Errorf("Treasures = %d, want 2", got)
	}
}

func TestMoonlitMeditationWaitsForTheNextTurnIfTokensCameFirst(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0].ID
	makeTokens(t, g, me, TreasureToken(), 1)
	mm2Moonlit(t, g, me)
	makeTokens(t, g, me, TreasureToken(), 1)
	if p := pendingOfKind(g, game.PendingChoiceOptionalReplacement); p != nil {
		t.Fatal("tokens were already created this turn, so this is not the first time")
	}
}

func TestMoonlitMeditationOffersOnlyTheFirstOfTwoWaitingCreations(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0].ID
	mm2Moonlit(t, g, me)

	// Two instructions in one resolution: the first pauses on the
	// question and its caller carries on to the second.
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(me, TreasureToken(), 1); err != nil {
			t.Fatalf("first creation: %v", err)
		}
		if err := g.CreateTokenForEffect(me, ClueToken(), 1); err != nil {
			t.Fatalf("second creation: %v", err)
		}
	})
	n := 0
	for _, p := range g.PendingChoices {
		if p != nil && p.Kind == game.PendingChoiceOptionalReplacement {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("%d Moonlit Meditation prompts, want 1", n)
	}
	if got := onBattlefieldNamed(g, "Clue"); got != 1 {
		t.Errorf("Clues = %d, want the second creation made as printed", got)
	}
	mm2AnswerOptional(t, g, me, true)
	if got := onBattlefieldNamed(g, "Host Golem"); got != 2 {
		t.Errorf("Host Golems = %d, want the original and one copy", got)
	}
}

func TestMoonlitMeditationAppliesAgainNextTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0].ID
	mm2Moonlit(t, g, me)
	makeTokens(t, g, me, TreasureToken(), 1)
	mm2AnswerOptional(t, g, me, false)

	advanceToUpkeepOf(t, g, 1)
	makeTokens(t, g, me, TreasureToken(), 1)
	mm2AnswerOptional(t, g, me, true)
	if got := onBattlefieldNamed(g, "Host Golem"); got != 2 {
		t.Errorf("Host Golems = %d, want a copy on the next turn's first creation", got)
	}
}
