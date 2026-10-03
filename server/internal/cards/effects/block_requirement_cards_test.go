package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

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

// reqCreature puts a creature onto the battlefield under `owner` through
// the zone-move path, so the layer pass and the catalog see it.
func reqCreature(g *game.Game, owner uuid.UUID, name, oracle string, p, t int, keywords ...string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Test", OracleID: oracle,
		Power: p, Toughness: t, Owner: owner, Controller: owner, Keywords: keywords,
	})
}

// reqToBlockers declares `attackers` against the next seat, walks into
// declare_blockers and hands the defender priority, so the next
// PassPriority is the defender's. Returns the defender.
//
// #1501: priority is parked while a defender declares, so the hand-over
// is the defenderHoldsPriority shape rather than the active player's
// pass. The defender's pass is the same CR 509.1c checkpoint as
// finish_blocks (game.blockCheckpointLocked); the game package's
// block_requirements tests pin all three completion paths.
func reqToBlockers(t *testing.T, g *game.Game, attackers ...uuid.UUID) *game.Player {
	t.Helper()
	seat := g.Turn.ActiveSeat
	opp := g.Seats[(seat+1)%len(g.Seats)]
	advanceToDeclareAttackersOf(t, g, seat)
	for _, a := range attackers {
		if err := g.DeclareAttacker(a, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	for g.Turn.Step != game.StepDeclareBlockers {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep to declare blockers: %v", err)
		}
	}
	defenderHoldsPriority(t, g, opp.ID)
	return opp
}

// defenderHoldsPriority builds the one shape in which a defender still
// declaring blockers holds priority (#1501: a restore point written
// before priority was parked for the declaration, or a player who became
// a defending player after the step began), so the next PassPriority is
// theirs and runs the declaration's checkpoint. The step must be parked
// for the declaration, which is what entering it with a pending defender
// does now.
func defenderHoldsPriority(t *testing.T, g *game.Game, defender uuid.UUID) {
	t.Helper()
	if g.Turn.PriorityHolder != game.NoPriority {
		t.Fatalf("priority holder %d, want it parked for the declaration", g.Turn.PriorityHolder)
	}
	idx := -1
	for i, s := range g.Seats {
		if s != nil && s.ID == defender {
			idx = i
		}
	}
	if idx < 0 {
		t.Fatalf("no seat for the defender %s", defender)
	}
	g.WithWriteLock(func() { g.Turn.PriorityHolder = idx })
}

// reqRefusal asserts err is a block_requirement refusal and returns it.
func reqRefusal(t *testing.T, err error) *game.BlockRefusedError {
	t.Helper()
	var br *game.BlockRefusedError
	if !errors.As(err, &br) || br.Reason != game.BlockReasonRequirement {
		t.Fatalf("err = %v, want a block_requirement refusal", err)
	}
	return br
}

func reqBlock(t *testing.T, g *game.Game, attacker uuid.UUID, blockers ...uuid.UUID) error {
	t.Helper()
	decls := make([]game.BlockDeclaration, 0, len(blockers))
	for _, b := range blockers {
		decls = append(decls, game.BlockDeclaration{Blocker: b, Attacker: attacker})
	}
	return g.DeclareBlockers(decls)
}

// TestLureMakesEveryCreatureBlockTheEnchantedCreature — three creatures
// able to block the enchanted creature all must, and one sent at the
// other attacker is refused.
func TestLureMakesEveryCreatureBlockTheEnchantedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := reqCreature(g, me.ID, "Bear", "", 2, 2, "haste")
	other := reqCreature(g, me.ID, "Other Bear", "", 2, 2, "haste")
	w1 := reqCreature(g, opp.ID, "Wall One", "", 0, 4)
	w2 := reqCreature(g, opp.ID, "Wall Two", "", 0, 4)
	w3 := reqCreature(g, opp.ID, "Wall Three", "", 0, 4)
	enchant(t, g, "Lure", lureOracle, bear)

	reqToBlockers(t, g, bear, other)
	br := reqRefusal(t, g.PassPriority())
	if got := br.Sentence(opp.ID); got != "Wall One must block Bear if able (Lure)." {
		t.Errorf("sentence = %q", got)
	}
	reqRefusal(t, reqBlock(t, g, other, w1))
	if err := reqBlock(t, g, bear, w1, w2); err != nil {
		t.Fatalf("two walls on the Lure'd Bear: %v", err)
	}
	if br = reqRefusal(t, g.PassPriority()); br.Blocker != w3 {
		t.Errorf("the refusal names %s, want the wall left home", br.Blocker)
	}
	if err := reqBlock(t, g, bear, w3); err != nil {
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushPermanentForTest(g, me.ID, "Grand Melee", grandMeleeOracle, "Enchantment")
	bear := reqCreature(g, me.ID, "Bear", "", 2, 2, "haste")
	a := reqCreature(g, opp.ID, "Their Wall", "", 0, 4)
	b := reqCreature(g, opp.ID, "Their Elf", "", 1, 1)

	advanceToDeclareAttackersOf(t, g, g.Turn.ActiveSeat)
	var re *game.AttackRequirementError
	if err := g.PassPriority(); !errors.As(err, &re) || re.Attacker != bear {
		t.Fatalf("attack half: pass = %v, want a refusal naming the Bear", err)
	}
	reqToBlockers(t, g, bear)
	br := reqRefusal(t, g.PassPriority())
	if got := br.Sentence(opp.ID); got != "Their Wall must block this combat if able (Grand Melee)." {
		t.Errorf("sentence = %q", got)
	}
	if err := reqBlock(t, g, bear, a); err != nil {
		t.Fatal(err)
	}
	reqRefusal(t, g.PassPriority())
	if err := reqBlock(t, g, bear, b); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with both blocking: %v", err)
	}
}

// TestPrizedUnicornLuresEveryAbleBlocker — Lure printed on the attacker.
func TestPrizedUnicornLuresEveryAbleBlocker(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	unicorn := reqCreature(g, me.ID, "Prized Unicorn", prizedUnicornOracle, 2, 2, "haste")
	a := reqCreature(g, opp.ID, "Wall A", "", 0, 4)
	b := reqCreature(g, opp.ID, "Wall B", "", 0, 4)
	reqToBlockers(t, g, unicorn)
	br := reqRefusal(t, g.PassPriority())
	if got := br.Sentence(opp.ID); got != "Wall A must block Prized Unicorn if able." {
		t.Errorf("sentence = %q", got)
	}
	if err := reqBlock(t, g, unicorn, a, b); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with both blocking: %v", err)
	}
}

// TestGaeasProtectorMustBeBlocked — one blocker is enough; none is not.
func TestGaeasProtectorMustBeBlocked(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	gp := reqCreature(g, me.ID, "Gaea's Protector", gaeasProtectorOracle, 4, 2, "haste")
	a := reqCreature(g, opp.ID, "Wall A", "", 0, 4)
	reqCreature(g, opp.ID, "Wall B", "", 0, 4)
	reqToBlockers(t, g, gp)
	br := reqRefusal(t, g.PassPriority())
	if got := br.Sentence(opp.ID); got != "Gaea's Protector must be blocked if able." {
		t.Errorf("sentence = %q", got)
	}
	if err := reqBlock(t, g, gp, a); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with one blocker: %v", err)
	}
}

// TestRazorgrassScreenBlocksEachCombat — the requirement is on the
// defender's Wall.
func TestRazorgrassScreenBlocksEachCombat(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := reqCreature(g, me.ID, "Bear", "", 2, 2, "haste")
	screen := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Razorgrass Screen", TypeLine: "Artifact Creature — Wall",
		OracleID: razorgrassScreenOracle, Power: 2, Toughness: 1, Owner: opp.ID, Controller: opp.ID,
	})
	reqToBlockers(t, g, bear)
	br := reqRefusal(t, g.PassPriority())
	if got := br.Sentence(opp.ID); got != "Razorgrass Screen must block this combat if able." {
		t.Errorf("sentence = %q", got)
	}
	if err := reqBlock(t, g, bear, screen); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with the Screen blocking: %v", err)
	}
}

// TestIrresistiblePreyMakesTheTargetBeBlockedAndDraws — the requirement
// rides a record on the target for the turn, and the card draws.
func TestIrresistiblePreyMakesTheTargetBeBlockedAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := reqCreature(g, me.ID, "Bear", "", 2, 2, "haste")
	a := reqCreature(g, opp.ID, "Wall A", "", 0, 4)
	castCatalogSpell(t, g, "Irresistible Prey", "Sorcery", irresistiblePreyOracle, cardTarget(bear))
	hand := len(me.Hand.Cards)
	passPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != hand+1 {
		t.Errorf("hand %d after resolution, want %d (draw a card)", got, hand+1)
	}
	reqToBlockers(t, g, bear)
	br := reqRefusal(t, g.PassPriority())
	if got := br.Sentence(opp.ID); got != "Bear must be blocked if able (Irresistible Prey)." {
		t.Errorf("sentence = %q", got)
	}
	if err := reqBlock(t, g, bear, a); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with the Bear blocked: %v", err)
	}
}

// TestTauntingChallengeLuresForATurn — Lure on the target, for the
// turn only: the record carries an end-of-turn duration.
func TestTauntingChallengeLuresForATurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := reqCreature(g, me.ID, "Bear", "", 2, 2, "haste")
	a := reqCreature(g, opp.ID, "Wall A", "", 0, 4)
	b := reqCreature(g, opp.ID, "Wall B", "", 0, 4)
	castCatalogSpell(t, g, "Taunting Challenge", "Sorcery", tauntingChallengeOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
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
	reqToBlockers(t, g, bear)
	reqRefusal(t, g.PassPriority())
	if err := reqBlock(t, g, bear, a, b); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with both walls blocking: %v", err)
	}
}
