package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pinned_retarget_test.go — #1743's card proofs: Spellskite, Mizzium
// Meddler and Hydroelectric Specimen, each "change a target of target
// spell or ability TO THIS CREATURE". The engine half (which slots are
// eligible, the which-one prompt) is game/retarget_pinned_test.go; this
// is the cards doing it on a real board.

const (
	spellskiteOracle     = "e0aa6ce0-ca31-433b-ac6c-32b8675cdb71"
	mizziumMeddlerOracle = "48a909b6-e6ee-4148-8b50-b35f11bc065f"
)

// pushSpellskite puts a 0/4 Spellskite on the battlefield under
// `owner`, through the zone-move event so the layer engine sees it.
func pushSpellskite(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Spellskite", TypeLine: "Artifact Creature — Phyrexian Horror",
		ManaCost: "{2}", OracleID: spellskiteOracle, Power: 0, Toughness: 4,
		Owner: owner, Controller: owner,
	})
}

// activateSpellskite announces the ability at `item`, paying {U/P}
// with two life.
func activateSpellskite(t *testing.T, g *game.Game, controller, skite, item uuid.UUID) {
	t.Helper()
	if err := g.ActivateCatalogAbility(controller, skite, 0, game.ActivateAbilityParams{
		Targets:       []game.TargetRef{{Kind: game.TargetCard, ID: item}},
		PhyrexianLife: 1,
	}); err != nil {
		t.Fatalf("activate Spellskite: %v", err)
	}
}

func boltAt(t *testing.T, g *game.Game, target game.TargetRef) uuid.UUID {
	t.Helper()
	return castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, []game.TargetRef{target})
}

// The card in one test: an opponent's Bolt at its controller's face is
// pulled onto Spellskite for 2 life, and the 0/4 takes the 3.
func TestSpellskiteRedirectsABoltToItself(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	skite := pushSpellskite(g, opp.ID)
	bolt := boltAt(t, g, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})

	life := opp.Life
	activateSpellskite(t, g, opp.ID, skite, bolt)
	if got := life - opp.Life; got != game.PhyrexianLifePerSymbol {
		t.Errorf("life paid = %d, want %d for {U/P}", got, game.PhyrexianLifePerSymbol)
	}
	passPriorityAroundTable(t, g)

	if latestRetarget(g, opp.ID) != nil {
		t.Error("one target and one place for it to go: nothing to ask")
	}
	if got := lifeOf(g, opp.ID); got != life-game.PhyrexianLifePerSymbol {
		t.Errorf("the Bolt still hit the player: life %d", got)
	}
	if got := damageMarkedOn(g, skite); got != 3 {
		t.Errorf("Spellskite has %d damage, want 3", got)
	}
}

// {U/P} paid with {U} costs no life.
func TestSpellskitePaysWithBlueMana(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	skite := pushSpellskite(g, opp.ID)
	bolt := boltAt(t, g, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	fillPoolColored(opp, "U", 1)
	life := opp.Life
	if err := g.ActivateCatalogAbility(opp.ID, skite, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bolt}},
	}); err != nil {
		t.Fatalf("activate with {U}: %v", err)
	}
	if opp.Life != life || len(opp.ManaPool) != 0 {
		t.Errorf("paying {U}: life %d -> %d, pool %d, want the blue spent and no life", life, opp.Life, len(opp.ManaPool))
	}
	passPriorityAroundTable(t, g)
	if got := damageMarkedOn(g, skite); got != 3 {
		t.Errorf("Spellskite has %d damage, want 3", got)
	}
}

// Ruling: Spellskite may target a spell or ability it is not a legal
// target for, and then nothing changes. "Target player" can never
// become a creature.
func TestSpellskiteDoesNothingWhereItIsNotALegalTarget(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	skite := pushSpellskite(g, opp.ID)
	shredder := b12Push(g, me.ID, "Codex Shredder", "Artifact", b22CodexShredderOracle, 0, 0)
	if err := g.ActivateCatalogAbility(me.ID, shredder, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("activate Codex Shredder: %v", err)
	}
	mill := acItemLabelled(g, "{T}: Target player mills a card.")
	if mill == uuid.Nil {
		t.Fatal("the mill is not on the stack")
	}
	library := opp.Library.Size()
	activateSpellskite(t, g, opp.ID, skite, mill)
	passPriorityAroundTable(t, g)
	if latestRetarget(g, opp.ID) != nil {
		t.Error("a prompt opened for a target Spellskite can never be")
	}
	if opp.Library.Size() != library-1 {
		t.Errorf("the mill should still have hit its original target: %d -> %d", library, opp.Library.Size())
	}
}

// "Target spell or ability" reaches an ABILITY on the stack: Goblin
// Bombardment's ping aimed at a creature is pulled onto Spellskite.
func TestSpellskiteRedirectsAnActivatedAbility(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	skite := pushSpellskite(g, opp.ID)
	bomb := pushCatalogPermanent(g, me.ID, "Goblin Bombardment", "Enchantment", goblinBombardmentOracle, false)
	fodder := pushCatalogPermanent(g, me.ID, "Fodder", "Creature — Goblin", "", false)
	victim := pushCatalogPermanent(g, opp.ID, "Victim", "Creature — Bear", "", false)
	if err := g.ActivateCatalogAbility(me.ID, bomb, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{fodder},
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}); err != nil {
		t.Fatalf("activate Goblin Bombardment: %v", err)
	}
	var ping uuid.UUID
	g.WithWriteLock(func() {
		for id, it := range g.StackMeta {
			if it != nil && it.Kind == game.StackItemActivated && it.SourceCardID == bomb {
				ping = id
			}
		}
	})
	if ping == uuid.Nil {
		t.Fatal("the ping is not on the stack")
	}
	activateSpellskite(t, g, opp.ID, skite, ping)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(victim) {
		t.Error("the 1/1 died: the ping was not redirected")
	}
	if got := damageMarkedOn(g, skite); got != 1 {
		t.Errorf("Spellskite has %d damage, want 1", got)
	}
}

// Arc Trail has two instances of "target": Spellskite's controller
// chooses which ONE becomes Spellskite, as the ability resolves.
// Taking the 2-damage half off the player leaves the creature its 1.
func TestSpellskiteChoosesWhichTargetOfATwoTargetSpell(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	skite := pushSpellskite(g, opp.ID)
	big := pushCostedPermanentForTest(g, opp.ID, "Big", "Creature — Beast", "{2}{G}")
	arc := castCatalogSpell(t, g, "Arc Trail", "Sorcery", arcTrailOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetCard, ID: big}})

	activateSpellskite(t, g, opp.ID, skite, arc)
	life := opp.Life
	passPriorityAroundTable(t, g)
	prompt := latestRetarget(g, opp.ID)
	if prompt == nil {
		t.Fatal("two targets that could each become Spellskite, and no choice offered")
	}
	if !hasID(prompt.PickTargetPlayers, opp.ID) || !hasID(prompt.PickTargetCards, big) {
		t.Fatalf("options = players %v cards %v, want the player and the creature",
			prompt.PickTargetPlayers, prompt.PickTargetCards)
	}
	if hasID(prompt.PickTargetCards, skite) {
		t.Error("Spellskite itself is offered — the answer is WHICH target, never where")
	}
	if err := g.ResolveRetarget(prompt.ID, opp.ID,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err != nil {
		t.Fatalf("ResolveRetarget: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Errorf("the player still took Arc Trail's 2: life %d -> %d", life, opp.Life)
	}
	if got := damageMarkedOn(g, skite); got != 2 {
		t.Errorf("Spellskite has %d damage, want the 2", got)
	}
	if got := damageMarkedOn(g, big); got != 1 {
		t.Errorf("the creature has %d damage, want its 1 — the other target is untouched", got)
	}
}

// Ruling: "If Spellskite leaves the battlefield before its ability
// resolves …, no targets are changed." That includes leaving and
// coming straight back: the card on the battlefield is then a new
// object (CR 400.7) with the same instance ID, which every target
// clause would accept — so this case is the source check's, not the
// gate's.
func TestSpellskiteThatLeftChangesNothing(t *testing.T) {
	cases := map[string]func(t *testing.T, g *game.Game, skite uuid.UUID){
		"destroyed": func(t *testing.T, g *game.Game, skite uuid.UUID) {
			g.WithWriteLock(func() {
				if err := g.DestroyPermanentForEffect(skite); err != nil {
					t.Fatalf("destroy Spellskite: %v", err)
				}
			})
		},
		"left and came back": func(t *testing.T, g *game.Game, skite uuid.UUID) {
			flickerInResponse(t, g, skite)
		},
	}
	for name, leave := range cases {
		t.Run(name, func(t *testing.T) {
			g := newCatalogGame(t)
			opp := g.Seats[1]
			skite := pushSpellskite(g, opp.ID)
			bolt := boltAt(t, g, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
			activateSpellskite(t, g, opp.ID, skite, bolt)
			leave(t, g, skite)
			life := opp.Life
			passPriorityAroundTable(t, g)
			if got := g.StackMeta[bolt]; got != nil {
				t.Fatalf("the Bolt never resolved: %+v", got)
			}
			if opp.Life != life-3 {
				t.Errorf("the Bolt should have hit its original target: life %d -> %d", life, opp.Life)
			}
		})
	}
}

// The redirected spell leaves first (countered above the ability):
// the ability resolves and does nothing, with no error.
func TestSpellskiteWhoseSpellIsGoneDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	skite := pushSpellskite(g, opp.ID)
	bolt := boltAt(t, g, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	activateSpellskite(t, g, opp.ID, skite, bolt)
	g.WithWriteLock(func() {
		if err := g.CounterTargetForEffect(bolt); err != nil {
			t.Fatalf("counter the Bolt: %v", err)
		}
	})
	before := len(g.Events)
	passPriorityAroundTable(t, g)
	for _, ev := range g.Events[before:] {
		if ev.Kind == game.EventEffectError {
			t.Errorf("the ability errored on a spell that is gone: %+v", ev)
		}
	}
	if got := damageMarkedOn(g, skite); got != 0 {
		t.Errorf("Spellskite has %d damage, want none", got)
	}
}

// Mizzium Meddler: flashed in, its trigger targets the Bolt, and "you
// may" is asked as it resolves — with the Bolt's one target as the
// only option. Saying yes pulls the Bolt onto the 1/4.
func TestMizziumMeddlerMayRedirectToItself(t *testing.T) {
	for _, accept := range []bool{true, false} {
		name := "declined"
		if accept {
			name = "accepted"
		}
		t.Run(name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			bolt := boltAt(t, g, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID})
			meddler := castCatalogSpell(t, g, "Mizzium Meddler", "Creature — Vedalken Wizard", mizziumMeddlerOracle, nil)
			g.WithWriteLock(func() {
				for i := range g.Stack.Cards {
					if g.Stack.Cards[i].InstanceID == meddler {
						g.Stack.Cards[i].Power, g.Stack.Cards[i].Toughness = 1, 4
					}
				}
			})
			for i := 0; i < 4 && latestPickTarget(g, me.ID) == nil && latestRetarget(g, me.ID) == nil; i++ {
				if err := g.PassPriority(); err != nil {
					break
				}
			}
			if latestPickTarget(g, me.ID) != nil {
				pickCard(t, g, me.ID, bolt)
			}
			passPriorityAroundTable(t, g)
			prompt := latestRetarget(g, me.ID)
			if prompt == nil {
				t.Fatal("Mizzium Meddler's trigger asked nothing")
			}
			if prompt.PickTargetMin != 0 || !hasID(prompt.PickTargetPlayers, me.ID) {
				t.Fatalf(`prompt = min %d over players %v, want a "you may" over the Bolt's target`,
					prompt.PickTargetMin, prompt.PickTargetPlayers)
			}
			var answer []game.TargetRef
			if accept {
				answer = []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}
			}
			if err := g.ResolveRetarget(prompt.ID, me.ID, answer); err != nil {
				t.Fatalf("ResolveRetarget: %v", err)
			}
			life := me.Life
			passPriorityAroundTable(t, g)
			if accept {
				if me.Life != life || damageMarkedOn(g, meddler) != 3 {
					t.Errorf("accepted: life %d -> %d, Meddler damage %d; want the Bolt on the Meddler",
						life, me.Life, damageMarkedOn(g, meddler))
				}
			} else if me.Life != life-3 {
				t.Errorf("declined: life %d -> %d, want the Bolt to land where it was aimed", life, me.Life)
			}
		})
	}
}

// Hydroelectric Specimen: the same, with a clause that only admits an
// instant or sorcery spell with a single target.
func TestHydroelectricSpecimenRedirectsASingleTargetInstant(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bolt := boltAt(t, g, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID})
	specimen := castCatalogSpell(t, g, "Hydroelectric Specimen", "Creature — Weird", hydroelectricSpecimenOracle, nil)
	g.WithWriteLock(func() {
		for i := range g.Stack.Cards {
			if g.Stack.Cards[i].InstanceID == specimen {
				g.Stack.Cards[i].Power, g.Stack.Cards[i].Toughness = 1, 4
			}
		}
	})
	for i := 0; i < 4 && latestPickTarget(g, me.ID) == nil && latestRetarget(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	if latestPickTarget(g, me.ID) != nil {
		pickCard(t, g, me.ID, bolt)
	}
	passPriorityAroundTable(t, g)
	prompt := latestRetarget(g, me.ID)
	if prompt == nil {
		t.Fatal("Hydroelectric Specimen's trigger asked nothing")
	}
	if err := g.ResolveRetarget(prompt.ID, me.ID,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}); err != nil {
		t.Fatalf("ResolveRetarget: %v", err)
	}
	life := me.Life
	passPriorityAroundTable(t, g)
	if me.Life != life || damageMarkedOn(g, specimen) != 3 {
		t.Errorf("life %d -> %d, Specimen damage %d; want the Bolt on the Specimen",
			life, me.Life, damageMarkedOn(g, specimen))
	}

	// The clause: a two-target spell is not a legal target at all.
	spec, _ := Lookup(hydroelectricSpecimenOracle)
	if len(spec.Triggered) != 1 || spec.Triggered[0].Targets == nil {
		t.Fatal("the enters trigger declares no target clause")
	}
	label := spec.Triggered[0].Targets.Label
	if label != "target instant or sorcery spell with a single target" {
		t.Errorf("clause label = %q", label)
	}
}
