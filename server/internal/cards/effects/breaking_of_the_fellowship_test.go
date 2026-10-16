package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// breaking_of_the_fellowship_test.go — #1943, the same-controller target
// relation: "another target creature that player controls".

const breakingOfTheFellowshipOracle = "96e081d2-ec72-41fc-ad81-2e31ab823ae0"

func fellowshipDamage(g *game.Game, id uuid.UUID) int {
	n := -1
	g.ReadSnapshot(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				n = g.Battlefield.Cards[i].DamageMarked
			}
		}
	})
	return n
}

// fellowshipTable is the caster with one creature of their own (so the
// tempt never asks), an opponent with a 3/3 and a 4/4, and a second
// opponent with a creature of their own.
func fellowshipTable(t *testing.T) (g *game.Game, me, opp, other *game.Player, giant, ogre, theirs uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me, opp, other = g.Seats[0], g.Seats[1], g.Seats[2]
	b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	giant = b12Creature(g, opp.ID, "Hill Giant", "Creature — Giant", 3, 3)
	ogre = b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	theirs = b12Creature(g, other.ID, "Elf", "Creature — Elf", 1, 1)
	return
}

func fellowshipTargets(first, second uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: first}, {Kind: game.TargetCard, ID: second, Slot: 1}}
}

func TestBreakingOfTheFellowshipBitesAndTempts(t *testing.T) {
	g, me, _, _, giant, ogre, _ := fellowshipTable(t)
	ringCastSettled(t, g, "Breaking of the Fellowship", "Sorcery", breakingOfTheFellowshipOracle, fellowshipTargets(giant, ogre)...)
	if got := fellowshipDamage(g, ogre); got != 3 {
		t.Fatalf("the Ogre has %d damage, want the Giant's 3", got)
	}
	if got := fellowshipDamage(g, giant); got != 0 {
		t.Fatalf("the Giant took %d damage; the bite is one-sided", got)
	}
	if ringCount(g, me.ID) != 1 {
		t.Fatal("the Ring did not tempt")
	}
}

// CR 601.2c: the second creature must be controlled by the first's
// controller, and must be another creature.
func TestBreakingOfTheFellowshipRefusesAnUnrelatedSecondTarget(t *testing.T) {
	for name, second := range map[string]func(giant, ogre, theirs, mine uuid.UUID) uuid.UUID{
		"another opponent's creature": func(_, _, theirs, _ uuid.UUID) uuid.UUID { return theirs },
		"the same creature":           func(giant, _, _, _ uuid.UUID) uuid.UUID { return giant },
		"the caster's creature":       func(_, _, _, mine uuid.UUID) uuid.UUID { return mine },
	} {
		t.Run(name, func(t *testing.T) {
			g, me, _, _, giant, ogre, theirs := fellowshipTable(t)
			mine := g.Battlefield.Cards[0].InstanceID
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].Controller == me.ID && g.Battlefield.Cards[i].Name == "Grizzly Bears" {
					mine = g.Battlefield.Cards[i].InstanceID
				}
			}
			err := castCatalogSpellErr(t, g, "Breaking of the Fellowship", "Sorcery", breakingOfTheFellowshipOracle,
				fellowshipTargets(giant, second(giant, ogre, theirs, mine)))
			if err == nil {
				t.Fatal("an unrelated second target was accepted")
			}
			if name == "another opponent's creature" && !errors.Is(err, game.ErrIllegalTarget) {
				t.Fatalf("want ErrIllegalTarget, got %v", err)
			}
		})
	}
}

// CR 608.2b: the relationship is checked again on resolution. Each case
// changes the board after the cast; the column says whether the Ogre is
// dealt damage and whether the spell resolved (so the Ring tempts).
func TestBreakingOfTheFellowshipRechecksTheRelationOnResolution(t *testing.T) {
	cases := []struct {
		name    string
		respond func(t *testing.T, g *game.Game, giant, ogre uuid.UUID, other *game.Player)
		damage  int
		tempts  bool
	}{
		{
			// The Giant still has an opponent's controller, but no
			// longer the Ogre's: only the Giant is a legal target, and
			// an illegal target is never dealt damage.
			name: "the first creature changes controller",
			respond: func(t *testing.T, g *game.Game, giant, _ uuid.UUID, other *game.Player) {
				if !g.GainControlForEffect(giant, giant, other.ID, game.IndefiniteDuration(), "test") {
					t.Fatal("control change refused")
				}
			},
			damage: 0, tempts: true,
		},
		{
			name: "the second creature changes controller",
			respond: func(t *testing.T, g *game.Game, _, ogre uuid.UUID, other *game.Player) {
				if !g.GainControlForEffect(ogre, ogre, other.ID, game.IndefiniteDuration(), "test") {
					t.Fatal("control change refused")
				}
			},
			damage: 0, tempts: true,
		},
		{
			// CR 608.2h: the Giant has left, so "that player" is its
			// last-known controller. The Ogre stays legal, the spell
			// resolves, and the Ring tempts, though a creature that
			// is not there deals no damage.
			name: "the first creature dies (last-known controller)",
			respond: func(t *testing.T, g *game.Game, giant, _ uuid.UUID, _ *game.Player) {
				if err := g.DestroyPermanentForEffect(giant); err != nil {
					t.Fatal(err)
				}
			},
			damage: 0, tempts: true,
		},
		{
			// The last-known controller is the Giant's opponent, not
			// whoever now controls the Ogre: both targets are illegal
			// and the spell does not resolve (no tempt).
			name: "the first creature dies and the second changes controller",
			respond: func(t *testing.T, g *game.Game, giant, ogre uuid.UUID, other *game.Player) {
				if err := g.DestroyPermanentForEffect(giant); err != nil {
					t.Fatal(err)
				}
				if !g.GainControlForEffect(ogre, ogre, other.ID, game.IndefiniteDuration(), "test") {
					t.Fatal("control change refused")
				}
			},
			damage: 0, tempts: false,
		},
		{
			// Still on the battlefield, so its current controller
			// decides, and that is unchanged.
			name: "the first creature gains hexproof",
			respond: func(t *testing.T, g *game.Game, giant, _ uuid.UUID, _ *game.Player) {
				if err := (ScopedEffectFor{Target: giant, Mods: []game.Mod{game.AddKeywordsMod("hexproof")}, Duration: game.IndefiniteDuration()}).Apply(NewContext(g, nil)); err != nil {
					t.Fatal(err)
				}
			},
			damage: 0, tempts: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g, me, _, other, giant, ogre, _ := fellowshipTable(t)
			castCatalogSpell(t, g, "Breaking of the Fellowship", "Sorcery", breakingOfTheFellowshipOracle, fellowshipTargets(giant, ogre))
			g.WithWriteLock(func() { tc.respond(t, g, giant, ogre, other) })
			ringSettle(t, g)
			if got := fellowshipDamage(g, ogre); got != tc.damage {
				t.Fatalf("the Ogre has %d damage, want %d", got, tc.damage)
			}
			if got := ringCount(g, me.ID) == 1; got != tc.tempts {
				t.Fatalf("the Ring tempted: %v, want %v", got, tc.tempts)
			}
		})
	}
}
