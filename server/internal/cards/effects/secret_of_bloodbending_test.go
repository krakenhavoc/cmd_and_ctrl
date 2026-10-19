package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// secret_of_bloodbending_test.go — #3999: Secret of Bloodbending's
// combat-phase control and its optional "waterbend {10}". The machinery
// is pinned in game/combat_control_test.go and actions/
// controlled_combat_test.go; these tests pin the card's use of it.

const secretOfBloodbendingOracle = "743fa824-18cb-4225-ba7b-5b0bbc3c8d41"

// castSecret puts Secret of Bloodbending in the active seat's hand, goes
// to the main phase and casts it at `target` with `params` (targets
// filled in). It returns the spell's ID and the cast's error.
func castSecret(t *testing.T, g *game.Game, target uuid.UUID, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Secret of Bloodbending", TypeLine: "Sorcery — Lesson",
		ManaCost: "{U}{U}{U}{U}", OracleID: secretOfBloodbendingOracle, Owner: me.ID, Controller: me.ID,
	})
	advanceToMain(t, g)
	params.Targets = []game.TargetRef{{Kind: game.TargetPlayer, ID: target}}
	return id, g.CastSpell(me.ID, id, params)
}

// stepToSeat walks the cursor until `seat` is active at `step`.
func stepToSeat(t *testing.T, g *game.Game, seat int, step game.Step) {
	t.Helper()
	for i := 0; i < 400; i++ {
		if g.Turn.Step == step && g.Turn.ActiveSeat == seat {
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep toward %s of seat %d: %v", step, seat, err)
		}
	}
	t.Fatalf("never reached %s of seat %d", step, seat)
}

func addPool(t *testing.T, g *game.Game, who uuid.UUID, mana string) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.AddManaForEffect(who, uuid.Nil, mana); err != nil {
			t.Fatalf("AddManaForEffect(%s): %v", mana, err)
		}
	})
}

// Unpaid: the caster controls the target during their next combat phase
// only, and the spell exiles itself.
func TestSecretOfBloodbendingControlsTheirNextCombatPhase(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	oppSeat := (g.Turn.ActiveSeat + 1) % len(g.Seats)
	opp := g.Seats[oppSeat]
	id, err := castSecret(t, g, opp.ID, game.CastSpellParams{})
	if err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !inExile(g, id) {
		t.Error("Secret of Bloodbending did not exile itself")
	}
	w := scheduledControls(g)
	if len(w) != 1 || w[0].Controller != me.ID || w[0].Controlled != opp.ID || w[0].UntilKind() != game.ControlUntilCombat || w[0].SourceName != "Secret of Bloodbending" {
		t.Fatalf("scheduled %+v, want a combat-phase control of %s by %s", w, opp.Name, me.Name)
	}

	stepToSeat(t, g, oppSeat, game.StepPrecombatMain)
	if got := g.ControllerOfPlayer(opp.ID); got != opp.ID {
		t.Fatalf("the target is controlled before their combat phase: %v", got)
	}
	stepToSeat(t, g, oppSeat, game.StepDeclareAttackers)
	if got := g.ControllerOfPlayer(opp.ID); got != me.ID {
		t.Fatalf("the target's combat phase is controlled by %v, want the caster", got)
	}
	if got := g.ControlUntilOfPlayer(opp.ID); got != game.ControlUntilCombat {
		t.Errorf("control lasts %q, want one combat phase", got)
	}
	stepToSeat(t, g, oppSeat, game.StepPostcombatMain)
	if got := g.ControllerOfPlayer(opp.ID); got != opp.ID {
		t.Errorf("the control outlived the combat phase: %v", got)
	}
	if len(g.PlayerControls) != 0 {
		t.Errorf("records left: %+v", g.PlayerControls)
	}
}

// Paid: "you control that player during their next turn instead", with
// no extra turn after it. Ten untapped creatures pay the whole {10}, so
// a strict cast spends only {U}{U}{U}{U} from the pool.
func TestSecretOfBloodbendingWaterbendPaidControlsTheirNextTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	taps := pushTapCostSoldiers(g, me.ID, 10)
	advanceToMain(t, g)
	addPool(t, g, me.ID, "{U}{U}{U}{U}")
	if _, err := castSecret(t, g, opp.ID, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}, TapIDs: taps}); err != nil {
		t.Fatalf("CastSpell with waterbend {10} paid by ten taps: %v", err)
	}
	for _, id := range taps {
		if !tapCostTapped(g, id) {
			t.Fatalf("a creature named to waterbend is untapped")
		}
	}
	passPriorityAroundTable(t, g)
	w := scheduledControls(g)
	if len(w) != 1 || w[0].UntilKind() != game.ControlUntilTurn || w[0].ExtraTurn != nil || w[0].Controlled != opp.ID {
		t.Fatalf("scheduled %+v, want a turn-long control of %s with no extra turn", w, opp.Name)
	}
	walkControlledTurnWithNoExtraTurn(t, g, me, opp)
}

// The waterbend's taps pay its {10} and nothing else, and only when the
// cost is chosen (CR 701.67b).
func TestSecretOfBloodbendingWaterbendPayment(t *testing.T) {
	cases := []struct {
		name     string
		pool     string
		taps     int
		optional bool
		ok       bool
	}{
		{"ten taps and the printed cost", "{U}{U}{U}{U}", 10, true, true},
		{"six taps and four mana", "{U}{U}{U}{U}{C}{C}{C}{C}", 6, true, true},
		{"nine taps leave {1} unpaid", "{U}{U}{U}{U}", 9, true, false},
		{"eleven taps are more than the waterbend is worth", "{U}{U}{U}{U}", 11, true, false},
		{"taps cannot pay the printed {U}{U}{U}{U}", "{U}{U}", 12, true, false},
		{"taps without choosing the waterbend", "{U}{U}{U}{U}", 1, false, false},
		{"no waterbend, the printed cost alone", "{U}{U}{U}{U}", 0, false, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			taps := pushTapCostSoldiers(g, me.ID, tc.taps)
			advanceToMain(t, g)
			addPool(t, g, me.ID, tc.pool)
			params := game.CastSpellParams{Strict: true, TapIDs: taps}
			if tc.optional {
				params.OptionalCosts = []int{0}
			}
			_, err := castSecret(t, g, opp.ID, params)
			if tc.ok && err != nil {
				t.Fatalf("refused: %v", err)
			}
			if !tc.ok && err == nil {
				t.Fatal("accepted")
			}
		})
	}
}

// The declared cost is the waterbend {10}, and the card ships Full.
func TestSecretOfBloodbendingOffersTheWaterbend(t *testing.T) {
	costs := game.OptionalCostsFor(secretOfBloodbendingOracle)
	if len(costs) != 1 {
		t.Fatalf("optional costs = %+v, want the one waterbend", costs)
	}
	oc := costs[0]
	if oc.Key != game.WaterbendKey || oc.ManaCost != "{10}" || oc.Waterbend.Empty() || oc.Waterbend.Extra != "{10}" || oc.Label != "Waterbend {10}" {
		t.Errorf("optional cost = %+v", oc)
	}
	if spec, ok := Lookup(secretOfBloodbendingOracle); !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Secret of Bloodbending is not Full without caveats: %v %v", spec.Completeness, spec.Caveats)
	}
}
