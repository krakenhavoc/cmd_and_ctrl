package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	s82NeoReitoSentinelOracle     = "e574bfdb-505b-4d61-a376-6d26f03cd9b1"
	s82NeoSevenTailMentorOracle   = "301b8265-9e17-48c3-942d-b7243112b394"
	s82NeoPapercraftDecoyOracle   = "b0494c34-f142-441f-b7f8-e981416f0bad"
	s82NeoGravelighterOracle      = "45ef5ea9-70a6-4999-8778-6a11cb97d28e"
	s82NeoBearerOfMemoryOracle    = "51e76d8f-95c6-428f-9030-3f8718139f93"
	s82NeoGeothermalKamiOracle    = "fbf8fdf5-0d7d-41a3-a6b9-27bb7a644c20"
	s82NeoNetworkDisruptorOracle  = "0d64d150-1716-4da4-9cfe-ee7749b89918"
	s82NeoKamiRestlessOracle      = "7e770c08-910e-4394-88c5-79c874750b6a"
	s82NeoUnstoppableOgreOracle   = "1733777d-87fe-4975-b099-1ff803a721e8"
	s82NeoPatchworkAutomatonOracl = "0ba86a50-13df-4b26-8b3a-9e3917ff850f"
	s82NeoGenerousVisitorOracle   = "0629747e-8dd8-4150-95ee-9d860264ccb0"
	s82NeoBlossomPrancerOracle    = "69e92c4f-df05-4297-a453-8e4644c31a90"
	s82NeoKamiOfIndustryOracle    = "b308fa9a-3f69-4e59-9d4b-8007d9a2b1f1"
)

func TestS82NeoB03EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s82NeoReitoSentinelOracle:     "Reito Sentinel",
		s82NeoSevenTailMentorOracle:   "Seven-Tail Mentor",
		s82NeoPapercraftDecoyOracle:   "Papercraft Decoy",
		s82NeoGravelighterOracle:      "Gravelighter",
		s82NeoBearerOfMemoryOracle:    "Bearer of Memory",
		s82NeoGeothermalKamiOracle:    "Geothermal Kami",
		s82NeoNetworkDisruptorOracle:  "Network Disruptor",
		s82NeoKamiRestlessOracle:      "Kami of Restless Shadows",
		s82NeoUnstoppableOgreOracle:   "Unstoppable Ogre",
		s82NeoPatchworkAutomatonOracl: "Patchwork Automaton",
		s82NeoGenerousVisitorOracle:   "Generous Visitor",
		s82NeoBlossomPrancerOracle:    "Blossom Prancer",
		s82NeoKamiOfIndustryOracle:    "Kami of Industry",
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

// s82NeoPickPlayer answers the open target prompt with a player.
func s82NeoPickPlayer(t *testing.T, g *game.Game, chooser, player uuid.UUID) {
	t.Helper()
	p := effectstest.LatestPickTarget(g, chooser)
	if p == nil {
		t.Fatal("no target prompt is open")
	}
	if err := g.ResolvePickTarget(p.ID, chooser, game.TargetRef{Kind: game.TargetPlayer, ID: player}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
}

func s82NeoPay(p *game.Player, colors ...string) {
	for _, c := range colors {
		p.ManaPool.AddMana(game.ManaToken{Color: c})
	}
}

func s82NeoGraveCard(p *game.Player, name, typeLine, manaCost string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: manaCost, Owner: p.ID, Controller: p.ID})
	return id
}

func TestS82NeoB03ReitoSentinelMillsThreeAndTucksAGraveyardCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(opp, "A", "B", "C", "D", "E")
	before := opp.Library.Size()
	sentinel := effectstest.CastAndResolveCreature(t, g, "Reito Sentinel", "Artifact Creature — Construct", s82NeoReitoSentinelOracle)
	s82NeoPickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := before - opp.Library.Size(); got != 3 {
		t.Errorf("the target player milled %d, want 3", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, sentinel, "defender") {
		t.Error("no defender")
	}
	card := s82NeoGraveCard(opp, "Old Bear", "Creature — Bear", "{1}{G}")
	if err := g.ActivateCatalogAbility(me.ID, sentinel, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(card)}); err == nil {
		t.Fatal("activated {3} with an empty pool")
	}
	s82NeoPay(me, "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, sentinel, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(card)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Graveyard.Contains(card) {
		t.Error("the card stayed in the graveyard")
	}
	if b, err := opp.Library.Bottom(); err != nil || b.InstanceID != card {
		t.Errorf("the card is not on the bottom of its owner's library")
	}
}

func TestS82NeoB03SevenTailMentorCountersACreatureOrVehicleOnEnterAndDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ride := effectstest.B12Permanent(g, me.ID, "Ride", "Artifact — Vehicle")
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	mentor := effectstest.CastAndResolveCreature(t, g, "Seven-Tail Mentor", "Creature — Fox Samurai", s82NeoSevenTailMentorOracle)
	if err := g.ResolvePickTarget(effectstest.LatestPickTarget(g, me.ID).ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: rock}); err == nil {
		t.Error("a noncreature, non-Vehicle artifact was a legal target")
	}
	effectstest.AnswerPickTargetIfOpen(t, g, ride)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, ride, game.CounterPlusOne); n != 1 {
		t.Errorf("the Vehicle has %d counters after the enter trigger, want 1", n)
	}
	effectstest.KillCreature(t, g, me.ID, mentor)
	effectstest.AnswerPickTargetIfOpen(t, g, ride)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, ride, game.CounterPlusOne); n != 2 {
		t.Errorf("the Vehicle has %d counters after the dies trigger, want 2", n)
	}
}

func TestS82NeoB03PapercraftDecoyDrawsOnlyIfYouPayWhenItLeaves(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.SeedLibrary(me, "A", "B", "C")
		decoy := effectstest.B12Push(g, me.ID, "Papercraft Decoy", "Artifact Creature — Frog", s82NeoPapercraftDecoyOracle, 2, 1)
		hand := me.Hand.Size()
		effectstest.KillCreature(t, g, me.ID, decoy)
		effectstest.PassPriorityAroundTable(t, g)
		s82NeoPay(me, "C", "C")
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		want := hand
		if pay {
			want++
		}
		if me.Hand.Size() != want {
			t.Errorf("pay=%v: hand %d, want %d", pay, me.Hand.Size(), want)
		}
		if pay && me.ManaPool.Total() != 0 {
			t.Errorf("the {2} was not spent")
		}
	}
}

func TestS82NeoB03GravelighterDrawsIfACreatureDiedElseEveryoneSacrifices(t *testing.T) {
	t.Run("a creature died", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.SeedLibrary(me, "A", "B", "C")
		fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
		effectstest.KillCreature(t, g, me.ID, fodder)
		keeper := effectstest.B12Creature(g, me.ID, "Keeper", "Creature — Bear", 2, 2)
		hand := me.Hand.Size()
		effectstest.CastAndResolveCreature(t, g, "Gravelighter", "Creature — Spirit", s82NeoGravelighterOracle)
		effectstest.PassPriorityAroundTable(t, g)
		if me.Hand.Size() != hand {
			// the cast moved the card out of hand during setup: only the draw adds one
			t.Logf("hand %d -> %d", hand, me.Hand.Size())
		}
		if !g.Battlefield.Contains(keeper) {
			t.Error("a creature was sacrificed although one had died this turn")
		}
		if me.Hand.Size() != hand+1 {
			t.Errorf("hand %d, want %d: one draw", me.Hand.Size(), hand+1)
		}
	})
	t.Run("none died", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.SeedLibrary(me, "A", "B", "C")
		mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
		theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
		hand := me.Hand.Size()
		light := effectstest.CastAndResolveCreature(t, g, "Gravelighter", "Creature — Spirit", s82NeoGravelighterOracle)
		effectstest.PassPriorityAroundTable(t, g)
		if c := effectstest.SacrificeChoiceFor(g, me.ID); c != nil {
			effectstest.AnswerSacrifice(t, g, me.ID, mine)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(theirs) {
			t.Error("the opponent kept their only creature")
		}
		if g.Battlefield.Contains(mine) {
			t.Error("my chosen creature was not sacrificed")
		}
		if !g.Battlefield.Contains(light) {
			t.Error("Gravelighter was sacrificed although another creature was chosen")
		}
		if me.Hand.Size() != hand {
			t.Errorf("hand %d, want %d: no draw", me.Hand.Size(), hand)
		}
	})
}

func TestS82NeoB03BearerOfMemoryCountersAnEnchantmentCreatureAndGivesTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bearer := effectstest.B12Push(g, me.ID, "Bearer of Memory", "Enchantment Creature — Human Monk", s82NeoBearerOfMemoryOracle, 3, 2)
	plain := effectstest.B12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	s82NeoPay(me, "G", "C", "C", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, bearer, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(plain)}); err == nil {
		t.Fatal("targeted a creature that is not an enchantment")
	}
	if err := g.ActivateCatalogAbility(me.ID, bearer, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bearer)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, bearer, game.CounterPlusOne); n != 1 {
		t.Errorf("%d counters, want 1", n)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bearer, "trample") {
		t.Error("no trample")
	}
}

func TestS82NeoB03GeothermalKamiMayReturnAnEnchantmentForThreeLife(t *testing.T) {
	for _, take := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		ench := effectstest.B12Permanent(g, me.ID, "Ench", "Enchantment")
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		life := me.Life
		effectstest.CastAndResolveCreature(t, g, "Geothermal Kami", "Creature — Spirit", s82NeoGeothermalKamiOracle)
		effectstest.PassPriorityAroundTable(t, g)
		pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
		if pick == nil {
			t.Fatalf("take=%v: no choice", take)
		}
		if pick.ChooseMin != 0 || pick.ChooseMax != 1 || effectstest.HasID(pick.ChooseCards, bear) || !effectstest.HasID(pick.ChooseCards, ench) {
			t.Errorf("take=%v: bounds %d..%d offered %v", take, pick.ChooseMin, pick.ChooseMax, pick.ChooseCards)
		}
		var picks []uuid.UUID
		if take {
			picks = []uuid.UUID{ench}
		}
		if err := g.ResolveOwnPermanents(pick.ID, me.ID, picks); err != nil {
			t.Fatalf("ResolveOwnPermanents: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if take {
			if !me.Hand.Contains(ench) || me.Life != life+3 {
				t.Errorf("returned=%v life %d, want %d", me.Hand.Contains(ench), me.Life, life+3)
			}
		} else if !g.Battlefield.Contains(ench) || me.Life != life {
			t.Errorf("declined: enchantment on battlefield %v, life %d", g.Battlefield.Contains(ench), me.Life)
		}
	}
}

func TestS82NeoB03NetworkDisruptorTapsATargetPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	disruptor := effectstest.CastAndResolveCreature(t, g, "Network Disruptor", "Artifact Creature — Moonfolk Rogue", s82NeoNetworkDisruptorOracle)
	effectstest.AnswerPickTargetIfOpen(t, g, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, rock) {
		t.Error("the target was not tapped")
	}
	if !effectstest.HasEffectiveKeyword(t, g, disruptor, "flying") {
		t.Error("no flying")
	}
	_ = me
}

func TestS82NeoB03KamiOfRestlessShadowsModes(t *testing.T) {
	setup := func(t *testing.T, mode int) (*game.Game, *game.Player, uuid.UUID, uuid.UUID) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		ninja := s82NeoGraveCard(me, "Ninja", "Creature — Human Ninja", "{1}")
		bear := s82NeoGraveCard(me, "Bear", "Creature — Bear", "{1}")
		effectstest.CastAndResolveCreature(t, g, "Kami of Restless Shadows", "Creature — Spirit", s82NeoKamiRestlessOracle)
		c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
		if c == nil {
			t.Fatalf("no mode prompt: %+v", g.PendingChoices)
		}
		if err := g.ResolveModePick(c.ID, me.ID, []int{mode}); err != nil {
			t.Fatalf("ResolveModePick: %v", err)
		}
		return g, me, ninja, bear
	}
	t.Run("return a Ninja or Rogue", func(t *testing.T) {
		g, me, ninja, bear := setup(t, 0)
		if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
			if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err == nil {
				t.Error("a non-Ninja non-Rogue creature card was a legal target")
			}
		}
		effectstest.AnswerPickTargetIfOpen(t, g, ninja)
		effectstest.PassPriorityAroundTable(t, g)
		if !me.Hand.Contains(ninja) || me.Hand.Contains(bear) {
			t.Error("the Ninja did not return (or the Bear did)")
		}
	})
	t.Run("put a creature on top", func(t *testing.T) {
		g, me, _, bear := setup(t, 1)
		effectstest.AnswerPickTargetIfOpen(t, g, bear)
		effectstest.PassPriorityAroundTable(t, g)
		if top, err := me.Library.Top(); err != nil || top.InstanceID != bear {
			t.Error("the creature card is not on top of the library")
		}
	})
}

func TestS82NeoB03UnstoppableOgreStopsACreatureBlocking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, opp.ID, "Other", "Creature — Bear", 2, 2)
	effectstest.CastAndResolveCreature(t, g, "Unstoppable Ogre", "Artifact Creature — Ogre Warrior", s82NeoUnstoppableOgreOracle)
	effectstest.AnswerPickTargetIfOpen(t, g, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if !s82NeoCantBlock(t, g, victim) {
		t.Error("the target can still block")
	}
	if s82NeoCantBlock(t, g, other) {
		t.Error("an untargeted creature can't block")
	}
}

func TestS82NeoB03PatchworkAutomatonWardsAndGrowsOnArtifactSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	auto := effectstest.B12Push(g, me.ID, "Patchwork Automaton", "Artifact Creature — Construct", s82NeoPatchworkAutomatonOracl, 1, 1)
	effectstest.CastCatalogSpell(t, g, "Shock", "Instant", effectstest.B5ShockOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, auto, game.CounterPlusOne); n != 0 {
		t.Fatalf("an instant gave %d counters", n)
	}
	effectstest.CastCatalogSpell(t, g, "Rock", "Artifact", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, auto, game.CounterPlusOne); n != 1 {
		t.Errorf("an artifact spell gave %d counters, want 1", n)
	}
}

func TestS82NeoB03GenerousVisitorCountersAnyCreatureOnEnchantmentSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Generous Visitor", "Creature — Spirit", s82NeoGenerousVisitorOracle, 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Rock", "Artifact", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, theirs, game.CounterPlusOne); n != 0 {
		t.Fatalf("an artifact spell triggered it")
	}
	effectstest.CastCatalogSpell(t, g, "Aura Thing", "Enchantment", "", nil)
	effectstest.AnswerPickTargetIfOpen(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, theirs, game.CounterPlusOne); n != 1 {
		t.Errorf("the target has %d counters, want 1", n)
	}
}

func TestS82NeoB03BlossomPrancerTakesACardOrGainsFour(t *testing.T) {
	t.Run("takes an enchantment", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.SeedLibrary(me, "A", "B", "C", "D", "E", "F")
		ench := effectstest.TopLibraryCard(me, "Ench", "Enchantment")
		life := me.Life
		effectstest.CastAndResolveCreature(t, g, "Blossom Prancer", "Creature — Spirit", s82NeoBlossomPrancerOracle)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerChooseCards(t, g, me.ID, ench)
		effectstest.PassPriorityAroundTable(t, g)
		if !me.Hand.Contains(ench) || me.Life != life {
			t.Errorf("in hand %v, life %d (want %d)", me.Hand.Contains(ench), me.Life, life)
		}
	})
	t.Run("nothing taken", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.SeedLibrary(me, "A", "B", "C", "D", "E", "F")
		life := me.Life
		lib := me.Library.Size()
		effectstest.CastAndResolveCreature(t, g, "Blossom Prancer", "Creature — Spirit", s82NeoBlossomPrancerOracle)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		if me.Life != life+4 {
			t.Errorf("life %d, want %d", me.Life, life+4)
		}
		if me.Library.Size() != lib {
			t.Errorf("library %d, want %d: the rest goes to the bottom", me.Library.Size(), lib)
		}
	})
	t.Run("declined", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.SeedLibrary(me, "A", "B", "C", "D", "E")
		effectstest.TopLibraryCard(me, "Bear", "Creature — Bear")
		life := me.Life
		effectstest.CastAndResolveCreature(t, g, "Blossom Prancer", "Creature — Spirit", s82NeoBlossomPrancerOracle)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerChooseCards(t, g, me.ID)
		effectstest.PassPriorityAroundTable(t, g)
		if me.Life != life+4 {
			t.Errorf("life %d, want %d", me.Life, life+4)
		}
	})
}

func TestS82NeoB03KamiOfIndustryReturnsAnArtifactWithHasteAndSacrificesItAtTheEnd(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cheap := s82NeoGraveCard(me, "Cheap Rock", "Artifact", "{3}")
	pricey := s82NeoGraveCard(me, "Pricey Rock", "Artifact", "{4}")
	effectstest.CastAndResolveCreature(t, g, "Kami of Industry", "Creature — Spirit", s82NeoKamiOfIndustryOracle)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: pricey}); err == nil {
			t.Error("a mana value 4 artifact was a legal target")
		}
	}
	effectstest.AnswerPickTargetIfOpen(t, g, cheap)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(cheap) {
		t.Fatal("the artifact did not return")
	}
	if !effectstest.HasEffectiveKeyword(t, g, cheap, "haste") {
		t.Error("no haste")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(cheap) {
		t.Error("the artifact was not sacrificed at the next end step")
	}
}

func s82NeoCantBlock(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	return effectstest.AuraRestrictions(t, g, id)&game.CantBlock != 0
}

func TestS82NeoB03PatchworkAutomatonWardCountersAnOpponentsSpellUnlessTheyPayTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	auto := effectstest.B12Push(g, me.ID, "Patchwork Automaton", "Artifact Creature — Construct", s82NeoPatchworkAutomatonOracl, 1, 1)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	blade := effectstest.CastAtWardedCreature(t, g, opp, auto)
	for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Fatal("no ward payment prompt for the spell's controller")
	}
	effectstest.AnswerPayUnless(t, g, opp.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(auto) || !opp.Graveyard.Contains(blade) {
		t.Error("an unpaid ward {2} must counter the spell")
	}
}
