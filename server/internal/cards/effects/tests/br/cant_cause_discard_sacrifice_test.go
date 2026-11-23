package testsbr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func pushTajuru(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.PushCatalogPermanent(g, owner, "Tajuru Preserver", "Creature — Elf Shaman", effectstest.TajuruPreserverOracle, false)
}

// An opponent's edict skips the protected player and still asks the
// other opponents (and its own controller, for "each player").
func TestOpponentsEdictSkipsProtectedPlayerOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, shielded, other := g.Seats[0], g.Seats[1], g.Seats[2]
	pushTajuru(g, shielded.ID)
	effectstest.PushCatalogPermanent(g, shielded.ID, "Shielded Bear", "Creature — Bear", "", false)
	otherBear := effectstest.PushCatalogPermanent(g, other.ID, "Other Bear", "Creature — Bear", "", false)
	mine := effectstest.PushCatalogPermanent(g, me.ID, "My Bear", "Creature — Bear", "", false)

	effectstest.CastCatalogSpell(t, g, "Fleshbag Marauder", "Creature — Zombie Warrior", effectstest.FleshbagMarauderOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if c := effectstest.SacrificeChoiceFor(g, shielded.ID); c != nil {
		t.Fatalf("the protected player was asked to sacrifice: %+v", c)
	}
	// The controller (Bear and Marauder) chooses; the other opponent's
	// only creature is a forced pick (ADR 0175, #3664).
	if effectstest.SacrificeChoiceFor(g, me.ID) == nil {
		t.Fatal("the controller is still asked")
	}
	effectstest.AnswerSacrifice(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(otherBear) || g.Battlefield.Contains(mine) {
		t.Error("the other opponent and the controller still sacrifice")
	}
	if n := effectstest.AnnCountOn(g, shielded.ID); n != 2 {
		t.Errorf("the protected player still controls %d permanents, want 2", n)
	}
}

// Your own edict still bites you: the clause covers OPPONENTS' spells.
func TestOwnEdictStillMakesYouSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	pushTajuru(g, me.ID)
	effectstest.CastCatalogSpell(t, g, "Fleshbag Marauder", "Creature — Zombie Warrior", effectstest.FleshbagMarauderOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SacrificeChoiceFor(g, me.ID) == nil {
		t.Fatal("Fleshbag's controller is protected from their OWN spell and ability")
	}
}

// The protection leaving the battlefield ends it.
func TestProtectionEndsWhenItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	shielded := g.Seats[1]
	preserver := pushTajuru(g, shielded.ID)
	bear := effectstest.PushCatalogPermanent(g, shielded.ID, "Shielded Bear", "Creature — Bear", "", false)
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(preserver) })
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.CastCatalogSpell(t, g, "Fleshbag Marauder", "Creature — Zombie Warrior", effectstest.FleshbagMarauderOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// Their only creature is a forced pick (ADR 0175, #3664).
	if g.Battlefield.Contains(bear) {
		t.Fatal("with the Preserver gone the player must sacrifice again")
	}
}

// A sacrifice paid as a cost is the player's own choice and is never
// blocked, even while an opponent's spell is the last thing that
// resolved.
func TestSacrificeCostIsNeverBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	shielded := g.Seats[1]
	pushTajuru(g, shielded.ID)
	bomb := effectstest.PushCatalogPermanent(g, shielded.ID, "Goblin Bombardment", "Enchantment", effectstest.GoblinBombardmentOracle, false)
	fodder := effectstest.PushCatalogPermanent(g, shielded.ID, "Fodder", "Creature — Goblin", "", false)

	// An opponent's spell resolves first.
	effectstest.CastCatalogSpell(t, g, "Fleshbag Marauder", "Creature — Zombie Warrior", effectstest.FleshbagMarauderOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if err := g.ActivateCatalogAbility(shielded.ID, bomb, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{fodder},
		Targets:      []game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[0].ID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrifice cost was blocked by the protection")
	}
}

// "Each opponent discards a card" from an opponent skips the player
// holding Tamiyo; the others are asked.
func TestOpponentsEachPlayerDiscardsSkipsProtectedPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	shielded := g.Seats[1]
	effectstest.PushCatalogWalker(g, shielded.ID, "Tamiyo, Collector of Tales", effectstest.TamiyoCollectorOracle, 5)

	effectstest.CastAndResolveCreature(t, g, "Burglar Rat", "Creature — Rat", "2f807301-37df-4724-871a-08e3512b07b3")
	effectstest.PassPriorityAroundTable(t, g)

	if n := effectstest.DiscardOwed(g, shielded.ID); n != 0 {
		t.Errorf("the protected player owes %d discards", n)
	}
	for _, p := range g.Seats[2:] {
		if effectstest.DiscardOwed(g, p.ID) != 1 {
			t.Errorf("seat %s should still owe one discard", p.ID)
		}
	}
}

// Annihilator from an opponent's attacker is stopped.
func TestAnnihilatorFromOpponentIsStopped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, shielded := g.Seats[0], g.Seats[1]
	eldrazi := effectstest.PushAnnihilatorCreature(g, me.ID, "annihilator 3")
	pushTajuru(g, shielded.ID)
	effectstest.PushAnnFodder(g, shielded.ID, 5)

	effectstest.DeclareAttack(t, g, shielded.ID, eldrazi)
	asked := effectstest.AnnSettle(t, g, nil)
	if len(asked) != 0 {
		t.Fatalf("the protected defender was asked to sacrifice: %d prompts", len(asked))
	}
	if n := effectstest.AnnCountOn(g, shielded.ID); n != 6 {
		t.Errorf("defender controls %d permanents, want all 6", n)
	}
}

// Sacrifice-all sweeps skip the protected player's permanents only.
func TestSacrificeAllSkipsProtectedPlayersPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, shielded, other := g.Seats[0], g.Seats[1], g.Seats[2]
	pushTajuru(g, shielded.ID)
	green := func(owner uuid.UUID) uuid.UUID {
		return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Green Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}",
			Colors: []string{"G"}, Power: 2, Toughness: 2, Owner: owner, Controller: owner,
		})
	}
	theirs, others, mine := green(shielded.ID), green(other.ID), green(me.ID)

	effectstest.CastCatalogSpell(t, g, "All Is Dust", "Kindred Sorcery — Eldrazi", effectstest.B08AllIsDustOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(theirs) {
		t.Error("the protected player's permanent was sacrificed")
	}
	if g.Battlefield.Contains(others) || g.Battlefield.Contains(mine) {
		t.Error("everyone else's coloured permanents are still sacrificed")
	}
}

// --- one test per card ---------------------------------------------

func TestTajuruPreserverIsRegisteredAndComplete(t *testing.T) {
	spec, ok := effects.Lookup(effectstest.TajuruPreserverOracle)
	if !ok || spec.Completeness != effects.CompletenessFull || !spec.OpponentEffectProtections[0].Sacrifice || spec.OpponentEffectProtections[0].Discard {
		t.Fatalf("Tajuru Preserver: %+v ok=%v", spec, ok)
	}
}

func TestSigardaHostOfHeronsProtectsFromSacrificeOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, shielded := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, shielded.ID, "Sigarda, Host of Herons", "Legendary Creature — Angel", effectstest.SigardaHostOracle, false)
	_ = me
	effectstest.CastCatalogSpell(t, g, "Fleshbag Marauder", "Creature — Zombie Warrior", effectstest.FleshbagMarauderOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SacrificeChoiceFor(g, shielded.ID) != nil {
		t.Error("Sigarda's controller was asked to sacrifice")
	}
	spec, _ := effects.Lookup(effectstest.SigardaHostOracle)
	if spec.Completeness != effects.CompletenessFull {
		t.Errorf("Sigarda is %v, the clause is live so it ships complete", spec.Completeness)
	}
	// Sigarda says nothing about discarding.
	if spec.OpponentEffectProtections[0].Discard {
		t.Error("Sigarda protects from discard; it prints only sacrifice")
	}
}

// Tamiyo +1: name a card, reveal four, matching nonland cards to hand,
// the rest (including lands named in error) to the graveyard.
func TestTamiyoPlusOneNamesRevealsAndTakes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tamiyo := effectstest.PushCatalogWalker(g, me.ID, "Tamiyo, Collector of Tales", effectstest.TamiyoCollectorOracle, 5)
	deep := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Sol Ring", TypeLine: "Artifact"})
	// Pushed last = top of the library.
	for _, c := range []game.Card{
		{Name: "Sol Ring", TypeLine: "Artifact"},
		{Name: "Grizzly Bears", TypeLine: "Creature — Bear"},
		{Name: "Sol Ring", TypeLine: "Artifact"},
		{Name: "Forest", TypeLine: "Basic Land — Forest"},
	} {
		c.Owner, c.Controller, c.InstanceID = me.ID, me.ID, uuid.New()
		me.Library.PushTop(c)
	}
	effectstest.AdvanceToMainOf(t, g, g.Turn.ActiveSeat)
	hand := len(me.Hand.Cards)
	grave := len(me.Graveyard.Cards)

	if err := g.ActivateCatalogAbility(me.ID, tamiyo, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.ActNameFor(t, g, me.ID, "sol ring")
	effectstest.PassPriorityAroundTable(t, g)

	if got := len(me.Hand.Cards) - hand; got != 2 {
		t.Errorf("two Sol Rings in the top four go to hand, got %d", got)
	}
	if got := len(me.Graveyard.Cards) - grave; got != 2 {
		t.Errorf("the other two of the four go to the graveyard, got %d", got)
	}
	if me.Library.Size() == 0 || me.Library.Cards[0].InstanceID != deep {
		t.Error("the fifth card is not revealed and stays in the library")
	}
	if effectstest.LoyaltyCount(g, tamiyo) != 6 {
		t.Errorf("loyalty = %d, want 6", effectstest.LoyaltyCount(g, tamiyo))
	}
}

// Naming a land takes nothing: the clause says nonland.
func TestTamiyoPlusOneNamingALandTakesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tamiyo := effectstest.PushCatalogWalker(g, me.ID, "Tamiyo, Collector of Tales", effectstest.TamiyoCollectorOracle, 5)
	for i := 0; i < 4; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	}
	effectstest.AdvanceToMainOf(t, g, g.Turn.ActiveSeat)
	hand := len(me.Hand.Cards)
	if err := g.ActivateCatalogAbility(me.ID, tamiyo, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.ActNameFor(t, g, me.ID, "Forest")
	effectstest.PassPriorityAroundTable(t, g)
	if len(me.Hand.Cards) != hand {
		t.Errorf("a land was taken to hand by naming it")
	}
}

// Tamiyo −3: return target card from your graveyard to your hand.
func TestTamiyoMinusThreeReturnsACard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tamiyo := effectstest.PushCatalogWalker(g, me.ID, "Tamiyo, Collector of Tales", effectstest.TamiyoCollectorOracle, 5)
	card := effectstest.PushCatalogGraveyardCard(me, "Grizzly Bears", "Creature — Bear", "", 2, 2)
	effectstest.AdvanceToMainOf(t, g, g.Turn.ActiveSeat)
	if err := g.ActivateCatalogAbility(me.ID, tamiyo, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: card}},
	}); err != nil {
		t.Fatalf("−3: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(card) || me.Graveyard.Contains(card) {
		t.Error("the targeted card did not return to hand")
	}
	if effectstest.LoyaltyCount(g, tamiyo) != 2 {
		t.Errorf("loyalty = %d, want 2", effectstest.LoyaltyCount(g, tamiyo))
	}
}

// Tamiyo's static covers discard and sacrifice, against opponents only.
func TestTamiyoStaticDeclaresBothClauses(t *testing.T) {
	spec, ok := effects.Lookup(effectstest.TamiyoCollectorOracle)
	if !ok || spec.Completeness != effects.CompletenessFull {
		t.Fatalf("Tamiyo: ok=%v %+v", ok, spec.Completeness)
	}
	p := spec.OpponentEffectProtections[0]
	if !p.Discard || !p.Sacrifice {
		t.Errorf("Tamiyo's clause = %+v, want discard and sacrifice", p)
	}
}
