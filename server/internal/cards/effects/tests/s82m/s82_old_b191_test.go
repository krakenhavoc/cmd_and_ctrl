package testss82m

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b191_test.go — Guilds of Ravnica, older-sets batch 191 (#5525).
// Every test plays the real catalog entry through the engine.

const (
	oldB191ConclaveGuildmage = "9d0eb1f4-08ef-4ac9-902b-47335ab46984"
	oldB191GrapplingSundew   = "a13e5a6c-e1aa-44dc-871e-120ea0390026"
	oldB191OmnispellAdept    = "b575449b-591d-4357-92b7-55137fc7966a"
	oldB191VividRevival      = "b5d856a3-0305-4c3b-b447-99570d6e8195"
	oldB191WandOfVertebrae   = "dbf7d799-98c9-4dc6-a58c-258bd9c19425"
	oldB191BloodOperative    = "ef75da10-1b33-4066-bd42-1111b5a94f75"
	oldB191SumalaWoodshaper  = "5a44062b-f1db-4122-8570-1241a16cb895"
	oldB191IntrusivePack     = "85a59833-b661-4599-a5d1-042dee033617"
	oldB191CandlelightVigil  = "7153bf15-53f8-445e-acbb-35b4f4b3d551"
	oldB191Demotion          = "c333eb79-300e-4dc0-9126-8dca1bac6ef2"
	oldB191VraskaRegalGorgon = "7616e834-4852-40c8-9de6-d89f96119f12"
	oldB191RalCallerOfStorms = "d7b10518-b6dc-4a39-881a-edd346034f28"
)

func TestOldB191EveryCardIsRegisteredFull(t *testing.T) {
	effectstest.NewCatalogGame(t)
	for _, id := range []string{oldB191ConclaveGuildmage, oldB191GrapplingSundew, oldB191OmnispellAdept, oldB191VividRevival,
		oldB191WandOfVertebrae, oldB191BloodOperative, oldB191SumalaWoodshaper, oldB191IntrusivePack, oldB191CandlelightVigil,
		oldB191Demotion, oldB191VraskaRegalGorgon, oldB191RalCallerOfStorms} {
		spec, ok := effects.Lookup(id)
		if !ok {
			t.Errorf("%s is not in the catalog", id)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v caveats %v", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB191ConclaveGuildmageGivesTrampleAndMakesAnElfKnight(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	mage := effectstest.B12Push(g, me.ID, "Conclave Guildmage", "Creature — Elf Cleric", oldB191ConclaveGuildmage, 2, 2)
	bear := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := effectstest.B12Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	effectstest.FloatForTest(g, me, "G")
	if err := g.ActivateCatalogAbility(me.ID, mage, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{mage, bear} {
		if !effectstest.HasEffectiveKeyword(t, g, id, "trample") {
			t.Error("one of my creatures lacks trample")
		}
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "trample") {
		t.Error("an opponent's creature gained trample")
	}
	if !effectstest.HmlTapped(t, g, mage) {
		t.Error("the Guildmage is not tapped")
	}
}

func TestOldB191ConclaveGuildmageMakesAnElfKnightToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	mage := effectstest.B12Push(g, me.ID, "Conclave Guildmage", "Creature — Elf Cleric", oldB191ConclaveGuildmage, 2, 2)
	effectstest.FloatForTest(g, me, "WWWWWW")
	if err := g.ActivateCatalogAbility(me.ID, mage, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	knights := effectstest.TokensNamed(g, "Elf Knight")
	if len(knights) != 1 || knights[0].Controller != me.ID {
		t.Fatalf("Elf Knight tokens: %d, want one under your control", len(knights))
	}
	id := knights[0].InstanceID
	if p, tough := effectstest.PtOf(t, g, id); p != 2 || tough != 2 || !effectstest.HasEffectiveKeyword(t, g, id, "vigilance") {
		t.Errorf("the token is %d/%d, want a 2/2 with vigilance", p, tough)
	}
	if !slices.Contains(knights[0].Colors, "G") || !slices.Contains(knights[0].Colors, "W") {
		t.Errorf("the token's colours are %v, want green and white", knights[0].Colors)
	}
}

func TestOldB191GrapplingSundewHasDefenderAndReachAndCanBecomeIndestructible(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	sundew := effectstest.B12Push(g, me.ID, "Grappling Sundew", "Creature — Plant", oldB191GrapplingSundew, 0, 4)
	for _, kw := range []string{"defender", "reach"} {
		if !effectstest.HasEffectiveKeyword(t, g, sundew, kw) {
			t.Errorf("the Sundew lacks %s", kw)
		}
	}
	if effectstest.HasEffectiveKeyword(t, g, sundew, "indestructible") {
		t.Fatal("indestructible before the ability")
	}
	effectstest.FloatForTest(g, me, "GGGGG")
	if err := g.ActivateCatalogAbility(me.ID, sundew, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, sundew, "indestructible") {
		t.Error("no indestructible after the ability")
	}
}

func TestOldB191OmnispellAdeptCastsAnInstantFromHandForFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	adept := effectstest.B12Push(g, me.ID, "Omnispell Adept", "Creature — Human Wizard", oldB191OmnispellAdept, 3, 4)
	bolt := effectstest.HandCardFull(me, "Lightning Bolt", "Instant", "{R}", effectstest.LightningBoltOracle, []string{"R"})
	bear := effectstest.HandCardFull(me, "Bear", "Creature — Bear", "{1}{G}", "", []string{"G"})
	effectstest.FloatForTest(g, me, "UUU")
	if err := g.ActivateCatalogAbility(me.ID, adept, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordHandCast, bolt)
	if slices.Contains(c.DiscardOptions, bear) {
		t.Fatal("a creature card is offered")
	}
	life := opp.Life
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{AlternativeCost: game.AltCostKeyGrantedCost, Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("cast free: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("the opponent is at %d life, want %d", opp.Life, life-3)
	}
}

func TestOldB191VividRevivalReturnsUpToThreeMulticoloredCardsAndExilesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	gold := func(name string, colors ...string) uuid.UUID {
		id := uuid.New()
		me.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Creature — Bear", Colors: colors, Owner: me.ID, Controller: me.ID})
		return id
	}
	a, b := gold("Gold A", "W", "U"), gold("Gold B", "B", "R", "G")
	mono := gold("Mono", "G")
	if err := effectstest.CastCatalogSpellErr(t, g, "Vivid Revival", "Sorcery", oldB191VividRevival, effectstest.CardTargets(mono)); err == nil {
		t.Fatal("a monocolored card was accepted")
	}
	vivid := effectstest.CastCatalogSpell(t, g, "Vivid Revival", "Sorcery", oldB191VividRevival, effectstest.CardTargets(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(a) || !me.Hand.Contains(b) {
		t.Error("the multicolored cards did not return to hand")
	}
	if !me.Graveyard.Contains(mono) {
		t.Error("the monocolored card moved")
	}
	if !g.Exile.Contains(vivid) {
		t.Error("Vivid Revival is not exiled")
	}
}

func TestOldB191WandOfVertebraeMillsAndShufflesCardsBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	wand := effectstest.PushCatalogPermanent(g, me.ID, "Wand of Vertebrae", "Artifact", oldB191WandOfVertebrae, false)
	effectstest.SeedLibrary(me, "A", "B", "C")
	grave := g.Seats[g.Turn.ActiveSeat].Graveyard.Size()
	if err := g.ActivateCatalogAbility(me.ID, wand, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("mill: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Size() != grave+1 {
		t.Errorf("graveyard %d, want %d", me.Graveyard.Size(), grave+1)
	}
	// The second ability needs the Wand untapped again.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == wand {
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
	var ids []uuid.UUID
	for i := 0; i < 3; i++ {
		ids = append(ids, effectstest.PushGraveyardCardTyped(me, "Old Card", "Instant"))
	}
	lib := me.Library.Size()
	effectstest.FloatForTest(g, me, "CC")
	if err := g.ActivateCatalogAbility(me.ID, wand, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTargets(ids...)}); err != nil {
		t.Fatalf("shuffle: %v", err)
	}
	if g.Battlefield.Contains(wand) {
		t.Error("the exile is a cost, paid on activation")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Library.Size() != lib+3 {
		t.Errorf("library %d, want %d", me.Library.Size(), lib+3)
	}
	for _, id := range ids {
		if me.Graveyard.Contains(id) {
			t.Error("a target stayed in the graveyard")
		}
	}
}

func TestOldB191BloodOperativeExilesAGraveyardCardWhenItEnters(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		dead := effectstest.PushGraveyardCardTyped(opp, "Dead Bear", "Creature — Bear")
		if _, err := effectstest.TrekB37CastFromHand(t, g, game.Card{Name: "Blood Operative", TypeLine: "Creature — Vampire Assassin",
			OracleID: oldB191BloodOperative, Power: 3, Toughness: 1}, 0); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.Cmdr2B07PickIfAsked(t, g, me.ID, dead)
		effectstest.PassPriorityAroundTable(t, g)
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceTriggerPrompt && c.Chooser == me.ID {
				if err := g.ResolveTriggerPrompt(c.ID, me.ID, yes); err != nil {
					t.Fatalf("ResolveTriggerPrompt: %v", err)
				}
				break
			}
		}
		effectstest.PassPriorityAroundTable(t, g)
		if opp.Graveyard.Contains(dead) == yes {
			t.Errorf("yes=%v: the card is in the graveyard = %v", yes, opp.Graveyard.Contains(dead))
		}
	}
}

func oldB191Surveil(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E", "F")
	effectstest.CastCatalogSpell(t, g, "Consider", "Instant", effectstest.ConsiderOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SurveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("no surveil prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveSurveil(c.ID, me.ID, nil, c.ScryCards); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
}

func TestOldB191BloodOperativeComesBackFromTheGraveyardWhenYouSurveilAndPayThreeLife(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		id := uuid.New()
		me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Blood Operative", TypeLine: "Creature — Vampire Assassin",
			OracleID: oldB191BloodOperative, Power: 3, Toughness: 1, Owner: me.ID, Controller: me.ID})
		life := me.Life
		oldB191Surveil(t, g, me)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		if pay {
			if !me.Hand.Contains(id) || me.Graveyard.Contains(id) {
				t.Error("Blood Operative did not return to hand")
			}
			if me.Life != life-3 {
				t.Errorf("life %d, want %d", me.Life, life-3)
			}
		} else if !me.Graveyard.Contains(id) || me.Life != life {
			t.Error("declining should leave the card in the graveyard and the life untouched")
		}
	}
}

func TestOldB191BloodOperativeDoesNothingOnTheBattlefieldOrForAnOpponentsSurveil(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Blood Operative", "Creature — Vampire Assassin", oldB191BloodOperative, 3, 1)
	oldB191Surveil(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("the graveyard trigger fired from the battlefield")
	}
}

func TestOldB191SumalaWoodshaperTakesACreatureOrEnchantmentAndBottomsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() { me.Library.Cards = nil })
	deep := effectstest.SeedLibraryTop(me, "Deep Bear", "Creature — Bear", "{G}")
	art := effectstest.SeedLibraryTop(me, "Rock", "Artifact", "{1}")
	land := effectstest.SeedLibraryTop(me, "Forest", "Basic Land — Forest", "")
	ench := effectstest.SeedLibraryTop(me, "Ench", "Enchantment", "{W}")
	bear := effectstest.SeedLibraryTop(me, "Bear", "Creature — Bear", "{G}")
	effectstest.CastCatalogSpell(t, g, "Sumala Woodshaper", "Creature — Elf Druid", oldB191SumalaWoodshaper, nil)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestChooseCardsFor(g, me.ID)
	if p == nil {
		t.Fatal("no choose_cards prompt")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range p.ChooseCards {
		offered[id] = true
	}
	if offered[land] || offered[art] || offered[deep] {
		t.Errorf("a land, an artifact or the fifth card was offered: %v", p.ChooseCards)
	}
	if !offered[ench] || !offered[bear] {
		t.Errorf("the creature and enchantment were not both offered: %v", p.ChooseCards)
	}
	hand := me.Hand.Size()
	if err := g.ResolveChooseCards(p.ID, me.ID, []uuid.UUID{ench}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(ench) || me.Hand.Size() != hand+1 {
		t.Error("the enchantment is not in hand")
	}
	n := len(me.Library.Cards)
	if me.Library.Cards[n-1].InstanceID != deep {
		t.Errorf("the fifth card is not on top: %s", me.Library.Cards[n-1].Name)
	}
	bottom := map[uuid.UUID]bool{}
	for _, c := range me.Library.Cards[:3] {
		bottom[c.InstanceID] = true
	}
	for _, id := range []uuid.UUID{bear, art, land} {
		if !bottom[id] {
			t.Error("a card that was not taken is not on the bottom")
		}
	}
}

func TestOldB191SumalaWoodshaperMayTakeNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() { me.Library.Cards = nil })
	bear := effectstest.SeedLibraryTop(me, "Bear", "Creature — Bear", "{G}")
	effectstest.CastCatalogSpell(t, g, "Sumala Woodshaper", "Creature — Elf Druid", oldB191SumalaWoodshaper, nil)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestChooseCardsFor(g, me.ID)
	if p == nil {
		t.Fatal("no choose_cards prompt")
	}
	if err := g.ResolveChooseCards(p.ID, me.ID, nil); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if me.Hand.Contains(bear) || !me.Library.Contains(bear) {
		t.Error("declining took the card")
	}
}

func TestOldB191IntrusivePackbeastTapsUpToTwoOpposingCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Push(g, me.ID, "Mine", "Creature — Bear", "", 2, 2)
	a := effectstest.B12Push(g, opp.ID, "A", "Creature — Bear", "", 2, 2)
	b := effectstest.B12Push(g, opp.ID, "B", "Creature — Bear", "", 2, 2)
	c := effectstest.B12Push(g, opp.ID, "C", "Creature — Bear", "", 2, 2)
	if _, err := effectstest.TrekB37CastFromHand(t, g, game.Card{Name: "Intrusive Packbeast", TypeLine: "Creature — Beast",
		OracleID: oldB191IntrusivePack, Power: 3, Toughness: 3}, 0); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerTriggerTargets(t, g, me.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HmlTapped(t, g, a) || !effectstest.HmlTapped(t, g, b) {
		t.Error("the chosen creatures are not tapped")
	}
	if effectstest.HmlTapped(t, g, c) || effectstest.HmlTapped(t, g, mine) {
		t.Error("an unchosen or own creature was tapped")
	}
}

func TestOldB191CandlelightVigilPumpsAndGrantsVigilance(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.AuraBear(g, me.ID)
	effectstest.AuraCast(t, g, "Candlelight Vigil", oldB191CandlelightVigil, bear)
	if p, tough := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 5 || tough != 4 {
		t.Errorf("the enchanted Bear is %d/%d, want 5/4", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "vigilance") {
		t.Error("the enchanted creature lacks vigilance")
	}
}

func TestOldB191DemotionStopsBlockingAndEveryActivatedAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	victim := effectstest.AuraBear(g, opp.ID)
	other := effectstest.AuraBear(g, opp.ID)
	effectstest.AuraCast(t, g, "Demotion", oldB191Demotion, victim)
	r := effectstest.AuraRestrictions(t, g, victim)
	for name, bit := range map[string]game.Restriction{"block": game.CantBlock, "activate": game.CantActivate, "activate mana abilities": game.CantActivateMana} {
		if r&bit == 0 {
			t.Errorf("the enchanted creature can still %s", name)
		}
	}
	if r&game.CantAttack != 0 {
		t.Error("Demotion stops attacking, which it does not print")
	}
	if effectstest.AuraRestrictions(t, g, other)&(game.CantBlock|game.CantActivate) != 0 {
		t.Error("an unenchanted creature is restricted")
	}
	_ = me
}

func TestOldB191VraskaPlusTwoPutsACounterAndMenaceOnTheTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	vraska := effectstest.S82Walker(g, me.ID, "Vraska, Regal Gorgon", oldB191VraskaRegalGorgon, 5)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, vraska, 0, game.ActivateAbilityParams{Targets: effectstest.S82TargetCard(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("%d counters, want 1", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "menace") || effectstest.HasEffectiveKeyword(t, g, other, "menace") {
		t.Error("only the target should have menace")
	}
	if got := effectstest.LoyaltyOf(g, vraska); got != 7 {
		t.Errorf("loyalty %d, want 7", got)
	}
}

func TestOldB191VraskaMinusThreeDestroysTheTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vraska := effectstest.S82Walker(g, me.ID, "Vraska, Regal Gorgon", oldB191VraskaRegalGorgon, 5)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, vraska, 1, game.ActivateAbilityParams{Targets: effectstest.S82TargetCard(theirs)})
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, theirs) {
		t.Error("the creature was not destroyed")
	}
	if got := effectstest.LoyaltyOf(g, vraska); got != 2 {
		t.Errorf("loyalty %d, want 2", got)
	}
}

func TestOldB191VraskaPlusTwoWithNoTargetStillAddsLoyalty(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	vraska := effectstest.S82Walker(g, me.ID, "Vraska, Regal Gorgon", oldB191VraskaRegalGorgon, 5)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, vraska, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LoyaltyOf(g, vraska); got != 7 {
		t.Errorf("loyalty %d, want 7", got)
	}
}

func TestOldB191VraskaUltimateCountsCreatureCardsInYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vraska := effectstest.S82Walker(g, me.ID, "Vraska, Regal Gorgon", oldB191VraskaRegalGorgon, 10)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	for i := 0; i < 3; i++ {
		effectstest.PushGraveyardCardTyped(me, "Dead", "Creature — Bear")
	}
	effectstest.PushGraveyardCardTyped(me, "Spell", "Instant")
	effectstest.PushGraveyardCardTyped(opp, "Their Dead", "Creature — Bear")
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, vraska, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if got := effectstest.CounterCount(g, id, game.CounterPlusOne); got != 3 {
			t.Errorf("%d counters on my creature, want 3", got)
		}
	}
	if got := effectstest.CounterCount(g, theirs, game.CounterPlusOne); got != 0 {
		t.Errorf("%d counters on an opponent's creature, want 0", got)
	}
}

func TestOldB191RalPlusOneDrawsACard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ral := effectstest.S82Walker(g, me.ID, "Ral, Caller of Storms", oldB191RalCallerOfStorms, 4)
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	effectstest.S82Activate(t, g, me.ID, ral, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
	if got := effectstest.LoyaltyOf(g, ral); got != 5 {
		t.Errorf("loyalty %d, want 5", got)
	}
}

func TestOldB191RalMinusTwoDividesThreeDamageAmongTheTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ral := effectstest.S82Walker(g, me.ID, "Ral, Caller of Storms", oldB191RalCallerOfStorms, 4)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Ogre", 4, 4)
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 1, 1)
	effectstest.AdvanceToMain(t, g)
	life := opp.Life
	targets := []game.TargetRef{{Kind: game.TargetCard, ID: small}, {Kind: game.TargetCard, ID: big}, {Kind: game.TargetPlayer, ID: opp.ID}}
	if err := g.ActivateCatalogAbility(me.ID, ral, 1, game.ActivateAbilityParams{Strict: true, Targets: targets,
		Distribution: map[uuid.UUID]int{small: 2, big: 1, opp.ID: 1}}); err == nil {
		t.Fatal("a division of 4 was accepted")
	}
	if err := g.ActivateCatalogAbility(me.ID, ral, 1, game.ActivateAbilityParams{Strict: true, Targets: targets,
		Distribution: map[uuid.UUID]int{small: 1, big: 1, opp.ID: 1}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, small) {
		t.Error("the 1/1 survived 1 damage")
	}
	if !effects.OnBattlefield(g, big) {
		t.Error("the 4/4 died to 1 damage")
	}
	if opp.Life != life-1 {
		t.Errorf("the player is at %d life, want %d", opp.Life, life-1)
	}
	if got := effectstest.LoyaltyOf(g, ral); got != 2 {
		t.Errorf("loyalty %d, want 2", got)
	}
}

func TestOldB191RalUltimateDrawsSevenAndHitsEveryOpposingCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ral := effectstest.S82Walker(g, me.ID, "Ral, Caller of Storms", oldB191RalCallerOfStorms, 7)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Ogre", 7, 8)
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Ogre", 7, 7)
	b := effectstest.B12Creature(g, g.Seats[2].ID, "B", "Creature — Bear", 2, 2)
	survivor := effectstest.B12Creature(g, opp.ID, "Tall", "Creature — Wall", 0, 8)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { me.Library.Cards = nil })
	effectstest.SeedLibrary(me, "1", "2", "3", "4", "5", "6", "7", "8")
	hand := me.Hand.Size()
	effectstest.S82Activate(t, g, me.ID, ral, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+7 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+7)
	}
	if effects.OnBattlefield(g, a) || effects.OnBattlefield(g, b) {
		t.Error("an opposing creature survived 7 damage")
	}
	if !effects.OnBattlefield(g, survivor) || !effects.OnBattlefield(g, mine) {
		t.Error("a creature with toughness 8, or my own creature, died")
	}
	if got := effectstest.LoyaltyOf(g, ral); got != 0 && effects.OnBattlefield(g, ral) {
		t.Errorf("Ral has %d loyalty and is still on the battlefield", got)
	}
}
