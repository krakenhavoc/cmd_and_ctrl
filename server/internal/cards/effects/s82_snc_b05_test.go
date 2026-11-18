package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_snc_b05_test.go — the casualty spells of SNC / NCC, S82 batch 5.

const (
	snc5MakeDisappear   = "5f12f665-f083-48ae-b61a-541051c996ec"
	snc5ALittleChat     = "1167fe30-32b8-4381-8067-a05cd345977b"
	snc5LightEmUp       = "c6932bcf-d66b-42c1-8ca4-6a4573e0f6bf"
	snc5GrislySigil     = "b9d9d118-a073-4e73-90fe-026462d1d895"
	snc5DigUpTheBody    = "144e3d90-8ed8-4c21-941d-fc59cec042a2"
	snc5RobTheArchives  = "df45a344-5937-4a3f-a80f-581ea456ea05"
	snc5RooftopNuisance = "8843faa9-0cd5-4ef3-8e7b-b8833db9dc84"
	snc5JoinTheMaestros = "849a7ae8-2b4e-4fa5-9607-2bfc961fee40"
	snc5IllicitShipment = "81d62f0a-b2ba-4e87-9003-e5472bd1101f"
	snc5CutYourLosses   = "54530a84-8278-4c93-a2cb-f5c86619a053"
	snc5CutOfTheProfits = "8ce63f60-0211-46c7-8fb6-c5484ebfa029"
	snc5AudaciousSwap   = "40b0b93d-4075-42d6-bcd8-b2f9910cc7d2"
)

func TestSNC5SNCBatch5CardsAreRegisteredAndWhole(t *testing.T) {
	for _, id := range []string{snc5MakeDisappear, snc5ALittleChat, snc5LightEmUp, snc5GrislySigil, snc5DigUpTheBody,
		snc5RobTheArchives, snc5RooftopNuisance, snc5JoinTheMaestros, snc5IllicitShipment, snc5CutYourLosses,
		snc5CutOfTheProfits, snc5AudaciousSwap} {
		spec, ok := Lookup(id)
		if !ok {
			t.Fatalf("%s: not registered", id)
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: Completeness = %v, Caveats = %v", spec.Name, spec.Completeness, spec.Caveats)
		}
		if len(spec.OptionalCosts) != 1 || spec.OptionalCosts[0].Key != game.CasualtyKey {
			t.Errorf("%s: no casualty cost", spec.Name)
		}
	}
}

// snc5Hand puts a card in the active seat's hand at main phase.
func snc5Hand(t *testing.T, g *game.Game, name, typeLine, cost, oracle string) (*game.Player, uuid.UUID) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
			OracleID: oracle, Owner: me.ID, Controller: me.ID})
	})
	return me, id
}

// snc5Drive passes priority and answers every prompt until the stack is
// empty: pick_target with `target`, the first card of a choose_cards or a
// search, a cast offer with castOffers, and pay_unless with pay.
func snc5Drive(t *testing.T, g *game.Game, target game.TargetRef, castOffers, pay bool) {
	t.Helper()
	for i := 0; i < 80; i++ {
		var pc *game.PendingChoice
		for _, c := range g.PendingChoices {
			if c != nil {
				pc = c
				break
			}
		}
		if pc == nil {
			if stackFullyEmpty(g) {
				return
			}
			passPriorityAroundTable(t, g)
			continue
		}
		var err error
		switch pc.Kind {
		case game.PendingChoicePickTarget:
			err = g.ResolvePickTarget(pc.ID, pc.Chooser, target)
		case game.PendingChoiceChooseCards:
			err = g.ResolveChooseCards(pc.ID, pc.Chooser, pc.ChooseCards[:1])
		case game.PendingChoiceSearchLibrary:
			err = g.ResolveSearchLibrary(pc.ID, pc.Chooser, pc.SearchCards[:1])
		case game.PendingChoicePayUnless:
			err = g.ResolvePayUnless(pc.ID, pc.Chooser, pay)
		case game.PendingChoiceResolutionCast:
			if !castOffers {
				err = g.ResolvePendingChoice(pc.ID, pc.Chooser, nil)
			} else {
				card := keywordCastCard(pc)
				err = g.CastForResolutionChoice(pc.Chooser, pc.ID, card, game.CastSpellParams{FromZone: string(offeredZone(g, pc.Chooser, card))})
			}
		default:
			t.Fatalf("unhandled prompt %s", pc.Kind)
		}
		if err != nil {
			t.Fatalf("answering %s: %v", pc.Kind, err)
		}
	}
	t.Fatal("the stack never emptied")
}

func snc5Sac(g *game.Game, owner uuid.UUID, power int) uuid.UUID {
	return pushVanillaCreature(g, owner, "Fodder", power, 1)
}

func TestSNC5LightEmUpCasualtyCopiesAndRetargets(t *testing.T) {
	g := newCatalogGame(t)
	me, id := snc5Hand(t, g, "Light 'Em Up", "Sorcery", "{1}{R}", snc5LightEmUp)
	opp := g.Seats[1]
	victim := pushVanillaCreature(g, opp.ID, "Wall", 0, 4)
	fodder := snc5Sac(g, me.ID, 2)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Targets:       []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
		OptionalCosts: []int{0}, SacrificeIDs: []uuid.UUID{fodder},
	}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if g.Battlefield.Contains(fodder) {
		t.Fatal("casualty creature not sacrificed")
	}
	snc5Drive(t, g, game.TargetRef{Kind: game.TargetCard, ID: victim}, false, false)
	if g.Battlefield.Contains(victim) {
		t.Error("2 damage from the spell and 2 from its copy should kill a 0/4")
	}
}

func TestSNC5LightEmUpWithoutCasualtyDealsTwoOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, id := snc5Hand(t, g, "Light 'Em Up", "Sorcery", "{1}{R}", snc5LightEmUp)
	victim := pushVanillaCreature(g, g.Seats[1].ID, "Wall", 0, 4)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	snc5Drive(t, g, game.TargetRef{}, false, false)
	if !g.Battlefield.Contains(victim) {
		t.Error("a 0/4 died to one 2-damage spell")
	}
}

func TestSNC5CasualtyTwoRefusesAPowerOneCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, id := snc5Hand(t, g, "Light 'Em Up", "Sorcery", "{1}{R}", snc5LightEmUp)
	victim := pushVanillaCreature(g, g.Seats[1].ID, "Wall", 0, 4)
	small := snc5Sac(g, me.ID, 1)
	err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Targets:       []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
		OptionalCosts: []int{0}, SacrificeIDs: []uuid.UUID{small}, Strict: true,
	})
	if err == nil {
		t.Error("a power-1 creature paid casualty 2")
	}
}

func TestSNC5GrislySigil(t *testing.T) {
	for _, hurt := range []bool{false, true} {
		g := newCatalogGame(t)
		me, bolt := sqBoltInHand(t, g)
		victim := pushVanillaCreature(g, g.Seats[1].ID, "Wall", 0, 9)
		if hurt {
			if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}}); err != nil {
				t.Fatalf("bolt: %v", err)
			}
			snc5Drive(t, g, game.TargetRef{}, false, false)
		}
		id := uuid.New()
		g.WithWriteLock(func() {
			me.Hand.PushTop(game.Card{InstanceID: id, Name: "Grisly Sigil", TypeLine: "Sorcery", ManaCost: "{B}",
				OracleID: snc5GrislySigil, Owner: me.ID, Controller: me.ID})
		})
		before := me.Life
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}}); err != nil {
			t.Fatalf("sigil: %v", err)
		}
		snc5Drive(t, g, game.TargetRef{}, false, false)
		want := 1
		if hurt {
			want = 3
		}
		if me.Life-before != want {
			t.Errorf("hurt=%v: gained %d, want %d", hurt, me.Life-before, want)
		}
	}
}

func TestSNC5GrislySigilIgnoresCombatAndRecycledDamage(t *testing.T) {
	g := newCatalogGame(t)
	victim := pushVanillaCreature(g, g.Seats[1].ID, "Wall", 0, 9)
	g.WithWriteLock(func() {
		if g.PermanentWasDealtNoncombatDamageThisTurn(victim) {
			t.Error("undamaged creature reads as dealt noncombat damage")
		}
	})
}

func TestSNC5CutYourLossesMillsHalf(t *testing.T) {
	g := newCatalogGame(t)
	me, id := snc5Hand(t, g, "Cut Your Losses", "Sorcery", "{4}{U}{U}", snc5CutYourLosses)
	opp := g.Seats[1]
	before := len(opp.Library.Cards)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	snc5Drive(t, g, game.TargetRef{}, false, false)
	if got := before - len(opp.Library.Cards); got != before/2 {
		t.Errorf("milled %d, want %d", got, before/2)
	}
}

func TestSNC5CutOfTheProfitsDrawsAndLosesX(t *testing.T) {
	g := newCatalogGame(t)
	me, id := snc5Hand(t, g, "Cut of the Profits", "Sorcery", "{X}{B}{B}", snc5CutOfTheProfits)
	hand, life := len(me.Hand.Cards)-1, me.Life
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 3}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	snc5Drive(t, g, game.TargetRef{}, false, false)
	if len(me.Hand.Cards) != hand+3 || me.Life != life-3 {
		t.Errorf("hand %d (want %d), life %d (want %d)", len(me.Hand.Cards), hand+3, me.Life, life-3)
	}
}

func TestSNC5CutOfTheProfitsCasualtyCopyHasTheSameX(t *testing.T) {
	g := newCatalogGame(t)
	me, id := snc5Hand(t, g, "Cut of the Profits", "Sorcery", "{X}{B}{B}", snc5CutOfTheProfits)
	fodder := snc5Sac(g, me.ID, 3)
	life := me.Life
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 2, OptionalCosts: []int{0}, SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	snc5Drive(t, g, game.TargetRef{}, false, false)
	if me.Life != life-4 {
		t.Errorf("life %d, want %d (X=2 twice)", me.Life, life-4)
	}
}

func TestSNC5JoinTheMaestrosMakesAnOgreWarrior(t *testing.T) {
	g := newCatalogGame(t)
	me, id := snc5Hand(t, g, "Join the Maestros", "Sorcery", "{4}{B}", snc5JoinTheMaestros)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	snc5Drive(t, g, game.TargetRef{}, false, false)
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Name == "Ogre Warrior" && c.Controller == me.ID && c.Power == 4 && c.Toughness == 3 {
			n++
		}
	}
	if n != 1 {
		t.Errorf("%d Ogre Warrior tokens, want 1", n)
	}
}

func TestSNC5IllicitShipmentTutors(t *testing.T) {
	g := newCatalogGame(t)
	me, id := snc5Hand(t, g, "Illicit Shipment", "Sorcery", "{3}{B}{B}", snc5IllicitShipment)
	hand := len(me.Hand.Cards) - 1
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	snc5Drive(t, g, game.TargetRef{}, false, false)
	if len(me.Hand.Cards) != hand+1 {
		t.Errorf("hand %d, want %d", len(me.Hand.Cards), hand+1)
	}
}

func TestSNC5RooftopNuisanceTapsDrawsAndFreezes(t *testing.T) {
	g := newCatalogGame(t)
	me, id := snc5Hand(t, g, "Rooftop Nuisance", "Sorcery", "{2}{U}", snc5RooftopNuisance)
	victim := pushVanillaCreature(g, g.Seats[1].ID, "Wall", 0, 4)
	hand := len(me.Hand.Cards) - 1
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	snc5Drive(t, g, game.TargetRef{}, false, false)
	c, _ := g.LookupCardForEffect(victim)
	if !c.Tapped {
		t.Error("creature not tapped")
	}
	if len(me.Hand.Cards) != hand+1 {
		t.Errorf("hand %d, want %d", len(me.Hand.Cards), hand+1)
	}
}

func TestSNC5RobTheArchivesExilesTwoYouMayPlay(t *testing.T) {
	g := newCatalogGame(t)
	me, id := snc5Hand(t, g, "Rob the Archives", "Sorcery", "{1}{R}", snc5RobTheArchives)
	lib := len(me.Library.Cards)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	snc5Drive(t, g, game.TargetRef{}, false, false)
	if len(me.Library.Cards) != lib-2 {
		t.Errorf("library %d, want %d", len(me.Library.Cards), lib-2)
	}
}

func TestSNC5DigUpTheBodyMillsTwoAndReturnsACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, id := snc5Hand(t, g, "Dig Up the Body", "Instant", "{2}{B}", snc5DigUpTheBody)
	bear := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: bear, Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
			Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	snc5Drive(t, g, game.TargetRef{}, false, false)
	inHand := false
	for _, c := range me.Hand.Cards {
		if c.InstanceID == bear {
			inHand = true
		}
	}
	if !inHand {
		t.Error("the creature card did not return to hand")
	}
}

func TestSNC5ALittleChatTakesOneOfTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, id := snc5Hand(t, g, "A Little Chat", "Instant", "{1}{U}", snc5ALittleChat)
	hand, lib := len(me.Hand.Cards)-1, len(me.Library.Cards)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	snc5Drive(t, g, game.TargetRef{}, false, false)
	if len(me.Hand.Cards) != hand+1 || len(me.Library.Cards) != lib-1 {
		t.Errorf("hand %d (want %d), library %d (want %d)", len(me.Hand.Cards), hand+1, len(me.Library.Cards), lib-1)
	}
}

func TestSNC5MakeDisappearCountersUnlessTwoPaid(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	me, bolt := sqBoltInHand(t, g)
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("bolt: %v", err)
	}
	id := uuid.New()
	g.WithWriteLock(func() {
		opp.Hand.PushTop(game.Card{InstanceID: id, Name: "Make Disappear", TypeLine: "Instant", ManaCost: "{1}{U}",
			OracleID: snc5MakeDisappear, Owner: opp.ID, Controller: opp.ID})
	})
	before := opp.Life
	if err := g.CastSpell(opp.ID, id, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bolt}}}); err != nil {
		t.Fatalf("Make Disappear: %v", err)
	}
	snc5Drive(t, g, game.TargetRef{}, false, false)
	if opp.Life != before {
		t.Errorf("the Bolt resolved (life %d -> %d); it should have been countered", before, opp.Life)
	}
}

// snc5SwapSetup targets an opponent's permanent `target`, with the
// owner's library holding exactly `top`. The target is shuffled into that
// library, so which of the two is exiled is random: the tests below make
// both cards the same kind, so the outcome does not depend on the shuffle.
func snc5SwapSetup(t *testing.T, target, top game.Card) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := newCatalogGame(t)
	me, id := snc5Hand(t, g, "Audacious Swap", "Instant", "{3}{R}", snc5AudaciousSwap)
	opp := g.Seats[1]
	tok := uuid.New()
	g.WithWriteLock(func() {
		target.InstanceID = tok
		target.Owner, target.Controller = opp.ID, opp.ID
		g.Battlefield.PushTop(target)
		top.InstanceID = uuid.New()
		top.Owner, top.Controller = opp.ID, opp.ID
		opp.Library.Cards = nil
		opp.Library.PushTop(top)
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: tok}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	return g, me, opp
}

func TestSNC5AudaciousSwapLandEntersUnderTheOwner(t *testing.T) {
	g, _, opp := snc5SwapSetup(t,
		game.Card{Name: "Island", TypeLine: "Basic Land — Island"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	snc5Drive(t, g, game.TargetRef{}, false, false)
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsLand() && c.Controller == opp.ID {
			n++
		}
	}
	if n != 1 {
		t.Errorf("%d lands under the owner, want 1 (the other was shuffled into the library)", n)
	}
	if len(opp.Library.Cards) != 1 {
		t.Errorf("library %d, want 1", len(opp.Library.Cards))
	}
}

func TestSNC5AudaciousSwapNonlandIsOfferedToTheOwnerNotTheCaster(t *testing.T) {
	g, me, opp := snc5SwapSetup(t,
		game.Card{Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4},
		game.Card{Name: "Divination", TypeLine: "Sorcery", ManaCost: "{2}{U}"})
	for i := 0; i < 8 && latestChoiceOfKindFor(g, game.PendingChoiceResolutionCast, opp.ID) == nil; i++ {
		passPriorityAroundTable(t, g)
	}
	pc := latestChoiceOfKindFor(g, game.PendingChoiceResolutionCast, opp.ID)
	if pc == nil {
		t.Fatal("the owner was not offered the free cast")
	}
	if latestChoiceOfKindFor(g, game.PendingChoiceResolutionCast, me.ID) != nil {
		t.Error("the caster was offered the free cast")
	}
}
