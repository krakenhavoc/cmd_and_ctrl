package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b14_cards_test.go — Secrets of Strixhaven Commander, batch 14
// (#4834). Each test plays the real catalog entry and asserts what its
// text says.

const (
	socB14SurgeOracle     = "8ea414c1-88cf-46d4-9a54-6dee9847c537"
	socB14TanazirOracle   = "73c44cf5-a601-4625-9550-e722fba2ec0b"
	socB14DrakeOracle     = "88d0a394-79e2-42ac-9a37-1a7b78231941"
	socB14TomikOracle     = "2ef0961a-0543-4371-8c2a-333881b735e5"
	socB14EnvoyOracle     = "f08f7814-4585-41dd-b82e-ed2ef5e24242"
	socB14TroyanOracle    = "1135ebc3-9537-4341-956b-6e22d8be5f20"
	socB14TrudgeOracle    = "be175e92-567a-42e3-ad54-136c0631fe0e"
	socB14TwinflameOracle = "83cf1169-5853-4332-b897-7b17d72d76ab"
	socB14UmbralOracle    = "5a1dfc60-645d-4bc0-8883-e06ed9c80706"
	socB14FlourishOracle  = "4db030f5-d7c4-4aa0-ae2e-12943e2683d7"
)

func TestSocB14EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		socB14SurgeOracle:     "Surge to Victory",
		socB14TanazirOracle:   "Tanazir Quandrix",
		socB14DrakeOracle:     "Thunderclap Drake",
		socB14TomikOracle:     "Tomik, Wielder of Law",
		socB14EnvoyOracle:     "Transcendent Envoy",
		socB14TroyanOracle:    "Troyan, Gutsy Explorer",
		socB14TrudgeOracle:    "Trudge Garden",
		socB14TwinflameOracle: "Twinflame",
		socB14UmbralOracle:    "Umbral Expanse",
		socB14FlourishOracle:  "Unbound Flourishing",
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

// socB14SettleKeepingTargets passes priority around the table until the stack
// is empty, answering every CR 603.3b order prompt as offered and every CR
// 707.10c new-target prompt by keeping `keep`. It returns how many target
// prompts it answered.
func socB14SettleKeepingTargets(t *testing.T, g *game.Game, caster uuid.UUID, keep game.TargetRef) int {
	t.Helper()
	answered := 0
	for i := 0; i < 64; i++ {
		if p := triggerOrderPromptFor(g, caster); p != nil {
			order := append([]uuid.UUID(nil), p.TriggerOrderIDs...)
			if err := g.ResolveTriggerOrder(p.ID, caster, order); err != nil {
				t.Fatalf("ResolveTriggerOrder: %v", err)
			}
			continue
		}
		if p := latestPickTarget(g, caster); p != nil {
			if err := g.ResolvePickTarget(p.ID, caster, keep); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			answered++
			continue
		}
		if stackFullyEmpty(g) {
			return answered
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the stack never emptied")
	return answered
}

// --- Transcendent Envoy ---

func TestSocB14TranscendentEnvoyDiscountsYourAuraSpellsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	envoy := b12Push(g, me.ID, "Transcendent Envoy", "Enchantment Creature — Griffin", socB14EnvoyOracle, 1, 2)
	if !hasEffectiveKeyword(t, g, envoy, "flying") {
		t.Error("Transcendent Envoy lacks flying")
	}
	bear := auraBear(g, me.ID)
	at := []game.TargetRef{{Kind: game.TargetCard, ID: bear}}

	aura := handCardFull(me, "Test Aura", auraTypeLine, "{2}{W}", "", []string{"W"})
	b06AddMana(me, "C", "W")
	if err := g.CastSpell(me.ID, aura, game.CastSpellParams{Strict: true, Targets: at}); err != nil {
		t.Fatalf("an Aura costing {2}{W} should cost {1}{W} with the Envoy out: %v", err)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("%d mana left, want 0", me.ManaPool.Total())
	}
	passPriorityAroundTable(t, g)

	// A plain enchantment spell is not an Aura spell.
	plain := handCardFull(me, "Test Enchantment", "Enchantment", "{2}{W}", "", []string{"W"})
	b06AddMana(me, "C", "W")
	if err := g.CastSpell(me.ID, plain, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("a non-Aura enchantment spell was discounted")
	}
}

func TestSocB14TranscendentEnvoyDoesNotDiscountAnOpponentsAuras(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	b12Push(g, opp.ID, "Transcendent Envoy", "Enchantment Creature — Griffin", socB14EnvoyOracle, 1, 2)
	bear := auraBear(g, me.ID)
	aura := handCardFull(me, "Test Aura", auraTypeLine, "{2}{W}", "", []string{"W"})
	b06AddMana(me, "C", "W")
	if err := g.CastSpell(me.ID, aura, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err == nil {
		t.Error("an opponent's Envoy discounted my Aura")
	}
}

// --- Umbral Expanse ---

func TestSocB14UmbralExpanseEntersTappedAndTapsForWhiteOrBlack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := playLandFromHand(t, g, "Umbral Expanse", socB14UmbralOracle)
	card, ok := battlefieldCard(g, id)
	if !ok || !card.Tapped {
		t.Fatal("Umbral Expanse must enter tapped")
	}
	if n := tapEventsFor(g, id); n != 0 {
		t.Errorf("%d tap events; the land should have ENTERED tapped", n)
	}
	for _, color := range []string{"W", "B"} {
		g.WithWriteLock(func() { _ = g.UntapTargetForEffect(id) })
		me.ManaPool = nil
		b28TapForMana(t, g, me.ID, id, color)
		if got := poolColors(me); len(got) != 1 || got[0] != color {
			t.Errorf("pool = %v, want {%s}", got, color)
		}
	}
}

func TestSocB14UmbralExpanseCycles(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hand := me.Hand.Size()
	id, _ := cycleFromHand(t, g, "Umbral Expanse", "Land — Plains Swamp", socB14UmbralOracle, "{C}{C}")
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand size = %d, want %d after cycling (one out, one drawn)", got, hand)
	}
	if !me.Graveyard.Contains(id) {
		t.Error("the cycled land is not in the graveyard")
	}
}

// --- Tomik, Wielder of Law ---

func socB14Tomik(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me, attacker := g.Seats[1], g.Seats[0]
	tomik := b12Push(g, me.ID, "Tomik, Wielder of Law", "Legendary Creature — Human Advisor", socB14TomikOracle, 2, 4)
	return g, me, attacker, tomik
}

func TestSocB14TomikHasFlyingAndVigilance(t *testing.T) {
	g, _, _, tomik := socB14Tomik(t)
	for _, kw := range []string{"flying", "vigilance"} {
		if !hasEffectiveKeyword(t, g, tomik, kw) {
			t.Errorf("Tomik lacks %s", kw)
		}
	}
}

func TestSocB14TomikAffinityForPlaneswalkers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	for i := 0; i < 2; i++ {
		pushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Test Walker", TypeLine: "Legendary Planeswalker — Test",
			Counters: map[string]int{game.CounterLoyalty: 5}, Owner: me.ID, Controller: me.ID,
		})
	}
	tomik := handCardFull(me, "Tomik, Wielder of Law", "Legendary Creature — Human Advisor", "{1}{W}{B}", socB14TomikOracle, []string{"W", "B"})
	b06AddMana(me, "W", "B")
	if err := g.CastSpell(me.ID, tomik, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("with two planeswalkers {1}{W}{B} should cost {W}{B}: %v", err)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("%d mana left, want 0", me.ManaPool.Total())
	}
}

func TestSocB14TomikCostsFullPriceWithoutPlaneswalkers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	tomik := handCardFull(me, "Tomik, Wielder of Law", "Legendary Creature — Human Advisor", "{1}{W}{B}", socB14TomikOracle, []string{"W", "B"})
	b06AddMana(me, "W", "B")
	if err := g.CastSpell(me.ID, tomik, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("Tomik cost {W}{B} with no planeswalker out")
	}
}

func TestSocB14TomikDrainsAndDrawsOnceForATwoCreatureAttackAtYou(t *testing.T) {
	g, me, attacker, _ := socB14Tomik(t)
	a := pushVanillaCreature(g, attacker.ID, "Bear A", 2, 2)
	b := pushVanillaCreature(g, attacker.ID, "Bear B", 2, 2)
	c := pushVanillaCreature(g, attacker.ID, "Bear C", 2, 2)
	hand, life := me.Hand.Size(), attacker.Life
	declareAttack(t, g, me.ID, a, b, c)
	passPriorityAroundTable(t, g)
	if got := attacker.Life; got != life-3 {
		t.Errorf("the attacker is at %d, want %d (3 lost once)", got, life-3)
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards, want exactly 1", got)
	}
}

func TestSocB14TomikNeedsTwoAttackersAtYou(t *testing.T) {
	t.Run("one attacker", func(t *testing.T) {
		g, me, attacker, _ := socB14Tomik(t)
		a := pushVanillaCreature(g, attacker.ID, "Bear A", 2, 2)
		hand, life := me.Hand.Size(), attacker.Life
		declareAttack(t, g, me.ID, a)
		passPriorityAroundTable(t, g)
		if attacker.Life != life || me.Hand.Size() != hand {
			t.Errorf("one attacker drained %d and drew %d", life-attacker.Life, me.Hand.Size()-hand)
		}
	})
	t.Run("two attackers at somebody else", func(t *testing.T) {
		g, me, attacker, _ := socB14Tomik(t)
		other := g.Seats[2]
		a := pushVanillaCreature(g, attacker.ID, "Bear A", 2, 2)
		b := pushVanillaCreature(g, attacker.ID, "Bear B", 2, 2)
		hand, life := me.Hand.Size(), attacker.Life
		declareAttack(t, g, other.ID, a, b)
		passPriorityAroundTable(t, g)
		if attacker.Life != life || me.Hand.Size() != hand {
			t.Errorf("an attack at someone else drained %d and drew %d", life-attacker.Life, me.Hand.Size()-hand)
		}
	})
	t.Run("one at you and one at somebody else", func(t *testing.T) {
		g, me, attacker, _ := socB14Tomik(t)
		other := g.Seats[2]
		a := pushVanillaCreature(g, attacker.ID, "Bear A", 2, 2)
		b := pushVanillaCreature(g, attacker.ID, "Bear B", 2, 2)
		hand, life := me.Hand.Size(), attacker.Life
		advanceTo(t, g, game.StepDeclareAttackers)
		if err := g.DeclareAttacker(a, me.ID); err != nil {
			t.Fatal(err)
		}
		if err := g.DeclareAttacker(b, other.ID); err != nil {
			t.Fatal(err)
		}
		lockInAttacks(t, g)
		passPriorityAroundTable(t, g)
		if attacker.Life != life || me.Hand.Size() != hand {
			t.Errorf("a split attack drained %d and drew %d", life-attacker.Life, me.Hand.Size()-hand)
		}
	})
}

func TestSocB14TomikCountsAttackersAtYourPlaneswalkers(t *testing.T) {
	g, me, attacker, _ := socB14Tomik(t)
	walker := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test Walker", TypeLine: "Legendary Planeswalker — Test",
		Counters: map[string]int{game.CounterLoyalty: 9}, Owner: me.ID, Controller: me.ID,
	})
	a := pushVanillaCreature(g, attacker.ID, "Bear A", 2, 2)
	b := pushVanillaCreature(g, attacker.ID, "Bear B", 2, 2)
	hand, life := me.Hand.Size(), attacker.Life
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(a, me.ID); err != nil {
		t.Fatal(err)
	}
	if err := g.DeclareAttacker(b, walker); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if attacker.Life != life-3 || me.Hand.Size() != hand+1 {
		t.Errorf("one at you and one at your planeswalker: lost %d and drew %d, want 3 and 1", life-attacker.Life, me.Hand.Size()-hand)
	}
}

func TestSocB14TomikChecksTheAttackersAgainOnResolution(t *testing.T) {
	g, me, attacker, _ := socB14Tomik(t)
	a := pushVanillaCreature(g, attacker.ID, "Bear A", 2, 2)
	b := pushVanillaCreature(g, attacker.ID, "Bear B", 2, 2)
	hand, life := me.Hand.Size(), attacker.Life
	declareAttack(t, g, me.ID, a, b)
	// One attacker dies with the trigger on the stack: only one is left
	// attacking, so the intervening "if" (CR 603.4) fails as it resolves.
	killCreature(t, g, attacker.ID, a)
	passPriorityAroundTable(t, g)
	if attacker.Life != life || me.Hand.Size() != hand {
		t.Errorf("lost %d and drew %d though only one attacker was left", life-attacker.Life, me.Hand.Size()-hand)
	}
}

func TestSocB14TomikIgnoresAnAttackFromYourOwnSide(t *testing.T) {
	g, me, _, _ := socB14Tomik(t)
	other := g.Seats[2]
	// Tomik's controller attacking another player is not "an opponent".
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "Bear B", 2, 2)
	for g.Turn.ActiveSeat != 1 {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	advanceTo(t, g, game.StepPrecombatMain)
	hand := me.Hand.Size()
	declareAttack(t, g, other.ID, a, b)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("drew %d on my own attack", me.Hand.Size()-hand)
	}
}

// --- Troyan, Gutsy Explorer ---

func socB14TroyanMana(t *testing.T) (*game.Game, *game.Player) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	troyan := pushCatalogPermanent(g, me.ID, "Troyan, Gutsy Explorer", "Legendary Creature — Vedalken Scout", socB14TroyanOracle, false)
	b28TapForMana(t, g, me.ID, troyan, "")
	if got := me.ManaPool.Total(); got != 2 {
		t.Fatalf("Troyan made %d mana, want {G}{U}", got)
	}
	want := game.ManaRestrictAnyOf(
		[]string{game.ManaRestrictCast, game.ManaRestrictManaValueAtLeast(5)},
		[]string{game.ManaRestrictSpellXCost},
	)
	colors := map[string]bool{}
	for _, tok := range me.ManaPool {
		colors[tok.Color] = true
		if len(tok.Restrictions) != 1 || tok.Restrictions[0] != want {
			t.Fatalf("pool token %+v is not restricted to %q", tok, want)
		}
	}
	if !colors["G"] || !colors["U"] {
		t.Fatalf("pool colours %v, want G and U", colors)
	}
	return g, me
}

func TestSocB14TroyanManaPaysForAManaValueFiveSpell(t *testing.T) {
	g, me := socB14TroyanMana(t)
	b06AddMana(me, "C", "C", "C")
	five := handCardFull(me, "Five Drop", "Sorcery", "{3}{G}{U}", "", []string{"G", "U"})
	if err := g.CastSpell(me.ID, five, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("a mana value 5 spell paid with Troyan's mana: %v", err)
	}
	if got := me.ManaPool.Total(); got != 0 {
		t.Errorf("%d mana left, want 0", got)
	}
}

func TestSocB14TroyanManaCannotPayForAManaValueFourSpell(t *testing.T) {
	g, me := socB14TroyanMana(t)
	b06AddMana(me, "C", "C")
	four := handCardFull(me, "Four Drop", "Sorcery", "{2}{G}{U}", "", []string{"G", "U"})
	if err := g.CastSpell(me.ID, four, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("a mana value 4 spell was paid with Troyan's mana")
	}
}

func TestSocB14TroyanManaPaysForASpellWithXInItsManaCost(t *testing.T) {
	g, me := socB14TroyanMana(t)
	xSpell := handCardFull(me, "X Spell", "Sorcery", "{X}{G}{U}", "", []string{"G", "U"})
	if err := g.CastSpell(me.ID, xSpell, game.CastSpellParams{Strict: true, XValue: 0}); err != nil {
		t.Fatalf("a spell with {X} in its mana cost paid with Troyan's mana: %v", err)
	}
	if got := me.ManaPool.Total(); got != 0 {
		t.Errorf("%d mana left, want 0", got)
	}
}

func TestSocB14TroyanManaCannotPayForItsOwnLoot(t *testing.T) {
	g, me := socB14TroyanMana(t)
	var troyan uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.OracleID == socB14TroyanOracle {
			troyan = c.InstanceID
		}
	}
	// Troyan is tapped for the mana; untap it. {U} of the restricted
	// kind may not pay for an activated ability.
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(troyan) })
	if err := g.ActivateCatalogAbility(me.ID, troyan, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("Troyan's restricted mana paid for an activated ability")
	}
}

func TestSocB14TroyanLoots(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	troyan := pushCatalogPermanent(g, me.ID, "Troyan, Gutsy Explorer", "Legendary Creature — Vedalken Scout", socB14TroyanOracle, false)
	advanceToMain(t, g)
	b06AddMana(me, "U")
	hand := me.Hand.Size()
	b16Activate(t, g, me.ID, troyan, 0, game.ActivateAbilityParams{Strict: true})
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("drew %d, want 1 before the discard", got-hand)
	}
	if discardOwed(g, me.ID) != 1 {
		t.Errorf("discard owed = %d, want 1", discardOwed(g, me.ID))
	}
	if !b20Tapped(t, g, troyan) {
		t.Error("the loot has a tap cost")
	}
}

// --- Trudge Garden ---

func TestSocB14TrudgeGardenMakesAFungusBeastWhenYouPayOnLifeGain(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Trudge Garden", "Enchantment", socB14TrudgeOracle, 0, 0)
	b14Gain(t, g, me.ID, 3)
	passPriorityAroundTable(t, g)
	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("gaining life asks the controller to pay {2}")
	}
	b06AddMana(me, "C", "C")
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	beasts := socB10Named(g, "Fungus Beast")
	if len(beasts) != 1 {
		t.Fatalf("%d Fungus Beasts, want 1", len(beasts))
	}
	c, _ := g.LookupCardForEffect(beasts[0])
	if !IsToken(c) || c.Controller != me.ID || c.Power != 4 || c.Toughness != 4 || !c.HasColor("G") || !c.HasSubtype("Fungus") {
		t.Errorf("the token is %+v, want a 4/4 green Fungus Beast token under my control", c)
	}
	if !hasEffectiveKeyword(t, g, beasts[0], "trample") {
		t.Error("the Fungus Beast lacks trample")
	}
	if me.ManaPool.Total() != 0 {
		t.Error("the {2} was not spent")
	}

	// Declining makes nothing.
	b14Gain(t, g, me.ID, 1)
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if n := len(socB10Named(g, "Fungus Beast")); n != 1 {
		t.Errorf("%d Fungus Beasts after declining, want still 1", n)
	}

	// An opponent's life gain is not yours.
	b14Gain(t, g, opp.ID, 3)
	passPriorityAroundTable(t, g)
	if hasPayUnlessFor(g, me.ID) {
		t.Error("an opponent's life gain asked me to pay")
	}
}

func TestSocB14TrudgeGardenAsksOncePerLifeGainEvent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Trudge Garden", "Enchantment", socB14TrudgeOracle, 0, 0)
	b14Gain(t, g, me.ID, 1)
	b14Gain(t, g, me.ID, 1)
	passPriorityAroundTable(t, g)
	b06AddMana(me, "C", "C", "C", "C")
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if n := len(socB10Named(g, "Fungus Beast")); n != 2 {
		t.Errorf("%d Fungus Beasts from two life gains, want 2", n)
	}
}
