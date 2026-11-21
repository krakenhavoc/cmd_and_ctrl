package testsr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsr "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/r"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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

// answerUnleash answers the open unleash "may".
func answerUnleash(t *testing.T, g *game.Game, apply bool) {
	t.Helper()
	c := effectstest.PendingOfKind(g, game.PendingChoiceOptionalReplacement)
	if c == nil || len(c.ReplacementEffectIDs) != 1 || game.EntryKeywordOfReplacement(c.ReplacementEffectIDs[0]) != game.KeywordUnleash {
		t.Fatalf("no unleash question is open: %+v", c)
	}
	if err := g.ResolveOptionalReplacement(c.ID, c.Chooser, apply); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
}

// A creature with riot stamped on the card — the deck importer's road,
// no catalog entry — is asked as it enters.
func TestAKeywordOnlyRiotCreatureIsAsked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Zhur-Taa Goblin", TypeLine: "Creature — Goblin Berserker",
		Power: 2, Toughness: 2, Keywords: []string{game.KeywordRiot}, Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerRiotFor(t, g, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOnBattlefield(t, g, id, "haste") || effectstest.CounterCount(g, id, game.CounterPlusOne) != 0 {
		t.Error("the haste answer did not give haste alone")
	}
}

// Rhythm of the Wild: a creature cast under it is asked (CR 614.12), a
// printed-riot one twice (CR 702.136b), and a token is not asked at all.
func TestRhythmOfTheWildGrantsRiot(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Rhythm of the Wild", TypeLine: "Enchantment",
		OracleID: riotRhythmOracle, Owner: me.ID, Controller: me.ID,
	})

	bear := effectstest.CastCatalogSpell(t, g, "Riot Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerRiotFor(t, g, true)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterCount(g, bear, game.CounterPlusOne) != 1 {
		t.Errorf("Rhythm's riot, counter answer: %d counters", effectstest.CounterCount(g, bear, game.CounterPlusOne))
	}

	beast := effectstest.CastCatalogSpell(t, g, "Gruul Beastmaster", "Creature — Human Shaman", riotBeastmasterOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerRiotFor(t, g, true)
	if g.Battlefield.Contains(beast) {
		t.Fatal("a printed riot under Rhythm entered after one question")
	}
	effectstest.AnswerRiotFor(t, g, false)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterCount(g, beast, game.CounterPlusOne) != 1 || !effectstest.HasKeywordOnBattlefield(t, g, beast, "haste") {
		t.Error("two riots answered counter and haste should give both")
	}

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.RedGoblinToken(), 1) })
	if effectstest.PendingOfKind(g, game.PendingChoiceEntryRiot) != nil {
		t.Error("a token was asked for riot — Rhythm grants it to nontoken creatures")
	}
	if !game.IsCatalogCard(riotRhythmOracle) {
		t.Fatal("Rhythm of the Wild is not registered")
	}
}

// Uncivil Unrest's riot is the same grant: an opponent's creature is
// not asked.
func TestUncivilUnrestRiotIsYoursOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Uncivil Unrest", TypeLine: "Enchantment",
		OracleID: effectstest.B28UncivilUnrestOracle, Owner: opp.ID, Controller: opp.ID,
	})
	bear := effectstest.CastCatalogSpell(t, g, "Plain Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PendingOfKind(g, game.PendingChoiceEntryRiot) != nil {
		t.Fatal("an opponent's Uncivil Unrest asked my creature for riot")
	}
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the bear did not enter")
	}
}

// Tesak: another Dog entering is asked unleash, and a creature with a
// counter has haste.
func TestTesakGrantsUnleashToDogsAndHasteToCounteredCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.RiotPush(g, me.ID, "Tesak, Judith's Hellhound", "Legendary Creature — Elemental Dog", unleashTesakOracle, 3, 3, 0)

	dog := effectstest.CastCatalogSpell(t, g, "Hound", "Creature — Dog", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	answerUnleash(t, g, true)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterCount(g, dog, game.CounterPlusOne) != 1 {
		t.Fatalf("the Dog's granted unleash: %d counters", effectstest.CounterCount(g, dog, game.CounterPlusOne))
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, dog, "haste") {
		t.Error("a creature with a counter under Tesak has no haste")
	}
	c := effectstest.FindBattlefieldCardByID(g, dog)
	if !game.Restricted(c, game.CantBlock) {
		t.Error("an unleashed Dog with its counter can block")
	}

	cat := effectstest.CastCatalogSpell(t, g, "Cat", "Creature — Cat", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PendingOfKind(g, game.PendingChoiceOptionalReplacement) != nil {
		t.Error("a non-Dog was asked for unleash")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, cat, "haste") {
		t.Error("a creature with no counter has haste under Tesak")
	}
}

// Exava: OTHER creatures with a +1/+1 counter have haste.
func TestExavaHastesOtherCounteredCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.RiotPush(g, me.ID, "Exava, Rakdos Blood Witch", "Legendary Creature — Human Cleric", unleashExavaOracle, 3, 3, 0)
	with := effectstest.RiotPush(g, me.ID, "Countered", "Creature — Bear", "", 2, 2, 1)
	without := effectstest.RiotPush(g, me.ID, "Plain", "Creature — Bear", "", 2, 2, 0)
	if !effectstest.HasKeywordOnBattlefield(t, g, with, "haste") || effectstest.HasKeywordOnBattlefield(t, g, without, "haste") {
		t.Error("Exava's haste goes to the creature with a +1/+1 counter only")
	}
}

// Chaos Imps: trample while it has a +1/+1 counter.
func TestChaosImpsTrampleWithACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	plain := effectstest.RiotPush(g, me.ID, "Chaos Imps", "Creature — Imp", unleashChaosImpsOID, 6, 5, 0)
	pumped := effectstest.RiotPush(g, me.ID, "Chaos Imps", "Creature — Imp", unleashChaosImpsOID, 6, 5, 1)
	if effectstest.HasKeywordOnBattlefield(t, g, plain, "trample") || !effectstest.HasKeywordOnBattlefield(t, g, pumped, "trample") {
		t.Error("Chaos Imps has trample exactly while it has a +1/+1 counter")
	}
	if c := effectstest.FindBattlefieldCardByID(g, pumped); !game.Restricted(c, game.CantBlock) {
		t.Error("an unleash creature with a +1/+1 counter can block")
	}
}

// Skarrgan Hellkite: the divided ping needs a +1/+1 counter.
func TestSkarrganHellkiteNeedsACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bare := effectstest.RiotPush(g, me.ID, "Skarrgan Hellkite", "Creature — Dragon", riotSkarrganOracle, 4, 4, 0)
	effectstest.FillPoolColored(me, "R", 4)
	if err := g.ActivateCatalogAbility(me.ID, bare, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}, Distribution: map[uuid.UUID]int{opp.ID: 2},
	}); err == nil {
		t.Fatal("a Hellkite with no +1/+1 counter activated its ping")
	}
	kite := effectstest.RiotPush(g, me.ID, "Skarrgan Hellkite", "Creature — Dragon", riotSkarrganOracle, 4, 4, 1)
	before := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, kite, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}, Distribution: map[uuid.UUID]int{opp.ID: 2},
	}); err != nil {
		t.Fatalf("activate with a counter: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("the ping dealt %d, want 2", before-opp.Life)
	}
}

// Hellhole Flailer: sacrificed, it deals its last-known power — the
// unleash counter included.
func TestHellholeFlailerDealsItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	flailer := effectstest.RiotPush(g, me.ID, "Hellhole Flailer", "Creature — Ogre Warrior", unleashFlailerOracle, 3, 2, 1)
	effectstest.FillPoolColored(me, "B", 2)
	effectstest.FillPoolColored(me, "R", 2)
	before := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, flailer, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-4 {
		t.Errorf("the Flailer dealt %d, want its 4 power", before-opp.Life)
	}
	if g.Battlefield.Contains(flailer) {
		t.Error("the Flailer was not sacrificed")
	}
}

// Grim Roustabout regenerates.
func TestGrimRoustaboutRegenerates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.RiotPush(g, me.ID, "Grim Roustabout", "Creature — Skeleton Warrior", unleashRoustabout, 1, 1, 0)
	effectstest.FillPoolColored(me, "B", 2)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.FindBattlefieldCardByID(g, id); c == nil || c.RegenerationShields != 1 {
		t.Errorf("no regeneration shield: %+v", c)
	}
}

// Frenzied Arynx's pump.
func TestFrenziedArynxPumps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.RiotPush(g, me.ID, "Frenzied Arynx", "Creature — Cat Beast", riotArynxOracle, 3, 3, 0)
	effectstest.FillPoolColored(me, "R", 5)
	effectstest.FillPoolColored(me, "G", 1)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, id); p != 6 || tt != 3 {
		t.Errorf("pumped to %d/%d, want 6/3", p, tt)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, id, "trample") {
		t.Error("Frenzied Arynx has no trample")
	}
}

// Gruul Beastmaster gives another creature +X/+0, X its power.
func TestGruulBeastmasterPumpsByItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	beast := effectstest.RiotPush(g, me.ID, "Gruul Beastmaster", "Creature — Human Shaman", riotBeastmasterOracle, 2, 2, 1)
	other := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, beast)
	if effectstest.OpenPickTarget(g) {
		effectstest.AnswerPickTarget(t, g, other)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, other); p != 5 {
		t.Errorf("the bear's power is %d, want 2+3", p)
	}
}

// Clamor Shaman stops an opposing creature blocking.
func TestClamorShamanStopsABlocker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shaman := effectstest.RiotPush(g, me.ID, "Clamor Shaman", "Creature — Goblin Shaman", riotClamorOracle, 1, 1, 0)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, shaman)
	if effectstest.OpenPickTarget(g) {
		effectstest.AnswerPickTarget(t, g, theirs)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.FindBattlefieldCardByID(g, theirs); !game.Restricted(c, game.CantBlock) {
		t.Error("the opposing bear can still block")
	}
}

// Burning-Tree Vandal's rummage is offered on its attack.
func TestBurningTreeVandalOffersARummage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vandal := effectstest.RiotPush(g, me.ID, "Burning-Tree Vandal", "Creature — Human Rogue", riotVandalOracle, 2, 1, 0)
	effectstest.DeclareAttack(t, g, opp.ID, vandal)
	effectstest.PassPriorityAroundTable(t, g)
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
	if cardsr.LandWithNonManaActivatedAbility(nil, uuid.Nil, plain) {
		t.Error("a basic land has a non-mana activated ability")
	}
	wilds := game.Card{Name: "Evolving Wilds", TypeLine: "Land", OracleID: evolvingWildsOracleForRiot()}
	if !cardsr.LandWithNonManaActivatedAbility(nil, uuid.Nil, wilds) {
		t.Error("Evolving Wilds' sacrifice-and-search is a non-mana activated ability")
	}
	if !game.IsCatalogCard(riotRavagerOracle) {
		t.Error("Ravager Wurm is not registered")
	}
}

// evolvingWildsOracleForRiot reads Evolving Wilds' oracle ID off the
// catalog by name, so this file names no second copy of it.
func evolvingWildsOracleForRiot() string {
	for _, s := range effects.All() {
		if s.Name == "Evolving Wilds" {
			return s.OracleID
		}
	}
	return ""
}
