package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b459_cards_test.go — New Phyrexia, batch 459 (S82, #5575).
// Each test plays the real catalog entry through the engine.

const (
	b459Verdict   = "0f575f18-4606-4b85-9a90-07f8d8d46d06"
	b459WarReport = "45dfb358-cd16-468e-bc89-ff17c6e244c1"
	b459Cell      = "56c558f3-535d-49aa-a2de-54e0bb598483"
	b459Invasion  = "88bbbc00-0704-4dc0-897f-93184b2e1eed"
	b459Dross     = "8ce2a750-1137-4f8d-803c-9aac30db4193"
	b459Dogs      = "a89994d0-2110-4ca0-abfc-5dd232981acf"
	b459Conduit   = "b48befc5-a9f2-4a64-a16d-c883b034ea9c"
	b459Parasite  = "c180f21b-f553-4839-80c7-a3ce2fae9b5a"
	b459Blessing  = "ea393815-0202-4edc-aa1e-6885fe9b20ab"
)

func TestB459EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		b459Verdict: "Geth's Verdict", b459WarReport: "War Report", b459Cell: "Isolation Cell",
		b459Invasion: "Ruthless Invasion", b459Dross: "Chancellor of the Dross", b459Dogs: "Mortis Dogs",
		b459Conduit: "Soul Conduit", b459Parasite: "Hex Parasite", b459Blessing: "Apostle's Blessing",
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

func b459Seats(g *game.Game) (*game.Player, *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func b459Cast(t *testing.T, g *game.Game, me *game.Player, name, typeLine, cost, oracle string, targets []game.TargetRef) uuid.UUID {
	t.Helper()
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, OracleID: oracle, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: targets}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return id
}

func b459Effective(g *game.Game, id uuid.UUID) (c game.Characteristic) {
	g.ReadSnapshot(func() {
		if card := effectstest.FindBattlefieldCardForTest(g, id); card != nil {
			c = card.Effective()
		}
	})
	return c
}

func TestB459GethsVerdictSacrificesAndDrainsOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := b459Seats(g)
	effectstest.AdvanceToMain(t, g)
	only := effectstest.B12Creature(g, opp.ID, "Lone Bear", "Creature — Bear", 2, 2)
	life := opp.Life
	effectstest.ApaMana(me, "B", "B")
	b459Cast(t, g, me, "Geth's Verdict", "Instant", "{B}{B}", b459Verdict, effectstest.PlayerRefs(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, opp.ID, only)
	if effectstest.FindBattlefieldCardForTest(g, only) != nil {
		t.Error("the creature was not sacrificed")
	}
	if got := opp.Life; got != life-1 {
		t.Errorf("life %d, want %d", got, life-1)
	}
	// With no creature at all the life is still lost.
	effectstest.ApaMana(me, "B", "B")
	b459Cast(t, g, me, "Geth's Verdict", "Instant", "{B}{B}", b459Verdict, effectstest.PlayerRefs(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Life; got != life-2 {
		t.Errorf("life %d, want %d with nothing to sacrifice", got, life-2)
	}
}

func TestB459WarReportCountsCreaturesPlusArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := b459Seats(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	effectstest.B12Push(g, opp.ID, "Rock", "Artifact", "", 0, 0)
	effectstest.B12Creature(g, me.ID, "Golem", "Artifact Creature — Golem", 1, 1)
	life := me.Life
	effectstest.ApaMana(me, "W", "C", "C", "C")
	b459Cast(t, g, me, "War Report", "Instant", "{3}{W}", b459WarReport, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// Creatures: Bear, Ogre, Golem = 3. Artifacts: Rock, Golem = 2.
	if got := me.Life - life; got != 5 {
		t.Errorf("gained %d, want 5", got)
	}
}

func TestB459IsolationCellTaxesAnOpponentsCreatureSpells(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := b459Seats(g)
		effectstest.AdvanceToMain(t, g)
		effectstest.B12Push(g, me.ID, "Isolation Cell", "Artifact", b459Cell, 0, 0)
		id := uuid.New()
		g.WithWriteLock(func() {
			opp.Hand.PushTop(game.Card{InstanceID: id, Name: "Flash Bear", TypeLine: "Creature — Bear", Keywords: []string{"flash"},
				Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
		})
		if err := g.CastSpell(opp.ID, id, game.CastSpellParams{}); err != nil {
			t.Fatalf("opponent cast: %v", err)
		}
		life := opp.Life
		if pay {
			effectstest.ApaMana(opp, "C", "C")
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerPayUnless(t, g, opp.ID, pay)
		want := life - 2
		if pay {
			want = life
		}
		if opp.Life != want {
			t.Errorf("pay=%v: life %d, want %d", pay, opp.Life, want)
		}
		if me.Life != g.Settings.StartingLife {
			t.Errorf("the Cell's controller's life changed: %d", me.Life)
		}
	}
}

func TestB459RuthlessInvasionStopsOnlyNonartifactBlockers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := b459Seats(g)
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	golem := effectstest.B12Creature(g, opp.ID, "Golem", "Artifact Creature — Golem", 2, 2)
	effectstest.ApaMana(me, "R", "C", "C", "C")
	b459Cast(t, g, me, "Ruthless Invasion", "Sorcery", "{3}{R}", b459Invasion, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if b459Effective(g, bear).Restrictions&game.CantBlock == 0 {
		t.Error("a nonartifact creature can still block")
	}
	if b459Effective(g, golem).Restrictions&game.CantBlock != 0 {
		t.Error("an artifact creature was stopped from blocking")
	}
	late := effectstest.B12Creature(g, opp.ID, "Latecomer", "Creature — Bear", 1, 1)
	if b459Effective(g, late).Restrictions&game.CantBlock == 0 {
		t.Error("a creature that arrived later can block")
	}
}

func TestB459ChancellorOfTheDrossDrainsAtTheFirstUpkeep(t *testing.T) {
	g := effectstest.OpeningTable(t)
	starter := g.Seats[g.StartingSeat]
	ids := effectstest.HandWith(g, g.StartingSeat,
		game.Card{Name: "Chancellor of the Dross", TypeLine: "Creature — Phyrexian Vampire", ManaCost: "{4}{B}{B}{B}", OracleID: b459Dross, Power: 6, Toughness: 6},
		effectstest.FillerCard("Spare"))
	effectstest.KeepAll(t, g)
	var offer *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c.Kind == game.PendingChoiceConfirm && c.Chooser == starter.ID {
			offer = c
		}
	}
	if offer == nil {
		t.Fatal("the Chancellor's owner was not offered the reveal")
	}
	if err := g.ResolveConfirm(offer.ID, starter.ID, true); err != nil {
		t.Fatalf("reveal: %v", err)
	}
	if !starter.Hand.Contains(ids[0]) {
		t.Fatal("revealing moved the card")
	}
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = p.Life
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		want := before[p.ID] - 3
		if p.ID == starter.ID {
			want = before[p.ID] + 3*(len(g.Seats)-1)
		}
		if p.Life != want {
			t.Errorf("%s has %d life, want %d", p.Name, p.Life, want)
		}
	}
}

func TestB459MortisDogsPumpsOnAttackAndDrainsItsPowerOnDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := b459Seats(g)
	effectstest.AdvanceToMain(t, g)
	dog := effectstest.B12Push(g, me.ID, "Mortis Dogs", "Creature — Phyrexian Dog", b459Dogs, 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(dog, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, dog); p != 4 {
		t.Fatalf("attacking Dog has power %d, want 4", p)
	}
	life := opp.Life
	effectstest.B14Kill(g, dog)
	effectstest.AnswerPickTargetPlayerIfOpen(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Life; got != life-4 {
		t.Errorf("life %d, want %d (the Dog's last-known power, attack bonus included)", got, life-4)
	}
}

func TestB459SoulConduitExchangesTwoPlayersLifeTotals(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := b459Seats(g)
	effectstest.AdvanceToMain(t, g)
	cond := effectstest.B12Push(g, me.ID, "Soul Conduit", "Artifact", b459Conduit, 0, 0)
	g.WithWriteLock(func() { me.Life, opp.Life = 7, 31 })
	effectstest.RavB2MustActivate(t, g, me.ID, cond, 0,
		game.ActivateAbilityParams{Targets: effectstest.PlayerRefs(me.ID, opp.ID)}, "C", "C", "C", "C", "C", "C")
	if me.Life != 31 || opp.Life != 7 {
		t.Errorf("life totals %d / %d, want 31 / 7", me.Life, opp.Life)
	}
}

func TestB459HexParasiteRemovesUpToXCountersAndGrowsForEach(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := b459Seats(g)
	effectstest.AdvanceToMain(t, g)
	para := effectstest.B12Push(g, me.ID, "Hex Parasite", "Artifact Creature — Phyrexian Insect", b459Parasite, 1, 1)
	victim := effectstest.B12Creature(g, opp.ID, "Hydra", "Creature — Hydra", 1, 1)
	g.WithWriteLock(func() {
		c := effectstest.FindBattlefieldCardForTest(g, victim)
		c.Counters = map[string]int{game.CounterPlusOne: 3, game.CounterCharge: 2}
	})
	if err := effectstest.RavB2Activate(t, g, me.ID, para, 0,
		game.ActivateAbilityParams{Targets: effectstest.CardTarget(victim), XValue: 4}, "B", "C", "C", "C", "C"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// Kinds are asked in sorted order: "+1/+1" then "charge".
	first := effectstest.AnswerNumber(t, g, me.ID, 3)
	if first.Max != 3 {
		t.Errorf("first question allows up to %d, want 3 (the +1/+1 counters, X being 4)", first.Max)
	}
	second := effectstest.AnswerNumber(t, g, me.ID, 1)
	if second.Max != 1 {
		t.Errorf("second question allows up to %d, want 1 (X is all spent but one)", second.Max)
	}
	if got := effectstest.CounterCount(g, victim, game.CounterPlusOne); got != 0 {
		t.Errorf("+1/+1 counters left: %d, want 0", got)
	}
	if got := effectstest.CounterCount(g, victim, game.CounterCharge); got != 1 {
		t.Errorf("charge counters left: %d, want 1", got)
	}
	if p, _ := effectstest.PtOf(t, g, para); p != 5 {
		t.Errorf("Hex Parasite has power %d, want 5 (+1/+0 for each of the four counters)", p)
	}
}

func TestB459ApostlesBlessingGrantsProtectionFromTheChosenThing(t *testing.T) {
	for i, want := range []string{"protection from artifacts", "protection from white", "protection from red"} {
		pick := []int{0, 1, 4}[i]
		g := effectstest.NewCatalogGame(t)
		me, _ := b459Seats(g)
		effectstest.AdvanceToMain(t, g)
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		effectstest.ApaMana(me, "W", "C")
		b459Cast(t, g, me, "Apostle's Blessing", "Instant", "{1}{W/P}", b459Blessing, effectstest.CardTarget(bear))
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerOptionPick(t, g, me.ID, pick)
		if !effectstest.HasEffectiveKeyword(t, g, bear, want) {
			t.Errorf("the creature lacks %q", want)
		}
	}
	g := effectstest.NewCatalogGame(t)
	me, opp := b459Seats(g)
	effectstest.AdvanceToMain(t, g)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.ApaMana(me, "W", "C")
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Apostle's Blessing", TypeLine: "Instant", ManaCost: "{1}{W/P}", OracleID: b459Blessing, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(theirs)}); err == nil {
		t.Error("it targeted an opponent's creature")
	}
}
