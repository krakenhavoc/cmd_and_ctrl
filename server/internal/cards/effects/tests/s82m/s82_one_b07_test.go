package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 set ONE, batch 7 (#5479): the corrupted commons and uncommons.

const (
	oneSphereOracle     = "15679808-baba-4a11-b111-f313ce24197f"
	oneAtlasOracle      = "b85d54e9-4f21-4440-9fd4-7b73b560c99d"
	oneGliderOracle     = "e51f89c8-3d7f-4f1a-bc35-67b92db20e10"
	oneDancerOracle     = "a605864a-f46f-4ebc-8e56-ce7322c918b0"
	oneEvangelistOracle = "007f3b20-d58b-4986-8de8-9e419af88628"
	oneApostleOracle    = "d56f1543-2b2d-4ce3-bcfe-41224a0cece4"
	oneSkirgeOracle     = "c42833e6-3e42-4bd6-acbc-d8e284e9833e"
	oneSkitterOracle    = "864fc52a-e9f4-4254-901e-b949d96b34eb"
	oneGladiatorOracle  = "8e33d6e5-30a2-48b4-826c-09aa75bbba27"
	oneTitanOracle      = "90109cb4-1e16-4d54-9332-b5fb1e71042c"
	oneConvictionOracle = "79ab207f-e7ff-4236-b260-5a109a73fbed"
	oneAnointOracle     = "402c3afe-7041-41a3-8fcd-64a060c12d4a"
	oneEndingOracle     = "57b37690-0fb6-4e78-9210-4a529a6ade7c"
)

func oneSetPoison(g *game.Game, player uuid.UUID, n int) {
	g.WithWriteLock(func() {
		p := g.PlayerByIDForEffect(player)
		if cur := p.Poison; cur != n {
			_ = g.AddPlayerCounterForEffect(player, game.CounterPoison, n-cur)
		}
	})
}

func onePush(g *game.Game, owner uuid.UUID, name, typeLine, oracle, cost string, p, t int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: cost,
		Power: p, Toughness: t, Owner: owner, Controller: owner,
	})
}

func TestOneB07CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oneSphereOracle: "Glistening Sphere", oneAtlasOracle: "Phyrexian Atlas", oneGliderOracle: "Incisor Glider",
		oneDancerOracle: "Sinew Dancer", oneEvangelistOracle: "Vivisection Evangelist", oneApostleOracle: "Apostle of Invasion",
		oneSkirgeOracle: "Bonepicker Skirge", oneSkitterOracle: "Chittering Skitterling", oneGladiatorOracle: "Fleshless Gladiator",
		oneTitanOracle: "Ravenous Necrotitan", oneConvictionOracle: "Zealot's Conviction", oneAnointOracle: "Anoint with Affliction",
		oneEndingOracle: "Bring the Ending",
	} {
		def, ok := effects.Lookup(oracle)
		if !ok || def.Name != name || def.Completeness != effects.CompletenessFull {
			t.Errorf("%s: registered=%v def=%+v", name, ok, def)
		}
	}
}

func TestOneGlisteningSphereMakesThreeManaOnlyWhenCorrupted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	sphere := onePush(g, me.ID, "Glistening Sphere", "Artifact", oneSphereOracle, "{3}", 0, 0)
	oneSetPoison(g, opp.ID, 2)
	if err := g.ActivateManaAbility(me.ID, sphere, 1, game.ManaAbilityParams{}); err == nil {
		t.Fatal("the corrupted mana ability worked with only two poison counters")
	}
	oneSetPoison(g, opp.ID, 3)
	if err := g.ActivateManaAbility(me.ID, sphere, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("corrupted mana ability: %v", err)
	}
	if c := effectstest.RiderLatestManaPick(g, me.ID); c != nil {
		if err := g.ResolveManaChoice(c.ID, me.ID, "G"); err != nil {
			t.Fatal(err)
		}
	}
	if got := me.ManaPool.Total(); got != 3 {
		t.Errorf("pool = %d, want 3", got)
	}
}

func TestOneGlisteningSphereEntersTappedAndProliferates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oneSetPoison(g, opp.ID, 1)
	id := effectstest.CastCatalogSpell(t, g, "Glistening Sphere", "Artifact", oneSphereOracle, nil)
	effectstest.SettleAnsweringProliferate(t, g)
	c, ok := effectstest.BattlefieldCard(g, id)
	if !ok || !c.Tapped {
		t.Fatalf("the Sphere should enter tapped (on battlefield: %v)", ok)
	}
	_ = me
	if opp.Poison != 2 {
		t.Errorf("poison = %d, want 2 after proliferate", opp.Poison)
	}
}

func TestOnePhyrexianAtlasDrainsOnlyCorruptedOpponentsWhenTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.AdvanceToMain(t, g)
	atlas := onePush(g, me.ID, "Phyrexian Atlas", "Artifact", oneAtlasOracle, "{3}", 0, 0)
	oneSetPoison(g, a.ID, 3)
	oneSetPoison(g, b.ID, 2)
	la, lb, lme := a.Life, b.Life, me.Life
	effectstest.TapForMana(t, g, me.ID, atlas)
	if c := effectstest.RiderLatestManaPick(g, me.ID); c != nil {
		_ = g.ResolveManaChoice(c.ID, me.ID, "G")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if a.Life != la-1 {
		t.Errorf("corrupted opponent life = %d, want %d", a.Life, la-1)
	}
	if b.Life != lb || me.Life != lme {
		t.Errorf("uncorrupted opponent %d (was %d) or controller %d (was %d) lost life", b.Life, lb, me.Life, lme)
	}
}

func TestOneIncisorGliderPumpsTheTeamOnlyWhenCorrupted(t *testing.T) {
	for _, poison := range []int{2, 3} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.AdvanceToMain(t, g)
		glider := onePush(g, me.ID, "Incisor Glider", "Artifact Creature — Phyrexian Construct", oneGliderOracle, "{1}{W}", 1, 3)
		bear := onePush(g, me.ID, "Bear", "Creature — Bear", "", "", 2, 2)
		oneSetPoison(g, opp.ID, poison)
		effectstest.DeclareAttack(t, g, opp.ID, glider)
		effectstest.PassPriorityAroundTable(t, g)
		want := 2
		if poison == 3 {
			want = 3
		}
		if got := effectstest.EffectivePower(t, g, bear); got != want {
			t.Errorf("poison %d: Bear power = %d, want %d", poison, got, want)
		}
	}
}

func TestOneSinewDancerCheapTapNeedsCorruption(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dancer := onePush(g, me.ID, "Sinew Dancer", "Creature — Phyrexian Soldier", oneDancerOracle, "{W}", 1, 1)
	victim := onePush(g, opp.ID, "Bear", "Creature — Bear", "", "", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "W")
	effectstest.AcRefused(t, g, me.ID, dancer, 1, game.ActivateAbilityParams{Targets: effectstest.CardRefs(victim)})
	oneSetPoison(g, opp.ID, 3)
	effectstest.B16Activate(t, g, me.ID, dancer, 1, game.ActivateAbilityParams{Targets: effectstest.CardRefs(victim)})
	if !effectstest.B16Tapped(t, g, victim) {
		t.Error("the target creature should be tapped")
	}
}

func TestOneSinewDancerFullPriceTapWorksUncorrupted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dancer := onePush(g, me.ID, "Sinew Dancer", "Creature — Phyrexian Soldier", oneDancerOracle, "{W}", 1, 1)
	victim := onePush(g, opp.ID, "Bear", "Creature — Bear", "", "", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "W", "C", "C", "C")
	effectstest.B16Activate(t, g, me.ID, dancer, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(victim)})
	if !effectstest.B16Tapped(t, g, victim) {
		t.Error("the target creature should be tapped")
	}
}

func TestOneVivisectionEvangelistDestroysOnlyWhenCorrupted(t *testing.T) {
	for _, poison := range []int{2, 3} {
		g := effectstest.NewCatalogGame(t)
		opp := g.Seats[1]
		victim := onePush(g, opp.ID, "Bear", "Creature — Bear", "", "", 2, 2)
		oneSetPoison(g, opp.ID, poison)
		effectstest.CastCatalogSpell(t, g, "Vivisection Evangelist", "Creature — Phyrexian Cleric", oneEvangelistOracle, nil)
		effectstest.PassPriorityAroundTable(t, g)
		if poison == 3 {
			effectstest.PickTargetRefs(t, g, g.Seats[0].ID, nil, game.TargetRef{Kind: game.TargetCard, ID: victim})
			effectstest.PassPriorityAroundTable(t, g)
		} else if effectstest.LatestPickTarget(g, g.Seats[0].ID) != nil {
			t.Error("uncorrupted: the trigger should not ask for a target")
		}
		alive := g.Battlefield.Contains(victim)
		if poison == 2 && !alive {
			t.Error("uncorrupted: the Bear should live")
		}
		if poison == 3 && alive {
			t.Error("corrupted: the Bear should be destroyed")
		}
	}
}

func TestOneApostleAndSkirgeGainKeywordsOnlyWhileCorrupted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	apostle := onePush(g, me.ID, "Apostle of Invasion", "Creature — Phyrexian Angel", oneApostleOracle, "{4}{W}{W}", 4, 4)
	skirge := onePush(g, me.ID, "Bonepicker Skirge", "Creature — Phyrexian Imp", oneSkirgeOracle, "{2}{B}", 2, 2)
	check := func(corrupted bool) {
		t.Helper()
		if effectstest.HasKeywordOf(g, apostle, "double strike") != corrupted {
			t.Errorf("corrupted=%v: Apostle double strike = %v", corrupted, !corrupted)
		}
		for _, kw := range []string{"deathtouch", "lifelink"} {
			if effectstest.HasKeywordOf(g, skirge, kw) != corrupted {
				t.Errorf("corrupted=%v: Skirge %s = %v", corrupted, kw, !corrupted)
			}
		}
	}
	oneSetPoison(g, opp.ID, 2)
	check(false)
	oneSetPoison(g, opp.ID, 3)
	check(true)
	oneSetPoison(g, opp.ID, 1)
	check(false)
	// Your own poison does not count: "an opponent".
	oneSetPoison(g, me.ID, 5)
	check(false)
}

func TestOneChitteringSkitterlingDrawsOncePerTurnWhenCorrupted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	skit := onePush(g, me.ID, "Chittering Skitterling", "Creature — Phyrexian Rat", oneSkitterOracle, "{2}{B}", 1, 4)
	fodder := onePush(g, me.ID, "Fodder", "Creature — Bear", "", "", 1, 1)
	fodder2 := onePush(g, me.ID, "Fodder Two", "Artifact", "", "", 0, 0)
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedLibrary(me, "A", "B", "C")
	effectstest.AcRefused(t, g, me.ID, skit, 0, game.ActivateAbilityParams{})
	oneSetPoison(g, opp.ID, 3)
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, skit, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed creature is still on the battlefield")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand = %d, want %d", me.Hand.Size(), hand+1)
	}
	_ = fodder2
	effectstest.AcRefused(t, g, me.ID, skit, 0, game.ActivateAbilityParams{})
}

func TestOneFleshlessGladiatorReturnsTappedAndCostsLifeOnlyWhenCorrupted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Fleshless Gladiator", TypeLine: "Creature — Phyrexian Skeleton",
		OracleID: oneGladiatorOracle, ManaCost: "{1}{B}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	effectstest.B06AddMana(me, "B", "C", "C")
	effectstest.AcRefused(t, g, me.ID, id, 0, game.ActivateAbilityParams{})
	oneSetPoison(g, opp.ID, 3)
	life := me.Life
	effectstest.B16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{})
	c, ok := effectstest.BattlefieldCard(g, id)
	if !ok || !c.Tapped {
		t.Fatalf("Gladiator should be on the battlefield tapped (present=%v)", ok)
	}
	if me.Life != life-1 {
		t.Errorf("life = %d, want %d", me.Life, life-1)
	}
}

func TestOneRavenousNecrotitanSacrificesUnlessCorrupted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	other := onePush(g, me.ID, "Bear", "Creature — Bear", "", "", 2, 2)
	id := effectstest.CastCatalogSpell(t, g, "Ravenous Necrotitan", "Creature — Phyrexian Horror", oneTitanOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, me.ID, other)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(other) || !g.Battlefield.Contains(id) {
		t.Errorf("uncorrupted: Bear gone=%v, Necrotitan present=%v", !g.Battlefield.Contains(other), g.Battlefield.Contains(id))
	}

	g2 := effectstest.NewCatalogGame(t)
	opp2 := g2.Seats[1]
	me2 := g2.Seats[0]
	other2 := onePush(g2, me2.ID, "Bear", "Creature — Bear", "", "", 2, 2)
	oneSetPoison(g2, opp2.ID, 3)
	id2 := effectstest.CastCatalogSpell(t, g2, "Ravenous Necrotitan", "Creature — Phyrexian Horror", oneTitanOracle, nil)
	effectstest.PassPriorityAroundTable(t, g2)
	if effectstest.SacrificeChoiceFor(g2, me2.ID) != nil || !g2.Battlefield.Contains(other2) || !g2.Battlefield.Contains(id2) {
		t.Error("corrupted: nothing is sacrificed")
	}
	_ = opp
}

func TestOneZealotsConvictionPumpsMoreWhileCorrupted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := onePush(g, me.ID, "Bear", "Creature — Bear", "", "", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Zealot's Conviction", "Enchantment — Aura", oneConvictionOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 3 || tt != 3 {
		t.Errorf("uncorrupted: %d/%d, want 3/3", p, tt)
	}
	if effectstest.HasKeywordOf(g, bear, "first strike") {
		t.Error("uncorrupted: first strike")
	}
	oneSetPoison(g, opp.ID, 3)
	if p, tt := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 4 || tt != 3 {
		t.Errorf("corrupted: %d/%d, want 4/3", p, tt)
	}
	if !effectstest.HasKeywordOf(g, bear, "first strike") {
		t.Error("corrupted: no first strike")
	}
}

func TestOneAnointWithAfflictionExilesByManaValueOrCorruptedController(t *testing.T) {
	type tc struct {
		name   string
		cost   string
		poison int
		exiled bool
	}
	for _, c := range []tc{
		{"cheap", "{2}", 0, true},
		{"expensive", "{4}", 0, false},
		{"expensive poisoned controller", "{4}", 3, true},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			opp := g.Seats[1]
			v := onePush(g, opp.ID, "Target", "Creature — Bear", "", c.cost, 2, 2)
			oneSetPoison(g, opp.ID, c.poison)
			effectstest.CastCatalogSpell(t, g, "Anoint with Affliction", "Instant", oneAnointOracle, effectstest.CardTarget(v))
			effectstest.PassPriorityAroundTable(t, g)
			if gone := !g.Battlefield.Contains(v); gone != c.exiled {
				t.Errorf("exiled = %v, want %v", gone, c.exiled)
			}
			if c.exiled && !g.Exile.Contains(v) {
				t.Error("the creature should be in exile")
			}
		})
	}
}

func TestOneAnointWithAfflictionIgnoresTheCastersOwnPoison(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	v := onePush(g, opp.ID, "Target", "Creature — Bear", "", "{4}", 2, 2)
	oneSetPoison(g, me.ID, 5)
	effectstest.CastCatalogSpell(t, g, "Anoint with Affliction", "Instant", oneAnointOracle, effectstest.CardTarget(v))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(v) {
		t.Error("the caster's poison is not the target's controller's")
	}
}

func TestOneBringTheEndingCountersOutrightWhenControllerIsCorrupted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	oneSetPoison(g, me.ID, 3)
	spell := effectstest.CastCatalogSpell(t, g, "A Bear", "Creature — Bear", "", nil)
	effectstest.CastCatalogSpell(t, g, "Bring the Ending", "Instant", oneEndingOracle, effectstest.CardTarget(spell))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("a corrupted controller is not offered the payment")
	}
	if g.Battlefield.Contains(spell) || g.Stack.Contains(spell) {
		t.Error("the spell should be countered")
	}
}

func TestOneBringTheEndingTaxesTwoOtherwise(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	spell := effectstest.CastCatalogSpell(t, g, "A Bear", "Creature — Bear", "", nil)
	effectstest.CastCatalogSpell(t, g, "Bring the Ending", "Instant", oneEndingOracle, effectstest.CardTarget(spell))
	effectstest.PassUntilTaxed(t, g, me.ID)
	if c := effectstest.AnswerPayUnless(t, g, me.ID, false); c == nil || c.PayCost != "{2}" {
		t.Fatalf("pay prompt = %+v, want {2}", c)
	}
	if g.Stack.Contains(spell) {
		t.Error("the unpaid spell should be countered")
	}
}
