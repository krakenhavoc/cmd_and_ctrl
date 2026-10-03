package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// riot_test.go — the riot and unleash cards of ADR 0109 §10 (#1556).
// The mechanic itself is pinned in game/riot_test.go.

const (
	riotRhythmOracle      = "acfa77fd-3610-4f12-9c3c-bd860ce91700"
	riotBeastmasterOracle = "924771b2-8566-4bdf-b089-85c3257b9900"
	riotSkarrganOracle    = "bfbe28a2-b656-4575-8b7e-9dbc22f8d4bc"
	riotRavagerOracle     = "40b872b5-3ec2-4fcc-b152-91966f67c2be"
	riotClamorOracle      = "7fd2dcb2-760d-461d-8e5b-693aacf548cf"
	riotArynxOracle       = "772873f4-74b9-44da-bd69-e4c8de797a26"
	riotVandalOracle      = "5bed3702-984d-436f-bb71-e16eaeffe902"
	unleashTesakOracle    = "4117e428-0c78-4dea-9019-dd1836fbb03e"
	unleashRoustabout     = "1adcc4a8-b2dd-4d9a-bded-188b81c84a10"
	unleashFlailerOracle  = "e13c1703-1184-4a8c-9310-f11e4e673597"
	unleashExavaOracle    = "524249cd-68d9-472a-89cd-5872641ca6de"
	unleashChaosImpsOID   = "31e99241-592d-4be4-9ef3-6dde522b1885"
)

// answerRiotFor answers the open entry_riot prompt, failing when there
// is none.
func answerRiotFor(t *testing.T, g *game.Game, counter bool) {
	t.Helper()
	c := pendingOfKind(g, game.PendingChoiceEntryRiot)
	if c == nil {
		t.Fatal("no riot question is open")
	}
	if err := g.ResolveEntryRiot(c.ID, c.Chooser, counter); err != nil {
		t.Fatalf("ResolveEntryRiot: %v", err)
	}
}

// answerUnleash answers the open unleash "may".
func answerUnleash(t *testing.T, g *game.Game, apply bool) {
	t.Helper()
	c := pendingOfKind(g, game.PendingChoiceOptionalReplacement)
	if c == nil || len(c.ReplacementEffectIDs) != 1 || game.EntryKeywordOfReplacement(c.ReplacementEffectIDs[0]) != game.KeywordUnleash {
		t.Fatalf("no unleash question is open: %+v", c)
	}
	if err := g.ResolveOptionalReplacement(c.ID, c.Chooser, apply); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
}

func riotPush(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int, counters int) uuid.UUID {
	c := game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	}
	if counters > 0 {
		c.Counters = map[string]int{game.CounterPlusOne: counters}
	}
	return pushBattlefieldCardWithTimestamp(g, c)
}

// A creature with riot stamped on the card — the deck importer's road,
// no catalog entry — is asked as it enters.
func TestAKeywordOnlyRiotCreatureIsAsked(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Zhur-Taa Goblin", TypeLine: "Creature — Goblin Berserker",
		Power: 2, Toughness: 2, Keywords: []string{game.KeywordRiot}, Owner: me.ID, Controller: me.ID,
	})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	answerRiotFor(t, g, false)
	passPriorityAroundTable(t, g)
	if !hasKeywordOnBattlefield(t, g, id, "haste") || counterCount(g, id, game.CounterPlusOne) != 0 {
		t.Error("the haste answer did not give haste alone")
	}
}

// Rhythm of the Wild: a creature cast under it is asked (CR 614.12), a
// printed-riot one twice (CR 702.136b), and a token is not asked at all.
func TestRhythmOfTheWildGrantsRiot(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Rhythm of the Wild", TypeLine: "Enchantment",
		OracleID: riotRhythmOracle, Owner: me.ID, Controller: me.ID,
	})

	bear := castCatalogSpell(t, g, "Riot Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	answerRiotFor(t, g, true)
	passPriorityAroundTable(t, g)
	if counterCount(g, bear, game.CounterPlusOne) != 1 {
		t.Errorf("Rhythm's riot, counter answer: %d counters", counterCount(g, bear, game.CounterPlusOne))
	}

	beast := castCatalogSpell(t, g, "Gruul Beastmaster", "Creature — Human Shaman", riotBeastmasterOracle, nil)
	passPriorityAroundTable(t, g)
	answerRiotFor(t, g, true)
	if g.Battlefield.Contains(beast) {
		t.Fatal("a printed riot under Rhythm entered after one question")
	}
	answerRiotFor(t, g, false)
	passPriorityAroundTable(t, g)
	if counterCount(g, beast, game.CounterPlusOne) != 1 || !hasKeywordOnBattlefield(t, g, beast, "haste") {
		t.Error("two riots answered counter and haste should give both")
	}

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, RedGoblinToken(), 1) })
	if pendingOfKind(g, game.PendingChoiceEntryRiot) != nil {
		t.Error("a token was asked for riot — Rhythm grants it to nontoken creatures")
	}
	if !game.IsCatalogCard(riotRhythmOracle) {
		t.Fatal("Rhythm of the Wild is not registered")
	}
}

// Uncivil Unrest's riot is the same grant: an opponent's creature is
// not asked.
func TestUncivilUnrestRiotIsYoursOnly(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Uncivil Unrest", TypeLine: "Enchantment",
		OracleID: b28UncivilUnrestOracle, Owner: opp.ID, Controller: opp.ID,
	})
	bear := castCatalogSpell(t, g, "Plain Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if pendingOfKind(g, game.PendingChoiceEntryRiot) != nil {
		t.Fatal("an opponent's Uncivil Unrest asked my creature for riot")
	}
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the bear did not enter")
	}
}

// Tesak: another Dog entering is asked unleash, and a creature with a
// counter has haste.
func TestTesakGrantsUnleashToDogsAndHasteToCounteredCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	riotPush(g, me.ID, "Tesak, Judith's Hellhound", "Legendary Creature — Elemental Dog", unleashTesakOracle, 3, 3, 0)

	dog := castCatalogSpell(t, g, "Hound", "Creature — Dog", "", nil)
	passPriorityAroundTable(t, g)
	answerUnleash(t, g, true)
	passPriorityAroundTable(t, g)
	if counterCount(g, dog, game.CounterPlusOne) != 1 {
		t.Fatalf("the Dog's granted unleash: %d counters", counterCount(g, dog, game.CounterPlusOne))
	}
	if !hasKeywordOnBattlefield(t, g, dog, "haste") {
		t.Error("a creature with a counter under Tesak has no haste")
	}
	c := findBattlefieldCardByID(g, dog)
	if !game.Restricted(c, game.CantBlock) {
		t.Error("an unleashed Dog with its counter can block")
	}

	cat := castCatalogSpell(t, g, "Cat", "Creature — Cat", "", nil)
	passPriorityAroundTable(t, g)
	if pendingOfKind(g, game.PendingChoiceOptionalReplacement) != nil {
		t.Error("a non-Dog was asked for unleash")
	}
	if hasKeywordOnBattlefield(t, g, cat, "haste") {
		t.Error("a creature with no counter has haste under Tesak")
	}
}

// Exava: OTHER creatures with a +1/+1 counter have haste.
func TestExavaHastesOtherCounteredCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	riotPush(g, me.ID, "Exava, Rakdos Blood Witch", "Legendary Creature — Human Cleric", unleashExavaOracle, 3, 3, 0)
	with := riotPush(g, me.ID, "Countered", "Creature — Bear", "", 2, 2, 1)
	without := riotPush(g, me.ID, "Plain", "Creature — Bear", "", 2, 2, 0)
	if !hasKeywordOnBattlefield(t, g, with, "haste") || hasKeywordOnBattlefield(t, g, without, "haste") {
		t.Error("Exava's haste goes to the creature with a +1/+1 counter only")
	}
}

// Chaos Imps: trample while it has a +1/+1 counter.
func TestChaosImpsTrampleWithACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	plain := riotPush(g, me.ID, "Chaos Imps", "Creature — Imp", unleashChaosImpsOID, 6, 5, 0)
	pumped := riotPush(g, me.ID, "Chaos Imps", "Creature — Imp", unleashChaosImpsOID, 6, 5, 1)
	if hasKeywordOnBattlefield(t, g, plain, "trample") || !hasKeywordOnBattlefield(t, g, pumped, "trample") {
		t.Error("Chaos Imps has trample exactly while it has a +1/+1 counter")
	}
	if c := findBattlefieldCardByID(g, pumped); !game.Restricted(c, game.CantBlock) {
		t.Error("an unleash creature with a +1/+1 counter can block")
	}
}

// Skarrgan Hellkite: the divided ping needs a +1/+1 counter.
func TestSkarrganHellkiteNeedsACounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bare := riotPush(g, me.ID, "Skarrgan Hellkite", "Creature — Dragon", riotSkarrganOracle, 4, 4, 0)
	fillPoolColored(me, "R", 4)
	if err := g.ActivateCatalogAbility(me.ID, bare, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}, Distribution: map[uuid.UUID]int{opp.ID: 2},
	}); err == nil {
		t.Fatal("a Hellkite with no +1/+1 counter activated its ping")
	}
	kite := riotPush(g, me.ID, "Skarrgan Hellkite", "Creature — Dragon", riotSkarrganOracle, 4, 4, 1)
	before := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, kite, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}, Distribution: map[uuid.UUID]int{opp.ID: 2},
	}); err != nil {
		t.Fatalf("activate with a counter: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("the ping dealt %d, want 2", before-opp.Life)
	}
}

// Hellhole Flailer: sacrificed, it deals its last-known power — the
// unleash counter included.
func TestHellholeFlailerDealsItsPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	flailer := riotPush(g, me.ID, "Hellhole Flailer", "Creature — Ogre Warrior", unleashFlailerOracle, 3, 2, 1)
	fillPoolColored(me, "B", 2)
	fillPoolColored(me, "R", 2)
	before := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, flailer, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != before-4 {
		t.Errorf("the Flailer dealt %d, want its 4 power", before-opp.Life)
	}
	if g.Battlefield.Contains(flailer) {
		t.Error("the Flailer was not sacrificed")
	}
}

// Grim Roustabout regenerates.
func TestGrimRoustaboutRegenerates(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	id := riotPush(g, me.ID, "Grim Roustabout", "Creature — Skeleton Warrior", unleashRoustabout, 1, 1, 0)
	fillPoolColored(me, "B", 2)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c := findBattlefieldCardByID(g, id); c == nil || c.RegenerationShields != 1 {
		t.Errorf("no regeneration shield: %+v", c)
	}
}

// Frenzied Arynx's pump.
func TestFrenziedArynxPumps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	id := riotPush(g, me.ID, "Frenzied Arynx", "Creature — Cat Beast", riotArynxOracle, 3, 3, 0)
	fillPoolColored(me, "R", 5)
	fillPoolColored(me, "G", 1)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tt := ptOf(t, g, id); p != 6 || tt != 3 {
		t.Errorf("pumped to %d/%d, want 6/3", p, tt)
	}
	if !hasKeywordOnBattlefield(t, g, id, "trample") {
		t.Error("Frenzied Arynx has no trample")
	}
}

// Gruul Beastmaster gives another creature +X/+0, X its power.
func TestGruulBeastmasterPumpsByItsPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	beast := riotPush(g, me.ID, "Gruul Beastmaster", "Creature — Human Shaman", riotBeastmasterOracle, 2, 2, 1)
	other := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	declareAttack(t, g, opp.ID, beast)
	if openPickTarget(g) {
		answerPickTarget(t, g, other)
	}
	passPriorityAroundTable(t, g)
	if p, _ := ptOf(t, g, other); p != 5 {
		t.Errorf("the bear's power is %d, want 2+3", p)
	}
}

// Clamor Shaman stops an opposing creature blocking.
func TestClamorShamanStopsABlocker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shaman := riotPush(g, me.ID, "Clamor Shaman", "Creature — Goblin Shaman", riotClamorOracle, 1, 1, 0)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	declareAttack(t, g, opp.ID, shaman)
	if openPickTarget(g) {
		answerPickTarget(t, g, theirs)
	}
	passPriorityAroundTable(t, g)
	if c := findBattlefieldCardByID(g, theirs); !game.Restricted(c, game.CantBlock) {
		t.Error("the opposing bear can still block")
	}
}

// Burning-Tree Vandal's rummage is offered on its attack.
func TestBurningTreeVandalOffersARummage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vandal := riotPush(g, me.ID, "Burning-Tree Vandal", "Creature — Human Rogue", riotVandalOracle, 2, 1, 0)
	declareAttack(t, g, opp.ID, vandal)
	passPriorityAroundTable(t, g)
	if len(g.PendingChoices) == 0 && me.Hand.Size() > 0 {
		t.Error("the Vandal's attack offered no discard")
	}
}

// The keyword-only riot and unleash creatures need no card file: with
// both keywords canonical, the importer does not flag their printed
// text (Zhur-Taa Goblin, Ghor-Clan Wrecker, Carnival Hellsteed,
// Thrill-Kill Assassin). A creature with any other line still is.
func TestKeywordOnlyRiotAndUnleashCreaturesNeedNoCatalogEntry(t *testing.T) {
	const (
		riot    = "Riot (This creature enters with your choice of a +1/+1 counter or haste.)"
		unleash = "Unleash (You may have this creature enter with a +1/+1 counter on it. It can't block as long as it has a +1/+1 counter on it.)"
	)
	for _, text := range []string{
		riot,
		riot + "\nMenace (This creature can't be blocked except by two or more creatures.)",
		"First strike, haste\n" + unleash,
		"Deathtouch\n" + unleash,
	} {
		if game.NeedsCatalogEffect("Creature — Test", text) {
			t.Errorf("a keyword-only creature is flagged: %q", text)
		}
	}
	if !game.NeedsCatalogEffect("Creature — Test", unleash+"\n{1}{B}: Regenerate this creature.") {
		t.Error("Grim Roustabout's second line was not flagged")
	}
}

// Ravager Wurm's land bullet only reaches a land with a non-mana
// activated ability.
func TestRavagerWurmLandPredicate(t *testing.T) {
	plain := game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"}
	if landWithNonManaActivatedAbility(nil, uuid.Nil, plain) {
		t.Error("a basic land has a non-mana activated ability")
	}
	wilds := game.Card{Name: "Evolving Wilds", TypeLine: "Land", OracleID: evolvingWildsOracleForRiot()}
	if !landWithNonManaActivatedAbility(nil, uuid.Nil, wilds) {
		t.Error("Evolving Wilds' sacrifice-and-search is a non-mana activated ability")
	}
	if !game.IsCatalogCard(riotRavagerOracle) {
		t.Error("Ravager Wurm is not registered")
	}
}

// evolvingWildsOracleForRiot reads Evolving Wilds' oracle ID off the
// catalog by name, so this file names no second copy of it.
func evolvingWildsOracleForRiot() string {
	for _, s := range All() {
		if s.Name == "Evolving Wilds" {
			return s.OracleID
		}
	}
	return ""
}
