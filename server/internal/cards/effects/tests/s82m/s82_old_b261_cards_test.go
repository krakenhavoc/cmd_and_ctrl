package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b261_cards_test.go — Kaladesh, batch 261 (#5539). Every test plays
// the real catalog entry through the engine.

const (
	kldB261WildestDreams = "0a25a72e-5603-4b50-ac20-36b286cc1a2b"
	kldB261Gearsmith     = "447c7fe0-6ec7-4163-9d9c-f08f0a4c7109"
	kldB261Find          = "8cd3815d-92a6-4336-9da7-b18b1afd5460"
	kldB261Daredevil     = "a33961bb-fff0-4f4a-aa35-f904dbf2b075"
	kldB261Scavenger     = "b9906b19-912e-4a9d-888f-3c895bbed2c5"
	kldB261Scrounger     = "d309e1e9-26e6-4a10-a3f4-352c7ba3f2ff"
	kldB261Bear          = "Creature — Bear"
)

func TestKldB261EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		kldB261WildestDreams: "Wildest Dreams",
		kldB261Gearsmith:     "Restoration Gearsmith",
		kldB261Find:          "Fortuitous Find",
		kldB261Daredevil:     "Ovalchase Daredevil",
		kldB261Scavenger:     "Dukhara Scavenger",
		kldB261Scrounger:     "Scrapheap Scrounger",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Wildest Dreams ------------------------------------------------------------------

func TestKldB261WildestDreamsReturnsXCardsAndExilesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.PushGraveyardPermanent(me, "A", kldB261Bear, "{1}{G}")
	b := effectstest.PushGraveyardPermanent(me, "B", "Artifact", "{1}")
	c := effectstest.PushGraveyardPermanent(me, "C", kldB261Bear, "{2}")
	spell := effectstest.CastXSpell(t, g, "Wildest Dreams", "Sorcery", kldB261WildestDreams, "{X}{X}{G}", 2, effectstest.CardRefs(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(a) || !me.Hand.Contains(b) {
		t.Error("a targeted card did not return to hand")
	}
	if !me.Graveyard.Contains(c) {
		t.Error("an untargeted card left the graveyard")
	}
	if !g.Exile.Contains(spell) || me.Graveyard.Contains(spell) {
		t.Error("Wildest Dreams was not exiled")
	}
}

func TestKldB261WildestDreamsPaysXTwiceAndTargetsExactlyX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.PushGraveyardPermanent(me, "A", kldB261Bear, "{1}{G}")
	b := effectstest.PushGraveyardPermanent(me, "B", kldB261Bear, "{1}{G}")
	effectstest.AdvanceToMain(t, g)
	cast := func(mana string, x int, targets ...uuid.UUID) error {
		effectstest.FloatMana(t, g, me, mana)
		_, err := effectstest.CastFromHandErr(g, me, "Wildest Dreams", "Sorcery", "{X}{X}{G}", kldB261WildestDreams,
			game.CastSpellParams{Strict: true, XValue: x, Targets: effectstest.CardRefs(targets...)})
		g.WithWriteLock(func() { me.ManaPool = game.ManaPool{} })
		return err
	}
	if err := cast("{G}{C}{C}", 2, a, b); err == nil {
		t.Error("X = 2 was paid with {X} once")
	}
	if err := cast("{G}{C}{C}{C}{C}", 2, a); err == nil {
		t.Error("one target was accepted for X = 2")
	}
	if err := cast("{G}{C}{C}{C}{C}", 1, a, b); err == nil {
		t.Error("two targets were accepted for X = 1")
	}
	if err := cast("{G}{C}{C}{C}{C}", 2, a, b); err != nil {
		t.Fatalf("X = 2 with two cards and {4}{G} was refused: %v", err)
	}
}

func TestKldB261WildestDreamsWithXZeroStillExilesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	spell := effectstest.CastXSpell(t, g, "Wildest Dreams", "Sorcery", kldB261WildestDreams, "{X}{X}{G}", 0, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(spell) || me.Graveyard.Contains(spell) {
		t.Error("Wildest Dreams with X = 0 was not exiled")
	}
}

// --- Restoration Gearsmith ---------------------------------------------------------------

func TestKldB261GearsmithReturnsAnArtifactOrCreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	creature := effectstest.GraveyardCardOf(g, me, "Dead Bear", kldB261Bear, "{2}", "")
	artifact := effectstest.GraveyardCardOf(g, me, "Dead Rock", "Artifact", "{1}", "")
	spell := effectstest.GraveyardCardOf(g, me, "Dead Bolt", "Instant", "{R}", "")
	effectstest.CastHoldCreature(t, g, "Restoration Gearsmith", "Creature — Human Artificer", kldB261Gearsmith, 3, 3)
	effectstest.B04WaitForPick(t, g, me.ID)
	if p := effectstest.LatestPickTarget(g, me.ID); p == nil || effectstest.HasID(p.PickTargetCards, spell) ||
		!effectstest.HasID(p.PickTargetCards, creature) || !effectstest.HasID(p.PickTargetCards, artifact) {
		t.Fatalf("the picker offers %+v, want the artifact and the creature but not the instant", p)
	}
	effectstest.PickCard(t, g, me.ID, artifact)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(artifact) {
		t.Error("the artifact card was not returned to hand")
	}
	if me.Hand.Contains(creature) || me.Hand.Contains(spell) {
		t.Error("an untargeted card was returned")
	}
}

// --- Dukhara Scavenger ---------------------------------------------------------------------

func TestKldB261ScavengerMayTopdeckAnArtifactOrCreatureCard(t *testing.T) {
	for _, accept := range []bool{true, false} {
		name := "accepted"
		if !accept {
			name = "declined"
		}
		t.Run(name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.SeedLibrary(me, "x", "y")
			creature := effectstest.GraveyardCardOf(g, me, "Dead Bear", kldB261Bear, "{2}", "")
			land := effectstest.GraveyardCardOf(g, me, "Dead Land", "Land", "", "")
			effectstest.CastHoldCreature(t, g, "Dukhara Scavenger", "Creature — Crocodile", kldB261Scavenger, 4, 6)
			effectstest.B04WaitForPick(t, g, me.ID)
			if p := effectstest.LatestPickTarget(g, me.ID); p != nil && effectstest.HasID(p.PickTargetCards, land) {
				t.Error("a land card is offered")
			}
			effectstest.PickCard(t, g, me.ID, creature)
			effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, accept)
			effectstest.PassPriorityAroundTable(t, g)
			if got := me.Library.Contains(creature); got != accept {
				t.Errorf("creature card in library = %v, want %v", got, accept)
			}
			if accept {
				if top := me.Library.Cards[len(me.Library.Cards)-1]; top.InstanceID != creature {
					t.Errorf("the top card is %q, want the put-back creature", top.Name)
				}
				if me.Graveyard.Contains(creature) {
					t.Error("the card is still in the graveyard")
				}
			}
		})
	}
}

// --- Fortuitous Find -------------------------------------------------------------------------

func TestKldB261FortuitousFindOneOrBoth(t *testing.T) {
	for _, c := range []struct {
		name         string
		modes        []int
		wantArtifact bool
		wantCreature bool
	}{
		{"artifact only", []int{0}, true, false},
		{"creature only", []int{1}, false, true},
		{"both", []int{0, 1}, true, true},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			artifact := effectstest.PushGraveyardPermanent(me, "Dead Rock", "Artifact", "{1}")
			creature := effectstest.PushGraveyardPermanent(me, "Dead Bear", kldB261Bear, "{2}")
			var targets []game.TargetRef
			if c.wantArtifact {
				targets = append(targets, effectstest.CardRefs(artifact)...)
			}
			if c.wantCreature {
				targets = append(targets, effectstest.CardRefs(creature)...)
			}
			effectstest.CastCatalogSpellWithModesTargets(t, g, kldB261Find, c.modes, targets)
			effectstest.PassPriorityAroundTable(t, g)
			if got := me.Hand.Contains(artifact); got != c.wantArtifact {
				t.Errorf("artifact in hand = %v, want %v", got, c.wantArtifact)
			}
			if got := me.Hand.Contains(creature); got != c.wantCreature {
				t.Errorf("creature in hand = %v, want %v", got, c.wantCreature)
			}
		})
	}
}

func TestKldB261FortuitousFindRefusesTheWrongTypeForABullet(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.PushGraveyardPermanent(me, "Dead Bear", kldB261Bear, "{2}")
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Fortuitous Find", TypeLine: "Sorcery", ManaCost: "{2}{B}",
			OracleID: kldB261Find, Owner: me.ID, Controller: me.ID})
	})
	effectstest.FloatMana(t, g, me, "{B}{C}{C}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Modes: []int{0}, Targets: effectstest.CardRefs(bear)}); err == nil {
		t.Error("a nonartifact creature card was accepted as the artifact bullet's target")
	}
}

// --- Ovalchase Daredevil ----------------------------------------------------------------------

func TestKldB261DaredevilReturnsWhenAnArtifactEnters(t *testing.T) {
	for _, accept := range []bool{true, false} {
		name := "accepted"
		if !accept {
			name = "declined"
		}
		t.Run(name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			dare := effectstest.GraveyardCardOf(g, me, "Ovalchase Daredevil", "Creature — Human Pilot", "{3}{B}", kldB261Daredevil)
			effectstest.CastHoldCreature(t, g, "Artifact Bear", "Artifact Creature — Bear", "", 2, 2)
			effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, accept)
			effectstest.PassPriorityAroundTable(t, g)
			if got := me.Hand.Contains(dare); got != accept {
				t.Errorf("Daredevil in hand = %v, want %v", got, accept)
			}
		})
	}
}

func TestKldB261DaredevilIgnoresNonartifactsAndOpponentsArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.Top100Opponent(g)
	dare := effectstest.GraveyardCardOf(g, me, "Ovalchase Daredevil", "Creature — Human Pilot", "{3}{B}", kldB261Daredevil)
	effectstest.CastHoldCreature(t, g, "Plain Bear", kldB261Bear, "", 2, 2)
	effectstest.B12Push(g, opp.ID, "Their Rock", "Artifact", "", 0, 0)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestConfirmFor(g, me.ID) != nil || me.Hand.Contains(dare) {
		t.Error("the Daredevil triggered off a nonartifact or an opponent's artifact")
	}
}

// --- Scrapheap Scrounger ------------------------------------------------------------------------

func TestKldB261ScroungerCantBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.B12Push(g, me.ID, "Scrapheap Scrounger", "Artifact Creature — Construct", kldB261Scrounger, 3, 2)
	if effectstest.AuraRestrictions(t, g, id)&game.CantBlock == 0 {
		t.Error("the Scrounger can block")
	}
}

func TestKldB261ScroungerReturnsFromTheGraveyardForAnotherCreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	scrounger := effectstest.GraveyardCardOf(g, me, "Scrapheap Scrounger", "Artifact Creature — Construct", "{2}", kldB261Scrounger)
	rock := effectstest.GraveyardCardOf(g, me, "Dead Rock", "Artifact", "{1}", "")
	bear := effectstest.GraveyardCardOf(g, me, "Dead Bear", kldB261Bear, "{2}", "")
	effectstest.FloatMana(t, g, me, "{B}{C}")
	if err := g.ActivateCatalogAbility(me.ID, scrounger, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{rock}}); err == nil {
		t.Fatal("a noncreature card paid for the exile cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, scrounger, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{scrounger}}); err == nil {
		t.Fatal("the Scrounger paid for its own cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, scrounger, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with nothing exiled")
	}
	if err := g.ActivateCatalogAbility(me.ID, scrounger, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !g.Exile.Contains(bear) {
		t.Error("the creature card was not exiled as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(scrounger) {
		t.Fatal("the Scrounger did not return to the battlefield")
	}
	if !me.Graveyard.Contains(rock) {
		t.Error("the rock left the graveyard")
	}
}

func TestKldB261ScroungerNeedsItsManaAndCannotActivateFromTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	scrounger := effectstest.GraveyardCardOf(g, me, "Scrapheap Scrounger", "Artifact Creature — Construct", "{2}", kldB261Scrounger)
	bear := effectstest.GraveyardCardOf(g, me, "Dead Bear", kldB261Bear, "{2}", "")
	if err := g.ActivateCatalogAbility(me.ID, scrounger, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{bear}}); err == nil {
		t.Error("activated with an empty pool")
	}
	live := effectstest.B12Push(g, me.ID, "Scrapheap Scrounger", "Artifact Creature — Construct", kldB261Scrounger, 3, 2)
	effectstest.FloatMana(t, g, me, "{B}{C}")
	if err := g.ActivateCatalogAbility(me.ID, live, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{bear}}); err == nil {
		t.Error("the graveyard ability was activated from the battlefield")
	}
}
