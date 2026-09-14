package game

import (
	"testing"

	"github.com/google/uuid"
)

// layer6_ability_removal_test.go is the engine-side suite for
// CR 613.1f: an ability-removing continuous effect is now visible to
// every Catalog* ability lookup, not just to the keyword list.
//
// These tests drive stubbed catalog hooks so they pin the ENGINE
// contract without depending on a particular card, the same way
// layer4_authoritative_test.go does for layer 4. The catalog side —
// Darksteel Mutation, Kenrith's Transformation, Song of the Dryads —
// is pinned in cards/effects/ability_removal_test.go.

const (
	silencerOracle = "static-ability-remover"
	victimOracle   = "static-ability-victim"
	anthemOracle   = "static-anthem"
)

// silencerStatics is "enchanted-style: the named victim loses all
// abilities", expressed as a bare layer-6 removal with no host
// relation so the fixture stays about the layer and not about
// attachments.
func silencerStatics(victim *uuid.UUID, keep ...string) []StaticAbility {
	return []StaticAbility{{
		Layer:            Layer6Ability,
		RemovesAbilities: true,
		AppliesTo: func(target *Card, _ *Game, _ *Card) bool {
			return target.InstanceID == *victim
		},
		Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
			c.Abilities = append(c.Abilities, keep...)
		},
	}}
}

// hasKeywordOnBattlefield forces a recompute and asks the layered
// card, which is the surface a keyword grant lands on.
func hasKeywordOnBattlefield(t *testing.T, g *Game, id uuid.UUID, kw string) bool {
	t.Helper()
	c := layeredBattlefieldCard(t, g, id)
	return HasKeyword(&c, kw)
}

// --- the accessor ------------------------------------------------

// CatalogAbilityKey is the seam. Everything else in this file is a
// consequence of it, so it is worth pinning on its own: it is
// CatalogKey until a removal applies, "" after, and the removal is
// unreachable off the battlefield (CR 113.6).
func TestCatalogAbilityKeyGoesBlankOnlyWhenAbilitiesAreRemoved(t *testing.T) {
	c := Card{OracleID: "oracle-1", Name: "Sol Ring", TypeLine: "Artifact"}

	if got := CatalogAbilityKey(c); got != "oracle-1" {
		t.Errorf("no layer cache: CatalogAbilityKey = %q, want the printed key", got)
	}

	eff := c.printedCharacteristic()
	c.effective = &eff
	if got := CatalogAbilityKey(c); got != "oracle-1" {
		t.Errorf("layered but not silenced: CatalogAbilityKey = %q, want the printed key", got)
	}

	eff.AbilitiesRemoved = true
	if got := CatalogAbilityKey(c); got != "" {
		t.Errorf("silenced: CatalogAbilityKey = %q, want the empty key", got)
	}
	if got := CatalogKey(c); got != "oracle-1" {
		t.Errorf("CatalogKey is the OTHER half and must not move: %q", got)
	}
}

// --- what goes quiet ---------------------------------------------

// The headline: a permanent whose abilities were removed offers no
// activated abilities, which is what stops internal/legal enumerating
// a move the engine would refuse. #544 is the shape of the failure
// when those two disagree — a seat that owes a decision is enumerated
// that decision's answers and nothing else, so an unanswerable move
// list is a hung table, not a wasted turn.
func TestActivatedAbilitiesGoQuietUnderAbilityRemoval(t *testing.T) {
	g := newActiveGame(t)
	seat := g.Seats[0].ID
	var victim uuid.UUID

	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		if oracleID == silencerOracle {
			return silencerStatics(&victim)
		}
		return nil
	})
	prev := CatalogActivatedAbilities
	CatalogActivatedAbilities = func(oracleID string) []ActivatedAbilityShape {
		if oracleID != victimOracle {
			return nil
		}
		return []ActivatedAbilityShape{{Label: "{T}: do a thing"}}
	}
	t.Cleanup(func() { CatalogActivatedAbilities = prev })

	victim = pushTypedTestCard(g, Card{
		Name: "Victim", TypeLine: "Artifact", OracleID: victimOracle,
		Owner: seat, Controller: seat,
	})

	before := layeredBattlefieldCard(t, g, victim)
	if len(ActivatedAbilitiesForCard(before)) != 1 {
		t.Fatalf("fixture is wrong: the card should offer one ability before the silencer lands")
	}

	pushTypedTestCard(g, Card{
		Name: "Silencer", TypeLine: "Enchantment", OracleID: silencerOracle,
		Owner: seat, Controller: seat,
	})

	after := layeredBattlefieldCard(t, g, victim)
	if got := ActivatedAbilitiesForCard(after); len(got) != 0 {
		t.Errorf("ActivatedAbilitiesForCard = %v, want none", got)
	}
}

// A token's intrinsic activated ability is an ability too — a Clue's
// "{2}, Sacrifice this: Draw a card" is not a catalog lookup, it is
// carried on the instance, and the removal has to reach it as well.
func TestIntrinsicActivatedAbilitiesGoQuietToo(t *testing.T) {
	g := newActiveGame(t)
	seat := g.Seats[0].ID
	var victim uuid.UUID

	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		if oracleID == silencerOracle {
			return silencerStatics(&victim)
		}
		return nil
	})

	victim = pushTypedTestCard(g, Card{
		Name: "Clue", TypeLine: "Artifact — Clue", Owner: seat, Controller: seat,
		ActivatedAbilities: []ActivatedAbilityShape{{Label: "{2}, Sacrifice this: Draw a card"}},
	})
	pushTypedTestCard(g, Card{
		Name: "Silencer", TypeLine: "Enchantment", OracleID: silencerOracle,
		Owner: seat, Controller: seat,
	})

	after := layeredBattlefieldCard(t, g, victim)
	if got := ActivatedAbilitiesForCard(after); len(got) != 0 {
		t.Errorf("a token's own ability is still an ability: got %v", got)
	}
}

// Triggered abilities are harvested through CatalogTriggers at event
// time, which is the reason this whole seam exists: the layer engine
// had no way to reach them.
func TestTriggersDoNotFireForASilencedPermanent(t *testing.T) {
	g := newActiveGame(t)
	seat := g.Seats[0].ID
	var victim uuid.UUID

	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		if oracleID == silencerOracle {
			return silencerStatics(&victim)
		}
		return nil
	})
	var fired int
	withCatalogTriggers(t, func(oracleID string) []TriggeredAbility {
		if oracleID != victimOracle {
			return nil
		}
		return []TriggeredAbility{{
			Watches: []EventKind{EventBeginUpkeep},
			AppliesTo: func(Event, *Card, Characteristic, *Game) bool {
				fired++
				return false
			},
		}}
	})

	victim = pushTypedTestCard(g, Card{
		Name: "Victim", TypeLine: "Enchantment", OracleID: victimOracle,
		Owner: seat, Controller: seat,
	})
	g.WithWriteLock(func() { g.EmitEvent(Event{Kind: EventBeginUpkeep, Actor: seat}) })
	if fired != 1 {
		t.Fatalf("fixture is wrong: the trigger should have been offered once, got %d", fired)
	}

	pushTypedTestCard(g, Card{
		Name: "Silencer", TypeLine: "Enchantment", OracleID: silencerOracle,
		Owner: seat, Controller: seat,
	})
	// Force the recompute so the silencing is stamped before the
	// harvest asks.
	_ = layeredBattlefieldCard(t, g, victim)

	fired = 0
	g.WithWriteLock(func() { g.EmitEvent(Event{Kind: EventBeginUpkeep, Actor: seat}) })
	if fired != 0 {
		t.Errorf("a silenced permanent's trigger was offered %d times, want 0", fired)
	}
}

// The source's OWN static abilities stop too, and this is the half
// that needs the second recompute pass: an anthem is layer 7c, the
// removal is layer 6, and the gather that collected the anthem
// happened before either ran.
func TestAbilityRemovalSilencesTheSourcesOwnStatics(t *testing.T) {
	g := newActiveGame(t)
	seat := g.Seats[0].ID
	var anthem uuid.UUID

	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		switch oracleID {
		case silencerOracle:
			return silencerStatics(&anthem)
		case anthemOracle:
			return []StaticAbility{{
				Layer:    Layer7PT,
				SubLayer: SubLayer7C_Modify,
				AppliesTo: func(target *Card, _ *Game, source *Card) bool {
					return target.IsCreature() && target.Controller == source.Controller
				},
				Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
					c.Power++
					c.Toughness++
				},
			}}
		}
		return nil
	})

	bear := pushTypedTestCard(g, Card{
		Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Owner: seat, Controller: seat,
	})
	anthem = pushTypedTestCard(g, Card{
		Name: "Glorious Anthem", TypeLine: "Enchantment", OracleID: anthemOracle,
		Owner: seat, Controller: seat,
	})
	if got := layeredBattlefieldCard(t, g, bear).Effective().Power; got != 3 {
		t.Fatalf("fixture is wrong: the anthem should be pumping; power = %d", got)
	}

	pushTypedTestCard(g, Card{
		Name: "Silencer", TypeLine: "Enchantment", OracleID: silencerOracle,
		Owner: seat, Controller: seat,
	})
	if got := layeredBattlefieldCard(t, g, bear).Effective().Power; got != 2 {
		t.Errorf("the silenced anthem is still pumping: power = %d, want 2", got)
	}
}

// --- what survives ------------------------------------------------

// CR 613.1f removes abilities. It does not remove the object, its
// counters, its name, its controller or its attachments — and a
// reader who has just seen everything else go quiet has every reason
// to want that stated as a test rather than as a comment.
func TestAbilityRemovalKeepsEverythingThatIsNotAnAbility(t *testing.T) {
	g := newActiveGame(t)
	seat := g.Seats[0].ID
	var victim uuid.UUID

	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		if oracleID == silencerOracle {
			return silencerStatics(&victim)
		}
		return nil
	})

	host := uuid.New()
	victim = pushTypedTestCard(g, Card{
		Name: "Victim", TypeLine: "Creature — Hydra", Power: 3, Toughness: 3,
		OracleID: victimOracle, Owner: seat, Controller: seat,
		Counters:   map[string]int{"+1/+1": 4},
		AttachedTo: TargetRef{Kind: TargetCard, ID: host},
	})
	pushTypedTestCard(g, Card{
		Name: "Silencer", TypeLine: "Enchantment", OracleID: silencerOracle,
		Owner: seat, Controller: seat,
	})

	c := layeredBattlefieldCard(t, g, victim)
	if !c.HasLostAllAbilities() {
		t.Fatal("fixture is wrong: the victim should have lost its abilities")
	}
	if c.Counters["+1/+1"] != 4 {
		t.Errorf("counters are objects on the permanent, not abilities: %v", c.Counters)
	}
	if c.Effective().Name != "Victim" {
		t.Errorf("the name is layer 3, not layer 6: %q", c.Effective().Name)
	}
	if c.Controller != seat {
		t.Errorf("controller moved: %s", c.Controller)
	}
	if !c.IsAttachedTo(host) {
		t.Error("ability removal is not an unattach")
	}
	if !c.IsCreature() {
		t.Error("ability removal is not a type change; layer 4 is a different layer")
	}
	if got := c.Effective().Power; got != 3 {
		t.Errorf("printed P/T is layer 7, applied after 6: power = %d, want 3", got)
	}
}

// CR 305.7's other half. An effect that turns a permanent into a
// Forest removes the abilities its rules text generated and grants
// the mana ability of the new land type — so the declared half goes
// and the intrinsic half arrives, from the same effect, in the same
// recompute.
func TestIntrinsicLandManaAbilitySurvivesAbilityRemoval(t *testing.T) {
	g := newActiveGame(t)
	seat := g.Seats[0].ID
	var victim uuid.UUID

	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		if oracleID != silencerOracle {
			return nil
		}
		out := silencerStatics(&victim)
		return append(out, StaticAbility{
			Layer: Layer4Type,
			AppliesTo: func(target *Card, _ *Game, _ *Card) bool {
				return target.InstanceID == victim
			},
			Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
				c.Types = []string{"Land"}
				c.Subtypes = []string{"Forest"}
			},
		})
	})
	prev := CatalogManaAbilities
	CatalogManaAbilities = func(oracleID string) []ManaAbilityShape {
		if oracleID != victimOracle {
			return nil
		}
		return []ManaAbilityShape{{TapCost: true, Produced: "{C}{C}", Label: "Add {C}{C}"}}
	}
	t.Cleanup(func() { CatalogManaAbilities = prev })

	victim = pushTypedTestCard(g, Card{
		Name: "Sol Ring", TypeLine: "Artifact", OracleID: victimOracle,
		Owner: seat, Controller: seat,
	})
	if got := ManaAbilitiesForCard(layeredBattlefieldCard(t, g, victim)); len(got) != 1 || got[0].Produced != "{C}{C}" {
		t.Fatalf("fixture is wrong: declared ability missing, got %v", got)
	}

	pushTypedTestCard(g, Card{
		Name: "Song", TypeLine: "Enchantment — Aura", OracleID: silencerOracle,
		Owner: seat, Controller: seat,
	})

	got := ManaAbilitiesForCard(layeredBattlefieldCard(t, g, victim))
	if len(got) != 1 {
		t.Fatalf("mana abilities = %v, want exactly the intrinsic Forest one", got)
	}
	if got[0].Produced != "{G}" {
		t.Errorf("produced %q, want {G} — the declared {C}{C} must go and the land type's {G} must arrive",
			got[0].Produced)
	}
}

// --- timestamps (CR 613.6) ---------------------------------------

// The point of the rule: an ability GRANTED by an effect with a later
// timestamp than the removal still applies. Both directions, because
// only having the surviving one would pass with a removal that did
// nothing at all.
func TestALaterGrantSurvivesTheRemovalAndAnEarlierOneDoesNot(t *testing.T) {
	for _, tc := range []struct {
		name       string
		grantFirst bool
		want       bool
	}{
		{name: "granted before the removal", grantFirst: true, want: false},
		{name: "granted after the removal", grantFirst: false, want: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			seat := g.Seats[0].ID
			var victim uuid.UUID

			withStaticAbilities(t, func(oracleID string) []StaticAbility {
				switch oracleID {
				case silencerOracle:
					return silencerStatics(&victim)
				case "static-flying-granter":
					return []StaticAbility{{
						Layer: Layer6Ability,
						AppliesTo: func(target *Card, _ *Game, _ *Card) bool {
							return target.InstanceID == victim
						},
						Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
							c.Abilities = append(c.Abilities, "flying")
						},
					}}
				}
				return nil
			})

			victim = pushTypedTestCard(g, Card{
				Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
				Owner: seat, Controller: seat,
			})
			grant := Card{
				Name: "Grant", TypeLine: "Enchantment", OracleID: "static-flying-granter",
				Owner: seat, Controller: seat,
			}
			remove := Card{
				Name: "Silencer", TypeLine: "Enchantment", OracleID: silencerOracle,
				Owner: seat, Controller: seat,
			}
			// EnteredBattlefieldAt is stamped from the clock in
			// entry order, so pushing in the wanted order IS the
			// timestamp order (CR 613.7).
			if tc.grantFirst {
				pushTypedTestCard(g, grant)
				pushTypedTestCard(g, remove)
			} else {
				pushTypedTestCard(g, remove)
				pushTypedTestCard(g, grant)
			}

			if got := hasKeywordOnBattlefield(t, g, victim, "flying"); got != tc.want {
				t.Errorf("flying = %v, want %v", got, tc.want)
			}
		})
	}
}

// Printed abilities have no such reprieve. They are part of the
// object rather than granted to it, so a removal that applies at all
// removes them whichever of the two arrived first — which is why
// AbilitiesRemoved is a bool and not a timestamp.
func TestPrintedAbilitiesAreRemovedRegardlessOfEntryOrder(t *testing.T) {
	for _, victimFirst := range []bool{true, false} {
		name := "victim entered first"
		if !victimFirst {
			name = "silencer entered first"
		}
		t.Run(name, func(t *testing.T) {
			g := newActiveGame(t)
			seat := g.Seats[0].ID
			var victim uuid.UUID

			withStaticAbilities(t, func(oracleID string) []StaticAbility {
				if oracleID == silencerOracle {
					return silencerStatics(&victim)
				}
				return nil
			})
			prev := CatalogActivatedAbilities
			CatalogActivatedAbilities = func(oracleID string) []ActivatedAbilityShape {
				if oracleID != victimOracle {
					return nil
				}
				return []ActivatedAbilityShape{{Label: "{T}: do a thing"}}
			}
			t.Cleanup(func() { CatalogActivatedAbilities = prev })

			victimCard := Card{
				InstanceID: uuid.New(),
				Name:       "Victim", TypeLine: "Artifact", OracleID: victimOracle,
				Owner: seat, Controller: seat,
			}
			victim = victimCard.InstanceID
			silencer := Card{
				Name: "Silencer", TypeLine: "Enchantment", OracleID: silencerOracle,
				Owner: seat, Controller: seat,
			}
			if victimFirst {
				pushTypedTestCard(g, victimCard)
				pushTypedTestCard(g, silencer)
			} else {
				pushTypedTestCard(g, silencer)
				pushTypedTestCard(g, victimCard)
			}

			if got := ActivatedAbilitiesForCard(layeredBattlefieldCard(t, g, victim)); len(got) != 0 {
				t.Errorf("printed abilities survived: %v", got)
			}
		})
	}
}

// Two ability removers pointed at each other. Layer 6 is walked in
// timestamp order and an effect whose source has already been
// silenced does not apply, so the EARLIER one wins and the board
// settles — rather than both cancelling, oscillating, or depending
// on which pass the fixed point stopped at.
func TestTwoAbilityRemoversPointedAtEachOtherSortByTimestamp(t *testing.T) {
	g := newActiveGame(t)
	seat := g.Seats[0].ID
	var first, second uuid.UUID

	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		switch oracleID {
		case "silencer-a":
			return silencerStatics(&second)
		case "silencer-b":
			return silencerStatics(&first)
		}
		return nil
	})

	first = pushTypedTestCard(g, Card{
		Name: "Song A", TypeLine: "Enchantment — Aura", OracleID: "silencer-a",
		Owner: seat, Controller: seat,
	})
	second = pushTypedTestCard(g, Card{
		Name: "Song B", TypeLine: "Enchantment — Aura", OracleID: "silencer-b",
		Owner: seat, Controller: seat,
	})

	if layeredBattlefieldCard(t, g, first).HasLostAllAbilities() {
		t.Error("the earlier remover applied first and must keep its own abilities")
	}
	if !layeredBattlefieldCard(t, g, second).HasLostAllAbilities() {
		t.Error("the later remover should have lost its abilities to the earlier one")
	}
}

// A chain: A silences B, B would silence C. B never gets to, because
// it has no abilities by the time its own effect would apply. C is
// untouched — including in layers 1-5, which is what the second
// recompute pass buys.
func TestASilencedRemoverDoesNotSilenceItsOwnTarget(t *testing.T) {
	g := newActiveGame(t)
	seat := g.Seats[0].ID
	var b, c uuid.UUID

	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		switch oracleID {
		case "silencer-a":
			return silencerStatics(&b)
		case "silencer-b":
			return silencerStatics(&c)
		}
		return nil
	})

	pushTypedTestCard(g, Card{
		Name: "Song A", TypeLine: "Enchantment — Aura", OracleID: "silencer-a",
		Owner: seat, Controller: seat,
	})
	b = pushTypedTestCard(g, Card{
		Name: "Song B", TypeLine: "Enchantment — Aura", OracleID: "silencer-b",
		Owner: seat, Controller: seat,
	})
	c = pushTypedTestCard(g, Card{
		Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Owner: seat, Controller: seat,
	})

	if !layeredBattlefieldCard(t, g, b).HasLostAllAbilities() {
		t.Fatal("B should have been silenced by A")
	}
	if layeredBattlefieldCard(t, g, c).HasLostAllAbilities() {
		t.Error("C was silenced by a permanent that has no abilities left to silence it with")
	}
}

// The fast path is load-bearing: the recompute runs the layer pass
// twice only when something was actually silenced, so an ordinary
// board pays one extra battlefield scan and nothing else.
func TestAnOrdinaryBoardStillSettlesInOnePass(t *testing.T) {
	g := newActiveGame(t)
	seat := g.Seats[0].ID
	withStaticAbilities(t, func(string) []StaticAbility { return nil })

	id := pushTypedTestCard(g, Card{
		Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Owner: seat, Controller: seat,
	})

	var passes int
	g.WithWriteLock(func() {
		g.layerVersion.Add(1)
		before := g.recomputeCount.Load()
		g.RecomputeLayersIfStaleLocked()
		passes = int(g.recomputeCount.Load() - before)
	})
	if passes != 1 {
		t.Errorf("recompute ran %d times, want 1", passes)
	}
	if layeredBattlefieldCard(t, g, id).HasLostAllAbilities() {
		t.Error("nothing should be silenced on a board with no removal effect")
	}
}
