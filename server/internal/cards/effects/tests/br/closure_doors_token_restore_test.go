package testsbr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mardu Charm's two Warriors under a Doubling Season and a Parallel Lives:
// the order prompt is a restore point, and the restored answer gives every
// Warrior made (eight) first strike (then/created-gain-keyword-until-eot,
// the keyword riding the tail as its Word).
func TestMarduCharmIsARestorePointAtTheTokenOrderPrompt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ds, pl := effectstest.TokenOrderBoard(g, me.ID)
	effectstest.PbModal(t, g, "Mardu Charm", "Instant", effectstest.MarduCharmOracle, 1)
	if !effectstest.OrderPromptOpen(g) {
		t.Fatalf("no order prompt is open: %+v", g.PendingChoices)
	}

	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeCreateTokens)
	if n := effectstest.OnBattlefieldNamed(g, "Warrior"); n != 0 {
		t.Fatalf("the restored table made %d Warriors before the order", n)
	}
	effectstest.AnswerOrderBySource(t, g, ds, pl)
	warriors := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Warrior" {
			continue
		}
		warriors++
		if !effectstest.HasKeywordOnBattlefield(t, g, c.InstanceID, "first strike") {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ds, pl := effectstest.TokenOrderBoard(g, me.ID)
	effectstest.CastCatalogSpell(t, g, "Case of the Stashed Skeleton", "Enchantment — Case", effectstest.SbStashedSkeleton, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.OrderPromptOpen(g) {
		t.Fatalf("no order prompt is open: %+v", g.PendingChoices)
	}

	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeCreateTokens)
	effectstest.AnswerOrderBySource(t, g, ds, pl)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ds, pl := effectstest.TokenOrderBoard(g, me.ID)
	skull := effectstest.CastCatalogSpell(t, g, "Batterskull", effectstest.EquipTypeLine, effectstest.BatterskullOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.OrderPromptOpen(g) {
		t.Fatalf("no order prompt is open: %+v", g.PendingChoices)
	}

	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeCreateTokens)
	effectstest.AnswerOrderBySource(t, g, ds, pl)
	effectstest.PassPriorityAroundTable(t, g)
	host := effectstest.AttachmentHostOf(t, g, skull)
	if host.Kind != game.TargetCard || !g.Battlefield.Contains(host.ID) {
		t.Fatalf("Batterskull is attached to %+v, want a Germ on the battlefield", host)
	}
	if got := effectstest.EffectivePower(t, g, host.ID); got != 4 {
		t.Errorf("the Germ is %d power, want the Equipment's +4", got)
	}
}

// Glissa's "incubate 2 twice": each round's creation asks for the order, and
// each is a restore point. The first answer makes the first round's
// Incubators and the tail (incubate/next-round, the rounds left packed in its
// N) starts the second round, whose prompt is a restore point too.
func TestIncubateTwiceIsARestorePointAtEachRoundsOrderPrompt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ds, pl := effectstest.TokenOrderBoard(g, me.ID)
	pick := effectstest.GlissaAtCombat(t, g, me)
	if err := g.ResolveModePick(pick.ID, me.ID, []int{0}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.OrderPromptOpen(g) {
		t.Fatalf("round one: no order prompt is open: %+v", g.PendingChoices)
	}

	for round := 1; round <= 2; round++ {
		g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeCreateTokens)
		if got := len(effectstest.IncubatorsOf(g, me.ID)); got != (round-1)*4 {
			t.Fatalf("round %d: %d Incubators before the order, want %d", round, got, (round-1)*4)
		}
		effectstest.AnswerOrderBySource(t, g, ds, pl)
		if round == 1 && !effectstest.OrderPromptOpen(g) {
			t.Fatalf("the restored tail did not start the second round: %+v", g.PendingChoices)
		}
	}
	if got := len(effectstest.IncubatorsOf(g, me.ID)); got != 8 {
		t.Errorf("%d Incubators, want two rounds of one doubled twice", got)
	}
}

// Widespread Brutality amasses with the Army's creation held on the order
// prompt: the prompt is a restore point, and the restored answer finds
// Armies to choose from, puts the counters on the chosen one and runs the
// card's own tail with it (engine/amass-created, with the sweep as its Next).
func TestAmassIsARestorePointAtTheArmysOrderPrompt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	ds, pl := effectstest.TokenOrderBoard(g, me.ID)
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Grizzly Bears", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, PrintedPTKnown: true, Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.CastCatalogSpell(t, g, "Widespread Brutality", "Sorcery", effectstest.WidespreadBrutalityOracl, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.OrderPromptOpen(g) {
		t.Fatalf("no order prompt is open: %+v", g.PendingChoices)
	}

	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeCreateTokens)
	effectstest.AnswerOrderBySource(t, g, ds, pl)
	var armies []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if game.IsArmy(c) {
			armies = append(armies, c.InstanceID)
		}
	}
	if len(armies) != 4 {
		t.Fatalf("%d Armies after the restored answer, want one doubled twice", len(armies))
	}
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("the restored tail did not ask which Army to amass onto")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, armies[0])
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, armies[0], game.CounterPlusOne); got < 2 {
		t.Errorf("the chosen Army has %d counters, want the amass's", got)
	}
	if effectstest.FindBattlefieldCardByID(g, bear) != nil {
		t.Error("the restored tail did not run the sweep: the 2/2 Bear survived the amassed Army")
	}
}

// Protege's Awakening empowers Jace with the token's creation held on the
// order prompt: the restored answer places the counters on the chosen Jace
// and then draws (empower-jace/place-counters, the draw its Next).
func TestEmpowerJaceIsARestorePointAtTheTokenOrderPrompt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ds, pl := effectstest.TokenOrderBoard(g, me.ID)
	before := effectstest.HandSize(me)
	effectstest.CastProtegesAwakening(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.OrderPromptOpen(g) {
		t.Fatalf("no order prompt is open: %+v", g.PendingChoices)
	}

	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeCreateTokens)
	me = g.Seats[g.Turn.ActiveSeat]
	effectstest.AnswerOrderBySource(t, g, ds, pl)
	jaces := effectstest.JaceTokensOf(g, me.ID)
	if len(jaces) != 4 {
		t.Fatalf("%d Jace tokens after the restored answer, want one doubled twice", len(jaces))
	}
	if effectstest.HandSize(me) != before {
		t.Fatal("the restored tail drew before the Jace was chosen")
	}
	effectstest.AnswerOwnPermanents(t, g, me.ID, jaces[0])
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LoyaltyCount(g, jaces[0]); got < 6 {
		t.Errorf("the chosen Jace has %d loyalty, want the six counters", got)
	}
	if got := effectstest.HandSize(me); got != before+1 {
		t.Errorf("hand %d after the answer, want the card Protege's Awakening draws (%d)", got, before+1)
	}
}

// A Role token (Monstrous Rage) made under the order prompt: the restored
// answer attaches the Role to the host that rode the tail as its Carry
// (role-tokens/attach-to-host), and only the newest of the same player's
// Roles stays (CR 704.5z).
func TestARoleTokenIsARestorePointAtTheTokenOrderPrompt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ds, pl := effectstest.TokenOrderBoard(g, me.ID)
	bear := effectstest.PushCreatureToBattlefieldForTest(g, me.ID, "Bear")
	effectstest.CastCatalogSpell(t, g, "Monstrous Rage", "Instant", effectstest.MonstrousRageOracle, []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.OrderPromptOpen(g) {
		t.Fatalf("no order prompt is open: %+v", g.PendingChoices)
	}

	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeCreateTokens)
	effectstest.AnswerOrderBySource(t, g, ds, pl)
	effectstest.PassPriorityAroundTable(t, g)
	roles := effectstest.RolesOn(g, bear)
	if len(roles) != 1 || roles[0].Name != "Monster Role" {
		t.Fatalf("roles on the bear = %+v, want the one Monster Role the restored tail attached", roles)
	}
}

// Double Major copies a creature spell under the token's entry order
// prompt: the copy's token entry is a restore point, and the restored
// answer puts exactly one token copy beside the original
// (engine/spell-copy-token). Doubling Season and Parallel Lives are on the
// board and do nothing to it, and it is still not "created" once restored
// (CR 111.13, #4994): the run's NotCreated rode the disk.
func TestACopiedPermanentSpellIsARestorePointAtTheTokenOrderPrompt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.TokenOrderBoard(g, me.ID)
	pa, pb := effectstest.SeedTokenEntryOrderProbes(g, me.ID)
	spell := effectstest.CastCreatureSpellForCopy(t, g, "Grizzly Bears", "Creature — Bear", "oracle-bears", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Double Major", "Instant", effectstest.OracleDoubleMajor,
		[]game.TargetRef{{Kind: game.TargetCard, ID: spell}})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.OrderPromptOpen(g) {
		t.Fatalf("no order prompt is open: %+v", g.PendingChoices)
	}

	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeMoveTail)
	if tokens, _ := effectstest.CopiesOnBattlefield(g, "Grizzly Bears"); len(tokens) != 0 {
		t.Fatalf("the restored table landed %d tokens before the order", len(tokens))
	}
	effectstest.AnswerOrderBySource(t, g, pa, pb)
	effectstest.PassPriorityAroundTable(t, g)
	tokens, cards := effectstest.CopiesOnBattlefield(g, "Grizzly Bears")
	if len(tokens) != 1 || len(cards) != 1 {
		t.Fatalf("%d token copies and %d cards after the restored answer, want one of each: the copy's token is not created, so nothing doubles it (CR 111.13)", len(tokens), len(cards))
	}
	if n := effectstest.TokenCreatedEventsFor(g, tokens[0].InstanceID); n != 0 {
		t.Errorf("the restored token was announced as created %d times, want none (CR 111.13)", n)
	}
}
