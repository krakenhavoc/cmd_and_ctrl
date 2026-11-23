package testsms

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// nghe_dimir_test.go — #4092 slice e (the Captain N'ghathrod deck
// request): Creeping Tar Pit, Restless Reef, Port of Karfell, Dimir
// Keyrune, and the Ravenous Chupacabra audit.

const (
	ngheTarPitOracle     = "250cb58b-2924-4dff-92fe-ac0ebbbeb218"
	ngheReefOracle       = "df20f85b-5f81-4ee4-8487-55d90109ac36"
	ngheKarfellOracle    = "f500a8a4-6135-448c-9116-bd2695a72229"
	ngheKeyruneOracle    = "7685ff3f-6d61-4b1d-a381-49bc6274aa01"
	ngheChupacabraOracle = "7b459306-149b-4f43-abc1-2dd70c748c0e"
)

func TestNgheCardsAreRegisteredFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		ngheTarPitOracle:     "Creeping Tar Pit",
		ngheReefOracle:       "Restless Reef",
		ngheKarfellOracle:    "Port of Karfell",
		ngheKeyruneOracle:    "Dimir Keyrune",
		ngheChupacabraOracle: "Ravenous Chupacabra",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: Lookup = %q, %v, completeness %v, caveats %v", name, spec.Name, ok, spec.Completeness, spec.Caveats)
		}
	}
}

// ngheTryBlock has `attacker` (the active seat's) attack the opponent
// and tries to block it with a fresh 2/2 of theirs. It returns the
// blocker's error.
func ngheTryBlock(t *testing.T, g *game.Game, opp *game.Player, attacker uuid.UUID) error {
	t.Helper()
	blocker := effectstest.B12Creature(g, opp.ID, "Wall of Test", "Creature — Wall", 0, 4)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	return g.DeclareBlocker(blocker, attacker)
}

// --- Creeping Tar Pit --------------------------------------------------

func TestCreepingTarPitEntersTappedAndTapsForBlueOrBlack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.B12PlayFromHand(t, g, "Creeping Tar Pit", "Land", ngheTarPitOracle, game.CastSpellParams{})
	if !effectstest.IsTapped(g, id) {
		t.Fatal("Creeping Tar Pit entered untapped")
	}
	effectstest.FindBattlefieldCardByID(g, id).Tapped = false
	for _, colour := range []string{"U", "B"} {
		me.ManaPool.EmptyPool()
		effectstest.FindBattlefieldCardByID(g, id).Tapped = false
		if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{Colors: []string{colour}}); err != nil {
			t.Fatalf("tap for %s: %v", colour, err)
		}
		if me.ManaPool.Total() != 1 || me.ManaPool.Units()[0].Color != colour {
			t.Errorf("pool = %v, want one %s", me.ManaPool.Units(), colour)
		}
	}
}

// The animation, layer by layer, and "it can't be blocked this turn"
// as part of the same effect.
func TestCreepingTarPitBecomesAnUnblockableThreeTwoElementalLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pit := effectstest.S58p6Push(g, me.ID, "Creeping Tar Pit", "Land", ngheTarPitOracle, "", 0, 0)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	if effectstest.S58p6View(t, g, pit).IsCreature() {
		t.Fatal("an idle Creeping Tar Pit is a creature")
	}
	effectstest.AssertRestrictions(t, g, pit, 0)

	effectstest.NgheAnimate(t, g, me, pit, 1, "U", "B")
	c := effectstest.S58p6View(t, g, pit)
	if !c.IsCreature() || !c.IsLand() {
		t.Fatalf("creature=%v land=%v, want both", c.IsCreature(), c.IsLand())
	}
	if !slices.Contains(c.Effective().Subtypes, "Elemental") {
		t.Errorf("subtypes %v, want Elemental", c.Effective().Subtypes)
	}
	if !slices.Equal(c.EffectiveColors(), []string{"U", "B"}) {
		t.Errorf("colours %v, want blue and black", c.EffectiveColors())
	}
	if c.CurrentPower() != 3 || c.CurrentToughness() != 2 {
		t.Errorf("%d/%d, want 3/2", c.CurrentPower(), c.CurrentToughness())
	}
	effectstest.AssertRestrictions(t, g, pit, game.CantBeBlocked)
	if len(game.ManaAbilitiesForCard(c)) == 0 {
		t.Error("the animated land lost its mana ability")
	}
}

func TestCreepingTarPitCantBeBlockedAndAPlainCreatureCan(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	pit := effectstest.S58p6Push(g, me.ID, "Creeping Tar Pit", "Land", ngheTarPitOracle, "", 0, 0)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.NgheAnimate(t, g, me, pit, 1, "U", "B")
	if err := ngheTryBlock(t, g, opp, pit); err == nil {
		t.Error("an animated Creeping Tar Pit was blocked")
	}
}

func TestNgheControlAPlainAttackerIsBlockable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	if err := ngheTryBlock(t, g, opp, bear); err != nil {
		t.Errorf("a plain attacker could not be blocked, so the block harness proves nothing: %v", err)
	}
}

// It is a land again, colourless, and blockable-in-principle again when
// the turn ends.
func TestCreepingTarPitReturnsToALandAtEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pit := effectstest.S58p6Push(g, me.ID, "Creeping Tar Pit", "Land", ngheTarPitOracle, "", 0, 0)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.NgheAnimate(t, g, me, pit, 1, "U", "B")
	effectstest.AdvanceToNextSeatsTurn(t, g)
	c := effectstest.S58p6View(t, g, pit)
	if c.IsCreature() || !c.IsLand() {
		t.Errorf("after the turn: creature=%v land=%v, want a land only", c.IsCreature(), c.IsLand())
	}
	if len(c.EffectiveColors()) != 0 {
		t.Errorf("colours %v, want none", c.EffectiveColors())
	}
	effectstest.AssertRestrictions(t, g, pit, 0)
}

// CR 302.6: animated the turn it entered, it can't attack.
func TestCreepingTarPitAnimatedTheTurnItEntersCantAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	pit := effectstest.S58p6Push(g, me.ID, "Creeping Tar Pit", "Land", ngheTarPitOracle, "", 0, 0)
	effectstest.FindBattlefieldCardByID(g, pit).SummonedThisTurn = true
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.NgheAnimate(t, g, me, pit, 1, "U", "B")
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(pit, opp.ID); err == nil {
		t.Error("a land that entered this turn attacked")
	}
}

// --- Dimir Keyrune -----------------------------------------------------

func TestDimirKeyruneTapsForBlueOrBlack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rune := effectstest.S58p6Push(g, me.ID, "Dimir Keyrune", "Artifact", ngheKeyruneOracle, "{3}", 0, 0)
	for _, colour := range []string{"U", "B"} {
		me.ManaPool.EmptyPool()
		effectstest.FindBattlefieldCardByID(g, rune).Tapped = false
		if err := g.ActivateManaAbility(me.ID, rune, 0, game.ManaAbilityParams{Colors: []string{colour}}); err != nil {
			t.Fatalf("tap for %s: %v", colour, err)
		}
		if me.ManaPool.Total() != 1 || me.ManaPool.Units()[0].Color != colour {
			t.Errorf("pool = %v, want one %s", me.ManaPool.Units(), colour)
		}
	}
}

func TestDimirKeyruneBecomesAnUnblockableTwoTwoHorrorArtifactCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	rune := effectstest.S58p6Push(g, me.ID, "Dimir Keyrune", "Artifact", ngheKeyruneOracle, "{3}", 0, 0)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	if effectstest.S58p6View(t, g, rune).IsCreature() {
		t.Fatal("an idle Dimir Keyrune is a creature")
	}
	effectstest.AssertRestrictions(t, g, rune, 0)

	effectstest.NgheAnimate(t, g, me, rune, 0, "U", "B")
	c := effectstest.S58p6View(t, g, rune)
	if !c.IsCreature() || !c.IsArtifact() {
		t.Fatalf("creature=%v artifact=%v, want both", c.IsCreature(), c.IsArtifact())
	}
	if !slices.Contains(c.Effective().Subtypes, "Horror") {
		t.Errorf("subtypes %v, want Horror", c.Effective().Subtypes)
	}
	if !slices.Equal(c.EffectiveColors(), []string{"U", "B"}) {
		t.Errorf("colours %v, want blue and black", c.EffectiveColors())
	}
	if c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
		t.Errorf("%d/%d, want 2/2", c.CurrentPower(), c.CurrentToughness())
	}
	effectstest.AssertRestrictions(t, g, rune, game.CantBeBlocked)
	if err := ngheTryBlock(t, g, opp, rune); err == nil {
		t.Error("an animated Dimir Keyrune was blocked")
	}
}

func TestDimirKeyruneIsAnArtifactAgainAtEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rune := effectstest.S58p6Push(g, me.ID, "Dimir Keyrune", "Artifact", ngheKeyruneOracle, "{3}", 0, 0)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.NgheAnimate(t, g, me, rune, 0, "U", "B")
	effectstest.AdvanceToNextSeatsTurn(t, g)
	c := effectstest.S58p6View(t, g, rune)
	if c.IsCreature() || !c.IsArtifact() {
		t.Errorf("after the turn: creature=%v artifact=%v, want an artifact only", c.IsCreature(), c.IsArtifact())
	}
	effectstest.AssertRestrictions(t, g, rune, 0)
}

// CR 302.6: a Keyrune cast this turn can animate but can't attack.
func TestDimirKeyruneCastThisTurnCantAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	rune := effectstest.S58p6Push(g, me.ID, "Dimir Keyrune", "Artifact", ngheKeyruneOracle, "{3}", 0, 0)
	effectstest.FindBattlefieldCardByID(g, rune).SummonedThisTurn = true
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.NgheAnimate(t, g, me, rune, 0, "U", "B")
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(rune, opp.ID); err == nil {
		t.Error("a Keyrune that entered this turn attacked")
	}
}

// --- Restless Reef -----------------------------------------------------

func TestRestlessReefBecomesAFourFourSharkWithDeathtouch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	reef := effectstest.S58p6Push(g, me.ID, "Restless Reef", "Land", ngheReefOracle, "", 0, 0)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	if c := effectstest.S58p6View(t, g, reef); c.IsCreature() || slices.Contains(c.Effective().Abilities, "deathtouch") {
		t.Fatal("an idle Restless Reef is a creature or has deathtouch")
	}
	effectstest.NgheAnimate(t, g, me, reef, 2, "U", "B")
	c := effectstest.S58p6View(t, g, reef)
	if !c.IsCreature() || !c.IsLand() {
		t.Fatalf("creature=%v land=%v, want both", c.IsCreature(), c.IsLand())
	}
	if !slices.Contains(c.Effective().Subtypes, "Shark") {
		t.Errorf("subtypes %v, want Shark", c.Effective().Subtypes)
	}
	if !slices.Equal(c.EffectiveColors(), []string{"U", "B"}) {
		t.Errorf("colours %v, want blue and black", c.EffectiveColors())
	}
	if c.CurrentPower() != 4 || c.CurrentToughness() != 4 {
		t.Errorf("%d/%d, want 4/4", c.CurrentPower(), c.CurrentToughness())
	}
	if !slices.Contains(c.Effective().Abilities, "deathtouch") {
		t.Errorf("abilities %v, want deathtouch", c.Effective().Abilities)
	}
	effectstest.AssertRestrictions(t, g, reef, 0)
}

func TestRestlessReefEntersTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.B12PlayFromHand(t, g, "Restless Reef", "Land", ngheReefOracle, game.CastSpellParams{})
	if !effectstest.IsTapped(g, id) {
		t.Error("Restless Reef entered untapped")
	}
}

// ngheLibrarySize is the cards left in a library.
func ngheLibrarySize(p *game.Player) int { return p.Library.Size() }

// Attacking: the targeted trigger mills the chosen player four cards.
func TestRestlessReefAttackMillsTheTargetPlayerFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	reef := effectstest.S58p6Push(g, me.ID, "Restless Reef", "Land", ngheReefOracle, "", 0, 0)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.NgheAnimate(t, g, me, reef, 2, "U", "B")

	// Attack one player, mill another: the target is its own choice.
	victim := g.Seats[(g.Turn.ActiveSeat+2)%4]
	lib, yard := ngheLibrarySize(victim), victim.Graveyard.Size()
	myLib, oppLib := ngheLibrarySize(me), ngheLibrarySize(opp)
	effectstest.DeclareAttack(t, g, opp.ID, reef)
	effectstest.AnswerPickTargetPlayer(t, g, victim.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := lib - ngheLibrarySize(victim); got != 4 {
		t.Errorf("target library lost %d cards, want 4", got)
	}
	if got := victim.Graveyard.Size() - yard; got != 4 {
		t.Errorf("target graveyard gained %d cards, want 4", got)
	}
	if ngheLibrarySize(me) != myLib || ngheLibrarySize(opp) != oppLib {
		t.Errorf("a player who was not the target milled")
	}
}

// "Target player" includes the controller.
func TestRestlessReefAttackCanTargetItsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	reef := effectstest.S58p6Push(g, me.ID, "Restless Reef", "Land", ngheReefOracle, "", 0, 0)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.NgheAnimate(t, g, me, reef, 2, "U", "B")
	lib := ngheLibrarySize(me)
	effectstest.DeclareAttack(t, g, opp.ID, reef)
	effectstest.AnswerPickTargetPlayer(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := lib - ngheLibrarySize(me); got != 4 {
		t.Errorf("controller milled %d, want 4", got)
	}
}

// --- Port of Karfell ---------------------------------------------------

func TestPortOfKarfellEntersTappedAndTapsForBlue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.B12PlayFromHand(t, g, "Port of Karfell", "Land", ngheKarfellOracle, game.CastSpellParams{})
	if !effectstest.IsTapped(g, id) {
		t.Fatal("Port of Karfell entered untapped")
	}
	effectstest.FindBattlefieldCardByID(g, id).Tapped = false
	me.ManaPool.EmptyPool()
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for U: %v", err)
	}
	if me.ManaPool.Total() != 1 || me.ManaPool.Units()[0].Color != "U" {
		t.Errorf("pool = %v, want one U", me.ManaPool.Units())
	}
}

func ngheActivatePort(t *testing.T, g *game.Game, me *game.Player, port uuid.UUID) {
	t.Helper()
	me.ManaPool.EmptyPool()
	effectstest.FillPool(me, 3)
	effectstest.FillPoolColored(me, "U", 1)
	effectstest.FillPoolColored(me, "B", 2)
	effectstest.S58p6Activate(t, g, me.ID, port, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
}

// The mill comes first and the creature is chosen after it, so a
// creature the mill just put into the graveyard is a legal pick; the
// land was sacrificed as a cost; the creature enters tapped.
func TestPortOfKarfellMillsFourThenReturnsAMilledCreatureTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	port := effectstest.S58p6Push(g, me.ID, "Port of Karfell", "Land", ngheKarfellOracle, "", 0, 0)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	// A creature on top of the library (milled) and one already in the
	// graveyard: both are candidates.
	milledCreature := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: milledCreature, Name: "Milled Ogre", TypeLine: "Creature — Ogre",
		Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	older := effectstest.PushGraveyardCardTyped(me, "Old Bear", "Creature — Bear")
	lib := ngheLibrarySize(me)

	ngheActivatePort(t, g, me, port)

	if g.Battlefield.Contains(port) {
		t.Error("the Port was not sacrificed")
	}
	if got := lib - ngheLibrarySize(me); got != 4 {
		t.Errorf("milled %d, want 4", got)
	}
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no creature-card prompt after the mill")
	}
	if !slices.Contains(pick.ChooseCards, milledCreature) || !slices.Contains(pick.ChooseCards, older) {
		t.Errorf("candidates %v, want both the milled and the older creature", pick.ChooseCards)
	}
	if pick.ChooseMin != 1 || pick.ChooseMax != 1 {
		t.Errorf("bounds %d..%d, want exactly one", pick.ChooseMin, pick.ChooseMax)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, milledCreature)
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(milledCreature) {
		t.Fatal("the chosen creature did not enter the battlefield")
	}
	if !effectstest.IsTapped(g, milledCreature) {
		t.Error("the returned creature entered untapped")
	}
	if !me.Graveyard.Contains(older) {
		t.Error("the card that was not chosen left the graveyard")
	}
}

// A non-creature card is not a legal pick, milled or not.
func TestPortOfKarfellRefusesANonCreaturePick(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	port := effectstest.S58p6Push(g, me.ID, "Port of Karfell", "Land", ngheKarfellOracle, "", 0, 0)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	sorcery := effectstest.PushGraveyardCardTyped(me, "Old Sorcery", "Sorcery")
	creature := effectstest.PushGraveyardCardTyped(me, "Old Bear", "Creature — Bear")

	ngheActivatePort(t, g, me, port)

	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt")
	}
	if slices.Contains(pick.ChooseCards, sorcery) {
		t.Fatalf("the sorcery was offered: %v", pick.ChooseCards)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{sorcery}); err == nil {
		t.Error("a non-candidate pick was accepted")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, creature)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(creature) {
		t.Error("the creature did not return")
	}
}

// With no creature card in the graveyard after the mill the ability
// asks nothing and does nothing more.
func TestPortOfKarfellWithNoCreatureAsksNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	port := effectstest.S58p6Push(g, me.ID, "Port of Karfell", "Land", ngheKarfellOracle, "", 0, 0)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	before := len(g.Battlefield.Cards)
	ngheActivatePort(t, g, me, port)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("a prompt was asked with no creature card in the graveyard")
	}
	if got := len(g.Battlefield.Cards); got != before-1 {
		t.Errorf("battlefield %d -> %d, want only the sacrificed Port gone", before, got)
	}
}

// The pick is graveyard-public: the chooser sees the milled creature by
// name, and no seat learns a library card the ability did not mill.
func TestPortOfKarfellViewsRevealOnlyTheMilledCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	port := effectstest.S58p6Push(g, me.ID, "Port of Karfell", "Land", ngheKarfellOracle, "", 0, 0)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	secret := uuid.New()
	me.Library.PushBottom(game.Card{InstanceID: secret, Name: "Secret Bottom Card", TypeLine: "Creature — Ogre",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	milled := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: milled, Name: "Milled Ogre", TypeLine: "Creature — Ogre",
		Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})

	ngheActivatePort(t, g, me, port)

	mine := protocol.ViewOfGameFor(g, me.ID.String())
	var found bool
	for _, pc := range mine.PendingChoices {
		for _, o := range pc.Options {
			if o.Name == "Milled Ogre" {
				found = true
			}
			if o.Name == "Secret Bottom Card" {
				t.Error("the chooser's pick offered a card that was never milled")
			}
		}
	}
	if !found {
		t.Error("the chooser's view did not offer the milled creature by name")
	}
	for _, seat := range []*game.Player{opp, g.Seats[(g.Turn.ActiveSeat+2)%4]} {
		other := protocol.ViewOfGameFor(g, seat.ID.String())
		for _, pc := range other.PendingChoices {
			for _, o := range pc.Options {
				if o.Name == "Secret Bottom Card" {
					t.Errorf("a seat saw the unmilled library card in %s", pc.Kind)
				}
			}
		}
	}
}

// --- Ravenous Chupacabra (the audit) ------------------------------------

// CR 608.2b: a target that is no longer "a creature an opponent
// controls" when the trigger resolves is illegal, and the ability does
// nothing.
func TestRavenousChupacabraDoesNothingToATargetThatChangedSides(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster, opp := g.Seats[0], g.Seats[1]
	prey := effectstest.SeedCreature(g, "Blightsteel Colossus", opp.ID)

	effectstest.CastCatalogSpell(t, g, "Ravenous Chupacabra", "Creature — Beast Horror", ngheChupacabraOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickCard(t, g, caster.ID, prey)

	// In response, the creature comes under the caster's control.
	c := effectstest.FindBattlefieldCardByID(g, prey)
	c.Controller = caster.ID
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(prey) {
		t.Error("Chupacabra destroyed a creature its controller controls")
	}
}
