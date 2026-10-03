package game

import (
	"testing"

	"github.com/google/uuid"
)

// damage_as_though_test.go — ADR 0108 §10 (#1889): damage dealt as though
// its source had wither (Everlasting Torment) or, to you while you have 0
// or less life, infect (Phyrexian Unlife). CR 120.3b, 120.3d, 609.4,
// 702.80a, 702.90b.

const asThoughOracle = "test-damage-as-though"

// withDamageAsThough stubs the catalog so a battlefield permanent with
// asThoughOracle has exactly these statics.
func withDamageAsThough(t *testing.T, statics ...DamageAsThoughStatic) {
	t.Helper()
	prev := CatalogDamageAsThough
	CatalogDamageAsThough = func(key string) []DamageAsThoughStatic {
		if key == asThoughOracle {
			return statics
		}
		return nil
	}
	t.Cleanup(func() { CatalogDamageAsThough = prev })
}

var (
	allWither   = DamageAsThoughStatic{Label: "Everlasting Torment", Wither: true}
	unlifeInfec = DamageAsThoughStatic{Label: "Phyrexian Unlife", Infect: true, ToYou: true, WhileAtOrBelowZeroLife: true}
)

// pushAsThoughEnchantment puts an enchantment carrying the stubbed statics
// onto the battlefield under `owner`.
func pushAsThoughEnchantment(g *Game, owner *Player) uuid.UUID {
	c := NewCard("As Though", owner.ID)
	c.TypeLine = "Enchantment"
	c.OracleID = asThoughOracle
	g.WithWriteLock(func() { g.Battlefield.PushTop(c) })
	return c.InstanceID
}

func setLife(g *Game, p *Player, n int) {
	g.WithWriteLock(func() { p.Life = n })
}

func toughCreature(g *Game, owner *Player, name string) uuid.UUID {
	id := pushColouredCreature(g, owner, name, []string{"R"})
	g.WithWriteLock(func() { findBattlefieldCard(g, id).Toughness = 20 })
	return id
}

// Everlasting Torment: "All damage is dealt as though its source had
// wither." Damage to a creature is -1/-1 counters, not marked damage; the
// source gains no ability (CR 609.4); damage to a player is life loss as
// before (CR 702.80a is about creatures).
func TestAllDamageAsThoughWitherIsCountersOnCreatures(t *testing.T) {
	withDamageAsThough(t, allWither)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushColouredCreature(g, me, "Plain Source", []string{"R"})
	victim := toughCreature(g, opp, "Victim")
	pushAsThoughEnchantment(g, opp)
	start := lifeOf(g, opp.ID)

	dealToCreature(t, g, src, victim, 3)
	dealToPlayer(t, g, src, opp.ID, 2)

	if got := minusOneCountersOn(g, victim); got != 3 {
		t.Errorf("-1/-1 counters on the victim = %d, want 3 (CR 120.3d)", got)
	}
	if got := damageOn(g, victim); got != 0 {
		t.Errorf("damage marked = %d, want 0: wither damage is counters instead", got)
	}
	if got := lifeOf(g, opp.ID); got != start-2 {
		t.Errorf("life %d, want %d: wither changes nothing about damage to a player", got, start-2)
	}
	g.ReadSnapshot(func() {
		if HasKeyword(findBattlefieldCard(g, src), KeywordWither) {
			t.Error("the source gained wither: \"as though\" gives it no ability (CR 609.4)")
		}
	})
	// Counters are placed by the source's controller (CR 120.3d).
	var placed []Event
	for _, ev := range iwtEventsOfKind(g, EventCounterPlaced) {
		if ev.Target == victim {
			placed = append(placed, ev)
		}
	}
	if len(placed) == 0 {
		t.Fatal("no counter event for the -1/-1 counters")
	}
	if placed[0].Actor != me.ID {
		t.Errorf("counters placed by %v, want the source's controller %v", placed[0].Actor, me.ID)
	}
	// The counters are dealt by a spell too: "Wither works everywhere."
	bolt := pushRedSpell(g, me)
	other := toughCreature(g, opp, "Other")
	dealToCreature(t, g, bolt, other, 2)
	if got := minusOneCountersOn(g, other); got != 2 {
		t.Errorf("a spell's 2 damage put %d counters, want 2", got)
	}
}

// The static is read as the damage lands, on the battlefield: once it
// has left, damage is marked again.
func TestDamageAsThoughStopsWhenThePermanentLeaves(t *testing.T) {
	withDamageAsThough(t, allWither)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushColouredCreature(g, me, "Plain Source", []string{"R"})
	victim := toughCreature(g, opp, "Victim")
	ench := pushAsThoughEnchantment(g, opp)
	if err := g.MoveCardByID(ZoneRef{Kind: ZoneBattlefield}, ZoneRef{Kind: ZoneGraveyard, Owner: opp.ID}, ench); err != nil {
		t.Fatal(err)
	}
	dealToCreature(t, g, src, victim, 3)
	if got, marked := minusOneCountersOn(g, victim), damageOn(g, victim); got != 0 || marked != 3 {
		t.Errorf("counters %d, marked %d; want 0 and 3 once the enchantment is gone", got, marked)
	}
}

// Phyrexian Unlife: only damage dealt to its controller, and only while
// that player has 0 or less life.
func TestDamageToYouAsThoughInfectAtOrBelowZeroLife(t *testing.T) {
	withDamageAsThough(t, unlifeInfec)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushColouredCreature(g, opp, "Source", []string{"R"})
	mine := toughCreature(g, me, "My Creature")
	pushAsThoughEnchantment(g, me)

	// Positive life: ordinary life loss.
	setLife(g, me, 5)
	dealToPlayer(t, g, src, me.ID, 2)
	if lifeOf(g, me.ID) != 3 || poisonOf(g, me.ID) != 0 {
		t.Fatalf("at 5 life: life %d poison %d, want 3 and 0", lifeOf(g, me.ID), poisonOf(g, me.ID))
	}
	// 0 life: poison instead.
	setLife(g, me, 0)
	dealToPlayer(t, g, src, me.ID, 3)
	if lifeOf(g, me.ID) != 0 || poisonOf(g, me.ID) != 3 {
		t.Fatalf("at 0 life: life %d poison %d, want 0 and 3", lifeOf(g, me.ID), poisonOf(g, me.ID))
	}
	// Not my creature ("to you"), and not an opponent at 0 life.
	dealToCreature(t, g, src, mine, 2)
	if got := minusOneCountersOn(g, mine); got != 0 {
		t.Errorf("your creature got %d -1/-1 counters, want 0: the static is about you", got)
	}
	setLife(g, opp, 0)
	dealToPlayer(t, g, mine, opp.ID, 2)
	if lifeOf(g, opp.ID) != -2 || poisonOf(g, opp.ID) != 0 {
		t.Errorf("opponent: life %d poison %d, want -2 and 0", lifeOf(g, opp.ID), poisonOf(g, opp.ID))
	}
}

// The ruling: "Phyrexian Unlife won't affect damage that reduces your
// life total from a positive number to 0 or less. For example, if you're
// at 3 life and are dealt 5 damage, you'll end up at -2 life. The next
// time you're dealt damage, it will be dealt as though its source had
// infect."
func TestUnlifeDoesNotAffectDamageThatTakesYouToZero(t *testing.T) {
	withDamageAsThough(t, unlifeInfec)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushColouredCreature(g, opp, "Source", []string{"R"})
	pushAsThoughEnchantment(g, me)
	setLife(g, me, 3)
	dealToPlayer(t, g, src, me.ID, 5)
	if lifeOf(g, me.ID) != -2 || poisonOf(g, me.ID) != 0 {
		t.Fatalf("life %d poison %d, want -2 and 0", lifeOf(g, me.ID), poisonOf(g, me.ID))
	}
	dealToPlayer(t, g, src, me.ID, 2)
	if lifeOf(g, me.ID) != -2 || poisonOf(g, me.ID) != 2 {
		t.Fatalf("next time: life %d poison %d, want -2 and 2", lifeOf(g, me.ID), poisonOf(g, me.ID))
	}
}

// Owner decision 4: the life total is read once per damage instance, as
// it began. Two sources dealing damage at the same time — one printed
// instruction, or one combat damage step (CR 510.2) — that take you from
// 3 to -1 are all life loss, though the first alone took you to 1 and the
// second's half alone would have found you at 0 or less.
func TestUnlifeReadsLifeOncePerInstance(t *testing.T) {
	withDamageAsThough(t, unlifeInfec)
	t.Run("one instruction", func(t *testing.T) {
		g := newActiveGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		a := pushColouredCreature(g, opp, "A", []string{"R"})
		b := pushColouredCreature(g, opp, "B", []string{"R"})
		pushAsThoughEnchantment(g, me)
		setLife(g, me, 3)
		g.WithWriteLock(func() {
			err := g.DamageInstanceForEffect(func() error {
				for _, src := range []uuid.UUID{a, b, a} {
					if err := g.DealDamageToPlayerForEffect(src, me.ID, 2); err != nil {
						return err
					}
				}
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
		})
		if lifeOf(g, me.ID) != -3 || poisonOf(g, me.ID) != 0 {
			t.Fatalf("life %d poison %d, want -3 and 0: one instance began at 3 life", lifeOf(g, me.ID), poisonOf(g, me.ID))
		}
	})
	t.Run("one combat damage step", func(t *testing.T) {
		g := newActiveGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		a := pushColouredCreature(g, opp, "A", []string{"R"})
		b := pushColouredCreature(g, opp, "B", []string{"R"})
		pushAsThoughEnchantment(g, me)
		setLife(g, me, 3)
		g.WithWriteLock(func() {
			g.beginEventBatchLocked()
			g.markCombatDamageToPlayerLocked(me.ID, a, 2, CombatStepRegular)
			g.markCombatDamageToPlayerLocked(me.ID, b, 2, CombatStepRegular)
		})
		if lifeOf(g, me.ID) != -1 || poisonOf(g, me.ID) != 0 {
			t.Fatalf("life %d poison %d, want -1 and 0: the step began at 3 life", lifeOf(g, me.ID), poisonOf(g, me.ID))
		}
		// The next step is a new instance, begun at -1.
		g.WithWriteLock(func() {
			g.beginEventBatchLocked()
			g.markCombatDamageToPlayerLocked(me.ID, a, 2, CombatStepRegular)
		})
		if lifeOf(g, me.ID) != -1 || poisonOf(g, me.ID) != 2 {
			t.Fatalf("next step: life %d poison %d, want -1 and 2", lifeOf(g, me.ID), poisonOf(g, me.ID))
		}
	})
	t.Run("two instructions", func(t *testing.T) {
		// "It deals 2 damage to you. Then it deals 2 damage to you" is two
		// instances (CR 608.2c): the second begins at 0 or less.
		g := newActiveGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		a := pushColouredCreature(g, opp, "A", []string{"R"})
		pushAsThoughEnchantment(g, me)
		setLife(g, me, 2)
		g.WithWriteLock(func() {
			for i := 0; i < 2; i++ {
				if err := g.DealDamageToPlayerForEffect(a, me.ID, 2); err != nil {
					t.Fatal(err)
				}
			}
		})
		if lifeOf(g, me.ID) != 0 || poisonOf(g, me.ID) != 2 {
			t.Fatalf("life %d poison %d, want 0 and 2", lifeOf(g, me.ID), poisonOf(g, me.ID))
		}
	})
}

// Toxic still adds its total on combat damage (CR 120.3g), on top of the
// infect poison.
func TestUnlifeInfectKeepsToxic(t *testing.T) {
	withDamageAsThough(t, unlifeInfec)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	atk := uuid.New()
	g.WithWriteLock(func() { pushPrintedKeywordCreature(g, atk, opp.ID, opp.ID, 2, 2, "toxic 1") })
	pushAsThoughEnchantment(g, me)
	setLife(g, me, 0)
	g.WithWriteLock(func() {
		g.beginEventBatchLocked()
		g.markCombatDamageToPlayerLocked(me.ID, atk, 2, CombatStepRegular)
	})
	if lifeOf(g, me.ID) != 0 || poisonOf(g, me.ID) != 3 {
		t.Fatalf("life %d poison %d, want 0 and 3 (2 infect + toxic 1)", lifeOf(g, me.ID), poisonOf(g, me.ID))
	}
}

// --- interactions -----------------------------------------------------

// Damage dealt as counters or poison is still damage, so a prevention
// shield prevents it, and no counter is put for a prevented point (CR
// 615.1, 120.4b before 120.4c).
func TestDamageAsThoughIsStillPreventable(t *testing.T) {
	withDamageAsThough(t, allWither, unlifeInfec)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushColouredCreature(g, opp, "Source", []string{"R"})
	victim := toughCreature(g, opp, "Victim")
	pushAsThoughEnchantment(g, me)
	setLife(g, me, 0)
	g.WithWriteLock(func() {
		if !g.PreventNextDamageThisTurnForEffect(uuid.Nil, victim, 2, false, "Mending Hands") {
			t.Fatal("no creature shield")
		}
		if !g.PreventNextDamageThisTurnForEffect(uuid.Nil, me.ID, 2, false, "Mending Hands") {
			t.Fatal("no player shield")
		}
	})
	dealToCreature(t, g, src, victim, 3)
	dealToPlayer(t, g, src, me.ID, 3)
	if got := minusOneCountersOn(g, victim); got != 1 {
		t.Errorf("-1/-1 counters %d, want 1: 2 of the 3 prevented", got)
	}
	if got := poisonOf(g, me.ID); got != 1 {
		t.Errorf("poison %d, want 1: 2 of the 3 prevented", got)
	}
}

// Under "damage can't be prevented" (ADR 0107 §5, Everlasting Torment's
// own second line) a shield prevents nothing and keeps its charge (CR
// 615.12), and every point is a counter.
func TestDamageAsThoughUnderDamageCantBePrevented(t *testing.T) {
	withDamageAsThough(t, allWither, unlifeInfec)
	withUnpreventableStatics(t, UnpreventableDamageStatic{Label: "Damage can't be prevented.", Scope: UnpreventableAll})
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushColouredCreature(g, opp, "Source", []string{"R"})
	victim := toughCreature(g, opp, "Victim")
	pushAsThoughEnchantment(g, me)
	pushUnpreventableCreature(g, opp)
	setLife(g, me, 0)
	g.WithWriteLock(func() {
		if !g.PreventNextDamageThisTurnForEffect(uuid.Nil, victim, 2, false, "Mending Hands") {
			t.Fatal("no shield")
		}
		g.PreventCombatDamageThisTurnForEffect(uuid.Nil, uuid.Nil, "Fog")
	})
	dealToCreature(t, g, src, victim, 3)
	g.WithWriteLock(func() {
		g.beginEventBatchLocked()
		g.markCombatDamageToPlayerLocked(me.ID, src, 2, CombatStepRegular)
	})
	if got := minusOneCountersOn(g, victim); got != 3 {
		t.Errorf("-1/-1 counters %d, want 3: nothing is prevented", got)
	}
	if got := poisonOf(g, me.ID); got != 2 {
		t.Errorf("poison %d, want 2: the Fog prevents nothing", got)
	}
	g.ReadSnapshot(func() {
		for _, r := range g.ScopedEffects {
			for _, m := range r.Mods {
				if m.Kind == ModPreventDamage && m.Amount != 2 {
					t.Errorf("the shield's charge is %d, want 2 (CR 615.12)", m.Amount)
				}
			}
		}
	})
}

// A damage multiplier (ADR 0108 §3) changes the damage, and the counters
// follow the damage: twice the damage is twice the counters or poison.
func TestDamageAsThoughFollowsAMultiplier(t *testing.T) {
	withDamageAsThough(t, allWither, unlifeInfec)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushColouredCreature(g, opp, "Source", []string{"R"})
	victim := toughCreature(g, me, "Victim")
	pushAsThoughEnchantment(g, me)
	setLife(g, me, 0)
	multiplier(t, g, DamageMultiplier{Controller: opp.ID, Factor: 2, Sources: DamageSourcesYours, Label: "Insult"})
	dealToCreature(t, g, src, victim, 3)
	dealToPlayer(t, g, src, me.ID, 2)
	if got := minusOneCountersOn(g, victim); got != 6 {
		t.Errorf("-1/-1 counters %d, want 6: 3 doubled", got)
	}
	if got := poisonOf(g, me.ID); got != 4 {
		t.Errorf("poison %d, want 4: 2 doubled", got)
	}
}

// ADR 0108 §7's shields take their part off the damage first: a charged
// source shield and Dark Sphere's half.
func TestDamageAsThoughAfterSourceShields(t *testing.T) {
	withDamageAsThough(t, unlifeInfec)
	t.Run("charged", func(t *testing.T) {
		g := newActiveGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		src := pushColouredCreature(g, opp, "Dragon", []string{"R"})
		pushAsThoughEnchantment(g, me)
		setLife(g, me, 0)
		sourceShield(t, g, me.ID, src, 2, BodyRef{})
		dealToPlayer(t, g, src, me.ID, 5)
		if lifeOf(g, me.ID) != 0 || poisonOf(g, me.ID) != 3 {
			t.Fatalf("life %d poison %d, want 0 and 3: 2 of the 5 prevented", lifeOf(g, me.ID), poisonOf(g, me.ID))
		}
	})
	t.Run("all this turn", func(t *testing.T) {
		g := newActiveGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		src := pushColouredCreature(g, opp, "Dragon", []string{"R"})
		pushAsThoughEnchantment(g, me)
		setLife(g, me, 0)
		sourceShield(t, g, me.ID, src, 0, BodyRef{})
		dealToPlayer(t, g, src, me.ID, 5)
		if lifeOf(g, me.ID) != 0 || poisonOf(g, me.ID) != 0 {
			t.Fatalf("life %d poison %d, want 0 and 0: all of it prevented", lifeOf(g, me.ID), poisonOf(g, me.ID))
		}
	})
	t.Run("half", func(t *testing.T) {
		g := newActiveGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		src := pushColouredCreature(g, opp, "Dragon", []string{"R"})
		pushAsThoughEnchantment(g, me)
		setLife(g, me, 0)
		g.WithWriteLock(func() {
			ref, zone, _ := g.DamageSourceRefLocked(src)
			if !g.PreventNextDamageFromSourceForEffect(NextDamageShield{
				Controller: me.ID, Source: ref, SourceZone: zone, ProtectPlayer: me.ID, Half: true, Label: "Dark Sphere",
			}) {
				t.Fatal("no shield")
			}
		})
		dealToPlayer(t, g, src, me.ID, 5)
		if lifeOf(g, me.ID) != 0 || poisonOf(g, me.ID) != 3 {
			t.Fatalf("life %d poison %d, want 0 and 3: 2 of the 5 prevented, rounded down", lifeOf(g, me.ID), poisonOf(g, me.ID))
		}
	})
}

// An undo rewinds the life totals the instances began with along with
// the counter: the record is cloned with the game.
func TestDamageInstanceLivesAreCloned(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	setLife(g, me, 7)
	var inst DamageInstance
	g.WithWriteLock(func() { inst = g.nextDamageInstanceLocked() })
	c := g.Clone()
	g.WithWriteLock(func() {
		for i := 0; i < maxDamageInstanceLives+2; i++ {
			g.nextDamageInstanceLocked()
		}
	})
	setLife(g, me, -4)
	c.WithWriteLock(func() {
		me := c.Seats[0]
		me.Life = -4
		if c.atOrBelowZeroLifeAsInstanceBeganLocked(inst, me.ID) {
			t.Error("the clone lost the record: the instance began at 7 life")
		}
	})
	g.WithWriteLock(func() {
		if !g.atOrBelowZeroLifeAsInstanceBeganLocked(inst, me.ID) {
			t.Error("an aged-out record should read the live total (-4)")
		}
		if n := len(g.damageInstanceLives); n > maxDamageInstanceLives {
			t.Errorf("%d records kept, want at most %d", n, maxDamageInstanceLives)
		}
	})
}
