package testsru

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bro_b13_cards2_test.go — the graveyard grants, the ward and the
// graveyard-entering cards of The Brothers' War batch 13 (#5482).

// --- Mishra, Tamer of Mak Fawa -----------------------------------------------

func TestS82Bro13MishraGivesArtifactCardsInYourGraveyardUnearth(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.ZgGraveyardCard(me, "Plain Rock", "Artifact")
	golem := effectstest.ZgGraveyardCard(me, "Plain Golem", "Artifact Creature — Golem")
	bear := effectstest.ZgGraveyardCard(me, "Plain Bear", "Creature — Bear")
	theirs := effectstest.ZgGraveyardCard(opp, "Their Rock", "Artifact")
	mishra := effectstest.B12Push(g, me.ID, "Mishra, Tamer of Mak Fawa", "Legendary Creature — Human Artificer", s82bro13MishraOracle, 4, 4)
	for _, tc := range []struct {
		id   uuid.UUID
		name string
		want int
	}{{rock, "artifact", 1}, {golem, "artifact creature", 1}, {bear, "nonartifact creature", 0}, {theirs, "opponent's artifact", 0}} {
		rows, _ := effectstest.ZgRows(g, tc.id)
		if len(rows) != tc.want || (tc.want == 1 && !strings.HasPrefix(rows[0].Label, "Unearth {1}{B}{R}")) {
			t.Errorf("%s: rows = %v, want %d Unearth {1}{B}{R}", tc.name, effectstest.ZgLabels(rows), tc.want)
		}
	}
	effectstest.ZgLeaveBattlefield(t, g, mishra)
	if rows, _ := effectstest.ZgRows(g, rock); len(rows) != 0 {
		t.Errorf("after Mishra left, rows = %v, want none", effectstest.ZgLabels(rows))
	}
}

func TestS82Bro13MishraGrantedUnearthReturnsAnArtifactCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.ToMain(t, g)
	rock := effectstest.ZgGraveyardCard(me, "Plain Rock", "Artifact")
	effectstest.B12Push(g, me.ID, "Mishra, Tamer of Mak Fawa", "Legendary Creature — Human Artificer", s82bro13MishraOracle, 4, 4)
	effectstest.FloatMana(t, g, me, "{C}{B}{R}")
	if err := g.ActivateCatalogAbility(me.ID, rock, 0, game.ActivateAbilityParams{
		Ref: "grant:mishra-tamer-of-mak-fawa/unearth:0:0", Strict: true,
	}); err != nil {
		t.Fatalf("activate the granted unearth: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(rock) {
		t.Fatal("the artifact did not return")
	}
	// Unearth only as a sorcery, and only with the full {1}{B}{R}.
	other := effectstest.ZgGraveyardCard(me, "Other Rock", "Artifact")
	effectstest.FloatMana(t, g, me, "{B}{R}")
	if err := g.ActivateCatalogAbility(me.ID, other, 0, game.ActivateAbilityParams{
		Ref: "grant:mishra-tamer-of-mak-fawa/unearth:0:0", Strict: true,
	}); err == nil {
		t.Error("unearth activated without {1}")
	}
}

func TestS82Bro13MishraWardsEveryPermanentYouControlWithASacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mishra := effectstest.B12Push(g, me.ID, "Mishra, Tamer of Mak Fawa", "Legendary Creature — Human Artificer", s82bro13MishraOracle, 4, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	for name, id := range map[string]uuid.UUID{"another creature": bear, "Mishra himself": mishra} {
		if id == mishra {
			continue // Mishra is black-red; Doom Blade can't target him
		}
		t.Run(name, func(t *testing.T) {
			fodder := effectstest.B12Creature(g, opp.ID, "Fodder", "Creature — Bear", 1, 1)
			effectstest.CastAtWardedCreature(t, g, opp, id)
			prompt := effectstest.PassUntilConfirmFor(t, g, opp.ID)
			if prompt == nil {
				t.Fatal("no ward prompt for the spell's controller")
			}
			if err := g.ResolveConfirm(prompt.ID, opp.ID, true); err != nil {
				t.Fatalf("ResolveConfirm: %v", err)
			}
			pick := effectstest.ChooseCardsChoiceFor(g, opp.ID)
			if pick == nil || pick.ChooseMin != 1 || pick.ChooseMax != 1 {
				t.Fatalf("want a one-permanent sacrifice pick: %+v", pick)
			}
			if err := g.ResolveChooseCards(pick.ID, opp.ID, []uuid.UUID{fodder}); err != nil {
				t.Fatalf("ResolveChooseCards: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if g.Battlefield.Contains(fodder) {
				t.Error("the sacrifice was not paid")
			}
			if g.Battlefield.Contains(id) {
				t.Error("the ward was paid, so the removal spell should have resolved")
			}
		})
	}
}

func TestS82Bro13MishraWardCountersTheSpellWhenDeclined(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Mishra, Tamer of Mak Fawa", "Legendary Creature — Human Artificer", s82bro13MishraOracle, 4, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, opp.ID, "Fodder", "Creature — Bear", 1, 1)
	effectstest.CastAtWardedCreature(t, g, opp, bear)
	prompt := effectstest.PassUntilConfirmFor(t, g, opp.ID)
	if prompt == nil {
		t.Fatal("no ward prompt")
	}
	if err := g.ResolveConfirm(prompt.ID, opp.ID, false); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("a declined ward must counter the spell")
	}
}

func TestS82Bro13MishraGivesNoWardToOpposingPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Mishra, Tamer of Mak Fawa", "Legendary Creature — Human Artificer", s82bro13MishraOracle, 4, 4)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastAtWardedCreature(t, g, me, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("an opposing permanent was warded")
	}
}

// --- Wire Surgeons ------------------------------------------------------------

func TestS82Bro13WireSurgeonsGivesArtifactCreatureCardsEncoreAtTheirManaCost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.ToMain(t, g)
	golem := s82bro13GraveCard(me, "Plain Golem", "Artifact Creature — Golem", effectstest.ZgPlainOracle, "{2}{G}", 3, 3)
	rock := s82bro13GraveCard(me, "Plain Rock", "Artifact", effectstest.ZgPlainOracle, "{2}", 0, 0)
	bear := s82bro13GraveCard(me, "Plain Bear", "Creature — Bear", effectstest.ZgPlainOracle, "{1}{G}", 2, 2)
	surgeons := effectstest.B12Push(g, me.ID, "Wire Surgeons", "Creature — Human Artificer", s82bro13SurgeonsOracle, 6, 5)
	if !effectstest.EffectiveAbilitiesContain(t, g, surgeons, "fear") {
		t.Error("no fear")
	}
	for _, tc := range []struct {
		id   uuid.UUID
		name string
		want int
	}{{golem, "artifact creature", 1}, {rock, "plain artifact", 0}, {bear, "plain creature", 0}} {
		rows, _ := effectstest.ZgRows(g, tc.id)
		if len(rows) != tc.want || (tc.want == 1 && !strings.HasPrefix(rows[0].Label, "Encore")) {
			t.Errorf("%s: rows = %v, want %d Encore", tc.name, effectstest.ZgLabels(rows), tc.want)
		}
	}

	// One short of {2}{G} cannot pay; the full mana cost does, and makes a
	// token copy for each of the three opponents.
	effectstest.FloatMana(t, g, me, "{C}{G}")
	if err := g.ActivateCatalogAbility(me.ID, golem, 0, game.ActivateAbilityParams{
		Ref: "grant:wire-surgeons/encore:0:0", Strict: true,
	}); err == nil {
		t.Fatal("encore activated for less than its mana cost")
	}
	effectstest.FloatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, golem, 0, game.ActivateAbilityParams{
		Ref: "grant:wire-surgeons/encore:0:0", Strict: true,
	}); err != nil {
		t.Fatalf("activate the granted encore: %v", err)
	}
	if !g.Exile.Contains(golem) {
		t.Fatal("encore's cost exiles the card")
	}
	effectstest.PassPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Name == "Plain Golem" && c.Controller == me.ID {
			n++
		}
	}
	if n != 3 {
		t.Errorf("%d token copies, want one per opponent (3)", n)
	}
}

// --- Dredging Claw ------------------------------------------------------------

func s82bro13AttachedTo(g *game.Game, id uuid.UUID) uuid.UUID {
	c, ok := effectstest.BattlefieldCard(g, id)
	if !ok {
		return uuid.Nil
	}
	return c.AttachedTo.ID
}

func TestS82Bro13DredgingClawAttachesToACreatureEnteringFromYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	claw := effectstest.B12Push(g, me.ID, "Dredging Claw", "Artifact — Equipment", s82bro13ClawOracle, 0, 0)
	effectstest.ToMain(t, g)
	id := s82bro13GraveCard(me, "Terror Ballista", "Artifact Creature — Construct", s82bro13BallistaOracle, "{7}", 5, 3)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{B}{B}")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("unearth: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("the creature did not unearth")
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := s82bro13AttachedTo(g, claw); got != id {
		t.Fatalf("the Claw is attached to %v, want the returned creature", got)
	}
	if got := effectstest.EffectivePower(t, g, id); got != 6 {
		t.Errorf("equipped power %d, want 6 (+1/+0)", got)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, id, "menace") {
		t.Error("the equipped creature has no menace")
	}
}

func TestS82Bro13DredgingClawMayDeclineAndIgnoresACastCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	claw := effectstest.B12Push(g, me.ID, "Dredging Claw", "Artifact — Equipment", s82bro13ClawOracle, 0, 0)
	effectstest.ToMain(t, g)
	// A creature that enters from the stack offers nothing.
	effectstest.CastFree(t, g, me, "Bear Cub", "Creature — Bear")
	if effectstest.ConfirmForTriggerPrompt(g, me.ID) != nil {
		t.Fatal("a cast creature offered the attach")
	}
	// A returned one may be declined.
	id := s82bro13GraveCard(me, "Terror Ballista", "Artifact Creature — Construct", s82bro13BallistaOracle, "{7}", 5, 3)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{B}{B}")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("unearth: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if got := s82bro13AttachedTo(g, claw); got != uuid.Nil {
		t.Errorf("declined, yet attached to %v", got)
	}
}

func TestS82Bro13DredgingClawIgnoresACreatureFromAnotherGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	claw := effectstest.B12Push(g, me.ID, "Dredging Claw", "Artifact — Equipment", s82bro13ClawOracle, 0, 0)
	effectstest.ToMain(t, g)
	// Reanimated from an opponent's graveyard: not "your graveyard".
	id := s82bro13GraveCard(opp, "Their Bear", "Creature — Bear", "", "{1}{G}", 2, 2)
	g.WithWriteLock(func() {
		if err := g.ReturnFromGraveyardUnderControlForEffect(id, game.ZoneBattlefield, me.ID); err != nil {
			t.Fatalf("reanimate: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ConfirmForTriggerPrompt(g, me.ID) != nil {
		t.Error("a creature from an opponent's graveyard offered the attach")
	}
	if got := s82bro13AttachedTo(g, claw); got != uuid.Nil {
		t.Errorf("attached to %v", got)
	}
}

// --- Thran Vigil --------------------------------------------------------------

func TestS82Bro13ThranVigilCountsArtifactAndCreatureCardsLeavingDuringYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Thran Vigil", "Enchantment", s82bro13VigilOracle, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.ToMain(t, g)
	golem := s82bro13GraveCard(me, "Plain Golem", "Artifact Creature — Golem", effectstest.ZgPlainOracle, "{3}", 3, 3)
	rock := s82bro13GraveCard(me, "Plain Rock", "Artifact", effectstest.ZgPlainOracle, "{2}", 0, 0)
	spell := s82bro13GraveCard(me, "Plain Spell", "Sorcery", effectstest.ZgPlainOracle, "{1}", 0, 0)

	// A sorcery card leaving does nothing.
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(spell) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectiveCounters(g, bear); got != 0 {
		t.Fatalf("a sorcery card leaving gave %d counters", got)
	}
	// Two cards leaving at once are one batch: one counter.
	g.WithWriteLock(func() { g.ExileCardsForEffect([]uuid.UUID{golem, rock}) })
	pickIfOpen := effectstest.LatestPickTarget(g, me.ID)
	if pickIfOpen != nil {
		effectstest.PickCard(t, g, me.ID, bear)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectiveCounters(g, bear); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
}

func TestS82Bro13ThranVigilDoesNothingOnAnOpponentsTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Thran Vigil", "Enchantment", s82bro13VigilOracle, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMainOf(t, g, 1)
	rock := s82bro13GraveCard(me, "Plain Rock", "Artifact", effectstest.ZgPlainOracle, "{2}", 0, 0)
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(rock) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectiveCounters(g, bear); got != 0 {
		t.Errorf("a card leaving on an opponent's turn gave %d counters", got)
	}
}

func effectiveCounters(g *game.Game, id uuid.UUID) int {
	c, ok := effectstest.BattlefieldCard(g, id)
	if !ok {
		return -1
	}
	return c.Counters[game.CounterPlusOne]
}

// --- Meticulous Excavation ----------------------------------------------------

func TestS82Bro13ExcavationBouncesAPermanentWithoutUnearth(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	exc := effectstest.B12Push(g, me.ID, "Meticulous Excavation", "Enchantment", s82bro13ExcavationOracle, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.ToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}{W}")
	if err := g.ActivateCatalogAbility(me.ID, exc, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !me.Hand.Contains(bear) {
		t.Error("the permanent did not return to hand")
	}
}

func TestS82Bro13ExcavationExilesThenReturnsAUnearthPermanentToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	exc := effectstest.B12Push(g, me.ID, "Meticulous Excavation", "Enchantment", s82bro13ExcavationOracle, 0, 0)
	tb := effectstest.B12Push(g, me.ID, "Terror Ballista", "Artifact Creature — Construct", s82bro13BallistaOracle, 5, 3)
	effectstest.ToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}{W}")
	if err := g.ActivateCatalogAbility(me.ID, exc, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(tb), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tb) || !me.Hand.Contains(tb) {
		t.Error("the unearth permanent did not end in its owner's hand")
	}
	var exiled bool
	for _, ev := range g.Events {
		if ev.CardID == tb && ev.NewZone == game.ZoneExile {
			exiled = true
		}
	}
	if !exiled {
		t.Error("the unearth permanent was bounced, not exiled first")
	}
}

func TestS82Bro13ExcavationBouncesAnOpponentsUnearthlessAndIsMyTurnOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	exc := effectstest.B12Push(g, me.ID, "Meticulous Excavation", "Enchantment", s82bro13ExcavationOracle, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.FloatMana(t, g, me, "{C}{C}{W}")
	if err := g.ActivateCatalogAbility(me.ID, exc, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); err == nil {
		t.Error("activated on an opponent's turn")
	}
	// It targets a permanent YOU control.
	effectstest.AdvanceToMainOf(t, g, 0)
	theirs := effectstest.B12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, exc, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(theirs), Strict: true}); err == nil {
		t.Error("targeted a permanent an opponent controls")
	}
}
