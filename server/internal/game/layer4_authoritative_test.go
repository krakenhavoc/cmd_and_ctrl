package game

import (
	"testing"

	"github.com/google/uuid"
)

// layer4_authoritative_test.go is the regression suite for #255 /
// #258 / #344 / #348: before this, the layer engine computed
// Effective().Types and then nothing in the engine read it. A
// Layer-4 type change was projected onto the wire and was
// simultaneously invisible to combat, state-based actions,
// targeting and the synthetic basic-land mana ability, because
// Card.IsCreature / IsLand / IsArtifact all read the PRINTED
// TypeLine.
//
// These tests drive the type predicates through a stubbed
// CatalogStaticAbilities hook so they pin the engine contract
// without depending on any particular catalog card. The catalog
// side (Urborg, Tomb of Yawgmoth) is pinned in
// cards/effects/urborg_test.go.

// withStaticAbilities installs a CatalogStaticAbilities stub for
// the duration of one test and restores the previous hook after.
func withStaticAbilities(t *testing.T, fn func(oracleID string) []StaticAbility) {
	t.Helper()
	prev := CatalogStaticAbilities
	CatalogStaticAbilities = fn
	t.Cleanup(func() { CatalogStaticAbilities = prev })
}

// pushTypedTestCard seeds a card on the battlefield and fires the
// zone-move event so the layer listener stamps the CR 613
// timestamp and invalidates the cached resolution.
func pushTypedTestCard(g *Game, c Card) uuid.UUID {
	if c.InstanceID == uuid.Nil {
		c.InstanceID = uuid.New()
	}
	g.Battlefield.PushTop(c)
	g.WithWriteLock(func() {
		g.EmitEvent(Event{
			Kind:    EventZoneMove,
			CardID:  c.InstanceID,
			OldZone: ZoneHand,
			NewZone: ZoneBattlefield,
		})
	})
	return c.InstanceID
}

// layeredBattlefieldCard returns a copy of the named battlefield card
// after forcing a layer recompute.
func layeredBattlefieldCard(t *testing.T, g *Game, id uuid.UUID) Card {
	t.Helper()
	var out Card
	found := false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				out, found = c, true
				return
			}
		}
	})
	if !found {
		t.Fatalf("card %s not on battlefield", id)
	}
	return out
}

// printedEffectiveWith is the shape a hand-built Card.effective has
// to have now that the type predicates read it: seeded from the
// card's PRINTED characteristic, with `abilities` layered on top.
//
// Fixtures that stamped `&Characteristic{Power, Toughness,
// Abilities}` directly used to be harmless — nothing read Types, so
// leaving them empty cost nothing. It declares a permanent with no
// card types at all, which is exactly what the layer engine would
// produce for a creature that had lost every type, so with the
// predicates rerouted the fixture's "creature" stopped being one.
// Going through the printed characteristic keeps a fixture saying
// what it means.
func printedEffectiveWith(c Card, abilities ...string) *Characteristic {
	eff := c.printedCharacteristic()
	eff.Abilities = append(eff.Abilities, abilities...)
	return &eff
}

const (
	typeAdderOracle   = "static-type-adder"
	typeRemoverOracle = "static-type-remover"
)

// TestLayer4TypeAddMakesIsCreatureTrue — "all lands are 1/1
// creatures" (Kormus Bell / Awakening-style). The land must report
// IsCreature() == true once the Layer-4 effect is in play.
func TestLayer4TypeAddMakesIsCreatureTrue(t *testing.T) {
	g := newActiveGame(t)
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		if oracleID != typeAdderOracle {
			return nil
		}
		return []StaticAbility{{
			Layer: Layer4Type,
			AppliesTo: func(target *Card, g *Game, source *Card) bool {
				return target.IsLand()
			},
			Apply: func(c *Characteristic, target *Card, g *Game, source *Card) {
				c.Types = append(c.Types, "Creature")
			},
		}}
	})
	seat := g.Seats[0].ID
	landID := pushTypedTestCard(g, Card{
		Name: "Forest", TypeLine: "Basic Land — Forest",
		Owner: seat, Controller: seat,
	})
	pushTypedTestCard(g, Card{
		Name: "Kormus Bell", TypeLine: "Artifact", OracleID: typeAdderOracle,
		Owner: seat, Controller: seat,
	})

	land := layeredBattlefieldCard(t, g, landID)
	if !land.IsCreature() {
		t.Errorf("land under a Layer-4 type-add: IsCreature() = false, want true")
	}
	if !land.IsLand() {
		t.Errorf("type-ADD must not strip the printed type: IsLand() = false, want true")
	}
}

// TestLayer4TypeRemovalMakesIsCreatureFalse — the devotion-gated
// "isn't a creature" clause every Theros god carries. #255 named
// this as the class of card the printed-TypeLine read made
// unimplementable.
func TestLayer4TypeRemovalMakesIsCreatureFalse(t *testing.T) {
	g := newActiveGame(t)
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		if oracleID != typeRemoverOracle {
			return nil
		}
		return []StaticAbility{{
			Layer: Layer4Type,
			AppliesTo: func(target *Card, g *Game, source *Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *Characteristic, target *Card, g *Game, source *Card) {
				kept := c.Types[:0]
				for _, ty := range c.Types {
					if ty != "Creature" {
						kept = append(kept, ty)
					}
				}
				c.Types = kept
			},
		}}
	})
	seat := g.Seats[0].ID
	godID := pushTypedTestCard(g, Card{
		Name: "Thassa, Deep-Dwelling", TypeLine: "Legendary Enchantment Creature — God",
		Power: 6, Toughness: 5, OracleID: typeRemoverOracle,
		Owner: seat, Controller: seat,
	})

	god := layeredBattlefieldCard(t, g, godID)
	if god.IsCreature() {
		t.Errorf("permanent under a Layer-4 type-removal: IsCreature() = true, want false")
	}
	if !god.IsEnchantment() {
		t.Errorf("IsEnchantment() = false, want true (only Creature was removed)")
	}
	if god.IsPermanent() != true {
		t.Errorf("IsPermanent() = false, want true")
	}
}

// TestLayer4TypeRemovalBlocksAttackDeclaration is the "it actually
// reaches combat" half. #344: a permanent that stops being a
// creature "would still attack, block, be targeted as a creature,
// and feed its own trigger."
func TestLayer4TypeRemovalBlocksAttackDeclaration(t *testing.T) {
	g := newActiveGame(t)
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		if oracleID != typeRemoverOracle {
			return nil
		}
		return []StaticAbility{{
			Layer: Layer4Type,
			AppliesTo: func(target *Card, g *Game, source *Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *Characteristic, target *Card, g *Game, source *Card) {
				c.Types = nil
			},
		}}
	})
	seat := g.Seats[0].ID
	defender := g.Seats[1].ID
	godID := pushTypedTestCard(g, Card{
		Name: "Thassa, Deep-Dwelling", TypeLine: "Legendary Enchantment Creature — God",
		Power: 6, Toughness: 5, OracleID: typeRemoverOracle,
		Owner: seat, Controller: seat,
	})
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			g.Battlefield.Cards[i].SummonedThisTurn = false
		}
		g.Turn.Step = StepDeclareAttackers
	})
	if err := g.DeclareAttacker(godID, defender); err == nil {
		t.Errorf("DeclareAttacker on a non-creature succeeded, want an error")
	}
}

// TestPrintedTypesUnaffectedOffBattlefield — Effective() falls back
// to the printed characteristic when the card is not on the
// battlefield, so a hand card keeps reading its printed type line
// even while a global Layer-4 effect is in play.
func TestPrintedTypesUnaffectedOffBattlefield(t *testing.T) {
	g := newActiveGame(t)
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		if oracleID != typeAdderOracle {
			return nil
		}
		return []StaticAbility{{
			Layer:     Layer4Type,
			AppliesTo: func(target *Card, g *Game, source *Card) bool { return true },
			Apply: func(c *Characteristic, target *Card, g *Game, source *Card) {
				c.Types = append(c.Types, "Creature")
			},
		}}
	})
	seat := g.Seats[0].ID
	pushTypedTestCard(g, Card{
		Name: "Kormus Bell", TypeLine: "Artifact", OracleID: typeAdderOracle,
		Owner: seat, Controller: seat,
	})
	inHand := Card{Name: "Swords to Plowshares", TypeLine: "Instant", Owner: seat, Controller: seat}
	g.ReadSnapshot(func() {})
	if inHand.IsCreature() {
		t.Errorf("off-battlefield card read a battlefield-only static: IsCreature() = true, want false")
	}
	if !inHand.IsInstant() {
		t.Errorf("off-battlefield card lost its printed type: IsInstant() = false, want true")
	}
}

// TestPrintedIsAccessorsIgnoreLayers pins the printed-vs-effective
// split: PrintedIs* always reads the printed type line, whatever
// the layer engine says.
func TestPrintedIsAccessorsIgnoreLayers(t *testing.T) {
	g := newActiveGame(t)
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		if oracleID != typeAdderOracle {
			return nil
		}
		return []StaticAbility{{
			Layer:     Layer4Type,
			AppliesTo: func(target *Card, g *Game, source *Card) bool { return target.IsLand() },
			Apply: func(c *Characteristic, target *Card, g *Game, source *Card) {
				c.Types = append(c.Types, "Creature")
			},
		}}
	})
	seat := g.Seats[0].ID
	landID := pushTypedTestCard(g, Card{
		Name: "Forest", TypeLine: "Basic Land — Forest",
		Owner: seat, Controller: seat,
	})
	pushTypedTestCard(g, Card{
		Name: "Kormus Bell", TypeLine: "Artifact", OracleID: typeAdderOracle,
		Owner: seat, Controller: seat,
	})
	land := layeredBattlefieldCard(t, g, landID)
	if !land.IsCreature() {
		t.Fatalf("precondition: effective IsCreature() = false, want true")
	}
	if land.PrintedIsCreature() {
		t.Errorf("PrintedIsCreature() = true, want false — printed accessors must ignore layers")
	}
	if !land.PrintedIsLand() {
		t.Errorf("PrintedIsLand() = false, want true")
	}
}

// TestLayer4LandTypeGrantsIntrinsicMana is the Urborg mechanism at
// the engine level, and the exact reason #258 refused to catalogue
// Urborg: the synthetic basic-land mana ability read the printed
// TypeLine, so a Layer-4 "is a Swamp" static "would apply cleanly
// and do nothing."
//
// CR 305.6: the intrinsic mana ability comes from the basic LAND
// TYPE, not from the Basic supertype — Urborg grants the Swamp
// subtype, never the supertype.
func TestLayer4LandTypeGrantsIntrinsicMana(t *testing.T) {
	g := newActiveGame(t)
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		if oracleID != typeAdderOracle {
			return nil
		}
		return []StaticAbility{{
			Layer: Layer4Type,
			AppliesTo: func(target *Card, g *Game, source *Card) bool {
				return target.IsLand()
			},
			Apply: func(c *Characteristic, target *Card, g *Game, source *Card) {
				c.Subtypes = append(c.Subtypes, "Swamp")
			},
		}}
	})
	seat := g.Seats[0].ID
	forestID := pushTypedTestCard(g, Card{
		Name: "Forest", TypeLine: "Basic Land — Forest",
		Owner: seat, Controller: seat,
	})

	// Baseline: without the static the Forest taps for {G} only.
	forest := layeredBattlefieldCard(t, g, forestID)
	if got := ManaAbilitiesForCard(forest); len(got) != 1 || got[0].Produced != "{G}" {
		t.Fatalf("bare Forest mana abilities = %+v, want one {G}", got)
	}

	pushTypedTestCard(g, Card{
		Name: "Urborg, Tomb of Yawgmoth", TypeLine: "Legendary Land", OracleID: typeAdderOracle,
		Owner: seat, Controller: seat,
	})
	forest = layeredBattlefieldCard(t, g, forestID)
	abilities := ManaAbilitiesForCard(forest)
	if len(abilities) != 2 {
		t.Fatalf("Forest under a Swamp-granting static: %d mana abilities, want 2 (%+v)", len(abilities), abilities)
	}
	// Printed type first so the auto-tapper's "first ability"
	// choice is unchanged for every land that has no static on it.
	if abilities[0].Produced != "{G}" {
		t.Errorf("ability[0].Produced = %q, want {G} (printed type keeps its slot)", abilities[0].Produced)
	}
	if abilities[1].Produced != "{B}" {
		t.Errorf("ability[1].Produced = %q, want {B}", abilities[1].Produced)
	}
}

// TestNonbasicLandTypeProducesMana pins the CR 305.6 half that is
// independent of layers: a nonbasic land printed with a basic land
// type has that type's mana ability. battle_lands.go documented
// the gap ("the engine's synthetic mana ability only fires for
// BASIC lands") as the reason the Battle land cycle declares its
// pipe ability by hand.
func TestNonbasicLandTypeProducesMana(t *testing.T) {
	c := Card{Name: "Bayou", TypeLine: "Land — Swamp Forest"}
	abilities := ManaAbilitiesForCard(c)
	if len(abilities) != 2 {
		t.Fatalf("Bayou mana abilities = %+v, want 2", abilities)
	}
	if abilities[0].Produced != "{B}" || abilities[1].Produced != "{G}" {
		t.Errorf("produced = %q,%q, want {B},{G}", abilities[0].Produced, abilities[1].Produced)
	}
}
