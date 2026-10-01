package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// granted_rebound_test.go — a spell GIVEN rebound (ADR 0107 §3
// decision 5, owner decision 3, #1854): the stack step of the layer
// pass is layers 2 and 6, keywords only (CR 613.1f).
//
// What a bug here would hide, worst first:
//
//  1. Rebound read off the wrong list. rebound.go asks HasKeyword of
//     the resolving spell; if the grant is not on that list, a spell
//     "that gains rebound" goes to the graveyard and nothing says so.
//  2. A static over spells leaking onto permanents, or the other way.
//  3. The grant outliving the spell (CR 400.7): a card that rebounded
//     once rebounding forever.
//  4. CR 616.1: a spell with rebound and buyback, or rebound and an
//     Adventure, quietly applying one of them on the controller's
//     behalf.
//  5. The restore point: a spell given rebound restored without it.

const testSpellStaticOracle = "test-instant-and-sorcery-spells-have-rebound"

// withReboundForYourSpells installs a Cast Through Time stand-in:
// "Instant and sorcery spells you control have rebound", a battlefield
// static over spells.
func withReboundForYourSpells(t *testing.T) {
	t.Helper()
	withStaticAbilities(t, func(key string) []StaticAbility {
		if key != testSpellStaticOracle {
			return nil
		}
		return []StaticAbility{{
			Layer:         Layer6Ability,
			AffectsSpells: true,
			AppliesTo: func(target *Card, _ *Game, source *Card) bool {
				return target.Controller == source.Controller &&
					(target.HasCardType("instant") || target.HasCardType("sorcery"))
			},
			Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
				c.Abilities = AppendKeywordAbility(c.Abilities, KeywordRebound)
			},
		}}
	})
}

// pushSpellStaticSource puts the Cast Through Time stand-in on p's
// battlefield.
func pushSpellStaticSource(g *Game, p *Player) uuid.UUID {
	return pushTypedTestCard(g, Card{
		Name:       "Cast Through Time",
		TypeLine:   "Enchantment",
		OracleID:   testSpellStaticOracle,
		Owner:      p.ID,
		Controller: p.ID,
	})
}

// handSpell puts a plain spell of `typeLine` (no rebound of its own)
// in p's hand.
func handSpell(g *Game, p *Player, typeLine string) uuid.UUID {
	c := NewCard("Plain "+typeLine, p.ID)
	c.TypeLine = typeLine
	c.ManaCost = "{0}"
	g.WithWriteLock(func() { p.Hand.PushTop(c) })
	return c.InstanceID
}

func castSpellFromHand(t *testing.T, g *Game, p *Player, id uuid.UUID, params CastSpellParams) {
	t.Helper()
	if err := g.CastSpell(p.ID, id, params); err != nil {
		t.Fatalf("cast from hand: %v", err)
	}
}

// grantRebound runs "that spell gains rebound" under the write lock.
func grantRebound(t *testing.T, g *Game, spell uuid.UUID) bool {
	t.Helper()
	var ok bool
	g.WithWriteLock(func() {
		ok = g.GrantKeywordsToSpellForEffect(uuid.New(), spell, []string{KeywordRebound}, "test — that spell gains rebound")
	})
	return ok
}

// stackSpellHasRebound is HasKeyword on the spell's stack card, after a
// recompute — the read the resolution makes.
func stackSpellHasRebound(t *testing.T, g *Game, id uuid.UUID) bool {
	t.Helper()
	var has bool
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		c, ok := g.stackCardLocked(id)
		if !ok {
			t.Fatalf("%s is not on the stack", id)
		}
		has = HasKeyword(c, KeywordRebound) && containsKeyword(c.Effective().Abilities, KeywordRebound)
	})
	return has
}

func resolutionChoiceOutstanding(g *Game) *PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceOptionPick {
			return c
		}
	}
	return nil
}

// CR 613.1f, 702.88a: "that spell gains rebound" is an ability-adding
// effect on the spell, so HasKeyword finds it and the resolution
// exiles the card with a delayed trigger — no rebound-specific read.
func TestASpellGivenReboundOnTheStackRebounds(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	id := handSpell(g, me, "Instant")
	castSpellFromHand(t, g, me, id, CastSpellParams{})
	if stackSpellHasRebound(t, g, id) {
		t.Fatal("setup: the spell has rebound before anything gave it rebound")
	}
	if !grantRebound(t, g, id) {
		t.Fatal("GrantKeywordsToSpellForEffect refused a spell on the stack")
	}
	if !stackSpellHasRebound(t, g, id) {
		t.Fatal("the spell does not have the rebound it was given")
	}
	resolveTop(t, g)
	if !g.Exile.Contains(id) || me.Graveyard.Contains(id) {
		t.Fatal("a spell given rebound and cast from hand was not exiled as it resolved")
	}
	if n := len(reboundTriggers(g)); n != 1 {
		t.Fatalf("rebound delayed triggers: got %d, want 1", n)
	}
}

// The grant names the OBJECT (CR 400.7): a spell given rebound that is
// countered goes to the graveyard, the record ends with it, and the
// card there does not have rebound.
func TestGrantedReboundEndsWithTheSpell(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	id := handSpell(g, me, "Sorcery")
	castSpellFromHand(t, g, me, id, CastSpellParams{})
	grantRebound(t, g, id)
	if err := g.CounterSpell(id, nil); err != nil {
		t.Fatalf("CounterSpell: %v", err)
	}
	if !me.Graveyard.Contains(id) {
		t.Fatal("a countered spell given rebound did not reach the graveyard")
	}
	g.WithWriteLock(func() {
		// The record is swept by the next pass, whatever prompts it.
		g.layerVersion.Add(1)
		g.RecomputeLayersIfStaleLocked()
		if n := len(g.ScopedEffects); n != 0 {
			t.Errorf("the grant outlived the spell: %d records left", n)
		}
		for i := range me.Graveyard.Cards {
			if c := &me.Graveyard.Cards[i]; c.InstanceID == id && HasKeyword(c, KeywordRebound) {
				t.Error("the card in the graveyard still has the rebound its spell was given")
			}
		}
	})
}

// A spell that has already left the stack is given nothing.
func TestGrantingReboundToASpellThatIsGoneDoesNothing(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	id := handSpell(g, me, "Instant")
	castSpellFromHand(t, g, me, id, CastSpellParams{})
	resolveTop(t, g)
	if grantRebound(t, g, id) {
		t.Error("a spell that resolved was given rebound")
	}
	if n := len(g.ScopedEffects); n != 0 {
		t.Errorf("registered %d records for a spell that is gone", n)
	}
}

// CR 613.1f, 611.3a: "Instant and sorcery spells you control have
// rebound" reaches every such spell its controller casts while the
// static is out — including one cast after it entered — and nobody
// else's.
func TestAStaticOverSpellsGivesYourSpellsRebound(t *testing.T) {
	withReboundForYourSpells(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceTo(t, g, StepPrecombatMain)
	pushSpellStaticSource(g, me)
	// A clean cache first, so the cast itself has to invalidate it.
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })

	mine := handSpell(g, me, "Sorcery")
	castSpellFromHand(t, g, me, mine, CastSpellParams{})
	theirs := handSpell(g, opp, "Instant")
	g.WithWriteLock(func() { g.Turn.PriorityHolder = 1 })
	castSpellFromHand(t, g, opp, theirs, CastSpellParams{})

	if stackSpellHasRebound(t, g, theirs) {
		t.Error("an opponent's spell has the rebound \"spells you control\" gives")
	}
	if !stackSpellHasRebound(t, g, mine) {
		t.Fatal("your sorcery does not have rebound")
	}
	resolveTop(t, g) // theirs
	if !opp.Graveyard.Contains(theirs) {
		t.Error("the opponent's instant did not go to the graveyard")
	}
	resolveTop(t, g) // mine
	if !g.Exile.Contains(mine) || len(reboundTriggers(g)) != 1 {
		t.Error("your sorcery did not rebound under the static")
	}
}

// The static's predicate is asked of spells only. The battlefield pass
// never gathers it, even with a predicate that would accept anything.
func TestAStaticOverSpellsDoesNotReachPermanents(t *testing.T) {
	withStaticAbilities(t, func(key string) []StaticAbility {
		if key != testSpellStaticOracle {
			return nil
		}
		return []StaticAbility{{
			Layer:         Layer6Ability,
			AffectsSpells: true,
			AppliesTo:     func(*Card, *Game, *Card) bool { return true },
			Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
				c.Abilities = AppendKeywordAbility(c.Abilities, "flying")
			},
		}}
	})
	g := newActiveGame(t)
	me := g.Seats[0]
	pushSpellStaticSource(g, me)
	bear := pushScopedTestCreature(g, me.ID, 2, 2)
	if c := layeredBattlefieldCard(t, g, bear); HasKeyword(&c, "flying") {
		t.Error("a static over spells gave a permanent flying")
	}
}

// CR 611.3a: a static's effect is never locked in. The source leaving
// before the spell resolves takes the rebound away.
func TestASpellLosesTheStaticsReboundWhenTheSourceLeaves(t *testing.T) {
	withReboundForYourSpells(t)
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	src := pushSpellStaticSource(g, me)
	id := handSpell(g, me, "Instant")
	castSpellFromHand(t, g, me, id, CastSpellParams{})
	if !stackSpellHasRebound(t, g, id) {
		t.Fatal("setup: the instant does not have rebound")
	}
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(src); err != nil {
			t.Fatalf("exile the source: %v", err)
		}
	})
	resolveTop(t, g)
	if !me.Graveyard.Contains(id) || g.Exile.Contains(id) {
		t.Error("the instant rebounded after the static's source left")
	}
}

// CR 616.1: rebound and buyback replace the same event, and the spell's
// controller chooses. Each answer applies that one replacement only
// (CR 616.1f).
func TestReboundAndBuybackAskTheController(t *testing.T) {
	for _, tc := range []struct {
		name   string
		index  int
		inHand bool
	}{
		{"rebound", 0, false},
		{"buyback", 1, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newRestorableGame(t)
			stubOptionalCosts(t, testBuybackOracle, []AdditionalCost{
				{Optional: true, Key: BuybackKey, ManaCost: "{0}", Label: "Buyback {0}"},
			})
			me := g.Seats[g.Turn.ActiveSeat]
			id, err := castOptional(t, g, "Bought Back", "Instant", testBuybackOracle, []int{0})
			if err != nil {
				t.Fatalf("cast with buyback: %v", err)
			}
			grantRebound(t, g, id)
			resolveTop(t, g)

			choice := resolutionChoiceOutstanding(g)
			if choice == nil {
				t.Fatal("rebound and buyback both applied and nobody was asked")
			}
			if choice.Chooser != me.ID || len(choice.PickOptions) != 2 {
				t.Fatalf("choice = chooser %v, %d options; want the controller, 2", choice.Chooser, len(choice.PickOptions))
			}
			if !g.Stack.Contains(id) {
				t.Fatal("the card left the stack before the choice was made")
			}
			if err := g.ResolveOptionPick(choice.ID, me.ID, tc.index); err != nil {
				t.Fatalf("ResolveOptionPick: %v", err)
			}
			if got := me.Hand.Contains(id); got != tc.inHand {
				t.Errorf("in hand = %v, want %v", got, tc.inHand)
			}
			if got := g.Exile.Contains(id); got == tc.inHand {
				t.Errorf("in exile = %v, want %v", got, !tc.inHand)
			}
			wantTriggers := 1
			if tc.inHand {
				wantTriggers = 0
			}
			if n := len(reboundTriggers(g)); n != wantTriggers {
				t.Errorf("rebound triggers = %d, want %d", n, wantTriggers)
			}
			if me.Graveyard.Contains(id) {
				t.Error("the card reached the graveyard as well")
			}
		})
	}
}

// CR 616.1: rebound and an Adventure's exile (CR 715.3d) replace the
// same event too.
func TestReboundAndTheAdventureExileAskTheController(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	id := handWithAdventure(t, g, me)
	castSpellFromHand(t, g, me, id, CastSpellParams{Face: adventureSpellFace})
	grantRebound(t, g, id)
	resolveTop(t, g)

	choice := resolutionChoiceOutstanding(g)
	if choice == nil {
		t.Fatal("rebound and the Adventure exile both applied and nobody was asked")
	}
	// The Adventure: exiled with CR 715.4's permission, no rebound.
	if err := g.ResolveOptionPick(choice.ID, me.ID, 1); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	if !g.Exile.Contains(id) {
		t.Fatal("the Adventure answer did not exile the card")
	}
	if perm := adventureGrantFor(g, id); !perm.Granted() {
		t.Error("the Adventure answer left no CR 715.4 permission")
	}
	if n := len(reboundTriggers(g)); n != 0 {
		t.Errorf("the Adventure answer also scheduled %d rebound triggers", n)
	}
}

// A spell given rebound is a restore point: the grant is a ScopedEffect
// record, and the restored game rebounds it.
func TestGrantedReboundSurvivesASnapshotRoundTrip(t *testing.T) {
	g := newRestorableGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceTo(t, g, StepPrecombatMain)
	id := handSpell(g, me, "Instant")
	castSpellFromHand(t, g, me, id, CastSpellParams{})
	grantRebound(t, g, id)

	_, restored := roundTrip(t, g)
	if !stackSpellHasRebound(t, restored, id) {
		t.Fatal("the restored spell lost the rebound it was given")
	}
	resolveTop(t, restored)
	if !restored.Exile.Contains(id) || len(reboundTriggers(restored)) != 1 {
		t.Error("the restored spell did not rebound")
	}
}

// ADR 0107 §3 snapshot impact: a stack pin carries only what the stack
// step applies. A file whose stack pin carries anything else is from a
// newer binary, and is refused rather than restored as a record that
// silently does nothing.
func TestAStackPinWithAnUnappliedModIsRefusedOnRestore(t *testing.T) {
	g := newRestorableGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceTo(t, g, StepPrecombatMain)
	id := handSpell(g, me, "Instant")
	castSpellFromHand(t, g, me, id, CastSpellParams{})
	grantRebound(t, g, id)
	snap := g.CaptureSnapshot()
	if _, err := snap.Restore(); err != nil {
		t.Fatalf("a keyword grant on a stack pin was refused: %v", err)
	}
	snap.ScopedEffects[0].Mods[0] = AddTypesMod("Creature")
	if _, err := snap.Restore(); !errors.Is(err, ErrUnknownEffectKey) {
		t.Errorf("Restore: err = %v, want ErrUnknownEffectKey", err)
	}
}

// CR 400.7a: an effect that changed a permanent spell's characteristics
// continues to apply to the permanent that spell becomes.
func TestAKeywordGivenToAPermanentSpellCarriesOntoThePermanent(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	item := castSpellForControlTest(t, g, me, "Creature — Bear")
	g.WithWriteLock(func() {
		if !g.GrantKeywordsToSpellForEffect(uuid.New(), item.ID, []string{"flying"}, "test — that spell gains flying") {
			t.Fatal("GrantKeywordsToSpellForEffect refused a creature spell")
		}
	})
	resolveTop(t, g)
	if c := layeredBattlefieldCard(t, g, item.ID); !HasKeyword(&c, "flying") {
		t.Error("the permanent did not keep the flying its spell was given")
	}
}
