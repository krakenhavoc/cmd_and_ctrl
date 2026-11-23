package testsa

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
	g := effectstest.NewCatalogGame(t)
	p0, p1 := g.Seats[0], g.Seats[1]
	bear := effectstest.PushCatalogPermanent(g, p0.ID, "Bears", "Creature — Bear", "", false)
	otherDeputy := effectstest.PushCatalogPermanent(g, p0.ID, "Deputy of Acquittals", "Creature — Human Wizard", effectstest.DeputyOfAcquittalsOracle, false)
	theirs := effectstest.PushCatalogPermanent(g, p1.ID, "Their Bear", "Creature — Bear", "", false)

	deputy := effectstest.CastAndResolveCreature(t, g, "Deputy of Acquittals", "Creature — Human Wizard", effectstest.DeputyOfAcquittalsOracle)

	pick := effectstest.LatestPickTarget(g, p0.ID)
	if pick == nil {
		t.Fatal("no pick_target prompt after answering yes")
	}
	if effectstest.HasID(pick.PickTargetCards, deputy) {
		t.Error("the entering Deputy is offered as its own \"another\" target")
	}
	if !effectstest.HasID(pick.PickTargetCards, otherDeputy) {
		t.Error("a second Deputy of Acquittals is not offered — \"another\" is not \"not named Deputy\"")
	}
	if !effectstest.HasID(pick.PickTargetCards, bear) {
		t.Error("an ordinary creature you control is not offered")
	}
	if effectstest.HasID(pick.PickTargetCards, theirs) {
		t.Error("an opponent's creature is offered to \"a creature you control\"")
	}

	effectstest.PickCard(t, g, p0.ID, otherDeputy)
	effectstest.AnswerLatestTriggerPrompt(t, g, p0.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !p0.Hand.Contains(otherDeputy) {
		t.Error("the picked Deputy did not return to its owner's hand")
	}
	if !effects.OnBattlefield(g, deputy) {
		t.Error("the entering Deputy left the battlefield")
	}
	if spec, _ := effects.Lookup(effectstest.DeputyOfAcquittalsOracle); spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Deputy of Acquittals is %v with caveats %v, want full", spec.Completeness, spec.Caveats)
	}
}

// --- Sink into Stupor // Soporific Springs -----------------------

const sinkIntoStuporOracle = "bcc6eece-75ea-494c-b33a-d4477d504e0b"

// sinkIntoStupor is the MH3 modal DFC, imported the way the deck
// importer builds it: an instant front and a land back.
func sinkIntoStupor(owner uuid.UUID) game.Card {
	return effectstest.MdfcCard(owner, sinkIntoStuporOracle, game.LayoutModalDFC,
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
			g, me, id := effectstest.HandWithMDFC(t, sinkIntoStupor)
			opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			var target uuid.UUID
			if half == "spell" {
				target = effectstest.B02dCastInstantAs(t, g, opp, "Their Instant", nil)
			} else {
				target = effectstest.PushCatalogPermanent(g, opp.ID, "Their Rock", "Artifact", "", false)
			}
			if err := g.CastSpell(me.ID, id, game.CastSpellParams{
				Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
			}); err != nil {
				t.Fatalf("cast Sink into Stupor at their %s: %v", half, err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if !opp.Hand.Contains(target) {
				t.Errorf("their %s did not return to its owner's hand", half)
			}
		})
	}
	if spec, ok := effects.Lookup(sinkIntoStuporOracle); !ok || spec.Name != "Sink into Stupor" || spec.Completeness != effects.CompletenessFull {
		t.Errorf("front face spec = %+v, want a full Sink into Stupor", spec)
	}
	if back, ok := effects.Lookup(sinkIntoStuporOracle + "#1"); !ok || back.Name != "Soporific Springs" {
		t.Errorf("back face spec = %+v, want Soporific Springs", back)
	}
}

// "An opponent controls" governs both halves: your own spell, your own
// permanent and an opponent's LAND are all refused at announce.
func TestSinkIntoStuporRefusesYourOwnThingsAndLands(t *testing.T) {
	for _, what := range []string{"my spell", "my permanent", "their land"} {
		t.Run(what, func(t *testing.T) {
			g, me, id := effectstest.HandWithMDFC(t, sinkIntoStupor)
			opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			var target uuid.UUID
			switch what {
			case "my spell":
				target = effectstest.B02dCastInstantAs(t, g, me, "My Instant", nil)
			case "my permanent":
				target = effectstest.PushCatalogPermanent(g, me.ID, "My Rock", "Artifact", "", false)
			default:
				target = effectstest.AangPushLand(g, opp.ID, "Island", false)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rock := effectstest.PushCatalogPermanent(g, opp.ID, "Their Rock", "Artifact", "", false)

	effectstest.ToMainForCost(t, g)
	theirSpell := effectstest.B02dCastInstantAs(t, g, opp, "Their Instant", nil)
	effectstest.B18CastFromHand(t, g, me, "Venser, Shaper Savant", "Legendary Creature — Human Wizard", effectstest.B18VenserShaperSavantOracle, game.CastSpellParams{})
	effectstest.B04WaitForPick(t, g, me.ID)

	pick := effectstest.LatestPickTarget(g, me.ID)
	if !effectstest.HasID(pick.PickTargetCards, theirSpell) {
		t.Fatal("the spell on the stack is not offered to \"target spell or permanent\"")
	}
	if !effectstest.HasID(pick.PickTargetCards, rock) {
		t.Error("a permanent is not offered to \"target spell or permanent\"")
	}
	counteredBefore := effectstest.CountEvents(g, game.EventCounterSpell)
	effectstest.PickCard(t, g, me.ID, theirSpell)
	effectstest.PassPriorityAroundTable(t, g)

	if !opp.Hand.Contains(theirSpell) {
		t.Error("the targeted spell did not return to its owner's hand")
	}
	if g.Stack.Contains(theirSpell) {
		t.Error("the targeted spell is still on the stack")
	}
	if n := effectstest.CountEvents(g, game.EventCounterSpell); n != counteredBefore {
		t.Errorf("returning the spell countered it (%d counter events)", n-counteredBefore)
	}
	if !effects.OnBattlefield(g, rock) {
		t.Error("the untargeted permanent moved")
	}
	if spec, _ := effects.Lookup(effectstest.B18VenserShaperSavantOracle); spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Venser is %v with caveats %v, want full", spec.Completeness, spec.Caveats)
	}
}

// --- Ambrosia Whiteheart -----------------------------------------

const ambrosiaWhiteheartOracle = "2bcc9f11-5b12-433e-9680-0f4b18aa521c"

// The ETB is an untargeted choice on resolution among the controller's
// OTHER permanents — lands included, Ambrosia and opponents' permanents
// not — with zero as a legal answer; landfall pumps her +1/+0.
func TestAmbrosiaWhiteheartReturnsAnotherPermanentAndGrowsOnLandfall(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.PushCatalogPermanent(g, me.ID, "Bears", "Creature — Bear", "", false)
	land := effectstest.AangPushLand(g, me.ID, "Plains", false)
	theirs := effectstest.PushCatalogPermanent(g, opp.ID, "Their Rock", "Artifact", "", false)

	ambrosia := effectstest.CastAndResolveCreature(t, g, "Ambrosia Whiteheart", "Legendary Creature — Bird", ambrosiaWhiteheartOracle)
	effectstest.PassPriorityAroundTable(t, g)

	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if pick == nil {
		t.Fatalf("no resolution-time permanent choice: %+v", g.PendingChoices)
	}
	if pick.ChooseMin != 0 || pick.ChooseMax != 1 {
		t.Errorf("bounds %d..%d, want 0..1 (\"you may return another permanent\")", pick.ChooseMin, pick.ChooseMax)
	}
	if effectstest.HasID(pick.ChooseCards, ambrosia) {
		t.Error("Ambrosia is offered as \"another\" permanent")
	}
	if !effectstest.HasID(pick.ChooseCards, bear) || !effectstest.HasID(pick.ChooseCards, land) {
		t.Error("a creature and a land you control are both \"another permanent you control\"")
	}
	if effectstest.HasID(pick.ChooseCards, theirs) {
		t.Error("an opponent's permanent is offered")
	}
	if err := g.ResolveOwnPermanents(pick.ID, me.ID, []uuid.UUID{bear}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) {
		t.Error("the chosen permanent did not return to its owner's hand")
	}
	if !effects.OnBattlefield(g, land) || !effects.OnBattlefield(g, ambrosia) {
		t.Error("an unchosen permanent moved")
	}

	before := effectstest.EffectivePower(t, g, ambrosia)
	effectstest.B13PlayAs(t, g, g.Turn.ActiveSeat, "Plains", "Basic Land — Plains", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, ambrosia); got != before+1 {
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	thor := effectstest.PushDiesCreatureForTest(g, me.ID, "The Mighty Thor, Jane Foster", mightyThorOracle, "Legendary Creature — Human God Hero", 3, 3)
	stolen := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Stolen Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: me.ID,
	})
	token := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Soldier", TypeLine: "Token Creature — Soldier",
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID,
	})

	effectstest.DeclareAttack(t, g, opp.ID, thor)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatalf("no pick_target prompt for the attack trigger: %+v", g.PendingChoices)
	}
	if effectstest.HasID(pick.PickTargetCards, token) {
		t.Error("a token is offered to \"nontoken artifact or creature\"")
	}
	if !effectstest.HasID(pick.PickTargetCards, stolen) || !effectstest.HasID(pick.PickTargetCards, thor) {
		t.Error("a nontoken creature (and Thor herself) must be offered")
	}
	effectstest.PickCard(t, g, me.ID, stolen)
	effectstest.PassPriorityAroundTable(t, g)

	if effects.OnBattlefield(g, stolen) {
		t.Fatal("the flickered creature is still the same object (CR 400.7)")
	}
	back := effectstest.FindBattlefieldByName(g, "Stolen Bear")
	if back == uuid.Nil {
		t.Fatal("the flickered creature did not return to the battlefield")
	}
	c, _ := effectstest.AangCardOnBF(g, back)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushDiesCreatureForTest(g, me.ID, "The Mighty Thor, Jane Foster", mightyThorOracle, "Legendary Creature — Human God Hero", 3, 3)

	effectstest.ToMainForCost(t, g)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Bonesplitter", "Artifact — Equipment", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d → %d after your Equipment entered, want +1", hand, got)
	}

	hand = me.Hand.Size()
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Sword", TypeLine: "Artifact — Equipment",
		Owner: opp.ID, Controller: opp.ID,
	})
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, CardID: theirs, Actor: opp.ID})
	})
	if len(g.PendingTriggers) != 0 {
		t.Errorf("an opponent's Equipment entering queued %d trigger(s)", len(g.PendingTriggers))
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d → %d after an opponent's Equipment entered, want unchanged", hand, got)
	}
}

// --- Thassa, Deep-Dwelling ---------------------------------------

const thassaDeepDwellingOracle = "2396299a-c031-4020-b13e-1f9bf9d64511"

func pushThassa(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Thassa, Deep-Dwelling", OracleID: thassaDeepDwellingOracle,
		TypeLine: "Legendary Enchantment Creature — God", ManaCost: "{3}{U}",
		Power: 6, Toughness: 5, Owner: owner, Controller: owner,
	})
}

// The God clause: not a creature below five blue devotion (her own {U}
// counts), a creature at five.
func TestThassaIsACreatureOnlyAtFiveBlueDevotion(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	thassa := pushThassa(g, me.ID)
	if effectstest.HasString(effectstest.EffectiveTypes(t, g, thassa), "Creature") {
		t.Fatal("Thassa is a creature at devotion 1")
	}
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Blue Pips", TypeLine: "Enchantment", ManaCost: "{U}{U}{U}",
		Owner: me.ID, Controller: me.ID,
	})
	if effectstest.HasString(effectstest.EffectiveTypes(t, g, thassa), "Creature") {
		t.Fatal("Thassa is a creature at devotion 4")
	}
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "One More", TypeLine: "Artifact", ManaCost: "{U}",
		Owner: me.ID, Controller: me.ID,
	})
	if !effectstest.HasString(effectstest.EffectiveTypes(t, g, thassa), "Creature") {
		t.Error("Thassa is not a creature at devotion 5")
	}
}

// The end-step blink offers up to one OTHER creature you control and
// returns it under YOUR control as a new object — a creature you had
// stolen stays yours. Thassa (a creature here) is never offered.
func TestThassaBlinksAnotherCreatureAtYourEndStepUnderYourControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	thassa := pushThassa(g, me.ID)
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Blue Pips", TypeLine: "Enchantment", ManaCost: "{U}{U}{U}{U}",
		Owner: me.ID, Controller: me.ID,
	})
	stolen := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Stolen Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: me.ID,
	})

	effectstest.AdvanceToEndStepOf(t, g, 0)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatalf("no pick_target prompt at your end step: %+v", g.PendingChoices)
	}
	if effectstest.HasID(pick.PickTargetCards, thassa) {
		t.Error("Thassa is offered to her own \"other target creature\"")
	}
	if !effectstest.HasID(pick.PickTargetCards, stolen) {
		t.Fatal("a creature you control is not offered")
	}
	effectstest.PickCard(t, g, me.ID, stolen)
	effectstest.PassPriorityAroundTable(t, g)

	back := effectstest.FindBattlefieldByName(g, "Stolen Bear")
	if back == uuid.Nil || back == stolen {
		t.Fatalf("blinked creature back=%v (old %v): want a new object on the battlefield", back, stolen)
	}
	if c, _ := effectstest.AangCardOnBF(g, back); c.Controller != me.ID {
		t.Errorf("returned under %v's control, want yours (%v)", c.Controller, me.ID)
	}
}

// {3}{U}: tap another target creature — an opponent's creature taps,
// Thassa herself is refused.
func TestThassaTapsAnotherTargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	thassa := pushThassa(g, me.ID)
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Blue Pips", TypeLine: "Enchantment", ManaCost: "{U}{U}{U}{U}",
		Owner: me.ID, Controller: me.ID,
	})
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	effectstest.ToMainForCost(t, g)

	if err := g.ActivateCatalogAbility(me.ID, thassa, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(thassa)}); !errors.Is(err, game.ErrIllegalTarget) {
		t.Errorf("Thassa targeting herself: err = %v, want ErrIllegalTarget", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, thassa, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(theirs)}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.AangCardOnBF(g, theirs); !c.Tapped {
		t.Error("the target creature did not become tapped")
	}
}

// "You may": choosing nothing returns nothing.
func TestAmbrosiaWhiteheartMayReturnNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushCatalogPermanent(g, me.ID, "Bears", "Creature — Bear", "", false)
	effectstest.CastAndResolveCreature(t, g, "Ambrosia Whiteheart", "Legendary Creature — Bird", ambrosiaWhiteheartOracle)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if pick == nil {
		t.Fatal("no resolution-time permanent choice")
	}
	if err := g.ResolveOwnPermanents(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("declining: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effects.OnBattlefield(g, bear) {
		t.Error("declining still returned a permanent")
	}
}
