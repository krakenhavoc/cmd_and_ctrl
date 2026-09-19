package game

import (
	"testing"

	"github.com/google/uuid"
)

// layer_dependency_test.go is the engine-side suite for ADR 0067:
// CR 613.8 dependency ordering inside a layer, and CR 613.6's rule
// about an effect whose source loses its abilities part-way through
// the pass.
//
// It drives stubbed catalog hooks, like layer4_authoritative_test.go
// and layer6_ability_removal_test.go, so the CR is pinned
// independently of any one card. The catalog side — Urborg + Song of
// the Dryads, Maskwood Nexus + a crewed Vehicle, Magus of the Moon
// under Kenrith's Transformation — is in
// cards/effects/layer_dependency_pairs_test.go and
// cards/effects/ability_removal_test.go.
//
// The two rulings these tests exist for:
//
//	Magus of the Moon, 2021-03-19   "If Magus of the Moon loses its
//	                                abilities, it continues to turn
//	                                nonbasic lands into Mountains."
//	Humility + Opalescence,         all creatures end up 1/1 with no
//	2009-10-01                      abilities when Opalescence is
//	                                older, and the animated
//	                                enchantments keep their own base
//	                                P/T when Humility is older.

const (
	moonOracle         = "layerdep-moon"
	moonSilencerOracle = "layerdep-silencer"
	anthemSourceOracle = "layerdep-anthem"
	humilityOracle     = "layerdep-humility"
	opalescenceOracle  = "layerdep-opalescence"
	swampAdderOracle   = "layerdep-swamp-adder"
	landMakerOracle    = "layerdep-land-maker"
	loopAOracle        = "layerdep-loop-a"
	loopBOracle        = "layerdep-loop-b"
)

// removalTargeting is a bare layer-6 "that permanent loses all
// abilities", with no host relation, so a fixture stays about the
// layer and not about attachments.
func removalTargeting(victim *uuid.UUID) StaticAbility {
	return StaticAbility{
		Layer:            Layer6Ability,
		RemovesAbilities: true,
		AppliesTo: func(target *Card, _ *Game, _ *Card) bool {
			return target.InstanceID == *victim
		},
		Apply: func(_ *Characteristic, _ *Card, _ *Game, _ *Card) {},
	}
}

// --- CR 613.6: a removal reaches forwards only ---------------------

// Magus of the Moon's ruling, as the engine contract: the Magus's
// layer-4 effect applies in layer 4, which is over by the time layer
// 6 takes its abilities away. Before #669 the recompute held a
// silenced source out of the GATHER, so the Magus contributed nothing
// to any layer and the Mountains disappeared.
func TestASilencedSourceKeepsItsEarlierLayerEffect(t *testing.T) {
	g := newActiveGame(t)
	var magus uuid.UUID
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		switch oracleID {
		case moonOracle:
			return []StaticAbility{{
				Layer: Layer4Type,
				AppliesTo: func(target *Card, _ *Game, _ *Card) bool {
					return target.IsLand() && !target.HasSupertype("basic")
				},
				Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
					c.SetSubtypes([]string{"Mountain"})
				},
			}}
		case moonSilencerOracle:
			return []StaticAbility{removalTargeting(&magus)}
		}
		return nil
	})
	seat := g.Seats[0].ID
	land := pushTypedTestCard(g, Card{
		Name: "Bojuka Bog", TypeLine: "Land", Owner: seat, Controller: seat,
	})
	magus = pushTypedTestCard(g, Card{
		Name: "Magus of the Moon", TypeLine: "Creature — Human Wizard",
		Power: 2, Toughness: 2, OracleID: moonOracle, Owner: seat, Controller: seat,
	})
	pushTypedTestCard(g, Card{
		Name: "Silencer", TypeLine: "Enchantment — Aura",
		OracleID: moonSilencerOracle, Owner: seat, Controller: seat,
	})

	if c := layeredBattlefieldCard(t, g, magus); !c.HasLostAllAbilities() {
		t.Fatal("fixture is wrong: the Magus was not silenced")
	}
	if c := layeredBattlefieldCard(t, g, land); !c.HasSubtype("Mountain") {
		t.Errorf("a silenced Magus stopped making Mountains: subtypes %v", c.Effective().Subtypes)
	}
}

// The other side of the same rule, and the reason it is not "a
// silenced source keeps doing everything": an ability whose only
// applicable layer comes AFTER the removal never starts applying, so
// it stops. An animated Glorious Anthem under a Humility gives
// nothing.
func TestASilencedSourceLosesItsLaterLayerEffect(t *testing.T) {
	g := newActiveGame(t)
	var anthem uuid.UUID
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		switch oracleID {
		case anthemSourceOracle:
			return []StaticAbility{{
				Layer:    Layer7PT,
				SubLayer: SubLayer7C_Modify,
				AppliesTo: func(target *Card, _ *Game, _ *Card) bool {
					return target.IsCreature()
				},
				Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
					c.Power++
					c.Toughness++
				},
			}}
		case moonSilencerOracle:
			return []StaticAbility{removalTargeting(&anthem)}
		}
		return nil
	})
	seat := g.Seats[0].ID
	bear := pushTypedTestCard(g, Card{
		Name: "Grizzly Bears", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: seat, Controller: seat,
	})
	anthem = pushTypedTestCard(g, Card{
		Name: "Glorious Anthem", TypeLine: "Enchantment",
		OracleID: anthemSourceOracle, Owner: seat, Controller: seat,
	})
	pushTypedTestCard(g, Card{
		Name: "Silencer", TypeLine: "Enchantment — Aura",
		OracleID: moonSilencerOracle, Owner: seat, Controller: seat,
	})

	if got := layeredBattlefieldCard(t, g, bear).Effective().Power; got != 2 {
		t.Errorf("Bear power = %d, want 2: a silenced anthem has no layer-7c effect to continue", got)
	}
}

// --- Humility + Opalescence ----------------------------------------

// humilityStatics is "All creatures lose all abilities and have base
// power and toughness 1/1" — one continuous effect, two layers, so
// the 7b half declares ContinuesAfterRemoval.
func humilityStatics() []StaticAbility {
	isCreature := func(target *Card, _ *Game, _ *Card) bool { return target.IsCreature() }
	return []StaticAbility{
		{
			Layer:                 Layer6Ability,
			RemovesAbilities:      true,
			ContinuesAfterRemoval: true,
			AppliesTo:             isCreature,
			Apply:                 func(_ *Characteristic, _ *Card, _ *Game, _ *Card) {},
		},
		{
			Layer:                 Layer7PT,
			SubLayer:              SubLayer7B_Set,
			ContinuesAfterRemoval: true,
			AppliesTo:             isCreature,
			Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
				c.Power, c.Toughness = 1, 1
			},
		},
	}
}

// opalescenceStatics is "Each other non-Aura enchantment is a
// creature in addition to its other types and has base power and
// toughness each equal to its mana value" — with the mana value
// fixed at 4 for every fixture card here, because this test is about
// layers and not about cost parsing.
func opalescenceStatics() []StaticAbility {
	other := func(target *Card, _ *Game, source *Card) bool {
		return target.InstanceID != source.InstanceID &&
			target.IsEnchantment() && !target.HasSubtype("Aura")
	}
	return []StaticAbility{
		{
			Layer:     Layer4Type,
			AppliesTo: other,
			Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
				c.Types = append(c.Types, "Creature")
			},
		},
		{
			Layer:                 Layer7PT,
			SubLayer:              SubLayer7B_Set,
			ContinuesAfterRemoval: true,
			AppliesTo:             other,
			Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
				c.Power, c.Toughness = 4, 4
			},
		},
	}
}

// humilityBoard seeds a Bear, a plain enchantment, and Humility and
// Opalescence in the requested timestamp order.
func humilityBoard(t *testing.T, humilityFirst bool) (g *Game, bear, plain, humility uuid.UUID) {
	t.Helper()
	g = newActiveGame(t)
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		switch oracleID {
		case humilityOracle:
			return humilityStatics()
		case opalescenceOracle:
			return opalescenceStatics()
		}
		return nil
	})
	seat := g.Seats[0].ID
	bear = pushTypedTestCard(g, Card{
		Name: "Grizzly Bears", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: seat, Controller: seat,
	})
	plain = pushTypedTestCard(g, Card{
		Name: "Ghostly Prison", TypeLine: "Enchantment", Owner: seat, Controller: seat,
	})
	pushHumility := func() {
		humility = pushTypedTestCard(g, Card{
			Name: "Humility", TypeLine: "Enchantment",
			OracleID: humilityOracle, Owner: seat, Controller: seat,
		})
	}
	pushOpalescence := func() {
		pushTypedTestCard(g, Card{
			Name: "Opalescence", TypeLine: "Enchantment",
			OracleID: opalescenceOracle, Owner: seat, Controller: seat,
		})
	}
	if humilityFirst {
		pushHumility()
		pushOpalescence()
	} else {
		pushOpalescence()
		pushHumility()
	}
	return g, bear, plain, humility
}

// Opalescence older: its layer-7b set runs first and Humility's runs
// after, so everything is 1/1 — including Humility itself, which
// Opalescence animated and which then removed its own abilities in
// layer 6. The 1/1 is CR 613.6 in one number: the ability that
// generated it no longer exists by the time layer 7b runs.
func TestHumilityNewerThanOpalescenceMakesEverythingOneOne(t *testing.T) {
	g, bear, plain, humility := humilityBoard(t, false)

	for _, tc := range []struct {
		name string
		id   uuid.UUID
	}{{"Bear", bear}, {"the animated Ghostly Prison", plain}, {"the animated Humility", humility}} {
		c := layeredBattlefieldCard(t, g, tc.id)
		if !c.IsCreature() {
			t.Errorf("%s is not a creature", tc.name)
			continue
		}
		if c.Effective().Power != 1 || c.Effective().Toughness != 1 {
			t.Errorf("%s is %d/%d, want 1/1", tc.name, c.Effective().Power, c.Effective().Toughness)
		}
		if !c.HasLostAllAbilities() {
			t.Errorf("%s kept its abilities under Humility", tc.name)
		}
	}
}

// Humility older: Humility's 1/1 runs first and Opalescence's
// mana-value set runs after it, so the animated enchantments are 4/4
// and only the ordinary creature stays 1/1. Opalescence is never a
// creature ("each OTHER"), so Humility never silences it.
func TestOpalescenceNewerThanHumilityKeepsTheEnchantmentsBig(t *testing.T) {
	g, bear, plain, humility := humilityBoard(t, true)

	if c := layeredBattlefieldCard(t, g, bear); c.Effective().Power != 1 || c.Effective().Toughness != 1 {
		t.Errorf("Bear is %d/%d, want 1/1", c.Effective().Power, c.Effective().Toughness)
	}
	for _, tc := range []struct {
		name string
		id   uuid.UUID
	}{{"the animated Ghostly Prison", plain}, {"the animated Humility", humility}} {
		c := layeredBattlefieldCard(t, g, tc.id)
		if c.Effective().Power != 4 || c.Effective().Toughness != 4 {
			t.Errorf("%s is %d/%d, want 4/4", tc.name, c.Effective().Power, c.Effective().Toughness)
		}
	}
}

// --- CR 613.8 ------------------------------------------------------

// The shape of every catalogued pair, as the engine contract: a
// layer-4 effect whose AppliesTo reads a type another layer-4 effect
// writes applies after it, whichever entered first.
func TestLayerFourDependencyAppliesAfterWhatItDependsOn(t *testing.T) {
	for _, adderFirst := range []bool{true, false} {
		name := "type-setter entered first"
		if adderFirst {
			name = "type-adder entered first"
		}
		t.Run(name, func(t *testing.T) {
			g := newActiveGame(t)
			var subject uuid.UUID
			withStaticAbilities(t, func(oracleID string) []StaticAbility {
				switch oracleID {
				case swampAdderOracle:
					return []StaticAbility{{
						Layer: Layer4Type,
						AppliesTo: func(target *Card, _ *Game, _ *Card) bool {
							return target.IsLand()
						},
						Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
							c.Subtypes = append(c.Subtypes, "Swamp")
						},
					}}
				case landMakerOracle:
					return []StaticAbility{{
						Layer: Layer4Type,
						AppliesTo: func(target *Card, _ *Game, _ *Card) bool {
							return target.InstanceID == subject
						},
						Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
							c.Types = []string{"Land"}
						},
					}}
				}
				return nil
			})
			seat := g.Seats[0].ID
			subject = pushTypedTestCard(g, Card{
				Name: "Sol Ring", TypeLine: "Artifact", Owner: seat, Controller: seat,
			})
			push := func(oracle, name string) {
				pushTypedTestCard(g, Card{
					Name: name, TypeLine: "Enchantment", OracleID: oracle,
					Owner: seat, Controller: seat,
				})
			}
			if adderFirst {
				push(swampAdderOracle, "Urborg-alike")
				push(landMakerOracle, "Song-alike")
			} else {
				push(landMakerOracle, "Song-alike")
				push(swampAdderOracle, "Urborg-alike")
			}

			c := layeredBattlefieldCard(t, g, subject)
			if !c.IsLand() {
				t.Fatal("the subject is not a land")
			}
			if !c.HasSubtype("Swamp") {
				t.Errorf("subtypes = %v, want Swamp: the adder depends on the type-setter (CR 613.8a)",
					c.Effective().Subtypes)
			}
		})
	}
}

// CR 613.8c's escape hatch. Two effects that each change what the
// other applies to cannot both go second, so the rules give up on
// dependency for that bucket and apply by timestamp. The assertion is
// that the pass TERMINATES with the timestamp answer rather than
// spinning or picking arbitrarily.
func TestDependencyLoopFallsBackToTimestampOrder(t *testing.T) {
	g := newActiveGame(t)
	var subject uuid.UUID
	// Each effect applies only while the subject is NOT the type the
	// other writes, so applying either one takes the other's
	// applicability away: a two-effect loop.
	loop := func(mine, theirs string) []StaticAbility {
		return []StaticAbility{{
			Layer: Layer4Type,
			AppliesTo: func(target *Card, _ *Game, _ *Card) bool {
				return target.InstanceID == subject && !target.HasSubtype(theirs)
			},
			Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
				c.SetSubtypes([]string{mine})
			},
		}}
	}
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		switch oracleID {
		case loopAOracle:
			return loop("Forest", "Island")
		case loopBOracle:
			return loop("Island", "Forest")
		}
		return nil
	})
	seat := g.Seats[0].ID
	subject = pushTypedTestCard(g, Card{
		Name: "Sol Ring", TypeLine: "Artifact Land", Owner: seat, Controller: seat,
	})
	pushTypedTestCard(g, Card{
		Name: "Loop A", TypeLine: "Enchantment", OracleID: loopAOracle,
		Owner: seat, Controller: seat,
	})
	pushTypedTestCard(g, Card{
		Name: "Loop B", TypeLine: "Enchantment", OracleID: loopBOracle,
		Owner: seat, Controller: seat,
	})

	// Timestamp order: A applies first and makes it a Forest, which
	// takes B's applicability away.
	if c := layeredBattlefieldCard(t, g, subject); !c.HasSubtype("Forest") || c.HasSubtype("Island") {
		t.Errorf("subtypes = %v, want the older effect's Forest (CR 613.8c loop fallback)",
			c.Effective().Subtypes)
	}
}

// --- #670: every creature type is a layer-4 fact -------------------

// The Maskwood Nexus ruling (2021-02-05) about a PRINTED changeling:
// "If an effect causes a creature with changeling to lose all
// abilities, it will remain all creature types … because changeling
// applies before the effect that removes it." A pure layer-6 removal
// with no type change is the shape that shows it.
func TestPrintedChangelingKeepsEveryCreatureTypeUnderAbilityRemoval(t *testing.T) {
	g := newActiveGame(t)
	var shifter uuid.UUID
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		if oracleID == moonSilencerOracle {
			return []StaticAbility{removalTargeting(&shifter)}
		}
		return nil
	})
	seat := g.Seats[0].ID
	shifter = pushTypedTestCard(g, Card{
		Name: "Woodland Changeling", TypeLine: "Creature — Shapeshifter",
		Power: 2, Toughness: 2, Keywords: []string{KeywordChangeling},
		Owner: seat, Controller: seat,
	})
	pushTypedTestCard(g, Card{
		Name: "Silencer", TypeLine: "Enchantment — Aura",
		OracleID: moonSilencerOracle, Owner: seat, Controller: seat,
	})

	c := layeredBattlefieldCard(t, g, shifter)
	if !c.HasLostAllAbilities() {
		t.Fatal("fixture is wrong: the changeling was not silenced")
	}
	if HasKeyword(&c, KeywordChangeling) {
		t.Error("the changeling KEYWORD survived a layer-6 removal; it is an ability and must not")
	}
	if !HasAllCreatureTypes(&c) {
		t.Error("the silenced changeling stopped being every creature type (CR 702.73a is a layer-4 fact)")
	}
	if !c.HasSubtype("Goblin") {
		t.Error("the silenced changeling is not a Goblin")
	}
}

// The storage decision in one assertion: the fact is on the
// Characteristic and the keyword is only the printed source of it, so
// emptying the ability list cannot take the types away.
func TestPrintedChangelingProjectsIntoTheLayerFourFact(t *testing.T) {
	c := Card{
		Name: "Woodland Changeling", TypeLine: "Creature — Shapeshifter",
		Keywords: []string{KeywordChangeling},
	}
	printed := c.printedCharacteristic()
	if !printed.AllCreatureTypes {
		t.Fatal("printed changeling did not set the layer-4 fact")
	}
	printed.Abilities = nil
	c.effective = &printed
	if !HasAllCreatureTypes(&c) {
		t.Error("emptying Abilities took the types away")
	}
	printed.SetSubtypes([]string{"Elk"})
	if HasAllCreatureTypes(&c) {
		t.Error("a layer-4 subtype SET must clear the fact (CR 205.1b)")
	}
}

// --- undo and clone ------------------------------------------------

// Nothing here is snapshotted. Characteristic is the layer engine's
// per-pass cache — snapshot_drift_test.go classifies `effective` as
// `rebuilt` — so the every-creature-type fact, the removal stamp and
// the CR 613.6 bookkeeping all come back from a recompute rather than
// from a file or an undo frame. This test says so out loud: a clone
// answers the same questions as the original, and it answers them
// after its own pass.
func TestCloneRecomputesTheSameLayerAnswer(t *testing.T) {
	g := newActiveGame(t)
	var shifter uuid.UUID
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		if oracleID == moonSilencerOracle {
			return []StaticAbility{removalTargeting(&shifter)}
		}
		return nil
	})
	seat := g.Seats[0].ID
	shifter = pushTypedTestCard(g, Card{
		Name: "Woodland Changeling", TypeLine: "Creature — Shapeshifter",
		Power: 2, Toughness: 2, Keywords: []string{KeywordChangeling},
		Owner: seat, Controller: seat,
	})
	pushTypedTestCard(g, Card{
		Name: "Silencer", TypeLine: "Enchantment — Aura",
		OracleID: moonSilencerOracle, Owner: seat, Controller: seat,
	})
	before := layeredBattlefieldCard(t, g, shifter)

	clone := g.Clone()
	after := layeredBattlefieldCard(t, clone, shifter)

	if after.HasLostAllAbilities() != before.HasLostAllAbilities() {
		t.Errorf("clone removal stamp = %v, want %v", after.HasLostAllAbilities(), before.HasLostAllAbilities())
	}
	if !HasAllCreatureTypes(&after) {
		t.Error("the clone lost the every-creature-type fact; it must be recomputed, not carried")
	}
}
