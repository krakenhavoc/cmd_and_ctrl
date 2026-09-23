package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// aang_g4_test.go — the "Aang is so flashy" deck's flicker / ETB /
// bounce group (#1306): the cards whose caveats closed and the cards
// that were missing.

// --- Deputy of Acquittals ----------------------------------------

// "Another target creature you control" excludes THIS Deputy by
// instance and nothing else: a second, same-named Deputy is offered
// (and returned when picked), the entering one is not.
func TestDeputyOfAcquittalsOffersAnotherCreatureButNeverItself(t *testing.T) {
	g := newCatalogGame(t)
	p0, p1 := g.Seats[0], g.Seats[1]
	bear := pushCatalogPermanent(g, p0.ID, "Bears", "Creature — Bear", "", false)
	otherDeputy := pushCatalogPermanent(g, p0.ID, "Deputy of Acquittals", "Creature — Human Wizard", deputyOfAcquittalsOracle, false)
	theirs := pushCatalogPermanent(g, p1.ID, "Their Bear", "Creature — Bear", "", false)

	deputy := castAndResolveCreature(t, g, "Deputy of Acquittals", "Creature — Human Wizard", deputyOfAcquittalsOracle)
	answerLatestTriggerPrompt(t, g, p0.ID, true)

	pick := latestPickTarget(g, p0.ID)
	if pick == nil {
		t.Fatal("no pick_target prompt after answering yes")
	}
	if hasID(pick.PickTargetCards, deputy) {
		t.Error("the entering Deputy is offered as its own \"another\" target")
	}
	if !hasID(pick.PickTargetCards, otherDeputy) {
		t.Error("a second Deputy of Acquittals is not offered — \"another\" is not \"not named Deputy\"")
	}
	if !hasID(pick.PickTargetCards, bear) {
		t.Error("an ordinary creature you control is not offered")
	}
	if hasID(pick.PickTargetCards, theirs) {
		t.Error("an opponent's creature is offered to \"a creature you control\"")
	}

	pickCard(t, g, p0.ID, otherDeputy)
	passPriorityAroundTable(t, g)
	if !p0.Hand.Contains(otherDeputy) {
		t.Error("the picked Deputy did not return to its owner's hand")
	}
	if !onBattlefield(g, deputy) {
		t.Error("the entering Deputy left the battlefield")
	}
	if spec, _ := Lookup(deputyOfAcquittalsOracle); spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Deputy of Acquittals is %v with caveats %v, want full", spec.Completeness, spec.Caveats)
	}
}

// --- Sink into Stupor // Soporific Springs -----------------------

const sinkIntoStuporOracle = "bcc6eece-75ea-494c-b33a-d4477d504e0b"

// sinkIntoStupor is the MH3 modal DFC, imported the way the deck
// importer builds it: an instant front and a land back.
func sinkIntoStupor(owner uuid.UUID) game.Card {
	return mdfcCard(owner, sinkIntoStuporOracle, game.LayoutModalDFC,
		game.Face{Name: "Sink into Stupor", TypeLine: "Instant", ManaCost: "{1}{U}{U}", Colors: []string{"U"}},
		game.Face{Name: "Soporific Springs", TypeLine: "Land"},
	)
}

// The front face is registered and resolves on both halves of its
// clause: an opponent's spell is returned from the stack, an
// opponent's nonland permanent is bounced. The back face's land spec
// is untouched.
func TestSinkIntoStuporReturnsAnOpponentsSpellOrNonlandPermanent(t *testing.T) {
	for _, half := range []string{"spell", "permanent"} {
		t.Run(half, func(t *testing.T) {
			g, me, id := handWithMDFC(t, sinkIntoStupor)
			opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			var target uuid.UUID
			if half == "spell" {
				target = b02dCastInstantAs(t, g, opp, "Their Instant", nil)
			} else {
				target = pushCatalogPermanent(g, opp.ID, "Their Rock", "Artifact", "", false)
			}
			if err := g.CastSpell(me.ID, id, game.CastSpellParams{
				Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
			}); err != nil {
				t.Fatalf("cast Sink into Stupor at their %s: %v", half, err)
			}
			passPriorityAroundTable(t, g)
			if !opp.Hand.Contains(target) {
				t.Errorf("their %s did not return to its owner's hand", half)
			}
		})
	}
	if spec, ok := Lookup(sinkIntoStuporOracle); !ok || spec.Name != "Sink into Stupor" || spec.Completeness != CompletenessFull {
		t.Errorf("front face spec = %+v, want a full Sink into Stupor", spec)
	}
	if back, ok := Lookup(sinkIntoStuporOracle + "#1"); !ok || back.Name != "Soporific Springs" {
		t.Errorf("back face spec = %+v, want Soporific Springs", back)
	}
}

// "An opponent controls" governs both halves: your own spell, your own
// permanent and an opponent's LAND are all refused at announce.
func TestSinkIntoStuporRefusesYourOwnThingsAndLands(t *testing.T) {
	for _, what := range []string{"my spell", "my permanent", "their land"} {
		t.Run(what, func(t *testing.T) {
			g, me, id := handWithMDFC(t, sinkIntoStupor)
			opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			var target uuid.UUID
			switch what {
			case "my spell":
				target = b02dCastInstantAs(t, g, me, "My Instant", nil)
			case "my permanent":
				target = pushCatalogPermanent(g, me.ID, "My Rock", "Artifact", "", false)
			default:
				target = aangPushLand(g, opp.ID, "Island", false)
			}
			err := g.CastSpell(me.ID, id, game.CastSpellParams{
				Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
			})
			if !errors.Is(err, game.ErrIllegalTarget) {
				t.Fatalf("Sink into Stupor at %s: err = %v, want ErrIllegalTarget", what, err)
			}
		})
	}
}

// --- Venser, Shaper Savant ---------------------------------------

// The spell half: flashed in over an opponent's instant, Venser's ETB
// offers that spell alongside the permanents, and picking it returns
// the spell to its owner's hand without it resolving — and without
// countering it.
func TestVenserReturnsATargetSpellToItsOwnersHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rock := pushCatalogPermanent(g, opp.ID, "Their Rock", "Artifact", "", false)

	toMainForCost(t, g)
	theirSpell := b02dCastInstantAs(t, g, opp, "Their Instant", nil)
	b18CastFromHand(t, g, me, "Venser, Shaper Savant", "Legendary Creature — Human Wizard", b18VenserShaperSavantOracle, game.CastSpellParams{})
	b04WaitForPick(t, g, me.ID)

	pick := latestPickTarget(g, me.ID)
	if !hasID(pick.PickTargetCards, theirSpell) {
		t.Fatal("the spell on the stack is not offered to \"target spell or permanent\"")
	}
	if !hasID(pick.PickTargetCards, rock) {
		t.Error("a permanent is not offered to \"target spell or permanent\"")
	}
	counteredBefore := countEvents(g, game.EventCounterSpell)
	pickCard(t, g, me.ID, theirSpell)
	passPriorityAroundTable(t, g)

	if !opp.Hand.Contains(theirSpell) {
		t.Error("the targeted spell did not return to its owner's hand")
	}
	if g.Stack.Contains(theirSpell) {
		t.Error("the targeted spell is still on the stack")
	}
	if n := countEvents(g, game.EventCounterSpell); n != counteredBefore {
		t.Errorf("returning the spell countered it (%d counter events)", n-counteredBefore)
	}
	if !onBattlefield(g, rock) {
		t.Error("the untargeted permanent moved")
	}
	if spec, _ := Lookup(b18VenserShaperSavantOracle); spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Venser is %v with caveats %v, want full", spec.Completeness, spec.Caveats)
	}
}

// --- Ambrosia Whiteheart -----------------------------------------

const ambrosiaWhiteheartOracle = "2bcc9f11-5b12-433e-9680-0f4b18aa521c"

// The ETB is an untargeted choice on resolution among the controller's
// OTHER permanents — lands included, Ambrosia and opponents' permanents
// not — with zero as a legal answer; landfall pumps her +1/+0.
func TestAmbrosiaWhiteheartReturnsAnotherPermanentAndGrowsOnLandfall(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pushCatalogPermanent(g, me.ID, "Bears", "Creature — Bear", "", false)
	land := aangPushLand(g, me.ID, "Plains", false)
	theirs := pushCatalogPermanent(g, opp.ID, "Their Rock", "Artifact", "", false)

	ambrosia := castAndResolveCreature(t, g, "Ambrosia Whiteheart", "Legendary Creature — Bird", ambrosiaWhiteheartOracle)
	passPriorityAroundTable(t, g)

	pick := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if pick == nil {
		t.Fatalf("no resolution-time permanent choice: %+v", g.PendingChoices)
	}
	if pick.ChooseMin != 0 || pick.ChooseMax != 1 {
		t.Errorf("bounds %d..%d, want 0..1 (\"you may return another permanent\")", pick.ChooseMin, pick.ChooseMax)
	}
	if hasID(pick.ChooseCards, ambrosia) {
		t.Error("Ambrosia is offered as \"another\" permanent")
	}
	if !hasID(pick.ChooseCards, bear) || !hasID(pick.ChooseCards, land) {
		t.Error("a creature and a land you control are both \"another permanent you control\"")
	}
	if hasID(pick.ChooseCards, theirs) {
		t.Error("an opponent's permanent is offered")
	}
	if err := g.ResolveOwnPermanents(pick.ID, me.ID, []uuid.UUID{bear}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) {
		t.Error("the chosen permanent did not return to its owner's hand")
	}
	if !onBattlefield(g, land) || !onBattlefield(g, ambrosia) {
		t.Error("an unchosen permanent moved")
	}

	before := effectivePower(t, g, ambrosia)
	b13PlayAs(t, g, g.Turn.ActiveSeat, "Plains", "Basic Land — Plains", "")
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, ambrosia); got != before+1 {
		t.Errorf("Ambrosia's power after a land entered = %d, want %d", got, before+1)
	}
}

// --- The Mighty Thor, Jane Foster --------------------------------

const mightyThorOracle = "57d02dc8-e22e-4874-9f02-490a2528a28f"

// The attack trigger flickers one nontoken artifact or creature and
// returns it TAPPED under its OWNER's control: a creature stolen from
// an opponent goes home, tapped, as a new object. Tokens are not
// offered.
func TestMightyThorAttackFlickersTappedUnderItsOwnersControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	thor := pushDiesCreatureForTest(g, me.ID, "The Mighty Thor, Jane Foster", mightyThorOracle, "Legendary Creature — Human God Hero", 3, 3)
	stolen := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Stolen Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: me.ID,
	})
	token := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Soldier", TypeLine: "Token Creature — Soldier",
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID,
	})

	declareAttack(t, g, opp.ID, thor)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatalf("no pick_target prompt for the attack trigger: %+v", g.PendingChoices)
	}
	if hasID(pick.PickTargetCards, token) {
		t.Error("a token is offered to \"nontoken artifact or creature\"")
	}
	if !hasID(pick.PickTargetCards, stolen) || !hasID(pick.PickTargetCards, thor) {
		t.Error("a nontoken creature (and Thor herself) must be offered")
	}
	pickCard(t, g, me.ID, stolen)
	passPriorityAroundTable(t, g)

	if onBattlefield(g, stolen) {
		t.Fatal("the flickered creature is still the same object (CR 400.7)")
	}
	back := findBattlefieldByName(g, "Stolen Bear")
	if back == uuid.Nil {
		t.Fatal("the flickered creature did not return to the battlefield")
	}
	c, _ := aangCardOnBF(g, back)
	if c.Controller != opp.ID {
		t.Errorf("returned under %v's control, want its owner %v", c.Controller, opp.ID)
	}
	if !c.Tapped {
		t.Error("the flickered creature returned untapped")
	}
}

// "Whenever an Equipment you control enters, draw a card" — yours
// draws, an opponent's does not.
func TestMightyThorDrawsWhenYourEquipmentEnters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushDiesCreatureForTest(g, me.ID, "The Mighty Thor, Jane Foster", mightyThorOracle, "Legendary Creature — Human God Hero", 3, 3)

	toMainForCost(t, g)
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Bonesplitter", "Artifact — Equipment", "", nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d → %d after your Equipment entered, want +1", hand, got)
	}

	hand = me.Hand.Size()
	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Sword", TypeLine: "Artifact — Equipment",
		Owner: opp.ID, Controller: opp.ID,
	})
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, CardID: theirs, Actor: opp.ID})
	})
	if len(g.PendingTriggers) != 0 {
		t.Errorf("an opponent's Equipment entering queued %d trigger(s)", len(g.PendingTriggers))
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d → %d after an opponent's Equipment entered, want unchanged", hand, got)
	}
}

// --- Thassa, Deep-Dwelling ---------------------------------------

const thassaDeepDwellingOracle = "2396299a-c031-4020-b13e-1f9bf9d64511"

func pushThassa(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Thassa, Deep-Dwelling", OracleID: thassaDeepDwellingOracle,
		TypeLine: "Legendary Enchantment Creature — God", ManaCost: "{3}{U}",
		Power: 6, Toughness: 5, Owner: owner, Controller: owner,
	})
}

// The God clause: not a creature below five blue devotion (her own {U}
// counts), a creature at five.
func TestThassaIsACreatureOnlyAtFiveBlueDevotion(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	thassa := pushThassa(g, me.ID)
	if hasString(effectiveTypes(t, g, thassa), "Creature") {
		t.Fatal("Thassa is a creature at devotion 1")
	}
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Blue Pips", TypeLine: "Enchantment", ManaCost: "{U}{U}{U}",
		Owner: me.ID, Controller: me.ID,
	})
	if hasString(effectiveTypes(t, g, thassa), "Creature") {
		t.Fatal("Thassa is a creature at devotion 4")
	}
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "One More", TypeLine: "Artifact", ManaCost: "{U}",
		Owner: me.ID, Controller: me.ID,
	})
	if !hasString(effectiveTypes(t, g, thassa), "Creature") {
		t.Error("Thassa is not a creature at devotion 5")
	}
}

// The end-step blink offers up to one OTHER creature you control and
// returns it under YOUR control as a new object — a creature you had
// stolen stays yours. Thassa (a creature here) is never offered.
func TestThassaBlinksAnotherCreatureAtYourEndStepUnderYourControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	thassa := pushThassa(g, me.ID)
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Blue Pips", TypeLine: "Enchantment", ManaCost: "{U}{U}{U}{U}",
		Owner: me.ID, Controller: me.ID,
	})
	stolen := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Stolen Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: me.ID,
	})

	advanceToEndStepOf(t, g, 0)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatalf("no pick_target prompt at your end step: %+v", g.PendingChoices)
	}
	if hasID(pick.PickTargetCards, thassa) {
		t.Error("Thassa is offered to her own \"other target creature\"")
	}
	if !hasID(pick.PickTargetCards, stolen) {
		t.Fatal("a creature you control is not offered")
	}
	pickCard(t, g, me.ID, stolen)
	passPriorityAroundTable(t, g)

	back := findBattlefieldByName(g, "Stolen Bear")
	if back == uuid.Nil || back == stolen {
		t.Fatalf("blinked creature back=%v (old %v): want a new object on the battlefield", back, stolen)
	}
	if c, _ := aangCardOnBF(g, back); c.Controller != me.ID {
		t.Errorf("returned under %v's control, want yours (%v)", c.Controller, me.ID)
	}
}

// {3}{U}: tap another target creature — an opponent's creature taps,
// Thassa herself is refused.
func TestThassaTapsAnotherTargetCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	thassa := pushThassa(g, me.ID)
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Blue Pips", TypeLine: "Enchantment", ManaCost: "{U}{U}{U}{U}",
		Owner: me.ID, Controller: me.ID,
	})
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	toMainForCost(t, g)

	if err := g.ActivateCatalogAbility(me.ID, thassa, 0, game.ActivateAbilityParams{Targets: cardRefs(thassa)}); !errors.Is(err, game.ErrIllegalTarget) {
		t.Errorf("Thassa targeting herself: err = %v, want ErrIllegalTarget", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, thassa, 0, game.ActivateAbilityParams{Targets: cardRefs(theirs)}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c, _ := aangCardOnBF(g, theirs); !c.Tapped {
		t.Error("the target creature did not become tapped")
	}
}

// "You may": choosing nothing returns nothing.
func TestAmbrosiaWhiteheartMayReturnNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushCatalogPermanent(g, me.ID, "Bears", "Creature — Bear", "", false)
	castAndResolveCreature(t, g, "Ambrosia Whiteheart", "Legendary Creature — Bird", ambrosiaWhiteheartOracle)
	passPriorityAroundTable(t, g)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if pick == nil {
		t.Fatal("no resolution-time permanent choice")
	}
	if err := g.ResolveOwnPermanents(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("declining: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !onBattlefield(g, bear) {
		t.Error("declining still returned a permanent")
	}
}
