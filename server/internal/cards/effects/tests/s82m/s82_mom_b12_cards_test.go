package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_mom_b12_cards_test.go — S82, March of the Machine batch 12 (#5477).

const (
	s82b12DoombringerOracle = "85c799e9-74d5-459e-b168-b95ed78cd4e9"
	s82b12TrackerOracle     = "ca49d6ad-97b1-4fd4-b1cf-fecd9cafee6d"
	s82b12RaptorOracle      = "450a07f5-709c-48a6-aaa8-ddbd7da4fd2e"
	s82b12HungerOracle      = "fe84ce0b-cd25-41ee-82c1-8ab03df5791b"
	s82b12TandemOracle      = "e8b0edb8-6a69-4760-9bcd-969c6c898a16"
	s82b12SpiteOracle       = "10426070-2e6f-47a5-bc8d-16dcfb058b26"
	s82b12FallOracle        = "22677153-bd87-4ff8-af92-ced4da2aac6c"
	s82b12WarcraftingOracle = "48838239-bec4-4a57-9858-425f0a937dc9"
	s82b12JavelineerOracle  = "e834b281-279f-4df3-9c0e-764502b809c5"
	s82b12GuerrillaOracle   = "24dbddad-998b-4755-b356-4c8aca3592b1"
	s82b12SkaabOracle       = "d7dcd1a3-3163-4e26-94e8-f4072e477e6a"
	s82b12EscortOracle      = "ce807ee3-a27c-4f3b-92a4-37faaca42aa1"
)

func TestS82B12CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s82b12DoombringerOracle: "Etched Host Doombringer",
		s82b12TrackerOracle:     "Portent Tracker",
		s82b12RaptorOracle:      "Rampaging Raptor",
		s82b12HungerOracle:      "Cosmic Hunger",
		s82b12TandemOracle:      "Tandem Takedown",
		s82b12SpiteOracle:       "Volcanic Spite",
		s82b12FallOracle:        "Atraxa's Fall",
		s82b12WarcraftingOracle: "Nahiri's Warcrafting",
		s82b12JavelineerOracle:  "Onakke Javelineer",
		s82b12GuerrillaOracle:   "Flitting Guerrilla",
		s82b12SkaabOracle:       "Halo-Charged Skaab",
		s82b12EscortOracle:      "Wildwood Escort",
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

// s82b12Battle puts a Siege on the battlefield under owner, protected by
// protector, with `defense` defense counters.
func s82b12Battle(g *game.Game, owner, protector uuid.UUID, defense int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test Siege", TypeLine: "Battle — Siege",
		Owner: owner, Controller: owner, ProtectorPlayerID: protector,
		Counters: map[string]int{game.CounterDefense: defense},
	})
}

func s82b12Defense(g *game.Game, id uuid.UUID) int {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return c.Counters[game.CounterDefense]
		}
	}
	return -1
}

// --- Atraxa's Fall -------------------------------------------------------------

func TestS82B12AtraxasFallDestroysTheFourKindsAndRefusesAGroundCreature(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		keywords       []string
	}{
		{"Test Rock", "Artifact", nil},
		{"Test Aura", "Enchantment", nil},
		{"Test Flier", "Creature — Bird", []string{"flying"}},
	} {
		g := effectstest.NewCatalogGame(t)
		opp := g.Seats[1]
		id := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: tc.name, TypeLine: tc.typeLine, Keywords: tc.keywords,
			Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
		})
		effectstest.CastCatalogSpell(t, g, "Atraxa's Fall", "Sorcery", s82b12FallOracle, []game.TargetRef{{Kind: game.TargetCard, ID: id}})
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(id) {
			t.Errorf("%s survived Atraxa's Fall", tc.name)
		}
	}
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	battle := s82b12Battle(g, opp.ID, g.Seats[0].ID, 4)
	effectstest.CastCatalogSpell(t, g, "Atraxa's Fall", "Sorcery", s82b12FallOracle, []game.TargetRef{{Kind: game.TargetCard, ID: battle}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(battle) {
		t.Error("a battle survived Atraxa's Fall")
	}

	g = effectstest.NewCatalogGame(t)
	ground := effectstest.B16Creature(g, g.Seats[1].ID, "Ground Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Atraxa's Fall", "Sorcery", s82b12FallOracle, []game.TargetRef{{Kind: game.TargetCard, ID: ground}}); err == nil {
		t.Error("a creature without flying is a legal target")
	}
}

// --- Onakke Javelineer ---------------------------------------------------------

func TestS82B12OnakkeJavelineerHitsAPlayerOrABattleButNotACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	j := effectstest.B12Push(g, me.ID, "Onakke Javelineer", "Creature — Ogre Spirit", s82b12JavelineerOracle, 5, 4)
	battle := s82b12Battle(g, opp.ID, me.ID, 5)
	effectstest.B16Activate(t, g, me.ID, j, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: battle}}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := s82b12Defense(g, battle); got != 3 {
		t.Errorf("battle defense %d, want 3", got)
	}
	// Tapped now: a second activation is refused, and a creature is no target.
	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	j2 := effectstest.B12Push(g2, me2.ID, "Onakke Javelineer", "Creature — Ogre Spirit", s82b12JavelineerOracle, 5, 4)
	bear := effectstest.B16Creature(g2, opp2.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g2.ActivateCatalogAbility(me2.ID, j2, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err == nil {
		t.Error("a creature was a legal target for the Javelineer")
	}
	life2 := opp2.Life
	effectstest.B16Activate(t, g2, me2.ID, j2, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp2.ID}}})
	effectstest.PassPriorityAroundTable(t, g2)
	if opp2.Life != life2-2 {
		t.Errorf("opponent life %d, want 2 less than %d", opp2.Life, life2)
	}
}

// --- Cosmic Hunger / Tandem Takedown -------------------------------------------

func TestS82B12CosmicHungerBitesACreatureAPlaneswalkerAndABattle(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B16Creature(g, me.ID, "My Beast", "Creature — Beast", 4, 4)
	bear := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	walker := effectstest.P1g2Planeswalker(g, opp.ID, 6)
	battle := s82b12Battle(g, opp.ID, me.ID, 5)
	effectstest.CastCatalogSpell(t, g, "Cosmic Hunger", "Instant", s82b12HungerOracle, effectstest.S82b12Pair(mine, bear))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the Bear took 4 and lived")
	}
	effectstest.CastCatalogSpell(t, g, "Cosmic Hunger", "Instant", s82b12HungerOracle, effectstest.S82b12Pair(mine, walker))
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == walker && c.Counters[game.CounterLoyalty] != 2 {
			t.Errorf("walker loyalty %d, want 2", c.Counters[game.CounterLoyalty])
		}
	}
	effectstest.CastCatalogSpell(t, g, "Cosmic Hunger", "Instant", s82b12HungerOracle, effectstest.S82b12Pair(mine, battle))
	effectstest.PassPriorityAroundTable(t, g)
	if got := s82b12Defense(g, battle); got != 1 {
		t.Errorf("battle defense %d, want 1", got)
	}
	// The same creature can't be both targets.
	if err := effectstest.CastCatalogSpellErr(t, g, "Cosmic Hunger", "Instant", s82b12HungerOracle, effectstest.S82b12Pair(mine, mine)); err == nil {
		t.Error("a creature was its own victim")
	}
}

func TestS82B12TandemTakedownPumpsBothAndTheyBothHit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B16Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B16Creature(g, me.ID, "B", "Creature — Bear", 1, 3)
	victim := effectstest.B16Creature(g, opp.ID, "Big", "Creature — Giant", 3, 5)
	effectstest.CastCatalogSpell(t, g, "Tandem Takedown", "Instant", s82b12TandemOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}, {Kind: game.TargetCard, ID: victim},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, a); p != 3 {
		t.Errorf("A has power %d, want 3", p)
	}
	if p := effectstest.EffectivePower(t, g, b); p != 2 {
		t.Errorf("B has power %d, want 2", p)
	}
	// 3 + 2 = 5 damage is lethal for the 3/5.
	if g.Battlefield.Contains(victim) {
		t.Error("the victim took 5 and lived")
	}
}

func TestS82B12TandemTakedownWithOneCreatureAndABattle(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B16Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	battle := s82b12Battle(g, opp.ID, me.ID, 5)
	effectstest.CastCatalogSpell(t, g, "Tandem Takedown", "Instant", s82b12TandemOracle, effectstest.S82b12Pair(a, battle))
	effectstest.PassPriorityAroundTable(t, g)
	if got := s82b12Defense(g, battle); got != 2 {
		t.Errorf("battle defense %d, want 2 (3 damage)", got)
	}
}

// --- Volcanic Spite ------------------------------------------------------------

func TestS82B12VolcanicSpiteDamagesThenMayRummage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B16Creature(g, opp.ID, "Bear", "Creature — Bear", 3, 3)
	effectstest.SeedLibrary(me, "Drawn")
	keep := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: keep, Name: "Spare", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	effectstest.CastCatalogSpell(t, g, "Volcanic Spite", "Instant", s82b12SpiteOracle, []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, keep)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(bear) {
		t.Fatal("the Bear took 3 and lived")
	}
	found := false
	for _, c := range me.Hand.Cards {
		if c.Name == "Drawn" {
			found = true
		}
		if c.InstanceID == keep {
			t.Error("the tucked card is still in hand")
		}
	}
	if !found {
		t.Error("no card was drawn after the tuck")
	}
	if n := len(me.Library.Cards); n == 0 || me.Library.Cards[0].InstanceID != keep {
		t.Error("the tucked card is not on the bottom of the library")
	}
}

func TestS82B12VolcanicSpiteDeclinedRummageDrawsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	walker := effectstest.P1g2Planeswalker(g, opp.ID, 5)
	effectstest.SeedLibrary(me, "Drawn")
	me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Spare", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	me.Hand.Cards = me.Hand.Cards[:0]
	me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Spare", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	hand := len(me.Hand.Cards)
	effectstest.CastCatalogSpell(t, g, "Volcanic Spite", "Instant", s82b12SpiteOracle, []game.TargetRef{{Kind: game.TargetCard, ID: walker}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	if len(me.Hand.Cards) != hand {
		t.Errorf("hand %d, want %d (the spell left, nothing drawn)", len(me.Hand.Cards), hand)
	}
}

// --- Nahiri's Warcrafting ------------------------------------------------------

func TestS82B12NahirisWarcraftingLooksAtTheExcessAndPlaysOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B16Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	me.Library.Cards = nil
	ids := effectstest.SeedLibrary(me, "One", "Two", "Three", "Four")
	effectstest.CastCatalogSpell(t, g, "Nahiri's Warcrafting", "Sorcery", s82b12WarcraftingOracle, []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	// 5 damage on a 2/2: 3 excess, so the top three are offered.
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no look prompt")
	}
	if len(pick.ChooseCards) != 3 {
		t.Errorf("%d cards offered, want 3", len(pick.ChooseCards))
	}
	effectstest.AnswerChooseCards(t, g, me.ID, ids[1])
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(bear) {
		t.Error("the Bear took 5 and lived")
	}
	exiled := false
	for _, c := range g.Exile.Cards {
		if c.InstanceID == ids[1] {
			exiled = true
		}
	}
	if !exiled {
		t.Error("the chosen card is not in exile")
	}
	if n := len(me.Library.Cards); n != 3 {
		t.Fatalf("library holds %d cards, want 3 (one exiled)", n)
	}
	if me.Library.Cards[2].InstanceID != ids[3] {
		t.Error("the unlooked fourth card is not the new top")
	}
	if a, b := me.Library.Cards[0].InstanceID, me.Library.Cards[1].InstanceID; !(a == ids[0] && b == ids[2]) && !(a == ids[2] && b == ids[0]) {
		t.Error("the two looked-at cards left over are not on the bottom")
	}
}

func TestS82B12NahirisWarcraftingExcessOnABattleCountsDefense(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	battle := s82b12Battle(g, opp.ID, me.ID, 3)
	me.Library.Cards = nil
	effectstest.SeedLibrary(me, "One", "Two", "Three", "Four")
	effectstest.CastCatalogSpell(t, g, "Nahiri's Warcrafting", "Sorcery", s82b12WarcraftingOracle, []game.TargetRef{{Kind: game.TargetCard, ID: battle}})
	effectstest.PassPriorityAroundTable(t, g)
	// 5 damage on defense 3: 2 excess, two cards looked at.
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no look prompt")
	}
	effectstest.AnswerChooseCards(t, g, me.ID)
	if n := len(me.Library.Cards); n != 4 {
		t.Errorf("library holds %d cards, want 4 (nothing exiled)", n)
	}
}

func TestS82B12NahirisWarcraftingWithNoExcessLooksAtNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ogre := effectstest.B16Creature(g, opp.ID, "Ogre", "Creature — Ogre", 5, 5)
	me.Library.Cards = nil
	effectstest.SeedLibrary(me, "One", "Two")
	effectstest.CastCatalogSpell(t, g, "Nahiri's Warcrafting", "Sorcery", s82b12WarcraftingOracle, []game.TargetRef{{Kind: game.TargetCard, ID: ogre}})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("exactly lethal damage looked at cards")
	}
}

// --- Portent Tracker -----------------------------------------------------------

func TestS82B12PortentTrackerUntapsALandAndTendsABattle(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tracker := effectstest.B12Push(g, me.ID, "Portent Tracker", "Creature — Satyr Scout", s82b12TrackerOracle, 1, 1)
	land := effectstest.AangPushLand(g, me.ID, "Forest", true)
	effectstest.B16Activate(t, g, me.ID, tracker, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: land}}})
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == land && c.Tapped {
			t.Error("the land is still tapped")
		}
	}

	// An opponent protects it: a counter comes off.
	g = effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tracker = effectstest.B12Push(g, me.ID, "Portent Tracker", "Creature — Satyr Scout", s82b12TrackerOracle, 1, 1)
	effectstest.AdvanceToMain(t, g)
	theirs := s82b12Battle(g, me.ID, opp.ID, 5)
	effectstest.B16Activate(t, g, me.ID, tracker, 1, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := s82b12Defense(g, theirs); got != 4 {
		t.Errorf("an opponent-protected battle has %d defense, want 4", got)
	}

	// Otherwise a counter goes on.
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	tracker = effectstest.B12Push(g, me.ID, "Portent Tracker", "Creature — Satyr Scout", s82b12TrackerOracle, 1, 1)
	effectstest.AdvanceToMain(t, g)
	mine := s82b12Battle(g, opp.ID, me.ID, 5)
	effectstest.B16Activate(t, g, me.ID, tracker, 1, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mine}}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := s82b12Defense(g, mine); got != 6 {
		t.Errorf("a battle you protect has %d defense, want 6", got)
	}

	// Sorcery speed only.
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	tracker = effectstest.B12Push(g, me.ID, "Portent Tracker", "Creature — Satyr Scout", s82b12TrackerOracle, 1, 1)
	battle := s82b12Battle(g, opp.ID, me.ID, 5)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.ActivateCatalogAbility(me.ID, tracker, 1, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: battle}}}); err == nil {
		t.Error("the battle ability was activated outside a main phase")
	}
}

// --- Etched Host Doombringer ---------------------------------------------------

func TestS82B12EtchedHostDoombringerModes(t *testing.T) {
	t.Run("drain", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		myLife, theirLife := me.Life, opp.Life
		effectstest.CastAndResolveCreature(t, g, "Etched Host Doombringer", "Creature — Phyrexian Demon", s82b12DoombringerOracle)
		c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
		if c == nil {
			t.Fatalf("no mode_pick prompt: %+v", g.PendingChoices)
		}
		if err := g.ResolveModePick(c.ID, me.ID, []int{0}); err != nil {
			t.Fatalf("ResolveModePick: %v", err)
		}
		effectstest.PickPlayer(t, g, me.ID, opp.ID)
		effectstest.PassPriorityAroundTable(t, g)
		if opp.Life != theirLife-2 || me.Life != myLife+2 {
			t.Errorf("life: opponent %d (was %d), you %d (was %d)", opp.Life, theirLife, me.Life, myLife)
		}
	})
	for _, tc := range []struct {
		name        string
		protectorMe bool
		want        int
	}{{"opponent protects", false, 2}, {"you protect", true, 8}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			protector := opp.ID
			if tc.protectorMe {
				protector = me.ID
			}
			battle := s82b12Battle(g, g.Seats[2].ID, protector, 5)
			effectstest.CastAndResolveCreature(t, g, "Etched Host Doombringer", "Creature — Phyrexian Demon", s82b12DoombringerOracle)
			c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
			if c == nil {
				t.Fatalf("no mode_pick prompt: %+v", g.PendingChoices)
			}
			if err := g.ResolveModePick(c.ID, me.ID, []int{1}); err != nil {
				t.Fatalf("ResolveModePick: %v", err)
			}
			effectstest.PickCard(t, g, me.ID, battle)
			effectstest.PassPriorityAroundTable(t, g)
			if got := s82b12Defense(g, battle); got != tc.want {
				t.Errorf("defense %d, want %d", got, tc.want)
			}
		})
	}
}

// --- Rampaging Raptor ----------------------------------------------------------

func TestS82B12RampagingRaptorHitsAPlaneswalkerOrBattleOfTheHitPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	raptor := effectstest.B12Push(g, me.ID, "Rampaging Raptor", "Creature — Dinosaur", s82b12RaptorOracle, 4, 4)
	effectstest.AssertKeywords(t, g, raptor, "trample", "haste")
	theirWalker := effectstest.P1g2Planeswalker(g, opp.ID, 6)
	otherWalker := effectstest.P1g2Planeswalker(g, other.ID, 6)
	theirBattle := s82b12Battle(g, other.ID, opp.ID, 6)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(raptor, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	g.RunStateChecksForTest()
	// The target is chosen from the hit player's walkers and battles only.
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the Raptor's trigger")
	}
	for _, bad := range []uuid.UUID{otherWalker} {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
			t.Error("another player's planeswalker was a legal target")
			return
		}
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: theirBattle}); err != nil {
		t.Fatalf("the battle the player protects was refused: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := s82b12Defense(g, theirBattle); got != 2 {
		t.Errorf("battle defense %d, want 2 (4 damage)", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == theirWalker && c.Counters[game.CounterLoyalty] != 6 {
			t.Errorf("the untargeted walker lost loyalty: %d", c.Counters[game.CounterLoyalty])
		}
	}
}

func TestS82B12RampagingRaptorPumpsForTwoRed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	raptor := effectstest.B12Push(g, me.ID, "Rampaging Raptor", "Creature — Dinosaur", s82b12RaptorOracle, 4, 4)
	effectstest.B16Activate(t, g, me.ID, raptor, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, raptor); p != 6 {
		t.Errorf("power %d, want 6", p)
	}
}

// --- Halo-Charged Skaab --------------------------------------------------------

func TestS82B12HaloChargedSkaabMillsEveryoneAndTucksASpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for _, p := range g.Seats {
		p.Library.Cards = nil
		effectstest.SeedLibrary(p, "Filler A", "Filler B", "Filler C")
	}
	// The top two of my library are an instant and a creature: both are
	// milled, and the instant is a legal pick.
	me.Library.Cards = nil
	ids := effectstest.SeedLibrary(me, "Milled Bolt", "Milled Bear", "Rest")
	for i := range me.Library.Cards {
		switch me.Library.Cards[i].InstanceID {
		case ids[0]:
			me.Library.Cards[i].TypeLine = "Instant"
		case ids[1]:
			me.Library.Cards[i].TypeLine = "Creature — Bear"
		}
	}
	effectstest.CastAndResolveCreature(t, g, "Halo-Charged Skaab", "Creature — Zombie", s82b12SkaabOracle)
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats[1:] {
		if n := len(p.Graveyard.Cards); n != 2 {
			t.Errorf("%s milled %d cards, want 2", p.Name, n)
		}
	}
	effectstest.AnswerChooseCards(t, g, me.ID, ids[0])
	if len(me.Library.Cards) == 0 || me.Library.Cards[len(me.Library.Cards)-1].InstanceID != ids[0] {
		t.Error("the instant is not on top of the library")
	}
	for _, c := range me.Graveyard.Cards {
		if c.InstanceID == ids[0] {
			t.Error("the instant is still in the graveyard")
		}
	}
}

// --- Wildwood Escort -----------------------------------------------------------

func TestS82B12WildwoodEscortReturnsACreatureOrBattleCardAndExilesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.SeedGraveyardCard(t, g, "Dead Bear", "Creature — Bear", "")
	siege := effectstest.SeedGraveyardCard(t, g, "Dead Siege", "Battle — Siege", "")
	land := effectstest.SeedGraveyardCard(t, g, "Dead Land", "Land", "")
	escort := effectstest.CastAndResolveCreature(t, g, "Wildwood Escort", "Creature — Elf Warrior", s82b12EscortOracle)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: land}); err == nil {
		t.Error("a land card was a legal target")
	}
	effectstest.PickCard(t, g, me.ID, siege)
	effectstest.PassPriorityAroundTable(t, g)
	inHand := false
	for _, c := range me.Hand.Cards {
		if c.InstanceID == siege {
			inHand = true
		}
	}
	if !inHand {
		t.Error("the battle card did not return to hand")
	}
	_ = bear
	// If it would die, it is exiled instead.
	effectstest.B27Kill(g, escort)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(escort) {
		t.Fatal("the Escort is still on the battlefield")
	}
	for _, c := range me.Graveyard.Cards {
		if c.InstanceID == escort {
			t.Error("the Escort died into the graveyard")
		}
	}
	exiled := false
	for _, c := range g.Exile.Cards {
		if c.InstanceID == escort {
			exiled = true
		}
	}
	if !exiled {
		t.Error("the Escort is not in exile")
	}
}

// --- Flitting Guerrilla --------------------------------------------------------

func TestS82B12FlittingGuerrillaMillsThenExilesItselfToTuckACreatureOrBattle(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for _, p := range g.Seats {
		p.Library.Cards = nil
		effectstest.SeedLibrary(p, "Filler A", "Filler B", "Filler C")
	}
	guerrilla := effectstest.B12Push(g, me.ID, "Flitting Guerrilla", "Creature — Faerie Rogue", s82b12GuerrillaOracle, 2, 2)
	effectstest.AssertKeywords(t, g, guerrilla, "flying")
	buried := effectstest.SeedGraveyardCard(t, g, "Buried Bear", "Creature — Bear", "")
	effectstest.B27Kill(g, guerrilla)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerConfirmFor(t, g, me, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickCard(t, g, me.ID, buried)
	effectstest.PassPriorityAroundTable(t, g)
	exiled := false
	for _, c := range g.Exile.Cards {
		if c.InstanceID == guerrilla {
			exiled = true
		}
	}
	if !exiled {
		t.Error("the Guerrilla was not exiled")
	}
	if me.Library.Cards[len(me.Library.Cards)-1].InstanceID != buried {
		t.Error("the Bear is not on top of the library")
	}
}

func TestS82B12FlittingGuerrillaDeclinedStaysInTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for _, p := range g.Seats {
		p.Library.Cards = nil
		effectstest.SeedLibrary(p, "Filler A", "Filler B", "Filler C")
	}
	guerrilla := effectstest.B12Push(g, me.ID, "Flitting Guerrilla", "Creature — Faerie Rogue", s82b12GuerrillaOracle, 2, 2)
	effectstest.B27Kill(g, guerrilla)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerConfirmFor(t, g, me, false)
	effectstest.PassPriorityAroundTable(t, g)
	inYard := false
	for _, c := range me.Graveyard.Cards {
		if c.InstanceID == guerrilla {
			inYard = true
		}
	}
	if !inYard {
		t.Error("a declined Guerrilla left the graveyard")
	}
	for _, p := range g.Seats {
		if p == me {
			continue
		}
		if n := len(p.Graveyard.Cards); n != 2 {
			t.Errorf("%s milled %d, want 2", p.Name, n)
		}
	}
}
