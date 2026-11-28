package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b703_cards_test.go — Urza's Destiny Auras, S82 OLD batch 703
// (#5621), played through the real catalog entries.

const (
	o703Disappear      = "039e2e39-8f7f-43f2-8448-de35ef7b8dc1"
	o703ChimeOfNight   = "09b41783-0583-4486-99e8-3f8a11ae37fd"
	o703MarkOfFury     = "1c350891-8810-4c14-b8ea-84486864637b"
	o703FesteringWound = "32638d03-c101-42f0-8a4b-8c6290257a28"
	o703PatternRebirth = "4b8856f8-3cd1-450c-a71b-eb0026e8e179"
	o703SigilOfSleep   = "51b47053-d169-4627-982f-f29b8630d75f"
	o703DyingWail      = "78a9cb8d-52bb-4182-ad7d-e4f8f0bb9797"
	o703Momentum       = "80875914-df4c-4b05-8c4d-658ff477c956"
	o703Treachery      = "8ed57194-7508-4aef-9373-64f7e80612d8"
	o703Incendiary     = "a893bee3-ea5c-4619-a875-e4af6d4396fe"
	o703ArcheryTrain   = "acab4fa9-d9ab-4a89-a50f-0cee9d7ca0c3"
	o703PrivateRes     = "e8578152-ee97-4471-bf32-77abe5ab80f1"
)

func TestS82OldB703CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		o703Disappear: "Disappear", o703ChimeOfNight: "Chime of Night", o703MarkOfFury: "Mark of Fury",
		o703FesteringWound: "Festering Wound", o703PatternRebirth: "Pattern of Rebirth", o703SigilOfSleep: "Sigil of Sleep",
		o703DyingWail: "Dying Wail", o703Momentum: "Momentum", o703Treachery: "Treachery",
		o703Incendiary: "Incendiary", o703ArcheryTrain: "Archery Training", o703PrivateRes: "Private Research",
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

// o703Kill destroys a creature and lets the state-based actions and the
// triggers they cause run.
func o703Kill(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	effectstest.B14Kill(g, id)
	g.RunStateChecksForTest()
}

// o703PickIfAsked answers an open pick_target prompt with the given target.
func o703PickIfAsked(t *testing.T, g *game.Game, chooser uuid.UUID, ref game.TargetRef) {
	t.Helper()
	if pick := effectstest.LatestPickTarget(g, chooser); pick != nil {
		if err := g.ResolvePickTargets(pick.ID, chooser, []game.TargetRef{ref}); err != nil {
			t.Fatalf("ResolvePickTargets: %v", err)
		}
	}
}

func o703Bear(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return effectstest.B12Creature(g, owner, name, "Creature — Bear", 2, 2)
}

func TestS82OldB703DisappearReturnsTheCreatureAndItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := o703Bear(g, me.ID, "My Bear")
	aura := effectstest.Bns44Aura(t, g, "Disappear", o703Disappear, bear)
	effectstest.AddPool(t, g, me.ID, "{U}")
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || g.Battlefield.Contains(aura) {
		t.Error("the creature or the Aura stayed on the battlefield")
	}
	if !me.Hand.Contains(bear) || !me.Hand.Contains(aura) {
		t.Error("the creature and the Aura were not both returned to their owner's hand")
	}
}

func TestS82OldB703ChimeOfNightDestroysANonblackCreatureWhenItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := o703Bear(g, me.ID, "My Bear")
	green := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Green Bear", TypeLine: "Creature — Bear",
		Colors: []string{"G"}, Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	black := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Black Bear", TypeLine: "Creature — Bear",
		Colors: []string{"B"}, Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	effectstest.Bns44Aura(t, g, "Chime of Night", o703ChimeOfNight, mine)
	o703Kill(t, g, mine)
	o703PickIfAsked(t, g, me.ID, game.TargetRef{Kind: game.TargetCard, ID: green})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(green) {
		t.Error("the nonblack creature survived")
	}
	if !g.Battlefield.Contains(black) {
		t.Error("the black creature was destroyed")
	}
}

func TestS82OldB703MarkOfFuryGivesHasteAndReturnsAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := o703Bear(g, me.ID, "My Bear")
	aura := effectstest.Bns44Aura(t, g, "Mark of Fury", o703MarkOfFury, bear)
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "haste") {
		t.Error("the enchanted creature has no haste")
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) || !me.Hand.Contains(aura) {
		t.Error("the Aura was not returned to its owner's hand at the end step")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, bear, "haste") {
		t.Error("the creature kept haste after the Aura left")
	}
}

func TestS82OldB703FesteringWoundGrowsAtYourUpkeepAndBurnsTheCreaturesController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := o703Bear(g, opp.ID, "Their Bear")
	aura := effectstest.Bns44Aura(t, g, "Festering Wound", o703FesteringWound, theirs)
	effectstest.WalkToUpkeepOfSeat0(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B12Counter(t, g, aura, "infection"); n != 1 {
		t.Fatalf("%d infection counters, want 1", n)
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(aura, "infection", 2) })
	mine, theirLife := me.Life, opp.Life
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != theirLife-3 {
		t.Errorf("the creature's controller is at %d, want %d", opp.Life, theirLife-3)
	}
	if me.Life != mine {
		t.Errorf("the Aura's controller lost life: %d", me.Life)
	}
}

func TestS82OldB703PatternOfRebirthLetsTheCreaturesControllerSearch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := o703Bear(g, me.ID, "My Bear")
	big := effectstest.SsCard(me.Library, me.ID, "Library Ogre", "Creature — Ogre", 5, 5)
	effectstest.Bns44Aura(t, g, "Pattern of Rebirth", o703PatternRebirth, mine)
	o703Kill(t, g, mine)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, big)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(big) {
		t.Error("the searched creature did not enter the battlefield")
	}
}

func TestS82OldB703SigilOfSleepBouncesACreatureOfTheDamagedPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := o703Bear(g, me.ID, "My Bear")
	theirs := o703Bear(g, opp.ID, "Their Bear")
	effectstest.Bns44Aura(t, g, "Sigil of Sleep", o703SigilOfSleep, mine)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(mine, opp.ID, 1) })
	o703PickIfAsked(t, g, me.ID, game.TargetRef{Kind: game.TargetCard, ID: theirs})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) || !opp.Hand.Contains(theirs) {
		t.Error("the damaged player's creature was not returned to its owner's hand")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("the enchanted creature was bounced")
	}
}

func TestS82OldB703DyingWailMakesTheTargetDiscardTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := o703Bear(g, me.ID, "My Bear")
	a := effectstest.SsCard(opp.Hand, opp.ID, "A", "Sorcery", 0, 0)
	b := effectstest.SsCard(opp.Hand, opp.ID, "B", "Sorcery", 0, 0)
	c := effectstest.SsCard(opp.Hand, opp.ID, "C", "Sorcery", 0, 0)
	effectstest.Bns44Aura(t, g, "Dying Wail", o703DyingWail, mine)
	o703Kill(t, g, mine)
	o703PickIfAsked(t, g, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, opp.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(a) || !opp.Graveyard.Contains(b) || !opp.Hand.Contains(c) {
		t.Error("the target did not discard exactly the two chosen cards")
	}
}

func TestS82OldB703MomentumGrowsAndPumpsPerCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := o703Bear(g, me.ID, "My Bear")
	aura := effectstest.Bns44Aura(t, g, "Momentum", o703Momentum, bear)
	if p, tough := effectstest.PtOf(t, g, bear); p != 2 || tough != 2 {
		t.Fatalf("with no counters the bear is %d/%d, want 2/2", p, tough)
	}
	effectstest.WalkToUpkeepOfSeat0(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B12Counter(t, g, aura, "growth"); n != 1 {
		t.Fatalf("%d growth counters, want 1", n)
	}
	if p, tough := effectstest.PtOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("with one counter the bear is %d/%d, want 3/3", p, tough)
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(aura, "growth", 2) })
	if p, tough := effectstest.PtOf(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("with three counters the bear is %d/%d, want 5/5", p, tough)
	}
}

func TestS82OldB703MomentumUpkeepCounterIsOptional(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := o703Bear(g, me.ID, "My Bear")
	aura := effectstest.Bns44Aura(t, g, "Momentum", o703Momentum, bear)
	effectstest.WalkToUpkeepOfSeat0(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B12Counter(t, g, aura, "growth"); n != 0 {
		t.Errorf("%d growth counters after declining, want 0", n)
	}
}

func TestS82OldB703TreacheryStealsTheCreatureAndUntapsLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := o703Bear(g, opp.ID, "Their Bear")
	var lands []uuid.UUID
	for i := 0; i < 6; i++ {
		land := effectstest.B12Permanent(g, me.ID, "Island", "Basic Land — Island")
		effectstest.B02bSetTapped(g, land, true)
		lands = append(lands, land)
	}
	effectstest.CastCatalogSpell(t, g, "Treachery", effectstest.AuraTypeLine, o703Treachery, effectstest.CardRefs(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, lands[:5]...)
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, theirs); c.Controller != me.ID {
		t.Error("the Aura's controller does not control the enchanted creature")
	}
	untapped := 0
	for _, l := range lands {
		if c, _ := effectstest.BattlefieldCard(g, l); !c.Tapped {
			untapped++
		}
	}
	if untapped != 5 {
		t.Errorf("%d lands untapped, want 5", untapped)
	}
}

func TestS82OldB703IncendiaryDealsDamageEqualToItsFuseCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := o703Bear(g, me.ID, "My Bear")
	aura := effectstest.Bns44Aura(t, g, "Incendiary", o703Incendiary, mine)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(aura, "fuse", 3) })
	life := opp.Life
	o703Kill(t, g, mine)
	o703PickIfAsked(t, g, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("the target is at %d, want %d (three fuse counters)", opp.Life, life-3)
	}
}

func TestS82OldB703ArcheryTrainingShootsAnAttackerForItsArrowCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	archer := o703Bear(g, me.ID, "Archer")
	ogre := effectstest.B12Creature(g, me.ID, "Attacking Ogre", "Creature — Ogre", 3, 4)
	aura := effectstest.Bns44Aura(t, g, "Archery Training", o703ArcheryTrain, archer)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(aura, "arrow", 2) })
	if err := g.ActivateCatalogAbility(me.ID, archer, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(ogre)}); err == nil {
		t.Fatal("the ability targeted a creature that is not attacking or blocking")
	}
	effectstest.BrAttack(t, g, ogre)
	if err := g.ActivateCatalogAbility(me.ID, archer, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(ogre)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, ogre); got != 2 {
		t.Errorf("the attacker has %d damage, want 2 (two arrow counters)", got)
	}
	if c, _ := effectstest.BattlefieldCard(g, archer); !c.Tapped {
		t.Error("the ability did not tap the enchanted creature")
	}
}

func TestS82OldB703PrivateResearchDrawsForEachPageCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := o703Bear(g, me.ID, "My Bear")
	aura := effectstest.Bns44Aura(t, g, "Private Research", o703PrivateRes, mine)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(aura, "page", 2) })
	hand := effectstest.HandSize(me)
	o703Kill(t, g, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandSize(me); got != hand+2 {
		t.Errorf("hand %d, want %d (two page counters)", got, hand+2)
	}
}
