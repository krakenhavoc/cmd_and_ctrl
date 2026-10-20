package effects

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const umezawasJitteOracle = "1da10d5c-36a8-473f-a5d4-782ad61d8057"

func jitteSetup(t *testing.T) (*game.Game, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 4)
	jitte := pushCatalogPermanent(g, me.ID, "Umezawa's Jitte", "Legendary Artifact — Equipment", umezawasJitteOracle, false)
	g.WithWriteLock(func() {
		if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}"); err != nil {
			t.Fatal(err)
		}
	})
	if err := g.ActivateCatalogAbility(me.ID, jitte, 1, game.ActivateAbilityParams{Targets: cardRefs(bear), Strict: true}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	return g, jitte, bear
}

func activateJitte(t *testing.T, g *game.Game, jitte uuid.UUID, mode int, targets ...uuid.UUID) {
	t.Helper()
	if err := g.ActivateCatalogAbility(g.Seats[0].ID, jitte, 0, game.ActivateAbilityParams{Modes: []int{mode}, Targets: cardRefs(targets...)}); err != nil {
		t.Fatal(err)
	}
}

func TestUmezawasJitteBanksTwoCountersForCombatDamageToAnyRecipient(t *testing.T) {
	for _, blocked := range []bool{false, true} {
		t.Run(map[bool]string{false: "player", true: "blocker"}[blocked], func(t *testing.T) {
			g, jitte, bear := jitteSetup(t)
			opp := g.Seats[1]
			wall := b16Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 5)
			attackIntoBlocks(t, g, opp.ID, bear)
			if blocked {
				if err := g.DeclareBlocker(wall, bear); err != nil {
					t.Fatal(err)
				}
			}
			advanceTo(t, g, game.StepCombatDamage)
			passPriorityAroundTable(t, g)
			if n := counterOn(g, jitte, "charge"); n != 2 {
				t.Fatalf("charge=%d, want 2", n)
			}
		})
	}
}

func TestUmezawasJitteIgnoresNoncombatDamageAndZeroDamage(t *testing.T) {
	g, jitte, bear := jitteSetup(t)
	opp := g.Seats[1]
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(bear, opp.ID, 1); err != nil {
			t.Fatal(err)
		}
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: bear, Target: opp.ID, Combat: true, Amount: 0})
	})
	passPriorityAroundTable(t, g)
	if n := counterOn(g, jitte, "charge"); n != 0 {
		t.Fatalf("charge=%d after noncombat or zero damage", n)
	}
}

func TestUmezawasJitteAllModesPayCounterAtActivation(t *testing.T) {
	for _, mode := range []int{0, 1, 2} {
		t.Run([]string{"pump", "shrink", "life"}[mode], func(t *testing.T) {
			g, jitte, bear := jitteSetup(t)
			me := g.Seats[0]
			target := b16Creature(g, g.Seats[1].ID, "Other", "Creature", 2, 2)
			setCounters(g, jitte, map[string]int{"charge": 1})
			life := me.Life
			if mode == 1 {
				activateJitte(t, g, jitte, mode, target)
			} else {
				activateJitte(t, g, jitte, mode)
			}
			if counterOn(g, jitte, "charge") != 0 {
				t.Fatal("counter was not paid at activation")
			}
			if err := g.ActivateCatalogAbility(me.ID, jitte, 0, game.ActivateAbilityParams{Modes: []int{2}}); err == nil {
				t.Fatal("same charge counter spent twice")
			}
			passPriorityAroundTable(t, g)
			switch mode {
			case 0:
				if effectivePower(t, g, bear) != 4 || effectiveToughness(t, g, bear) != 6 {
					t.Fatal("missing +2/+2")
				}
			case 1:
				if effectivePower(t, g, target) != 1 || effectiveToughness(t, g, target) != 1 {
					t.Fatal("missing -1/-1")
				}
			case 2:
				if me.Life != life+2 {
					t.Fatal("missing life gain")
				}
			}
			advanceToNextSeatsTurn(t, g)
			if effectivePower(t, g, bear) != 2 || effectivePower(t, g, target) != 2 {
				t.Fatal("pump or shrink did not expire")
			}
		})
	}
}

func TestUmezawasJitteTargetsOnlyItsShrinkMode(t *testing.T) {
	g, jitte, _ := jitteSetup(t)
	setCounters(g, jitte, map[string]int{"charge": 1})
	if err := g.ActivateCatalogAbility(g.Seats[0].ID, jitte, 0, game.ActivateAbilityParams{Modes: []int{1}}); err == nil {
		t.Fatal("targetless shrink accepted")
	}
	if counterOn(g, jitte, "charge") != 1 {
		t.Fatal("invalid activation paid its counter")
	}
	victim := b16Creature(g, g.Seats[1].ID, "Victim", "Creature", 1, 1)
	activateJitte(t, g, jitte, 1, victim)
	passPriorityAroundTable(t, g)
	if !g.Seats[1].Graveyard.Contains(victim) {
		t.Fatal("-1/-1 did not kill a 1/1")
	}
}

func TestUmezawasJitteUnattachedModesAndReattachment(t *testing.T) {
	for _, action := range []string{"move", "unattach", "destroy", "blink-source", "blink-host"} {
		t.Run(action, func(t *testing.T) {
			g, jitte, bear := jitteSetup(t)
			me := g.Seats[0]
			other := b16Creature(g, me.ID, "Other", "Creature", 2, 4)
			setCounters(g, jitte, map[string]int{"charge": 1})
			activateJitte(t, g, jitte, 0)
			g.WithWriteLock(func() {
				switch action {
				case "move":
					if err := g.AttachForEffect(jitte, game.TargetRef{Kind: game.TargetCard, ID: other}); err != nil {
						t.Fatal(err)
					}
				case "unattach":
					if err := g.UnattachForEffect(jitte); err != nil {
						t.Fatal(err)
					}
				case "destroy":
					if err := g.DestroyPermanentForEffect(jitte); err != nil {
						t.Fatal(err)
					}
				case "blink-source":
					if err := g.ExileCardForEffect(jitte); err != nil {
						t.Fatal(err)
					}
					returned, err := g.ReturnFromExileToBattlefieldForEffect(jitte, me.ID, false)
					if err != nil {
						t.Fatal(err)
					}
					jitte = returned
					if err := g.AttachForEffect(jitte, game.TargetRef{Kind: game.TargetCard, ID: other}); err != nil {
						t.Fatal(err)
					}
				case "blink-host":
					if err := g.ExileCardForEffect(bear); err != nil {
						t.Fatal(err)
					}
					returned, err := g.ReturnFromExileToBattlefieldForEffect(bear, me.ID, false)
					if err != nil {
						t.Fatal(err)
					}
					bear = returned
				}
			})
			passPriorityAroundTable(t, g)
			wantBear, wantOther := 2, 2
			switch action {
			case "move":
				wantOther = 4
			case "destroy", "blink-source":
				wantBear = 4
			}
			if p := effectivePower(t, g, bear); p != wantBear {
				t.Fatalf("Bear power=%d want %d", p, wantBear)
			}
			if p := effectivePower(t, g, other); p != wantOther {
				t.Fatalf("Other power=%d want %d", p, wantOther)
			}
		})
	}
}

func TestUmezawasJitteCanActivateWhileUnattached(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	jitte := pushCatalogPermanent(g, me.ID, "Umezawa's Jitte", "Legendary Artifact — Equipment", umezawasJitteOracle, false)
	setCounters(g, jitte, map[string]int{"charge": 2})
	activateJitte(t, g, jitte, 0)
	passPriorityAroundTable(t, g)
	before := me.Life
	activateJitte(t, g, jitte, 2)
	passPriorityAroundTable(t, g)
	if me.Life != before+2 || counterOn(g, jitte, "charge") != 0 {
		t.Fatal("unattached modes did not work")
	}
}

func TestUmezawasJitteDepartedSourceAndPendingModeRestore(t *testing.T) {
	g, jitte, bear := jitteSetup(t)
	setCounters(g, jitte, map[string]int{"charge": 1})
	activateJitte(t, g, jitte, 0)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(jitte); err != nil {
			t.Fatal(err)
		}
	})
	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("pending modal ability not restorable: %+v", snap.Continuations)
	}
	raw, err := json.Marshal(snap)
	if err != nil {
		t.Fatal(err)
	}
	var decoded game.GameSnapshot
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	restored, err := decoded.RestoreStrict()
	if err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, restored)
	if effectivePower(t, restored, bear) != 4 {
		t.Fatal("restored pending pump forgot last equipped creature")
	}
}

func TestUmezawasJitteCounterTriggerDoesNotFollowABlinkedSource(t *testing.T) {
	g, jitte, bear := jitteSetup(t)
	me, opp := g.Seats[0], g.Seats[1]
	attackIntoBlocks(t, g, opp.ID, bear)
	advanceTo(t, g, game.StepCombatDamage)
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(jitte); err != nil {
			t.Fatal(err)
		}
		returned, err := g.ReturnFromExileToBattlefieldForEffect(jitte, me.ID, false)
		if err != nil {
			t.Fatal(err)
		}
		jitte = returned
	})
	passPriorityAroundTable(t, g)
	if counterOn(g, jitte, "charge") != 0 {
		t.Fatal("old combat trigger put counters on a new Jitte object")
	}
}

func TestUmezawasJitteCanSpendFirstStrikeCountersBeforeNormalDamage(t *testing.T) {
	g, jitte, bear := jitteSetup(t)
	me, opp := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() {
		if err := (GrantKeywordUntilEOT{Target: bear, Keywords: []string{"double strike"}}).Apply(NewContext(g, &game.StackItem{Controller: me.ID})); err != nil {
			t.Fatal(err)
		}
	})
	attackIntoBlocks(t, g, opp.ID, bear)
	advanceTo(t, g, game.StepFirstStrikeDamage)
	passPriorityAroundTable(t, g)
	if counterOn(g, jitte, "charge") != 2 {
		t.Fatal("first-strike damage did not add two counters")
	}
	activateJitte(t, g, jitte, 0)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if counterOn(g, jitte, "charge") != 3 || opp.Life != 34 {
		t.Fatalf("counter/damage timing: counters=%d defender life=%d", counterOn(g, jitte, "charge"), opp.Life)
	}
}

func TestUmezawasJitteIllegalShrinkTargetKeepsItsCostPaid(t *testing.T) {
	g, jitte, _ := jitteSetup(t)
	setCounters(g, jitte, map[string]int{"charge": 1})
	victim := b16Creature(g, g.Seats[1].ID, "Victim", "Creature", 2, 2)
	activateJitte(t, g, jitte, 1, victim)
	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(victim); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	if counterOn(g, jitte, "charge") != 0 {
		t.Fatal("counter was refunded after the target left")
	}
}
