package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	gapShriekingDrakeOracle  = "2a8e14be-6f93-428d-a204-e2399dc93797"
	gapMessengerOracle       = "f1409c9e-f0d9-49e1-b02e-d0f04a658331"
	gapDrownyardLurkerOracle = "7cf5c3dd-5c07-4eaf-8227-6f4e7f9b134d"
	gapMeddlerOracle         = "bd9b3009-f389-4831-b250-80c4bb8f25e5"
	gapCoatlOracle           = "57cf576b-5a03-4c60-a531-544df2638323"
	gapGuardianOracle        = "89b88a6c-92be-43eb-9de4-f79cc75c6d90"
	gapEnvoyOracle           = "678ef5e4-800d-4d3a-8761-2abaad70e0c9"
	gapSageOracle            = "359a570e-92a7-4954-9a05-89cc75df57e7"
)

// gapCastCreature puts a creature card with printed power and toughness
// and no mana cost in `me`'s hand and casts it, resolving the spell.
func gapCastCreature(t *testing.T, g *game.Game, me *game.Player, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	effectstest.ToMainForCost(t, g)
	id := effectstest.PushHandCardWithManaCost(me, name, typeLine, "")
	gapB54StampHandCard(me, id, oracle, power, toughness)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return id
}

func gapFaceDownCount(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.FaceDownIsPermanent() {
			n++
		}
	}
	return n
}

func TestGapShriekingDrakeReturnsACreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	drake := gapCastCreature(t, g, me, "Shrieking Drake", "Creature — Drake", gapShriekingDrakeOracle, 1, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !g.Battlefield.Contains(drake) {
		t.Errorf("bear on battlefield %v, drake on battlefield %v: want the Bear returned", g.Battlefield.Contains(bear), g.Battlefield.Contains(drake))
	}
	effectstest.AssertKeywords(t, g, drake, "flying")
}

func TestGapShriekingDrakeMayReturnItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	drake := gapCastCreature(t, g, me, "Shrieking Drake", "Creature — Drake", gapShriekingDrakeOracle, 1, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID, drake)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(drake) {
		t.Error("the Drake did not return itself")
	}
}

func TestGapEmrakulsMessengerMakesASpawnOnTheSecondDrawOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Emrakul's Messenger", "Creature — Eldrazi Faerie Rogue", gapMessengerOracle, false)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	draws := func() {
		g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 1) })
		effectstest.PassPriorityAroundTable(t, g)
	}
	before := g.TurnTallyFor(me.ID).CardsDrawn
	for i := before; i < 1; i++ {
		draws()
	}
	if n := effectstest.TokensNamedControlledBy(g, "Eldrazi Spawn", me.ID); n != 0 {
		t.Fatalf("after the first draw: %d Spawn, want 0", n)
	}
	draws()
	if n := effectstest.TokensNamedControlledBy(g, "Eldrazi Spawn", me.ID); n != 1 {
		t.Fatalf("after the second draw: %d Spawn, want 1", n)
	}
	draws()
	if n := effectstest.TokensNamedControlledBy(g, "Eldrazi Spawn", me.ID); n != 1 {
		t.Errorf("after the third draw: %d Spawn, want 1", n)
	}
}

func TestGapDrownyardLurkerMakesASpawnWhenCastAndWhenCycled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	lurker := effectstest.PushHandCardWithManaCost(me, "Drownyard Lurker", "Creature — Eldrazi Trilobite", "")
	gapB54StampHandCard(me, lurker, gapDrownyardLurkerOracle, 7, 7)
	if err := g.CastSpell(me.ID, lurker, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.TokensNamedControlledBy(g, "Eldrazi Spawn", me.ID); n != 1 {
		t.Fatalf("cast: %d Spawn, want 1", n)
	}

	cycler := effectstest.PushCatalogHandCard(me, "Drownyard Lurker", "Creature — Eldrazi Trilobite", gapDrownyardLurkerOracle)
	if err := g.AddManaForEffect(me.ID, me.ID, "{U}{U}{U}"); err != nil {
		t.Fatalf("mana: %v", err)
	}
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, cycler, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("cycle: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.TokensNamedControlledBy(g, "Eldrazi Spawn", me.ID); n != 2 {
		t.Errorf("cycled: %d Spawn, want 2", n)
	}
	if me.Hand.Size() != hand {
		// one card cycled away, one drawn
		t.Errorf("hand %d -> %d after cycling", hand, me.Hand.Size())
	}
}

func TestGapPetrifyingMeddlerTapsAndStunsAtCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	effectstest.ToMainForCost(t, g)
	meddler := effectstest.PushHandCardWithManaCost(me, "Petrifying Meddler", "Creature — Eldrazi", "")
	gapB54StampHandCard(me, meddler, gapMeddlerOracle, 4, 5)
	if err := g.CastSpell(me.ID, meddler, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.AnswerTriggerTargets(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	c, _ := effectstest.BattlefieldCard(g, victim)
	if !c.Tapped || effectstest.CounterCount(g, victim, game.CounterStun) != 1 {
		t.Errorf("tapped %v, stun counters %d: want tapped with one stun counter", c.Tapped, effectstest.CounterCount(g, victim, game.CounterStun))
	}
	effectstest.AssertKeywords(t, g, meddler, "reach", "devoid")
}

func TestGapHopeEnderCoatlCountersASpellUnlessTheyPay(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		life := me.Life
		victim := effectstest.CastFor(t, g, opp, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, game.CastSpellParams{Targets: []game.TargetRef{effectstest.PlayerTarget(me.ID)}})
		coatl := effectstest.PushHandCardWithManaCost(me, "Hope-Ender Coatl", "Creature — Eldrazi Snake", "")
		gapB54StampHandCard(me, coatl, gapCoatlOracle, 2, 2)
		if err := g.CastSpell(me.ID, coatl, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.AnswerTriggerTargets(t, g, me.ID, victim)
		effectstest.PassPriorityAroundTable(t, g)
		if pay {
			if err := g.AddManaForEffect(opp.ID, opp.ID, "{C}"); err != nil {
				t.Fatalf("mana: %v", err)
			}
		}
		effectstest.AnswerPayUnless(t, g, opp.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		resolved := me.Life == life-3
		if resolved != pay {
			t.Errorf("pay=%v: the Bolt resolved=%v (life %d -> %d)", pay, resolved, life, me.Life)
		}
		if !g.Battlefield.Contains(coatl) {
			t.Errorf("pay=%v: the Coatl did not resolve", pay)
		}
	}
}

func TestGapGuardianManifestsWhenAModifiedCreatureYouControlDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Guardian of the Forgotten", "Creature — Elephant Warrior", gapGuardianOracle, 4, 4)
	plain := effectstest.PushVanillaCreature(g, me.ID, "Plain Bear", 2, 2)
	modified := effectstest.PushVanillaCreature(g, me.ID, "Counter Bear", 2, 2)
	effectstest.PutCounters(t, g, modified, game.CounterPlusOne, 1)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	effectstest.PutCounters(t, g, theirs, game.CounterPlusOne, 1)

	effectstest.B27Kill(g, plain)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B27Kill(g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if n := gapFaceDownCount(g, me.ID); n != 0 {
		t.Fatalf("an unmodified creature of mine or a modified one of theirs died: %d manifests, want 0", n)
	}
	top := me.Library.Cards[len(me.Library.Cards)-1].InstanceID
	if got := me.Library.Cards[0].InstanceID; got == top && len(me.Library.Cards) > 1 {
		t.Log("library orientation: top is the last slice element")
	}
	libBefore := len(me.Library.Cards)
	effectstest.B27Kill(g, modified)
	effectstest.PassPriorityAroundTable(t, g)
	if n := gapFaceDownCount(g, me.ID); n != 1 {
		t.Fatalf("a modified creature died: %d manifests, want 1", n)
	}
	if len(me.Library.Cards) != libBefore-1 {
		t.Errorf("library %d -> %d, want one card manifested", libBefore, len(me.Library.Cards))
	}
}

func TestGapEnvoyGivesLifelinkToModifiedCreaturesAndOutlasts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	envoy := effectstest.PushCatalogPermanent(g, me.ID, "Envoy of the Ancestors", "Creature — Human Cleric", gapEnvoyOracle, false)
	plain := effectstest.PushVanillaCreature(g, me.ID, "Plain Bear", 2, 2)
	modified := effectstest.PushVanillaCreature(g, me.ID, "Counter Bear", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	effectstest.PutCounters(t, g, modified, game.CounterPlusOne, 1)
	effectstest.PutCounters(t, g, theirs, game.CounterPlusOne, 1)
	if effectstest.HasEffectiveKeyword(t, g, plain, "lifelink") || effectstest.HasEffectiveKeyword(t, g, envoy, "lifelink") {
		t.Error("an unmodified creature has lifelink")
	}
	if !effectstest.HasEffectiveKeyword(t, g, modified, "lifelink") {
		t.Error("a modified creature I control lacks lifelink")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "lifelink") {
		t.Error("an opponent's modified creature has lifelink")
	}
	// Outlast {W}: sorcery speed, taps, a counter modifies the Envoy itself.
	effectstest.ToMainForCost(t, g)
	if err := g.AddManaForEffect(me.ID, me.ID, "{W}"); err != nil {
		t.Fatalf("mana: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, envoy, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("outlast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, envoy, "lifelink") {
		t.Error("the Envoy with a counter on it is modified and should have lifelink")
	}
}

// gapB54StampHandCard stamps an oracle ID and printed power and toughness
// onto a card already in a hand.
func gapB54StampHandCard(p *game.Player, id uuid.UUID, oracle string, power, toughness int) {
	for i := range p.Hand.Cards {
		if p.Hand.Cards[i].InstanceID == id {
			p.Hand.Cards[i].OracleID = oracle
			p.Hand.Cards[i].Power, p.Hand.Cards[i].Toughness = power, toughness
		}
	}
}

func TestGapSageOfTheUnknowableManaPaysColorlessSpellsAndAbilitiesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sage := effectstest.PushCatalogPermanent(g, me.ID, "Sage of the Unknowable", "Creature — Human Wizard", gapSageOracle, false)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, sage, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap the Sage: %v", err)
	}
	cost, _ := game.ParseCost("{C}")
	rock := game.Card{Name: "Rock", TypeLine: "Artifact", ManaCost: "{1}"}
	bear := game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Colors: []string{"G"}}
	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(rock)) {
		t.Error("refused a colorless spell")
	}
	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(bear)) {
		t.Error("refused an ability")
	}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(bear)) {
		t.Error("paid for a coloured spell")
	}
}

const (
	gapGlyphElementalOracle  = "288ea954-a880-4731-8fe7-93d2ff6a043f"
	gapTritonWavebreakerOrac = "d2309fba-4273-4084-b206-642a43975c34"
)

func TestGapGlyphElementalBestowedGrowsItsHostWithEachLandfall(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	glyph := effectstest.BestowCard(t, g, "Glyph Elemental", "Enchantment Creature — Elemental", gapGlyphElementalOracle, 2, 2)
	effectstest.CastBestowed(t, g, glyph, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if s := effectstest.ReadBestowState(g, glyph); !s.OnBattlefield || s.Creature || !s.Aura || s.AttachedTo != bear {
		t.Fatalf("bestowed state %+v, want an Aura attached to the Bear and not a creature", s)
	}
	if p := effectstest.EffectivePowerWithCounters(t, g, bear); p != 2 {
		t.Fatalf("host power %d with no counters on the Aura, want 2", p)
	}
	effectstest.B31PlayLand(t, g, "Forest", "Basic Land — Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, glyph, game.CounterPlusOne); got != 1 {
		t.Fatalf("Aura counters %d after landfall, want 1", got)
	}
	if p := effectstest.EffectivePowerWithCounters(t, g, bear); p != 3 {
		t.Errorf("host power %d, want 3 (+1/+1 for the counter on the Aura)", p)
	}
	// The host leaves: the Aura is a creature again, 2/2 with its counter.
	effectstest.B27Kill(g, bear)
	g.RunStateChecksForTest()
	if s := effectstest.ReadBestowState(g, glyph); !s.Creature || !s.OnBattlefield {
		t.Fatalf("after the host left: %+v, want a creature on the battlefield", s)
	}
	if p := effectstest.EffectivePowerWithCounters(t, g, glyph); p != 3 {
		t.Errorf("unattached power %d, want 3 (2/2 with one counter)", p)
	}
}

func TestGapGlyphElementalIsA2By2WithLandfallWhenCastNormally(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	glyph := effectstest.B12Push(g, me.ID, "Glyph Elemental", "Enchantment Creature — Elemental", gapGlyphElementalOracle, 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.B31PlayLand(t, g, "Forest", "Basic Land — Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, glyph, game.CounterPlusOne); got != 1 {
		t.Errorf("counters %d, want 1", got)
	}
}

func TestGapTritonWavebreakerGivesItsHostPlusOneAndProwessAndHasNoneItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	tri := effectstest.BestowCard(t, g, "Triton Wavebreaker", "Enchantment Creature — Merfolk Wizard", gapTritonWavebreakerOrac, 1, 1)
	effectstest.CastBestowed(t, g, tri, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, bear); p != 3 {
		t.Errorf("host power %d, want 3", p)
	}
	effectstest.AssertKeywords(t, g, bear, "prowess")
	if effectstest.HasEffectiveKeyword(t, g, tri, "prowess") {
		t.Error("the bestowed Aura has prowess though it is not a creature")
	}
	// Unattached it is a creature again and has prowess.
	effectstest.B27Kill(g, bear)
	g.RunStateChecksForTest()
	effectstest.AssertKeywords(t, g, tri, "prowess")
	if p := effectstest.EffectivePower(t, g, tri); p != 1 {
		t.Errorf("unattached power %d, want 1", p)
	}
}

const gapVolatileStormdrakeOracle = "08d64b16-9221-4532-890c-d5cdbf181e28"

func TestGapVolatileStormdrakeRefusesOpponentsAbilitiesButNotSpellsOrYours(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	drake := effectstest.B27Push(g, me.ID, "Volatile Stormdrake", "Creature — Drake", gapVolatileStormdrakeOracle, "{1}{U}", 3, 2, "U")
	theirBallista := effectstest.B27Push(g, opp.ID, "Walking Ballista", "Artifact Creature — Construct", effectstest.WalkingBallistaOracle, "{X}{X}", 0, 0)
	myBallista := effectstest.B27Push(g, me.ID, "Walking Ballista", "Artifact Creature — Construct", effectstest.WalkingBallistaOracle, "{X}{X}", 0, 0)
	for _, id := range []uuid.UUID{theirBallista, myBallista} {
		effectstest.SetCountersOn(g, id, map[string]int{game.CounterPlusOne: 2})
	}
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(opp.ID, theirBallista, 1, game.ActivateAbilityParams{Targets: effectstest.CardRefs(drake), Strict: true}); err == nil {
		t.Error("an opponent's ability targeted the Stormdrake")
	}
	if err := g.ActivateCatalogAbility(me.ID, myBallista, 1, game.ActivateAbilityParams{Targets: effectstest.CardRefs(drake), Strict: true}); err != nil {
		t.Errorf("my own ability could not target the Stormdrake: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	stoke := effectstest.PutInHand(opp, game.Card{Name: "Stoke the Flames", TypeLine: "Instant", OracleID: effectstest.RprB03StokeOracle, Owner: opp.ID, Controller: opp.ID})
	if err := g.CastSpell(opp.ID, stoke, game.CastSpellParams{Targets: effectstest.CardTarget(drake), Strict: true}); err != nil {
		t.Errorf("an opponent's spell could not target the Stormdrake: %v", err)
	}
	effectstest.AssertKeywords(t, g, drake, "flying")
}

// gapStormdrakeEnter casts the Stormdrake with `victim` as the target of
// its enters trigger and returns the Stormdrake's id.
func gapStormdrakeEnter(t *testing.T, g *game.Game, me *game.Player, victim uuid.UUID) uuid.UUID {
	t.Helper()
	drake := gapCastCreature(t, g, me, "Volatile Stormdrake", "Creature — Drake", gapVolatileStormdrakeOracle, 3, 2)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerTriggerTargets(t, g, me.ID, victim)
	return drake
}

func TestGapVolatileStormdrakeExchangesThenKeepsTheCreatureForEnergy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SetEnergy(t, g, me, 0)
	victim := effectstest.B27Push(g, opp.ID, "Their Bear", "Creature — Bear", "", "{1}{G}", 2, 2, "G")
	drake := gapStormdrakeEnter(t, g, me, victim)
	if n := effectstest.SettleEnergy(t, g, me.ID, effectstest.ErPlay{Pay: true}); n != 1 {
		t.Fatalf("answered %d energy prompts, want 1", n)
	}
	if got := effectstest.ControllerOf(t, g, victim); got != me.ID {
		t.Errorf("the Bear is controlled by %s, want me", got)
	}
	if got := effectstest.ControllerOf(t, g, drake); got != opp.ID {
		t.Errorf("the Stormdrake is controlled by %s, want the opponent", got)
	}
	if effectstest.EnergyOf(me) != 2 {
		t.Errorf("energy %d, want 2 (four gained, two paid)", effectstest.EnergyOf(me))
	}
}

func TestGapVolatileStormdrakeSacrificesThatCreatureWhenYouDeclineOrCannotPay(t *testing.T) {
	for _, tc := range []struct {
		name string
		cost string
		pay  bool
	}{
		{"declined", "{1}{G}", false},
		{"cannot fund", "{4}{G}{G}", true},
	} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.SetEnergy(t, g, me, 0)
		victim := effectstest.B27Push(g, opp.ID, "Their Big Bear", "Creature — Bear", "", tc.cost, 2, 2, "G")
		gapStormdrakeEnter(t, g, me, victim)
		effectstest.SettleEnergy(t, g, me.ID, effectstest.ErPlay{Pay: tc.pay})
		if g.Battlefield.Contains(victim) {
			t.Errorf("%s: the creature was not sacrificed", tc.name)
		}
		if effectstest.EnergyOf(me) != 4 {
			t.Errorf("%s: energy %d, want 4", tc.name, effectstest.EnergyOf(me))
		}
	}
}

func TestGapVolatileStormdrakeIsPlainWhenItsTargetIsGone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SetEnergy(t, g, me, 0)
	victim := effectstest.B27Push(g, opp.ID, "Their Bear", "Creature — Bear", "", "{1}{G}", 2, 2, "G")
	drake := gapStormdrakeEnter(t, g, me, victim)
	effectstest.B27Kill(g, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.ControllerOf(t, g, drake); got != me.ID {
		t.Errorf("no exchange was possible but the Stormdrake is controlled by %s", got)
	}
	if effectstest.EnergyOf(me) != 0 {
		t.Errorf("energy %d, want 0 with no exchange", effectstest.EnergyOf(me))
	}
}
