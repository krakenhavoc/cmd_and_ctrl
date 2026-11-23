package testso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// otj_uz_cards_test.go — Outlaws of Thunder Junction (OTJ), the U–Z cards
// (#4468). Each test plays the real catalog entry through the engine and
// asserts what the printed text says.

const (
	otjUZUnfortunateAccidentOracle = "011558f3-3c23-430e-a08b-8f91238e7971"
	otjUZContractorOracle          = "00ff5598-2f7f-4e28-8b63-5ae1ce18f0c5"
	otjUZVadmirOracle              = "b000bcf5-529a-447a-bc2e-ae6f42d6073b"
	otjUZVaultPlundererOracle      = "80e5d768-cc1f-48f8-b0fe-ed974e4d1a56"
	otjUZTownsfolkOracle           = "c649d5ae-2f38-4737-8123-8069a2ba0bde"
	otjUZVialSmasherOracle         = "a66d8dc6-ced4-4b24-b0c2-3557b2a851be"
	otjUZVisageBanditOracle        = "db11596a-6704-427b-aaef-0211e7822127"
	otjUZVarmintOracle             = "bd785a7e-df65-48e6-843b-7831f3c1e5d3"
	otjUZWantedGriffinOracle       = "c1348afe-4dc7-41bb-9d3f-1e8751abd7db"
	otjUZWylieDukeOracle           = "b3ac27d6-7e18-404d-9af2-b03897a5f1d9"
)

func TestOtjUZEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		otjUZUnfortunateAccidentOracle: "Unfortunate Accident",
		otjUZContractorOracle:          "Unscrupulous Contractor",
		otjUZVadmirOracle:              "Vadmir, New Blood",
		otjUZVaultPlundererOracle:      "Vault Plunderer",
		otjUZTownsfolkOracle:           "Vengeful Townsfolk",
		otjUZVialSmasherOracle:         "Vial Smasher, Gleeful Grenadier",
		otjUZVisageBanditOracle:        "Visage Bandit",
		otjUZVarmintOracle:             "Voracious Varmint",
		otjUZWantedGriffinOracle:       "Wanted Griffin",
		otjUZWylieDukeOracle:           "Wylie Duke, Atiin Hero",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
	for oracle, cost := range map[string]string{
		otjUZContractorOracle:   "{2}{B}",
		otjUZVisageBanditOracle: "{2}{U}",
	} {
		if sa := effectstest.PlotOf(oracle); sa == nil || sa.Cost != cost {
			t.Errorf("%s: plot %+v, want %s", oracle, sa, cost)
		}
	}
}

// --- Unfortunate Accident ----------------------------------------------------

func TestOtjUZUnfortunateAccidentDestroysAndMakesAMercenary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastModal(t, g, "Unfortunate Accident", "Instant", otjUZUnfortunateAccidentOracle,
		[]int{0, 1}, []game.TargetRef{effectstest.ModeRef(game.TargetCard, theirs, 0, 0)})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the targeted creature was not destroyed")
	}
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Mercenary"); n != 1 {
		t.Errorf("%d Mercenary tokens, want 1", n)
	}
}

func TestOtjUZUnfortunateAccidentMercenaryAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastModal(t, g, "Unfortunate Accident", "Instant", otjUZUnfortunateAccidentOracle, []int{1}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(theirs) {
		t.Error("the destroy bullet was not chosen")
	}
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Mercenary"); n != 1 {
		t.Errorf("%d Mercenary tokens, want 1", n)
	}
}

// --- Unscrupulous Contractor -------------------------------------------------

func TestOtjUZUnscrupulousContractorSacrificesForTwoCardsAndTwoLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	contractor := effectstest.CastHoldCreature(t, g, "Unscrupulous Contractor", "Creature — Human Assassin", otjUZContractorOracle, 3, 2)
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if c == nil {
		t.Fatal("no sacrifice prompt")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range c.ChooseCards {
		offered[id] = true
	}
	if !offered[bear] || !offered[contractor] {
		t.Errorf("offered %v: want any creature you control, the Contractor included", c.ChooseCards)
	}
	hand, life := opp.Hand.Size(), opp.Life
	effectstest.AnswerOwnPermanents(t, g, me.ID, bear)
	effectstest.B04WaitForPick(t, g, me.ID)
	if g.Battlefield.Contains(bear) {
		t.Fatal("the Bear was not sacrificed before the reflexive trigger")
	}
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Size() != hand+2 || opp.Life != life-2 {
		t.Errorf("target player: hand %d → %d, life %d → %d, want +2 cards and -2 life", hand, opp.Hand.Size(), life, opp.Life)
	}
}

func TestOtjUZUnscrupulousContractorMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastHoldCreature(t, g, "Unscrupulous Contractor", "Creature — Human Assassin", otjUZContractorOracle, 3, 2)
	life := me.Life
	effectstest.AnswerOwnPermanents(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) || me.Life != life {
		t.Error("declining changed the board")
	}
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("no sacrifice, no reflexive trigger")
	}
}

// --- Vadmir, New Blood -------------------------------------------------------

func TestOtjUZVadmirGrowsOnceEachTurnAndGainsKeywordsAtFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vadmir := effectstest.B12Push(g, me.ID, "Vadmir, New Blood", "Legendary Creature — Vampire Rogue", otjUZVadmirOracle, 2, 2)
	effectstest.AdvanceToMain(t, g)
	otjB1BoltOpponent(t, g, opp.ID)
	if got := effectstest.PlusOneCounters(g, vadmir); got != 1 {
		t.Fatalf("after a crime: %d counters, want 1", got)
	}
	otjB1BoltOpponent(t, g, opp.ID)
	if got := effectstest.PlusOneCounters(g, vadmir); got != 1 {
		t.Fatalf("a second crime this turn: %d counters, want still 1", got)
	}
	ctx := effectstest.CtxFor(g, &game.StackItem{Controller: me.ID, SourceCardID: vadmir})
	if err := (effects.AddCounter{Target: vadmir, Kind: game.CounterPlusOne, N: 2}).Apply(ctx); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	for _, kw := range []string{"menace", "lifelink"} {
		if effectstest.HasEffectiveKeyword(t, g, vadmir, kw) {
			t.Errorf("three counters: %s too early", kw)
		}
	}
	if err := (effects.AddCounter{Target: vadmir, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	for _, kw := range []string{"menace", "lifelink"} {
		if !effectstest.HasEffectiveKeyword(t, g, vadmir, kw) {
			t.Errorf("four counters: no %s", kw)
		}
	}
}

// --- Vault Plunderer ---------------------------------------------------------

func TestOtjUZVaultPlundererTargetPlayerDrawsAndLosesOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.CastHoldCreature(t, g, "Vault Plunderer", "Creature — Human Rogue", otjUZVaultPlundererOracle, 3, 1)
	myHand, myLife, hand, life := me.Hand.Size(), me.Life, opp.Hand.Size(), opp.Life
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Size() != hand+1 || opp.Life != life-1 {
		t.Errorf("target player: hand %d → %d, life %d → %d, want +1 and -1", hand, opp.Hand.Size(), life, opp.Life)
	}
	if me.Hand.Size() != myHand || me.Life != myLife {
		t.Error("the controller drew or lost life, not the target")
	}
}

// --- Vengeful Townsfolk ------------------------------------------------------

func TestOtjUZVengefulTownsfolkCountsBatchesOfOtherCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	town := effectstest.B12Push(g, me.ID, "Vengeful Townsfolk", "Creature — Human Citizen", otjUZTownsfolkOracle, 3, 3)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, me.ID, "Bear C", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	ctx := effectstest.CtxFor(g, &game.StackItem{Controller: me.ID})
	for _, id := range []uuid.UUID{a, b} {
		if err := (effects.DestroyTarget{Target: id}).Apply(ctx); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOneCounters(g, town); got != 1 {
		t.Fatalf("two creatures died together: %d counters, want 1", got)
	}
	if err := (effects.DestroyTarget{Target: theirs}).Apply(ctx); err != nil {
		t.Fatalf("destroy: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOneCounters(g, town); got != 1 {
		t.Fatalf("an opponent's creature died: %d counters, want still 1", got)
	}
	if err := (effects.DestroyTarget{Target: c}).Apply(ctx); err != nil {
		t.Fatalf("destroy: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOneCounters(g, town); got != 2 {
		t.Errorf("a later death: %d counters, want 2", got)
	}
}

// --- Vial Smasher, Gleeful Grenadier -----------------------------------------

func TestOtjUZVialSmasherPingsAnOpponentWhenAnotherOutlawEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.CastHoldCreature(t, g, "Vial Smasher, Gleeful Grenadier", "Legendary Creature — Goblin Mercenary", otjUZVialSmasherOracle, 3, 2)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("Vial Smasher triggered on its own entry")
	}
	effectstest.CastHoldCreature(t, g, "Bear", "Creature — Bear", "", 2, 2)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("a Bear is no outlaw")
	}
	life := opp.Life
	effectstest.CastHoldCreature(t, g, "Rogue", "Creature — Human Rogue", "", 1, 1)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("opponent life %d → %d, want -1", life, opp.Life)
	}
}

// --- Visage Bandit -----------------------------------------------------------

func TestOtjUZVisageBanditCopiesYourCreatureAsAShapeshifterRogue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	angel := effectstest.SeedCopyableCreature(g, me.ID, "Serra Angel", "Creature — Angel", 4, 4)
	theirs := effectstest.SeedCopyableCreature(g, opp.ID, "Their Dragon", "Creature — Dragon", 5, 5)
	id := effectstest.CastCatalogSpell(t, g, "Visage Bandit", "Creature — Shapeshifter Rogue", otjUZVisageBanditOracle, nil)
	for i := 0; i < 8 && effectstest.CopyPrompt(g) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	pc := effectstest.CopyPrompt(g)
	if pc == nil {
		t.Fatal("no copy prompt")
	}
	if len(pc.CopyOptions) != 1 || pc.CopyOptions[0] != angel {
		t.Fatalf("CopyOptions = %v, want only your creature (%s), never %s", pc.CopyOptions, angel, theirs)
	}
	effectstest.ResolveWithCopyChoice(t, g, angel)
	got := effectstest.CopyBattlefieldCard(t, g, id)
	if got.Name != "Serra Angel" || !got.IsCopy() {
		t.Fatalf("name %q, copy %v, want a copy of Serra Angel", got.Name, got.IsCopy())
	}
	subs := map[string]bool{}
	for _, s := range effectstest.EffectiveSubtypes(t, g, id) {
		subs[s] = true
	}
	for _, want := range []string{"Angel", "Shapeshifter", "Rogue"} {
		if !subs[want] {
			t.Errorf("subtypes %v, missing %q", subs, want)
		}
	}
}

// --- Voracious Varmint -------------------------------------------------------

func TestOtjUZVoraciousVarmintSacrificesToDestroyAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	varmint := effectstest.B12Push(g, me.ID, "Voracious Varmint", "Creature — Varmint", otjUZVarmintOracle, 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, varmint, "vigilance") {
		t.Error("vigilance is printed")
	}
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{G}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, varmint, 0, game.ActivateAbilityParams{
		Targets: effectstest.CardRefs(bear), Strict: true,
	}); err == nil {
		t.Fatal("a creature is not an artifact or enchantment")
	}
	if err := g.ActivateCatalogAbility(me.ID, varmint, 0, game.ActivateAbilityParams{
		Targets: effectstest.CardRefs(rock), Strict: true,
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(varmint) || g.Battlefield.Contains(rock) {
		t.Error("the Varmint was not sacrificed or the artifact not destroyed")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the Bear left")
	}
}

// --- Wanted Griffin ----------------------------------------------------------

func TestOtjUZWantedGriffinLeavesAMercenaryWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	griffin := effectstest.B12Push(g, me.ID, "Wanted Griffin", "Creature — Griffin", otjUZWantedGriffinOracle, 3, 2)
	if !effectstest.HasEffectiveKeyword(t, g, griffin, "flying") {
		t.Error("flying is printed")
	}
	if err := (effects.DestroyTarget{Target: griffin}).Apply(effectstest.CtxFor(g, &game.StackItem{Controller: me.ID})); err != nil {
		t.Fatalf("destroy: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Mercenary"); n != 1 {
		t.Errorf("%d Mercenary tokens, want 1", n)
	}
}

// --- Wylie Duke, Atiin Hero --------------------------------------------------

func TestOtjUZWylieDukeGainsAndDrawsWhenTappedButNotOnAVigilantAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wylie := effectstest.B12Push(g, me.ID, "Wylie Duke, Atiin Hero", "Legendary Creature — Human Ranger", otjUZWylieDukeOracle, 4, 2)
	if !effectstest.HasEffectiveKeyword(t, g, wylie, "vigilance") {
		t.Error("vigilance is printed")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	hand, life := me.Hand.Size(), me.Life
	if err := g.DeclareAttacker(wylie, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand || me.Life != life {
		t.Fatalf("a vigilant attack taps nothing: hand %d → %d, life %d → %d", hand, me.Hand.Size(), life, me.Life)
	}
	if err := g.TapCard(wylie, true); err != nil {
		t.Fatalf("TapCard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || me.Life != life+1 {
		t.Errorf("tapped: hand %d → %d, life %d → %d, want +1 each", hand, me.Hand.Size(), life, me.Life)
	}
}
