package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trigger_independence_damage_test.go — the aristocrats batches the
// arena kept asking about: a damage step whose source died in the same
// event it triggers on (Syr Konrad in the wipe), "~ deals N damage to
// you" (City of Brass, Midnight Reaper), and the targeted drain (Blood
// Artist, Falkenrath Noble). Each now has a footprint, so a batch whose
// order cannot show goes on with no CR 603.3b prompt.

const (
	syrKonradOracle         = "14c3ff84-1e82-4606-a433-869fc52cc382"
	falkenrathNobleOracle   = "3739b179-bc81-4737-8376-66a57e16b942"
	tatyovaOracle           = "0715e860-3b3b-4331-9718-207973e94fee"
	vindictiveVampireOracle = "11ad664e-36d3-4d5b-8a87-59ea715877e0"
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
		{"Tatyova, Benthic Druid", tatyovaOracle, []game.FootprintStep{gain, draw}},
		{"Vindictive Vampire", vindictiveVampireOracle, []game.FootprintStep{step(game.FootprintDamageEachOpponent, 1), gain}},
		{"City of Brass", cityOfBrassOracle, []game.FootprintStep{toYou}},
		{"Midnight Reaper", midnightReaperOracle, []game.FootprintStep{toYou, draw}},
		{"Blood Artist", bloodArtistOracle, []game.FootprintStep{targetLoses, gain}},
		{"Falkenrath Noble", falkenrathNobleOracle, []game.FootprintStep{targetLoses, gain}},
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
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Syr Konrad, the Grim", TypeLine: "Legendary Creature — Human Knight",
		OracleID: syrKonradOracle, Power: 5, Toughness: 4, Owner: owner, Controller: owner,
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	konrad := pushCatalogPermanent(g, me.ID, "Syr Konrad, the Grim", "Creature — Human Knight", syrKonradOracle, false)
	reaper := pushCatalogPermanent(g, me.ID, "Midnight Reaper", "Creature — Zombie Knight", midnightReaperOracle, false)
	bear := pushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	life, oppLife, hand := me.Life, opp.Life, len(me.Hand.Cards)

	wipe(g, konrad, reaper, bear)
	if ch := settleUntilOrderPrompt(t, g); ch != nil {
		t.Fatalf("the wipe's batch asked for an order: %+v", ch)
	}
	settleWithoutPrompts(t, g)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	konrad := pushCatalogPermanent(g, me.ID, "Syr Konrad, the Grim", "Creature — Human Knight", syrKonradOracle, false)
	vampire := pushCatalogPermanent(g, me.ID, "Vindictive Vampire", "Creature — Vampire", vindictiveVampireOracle, false)
	bear := pushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	life, oppLife := me.Life, opp.Life

	wipe(g, konrad, vampire, bear)
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(konrad); err != nil {
			t.Fatalf("exiling Konrad: %v", err)
		}
	})
	if ch := settleUntilOrderPrompt(t, g); ch != nil {
		t.Fatalf("the batch asked for an order: %+v", ch)
	}
	settleWithoutPrompts(t, g)
	// Konrad saw the Vampire and the Bear die, the Vampire saw Konrad
	// and the Bear: four points to each opponent, two life for me.
	if opp.Life != oppLife-4 || me.Life != life+2 {
		t.Errorf("life %d → %d and opponent %d → %d, want +2 and -4", life, me.Life, oppLife, opp.Life)
	}
}

// "Another creature you control": the Vampire does not see its own death.
// Until the footprint work it was written with the helper that does.
func TestVindictiveVampireDoesNotSeeItsOwnDeath(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vampire := pushCatalogPermanent(g, me.ID, "Vindictive Vampire", "Creature — Vampire", vindictiveVampireOracle, false)
	life, oppLife := me.Life, opp.Life
	wipe(g, vampire)
	settleWithoutPrompts(t, g)
	if me.Life != life || opp.Life != oppLife {
		t.Errorf("the Vampire's own death drained: life %d → %d, opponent %d → %d", life, me.Life, oppLife, opp.Life)
	}
}

// The record is what the damage reads: a Konrad that had lifelink as he
// died deals lifelink damage from the graveyard, which the footprint does
// not describe, so the batch keeps its prompt.
func TestADepartedLifelinkSourceStillAsks(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	konrad := pushLifelinkKonrad(g, me.ID)
	reaper := pushCatalogPermanent(g, me.ID, "Midnight Reaper", "Creature — Zombie Knight", midnightReaperOracle, false)
	bear := pushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	before := promptsByReason(t, game.TriggerOrderReasonUndescribed)

	wipe(g, konrad, reaper, bear)
	if ch := settleUntilOrderPrompt(t, g); ch == nil {
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
		if ch := triggerOrderPrompt(g); ch != nil {
			return ch
		}
		if latestPickTarget(g, me) != nil {
			pickPlayer(t, g, me, target)
			continue
		}
		if stackFullyEmpty(g) {
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Blood Artist", "Creature — Vampire", bloodArtistOracle, false)
	pushCatalogPermanent(g, me.ID, "Falkenrath Noble", "Creature — Vampire", falkenrathNobleOracle, false)
	bear := pushCatalogPermanent(g, opp.ID, "Bear", "Creature — Bear", "", false)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Blood Artist", "Creature — Vampire", bloodArtistOracle, false)
	pushCatalogPermanent(g, me.ID, "Falkenrath Noble", "Creature — Vampire", falkenrathNobleOracle, false)
	bear := pushCatalogPermanent(g, opp.ID, "Bear", "Creature — Bear", "", false)
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
	g := newCatalogGame(t)
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
