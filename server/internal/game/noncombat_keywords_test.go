package game

import (
	"testing"

	"github.com/google/uuid"
)

// noncombat_keywords_test.go is the #711 regression suite: CR 702.15b
// lifelink and CR 702.2b deathtouch on damage that is not combat
// damage.
//
// Neither rule mentions combat. "Damage dealt by a source with
// lifelink also causes that source's controller to gain that much
// life"; "any nonzero amount of damage a source with deathtouch deals
// to a creature is considered to be lethal damage". A fight, a
// pinger's ability and "target creature you control deals damage
// equal to its power to any target" all carry the source's keywords,
// and until #711 the two *ForEffect entry points carried neither —
// while three card files (Bite Down, Soul's Fire, Chandra's Ignition)
// already documented that they did.
//
// The paused variants reuse damage_tail_test.go's harness: the same
// scenario is played twice, once through a CR 616 ordering prompt and
// once with no prompt at all, and the whole observable outcome has to
// match. A keyword that only survives the unpaused path would be the
// #694 bug wearing a new hat.

// changeLifeRecorder captures the positive EventChangeLife the engine
// emits, which on a damage path is exactly the lifelink credits —
// the life LOSS from damage to a player is folded into the
// EventDealDamage rather than emitted separately.
type changeLifeRecorder struct{ gains []Event }

func (r *changeLifeRecorder) OnEvent(_ *Game, ev Event) {
	if ev.Kind == EventChangeLife && ev.Amount > 0 {
		r.gains = append(r.gains, ev)
	}
}

// TestNonCombatDamageToAPlayerAppliesLifelink — the pinger case.
// A lifelinking creature's activated ability aimed at a player pays
// its controller, exactly as a swing does.
func TestNonCombatDamageToAPlayerAppliesLifelink(t *testing.T) {
	srcID := uuid.New()
	g := assertPausedMatchesUnpaused(t, damageScenario{
		setUp: func(g *Game) {
			pushTestCreature(g, srcID, g.Seats[0], 3, 3, "lifelink")
		},
		deal: func(g *Game) {
			_ = g.DealDamageToPlayerForEffect(srcID, g.Seats[1].ID, 3)
		},
	})
	if got := g.Seats[1].Life; got != StartingLife-4 {
		t.Errorf("target life = %d, want %d", got, StartingLife-4)
	}
	if got := g.Seats[0].Life; got != StartingLife+4 {
		t.Errorf("source controller life = %d, want %d — CR 702.15b is not about combat (#711)",
			got, StartingLife+4)
	}
}

// TestNonCombatDamageToACreatureAppliesLifelinkAndDeathtouch — the
// fight case. A Wurmcoil-shaped source damaging a creature gains its
// controller the damage dealt and makes that damage lethal however
// small it is, so the 1/10 dies to the state-based check.
func TestNonCombatDamageToACreatureAppliesLifelinkAndDeathtouch(t *testing.T) {
	srcID, victimID := uuid.New(), uuid.New()
	g := assertPausedMatchesUnpaused(t, damageScenario{
		setUp: func(g *Game) {
			pushTestCreature(g, srcID, g.Seats[0], 2, 2, "deathtouch", "lifelink")
			pushTestCreature(g, victimID, g.Seats[1], 1, 10)
		},
		deal: func(g *Game) {
			// 2 damage in, (2-1)*2 = 2 out — well short of the
			// victim's 10 toughness, so only deathtouch can kill it.
			_ = g.DealDamageToCreatureForEffect(srcID, victimID, 2)
		},
	})
	if findBattlefieldCard(g, victimID) != nil {
		t.Error("the 1/10 survived 2 non-combat damage from a deathtouch source — " +
			"CR 702.2b makes any nonzero damage from it lethal (#711)")
	}
	if got := g.Seats[1].Graveyard.Size(); got != 1 {
		t.Errorf("victim's graveyard holds %d cards, want 1", got)
	}
	if got := g.Seats[0].Life; got != StartingLife+2 {
		t.Errorf("source controller life = %d, want %d — lifelink on damage to a "+
			"creature too (#711)", got, StartingLife+2)
	}
}

// TestNonCombatLifelinkGainsLifeExactlyOnce — the gain is one event of
// the damage dealt, not one per replacement in the chain and not one
// per SBA pass. Cards that watch lifegain (Ajani's Pridemate, Dawn of
// Hope) count events, so a doubled credit would be visible twice over.
func TestNonCombatLifelinkGainsLifeExactlyOnce(t *testing.T) {
	g := newActiveGame(t)
	srcID := uuid.New()
	seen := &changeLifeRecorder{}
	g.WithWriteLock(func() {
		pushTestCreature(g, srcID, g.Seats[0], 3, 3, "lifelink")
		g.Listeners = append(g.Listeners, seen)
		_ = g.DealDamageToPlayerForEffect(srcID, g.Seats[1].ID, 3)
	})
	if len(seen.gains) != 1 {
		t.Fatalf("%d life-gain events, want exactly 1", len(seen.gains))
	}
	if got := seen.gains[0].Amount; got != 3 {
		t.Errorf("life gained = %d, want 3 — the damage actually dealt", got)
	}
	if got := seen.gains[0].Source; got != srcID {
		t.Errorf("gain event source = %s, want the damage source %s", got, srcID)
	}
	if got := g.Seats[0].Life; got != StartingLife+3 {
		t.Errorf("controller life = %d, want %d", got, StartingLife+3)
	}
}

// TestNonCombatDeathtouchMakesOneDamageLethal — the literal rule, with
// no replacement effects in the way: one damage from a deathtouch
// source kills a creature of any toughness (CR 702.2b feeding the
// CR 704.5h state-based action).
func TestNonCombatDeathtouchMakesOneDamageLethal(t *testing.T) {
	g := newActiveGame(t)
	srcID, victimID := uuid.New(), uuid.New()
	g.WithWriteLock(func() {
		pushTestCreature(g, srcID, g.Seats[0], 1, 1, "deathtouch")
		pushTestCreature(g, victimID, g.Seats[1], 6, 6)
		if err := g.DealDamageToCreatureForEffect(srcID, victimID, 1); err != nil {
			t.Fatalf("DealDamageToCreatureForEffect: %v", err)
		}
	})
	g.WithWriteLock(func() {
		if c := findBattlefieldCard(g, victimID); c == nil {
			t.Fatal("the victim left the battlefield before the state-based check")
		} else if !c.MarkedLethalByDeathtouch {
			t.Error("1 non-combat damage from a deathtouch source did not mark the " +
				"creature lethal (CR 702.2b, #711)")
		}
		g.runStateChecksLocked()
	})
	if findBattlefieldCard(g, victimID) != nil {
		t.Error("the 6/6 survived 1 deathtouch damage — CR 704.5h destroys a creature " +
			"marked lethal")
	}
}

// TestNonCombatDamageFromANonPermanentSourceAppliesNeither — the
// guard rail. The keywords are read off the SOURCE, and a spell, an
// emblem or an absent source is not a permanent with characteristics
// to read. Lightning Bolt still just deals 3.
//
// The stack case is the sharp one: the card is real, it is in a zone,
// and it carries both keywords on its printed Keywords slice — which
// HasKeyword would happily return true for. The snapshot only ever
// looks at the battlefield, so it finds nothing, which is the same
// answer the rules give.
func TestNonCombatDamageFromANonPermanentSourceAppliesNeither(t *testing.T) {
	for _, tc := range []struct {
		name  string
		build func(g *Game) uuid.UUID
	}{
		{
			name:  "no source at all",
			build: func(*Game) uuid.UUID { return uuid.Nil },
		},
		{
			name: "a spell on the stack with both keywords",
			build: func(g *Game) uuid.UUID {
				c := NewCard("Test Spell", g.Seats[0].ID)
				c.TypeLine = "Instant"
				c.Controller = g.Seats[0].ID
				c.Keywords = []string{"deathtouch", "lifelink"}
				g.Stack.PushTop(c)
				return c.InstanceID
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			victimID := uuid.New()
			g.WithWriteLock(func() {
				pushTestCreature(g, victimID, g.Seats[1], 4, 4)
				src := tc.build(g)
				if err := g.DealDamageToCreatureForEffect(src, victimID, 1); err != nil {
					t.Fatalf("DealDamageToCreatureForEffect: %v", err)
				}
				if err := g.DealDamageToPlayerForEffect(src, g.Seats[1].ID, 1); err != nil {
					t.Fatalf("DealDamageToPlayerForEffect: %v", err)
				}
				g.runStateChecksLocked()
			})
			victim := findBattlefieldCard(g, victimID)
			if victim == nil {
				t.Fatal("the 4/4 died to 1 damage from a sourceless effect")
			}
			if victim.MarkedLethalByDeathtouch {
				t.Error("a non-permanent source marked the creature lethal — it has no " +
					"deathtouch to read")
			}
			if got := g.Seats[0].Life; got != StartingLife {
				t.Errorf("caster life = %d, want %d — a non-permanent source has no "+
					"lifelink to credit", got, StartingLife)
			}
			if got := g.Seats[1].Life; got != StartingLife-1 {
				t.Errorf("target life = %d, want %d — the damage itself still lands",
					got, StartingLife-1)
			}
		})
	}
}
