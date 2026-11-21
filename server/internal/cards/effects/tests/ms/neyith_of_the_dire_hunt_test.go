package testsms

import (
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const neyithOracle = "aba04455-c13f-48f7-9674-56803b39a42b"

func neyithPut(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.B12Push(g, owner, "Neyith of the Dire Hunt", "Legendary Creature — Human Warrior", neyithOracle, 3, 3)
}
func neyithFight(t *testing.T, g *game.Game, a, b uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := effects.B10Fight(effects.NewContext(g, &game.StackItem{Controller: g.Seats[0].ID}), a, b); err != nil {
			t.Fatal(err)
		}
	})
}
func neyithFightEvents(g *game.Game) []game.Event {
	var out []game.Event
	for _, ev := range g.Events {
		if ev.Kind == game.EventFight {
			out = append(out, ev)
		}
	}
	return out
}
func TestNeyithFightDrawsOnceEvenWithoutDamage(t *testing.T) {
	for _, tc := range []struct {
		name          string
		power         int
		prevent, both bool
	}{
		{"ordinary", 2, false, false}, {"zero power", 0, false, false}, {"prevented", 2, true, false}, {"both yours", 2, false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			neyithPut(g, me.ID)
			owner := opp.ID
			if tc.both {
				owner = me.ID
			}
			a := effectstest.PushVanillaCreature(g, me.ID, "First", tc.power, 8)
			b := effectstest.PushVanillaCreature(g, owner, "Second", tc.power, 8)
			if tc.prevent {
				effectstest.P1ShieldCreature(g, a)
				effectstest.P1ShieldCreature(g, b)
			}
			before := len(me.Hand.Cards)
			neyithFight(t, g, a, b)
			events := neyithFightEvents(g)
			if len(events) != 2 || events[0].BatchSeq == 0 || events[0].BatchSeq != events[1].BatchSeq || events[0].Actor != me.ID || events[1].Actor != owner {
				t.Fatalf("fight participation: %+v", events)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if got := len(me.Hand.Cards) - before; got != 1 {
				t.Fatalf("drew %d, want 1", got)
			}
			want := tc.power
			if tc.prevent {
				want = 0
			}
			if effectstest.GaLayered(t, g, a).DamageMarked != want || effectstest.GaLayered(t, g, b).DamageMarked != want {
				t.Fatal("fight damage")
			}
		})
	}
}
func TestNeyithSeparateFightsInOneBatchAndOpponentOnlyFight(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	neyithPut(g, me.ID)
	a := effectstest.PushVanillaCreature(g, me.ID, "Mine", 0, 8)
	b := effectstest.PushVanillaCreature(g, opp.ID, "Theirs", 0, 8)
	c := effectstest.PushVanillaCreature(g, opp.ID, "Also theirs", 0, 8)
	before := len(me.Hand.Cards)
	neyithFight(t, g, a, b)
	neyithFight(t, g, a, b)
	neyithFight(t, g, b, c)
	evs := neyithFightEvents(g)
	if evs[0].Batch != evs[2].Batch || evs[0].BatchSeq == evs[2].BatchSeq {
		t.Fatal("fight occurrence grouping")
	}
	g = effectstest.RestoreRoundTrip(t, g, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(g.Seats[0].Hand.Cards) - before; got != 2 {
		t.Fatalf("drew %d, want two separate fights", got)
	}
}
func TestNeyithInvalidFightersNeverFight(t *testing.T) {
	for _, mode := range []string{"absent", "blink", "noncreature", "phased out"} {
		t.Run(mode, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			neyithPut(g, me.ID)
			a := effectstest.PushVanillaCreature(g, me.ID, "Mine", 2, 8)
			b := effectstest.PushVanillaCreature(g, opp.ID, "Theirs", 2, 8)
			var ra, rb game.ObjectRef
			g.WithWriteLock(func() { ra, _ = g.PermanentRefForEffect(a); rb, _ = g.PermanentRefForEffect(b) })
			switch mode {
			case "absent":
				effectstest.P1Destroy(t, g, b)
			case "blink":
				effectstest.DdFlicker(t, g, b)
			case "noncreature":
				g.WithWriteLock(func() {
					for i := range g.Battlefield.Cards {
						if g.Battlefield.Cards[i].InstanceID == b {
							g.Battlefield.Cards[i].TypeLine = "Artifact"
						}
					}
					g.RecomputeLayersIfStaleLocked()
				})
			case "phased out":
				g.WithWriteLock(func() {
					if err := g.PhaseOutForEffect(uuid.Nil, b); err != nil {
						t.Fatal(err)
					}
				})
			}
			before := len(me.Hand.Cards)
			g.WithWriteLock(func() {
				if err := g.FightForEffect(ra, rb); err != nil {
					t.Fatal(err)
				}
			})
			effectstest.PassPriorityAroundTable(t, g)
			if len(neyithFightEvents(g)) != 0 || len(me.Hand.Cards) != before || effectstest.GaLayered(t, g, a).DamageMarked != 0 {
				t.Fatal("invalid fighters fought")
			}
		})
	}
}
func TestNeyithBlockedCreaturesDrawOncePerDeclaration(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.BrCreature(g, me.ID, "A", "Creature", 2, 8)
	b := effectstest.BrCreature(g, me.ID, "B", "Creature", 2, 8)
	x := effectstest.BrCreature(g, opp.ID, "X", "Creature", 1, 8)
	y := effectstest.BrCreature(g, opp.ID, "Y", "Creature", 1, 8)
	z := effectstest.BrCreature(g, opp.ID, "Z", "Creature", 1, 8)
	effectstest.BrAttack(t, g, a, b)
	neyithPut(g, me.ID)
	before := len(me.Hand.Cards)
	if err := g.DeclareBlockers([]game.BlockDeclaration{{Blocker: x, Attacker: a}, {Blocker: y, Attacker: a}, {Blocker: z, Attacker: b}}); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if len(me.Hand.Cards) != before+1 {
		t.Fatal("multiple blocked creatures and a double block should draw one")
	}
}
func neyithCombatTarget(t *testing.T, g *game.Game, target uuid.UUID) *game.Game {
	t.Helper()
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, g.Seats[0].ID) == nil {
		t.Fatal("no beginning-combat target choice")
	}
	g = effectstest.RestoreRoundTrip(t, g, true)
	effectstest.AnswerTriggerTargets(t, g, g.Seats[0].ID, target)
	return g
}
func TestNeyithCombatPaymentAndDurationsRestore(t *testing.T) {
	for _, mana := range []string{"{C}{C}{R}", "{C}{C}{G}"} {
		t.Run(mana, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			neyithPut(g, me.ID)
			target := effectstest.PushVanillaCreature(g, opp.ID, "Opponent target", 3, 8)
			g = neyithCombatTarget(t, g, target)
			g.WithWriteLock(func() {
				if err := g.AddManaForEffect(me.ID, uuid.Nil, mana); err != nil {
					t.Fatal(err)
				}
			})
			effectstest.PassPriorityAroundTable(t, g)
			if !effectstest.HasPayUnlessFor(g, me.ID) {
				t.Fatal("no payment prompt")
			}
			g = effectstest.RestoreRoundTrip(t, g, true)
			effectstest.AnswerPayUnless(t, g, me.ID, true)
			c := effectstest.GaLayered(t, g, target)
			if c.CurrentPower() != 6 || c.CurrentToughness() != 8 || len(c.Effective().BlockRequirements) != 1 {
				t.Fatalf("paid result: %+v", c.Effective())
			}
			if g.Seats[0].ManaPool.Total() != 0 {
				t.Fatal("payment was not collected")
			}
			g = effectstest.RestoreRoundTrip(t, g, true)
			effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
			c = effectstest.GaLayered(t, g, target)
			if c.CurrentPower() != 6 || len(c.Effective().BlockRequirements) != 0 {
				t.Fatal("requirement must expire at combat end while power remains doubled")
			}
		})
	}
}
func TestNeyithCombatDeclineUnpayableAndInvalidTarget(t *testing.T) {
	for _, mode := range []string{"decline", "unpayable", "gone", "blink"} {
		t.Run(mode, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			neyithPut(g, me.ID)
			target := effectstest.PushVanillaCreature(g, me.ID, "Target", 3, 8)
			g = neyithCombatTarget(t, g, target)
			me = g.Seats[0]
			if mode != "unpayable" {
				g.WithWriteLock(func() { _ = g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{G}") })
			}
			if mode == "gone" {
				effectstest.P1Destroy(t, g, target)
			}
			if mode == "blink" {
				effectstest.DdFlicker(t, g, target)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if mode == "gone" || mode == "blink" {
				if effectstest.HasPayUnlessFor(g, me.ID) {
					t.Fatal("illegal target still offered payment")
				}
				if me.ManaPool.Total() != 3 {
					t.Fatal("fizzle spent mana")
				}
				return
			}
			if effectstest.HasPayUnlessFor(g, me.ID) {
				effectstest.AnswerPayUnless(t, g, me.ID, mode == "unpayable")
			}
			c := effectstest.GaLayered(t, g, target)
			if c.CurrentPower() != 3 || len(c.Effective().BlockRequirements) != 0 {
				t.Fatal("unpaid ability changed the target")
			}
		})
	}
}

func TestNeyithSelfFightDealsTwicePowerAsOneEvent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	neyithPut(g, me.ID)
	a := effectstest.PushVanillaCreature(g, me.ID, "Self fighter", 2, 8)
	before := len(me.Hand.Cards)
	neyithFight(t, g, a, a)
	effectstest.PassPriorityAroundTable(t, g)
	if len(neyithFightEvents(g)) != 1 || len(me.Hand.Cards) != before+1 || effectstest.GaLayered(t, g, a).DamageMarked != 4 {
		t.Fatal("self fight should deal twice power and draw once")
	}
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventDealDamage && ev.Source == a {
			n++
			if ev.Amount != 4 {
				t.Fatal("self fight damage was split")
			}
		}
	}
	if n != 1 {
		t.Fatalf("%d damage events, want one", n)
	}
}

func TestNeyithFightReplacementChoiceStrictRestore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	neyithPut(g, me.ID)
	a := effectstest.PushVanillaCreature(g, me.ID, "First", 2, 8)
	b := effectstest.PushVanillaCreature(g, opp.ID, "Second", 3, 8)
	effectstest.B12Push(g, me.ID, "Angrath's Marauders", "Creature — Human Pirate", effectstest.B14AngrathsMaraudersOracle, 4, 4)
	g.WithWriteLock(func() { g.PreventNextDamageThisTurnForEffect(uuid.Nil, b, 1, false, "prevent the next 1 damage") })
	before := len(me.Hand.Cards)
	neyithFight(t, g, a, b)
	if effectstest.LatestChoiceOfKind(g, game.PendingChoiceReplacementOrder) == nil {
		t.Fatal("fight did not pause on damage replacement order")
	}
	g = effectstest.RestoreRoundTrip(t, g, true)
	effectstest.AnswerOrderingPromptDoublerFirst(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.GaLayered(t, g, a).DamageMarked != 3 || effectstest.GaLayered(t, g, b).DamageMarked != 3 {
		t.Fatal("both fight blows must survive a replacement prompt and restore")
	}
	if len(g.Seats[0].Hand.Cards) != before+1 || len(neyithFightEvents(g)) != 2 {
		t.Fatal("fight trigger duplicated or lost across restore")
	}
}
