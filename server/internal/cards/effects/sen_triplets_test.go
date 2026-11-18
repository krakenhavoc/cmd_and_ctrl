package effects

import (
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

const senTripletsOracle = "f0c5864f-83c3-4d99-8f97-0ab92608fff3"

func TestSenTripletsIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(senTripletsOracle)
	if !ok || spec.Name != "Sen Triplets" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("spec = %+v, %v", spec, ok)
	}
}

// senTripletsUpkeep puts the Triplets under seat 1 and walks to seat 1's
// upkeep, answering the trigger's target prompt with seat 2. It checks
// that the prompt offers opponents only.
func senTripletsResolve(t *testing.T, g *game.Game) (me, opp, third *game.Player) {
	t.Helper()
	me, opp, third = g.Seats[1], g.Seats[2], g.Seats[3]
	b12Push(g, me.ID, "Sen Triplets", "Legendary Artifact Creature — Human Wizard", senTripletsOracle, 3, 3)
	advanceToUpkeepOf(t, g, 1)
	for i := 0; i < 16; i++ {
		passPriorityAroundTable(t, g)
		if p := latestPickTarget(g, me.ID); p != nil {
			if slices.Contains(p.PickTargetPlayers, me.ID) {
				t.Fatal("the trigger offered its controller as the target opponent")
			}
			if !slices.Contains(p.PickTargetPlayers, opp.ID) {
				t.Fatalf("the trigger did not offer the opponent: %v", p.PickTargetPlayers)
			}
			if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			continue
		}
		if stackFullyEmpty(g) {
			break
		}
	}
	if !stackFullyEmpty(g) {
		t.Fatal("the trigger did not resolve")
	}
	if g.Turn.Step != game.StepUpkeep || g.Turn.ActiveSeat != 1 {
		t.Fatalf("left the upkeep: step %v, seat %d", g.Turn.Step, g.Turn.ActiveSeat)
	}
	return me, opp, third
}

func senHandRevealed(g *game.Game, p uuid.UUID) []string {
	var out []string
	g.ReadSnapshot(func() { out = g.HandRevealedLocked(p) })
	return out
}

// The whole trigger: the opponent can't cast or activate, their hand is
// revealed, and the Triplets' controller casts out of it at the card's
// own timing, paying its cost, controlling the spell; the card goes to
// its owner's graveyard. Nobody else may.
func TestSenTripletsLocksTheOpponentAndCastsFromTheirHand(t *testing.T) {
	g := newCatalogGame(t)
	oppInstant0 := pushHandCardWithManaCost(g.Seats[2], "Their Shock", "Instant", "{1}")
	me, opp, third := senTripletsResolve(t, g)
	sorcery := pushHandCardWithManaCost(opp, "Their Sorcery", "Sorcery", "{2}")
	instant := pushHandCardWithManaCost(opp, "Their Instant", "Instant", "{1}")
	forest := pushLandFor(g, opp.ID, "Their Forest", "Basic Land — Forest")

	if got := senHandRevealed(g, opp.ID); !slices.Contains(got, "Sen Triplets") {
		t.Errorf("the opponent's hand is not revealed: %v", got)
	}
	if got := senHandRevealed(g, third.ID); len(got) != 0 {
		t.Errorf("a third player's hand is revealed: %v", got)
	}

	// The opponent can't cast spells...
	fillPool(opp, 3)
	if err := g.CastSpell(opp.ID, oppInstant0, game.CastSpellParams{FromZone: "hand", Strict: true}); !errors.Is(err, game.ErrCantCast) {
		t.Errorf("the opponent cast a spell: err = %v, want ErrCantCast", err)
	}
	// ...or activate abilities, a mana ability included.
	if err := g.ActivateManaAbility(opp.ID, forest, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("the opponent activated a mana ability")
	}
	// A third player is under no ban.
	thirdForest := pushLandFor(g, third.ID, "Third Forest", "Basic Land — Forest")
	if err := g.ActivateManaAbility(third.ID, thirdForest, 0, game.ManaAbilityParams{}); err != nil {
		t.Errorf("a third player could not tap a Forest: %v", err)
	}

	// The card's own timing holds: a sorcery waits for the main phase,
	// an instant does not.
	fillPool(me, 3)
	if err := g.CastSpell(me.ID, sorcery, game.CastSpellParams{FromZone: "hand", Strict: true}); !errors.Is(err, game.ErrSorcerySpeedRequired) {
		t.Errorf("a sorcery from their hand in the upkeep: err = %v, want ErrSorcerySpeedRequired", err)
	}
	// A third player has no permission over that hand.
	fillPool(third, 3)
	if err := g.CastSpell(third.ID, instant, game.CastSpellParams{FromZone: "hand", Strict: true}); err == nil {
		t.Error("a third player cast out of the opponent's hand")
	}
	if err := g.CastSpell(me.ID, instant, game.CastSpellParams{FromZone: "hand", Strict: true}); err != nil {
		t.Fatalf("the instant from their hand: %v", err)
	}
	if !g.Stack.Contains(instant) || opp.Hand.Contains(instant) {
		t.Fatal("the instant is not on the stack")
	}
	var ctrl uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Stack.Cards {
			if c.InstanceID == instant {
				ctrl = c.Controller
			}
		}
	})
	if ctrl != me.ID {
		t.Errorf("the spell's controller is %v, want the Triplets' controller", ctrl)
	}
	if me.ManaPool.Total() != 2 {
		t.Errorf("my pool holds %d after a {1} spell from 3, want 2 (I pay its cost)", me.ManaPool.Total())
	}
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(instant) {
		t.Error("the resolved card did not go to its owner's graveyard")
	}

	// In the main phase the sorcery is castable, for its cost.
	advanceToMain(t, g)
	me.ManaPool.EmptyPool()
	if err := g.CastSpell(me.ID, sorcery, game.CastSpellParams{FromZone: "hand", Strict: true}); err == nil {
		t.Error("cast the sorcery from their hand without paying for it")
	}
	fillPool(me, 2)
	if err := g.CastSpell(me.ID, sorcery, game.CastSpellParams{FromZone: "hand", Strict: true}); err != nil {
		t.Fatalf("the sorcery from their hand in my main phase: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(sorcery) {
		t.Error("the sorcery did not go to its owner's graveyard")
	}
}

// A land from their hand is played with my land drop, under my control,
// and only one of them (CR 305.2).
func TestSenTripletsPlaysALandFromTheirHandWithMyLandDrop(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := senTripletsResolve(t, g)
	island := pushHandCardWithManaCost(opp, "Their Island", "Basic Land — Island", "")
	swamp := pushHandCardWithManaCost(opp, "Their Swamp", "Basic Land — Swamp", "")
	if err := g.CastSpell(me.ID, island, game.CastSpellParams{FromZone: "hand"}); !errors.Is(err, game.ErrSorcerySpeedRequired) {
		t.Errorf("a land played in the upkeep: err = %v, want ErrSorcerySpeedRequired", err)
	}
	advanceToMain(t, g)
	// The bot enumerator offers the play, as it would one from my hand.
	if !senOffersMoveFor(g, me.ID, island) {
		t.Error("the enumerator does not offer playing their land")
	}
	if err := g.CastSpell(me.ID, island, game.CastSpellParams{FromZone: "hand"}); err != nil {
		t.Fatalf("playing a land from their hand: %v", err)
	}
	var onField game.Card
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == island {
				onField = c
			}
		}
	})
	if onField.InstanceID != island || onField.Controller != me.ID || onField.Owner != opp.ID {
		t.Fatalf("the land is not on the battlefield under my control: %+v", onField)
	}
	if got := g.LandsPlayedThisTurnFor(me.ID); got != 1 {
		t.Errorf("my lands played this turn = %d, want 1", got)
	}
	if err := g.CastSpell(me.ID, swamp, game.CastSpellParams{FromZone: "hand"}); !errors.Is(err, game.ErrLandDropUnavailable) {
		t.Errorf("a second land: err = %v, want ErrLandDropUnavailable", err)
	}
	if senOffersMoveFor(g, me.ID, swamp) {
		t.Error("the enumerator offers a second land play")
	}
}

// "This turn": at the next turn the opponent casts again, the hand is
// hidden again, and the permission is gone. A card that reached the hand
// during the turn was covered while it lasted.
func TestSenTripletsEndsWithTheTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := senTripletsResolve(t, g)
	late := pushHandCardWithManaCost(opp, "Drawn Later", "Instant", "{0}")
	if !senOffersMoveFor(g, me.ID, late) {
		t.Error("a card that arrived in the hand this turn is not castable from it")
	}
	advanceToUpkeepOf(t, g, 2)
	passPriorityAroundTable(t, g)
	if got := senHandRevealed(g, opp.ID); len(got) != 0 {
		t.Errorf("the hand is still revealed next turn: %v", got)
	}
	if err := g.CastSpell(me.ID, late, game.CastSpellParams{FromZone: "hand", Strict: true}); err == nil {
		t.Error("cast from their hand on a later turn")
	}
	if err := g.CastSpell(opp.ID, late, game.CastSpellParams{FromZone: "hand", Strict: true}); err != nil {
		t.Errorf("the opponent cannot cast on their own turn: %v", err)
	}
}

// senOffersMoveFor reports whether the enumerator offers seat a move on
// the card.
func senOffersMoveFor(g *game.Game, seat, card uuid.UUID) bool {
	for _, m := range legal.EnumerateFor(g, seat) {
		if m.Type == legal.TypeCastSpell && m.Source == card {
			return true
		}
	}
	return false
}
