package testsr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestSeasingerEndsWhenEitherHalfStops is ADR 0109 §3's headline: one
// duration, two conditions, and the effect ends when either stops.
func TestSeasingerEndsWhenEitherHalfStops(t *testing.T) {
	cases := map[string]func(g *game.Game, seasinger, opp uuid.UUID){
		"Seasinger untaps": func(g *game.Game, seasinger, _ uuid.UUID) { effectstest.RtUntap(g, seasinger) },
		"you lose control of Seasinger": func(g *game.Game, seasinger, opp uuid.UUID) {
			g.WithWriteLock(func() { g.GainControlForEffect(uuid.Nil, seasinger, opp, game.IndefiniteDuration(), "theft") })
		},
	}
	for name, stop := range cases {
		t.Run(name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.PushLandFor(g, me.ID, "Island", "Basic Land — Island")
			effectstest.PushLandFor(g, opp.ID, "Island", "Basic Land — Island")
			seasinger := effectstest.PushCatalogPermanent(g, me.ID, "Seasinger", "Creature — Merfolk", effectstest.RtSeasingerOracle, false)
			victim := effectstest.CtrlPushCreature(g, opp.ID, "Bear")
			effectstest.RtActivate(t, g, me.ID, seasinger, 0, effectstest.LtCardTarget(victim))
			if got := effectstest.ControllerOf(t, g, victim); got != me.ID {
				t.Fatalf("setup: controller %s, want %s", got, me.ID)
			}
			stop(g, seasinger, opp.ID)
			if got := effectstest.ControllerOf(t, g, victim); got != opp.ID {
				t.Errorf("%s and the creature stayed stolen (CR 611.2b)", name)
			}
		})
	}
}

// TestSeasingerNeedsAnIslandUnderTheTargetsController is its target
// clause: "whose controller controls an Island".
func TestSeasingerNeedsAnIslandUnderTheTargetsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushLandFor(g, me.ID, "Island", "Basic Land — Island")
	seasinger := effectstest.PushCatalogPermanent(g, me.ID, "Seasinger", "Creature — Merfolk", effectstest.RtSeasingerOracle, false)
	victim := effectstest.CtrlPushCreature(g, opp.ID, "Bear")
	if err := g.ActivateCatalogAbility(me.ID, seasinger, 0, game.ActivateAbilityParams{Targets: effectstest.LtCardTarget(victim)}); err == nil {
		t.Error("Seasinger targeted a creature whose controller controls no Island")
	}
}

// TestHelmOfPossessionHoldsBothConditions is the same pair on an
// artifact with a sacrifice cost.
func TestHelmOfPossessionHoldsBothConditions(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	helm := effectstest.PushCatalogPermanent(g, me.ID, "Helm of Possession", "Artifact", effectstest.RtHelmOracle, false)
	fodder := effectstest.CtrlPushCreature(g, me.ID, "Fodder")
	victim := effectstest.CtrlPushCreature(g, opp.ID, "Bear")
	effectstest.B06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, helm, 0, game.ActivateAbilityParams{
		Targets: effectstest.LtCardTarget(victim), SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.ControllerOf(t, g, victim); got != me.ID {
		t.Fatalf("controller %s, want %s", got, me.ID)
	}
	effectstest.RtUntap(g, helm)
	if got := effectstest.ControllerOf(t, g, victim); got != opp.ID {
		t.Error("the Helm untapped and the creature stayed stolen")
	}
}

// TestVedalkenShacklesCountsIslands: the target bound is the number of
// Islands you control, and the theft lasts while the Shackles stay
// tapped.
func TestVedalkenShacklesCountsIslands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushLandFor(g, me.ID, "Island", "Basic Land — Island")
	effectstest.PushLandFor(g, me.ID, "Island", "Basic Land — Island")
	shackles := effectstest.PushCatalogPermanent(g, me.ID, "Vedalken Shackles", "Artifact", effectstest.RtShacklesOracle, false)
	small := effectstest.CtrlPushCreature(g, opp.ID, "Bear")
	big := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Giant",
		TypeLine: "Creature — Giant", Power: 3, Toughness: 3, Owner: opp.ID, Controller: opp.ID})
	effectstest.B06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, shackles, 0, game.ActivateAbilityParams{Targets: effectstest.LtCardTarget(big)}); err == nil {
		t.Fatal("the Shackles targeted a 3-power creature with two Islands")
	}
	effectstest.RtActivate(t, g, me.ID, shackles, 0, effectstest.LtCardTarget(small))
	if got := effectstest.ControllerOf(t, g, small); got != me.ID {
		t.Fatalf("controller %s, want %s", got, me.ID)
	}
	effectstest.RtUntap(g, shackles)
	if got := effectstest.ControllerOf(t, g, small); got != opp.ID {
		t.Error("the Shackles untapped and the creature stayed stolen")
	}
}

// TestEvergloveCourierPumpsAnElfWhileTapped is the Couriers' shape.
func TestEvergloveCourierPumpsAnElfWhileTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	courier := effectstest.PushCatalogPermanent(g, me.ID, "Everglove Courier", "Creature — Elf", effectstest.RtEvergloveOracle, false)
	elf := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Llanowar Elves",
		TypeLine: "Creature — Elf Druid", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	effectstest.RtActivate(t, g, me.ID, courier, 0, effectstest.LtCardTarget(elf), "G", "C", "C")
	c := effectstest.LayeredCard(t, g, elf)
	if c.CurrentPower() != 3 || !game.HasKeyword(&c, "trample") {
		t.Fatalf("pumped elf: power %d, trample %v; want 3 and trample", c.CurrentPower(), game.HasKeyword(&c, "trample"))
	}
	effectstest.RtUntap(g, courier)
	c = effectstest.LayeredCard(t, g, elf)
	if c.CurrentPower() != 1 || game.HasKeyword(&c, "trample") {
		t.Errorf("after the Courier untapped: power %d, trample %v; want 1 and none", c.CurrentPower(), game.HasKeyword(&c, "trample"))
	}
}

// TestEndoskeletonNeverStartsOnceUntapped is CR 611.2b's "never
// starts": the source untapped in response, so nothing is registered.
func TestEndoskeletonNeverStartsOnceUntapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	endo := effectstest.PushCatalogPermanent(g, me.ID, "Endoskeleton", "Artifact", effectstest.RtEndoskeletonOracle, false)
	bear := effectstest.CtrlPushCreature(g, me.ID, "Bear")
	effectstest.B06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, endo, 0, game.ActivateAbilityParams{Targets: effectstest.LtCardTarget(bear)}); err != nil {
		t.Fatal(err)
	}
	effectstest.RtUntap(g, endo)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LayeredCard(t, g, bear).CurrentToughness(); got != 2 {
		t.Errorf("toughness %d, want 2: the Endoskeleton was untapped as its ability resolved", got)
	}
}

// TestManaLeechHoldsTheLandWhileItRemainsTapped records Rust Tick's
// hold on the land.
func TestManaLeechHoldsTheLandWhileItRemainsTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	leech := effectstest.PushCatalogPermanent(g, me.ID, "Mana Leech", "Creature — Leech", effectstest.RtManaLeechOracle, false)
	land := effectstest.PushLandFor(g, opp.ID, "Forest", "Basic Land — Forest")
	effectstest.RtActivate(t, g, me.ID, leech, 0, effectstest.LtCardTarget(land))
	c := effectstest.LayeredCard(t, g, land)
	if !c.Tapped || len(c.NextUntapSkips) != 1 || c.NextUntapSkips[0].While == nil ||
		c.NextUntapSkips[0].While.Condition != game.WhileSourceRemainsTapped {
		t.Fatalf("land: tapped %v, holds %+v; want tapped and held while the Leech remains tapped", c.Tapped, c.NextUntapSkips)
	}
}

// TestKillSwitchTapsEveryOtherArtifact: the set is every other
// artifact as the ability resolves, each held.
func TestKillSwitchTapsEveryOtherArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ks := effectstest.PushCatalogPermanent(g, me.ID, "Kill Switch", "Artifact", effectstest.RtKillSwitchOracle, false)
	mine := pushTappedPermanentForTest(g, me.ID, "Sol Ring", "Artifact")
	theirs := pushTappedPermanentForTest(g, opp.ID, "Mind Stone", "Artifact")
	bear := effectstest.CtrlPushCreature(g, opp.ID, "Bear")
	effectstest.RtActivate(t, g, me.ID, ks, 0, nil, "C", "C")
	for _, id := range []uuid.UUID{mine, theirs} {
		if c := effectstest.LayeredCard(t, g, id); !c.Tapped || len(c.NextUntapSkips) != 1 {
			t.Errorf("%s: tapped %v, holds %d; want tapped and held", c.Name, c.Tapped, len(c.NextUntapSkips))
		}
	}
	if effectstest.RtTapped(t, g, bear) {
		t.Error("Kill Switch tapped a creature")
	}
	if len(effectstest.LayeredCard(t, g, ks).NextUntapSkips) != 0 {
		t.Error("Kill Switch held itself")
	}
}

// pushTappedPermanentForTest seeds an untapped permanent with the
// entry event, so it is stamped.
func pushTappedPermanentForTest(g *game.Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
		Owner: owner, Controller: owner})
}

// TestZygonInfiltratorIsACopyWhileThatCreatureRemainsTapped is the
// pinned condition: the copy is about the creature copied. The stun
// counter spends its first untap.
func TestZygonInfiltratorIsACopyWhileThatCreatureRemainsTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	zygon := effectstest.PushCatalogPermanent(g, me.ID, "Zygon Infiltrator", "Creature — Alien Shapeshifter Soldier", effectstest.RtZygonOracle, false)
	model := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Hill Giant",
		TypeLine: "Creature — Giant", Power: 3, Toughness: 3, Owner: opp.ID, Controller: opp.ID})
	effectstest.RtActivate(t, g, me.ID, zygon, 0, effectstest.LtCardTarget(model), "U", "C", "C")
	if got := effectstest.LayeredCard(t, g, zygon).Name; got != "Hill Giant" {
		t.Fatalf("Zygon is %q, want a copy of Hill Giant", got)
	}
	m := effectstest.LayeredCard(t, g, model)
	if !m.Tapped || m.Counters[game.CounterStun] != 1 {
		t.Fatalf("model: tapped %v, stun %d; want tapped with one stun counter", m.Tapped, m.Counters[game.CounterStun])
	}
	effectstest.RtUntap(g, model) // the stun counter goes instead (CR 122.1d)
	if got := effectstest.LayeredCard(t, g, zygon).Name; got != "Hill Giant" {
		t.Fatalf("after the stun counter went: Zygon is %q, want still Hill Giant", got)
	}
	effectstest.RtUntap(g, model)
	if got := effectstest.LayeredCard(t, g, zygon).Name; got != "Zygon Infiltrator" {
		t.Errorf("after the model untapped: Zygon is %q, want itself again", got)
	}
}

// TestHisokasGuardCannotTargetItself is "other than this creature".
func TestHisokasGuardCannotTargetItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	guard := effectstest.PushCatalogPermanent(g, me.ID, "Hisoka's Guard", "Creature — Human Wizard", effectstest.RtHisokasGuardOracle, false)
	effectstest.B06AddMana(me, "U", "C")
	if err := g.ActivateCatalogAbility(me.ID, guard, 0, game.ActivateAbilityParams{Targets: effectstest.LtCardTarget(guard)}); err == nil {
		t.Error("Hisoka's Guard targeted itself")
	}
}

// TestBlackstaffAnimatesAnArtifactWhileTapped: a 4/4 artifact creature
// that keeps its own types.
func TestBlackstaffAnimatesAnArtifactWhileTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[0]
	staff := effectstest.PushCatalogPermanent(g, me.ID, "The Blackstaff of Waterdeep", "Legendary Artifact", effectstest.RtBlackstaffOracle, false)
	rock := pushTappedPermanentForTest(g, me.ID, "Mind Stone", "Artifact")
	effectstest.RtActivate(t, g, me.ID, staff, 0, effectstest.LtCardTarget(rock), "U", "C")
	c := effectstest.LayeredCard(t, g, rock)
	if !c.IsCreature() || !c.IsArtifact() || c.CurrentPower() != 4 || c.CurrentToughness() != 4 {
		t.Fatalf("Mind Stone: creature %v artifact %v %d/%d; want a 4/4 artifact creature",
			c.IsCreature(), c.IsArtifact(), c.CurrentPower(), c.CurrentToughness())
	}
	effectstest.RtUntap(g, staff)
	if effectstest.LayeredCard(t, g, rock).IsCreature() {
		t.Error("the Blackstaff untapped and the artifact is still a creature")
	}
}

// TestIceFloeTargetsOnlyAttackersOfYou is its target clause.
func TestIceFloeTargetsOnlyAttackersOfYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	floe := effectstest.PushCatalogPermanent(g, me.ID, "Ice Floe", "Land", effectstest.RtIceFloeOracle, false)
	idle := effectstest.CtrlPushCreature(g, opp.ID, "Idle Bear")
	if err := g.ActivateCatalogAbility(me.ID, floe, 0, game.ActivateAbilityParams{Targets: effectstest.LtCardTarget(idle)}); err == nil {
		t.Error("Ice Floe targeted a creature that is not attacking")
	}
	pred := effects.And(effects.WithoutKeyword("flying"), effects.AttackingYou())
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
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	forest := effectstest.PushLandFor(g, opp.ID, "Forest", "Basic Land — Forest")
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Merfolk",
		TypeLine: "Creature — Merfolk", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	hand := len(me.Hand.Cards)
	effectstest.CastCatalogSpell(t, g, "Aquitect's Will", "Kindred Sorcery — Merfolk", effectstest.RtAquitectsWillOracle, effectstest.LtCardTarget(forest))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectiveSubtypes(t, g, forest); !typeListHasForTest(got, "Island") || !typeListHasForTest(got, "Forest") {
		t.Fatalf("subtypes %v, want Forest and Island", got)
	}
	if got := len(me.Hand.Cards); got != hand+1 {
		t.Errorf("hand %d, want %d: a Merfolk under you draws a card", got, hand+1)
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(forest, "flood", -1) })
	if got := effectstest.EffectiveSubtypes(t, g, forest); typeListHasForTest(got, "Island") {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	morgul := effectstest.PushCatalogPermanent(g, me.ID, "Minas Morgul, Dark Fortress", "Legendary Land", effectstest.RtMinasMorgulOracle, false)
	bear := effectstest.CtrlPushCreature(g, me.ID, "Bear")
	effectstest.RtActivate(t, g, me.ID, morgul, 0, effectstest.LtCardTarget(bear), "B", "C", "C", "C")
	c := effectstest.LayeredCard(t, g, bear)
	if !game.HasKeyword(&c, "shadow") || !typeListHasForTest(effectstest.EffectiveSubtypes(t, g, bear), "Wraith") {
		t.Fatalf("bear: shadow %v, subtypes %v; want shadow and Wraith", game.HasKeyword(&c, "shadow"), effectstest.EffectiveSubtypes(t, g, bear))
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, game.CounterShadow, -1) })
	if typeListHasForTest(effectstest.EffectiveSubtypes(t, g, bear), "Wraith") {
		t.Error("the shadow counter went and the creature is still a Wraith")
	}
}

// TestMakeshiftMannequinSacrificesWhenTargeted: the creature returns
// with its counter and the granted trigger.
func TestMakeshiftMannequinSacrificesWhenTargeted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dead := effectstest.PushGraveyardCardTyped(me, "Bear", "Creature — Bear")
	effectstest.CastCatalogSpell(t, g, "Makeshift Mannequin", "Instant", effectstest.RtMannequinOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: dead}})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LayeredCard(t, g, dead)
	if c.Counters["mannequin"] != 1 {
		t.Fatalf("returned with counters %v, want a mannequin counter", c.Counters)
	}
	endo := effectstest.PushCatalogPermanent(g, me.ID, "Endoskeleton", "Artifact", effectstest.RtEndoskeletonOracle, false)
	effectstest.RtActivate(t, g, me.ID, endo, 0, effectstest.LtCardTarget(dead), "C", "C")
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, dead) {
		t.Error("the returned creature was targeted and is still on the battlefield")
	}
}

// TestMathasBountyRewardsTheOpponents: the dies trigger is the bounty
// creature's, so "each opponent" is its controller's opponents.
func TestMathasBountyRewardsTheOpponents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Mathas, Fiend Seeker", "Legendary Creature — Vampire", effectstest.RtMathasOracle, false)
	victim := effectstest.CtrlPushCreature(g, opp.ID, "Bear")
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LayeredCard(t, g, victim).Counters["bounty"]; got != 1 {
		t.Fatalf("bounty counters %d, want 1", got)
	}
	effectstest.AdvanceToMain(t, g)
	hand, life := len(me.Hand.Cards), me.Life
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(victim) })
	effectstest.PassPriorityAroundTable(t, g)
	if len(me.Hand.Cards) != hand+1 || me.Life != life+2 {
		t.Errorf("after the bounty creature died: hand %d (was %d), life %d (was %d); want +1 and +2",
			len(me.Hand.Cards), hand, me.Life, life)
	}
}

// TestObsidianFireheartGrantsTheBlaze: the land carries the upkeep
// trigger for as long as it has a blaze counter, and may not be chosen
// again while it does.
func TestObsidianFireheartGrantsTheBlaze(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fire := effectstest.PushCatalogPermanent(g, me.ID, "Obsidian Fireheart", "Creature — Elemental", effectstest.RtFireheartOracle, false)
	land := effectstest.PushLandFor(g, opp.ID, "Plains", "Basic Land — Plains")
	effectstest.RtActivate(t, g, me.ID, fire, 0, effectstest.LtCardTarget(land), "R", "R", "C")
	c := effectstest.LayeredCard(t, g, land)
	if c.Counters["blaze"] != 1 || len(c.Effective().GrantedAbilities) == 0 {
		t.Fatalf("land: counters %v, granted %v; want a blaze counter and the grant", c.Counters, c.Effective().GrantedAbilities)
	}
	effectstest.B06AddMana(me, "R", "R", "C")
	if err := g.ActivateCatalogAbility(me.ID, fire, 0, game.ActivateAbilityParams{Targets: effectstest.LtCardTarget(land)}); err == nil {
		t.Error("the Fireheart targeted a land that already has a blaze counter")
	}
}

// TestXolatoyacFloodsALandAsItEnters is Aquitect's Will's flood on an
// enters trigger.
func TestXolatoyacFloodsALandAsItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[0]
	forest := effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.CastAndResolveCreature(t, g, "Xolatoyac, the Smiling Flood", "Legendary Creature — Salamander Serpent", effectstest.RtXolatoyacOracle)
	effectstest.PickCard(t, g, me.ID, forest)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectiveSubtypes(t, g, forest); !typeListHasForTest(got, "Island") {
		t.Errorf("subtypes %v, want an Island in addition", got)
	}
}
