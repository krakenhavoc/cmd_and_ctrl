package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// chandra_torch_of_defiance_test.go — Chandra, Torch of Defiance (#3510):
// the first +1's paid cast from exile during the resolution and its "if
// you don't", the mana +1, the −3 and the emblem.

const chandraTorchOracle = "12cc800c-c7be-4811-8207-fd01a99cc892"

// chandraTorchTable seeds Chandra with `loyalty` under seat 0 in a main
// phase, with `top` on top of seat 0's library (nil for none).
func chandraTorchTable(t *testing.T, loyalty int, top *game.Card) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[0]
	chandra := pushCatalogPermanent(g, me.ID, "Chandra, Torch of Defiance", "Legendary Planeswalker — Chandra", chandraTorchOracle, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(chandra, game.CounterLoyalty, loyalty) })
	if top != nil {
		top.Owner, top.Controller = me.ID, me.ID
		me.Library.PushTop(*top)
	}
	advanceToMain(t, g)
	return g, me, chandra
}

// assertOpponentsLost checks each opponent lost exactly `n` since `before`.
func assertOpponentsLost(t *testing.T, g *game.Game, me uuid.UUID, before map[uuid.UUID]int, n int) {
	t.Helper()
	for id, life := range opponentLives(g, me) {
		if got := before[id] - life; got != n {
			t.Errorf("opponent %s lost %d, want %d", id, got, n)
		}
	}
}

func chandraTorchGrant(me *game.Player, card uuid.UUID) *game.CastPermission {
	for i := range me.CastPermissions {
		p := &me.CastPermissions[i]
		for _, ref := range p.Cards {
			if ref.ID == card {
				return p
			}
		}
	}
	return nil
}

func TestChandraTorchDeclineDealsTwoToEachOpponent(t *testing.T) {
	spell := uuid.New()
	g, me, chandra := chandraTorchTable(t, 4, &game.Card{InstanceID: spell, Name: "Divination", TypeLine: "Sorcery", ManaCost: "{2}{U}"})
	before := opponentLives(g, me.ID)
	life := me.Life

	b16Activate(t, g, me.ID, chandra, 0, game.ActivateAbilityParams{})
	if !g.Exile.Contains(spell) {
		t.Fatal("the top card was not exiled")
	}
	var prompt *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceMayCast {
			prompt = c
		}
	}
	if prompt == nil || prompt.MayCastCard != spell || prompt.MayCastKeyword != game.MayCastKeywordPaidOrElse {
		t.Fatalf("a paid_or_else may_cast over the exiled card: %+v", prompt)
	}
	assertOpponentsLost(t, g, me.ID, before, 0)

	answerMayCastPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	assertOpponentsLost(t, g, me.ID, before, 2)
	if me.Life != life {
		t.Errorf("Chandra's controller is not an opponent: life %d → %d", life, me.Life)
	}
	if !g.Exile.Contains(spell) {
		t.Error("the uncast card stays in exile")
	}
	if chandraTorchGrant(me, spell) != nil {
		t.Error("a declined offer grants nothing")
	}
	if loyaltyOf(g, chandra) != 5 {
		t.Errorf("+1 loyalty: %d", loyaltyOf(g, chandra))
	}
}

func TestChandraTorchAcceptPaysTheCostAndDealsNoDamage(t *testing.T) {
	spell := uuid.New()
	g, me, chandra := chandraTorchTable(t, 4, &game.Card{InstanceID: spell, Name: "Divination", TypeLine: "Sorcery", ManaCost: "{1}{R}"})
	mountains := []uuid.UUID{
		pushLandFor(g, me.ID, "Mountain", "Basic Land — Mountain"),
		pushLandFor(g, me.ID, "Mountain", "Basic Land — Mountain"),
	}
	before := opponentLives(g, me.ID)

	b16Activate(t, g, me.ID, chandra, 0, game.ActivateAbilityParams{})
	answerMayCastPrompt(t, g, me.ID, true)
	grant := chandraTorchGrant(me, spell)
	if grant == nil {
		t.Fatal("accepting granted no cast over the exiled card")
	}
	if grant.Cost != "" || grant.Timing != game.TimingFlash || !grant.CastOnly ||
		grant.LapseOnPass != game.LapseStaysInExile || grant.IfNotCast == "" {
		t.Errorf("grant = %+v, want printed cost, flash timing, cast-only, pass-closed with an if-not-cast", *grant)
	}

	// The cast is PAID: with an empty pool it is refused.
	err := g.CastSpell(me.ID, spell, game.CastSpellParams{FromZone: string(game.ZoneExile), Strict: true})
	if !errors.Is(err, game.ErrInsufficientMana) {
		t.Fatalf("a free cast went through: %v", err)
	}
	// CR 601.2g: mana abilities during the cast.
	for _, m := range mountains {
		if err := g.ActivateManaAbility(me.ID, m, 0, game.ManaAbilityParams{}); err != nil {
			t.Fatalf("tap Mountain: %v", err)
		}
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{FromZone: string(game.ZoneExile), Strict: true}); err != nil {
		t.Fatalf("cast from exile: %v", err)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("the cost was not paid: pool %+v", me.ManaPool)
	}
	if chandraTorchGrant(me, spell) != nil {
		t.Error("the cast spent the grant")
	}
	passPriorityAroundTable(t, g)
	assertOpponentsLost(t, g, me.ID, before, 0)
	if !me.Graveyard.Contains(spell) {
		t.Error("the cast sorcery resolved into the graveyard")
	}
}

func TestChandraTorchAcceptThenPassDealsTheDamage(t *testing.T) {
	spell := uuid.New()
	g, me, chandra := chandraTorchTable(t, 4, &game.Card{InstanceID: spell, Name: "Divination", TypeLine: "Sorcery", ManaCost: "{2}{U}"})
	before := opponentLives(g, me.ID)

	b16Activate(t, g, me.ID, chandra, 0, game.ActivateAbilityParams{})
	answerMayCastPrompt(t, g, me.ID, true)
	assertOpponentsLost(t, g, me.ID, before, 0)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass: %v", err)
	}
	assertOpponentsLost(t, g, me.ID, before, 2)
	if chandraTorchGrant(me, spell) != nil {
		t.Error("the pass closed the window")
	}
	if !g.Exile.Contains(spell) {
		t.Error("the uncast card stays in exile")
	}
}

func TestChandraTorchLandIsNotAskedAndDealsTheDamage(t *testing.T) {
	land := uuid.New()
	g, me, chandra := chandraTorchTable(t, 4, &game.Card{InstanceID: land, Name: "Mountain", TypeLine: "Basic Land — Mountain"})
	before := opponentLives(g, me.ID)

	b16Activate(t, g, me.ID, chandra, 0, game.ActivateAbilityParams{})
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceMayCast {
			t.Fatal("a land can't be cast, so nothing is asked")
		}
	}
	assertOpponentsLost(t, g, me.ID, before, 2)
	if !g.Exile.Contains(land) {
		t.Error("the land stays in exile")
	}
}

func TestChandraTorchEmptyLibraryDealsTheDamage(t *testing.T) {
	g, me, chandra := chandraTorchTable(t, 4, nil)
	me.Library.Cards = nil
	before := opponentLives(g, me.ID)
	b16Activate(t, g, me.ID, chandra, 0, game.ActivateAbilityParams{})
	assertOpponentsLost(t, g, me.ID, before, 2)
}

func TestChandraTorchSecondPlusOneAddsTwoRed(t *testing.T) {
	g, me, chandra := chandraTorchTable(t, 4, nil)
	b16Activate(t, g, me.ID, chandra, 1, game.ActivateAbilityParams{})
	var toks []game.ManaToken
	g.ReadSnapshot(func() { toks = append(toks, me.ManaPool...) })
	if len(toks) != 2 || toks[0].Color != "R" || toks[1].Color != "R" {
		t.Fatalf("{R}{R}: %+v", toks)
	}
	if loyaltyOf(g, chandra) != 5 {
		t.Errorf("+1 loyalty: %d", loyaltyOf(g, chandra))
	}
}

func TestChandraTorchMinusThreeDealsFourToTargetCreature(t *testing.T) {
	g, me, chandra := chandraTorchTable(t, 4, nil)
	opp := g.Seats[1]
	big := b16Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 5)
	b16Activate(t, g, me.ID, chandra, 2, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: big}},
	})
	c, ok := battlefieldCard(g, big)
	if !ok || c.DamageMarked != 4 {
		t.Errorf("4 damage on the 0/5: %+v", c)
	}
	if loyaltyOf(g, chandra) != 1 {
		t.Errorf("−3 loyalty: %d", loyaltyOf(g, chandra))
	}
}

func TestChandraTorchEmblemDealsFiveOnEverySpell(t *testing.T) {
	g, me, chandra := chandraTorchTable(t, 7, nil)
	opp := g.Seats[1]
	b16Activate(t, g, me.ID, chandra, 3, game.ActivateAbilityParams{})
	if me.Emblems.Size() != 1 {
		t.Fatalf("one emblem: %d", me.Emblems.Size())
	}
	life := opp.Life
	castCatalogSpell(t, g, "Some Sorcery", "Sorcery", "", nil)
	answerPickTargetPlayer(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != life-5 {
		t.Errorf("whenever you cast a spell, 5 damage to any target: %d → %d", life, opp.Life)
	}
}
