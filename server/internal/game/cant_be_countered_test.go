package game

import (
	"testing"

	"github.com/google/uuid"
)

// cant_be_countered_test.go — ADR 0106 §4 (#1806), delivery PR 2: the
// counter gate's battlefield statics, "Spells you control can't be
// countered" and its "you cast" and "any player" forms. Stubbed
// catalog hooks, so this pins the engine contract; the printed cards
// (Chimil, Gaea's Herald, Thryx …) are pinned end to end in
// cards/effects/counter_shields_test.go.

const shieldOracle = "test-counter-shield"

// withCounterShields stubs the catalog so a battlefield permanent with
// shieldOracle has exactly these shields.
func withCounterShields(t *testing.T, shields ...CounterShieldStatic) {
	t.Helper()
	prev := CatalogCounterShields
	CatalogCounterShields = func(key string) []CounterShieldStatic {
		if key == shieldOracle {
			return shields
		}
		return nil
	}
	t.Cleanup(func() { CatalogCounterShields = prev })
}

// pushShieldSource puts the permanent with the static onto the
// battlefield under `controller`.
func pushShieldSource(g *Game, controller *Player) uuid.UUID {
	c := NewCard("Shield", controller.ID)
	c.TypeLine = "Legendary Artifact"
	c.OracleID = shieldOracle
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// pushSpellFor puts a spell on the stack: cast by `caster`, controlled
// now by `controller`.
func pushSpellFor(t *testing.T, g *Game, typeLine string, caster, controller *Player) uuid.UUID {
	t.Helper()
	id := uuid.New()
	pushStackSpell(t, g, Card{InstanceID: id, Name: "Spell", TypeLine: typeLine,
		Owner: caster.ID, Controller: controller.ID})
	if caster.ID != controller.ID {
		g.StackMeta[id].BaseController = caster.ID
	}
	return id
}

func cantBeCountered(g *Game, id uuid.UUID) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.spellCantBeCounteredLocked(id)
}

// Every counter verb asks the gate, so a battlefield static stops each
// of them; without the static each one counters.
func TestEveryCounterVerbHonoursABattlefieldShield(t *testing.T) {
	verbs := []struct {
		name    string
		counter func(g *Game, id uuid.UUID) error
	}{
		{"counter", func(g *Game, id uuid.UUID) error { return g.CounterTargetForEffect(id) }},
		{"counter to exile", func(g *Game, id uuid.UUID) error {
			return g.CounterTargetToZoneForEffect(id, ZoneRef{Kind: ZoneExile})
		}},
		{"counter to the bottom of the library", func(g *Game, id uuid.UUID) error {
			return g.CounterSpellToLibraryKeyedForEffect(id, TuckOptions{ToBottom: true}, KeyedTail{})
		}},
	}
	for _, v := range verbs {
		for _, shielded := range []bool{true, false} {
			name := v.name + "/unshielded"
			if shielded {
				name = v.name + "/shielded"
			}
			t.Run(name, func(t *testing.T) {
				withCounterShields(t, CounterShieldStatic{Label: "Spells you control can't be countered."})
				g := newActiveGame(t)
				me := g.Seats[0]
				if shielded {
					pushShieldSource(g, me)
				}
				spell := pushSpellFor(t, g, "Instant", me, me)
				g.mu.Lock()
				counterable := g.counterableSpellOnStackLocked(spell)
				err := v.counter(g, spell)
				g.mu.Unlock()
				if err != nil {
					t.Fatalf("counter: %v", err)
				}
				if counterable == shielded {
					t.Errorf("the put_in_library prompt's check says counterable=%v", counterable)
				}
				if g.Stack.Contains(spell) != shielded {
					t.Errorf("on the stack after the counter: %v, want %v", g.Stack.Contains(spell), shielded)
				}
				if hasEventFor(g, EventCounterSpell, spell) == shielded {
					t.Errorf("EventCounterSpell fired: %v", !shielded)
				}
			})
		}
	}
}

// The view's chip reads the same gate.
func TestTheStackChipFollowsABattlefieldShield(t *testing.T) {
	withCounterShields(t, CounterShieldStatic{Label: "Spells you control can't be countered."})
	g := newActiveGame(t)
	me := g.Seats[0]
	spell := pushSpellFor(t, g, "Sorcery", me, me)
	if g.SpellCantBeCounteredForEffect(spell) {
		t.Fatal("no shield yet")
	}
	pushShieldSource(g, me)
	if !g.SpellCantBeCounteredForEffect(spell) {
		t.Error("the chip must show while the shield is out")
	}
}

// The three forms, against the four relations a spell can have to the
// static's controller: cast and controlled, cast but stolen (ADR
// 0104), stolen from someone else, and someone else's altogether.
func TestCounterShieldWhoseSpells(t *testing.T) {
	type rel struct {
		name               string
		caster, controller int // seat index; 0 controls the shield
	}
	mine := rel{"cast and controlled by the static's controller", 0, 0}
	castStolen := rel{"cast by the static's controller, now another player's", 0, 1}
	stolenIn := rel{"cast by another player, now the static's controller's", 1, 0}
	theirs := rel{"another player's", 1, 1}
	for _, tc := range []struct {
		whose CounterShieldWhose
		want  map[rel]bool
	}{
		{CounterShieldYouControl, map[rel]bool{mine: true, castStolen: false, stolenIn: true, theirs: false}},
		{CounterShieldYouCast, map[rel]bool{mine: true, castStolen: true, stolenIn: false, theirs: false}},
		{CounterShieldAnyPlayer, map[rel]bool{mine: true, castStolen: true, stolenIn: true, theirs: true}},
	} {
		for r, want := range tc.want {
			t.Run(r.name, func(t *testing.T) {
				withCounterShields(t, CounterShieldStatic{Whose: tc.whose})
				g := newActiveGame(t)
				pushShieldSource(g, g.Seats[0])
				spell := pushSpellFor(t, g, "Instant", g.Seats[r.caster], g.Seats[r.controller])
				if got := cantBeCountered(g, spell); got != want {
					t.Errorf("whose=%d: can't be countered = %v, want %v", tc.whose, got, want)
				}
			})
		}
	}
}

// A copy is not cast (CR 707.10): "spells you cast" never covers one,
// while "spells you control" covers a copy you control.
func TestCounterShieldOnACopy(t *testing.T) {
	for _, tc := range []struct {
		whose CounterShieldWhose
		want  bool
	}{
		{CounterShieldYouCast, false},
		{CounterShieldYouControl, true},
		{CounterShieldAnyPlayer, true},
	} {
		withCounterShields(t, CounterShieldStatic{Whose: tc.whose})
		g := newActiveGame(t)
		me := g.Seats[0]
		pushShieldSource(g, me)
		copyID := pushSpellFor(t, g, "Instant", me, me)
		g.StackMeta[copyID].IsCopy = true
		if got := cantBeCountered(g, copyID); got != tc.want {
			t.Errorf("whose=%d: a copy can't be countered = %v, want %v", tc.whose, got, tc.want)
		}
	}
}

// The spell filter narrows the static to the spells it names.
func TestCounterShieldSpellFilter(t *testing.T) {
	withCounterShields(t, CounterShieldStatic{
		Label: "Creature spells you control can't be countered.",
		Spell: func(_ *Game, spell Card, _ *Card) bool { return spell.IsCreature() },
	})
	g := newActiveGame(t)
	me := g.Seats[0]
	pushShieldSource(g, me)
	creature := pushSpellFor(t, g, "Creature — Bear", me, me)
	instant := pushSpellFor(t, g, "Instant", me, me)
	if !cantBeCountered(g, creature) {
		t.Error("a creature spell is covered")
	}
	if cantBeCountered(g, instant) {
		t.Error("an instant is not")
	}
}

// CR 611.3a / 604.2: the static applies while its source is on the
// battlefield with the ability. Removing the source, phasing it out
// (CR 702.26b) or taking its abilities away (CR 613.1f) each makes the
// same spell counterable again.
func TestCounterShieldEndsWithItsSource(t *testing.T) {
	for _, tc := range []struct {
		name string
		gone func(g *Game, src uuid.UUID)
	}{
		{"it leaves the battlefield", func(g *Game, src uuid.UUID) {
			if _, err := g.Battlefield.Remove(src); err != nil {
				t.Fatal(err)
			}
		}},
		{"it phases out", func(g *Game, src uuid.UUID) {
			g.mu.Lock()
			defer g.mu.Unlock()
			_ = g.PhaseOutForEffect(uuid.Nil, src)
		}},
		{"it loses all its abilities", func(g *Game, src uuid.UUID) {
			for i := range g.Battlefield.Cards {
				c := &g.Battlefield.Cards[i]
				if c.InstanceID == src {
					eff := c.printedCharacteristic()
					eff.AbilitiesRemoved = true
					c.effective = &eff
				}
			}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			withCounterShields(t, CounterShieldStatic{Label: "Spells you control can't be countered."})
			g := newActiveGame(t)
			me := g.Seats[0]
			src := pushShieldSource(g, me)
			spell := pushSpellFor(t, g, "Instant", me, me)
			if !cantBeCountered(g, spell) {
				t.Fatal("setup: the shield covers the spell")
			}
			tc.gone(g, src)
			if cantBeCountered(g, spell) {
				t.Error("the shield outlived its source")
			}
		})
	}
}

// The static follows its source's controller: a Chimil that changes
// hands shields its new controller's spells.
func TestCounterShieldFollowsItsSourcesController(t *testing.T) {
	withCounterShields(t, CounterShieldStatic{})
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushShieldSource(g, me)
	mySpell := pushSpellFor(t, g, "Instant", me, me)
	theirSpell := pushSpellFor(t, g, "Instant", opp, opp)
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == src {
			g.Battlefield.Cards[i].Controller = opp.ID
		}
	}
	if cantBeCountered(g, mySpell) || !cantBeCountered(g, theirSpell) {
		t.Error("the shield covers its CURRENT controller's spells")
	}
}

// A spell that can't be countered for another reason is untouched by
// the shield coming or going: the printed rider and the mana rider
// still answer on their own.
func TestOtherSourcesAreUnaffectedByAShield(t *testing.T) {
	withCounterShields(t, CounterShieldStatic{})
	old := CatalogCantBeCountered
	CatalogCantBeCountered = func(oracle string) bool { return oracle == "test-printed-rider-1806" }
	t.Cleanup(func() { CatalogCantBeCountered = old })

	g := newActiveGame(t)
	me := g.Seats[0]
	printed := uuid.New()
	pushStackSpell(t, g, Card{InstanceID: printed, Name: "Supreme Verdict", TypeLine: "Sorcery",
		OracleID: "test-printed-rider-1806", Owner: me.ID, Controller: me.ID})
	mana := pushSpellFor(t, g, "Creature — Elf", me, me)
	g.StackMeta[mana].Paid = PaidCost{Mana: []ManaToken{{Color: "G",
		Riders: []ManaSpendRider{{Kind: ManaRiderCantBeCountered, Applied: true}}}}}

	src := pushShieldSource(g, me)
	for _, id := range []uuid.UUID{printed, mana} {
		if !cantBeCountered(g, id) {
			t.Fatal("setup: covered with the shield out")
		}
	}
	if _, err := g.Battlefield.Remove(src); err != nil {
		t.Fatal(err)
	}
	if !cantBeCountered(g, printed) {
		t.Error("the printed rider stopped answering when the shield left")
	}
	if !cantBeCountered(g, mana) {
		t.Error("the mana rider stopped answering when the shield left")
	}
}

// "Spells … can't be countered" is about spells. An ability is still
// countered (CR 701.6a) unless a static says abilities too (#2014,
// below).
func TestCounterShieldDoesNotProtectAbilities(t *testing.T) {
	withCounterShields(t, CounterShieldStatic{Whose: CounterShieldAnyPlayer})
	g := newActiveGame(t)
	me := g.Seats[0]
	src := pushShieldSource(g, me)
	ability := uuid.New()
	if g.StackMeta == nil {
		g.StackMeta = map[uuid.UUID]*StackItem{}
	}
	g.StackMeta[ability] = &StackItem{ID: ability, Kind: StackItemActivated,
		Controller: me.ID, Owner: me.ID, SourceCardID: src}
	if g.SpellCantBeCounteredForEffect(ability) {
		t.Error("the chip must not show on an ability")
	}
	g.mu.Lock()
	err := g.CounterTargetForEffect(ability)
	g.mu.Unlock()
	if err != nil {
		t.Fatalf("counter the ability: %v", err)
	}
	if _, ok := g.StackMeta[ability]; ok {
		t.Error("the ability was not countered")
	}
}

// --- #2014: "Spells and abilities can't be countered" ---------------

// pushAbilityFor puts an activated or triggered ability on the stack,
// controlled by `controller`, from the permanent `src`.
func pushAbilityFor(g *Game, kind StackItemKind, controller *Player, src uuid.UUID) uuid.UUID {
	id := uuid.New()
	if g.StackMeta == nil {
		g.StackMeta = map[uuid.UUID]*StackItem{}
	}
	g.StackMeta[id] = &StackItem{ID: id, Kind: kind, Controller: controller.ID,
		Owner: controller.ID, SourceCardID: src, Label: "Test ability"}
	return id
}

func abilityCantBeCountered(g *Game, id uuid.UUID) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.abilityCantBeCounteredLocked(g.StackMeta[id])
}

// A static that covers abilities stops every counter verb from
// countering an activated or a triggered ability (CR 701.6a, CR
// 101.2); the verb returns cleanly, as it does for a covered spell, so
// the Stifle that asked resolves and does nothing. Without the static
// each verb counters.
func TestAnAbilityShieldStopsEveryCounterVerb(t *testing.T) {
	verbs := []struct {
		name    string
		counter func(g *Game, id uuid.UUID) error
	}{
		{"counter", func(g *Game, id uuid.UUID) error { return g.CounterTargetForEffect(id) }},
		{"counter to exile", func(g *Game, id uuid.UUID) error {
			return g.CounterTargetToZoneForEffect(id, ZoneRef{Kind: ZoneExile})
		}},
		{"counter to the bottom of the library", func(g *Game, id uuid.UUID) error {
			return g.CounterSpellToLibraryKeyedForEffect(id, TuckOptions{ToBottom: true}, KeyedTail{})
		}},
		{"counter unless paid, declined", func(g *Game, id uuid.UUID) error { return g.counterIfStillOnStackLocked(id) }},
	}
	for _, v := range verbs {
		for _, kind := range []StackItemKind{StackItemActivated, StackItemTriggered} {
			for _, shielded := range []bool{true, false} {
				name := v.name + "/" + string(kind) + "/unshielded"
				if shielded {
					name = v.name + "/" + string(kind) + "/shielded"
				}
				t.Run(name, func(t *testing.T) {
					withCounterShields(t, CounterShieldStatic{Label: "Spells and abilities can't be countered.",
						Whose: CounterShieldAnyPlayer, Abilities: true})
					g := newActiveGame(t)
					me, opp := g.Seats[0], g.Seats[1]
					src := uuid.New()
					if shielded {
						src = pushShieldSource(g, me)
					}
					// An opponent's ability: the static says "abilities",
					// with no "you".
					ability := pushAbilityFor(g, kind, opp, src)
					g.mu.Lock()
					err := v.counter(g, ability)
					g.mu.Unlock()
					if err != nil {
						t.Fatalf("counter: %v", err)
					}
					if _, onStack := g.StackMeta[ability]; onStack != shielded {
						t.Errorf("on the stack after the counter: %v, want %v", onStack, shielded)
					}
					countered := false
					for _, ev := range g.Events {
						if ev.Kind == EventCounterSpell && ev.Target == ability {
							countered = true
						}
					}
					if countered == shielded {
						t.Errorf("EventCounterSpell fired: %v, want %v", countered, !shielded)
					}
				})
			}
		}
	}
}

// Whose abilities: "any player" covers every player's; "you control"
// covers the static's controller's only; "you cast" covers none,
// because an ability is never cast.
func TestAnAbilityShieldWhose(t *testing.T) {
	for _, tc := range []struct {
		name         string
		whose        CounterShieldWhose
		mine, theirs bool
	}{
		{"any player", CounterShieldAnyPlayer, true, true},
		{"you control", CounterShieldYouControl, true, false},
		{"you cast", CounterShieldYouCast, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			withCounterShields(t, CounterShieldStatic{Whose: tc.whose, Abilities: true})
			g := newActiveGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			src := pushShieldSource(g, me)
			mine := pushAbilityFor(g, StackItemActivated, me, src)
			theirs := pushAbilityFor(g, StackItemTriggered, opp, src)
			if got := abilityCantBeCountered(g, mine); got != tc.mine {
				t.Errorf("my ability covered: %v, want %v", got, tc.mine)
			}
			if got := abilityCantBeCountered(g, theirs); got != tc.theirs {
				t.Errorf("an opponent's ability covered: %v, want %v", got, tc.theirs)
			}
		})
	}
}

// A static with Abilities still covers spells as its Whose and Spell
// say, and the ability gate answers false for a spell item.
func TestAnAbilityShieldStillCoversSpells(t *testing.T) {
	withCounterShields(t, CounterShieldStatic{Whose: CounterShieldAnyPlayer, Abilities: true})
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushShieldSource(g, me)
	spell := pushSpellFor(t, g, "Instant", opp, opp)
	if !cantBeCountered(g, spell) {
		t.Error("the spell half must cover an opponent's spell")
	}
	if abilityCantBeCountered(g, spell) {
		t.Error("the ability gate answered for a spell")
	}
	if g.AbilityCantBeCounteredForEffect(spell) {
		t.Error("the ability chip must not show on a spell")
	}
}

// The chip on an ability reads the same gate, and the static ends with
// its source (CR 611.3a, CR 604.2): an ability already on the stack can
// be countered once the permanent has left.
func TestAnAbilityShieldEndsWithItsSource(t *testing.T) {
	withCounterShields(t, CounterShieldStatic{Whose: CounterShieldAnyPlayer, Abilities: true})
	g := newActiveGame(t)
	me := g.Seats[0]
	src := pushShieldSource(g, me)
	ability := pushAbilityFor(g, StackItemTriggered, me, src)
	if !g.AbilityCantBeCounteredForEffect(ability) {
		t.Fatal("the chip must show while the shield is out")
	}
	if _, err := g.Battlefield.Remove(src); err != nil {
		t.Fatal(err)
	}
	if g.AbilityCantBeCounteredForEffect(ability) {
		t.Error("the shield outlived its source")
	}
	g.mu.Lock()
	err := g.CounterTargetForEffect(ability)
	g.mu.Unlock()
	if err != nil {
		t.Fatalf("counter: %v", err)
	}
	if _, ok := g.StackMeta[ability]; ok {
		t.Error("the ability was not countered once the shield left")
	}
}
