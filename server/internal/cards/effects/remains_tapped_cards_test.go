package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// remains_tapped_cards_test.go is ADR 0109 Delivery PR 2 (#1894): "for
// as long as you control this and it remains tapped" (a duration with
// two conditions), "for as long as that creature remains tapped", and
// the single "for as long as this remains tapped" cards of owner
// decision 4.

const (
	rtSeasingerOracle     = "13745b92-1e30-4075-9733-5d00da11408c"
	rtShacklesOracle      = "fddac567-c42c-453b-9c82-c3140b374a06"
	rtEvergloveOracle     = "31da3df3-5b79-42d0-9018-76bc73411954"
	rtManaLeechOracle     = "cad05528-917b-4296-9d23-ba9433612709"
	rtKillSwitchOracle    = "0cd5eef6-6830-4c6d-8578-2726900c39a5"
	rtZygonOracle         = "693c071a-d6b5-4a96-8fe3-8c61fa9406bc"
	rtHisokasGuardOracle  = "6136f019-da8f-471e-9ac5-41e535aa8a21"
	rtBlackstaffOracle    = "a4faed58-4b1b-4cab-8a12-4fcd20b320ab"
	rtHelmOracle          = "78f22413-4e61-4191-9d30-7b5d6538446e"
	rtEndoskeletonOracle  = "e9ef7f8c-21ef-45f0-af64-8ffa9f0089d1"
	rtIceFloeOracle       = "cfaaead2-09e8-47cb-9e39-8570b8d8de86"
	rtAquitectsWillOracle = "d02befcb-3e2c-4cfe-a913-9bb648e5bb2c"
	rtMinasMorgulOracle   = "867dbd5a-c3cf-41ce-980b-c9babc6f30f2"
	rtMannequinOracle     = "d9dfdbbb-75b3-43d3-8ed6-371cb98a0124"
	rtMathasOracle        = "1e8d9d0c-185b-4149-9ec9-74a2f2fe8764"
	rtFireheartOracle     = "c3b3e070-28f6-4852-816e-caffe358b6fe"
	rtXolatoyacOracle     = "2c0fc559-857f-4239-97fa-379ced3dfc1f"
)

// rtActivate pays `mana` into the controller's pool and activates the
// source's ability `index`, then resolves it.
func rtActivate(t *testing.T, g *game.Game, controller, source uuid.UUID, index int, targets []game.TargetRef, mana ...string) {
	t.Helper()
	b06AddMana(g.PlayerByIDForEffect(controller), mana...)
	if err := g.ActivateCatalogAbility(controller, source, index, game.ActivateAbilityParams{Targets: targets}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g)
}

func rtUntap(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(id) })
}

func rtTapped(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	return layeredCard(t, g, id).Tapped
}

// TestSeasingerEndsWhenEitherHalfStops is ADR 0109 §3's headline: one
// duration, two conditions, and the effect ends when either stops.
func TestSeasingerEndsWhenEitherHalfStops(t *testing.T) {
	cases := map[string]func(g *game.Game, seasinger, opp uuid.UUID){
		"Seasinger untaps": func(g *game.Game, seasinger, _ uuid.UUID) { rtUntap(g, seasinger) },
		"you lose control of Seasinger": func(g *game.Game, seasinger, opp uuid.UUID) {
			g.WithWriteLock(func() { g.GainControlForEffect(uuid.Nil, seasinger, opp, game.IndefiniteDuration(), "theft") })
		},
	}
	for name, stop := range cases {
		t.Run(name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			pushLandFor(g, me.ID, "Island", "Basic Land — Island")
			pushLandFor(g, opp.ID, "Island", "Basic Land — Island")
			seasinger := pushCatalogPermanent(g, me.ID, "Seasinger", "Creature — Merfolk", rtSeasingerOracle, false)
			victim := ctrlPushCreature(g, opp.ID, "Bear")
			rtActivate(t, g, me.ID, seasinger, 0, ltCardTarget(victim))
			if got := controllerOf(t, g, victim); got != me.ID {
				t.Fatalf("setup: controller %s, want %s", got, me.ID)
			}
			stop(g, seasinger, opp.ID)
			if got := controllerOf(t, g, victim); got != opp.ID {
				t.Errorf("%s and the creature stayed stolen (CR 611.2b)", name)
			}
		})
	}
}

// TestSeasingerNeedsAnIslandUnderTheTargetsController is its target
// clause: "whose controller controls an Island".
func TestSeasingerNeedsAnIslandUnderTheTargetsController(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushLandFor(g, me.ID, "Island", "Basic Land — Island")
	seasinger := pushCatalogPermanent(g, me.ID, "Seasinger", "Creature — Merfolk", rtSeasingerOracle, false)
	victim := ctrlPushCreature(g, opp.ID, "Bear")
	if err := g.ActivateCatalogAbility(me.ID, seasinger, 0, game.ActivateAbilityParams{Targets: ltCardTarget(victim)}); err == nil {
		t.Error("Seasinger targeted a creature whose controller controls no Island")
	}
}

// TestHelmOfPossessionHoldsBothConditions is the same pair on an
// artifact with a sacrifice cost.
func TestHelmOfPossessionHoldsBothConditions(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	helm := pushCatalogPermanent(g, me.ID, "Helm of Possession", "Artifact", rtHelmOracle, false)
	fodder := ctrlPushCreature(g, me.ID, "Fodder")
	victim := ctrlPushCreature(g, opp.ID, "Bear")
	b06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, helm, 0, game.ActivateAbilityParams{
		Targets: ltCardTarget(victim), SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := controllerOf(t, g, victim); got != me.ID {
		t.Fatalf("controller %s, want %s", got, me.ID)
	}
	rtUntap(g, helm)
	if got := controllerOf(t, g, victim); got != opp.ID {
		t.Error("the Helm untapped and the creature stayed stolen")
	}
}

// TestVedalkenShacklesCountsIslands: the target bound is the number of
// Islands you control, and the theft lasts while the Shackles stay
// tapped.
func TestVedalkenShacklesCountsIslands(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushLandFor(g, me.ID, "Island", "Basic Land — Island")
	pushLandFor(g, me.ID, "Island", "Basic Land — Island")
	shackles := pushCatalogPermanent(g, me.ID, "Vedalken Shackles", "Artifact", rtShacklesOracle, false)
	small := ctrlPushCreature(g, opp.ID, "Bear")
	big := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Giant",
		TypeLine: "Creature — Giant", Power: 3, Toughness: 3, Owner: opp.ID, Controller: opp.ID})
	b06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, shackles, 0, game.ActivateAbilityParams{Targets: ltCardTarget(big)}); err == nil {
		t.Fatal("the Shackles targeted a 3-power creature with two Islands")
	}
	rtActivate(t, g, me.ID, shackles, 0, ltCardTarget(small))
	if got := controllerOf(t, g, small); got != me.ID {
		t.Fatalf("controller %s, want %s", got, me.ID)
	}
	rtUntap(g, shackles)
	if got := controllerOf(t, g, small); got != opp.ID {
		t.Error("the Shackles untapped and the creature stayed stolen")
	}
}

// TestEvergloveCourierPumpsAnElfWhileTapped is the Couriers' shape.
func TestEvergloveCourierPumpsAnElfWhileTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	courier := pushCatalogPermanent(g, me.ID, "Everglove Courier", "Creature — Elf", rtEvergloveOracle, false)
	elf := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Llanowar Elves",
		TypeLine: "Creature — Elf Druid", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	rtActivate(t, g, me.ID, courier, 0, ltCardTarget(elf), "G", "C", "C")
	c := layeredCard(t, g, elf)
	if c.CurrentPower() != 3 || !game.HasKeyword(&c, "trample") {
		t.Fatalf("pumped elf: power %d, trample %v; want 3 and trample", c.CurrentPower(), game.HasKeyword(&c, "trample"))
	}
	rtUntap(g, courier)
	c = layeredCard(t, g, elf)
	if c.CurrentPower() != 1 || game.HasKeyword(&c, "trample") {
		t.Errorf("after the Courier untapped: power %d, trample %v; want 1 and none", c.CurrentPower(), game.HasKeyword(&c, "trample"))
	}
}

// TestEndoskeletonNeverStartsOnceUntapped is CR 611.2b's "never
// starts": the source untapped in response, so nothing is registered.
func TestEndoskeletonNeverStartsOnceUntapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	endo := pushCatalogPermanent(g, me.ID, "Endoskeleton", "Artifact", rtEndoskeletonOracle, false)
	bear := ctrlPushCreature(g, me.ID, "Bear")
	b06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, endo, 0, game.ActivateAbilityParams{Targets: ltCardTarget(bear)}); err != nil {
		t.Fatal(err)
	}
	rtUntap(g, endo)
	passPriorityAroundTable(t, g)
	if got := layeredCard(t, g, bear).CurrentToughness(); got != 2 {
		t.Errorf("toughness %d, want 2: the Endoskeleton was untapped as its ability resolved", got)
	}
}

// TestManaLeechHoldsTheLandWhileItRemainsTapped records Rust Tick's
// hold on the land.
func TestManaLeechHoldsTheLandWhileItRemainsTapped(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	leech := pushCatalogPermanent(g, me.ID, "Mana Leech", "Creature — Leech", rtManaLeechOracle, false)
	land := pushLandFor(g, opp.ID, "Forest", "Basic Land — Forest")
	rtActivate(t, g, me.ID, leech, 0, ltCardTarget(land))
	c := layeredCard(t, g, land)
	if !c.Tapped || len(c.NextUntapSkips) != 1 || c.NextUntapSkips[0].While == nil ||
		c.NextUntapSkips[0].While.Condition != game.WhileSourceRemainsTapped {
		t.Fatalf("land: tapped %v, holds %+v; want tapped and held while the Leech remains tapped", c.Tapped, c.NextUntapSkips)
	}
}

// TestKillSwitchTapsEveryOtherArtifact: the set is every other
// artifact as the ability resolves, each held.
func TestKillSwitchTapsEveryOtherArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ks := pushCatalogPermanent(g, me.ID, "Kill Switch", "Artifact", rtKillSwitchOracle, false)
	mine := pushTappedPermanentForTest(g, me.ID, "Sol Ring", "Artifact")
	theirs := pushTappedPermanentForTest(g, opp.ID, "Mind Stone", "Artifact")
	bear := ctrlPushCreature(g, opp.ID, "Bear")
	rtActivate(t, g, me.ID, ks, 0, nil, "C", "C")
	for _, id := range []uuid.UUID{mine, theirs} {
		if c := layeredCard(t, g, id); !c.Tapped || len(c.NextUntapSkips) != 1 {
			t.Errorf("%s: tapped %v, holds %d; want tapped and held", c.Name, c.Tapped, len(c.NextUntapSkips))
		}
	}
	if rtTapped(t, g, bear) {
		t.Error("Kill Switch tapped a creature")
	}
	if len(layeredCard(t, g, ks).NextUntapSkips) != 0 {
		t.Error("Kill Switch held itself")
	}
}

// pushTappedPermanentForTest seeds an untapped permanent with the
// entry event, so it is stamped.
func pushTappedPermanentForTest(g *game.Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
		Owner: owner, Controller: owner})
}

// TestZygonInfiltratorIsACopyWhileThatCreatureRemainsTapped is the
// pinned condition: the copy is about the creature copied. The stun
// counter spends its first untap.
func TestZygonInfiltratorIsACopyWhileThatCreatureRemainsTapped(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	zygon := pushCatalogPermanent(g, me.ID, "Zygon Infiltrator", "Creature — Alien Shapeshifter Soldier", rtZygonOracle, false)
	model := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Hill Giant",
		TypeLine: "Creature — Giant", Power: 3, Toughness: 3, Owner: opp.ID, Controller: opp.ID})
	rtActivate(t, g, me.ID, zygon, 0, ltCardTarget(model), "U", "C", "C")
	if got := layeredCard(t, g, zygon).Name; got != "Hill Giant" {
		t.Fatalf("Zygon is %q, want a copy of Hill Giant", got)
	}
	m := layeredCard(t, g, model)
	if !m.Tapped || m.Counters[game.CounterStun] != 1 {
		t.Fatalf("model: tapped %v, stun %d; want tapped with one stun counter", m.Tapped, m.Counters[game.CounterStun])
	}
	rtUntap(g, model) // the stun counter goes instead (CR 122.1d)
	if got := layeredCard(t, g, zygon).Name; got != "Hill Giant" {
		t.Fatalf("after the stun counter went: Zygon is %q, want still Hill Giant", got)
	}
	rtUntap(g, model)
	if got := layeredCard(t, g, zygon).Name; got != "Zygon Infiltrator" {
		t.Errorf("after the model untapped: Zygon is %q, want itself again", got)
	}
}

// TestHisokasGuardCannotTargetItself is "other than this creature".
func TestHisokasGuardCannotTargetItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	guard := pushCatalogPermanent(g, me.ID, "Hisoka's Guard", "Creature — Human Wizard", rtHisokasGuardOracle, false)
	b06AddMana(me, "U", "C")
	if err := g.ActivateCatalogAbility(me.ID, guard, 0, game.ActivateAbilityParams{Targets: ltCardTarget(guard)}); err == nil {
		t.Error("Hisoka's Guard targeted itself")
	}
}

// TestBlackstaffAnimatesAnArtifactWhileTapped: a 4/4 artifact creature
// that keeps its own types.
func TestBlackstaffAnimatesAnArtifactWhileTapped(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[0]
	staff := pushCatalogPermanent(g, me.ID, "The Blackstaff of Waterdeep", "Legendary Artifact", rtBlackstaffOracle, false)
	rock := pushTappedPermanentForTest(g, me.ID, "Mind Stone", "Artifact")
	rtActivate(t, g, me.ID, staff, 0, ltCardTarget(rock), "U", "C")
	c := layeredCard(t, g, rock)
	if !c.IsCreature() || !c.IsArtifact() || c.CurrentPower() != 4 || c.CurrentToughness() != 4 {
		t.Fatalf("Mind Stone: creature %v artifact %v %d/%d; want a 4/4 artifact creature",
			c.IsCreature(), c.IsArtifact(), c.CurrentPower(), c.CurrentToughness())
	}
	rtUntap(g, staff)
	if layeredCard(t, g, rock).IsCreature() {
		t.Error("the Blackstaff untapped and the artifact is still a creature")
	}
}

// TestIceFloeTargetsOnlyAttackersOfYou is its target clause.
func TestIceFloeTargetsOnlyAttackersOfYou(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	floe := pushCatalogPermanent(g, me.ID, "Ice Floe", "Land", rtIceFloeOracle, false)
	idle := ctrlPushCreature(g, opp.ID, "Idle Bear")
	if err := g.ActivateCatalogAbility(me.ID, floe, 0, game.ActivateAbilityParams{Targets: ltCardTarget(idle)}); err == nil {
		t.Error("Ice Floe targeted a creature that is not attacking")
	}
	pred := And(WithoutKeyword("flying"), AttackingYou())
	attacker := game.Card{Name: "Raider", TypeLine: "Creature — Orc", Controller: opp.ID, AttackingTarget: me.ID}
	if !pred(g, me.ID, attacker) {
		t.Error("a creature attacking you is not a legal target")
	}
	attacker.AttackingTarget = uuid.New() // a planeswalker, say
	if pred(g, me.ID, attacker) {
		t.Error("a creature attacking something else is a legal target")
	}
}

// --- counter-held (ADR 0109 §2, #1604) ---------------------------

// TestAquitectsWillFloodsALandUntilTheCounterGoes: an Island in
// addition to its other types, for as long as it has a flood counter.
func TestAquitectsWillFloodsALandUntilTheCounterGoes(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	forest := pushLandFor(g, opp.ID, "Forest", "Basic Land — Forest")
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Merfolk",
		TypeLine: "Creature — Merfolk", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	hand := len(me.Hand.Cards)
	castCatalogSpell(t, g, "Aquitect's Will", "Kindred Sorcery — Merfolk", rtAquitectsWillOracle, ltCardTarget(forest))
	passPriorityAroundTable(t, g)
	if got := effectiveSubtypes(t, g, forest); !typeListHasForTest(got, "Island") || !typeListHasForTest(got, "Forest") {
		t.Fatalf("subtypes %v, want Forest and Island", got)
	}
	if got := len(me.Hand.Cards); got != hand+1 {
		t.Errorf("hand %d, want %d: a Merfolk under you draws a card", got, hand+1)
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(forest, "flood", -1) })
	if got := effectiveSubtypes(t, g, forest); typeListHasForTest(got, "Island") {
		t.Errorf("the flood counter went and the land is still an Island: %v", got)
	}
}

func typeListHasForTest(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}

// TestMinasMorgulGivesShadowAndWraith: the shadow counter is a keyword
// counter, the Wraith type is the counter-held effect.
func TestMinasMorgulGivesShadowAndWraith(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	morgul := pushCatalogPermanent(g, me.ID, "Minas Morgul, Dark Fortress", "Legendary Land", rtMinasMorgulOracle, false)
	bear := ctrlPushCreature(g, me.ID, "Bear")
	rtActivate(t, g, me.ID, morgul, 0, ltCardTarget(bear), "B", "C", "C", "C")
	c := layeredCard(t, g, bear)
	if !game.HasKeyword(&c, "shadow") || !typeListHasForTest(effectiveSubtypes(t, g, bear), "Wraith") {
		t.Fatalf("bear: shadow %v, subtypes %v; want shadow and Wraith", game.HasKeyword(&c, "shadow"), effectiveSubtypes(t, g, bear))
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, game.CounterShadow, -1) })
	if typeListHasForTest(effectiveSubtypes(t, g, bear), "Wraith") {
		t.Error("the shadow counter went and the creature is still a Wraith")
	}
}

// TestMakeshiftMannequinSacrificesWhenTargeted: the creature returns
// with its counter and the granted trigger.
func TestMakeshiftMannequinSacrificesWhenTargeted(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dead := pushGraveyardCardTyped(me, "Bear", "Creature — Bear")
	castCatalogSpell(t, g, "Makeshift Mannequin", "Instant", rtMannequinOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: dead}})
	passPriorityAroundTable(t, g)
	c := layeredCard(t, g, dead)
	if c.Counters["mannequin"] != 1 {
		t.Fatalf("returned with counters %v, want a mannequin counter", c.Counters)
	}
	endo := pushCatalogPermanent(g, me.ID, "Endoskeleton", "Artifact", rtEndoskeletonOracle, false)
	rtActivate(t, g, me.ID, endo, 0, ltCardTarget(dead), "C", "C")
	passPriorityAroundTable(t, g)
	if onBattlefield(g, dead) {
		t.Error("the returned creature was targeted and is still on the battlefield")
	}
}

// TestMathasBountyRewardsTheOpponents: the dies trigger is the bounty
// creature's, so "each opponent" is its controller's opponents.
func TestMathasBountyRewardsTheOpponents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Mathas, Fiend Seeker", "Legendary Creature — Vampire", rtMathasOracle, false)
	victim := ctrlPushCreature(g, opp.ID, "Bear")
	advanceToEndStep(t, g)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if got := layeredCard(t, g, victim).Counters["bounty"]; got != 1 {
		t.Fatalf("bounty counters %d, want 1", got)
	}
	advanceToMain(t, g)
	hand, life := len(me.Hand.Cards), me.Life
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(victim) })
	passPriorityAroundTable(t, g)
	if len(me.Hand.Cards) != hand+1 || me.Life != life+2 {
		t.Errorf("after the bounty creature died: hand %d (was %d), life %d (was %d); want +1 and +2",
			len(me.Hand.Cards), hand, me.Life, life)
	}
}

// TestObsidianFireheartGrantsTheBlaze: the land carries the upkeep
// trigger for as long as it has a blaze counter, and may not be chosen
// again while it does.
func TestObsidianFireheartGrantsTheBlaze(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fire := pushCatalogPermanent(g, me.ID, "Obsidian Fireheart", "Creature — Elemental", rtFireheartOracle, false)
	land := pushLandFor(g, opp.ID, "Plains", "Basic Land — Plains")
	rtActivate(t, g, me.ID, fire, 0, ltCardTarget(land), "R", "R", "C")
	c := layeredCard(t, g, land)
	if c.Counters["blaze"] != 1 || len(c.Effective().GrantedAbilities) == 0 {
		t.Fatalf("land: counters %v, granted %v; want a blaze counter and the grant", c.Counters, c.Effective().GrantedAbilities)
	}
	b06AddMana(me, "R", "R", "C")
	if err := g.ActivateCatalogAbility(me.ID, fire, 0, game.ActivateAbilityParams{Targets: ltCardTarget(land)}); err == nil {
		t.Error("the Fireheart targeted a land that already has a blaze counter")
	}
}

// TestXolatoyacFloodsALandAsItEnters is Aquitect's Will's flood on an
// enters trigger.
func TestXolatoyacFloodsALandAsItEnters(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[0]
	forest := pushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	castAndResolveCreature(t, g, "Xolatoyac, the Smiling Flood", "Legendary Creature — Salamander Serpent", rtXolatoyacOracle)
	pickCard(t, g, me.ID, forest)
	passPriorityAroundTable(t, g)
	if got := effectiveSubtypes(t, g, forest); !typeListHasForTest(got, "Island") {
		t.Errorf("subtypes %v, want an Island in addition", got)
	}
}
