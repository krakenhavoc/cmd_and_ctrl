package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b627_test.go — Judgment cards, S82 OLD batch 627 (#5611), played
// through the real catalog entries.

const (
	o627Rootcaster = "0aaefa00-86f4-4e12-87ae-3a362d862fa4"
	o627Researcher = "240a349a-79f8-4e11-9a05-a78148e70a62"
	o627Barbarian  = "9c0fd5ad-fbba-4598-80cf-6540888378cb"
	o627Driller    = "5967c3d3-2ef1-4496-b51c-23478fa77f09"
	o627Vines      = "6a3052a1-7a0b-4364-883d-cd2abca39c06"
	o627Behemoth   = "77a000de-76d9-44f8-9789-875aa9f576e7"
	o627Fogbringer = "3fb3d4b3-1034-449c-864b-633e7f39d73f"
	o627Spelljack  = "7687b2a7-816d-4416-979b-675e35e235fc"
	o627Aerie      = "69565f6e-2e23-4cd8-bd21-1283e06729a3"
	o627Advocate   = "c8c5cd28-4f27-424d-92bc-e7d6f15a5f9f"
	o627Morality   = "82e36ba9-9652-4594-8660-86c766b47709"
	o627LeadAstray = "9e8e5ad4-05a3-4757-8797-c37bf49506d5"
)

func TestS82OldB627CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		o627Rootcaster: "Centaur Rootcaster", o627Researcher: "Hapless Researcher", o627Barbarian: "Spellgorger Barbarian",
		o627Driller: "Dwarven Driller", o627Vines: "Venomous Vines", o627Behemoth: "Wormfang Behemoth",
		o627Fogbringer: "Aven Fogbringer", o627Spelljack: "Spelljack", o627Aerie: "Soulcatchers' Aerie",
		o627Advocate: "Forcemage Advocate", o627Morality: "Morality Shift", o627LeadAstray: "Lead Astray",
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

func TestS82OldB627CentaurRootcasterFetchesATappedBasicOnCombatDamageToAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	centaur := effectstest.B12Push(g, me.ID, "Centaur Rootcaster", "Creature — Centaur Druid", o627Rootcaster, 2, 2)
	forest := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: forest, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	effectstest.DealCombatDamageToPlayer(g, centaur, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := g.LookupCardForEffect(forest)
	if !ok || !g.Battlefield.Contains(forest) || !c.Tapped {
		t.Fatal("the Forest is not on the battlefield tapped")
	}
}

func TestS82OldB627HaplessResearcherLoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	researcher := effectstest.B12Push(g, me.ID, "Hapless Researcher", "Creature — Human Wizard", o627Researcher, 1, 1)
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, researcher, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityUntilChoice(t, g)
	if effectstest.FindBattlefieldCardForTest(g, researcher) != nil {
		t.Error("the Researcher was not sacrificed")
	}
	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand after the draw = %d, want %d", me.Hand.Size(), hand+1)
	}
	effectstest.DiscardFromHand(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand after the discard = %d, want %d", me.Hand.Size(), hand)
	}
}

func TestS82OldB627SpellgorgerBarbarianDiscardsAtRandomThenDrawsWhenItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hand := me.Hand.Size()
	barb := effectstest.CastCatalogSpell(t, g, "Spellgorger Barbarian", "Creature — Human Nightmare Barbarian", o627Barbarian, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand-1 {
		t.Fatalf("hand after entering = %d, want %d (a card discarded at random)", me.Hand.Size(), hand-1)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(barb) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand after leaving = %d, want %d (a card drawn)", me.Hand.Size(), hand)
	}
}

func TestS82OldB627DwarvenDrillerDestroysTheLandUnlessItsControllerTakesTwo(t *testing.T) {
	for _, takeDamage := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		driller := effectstest.B12Push(g, me.ID, "Dwarven Driller", "Creature — Dwarf", o627Driller, 2, 2)
		land := effectstest.B12Push(g, opp.ID, "Their Island", "Basic Land — Island", "", 0, 0)
		life := opp.Life
		if err := g.ActivateCatalogAbility(me.ID, driller, 0, game.ActivateAbilityParams{Strict: true, Targets: o627Target(land)}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		effectstest.PassPriorityUntilChoice(t, g)
		effectstest.AnswerMayChoice(t, g, opp.ID, takeDamage)
		effectstest.PassPriorityAroundTable(t, g)
		gone := effectstest.FindBattlefieldCardForTest(g, land) == nil
		if takeDamage {
			if gone || opp.Life != life-2 {
				t.Errorf("took damage: land gone %v, life %d -> %d, want the land kept and 2 damage", gone, life, opp.Life)
			}
		} else if !gone || opp.Life != life {
			t.Errorf("refused: land gone %v, life %d -> %d, want the land destroyed and no damage", gone, life, opp.Life)
		}
	}
}

func TestS82OldB627VenomousVinesNeedsAnEnchantedPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	plain := effectstest.AuraBear(g, opp.ID)
	enchanted := effectstest.AuraBear(g, opp.ID)
	effectstest.Cmdr2B19Attach(g, me.ID, enchanted, "Pacifism-like Aura", "")
	if err := effectstest.CastCatalogSpellErr(t, g, "Venomous Vines", "Sorcery", o627Vines, o627Target(plain)); err == nil {
		t.Fatal("Venomous Vines targeted an unenchanted creature")
	}
	effectstest.CastCatalogSpell(t, g, "Venomous Vines", "Sorcery", o627Vines, o627Target(enchanted))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldCardForTest(g, enchanted) != nil {
		t.Error("the enchanted creature survived")
	}
	if effectstest.FindBattlefieldCardForTest(g, plain) == nil {
		t.Error("the unenchanted creature was destroyed")
	}
}

func TestS82OldB627WormfangBehemothExilesYourHandAndReturnsItWhenItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	behemoth := effectstest.CastCatalogSpell(t, g, "Wormfang Behemoth", "Creature — Nightmare Fish Beast", o627Behemoth, nil)
	hand := me.Hand.Size()
	if hand == 0 {
		t.Fatal("no cards in hand to exile")
	}
	exileBefore := g.Exile.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != 0 {
		t.Fatalf("hand after entering = %d, want 0", me.Hand.Size())
	}
	if g.Exile.Size() != exileBefore+hand {
		t.Fatalf("exile holds %d, want %d", g.Exile.Size(), exileBefore+hand)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(behemoth) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand after the Behemoth left = %d, want %d", me.Hand.Size(), hand)
	}
}

func TestS82OldB627AvenFogbringerBouncesATargetLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	land := effectstest.B12Push(g, opp.ID, "Their Island", "Basic Land — Island", "", 0, 0)
	effectstest.CastCatalogSpell(t, g, "Aven Fogbringer", "Creature — Bird Wizard", o627Fogbringer, nil)
	effectstest.PassPriorityUntilChoice(t, g)
	effectstest.AnswerPickTargetIfOpen(t, g, land)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldCardForTest(g, land) != nil {
		t.Error("the land is still on the battlefield")
	}
	if !opp.Hand.Contains(land) {
		t.Error("the land is not in its owner's hand")
	}
}

func TestS82OldB627SpelljackCountersExilesAndLetsYouPlayItFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	spell := effectstest.CastCatalogSpell(t, g, "Test Enchantment", "Enchantment", "", nil)
	jack := effectstest.HandCardForTest(me, "Spelljack", "Instant", o627Spelljack)
	if err := g.CastSpell(me.ID, jack, game.CastSpellParams{Targets: o627Target(spell)}); err != nil {
		t.Fatalf("cast Spelljack: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(spell) {
		t.Fatal("the countered spell was not exiled")
	}
	if me.Graveyard.Contains(spell) {
		t.Error("the countered spell went to the graveyard")
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Fatalf("play the exiled spell for free: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldCardForTest(g, spell) == nil {
		t.Error("the spell cast from exile did not resolve")
	}
}

func TestS82OldB627SoulcatchersAerieCountsYourBirdsAndPumpsEveryBird(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	aerie := effectstest.PushCatalogPermanent(g, me.ID, "Soulcatchers' Aerie", "Enchantment", o627Aerie, false)
	mine := effectstest.B12Creature(g, me.ID, "My Bird", "Creature — Bird", 1, 1)
	mine2 := effectstest.B12Creature(g, me.ID, "My Other Bird", "Creature — Bird", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bird", "Creature — Bird", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.KillCreature(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, aerie, "feather"); got != 0 {
		t.Fatalf("an opponent's Bird went to their graveyard and the Aerie has %d feather counters", got)
	}
	effectstest.KillCreature(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, aerie, "feather"); got != 1 {
		t.Fatalf("feather counters = %d, want 1", got)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, mine2); p != 2 || tough != 2 {
		t.Errorf("remaining Bird = %d/%d, want 2/2", p, tough)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("a non-Bird = %d/%d, want 2/2", p, tough)
	}
	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, aerie, "feather"); got != 1 {
		t.Errorf("a non-Bird added a feather counter: %d", got)
	}
}

func TestS82OldB627ForcemageAdvocateReturnsAnOpponentsCardAndAddsACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advocate := effectstest.B12Push(g, me.ID, "Forcemage Advocate", "Creature — Centaur Shaman", o627Advocate, 2, 1)
	card := effectstest.PushGraveyardCardWithTypeLineID(opp, "Their Card", "Sorcery")
	mine := effectstest.PushGraveyardCardWithTypeLineID(me, "My Card", "Sorcery")
	bear := effectstest.AuraBear(g, me.ID)
	if err := g.ActivateCatalogAbility(me.ID, advocate, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: bear},
	}}); err == nil {
		t.Fatal("the Advocate targeted a card in its own controller's graveyard")
	}
	if err := g.ActivateCatalogAbility(me.ID, advocate, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: card}, {Kind: game.TargetCard, ID: bear},
	}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(card) {
		t.Error("the card did not go to its owner's hand")
	}
	if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters on the Bear = %d, want 1", got)
	}
}

func TestS82OldB627MoralityShiftSwapsGraveyardAndLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 3; i++ {
		effectstest.PushGraveyardCardWithTypeLineID(me, "Grave Card", "Sorcery")
	}
	lib, gy := me.Library.Size(), me.Graveyard.Size()
	effectstest.CastCatalogSpell(t, g, "Morality Shift", "Sorcery", o627Morality, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// The spell itself is in the graveyard once it resolves.
	if me.Library.Size() != gy {
		t.Errorf("library = %d, want the old graveyard's %d", me.Library.Size(), gy)
	}
	if me.Graveyard.Size() != lib+1 {
		t.Errorf("graveyard = %d, want the old library's %d plus Morality Shift", me.Graveyard.Size(), lib)
	}
	for _, c := range me.Library.Cards {
		if c.Name != "Grave Card" {
			t.Errorf("a card named %q is in the library, want only the old graveyard cards", c.Name)
		}
	}
}

func TestS82OldB627LeadAstrayTapsUpToTwoCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, opp.ID, "C", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Lead Astray", "Instant", o627LeadAstray, []game.TargetRef{
		{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}, {Kind: game.TargetCard, ID: c},
	}); err == nil {
		t.Fatal("Lead Astray took three targets")
	}
	effectstest.CastCatalogSpell(t, g, "Lead Astray", "Instant", o627LeadAstray, []game.TargetRef{
		{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b},
	})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if tapped, _ := effectstest.BattlefieldCardTapped(g, id); !tapped {
			t.Errorf("%s was not tapped", id)
		}
	}
	if tapped, _ := effectstest.BattlefieldCardTapped(g, c); tapped {
		t.Error("the untargeted creature was tapped")
	}
}

func o627Target(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}
