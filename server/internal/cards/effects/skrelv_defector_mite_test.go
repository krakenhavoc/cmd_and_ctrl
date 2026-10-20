package effects

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const skrelvOracle = "20053847-6623-493c-8cdb-a69cda3b1577"

func skrelvSetup(t *testing.T) (*game.Game, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	skrelv := pushCatalogPermanent(g, me.ID, "Skrelv, Defector Mite", "Legendary Artifact Creature — Phyrexian Mite", skrelvOracle, false)
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 4)
	return g, skrelv, bear
}

func skrelvActivate(t *testing.T, g *game.Game, source, target uuid.UUID, color string) {
	t.Helper()
	if err := g.ActivateCatalogAbility(g.Seats[0].ID, source, 0, game.ActivateAbilityParams{Targets: cardRefs(target), PhyrexianLife: 1, Strict: true}); err != nil {
		t.Fatal(err)
	}
	if pendingOfKind(g, game.PendingChoiceColor) != nil {
		t.Fatal("color chosen at announcement instead of resolution")
	}
	passPriorityAroundTable(t, g)
	answerColor(t, g, g.Seats[0].ID, color)
}

func pairRefusal(g *game.Game, attacker, blocker uuid.UUID) game.BlockRefusal {
	var result game.BlockRefusal
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		a, _ := g.LookupCardForEffect(attacker)
		b, _ := g.LookupCardForEffect(blocker)
		result = g.BlockPairRefusalLocked(&a, &b)
	})
	return result
}

func toxicOn(g *game.Game, id uuid.UUID) int {
	n := 0
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked(); c, _ := g.LookupCardForEffect(id); n = game.ToxicTotal(&c) })
	return n
}

func TestSkrelvPaysEitherWhiteOrLifeAndChoosesColorAtResolution(t *testing.T) {
	for _, life := range []bool{false, true} {
		t.Run(map[bool]string{false: "white", true: "life"}[life], func(t *testing.T) {
			g, skrelv, bear := skrelvSetup(t)
			me, opp := g.Seats[0], g.Seats[1]
			before := me.Life
			params := game.ActivateAbilityParams{Targets: cardRefs(bear), Strict: true}
			if life {
				params.PhyrexianLife = 1
			} else {
				g.WithWriteLock(func() {
					if err := g.AddManaForEffect(me.ID, uuid.Nil, "{W}"); err != nil {
						t.Fatal(err)
					}
				})
			}
			if err := g.ActivateCatalogAbility(me.ID, skrelv, 0, params); err != nil {
				t.Fatal(err)
			}
			expected := before
			if life {
				expected -= 2
			}
			if me.Life != expected || !b16Tapped(t, g, skrelv) {
				t.Fatalf("wrong payment: life=%d tapped=%v", me.Life, b16Tapped(t, g, skrelv))
			}
			if pendingOfKind(g, game.PendingChoiceColor) != nil {
				t.Fatal("premature color choice")
			}
			passPriorityAroundTable(t, g)
			prompt := pendingOfKind(g, game.PendingChoiceColor)
			if prompt == nil {
				t.Fatal("no color choice at resolution")
			}
			if err := g.ResolveColorChoice(prompt.ID, me.ID, "C"); err == nil {
				t.Fatal("colorless accepted as a color")
			}
			answerColor(t, g, me.ID, "R")
			if toxicOn(g, bear) != 1 || toxicOn(g, skrelv) != 1 {
				t.Fatal("missing toxic 1 on source or recipient")
			}
			if creatureTargetableFrom(g, opp.ID, bear, srcRed) {
				t.Fatal("opponent's red source can target hexproof-from-red creature")
			}
			if !creatureTargetableFrom(g, opp.ID, bear, srcBlue) || !creatureTargetableFrom(g, me.ID, bear, srcRed) {
				t.Fatal("hexproof incorrectly barred another color or own controller")
			}
			if hasEffectiveKeyword(t, g, bear, "protection from red") {
				t.Fatal("Skrelv incorrectly granted protection")
			}
			red := b16Creature(g, opp.ID, "Red", "Creature", 2, 2, "R")
			gold := b16Creature(g, opp.ID, "Gold", "Creature", 2, 2, "R", "U")
			blue := b16Creature(g, opp.ID, "Blue", "Creature", 2, 2, "U")
			colorless := b16Creature(g, opp.ID, "Colorless", "Artifact Creature", 2, 2)
			for _, id := range []uuid.UUID{red, gold} {
				if pairRefusal(g, bear, id).Reason != game.BlockReasonCantBeBlockedBy {
					t.Fatal("red blocker allowed")
				}
			}
			for _, id := range []uuid.UUID{blue, colorless} {
				if !pairRefusal(g, bear, id).Legal() {
					t.Fatal("unrelated blocker refused")
				}
			}
			// This is hexproof and evasion, not damage prevention.
			g.WithWriteLock(func() {
				if err := (DealDamage{Source: red, Target: bear, Amount: 1}).Apply(NewContext(g, &game.StackItem{Controller: opp.ID, SourceCardID: red})); err != nil {
					t.Fatal(err)
				}
			})
			c, _ := battlefieldCard(g, bear)
			if c.DamageMarked != 1 {
				t.Fatal("red damage was incorrectly prevented")
			}
		})
	}
}

func TestSkrelvRejectsSelfOpponentAndMissingPayment(t *testing.T) {
	for _, kind := range []string{"self", "opponent", "unpaid", "sick"} {
		t.Run(kind, func(t *testing.T) {
			g, skrelv, bear := skrelvSetup(t)
			me := g.Seats[0]
			target := bear
			params := game.ActivateAbilityParams{PhyrexianLife: 1, Strict: true}
			switch kind {
			case "self":
				target = skrelv
			case "opponent":
				target = b16Creature(g, g.Seats[1].ID, "Other", "Creature", 2, 2)
			case "unpaid":
				params.PhyrexianLife = 0
			case "sick":
				for i := range g.Battlefield.Cards {
					if g.Battlefield.Cards[i].InstanceID == skrelv {
						g.Battlefield.Cards[i].SummonedThisTurn = true
					}
				}
			}
			params.Targets = cardRefs(target)
			life := me.Life
			if err := g.ActivateCatalogAbility(me.ID, skrelv, 0, params); err == nil {
				t.Fatal("illegal activation accepted")
			}
			if me.Life != life || b16Tapped(t, g, skrelv) {
				t.Fatal("refused activation paid a cost")
			}
		})
	}
}

func TestSkrelvIllegalTargetDoesNotAskForColor(t *testing.T) {
	g, skrelv, bear := skrelvSetup(t)
	me := g.Seats[0]
	if err := g.ActivateCatalogAbility(me.ID, skrelv, 0, game.ActivateAbilityParams{PhyrexianLife: 1, Targets: cardRefs(bear), Strict: true}); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(bear); err != nil {
			t.Fatal(err)
		}
		returned, err := g.ReturnFromExileToBattlefieldForEffect(bear, me.ID, false)
		if err != nil {
			t.Fatal(err)
		}
		bear = returned
	})
	passPriorityAroundTable(t, g)
	if pendingOfKind(g, game.PendingChoiceColor) != nil || toxicOn(g, bear) != 0 {
		t.Fatal("blinked target received the old ability")
	}
}

func TestSkrelvToxicStacksDealsPoisonAndExpires(t *testing.T) {
	g, skrelv, bear := skrelvSetup(t)
	me, opp := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() {
		if err := (GrantKeywordUntilEOT{Target: bear, Keywords: []string{"toxic 1"}}).Apply(NewContext(g, &game.StackItem{Controller: me.ID})); err != nil {
			t.Fatal(err)
		}
	})
	skrelvActivate(t, g, skrelv, bear, "R")
	if toxicOn(g, bear) != 2 {
		t.Fatal("toxic grants did not stack")
	}
	attackIntoBlocks(t, g, opp.ID, bear)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if poisonOn(opp) != 2 {
		t.Fatalf("poison=%d, want 2", poisonOn(opp))
	}
	advanceToNextSeatsTurn(t, g)
	if toxicOn(g, bear) != 0 || hasEffectiveKeyword(t, g, bear, game.HexproofFromColor("R")) {
		t.Fatal("grant outlasted the turn")
	}
}

func TestSkrelvGrantSurvivesSourceDepartureAndSnapshot(t *testing.T) {
	g, skrelv, bear := skrelvSetup(t)
	opp := g.Seats[1]
	red := b16Creature(g, opp.ID, "Red", "Creature", 2, 2, "R")
	skrelvActivate(t, g, skrelv, bear, "R")
	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(skrelv); err != nil {
			t.Fatal(err)
		}
	})
	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("not restorable: %+v", snap.Continuations)
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
	if toxicOn(restored, bear) != 1 || pairRefusal(restored, bear, red).Legal() {
		t.Fatal("grant lost when source left or after restore")
	}
	advanceToNextSeatsTurn(t, restored)
	if toxicOn(restored, bear) != 0 || !pairRefusal(restored, bear, red).Legal() {
		t.Fatal("restored grant did not expire")
	}
}

func TestSkrelvCannotBlockAndItsPrintedToxicWorks(t *testing.T) {
	g, skrelv, _ := skrelvSetup(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := b16Creature(g, opp.ID, "Attacker", "Creature", 2, 2)
	if pairRefusal(g, attacker, skrelv).Legal() {
		t.Fatal("Skrelv can block")
	}
	attackIntoBlocks(t, g, opp.ID, skrelv)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if poisonOn(opp) != 1 || me.Life != 40 {
		t.Fatalf("printed toxic failed: poison=%d", poisonOn(opp))
	}
}
