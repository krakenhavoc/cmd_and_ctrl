package testsav

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// block_requirement_cards_test.go — #1597's proof cards on the CR 509.1c
// machinery: Lure, Grand Melee's block half, Prized Unicorn, Gaea's
// Protector, Razorgrass Screen, Irresistible Prey and Taunting
// Challenge. One test per card, each driven through a real attack and
// the defending player's pass, which is the declaration's checkpoint.

const (
	lureOracle              = "7a7425ba-4478-4bc4-855f-abf947ea4fa2"
	prizedUnicornOracle     = "766787ef-654a-4544-b5d7-d2f96949edf1"
	gaeasProtectorOracle    = "d91d5d22-9f94-4a41-9970-0a8ddd2f7711"
	razorgrassScreenOracle  = "1818324c-2738-42e2-a78b-63abf325d8e9"
	irresistiblePreyOracle  = "cb102767-f8d8-4477-9a24-07686870bda7"
	tauntingChallengeOracle = "50c52320-9f84-4655-a91a-e09b71f9025c"
)

// TestLureMakesEveryCreatureBlockTheEnchantedCreature — three creatures
// able to block the enchanted creature all must, and one sent at the
// other attacker is refused.
func TestLureMakesEveryCreatureBlockTheEnchantedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.ReqCreature(g, me.ID, "Bear", "", 2, 2, "haste")
	other := effectstest.ReqCreature(g, me.ID, "Other Bear", "", 2, 2, "haste")
	w1 := effectstest.ReqCreature(g, opp.ID, "Wall One", "", 0, 4)
	w2 := effectstest.ReqCreature(g, opp.ID, "Wall Two", "", 0, 4)
	w3 := effectstest.ReqCreature(g, opp.ID, "Wall Three", "", 0, 4)
	effectstest.Enchant(t, g, "Lure", lureOracle, bear)

	effectstest.ReqToBlockers(t, g, bear, other)
	br := effectstest.ReqRefusal(t, g.PassPriority())
	if got := br.Sentence(opp.ID); got != "Wall One must block Bear if able (Lure)." {
		t.Errorf("sentence = %q", got)
	}
	effectstest.ReqRefusal(t, effectstest.ReqBlock(t, g, other, w1))
	if err := effectstest.ReqBlock(t, g, bear, w1, w2); err != nil {
		t.Fatalf("two walls on the Lure'd Bear: %v", err)
	}
	if br = effectstest.ReqRefusal(t, g.PassPriority()); br.Blocker != w3 {
		t.Errorf("the refusal names %s, want the wall left home", br.Blocker)
	}
	if err := effectstest.ReqBlock(t, g, bear, w3); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with all three blocking: %v", err)
	}
}

// TestGrandMeleeEnforcesBothHalves — every creature attacks, and every
// creature of the defending player blocks, each refusal naming Grand
// Melee.
func TestGrandMeleeEnforcesBothHalves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.PushPermanentForTest(g, me.ID, "Grand Melee", effectstest.GrandMeleeOracle, "Enchantment")
	bear := effectstest.ReqCreature(g, me.ID, "Bear", "", 2, 2, "haste")
	a := effectstest.ReqCreature(g, opp.ID, "Their Wall", "", 0, 4)
	b := effectstest.ReqCreature(g, opp.ID, "Their Elf", "", 1, 1)

	effectstest.AdvanceToDeclareAttackersOf(t, g, g.Turn.ActiveSeat)
	var re *game.AttackRequirementError
	if err := g.PassPriority(); !errors.As(err, &re) || re.Attacker != bear {
		t.Fatalf("attack half: pass = %v, want a refusal naming the Bear", err)
	}
	effectstest.ReqToBlockers(t, g, bear)
	br := effectstest.ReqRefusal(t, g.PassPriority())
	if got := br.Sentence(opp.ID); got != "Their Wall must block this combat if able (Grand Melee)." {
		t.Errorf("sentence = %q", got)
	}
	if err := effectstest.ReqBlock(t, g, bear, a); err != nil {
		t.Fatal(err)
	}
	effectstest.ReqRefusal(t, g.PassPriority())
	if err := effectstest.ReqBlock(t, g, bear, b); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with both blocking: %v", err)
	}
}

// TestPrizedUnicornLuresEveryAbleBlocker — Lure printed on the attacker.
func TestPrizedUnicornLuresEveryAbleBlocker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	unicorn := effectstest.ReqCreature(g, me.ID, "Prized Unicorn", prizedUnicornOracle, 2, 2, "haste")
	a := effectstest.ReqCreature(g, opp.ID, "Wall A", "", 0, 4)
	b := effectstest.ReqCreature(g, opp.ID, "Wall B", "", 0, 4)
	effectstest.ReqToBlockers(t, g, unicorn)
	br := effectstest.ReqRefusal(t, g.PassPriority())
	if got := br.Sentence(opp.ID); got != "Wall A must block Prized Unicorn if able." {
		t.Errorf("sentence = %q", got)
	}
	if err := effectstest.ReqBlock(t, g, unicorn, a, b); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with both blocking: %v", err)
	}
}

// TestGaeasProtectorMustBeBlocked — one blocker is enough; none is not.
func TestGaeasProtectorMustBeBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	gp := effectstest.ReqCreature(g, me.ID, "Gaea's Protector", gaeasProtectorOracle, 4, 2, "haste")
	a := effectstest.ReqCreature(g, opp.ID, "Wall A", "", 0, 4)
	effectstest.ReqCreature(g, opp.ID, "Wall B", "", 0, 4)
	effectstest.ReqToBlockers(t, g, gp)
	br := effectstest.ReqRefusal(t, g.PassPriority())
	if got := br.Sentence(opp.ID); got != "Gaea's Protector must be blocked if able." {
		t.Errorf("sentence = %q", got)
	}
	if err := effectstest.ReqBlock(t, g, gp, a); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with one blocker: %v", err)
	}
}

// TestRazorgrassScreenBlocksEachCombat — the requirement is on the
// defender's Wall.
func TestRazorgrassScreenBlocksEachCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.ReqCreature(g, me.ID, "Bear", "", 2, 2, "haste")
	screen := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Razorgrass Screen", TypeLine: "Artifact Creature — Wall",
		OracleID: razorgrassScreenOracle, Power: 2, Toughness: 1, Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.ReqToBlockers(t, g, bear)
	br := effectstest.ReqRefusal(t, g.PassPriority())
	if got := br.Sentence(opp.ID); got != "Razorgrass Screen must block this combat if able." {
		t.Errorf("sentence = %q", got)
	}
	if err := effectstest.ReqBlock(t, g, bear, screen); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with the Screen blocking: %v", err)
	}
}

// TestIrresistiblePreyMakesTheTargetBeBlockedAndDraws — the requirement
// rides a record on the target for the turn, and the card draws.
func TestIrresistiblePreyMakesTheTargetBeBlockedAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.ReqCreature(g, me.ID, "Bear", "", 2, 2, "haste")
	a := effectstest.ReqCreature(g, opp.ID, "Wall A", "", 0, 4)
	effectstest.CastCatalogSpell(t, g, "Irresistible Prey", "Sorcery", irresistiblePreyOracle, effectstest.CardTarget(bear))
	hand := len(me.Hand.Cards)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != hand+1 {
		t.Errorf("hand %d after resolution, want %d (draw a card)", got, hand+1)
	}
	effectstest.ReqToBlockers(t, g, bear)
	br := effectstest.ReqRefusal(t, g.PassPriority())
	if got := br.Sentence(opp.ID); got != "Bear must be blocked if able (Irresistible Prey)." {
		t.Errorf("sentence = %q", got)
	}
	if err := effectstest.ReqBlock(t, g, bear, a); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with the Bear blocked: %v", err)
	}
}

// TestTauntingChallengeLuresForATurn — Lure on the target, for the
// turn only: the record carries an end-of-turn duration.
func TestTauntingChallengeLuresForATurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.ReqCreature(g, me.ID, "Bear", "", 2, 2, "haste")
	a := effectstest.ReqCreature(g, opp.ID, "Wall A", "", 0, 4)
	b := effectstest.ReqCreature(g, opp.ID, "Wall B", "", 0, 4)
	effectstest.CastCatalogSpell(t, g, "Taunting Challenge", "Sorcery", tauntingChallengeOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	found := false
	for _, e := range g.ScopedEffects {
		for _, m := range e.Mods {
			if m.Kind == game.ModAddBlockRequirement && m.Text == string(game.BlockRequirementLure) {
				found = true
				if e.Duration.Kind != game.UntilEndOfTurn {
					t.Errorf("duration %v, want until end of turn", e.Duration.Kind)
				}
			}
		}
	}
	if !found {
		t.Fatal("Taunting Challenge registered no Lure record")
	}
	effectstest.ReqToBlockers(t, g, bear)
	effectstest.ReqRefusal(t, g.PassPriority())
	if err := effectstest.ReqBlock(t, g, bear, a, b); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with both walls blocking: %v", err)
	}
}
