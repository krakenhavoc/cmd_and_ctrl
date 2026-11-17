package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// m3c_5468_cards_test.go — Modern Horizons 3 Commander, batch 1 (#5468,
// S82): twelve cards built from seams the engine already has. Each test
// plays the real catalog entry.

const (
	m3cb1TrenchpostOracle       = "42f1ccb8-eda0-4828-ac07-82d4e950d7e1"
	m3cb1BroodmateTyrantOracle  = "9b984236-39cf-4552-827b-e81c26cfb388"
	m3cb1DreadsireOracle        = "2afefaab-c614-4fca-afba-d8d206d9a8b9"
	m3cb1BloodbraidOracle       = "6e111622-c7cc-49cf-83e8-3555e06518c1"
	m3cb1DispatcherOracle       = "c1b276b7-f1b3-4f6f-b6a2-1315d805e0f2"
	m3cb1EldritchImmunityOracle = "7dce1e8c-5cfd-4626-9f23-7bb3f1b10c1e"
	m3cb1SiegeGangOracle        = "9e9d538b-a9e5-4d59-b7c1-208fe18110e3"
	m3cb1ConfluenceOracle       = "01160cb5-6d83-4317-9522-74083b3a83bb"
	m3cb1PolygoyfOracle         = "6021f2e3-f915-428c-8289-2aa675b5bc78"
	m3cb1PyrogoyfOracle         = "7fd7457a-388d-4cca-a7cf-86b4ea922037"
	m3cb1JyotiOracle            = "67f0cef6-8ef4-4c39-b306-2b580036894e"
	m3cb1HorizonOracle          = "59a82f57-fe2f-4834-a4ee-4b948eef1e12"
)

func TestM3cb1CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		m3cb1TrenchpostOracle:       "Trenchpost",
		m3cb1BroodmateTyrantOracle:  "Broodmate Tyrant",
		m3cb1DreadsireOracle:        "Ulamog's Dreadsire",
		m3cb1BloodbraidOracle:       "Bloodbraid Challenger",
		m3cb1DispatcherOracle:       "Chittering Dispatcher",
		m3cb1EldritchImmunityOracle: "Eldritch Immunity",
		m3cb1SiegeGangOracle:        "Siege-Gang Lieutenant",
		m3cb1ConfluenceOracle:       "Eldrazi Confluence",
		m3cb1PolygoyfOracle:         "Polygoyf",
		m3cb1PyrogoyfOracle:         "Pyrogoyf",
		m3cb1JyotiOracle:            "Jyoti, Moag Ancient",
		m3cb1HorizonOracle:          "Horizon of Progress",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// m3cb1Opponent is the seat after the active one.
func m3cb1Opponent(g *game.Game) *game.Player {
	return g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

// m3cb1Push seeds a permanent, catalog row or not, untapped and able to
// act.
func m3cb1Push(g *game.Game, owner uuid.UUID, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = owner, owner
	return pushBattlefieldCardWithTimestamp(g, c)
}

// --- Trenchpost --------------------------------------------------------------

func TestM3cb1TrenchpostMillsOneCardPerLocusYouControl(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := m3cb1Opponent(g)
	trench := pushCatalogPermanent(g, me.ID, "Trenchpost", "Land — Locus", m3cb1TrenchpostOracle, false)
	m3cb1Push(g, me.ID, game.Card{Name: "Test Locus", TypeLine: "Land — Locus"})
	m3cb1Push(g, me.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	m3cb1Push(g, opp.ID, game.Card{Name: "Their Locus", TypeLine: "Land — Locus"})
	lib, yard := opp.Library.Size(), opp.Graveyard.Size()

	floatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, trench, 0, game.ActivateAbilityParams{Targets: playerRefs(opp.ID), Strict: true}); err == nil {
		t.Fatal("activated the {3} ability for two mana")
	}
	floatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, trench, 0, game.ActivateAbilityParams{Targets: playerRefs(opp.ID), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	// Trenchpost and the Test Locus are yours; the opponent's Locus is not.
	if got := opp.Library.Size(); got != lib-2 {
		t.Errorf("library %d -> %d, want two milled", lib, got)
	}
	if got := opp.Graveyard.Size(); got != yard+2 {
		t.Errorf("graveyard %d -> %d, want +2", yard, got)
	}
	if !layeredCard(t, g, trench).Tapped {
		t.Error("the {T} cost did not tap Trenchpost")
	}
}

func TestM3cb1TrenchpostTapsForColorless(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	trench := pushCatalogPermanent(g, me.ID, "Trenchpost", "Land — Locus", m3cb1TrenchpostOracle, false)
	if err := g.ActivateManaAbility(me.ID, trench, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if me.ManaPool.Total() != 1 {
		t.Errorf("pool %v, want {C}", me.ManaPool)
	}
}

// --- Broodmate Tyrant --------------------------------------------------------

func TestM3cb1BroodmateTyrantEntersWithADragon(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	tyrant := trekB31Enter(t, g, me, game.Card{Name: "Broodmate Tyrant", TypeLine: "Creature — Dragon",
		OracleID: m3cb1BroodmateTyrantOracle, ManaCost: "{4}{B}{R}{G}", Power: 5, Toughness: 5})
	passPriorityAroundTable(t, g)
	if !effectiveAbilitiesContain(t, g, tyrant, "flying") {
		t.Error("the Tyrant has no flying")
	}
	dragons := s46Tokens(g, me.ID, "Dragon")
	if len(dragons) != 1 {
		t.Fatalf("%d Dragon tokens, want 1", len(dragons))
	}
	if p, tg := ptOf(t, g, dragons[0].InstanceID); p != 5 || tg != 5 {
		t.Errorf("the Dragon is %d/%d, want 5/5", p, tg)
	}
	if !dragons[0].HasColor("R") || !effectiveAbilitiesContain(t, g, dragons[0].InstanceID, "flying") {
		t.Error("the Dragon is not a red flyer")
	}
}

func TestM3cb1BroodmateTyrantEncoreCopiesEachMakeADragon(t *testing.T) {
	g := newCatalogGame(t)
	_, me := s46Encore(t, g, "Broodmate Tyrant", "Creature — Dragon", m3cb1BroodmateTyrantOracle, 5, 5, "{C}{C}{C}{C}{C}{B}{R}{G}")
	s46Settle(t, g)
	if n := len(s46Tokens(g, me.ID, "Broodmate Tyrant")); n != 3 {
		t.Fatalf("%d encore tokens, want one per opponent", n)
	}
	if n := len(s46Tokens(g, me.ID, "Dragon")); n != 3 {
		t.Errorf("%d Dragon tokens, want one per encore copy", n)
	}
}

// --- Ulamog's Dreadsire ------------------------------------------------------

func TestM3cb1DreadsireTapsForATenTenEldrazi(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	sire := pushCatalogPermanent(g, me.ID, "Ulamog's Dreadsire", "Creature — Eldrazi", m3cb1DreadsireOracle, false)
	if !effectiveAbilitiesContain(t, g, sire, "vigilance") {
		t.Error("the Dreadsire has no vigilance")
	}
	if err := g.ActivateCatalogAbility(me.ID, sire, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	toks := s46Tokens(g, me.ID, "Eldrazi")
	if len(toks) != 1 {
		t.Fatalf("%d Eldrazi tokens, want 1", len(toks))
	}
	if p, tg := ptOf(t, g, toks[0].InstanceID); p != 10 || tg != 10 {
		t.Errorf("the Eldrazi is %d/%d, want 10/10", p, tg)
	}
}

func TestM3cb1DreadsireSummoningSickCannotTap(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	sire := pushCatalogPermanent(g, me.ID, "Ulamog's Dreadsire", "Creature — Eldrazi", m3cb1DreadsireOracle, true)
	if err := g.ActivateCatalogAbility(me.ID, sire, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("a summoning-sick Dreadsire tapped for a token")
	}
}

// The ward charges a permanent with mana value 1 or greater: the caster's
// land and their {0} artifact creature are not offered, the two-drop is.
func TestM3cb1DreadsireWardTakesAPermanentWithManaValueOneOrMore(t *testing.T) {
	g, _, opp, sire := wardTable(t, "Ulamog's Dreadsire", "Creature — Eldrazi", m3cb1DreadsireOracle)
	land := m3cb1Push(g, opp.ID, game.Card{Name: "Their Forest", TypeLine: "Basic Land — Forest"})
	zero := m3cb1Push(g, opp.ID, game.Card{Name: "Their Ornithopter", TypeLine: "Artifact Creature — Thopter", ManaCost: "{0}", Power: 0, Toughness: 2})
	bear := m3cb1Push(g, opp.ID, game.Card{Name: "Their Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2})
	castAtWardedCreature(t, g, opp, sire)

	prompt := passUntilConfirmFor(t, g, opp.ID)
	if prompt == nil {
		t.Fatal("no ward prompt for the spell's controller")
	}
	if err := g.ResolveConfirm(prompt.ID, opp.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	pick := chooseCardsChoiceFor(g, opp.ID)
	if pick == nil {
		t.Fatal("accepting the ward queued no sacrifice pick")
	}
	if len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != bear {
		t.Fatalf("candidates = %v, want only the two-drop %s", pick.ChooseCards, bear)
	}
	if err := g.ResolveChooseCards(pick.ID, opp.ID, []uuid.UUID{bear}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the chosen permanent was not sacrificed")
	}
	if !g.Battlefield.Contains(land) || !g.Battlefield.Contains(zero) {
		t.Error("a mana value 0 permanent was sacrificed")
	}
	if g.Battlefield.Contains(sire) {
		t.Error("the ward was paid, so the removal should have resolved")
	}
}

func TestM3cb1DreadsireWardCountersWhenOnlyManaValueZeroIsLeft(t *testing.T) {
	g, _, opp, sire := wardTable(t, "Ulamog's Dreadsire", "Creature — Eldrazi", m3cb1DreadsireOracle)
	m3cb1Push(g, opp.ID, game.Card{Name: "Their Forest", TypeLine: "Basic Land — Forest"})
	m3cb1Push(g, opp.ID, game.Card{Name: "Their Ornithopter", TypeLine: "Artifact Creature — Thopter", ManaCost: "{0}", Power: 0, Toughness: 2})
	blade := castAtWardedCreature(t, g, opp, sire)
	if prompt := passUntilConfirmFor(t, g, opp.ID); prompt != nil {
		t.Fatalf("a caster with nothing of mana value 1 or more was offered the sacrifice: %q", prompt.Reason)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(sire) || !opp.Graveyard.Contains(blade) {
		t.Error("a ward the caster cannot pay must counter the spell")
	}
}

// --- Bloodbraid Challenger ---------------------------------------------------

func TestM3cb1BloodbraidChallengerDeclaresEscape(t *testing.T) {
	if !game.CardCastableFromZone(m3cb1BloodbraidOracle, game.ZoneGraveyard) {
		t.Fatal("the graveyard is not castable")
	}
	offers := game.AlternativeCostsOfferedFromZone(m3cb1BloodbraidOracle, game.ZoneGraveyard)
	if len(offers) != 1 || offers[0].Key != "escape" {
		t.Fatalf("graveyard offers = %+v, want one escape", offers)
	}
	if offers[0].ManaCost != "{3}{R}{G}" {
		t.Errorf("escape cost %q, want {3}{R}{G}", offers[0].ManaCost)
	}
	if offers[0].ExileFromGraveyard == nil || offers[0].ExileFromGraveyard.Min != 3 || offers[0].ExileFromGraveyard.Max != 3 {
		t.Errorf("escape exile clause %+v, want exactly three", offers[0].ExileFromGraveyard)
	}
	if offers[0].EntersWithCounterCount != 0 {
		t.Errorf("escapes with %d counters, want none", offers[0].EntersWithCounterCount)
	}
}

// Escaped from the graveyard it still cascades, and it arrives with haste.
func TestM3cb1BloodbraidChallengerEscapesAndCascades(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	seedCheapLibrary(g, me, 3)
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Bloodbraid Challenger", TypeLine: "Creature — Elf Berserker",
		OracleID: m3cb1BloodbraidOracle, ManaCost: "{3}{R}{G}", Power: 4, Toughness: 3, Owner: me.ID, Controller: me.ID})
	pay := seedGraveyardFodder(t, g, 3)

	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "escape", AltCostIDs: pay,
	}); err != nil {
		t.Fatalf("escape cast: %v", err)
	}
	for _, p := range pay {
		if !g.Exile.Contains(p) {
			t.Errorf("escape fodder %s was not exiled", p)
		}
	}
	passPriorityAroundTable(t, g)
	if n := answerAllMayCast(t, g, me.ID, false); n != 1 {
		t.Errorf("cascade offers %d, want 1", n)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("the escaped Challenger is not on the battlefield")
	}
	if !effectiveAbilitiesContain(t, g, id, "haste") {
		t.Error("the Challenger has no haste")
	}
}

// --- Chittering Dispatcher ---------------------------------------------------

func TestM3cb1ChitteringDispatcherLeavesASpawnBehind(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	disp := b12Push(g, me.ID, "Chittering Dispatcher", "Creature — Eldrazi Drone", m3cb1DispatcherOracle, 2, 3)
	trekB18BDestroy(t, g, disp)
	passPriorityAroundTable(t, g)
	spawns := s46Tokens(g, me.ID, "Eldrazi Spawn")
	if len(spawns) != 1 {
		t.Fatalf("%d Eldrazi Spawn tokens, want 1", len(spawns))
	}
	if p, tg := ptOf(t, g, spawns[0].InstanceID); p != 0 || tg != 1 {
		t.Errorf("the Spawn is %d/%d, want 0/1", p, tg)
	}
}

// Myriad copies one Dispatcher per other opponent, and a copy leaving
// makes a Spawn too.
func TestM3cb1ChitteringDispatcherMyriadCopiesLeaveSpawns(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	disp := b12Push(g, me.ID, "Chittering Dispatcher", "Creature — Eldrazi Drone", m3cb1DispatcherOracle, 2, 3)
	declareAttack(t, g, opp.ID, disp)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	answerMayChoice(t, g, me.ID, true)
	copies := s46Tokens(g, me.ID, "Chittering Dispatcher")
	if len(copies) != 2 {
		t.Fatalf("%d myriad copies, want 2", len(copies))
	}
	trekB18BDestroy(t, g, copies[0].InstanceID)
	s46Settle(t, g)
	if n := len(s46Tokens(g, me.ID, "Eldrazi Spawn")); n != 1 {
		t.Errorf("%d Spawns after a copy left, want 1", n)
	}
}

// --- Eldritch Immunity -------------------------------------------------------

func m3cb1HasProtectionFromEachColor(g *game.Game, id uuid.UUID) bool {
	for _, c := range game.AllColors {
		if !p1g2HasKeyword(g, id, game.ProtectionFromColor(c)) {
			return false
		}
	}
	return true
}

func TestM3cb1EldritchImmunityProtectsTargetCreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := m3cb1Opponent(g)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	other := b12Creature(g, me.ID, "My Other Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := castCatalogSpellErr(t, g, "Eldritch Immunity", "Kindred Instant — Eldrazi", m3cb1EldritchImmunityOracle, cardRefs(theirs)); err == nil {
		t.Fatal("an opponent's creature is not a legal target")
	}
	castCatalogSpell(t, g, "Eldritch Immunity", "Kindred Instant — Eldrazi", m3cb1EldritchImmunityOracle, cardRefs(mine))
	passPriorityAroundTable(t, g)
	if !m3cb1HasProtectionFromEachColor(g, mine) {
		t.Error("the target lacks protection from each color")
	}
	if p1g2HasKeyword(g, other, game.ProtectionFromColor("R")) {
		t.Error("an untargeted creature gained protection")
	}
	dftMPPassTurn(t, g)
	if p1g2HasKeyword(g, mine, game.ProtectionFromColor("R")) {
		t.Error("the protection outlasted the turn")
	}
}

func TestM3cb1EldritchImmunityOverloadedProtectsEachCreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := m3cb1Opponent(g)
	a := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "My Other Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castWithAltCost(t, g, "Eldritch Immunity", "Kindred Instant — Eldrazi", m3cb1EldritchImmunityOracle, "overload")
	passPriorityAroundTable(t, g)
	if !m3cb1HasProtectionFromEachColor(g, a) || !m3cb1HasProtectionFromEachColor(g, b) {
		t.Error("overloaded, each creature you control should have protection from each color")
	}
	if p1g2HasKeyword(g, theirs, game.ProtectionFromColor("R")) {
		t.Error("an opponent's creature gained protection")
	}
}

// --- Siege-Gang Lieutenant ---------------------------------------------------

func TestM3cb1SiegeGangLieutenantMakesHastyGoblinsOnlyWithYourCommander(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Siege-Gang Lieutenant", "Creature — Goblin", m3cb1SiegeGangOracle, 2, 2)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if got := len(gl1bNamed(g, me.ID, "Goblin")); got != 0 {
		t.Fatalf("Goblins without a commander = %d, want 0", got)
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	b12Push(g2, me2.ID, "Siege-Gang Lieutenant", "Creature — Goblin", m3cb1SiegeGangOracle, 2, 2)
	m3cb1Push(g2, me2.ID, game.Card{Name: "My Commander", TypeLine: "Legendary Creature — Bear",
		Power: 2, Toughness: 2, IsCommander: true})
	advanceTo(t, g2, game.StepBeginCombat)
	passPriorityAroundTable(t, g2)
	goblins := gl1bNamed(g2, me2.ID, "Goblin")
	if len(goblins) != 2 {
		t.Fatalf("Goblins with a commander = %d, want 2", len(goblins))
	}
	for _, gob := range goblins {
		if !effectiveAbilitiesContain(t, g2, gob.InstanceID, "haste") {
			t.Error("a Goblin token has no haste this turn")
		}
	}
}

func TestM3cb1SiegeGangLieutenantSacrificesAGoblinToPing(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := m3cb1Opponent(g)
	lt := b12Push(g, me.ID, "Siege-Gang Lieutenant", "Creature — Goblin", m3cb1SiegeGangOracle, 2, 2)
	goblin := b12Creature(g, me.ID, "My Goblin", "Creature — Goblin", 1, 1)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	floatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, lt, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{bear}, Targets: playerRefs(opp.ID), Strict: true,
	}); err == nil {
		t.Fatal("a non-Goblin was sacrificed to the Lieutenant")
	}
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, lt, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{goblin}, Targets: playerRefs(opp.ID), Strict: true,
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(goblin) {
		t.Error("the Goblin was not sacrificed as the cost")
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("opponent life %d -> %d, want -1", life, opp.Life)
	}
}

// --- Eldrazi Confluence ------------------------------------------------------

func TestM3cb1EldraziConfluencePumpsAndMakesTwoScions(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := m3cb1Opponent(g)
	big := b12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 2, 4)
	castModal(t, g, "Eldrazi Confluence", "Instant", m3cb1ConfluenceOracle,
		[]int{0, 2, 2}, []game.TargetRef{modeRef(game.TargetCard, big, 0, 0)})
	passPriorityAroundTable(t, g)
	if p, tg := ptOf(t, g, big); p != 5 || tg != 1 {
		t.Errorf("the Wall is %d/%d, want 5/1 after +3/-3", p, tg)
	}
	if n := len(s46Tokens(g, me.ID, "Eldrazi Scion")); n != 2 {
		t.Errorf("%d Eldrazi Scions, want 2", n)
	}
}

// The flicker returns the permanent tapped, as a new object, under its
// OWNER's control; +3/-3 kills a 3-toughness creature.
func TestM3cb1EldraziConfluenceFlickersTappedToItsOwner(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := m3cb1Opponent(g)
	stolen := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Borrowed Rock",
		TypeLine: "Artifact", Owner: opp.ID, Controller: me.ID})
	ox := b12Creature(g, opp.ID, "Their Ox", "Creature — Ox", 2, 3)
	castModal(t, g, "Eldrazi Confluence", "Instant", m3cb1ConfluenceOracle,
		[]int{1, 0, 2}, []game.TargetRef{
			modeRef(game.TargetCard, stolen, 0, 0),
			modeRef(game.TargetCard, ox, 1, 0),
		})
	passPriorityAroundTable(t, g)
	if n := len(s46Tokens(g, me.ID, "Eldrazi Scion")); n != 1 {
		t.Errorf("%d Eldrazi Scions, want 1", n)
	}
	back := roonaOnly(t, g, "Borrowed Rock")
	if back.InstanceID == stolen {
		t.Error("the permanent came back as the same object (CR 400.7)")
	}
	if !back.Tapped || back.Controller != opp.ID {
		t.Errorf("returned tapped=%v under %v, want tapped under its owner %v", back.Tapped, back.Controller, opp.ID)
	}
	if g.Battlefield.Contains(ox) {
		t.Error("+3/-3 should have killed the 3-toughness Ox")
	}
}

func TestM3cb1EldraziConfluenceCannotFlickerALand(t *testing.T) {
	g := newCatalogGame(t)
	opp := m3cb1Opponent(g)
	land := m3cb1Push(g, opp.ID, game.Card{Name: "Their Forest", TypeLine: "Basic Land — Forest"})
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: "Eldrazi Confluence", TypeLine: "Instant",
		OracleID: m3cb1ConfluenceOracle, Owner: active.ID, Controller: active.ID})
	advanceToMain(t, g)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{Modes: []int{1, 2, 2},
		Targets: []game.TargetRef{modeRef(game.TargetCard, land, 0, 0)}}); err == nil {
		t.Fatal("a land is not a nonland permanent")
	}
}

// --- Polygoyf ----------------------------------------------------------------

func TestM3cb1PolygoyfCountsCardTypesAndMyriads(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	goyf := b12Push(g, me.ID, "Polygoyf", "Creature — Lhurgoyf", m3cb1PolygoyfOracle, 0, 1)
	pushGraveyardCardWithTypeLine(g, opp.ID, "Grizzly Bears", "Creature — Bear")
	pushGraveyardCardWithTypeLine(g, opp.ID, "Forest", "Basic Land — Forest")
	pushGraveyardCardWithTypeLine(g, me.ID, "Shock", "Instant")
	if p, tg := ptOf(t, g, goyf); p != 3 || tg != 4 {
		t.Fatalf("Polygoyf is %d/%d, want 3/4 (creature, land, instant)", p, tg)
	}
	if !effectiveAbilitiesContain(t, g, goyf, "trample") {
		t.Error("Polygoyf has no trample")
	}
	declareAttack(t, g, opp.ID, goyf)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	answerMayChoice(t, g, me.ID, false)
	copies := s46Tokens(g, me.ID, "Polygoyf")
	if len(copies) != 1 {
		t.Fatalf("%d myriad copies, want the one accepted", len(copies))
	}
	if p, tg := ptOf(t, g, copies[0].InstanceID); p != 3 || tg != 4 {
		t.Errorf("the copy is %d/%d, want 3/4", p, tg)
	}
}

// --- Pyrogoyf ----------------------------------------------------------------

func TestM3cb1PyrogoyfAndEachLhurgoyfYouControlDealTheirPowerOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := m3cb1Opponent(g)
	pushGraveyardCardWithTypeLine(g, opp.ID, "Grizzly Bears", "Creature — Bear")
	pushGraveyardCardWithTypeLine(g, opp.ID, "Forest", "Basic Land — Forest")

	life := opp.Life
	goyf := trekB31Enter(t, g, me, game.Card{Name: "Pyrogoyf", TypeLine: "Creature — Lhurgoyf",
		OracleID: m3cb1PyrogoyfOracle, ManaCost: "{3}{R}", Power: 0, Toughness: 1})
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if p, _ := ptOf(t, g, goyf); p != 2 {
		t.Fatalf("Pyrogoyf has power %d, want 2", p)
	}
	if opp.Life != life-2 {
		t.Fatalf("opponent life %d -> %d, want -2 from Pyrogoyf itself", life, opp.Life)
	}

	life = opp.Life
	trekB31Enter(t, g, me, game.Card{Name: "Test Lhurgoyf", TypeLine: "Creature — Lhurgoyf", Power: 3, Toughness: 3})
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("opponent life %d -> %d, want -3 from the other Lhurgoyf", life, opp.Life)
	}

	life = opp.Life
	trekB31Enter(t, g, me, game.Card{Name: "Test Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	trekB31Enter(t, g, opp, game.Card{Name: "Their Lhurgoyf", TypeLine: "Creature — Lhurgoyf", Power: 4, Toughness: 4})
	if latestPickTarget(g, me.ID) != nil || g.Stack.Size() != 0 {
		t.Error("a non-Lhurgoyf or an opponent's Lhurgoyf triggered Pyrogoyf")
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Errorf("opponent life %d -> %d, want unchanged", life, opp.Life)
	}
}

// --- Jyoti, Moag Ancient -----------------------------------------------------

func TestM3cb1JyotiMakesADryadPerCommanderCastAndPumpsLandCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	me.CommanderCasts = map[uuid.UUID]int{uuid.New(): 2}
	trekB31Enter(t, g, me, game.Card{Name: "Jyoti, Moag Ancient", TypeLine: "Legendary Creature — Elemental",
		OracleID: m3cb1JyotiOracle, ManaCost: "{2}{G}{U}", Power: 2, Toughness: 4})
	passPriorityAroundTable(t, g)
	dryads := s46Tokens(g, me.ID, "Forest Dryad")
	if len(dryads) != 2 {
		t.Fatalf("%d Forest Dryads, want one per commander cast (2)", len(dryads))
	}
	for _, d := range dryads {
		if !d.IsLand() || !d.IsCreature() {
			t.Errorf("a Dryad is not a land creature: %q", d.TypeLine)
		}
	}
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := m3cb1Push(g, opp.ID, game.Card{Name: "Their Dryad", TypeLine: "Land Creature — Forest Dryad", Power: 1, Toughness: 1})

	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	for _, d := range dryads {
		if p, tg := ptOf(t, g, d.InstanceID); p != 3 || tg != 3 {
			t.Errorf("a Dryad is %d/%d, want 3/3 (+2/+2 for Jyoti's power)", p, tg)
		}
	}
	if p, tg := ptOf(t, g, bear); p != 2 || tg != 2 {
		t.Errorf("a non-land creature was pumped to %d/%d", p, tg)
	}
	if p, tg := ptOf(t, g, theirs); p != 1 || tg != 1 {
		t.Errorf("an opponent's land creature was pumped to %d/%d", p, tg)
	}
}

func TestM3cb1JyotiMakesNoDryadsWithoutACommanderCast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	trekB31Enter(t, g, me, game.Card{Name: "Jyoti, Moag Ancient", TypeLine: "Legendary Creature — Elemental",
		OracleID: m3cb1JyotiOracle, ManaCost: "{2}{G}{U}", Power: 2, Toughness: 4})
	passPriorityAroundTable(t, g)
	if n := len(s46Tokens(g, me.ID, "Forest Dryad")); n != 0 {
		t.Errorf("%d Forest Dryads with no commander cast, want 0", n)
	}
}

// --- Horizon of Progress -----------------------------------------------------

func TestM3cb1HorizonOfProgressPaysLifeForAnyTypeYourLandsCouldMake(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := m3cb1Opponent(g)
	horizon := pushCatalogPermanent(g, me.ID, "Horizon of Progress", "Land", m3cb1HorizonOracle, false)
	seedManaLand(g, me.ID, "Wastes", "Basic Land — Wastes", "C")
	seedManaLand(g, me.ID, "Plains", "Basic Land — Plains", "W")
	seedManaLand(g, opp.ID, "Island", "Basic Land — Island", "U")
	life := me.Life
	if err := g.ActivateManaAbility(me.ID, horizon, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if me.Life != life-1 {
		t.Errorf("life %d -> %d, want 1 paid", life, me.Life)
	}
	pick := manaPickFor(g, me.ID)
	if pick == nil {
		t.Fatal("no mana pick queued")
	}
	want := map[string]bool{"C": true, "W": true}
	if len(pick.ColorOptions) != 2 {
		t.Fatalf("options = %v, want exactly {C, W}", pick.ColorOptions)
	}
	for _, o := range pick.ColorOptions {
		if !want[o] {
			t.Errorf("options = %v, contains %q which no land you control could produce", pick.ColorOptions, o)
		}
	}
}

func TestM3cb1HorizonOfProgressPutsALandFromHandTapped(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	horizon := pushCatalogPermanent(g, me.ID, "Horizon of Progress", "Land", m3cb1HorizonOracle, false)
	forest := handCard(me, "Forest", "Basic Land — Forest")
	floatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, horizon, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	ch := latestChooseCards(g, me.ID)
	if ch == nil || ch.ChooseMin != 0 {
		t.Fatalf("want an optional land pick: %+v", ch)
	}
	if err := g.ResolveChooseCards(ch.ID, me.ID, []uuid.UUID{forest}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if c, ok := battlefieldCard(g, forest); !ok || !c.Tapped {
		t.Errorf("the Forest is on the battlefield %v, tapped %v; want both", ok, c.Tapped)
	}
}

func TestM3cb1HorizonOfProgressSacrificesToDraw(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	horizon := pushCatalogPermanent(g, me.ID, "Horizon of Progress", "Land", m3cb1HorizonOracle, false)
	hand := handSize(me)
	floatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, horizon, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(horizon) {
		t.Error("Horizon of Progress was not sacrificed as the cost")
	}
	passPriorityAroundTable(t, g)
	if got := handSize(me); got != hand+1 {
		t.Errorf("hand %d -> %d, want +1", hand, got)
	}
}
