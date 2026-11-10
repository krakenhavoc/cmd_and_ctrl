package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// eye_of_duskmantle_test.go — #5429: "You may play lands and cast spells
// from among cards in your graveyard you've surveilled this turn. If you
// cast a spell this way, you pay life equal to its mana value rather than
// paying its mana cost."

const eyeOfDuskmantleOracle = "04e7f739-fe50-440d-aa3d-e547295044a8"

// eyeTable is seat 0's main phase with the Eye on its battlefield.
func eyeTable(t *testing.T) (g *game.Game, me, opp *game.Player) {
	t.Helper()
	g = newCatalogGame(t)
	advanceToMain(t, g)
	me, opp = g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Eye of Duskmantle", "Creature — Eye", eyeOfDuskmantleOracle, false)
	return
}

// eyeLibrary puts a card on top of p's library.
func eyeLibrary(p *game.Player, name, typeLine, cost string) uuid.UUID {
	id := uuid.New()
	p.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
		Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID})
	return id
}

// eyeSurveil has p surveil len(bin)+len(keep), putting `bin` into the
// graveyard and `keep` back on top.
func eyeSurveil(t *testing.T, g *game.Game, p *game.Player, bin, keep []uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		g.SurveilKeyedForEffect(p.ID, uuid.Nil, len(bin)+len(keep), game.KeyedTail{})
	})
	c := latestChoiceOfKindFor(g, game.PendingChoiceSurveil, p.ID)
	if c == nil {
		t.Fatal("no surveil prompt")
	}
	if err := g.ResolveSurveil(c.ID, p.ID, bin, keep); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
}

func eyeCast(g *game.Game, who *game.Player, id uuid.UUID, targets ...game.TargetRef) error {
	return g.CastSpell(who.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: eyeOfDuskmantleAltCost,
		Strict: true, Targets: targets})
}

// A card surveilled into the graveyard this turn is cast from there for
// life equal to its mana value and no mana.
func TestEyeOfDuskmantleCastsASurveilledCardForLife(t *testing.T) {
	g, me, _ := eyeTable(t)
	kept := eyeLibrary(me, "Kept Bear", "Creature — Bear", "{1}{G}")
	bear := eyeLibrary(me, "Binned Giant", "Creature — Giant", "{3}{G}{G}")
	eyeSurveil(t, g, me, []uuid.UUID{bear}, []uuid.UUID{kept})
	if !enumeratedFor(g, me.ID, bear) {
		t.Fatal("the enumerator does not offer the surveilled card")
	}
	life := me.Life
	if err := eyeCast(g, me, bear); err != nil {
		t.Fatalf("cast the surveilled creature for life: %v", err)
	}
	if paid := life - me.Life; paid != 5 {
		t.Errorf("life paid = %d, want 5 (its mana value)", paid)
	}
	passPriorityAroundTable(t, g)
	if c, ok := battlefieldCard(g, bear); !ok || c.Controller != me.ID {
		t.Fatal("the creature did not resolve onto the battlefield under my control")
	}
}

// "Rather than paying its mana cost": the Eye does not let the card be
// cast for mana.
func TestEyeOfDuskmantleCannotPayTheManaCostInstead(t *testing.T) {
	g, me, _ := eyeTable(t)
	bear := eyeLibrary(me, "Binned Bear", "Creature — Bear", "{1}{G}")
	eyeSurveil(t, g, me, []uuid.UUID{bear}, nil)
	floatForTest(g, me, "GG")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{FromZone: "graveyard", Strict: true}); err == nil {
		t.Fatal("a surveilled card was cast from the graveyard for its mana cost")
	}
}

// Only cards this turn's surveil put into the graveyard: not a milled
// card, not a card already there, not one that left and came back.
func TestEyeOfDuskmantleOffersOnlyCardsSurveilledIntoTheGraveyard(t *testing.T) {
	g, me, _ := eyeTable(t)
	old := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: old, Name: "Old Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}",
		Owner: me.ID, Controller: me.ID})
	milled := eyeLibrary(me, "Milled Bear", "Creature — Bear", "{1}{G}")
	g.WithWriteLock(func() {
		if err := g.MillNForEffect(me.ID, 1); err != nil {
			t.Fatalf("mill: %v", err)
		}
	})
	if !me.Graveyard.Contains(milled) {
		t.Fatal("non-vacuity: the card was not milled")
	}
	for name, id := range map[string]uuid.UUID{"a card already in the graveyard": old, "a milled card": milled} {
		if err := eyeCast(g, me, id); err == nil {
			t.Errorf("%s was cast through the Eye", name)
		}
		if enumeratedFor(g, me.ID, id) {
			t.Errorf("the enumerator offers %s", name)
		}
	}

	// Surveilled, returned to hand, discarded: a new object (CR 400.7).
	bounced := eyeLibrary(me, "Bounced Bear", "Creature — Bear", "{1}{G}")
	eyeSurveil(t, g, me, []uuid.UUID{bounced}, nil)
	if !enumeratedFor(g, me.ID, bounced) {
		t.Fatal("non-vacuity: the surveilled card is offered before it leaves")
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, bounced); err != nil {
		t.Fatal(err)
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, bounced); err != nil {
		t.Fatal(err)
	}
	if err := eyeCast(g, me, bounced); err == nil {
		t.Error("a card that left the graveyard and came back was cast through the Eye")
	}
}

// Lands are played, not cast: no life, and the turn's land play is spent.
func TestEyeOfDuskmantlePlaysASurveilledLand(t *testing.T) {
	g, me, _ := eyeTable(t)
	land := eyeLibrary(me, "Binned Swamp", "Basic Land — Swamp", "")
	eyeSurveil(t, g, me, []uuid.UUID{land}, nil)
	life, drops := me.Life, g.LandDropsRemainingFor(me.ID)
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{FromZone: "graveyard", Strict: true}); err != nil {
		t.Fatalf("play the surveilled land: %v", err)
	}
	if me.Life != life {
		t.Errorf("a land cost %d life", life-me.Life)
	}
	if _, ok := battlefieldCard(g, land); !ok {
		t.Fatal("the land is not on the battlefield")
	}
	if got := g.LandDropsRemainingFor(me.ID); got != drops-1 {
		t.Errorf("land plays %d -> %d, want one spent", drops, got)
	}
	// With the land play spent, a second surveilled land is refused.
	second := eyeLibrary(me, "Second Swamp", "Basic Land — Swamp", "")
	eyeSurveil(t, g, me, []uuid.UUID{second}, nil)
	if err := g.CastSpell(me.ID, second, game.CastSpellParams{FromZone: "graveyard", Strict: true}); err == nil {
		t.Fatal("a second land was played from the graveyard in one turn")
	}
}

// CR 119.4: a player who cannot pay the life cannot cast it.
func TestEyeOfDuskmantleRefusesACastYouCannotPayLifeFor(t *testing.T) {
	g, me, _ := eyeTable(t)
	big := eyeLibrary(me, "Binned Colossus", "Creature — Giant", "{9}{G}")
	eyeSurveil(t, g, me, []uuid.UUID{big}, nil)
	g.WithWriteLock(func() { me.Life = 4 })
	if err := eyeCast(g, me, big); err == nil {
		t.Fatal("cast at 4 life for mana value 10")
	}
	if me.Life != 4 || !me.Graveyard.Contains(big) {
		t.Errorf("a refused cast changed the game: life %d, in graveyard %v", me.Life, me.Graveyard.Contains(big))
	}
}

// "This turn": the cards are no longer offered once the turn has ended.
func TestEyeOfDuskmantleLapsesWithTheTurn(t *testing.T) {
	g, me, _ := eyeTable(t)
	bear := eyeLibrary(me, "Binned Bear", "Creature — Bear", "{1}{G}")
	eyeSurveil(t, g, me, []uuid.UUID{bear}, nil)
	if !enumeratedFor(g, me.ID, bear) {
		t.Fatal("non-vacuity: the card is offered the turn it was surveilled")
	}
	b39NextTurnOf(t, g, 0)
	if err := eyeCast(g, me, bear); err == nil {
		t.Fatal("a card surveilled on an earlier turn was cast through the Eye")
	}
	if enumeratedFor(g, me.ID, bear) {
		t.Error("the enumerator still offers it on a later turn")
	}
}

// The permission changes where from, not when: a surveilled instant is
// cast on an opponent's turn, a sorcery is not.
func TestEyeOfDuskmantleKeepsTheCardsOwnTiming(t *testing.T) {
	g, me, opp := eyeTable(t)
	advanceToMainOf(t, g, 1)
	sorcery := eyeLibrary(me, "Binned Sorcery", "Sorcery", "{2}{B}")
	bolt := eyeLibrary(me, "Lightning Bolt", "Instant", "{R}")
	g.WithWriteLock(func() {
		for i := range me.Library.Cards {
			if me.Library.Cards[i].InstanceID == bolt {
				me.Library.Cards[i].OracleID = lightningBoltOracle
			}
		}
	})
	eyeSurveil(t, g, me, []uuid.UUID{bolt, sorcery}, nil)
	for i := 0; i < len(g.Seats) && g.Turn.PriorityHolder != 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if g.Turn.PriorityHolder != 0 || g.Turn.ActiveSeat != 1 {
		t.Fatal("the harness did not hand me priority on seat 1's turn")
	}
	if err := eyeCast(g, me, sorcery); err == nil {
		t.Fatal("a sorcery was cast on an opponent's turn")
	}
	life := me.Life
	if err := eyeCast(g, me, bolt, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("cast the surveilled instant on an opponent's turn: %v", err)
	}
	if me.Life != life-1 {
		t.Errorf("life %d -> %d, want 1 paid", life, me.Life)
	}
}

// "Your graveyard ... you've surveilled": the Eye's controller gets nothing
// from a card another player surveilled, and that player has no permission.
func TestEyeOfDuskmantleIsItsControllersAndTheirGraveyardOnly(t *testing.T) {
	g, me, opp := eyeTable(t)
	theirs := eyeLibrary(opp, "Their Bear", "Creature — Bear", "{1}{G}")
	eyeSurveil(t, g, opp, []uuid.UUID{theirs}, nil)
	if err := eyeCast(g, me, theirs); err == nil {
		t.Fatal("I cast a card from another player's graveyard through the Eye")
	}
	if enumeratedFor(g, me.ID, theirs) || enumeratedFor(g, opp.ID, theirs) {
		t.Error("the enumerator offers a card another player surveilled")
	}
}

// The surveil need not happen while the Eye is out: one that arrives later
// in the turn offers what was surveilled earlier that turn.
func TestEyeOfDuskmantleArrivingLaterSeesTheTurnsEarlierSurveil(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[0]
	bear := eyeLibrary(me, "Binned Bear", "Creature — Bear", "{1}{G}")
	eyeSurveil(t, g, me, []uuid.UUID{bear}, nil)
	if enumeratedFor(g, me.ID, bear) {
		t.Fatal("the card is offered with no Eye on the battlefield")
	}
	eye := pushCatalogPermanent(g, me.ID, "Eye of Duskmantle", "Creature — Eye", eyeOfDuskmantleOracle, false)
	life := me.Life
	if err := eyeCast(g, me, bear); err != nil {
		t.Fatalf("cast through an Eye that arrived after the surveil: %v", err)
	}
	if me.Life != life-2 {
		t.Errorf("life %d -> %d, want 2 paid", life, me.Life)
	}
	// And the permission goes with the Eye.
	second := eyeLibrary(me, "Second Bear", "Creature — Bear", "{1}{G}")
	passPriorityAroundTable(t, g)
	eyeSurveil(t, g, me, []uuid.UUID{second}, nil)
	killCreature(t, g, me.ID, eye)
	if err := eyeCast(g, me, second); err == nil {
		t.Fatal("a surveilled card was cast after the Eye left the battlefield")
	}
}

// The cast is on the wire as the one life offer on the graveyard card, so
// the client prices it the way the engine does, and only its holder sees it.
func TestEyeOfDuskmantleOffersExactlyTheLifeCostOnTheWire(t *testing.T) {
	g, me, opp := eyeTable(t)
	bear := eyeLibrary(me, "Binned Bear", "Creature — Bear", "{1}{G}")
	eyeSurveil(t, g, me, []uuid.UUID{bear}, nil)
	view := func(viewer uuid.UUID) *protocol.CardView {
		v := protocol.ViewOfGameFor(g, viewer.String())
		for _, s := range v.Seats {
			for i := range s.Graveyard.Cards {
				if s.Graveyard.Cards[i].InstanceID == bear.String() {
					return &s.Graveyard.Cards[i]
				}
			}
		}
		t.Fatal("the surveilled card is not in a graveyard view")
		return nil
	}
	mine := view(me.ID)
	if !mine.CastableHere {
		t.Error("my view has no castable_here on the surveilled card")
	}
	if len(mine.AlternativeCosts) != 1 || mine.AlternativeCosts[0].Key != eyeOfDuskmantleAltCost || mine.AlternativeCosts[0].Life != 2 {
		t.Errorf("view alternative costs = %+v, want the one 2-life offer", mine.AlternativeCosts)
	}
	if theirs := view(opp.ID); theirs.CastableHere {
		t.Error("an opponent's view says they may cast my surveilled card")
	}
}
