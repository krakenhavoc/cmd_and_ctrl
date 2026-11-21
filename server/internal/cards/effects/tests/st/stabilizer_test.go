package testsst

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// stabilizer_test.go — #5843, "Players can't cycle cards." Driven
// through the real catalog and the real activation path: the engine's
// refusal, the enumerator's offer (the bot's half, #544) and the view's
// cant_activate stamp (the client's greyed row) must all agree, for the
// Stabilizer's controller and for an opponent, for cycling and for
// typecycling (CR 702.29f), and all three must lift once Stabilizer
// leaves the battlefield.

const stabilizerOracle = "856c79ae-feab-4df9-8269-fffee7a44193"

const stabilizerClause = "Stabilizer — players can't cycle cards."

// stabilizerHandRowReason is the cant_activate stamp on row 0 of hand
// card `id`, as `viewer` receives it on the wire.
func stabilizerHandRowReason(t *testing.T, g *game.Game, viewer, id uuid.UUID) string {
	t.Helper()
	v := protocol.FilterViewFor(protocol.ViewOfGame(g), viewer.String())
	for _, seat := range v.Seats {
		for _, c := range seat.Hand.Cards {
			if c.InstanceID != id.String() {
				continue
			}
			if len(c.ZoneAbilities) == 0 {
				t.Fatalf("%s has no zone ability rows in %s's view", id, viewer)
			}
			return c.ZoneAbilities[0].CantActivate
		}
	}
	t.Fatalf("%s is not in %s's hand view", id, viewer)
	return ""
}

// stabilizerSeedCycler puts a cycling card in `p`'s hand and floats
// the mana its cycling cost needs, so a refusal can only be the gate.
func stabilizerSeedCycler(t *testing.T, g *game.Game, p *game.Player, name, typeLine, oracle, mana string) uuid.UUID {
	t.Helper()
	id := effectstest.PushCatalogHandCard(p, name, typeLine, oracle)
	if err := g.AddManaForEffect(p.ID, uuid.Nil, mana); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	return id
}

func TestStabilizerStopsEveryPlayerCyclingUntilItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	them := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	stabilizer := effectstest.B21Push(g, me.ID, "Stabilizer", "Artifact", stabilizerOracle, 0, 0)

	// Cycling for the controller, cycling for an opponent, and
	// typecycling (Sylvan Reclamation's basic landcycling), which CR
	// 702.29f says the same clause stops.
	mine := stabilizerSeedCycler(t, g, me, "Ketria Triome", "Land — Forest Island Mountain", effectstest.KetriaTriomeOracle, "{C}{C}{C}")
	landcycler := stabilizerSeedCycler(t, g, me, "Sylvan Reclamation", "Instant", effectstest.SylvanReclamationOracle, "{C}{C}")
	theirs := stabilizerSeedCycler(t, g, them, "Ketria Triome", "Land — Forest Island Mountain", effectstest.KetriaTriomeOracle, "{C}{C}{C}")

	for _, tc := range []struct {
		what string
		p    *game.Player
		id   uuid.UUID
	}{
		{"the controller's cycling", me, mine},
		{"the controller's basic landcycling", me, landcycler},
		{"an opponent's cycling", them, theirs},
	} {
		err := g.ActivateCatalogAbility(tc.p.ID, tc.id, 0, game.ActivateAbilityParams{})
		var cant *game.CantActivateError
		if !errors.As(err, &cant) || cant.Reason != stabilizerClause || cant.Source != stabilizer {
			t.Errorf("%s: err = %v, want Stabilizer's CantActivateError", tc.what, err)
		}
		if !tc.p.Hand.Contains(tc.id) {
			t.Errorf("%s: the card left the hand, so a cost was paid for a refused activation", tc.what)
		}
		if why := stabilizerHandRowReason(t, g, tc.p.ID, tc.id); why != stabilizerClause {
			t.Errorf("%s: the view stamps cant_activate %q, want %q", tc.what, why, stabilizerClause)
		}
	}
	if effectstest.ActOffersActivationFrom(g, me.ID, mine) || effectstest.ActOffersActivationFrom(g, me.ID, landcycler) {
		t.Error("the enumerator offers a cycling activation Stabilizer refuses")
	}

	// It stops cycling and nothing else: a non-cycling activated
	// ability on the same board is untouched.
	krenko := effectstest.ActPushKrenko(g, me.ID)
	if err := g.ActivateCatalogAbility(me.ID, krenko, 0, game.ActivateAbilityParams{}); errors.Is(err, game.ErrCantActivate) {
		t.Errorf("Krenko's {T} under Stabilizer: err = %v, want no activation ban — only cycling is stopped", err)
	}

	// Stabilizer leaves: nothing was stored, so the next query is clean.
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(stabilizer); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
	if g.Battlefield.Contains(stabilizer) {
		t.Fatal("Stabilizer is still on the battlefield")
	}
	if why := stabilizerHandRowReason(t, g, them.ID, theirs); why != "" {
		t.Errorf("with Stabilizer gone, the opponent's cycling row is greyed with %q", why)
	}
	if !effectstest.ActOffersActivationFrom(g, me.ID, mine) {
		t.Error("with Stabilizer gone, the enumerator does not offer cycling")
	}
	if err := g.ActivateCatalogAbility(me.ID, mine, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("cycling with Stabilizer gone: err = %v, want nil", err)
	}
	if !me.Graveyard.Contains(mine) {
		t.Error("the cycled card is not in the graveyard")
	}
	if err := g.ActivateCatalogAbility(me.ID, landcycler, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("landcycling with Stabilizer gone: err = %v, want nil", err)
	}
	if err := g.ActivateCatalogAbility(them.ID, theirs, 0, game.ActivateAbilityParams{}); errors.Is(err, game.ErrCantActivate) {
		t.Errorf("the opponent's cycling with Stabilizer gone: err = %v, want no activation ban", err)
	}
}
