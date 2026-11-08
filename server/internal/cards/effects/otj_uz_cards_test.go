package effects

import (
	"testing"

	"github.com/google/uuid"

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
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
	for oracle, cost := range map[string]string{
		otjUZContractorOracle:   "{2}{B}",
		otjUZVisageBanditOracle: "{2}{U}",
	} {
		if sa := plotOf(oracle); sa == nil || sa.Cost != cost {
			t.Errorf("%s: plot %+v, want %s", oracle, sa, cost)
		}
	}
}

// --- Unfortunate Accident ----------------------------------------------------

func TestOtjUZUnfortunateAccidentDestroysAndMakesAMercenary(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castModal(t, g, "Unfortunate Accident", "Instant", otjUZUnfortunateAccidentOracle,
		[]int{0, 1}, []game.TargetRef{modeRef(game.TargetCard, theirs, 0, 0)})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the targeted creature was not destroyed")
	}
	if n, _ := countTokensNamed(g, me.ID, "Mercenary"); n != 1 {
		t.Errorf("%d Mercenary tokens, want 1", n)
	}
}

func TestOtjUZUnfortunateAccidentMercenaryAlone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castModal(t, g, "Unfortunate Accident", "Instant", otjUZUnfortunateAccidentOracle, []int{1}, nil)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(theirs) {
		t.Error("the destroy bullet was not chosen")
	}
	if n, _ := countTokensNamed(g, me.ID, "Mercenary"); n != 1 {
		t.Errorf("%d Mercenary tokens, want 1", n)
	}
}

// --- Unscrupulous Contractor -------------------------------------------------

func TestOtjUZUnscrupulousContractorSacrificesForTwoCardsAndTwoLife(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	contractor := castHoldCreature(t, g, "Unscrupulous Contractor", "Creature — Human Assassin", otjUZContractorOracle, 3, 2)
	c := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
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
	answerOwnPermanents(t, g, me.ID, bear)
	b04WaitForPick(t, g, me.ID)
	if g.Battlefield.Contains(bear) {
		t.Fatal("the Bear was not sacrificed before the reflexive trigger")
	}
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Hand.Size() != hand+2 || opp.Life != life-2 {
		t.Errorf("target player: hand %d → %d, life %d → %d, want +2 cards and -2 life", hand, opp.Hand.Size(), life, opp.Life)
	}
}

func TestOtjUZUnscrupulousContractorMayDecline(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	castHoldCreature(t, g, "Unscrupulous Contractor", "Creature — Human Assassin", otjUZContractorOracle, 3, 2)
	life := me.Life
	answerOwnPermanents(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) || me.Life != life {
		t.Error("declining changed the board")
	}
	if latestPickTarget(g, me.ID) != nil {
		t.Error("no sacrifice, no reflexive trigger")
	}
}

// --- Vadmir, New Blood -------------------------------------------------------

func TestOtjUZVadmirGrowsOnceEachTurnAndGainsKeywordsAtFour(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vadmir := b12Push(g, me.ID, "Vadmir, New Blood", "Legendary Creature — Vampire Rogue", otjUZVadmirOracle, 2, 2)
	advanceToMain(t, g)
	otjB1BoltOpponent(t, g, opp.ID)
	if got := plusOneCounters(g, vadmir); got != 1 {
		t.Fatalf("after a crime: %d counters, want 1", got)
	}
	otjB1BoltOpponent(t, g, opp.ID)
	if got := plusOneCounters(g, vadmir); got != 1 {
		t.Fatalf("a second crime this turn: %d counters, want still 1", got)
	}
	ctx := ctxFor(g, &game.StackItem{Controller: me.ID, SourceCardID: vadmir})
	if err := (AddCounter{Target: vadmir, Kind: game.CounterPlusOne, N: 2}).Apply(ctx); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	for _, kw := range []string{"menace", "lifelink"} {
		if hasEffectiveKeyword(t, g, vadmir, kw) {
			t.Errorf("three counters: %s too early", kw)
		}
	}
	if err := (AddCounter{Target: vadmir, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	for _, kw := range []string{"menace", "lifelink"} {
		if !hasEffectiveKeyword(t, g, vadmir, kw) {
			t.Errorf("four counters: no %s", kw)
		}
	}
}

// --- Vault Plunderer ---------------------------------------------------------

func TestOtjUZVaultPlundererTargetPlayerDrawsAndLosesOne(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	castHoldCreature(t, g, "Vault Plunderer", "Creature — Human Rogue", otjUZVaultPlundererOracle, 3, 1)
	myHand, myLife, hand, life := me.Hand.Size(), me.Life, opp.Hand.Size(), opp.Life
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Hand.Size() != hand+1 || opp.Life != life-1 {
		t.Errorf("target player: hand %d → %d, life %d → %d, want +1 and -1", hand, opp.Hand.Size(), life, opp.Life)
	}
	if me.Hand.Size() != myHand || me.Life != myLife {
		t.Error("the controller drew or lost life, not the target")
	}
}

// --- Vengeful Townsfolk ------------------------------------------------------

func TestOtjUZVengefulTownsfolkCountsBatchesOfOtherCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	town := b12Push(g, me.ID, "Vengeful Townsfolk", "Creature — Human Citizen", otjUZTownsfolkOracle, 3, 3)
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	c := b12Creature(g, me.ID, "Bear C", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	ctx := ctxFor(g, &game.StackItem{Controller: me.ID})
	for _, id := range []uuid.UUID{a, b} {
		if err := (DestroyTarget{Target: id}).Apply(ctx); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	if got := plusOneCounters(g, town); got != 1 {
		t.Fatalf("two creatures died together: %d counters, want 1", got)
	}
	if err := (DestroyTarget{Target: theirs}).Apply(ctx); err != nil {
		t.Fatalf("destroy: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := plusOneCounters(g, town); got != 1 {
		t.Fatalf("an opponent's creature died: %d counters, want still 1", got)
	}
	if err := (DestroyTarget{Target: c}).Apply(ctx); err != nil {
		t.Fatalf("destroy: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := plusOneCounters(g, town); got != 2 {
		t.Errorf("a later death: %d counters, want 2", got)
	}
}

// --- Vial Smasher, Gleeful Grenadier -----------------------------------------

func TestOtjUZVialSmasherPingsAnOpponentWhenAnotherOutlawEnters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	castHoldCreature(t, g, "Vial Smasher, Gleeful Grenadier", "Legendary Creature — Goblin Mercenary", otjUZVialSmasherOracle, 3, 2)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("Vial Smasher triggered on its own entry")
	}
	castHoldCreature(t, g, "Bear", "Creature — Bear", "", 2, 2)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("a Bear is no outlaw")
	}
	life := opp.Life
	castHoldCreature(t, g, "Rogue", "Creature — Human Rogue", "", 1, 1)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("opponent life %d → %d, want -1", life, opp.Life)
	}
}

// --- Visage Bandit -----------------------------------------------------------

func TestOtjUZVisageBanditCopiesYourCreatureAsAShapeshifterRogue(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	angel := seedCopyableCreature(g, me.ID, "Serra Angel", "Creature — Angel", 4, 4)
	theirs := seedCopyableCreature(g, opp.ID, "Their Dragon", "Creature — Dragon", 5, 5)
	id := castCatalogSpell(t, g, "Visage Bandit", "Creature — Shapeshifter Rogue", otjUZVisageBanditOracle, nil)
	for i := 0; i < 8 && copyPrompt(g) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	pc := copyPrompt(g)
	if pc == nil {
		t.Fatal("no copy prompt")
	}
	if len(pc.CopyOptions) != 1 || pc.CopyOptions[0] != angel {
		t.Fatalf("CopyOptions = %v, want only your creature (%s), never %s", pc.CopyOptions, angel, theirs)
	}
	resolveWithCopyChoice(t, g, angel)
	got := copyBattlefieldCard(t, g, id)
	if got.Name != "Serra Angel" || !got.IsCopy() {
		t.Fatalf("name %q, copy %v, want a copy of Serra Angel", got.Name, got.IsCopy())
	}
	subs := map[string]bool{}
	for _, s := range effectiveSubtypes(t, g, id) {
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	varmint := b12Push(g, me.ID, "Voracious Varmint", "Creature — Varmint", otjUZVarmintOracle, 2, 2)
	if !hasEffectiveKeyword(t, g, varmint, "vigilance") {
		t.Error("vigilance is printed")
	}
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{G}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, varmint, 0, game.ActivateAbilityParams{
		Targets: cardRefs(bear), Strict: true,
	}); err == nil {
		t.Fatal("a creature is not an artifact or enchantment")
	}
	if err := g.ActivateCatalogAbility(me.ID, varmint, 0, game.ActivateAbilityParams{
		Targets: cardRefs(rock), Strict: true,
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(varmint) || g.Battlefield.Contains(rock) {
		t.Error("the Varmint was not sacrificed or the artifact not destroyed")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the Bear left")
	}
}

// --- Wanted Griffin ----------------------------------------------------------

func TestOtjUZWantedGriffinLeavesAMercenaryWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	griffin := b12Push(g, me.ID, "Wanted Griffin", "Creature — Griffin", otjUZWantedGriffinOracle, 3, 2)
	if !hasEffectiveKeyword(t, g, griffin, "flying") {
		t.Error("flying is printed")
	}
	if err := (DestroyTarget{Target: griffin}).Apply(ctxFor(g, &game.StackItem{Controller: me.ID})); err != nil {
		t.Fatalf("destroy: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Mercenary"); n != 1 {
		t.Errorf("%d Mercenary tokens, want 1", n)
	}
}

// --- Wylie Duke, Atiin Hero --------------------------------------------------

func TestOtjUZWylieDukeGainsAndDrawsWhenTappedButNotOnAVigilantAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wylie := b12Push(g, me.ID, "Wylie Duke, Atiin Hero", "Legendary Creature — Human Ranger", otjUZWylieDukeOracle, 4, 2)
	if !hasEffectiveKeyword(t, g, wylie, "vigilance") {
		t.Error("vigilance is printed")
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	hand, life := me.Hand.Size(), me.Life
	if err := g.DeclareAttacker(wylie, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand || me.Life != life {
		t.Fatalf("a vigilant attack taps nothing: hand %d → %d, life %d → %d", hand, me.Hand.Size(), life, me.Life)
	}
	if err := g.TapCard(wylie, true); err != nil {
		t.Fatalf("TapCard: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || me.Life != life+1 {
		t.Errorf("tapped: hand %d → %d, life %d → %d, want +1 each", hand, me.Hand.Size(), life, me.Life)
	}
}
