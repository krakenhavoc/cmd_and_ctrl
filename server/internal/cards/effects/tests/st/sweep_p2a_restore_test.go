package testsst

import (
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sweep_p2a_restore_test.go — #4253's second pass, group A: the shared
// primitives. A resolution-time colour and creature-type prompt hand their
// answer to a registered tail (ChooseColorTail, ChooseCreatureTypeTail); the
// land-type choice, the text-word change, the reveal pick and populate's
// "except it has haste" are data. Each test takes a STRICT restore while the
// prompt is open (restoreAtKeyedPrompt checks the snapshot names the key and
// counts no closure frame) and answers it on the restored table.

// The Selective Obliteration chain is a restore point between two players'
// colours: the answers so far ride the tail, and the exile on the restored
// table reads them.
func TestSelectiveObliterationMidChainIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seats := g.Seats
	whiteMine := effectstest.B31Push(g, seats[0].ID, "Savannah Lions", "Creature — Cat", "", "{W}", 2, 1, "W")
	azorius := effectstest.B31Push(g, seats[0].ID, "Azorius Guildmage", "Creature — Vedalken Wizard", "", "{W/U}{W/U}", 2, 2, "W", "U")
	greenTheirs := effectstest.B31Push(g, seats[1].ID, "Grizzly Bears", "Creature — Bear", "", "{1}{G}", 2, 2, "G")
	redTheirs := effectstest.B31Push(g, seats[3].ID, "Goblin Guide", "Creature — Goblin Scout", "", "{R}", 2, 2, "R")

	effectstest.CastCatalogSpell(t, g, "Selective Obliteration", "Sorcery", "9f95027a-5f04-48b4-99a0-8913be3a520a", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerColor(t, g, seats[0].ID, "W")

	g = effectstest.RestoreAtKeyedPrompt(t, g, "then/selective-obliteration-color-chosen")
	seats = g.Seats
	effectstest.AnswerColor(t, g, seats[1].ID, "U")
	effectstest.AnswerColor(t, g, seats[2].ID, "B")
	effectstest.AnswerColor(t, g, seats[3].ID, "R")

	for id, wantKept := range map[uuid.UUID]bool{whiteMine: true, azorius: false, greenTheirs: false, redTheirs: true} {
		if _, ok := effectstest.BattlefieldCard(g, id); ok != wantKept {
			t.Errorf("%s kept = %v, want %v", id, ok, wantKept)
		}
	}
}

// Oona's colour is a restore point: X and the opponent ride the tail.
func TestOonaColorPromptIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	effectstest.AangAdvanceToMain(t, g, 0)
	oona := effectstest.B31Push(g, me.ID, "Oona, Queen of the Fae", "Legendary Creature — Faerie Wizard",
		"6052822d-47a2-4d69-a32d-40cdd600d7a9", "{3}{U/B}{U/B}{U/B}", 5, 5, "U", "B")
	them.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Island", TypeLine: "Basic Land — Island", Owner: them.ID, Controller: them.ID})
	them.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Counterspell", TypeLine: "Instant", ManaCost: "{U}{U}", Owner: them.ID, Controller: them.ID})
	them.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Brainstorm", TypeLine: "Instant", ManaCost: "{U}", Owner: them.ID, Controller: them.ID})
	effectstest.B31AddMana(me, "U", "U", "U")
	if err := g.ActivateCatalogAbility(me.ID, oona, 0, game.ActivateAbilityParams{
		XValue:  2,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: them.ID}},
	}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	g = effectstest.RestoreAtKeyedPrompt(t, g, "then/oona-queen-of-the-fae-color-chosen")
	me, them = g.Seats[0], g.Seats[1]
	libBefore := them.Library.Size()
	effectstest.AnswerColor(t, g, me.ID, "U")
	if got := libBefore - them.Library.Size(); got != 2 {
		t.Errorf("exiled %d cards after the restored answer, want X=2", got)
	}
	faeries := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Faerie Rogue" && c.Controller == me.ID {
			faeries++
		}
	}
	if faeries != 2 {
		t.Errorf("Faerie Rogues = %d, want 2 (two blue cards)", faeries)
	}
}

// Center Soul's colour rides the shared protection tail: the restored
// answer grants protection to the target.
func TestCenterSoulColorPromptIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Grizzly Bears", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Owner: me.ID, Controller: me.ID,
	})
	effectstest.CastCatalogSpell(t, g, "Center Soul", "Instant", "686b44ec-3446-4e1f-a15f-9d8557db6d70", []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)

	g = effectstest.RestoreAtKeyedPrompt(t, g, "then/protection-from-chosen-color-until-eot")
	effectstest.AnswerColor(t, g, g.Seats[g.Turn.ActiveSeat].ID, "R")
	if got := effectstest.ProtectionsOn(t, g, bear); !effectstest.HasString(got, "red") {
		t.Errorf("the restored answer granted no protection from red: %q", got)
	}
}

// Kindred Dominance's creature type is a restore point: the restored answer
// destroys every creature not of it.
func TestKindredDominanceTypePromptIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elf := effectstest.PushTribalCreature(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", 1, 1)
	goblin := effectstest.PushTribalCreature(g, opp.ID, "Goblin Guide", "Creature — Goblin Scout", 2, 2)

	effectstest.CastAndPause(t, g, "Kindred Dominance", "Sorcery", effectstest.KindredDominanceOracle)
	g = effectstest.RestoreAtKeyedPrompt(t, g, "then/kindred-dominance-type-chosen")
	effectstest.AnswerCreatureType(t, g, g.Seats[0].ID, "Elf")
	if !effectstest.BattlefieldHas(g, elf) || effectstest.BattlefieldHas(g, goblin) {
		t.Errorf("after the restored answer: Elf kept %v (want true), Goblin kept %v (want false)",
			effectstest.BattlefieldHas(g, elf), effectstest.BattlefieldHas(g, goblin))
	}
}

// Patriarch's Bidding is a restore point between seats: the seats still to
// ask and the types so far ride the tail.
func TestPatriarchsBiddingMidChainIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	s := g.Seats
	effectstest.PushGraveyardCardTyped(s[0], "Dead Elf", "Creature — Elf")
	effectstest.PushGraveyardCardTyped(s[1], "Dead Goblin", "Creature — Goblin")
	effectstest.PushGraveyardCardTyped(s[3], "Dead Bear", "Creature — Bear")

	effectstest.CastAndPause(t, g, "Patriarch's Bidding", "Sorcery", effectstest.PatriarchsBiddingOracle)
	effectstest.AnswerCreatureType(t, g, s[0].ID, "Elf")
	g = effectstest.RestoreAtKeyedPrompt(t, g, "then/patriarchs-bidding-type-chosen")
	s = g.Seats
	for i, tribe := range []string{"Goblin", "Sliver", "Sliver"} {
		c := effectstest.PendingOfKind(g, game.PendingChoiceCreatureType)
		if c == nil || c.Chooser != s[i+1].ID {
			t.Fatalf("prompt %d after the restore: want one owed by seat %d, got %+v", i, i+1, c)
		}
		effectstest.AnswerCreatureType(t, g, s[i+1].ID, tribe)
	}
	for name, want := range map[string]bool{"Dead Elf": true, "Dead Goblin": true, "Dead Bear": false} {
		if got := effectstest.BattlefieldHasName(g, name); got != want {
			t.Errorf("%s on the battlefield = %v, want %v", name, got, want)
		}
	}
}

// Shimmering Mirage's land-type pick is a restore point, and its draw (the
// tail's Next) still comes after the restored answer.
func TestShimmeringMirageLandTypePickIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	forest := effectstest.PushLandFor(g, opp.ID, "Forest", "Basic Land — Forest")
	effectstest.CastCatalogSpell(t, g, "Shimmering Mirage", "Instant", effectstest.LtShimmeringMirageOracle, effectstest.LtCardTarget(forest))
	hand := len(me.Hand.Cards)
	effectstest.PassPriorityAroundTable(t, g)

	g = effectstest.RestoreAtKeyedPrompt(t, g, "option-pick/choose-land-type")
	me = g.Seats[0]
	effectstest.AnswerOptionPick(t, g, me.ID, 1)
	if len(me.Hand.Cards) != hand+1 {
		t.Errorf("hand %d after the restored answer, want %d (the draw follows)", len(me.Hand.Cards), hand+1)
	}
	if got := effectstest.EffectiveSubtypes(t, g, forest); !reflect.DeepEqual(got, []string{"Island"}) {
		t.Errorf("%v, want [Island]", got)
	}
}

// Crystal Spray's second word is a restore point: the change and the draw
// after it ride the keyed picks.
func TestCrystalSprayWordPickIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	knight := effectstest.TwPut(g, me.ID, "White Knight", "Creature — Human Knight", []string{"W"}, "protection from black")
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Crystal Spray", "Instant", effectstest.TwSprayOracle, effectstest.TwTargetRef(knight))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, effectstest.TwBlack)

	g = effectstest.RestoreAtKeyedPrompt(t, g, "option-pick/change-text-word-with")
	me = g.Seats[g.Turn.ActiveSeat]
	effectstest.AnswerOptionPick(t, g, me.ID, effectstest.TwGreen-1)
	if me.Hand.Size() != hand+1 {
		t.Errorf("Crystal Spray draws after the restored answer (hand %d → %d)", hand, me.Hand.Size())
	}
	if !effectstest.HasEffectiveKeyword(t, g, knight, "protection from green") {
		t.Error("protection from black does not read protection from green after the restored answer")
	}
}

// Artificial Evolution's second creature type is a restore point, and the
// refused types still are refused on the restored table.
func TestArtificialEvolutionSecondTypeIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	gob := effectstest.TwPut(g, me.ID, "Goblin Piker", "Creature — Goblin", []string{"R"})
	effectstest.CastCatalogSpell(t, g, "Artificial Evolution", "Instant", effectstest.TwEvolutionOracle, effectstest.TwTargetRef(gob))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerCreatureType(t, g, me.ID, "Goblin")

	g = effectstest.RestoreAtKeyedPrompt(t, g, "then/change-creature-type-word-with")
	me = g.Seats[g.Turn.ActiveSeat]
	c := effectstest.PendingOfKind(g, game.PendingChoiceCreatureType)
	if c == nil {
		t.Fatal("no creature-type prompt after the restore")
	}
	if err := g.ResolveCreatureTypeChoice(c.ID, me.ID, "Wall"); err == nil {
		t.Error("the restored prompt accepted Wall, a refused type")
	}
	effectstest.AnswerCreatureType(t, g, me.ID, "Elf")
	if subs := effectstest.EffectiveSubtypes(t, g, gob); !reflect.DeepEqual(subs, []string{"Elf"}) {
		t.Errorf("the Goblin is a %v after the restored answer, want [Elf]", subs)
	}
}

// Thassa's Oracle's pick is a restore point: X rides the prompt, and the
// restored answer still wins.
func TestThassasOraclePickIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for me.Library.Size() > 2 {
		_, _ = me.Library.PopTop()
	}
	oracle := effectstest.CastThassasOracle(t, g, "{U}{U}")

	g = effectstest.RestoreAtKeyedPrompt(t, g, "choose-cards/thassas-oracle-arranged")
	me = g.Seats[0]
	pick := effectstest.LatestRevealPickFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick prompt after the restore")
	}
	if err := g.ResolveRevealPick(pick.ID, me.ID, pick.ChooseCards[1:2]); err != nil {
		t.Fatalf("ResolveRevealPick: %v", err)
	}
	want := game.GameOutcome{Kind: game.OutcomeWin, Winner: me.ID, Cause: game.OutcomeCauseEffect, Source: oracle}
	if g.Outcome == nil || *g.Outcome != want {
		t.Fatalf("outcome %+v after the restored answer, want %+v", g.Outcome, want)
	}
}

// Determined Iteration's populate pick is a restore point: "except it has
// haste" rides the pick as its word, and the sacrifice as its tail.
func TestDeterminedIterationPopulatePickIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushPermanentForTest(g, me.ID, "Determined Iteration", effectstest.DeterminedIterationOracle, "Enchantment")
	centaur := effectstest.PushToken(g, me.ID, effects.TokenCard("3/3 green Centaur"))
	effectstest.PushToken(g, me.ID, effects.TokenCard("1/1 white Soldier"))

	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)

	g = effectstest.RestoreAtKeyedPrompt(t, g, "choose-cards/populate")
	me = g.Seats[0]
	c := effectstest.PendingOfKind(g, game.PendingChoiceChooseCards)
	if c == nil {
		t.Fatal("no populate pick after the restore")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{centaur}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	var copyID uuid.UUID
	for _, id := range effectstest.PopulateTokens(g, me.ID, "Centaur") {
		if id != centaur {
			copyID = id
		}
	}
	if copyID == uuid.Nil {
		t.Fatal("no copy was made after the restored answer")
	}
	if cp, _ := g.LookupCardForEffect(copyID); !effectstest.HasKeywordForTest(cp.Keywords, "haste") {
		t.Errorf("the restored copy has no haste: %v", cp.Keywords)
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(copyID) {
		t.Error("the copy was not sacrificed at the end step")
	}
}
