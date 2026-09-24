package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch02_fourth_pass_test.go — the second slice of roadmap batch 02's
// re-triage (#295): the six cards #1241's triage confirmed ready and
// left for the next PR. The Ozolith shipped early as #1244's own proof
// card (the_ozolith_test.go), so this file covers the other five.
//
// Naming is `b02d`-prefixed to stay clear of batch02_test.go's `b02`
// helpers, batch02_second_pass_test.go's `b02b` helpers and
// batch02_third_pass_test.go's `b02c` helpers.

const (
	b02dBraidsOracle              = "e0445c80-fa53-4c3e-881e-940e9fce7f57"
	b02dHullbreakerHorrorOracle   = "d4a84e78-d9b9-4c67-8a4b-4329e65f0f15"
	b02dFinaleOfDevastationOracle = "69872a9a-fe54-4e58-940c-89395af71acd"
	b02dKindredDiscoveryOracle    = "005ee549-1bf5-478f-bc3f-3e791bd7eecf"
	b02dDawnsTruceOracle          = "37c06f89-db36-4937-9404-2b07cd22e1a6"
)

func TestBatch02FourthPassCardsAreRegistered(t *testing.T) {
	want := map[string]string{
		b02dBraidsOracle:              "Braids, Arisen Nightmare",
		b02dHullbreakerHorrorOracle:   "Hullbreaker Horror",
		b02dFinaleOfDevastationOracle: "Finale of Devastation",
		b02dKindredDiscoveryOracle:    "Kindred Discovery",
		b02dDawnsTruceOracle:          "Dawn's Truce",
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("oracle %s registered as %q, want %q", oracle, spec.Name, name)
		}
	}
}

// --- Braids, Arisen Nightmare ----------------------------------------

// b02dCastInstantAs pushes a costless instant into player's hand and
// casts it directly, bypassing castCatalogSpell's "advance to main
// phase" loop — an instant has no sorcery-speed gate, and advancing
// steps drains the stack (#914 / CR 117.4), which would resolve away
// whatever this leaves open for the next cast.
func b02dCastInstantAs(t *testing.T, g *game.Game, player *game.Player, name string, targets []game.TargetRef) uuid.UUID {
	t.Helper()
	id := uuid.New()
	player.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Instant",
		Owner: player.ID, Controller: player.ID,
	})
	if err := g.CastSpell(player.ID, id, game.CastSpellParams{Targets: targets}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// TestB02dBraidsSacrificesAndEachOpponentAnswersDifferently is the
// whole chain in one board: the controller accepts the "you may"
// and picks an Artifact to sacrifice; the opponent who controls a
// matching Artifact is asked and chooses to sacrifice it (no life
// lost, no card drawn); the two opponents with nothing that shares a
// card type are never asked at all and each costs its controller 2
// life and the caster a card.
func TestB02dBraidsSacrificesAndEachOpponentAnswersDifferently(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp1 := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	opp2 := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	opp3 := g.Seats[(g.Turn.ActiveSeat+3)%len(g.Seats)]

	seedPermanentWithOracle(g, me.ID, "Braids, Arisen Nightmare", "Legendary Creature — Nightmare", b02dBraidsOracle)
	myRock := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "My Rock", TypeLine: "Artifact",
		Owner: me.ID, Controller: me.ID,
	})
	theirRock := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Rock", TypeLine: "Artifact",
		Owner: opp1.ID, Controller: opp1.ID,
	})
	opp1Life, opp2Life, opp3Life := opp1.Life, opp2.Life, opp3.Life
	handBefore := me.Hand.Size()

	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)

	ask := latestChoiceOfKind(g, game.PendingChoiceConfirm)
	if ask == nil || ask.Chooser != me.ID {
		t.Fatalf("Braids must ask its controller whether to sacrifice: %+v", g.PendingChoices)
	}
	if err := g.ResolveConfirm(ask.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveConfirm(yes): %v", err)
	}

	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatalf("Braids must ask WHICH permanent to sacrifice: %+v", g.PendingChoices)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{myRock}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if g.Battlefield.Contains(myRock) {
		t.Fatal("the chosen permanent must be sacrificed")
	}

	// Opponent 1 controls a matching Artifact and is asked; choosing to
	// sacrifice it costs no life and draws the caster nothing.
	op1 := latestOptionPickFor(g, opp1.ID)
	if op1 == nil {
		t.Fatalf("opponent 1 (a matching Artifact) must be asked: %+v", g.PendingChoices)
	}
	if len(op1.PickOptions) != 2 {
		t.Fatalf("opponent 1's option list: %+v", op1.PickOptions)
	}
	answerOptionPick(t, g, opp1.ID, 1)
	sacPick := latestChooseCardsFor(g, opp1.ID)
	if sacPick == nil {
		t.Fatalf("the sacrifice branch must ask WHICH permanent: %+v", g.PendingChoices)
	}
	if err := g.ResolveChooseCards(sacPick.ID, opp1.ID, []uuid.UUID{theirRock}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if g.Battlefield.Contains(theirRock) {
		t.Error("opponent 1's matching Artifact must be sacrificed")
	}
	if opp1.Life != opp1Life {
		t.Errorf("opponent 1 sacrificed instead of losing life: %d, want %d", opp1.Life, opp1Life)
	}

	// Opponents 2 and 3 control nothing that shares a card type with an
	// Artifact and are never asked at all (CR 608.2's "as much as
	// possible", enforced at build time) — they simply take the
	// consequence.
	if latestOptionPickFor(g, opp2.ID) != nil || latestOptionPickFor(g, opp3.ID) != nil {
		t.Error("an opponent with no matching permanent must not be prompted")
	}
	if opp2.Life != opp2Life-2 {
		t.Errorf("opponent 2 life = %d, want %d", opp2.Life, opp2Life-2)
	}
	if opp3.Life != opp3Life-2 {
		t.Errorf("opponent 3 life = %d, want %d", opp3.Life, opp3Life-2)
	}
	if got := me.Hand.Size(); got != handBefore+2 {
		t.Errorf("hand = %d, want %d (one draw per opponent who didn't sacrifice)", got, handBefore+2)
	}
}

// --- Hullbreaker Horror ------------------------------------------------

// TestB02dHullbreakerHorrorBouncesATargetSpellYouDontControl proves the
// first bullet, the flash keyword and the can't-be-countered
// declaration together: an opponent's instant is left open on the
// stack, casting a spell of your own puts the modal trigger up
// (#764's real mode_pick for a trigger), and choosing the first bullet
// bounces the opponent's spell rather than countering it.
func TestB02dHullbreakerHorrorBouncesATargetSpellYouDontControl(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	seedPermanentWithOracle(g, me.ID, "Hullbreaker Horror", "Creature — Kraken Horror", b02dHullbreakerHorrorOracle)

	spec, ok := Lookup(b02dHullbreakerHorrorOracle)
	if !ok || !spec.CantBeCountered {
		t.Fatalf("Hullbreaker Horror must declare CantBeCountered")
	}
	if !hasKeyword(spec.PrintedKeywords, "flash") {
		t.Fatalf("Hullbreaker Horror must print flash: %v", spec.PrintedKeywords)
	}

	toMainForCost(t, g)
	theirSpell := b02dCastInstantAs(t, g, opp, "Their Instant", nil)
	mySpell := b02dCastInstantAs(t, g, me, "My Instant", nil)

	c := modePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("casting a spell must offer Hullbreaker Horror's modal trigger: %+v", g.PendingChoices)
	}
	if c.ModeMin != 0 || c.ModeMax != 1 {
		t.Fatalf("choose up to one: min=%d max=%d", c.ModeMin, c.ModeMax)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{0}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatalf("the bounce-a-spell bullet must ask for its target: %+v", g.PendingChoices)
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: theirSpell}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	passPriorityAroundTable(t, g)

	if !opp.Hand.Contains(theirSpell) {
		t.Error("the targeted spell must be returned to its owner's hand")
	}
	if g.Battlefield.Contains(mySpell) {
		t.Error("My Instant is not a permanent and should not be on the battlefield")
	}
}

// TestB02dHullbreakerHorrorBouncesATargetNonlandPermanent proves the
// second bullet, over the battlefield rather than the stack.
func TestB02dHullbreakerHorrorBouncesATargetNonlandPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	seedPermanentWithOracle(g, me.ID, "Hullbreaker Horror", "Creature — Kraken Horror", b02dHullbreakerHorrorOracle)
	theirRock := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Rock", TypeLine: "Artifact",
		Owner: opp.ID, Controller: opp.ID,
	})

	castCatalogSpell(t, g, "My Sorcery", "Sorcery", "", nil)
	c := modePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("casting a spell must offer Hullbreaker Horror's modal trigger: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatalf("the bounce-a-permanent bullet must ask for its target: %+v", g.PendingChoices)
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: theirRock}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	passPriorityAroundTable(t, g)

	if !opp.Hand.Contains(theirRock) {
		t.Error("the targeted nonland permanent must be returned to its owner's hand")
	}
}

// hasKeyword reports whether list contains kw. Only used to read a
// Spec's PrintedKeywords back in a test.
func hasKeyword(list []string, kw string) bool {
	for _, k := range list {
		if k == kw {
			return true
		}
	}
	return false
}

// --- Finale of Devastation ---------------------------------------------

// TestB02dFinaleOfDevastationFindsACreatureAndPumpsAtXTen proves the
// library-only search (declared caveat) and the unconditional X>=10
// pump together: a creature at mana value <= X is a legal find, one
// above X is not, and every creature you control gets +X/+X and haste
// until end of turn regardless of what the search turned up.
func TestB02dFinaleOfDevastationFindsACreatureAndPumpsAtXTen(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	other := pushVanillaCreature(g, me.ID, "Already Here", 2, 2)
	within := pushLibraryCardForTest(me, game.Card{Name: "Nine-Cost Beast", TypeLine: "Creature — Beast", ManaCost: "{9}"})
	tooExpensive := pushLibraryCardForTest(me, game.Card{Name: "Eleven-Cost Titan", TypeLine: "Creature — Giant", ManaCost: "{11}"})

	castXSpell(t, g, "Finale of Devastation", "Sorcery", b02dFinaleOfDevastationOracle, "{X}{G}{G}", 10, nil)
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(within) {
		t.Error("the creature card with mana value <= X should have been put onto the battlefield")
	}
	if g.Battlefield.Contains(tooExpensive) {
		t.Error("a mana value 11 creature is not a legal find for X=10")
	}
	if p, tg := effectivePower(t, g, other), effectiveToughness(t, g, other); p != 12 || tg != 12 {
		t.Errorf("Already Here P/T = %d/%d, want 12/12 (2/2 base, +10/+10 for X=10)", p, tg)
	}
	if !eotHasAbility(effectiveAbilities(t, g, other), "haste") {
		t.Error("X >= 10: creatures you control must gain haste until end of turn")
	}
}

// --- Kindred Discovery ---------------------------------------------------

// TestB02dKindredDiscoveryDrawsForChosenTypeEntersOrAttacks proves
// both trigger conditions and the chosen-type gate: an Elf entering
// draws, a Bear entering does not, and of two attackers only the Elf
// draws again.
func TestB02dKindredDiscoveryDrawsForChosenTypeEntersOrAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushNamedTribePermanent(t, g, me.ID, "Kindred Discovery", "Enchantment", b02dKindredDiscoveryOracle, "Elf")
	handBefore := me.Hand.Size()

	castCatalogSpell(t, g, "My Elf", "Creature — Elf Druid", "", nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != handBefore+1 {
		t.Fatalf("hand = %d, want %d (the entering Elf draws)", got, handBefore+1)
	}

	castCatalogSpell(t, g, "My Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != handBefore+1 {
		t.Fatalf("hand = %d, want %d (a non-Elf entering must not draw)", got, handBefore+1)
	}

	elfAttacker := pushDiesCreatureForTest(g, me.ID, "Ready Elf", "", "Creature — Elf Warrior", 2, 2)
	bearAttacker := pushDiesCreatureForTest(g, me.ID, "Ready Bear", "", "Creature — Bear", 2, 2)
	attackWith(t, g, opp.ID, elfAttacker, bearAttacker)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != handBefore+2 {
		t.Fatalf("hand = %d, want %d (only the attacking Elf draws)", got, handBefore+2)
	}
}

// --- Dawn's Truce --------------------------------------------------------

// TestB02dDawnsTruceGrantsHexproofWithoutTheGift proves the
// guaranteed half end to end (a player-level hexproof grant an
// opponent's Bolt can't get past, and the same grant on permanents you
// control) and that an UNPROMISED Truce gives no indestructible. The
// promised half is TestGiftDawnsTrucePromisedDrawsForTheOpponentAndAddsIndestructible.
func TestB02dDawnsTruceGrantsHexproofWithoutTheGift(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushVanillaCreature(g, me.ID, "My Guy", 2, 2)

	castCatalogSpell(t, g, "Dawn's Truce", "Instant", b02dDawnsTruceOracle, nil)
	passPriorityAroundTable(t, g)

	if got := playerAbilities(g, me); !hasPlayerAbility(got, KeywordHexproof) {
		t.Fatalf("Dawn's Truce's caster has %v, want hexproof", got)
	}
	if err := boltAtPlayerErr(t, g, opp, me.ID); err != game.ErrIllegalTarget {
		t.Errorf("an opponent's Bolt at the caster: got %v, want ErrIllegalTarget", err)
	}
	if !eotHasAbility(effectiveAbilities(t, g, mine), "hexproof") {
		t.Error("permanents you control must also gain hexproof")
	}
	if eotHasAbility(effectiveAbilities(t, g, mine), "indestructible") {
		t.Error("unpromised, permanents must not gain indestructible")
	}
}
