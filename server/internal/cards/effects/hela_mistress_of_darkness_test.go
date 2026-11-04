package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hela_mistress_of_darkness_test.go — Hela, Mistress of Darkness // Dark
// Tutor (#4460). Each test plays the real catalog entries.

// helaCard builds the printed card the way deck.toGameCard does: per-face
// data, then SetFace(0).
func helaCard(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   helaOracleID,
		Layout:     game.LayoutAdventure,
		Owner:      owner,
		Controller: owner,
		Faces: []game.Face{
			{
				Name:      "Hela, Mistress of Darkness",
				TypeLine:  "Legendary Creature — God Sorcerer Villain",
				ManaCost:  "{4}{B}{B}",
				Colors:    []string{"B"},
				Power:     6,
				Toughness: 6,
			},
			{
				Name:       "Dark Tutor",
				TypeLine:   "Sorcery — Vengeance",
				ManaCost:   "{1}{B}{B}",
				Colors:     []string{"B"},
				OracleText: "(Cast only from your graveyard. Then exile it.)\nSearch your library for a card, put it into your hand or graveyard, then shuffle. You lose 3 life.",
			},
		},
	}
	c.SetFace(0)
	return c
}

func helaMain(t *testing.T, g *game.Game) {
	t.Helper()
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
}

func helaZombieTokens(g *game.Game, controller uuid.UUID) (n, tapped int) {
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Zombie" && c.Controller == controller {
			n++
			if c.Tapped {
				tapped++
			}
		}
	}
	return n, tapped
}

// --- Dark Tutor: Vengeance --------------------------------------------------

func TestDarkTutorCannotBeCastFromHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	helaMain(t, g)
	c := helaCard(me.ID)
	me.Hand.PushTop(c)
	floatMana(t, g, me, "{C}{B}{B}")

	err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Face: 1, Strict: true})
	if !errors.Is(err, game.ErrCastZoneNotAllowed) {
		t.Fatalf("Dark Tutor from hand: %v, want ErrCastZoneNotAllowed", err)
	}
	if !me.Hand.Contains(c.InstanceID) {
		t.Error("the refused cast moved the card out of the hand")
	}
}

func TestHelaCannotBeCastFromTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	helaMain(t, g)
	c := helaCard(me.ID)
	me.Graveyard.PushTop(c)
	floatMana(t, g, me, "{C}{C}{C}{C}{B}{B}")

	err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Face: 0, FromZone: string(game.ZoneGraveyard), Strict: true})
	if err == nil {
		t.Fatal("Hela was cast from the graveyard")
	}
}

func TestDarkTutorFromTheGraveyardTutorsToHandCostsThreeAndExilesTheCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	helaMain(t, g)
	needle := stapleLibraryCard(me, "Needle", "Artifact")
	c := helaCard(me.ID)
	me.Graveyard.PushTop(c)
	floatMana(t, g, me, "{C}{B}{B}")
	before := me.Life

	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Face: 1, FromZone: string(game.ZoneGraveyard), Strict: true}); err != nil {
		t.Fatalf("cast Dark Tutor from the graveyard: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, needle)
	answerOptionPick(t, g, me.ID, 0)
	passPriorityAroundTable(t, g)

	if !me.Hand.Contains(needle) {
		t.Error("the chosen card did not reach the hand")
	}
	if me.Life != before-3 {
		t.Errorf("life %d -> %d, want -3", before, me.Life)
	}
	if !g.Exile.Contains(c.InstanceID) {
		t.Error("Dark Tutor did not end in exile")
	}
	if me.Graveyard.Contains(c.InstanceID) {
		t.Error("Dark Tutor went back to the graveyard")
	}
	if perm := g.CastPermissionOnCardByIDForEffect(c.InstanceID); perm != nil {
		t.Errorf("exiled Dark Tutor left a cast permission %+v: it is not an Adventure", perm)
	}
}

func TestDarkTutorCanPutTheCardIntoTheGraveyardInstead(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	helaMain(t, g)
	needle := stapleLibraryCard(me, "Needle", "Artifact")
	c := helaCard(me.ID)
	me.Graveyard.PushTop(c)
	floatMana(t, g, me, "{C}{B}{B}")
	before := me.Life

	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Face: 1, FromZone: string(game.ZoneGraveyard), Strict: true}); err != nil {
		t.Fatalf("cast Dark Tutor: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, needle)
	answerOptionPick(t, g, me.ID, 1)
	passPriorityAroundTable(t, g)

	if !me.Graveyard.Contains(needle) {
		t.Error("the chosen card did not reach the graveyard")
	}
	if me.Hand.Contains(needle) {
		t.Error("the chosen card also reached the hand")
	}
	if me.Life != before-3 {
		t.Errorf("life %d -> %d, want -3", before, me.Life)
	}
	if !g.Exile.Contains(c.InstanceID) {
		t.Error("Dark Tutor did not end in exile")
	}
}

func TestDarkTutorCounteredIsStillExiled(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	helaMain(t, g)
	stapleLibraryCard(me, "Needle", "Artifact")
	c := helaCard(me.ID)
	me.Graveyard.PushTop(c)
	floatMana(t, g, me, "{C}{B}{B}")

	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Face: 1, FromZone: string(game.ZoneGraveyard), Strict: true}); err != nil {
		t.Fatalf("cast Dark Tutor: %v", err)
	}
	if err := g.CounterSpell(c.InstanceID, nil); err != nil {
		t.Fatalf("CounterSpell: %v", err)
	}
	if !g.Exile.Contains(c.InstanceID) {
		t.Error("a countered Dark Tutor did not end in exile")
	}
	if me.Graveyard.Contains(c.InstanceID) {
		t.Error("a countered Dark Tutor went to the graveyard")
	}
}

// --- Hela, the creature half --------------------------------------------------

func TestHelaCastFromHandEntersTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	helaMain(t, g)
	c := helaCard(me.ID)
	me.Hand.PushTop(c)
	floatMana(t, g, me, "{C}{C}{C}{C}{B}{B}")

	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Face: 0, Strict: true}); err != nil {
		t.Fatalf("cast Hela: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(c.InstanceID) {
		t.Fatal("Hela is not on the battlefield")
	}
}

// --- Hela's trigger -----------------------------------------------------------

func helaTable(t *testing.T) (g *game.Game, me, opp *game.Player) {
	t.Helper()
	g = newCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	opp = g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	helaMain(t, g)
	pushCatalogPermanent(g, me.ID, "Hela, Mistress of Darkness", "Legendary Creature — God Sorcerer Villain", helaOracleID, false)
	return g, me, opp
}

func boltPlayer(t *testing.T, g *game.Game, target *game.Player) {
	t.Helper()
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: target.ID}})
	passPriorityAroundTable(t, g)
}

func TestHelaMakesThatManyTappedZombiesWhenAPlayerLosesLifeOnYourTurn(t *testing.T) {
	g, me, opp := helaTable(t)
	boltPlayer(t, g, opp)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)

	n, tapped := helaZombieTokens(g, me.ID)
	if n != 3 || tapped != 3 {
		t.Errorf("Zombies = %d (%d tapped), want 3 tapped for the 3 life lost", n, tapped)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Zombie" && (c.Power != 2 || c.Toughness != 2 || !c.HasColor("B")) {
			t.Errorf("Zombie is %d/%d colours %v, want a black 2/2", c.Power, c.Toughness, c.Colors)
		}
	}
}

func TestHelaSeesYourOwnLifeLossToo(t *testing.T) {
	g, me, _ := helaTable(t)
	boltPlayer(t, g, me)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if n, _ := helaZombieTokens(g, me.ID); n != 3 {
		t.Errorf("Zombies = %d, want 3: a player includes you", n)
	}
}

func TestHelaIsOptionalAndDecliningDoesNotUseTheTurnUp(t *testing.T) {
	g, me, opp := helaTable(t)
	boltPlayer(t, g, opp)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if n, _ := helaZombieTokens(g, me.ID); n != 0 {
		t.Fatalf("declining still made %d Zombies", n)
	}

	boltPlayer(t, g, opp)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if n, _ := helaZombieTokens(g, me.ID); n != 3 {
		t.Errorf("Zombies after declining once = %d, want 3", n)
	}
}

func TestHelaDoesThisOnlyOnceEachTurn(t *testing.T) {
	g, me, opp := helaTable(t)
	boltPlayer(t, g, opp)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)

	boltPlayer(t, g, opp)
	if latestTriggerPrompt(g, me.ID) != nil {
		t.Error("a second loss the same turn was offered Zombies")
	}
	if n, _ := helaZombieTokens(g, me.ID); n != 3 {
		t.Errorf("Zombies = %d, want 3: only the first loss counts", n)
	}
}

func TestHelaCountsEverythingLostAtTheSameTime(t *testing.T) {
	g, me, opp := helaTable(t)
	g.WithWriteLock(func() {
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -2)
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -3)
	})
	g.RunStateChecksForTest()
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if n, _ := helaZombieTokens(g, me.ID); n != 5 {
		t.Errorf("Zombies = %d, want 5: 2 and 3 life lost together is one loss of 5", n)
	}
}

func TestHelaIgnoresLifeLostOnAnOpponentsTurn(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	third := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	helaMain(t, g)
	pushCatalogPermanent(g, other.ID, "Hela, Mistress of Darkness", "Legendary Creature — God Sorcerer Villain", helaOracleID, false)

	boltPlayer(t, g, third)
	if latestTriggerPrompt(g, other.ID) != nil || latestTriggerPrompt(g, active.ID) != nil {
		t.Error("a loss on another player's turn offered Zombies")
	}
	if n, _ := helaZombieTokens(g, other.ID); n != 0 {
		t.Errorf("Zombies = %d, want 0: it is not Hela's controller's turn", n)
	}
}
