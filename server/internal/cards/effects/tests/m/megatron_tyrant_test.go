package testsm

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	cardslu "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/lu"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func megatronRow() cards.Card {
	return effectstest.TransformerRow(cardslu.MegatronOracleID,
		effectstest.DisturbFace{Name: "Megatron, Tyrant", TypeLine: "Legendary Artifact Creature — Robot", Cost: "{3}{R}{W}{B}", Power: "7", Toughness: "5"},
		effectstest.DisturbFace{Name: "Megatron, Destructive Force", TypeLine: "Legendary Artifact — Vehicle", Cost: "", Power: "4", Toughness: "5"},
		"More Than Meets the Eye {1}{R}{W}{B}\nYour opponents can't cast spells during combat.\nAt the beginning of each of your postcombat main phases, you may convert Megatron. If you do, add {C} for each 1 life your opponents have lost this turn.",
		"Living metal\nWhenever Megatron attacks, you may sacrifice another artifact. When you do, Megatron deals damage equal to the sacrificed artifact's mana value to target creature. If excess damage would be dealt to that creature this way, instead that damage is dealt to that creature's controller and you convert Megatron.",
		"Living metal", "Convert")
}

// "Your opponents can't cast spells during combat": every combat step,
// opponents only.
func TestMegatronOpponentsCantCastDuringCombat(t *testing.T) {
	g, me, opp := effectstest.FlipMain(t)
	effectstest.TransformerIn(t, g, megatronRow(), me)

	if err := effectstest.E7CastInstant(g, opp); err != nil {
		t.Errorf("an opponent was refused in the main phase: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	for _, step := range []game.Step{game.StepBeginCombat, game.StepDeclareAttackers} {
		effectstest.AdvanceTo(t, g, step)
		effectstest.AssertCantCast(t, effectstest.E7CastInstant(g, opp), "during combat")
	}
	if err := effectstest.E7CastInstant(g, me); err != nil {
		t.Errorf("Megatron's controller was refused in combat: %v", err)
	}
}

// The restriction is lifted once Megatron is the Vehicle (it is only the
// front face's ability).
func TestMegatronDestructiveForceDoesNotLockOpponentsOut(t *testing.T) {
	g, me, opp := effectstest.FlipMain(t)
	effectstest.TransformerConverted(t, g, megatronRow(), me)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if err := effectstest.E7CastInstant(g, opp); err != nil {
		t.Errorf("the Vehicle still locked opponents out: %v", err)
	}
}

// Postcombat: convert and add {C} for all the life the opponents lost.
func TestMegatronConvertsPostcombatForManaPerLifeLost(t *testing.T) {
	g, me, opp := effectstest.FlipMain(t)
	id := effectstest.TransformerIn(t, g, megatronRow(), me)
	other := g.Seats[2]
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(id, opp.ID, -3); err != nil {
			t.Fatal(err)
		}
		if err := g.ChangePlayerLifeForEffect(id, other.ID, -2); err != nil {
			t.Fatal(err)
		}
		me.ManaPool = nil
	})
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	effectstest.TfSettle(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.TfSettle(t, g)
	effectstest.WantFace(t, g, id, 1, "Megatron, Destructive Force")
	if n := me.ManaPool.Total(); n != 5 {
		t.Errorf("%d mana floating, want {C} x5 for 3 + 2 life lost", n)
	}
}

// Declining converts nothing and adds nothing.
func TestMegatronPostcombatDeclined(t *testing.T) {
	g, me, opp := effectstest.FlipMain(t)
	id := effectstest.TransformerIn(t, g, megatronRow(), me)
	g.WithWriteLock(func() {
		_ = g.ChangePlayerLifeForEffect(id, opp.ID, -3)
		me.ManaPool = nil
	})
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	effectstest.TfSettle(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	effectstest.TfSettle(t, g)
	effectstest.WantFace(t, g, id, 0, "Megatron, Tyrant")
	if me.ManaPool.Total() != 0 {
		t.Errorf("%d mana floating after declining", me.ManaPool.Total())
	}
}

// Destructive Force: sacrifice another artifact, damage equal to its
// mana value, excess to the controller and a convert.
func TestMegatronDestructiveForceSacrificesForDamageAndConvertsOnExcess(t *testing.T) {
	for _, tc := range []struct {
		name        string
		cost        string
		wantLife    int // opponent's life change
		wantDead    bool
		wantConvert bool
	}{
		{"excess", "{3}", -1, true, true},
		{"exactly lethal", "{2}", 0, true, false},
		{"short of lethal", "{1}", 0, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, opp := effectstest.FlipMain(t)
			id := effectstest.TransformerConverted(t, g, megatronRow(), me)
			rock := effectstest.PushCatalogPermanent(g, me.ID, "Fuel Rock", "Artifact", "", false)
			g.WithWriteLock(func() {
				for i := range g.Battlefield.Cards {
					if g.Battlefield.Cards[i].InstanceID == rock {
						g.Battlefield.Cards[i].ManaCost = tc.cost
					}
				}
			})
			bear := effectstest.PushVanillaCreature(g, opp.ID, "Target Bear", 2, 2)
			life := opp.Life

			effectstest.TransformerAttacks(t, g, id, opp.ID)
			effectstest.AnswerOwnPermanents(t, g, me.ID, rock)
			effectstest.B04WaitForPick(t, g, me.ID)
			effectstest.PickTriggerTarget(t, g, me.ID, bear)
			effectstest.TfSettle(t, g)

			if zone, _ := effectstest.CardWhere(g, me, rock); zone != "graveyard" {
				t.Errorf("the sacrificed artifact is in %s", zone)
			}
			if dead := !effectstest.OnBattlefieldNow(g, bear); dead != tc.wantDead {
				t.Errorf("bear dead = %v, want %v", dead, tc.wantDead)
			}
			if opp.Life-life != tc.wantLife {
				t.Errorf("opponent life changed by %d, want %d", opp.Life-life, tc.wantLife)
			}
			if tc.wantConvert {
				effectstest.WantFace(t, g, id, 0, "Megatron, Tyrant")
			} else {
				effectstest.WantFace(t, g, id, 1, "Megatron, Destructive Force")
			}
		})
	}
}

// Declining the sacrifice does nothing.
func TestMegatronDestructiveForceMaySkipTheSacrifice(t *testing.T) {
	g, me, opp := effectstest.FlipMain(t)
	id := effectstest.TransformerConverted(t, g, megatronRow(), me)
	rock := effectstest.PushCatalogPermanent(g, me.ID, "Fuel Rock", "Artifact", "", false)
	effectstest.TransformerAttacks(t, g, id, opp.ID)
	effectstest.AnswerOwnPermanents(t, g, me.ID)
	effectstest.TfSettle(t, g)
	if zone, _ := effectstest.CardWhere(g, me, rock); zone != "battlefield" {
		t.Errorf("declined, yet the artifact is in %s", zone)
	}
	effectstest.WantFace(t, g, id, 1, "Megatron, Destructive Force")
}
