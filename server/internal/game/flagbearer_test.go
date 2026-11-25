package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// flagbearer_test.go — #6236, Coalition Flag and Standard Bearer (CR
// 601.2c): an opponent of the rule's controller who is choosing targets
// for a spell they cast must choose a Flagbearer if one is a legal choice.

const (
	testFlagbearerRuleOracle = "test-flagbearer-rule"
	testFlagbearerShock      = "test-flagbearer-shock"
)

func withFlagbearerRule(t *testing.T) {
	t.Helper()
	prev := CatalogStaticAbilities
	t.Cleanup(func() { CatalogStaticAbilities = prev })
	CatalogStaticAbilities = func(key string) []StaticAbility {
		if key == testFlagbearerRuleOracle {
			return []StaticAbility{{
				Layer:     Layer6Ability,
				AppliesTo: func(target *Card, _ *Game, source *Card) bool { return target.InstanceID == source.InstanceID },
				Apply:     func(c *Characteristic, _ *Card, _ *Game, _ *Card) { c.FlagbearerTargetRule = true },
			}}
		}
		if prev == nil {
			return nil
		}
		return prev(key)
	}
	withCatalogTargetSpec(t, func(id string) *TargetSpec {
		if id == testFlagbearerShock {
			return anyCreatureSpec()
		}
		return nil
	})
}

// pushFlagbearer puts a creature of type Flagbearer on the battlefield;
// with rule set it also carries the targeting rule (Standard Bearer).
func pushFlagbearer(t *testing.T, g *Game, owner *Player, name string, rule bool) uuid.UUID {
	t.Helper()
	c := NewCard(name, owner.ID)
	c.TypeLine = "Creature — Human Flagbearer"
	c.Power, c.Toughness = 1, 1
	if rule {
		c.OracleID = testFlagbearerRuleOracle
	}
	g.Battlefield.PushTop(c)
	g.WithWriteLock(func() { g.layerVersion.Add(1) })
	return c.InstanceID
}

func castFlagbearerShock(g *Game, caster *Player, at uuid.UUID) error {
	spell := NewCard("Test Shock", caster.ID)
	spell.TypeLine = "Instant"
	spell.OracleID = testFlagbearerShock
	caster.Hand.PushTop(spell)
	return g.CastSpell(caster.ID, spell.InstanceID, CastSpellParams{Targets: []TargetRef{{Kind: TargetCard, ID: at}}})
}

// An opponent's spell must target the Flagbearer while it is a legal
// target; once it can't be targeted, the spell is free again.
func TestOpponentsSpellMustTargetAFlagbearerIfAble(t *testing.T) {
	withFlagbearerRule(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bearer := pushFlagbearer(t, g, opp, "Standard Bearer", true)
	bear := pushCombatant(t, g, opp, "Bear", 2, 2)

	if err := castFlagbearerShock(g, me, bear); !errors.Is(err, ErrMustTargetFlagbearer) || !errors.Is(err, ErrIllegalTarget) {
		t.Fatalf("a Shock at the Bear: got %v, want ErrMustTargetFlagbearer", err)
	}
	if err := castFlagbearerShock(g, me, bearer); err != nil {
		t.Fatalf("a Shock at the Flagbearer: %v", err)
	}
	resolveTop(t, g)

	// The Flagbearer gains shroud: no longer a legal target, so the rule
	// asks nothing. Another Flagbearer without the rule still counts as
	// "a Flagbearer", which a second test pins.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bearer {
				g.Battlefield.Cards[i].Keywords = []string{KeywordShroud}
			}
		}
		g.layerVersion.Add(1)
	})
	if err := castFlagbearerShock(g, me, bear); err != nil {
		t.Fatalf("a Shock at the Bear with no targetable Flagbearer: %v", err)
	}
}

// The rule binds only the rule controller's opponents, and any Flagbearer
// on the battlefield answers it, whoever controls it.
func TestFlagbearerRuleBindsOnlyOpponentsAndAnyFlagbearerAnswers(t *testing.T) {
	withFlagbearerRule(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	// My own rule does not bind me.
	pushFlagbearer(t, g, me, "My Standard Bearer", true)
	bear := pushCombatant(t, g, opp, "Bear", 2, 2)
	if err := castFlagbearerShock(g, me, bear); err != nil {
		t.Fatalf("my own rule bound me: %v", err)
	}
	resolveTop(t, g)
	// An opponent's rule binds me; my own Flagbearer (without the rule)
	// satisfies it.
	pushFlagbearer(t, g, opp, "Their Standard Bearer", true)
	mine := pushFlagbearer(t, g, me, "My Plain Flagbearer", false)
	if err := castFlagbearerShock(g, me, bear); !errors.Is(err, ErrMustTargetFlagbearer) {
		t.Fatalf("a Shock at the Bear under an opponent's rule: got %v", err)
	}
	if err := castFlagbearerShock(g, me, mine); err != nil {
		t.Fatalf("a Shock at my own Flagbearer: %v", err)
	}
}
