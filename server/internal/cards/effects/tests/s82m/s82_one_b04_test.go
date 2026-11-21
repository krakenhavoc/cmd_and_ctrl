package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	s82ShepherdOracle    = "929c67aa-49ff-4401-85e8-76990fe1a369"
	s82ChorusOracle      = "99ba0d4f-b2bf-48ea-95d6-b6b85ee077ab"
	s82HivemasterOracle  = "6b1f750c-0f01-4b1c-ad8f-260e26fbbfea"
	s82AttendantOracle   = "9587bdb3-fb6a-47c0-898d-404bf60fcb6b"
	s82ChargeOracle      = "fa47c190-d66a-41b0-aa08-21036cda4223"
	s82FleshcutterOracle = "e9c03ed6-34dd-46f7-ab41-8004348328cc"
	s82OverseerOracle    = "d92411b3-cb25-4123-a78e-4f785dfdae0c"
	s82VishgrazOracle    = "7b1599c0-bb4f-48f8-a358-dc60166b13e7"
	s82PrologueOracle    = "3f661b58-e8ca-412c-bdac-af151162de12"
	s82InquiryOracle     = "66051196-53ea-4461-b241-205df9395b3b"
	s82VraskasFallOracle = "33fd9192-3e9b-4d3a-92aa-06de29d449df"
	s82InfectBiteOracle  = "7b740cd6-e07d-4273-8a7f-e084bdaf0f1c"
	s82OutbreakOracle    = "47148b5d-3850-4614-822e-6ef2755996f8"
)

func s82Mites(g *game.Game, owner uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Phyrexian Mite" && c.Controller == owner {
				out = append(out, c.InstanceID)
			}
		}
	})
	return out
}

func s82Others(g *game.Game, me uuid.UUID) []*game.Player {
	var out []*game.Player
	for _, p := range g.Seats {
		if p.ID != me {
			out = append(out, p)
		}
	}
	return out
}

func TestBasilicaShepherdMakesTwoMites(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Basilica Shepherd", "Creature — Phyrexian Angel", s82ShepherdOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	mites := s82Mites(g, me.ID)
	if len(mites) != 2 {
		t.Fatalf("%d Mites, want 2", len(mites))
	}
	for _, id := range mites {
		if effectstest.ToxicOn(g, id) != 1 {
			t.Error("a Mite lacks toxic 1")
		}
	}
}

func TestToxicDiesMiteCreatures(t *testing.T) {
	for _, tc := range []struct{ name, oracle string }{
		{"Crawling Chorus", s82ChorusOracle}, {"Stinging Hivemaster", s82HivemasterOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			id := effectstest.PushDiesCreatureForTest(g, me.ID, tc.name, tc.oracle, "Creature — Phyrexian", 1, 1)
			if effectstest.ToxicOn(g, id) != 1 {
				t.Fatal("no printed toxic 1")
			}
			effectstest.KillCreature(t, g, me.ID, id)
			effectstest.PassPriorityAroundTable(t, g)
			if n := len(s82Mites(g, me.ID)); n != 1 {
				t.Fatalf("%d Mites after death, want 1", n)
			}
		})
	}
}

func TestMiteCantBlockAndDealsPoison(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.CastCatalogSpell(t, g, "Basilica Shepherd", "Creature — Phyrexian Angel", s82ShepherdOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	mites := s82Mites(g, me.ID)
	if len(mites) == 0 {
		t.Fatal("no Mites")
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			g.Battlefield.Cards[i].SummonedThisTurn = false
			g.Battlefield.Cards[i].EnteredBattlefieldAt = 0
		}
	})
	effectstest.AttackWith(t, g, opp.ID, mites[0])
	if effectstest.PoisonOn(opp) != 1 {
		t.Errorf("poison %d, want 1", effectstest.PoisonOn(opp))
	}
}

func TestChargeOfTheMitesModes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.CastCatalogSpellWithModes(t, g, "Charge of the Mites", "Instant", s82ChargeOracle, []int{1})
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(s82Mites(g, me.ID)); n != 2 {
		t.Fatalf("%d Mites, want 2", n)
	}
	// Mode 0: three creatures (two Mites + one more) -> 3 damage.
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	victim := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: "Charge of the Mites", TypeLine: "Instant", OracleID: s82ChargeOracle, Owner: active.ID, Controller: active.ID})
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{Modes: []int{0}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == victim && c.DamageMarked != 3 {
				t.Errorf("damage %d, want 3", c.DamageMarked)
			}
		}
	})
}

func TestInfestedFleshcutterPumpsAndMakesMiteOnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	cutter := effectstest.PushCatalogPermanent(g, me.ID, "Infested Fleshcutter", "Artifact — Equipment", s82FleshcutterOracle, false)
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(cutter, game.TargetRef{Kind: game.TargetCard, ID: bear}); err != nil {
			t.Fatal(err)
		}
	})
	if p, _ := effectstest.EffectivePTOf(t, g, bear); p != 4 {
		t.Fatalf("power %d, want 4", p)
	}
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(s82Mites(g, me.ID)); n != 1 {
		t.Fatalf("%d Mites, want 1", n)
	}
}

func TestMiteOverseerBuffsTokensOnlyOnYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushDiesCreatureForTest(g, me.ID, "Mite Overseer", s82OverseerOracle, "Creature — Phyrexian Soldier", 4, 2)
	effectstest.CastCatalogSpell(t, g, "Basilica Shepherd", "Creature — Phyrexian Angel", s82ShepherdOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	mite := s82Mites(g, me.ID)[0]
	if p, _ := effectstest.EffectivePTOf(t, g, mite); p != 2 {
		t.Errorf("Mite power %d on my turn, want 2", p)
	}
	if !effectstest.HasEffectiveKeyword(t, g, mite, "first strike") {
		t.Error("Mite lacks first strike on my turn")
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if p, _ := effectstest.EffectivePTOf(t, g, mite); p != 1 {
		t.Errorf("Mite power %d on their turn, want 1", p)
	}
	if effectstest.HasEffectiveKeyword(t, g, mite, "first strike") {
		t.Error("Mite has first strike on their turn")
	}
}

func TestVishgrazGrowsWithOpponentPoison(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	vish := effectstest.PushDiesCreatureForTest(g, me.ID, "Vishgraz, the Doomhive", s82VishgrazOracle, "Legendary Creature — Phyrexian Insect", 3, 3)
	opps := s82Others(g, me.ID)
	g.WithWriteLock(func() {
		_ = g.AddPlayerCounterByForEffect(me.ID, opps[0].ID, game.CounterPoison, 2)
		_ = g.AddPlayerCounterByForEffect(me.ID, opps[1].ID, game.CounterPoison, 3)
		_ = g.AddPlayerCounterByForEffect(opps[0].ID, me.ID, game.CounterPoison, 4)
	})
	if p, tt := effectstest.EffectivePTOf(t, g, vish); p != 8 || tt != 8 {
		t.Errorf("Vishgraz %d/%d, want 8/8", p, tt)
	}
}

func TestVishgrazMakesThreeMites(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Vishgraz, the Doomhive", "Legendary Creature — Phyrexian Insect", s82VishgrazOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(s82Mites(g, me.ID)); n != 3 {
		t.Fatalf("%d Mites, want 3", n)
	}
}

func TestPoisonSpells(t *testing.T) {
	t.Run("Prologue to Phyresis", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		hand := me.Hand.Size()
		effectstest.CastCatalogSpell(t, g, "Prologue to Phyresis", "Instant", s82PrologueOracle, nil)
		effectstest.PassPriorityAroundTable(t, g)
		for _, o := range s82Others(g, me.ID) {
			if effectstest.PoisonOn(o) != 1 {
				t.Errorf("%s poison %d", o.Name, effectstest.PoisonOn(o))
			}
		}
		if effectstest.PoisonOn(me) != 0 || me.Hand.Size() != hand+1 { // the cast card left, one was drawn
			t.Errorf("me: poison %d hand %d (was %d, want +1)", effectstest.PoisonOn(me), me.Hand.Size(), hand)
		}
	})
	t.Run("Infectious Inquiry", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		life := me.Life
		hand := me.Hand.Size()
		effectstest.CastCatalogSpell(t, g, "Infectious Inquiry", "Sorcery", s82InquiryOracle, nil)
		effectstest.PassPriorityAroundTable(t, g)
		if me.Life != life-2 || me.Hand.Size() != hand+2 {
			t.Errorf("life %d (want %d), hand %d (want %d)", me.Life, life-2, me.Hand.Size(), hand+2)
		}
		for _, o := range s82Others(g, me.ID) {
			if effectstest.PoisonOn(o) != 1 {
				t.Errorf("%s poison %d", o.Name, effectstest.PoisonOn(o))
			}
		}
	})
	t.Run("Phyresis Outbreak", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		opps := s82Others(g, me.ID)
		g.WithWriteLock(func() { _ = g.AddPlayerCounterByForEffect(me.ID, opps[0].ID, game.CounterPoison, 1) })
		a := effectstest.B12Creature(g, opps[0].ID, "A", "Creature — Bear", 5, 5)
		b := effectstest.B12Creature(g, opps[1].ID, "B", "Creature — Bear", 5, 5)
		mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 5, 5)
		effectstest.CastCatalogSpell(t, g, "Phyresis Outbreak", "Sorcery", s82OutbreakOracle, nil)
		effectstest.PassPriorityAroundTable(t, g)
		for id, want := range map[uuid.UUID]int{a: 3, b: 4, mine: 5} {
			if p, _ := effectstest.EffectivePTOf(t, g, id); p != want {
				t.Errorf("power %d, want %d", p, want)
			}
		}
	})
	t.Run("Infectious Bite", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 3, 3)
		foe := effectstest.B12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 2)
		effectstest.CastCatalogSpell(t, g, "Infectious Bite", "Instant", s82InfectBiteOracle,
			[]game.TargetRef{{Kind: game.TargetCard, ID: bear}, {Kind: game.TargetCard, ID: foe}})
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.OnBattlefieldNow(g, foe) {
			t.Error("the Foe survived")
		}
		for _, o := range s82Others(g, me.ID) {
			if effectstest.PoisonOn(o) != 1 {
				t.Errorf("%s poison %d", o.Name, effectstest.PoisonOn(o))
			}
		}
		// Strict: a creature you control can't be the second target.
		err := effectstest.CastCatalogSpellErr(t, g, "Infectious Bite", "Instant", s82InfectBiteOracle,
			[]game.TargetRef{{Kind: game.TargetCard, ID: bear}, {Kind: game.TargetCard, ID: bear}})
		if err == nil {
			t.Error("targeting my own creature as the victim was legal")
		}
	})
	t.Run("Vraska's Fall", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		opps := s82Others(g, me.ID)
		keep := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
		lone := effectstest.B12Creature(g, opps[0].ID, "Lone", "Creature — Bear", 2, 2)
		effectstest.CastCatalogSpell(t, g, "Vraska's Fall", "Instant", s82VraskasFallOracle, nil)
		effectstest.PassPriorityAroundTable(t, g)
		for _, o := range opps {
			effectstest.AnswerSacrifice(t, g, o.ID)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.OnBattlefieldNow(g, lone) || !effectstest.OnBattlefieldNow(g, keep) {
			t.Error("the edict hit the wrong board")
		}
		for _, o := range opps {
			if effectstest.PoisonOn(o) != 1 {
				t.Errorf("%s poison %d, want 1 (even with nothing to sacrifice)", o.Name, effectstest.PoisonOn(o))
			}
		}
	})
}

func TestIndoctrinationAttendantBouncesForAMite(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	other := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Indoctrination Attendant", "Creature — Phyrexian Cleric", s82AttendantOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	answerOwnPerm(t, g, me.ID, other)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OnBattlefieldNow(g, other) {
		t.Error("the Bear stayed")
	}
	if n := len(s82Mites(g, me.ID)); n != 1 {
		t.Errorf("%d Mites, want 1", n)
	}
}

func TestIndoctrinationAttendantDecliningMakesNoMite(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	other := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Indoctrination Attendant", "Creature — Phyrexian Cleric", s82AttendantOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	answerOwnPerm(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.OnBattlefieldNow(g, other) || len(s82Mites(g, me.ID)) != 0 {
		t.Error("declining still bounced or made a Mite")
	}
}

func answerOwnPerm(t *testing.T, g *game.Game, chooser uuid.UUID, picks ...uuid.UUID) {
	t.Helper()
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, chooser)
	if pick == nil {
		t.Fatal("no permanent choice is open")
	}
	if err := g.ResolveOwnPermanents(pick.ID, chooser, picks); err != nil {
		t.Fatal(err)
	}
}
