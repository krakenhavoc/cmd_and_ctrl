package testso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// old_b814_cards_test.go — S82 OLD batch 814, Alliances card draw and
// library (#5634).

const (
	b814Sage      = "1f612df3-53b6-4317-9d63-1f903ee3f0c4"
	b814Gargoyle  = "e8cc9f83-4e4c-4820-9a7a-f82f728e9767"
	b814Lord      = "283e7a23-ae25-48d1-abb5-cd021f66cc29"
	b814WarCry    = "60a7f61c-41ad-4dce-b556-a99bf38c3d73"
	b814Bauble    = "6048b9ac-7d5e-4486-a89a-30adb45aab27"
	b814Sentry    = "d2a61132-d881-4c80-92f1-a151d835832a"
	b814Astrolabe = "d19567ac-1f52-40eb-b2cf-4c791823ec92"
	b814Heretic   = "eed353a6-757d-47f6-a118-b951ccf95600"
	b814Heritance = "742cb752-dd20-4b1f-8d96-7e9a8f864e75"
	b814Foresight = "c9640b04-9751-46d5-b7bf-5b8a8ecf145f"
	b814LatNam    = "f885ee54-3b26-4d61-9783-3930b0e625aa"
	b814Suffocate = "57aeb106-dc4f-4069-b596-f997263dc0cf"
	b814Scarab    = "5f085af7-ce66-4019-9516-d1650ee7b6e0"
)

func TestB814CardsAreRegisteredFull(t *testing.T) {
	for id, name := range map[string]string{
		b814Sage: "Soldevi Sage", b814Gargoyle: "Ivory Gargoyle", b814Lord: "Lord of Tresserhorn",
		b814WarCry: "Gorilla War Cry", b814Bauble: "Lodestone Bauble", b814Sentry: "Soldevi Sentry",
		b814Astrolabe: "Astrolabe", b814Heretic: "Soldevi Heretic", b814Heritance: "Inheritance",
		b814Foresight: "Foresight", b814LatNam: "Lat-Nam's Legacy", b814Suffocate: "Suffocation",
		b814Scarab: "Scarab of the Unseen",
	} {
		spec, ok := effects.Lookup(id)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func b814Pool(p *game.Player, color string) int {
	n := 0
	for _, m := range p.ManaPool {
		if m.Color == color {
			n += max(m.Count, 1)
		}
	}
	return n
}

func b814Hand(p *game.Player) int { return p.Hand.Size() }

// b814ToNextUpkeep walks to the next upkeep at the table and resolves what it
// triggers.
func b814ToNextUpkeep(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 400; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
		if g.Turn.Step == game.StepUpkeep {
			break
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func b814Cast(t *testing.T, g *game.Game, me *game.Player, name, typeLine, oracle string) uuid.UUID {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func TestB814SoldeviSageDrawsThreeThenDiscardsOneOfThose(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AwbPadLibrary(me)
	sage := effectstest.B12Push(g, me.ID, "Soldevi Sage", "Creature — Human Wizard", b814Sage, 1, 1)
	l1 := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	l2 := effectstest.B12Permanent(g, me.ID, "Island", "Basic Land — Island")
	old := effectstest.B02TopOfLibrary(me, "Old card", "Sorcery")
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneLibrary, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, old); err != nil {
		t.Fatal(err)
	}
	before := b814Hand(me)
	if err := g.ActivateCatalogAbility(me.ID, sage, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{l1, l2}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(l1) || g.Battlefield.Contains(l2) {
		t.Error("the two lands were not sacrificed")
	}
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil || len(pick.ChooseCards) != 3 {
		t.Fatalf("pick = %+v, want a pick among the three drawn cards", pick)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{old}); err == nil {
		t.Fatal("a card that was already in hand could be discarded")
	}
	chosen := pick.ChooseCards[1]
	effectstest.AnswerChooseCards(t, g, me.ID, chosen)
	if got := b814Hand(me); got != before+2 {
		t.Errorf("hand %d, want %d (three drawn, one discarded)", got, before+2)
	}
	if !me.Graveyard.Contains(chosen) {
		t.Error("the chosen card is not in the graveyard")
	}
}

func TestB814SoldeviSageNeedsTwoLandsToSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sage := effectstest.B12Push(g, me.ID, "Soldevi Sage", "Creature — Human Wizard", b814Sage, 1, 1)
	l1 := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	if err := g.ActivateCatalogAbility(me.ID, sage, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{l1}}); err == nil {
		t.Error("the Sage activated with one land to sacrifice")
	}
}

func TestB814IvoryGargoyleReturnsAtTheNextEndStepAndSkipsYourNextDrawStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AwbPadLibrary(me)
	id := effectstest.B12Push(g, me.ID, "Ivory Gargoyle", "Creature — Gargoyle", b814Gargoyle, 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.B14Kill(g, id)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(id) || !me.Graveyard.Contains(id) {
		t.Fatalf("the Gargoyle did not die")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("the Gargoyle did not return at the end step")
	}
	handBefore := b814Hand(me)
	// My next turn: its draw step is skipped, so I reach its main phase with no new card.
	for i := 0; i < 400; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
		if g.Turn.ActiveSeat == 0 && g.Turn.Step == game.StepPrecombatMain {
			break
		}
	}
	if got := b814Hand(me); got != handBefore {
		t.Errorf("hand %d in my next main phase, want %d (the draw step was skipped)", got, handBefore)
	}
	// The turn after that, I draw again.
	for i := 0; i < 400; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
		if g.Turn.ActiveSeat == 0 && g.Turn.Step == game.StepPrecombatMain {
			break
		}
	}
	if got := b814Hand(me); got != handBefore+1 {
		t.Errorf("hand %d a turn later, want %d (only one draw step is skipped)", got, handBefore+1)
	}
}

func TestB814IvoryGargoyleExilesItselfForFivePlusAWhite(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.B12Push(g, me.ID, "Ivory Gargoyle", "Creature — Gargoyle", b814Gargoyle, 2, 2)
	effectstest.ApaMana(me, "W", "W", "W", "W", "W")
	effectstest.P7Activate(t, g, me, id, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(id) || me.Graveyard.Contains(id) {
		t.Error("the Gargoyle did not go to exile")
	}
}

func TestB814LordOfTresserhornCostsLifeCreaturesAndGivesTheOpponentCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AwbPadLibrary(opp)
	bear1 := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	bear2 := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	bear3 := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	life, oppHand := me.Life, b814Hand(opp)
	lord := effectstest.EnterFromHand(t, g, me.ID, "Lord of Tresserhorn", "Legendary Creature — Zombie", b814Lord)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{effectstest.PlayerTarget(opp.ID)}); err != nil {
			t.Fatalf("pick the opponent: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-2 {
		t.Errorf("life %d, want %d", me.Life, life-2)
	}
	effectstest.AnswerPermanentPicks(t, g, me.ID, map[uuid.UUID][]uuid.UUID{me.ID: {bear1, bear2}})
	if g.Battlefield.Contains(bear1) || g.Battlefield.Contains(bear2) || !g.Battlefield.Contains(bear3) || !g.Battlefield.Contains(lord) {
		t.Errorf("sacrifice: bear1 %v bear2 %v bear3 %v lord %v", g.Battlefield.Contains(bear1), g.Battlefield.Contains(bear2), g.Battlefield.Contains(bear3), g.Battlefield.Contains(lord))
	}
	if got := b814Hand(opp); got != oppHand+2 {
		t.Errorf("the opponent's hand %d, want %d", got, oppHand+2)
	}
}

func TestB814GorillaWarCryGivesEveryCreatureMenaceAndDrawsAtTheNextUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AwbPadLibrary(me)
	atk := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Gorilla War Cry", TypeLine: "Instant", OracleID: b814WarCry, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast at the beginning of combat: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range []uuid.UUID{atk, theirs} {
		lc := effectstest.LayeredCard(t, g, c)
		if !game.HasKeyword(&lc, "menace") {
			t.Errorf("%s has no menace", lc.Name)
		}
	}
	before := b814Hand(me)
	b814ToNextUpkeep(t, g)
	if got := b814Hand(me); got != before+1 {
		t.Errorf("hand %d, want %d after the delayed draw", got, before+1)
	}
}

func TestB814GorillaWarCryCannotBeCastOutsideCombatBeforeBlockers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Gorilla War Cry", TypeLine: "Instant", OracleID: b814WarCry, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("cast in the main phase")
	}
}

func TestB814LodestoneBaubleStacksBasicLandsOnTopAndTheOwnerDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AwbPadLibrary(opp)
	bauble := effectstest.PushCatalogPermanent(g, me.ID, "Lodestone Bauble", "Artifact", b814Bauble, false)
	var lands []uuid.UUID
	for i := 0; i < 2; i++ {
		id := uuid.New()
		lands = append(lands, id)
		g.WithWriteLock(func() {
			opp.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: opp.ID, Controller: opp.ID})
		})
	}
	effectstest.ApaMana(me, "G")
	targets := []game.TargetRef{{Kind: game.TargetCard, ID: lands[0]}, {Kind: game.TargetCard, ID: lands[1]}}
	if err := g.ActivateCatalogAbility(me.ID, bauble, 0, game.ActivateAbilityParams{Targets: targets}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	order := effectstest.PutInLibraryChoiceFor(g, me.ID)
	if order == nil {
		t.Fatalf("no ordering prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, nil, []uuid.UUID{lands[1], lands[0]}); err != nil {
		t.Fatalf("order the lands: %v", err)
	}
	if g.Battlefield.Contains(bauble) {
		t.Error("the Bauble was not sacrificed")
	}
	if opp.Graveyard.Contains(lands[0]) || opp.Graveyard.Contains(lands[1]) {
		t.Error("the lands stayed in the graveyard")
	}
	top, _ := opp.Library.Top()
	if top.InstanceID != lands[0] && top.InstanceID != lands[1] {
		t.Errorf("the top of their library is %q, want a Forest", top.Name)
	}
	before := b814Hand(opp)
	b814ToNextUpkeep(t, g)
	if got := b814Hand(opp); got != before+1 {
		t.Errorf("the graveyard's owner has %d cards, want %d after the draw", got, before+1)
	}
}

func TestB814LodestoneBaubleRefusesCardsFromTwoGraveyards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bauble := effectstest.PushCatalogPermanent(g, me.ID, "Lodestone Bauble", "Artifact", b814Bauble, false)
	a, b := uuid.New(), uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: a, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
		opp.Graveyard.PushTop(game.Card{InstanceID: b, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: opp.ID, Controller: opp.ID})
	})
	effectstest.ApaMana(me, "G")
	targets := []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}
	if err := g.ActivateCatalogAbility(me.ID, bauble, 0, game.ActivateAbilityParams{Strict: true, Targets: targets}); err == nil {
		t.Error("targeted lands from two different graveyards")
	}
}

func TestB814AstrolabeMakesTwoManaOfOneColorAndDrawsAtTheNextUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AwbPadLibrary(me)
	id := effectstest.PushCatalogPermanent(g, me.ID, "Astrolabe", "Artifact", b814Astrolabe, false)
	effectstest.ApaMana(me, "G")
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.B10ResolveAllManaPicks(t, g, me.ID, "U")
	if b814Pool(me, "U") != 2 {
		t.Errorf("pool %v, want two {U}", me.ManaPool)
	}
	if g.Battlefield.Contains(id) {
		t.Error("Astrolabe was not sacrificed")
	}
	before := b814Hand(me)
	b814ToNextUpkeep(t, g)
	if got := b814Hand(me); got != before+1 {
		t.Errorf("hand %d, want %d", got, before+1)
	}
}

func TestB814InheritanceDrawsOnlyWhenYouPay(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.AwbPadLibrary(me)
		effectstest.PushCatalogPermanent(g, me.ID, "Inheritance", "Enchantment", b814Heritance, false)
		bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
		effectstest.B10AddMana(me, "W", "W", "W")
		before := b814Hand(me)
		effectstest.B14Kill(g, bear)
		g.RunStateChecksForTest()
		effectstest.PassPriorityAroundTable(t, g)
		c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePayUnless, me.ID)
		if c == nil {
			t.Fatalf("no payment prompt: %+v", g.PendingChoices)
		}
		if err := g.ResolvePayUnless(c.ID, me.ID, pay); err != nil {
			t.Fatal(err)
		}
		want := before
		if pay {
			want++
		}
		if got := b814Hand(me); got != want {
			t.Errorf("pay=%v: hand %d, want %d", pay, got, want)
		}
	}
}

func TestB814ForesightExilesThreeCardsAndDrawsAtTheNextUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AwbPadLibrary(me)
	var picks []uuid.UUID
	for _, n := range []string{"A", "B", "C"} {
		picks = append(picks, effectstest.B02TopOfLibrary(me, "Pick "+n, "Sorcery"))
	}
	size := me.Library.Size()
	b814Cast(t, g, me, "Foresight", "Sorcery", b814Foresight)
	effectstest.AnswerSearchByID(t, g, me.ID, picks...)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Library.Size(); got != size-3 {
		t.Errorf("library %d, want %d", got, size-3)
	}
	for _, id := range picks {
		if !g.Exile.Contains(id) {
			t.Errorf("a searched card is not in exile")
		}
	}
	before := b814Hand(me)
	b814ToNextUpkeep(t, g)
	if got := b814Hand(me); got != before+1 {
		t.Errorf("hand %d, want %d", got, before+1)
	}
}

func TestB814LatNamsLegacyShufflesAHandCardAwayForTwoCardsLater(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AwbPadLibrary(me)
	keep := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: keep, Name: "Put Away", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	size := me.Library.Size()
	b814Cast(t, g, me, "Lat-Nam's Legacy", "Instant", b814LatNam)
	effectstest.AnswerChooseCards(t, g, me.ID, keep)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(keep) || me.Library.Size() != size+1 {
		t.Fatalf("the card did not reach the library (library %d, want %d)", me.Library.Size(), size+1)
	}
	before := b814Hand(me)
	b814ToNextUpkeep(t, g)
	if got := b814Hand(me); got != before+2 {
		t.Errorf("hand %d, want %d (two cards at the next upkeep)", got, before+2)
	}
}

func TestB814LatNamsLegacyWithAnEmptyHandDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AwbPadLibrary(me)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	b814Cast(t, g, me, "Lat-Nam's Legacy", "Instant", b814LatNam)
	before := b814Hand(me)
	b814ToNextUpkeep(t, g)
	if got := b814Hand(me); got != before {
		t.Errorf("hand %d, want %d (no card was shuffled away)", got, before)
	}
}

func TestB814SoldeviSentryRegeneratesAndTheNamedOpponentMayDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AwbPadLibrary(opp)
	id := effectstest.B12Push(g, me.ID, "Soldevi Sentry", "Artifact Creature — Soldier", b814Sentry, 1, 1)
	effectstest.ApaMana(me, "C")
	targets := []game.TargetRef{effectstest.PlayerTarget(opp.ID)}
	effectstest.P7Activate(t, g, me, id, 0, game.ActivateAbilityParams{Targets: targets})
	before := b814Hand(opp)
	effectstest.B14Kill(g, id)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("the Sentry did not regenerate")
	}
	effectstest.AnswerMayChoice(t, g, opp.ID, true)
	if got := b814Hand(opp); got != before+1 {
		t.Errorf("the named opponent's hand %d, want %d", got, before+1)
	}
}

func TestB814SoldeviSentryDeclinedDrawsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AwbPadLibrary(opp)
	id := effectstest.B12Push(g, me.ID, "Soldevi Sentry", "Artifact Creature — Soldier", b814Sentry, 1, 1)
	effectstest.ApaMana(me, "C")
	effectstest.P7Activate(t, g, me, id, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}})
	before := b814Hand(opp)
	effectstest.B14Kill(g, id)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, opp.ID, false)
	if got := b814Hand(opp); got != before {
		t.Errorf("hand %d, want %d", got, before)
	}
}

func TestB814SoldeviHereticPreventsTwoAndLetsTheOpponentDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AwbPadLibrary(opp)
	heretic := effectstest.B12Push(g, me.ID, "Soldevi Heretic", "Creature — Human Cleric", b814Heretic, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 3, 3)
	effectstest.ApaMana(me, "W")
	before := b814Hand(opp)
	targets := []game.TargetRef{{Kind: game.TargetCard, ID: bear}, effectstest.PlayerTarget(opp.ID)}
	effectstest.P7Activate(t, g, me, heretic, 0, game.ActivateAbilityParams{Targets: targets})
	effectstest.AnswerMayChoice(t, g, opp.ID, true)
	if got := b814Hand(opp); got != before+1 {
		t.Errorf("the opponent's hand %d, want %d", got, before+1)
	}
	g.WithWriteLock(func() {
		_ = g.DealDamageToCreatureForEffect(uuid.Nil, bear, 3)
	})
	g.RunStateChecksForTest()
	if c := effectstest.ApaLive(g, bear); c == nil || c.DamageMarked != 1 {
		t.Errorf("the Bear took %v damage, want 1 (3 less the 2 prevented)", c)
	}
}

func TestB814ScarabOfTheUnseenReturnsEveryAuraOnTheTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AwbPadLibrary(me)
	scarab := effectstest.PushCatalogPermanent(g, me.ID, "Scarab of the Unseen", "Artifact", b814Scarab, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	mine := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "My Aura", TypeLine: "Enchantment — Aura"})
	theirs := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Their Aura", TypeLine: "Enchantment — Aura"})
	for _, a := range []uuid.UUID{mine, theirs} {
		g.WithWriteLock(func() {
			if err := g.AttachForEffect(a, game.TargetRef{Kind: game.TargetCard, ID: bear}); err != nil {
				t.Fatal(err)
			}
		})
	}
	effectstest.P7Activate(t, g, me, scarab, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
	if !me.Hand.Contains(mine) || !opp.Hand.Contains(theirs) {
		t.Error("the Auras did not return to their owners' hands")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the Bear left")
	}
	before := b814Hand(me)
	b814ToNextUpkeep(t, g)
	if got := b814Hand(me); got != before+1 {
		t.Errorf("hand %d, want %d", got, before+1)
	}
}

func TestB814ScarabOfTheUnseenCannotTargetAPermanentYouOnlyControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	scarab := effectstest.PushCatalogPermanent(g, me.ID, "Scarab of the Unseen", "Artifact", b814Scarab, false)
	stolen := effectstest.ApaPush(g, opp.ID, me.ID, game.Card{Name: "Stolen Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	err := g.ActivateCatalogAbility(me.ID, scarab, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: stolen}}})
	if err == nil {
		t.Error("targeted a permanent the controller does not own")
	}
}

func TestB814SuffocationNeedsRedSpellDamageAndHitsThatSpellsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AwbPadLibrary(me)
	effectstest.AdvanceToMain(t, g)
	// Without damage from a red instant or sorcery it cannot be cast.
	early := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: early, Name: "Suffocation", TypeLine: "Instant", OracleID: b814Suffocate, Colors: []string{"U"}, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, early, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast with no damage dealt to me this turn")
	}
	// Their red sorcery deals 3 to me.
	bolt := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: bolt, Name: "Burn", TypeLine: "Sorcery", Colors: []string{"R"}, Owner: opp.ID, Controller: opp.ID})
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: opp.ID}, game.ZoneRef{Kind: game.ZoneStack}, bolt); err != nil {
		t.Fatalf("to the stack: %v", err)
	}
	g.WithWriteLock(func() {
		if g.StackMeta == nil {
			g.StackMeta = map[uuid.UUID]*game.StackItem{}
		}
		g.StackMeta[bolt] = &game.StackItem{ID: bolt, Kind: game.StackItemSpell, SourceCardID: bolt, Controller: opp.ID}
		_ = g.DealDamageToPlayerForEffect(bolt, me.ID, 3)
	})
	oppLife := opp.Life
	if err := g.CastSpell(me.ID, early, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast after a red sorcery hit me: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-4 {
		t.Errorf("their life %d, want %d", opp.Life, oppLife-4)
	}
	before := b814Hand(me)
	b814ToNextUpkeep(t, g)
	if got := b814Hand(me); got != before+1 {
		t.Errorf("hand %d, want %d", got, before+1)
	}
}
