package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b722_cards_test.go — Urza's Saga card draw and library, S82 OLD batch
// 722 (#5624). Every test plays the real catalog entry through the engine.

const (
	b722Remembrance = "02d0b0fd-ff73-4e26-be44-d85c743d51e4"
	b722Scrap       = "0ad3b788-eb3d-4876-9c89-893429595e0e"
	b722Serpent     = "30fce201-03a8-4afc-a1ad-7a9b3fd59836"
	b722Rejuvenate  = "35069cdb-e5bd-4224-a1a8-b329054e003b"
	b722Rescind     = "3f487c8e-a923-4031-b0a7-04f778544fe0"
	b722Expunge     = "4b1fee86-360d-41d8-9238-eff216376732"
	b722Skullcap    = "7694887a-2f3b-4ef1-b598-f3323f018ef4"
	b722Merfolk     = "7d6e5a88-a7fd-466c-a9f8-5015e3400176"
	b722Drake       = "8d8981df-4cdc-4fd4-acd8-095fa3281281"
	b722Barrier     = "94674840-c73c-4659-80b9-11140d1deaa3"
	b722Bounty      = "c5a61e61-d5fe-41ef-afd6-ba5e0399b986"
	b722TimeSpiral  = "cac55e46-b730-4569-b92c-a4b5922fc20c"
	b722Hush        = "eaa0c6c2-eee5-443a-867f-ab77b995b6ef"
)

type b722Card struct {
	oracle, name, typeLine, cost string
	cycles                       bool
}

var b722Cards = []b722Card{
	{b722Remembrance, "Remembrance", "Enchantment", "{3}{W}", false},
	{b722Scrap, "Scrap", "Instant", "{2}{R}", true},
	{b722Serpent, "Sandbar Serpent", "Creature — Serpent", "{4}{U}", true},
	{b722Rejuvenate, "Rejuvenate", "Sorcery", "{3}{G}", true},
	{b722Rescind, "Rescind", "Instant", "{1}{U}{U}", true},
	{b722Expunge, "Expunge", "Instant", "{2}{B}", true},
	{b722Skullcap, "Grafted Skullcap", "Artifact", "{4}", false},
	{b722Merfolk, "Sandbar Merfolk", "Creature — Merfolk", "{U}", true},
	{b722Drake, "Pendrell Drake", "Creature — Drake", "{3}{U}", true},
	{b722Barrier, "Shimmering Barrier", "Creature — Wall", "{1}{W}", true},
	{b722Bounty, "Gaea's Bounty", "Sorcery", "{2}{G}", false},
	{b722TimeSpiral, "Time Spiral", "Sorcery", "{4}{U}{U}", false},
	{b722Hush, "Hush", "Sorcery", "{3}{G}", true},
}

func TestB722EveryCardIsRegisteredFull(t *testing.T) {
	for _, c := range b722Cards {
		spec, ok := effects.Lookup(c.oracle)
		if !ok {
			t.Errorf("%s is not registered", c.name)
			continue
		}
		if spec.Name != c.name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", c.name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// Cycling {2}: the card is discarded as a cost and its owner draws a card; it
// works only from the hand and only with {2} to pay.
func TestB722CyclingCardsCycleForTwo(t *testing.T) {
	for _, c := range b722Cards {
		if !c.cycles {
			continue
		}
		t.Run(c.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			id := uuid.New()
			me.Hand.PushTop(game.Card{InstanceID: id, Name: c.name, TypeLine: c.typeLine, OracleID: c.oracle, ManaCost: c.cost, Owner: me.ID, Controller: me.ID})
			// Not with a single mana.
			effectstest.FloatForTest(g, me, "{C}")
			if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
				t.Fatal("cycled for {1}")
			}
			if !me.Hand.Contains(id) {
				t.Fatal("a refused cycle moved the card")
			}
			effectstest.FloatForTest(g, me, "{C}")
			hand := len(me.Hand.Cards)
			if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
				t.Fatalf("cycle: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if me.Hand.Contains(id) || !me.Graveyard.Contains(id) {
				t.Error("the cycled card is not in the graveyard")
			}
			if len(me.Hand.Cards) != hand {
				t.Errorf("hand %d → %d, want the discard and the draw to cancel", hand, len(me.Hand.Cards))
			}
		})
	}
}

func TestB722ScrapDestroysOnlyAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Scrap", "Instant", b722Scrap, effectstest.CardTarget(bear)); err == nil {
		t.Fatal("a nonartifact creature was a legal target")
	}
	effectstest.CastCatalogSpell(t, g, "Scrap", "Instant", b722Scrap, effectstest.CardTarget(rock))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || !opp.Graveyard.Contains(rock) {
		t.Error("the artifact was not destroyed")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the bear died")
	}
}

func TestB722ExpungeDestroysOnlyANonartifactNonblackCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	black := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Black Bear", TypeLine: "Creature — Bear", Colors: []string{"B"}, Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	golem := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Golem", TypeLine: "Artifact Creature — Golem", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	green := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Green Bear", TypeLine: "Creature — Bear", Colors: []string{"G"}, Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	for _, bad := range []uuid.UUID{black, golem} {
		if err := effectstest.CastCatalogSpellErrStrict(t, g, "Expunge", "Instant", b722Expunge, effectstest.CardTarget(bad)); err == nil {
			t.Fatal("a black or artifact creature was a legal target")
		}
	}
	effectstest.CastCatalogSpell(t, g, "Expunge", "Instant", b722Expunge, effectstest.CardTarget(green))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(green) || !opp.Graveyard.Contains(green) {
		t.Error("the green creature was not destroyed")
	}
}

func TestB722RescindReturnsAnyPermanentToItsOwnersHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	land := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	effectstest.CastCatalogSpell(t, g, "Rescind", "Instant", b722Rescind, effectstest.CardTarget(land))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) || !opp.Hand.Contains(land) {
		t.Error("the land did not return to its owner's hand")
	}
}

func TestB722HushDestroysEveryEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Permanent(g, me.ID, "My Aura", "Enchantment — Aura")
	theirs := effectstest.B12Permanent(g, opp.ID, "Their Enchantment", "Enchantment")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Hush", "Sorcery", b722Hush, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("an enchantment survived")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("a nonenchantment was destroyed")
	}
}

func TestB722RejuvenateGainsSixLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Rejuvenate", "Sorcery", b722Rejuvenate, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+6 {
		t.Errorf("life %d, want %d", me.Life, life+6)
	}
}

func TestB722GaeasBountyFindsUpToTwoForests(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Stomping Ground", TypeLine: "Land — Mountain Forest"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Island", TypeLine: "Basic Land — Island"},
	)
	hand := len(me.Hand.Cards)
	effectstest.CastCatalogSpell(t, g, "Gaea's Bounty", "Sorcery", b722Bounty, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if len(c.SearchCards) != 3 || effectstest.SearchOptionNamed(g, c, "Island") != uuid.Nil {
		t.Fatalf("%d cards offered, want the three Forest cards only", len(c.SearchCards))
	}
	effectstest.BeginSearchIfAsked(t, g, c)
	if err := g.ResolveSearchLibrary(c.ID, me.ID, c.SearchCards[:3]); err == nil {
		t.Error("three cards were accepted for an up-to-two search")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, c.SearchCards[:2]); err != nil {
		t.Fatalf("taking two Forests: %v", err)
	}
	if len(me.Hand.Cards) != hand+2 {
		t.Errorf("hand %d → %d, want two cards in", hand, len(me.Hand.Cards))
	}
}

func TestB722GaeasBountyMayFindFewer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Island", TypeLine: "Basic Land — Island"},
	)
	hand := len(me.Hand.Cards)
	effectstest.CastCatalogSpell(t, g, "Gaea's Bounty", "Sorcery", b722Bounty, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	effectstest.BeginSearchIfAsked(t, g, c)
	if err := g.ResolveSearchLibrary(c.ID, me.ID, c.SearchCards[:1]); err != nil {
		t.Fatalf("taking the one Forest: %v", err)
	}
	if len(me.Hand.Cards) != hand+1 {
		t.Errorf("hand %d → %d, want one card in", hand, len(me.Hand.Cards))
	}
}

func TestB722TimeSpiralShufflesEveryHandAndGraveyardDrawsSevenAndUntapsSix(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	for _, p := range []*game.Player{me, opp} {
		p.Library.Cards = nil
		for i := 0; i < 20; i++ {
			effectstest.PushLibraryCardForTest(p, game.Card{Name: "Filler", TypeLine: "Sorcery"})
		}
		p.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Old Card", TypeLine: "Instant", Owner: p.ID, Controller: p.ID})
	}
	var lands []uuid.UUID
	for i := 0; i < 8; i++ {
		lands = append(lands, effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Tapped: true, Owner: me.ID, Controller: me.ID,
		}))
	}
	spell := effectstest.CastCatalogSpell(t, g, "Time Spiral", "Sorcery", b722TimeSpiral, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, spell) || me.Graveyard.Contains(spell) {
		t.Error("Time Spiral is not exiled")
	}
	for _, p := range []*game.Player{me, opp} {
		if len(p.Hand.Cards) != 7 {
			t.Errorf("%s has %d cards in hand, want 7", p.ID, len(p.Hand.Cards))
		}
		if len(p.Graveyard.Cards) != 0 {
			t.Errorf("%s has %d cards in their graveyard, want none", p.ID, len(p.Graveyard.Cards))
		}
	}
	if effectstest.ChooseCardsChoiceFor(g, me.ID) == nil {
		t.Fatal("no untap prompt: \"untap up to six lands\" is a resolution-time choice")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, lands[:6]...)
	untapped := 0
	for _, id := range lands {
		if !effectstest.IsTapped(g, id) {
			untapped++
		}
	}
	if untapped != 6 {
		t.Errorf("%d lands untapped, want 6", untapped)
	}
}

func TestB722GraftedSkullcapDrawsAnExtraCardInYourDrawStepOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.PushCatalogPermanent(g, g.Seats[0].ID, "Grafted Skullcap", "Artifact", b722Skullcap, false)
	other := g.Seats[1]
	hand := other.Hand.Size()
	effectstest.Batch01AdvanceToStepOf(t, g, 1, game.StepDraw)
	effectstest.PassPriorityAroundTable(t, g)
	if other.Hand.Size() != hand+1 {
		t.Errorf("an opponent drew %d on their own turn, want the ordinary 1", other.Hand.Size()-hand)
	}
	me := g.Seats[0]
	hand = me.Hand.Size()
	effectstest.Batch01AdvanceToStepOf(t, g, 0, game.StepDraw)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("the controller drew %d, want 2", me.Hand.Size()-hand)
	}
}

func TestB722GraftedSkullcapDiscardsYourHandAtYourEndStepOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Grafted Skullcap", "Artifact", b722Skullcap, false)
	// Past my own end step (the game starts on my turn), into an opponent's turn.
	effectstest.Batch01AdvanceToStepOf(t, g, 1, game.StepDraw)
	effectstest.GiveHandCards(me, 3)
	effectstest.Batch01AdvanceToStepOf(t, g, 1, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() == 0 {
		t.Fatal("the controller's hand was emptied on an opponent's end step")
	}
	effectstest.Batch01AdvanceToStepOf(t, g, 0, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != 0 {
		t.Errorf("%d cards left in hand at the controller's end step, want none", me.Hand.Size())
	}
}

func TestB722RemembranceFetchesACopyOfADyingNontokenCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Remembrance", "Enchantment", b722Remembrance, false)
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear"},
		game.Card{Name: "Runeclaw Bear", TypeLine: "Creature — Bear"},
	)
	// A token, and an opponent's creature, do nothing.
	token := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Grizzly Bears", TypeLine: "Token Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	theirs := effectstest.B12Creature(g, opp.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	effectstest.KillCreature(t, g, me.ID, token)
	effectstest.KillCreature(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SearchChoiceFor(g, me.ID) != nil {
		t.Fatal("a token, or an opponent's creature, triggered Remembrance")
	}
	mine := effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	hand := len(me.Hand.Cards)
	effectstest.KillCreature(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if len(c.SearchCards) != 1 || effectstest.SearchOptionNamed(g, c, "Grizzly Bears") == uuid.Nil {
		t.Fatalf("%d cards offered, want the one Grizzly Bears only", len(c.SearchCards))
	}
	effectstest.BeginSearchIfAsked(t, g, c)
	if err := g.ResolveSearchLibrary(c.ID, me.ID, c.SearchCards[:1]); err != nil {
		t.Fatalf("taking the copy: %v", err)
	}
	if len(me.Hand.Cards) != hand+1 {
		t.Errorf("hand %d → %d, want the copy in", hand, len(me.Hand.Cards))
	}
}
