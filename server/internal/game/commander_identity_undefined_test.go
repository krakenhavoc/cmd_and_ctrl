package game

import (
	"reflect"
	"testing"

	"github.com/google/uuid"
)

// CR 903.4f: "If an ability refers to the colors or number of colors in
// a commander's color identity, that quality is undefined if that
// player doesn't have a commander. That part of the ability won't do
// anything." The official rulings on Command Tower, Arcane Signet,
// Commander's Sphere and Path of Ancestry say the same, and a
// colourless commander (Kozilek, Karn) leaves the same nothing to add.
//
// Issue #844: the engine used to hand the printed five colours back
// instead, so Command Tower was stronger than printed for a
// commanderless or colourless deck — and it could not tell a real
// colourless identity from a placeholder commander with no colour data
// at all. commanderIdentity's tri-state is that distinction.

// setImportedCommanderForTest stamps the seat's commander with the
// shape deck.ToGameCard produces for a REAL card: a Scryfall printing
// behind it, and Scryfall's own color_identity — empty included, which
// is Kozilek's actual identity.
func setImportedCommanderForTest(t *testing.T, p *Player, name, cost string, identity []string) {
	t.Helper()
	for i := range p.Command.Cards {
		if !p.Command.Cards[i].IsCommander {
			continue
		}
		c := &p.Command.Cards[i]
		c.Name = name
		c.ScryfallID, c.OracleID = uuid.NewString(), uuid.NewString()
		c.ManaCost, c.Colors, c.ColorIdentity = cost, nil, identity
		return
	}
	t.Fatalf("no commander in %s's command zone", p.Name)
}

// removeCommanderForTest empties the command zone: the deck's only
// IsCommander card starts there, so the seat then owns no commander in
// any zone.
func removeCommanderForTest(t *testing.T, p *Player) {
	t.Helper()
	found := false
	for _, c := range p.Command.Cards {
		found = found || c.IsCommander
	}
	if !found {
		t.Fatalf("no commander in %s's command zone to remove", p.Name)
	}
	p.Command.Cards = nil
}

func TestCommanderIdentityTriState(t *testing.T) {
	t.Run("no commander is undefined", func(t *testing.T) {
		g := newActiveGame(t)
		p := g.Seats[0]
		removeCommanderForTest(t, p)
		got := commanderIdentityFor(g, p)
		if got.State != identityNoCommander || len(got.Colors) != 0 {
			t.Errorf("identity = %+v, want no commander and no colours", got)
		}
	})

	t.Run("an imported colourless commander is known and empty", func(t *testing.T) {
		g := newActiveGame(t)
		p := g.Seats[0]
		setImportedCommanderForTest(t, p, "Kozilek, Butcher of Truth", "{10}", nil)
		got := commanderIdentityFor(g, p)
		if got.State != identityKnown || len(got.Colors) != 0 {
			t.Errorf("identity = %+v, want known and colourless", got)
		}
	})

	t.Run("an imported coloured commander is known", func(t *testing.T) {
		g := newActiveGame(t)
		p := g.Seats[0]
		setImportedCommanderForTest(t, p, "Test Commander", "{1}{W}{U}", []string{"W", "U"})
		got := commanderIdentityFor(g, p)
		if got.State != identityKnown || !reflect.DeepEqual(got.Colors, []string{"W", "U"}) {
			t.Errorf("identity = %+v, want known [W U]", got)
		}
	})

	t.Run("a placeholder commander is unknown, not colourless", func(t *testing.T) {
		g := newActiveGame(t)
		p := g.Seats[0]
		// The shared test deck's commander: a name and nothing else,
		// like the demo seed's. No Scryfall record stands behind it,
		// so its empty colour list says "nobody stamped one".
		got := commanderIdentityFor(g, p)
		if got.State != identityUnknown || len(got.Colors) != 0 {
			t.Errorf("identity = %+v, want unknown and no colours", got)
		}
	})

	t.Run("a partner with colour data keeps the pair known", func(t *testing.T) {
		g := newActiveGame(t)
		p := g.Seats[0]
		setImportedCommanderForTest(t, p, "Kozilek, Butcher of Truth", "{10}", nil)
		partner := NewCard("Second Partner", p.ID)
		partner.TypeLine = "Legendary Creature — Human"
		partner.ManaCost = "{2}{U}"
		partner.IsCommander = true
		g.WithWriteLock(func() { g.Battlefield.PushTop(partner) })
		got := commanderIdentityFor(g, p)
		if got.State != identityKnown || !reflect.DeepEqual(got.Colors, []string{"U"}) {
			t.Errorf("identity = %+v, want known [U] (the union, CR 903.4a)", got)
		}
	})
}

// The whole of #844 through Command Tower: with no identity to narrow
// against, the ability adds no mana, prompts for nothing, and nothing
// offers it.
func TestIdentityManaAddsNothingWithoutAnIdentity(t *testing.T) {
	withCatalogHook(t, commandTowerHook)

	for _, tc := range []struct {
		name  string
		seat  func(t *testing.T, p *Player)
		about string
	}{
		{
			name:  "no commander",
			seat:  func(t *testing.T, p *Player) { removeCommanderForTest(t, p) },
			about: "CR 903.4f: the quality is undefined",
		},
		{
			name: "colourless commander",
			seat: func(t *testing.T, p *Player) {
				setImportedCommanderForTest(t, p, "Kozilek, Butcher of Truth", "{10}", nil)
			},
			about: "a real identity that names no colours",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			setup := func(t *testing.T) (*Game, *Player, uuid.UUID) {
				g := newActiveGame(t)
				advanceTo(t, g, StepPrecombatMain)
				p := g.Seats[0]
				tc.seat(t, p)
				tower := pushBattlefieldForTest(g, p.ID, "Command Tower", "Land", commandTowerOracleID)
				return g, p, tower
			}

			t.Run("activation adds nothing and prompts for nothing", func(t *testing.T) {
				g, p, tower := setup(t)
				if err := g.ActivateManaAbility(p.ID, tower, 0, ManaAbilityParams{}); err != nil {
					t.Fatalf("activate Command Tower: %v", err)
				}
				if c := choiceByKind(g, PendingChoiceMana); c != nil {
					t.Errorf("queued a mana pick (%v) for a source that adds nothing — %s", c.ColorOptions, tc.about)
				}
				if len(p.ManaPool) != 0 {
					t.Errorf("pool = %+v, want empty", p.ManaPool)
				}
			})

			t.Run("the ability is not offered", func(t *testing.T) {
				g, p, tower := setup(t)
				ab := ManaAbilitiesForCard(*findBattlefieldCard(g, tower))[0]
				if !ManaAbilityAddsNoMana(g, p.ID, tower, ab) {
					t.Error("ManaAbilityAddsNoMana = false, want true")
				}
			})

			t.Run("the planner skips the source", func(t *testing.T) {
				g, p, _ := setup(t)
				var sources []tapSource
				g.WithWriteLock(func() { sources = gatherTapSources(g, p.ID, nil) })
				if len(sources) != 0 {
					t.Errorf("planned sources = %+v, want none", sources)
				}
			})

			t.Run("the executor leaves it untapped", func(t *testing.T) {
				g, p, tower := setup(t)
				g.WithWriteLock(func() {
					g.materializePlanLocked(p, []uuid.UUID{tower}, costFor(t, "{1}"))
				})
				if findBattlefieldCard(g, tower).Tapped {
					t.Error("the auto-tapper tapped a source that adds no mana")
				}
				if len(p.ManaPool) != 0 {
					t.Errorf("pool = %+v, want empty", p.ManaPool)
				}
			})

			t.Run("a cast cannot be paid from it", func(t *testing.T) {
				g, p, tower := setup(t)
				id := pushTypedCardToHandWithCost(p, "Sol Ring", "Artifact", "{1}")
				if err := g.CastSpell(p.ID, id, CastSpellParams{Strict: true, AutoTap: true}); err == nil {
					t.Error("auto-tapped {1} out of a source that adds no mana")
				}
				if findBattlefieldCard(g, tower).Tapped {
					t.Error("the failed cast tapped the Tower anyway")
				}
			})
		})
	}
}

// The two identities that still narrow to something, so the fix stops
// where the rule does.
func TestIdentityManaStillNarrowsWhenItCan(t *testing.T) {
	withCatalogHook(t, commandTowerHook)

	t.Run("an imported two-colour commander narrows to its colours", func(t *testing.T) {
		g := newActiveGame(t)
		advanceTo(t, g, StepPrecombatMain)
		p := g.Seats[0]
		setImportedCommanderForTest(t, p, "Test Commander", "{1}{W}{U}", []string{"W", "U"})
		tower := pushBattlefieldForTest(g, p.ID, "Command Tower", "Land", commandTowerOracleID)
		if err := g.ActivateManaAbility(p.ID, tower, 0, ManaAbilityParams{}); err != nil {
			t.Fatalf("activate Command Tower: %v", err)
		}
		c := choiceByKind(g, PendingChoiceMana)
		if c == nil || !reflect.DeepEqual(c.ColorOptions, []string{"W", "U"}) {
			t.Fatalf("Command Tower options = %+v, want [W U] in printed order", c)
		}
	})

	// The documented data-gap fallback: a commander with no colour data
	// anywhere is a broken import, not a colourless commander, and a
	// data gap must not switch a printed card off.
	t.Run("a placeholder commander keeps the printed five", func(t *testing.T) {
		g := newActiveGame(t)
		advanceTo(t, g, StepPrecombatMain)
		p := g.Seats[0]
		tower := pushBattlefieldForTest(g, p.ID, "Command Tower", "Land", commandTowerOracleID)
		if err := g.ActivateManaAbility(p.ID, tower, 0, ManaAbilityParams{}); err != nil {
			t.Fatalf("activate Command Tower: %v", err)
		}
		c := choiceByKind(g, PendingChoiceMana)
		if c == nil || !reflect.DeepEqual(c.ColorOptions, []string{"W", "U", "B", "R", "G"}) {
			t.Fatalf("Command Tower options = %+v, want the printed five", c)
		}
		ab := ManaAbilitiesForCard(*findBattlefieldCard(g, tower))[0]
		if ManaAbilityAddsNoMana(g, p.ID, tower, ab) {
			t.Error("ManaAbilityAddsNoMana = true for a source that still offers five colours")
		}
	})

	// An "any color" source is untouched by any of this: it prints its
	// own colours and CR 903.4f has nothing to say about it.
	t.Run("Birds still offers five without a commander", func(t *testing.T) {
		withCatalogHook(t, birdsHook)
		g := newActiveGame(t)
		advanceTo(t, g, StepPrecombatMain)
		p := g.Seats[0]
		removeCommanderForTest(t, p)
		birds := pushBattlefieldForTest(g, p.ID, "Birds of Paradise", "Creature — Bird", birdsOracleID)
		if err := g.ActivateManaAbility(p.ID, birds, 0, ManaAbilityParams{}); err != nil {
			t.Fatalf("activate Birds: %v", err)
		}
		c := choiceByKind(g, PendingChoiceMana)
		if c == nil || !reflect.DeepEqual(c.ColorOptions, []string{"W", "U", "B", "R", "G"}) {
			t.Fatalf("Birds options = %+v, want the printed five", c)
		}
	})
}
