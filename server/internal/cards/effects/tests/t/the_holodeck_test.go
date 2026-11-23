package testst

import (
	"testing"

	"github.com/google/uuid"

	cardsth "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/th"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// the_holodeck_test.go — The Holodeck: "Once during each of your turns,
// you may play a permanent card you own from outside the game or from
// your graveyard." Two standing permissions, one per place, share ONE use
// per turn.

// holodeckTable is a two-seat table in seat 0's precombat main, with The
// Holodeck on seat 0's battlefield and `side` in seat 0's sideboard.
func holodeckTable(t *testing.T, side ...game.Card) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewSideboardCatalogGame(t, 2, side...)
	effectstest.AdvanceToMainOf(t, g, 0)
	me := g.Seats[0]
	deck := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "The Holodeck", TypeLine: "Legendary Artifact", ManaCost: "{7}",
		OracleID: cardsth.TheHolodeckOracleID, Owner: me.ID, Controller: me.ID,
	})
	return g, me, deck
}

func holodeckSideBear(name string) game.Card {
	c := effectstest.WithManaCost(effectstest.OutsideCard(name, "Creature — Bear"), "{1}")
	c.Power, c.Toughness = 2, 2
	return c
}

func holodeckFromGraveyard(g *game.Game, p *game.Player, id uuid.UUID) error {
	return g.CastSpell(p.ID, id, game.CastSpellParams{FromZone: "graveyard", Strict: true})
}

// A permanent card from the graveyard spends the turn's one use, and the
// sideboard is closed for the rest of the turn.
func TestTheHolodeckGraveyardPlaySpendsTheOutsideUseToo(t *testing.T) {
	g, me, _ := holodeckTable(t, holodeckSideBear("Side Bear"))
	grave := effectstest.SocB12GraveCard(me, "Grave Bear", "Creature — Bear", "{1}")
	effectstest.FundOutside(me, 1)
	if err := holodeckFromGraveyard(g, me, grave); err != nil {
		t.Fatalf("a permanent card from the graveyard: %v", err)
	}
	effectstest.ResolveAllPT(t, g)
	if !g.Battlefield.Contains(grave) {
		t.Fatal("the creature did not resolve")
	}
	side := effectstest.SideCardID(me, "Side Bear")
	effectstest.FundOutside(me, 1)
	if err := effectstest.CastFromOutside(g, me, side, true); err == nil {
		t.Error("a card was played from outside the game after the turn's use went to the graveyard")
	}
	if effectstest.EnumeratedFor(g, me.ID, side) {
		t.Error("the enumerator still offers the sideboard card")
	}
}

// And the reverse: a card from outside the game closes the graveyard.
func TestTheHolodeckOutsidePlaySpendsTheGraveyardUseToo(t *testing.T) {
	g, me, _ := holodeckTable(t, holodeckSideBear("Side Bear"))
	grave := effectstest.SocB12GraveCard(me, "Grave Bear", "Creature — Bear", "{1}")
	side := effectstest.SideCardID(me, "Side Bear")
	effectstest.FundOutside(me, 1)
	if !effectstest.EnumeratedFor(g, me.ID, side) {
		t.Error("the enumerator does not offer the sideboard card")
	}
	if err := effectstest.CastFromOutside(g, me, side, true); err != nil {
		t.Fatalf("a permanent card from outside the game: %v", err)
	}
	effectstest.ResolveAllPT(t, g)
	if card, ok := effectstest.BattlefieldCard(g, side); !ok || card.Controller != me.ID {
		t.Fatal("the sideboard creature did not resolve onto the battlefield")
	}
	effectstest.FundOutside(me, 1)
	if err := holodeckFromGraveyard(g, me, grave); err == nil {
		t.Error("a graveyard card was played after the turn's use went outside the game")
	}
	if effectstest.EnumeratedFor(g, me.ID, grave) {
		t.Error("the enumerator still offers the graveyard card")
	}
}

// A land is a permanent card and is PLAYED: it spends the one use, and it
// is the turn's land play.
func TestTheHolodeckLandFromOutsideIsTheUseAndTheLandDrop(t *testing.T) {
	g, me, _ := holodeckTable(t, effectstest.OutsideCard("Forest", "Basic Land — Forest"))
	grave := effectstest.SocB12GraveCard(me, "Grave Bear", "Creature — Bear", "{1}")
	forest := effectstest.SideCardID(me, "Forest")
	var left int
	g.WithWriteLock(func() { left = g.LandDropsRemainingLocked(me.ID) })
	if err := effectstest.CastFromOutside(g, me, forest, true); err != nil {
		t.Fatalf("play a land from outside the game: %v", err)
	}
	if !g.Battlefield.Contains(forest) {
		t.Fatal("the land is not on the battlefield")
	}
	var after int
	g.WithWriteLock(func() { after = g.LandDropsRemainingLocked(me.ID) })
	if after != left-1 {
		t.Errorf("land drops left %d, want %d", after, left-1)
	}
	effectstest.FundOutside(me, 1)
	if err := holodeckFromGraveyard(g, me, grave); err == nil {
		t.Error("the graveyard opened after a land from outside the game spent the use")
	}
}

// Only permanent cards: an instant or a sorcery opens from neither place,
// and refusing one spends nothing.
func TestTheHolodeckOpensOnlyPermanentCards(t *testing.T) {
	g, me, _ := holodeckTable(t, effectstest.WithManaCost(effectstest.OutsideCard("Side Shock", "Instant"), "{1}"), holodeckSideBear("Side Bear"))
	graveShock := effectstest.SocB12GraveCard(me, "Grave Shock", "Sorcery", "{1}")
	effectstest.FundOutside(me, 2)
	if err := effectstest.CastFromOutside(g, me, effectstest.SideCardID(me, "Side Shock"), true); err == nil {
		t.Error("an instant was cast from outside the game")
	}
	if err := holodeckFromGraveyard(g, me, graveShock); err == nil {
		t.Error("a sorcery was cast from the graveyard")
	}
	if err := effectstest.CastFromOutside(g, me, effectstest.SideCardID(me, "Side Bear"), true); err != nil {
		t.Errorf("the refusals spent the use: %v", err)
	}
}

// Only during your turn, and the use renews on your next one.
func TestTheHolodeckOnlyOnYourTurnAndRenewsEachTurn(t *testing.T) {
	g, me, _ := holodeckTable(t, holodeckSideBear("Side Bear"), holodeckSideBear("Later Bear"))
	effectstest.FundOutside(me, 1)
	if err := effectstest.CastFromOutside(g, me, effectstest.SideCardID(me, "Side Bear"), true); err != nil {
		t.Fatalf("first use: %v", err)
	}
	effectstest.ResolveAllPT(t, g)
	grave := effectstest.SocB12GraveCard(me, "Grave Bear", "Creature — Bear", "{1}")

	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceToMain(t, g)
	effectstest.FundOutside(me, 1)
	if err := holodeckFromGraveyard(g, me, grave); err == nil {
		t.Error("a graveyard card was played on another player's turn")
	}
	if err := effectstest.CastFromOutside(g, me, effectstest.SideCardID(me, "Later Bear"), true); err == nil {
		t.Error("a sideboard card was played on another player's turn")
	}

	effectstest.AdvanceToMainOf(t, g, 0)
	effectstest.FundOutside(me, 1)
	if err := holodeckFromGraveyard(g, me, grave); err != nil {
		t.Fatalf("the use did not renew on my next turn: %v", err)
	}
	effectstest.ResolveAllPT(t, g)
	effectstest.FundOutside(me, 1)
	if err := effectstest.CastFromOutside(g, me, effectstest.SideCardID(me, "Later Bear"), true); err == nil {
		t.Error("the renewed turn gave two uses")
	}
}

// The spent use survives a capture → JSON → restore.
func TestTheHolodeckSpentUseSurvivesARestore(t *testing.T) {
	g, me, _ := holodeckTable(t, holodeckSideBear("Side Bear"))
	effectstest.SocB12GraveCard(me, "Grave Bear", "Creature — Bear", "{1}")
	effectstest.FundOutside(me, 1)
	if err := effectstest.CastFromOutside(g, me, effectstest.SideCardID(me, "Side Bear"), true); err != nil {
		t.Fatalf("first use: %v", err)
	}
	effectstest.ResolveAllPT(t, g)

	r := effectstest.RestoreRoundTrip(t, g, true)
	rme := r.Seats[0]
	var grave uuid.UUID
	for _, c := range rme.Graveyard.Cards {
		if c.Name == "Grave Bear" {
			grave = c.InstanceID
		}
	}
	effectstest.FundOutside(rme, 1)
	if err := holodeckFromGraveyard(r, rme, grave); err == nil {
		t.Error("the restored table forgot the turn's use was spent")
	}
}

// A database pool (no sideboard) names the card through the outside_play
// action, which The Holodeck opens like Wish; after the use is spent the
// action is closed.
func TestTheHolodeckNamesACardFromTheDatabase(t *testing.T) {
	effectstest.InstallOutsideCollection(t)
	g, me, _ := holodeckTable(t)
	effectstest.NameWished(t, g, me, "Forest")
	forest := effectstest.SideCardID(me, "Forest")
	if err := effectstest.CastFromOutside(g, me, forest, true); err != nil {
		t.Fatalf("play the named land: %v", err)
	}
	if !g.Battlefield.Contains(forest) {
		t.Fatal("the named land is not on the battlefield")
	}
	if err := g.NameOutsideCardToPlay(me.ID); err == nil {
		t.Error("a second card could be named after the use was spent")
	}
}

// The database pick offers only permanent cards, and a use spent from the
// graveyard closes the naming action for the turn.
func TestTheHolodeckNamingOffersPermanentsAndClosesWithTheGraveyardUse(t *testing.T) {
	effectstest.InstallOutsideCollection(t)
	g, me, _ := holodeckTable(t)
	if err := g.NameOutsideCardToPlay(me.ID); err != nil {
		t.Fatalf("outside_play: %v", err)
	}
	c := effectstest.LatestCollectionPickFor(g, me.ID)
	if c == nil {
		t.Fatal("no collection pick")
	}
	if err := g.ResolveOutsideCollection(c.ID, me.ID, []string{"Lightning Bolt"}); err == nil && effectstest.SideCardID(me, "Lightning Bolt") != uuid.Nil {
		t.Error("an instant was named for a permanent-only play")
	}
	if c := effectstest.LatestCollectionPickFor(g, me.ID); c != nil {
		if err := g.ResolveOutsideCollection(c.ID, me.ID, nil); err != nil {
			t.Fatalf("decline: %v", err)
		}
	}

	grave := effectstest.SocB12GraveCard(me, "Grave Bear", "Creature — Bear", "{1}")
	effectstest.FundOutside(me, 1)
	if err := holodeckFromGraveyard(g, me, grave); err != nil {
		t.Fatalf("a permanent card from the graveyard: %v", err)
	}
	if err := g.NameOutsideCardToPlay(me.ID); err == nil {
		t.Error("the naming action stayed open after the graveyard spent the use")
	}
}
