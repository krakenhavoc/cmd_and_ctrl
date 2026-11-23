package teststo

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestTheDamageAndTargetedDrainsDeclareFootprints(t *testing.T) {
	step := func(k game.FootprintKind, n int) game.FootprintStep { return game.FootprintStep{Kind: k, N: n} }
	gain, draw := step(game.FootprintGainLife, 1), step(game.FootprintDraw, 1)
	toYou := step(game.FootprintDamageController, 1)
	targetLoses := step(game.FootprintTargetPlayerLosesLife, 1)
	for _, c := range []struct {
		name   string
		oracle string
		want   []game.FootprintStep
	}{
		{"Tatyova, Benthic Druid", effectstest.TatyovaOracle, []game.FootprintStep{gain, draw}},
		{"Vindictive Vampire", effectstest.VindictiveVampireOracle, []game.FootprintStep{step(game.FootprintDamageEachOpponent, 1), gain}},
		{"City of Brass", effectstest.CityOfBrassOracle, []game.FootprintStep{toYou}},
		{"Midnight Reaper", effectstest.MidnightReaperOracle, []game.FootprintStep{toYou, draw}},
		{"Blood Artist", effectstest.BloodArtistOracle, []game.FootprintStep{targetLoses, gain}},
		{"Falkenrath Noble", effectstest.FalkenrathNobleOracle, []game.FootprintStep{targetLoses, gain}},
	} {
		rows := game.CatalogTriggers(c.oracle)
		if len(rows) != 1 {
			t.Fatalf("%s: %d rows", c.name, len(rows))
		}
		got := rows[0].Footprint
		if len(got) != len(c.want) {
			t.Errorf("%s: footprint %+v, want %+v", c.name, got, c.want)
			continue
		}
		for i := range got {
			if got[i].Kind != c.want[i].Kind || got[i].N != c.want[i].N {
				t.Errorf("%s: step %d is %+v, want %+v", c.name, i, got[i], c.want[i])
			}
		}
	}
}

// pushLifelinkKonrad puts a Syr Konrad with lifelink onto the
// battlefield under owner.
func pushLifelinkKonrad(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Syr Konrad, the Grim", TypeLine: "Legendary Creature — Human Knight",
		OracleID: effectstest.SyrKonradOracle, Power: 5, Toughness: 4, Owner: owner, Controller: owner,
		Keywords: []string{"lifelink"},
	})
}

// wipe destroys ids at once, as a Wrath does.
func wipe(g *game.Game, ids ...uuid.UUID) {
	g.WithWriteLock(func() { g.DestroyPermanentsForEffect(ids) })
}

// Syr Konrad, Midnight Reaper and a Bear die together. Konrad sees the
// other two die and Reaper sees all three, so the batch is two "1 damage
// to each opponent" from a Konrad already in the graveyard and three "1
// damage to you, draw a card". The damage lands on different players and
// only the draws touch a library, so every order gives the same game.
func TestSyrKonradDyingInTheWipeHeTriggersOnDoesNotAsk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	konrad := effectstest.PushCatalogPermanent(g, me.ID, "Syr Konrad, the Grim", "Creature — Human Knight", effectstest.SyrKonradOracle, false)
	reaper := effectstest.PushCatalogPermanent(g, me.ID, "Midnight Reaper", "Creature — Zombie Knight", effectstest.MidnightReaperOracle, false)
	bear := effectstest.PushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	life, oppLife, hand := me.Life, opp.Life, len(me.Hand.Cards)

	wipe(g, konrad, reaper, bear)
	if ch := effectstest.SettleUntilOrderPrompt(t, g); ch != nil {
		t.Fatalf("the wipe's batch asked for an order: %+v", ch)
	}
	effectstest.SettleWithoutPrompts(t, g)
	if opp.Life != oppLife-2 {
		t.Errorf("opponent life %d → %d, want -2 (Konrad saw Reaper and the Bear die)", oppLife, opp.Life)
	}
	if me.Life != life-3 {
		t.Errorf("life %d → %d, want -3 (Reaper saw all three die)", life, me.Life)
	}
	if got := len(me.Hand.Cards) - hand; got != 3 {
		t.Errorf("drew %d, want 3", got)
	}
}

// A Konrad that has moved on from the graveyard (as a commander does, to
// the command zone) matches no last-known record, and the damage tail
// deals his damage with no traits at all: still plain damage, so the
// batch with Vindictive Vampire's drains does not ask either.
func TestSyrKonradMovedOnFromTheGraveyardDoesNotAsk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	konrad := effectstest.PushCatalogPermanent(g, me.ID, "Syr Konrad, the Grim", "Creature — Human Knight", effectstest.SyrKonradOracle, false)
	vampire := effectstest.PushCatalogPermanent(g, me.ID, "Vindictive Vampire", "Creature — Vampire", effectstest.VindictiveVampireOracle, false)
	bear := effectstest.PushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	life, oppLife := me.Life, opp.Life

	wipe(g, konrad, vampire, bear)
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(konrad); err != nil {
			t.Fatalf("exiling Konrad: %v", err)
		}
	})
	if ch := effectstest.SettleUntilOrderPrompt(t, g); ch != nil {
		t.Fatalf("the batch asked for an order: %+v", ch)
	}
	effectstest.SettleWithoutPrompts(t, g)
	// Konrad saw the Vampire and the Bear die, the Vampire saw Konrad
	// and the Bear: four points to each opponent, two life for me.
	if opp.Life != oppLife-4 || me.Life != life+2 {
		t.Errorf("life %d → %d and opponent %d → %d, want +2 and -4", life, me.Life, oppLife, opp.Life)
	}
}

// "Another creature you control": the Vampire does not see its own death.
// Until the footprint work it was written with the helper that does.
func TestVindictiveVampireDoesNotSeeItsOwnDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vampire := effectstest.PushCatalogPermanent(g, me.ID, "Vindictive Vampire", "Creature — Vampire", effectstest.VindictiveVampireOracle, false)
	life, oppLife := me.Life, opp.Life
	wipe(g, vampire)
	effectstest.SettleWithoutPrompts(t, g)
	if me.Life != life || opp.Life != oppLife {
		t.Errorf("the Vampire's own death drained: life %d → %d, opponent %d → %d", life, me.Life, oppLife, opp.Life)
	}
}

// The record is what the damage reads: a Konrad that had lifelink as he
// died deals lifelink damage from the graveyard, which the footprint does
// not describe, so the batch keeps its prompt.
func TestADepartedLifelinkSourceStillAsks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	konrad := pushLifelinkKonrad(g, me.ID)
	reaper := effectstest.PushCatalogPermanent(g, me.ID, "Midnight Reaper", "Creature — Zombie Knight", effectstest.MidnightReaperOracle, false)
	bear := effectstest.PushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	before := promptsByReason(t, game.TriggerOrderReasonUndescribed)

	wipe(g, konrad, reaper, bear)
	if ch := effectstest.SettleUntilOrderPrompt(t, g); ch == nil {
		t.Fatal("damage from a source that died with lifelink must keep the prompt")
	}
	if got := promptsByReason(t, game.TriggerOrderReasonUndescribed) - before; got != 1 {
		t.Errorf("counted %v under undescribed, want 1", got)
	}
}

// settleTargetingPlayer passes priority until the batch is on the stack,
// answering each pick_target prompt of me's with target. It returns the
// first trigger-order prompt it meets, or nil once the stack is empty.
func settleTargetingPlayer(t *testing.T, g *game.Game, me, target uuid.UUID) *game.PendingChoice {
	t.Helper()
	for i := 0; i < 24; i++ {
		if ch := effectstest.TriggerOrderPrompt(g); ch != nil {
			return ch
		}
		if effectstest.LatestPickTarget(g, me) != nil {
			effectstest.PickPlayer(t, g, me, target)
			continue
		}
		if effectstest.StackFullyEmpty(g) {
			return nil
		}
		if err := g.PassPriority(); err != nil && !errors.Is(err, game.ErrChoicePending) {
			t.Fatal(err)
		}
	}
	t.Fatal("the batch did not settle")
	return nil
}

// Blood Artist and Falkenrath Noble each drain the opponent they target
// for one death: two losses for one player and two gains for me, which
// commute.
func TestTwoTargetedDrainsDoNotAsk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Blood Artist", "Creature — Vampire", effectstest.BloodArtistOracle, false)
	effectstest.PushCatalogPermanent(g, me.ID, "Falkenrath Noble", "Creature — Vampire", effectstest.FalkenrathNobleOracle, false)
	bear := effectstest.PushCatalogPermanent(g, opp.ID, "Bear", "Creature — Bear", "", false)
	life, oppLife := me.Life, opp.Life

	wipe(g, bear)
	if ch := settleTargetingPlayer(t, g, me.ID, opp.ID); ch != nil {
		t.Fatalf("two targeted drains asked for an order: %+v", ch)
	}
	if opp.Life != oppLife-2 || me.Life != life+2 {
		t.Errorf("life %d → %d and opponent %d → %d, want +2 and -2", life, me.Life, oppLife, opp.Life)
	}
}

// At 1 life the first drain takes the target out of the game, and the
// second then has no legal target and gains nothing (CR 608.2b): where
// the loss lands in the batch can show, so the check keeps the prompt.
func TestTwoTargetedDrainsAskWhenTheTargetCanDie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Blood Artist", "Creature — Vampire", effectstest.BloodArtistOracle, false)
	effectstest.PushCatalogPermanent(g, me.ID, "Falkenrath Noble", "Creature — Vampire", effectstest.FalkenrathNobleOracle, false)
	bear := effectstest.PushCatalogPermanent(g, opp.ID, "Bear", "Creature — Bear", "", false)
	opp.Life = 1

	wipe(g, bear)
	if ch := settleTargetingPlayer(t, g, me.ID, opp.ID); ch == nil {
		t.Fatal("a drain that can eliminate its target must keep the prompt")
	}
}

// Two draws from one library put the same cards in hand in either order,
// so two Llanowar Visionaries entering together no longer ask (a draw
// beside a scry still does: trigger_independence_draws_test.go).
func TestTwoDrawsOfOneLibraryDoNotAsk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hand := me.Hand.Size()
	if ch := enterTwo(t, g, "Llanowar Visionary", "Creature — Elf Druid Shaman", llanowarVisionaryOracle,
		"Llanowar Visionary", "Creature — Elf Druid Shaman", llanowarVisionaryOracle); ch != nil {
		t.Fatalf("two draws asked for an order: %+v", ch)
	}
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d → %d, want +2", hand, me.Hand.Size())
	}
}
