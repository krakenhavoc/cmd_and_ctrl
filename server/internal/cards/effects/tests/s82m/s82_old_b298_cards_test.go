package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 OLD batch 298 (Oath of the Gatewatch: card draw and library, #5545).
// Every test plays the real catalog entry through the engine.

const (
	oldB298Jori      = "4cd9b48e-5774-4a1c-857f-9f4a92db9b41"
	oldB298Call      = "652bfa31-b2eb-4cba-850f-1db02a83bb35"
	oldB298Waters    = "77fef82d-e8d0-444b-8d76-68e9b533f38b"
	oldB298Tazri     = "b0f19cba-1339-4518-8320-d7b1dcaf2eb0"
	oldB298Bonds     = "d93ba2f2-bf9b-444b-adbb-3835662734e7"
	oldB298Flameseek = "efae637f-3232-46f2-9839-f3386e2f447d"
)

func TestOldB298EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB298Jori: "Jori En, Ruin Diver", oldB298Call: "Call the Gatewatch",
		oldB298Waters: "Roiling Waters", oldB298Tazri: "General Tazri",
		oldB298Bonds: "Bonds of Mortality", oldB298Flameseek: "Akoum Flameseeker",
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

func TestOldB298JoriEnDrawsOnTheSecondSpellOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	effectstest.B12Push(g, me.ID, "Jori En, Ruin Diver", "Legendary Creature — Merfolk Wizard", oldB298Jori, 2, 3)
	hand := me.Hand.Size()
	for i, want := range []int{hand, hand + 1, hand + 1} { // first, second, third spell
		effectstest.CastCatalogSpell(t, g, "Test Instant", "Instant", "", nil)
		effectstest.PassPriorityAroundTable(t, g)
		if got := me.Hand.Size(); got != want {
			t.Errorf("after spell %d: hand %d, want %d", i+1, got, want)
		}
	}
}

func TestOldB298CallTheGatewatchFindsOnlyAPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"},
		game.Card{Name: "Walker", TypeLine: "Legendary Planeswalker — Gideon", ManaCost: "{2}{W}{W}"},
	)
	effectstest.CastCatalogSpell(t, g, "Call the Gatewatch", "Sorcery", oldB298Call, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if effectstest.SearchOptionNamed(g, c, "Bear") != uuid.Nil {
		t.Error("a creature card is offered")
	}
	pick := effectstest.SearchOptionNamed(g, c, "Walker")
	if pick == uuid.Nil {
		t.Fatal("the planeswalker is not offered")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, pick)
	if !me.Hand.Contains(pick) {
		t.Error("the planeswalker did not reach the hand")
	}
}

func TestOldB298RoilingWatersBouncesTwoOpposingCreaturesAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, foe := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, foe.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, foe.ID, "B", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Roiling Waters", "Sorcery", oldB298Waters,
		append(effectstest.CardRefs(mine), effectstest.PlayerRefs(me.ID)...)); err == nil {
		t.Error("a creature I control was accepted as a target")
	}
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Roiling Waters", "Sorcery", oldB298Waters,
		append(effectstest.CardRefs(a, b), game.TargetRef{Kind: game.TargetPlayer, ID: me.ID, Slot: 1}))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OnBattlefieldByID(g, a) || effectstest.OnBattlefieldByID(g, b) {
		t.Error("an opposing creature stayed on the battlefield")
	}
	if !foe.Hand.Contains(a) || !foe.Hand.Contains(b) {
		t.Error("the creatures did not return to their owner's hand")
	}
	if !effectstest.OnBattlefieldByID(g, mine) {
		t.Error("my creature was bounced")
	}
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d, want %d (two drawn, the spell cast)", got, hand+2)
	}
}

func TestOldB298RoilingWatersStillDrawsWithNoCreatureChosen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Roiling Waters", "Sorcery", oldB298Waters,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID, Slot: 1}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d, want %d", got, hand+2)
	}
}

func TestOldB298GeneralTazriSearchesForAnAllyOnEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"},
		game.Card{Name: "Kor", TypeLine: "Creature — Kor Ally", ManaCost: "{1}{W}"},
		game.Card{Name: "Ally Spell", TypeLine: "Sorcery", ManaCost: "{W}"},
	)
	effectstest.CastAndResolveCreature(t, g, "General Tazri", "Legendary Creature — Human Ally", oldB298Tazri)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if effectstest.SearchOptionNamed(g, c, "Bear") != uuid.Nil || effectstest.SearchOptionNamed(g, c, "Ally Spell") != uuid.Nil {
		t.Error("a non-Ally creature card is offered")
	}
	pick := effectstest.SearchOptionNamed(g, c, "Kor")
	if pick == uuid.Nil {
		t.Fatal("the Ally is not offered")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, pick)
	if !me.Hand.Contains(pick) {
		t.Error("the Ally did not reach the hand")
	}
}

func TestOldB298GeneralTazriPumpsAlliesByTheColorsAmongThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[0]
	tazri := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "General Tazri",
		TypeLine: "Legendary Creature — Human Ally", OracleID: oldB298Tazri, Colors: []string{"W"},
		Power: 3, Toughness: 4, Owner: me.ID, Controller: me.ID})
	red := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Red Ally",
		TypeLine: "Creature — Goblin Ally", Colors: []string{"R"}, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	both := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Gold Ally",
		TypeLine: "Creature — Elf Ally", Colors: []string{"R", "G"}, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Blue Bear",
		TypeLine: "Creature — Bear", Colors: []string{"U"}, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	effectstest.FloatForTest(g, me, "WUBRG")
	effectstest.ApaActivate(t, g, me, tazri, 0, game.ActivateAbilityParams{})
	// W, R and G among the Allies: X = 3. The blue Bear is no Ally and adds no color.
	for id, want := range map[uuid.UUID]int{tazri: 6, red: 4, both: 4, bear: 2} {
		if p, _ := effectstest.EffectivePTOf(t, g, id); p != want {
			t.Errorf("power %d, want %d", p, want)
		}
	}
}

func TestOldB298BondsOfMortalityDrawsAndStripsOpposingCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, foe := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Shielded", TypeLine: "Creature — Bear",
		Keywords: []string{"hexproof", "indestructible"}, Power: 2, Toughness: 2, Owner: foe.ID, Controller: foe.ID})
	mine := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Mine", TypeLine: "Creature — Bear",
		Keywords: []string{"hexproof"}, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	hand := me.Hand.Size()
	bonds := effectstest.CastCatalogSpell(t, g, "Bonds of Mortality", "Enchantment", oldB298Bonds, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.OnBattlefieldByID(g, bonds) {
		t.Fatal("Bonds of Mortality did not resolve")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d (one drawn)", got, hand+1)
	}
	effectstest.FloatForTest(g, me, "G")
	effectstest.ApaActivate(t, g, me, bonds, 0, game.ActivateAbilityParams{})
	for _, kw := range []string{"hexproof", "indestructible"} {
		if effectstest.HasKeywordOnBattlefield(t, g, theirs, kw) {
			t.Errorf("the opposing creature kept %s", kw)
		}
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, mine, "hexproof") {
		t.Error("my own creature lost hexproof")
	}
}

func TestOldB298AkoumFlameseekerLoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[0]
	seeker := effectstest.B12Push(g, me.ID, "Akoum Flameseeker", "Creature — Human Shaman Ally", oldB298Flameseek, 3, 2)
	ally := effectstest.B12Creature(g, me.ID, "Ally", "Creature — Kor Ally", 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, seeker, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("cohort was paid without tapping an Ally")
	}
	hand := me.Hand.Size()
	effectstest.ApaActivate(t, g, me, seeker, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{ally}})
	if effectstest.DiscardChoiceFor(g, me.ID) == nil {
		t.Fatal("no discard prompt")
	}
	effectstest.DiscardFromHand(t, g, me.ID)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d, want %d (discard one, draw one)", got, hand)
	}
}
