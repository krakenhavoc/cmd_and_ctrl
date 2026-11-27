package testss82m

import (
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b557_cards_test.go — Lorwyn cards, S82 OLD batch 557, played through
// the real catalog entries.

const (
	oldB557EntanglingTrap = "229b1e23-8e91-46e8-b0c3-1c5b2c9411da"
	oldB557Favor          = "f2ad82e6-4669-4d87-bea3-f2ac6ebe5ba9"
	oldB557NeckSnap       = "b8ad4276-759d-4d22-a47e-bd27edb34bff"
	oldB557Loggers        = "f8fac28d-adf4-47e9-bdd3-8402cbf0f228"
	oldB557Stalactite     = "1c169307-71f5-4679-8b5a-d60c1109b9db"
	oldB557Trickery       = "bbe370ba-412b-4052-89d9-2d0ae4928118"
	oldB557Whelm          = "7f527852-bf7d-43a3-be0b-784c73ed88aa"
)

func TestOldB557CardsAreFullWithNoCaveats(t *testing.T) {
	for id, name := range map[string]string{
		oldB557EntanglingTrap: "Entangling Trap", oldB557Favor: "Favor of the Mighty",
		oldB557NeckSnap: "Neck Snap", oldB557Loggers: "Boggart Loggers",
		oldB557Stalactite: "Runed Stalactite", oldB557Trickery: "Faerie Trickery",
	} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v as %q (%v, caveats %v)", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB557NeckSnapDestroysOnlyAnAttackerOrBlocker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := effectstest.B12Creature(g, me.ID, "Big", "Creature — Bear", 7, 7)
	idle := effectstest.B12Creature(g, opp.ID, "Idle", "Creature — Bear", 1, 1)
	effectstest.DeclareAttack(t, g, opp.ID, big)
	cast := func(target uuid.UUID, strict bool) error {
		id := effectstest.HandCardOf(g, me, "Neck Snap", "Instant", "", oldB557NeckSnap)
		return g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.CardRefs(target), Strict: strict})
	}
	if err := cast(idle, true); err == nil {
		t.Error("targeted a creature that is neither attacking nor blocking")
	}
	if err := cast(big, true); err != nil {
		t.Fatalf("CastSpell on an attacker: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("the attacker survived")
	}
}

func TestOldB557BoggartLoggersSacrificesToDestroyATreefolkOrForest(t *testing.T) {
	for _, tc := range []struct{ name, typeLine string }{
		{"Treefolk", "Creature — Treefolk"},
		{"Forest", "Basic Land — Forest"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			loggers := effectstest.B21Push(g, me.ID, "Boggart Loggers", "Creature — Goblin Rogue", oldB557Loggers, 2, 1, "B")
			bear := effectstest.B16Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2, "G")
			victim := effectstest.B12Permanent(g, opp.ID, "Victim", tc.typeLine)
			effectstest.AdvanceToMain(t, g)
			effectstest.B06AddMana(me, "B", "B", "B")
			if err := g.ActivateCatalogAbility(me.ID, loggers, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(bear)}); err == nil {
				t.Fatal("a Bear is not a legal target")
			}
			effectstest.B16Activate(t, g, me.ID, loggers, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(victim)})
			if g.Battlefield.Contains(victim) {
				t.Error("the target survived")
			}
			if g.Battlefield.Contains(loggers) || !me.Graveyard.Contains(loggers) {
				t.Error("the Loggers are not sacrificed as part of the cost")
			}
		})
	}
}

func TestOldB557RunedStalactitePumpsAndMakesTheHostEveryType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	host := effectstest.B12Creature(g, me.ID, "Grizzly", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Cub", "Creature — Bear", 2, 2)
	eq := effectstest.PushCatalogPermanent(g, me.ID, "Runed Stalactite", "Artifact — Equipment", oldB557Stalactite, false)
	effectstest.BrAttach(t, g, eq, host)
	if p, tough := effectstest.EffectivePower(t, g, host), effectstest.EffectiveToughness(t, g, host); p != 3 || tough != 3 {
		t.Errorf("equipped creature %d/%d, want 3/3", p, tough)
	}
	if !effectstest.LayeredCard(t, g, host).Effective().AllCreatureTypes {
		t.Error("equipped creature is not every creature type")
	}
	if effectstest.LayeredCard(t, g, other).Effective().AllCreatureTypes || effectstest.EffectivePower(t, g, other) != 2 {
		t.Error("an unequipped creature changed")
	}
}

func TestOldB557FaerieTrickeryCountersAndExilesANonFaerieSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := effectstest.Batch01OpponentCasts(t, g, opp, "Fireball", "", "{2}{R}", []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	effectstest.CastFromSeat(t, g, me, "Faerie Trickery", "Kindred Instant — Faerie", oldB557Trickery, []game.TargetRef{{Kind: game.TargetCard, ID: bolt}})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.P7Countered(g, bolt) {
		t.Fatal("the spell was not countered")
	}
	if zone, _ := effectstest.CardWhere(g, opp, bolt); zone != "exile" {
		t.Errorf("the countered spell is in %s, want exile", zone)
	}
}

func TestOldB557FaerieTrickeryCannotTargetAFaerieSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	faerie := effectstest.Batch01OpponentCasts(t, g, opp, "Spiteful Sprite", "", "{U}", nil)
	g.WithWriteLock(func() {
		for i := range g.Stack.Cards {
			if g.Stack.Cards[i].InstanceID == faerie {
				g.Stack.Cards[i].TypeLine = "Kindred Instant — Faerie"
			}
		}
	})
	id := effectstest.HandCardOf(g, me, "Faerie Trickery", "Kindred Instant — Faerie", "", oldB557Trickery)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: faerie}}, Strict: true}); err == nil {
		t.Error("Faerie Trickery targeted a Faerie spell")
	}
}

func oldB557Pushed(g *game.Game, owner uuid.UUID, name, cost string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Bear", ManaCost: cost,
		Power: 2, Toughness: 2, Owner: owner, Controller: owner,
	})
}

func TestOldB557FavorGivesProtectionToEveryCreatureWithTheGreatestManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	small := oldB557Pushed(g, me.ID, "Small", "{1}")
	bigA := oldB557Pushed(g, me.ID, "Big A", "{4}")
	bigB := oldB557Pushed(g, opp.ID, "Big B", "{4}")
	effectstest.PushCatalogPermanent(g, me.ID, "Favor of the Mighty", "Kindred Enchantment — Giant", oldB557Favor, false)
	red := game.ProtectionFromColor("R")
	if effectstest.HasKeywordOnBattlefield(t, g, small, red) {
		t.Error("a smaller creature has protection")
	}
	for _, id := range []uuid.UUID{bigA, bigB} {
		for _, col := range game.AllColors {
			if !effectstest.HasKeywordOnBattlefield(t, g, id, game.ProtectionFromColor(col)) {
				t.Errorf("a creature tied for the greatest mana value lacks protection from %s", col)
			}
		}
	}
	oldB557Pushed(g, opp.ID, "Huge", "{7}")
	if effectstest.HasKeywordOnBattlefield(t, g, bigA, red) {
		t.Error("protection stayed after a greater creature arrived")
	}
}

func TestOldB557EntanglingTrapTapsOnAnyClashAndFreezesOnAWin(t *testing.T) {
	for _, tc := range []struct {
		name         string
		mine, theirs string
		frozen       bool
	}{
		{"won", "{3}{U}{U}", "{U}", true},
		{"lost", "{U}", "{3}{U}{U}", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.AdvanceToMain(t, g)
			effectstest.PushCatalogPermanent(g, me.ID, "Entangling Trap", "Enchantment", oldB557EntanglingTrap, false)
			bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
			// Whirlpool Whelm's own target, so the Trap's target is never bounced.
			mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
			mineTop := game.NewCard("My Top", me.ID)
			mineTop.TypeLine, mineTop.ManaCost = "Sorcery", tc.mine
			me.Library.PushTop(mineTop)
			theirTop := game.NewCard("Their Top", opp.ID)
			theirTop.TypeLine, theirTop.ManaCost = "Sorcery", tc.theirs
			opp.Library.PushTop(theirTop)
			effectstest.CastCatalogSpell(t, g, "Whirlpool Whelm", "Instant", oldB557Whelm, effectstest.CardTargets(mine))
			effectstest.PassPriorityAroundTable(t, g)
			for i := 0; i < 8; i++ {
				var open *game.PendingChoice
				for _, c := range g.PendingChoices {
					if c != nil && c.Kind == game.PendingChoiceOptionPick {
						open = c
						break
					}
				}
				if open == nil {
					break
				}
				pick := 0
				for j, o := range open.PickOptions {
					if o.Player == opp.ID {
						pick = j
					}
				}
				if err := g.ResolveOptionPick(open.ID, open.Chooser, pick); err != nil {
					t.Fatalf("answer the clash question: %v", err)
				}
			}
			pick := game.TargetRef{Kind: game.TargetCard, ID: bear}
			effectstest.AssistCardSettle(t, g, me.ID, &pick, true)
			if tc.frozen {
				// Won: Whirlpool Whelm offers the tuck; decline it.
				effectstest.AnswerMayChoice(t, g, me.ID, false)
			}
			effectstest.PassPriorityAroundTable(t, g)
			c, ok := effectstest.BattlefieldCard(g, bear)
			if !ok {
				t.Fatal("their creature left the battlefield")
			}
			if !c.Tapped {
				t.Error("the creature was not tapped")
			}
			if got := len(c.NextUntapSkips) > 0; got != tc.frozen {
				t.Errorf("untap skips = %v, want frozen=%v", c.NextUntapSkips, tc.frozen)
			}
		})
	}
}
