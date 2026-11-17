package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bro_b13_cards2_test.go — the graveyard grants, the ward and the
// graveyard-entering cards of The Brothers' War batch 13 (#5482).

// --- Mishra, Tamer of Mak Fawa -----------------------------------------------

func TestS82Bro13MishraGivesArtifactCardsInYourGraveyardUnearth(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := zgGraveyardCard(me, "Plain Rock", "Artifact")
	golem := zgGraveyardCard(me, "Plain Golem", "Artifact Creature — Golem")
	bear := zgGraveyardCard(me, "Plain Bear", "Creature — Bear")
	theirs := zgGraveyardCard(opp, "Their Rock", "Artifact")
	mishra := b12Push(g, me.ID, "Mishra, Tamer of Mak Fawa", "Legendary Creature — Human Artificer", s82bro13MishraOracle, 4, 4)
	for _, tc := range []struct {
		id   uuid.UUID
		name string
		want int
	}{{rock, "artifact", 1}, {golem, "artifact creature", 1}, {bear, "nonartifact creature", 0}, {theirs, "opponent's artifact", 0}} {
		rows, _ := zgRows(g, tc.id)
		if len(rows) != tc.want || (tc.want == 1 && !strings.HasPrefix(rows[0].Label, "Unearth {1}{B}{R}")) {
			t.Errorf("%s: rows = %v, want %d Unearth {1}{B}{R}", tc.name, zgLabels(rows), tc.want)
		}
	}
	zgLeaveBattlefield(t, g, mishra)
	if rows, _ := zgRows(g, rock); len(rows) != 0 {
		t.Errorf("after Mishra left, rows = %v, want none", zgLabels(rows))
	}
}

func TestS82Bro13MishraGrantedUnearthReturnsAnArtifactCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	rock := zgGraveyardCard(me, "Plain Rock", "Artifact")
	b12Push(g, me.ID, "Mishra, Tamer of Mak Fawa", "Legendary Creature — Human Artificer", s82bro13MishraOracle, 4, 4)
	floatMana(t, g, me, "{C}{B}{R}")
	if err := g.ActivateCatalogAbility(me.ID, rock, 0, game.ActivateAbilityParams{
		Ref: "grant:mishra-tamer-of-mak-fawa/unearth:0:0", Strict: true,
	}); err != nil {
		t.Fatalf("activate the granted unearth: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(rock) {
		t.Fatal("the artifact did not return")
	}
	// Unearth only as a sorcery, and only with the full {1}{B}{R}.
	other := zgGraveyardCard(me, "Other Rock", "Artifact")
	floatMana(t, g, me, "{B}{R}")
	if err := g.ActivateCatalogAbility(me.ID, other, 0, game.ActivateAbilityParams{
		Ref: "grant:mishra-tamer-of-mak-fawa/unearth:0:0", Strict: true,
	}); err == nil {
		t.Error("unearth activated without {1}")
	}
}

func TestS82Bro13MishraWardsEveryPermanentYouControlWithASacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	mishra := b12Push(g, me.ID, "Mishra, Tamer of Mak Fawa", "Legendary Creature — Human Artificer", s82bro13MishraOracle, 4, 4)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	for name, id := range map[string]uuid.UUID{"another creature": bear, "Mishra himself": mishra} {
		if id == mishra {
			continue // Mishra is black-red; Doom Blade can't target him
		}
		t.Run(name, func(t *testing.T) {
			fodder := b12Creature(g, opp.ID, "Fodder", "Creature — Bear", 1, 1)
			castAtWardedCreature(t, g, opp, id)
			prompt := passUntilConfirmFor(t, g, opp.ID)
			if prompt == nil {
				t.Fatal("no ward prompt for the spell's controller")
			}
			if err := g.ResolveConfirm(prompt.ID, opp.ID, true); err != nil {
				t.Fatalf("ResolveConfirm: %v", err)
			}
			pick := chooseCardsChoiceFor(g, opp.ID)
			if pick == nil || pick.ChooseMin != 1 || pick.ChooseMax != 1 {
				t.Fatalf("want a one-permanent sacrifice pick: %+v", pick)
			}
			if err := g.ResolveChooseCards(pick.ID, opp.ID, []uuid.UUID{fodder}); err != nil {
				t.Fatalf("ResolveChooseCards: %v", err)
			}
			passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Mishra, Tamer of Mak Fawa", "Legendary Creature — Human Artificer", s82bro13MishraOracle, 4, 4)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b12Creature(g, opp.ID, "Fodder", "Creature — Bear", 1, 1)
	castAtWardedCreature(t, g, opp, bear)
	prompt := passUntilConfirmFor(t, g, opp.ID)
	if prompt == nil {
		t.Fatal("no ward prompt")
	}
	if err := g.ResolveConfirm(prompt.ID, opp.ID, false); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("a declined ward must counter the spell")
	}
}

func TestS82Bro13MishraGivesNoWardToOpposingPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Mishra, Tamer of Mak Fawa", "Legendary Creature — Human Artificer", s82bro13MishraOracle, 4, 4)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castAtWardedCreature(t, g, me, theirs)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("an opposing permanent was warded")
	}
}

// --- Wire Surgeons ------------------------------------------------------------

func TestS82Bro13WireSurgeonsGivesArtifactCreatureCardsEncoreAtTheirManaCost(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	golem := s82bro13GraveCard(me, "Plain Golem", "Artifact Creature — Golem", zgPlainOracle, "{2}{G}", 3, 3)
	rock := s82bro13GraveCard(me, "Plain Rock", "Artifact", zgPlainOracle, "{2}", 0, 0)
	bear := s82bro13GraveCard(me, "Plain Bear", "Creature — Bear", zgPlainOracle, "{1}{G}", 2, 2)
	surgeons := b12Push(g, me.ID, "Wire Surgeons", "Creature — Human Artificer", s82bro13SurgeonsOracle, 6, 5)
	if !effectiveAbilitiesContain(t, g, surgeons, "fear") {
		t.Error("no fear")
	}
	for _, tc := range []struct {
		id   uuid.UUID
		name string
		want int
	}{{golem, "artifact creature", 1}, {rock, "plain artifact", 0}, {bear, "plain creature", 0}} {
		rows, _ := zgRows(g, tc.id)
		if len(rows) != tc.want || (tc.want == 1 && !strings.HasPrefix(rows[0].Label, "Encore")) {
			t.Errorf("%s: rows = %v, want %d Encore", tc.name, zgLabels(rows), tc.want)
		}
	}

	// One short of {2}{G} cannot pay; the full mana cost does, and makes a
	// token copy for each of the three opponents.
	floatMana(t, g, me, "{C}{G}")
	if err := g.ActivateCatalogAbility(me.ID, golem, 0, game.ActivateAbilityParams{
		Ref: "grant:wire-surgeons/encore:0:0", Strict: true,
	}); err == nil {
		t.Fatal("encore activated for less than its mana cost")
	}
	floatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, golem, 0, game.ActivateAbilityParams{
		Ref: "grant:wire-surgeons/encore:0:0", Strict: true,
	}); err != nil {
		t.Fatalf("activate the granted encore: %v", err)
	}
	if !g.Exile.Contains(golem) {
		t.Fatal("encore's cost exiles the card")
	}
	passPriorityAroundTable(t, g)
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
	c, ok := battlefieldCard(g, id)
	if !ok {
		return uuid.Nil
	}
	return c.AttachedTo.ID
}

func TestS82Bro13DredgingClawAttachesToACreatureEnteringFromYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	claw := b12Push(g, me.ID, "Dredging Claw", "Artifact — Equipment", s82bro13ClawOracle, 0, 0)
	toMain(t, g)
	id := s82bro13GraveCard(me, "Terror Ballista", "Artifact Creature — Construct", s82bro13BallistaOracle, "{7}", 5, 3)
	floatMana(t, g, me, "{C}{C}{C}{B}{B}")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("unearth: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("the creature did not unearth")
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := s82bro13AttachedTo(g, claw); got != id {
		t.Fatalf("the Claw is attached to %v, want the returned creature", got)
	}
	if got := effectivePower(t, g, id); got != 6 {
		t.Errorf("equipped power %d, want 6 (+1/+0)", got)
	}
	if !effectiveAbilitiesContain(t, g, id, "menace") {
		t.Error("the equipped creature has no menace")
	}
}

func TestS82Bro13DredgingClawMayDeclineAndIgnoresACastCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	claw := b12Push(g, me.ID, "Dredging Claw", "Artifact — Equipment", s82bro13ClawOracle, 0, 0)
	toMain(t, g)
	// A creature that enters from the stack offers nothing.
	castFree(t, g, me, "Bear Cub", "Creature — Bear")
	if confirmForTriggerPrompt(g, me.ID) != nil {
		t.Fatal("a cast creature offered the attach")
	}
	// A returned one may be declined.
	id := s82bro13GraveCard(me, "Terror Ballista", "Artifact Creature — Construct", s82bro13BallistaOracle, "{7}", 5, 3)
	floatMana(t, g, me, "{C}{C}{C}{B}{B}")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("unearth: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if got := s82bro13AttachedTo(g, claw); got != uuid.Nil {
		t.Errorf("declined, yet attached to %v", got)
	}
}

func TestS82Bro13DredgingClawIgnoresACreatureFromAnotherGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	claw := b12Push(g, me.ID, "Dredging Claw", "Artifact — Equipment", s82bro13ClawOracle, 0, 0)
	toMain(t, g)
	// Reanimated from an opponent's graveyard: not "your graveyard".
	id := s82bro13GraveCard(opp, "Their Bear", "Creature — Bear", "", "{1}{G}", 2, 2)
	g.WithWriteLock(func() {
		if err := g.ReturnFromGraveyardUnderControlForEffect(id, game.ZoneBattlefield, me.ID); err != nil {
			t.Fatalf("reanimate: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if confirmForTriggerPrompt(g, me.ID) != nil {
		t.Error("a creature from an opponent's graveyard offered the attach")
	}
	if got := s82bro13AttachedTo(g, claw); got != uuid.Nil {
		t.Errorf("attached to %v", got)
	}
}

// --- Thran Vigil --------------------------------------------------------------

func TestS82Bro13ThranVigilCountsArtifactAndCreatureCardsLeavingDuringYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Thran Vigil", "Enchantment", s82bro13VigilOracle, 0, 0)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	toMain(t, g)
	golem := s82bro13GraveCard(me, "Plain Golem", "Artifact Creature — Golem", zgPlainOracle, "{3}", 3, 3)
	rock := s82bro13GraveCard(me, "Plain Rock", "Artifact", zgPlainOracle, "{2}", 0, 0)
	spell := s82bro13GraveCard(me, "Plain Spell", "Sorcery", zgPlainOracle, "{1}", 0, 0)

	// A sorcery card leaving does nothing.
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(spell) })
	passPriorityAroundTable(t, g)
	if got := effectiveCounters(g, bear); got != 0 {
		t.Fatalf("a sorcery card leaving gave %d counters", got)
	}
	// Two cards leaving at once are one batch: one counter.
	g.WithWriteLock(func() { g.ExileCardsForEffect([]uuid.UUID{golem, rock}) })
	pickIfOpen := latestPickTarget(g, me.ID)
	if pickIfOpen != nil {
		pickCard(t, g, me.ID, bear)
	}
	passPriorityAroundTable(t, g)
	if got := effectiveCounters(g, bear); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
}

func TestS82Bro13ThranVigilDoesNothingOnAnOpponentsTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Thran Vigil", "Enchantment", s82bro13VigilOracle, 0, 0)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMainOf(t, g, 1)
	rock := s82bro13GraveCard(me, "Plain Rock", "Artifact", zgPlainOracle, "{2}", 0, 0)
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(rock) })
	passPriorityAroundTable(t, g)
	if got := effectiveCounters(g, bear); got != 0 {
		t.Errorf("a card leaving on an opponent's turn gave %d counters", got)
	}
}

func effectiveCounters(g *game.Game, id uuid.UUID) int {
	c, ok := battlefieldCard(g, id)
	if !ok {
		return -1
	}
	return c.Counters[game.CounterPlusOne]
}

// --- Meticulous Excavation ----------------------------------------------------

func TestS82Bro13ExcavationBouncesAPermanentWithoutUnearth(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	exc := b12Push(g, me.ID, "Meticulous Excavation", "Enchantment", s82bro13ExcavationOracle, 0, 0)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	toMain(t, g)
	floatMana(t, g, me, "{C}{C}{W}")
	if err := g.ActivateCatalogAbility(me.ID, exc, 0, game.ActivateAbilityParams{Targets: cardRefs(bear), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !me.Hand.Contains(bear) {
		t.Error("the permanent did not return to hand")
	}
}

func TestS82Bro13ExcavationExilesThenReturnsAUnearthPermanentToHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	exc := b12Push(g, me.ID, "Meticulous Excavation", "Enchantment", s82bro13ExcavationOracle, 0, 0)
	tb := b12Push(g, me.ID, "Terror Ballista", "Artifact Creature — Construct", s82bro13BallistaOracle, 5, 3)
	toMain(t, g)
	floatMana(t, g, me, "{C}{C}{W}")
	if err := g.ActivateCatalogAbility(me.ID, exc, 0, game.ActivateAbilityParams{Targets: cardRefs(tb), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	exc := b12Push(g, me.ID, "Meticulous Excavation", "Enchantment", s82bro13ExcavationOracle, 0, 0)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMainOf(t, g, 1)
	floatMana(t, g, me, "{C}{C}{W}")
	if err := g.ActivateCatalogAbility(me.ID, exc, 0, game.ActivateAbilityParams{Targets: cardRefs(bear), Strict: true}); err == nil {
		t.Error("activated on an opponent's turn")
	}
	// It targets a permanent YOU control.
	advanceToMainOf(t, g, 0)
	theirs := b12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, exc, 0, game.ActivateAbilityParams{Targets: cardRefs(theirs), Strict: true}); err == nil {
		t.Error("targeted a permanent an opponent controls")
	}
}
