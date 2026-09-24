package game

import (
	"testing"

	"github.com/google/uuid"
)

// layer_invalidation_cast_test.go covers #1325: a static ability
// reading "you haven't cast a spell this turn" (Stoic Sphinx) has a
// new answer the instant a spell is cast, and nothing bumped
// `layerVersion` for that — hand-size, life-total and attacking-status
// invalidation (#74, #1117, #1218) all shipped before this gap was
// noticed. Every assertion here is made with NO other event in
// between: a battlefield move would drop the cached resolution on its
// own and hide a missing bump.

const spellsCastStaticOracle = "spells-cast-static"

// spellsCastGatedHexproofForTest is a stub in the Stoic Sphinx shape:
// a Layer 6 self-only hexproof grant that is live only while its
// controller hasn't cast a spell this turn, declaring the dependency
// so the listener knows to invalidate.
func spellsCastGatedHexproofForTest(declare bool) StaticAbility {
	return StaticAbility{
		Layer:               Layer6Ability,
		DependsOnSpellsCast: declare,
		AppliesTo: func(target *Card, g *Game, source *Card) bool {
			return target.InstanceID == source.InstanceID && g.CastTallyFor(source.Controller).Total == 0
		},
		Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
			c.Abilities = append(c.Abilities, "hexproof")
		},
	}
}

// pushSpellsCastGatedCreature seeds a creature carrying the stub and
// returns it.
func pushSpellsCastGatedCreature(t *testing.T, g *Game, owner *Player) uuid.UUID {
	t.Helper()
	return pushTypedTestCard(g, Card{
		Name:       "Sphinx Stand-In",
		TypeLine:   "Creature — Sphinx",
		OracleID:   spellsCastStaticOracle,
		Power:      5,
		Toughness:  3,
		Owner:      owner.ID,
		Controller: owner.ID,
	})
}

// pushFreeInstant puts a zero-mana-cost instant in p's hand — cheap
// enough to cast in these tests without any mana-pool setup, and an
// instant so it can be cast at any point priority is open.
func pushFreeInstant(g *Game, p *Player, name string) uuid.UUID {
	c := NewCard(name, p.ID)
	c.TypeLine = "Instant"
	p.Hand.PushTop(c)
	return c.InstanceID
}

func TestLayerVersionBumpsOnEventCastWhenASpellsCastStaticIsLive(t *testing.T) {
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		if oracleID == spellsCastStaticOracle {
			return []StaticAbility{spellsCastGatedHexproofForTest(true)}
		}
		return nil
	})
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	pushSpellsCastGatedCreature(t, g, me)
	spell := pushFreeInstant(g, me, "Nothing Much")

	before := readLayerVersion(g)
	if err := g.CastSpell(me.ID, spell, CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	if got := readLayerVersion(g); got <= before {
		t.Errorf("layerVersion did not bump on EventCast while a spells-cast-keyed static was in play: was %d, now %d", before, got)
	}
}

func TestLayerVersionIgnoresEventCastForAStaticThatDoesNotDeclareIt(t *testing.T) {
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		if oracleID == spellsCastStaticOracle {
			return []StaticAbility{spellsCastGatedHexproofForTest(false)}
		}
		return nil
	})
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	pushSpellsCastGatedCreature(t, g, me)
	spell := pushFreeInstant(g, me, "Nothing Much")

	before := readLayerVersion(g)
	if err := g.CastSpell(me.ID, spell, CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	if got := readLayerVersion(g); got != before {
		t.Errorf("layerVersion bumped on EventCast with no spells-cast-keyed static in play: was %d, now %d", before, got)
	}
}

// TestStoicSphinxStyleHexproofTurnsOffAfterACast is the end-to-end
// shape: the gated grant reads as live before any spell is cast this
// turn, and stops applying — not merely "the cache didn't update", but
// the ACTUAL resolved characteristic — the moment one is.
func TestStoicSphinxStyleHexproofTurnsOffAfterACast(t *testing.T) {
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		if oracleID == spellsCastStaticOracle {
			return []StaticAbility{spellsCastGatedHexproofForTest(true)}
		}
		return nil
	})
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	id := pushSpellsCastGatedCreature(t, g, me)
	spell := pushFreeInstant(g, me, "Nothing Much")

	if got := layeredBattlefieldCard(t, g, id).Effective().Abilities; !eotHasAbilityForTest(got, "hexproof") {
		t.Fatalf("expected hexproof before any spell was cast this turn, got %v", got)
	}
	if err := g.CastSpell(me.ID, spell, CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	if got := layeredBattlefieldCard(t, g, id).Effective().Abilities; eotHasAbilityForTest(got, "hexproof") {
		t.Fatalf("expected hexproof to be gone after casting a spell this turn, got %v", got)
	}
}
