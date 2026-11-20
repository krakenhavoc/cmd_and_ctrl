package testsu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// woe_std51_cards_test.go — Standard-set leftovers, batch 51 (#4452): five
// Lost Caverns of Ixalan cards (#4470) and five Wilds of Eldraine cards
// (#4471). Each test plays the real catalog entry and asserts what its text
// says.

const (
	std51PaleontologistOracle = "d6026309-aae9-43e7-938b-69db7af7139c"
	std51SpelunkerOracle      = "0a72d6cd-b4a7-4c42-90ba-a14f1733d00e"
	std51EchoOracle           = "0c0d6c7e-a7cc-49e0-bf84-3fc33dd10b90"
	std51MatzalantliOracle    = "16182e01-22ff-4786-985d-919b47c4aa4d"
	std51BatOracle            = "e46f6644-6bf1-47b1-817c-906bd3eca99f"
	std51ArchonOracle         = "180efbd4-0f77-41cc-8ae2-2e6ccf853d10"
	std51GloryOracle          = "bc5631ae-494e-4987-8d17-57674bfbdb9f"
	std51BreakOracle          = "b9d51e07-851f-4b7b-bffe-7c1efff0714c"
	std51MouseOracle          = "db70ddae-0ac1-47f2-9158-8448e337a08f"
	std51CoopedUpOracle       = "3e8bcb44-b875-4fe4-b96f-97a20cbfec93"
)

func TestStd51EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		std51PaleontologistOracle:     "Intrepid Paleontologist",
		std51SpelunkerOracle:          "Jadelight Spelunker",
		std51EchoOracle:               "Deepfathom Echo",
		std51MatzalantliOracle:        "Matzalantli, the Great Door",
		std51MatzalantliOracle + "#1": "The Core",
		std51BatOracle:                "Deep-Cavern Bat",
		std51ArchonOracle:             "Archon of the Wild Rose",
		std51GloryOracle:              "Archon's Glory",
		std51BreakOracle:              "Break the Spell",
		std51MouseOracle:              "Cheeky House-Mouse",
		std51MouseOracle + "#1":       "Squeak By",
		std51CoopedUpOracle:           "Cooped Up",
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

// std51Me is the active seat; std51Opp the next.
func std51Seats(g *game.Game) (*game.Player, *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

// std51CastCreatureX casts a creature card from the active seat's hand for X.
func std51CastCreatureX(t *testing.T, g *game.Game, name, typeLine, oracle, cost string, power, toughness, x int) uuid.UUID {
	t.Helper()
	me, _ := std51Seats(g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: cost,
		Power: power, Toughness: toughness, Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: x}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// --- Jadelight Spelunker ------------------------------------------------------

func TestStd51JadelightSpelunkerExploresXTimesOneAfterAnother(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := std51Seats(g)
	top := effectstest.PutOnTopForTest(me, "Some Sorcery", "Sorcery")
	spelunker := std51CastCreatureX(t, g, "Jadelight Spelunker", "Creature — Merfolk Scout", std51SpelunkerOracle, "{X}{G}", 1, 1, 2)
	effectstest.PassPriorityAroundTable(t, g)
	// The first explore: a nonland, a counter, and the card is left on top.
	if got := effectstest.PlusOneCounters(g, spelunker); got != 1 {
		t.Fatalf("after the first explore: %d counters, want 1", got)
	}
	effectstest.AnswerExploreQuestion(t, g, me.ID, false)
	// The second explore reveals the same card again (the ruling).
	if got := effectstest.PlusOneCounters(g, spelunker); got != 2 {
		t.Fatalf("after the second explore: %d counters, want 2", got)
	}
	effectstest.AnswerExploreQuestion(t, g, me.ID, true)
	if !me.Graveyard.Contains(top) {
		t.Error("the card revealed twice was not put into the graveyard")
	}
	if got := len(effectstest.EventsOfKind(g, game.EventExplored)); got != 2 {
		t.Errorf("explored %d times, want 2", got)
	}
}

func TestStd51JadelightSpelunkerXZeroDoesNotExplore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := std51Seats(g)
	effectstest.PutOnTopForTest(me, "Some Sorcery", "Sorcery")
	spelunker := std51CastCreatureX(t, g, "Jadelight Spelunker", "Creature — Merfolk Scout", std51SpelunkerOracle, "{X}{G}", 1, 1, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(effectstest.EventsOfKind(g, game.EventExplored)); got != 0 {
		t.Errorf("explored %d times with X=0, want none", got)
	}
	if got := effectstest.PlusOneCounters(g, spelunker); got != 0 {
		t.Errorf("%d counters with X=0, want none", got)
	}
}

func TestStd51JadelightSpelunkerLandsGoToHandWithoutCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := std51Seats(g)
	l1 := effectstest.PutOnTopForTest(me, "Forest", "Basic Land — Forest")
	spelunker := std51CastCreatureX(t, g, "Jadelight Spelunker", "Creature — Merfolk Scout", std51SpelunkerOracle, "{X}{G}", 1, 1, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(l1) || effectstest.PlusOneCounters(g, spelunker) != 0 {
		t.Errorf("a land explore: in hand %v, counters %d", me.Hand.Contains(l1), effectstest.PlusOneCounters(g, spelunker))
	}
}

// --- Deepfathom Echo ----------------------------------------------------------

// std51ChooseCards answers the open choose-cards prompt with these picks.
func std51ChooseCards(t *testing.T, g *game.Game, chooser uuid.UUID, picks ...uuid.UUID) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceChooseCards && c.Chooser == chooser {
			if err := g.ResolveChooseCards(c.ID, chooser, picks); err != nil {
				t.Fatalf("ResolveChooseCards: %v", err)
			}
			return
		}
	}
	t.Fatalf("no choose-cards prompt for %s: %+v", chooser, g.PendingChoices)
}

func TestStd51DeepfathomEchoExploresThenMayCopyAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := std51Seats(g)
	echo := effectstest.PushCatalogPermanent(g, me.ID, "Deepfathom Echo", "Creature — Merfolk Spirit", std51EchoOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Grizzly Bear", "Creature — Bear", 2, 2)
	effectstest.PutOnTopForTest(me, "Some Sorcery", "Sorcery")
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	// The explore comes first, and its question is answered before the copy.
	if got := effectstest.PlusOneCounters(g, echo); got != 1 {
		t.Fatalf("the Echo has %d counters after exploring, want 1", got)
	}
	effectstest.AnswerExploreQuestion(t, g, me.ID, false)
	std51ChooseCards(t, g, me.ID, bear)
	c, _ := g.LookupCardForEffect(echo)
	if c.Name != "Grizzly Bear" {
		t.Errorf("the Echo is %q, want a copy of Grizzly Bear", c.Name)
	}
	if got := effectstest.PlusOneCounters(g, echo); got != 1 {
		t.Errorf("the copy lost its counter: %d", got)
	}
}

func TestStd51DeepfathomEchoCanDeclineTheCopy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := std51Seats(g)
	echo := effectstest.PushCatalogPermanent(g, me.ID, "Deepfathom Echo", "Creature — Merfolk Spirit", std51EchoOracle, false)
	effectstest.B12Creature(g, me.ID, "Grizzly Bear", "Creature — Bear", 2, 2)
	effectstest.PutOnTopForTest(me, "Forest", "Basic Land — Forest")
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	std51ChooseCards(t, g, me.ID)
	if c, _ := g.LookupCardForEffect(echo); c.Name != "Deepfathom Echo" {
		t.Errorf("the Echo is %q, want itself", c.Name)
	}
}

// std51Pay puts mana such as "{2}{W}" into the pool, generic mana as colorless.
// std51BlockRules counts live power-at-least block-rule records.
func std51BlockRules(g *game.Game) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, e := range g.ScopedEffects {
			for _, m := range e.Mods {
				if m.Kind == game.ModCantBeBlockedByPowerAtLeast {
					n++
					break
				}
			}
		}
	})
	return n
}

func TestStd51MatzalantliTransformsOnlyWithFourPermanentTypesInYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := std51Seats(g)
	effectstest.AdvanceToMain(t, g)
	door := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), OracleID: std51MatzalantliOracle, Layout: game.LayoutTransform,
		Owner: me.ID, Controller: me.ID,
		Faces: []game.Face{
			{Name: "Matzalantli, the Great Door", TypeLine: "Legendary Artifact", ManaCost: "{3}"},
			{Name: "The Core", TypeLine: "Legendary Land"},
		},
	})
	// An artifact creature is two types in one card; an instant is none.
	effectstest.Std51Graveyard(me, "Artifact Creature — Golem", "Enchantment", "Instant")
	effectstest.Std51Pay(g, me, "{4}")
	if err := g.ActivateCatalogAbility(me.ID, door, 1, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("transformed with three permanent types in the graveyard")
	}
	effectstest.Std51Graveyard(me, "Basic Land — Forest")
	if err := g.ActivateCatalogAbility(me.ID, door, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("four permanent types: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := g.LookupCardForEffect(door); c.Name != "The Core" {
		t.Errorf("Matzalantli is %q after transforming, want The Core", c.Name)
	}
}

func TestStd51TheCoreAddsXManaOfOneColorForPermanentCardsInYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := std51Seats(g)
	effectstest.AdvanceToMain(t, g)
	core := effectstest.PushCatalogPermanent(g, me.ID, "The Core", "Legendary Land", std51MatzalantliOracle+"#1", false)
	effectstest.Std51Graveyard(me, "Artifact", "Creature — Bear", "Instant", "Basic Land — Forest", "Sorcery")
	effectstest.B28TapForMana(t, g, me.ID, core, "G")
	got := effectstest.PoolColors(me)
	if len(got) != 3 {
		t.Fatalf("pool %v, want three mana (the instant and sorcery do not count)", got)
	}
	for _, c := range got {
		if c != "G" {
			t.Errorf("pool %v, want one colour throughout", got)
		}
	}
}

// --- Deep-Cavern Bat ---------------------------------------------------------------

func std51BatTable(t *testing.T) (g *game.Game, me, opp *game.Player, hand []uuid.UUID, bat uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me, opp = std51Seats(g)
	hand = effectstest.RevealHand(opp, effectstest.RhForest(), effectstest.RhBolt())
	bat = effectstest.CastAndResolveCreature(t, g, "Deep-Cavern Bat", "Creature — Bat", std51BatOracle)
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	return
}

func TestStd51DeepCavernBatLooksAtTheHandAndExilesANonlandCardUntilItLeaves(t *testing.T) {
	g, me, opp, hand, bat := std51BatTable(t)
	for _, id := range hand {
		card, _ := g.LookupCardForEffect(id)
		if !card.IsKnownTo(me.ID) {
			t.Errorf("%s was not looked at", card.Name)
		}
		if other := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]; card.IsKnownTo(other.ID) {
			t.Errorf("%s was shown to the table: it is a look, not a reveal", card.Name)
		}
	}
	pick := effectstest.OpenVariant(t, g)
	if !effectstest.SameIDs(pick.DiscardOptions, hand[1:]) {
		t.Fatalf("offered %v, want the nonland card only", pick.DiscardOptions)
	}
	if err := g.ResolvePendingChoice(pick.ID, me.ID, []uuid.UUID{hand[1]}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(hand[1]) || opp.Hand.Contains(hand[1]) {
		t.Fatal("the Bolt was not exiled from their hand")
	}
	effectstest.LeaveBattlefield(t, g, me.ID, bat)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Exile.Contains(hand[1]) || !opp.Hand.Contains(hand[1]) {
		t.Error("the Bolt did not return to its owner's hand when the Bat left")
	}
}

func TestStd51DeepCavernBatMayExileNothing(t *testing.T) {
	g, me, opp, hand, _ := std51BatTable(t)
	pick := effectstest.OpenVariant(t, g)
	if err := g.ResolvePendingChoice(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("declining: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Size() != len(hand) {
		t.Errorf("their hand has %d cards, want %d", opp.Hand.Size(), len(hand))
	}
}

// --- Archon of the Wild Rose ---------------------------------------------------------

func TestStd51ArchonMakesYourEnchantedCreaturesFourFourFlyers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := std51Seats(g)
	archon := effectstest.PushCatalogPermanent(g, me.ID, "Archon of the Wild Rose", "Creature — Archon", std51ArchonOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	plain := effectstest.B12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	enemyAura := effectstest.B12Permanent(g, opp.ID, "Their Aura", "Enchantment — Aura")
	woeB1Attach(t, g, enemyAura, plain)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 2 || tough != 2 {
		t.Fatalf("an unenchanted creature is %d/%d, want 2/2", p, tough)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, plain); p != 2 || tough != 2 {
		t.Errorf("an opponent's Aura enchanting my creature counts: %d/%d", p, tough)
	}
	myAura := effectstest.B12Permanent(g, me.ID, "My Aura", "Enchantment — Aura")
	woeB1Attach(t, g, myAura, bear)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("an enchanted creature is %d/%d, want 4/4", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("an enchanted creature has no flying")
	}
	// "Other": the Archon enchanted by its own controller's Aura is unchanged,
	// and an opponent's enchanted creature is not "you control".
	archonAura := effectstest.B12Permanent(g, me.ID, "Archon Aura", "Enchantment — Aura")
	woeB1Attach(t, g, archonAura, archon)
	if p, tough := effectstest.EffectivePTOf(t, g, archon); p != 1 || tough != 1 {
		t.Errorf("the Archon is %d/%d, want its own 1/1", p, tough)
	}
	theirAuraOfMine := effectstest.B12Permanent(g, me.ID, "Aura on theirs", "Enchantment — Aura")
	woeB1Attach(t, g, theirAuraOfMine, theirs)
	if p, tough := effectstest.EffectivePTOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("an opponent's creature is %d/%d, want 2/2", p, tough)
	}
	// It follows the Aura out.
	effectstest.WoeB1Destroy(t, g, myAura)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("after the Aura left: %d/%d, want 2/2", p, tough)
	}
}

// --- Archon's Glory ------------------------------------------------------------------

func TestStd51ArchonsGloryPumpsAndOnlyBargainedGivesFlyingAndLifelink(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := std51Seats(g)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Archon's Glory", "Instant", std51GloryOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the 2/2 is %d/%d, want 4/4", p, tough)
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "flying") || effectstest.HasEffectiveKeyword(t, g, bear, "lifelink") {
		t.Error("an unbargained cast gave keywords")
	}
}

func TestStd51ArchonsGloryBargainedAlsoGainsFlyingAndLifelink(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := std51Seats(g)
	token := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Treasure", TypeLine: "Token Artifact — Treasure", Owner: me.ID, Controller: me.ID,
	})
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.P1g2Cast(t, g, "Archon's Glory", "Instant", std51GloryOracle, "", game.CastSpellParams{
		Targets: effectstest.CardRefs(bear), OptionalCosts: []int{0}, SacrificeIDs: []uuid.UUID{token},
	})
	if g.Battlefield.Contains(token) {
		t.Fatal("the bargained token was not sacrificed")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the 2/2 is %d/%d, want 4/4", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "flying") || !effectstest.HasEffectiveKeyword(t, g, bear, "lifelink") {
		t.Error("a bargained cast did not give flying and lifelink")
	}
}

// --- Break the Spell -----------------------------------------------------------------

// std51Break casts Break the Spell at an enchantment and returns the caster's hand
// size before and after it resolves.
func std51Break(t *testing.T, g *game.Game, me *game.Player, target uuid.UUID) (before, after int) {
	t.Helper()
	effectstest.CastCatalogSpell(t, g, "Break the Spell", "Instant", std51BreakOracle, effectstest.CardRefs(target))
	before = effectstest.HandSize(me)
	effectstest.PassPriorityAroundTable(t, g)
	return before, effectstest.HandSize(me)
}

func TestStd51BreakTheSpellDrawsOnlyForYourPermanentOrAToken(t *testing.T) {
	for _, tc := range []struct {
		name  string
		owner func(me, opp *game.Player) *game.Player
		token bool
		draws bool
	}{
		{"an opponent's enchantment", func(_, opp *game.Player) *game.Player { return opp }, false, false},
		{"your own enchantment", func(me, _ *game.Player) *game.Player { return me }, false, true},
		{"an opponent's token enchantment", func(_, opp *game.Player) *game.Player { return opp }, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := std51Seats(g)
			tl := "Enchantment"
			if tc.token {
				tl = "Token Enchantment"
			}
			target := effectstest.B12Permanent(g, tc.owner(me, opp).ID, "Some Enchantment", tl)
			before, after := std51Break(t, g, me, target)
			if g.Battlefield.Contains(target) {
				t.Fatal("the enchantment was not destroyed")
			}
			if want := before; tc.draws {
				want++
				if after != want {
					t.Errorf("hand %d -> %d, want a card drawn", before, after)
				}
			} else if after != want {
				t.Errorf("hand %d -> %d, want no draw", before, after)
			}
		})
	}
}

func TestStd51BreakTheSpellDrawsNothingForAnIndestructibleEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := std51Seats(g)
	target := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Eternal Anthem", TypeLine: "Enchantment",
		Keywords: []string{"indestructible"}, Owner: me.ID, Controller: me.ID,
	})
	before, after := std51Break(t, g, me, target)
	if !g.Battlefield.Contains(target) {
		t.Fatal("an indestructible enchantment was destroyed")
	}
	if after != before {
		t.Errorf("hand %d -> %d, want no draw", before, after)
	}
}

// --- Cheeky House-Mouse // Squeak By ----------------------------------------------------

func TestStd51SqueakByPumpsAndBarsBlockersWithPowerThreeOrMore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := std51Seats(g)
	runner := effectstest.PushVanillaCreature(g, me.ID, "Runner", 2, 2)
	small := effectstest.PushVanillaCreature(g, opp.ID, "Power Two", 2, 2)
	big := effectstest.PushVanillaCreature(g, opp.ID, "Power Three", 3, 3)
	card := woeB1Adventure(me, std51MouseOracle,
		game.Face{Name: "Cheeky House-Mouse", TypeLine: "Creature — Mouse", ManaCost: "{W}", Colors: []string{"W"}, Power: 2, Toughness: 1},
		game.Face{Name: "Squeak By", TypeLine: "Sorcery — Adventure", ManaCost: "{W}", Colors: []string{"W"}})
	woeB1CastAdventure(t, g, me, card, effectstest.CardRefs(runner))
	if p, tough := effectstest.EffectivePTOf(t, g, runner); p != 3 || tough != 3 {
		t.Fatalf("the 2/2 is %d/%d, want 3/3", p, tough)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(runner, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	effectstest.BrRefusal(t, g.DeclareBlocker(big, runner), game.BlockReasonCantBeBlockedBy)
	if err := g.DeclareBlocker(small, runner); err != nil {
		t.Fatalf("a power 2 creature blocks: %v", err)
	}
	if n := std51BlockRules(g); n != 1 {
		t.Errorf("%d block rules, want the one record", n)
	}
	effectstest.AdvanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	if n := std51BlockRules(g); n != 0 {
		t.Errorf("%d block rules outlived the turn", n)
	}
}

// --- Cooped Up ---------------------------------------------------------------------------

func TestStd51CoopedUpStopsAttackingAndItsControllerMayExileTheCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := std51Seats(g)
	victim := effectstest.PushVanillaCreature(g, me.ID, "Enchanted", 2, 2)
	aura := effectstest.CastCatalogSpell(t, g, "Cooped Up", "Enchantment — Aura", std51CoopedUpOracle, effectstest.CardRefs(victim))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(aura) {
		t.Fatal("Cooped Up did not resolve")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(victim, opp.ID); err == nil {
		t.Error("the enchanted creature attacked")
	}
	// Only the Aura's controller may activate it.
	effectstest.Std51Pay(g, opp, "{2}{W}")
	if err := g.ActivateCatalogAbility(opp.ID, aura, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("an opponent activated Cooped Up")
	}
	effectstest.Std51Pay(g, me, "{2}{W}")
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(victim) || g.Battlefield.Contains(victim) {
		t.Error("the enchanted creature was not exiled")
	}
}

func TestStd51CoopedUpStillExilesTheCreatureIfItLeavesInResponse(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := std51Seats(g)
	victim := effectstest.PushVanillaCreature(g, me.ID, "Enchanted", 2, 2)
	aura := effectstest.CastCatalogSpell(t, g, "Cooped Up", "Enchantment — Aura", std51CoopedUpOracle, effectstest.CardRefs(victim))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.Std51Pay(g, me, "{2}{W}")
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.LeaveBattlefield(t, g, me.ID, aura)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(victim) {
		t.Error("the creature it enchanted was not exiled by the ability that was already on the stack")
	}
}

// --- Intrepid Paleontologist ---------------------------------------------------------------

func std51PaleoTable(t *testing.T) (g *game.Game, me, opp *game.Player, paleo uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me, opp = std51Seats(g)
	effectstest.AdvanceToMain(t, g)
	paleo = effectstest.PushCatalogPermanent(g, me.ID, "Intrepid Paleontologist", "Creature — Human Druid", std51PaleontologistOracle, false)
	return
}

func std51Dino(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: "{1}{G}", Power: 2, Toughness: 2,
		Owner: p.ID, Controller: p.ID,
	})
	return id
}

func std51ExileWith(t *testing.T, g *game.Game, me *game.Player, paleo, target uuid.UUID) {
	t.Helper()
	effectstest.Std51Pay(g, me, "{2}")
	if err := g.ActivateCatalogAbility(me.ID, paleo, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
	}); err != nil {
		t.Fatalf("exile ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(target) {
		t.Fatal("the card was not exiled")
	}
}

func TestStd51PaleontologistTapsForAnyColor(t *testing.T) {
	g, me, _, paleo := std51PaleoTable(t)
	effectstest.B28TapForMana(t, g, me.ID, paleo, "R")
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "R" {
		t.Errorf("pool %v, want one red mana", got)
	}
}

func TestStd51PaleontologistCastsAnExiledDinosaurWithAFinalityCounter(t *testing.T) {
	g, me, _, paleo := std51PaleoTable(t)
	raptor := std51Dino(me, "Raptor", "Creature — Dinosaur")
	std51ExileWith(t, g, me, paleo, raptor)
	effectstest.Std51Pay(g, me, "{1}{G}")
	if err := g.CastSpell(me.ID, raptor, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("cast the exiled Dinosaur: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(raptor) {
		t.Fatal("the Dinosaur did not resolve")
	}
	if got := effectstest.CountersOn(g, raptor, game.CounterFinality); got != 1 {
		t.Errorf("finality counters = %d, want 1", got)
	}
	// Finality: it is exiled rather than dying.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(raptor) })
	if !g.Exile.Contains(raptor) || me.Graveyard.Contains(raptor) {
		t.Error("the Dinosaur with a finality counter was not exiled instead of dying")
	}
}

func TestStd51PaleontologistOnlyCastsDinosaurCardsYouOwnExiledWithIt(t *testing.T) {
	g, me, opp, paleo := std51PaleoTable(t)
	bear := std51Dino(me, "Bear", "Creature — Bear")
	std51ExileWith(t, g, me, paleo, bear)
	effectstest.Std51Pay(g, me, "{1}{G}")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Error("cast a non-Dinosaur card exiled with it")
	}
	theirs := std51Dino(opp, "Their Raptor", "Creature — Dinosaur")
	std51ExileWith(t, g, me, paleo, theirs)
	if err := g.CastSpell(me.ID, theirs, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Error("cast a Dinosaur card it does not own")
	}
	// Exiled by something else: not "exiled with this creature".
	other := uuid.New()
	g.Exile.PushTop(game.Card{InstanceID: other, Name: "Loose Raptor", TypeLine: "Creature — Dinosaur",
		ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, other, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Error("cast a Dinosaur it did not exile")
	}
}
