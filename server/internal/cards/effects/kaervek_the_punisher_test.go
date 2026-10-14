package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// kaervek_the_punisher_test.go — Kaervek, the Punisher (#2947, #2910).

const kaervekOracle = "fa9a6083-75ab-41ae-b607-39d21d9a3d3a"

func kaervekSetup(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	kaervek := pushDiesCreatureForTest(g, me.ID, "Kaervek, the Punisher", kaervekOracle, "Legendary Creature — Human Warlock", 3, 3)
	toMainForCost(t, g)
	return g, me, opp, kaervek
}

// kaervekResolveOn commits a crime for me, picks target as Kaervek's
// target and resolves the trigger, leaving the copy offer open.
func kaervekResolveOn(t *testing.T, g *game.Game, me *game.Player, target uuid.UUID) {
	t.Helper()
	commitCrimeForMinerTest(g, me.ID)
	passPriorityAroundTable(t, g)
	answerTriggerTargets(t, g, me.ID, target)
	passPriorityAroundTable(t, g)
}

// Casting the copy costs its printed mana cost, makes the copy do what
// the card does, and costs 2 life; the original stays in exile.
func TestKaervekCastsACopyAtItsCostAndLosesTwoLife(t *testing.T) {
	g, me, _, _ := kaervekSetup(t)
	ritual := gyCard(me, "Dark Ritual", "Instant", "{B}", darkRitualOracle)
	life := me.Life

	kaervekResolveOn(t, g, me, ritual)
	if !inExile(g, ritual) {
		t.Fatal("the targeted black card is exiled")
	}
	cp := copyOf(t, g, "Dark Ritual")
	g.ReadSnapshot(func() {
		perm := g.CastPermissionOnCardForEffect(cp, game.ZoneExile)
		if perm == nil || perm.Player != me.ID {
			t.Fatalf("no permission for Kaervek's controller over the copy: %+v", perm)
		}
		if perm.Cost != "" {
			t.Errorf("permission cost = %q, want the printed cost (not free)", perm.Cost)
		}
	})
	if me.Life != life {
		t.Fatalf("life moved to %d before the copy was cast", me.Life)
	}
	if err := castCopy(g, me.ID, cp.InstanceID, true); err == nil {
		t.Fatal("the copy was cast with no {B}: the cast is not free")
	}
	floatForTest(g, me, "B")
	if err := castCopy(g, me.ID, cp.InstanceID, true); err != nil {
		t.Fatalf("cast the copy for {B}: %v", err)
	}
	if me.Life != life-2 {
		t.Errorf("life %d -> %d, want -2 for casting the copy", life, me.Life)
	}
	passPriorityAroundTable(t, g)
	if got := len(me.ManaPool); got != 3 {
		t.Errorf("mana pool holds %d, want {B}{B}{B} from the copied Dark Ritual", got)
	}
	if !inExile(g, ritual) {
		t.Error("the original card stays in exile")
	}
	if n := len(cardCopies(g)); n != 0 {
		t.Errorf("%d copies left after the copy resolved", n)
	}
}

// Declining the cast (passing priority) loses no life and the copy
// ceases to exist. The crime here is real: Lightning Bolt at an
// opponent, and Kaervek's trigger goes above it.
func TestKaervekDecliningTheCopyLosesNoLife(t *testing.T) {
	g, me, opp, kaervek := kaervekSetup(t)
	ritual := gyCard(me, "Dark Ritual", "Instant", "{B}", darkRitualOracle)
	life := me.Life

	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, []game.TargetRef{playerTarget(opp.ID)})
	if latestPickTarget(g, me.ID) == nil {
		passPriorityAroundTable(t, g)
	}
	if latestPickTarget(g, me.ID) == nil {
		t.Fatal("targeting an opponent is a crime: Kaervek should ask for a target")
	}
	answerTriggerTargets(t, g, me.ID, ritual)
	if triggerOnStack(g, kaervek) == nil {
		t.Fatal("Kaervek's trigger is not on the stack")
	}
	passPriorityAroundTable(t, g)

	if !inExile(g, ritual) {
		t.Error("the card is exiled whether or not the copy is cast")
	}
	if n := len(cardCopies(g)); n != 0 {
		t.Errorf("%d copies left after passing: an uncast copy ceases to exist", n)
	}
	if me.Life != life {
		t.Errorf("life %d -> %d: no life is lost when the copy is not cast", life, me.Life)
	}
}

// A copy of a permanent card resolves as a token (CR 707.10b).
func TestKaervekCopyOfACreatureCardBecomesAToken(t *testing.T) {
	g, me, _, _ := kaervekSetup(t)
	nighthawk := gyCard(me, "Vampire Nighthawk", "Creature — Vampire Shaman", "{1}{B}{B}", vampireNighthawkOracle)

	kaervekResolveOn(t, g, me, nighthawk)
	cp := copyOf(t, g, "Vampire Nighthawk")
	floatForTest(g, me, "BBB")
	if err := castCopy(g, me.ID, cp.InstanceID, true); err != nil {
		t.Fatalf("cast the creature copy: %v", err)
	}
	passPriorityAroundTable(t, g)
	var token *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Vampire Nighthawk" {
			token = &g.Battlefield.Cards[i]
		}
	}
	if token == nil {
		t.Fatal("the creature copy did not resolve onto the battlefield")
	}
	if !token.IsToken() {
		t.Error("a copy of a permanent spell becomes a token")
	}
	if !inExile(g, nighthawk) {
		t.Error("the original card stays in exile")
	}
}

// Only black cards in your own graveyard can be chosen, and with none
// there the trigger is not put on the stack. Another player's crime
// does not trigger it.
func TestKaervekTargetsOnlyYourBlackCards(t *testing.T) {
	g, me, opp, kaervek := kaervekSetup(t)
	gyCard(me, "Lightning Bolt", "Instant", "{R}", lightningBoltOracle)
	theirs := gyCard(opp, "Dark Ritual", "Instant", "{B}", darkRitualOracle)

	commitCrimeForMinerTest(g, opp.ID)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil || triggerOnStack(g, kaervek) != nil {
		t.Fatal("another player's crime triggered Kaervek")
	}

	commitCrimeForMinerTest(g, me.ID)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("with no black card in your graveyard there is nothing to target")
	}
	if !opp.Graveyard.Contains(theirs) {
		t.Error("an opponent's graveyard card was touched")
	}

	ritual := gyCard(me, "Dark Ritual", "Instant", "{B}", darkRitualOracle)
	commitCrimeForMinerTest(g, me.ID)
	passPriorityAroundTable(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("a black card in your graveyard: Kaervek should ask for a target")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}); err == nil {
		t.Error("a card in an opponent's graveyard is not a legal target")
	}
	answerTriggerTargets(t, g, me.ID, ritual)
	passPriorityAroundTable(t, g)
	if !inExile(g, ritual) {
		t.Error("the chosen black card is exiled")
	}
}

// A target that left the graveyard in response is illegal: nothing is
// exiled or copied (CR 608.2b).
func TestKaervekDoesNothingWhenItsTargetLeaves(t *testing.T) {
	g, me, _, kaervek := kaervekSetup(t)
	ritual := gyCard(me, "Dark Ritual", "Instant", "{B}", darkRitualOracle)

	commitCrimeForMinerTest(g, me.ID)
	passPriorityAroundTable(t, g)
	answerTriggerTargets(t, g, me.ID, ritual)
	if triggerOnStack(g, kaervek) == nil {
		t.Fatal("Kaervek's trigger is not on the stack")
	}
	if err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID},
		game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, ritual,
	); err != nil {
		t.Fatalf("move the target away: %v", err)
	}
	passPriorityAroundTable(t, g)
	if inExile(g, ritual) || !me.Hand.Contains(ritual) {
		t.Error("a target that left the graveyard is not exiled")
	}
	if n := len(cardCopies(g)); n != 0 {
		t.Errorf("%d copies made of a card that was never exiled", n)
	}
}
