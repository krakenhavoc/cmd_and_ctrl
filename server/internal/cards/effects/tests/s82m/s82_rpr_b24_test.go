package testss82m

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_rpr_b24_test.go — the reprint pass, batch 24 (#5908).

const (
	rprB24UrbanEvolutionOracle = "c715e525-fc3a-4cab-94a0-04e0a22a5013"
	rprB24VividMeadowOracle    = "dee99df5-628f-4a4e-a203-4dfddc927373"
	rprB24WingspanOracle       = "2b3b0e56-0b83-403b-bdfd-b5b7c0accf1b"
	rprB24WrexialOracle        = "fd36cc16-d3d9-4c9d-9d28-bbe7e5459d75"
	rprB24ZurzothOracle        = "a7da9ec5-f82f-4a6d-9737-06a1556e2dd6"
	rprB24CaptureOracle        = "0f730bd9-2060-46b1-9208-0ac6562e8b2a"
	rprB24FadeOracle           = "3bd3156f-dbc7-48f7-9e50-1f2b8b227f97"
	rprB24AdmiralOracle        = "dcde1a62-33a5-4454-a389-9bf7a3521a4c"
	rprB24OwlOracle            = "4bd38337-fb79-4809-848a-308c682825d1"
)

func rprB24Library(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		effectstest.PushLibraryCardForTest(p, game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Land"})
	}
}

func TestRprB24UrbanEvolutionDrawsThreeAndGrantsALandPlay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rprB24Library(me, 5)
	before := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Urban Evolution", "Sorcery", rprB24UrbanEvolutionOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 3 {
		t.Errorf("drew %d, want 3", got)
	}
	if got := g.ExtraLandDropsThisTurn[me.ID]; got != 1 {
		t.Errorf("extra land plays = %d, want 1", got)
	}
}

func TestRprB24FadeIntoAntiquityExilesAnArtifactOrEnchantmentOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Fade into Antiquity", "Sorcery", rprB24FadeOracle, effectstest.B16TargetCard(bear)); err == nil {
		t.Fatal("a plain creature was a legal target")
	}
	effectstest.CastCatalogSpell(t, g, "Fade into Antiquity", "Sorcery", rprB24FadeOracle, effectstest.B16TargetCard(rock))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, rock) {
		t.Error("the artifact was not exiled")
	}
}

func TestRprB24VividMeadowEntersTappedWithTwoChargeCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Vivid Meadow", TypeLine: "Land", OracleID: rprB24VividMeadowOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("play the land: %v", err)
	}
	if got := effectstest.CountersOn(g, id, game.CounterCharge); got != 2 {
		t.Errorf("%d charge counters, want 2", got)
	}
	if c := effectstest.BattlefieldCardPtr(t, g, id); !c.Tapped {
		t.Error("it did not enter tapped")
	}
	if n := len(game.ManaAbilitiesForCard(game.Card{OracleID: rprB24VividMeadowOracle})); n != 2 {
		t.Errorf("%d mana abilities, want 2", n)
	}
}

func TestRprB24AeronautAdmiralGivesVehiclesFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	adm := effectstest.PushCatalogPermanent(g, me.ID, "Aeronaut Admiral", "Creature — Human Pilot", rprB24AdmiralOracle, false)
	cart := effectstest.B12Permanent(g, me.ID, "My Cart", "Artifact — Vehicle")
	theirs := effectstest.B12Permanent(g, opp.ID, "Their Cart", "Artifact — Vehicle")
	effectstest.AssertKeywords(t, g, adm, "flying")
	effectstest.AssertKeywords(t, g, cart, "flying")
	if game.HasKeyword(effectstest.BattlefieldCardPtr(t, g, theirs), "flying") {
		t.Error("an opponent's Vehicle got flying")
	}
}

func TestRprB24ArcanistsOwlTakesAnArtifactOrEnchantmentFromTheTopFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rprB24Library(me, 6)
	rock := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: rock, Name: "Top Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Arcanist's Owl", TypeLine: "Artifact Creature — Bird", OracleID: rprB24OwlOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, rock)
	if !me.Hand.Contains(rock) {
		t.Error("the artifact did not reach the hand")
	}
}

func TestRprB24WingspanMentorPutsAFlyingCounterAndPumpsFliers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	human := effectstest.B12Creature(g, me.ID, "Farmer", "Creature — Human", 1, 1)
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Wingspan Mentor", TypeLine: "Creature — Human Wizard", OracleID: rprB24WingspanOracle, Power: 1, Toughness: 3, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	ch := effectstest.PickTargetPrompt(t, g)
	for _, c := range ch.PickTargetCards {
		if c == human || c == id {
			t.Fatalf("a Human was offered: %v", ch.PickTargetCards)
		}
	}
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertKeywords(t, g, bear, "flying")
	effectstest.FloatForTest(g, me, "UUU")
	mentor := effectstest.PushCatalogPermanent(g, me.ID, "Wingspan Mentor", "Creature — Human Wizard", rprB24WingspanOracle, false)
	effectstest.B16Activate(t, g, me.ID, mentor, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("the flier has %d +1/+1 counters, want 1", got)
	}
	if got := effectstest.CountersOn(g, human, game.CounterPlusOne); got != 0 {
		t.Errorf("a non-flier got %d +1/+1 counters", got)
	}
}

func TestRprB24EssenceCaptureCountersACreatureSpellAndPumps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	bolt := effectstest.B13OpponentCasts(t, g, opp, "Their Bolt", "Instant", "", "{R}", nil)
	id := effectstest.HandCardFull(me, "Essence Capture", "Instant", "{U}{U}", rprB24CaptureOracle, nil)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: bolt}, {Kind: game.TargetCard, ID: mine}}}); err == nil {
		t.Fatal("a noncreature spell was a legal target")
	}
	effectstest.PassPriorityAroundTable(t, g)
	bear := effectstest.B13PlayAs(t, g, 1, "Their Bear", "Creature — Bear", "")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: bear}, {Kind: game.TargetCard, ID: mine}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opp.Graveyard.Contains(bear) {
		t.Error("the creature spell was not countered")
	}
	if got := effectstest.CountersOn(g, mine, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
}

func TestRprB24ZurzothMakesADevilOnAnOpponentsFirstDrawOffTurnOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	me := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	other := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	effectstest.PushCatalogPermanent(g, me.ID, "Zurzoth, Chaos Rider", "Legendary Creature — Devil", rprB24ZurzothOracle, false)
	rprB24Library(active, 3)
	rprB24Library(other, 3)
	devils := func() int {
		n := 0
		for _, c := range g.Battlefield.Cards {
			if c.IsToken() && c.Name == "Devil" {
				n++
			}
		}
		return n
	}
	if err := g.DrawCard(active.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if devils() != 0 {
		t.Fatal("a draw on their own turn made a Devil")
	}
	if err := g.DrawCard(other.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if devils() != 1 {
		t.Fatalf("%d Devils after an off-turn first draw, want 1", devils())
	}
	if err := g.DrawCard(other.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if devils() != 1 {
		t.Errorf("a second draw made another Devil: %d", devils())
	}
}

func TestRprB24ZurzothDevilsAttackingDrawsAndDiscardsAtRandomForYouAndThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bystander := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	z := effectstest.PushCatalogPermanent(g, me.ID, "Zurzoth, Chaos Rider", "Legendary Creature — Devil", rprB24ZurzothOracle, false)
	rprB24Library(me, 4)
	rprB24Library(opp, 4)
	rprB24Library(bystander, 4)
	myHand, oppHand, byHand := me.Hand.Size(), opp.Hand.Size(), bystander.Hand.Size()
	myGY, oppGY := me.Graveyard.Size(), opp.Graveyard.Size()
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(z, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != myHand || opp.Hand.Size() != oppHand {
		t.Errorf("hands %d/%d, want unchanged (draw then discard)", me.Hand.Size(), opp.Hand.Size())
	}
	if me.Graveyard.Size() != myGY+1 || opp.Graveyard.Size() != oppGY+1 {
		t.Errorf("graveyards +%d/+%d, want +1/+1", me.Graveyard.Size()-myGY, opp.Graveyard.Size()-oppGY)
	}
	if bystander.Hand.Size() != byHand {
		t.Error("a player who was not attacked drew")
	}
}

func TestRprB24WrexialCastsAnInstantFromTheDamagedPlayersGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	w := effectstest.PushCatalogPermanent(g, me.ID, "Wrexial, the Risen Deep", "Legendary Creature — Kraken", rprB24WrexialOracle, false)
	bolt := effectstest.GyCard(opp, "Opp Instant", "Instant", "{U}", "")
	mine := effectstest.GyCard(me, "My Instant", "Instant", "{U}", "")
	effectstest.AttackWith(t, g, opp.ID, w)
	ch := effectstest.PickTargetPrompt(t, g)
	if len(ch.PickTargetCards) != 1 || ch.PickTargetCards[0] != bolt {
		t.Fatalf("only the damaged player's instant: %v (mine %s)", ch.PickTargetCards, mine)
	}
	effectstest.AnswerPickTarget(t, g, bolt)
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{FromZone: "graveyard", Strict: true,
		AlternativeCost: game.AltCostKeyGrantedCost}); err != nil {
		t.Fatalf("cast it free: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, bolt) {
		t.Error("the spell was not exiled instead of going to a graveyard")
	}
}

const (
	rprB24ArmedOracle    = "3d5b6151-d985-423b-8ade-f00958bbd3f6"
	rprB24JudgmentOracle = "7aacfc42-be6c-46a8-b474-f43a1b6b2ef8"
)

func TestRprB24ArmedAndArmoredAnimatesVehiclesAndAttachesChosenEquipmentToADwarf(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cart := effectstest.B12Permanent(g, me.ID, "My Cart", "Artifact — Vehicle")
	theirCart := effectstest.B12Permanent(g, opp.ID, "Their Cart", "Artifact — Vehicle")
	dwarf := effectstest.B12Creature(g, me.ID, "Dwarf A", "Creature — Dwarf", 1, 1)
	other := effectstest.B12Creature(g, me.ID, "Dwarf B", "Creature — Dwarf", 1, 1)
	human := effectstest.B12Creature(g, me.ID, "Farmer", "Creature — Human", 1, 1)
	blade := effectstest.B12Permanent(g, me.ID, "Blade", "Artifact — Equipment")
	shield := effectstest.B12Permanent(g, me.ID, "Shield", "Artifact — Equipment")
	effectstest.CastCatalogSpell(t, g, "Armed and Armored", "Instant", rprB24ArmedOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if types := effectstest.EffectiveTypes(t, g, cart); !slices.Contains(types, "Creature") || !slices.Contains(types, "Artifact") {
		t.Error("my Vehicle did not become a creature")
	}
	if slices.Contains(effectstest.EffectiveTypes(t, g, theirCart), "Creature") {
		t.Error("an opponent's Vehicle became a creature")
	}
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no Dwarf prompt")
	}
	for _, id := range []uuid.UUID{human, cart} {
		if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{id}); err == nil {
			t.Fatal("a non-Dwarf was a legal choice")
		}
	}
	effectstest.AnswerChooseCards(t, g, me.ID, dwarf)
	effectstest.AnswerChooseCards(t, g, me.ID, blade)
	if c := effectstest.BattlefieldCardPtr(t, g, blade); c.AttachedTo.ID != dwarf {
		t.Errorf("the chosen Equipment is not on the Dwarf: %+v", c.AttachedTo)
	}
	if c := effectstest.BattlefieldCardPtr(t, g, shield); c.AttachedTo.ID != uuid.Nil {
		t.Error("an Equipment that was not chosen was attached")
	}
	_ = other
}

func TestRprB24ZndrspltsJudgmentFriendsCopyAndFoesBounce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	foe := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	friend := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, foe.ID, "Foe Bear", "Creature — Bear", 2, 2)
	pals := effectstest.B12Creature(g, friend.ID, "Friend Wolf", "Creature — Wolf", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Zndrsplt's Judgment", "Sorcery", rprB24JudgmentOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// Seats are asked from the caster: me, then the next seats in order.
	for _, p := range orderedFrom(g, me.ID) {
		idx := 0 // friend
		if p == foe.ID {
			idx = 1
		}
		effectstest.AnswerOptionPick(t, g, me.ID, idx)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, mine)
	effectstest.AnswerChooseCards(t, g, friend.ID, pals)
	effectstest.AnswerChooseCards(t, g, foe.ID, theirs)
	count := func(p *game.Player, name string) int {
		n := 0
		for _, c := range g.Battlefield.Cards {
			if c.Controller == p.ID && c.Name == name {
				n++
			}
		}
		return n
	}
	if count(me, "My Bear") != 2 || count(friend, "Friend Wolf") != 2 {
		t.Errorf("friends did not copy: %d bears, %d wolves", count(me, "My Bear"), count(friend, "Friend Wolf"))
	}
	if g.Battlefield.Contains(theirs) || !foe.Hand.Contains(theirs) {
		t.Error("the foe's creature was not returned to its owner's hand")
	}
}

// orderedFrom lists the living seats from `first` in turn order.
func orderedFrom(g *game.Game, first uuid.UUID) []uuid.UUID {
	return effects.SeatsFromController(g, first)
}
