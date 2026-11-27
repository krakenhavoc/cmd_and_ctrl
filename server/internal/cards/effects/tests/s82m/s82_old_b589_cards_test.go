package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b589_cards_test.go — Saviors of Kamigawa, batch 589 (#5601). Each
// test plays the real catalog entry and asserts what its text says.

const (
	b589AyumiOracle      = "12acd466-deb9-4e68-b75b-aed86680b917"
	b589SkyfireOracle    = "169c468e-0e17-450b-8954-df47d78ad8f8"
	b589InfernalOracle   = "4f07a2e6-692b-4d8c-9ecf-4d94b03706a0"
	b589FiddleheadOracle = "7392c58b-1ba9-4cf3-8449-aa5de3c42130"
	b589CloudhoofOracle  = "82d8732b-8246-4fbf-96aa-e192daf527f1"
	b589SoramaroOracle   = "1cf3126e-fbac-4245-beba-3d4eedda24bf"
	b589WineOracle       = "22d85711-485b-47a2-b161-eaed135f1ddc"
	b589BonesetterOracle = "a64684bc-61b2-4261-8749-fc8cae9e2fd5"
	b589EnvoyOracle      = "be70c6e8-6f9f-49fb-ab40-c6ce0ec2077c"
	b589KagemaroOracle   = "da3e399f-6801-4e4d-9b1d-ba15fcfb7ee4"
)

func TestOldB589EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		b589AyumiOracle:      "Ayumi, the Last Visitor",
		b589SkyfireOracle:    "Skyfire Kirin",
		b589InfernalOracle:   "Infernal Kirin",
		b589FiddleheadOracle: "Fiddlehead Kami",
		b589CloudhoofOracle:  "Cloudhoof Kirin",
		b589SoramaroOracle:   "Descendant of Soramaro",
		b589WineOracle:       "Wine of Blood and Iron",
		b589BonesetterOracle: "Kitsune Bonesetter",
		b589EnvoyOracle:      "Oboro Envoy",
		b589KagemaroOracle:   "Kagemaro, First to Suffer",
	}
	for oracle, name := range want {
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

// b589Cast casts a spell of the given cost from the active player's hand
// and leaves it on the stack with its cast triggers pending.
func b589Cast(t *testing.T, g *game.Game, name, typeLine, cost string) uuid.UUID {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return id
}

func b589Creature(g *game.Game, owner uuid.UUID, name, cost string, power, toughness int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Bear", ManaCost: cost,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

func TestOldB589AyumiHasLegendaryLandwalk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	spec, _ := effects.Lookup(b589AyumiOracle)
	if len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "legendary landwalk" {
		t.Errorf("Ayumi's keywords %v, want legendary landwalk", spec.PrintedKeywords)
	}
	id := effectstest.B12Push(g, g.Seats[0].ID, "Ayumi, the Last Visitor", "Legendary Creature — Spirit", b589AyumiOracle, 7, 3)
	if !effectstest.HasKeywordOnBattlefield(t, g, id, "legendary landwalk") {
		t.Error("Ayumi lacks legendary landwalk")
	}
}

func TestOldB589SkyfireKirinStealsACreatureWithTheSpellsManaValue(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.B12Push(g, me.ID, "Skyfire Kirin", "Legendary Creature — Kirin Spirit", b589SkyfireOracle, 3, 3)
		three := b589Creature(g, opp.ID, "Three", "{2}{R}", 2, 2)
		two := b589Creature(g, opp.ID, "Two", "{1}{R}", 2, 2)
		// A Bear spell does not trigger it.
		b589Cast(t, g, "Bear", "Creature — Bear", "{1}")
		if effectstest.LatestPickTarget(g, me.ID) != nil {
			t.Fatal("a Bear spell triggered Skyfire Kirin")
		}
		effectstest.PassPriorityAroundTable(t, g)
		b589Cast(t, g, "A Spirit", "Creature — Spirit", "{1}{U}{U}")
		p := effectstest.LatestPickTarget(g, me.ID)
		if p == nil {
			t.Fatal("no target prompt for a mana value 3 Spirit spell")
		}
		if !effectstest.HasID(p.PickTargetCards, three) || effectstest.HasID(p.PickTargetCards, two) {
			t.Fatalf("candidates %v: want the mana value 3 creature only", p.PickTargetCards)
		}
		effectstest.PickCard(t, g, me.ID, three)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, yes)
		effectstest.PassPriorityAroundTable(t, g)
		got := effectstest.ControllerOf(t, g, three)
		if yes && got != me.ID {
			t.Error("yes: the creature is not under your control")
		}
		if !yes && got != opp.ID {
			t.Error("no: the creature changed hands")
		}
		if yes {
			effectstest.DcCleanup(g)
			if effectstest.ControllerOf(t, g, three) != opp.ID {
				t.Error("control outlived the end of turn")
			}
		}
	}
}

func TestOldB589InfernalKirinDiscardsEveryCardWithThatManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Infernal Kirin", "Legendary Creature — Kirin Spirit", b589InfernalOracle, 3, 3)
	twoA := effectstest.PushHandCardWithManaCost(opp, "Two A", "Instant", "{1}{R}")
	twoB := effectstest.PushHandCardWithManaCost(opp, "Two B", "Creature — Elf", "{G}{G}")
	three := effectstest.PushHandCardWithManaCost(opp, "Three", "Instant", "{2}{R}")
	land := effectstest.PushHandCardOfType(opp, "Forest", "Basic Land — Forest")
	b589Cast(t, g, "An Arcane Spell", "Sorcery — Arcane", "{1}{B}")
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Contains(twoA) || opp.Hand.Contains(twoB) {
		t.Error("a card with mana value 2 stayed in the hand")
	}
	if !opp.Hand.Contains(three) || !opp.Hand.Contains(land) {
		t.Error("a card with another mana value was discarded")
	}
	if !opp.Graveyard.Contains(twoA) || !opp.Graveyard.Contains(twoB) {
		t.Error("the discarded cards are not in the graveyard")
	}
}

func TestOldB589FiddleheadKamiRegeneratesOnASpiritOrArcaneSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	kami := effectstest.B12Push(g, me.ID, "Fiddlehead Kami", "Creature — Spirit", b589FiddleheadOracle, 3, 3)
	b589Cast(t, g, "A Spirit", "Creature — Spirit", "{1}")
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(kami); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	c := effectstest.FindBattlefieldCardByID(g, kami)
	if c == nil || !c.Tapped {
		t.Fatal("the Kami did not regenerate")
	}
}

func TestOldB589FiddleheadKamiIsNotShieldedByABear(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	kami := effectstest.B12Push(g, me.ID, "Fiddlehead Kami", "Creature — Spirit", b589FiddleheadOracle, 3, 3)
	b589Cast(t, g, "Bear", "Creature — Bear", "{1}")
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(kami); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	if g.Battlefield.Contains(kami) {
		t.Fatal("a Bear spell regenerated the Kami")
	}
}

func TestOldB589CloudhoofKirinMillsThatManyCards(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.B12Push(g, me.ID, "Cloudhoof Kirin", "Legendary Creature — Kirin Spirit", b589CloudhoofOracle, 4, 4)
		opp.Library.Cards = nil
		effectstest.SeedLibrary(opp, "A", "B", "C", "D", "E", "F")
		b589Cast(t, g, "A Spirit", "Creature — Spirit", "{2}{U}")
		effectstest.PickPlayer(t, g, me.ID, opp.ID)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, yes)
		effectstest.PassPriorityAroundTable(t, g)
		want := 6
		if yes {
			want = 3
		}
		if opp.Library.Size() != want {
			t.Errorf("yes=%v: library holds %d, want %d", yes, opp.Library.Size(), want)
		}
	}
}

func TestOldB589DescendantOfSoramaroLooksAtAsManyCardsAsYourHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.B12Push(g, me.ID, "Descendant of Soramaro", "Creature — Human Wizard", b589SoramaroOracle, 2, 3)
	effectstest.AdvanceToMain(t, g)
	me.Hand.Cards = nil
	effectstest.FillHand(me, "H1", "H2")
	lib := effectstest.SeedLibrary(me, "A", "B", "C", "D")
	effectstest.C2Pay(t, g, me, "{C}{U}")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	ask := effectstest.PutInLibraryChoiceFor(g, me.ID)
	if ask == nil || len(ask.ScryCards) != 2 || ask.LibraryPlacement != game.LibraryPlaceTop {
		t.Fatalf("order prompt %+v, want the top two cards to put back on top", ask)
	}
	if err := g.ResolvePutInLibrary(ask.ID, me.ID, nil, []uuid.UUID{lib[1], lib[0]}); err != nil {
		t.Fatal(err)
	}
	if got := effectstest.LibraryTopNames(me, 3); got[0] != "B" || got[1] != "A" || got[2] != "C" {
		t.Errorf("top of library %v, want B, A, C", got)
	}
}

func TestOldB589WineOfBloodAndIronDoublesPowerThenSacrificesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	wine := effectstest.PushCatalogPermanent(g, me.ID, "Wine of Blood and Iron", "Artifact", b589WineOracle, false)
	bear := b589Creature(g, me.ID, "Bear", "{1}{G}", 3, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.C2Pay(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, wine, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err == nil {
		t.Fatal("activated for {3}")
	}
	effectstest.C2Pay(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, wine, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.EffectivePTOf(t, g, bear); p != 6 {
		t.Errorf("the Bear has power %d, want 6", p)
	}
	if !g.Battlefield.Contains(wine) {
		t.Fatal("the Wine was sacrificed at once")
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(wine) {
		t.Error("the Wine was not sacrificed at the beginning of the end step")
	}
}

func TestOldB589KitsuneBonesetterNeedsMoreCardsThanEachOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	healer := effectstest.PushCatalogPermanent(g, me.ID, "Kitsune Bonesetter", "Creature — Fox Cleric", b589BonesetterOracle, false)
	bear := b589Creature(g, me.ID, "Bear", "{1}{G}", 2, 5)
	effectstest.AdvanceToMain(t, g)
	me.Hand.Cards, opp.Hand.Cards = nil, nil
	for _, s := range g.Seats[1:] {
		s.Hand.Cards = nil
	}
	effectstest.FillHand(me, "A", "B")
	effectstest.FillHand(opp, "C", "D")
	params := game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}
	if err := g.ActivateCatalogAbility(me.ID, healer, 0, params); err == nil {
		t.Fatal("activated with as many cards as an opponent")
	}
	effectstest.FillHand(me, "E")
	if err := g.ActivateCatalogAbility(me.ID, healer, 0, params); err != nil {
		t.Fatalf("activate with more cards than each opponent: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(healer, bear, 4); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	if got := effectstest.DamageMarkedOn(g, bear); got != 1 {
		t.Errorf("the Bear took %d damage from 4, want 1 (3 prevented)", got)
	}
}

func TestOldB589OboroEnvoyReturnsALandAndShrinksByHandSize(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	envoy := effectstest.PushCatalogPermanent(g, me.ID, "Oboro Envoy", "Creature — Moonfolk Wizard", b589EnvoyOracle, false)
	land := effectstest.AangPushLand(g, me.ID, "Island", false)
	victim := b589Creature(g, opp.ID, "Ogre", "{3}{R}", 5, 4)
	effectstest.AdvanceToMain(t, g)
	me.Hand.Cards = nil
	effectstest.FillHand(me, "A", "B")
	effectstest.C2Pay(t, g, me, "{C}{C}")
	params := game.ActivateAbilityParams{Strict: true, ReturnIDs: []uuid.UUID{land},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}}
	if err := g.ActivateCatalogAbility(me.ID, envoy, 0, params); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) || !me.Hand.Contains(land) {
		t.Error("the land did not return to its owner's hand")
	}
	// Two cards in hand plus the returned land.
	if p, _ := effectstest.EffectivePTOf(t, g, victim); p != 2 {
		t.Errorf("the Ogre has power %d, want 2 (5 - 3 cards in hand)", p)
	}
	effectstest.DcCleanup(g)
	if p, _ := effectstest.EffectivePTOf(t, g, victim); p != 5 {
		t.Errorf("the shrink outlived the turn (%d)", p)
	}
}

func TestOldB589KagemaroIsAsBigAsYourHandAndSacrificesForAMinusX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	me.Hand.Cards = nil
	effectstest.FillHand(me, "A", "B", "C")
	kagemaro := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Kagemaro, First to Suffer", TypeLine: "Legendary Creature — Demon Spirit",
		OracleID: b589KagemaroOracle, Owner: me.ID, Controller: me.ID,
	})
	mine := b589Creature(g, me.ID, "Mine", "{1}", 2, 2)
	theirs := b589Creature(g, opp.ID, "Theirs", "{1}", 5, 5)
	if p, tough := effectstest.EffectivePTOf(t, g, kagemaro); p != 3 || tough != 3 {
		t.Errorf("Kagemaro is %d/%d with three cards in hand, want 3/3", p, tough)
	}
	// The count follows the hand live: a draw with nothing else touching the board.
	if err := g.DrawCard(me.ID); err != nil {
		t.Fatalf("DrawCard: %v", err)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, kagemaro); p != 4 || tough != 4 {
		t.Errorf("Kagemaro is %d/%d after a draw, want 4/4", p, tough)
	}
	effectstest.C2Pay(t, g, me, "{B}")
	if err := g.ActivateCatalogAbility(me.ID, kagemaro, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(kagemaro) {
		t.Error("Kagemaro was not sacrificed")
	}
	if g.Battlefield.Contains(mine) {
		t.Error("the 2/2 survived -4/-4")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, theirs); p != 1 || tough != 1 {
		t.Errorf("the 5/5 is %d/%d, want 1/1", p, tough)
	}
}
