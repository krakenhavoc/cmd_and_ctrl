package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// attachments_refusal_test.go — #3160, CR 701.3b / 303.4j. An effect
// that tries to attach an Aura or Equipment to something it can't be
// attached to does nothing: the attachment keeps its host. Before, the
// attach happened and the CR 704.5m/n sweep then buried the Aura or
// dropped the Equipment.
//
// The catalog cards whose outcome changes are the ones that MOVE an
// attachment that is already on a host: Magnetic Theft and Brass
// Squire's ability (an Equipment aimed at a protected creature),
// Maddening Hex (a Curse aimed at a protected player), and the equip
// family (attachments.go). The cards that put an attachment onto the
// battlefield attached (Light-Paws, Mantle of the Ancients, Gift of
// Immortality, Hammer of Nazahn, Rohirrim Chargers, Role tokens) have
// no host to keep: a refused attach leaves them unattached, and the
// CR 704.5m sweep buries an Aura exactly as it did when the attach
// was made and then undone.

// attachNow runs AttachForEffect the way a resolving effect does.
func attachNow(t *testing.T, g *game.Game, attachment uuid.UUID, host game.TargetRef) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(attachment, host); err != nil {
			t.Fatalf("AttachForEffect: %v", err)
		}
	})
}

func skippedWhy(g *game.Game, attachment uuid.UUID) string {
	var out string
	g.ReadSnapshot(func() {
		for _, ev := range g.Events {
			if ev.Kind == game.EventAttachSkipped && ev.CardID == attachment {
				out = ev.Label
			}
		}
	})
	return out
}

// An Aura whose enchant clause rejects the host stays where it is: Wild
// Growth says "Enchant land", and an effect that aims it at a creature
// does nothing (CR 303.4j).
func TestAnAuraWhoseEnchantClauseRejectsTheHostDoesNotMove(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	forest := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest",
		Owner: me.ID, Controller: me.ID,
	})
	bear := auraBear(g, me.ID)
	growth := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Wild Growth", TypeLine: auraTypeLine,
		OracleID: wildGrowthAuraOracle, Owner: me.ID, Controller: me.ID,
	})
	attachNow(t, g, growth, game.TargetRef{Kind: game.TargetCard, ID: forest})
	if host := attachmentHostOf(t, g, growth); host.ID != forest {
		t.Fatalf("setup: Wild Growth not on the Forest: %+v", host)
	}

	attachNow(t, g, growth, game.TargetRef{Kind: game.TargetCard, ID: bear})
	g.RunStateChecksForTest()

	if host := attachmentHostOf(t, g, growth); host.ID != forest {
		t.Errorf("Wild Growth moved to %+v although it can't enchant a creature", host)
	}
	if skippedWhy(g, growth) == "" {
		t.Error("no EventAttachSkipped for the refused attach")
	}
}

// The unchanged half: a legal move of the same Aura still happens.
func TestAnAuraWhoseEnchantClauseAcceptsTheHostMoves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	first := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest",
		Owner: me.ID, Controller: me.ID,
	})
	second := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Mountain", TypeLine: "Basic Land — Mountain",
		Owner: me.ID, Controller: me.ID,
	})
	growth := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Wild Growth", TypeLine: auraTypeLine,
		OracleID: wildGrowthAuraOracle, Owner: me.ID, Controller: me.ID,
	})
	attachNow(t, g, growth, game.TargetRef{Kind: game.TargetCard, ID: first})
	attachNow(t, g, growth, game.TargetRef{Kind: game.TargetCard, ID: second})

	if host := attachmentHostOf(t, g, growth); host.ID != second {
		t.Errorf("a legal move did not happen: Wild Growth on %+v", host)
	}
	if why := skippedWhy(g, growth); why != "" {
		t.Errorf("a legal attach was skipped: %q", why)
	}
}

// Pacifism is white; an effect that aims it at a creature with
// protection from white does nothing, and the creature it was holding
// stays pacified.
func TestAProtectedCreatureKeepsTheAuraOffItAndTheOldHostKeepsIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	old := auraBear(g, me.ID)
	pro := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Knight of Red Cloaks", TypeLine: "Creature — Knight",
		Power: 2, Toughness: 2, Keywords: []string{"protection from white"},
		Owner: me.ID, Controller: me.ID,
	})
	pac := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Pacifism", TypeLine: auraTypeLine,
		OracleID: pacifismAuraOracle, Colors: []string{"W"}, Owner: me.ID, Controller: me.ID,
	})
	attachNow(t, g, pac, game.TargetRef{Kind: game.TargetCard, ID: old})

	attachNow(t, g, pac, game.TargetRef{Kind: game.TargetCard, ID: pro})
	g.RunStateChecksForTest()

	if host := attachmentHostOf(t, g, pac); host.ID != old {
		t.Errorf("Pacifism moved to %+v although the creature has protection from white", host)
	}
	if got := auraRestrictions(t, g, old); got&game.CantAttackOrBlock == 0 {
		t.Error("the creature Pacifism was already holding was freed")
	}
	if skippedWhy(g, pac) == "" {
		t.Error("no EventAttachSkipped for the refused attach")
	}
}

// Magnetic Theft: "Attach target Equipment to target creature." A
// creature with protection from artifacts can't be equipped, and the
// Equipment stays on the creature it was on.
func TestMagneticTheftCannotEquipACreatureProtectedFromArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	holder := auraBear(g, me.ID)
	pro := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Protected Knight", TypeLine: testCreatureTypeLine,
		Power: 2, Toughness: 2, Keywords: []string{"protection from artifacts"},
		Owner: me.ID, Controller: me.ID,
	})
	sword := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bonesplitter", TypeLine: equipTypeLine,
		OracleID: bonesplitterOracle, Owner: me.ID, Controller: me.ID,
	})
	attachNow(t, g, sword, game.TargetRef{Kind: game.TargetCard, ID: holder})

	castCatalogSpell(t, g, "Magnetic Theft", "Instant", magneticTheftOracle,
		[]game.TargetRef{
			{Kind: game.TargetCard, ID: sword},
			{Kind: game.TargetCard, ID: pro},
		})
	passPriorityAroundTable(t, g)

	if host := attachmentHostOf(t, g, sword); host.ID != holder {
		t.Errorf("Bonesplitter is on %+v, want it to stay on the creature it was equipping", host)
	}
	if got := effectivePower(t, g, holder); got != 4 {
		t.Errorf("the holder lost the Equipment bonus: power %d, want 4", got)
	}
}

// Maddening Hex moves to "another one of your opponents chosen at
// random". Every other opponent has protection from everything, so the
// Aura can't be attached to any of them and stays on its player
// (CR 702.16i, 701.3b); before, it moved to one of them and was then
// put into the graveyard.
func TestMaddeningHexStaysWhenEveryOtherOpponentIsProtected(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hex := pushCatalogPermanent(g, me.ID, "Maddening Hex", "Enchantment — Aura Curse", s58MaddeningHex, false)
	g.WithWriteLock(func() {
		findBattlefieldCardForTest(g, hex).AttachedTo = game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}
		for _, p := range g.Seats[2:] {
			g.GrantPlayerStaticForEffect(p.ID, "protection from everything",
				"Test — protection from everything", uuid.Nil, g.UntilYourNextTurnDuration(p.ID))
		}
	})
	advanceToUpkeepOf(t, g, 1)
	s58CastFromHand(t, g, opp, "Shock", "Instant", "{R}")

	if !g.Battlefield.Contains(hex) {
		t.Fatal("Maddening Hex left the battlefield; it should not have moved")
	}
	if at := findBattlefieldCardForTest(g, hex).AttachedTo; at.Kind != game.TargetPlayer || at.ID != opp.ID {
		t.Errorf("Maddening Hex is on %+v, want it to stay on the first opponent", at)
	}
}
