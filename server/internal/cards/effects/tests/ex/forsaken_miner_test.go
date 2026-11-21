package testsex

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const forsakenMinerOracle = "d37e0127-6f68-438e-8b9e-a71bb3c34733"

func putForsakenMinerInGraveyard(p *game.Player) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id,
		Name:       "Forsaken Miner",
		TypeLine:   "Creature — Skeleton Rogue",
		OracleID:   forsakenMinerOracle,
		Power:      2,
		Toughness:  2,
		Owner:      p.ID,
		Controller: p.ID,
	})
	return id
}

func TestForsakenMinerTriggersOnlyFromItsGraveyardAndForItsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opponent := g.Seats[0], g.Seats[1]
	miner := putForsakenMinerInGraveyard(me)

	effectstest.CommitCrimeForMinerTest(g, opponent.ID)
	if effectstest.TriggerOnStack(g, miner) != nil {
		t.Fatal("another player's crime triggered Forsaken Miner")
	}

	// A successful targeted spell against your own permanent is not a
	// crime, even though it has a target.
	mine := effectstest.PushPermanentForTest(g, me.ID, "My Bear", "", "Creature — Bear")
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.B16TargetCard(mine))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.TriggerOnStack(g, miner) != nil {
		t.Fatal("targeting your own permanent triggered Forsaken Miner")
	}

	// Two opponent-controlled objects in one announcement still commit
	// one crime, and trigger Miner once.
	opponentA := effectstest.PushCostedPermanentForTest(g, opponent.ID, "Opponent A", "Creature — Bear", "{1}{G}")
	opponentB := effectstest.PushCostedPermanentForTest(g, opponent.ID, "Opponent B", "Creature — Bear", "{1}{G}")
	effectstest.CastCatalogSpell(t, g, "Arc Trail", "Sorcery", effectstest.ArcTrailOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: opponentA}, {Kind: game.TargetCard, ID: opponentB, Slot: 1}})
	if n := effectstest.TriggersOnStackFrom(g, miner); n != 1 {
		t.Fatalf("one crime with two qualifying targets produced %d Miner triggers, want one", n)
	}
	if effectstest.TriggerOnStack(g, miner) == nil {
		t.Fatal("Forsaken Miner in its owner's graveyard did not trigger for their crime")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)

	// Its printed ability does not function from the battlefield.
	if err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID},
		game.ZoneRef{Kind: game.ZoneBattlefield}, miner,
	); err != nil {
		t.Fatalf("move Miner to battlefield: %v", err)
	}
	effectstest.CommitCrimeForMinerTest(g, me.ID)
	if n := effectstest.TriggersOnStackFrom(g, miner); n != 0 {
		t.Fatalf("crime while Miner was on the battlefield produced %d triggers, want 0", n)
	}
}

func TestForsakenMinerMayPayReturnsItAndMayDecline(t *testing.T) {
	for _, tc := range []struct {
		name string
		pay  bool
	}{
		{name: "pay", pay: true},
		{name: "decline", pay: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			miner := putForsakenMinerInGraveyard(me)

			effectstest.CommitCrimeForMinerTest(g, me.ID)
			effectstest.PassPriorityAroundTable(t, g)
			if tc.pay {
				effectstest.B06AddMana(me, "B")
			}
			prompt := effectstest.AnswerPayUnless(t, g, me.ID, tc.pay)
			if prompt.PayCost != "{B}" {
				t.Errorf("optional cost = %q, want {B}", prompt.PayCost)
			}
			if tc.pay {
				if !g.Battlefield.Contains(miner) || me.Graveyard.Contains(miner) {
					for _, ev := range g.EventsThisTurn() {
						if ev.Kind == game.EventEffectError {
							t.Logf("effect error: %s", ev.ErrorMsg)
						}
					}
					t.Fatalf("paying {B} did not return the same card to the battlefield (battlefield=%t graveyard=%t mana=%v)", g.Battlefield.Contains(miner), me.Graveyard.Contains(miner), me.ManaPool)
				}
			} else if !me.Graveyard.Contains(miner) {
				t.Fatal("declining {B} moved Forsaken Miner out of the graveyard")
			}
		})
	}
}

func TestForsakenMinerTriggerDoesNotFollowNewGraveyardObject(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	miner := putForsakenMinerInGraveyard(me)
	effectstest.CommitCrimeForMinerTest(g, me.ID)

	if err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID},
		game.ZoneRef{Kind: game.ZoneBattlefield}, miner,
	); err != nil {
		t.Fatalf("move Miner out before trigger resolution: %v", err)
	}
	if err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneBattlefield},
		game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, miner,
	); err != nil {
		t.Fatalf("return Miner as a new graveyard object: %v", err)
	}

	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("the stale trigger offered a payment for a new graveyard object")
	}
	if !me.Graveyard.Contains(miner) {
		t.Fatal("the new graveyard object moved while resolving the stale trigger")
	}
}

func TestForsakenMinerChecksObjectAgainAfterPaymentPrompt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	miner := putForsakenMinerInGraveyard(me)
	effectstest.CommitCrimeForMinerTest(g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("the resolving trigger did not ask whether to pay {B}")
	}

	if err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID},
		game.ZoneRef{Kind: game.ZoneBattlefield}, miner,
	); err != nil {
		t.Fatalf("move Miner out during payment prompt: %v", err)
	}
	if err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneBattlefield},
		game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, miner,
	); err != nil {
		t.Fatalf("return Miner as a new graveyard object: %v", err)
	}

	effectstest.B06AddMana(me, "B")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	if !me.Graveyard.Contains(miner) || g.Battlefield.Contains(miner) {
		t.Fatal("paying for a stale trigger returned the new graveyard object")
	}
}

func TestForsakenMinerCannotBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	miner := effectstest.CastCatalogSpell(t, g, "Forsaken Miner", "Creature — Skeleton Rogue", forsakenMinerOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertRestrictions(t, g, miner, game.CantBlock)
}
