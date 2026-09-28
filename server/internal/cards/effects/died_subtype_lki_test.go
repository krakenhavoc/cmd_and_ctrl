package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// died_subtype_lki_test.go — #1679, CR 603.10a: a tribal dies-trigger
// is judged on the creature's subtypes as it last existed on the
// battlefield. In the graveyard no grant applies any more — a Bear
// that was a Zombie under Maskwood Nexus is a Bear — so every tribal
// dies condition reads the subtypes the exit stamped on the EventLTB
// (leftAsSubtype), never the graveyard card.

// The issue's case: under Maskwood Nexus a non-Zombie creature dies
// with Diregraf Captain out, and the drain triggers — though the card
// in the graveyard is a plain Bear.
func TestDiregrafCaptainDrainsForANonZombieUnderMaskwoodNexus(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b21Push(g, me.ID, "Diregraf Captain", "Creature — Zombie Soldier", b21DiregrafCaptainOracle, 2, 2, "U", "B")
	bear := b16Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2, "G")
	pushNexusFor(g, me.ID)
	if !layeredCard(t, g, bear).HasSubtype("Zombie") {
		t.Fatal("setup: the Bear is not a Zombie under Maskwood Nexus")
	}
	oppBefore := opp.Life

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	if c, _ := g.LookupCardForEffect(bear); c.HasSubtype("Zombie") {
		t.Fatal("setup: the Bear in the graveyard should not be a Zombie — the point of the test")
	}
	b04WaitForPick(t, g, me.ID)
	if latestPickTarget(g, me.ID) == nil {
		t.Fatal("Diregraf Captain did not trigger on a creature that died a Zombie")
	}
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != oppBefore-1 {
		t.Errorf("opponent life %d → %d, want -1", oppBefore, opp.Life)
	}
}

// The other direction: a Zombie that an effect made an Elk before it
// died (Kenrith's Transformation sets the subtypes) was not a Zombie
// when it died, though the card in the graveyard reads as one.
func TestDiregrafCaptainIgnoresAZombieThatStoppedBeingOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b21Push(g, me.ID, "Diregraf Captain", "Creature — Zombie Soldier", b21DiregrafCaptainOracle, 2, 2, "U", "B")
	zombie := b16Creature(g, me.ID, "My Zombie", "Creature — Zombie", 2, 2, "B")
	enchant(t, g, "Kenrith's Transformation", kenrithTransformOracle, zombie)
	if layeredCard(t, g, zombie).HasSubtype("Zombie") {
		t.Fatal("setup: the enchanted Zombie should be an Elk and nothing else")
	}
	lifeBefore := lifeSnapshot(g)

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(zombie) })
	if c, _ := g.LookupCardForEffect(zombie); !c.HasSubtype("Zombie") {
		t.Fatal("setup: the graveyard card should read as the Zombie it prints — the point of the test")
	}
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("Diregraf Captain triggered on an Elk")
	}
	for i, p := range g.Seats {
		if p.Life != lifeBefore[i] {
			t.Errorf("seat %d life %d → %d; an Elk dying drains nobody", i, lifeBefore[i], p.Life)
		}
	}
}

// tribalDiesCheck is one switched tribal dies condition and the tribe
// it asks about.
type tribalDiesCheck struct {
	name  string
	tribe string
	check func(ev game.Event, source *game.Card, g *game.Game) bool
}

// tribalDiesChecks is every tribal dies / leaves-the-battlefield
// condition #1679 switched to leftAsSubtype. A new one belongs here.
func tribalDiesChecks(t *testing.T) []tribalDiesCheck {
	omnath, ok := Lookup("1816eede-c5bd-49df-958f-a3af64cb2932")
	if !ok || len(omnath.Triggered) < 2 {
		t.Fatal("Omnath, Locus of Rage is not registered with its dies trigger")
	}
	omnathDies := omnath.Triggered[1].AppliesTo
	var none game.Characteristic
	return []tribalDiesCheck{
		{"b17SelfOrZombieYouControlDied (Undead Augur)", "Zombie", b17SelfOrZombieYouControlDied},
		{"b27SelfOrNontokenZombieYouControlDied (Headless Rider)", "Zombie", b27SelfOrNontokenZombieYouControlDied},
		{"anotherZombieYouControlDied (Diregraf Captain, Plague Belcher)", "Zombie", anotherZombieYouControlDied},
		{"b25AnotherGoblinYouControlDied (Pashalik Mons)", "Goblin", b25AnotherGoblinYouControlDied},
		{"b34VampireYouControlDied (Crossway Troublemakers)", "Vampire", b34VampireYouControlDied},
		{"b36AngelYouControlDied (Bishop of Wings)", "Angel", b36AngelYouControlDied},
		{"b36AnotherFaerieYouControlDied (Tegwyll)", "Faerie", b36AnotherFaerieYouControlDied},
		{"anEggYouControlDied (Atla Palani)", "Egg", func(ev game.Event, source *game.Card, g *game.Game) bool {
			return anEggYouControlDied(ev, source, none, g)
		}},
		{"Omnath, Locus of Rage's dies trigger", "Elemental", func(ev game.Event, source *game.Card, g *game.Game) bool {
			return omnathDies(ev, source, none, g)
		}},
	}
}

// Every switched check against three deaths whose subtypes on the
// battlefield differ from — or, for the changeling, only match through
// the keyword — the card left behind:
//
//   - granted: a Bear given the tribe by a layer-4 grant (a lord's type
//     grant) — a plain Bear in the graveyard; counts;
//   - lost: a printed member of the tribe that Kenrith's
//     Transformation made an Elk — reads as the tribe in the graveyard;
//     does not count;
//   - changeling: a printed changeling (CR 702.73a) — every creature
//     type, stamped as the flag; counts for every tribe.
func TestTribalDiesChecksReadLastKnownSubtypes(t *testing.T) {
	for _, c := range tribalDiesChecks(t) {
		t.Run(c.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			source := &game.Card{InstanceID: uuid.New(), Controller: me.ID, Owner: me.ID}

			granted := b16Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2, "G")
			g.WithWriteLock(func() {
				if !g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(granted),
					[]game.Mod{game.AddSubtypesMod(c.tribe)},
					g.UntilEndOfTurnDuration(), "test — the Bear gains the tribe") {
					t.Fatal("setup: the grant registered nothing")
				}
				g.RecomputeLayersIfStaleLocked()
			})
			lost := b16Creature(g, me.ID, "My "+c.tribe, "Creature — "+c.tribe, 2, 2, "B")
			enchant(t, g, "Kenrith's Transformation", kenrithTransformOracle, lost)
			changeling := pushBattlefieldCardWithTimestamp(g, game.Card{
				InstanceID: uuid.New(), Name: "My Changeling", TypeLine: "Creature — Shapeshifter",
				Keywords: []string{game.KeywordChangeling}, Power: 1, Toughness: 1,
				Owner: me.ID, Controller: me.ID,
			})
			if !layeredCard(t, g, granted).HasSubtype(c.tribe) || layeredCard(t, g, lost).HasSubtype(c.tribe) {
				t.Fatal("setup: the grant or the Elk set did not take")
			}

			died := map[string]game.Event{}
			for label, id := range map[string]uuid.UUID{"granted": granted, "lost": lost, "changeling": changeling} {
				g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id) })
				died[label] = lastLTBOf(t, g, id)
			}
			if gy, _ := g.LookupCardForEffect(granted); gy.HasSubtype(c.tribe) {
				t.Fatal("setup: the granted Bear still has the tribe in the graveyard")
			}
			if gy, _ := g.LookupCardForEffect(lost); !gy.HasSubtype(c.tribe) {
				t.Fatal("setup: the Elk's graveyard card should read as the printed tribe")
			}

			for label, want := range map[string]bool{"granted": true, "lost": false, "changeling": true} {
				if got := c.check(died[label], source, g); got != want {
					t.Errorf("%s died: %s = %v, want %v", label, c.name, got, want)
				}
			}
		})
	}
}
