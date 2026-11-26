package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b400_cards_test.go — Magic 2014, S82 older sets batch 400 (#5563):
// Burning Earth, Fiendslayer Paladin, Pyromancer's Gauntlet, Rod of Ruin,
// Merfolk Spy, Ratchet Bomb, Voracious Wurm, Oath of the Ancient Wood,
// Witchstalker, Lifebane Zombie, Elite Arcanist and Jace's Mindseeker, through
// the real cast, activation, combat and trigger paths.

const (
	m14b400Earth     = "21b8192e-46a2-4a6a-abd5-3a074efa30a9"
	m14b400Paladin   = "3eb02993-b2fc-4185-aa67-08c8f133ce42"
	m14b400Gauntlet  = "5cbb70e7-a52c-4b35-bc31-d9a0a7df893c"
	m14b400Rod       = "c29a04ab-4e14-45d8-a993-3fffd0e2f6eb"
	m14b400Spy       = "d004b481-a1d5-459d-8282-76baa9d13652"
	m14b400Bomb      = "255d752a-8251-4e4c-99d6-4a1648de1474"
	m14b400Wurm      = "2a249c4b-beaa-4e7f-b4f7-496d282327bd"
	m14b400Oath      = "385d70ed-07ab-4e01-a473-4408f4bcf6d0"
	m14b400Witch     = "a5d3af88-2ca6-4e3d-8b6e-67482136a5c9"
	m14b400Lifebane  = "5b31ac7b-7901-484c-b59f-556284fd562c"
	m14b400Arcanist  = "5bc3c4e2-087d-4255-a576-0b17ed1e8c3f"
	m14b400Mindseek  = "8ebf337d-eac0-40de-978c-3a8fc3bbe03c"
	m14b400AirTemple = "9575d7ce-f26d-4b90-87a3-6329e9799572" // a nonbasic land with a {T}: Add {W} ability
)

func TestS82OldB400EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		m14b400Earth: "Burning Earth", m14b400Paladin: "Fiendslayer Paladin", m14b400Gauntlet: "Pyromancer's Gauntlet",
		m14b400Rod: "Rod of Ruin", m14b400Spy: "Merfolk Spy", m14b400Bomb: "Ratchet Bomb",
		m14b400Wurm: "Voracious Wurm", m14b400Oath: "Oath of the Ancient Wood", m14b400Witch: "Witchstalker",
		m14b400Lifebane: "Lifebane Zombie", m14b400Arcanist: "Elite Arcanist", m14b400Mindseek: "Jace's Mindseeker",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Burning Earth ---

func TestS82OldB400BurningEarthPingsAPlayerWhoTapsANonbasicLandButNotABasic(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Burning Earth", "Enchantment", m14b400Earth, 0, 0)
	basic := effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	nonbasic := effectstest.PushCatalogPermanent(g, opp.ID, "Abandoned Air Temple", "Land", m14b400AirTemple, false)
	effectstest.AdvanceToMain(t, g)
	lives := effectstest.B20Lives(g)
	if err := g.ActivateManaAbility(me.ID, basic, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap the basic: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != lives[0] {
		t.Errorf("tapping a basic land cost %d life", lives[0]-me.Life)
	}
	if err := g.ActivateManaAbility(opp.ID, nonbasic, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap the nonbasic: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != lives[1]-1 {
		t.Errorf("opponent life %d → %d, want 1 damage", lives[1], opp.Life)
	}
	if me.Life != lives[0] {
		t.Error("the enchantment's controller was hurt by an opponent's tap")
	}
}

func TestS82OldB400BurningEarthHurtsItsOwnControllerForTappingANonbasicLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Burning Earth", "Enchantment", m14b400Earth, 0, 0)
	nonbasic := effectstest.PushCatalogPermanent(g, me.ID, "Abandoned Air Temple", "Land", m14b400AirTemple, false)
	effectstest.AdvanceToMain(t, g)
	before := me.Life
	if err := g.ActivateManaAbility(me.ID, nonbasic, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before-1 {
		t.Errorf("life %d → %d, want 1 damage", before, me.Life)
	}
}

// --- Fiendslayer Paladin ---

// castAt400 casts a spell that needs a target (Lightning Bolt's text, in the
// given colours) at a card, so the targeting rules have something to refuse.
func castAt400(t *testing.T, g *game.Game, caster *game.Player, name, typeLine string, colors []string, target uuid.UUID) error {
	t.Helper()
	id := uuid.New()
	caster.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: effectstest.LightningBoltOracle, Colors: colors, Owner: caster.ID, Controller: caster.ID})
	return g.CastSpell(caster.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(target)})
}

func TestS82OldB400FiendslayerPaladinCannotBeTargetedByBlackOrRedSpellsOfOpponents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	paladin := effectstest.CastHoldCreature(t, g, "Fiendslayer Paladin", "Creature — Human Knight", m14b400Paladin, 2, 2)
	for _, colors := range [][]string{{"B"}, {"R"}, {"B", "G"}} {
		if err := castAt400(t, g, opp, "Hostile Spell", "Instant", colors, paladin); err == nil {
			t.Errorf("a %v spell of an opponent targeted the Paladin", colors)
		}
	}
	for _, colors := range [][]string{{"U"}, {"G"}, {"W"}, nil} {
		if err := castAt400(t, g, opp, "Fine Spell", "Instant", colors, paladin); err != nil {
			t.Errorf("a %v spell of an opponent was refused: %v", colors, err)
		}
	}
	if err := castAt400(t, g, me, "Own Black Spell", "Instant", []string{"B"}, paladin); err != nil {
		t.Errorf("the controller's own black spell was refused: %v", err)
	}
}

// --- Pyromancer's Gauntlet ---

func TestS82OldB400PyromancersGauntletAddsTwoToRedInstantsAndSorceriesAndPlaneswalkers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B27Push(g, me.ID, "Pyromancer's Gauntlet", "Artifact", m14b400Gauntlet, "{5}", 0, 0)
	effectstest.AdvanceToMain(t, g)
	before := opp.Life
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Lightning Bolt", TypeLine: "Instant", OracleID: effectstest.LightningBoltOracle,
		ManaCost: "{R}", Colors: []string{"R"}, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{effectstest.PbPlayer(opp.ID)}}); err != nil {
		t.Fatalf("cast the Bolt: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-5 {
		t.Errorf("a red Bolt dealt %d, want 5", before-opp.Life)
	}
}

func TestS82OldB400PyromancersGauntletLeavesOtherSourcesAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B27Push(g, me.ID, "Pyromancer's Gauntlet", "Artifact", m14b400Gauntlet, "{5}", 0, 0)
	effectstest.AdvanceToMain(t, g)
	// A red creature is not an instant, sorcery or planeswalker.
	redCreature := effectstest.B27Push(g, me.ID, "Red Bear", "Creature — Bear", "", "{1}{R}", 2, 2, "R")
	before := opp.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(redCreature, opp.ID, 2) })
	if opp.Life != before-2 {
		t.Errorf("a red creature dealt %d, want 2", before-opp.Life)
	}
	// A non-red instant, and a red instant the opponent controls.
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Black Bolt", TypeLine: "Instant", OracleID: effectstest.LightningBoltOracle,
		ManaCost: "{B}", Colors: []string{"B"}, Owner: me.ID, Controller: me.ID})
	before = opp.Life
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{effectstest.PbPlayer(opp.ID)}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-3 {
		t.Errorf("a black Bolt dealt %d, want 3", before-opp.Life)
	}
}

// --- Rod of Ruin ---

func TestS82OldB400RodOfRuinPingsAnyTargetForThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rod := effectstest.PushCatalogPermanent(g, me.ID, "Rod of Ruin", "Artifact", m14b400Rod, false)
	bear := effectstest.B12Creature(g, opp.ID, "Elf", "Creature — Elf", 1, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	effectstest.B16Activate(t, g, me.ID, rod, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the 1/1 survived")
	}
	if !effectstest.B16Tapped(t, g, rod) {
		t.Error("the {T} cost was not paid")
	}
	effectstest.UntapForTest(g, rod)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, rod, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{effectstest.PbPlayer(opp.ID)}}); err == nil {
		t.Error("activated with two mana")
	}
}

// --- Merfolk Spy ---

func TestS82OldB400MerfolkSpyRevealsARandomCardFromThePlayerItHit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spy := effectstest.B12Push(g, me.ID, "Merfolk Spy", "Creature — Merfolk Rogue", m14b400Spy, 1, 1)
	g.WithWriteLock(func() { opp.Hand.Cards = nil })
	hand := []uuid.UUID{
		effectstest.PutInHand(opp, game.Card{Name: "A", TypeLine: "Sorcery"}),
		effectstest.PutInHand(opp, game.Card{Name: "B", TypeLine: "Sorcery"}),
		effectstest.PutInHand(opp, game.Card{Name: "C", TypeLine: "Sorcery"}),
	}
	effectstest.BrAttack(t, g, spy)
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	revealed := map[uuid.UUID]bool{}
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind == game.EventRevealCards {
			revealed[ev.CardID] = true
		}
	}
	if len(revealed) != 1 {
		t.Fatalf("%d cards revealed, want exactly 1", len(revealed))
	}
	for id := range revealed {
		if !(id == hand[0] || id == hand[1] || id == hand[2]) {
			t.Errorf("the revealed card %v is not from their hand", id)
		}
	}
	if opp.Hand.Size() != 3 {
		t.Errorf("their hand has %d cards, want 3: revealing is not discarding", opp.Hand.Size())
	}
}

func TestS82OldB400MerfolkSpyWithAnEmptyHandRevealsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spy := effectstest.B12Push(g, me.ID, "Merfolk Spy", "Creature — Merfolk Rogue", m14b400Spy, 1, 1)
	g.WithWriteLock(func() { opp.Hand.Cards = nil })
	effectstest.BrAttack(t, g, spy)
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind == game.EventRevealCards {
			t.Error("a card was revealed from an empty hand")
		}
	}
}

// --- Ratchet Bomb ---

func TestS82OldB400RatchetBombDestroysNonlandPermanentsWithManaValueEqualToItsCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bomb := effectstest.PushCatalogPermanent(g, me.ID, "Ratchet Bomb", "Artifact", m14b400Bomb, false)
	two := effectstest.B27Push(g, opp.ID, "Two", "Creature — Bear", "", "{1}{G}", 2, 2, "G")
	three := effectstest.B27Push(g, opp.ID, "Three", "Creature — Bear", "", "{2}{G}", 3, 3, "G")
	land := effectstest.B27Push(g, opp.ID, "Rock Land", "Land", "", "", 0, 0)
	effectstest.AdvanceToMain(t, g)
	for i := 0; i < 2; i++ {
		effectstest.UntapForTest(g, bomb)
		effectstest.B16Activate(t, g, me.ID, bomb, 0, game.ActivateAbilityParams{})
		effectstest.PassPriorityAroundTable(t, g)
	}
	if n := effectstest.BattlefieldCountersOf(g, bomb, game.CounterCharge); n != 2 {
		t.Fatalf("%d charge counters, want 2", n)
	}
	effectstest.UntapForTest(g, bomb)
	effectstest.B16Activate(t, g, me.ID, bomb, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(two) {
		t.Error("the mana value 2 creature survived")
	}
	if !g.Battlefield.Contains(three) || !g.Battlefield.Contains(land) {
		t.Error("a permanent of another mana value, or a land, was destroyed")
	}
	if g.Battlefield.Contains(bomb) {
		t.Error("the Bomb was not sacrificed")
	}
}

func TestS82OldB400RatchetBombWithNoCountersDestroysTokensAndSparesLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bomb := effectstest.PushCatalogPermanent(g, me.ID, "Ratchet Bomb", "Artifact", m14b400Bomb, false)
	tok := effectstest.PushToken(g, opp.ID, effects.TokenCard("1/1 black Rat"))
	one := effectstest.B27Push(g, opp.ID, "One", "Creature — Bear", "", "{1}", 1, 1)
	land := effectstest.B27Push(g, opp.ID, "Rock Land", "Land", "", "", 0, 0)
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, bomb, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tok) {
		t.Error("the token (mana value 0) survived")
	}
	if !g.Battlefield.Contains(one) || !g.Battlefield.Contains(land) {
		t.Error("the mana value 1 creature or the land was destroyed")
	}
}

// --- Voracious Wurm ---

func TestS82OldB400VoraciousWurmEntersWithCountersForLifeGainedThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.GptB1Seats(g)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() {
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3)
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -1)
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 2)
	})
	id := effectstest.CastHoldCreature(t, g, "Voracious Wurm", "Creature — Wurm", m14b400Wurm, 2, 2)
	if n := effectstest.CounterCount(g, id, game.CounterPlusOne); n != 5 {
		t.Errorf("%d +1/+1 counters, want 5 (3 + 2 gained; the 1 lost does not reduce it)", n)
	}
}

func TestS82OldB400VoraciousWurmEntersBareWithNoLifeGained(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.CastHoldCreature(t, g, "Voracious Wurm", "Creature — Wurm", m14b400Wurm, 2, 2)
	if n := effectstest.CounterCount(g, id, game.CounterPlusOne); n != 0 {
		t.Errorf("%d +1/+1 counters, want 0", n)
	}
}

// --- Oath of the Ancient Wood ---

func TestS82OldB400OathPutsACounterOnTargetCreatureWhenAnEnchantmentEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Oath of the Ancient Wood", "Enchantment", m14b400Oath, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickTriggerTarget(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, bear, game.CounterPlusOne); n != 1 {
		t.Errorf("%d counters on the target, want 1 from the Oath's own entry", n)
	}
}

func TestS82OldB400OathTriggersForAnotherEnchantmentYouControlButNotAnOpponents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Oath of the Ancient Wood", "Enchantment", m14b400Oath, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	// An opponent's enchantment does nothing.
	effectstest.B12Push(g, opp.ID, "Their Enchantment", "Enchantment", "", 0, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("an opponent's enchantment entering raised the trigger")
	}
	// Mine does.
	effectstest.CastCatalogSpell(t, g, "My Enchantment", "Enchantment", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickTriggerTarget(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, bear, game.CounterPlusOne); n != 1 {
		t.Errorf("%d counters, want 1", n)
	}
}

// --- Witchstalker ---

func TestS82OldB400WitchstalkerGrowsOnlyForBlueOrBlackSpellsOfOpponentsDuringYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	stalker := effectstest.B27Push(g, me.ID, "Witchstalker", "Creature — Wolf", m14b400Witch, "{1}{G}{G}", 3, 3, "G")
	effectstest.AdvanceToMain(t, g)
	cast := func(caster *game.Player, colors []string) {
		id := uuid.New()
		caster.Hand.PushTop(game.Card{InstanceID: id, Name: "Spell", TypeLine: "Instant", Colors: colors, Owner: caster.ID, Controller: caster.ID})
		if err := g.CastSpell(caster.ID, id, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	cast(opp, []string{"G"})
	cast(opp, []string{"R"})
	cast(me, []string{"U"})
	if n := effectstest.CounterCount(g, stalker, game.CounterPlusOne); n != 0 {
		t.Fatalf("%d counters after spells that should not count, want 0", n)
	}
	cast(opp, []string{"U"})
	cast(opp, []string{"B"})
	cast(opp, []string{"U", "B"})
	if n := effectstest.CounterCount(g, stalker, game.CounterPlusOne); n != 3 {
		t.Errorf("%d counters, want 3", n)
	}
}

func TestS82OldB400WitchstalkerDoesNotGrowDuringAnOpponentsTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	stalker := effectstest.B27Push(g, me.ID, "Witchstalker", "Creature — Wolf", m14b400Witch, "{1}{G}{G}", 3, 3, "G")
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	third := g.Seats[2]
	third.Hand.PushTop(game.Card{InstanceID: id, Name: "Spell", TypeLine: "Instant", Colors: []string{"U"}, Owner: third.ID, Controller: third.ID})
	if err := g.CastSpell(third.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	_ = opp
	if n := effectstest.CounterCount(g, stalker, game.CounterPlusOne); n != 0 {
		t.Errorf("%d counters from a spell cast on another player's turn, want 0", n)
	}
}

func TestS82OldB400WitchstalkerHasHexproof(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	stalker := effectstest.CastHoldCreature(t, g, "Witchstalker", "Creature — Wolf", m14b400Witch, 3, 3)
	if err := castAt400(t, g, opp, "Hostile Spell", "Instant", []string{"G"}, stalker); err == nil {
		t.Error("an opponent's spell targeted the Witchstalker")
	}
}

// --- Lifebane Zombie ---

func TestS82OldB400LifebaneZombieExilesAGreenOrWhiteCreatureCardFromTheirHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hand := effectstest.RevealHand(opp,
		game.Card{Name: "Green Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Colors: []string{"G"}},
		game.Card{Name: "White Knight", TypeLine: "Creature — Knight", ManaCost: "{1}{W}", Colors: []string{"W"}},
		game.Card{Name: "Black Rat", TypeLine: "Creature — Rat", ManaCost: "{1}{B}", Colors: []string{"B"}},
		game.Card{Name: "Green Spell", TypeLine: "Sorcery", ManaCost: "{G}", Colors: []string{"G"}})
	effectstest.CastHoldCreatureNoResolve(t, g, "Lifebane Zombie", "Creature — Zombie Warrior", m14b400Lifebane, 3, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickTriggerTargetPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.OpenVariant(t, g)
	if !effectstest.SameIDs(pick.DiscardOptions, []uuid.UUID{hand[0], hand[1]}) {
		t.Fatalf("options = %v, want only the green and the white creature card", pick.DiscardOptions)
	}
	if err := g.ResolvePendingChoice(pick.ID, me.ID, []uuid.UUID{hand[1]}); err != nil {
		t.Fatalf("resolve: %v", err)
	}
	g.SettleResolution()
	if !g.Exile.Contains(hand[1]) {
		t.Error("the chosen card was not exiled")
	}
	if !opp.Hand.Contains(hand[0]) || !opp.Hand.Contains(hand[2]) || !opp.Hand.Contains(hand[3]) {
		t.Error("another card left their hand")
	}
}

func TestS82OldB400LifebaneZombieWithNoGreenOrWhiteCreatureCardExilesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.RevealHand(opp,
		game.Card{Name: "Black Rat", TypeLine: "Creature — Rat", ManaCost: "{1}{B}", Colors: []string{"B"}},
		game.Card{Name: "Green Spell", TypeLine: "Sorcery", ManaCost: "{G}", Colors: []string{"G"}})
	effectstest.CastHoldCreatureNoResolve(t, g, "Lifebane Zombie", "Creature — Zombie Warrior", m14b400Lifebane, 3, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickTriggerTargetPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Size() != 2 {
		t.Errorf("their hand has %d cards, want 2", opp.Hand.Size())
	}
}

// --- Elite Arcanist ---

func TestS82OldB400EliteArcanistImprintsAnInstantAndCopiesItForItsManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	bolt := effectstest.PutInHand(me, game.Card{Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}", OracleID: effectstest.LightningBoltOracle, Colors: []string{"R"}})
	big := effectstest.PutInHand(me, game.Card{Name: "Big Instant", TypeLine: "Instant", ManaCost: "{2}{U}{U}", Colors: []string{"U"}})
	sorcery := effectstest.PutInHand(me, game.Card{Name: "A Sorcery", TypeLine: "Sorcery", ManaCost: "{1}"})
	arcanist := effectstest.CastHoldCreatureNoResolve(t, g, "Elite Arcanist", "Creature — Human Wizard", m14b400Arcanist, 1, 1)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no imprint prompt")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range c.ChooseCards {
		offered[id] = true
	}
	if !offered[bolt] || !offered[big] || offered[sorcery] {
		t.Errorf("imprint offers %v, want both instants and not the sorcery", c.ChooseCards)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, bolt)
	if !effectstest.InExile(g, bolt) {
		t.Fatal("the imprinted card is not in exile")
	}
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, arcanist).SummonedThisTurn = false })
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.SweepFAddMana(g, me.ID, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, arcanist, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate for the Bolt's mana value of 1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	cp := effectstest.CopyOf(t, g, "Lightning Bolt")
	before := opp.Life
	if err := effectstest.CastCopy(g, me.ID, cp.InstanceID, true, effectstest.PlayerTarget(opp.ID)); err != nil {
		t.Fatalf("casting the copy free: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-3 {
		t.Errorf("opponent life %d → %d, want the copy to deal 3", before, opp.Life)
	}
	if !effectstest.InExile(g, bolt) {
		t.Error("the imprinted card left exile")
	}
}

func TestS82OldB400EliteArcanistActivationCostsTheManaValueOfTheExiledCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	big := effectstest.PutInHand(me, game.Card{Name: "Big Instant", TypeLine: "Instant", ManaCost: "{2}{U}{U}", Colors: []string{"U"}})
	arcanist := effectstest.CastHoldCreatureNoResolve(t, g, "Elite Arcanist", "Creature — Human Wizard", m14b400Arcanist, 1, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, big)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, arcanist).SummonedThisTurn = false })
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.SweepFAddMana(g, me.ID, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, arcanist, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated for 3 mana when the exiled card has mana value 4")
	}
}

// --- Jace's Mindseeker ---

func TestS82OldB400JacesMindseekerMillsFiveAndOffersAFreeInstantOrSorceryFromAmongThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	opp.Library.Cards = nil
	bolt := effectstest.PushLibraryCardForTest(opp, game.Card{Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}", OracleID: effectstest.LightningBoltOracle, Colors: []string{"R"}})
	var milledOthers []uuid.UUID
	for i := 0; i < 3; i++ {
		milledOthers = append(milledOthers, effectstest.PushLibraryCardForTest(opp, game.Card{Name: "Filler", TypeLine: "Creature — Bear", Power: 1, Toughness: 1}))
	}
	milledOthers = append(milledOthers, effectstest.PushLibraryCardForTest(opp, game.Card{Name: "Land", TypeLine: "Basic Land — Forest"}))
	deepSorcery := effectstest.PushLibraryCardForTest(opp, game.Card{Name: "Deep Sorcery", TypeLine: "Sorcery", ManaCost: "{1}"})
	oldGrave := uuid.New()
	g.WithWriteLock(func() {
		opp.Graveyard.PushTop(game.Card{InstanceID: oldGrave, Name: "Old Instant", TypeLine: "Instant", ManaCost: "{1}", Owner: opp.ID, Controller: opp.ID})
	})
	effectstest.CastHoldCreatureNoResolve(t, g, "Jace's Mindseeker", "Creature — Fish Illusion", m14b400Mindseek, 4, 4)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickTriggerTargetPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range append(milledOthers, bolt) {
		if !opp.Graveyard.Contains(id) {
			t.Fatalf("%v was not milled", id)
		}
	}
	if opp.Graveyard.Contains(deepSorcery) {
		t.Fatal("a sixth card was milled")
	}
	effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, bolt)
	before := opp.Life
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, bolt, game.CastSpellParams{
		Strict: true, AlternativeCost: game.AltCostKeyGrantedCost, Targets: []game.TargetRef{effectstest.PbPlayer(opp.ID)}}); err != nil {
		t.Fatalf("free cast from among the milled cards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-3 {
		t.Errorf("the Bolt dealt %d, want 3", before-opp.Life)
	}
}

func TestS82OldB400JacesMindseekerOffersNothingWhenNoInstantOrSorceryWasMilled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	opp.Library.Cards = nil
	for i := 0; i < 6; i++ {
		effectstest.PushLibraryCardForTest(opp, game.Card{Name: "Filler", TypeLine: "Creature — Bear", Power: 1, Toughness: 1})
	}
	effectstest.CastHoldCreatureNoResolve(t, g, "Jace's Mindseeker", "Creature — Fish Illusion", m14b400Mindseek, 4, 4)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickTriggerTargetPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ResolutionCastOfferFor(g, me.ID) != nil {
		t.Error("a free cast was offered with no instant or sorcery among the milled cards")
	}
}
