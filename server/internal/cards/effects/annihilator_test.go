package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// annihilator_test.go — #2073 (ADR 0113 §2) end to end: the engine's
// annihilator trigger (CR 702.86), the defending player it asks (CR
// 508.5), the from-anywhere shuffle trigger, and the three titans the
// Jodi deck asked for. The token and the two removed-from-combat cases
// are in game/annihilator_test.go.

const (
	kozilekButcherOracle = "7b8528b0-71eb-4c9e-bed9-aa2d2e84038f"
	ulamogGyreOracle     = "b817bc56-9b4d-4c50-bafa-3c652b99578f"
	ulamogDefilerOracle  = "97836c48-8777-4b4e-98fb-e99204f38bdd"
)

// pushAnnihilatorCreature puts a ready Eldrazi with the given keyword
// tokens printed on it onto the battlefield.
func pushAnnihilatorCreature(g *game.Game, controller uuid.UUID, keywords ...string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test Eldrazi", TypeLine: "Creature — Eldrazi",
		Power: 5, Toughness: 5, Keywords: keywords,
		Owner: controller, Controller: controller,
	})
}

// pushAnnFodder puts n plain artifacts under owner's control.
func pushAnnFodder(g *game.Game, owner uuid.UUID, n int) []uuid.UUID {
	out := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, pushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Fodder", TypeLine: "Artifact",
			Owner: owner, Controller: owner,
		}))
	}
	return out
}

// annOnBattlefield reports whether the permanent is on the battlefield.
func annOnBattlefield(g *game.Game, id uuid.UUID) bool {
	_, ok := battlefieldCard(g, id)
	return ok
}

// annCountOn counts the permanents a player controls.
func annCountOn(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller {
			n++
		}
	}
	return n
}

// annihilatorItems are the annihilator triggers on the stack.
func annihilatorItems(g *game.Game) []*game.StackItem {
	var out []*game.StackItem
	for _, it := range g.StackMeta {
		if it != nil && (it.Body == "annihilator/sacrifice" ||
			(it.Params.Ability != nil && it.Kind == game.StackItemTriggered && it.Params.Player != uuid.Nil)) {
			out = append(out, it)
		}
	}
	return out
}

// annSettle resolves the stack, answering trigger order in the order
// offered and every own_permanents prompt with `pick` (nil: the first
// ChooseMax candidates). It records each prompt it answered.
func annSettle(t *testing.T, g *game.Game, pick func(c *game.PendingChoice) []uuid.UUID) []*game.PendingChoice {
	t.Helper()
	var asked []*game.PendingChoice
	for i := 0; i < 64; i++ {
		answered := false
		for _, c := range g.PendingChoices {
			if c == nil {
				continue
			}
			switch c.Kind {
			case game.PendingChoiceTriggerOrder:
				if err := g.ResolveTriggerOrder(c.ID, c.Chooser, append([]uuid.UUID(nil), c.TriggerOrderIDs...)); err != nil {
					t.Fatalf("ResolveTriggerOrder: %v", err)
				}
				answered = true
			case game.PendingChoiceOwnPermanents:
				cp := *c
				asked = append(asked, &cp)
				var ids []uuid.UUID
				if pick != nil {
					ids = pick(c)
				} else {
					ids = append(ids, c.ChooseCards[:c.ChooseMax]...)
				}
				if err := g.ResolveOwnPermanents(c.ID, c.Chooser, ids); err != nil {
					t.Fatalf("ResolveOwnPermanents: %v", err)
				}
				answered = true
			}
			if answered {
				break
			}
		}
		if answered {
			continue
		}
		if stackFullyEmpty(g) {
			return asked
		}
		passPriorityAroundTable(t, g)
	}
	t.Fatal("the stack did not settle")
	return asked
}

// --- the keyword ------------------------------------------------------

// A printed "annihilator 4" makes the defending player sacrifice four
// permanents of their choice, in one choice, before blockers.
func TestAnnihilatorSacrificesNPermanentsInOneChoice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	eldrazi := pushAnnihilatorCreature(g, me.ID, "annihilator 4")
	fodder := pushAnnFodder(g, opp.ID, 6)

	declareAttack(t, g, opp.ID, eldrazi)
	if n := len(annihilatorItems(g)); n != 1 {
		t.Fatalf("annihilator triggers on the stack = %d, want 1", n)
	}
	asked := annSettle(t, g, nil)
	if len(asked) != 1 {
		t.Fatalf("prompts = %d, want ONE choice of four", len(asked))
	}
	if c := asked[0]; c.Chooser != opp.ID || c.ChooseMin != 4 || c.ChooseMax != 4 || len(c.ChooseCards) != 6 {
		t.Errorf("prompt = chooser %s, %d..%d of %d; want the defender choosing exactly 4 of 6",
			c.Chooser, c.ChooseMin, c.ChooseMax, len(c.ChooseCards))
	}
	if got := annCountOn(g, opp.ID); got != 2 {
		t.Errorf("defender has %d permanents left, want 2", got)
	}
	sacrificed := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventSacrifice {
			for _, id := range fodder {
				if ev.CardID == id {
					sacrificed++
				}
			}
		}
	}
	if sacrificed != 4 {
		t.Errorf("sacrifice events = %d, want 4 — the permanents are SACRIFICED", sacrificed)
	}
	if g.Turn.Step != game.StepDeclareAttackers {
		t.Errorf("step = %s, want the sacrifice to happen in declare attackers, before blocks", g.Turn.Step)
	}
}

// CR 702.86b: several instances trigger separately. A granted
// annihilator 2 on a printed annihilator 4 is two triggers, and the
// defender sacrifices six.
func TestAnnihilatorGrantedBesidePrintedTriggersTwice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	eldrazi := pushAnnihilatorCreature(g, me.ID, "annihilator 4")
	pushAnnFodder(g, opp.ID, 8)
	g.WithWriteLock(func() {
		c, _ := battlefieldCard(g, eldrazi)
		g.RegisterScopedEffectForEffect(eldrazi,
			[]game.AffectedObject{game.PinObject(eldrazi, c.EnteredBattlefieldAt)},
			[]game.Mod{game.AddKeywordsMod("annihilator 2")}, g.UntilEndOfTurnDuration(), "test grant")
	})

	declareAttack(t, g, opp.ID, eldrazi)
	asked := annSettle(t, g, nil)
	if len(asked) != 2 {
		t.Fatalf("prompts = %d, want one per instance", len(asked))
	}
	seen := map[int]bool{}
	for _, c := range asked {
		seen[c.ChooseMax] = true
	}
	if !seen[4] || !seen[2] {
		t.Errorf("prompt sizes %v, want one of 4 and one of 2", seen)
	}
	if got := annCountOn(g, opp.ID); got != 2 {
		t.Errorf("defender has %d permanents left, want 8-6 = 2", got)
	}
}

// Two identical instances on one creature need no ordering prompt.
func TestTwoIdenticalAnnihilatorsNeedNoOrderingPrompt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	eldrazi := pushAnnihilatorCreature(g, me.ID, "annihilator 1", "annihilator 1")
	pushAnnFodder(g, opp.ID, 3)

	declareAttack(t, g, opp.ID, eldrazi)
	if c := pendingOfKind(g, game.PendingChoiceTriggerOrder); c != nil {
		t.Errorf("an ordering prompt was asked for two identical triggers: %+v", c)
	}
	asked := annSettle(t, g, nil)
	if len(asked) != 2 || annCountOn(g, opp.ID) != 1 {
		t.Errorf("prompts %d, defender left with %d; want 2 and 1", len(asked), annCountOn(g, opp.ID))
	}
}

// CR 508.3a: a creature put onto the battlefield attacking was never
// declared as an attacker, so "whenever this creature attacks" does
// not trigger.
func TestAnnihilatorDoesNotTriggerForACreaturePutOntoTheBattlefieldAttacking(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushAnnFodder(g, opp.ID, 3)
	advanceTo(t, g, game.StepDeclareAttackers)
	lockInAttacks(t, g)
	g.WithWriteLock(func() {
		if err := g.CreateTokensAttackingForEffect(me.ID, game.Card{
			Name: "Eldrazi Token", TypeLine: "Token Creature — Eldrazi", Power: 5, Toughness: 5,
			Keywords: []string{"annihilator 2"},
		}, 1, opp.ID); err != nil {
			t.Fatalf("CreateTokensAttackingForEffect: %v", err)
		}
	})
	if n := len(annihilatorItems(g)) + len(g.PendingTriggers); n != 0 {
		t.Errorf("annihilator triggered %d times for a creature put onto the battlefield attacking", n)
	}
	annSettle(t, g, nil)
	if got := annCountOn(g, opp.ID); got != 3 {
		t.Errorf("defender lost permanents (%d left of 3)", got)
	}
}

// CR 508.5a: in a four-player game, each attacker's trigger asks its
// own defending player.
func TestAnnihilatorAsksEachAttackersOwnDefender(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	defenders := []*game.Player{g.Seats[1], g.Seats[2], g.Seats[3]}
	var attackers []uuid.UUID
	for i, d := range defenders {
		attackers = append(attackers, pushAnnihilatorCreature(g, me.ID, "annihilator "+string(rune('1'+i))))
		pushAnnFodder(g, d.ID, 4)
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	for i, a := range attackers {
		if err := g.DeclareAttacker(a, defenders[i].ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
	asked := annSettle(t, g, nil)
	if len(asked) != 3 {
		t.Fatalf("prompts = %d, want 3", len(asked))
	}
	for i, d := range defenders {
		want := i + 1
		found := false
		for _, c := range asked {
			if c.Chooser == d.ID {
				found = true
				if c.ChooseMax != want {
					t.Errorf("defender %d was asked for %d, want %d", i+1, c.ChooseMax, want)
				}
			}
		}
		if !found {
			t.Errorf("defender %d was never asked", i+1)
		}
		if got := annCountOn(g, d.ID); got != 4-want {
			t.Errorf("defender %d has %d permanents, want %d", i+1, got, 4-want)
		}
	}
}

// CR 701.21a / 609.3: a defender with fewer than N permanents
// sacrifices all of them.
func TestAnnihilatorWithFewerPermanentsThanN(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	eldrazi := pushAnnihilatorCreature(g, me.ID, "annihilator 4")
	pushAnnFodder(g, opp.ID, 2)

	declareAttack(t, g, opp.ID, eldrazi)
	asked := annSettle(t, g, nil)
	if len(asked) != 1 || asked[0].ChooseMin != 2 || asked[0].ChooseMax != 2 {
		t.Fatalf("prompts %+v, want one choice of exactly the two they have", asked)
	}
	if got := annCountOn(g, opp.ID); got != 0 {
		t.Errorf("defender has %d permanents left, want 0", got)
	}
}

// A defender with no permanents at all is asked nothing.
func TestAnnihilatorAgainstAnEmptyBoardAsksNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	eldrazi := pushAnnihilatorCreature(g, me.ID, "annihilator 2")
	declareAttack(t, g, opp.ID, eldrazi)
	if asked := annSettle(t, g, nil); len(asked) != 0 {
		t.Errorf("an empty board was asked %d prompts", len(asked))
	}
}

// CR 508.5: attacking a planeswalker, the defending player is its
// controller, who may sacrifice that planeswalker. The attacker keeps
// attacking, and a second instance still asks the same player.
func TestAnnihilatorAttackingAPlaneswalker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	eldrazi := pushAnnihilatorCreature(g, me.ID, "annihilator 1", "annihilator 1")
	walker := pushWalkerForTest(g, opp.ID, "Test Walker", "", 5)
	pushAnnFodder(g, opp.ID, 2)

	declareAttack(t, g, walker, eldrazi)
	for _, it := range annihilatorItems(g) {
		if it.Params.Player != opp.ID {
			t.Errorf("the trigger recorded %s as the defending player, want the walker's controller", it.Params.Player)
		}
	}
	first := true
	asked := annSettle(t, g, func(c *game.PendingChoice) []uuid.UUID {
		if first {
			first = false
			return []uuid.UUID{walker}
		}
		return c.ChooseCards[:1]
	})
	if len(asked) != 2 {
		t.Fatalf("prompts = %d, want 2", len(asked))
	}
	for _, c := range asked {
		if c.Chooser != opp.ID {
			t.Errorf("prompt went to %s, want the walker's controller", c.Chooser)
		}
	}
	if annOnBattlefield(g, walker) {
		t.Error("the attacked planeswalker could not be sacrificed")
	}
	if c, ok := battlefieldCard(g, eldrazi); !ok || c.AttackingTarget == uuid.Nil {
		t.Error("the attacker stopped attacking when the planeswalker was sacrificed")
	}
	if got := annCountOn(g, opp.ID); got != 1 {
		t.Errorf("defender has %d permanents left, want 1", got)
	}
}

// CR 508.5: attacking a battle, the defending player is its protector,
// not its controller.
func TestAnnihilatorAttackingABattleAsksItsProtector(t *testing.T) {
	g := newCatalogGame(t)
	me, owner, protector := g.Seats[0], g.Seats[1], g.Seats[2]
	eldrazi := pushAnnihilatorCreature(g, me.ID, "annihilator 1")
	battle := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test Siege", TypeLine: "Battle — Siege",
		Owner: owner.ID, Controller: owner.ID, ProtectorPlayerID: protector.ID,
		Counters: map[string]int{game.CounterDefense: 5},
	})
	pushAnnFodder(g, protector.ID, 2)
	pushAnnFodder(g, owner.ID, 2)

	declareAttack(t, g, battle, eldrazi)
	asked := annSettle(t, g, nil)
	if len(asked) != 1 || asked[0].Chooser != protector.ID {
		t.Fatalf("prompts %+v, want one for the battle's protector", asked)
	}
	if annCountOn(g, protector.ID) != 1 {
		t.Errorf("the protector did not sacrifice")
	}
}

// CR 800.4a: a defender who leaves the game before the trigger
// resolves controls nothing and is asked nothing.
func TestAnnihilatorAgainstADefenderWhoHasLeftTheGame(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	eldrazi := pushAnnihilatorCreature(g, me.ID, "annihilator 2")
	pushAnnFodder(g, opp.ID, 3)
	declareAttack(t, g, opp.ID, eldrazi)
	if err := g.Concede(opp.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if asked := annSettle(t, g, nil); len(asked) != 0 {
		t.Errorf("a player who left the game was asked %d prompts", len(asked))
	}
}

// --- the titans ---------------------------------------------------------

func TestKozilekButcherOfTruthDrawsFourOnCastEvenIfCountered(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := putInHand(me, game.Card{Name: "Kozilek, Butcher of Truth", TypeLine: "Legendary Creature — Eldrazi",
		OracleID: kozilekButcherOracle, ManaCost: "{10}", Power: 12, Toughness: 12})
	toMain(t, g)
	handBefore := me.Hand.Size()
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	g.WithWriteLock(func() {
		if err := g.CounterTargetForEffect(id); err != nil {
			t.Fatalf("counter: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	// -1 for the cast, +4 for the trigger.
	if got := me.Hand.Size(); got != handBefore+3 {
		t.Errorf("hand %d -> %d, want the four cards drawn even though Kozilek was countered", handBefore, got)
	}
}

// Kozilek put into a graveyard from anywhere — killed, discarded,
// milled or countered — shuffles its owner's graveyard into their
// library, Kozilek included.
func TestKozilekShufflesTheGraveyardFromAnywhere(t *testing.T) {
	kozilek := func(owner uuid.UUID) game.Card {
		return game.Card{InstanceID: uuid.New(), Name: "Kozilek, Butcher of Truth", TypeLine: "Legendary Creature — Eldrazi",
			OracleID: kozilekButcherOracle, ManaCost: "{10}", Power: 12, Toughness: 12, Owner: owner, Controller: owner}
	}
	for _, tc := range []struct {
		name string
		send func(t *testing.T, g *game.Game, me *game.Player) uuid.UUID
	}{
		{"killed", func(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
			id := pushBattlefieldCardWithTimestamp(g, kozilek(me.ID))
			g.WithWriteLock(func() {
				if err := g.DestroyPermanentForEffect(id); err != nil {
					t.Fatalf("destroy: %v", err)
				}
			})
			return id
		}},
		{"discarded", func(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
			me.Hand.Cards = nil
			c := kozilek(me.ID)
			me.Hand.PushTop(c)
			g.WithWriteLock(func() {
				if err := g.DiscardRandomForEffect(me.ID, 1); err != nil {
					t.Fatalf("discard: %v", err)
				}
			})
			return c.InstanceID
		}},
		{"milled", func(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
			c := kozilek(me.ID)
			me.Library.PushTop(c)
			g.WithWriteLock(func() {
				if err := g.MillNForEffect(me.ID, 1); err != nil {
					t.Fatalf("mill: %v", err)
				}
			})
			return c.InstanceID
		}},
		{"countered", func(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
			id := putInHand(me, kozilek(me.ID))
			toMain(t, g)
			if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			g.WithWriteLock(func() {
				if err := g.CounterTargetForEffect(id); err != nil {
					t.Fatalf("counter: %v", err)
				}
			})
			return id
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			pushGraveyardPermanent(me, "Old Bear", "Creature — Bear", "{1}{G}")
			pushGraveyardPermanent(me, "Old Wolf", "Creature — Wolf", "{2}{G}")
			id := tc.send(t, g, me)
			passPriorityAroundTable(t, g)
			if n := me.Graveyard.Size(); n != 0 {
				t.Errorf("graveyard still holds %d cards; the shuffle trigger did not run", n)
			}
			if !me.Library.Contains(id) {
				t.Error("Kozilek is not in its owner's library")
			}
			// One arrival, one trigger: a death is announced by both a
			// zone move and a leaves-the-battlefield event.
			shuffles := 0
			for _, ev := range g.Events {
				if ev.Kind == game.EventTrigger && ev.Source == id && strings.Contains(ev.Label, "shuffles their graveyard") {
					shuffles++
				}
			}
			if shuffles != 1 {
				t.Errorf("the shuffle triggered %d times, want 1", shuffles)
			}
		})
	}
}

// Exiled is not "put into a graveyard": no shuffle.
func TestKozilekExiledDoesNotShuffle(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushGraveyardPermanent(me, "Old Bear", "Creature — Bear", "{1}{G}")
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Kozilek, Butcher of Truth",
		TypeLine: "Legendary Creature — Eldrazi", OracleID: kozilekButcherOracle, Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() {
		if _, err := g.MillToZoneForEffect(me.ID, 1, game.ZoneExile); err != nil {
			t.Fatalf("exile top: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if n := me.Graveyard.Size(); n != 1 {
		t.Errorf("graveyard holds %d cards, want the 1 untouched", n)
	}
}

func TestUlamogTheInfiniteGyreDestroysOnCastAndAnnihilatesFour(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushCatalogPermanent(g, opp.ID, "Bear", "Creature — Bear", "", false)
	id := putInHand(me, game.Card{Name: "Ulamog, the Infinite Gyre", TypeLine: "Legendary Creature — Eldrazi",
		OracleID: ulamogGyreOracle, ManaCost: "{11}", Power: 10, Toughness: 10})
	toMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if annOnBattlefield(g, victim) {
		t.Error("the cast trigger did not destroy its target")
	}
	card, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatal("Ulamog did not resolve")
	}
	if !game.HasKeyword(&card, "indestructible") {
		t.Error("Ulamog is not indestructible")
	}
	if got := game.AnnihilatorAmounts(&card); len(got) != 1 || got[0] != 4 {
		t.Errorf("annihilator instances = %v, want [4]", got)
	}
}

// --- Ulamog, the Defiler -------------------------------------------------

func TestUlamogTheDefilerCastTriggerExilesHalfALibraryRoundedUp(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	opp.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Extra", TypeLine: "Land", Owner: opp.ID, Controller: opp.ID})
	libBefore := opp.Library.Size() // 21 after the draw step's dealing
	id := putInHand(me, game.Card{Name: "Ulamog, the Defiler", TypeLine: "Legendary Creature — Eldrazi",
		OracleID: ulamogDefilerOracle, ManaCost: "{10}", Power: 7, Toughness: 7})
	toMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the cast trigger did not ask for its target opponent")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	passPriorityAroundTable(t, g)
	want := libBefore - (libBefore+1)/2
	if got := opp.Library.Size(); got != want {
		t.Errorf("library %d -> %d, want %d (half rounded up exiled)", libBefore, got, want)
	}
}

// The entry counters read the cards in exile, ignoring a face-down one
// (CR 406.3a), before Ulamog moves — so from exile it sees itself.
func TestUlamogTheDefilerEntersWithTheGreatestManaValueInExile(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	g.Exile.PushTop(game.Card{InstanceID: uuid.New(), Name: "Seven", TypeLine: "Sorcery", ManaCost: "{5}{U}{U}", Owner: opp.ID})
	faceDown := game.Card{InstanceID: uuid.New(), Name: "Hidden", TypeLine: "Sorcery", ManaCost: "{12}", Owner: opp.ID}
	faceDown.FaceDown, faceDown.FaceDownKind = true, game.FaceDownExiled
	g.Exile.PushTop(faceDown)

	fromHand := enterCard(t, g, me.ID, game.Card{Name: "Ulamog, the Defiler", TypeLine: "Legendary Creature — Eldrazi",
		OracleID: ulamogDefilerOracle, ManaCost: "{10}", Power: 7, Toughness: 7})
	if got := countersOn(g, fromHand, game.CounterPlusOne); got != 7 {
		t.Errorf("from hand: %d +1/+1 counters, want 7", got)
	}

	exiled := game.Card{InstanceID: uuid.New(), Name: "Ulamog, the Defiler", TypeLine: "Legendary Creature — Eldrazi",
		OracleID: ulamogDefilerOracle, ManaCost: "{10}", Power: 7, Toughness: 7, Owner: opp.ID, Controller: opp.ID}
	g.Exile.PushTop(exiled)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneExile}, game.ZoneRef{Kind: game.ZoneBattlefield}, exiled.InstanceID); err != nil {
		t.Fatalf("move from exile: %v", err)
	}
	if got := countersOn(g, exiled.InstanceID, game.CounterPlusOne); got != 10 {
		t.Errorf("from exile: %d +1/+1 counters, want 10 (it sees itself)", got)
	}
}

// "Annihilator X" reads the counters as it resolves: a counter added in
// response is counted.
func TestUlamogTheDefilerAnnihilatorReadsCountersOnResolution(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ulamog := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Ulamog, the Defiler", TypeLine: "Legendary Creature — Eldrazi",
		OracleID: ulamogDefilerOracle, Power: 7, Toughness: 7, Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterPlusOne: 2},
	})
	pushAnnFodder(g, opp.ID, 6)
	declareAttack(t, g, opp.ID, ulamog)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(ulamog, game.CounterPlusOne, 1); err != nil {
			t.Fatalf("add counter: %v", err)
		}
	})
	asked := annSettle(t, g, nil)
	if len(asked) != 1 || asked[0].Chooser != opp.ID || asked[0].ChooseMax != 3 {
		t.Fatalf("prompts %+v, want the defender choosing 3", asked)
	}
	if got := annCountOn(g, opp.ID); got != 3 {
		t.Errorf("defender has %d permanents left, want 3", got)
	}
}

// With no counters, X is zero and nobody is asked anything.
func TestUlamogTheDefilerWithNoCountersAnnihilatesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ulamog := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Ulamog, the Defiler", TypeLine: "Legendary Creature — Eldrazi",
		OracleID: ulamogDefilerOracle, Power: 7, Toughness: 7, Owner: me.ID, Controller: me.ID,
	})
	pushAnnFodder(g, opp.ID, 2)
	declareAttack(t, g, opp.ID, ulamog)
	if asked := annSettle(t, g, nil); len(asked) != 0 || annCountOn(g, opp.ID) != 2 {
		t.Errorf("annihilator 0 asked %d prompts and left %d permanents", len(asked), annCountOn(g, opp.ID))
	}
}

// --- snapshots ------------------------------------------------------------

// A table with an annihilator trigger waiting on the stack is a restore
// point, and the restored table resolves it. While the defender is
// choosing, it holds Lotus Field's pick prompt and is not.
func TestAnnihilatorTriggerOnTheStackRestores(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	eldrazi := pushAnnihilatorCreature(g, me.ID, "annihilator 2")
	pushAnnFodder(g, opp.ID, 3)
	declareAttack(t, g, opp.ID, eldrazi)
	if n := len(annihilatorItems(g)); n != 1 {
		t.Fatalf("setup: %d annihilator triggers on the stack", n)
	}
	restored := restoreThroughJSON(t, g)
	asked := annSettle(t, restored, nil)
	if len(asked) != 1 || asked[0].ChooseMax != 2 {
		t.Fatalf("restored table asked %+v, want one choice of 2", asked)
	}
	if got := annCountOn(restored, opp.ID); got != 1 {
		t.Errorf("restored defender has %d permanents left, want 1", got)
	}

	// The pick open: the same restore-point class as Lotus Field's.
	passPriorityAroundTable(t, g)
	if c := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, opp.ID); c == nil {
		t.Fatal("no pick prompt on the original table")
	}
	if open := g.CaptureSnapshot(); open.Restorable() || open.Continuations.ChoiceResumeFrames == 0 {
		t.Errorf("with the pick open: restorable %v, resume frames %d; want a continuation-blocked capture",
			open.Restorable(), open.Continuations.ChoiceResumeFrames)
	}
}
