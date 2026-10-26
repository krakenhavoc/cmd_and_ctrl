package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// closure_doors_token_restore_test.go — #4253 slice 6, the token half: the
// "then" of a token creation is a registered tail, so the CR 616 order
// prompt a Doubling Season beside a Parallel Lives asks is a restore point.
// Each test takes a STRICT restore while the prompt is open, answers it on
// the restored table and checks the card still finishes off the disk.
// (The helpers libTop and tokenOrderBoard are in
// closure_doors_restore_test.go.)

// orderPromptOpen reports whether the one open prompt is a CR 616 order.
func orderPromptOpen(g *game.Game) bool {
	return len(g.PendingChoices) == 1 && g.PendingChoices[0].Kind == game.PendingChoiceReplacementOrder
}

// Mardu Charm's two Warriors under a Doubling Season and a Parallel Lives:
// the order prompt is a restore point, and the restored answer gives every
// Warrior made (eight) first strike (then/created-gain-keyword-until-eot,
// the keyword riding the tail as its Word).
func TestMarduCharmIsARestorePointAtTheTokenOrderPrompt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ds, pl := tokenOrderBoard(g, me.ID)
	pbModal(t, g, "Mardu Charm", "Instant", marduCharmOracle, 1)
	if !orderPromptOpen(g) {
		t.Fatalf("no order prompt is open: %+v", g.PendingChoices)
	}

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeCreateTokens)
	if n := onBattlefieldNamed(g, "Warrior"); n != 0 {
		t.Fatalf("the restored table made %d Warriors before the order", n)
	}
	answerOrderBySource(t, g, ds, pl)
	warriors := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Warrior" {
			continue
		}
		warriors++
		if !hasKeywordOnBattlefield(t, g, c.InstanceID, "first strike") {
			t.Errorf("a Warrior lacks first strike: the restored tail did not run on every token")
		}
	}
	if warriors != 8 {
		t.Errorf("%d Warriors, want two doubled twice", warriors)
	}
}

// Case of the Stashed Skeleton: the Skeletons are suspected once the order
// prompt of their creation is answered (then/suspect-each over tokens).
func TestCaseOfTheStashedSkeletonIsARestorePointAtTheTokenOrderPrompt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ds, pl := tokenOrderBoard(g, me.ID)
	castCatalogSpell(t, g, "Case of the Stashed Skeleton", "Enchantment — Case", sbStashedSkeleton, nil)
	passPriorityAroundTable(t, g)
	if !orderPromptOpen(g) {
		t.Fatalf("no order prompt is open: %+v", g.PendingChoices)
	}

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeCreateTokens)
	answerOrderBySource(t, g, ds, pl)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Skeleton" {
			continue
		}
		n++
		if !c.Suspected {
			t.Errorf("a Skeleton token is not suspected: the restored tail did not run on every token")
		}
	}
	if n != 4 {
		t.Errorf("%d Skeletons, want one doubled twice", n)
	}
}

// Batterskull's living weapon: the Germ's creation order prompt is a restore
// point, and the restored answer attaches the Equipment to the first Germ
// made (token-and-attach-source/attach, the asking item riding the tail), so
// it survives the state-based check that kills the other three 0/0s.
func TestLivingWeaponIsARestorePointAtTheTokenOrderPrompt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ds, pl := tokenOrderBoard(g, me.ID)
	skull := castCatalogSpell(t, g, "Batterskull", equipTypeLine, batterskullOracle, nil)
	passPriorityAroundTable(t, g)
	if !orderPromptOpen(g) {
		t.Fatalf("no order prompt is open: %+v", g.PendingChoices)
	}

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeCreateTokens)
	answerOrderBySource(t, g, ds, pl)
	passPriorityAroundTable(t, g)
	host := attachmentHostOf(t, g, skull)
	if host.Kind != game.TargetCard || !g.Battlefield.Contains(host.ID) {
		t.Fatalf("Batterskull is attached to %+v, want a Germ on the battlefield", host)
	}
	if got := effectivePower(t, g, host.ID); got != 4 {
		t.Errorf("the Germ is %d power, want the Equipment's +4", got)
	}
}

// Glissa's "incubate 2 twice": each round's creation asks for the order, and
// each is a restore point. The first answer makes the first round's
// Incubators and the tail (incubate/next-round, the rounds left packed in its
// N) starts the second round, whose prompt is a restore point too.
func TestIncubateTwiceIsARestorePointAtEachRoundsOrderPrompt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ds, pl := tokenOrderBoard(g, me.ID)
	pick := glissaAtCombat(t, g, me)
	if err := g.ResolveModePick(pick.ID, me.ID, []int{0}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !orderPromptOpen(g) {
		t.Fatalf("round one: no order prompt is open: %+v", g.PendingChoices)
	}

	for round := 1; round <= 2; round++ {
		g = strictRestoreWithFrame(t, g, game.ReplacementResumeCreateTokens)
		if got := len(incubatorsOf(g, me.ID)); got != (round-1)*4 {
			t.Fatalf("round %d: %d Incubators before the order, want %d", round, got, (round-1)*4)
		}
		answerOrderBySource(t, g, ds, pl)
		if round == 1 && !orderPromptOpen(g) {
			t.Fatalf("the restored tail did not start the second round: %+v", g.PendingChoices)
		}
	}
	if got := len(incubatorsOf(g, me.ID)); got != 8 {
		t.Errorf("%d Incubators, want two rounds of one doubled twice", got)
	}
}

// Widespread Brutality amasses with the Army's creation held on the order
// prompt: the prompt is a restore point, and the restored answer finds
// Armies to choose from, puts the counters on the chosen one and runs the
// card's own tail with it (engine/amass-created, with the sweep as its Next).
func TestAmassIsARestorePointAtTheArmysOrderPrompt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	ds, pl := tokenOrderBoard(g, me.ID)
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Grizzly Bears", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, PrintedPTKnown: true, Owner: opp.ID, Controller: opp.ID,
	})
	castCatalogSpell(t, g, "Widespread Brutality", "Sorcery", widespreadBrutalityOracl, nil)
	passPriorityAroundTable(t, g)
	if !orderPromptOpen(g) {
		t.Fatalf("no order prompt is open: %+v", g.PendingChoices)
	}

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeCreateTokens)
	answerOrderBySource(t, g, ds, pl)
	var armies []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if game.IsArmy(c) {
			armies = append(armies, c.InstanceID)
		}
	}
	if len(armies) != 4 {
		t.Fatalf("%d Armies after the restored answer, want one doubled twice", len(armies))
	}
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("the restored tail did not ask which Army to amass onto")
	}
	answerChooseCards(t, g, me.ID, armies[0])
	passPriorityAroundTable(t, g)
	if got := countersOn(g, armies[0], game.CounterPlusOne); got < 2 {
		t.Errorf("the chosen Army has %d counters, want the amass's", got)
	}
	if findBattlefieldCardByID(g, bear) != nil {
		t.Error("the restored tail did not run the sweep: the 2/2 Bear survived the amassed Army")
	}
}

// Protege's Awakening empowers Jace with the token's creation held on the
// order prompt: the restored answer places the counters on the chosen Jace
// and then draws (empower-jace/place-counters, the draw its Next).
func TestEmpowerJaceIsARestorePointAtTheTokenOrderPrompt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ds, pl := tokenOrderBoard(g, me.ID)
	before := handSize(me)
	castProtegesAwakening(t, g)
	passPriorityAroundTable(t, g)
	if !orderPromptOpen(g) {
		t.Fatalf("no order prompt is open: %+v", g.PendingChoices)
	}

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeCreateTokens)
	me = g.Seats[g.Turn.ActiveSeat]
	answerOrderBySource(t, g, ds, pl)
	jaces := jaceTokensOf(g, me.ID)
	if len(jaces) != 4 {
		t.Fatalf("%d Jace tokens after the restored answer, want one doubled twice", len(jaces))
	}
	if handSize(me) != before {
		t.Fatal("the restored tail drew before the Jace was chosen")
	}
	answerOwnPermanents(t, g, me.ID, jaces[0])
	passPriorityAroundTable(t, g)
	if got := loyaltyCount(g, jaces[0]); got < 6 {
		t.Errorf("the chosen Jace has %d loyalty, want the six counters", got)
	}
	if got := handSize(me); got != before+1 {
		t.Errorf("hand %d after the answer, want the card Protege's Awakening draws (%d)", got, before+1)
	}
}

// A Role token (Monstrous Rage) made under the order prompt: the restored
// answer attaches the Role to the host that rode the tail as its Carry
// (role-tokens/attach-to-host), and only the newest of the same player's
// Roles stays (CR 704.5z).
func TestARoleTokenIsARestorePointAtTheTokenOrderPrompt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ds, pl := tokenOrderBoard(g, me.ID)
	bear := pushCreatureToBattlefieldForTest(g, me.ID, "Bear")
	castCatalogSpell(t, g, "Monstrous Rage", "Instant", monstrousRageOracle, []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if !orderPromptOpen(g) {
		t.Fatalf("no order prompt is open: %+v", g.PendingChoices)
	}

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeCreateTokens)
	answerOrderBySource(t, g, ds, pl)
	passPriorityAroundTable(t, g)
	roles := rolesOn(g, bear)
	if len(roles) != 1 || roles[0].Name != "Monster Role" {
		t.Fatalf("roles on the bear = %+v, want the one Monster Role the restored tail attached", roles)
	}
}

// Double Major copies a creature spell under the order prompt: the copy's
// token creation is a restore point, and the restored answer makes four
// token copies beside the original (engine/spell-copy-token).
func TestACopiedPermanentSpellIsARestorePointAtTheTokenOrderPrompt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ds, pl := tokenOrderBoard(g, me.ID)
	spell := castCreatureSpellForCopy(t, g, "Grizzly Bears", "Creature — Bear", "oracle-bears", 2, 2)
	castCatalogSpell(t, g, "Double Major", "Instant", oracleDoubleMajor,
		[]game.TargetRef{{Kind: game.TargetCard, ID: spell}})
	passPriorityAroundTable(t, g)
	if !orderPromptOpen(g) {
		t.Fatalf("no order prompt is open: %+v", g.PendingChoices)
	}

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeCreateTokens)
	answerOrderBySource(t, g, ds, pl)
	passPriorityAroundTable(t, g)
	tokens, cards := copiesOnBattlefield(g, "Grizzly Bears")
	if len(tokens) != 4 || len(cards) != 1 {
		t.Errorf("%d token copies and %d cards after the restored answer, want four and the original", len(tokens), len(cards))
	}
}
