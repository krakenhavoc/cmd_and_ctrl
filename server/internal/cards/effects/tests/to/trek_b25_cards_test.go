package teststo

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b25_cards_test.go — Star Trek, batch 25 (#4973). Each test plays
// the real catalog entry and asserts what its text says.

const (
	trekB25KirkOracle    = "e85ad29a-ebd6-4f81-af99-556deaedbfe0"
	trekB25LLAPOracle    = "ff6bc9f5-5a83-42d9-b1a4-f0979aed28ad"
	trekB25LocutusOracle = "46ba09a7-42e4-42cd-911c-71d0ade01180"
	trekB25LoreOracle    = "ee712451-01de-4486-9048-fafb5aa5a584"
	trekB25LRellOracle   = "ea684e78-483c-4b89-b184-b2aeede83cc7"
	trekB25MirrorOracle  = "007d08e8-4db2-47d1-ba39-8bb8efa491a1"
	trekB25NogOracle     = "5a8ae1c0-1935-4d03-aa94-9e9088650821"
	trekB25PhloxOracle   = "b04cb333-ca7a-464f-a084-03cf8f468baf"
	trekB25SarekOracle   = "72c51539-5f15-4636-8cb8-b688cf53f431"
)

func TestTrekB25EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		trekB25KirkOracle:    "Kirk, Commanding Officer",
		trekB25LLAPOracle:    "Live Long and Prosper",
		trekB25LocutusOracle: "Locutus, the One Who Speaks",
		trekB25LoreOracle:    "Lore, Positronic Provocateur",
		trekB25LRellOracle:   "L'Rell, Fierce Visionary",
		trekB25MirrorOracle:  "Mirror Universe Intrusion",
		trekB25NogOracle:     "Nog, Ambitious Cadet",
		trekB25PhloxOracle:   "Phlox, Unorthodox Doctor",
		trekB25SarekOracle:   "Sarek, Federation Ambassador",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- fixtures ------------------------------------------------------------------

// trekB25Seats is the active seat and the next one.
func trekB25Seats(g *game.Game) (*game.Player, *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

// trekB25CastCreature puts a creature card with real P/T into the active
// seat's hand, casts it in a main phase, and passes until the stack is
// empty or a prompt stops the table.
func trekB25CastCreature(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	for i := 0; i < 8 && g.Stack.Size() > 0; i++ {
		if err := g.PassPriority(); err != nil {
			if errors.Is(err, game.ErrChoicePending) {
				break
			}
			t.Fatalf("PassPriority: %v", err)
		}
	}
	return id
}

// --- Kirk, Commanding Officer ------------------------------------------------------

func TestTrekB25KirkExilesTheTopCardWhenYouAttackWithTotalPowerEight(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := trekB25Seats(g)
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Impulse Land", TypeLine: "Basic Land — Mountain"},
		game.Card{Name: "Deep Card", TypeLine: "Sorcery"},
	)
	kirk := effectstest.B12Push(g, me.ID, "Kirk, Commanding Officer", "Legendary Creature — Human Officer", trekB25KirkOracle, 3, 3)
	effectstest.B12Creature(g, me.ID, "Big Friend", "Creature — Giant", 5, 5)
	effectstest.TrekB18AAttack(t, g, opp.ID, kirk)
	if !g.Exile.Contains(ids[0]) {
		t.Fatal("the top card was not exiled")
	}
	if !me.Library.Contains(ids[1]) {
		t.Error("the second card left the library")
	}
	if !effectstest.GrantedPermissionOn(g, me.ID, ids[0], game.ZoneExile).Granted() {
		t.Error("the exiled land is not playable this turn")
	}
}

func TestTrekB25KirkDoesNothingBelowTotalPowerEight(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := trekB25Seats(g)
	ids := effectstest.SeedSearchLibrary(me, game.Card{Name: "Top", TypeLine: "Sorcery"})
	kirk := effectstest.B12Push(g, me.ID, "Kirk, Commanding Officer", "Legendary Creature — Human Officer", trekB25KirkOracle, 3, 3)
	effectstest.B12Creature(g, me.ID, "Friend", "Creature — Giant", 4, 4)
	effectstest.TrekB18AAttack(t, g, opp.ID, kirk)
	if g.Exile.Contains(ids[0]) || !me.Library.Contains(ids[0]) {
		t.Error("Kirk triggered with total power 7")
	}
}

func TestTrekB25KirkSacrificesToDestroyAnArtifactOrEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := trekB25Seats(g)
	kirk := effectstest.B12Push(g, me.ID, "Kirk, Commanding Officer", "Legendary Creature — Human Officer", trekB25KirkOracle, 3, 3)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPool(me, 2)
	if err := g.ActivateCatalogAbility(me.ID, kirk, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{effectstest.CardRef(bear)},
	}); err == nil {
		t.Fatal("a creature was accepted as the artifact-or-enchantment target")
	}
	if err := g.ActivateCatalogAbility(me.ID, kirk, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{effectstest.CardRef(rock)},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(kirk) {
		t.Error("Kirk was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
}

// --- Live Long and Prosper ----------------------------------------------------------

func TestTrekB25LiveLongAndProsperDoublesLifeDrawsTheHandAndLiftsTheMaximum(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := trekB25Seats(g)
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E", "F", "G", "H")
	for _, name := range []string{"Held 1", "Held 2", "Held 3"} {
		me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: name, TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	if _, err := g.ChangePlayerLife(me.ID, -10); err != nil {
		t.Fatal(err)
	}
	life, hand := me.Life, effectstest.HandSize(me)
	effectstest.CastCatalogSpell(t, g, "Live Long and Prosper", "Sorcery", trekB25LLAPOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != 2*life {
		t.Errorf("life %d, want %d", me.Life, 2*life)
	}
	if got := effectstest.HandSize(me); got != 2*hand {
		t.Errorf("hand %d after drawing, want %d (it drew as many as it held)", got, 2*hand)
	}
	var limit int
	g.WithWriteLock(func() { limit = g.EffectiveMaxHandSizeLocked(me) })
	if limit != game.NoMaxHandSize {
		t.Errorf("maximum hand size is %d, want no maximum until the next turn", limit)
	}
}

func TestTrekB25LiveLongAndProsperCantBeCountered(t *testing.T) {
	spec, ok := effects.Lookup(trekB25LLAPOracle)
	if !ok || !spec.CantBeCountered {
		t.Fatal("Live Long and Prosper is counterable")
	}
}

// --- Locutus, the One Who Speaks --------------------------------------------------

func TestTrekB25LocutusPumpsArtifactCreaturesYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := trekB25Seats(g)
	locutus := effectstest.B12Push(g, me.ID, "Locutus, the One Who Speaks", "Legendary Artifact Creature — Borg", trekB25LocutusOracle, 1, 1)
	drone := effectstest.B12Creature(g, me.ID, "Drone", "Artifact Creature — Construct", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Drone", "Artifact Creature — Construct", 1, 1)
	for id, want := range map[uuid.UUID]int{locutus: 2, drone: 2, bear: 2, theirs: 1} {
		if p := effectstest.EffectivePower(t, g, id); p != want {
			t.Errorf("power %d, want %d", p, want)
		}
	}
	if tough := effectstest.EffectiveToughness(t, g, drone); tough != 2 {
		t.Errorf("the Drone's toughness %d, want 2", tough)
	}
}

func TestTrekB25LocutusDiscardsToDealTwoAndGainTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := trekB25Seats(g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	id := effectstest.PushCatalogHandCard(me, "Locutus, the One Who Speaks", "Legendary Artifact Creature — Borg", trekB25LocutusOracle)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{effectstest.CardRef(bear)},
	}); err == nil {
		t.Fatal("activated with no {W}{B} to pay")
	}
	effectstest.FillPoolColored(me, "W", 1)
	effectstest.FillPoolColored(me, "B", 1)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err == nil {
		t.Fatal("a player was accepted as the target creature")
	}
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{effectstest.CardRef(bear)},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !me.Graveyard.Contains(id) {
		t.Error("Locutus was not discarded as the cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("2 damage did not kill the 2/2")
	}
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
}

// --- Lore, Positronic Provocateur ---------------------------------------------------

func TestTrekB25LorePumpsAndMakesUnblockable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := trekB25Seats(g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	trekB25CastCreature(t, g, "Lore, Positronic Provocateur", "Legendary Artifact Creature — Android Rogue", trekB25LoreOracle, 2, 3)
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if c == nil {
		t.Fatalf("no mode_pick prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{0}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, bear); p != 3 {
		t.Errorf("the Bear has power %d, want 3", p)
	}
	effectstest.AssertRestrictions(t, g, bear, game.CantBeBlocked)
}

func TestTrekB25LoreSurveilsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := trekB25Seats(g)
	effectstest.SeedLibrary(me, "One", "Two", "Three")
	trekB25CastCreature(t, g, "Lore, Positronic Provocateur", "Legendary Artifact Creature — Android Rogue", trekB25LoreOracle, 2, 3)
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if c == nil {
		t.Fatalf("no mode_pick prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	sc := effectstest.SurveilChoiceFor(g, me.ID)
	if sc == nil || len(sc.ScryCards) != 2 {
		t.Fatalf("want a surveil 2 prompt, got %+v", sc)
	}
}

// --- L'Rell, Fierce Visionary --------------------------------------------------------

func TestTrekB25LRellMakesAKlingonEachCombatAndSacrificesItAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := trekB25Seats(g)
	effectstest.B12Push(g, me.ID, "L'Rell, Fierce Visionary", "Legendary Creature — Klingon Warrior", trekB25LRellOracle, 1, 4)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	tokens := trekB15Tokens(g, me.ID, "Klingon Warrior")
	if len(tokens) != 1 {
		t.Fatalf("%d Klingon Warriors, want 1", len(tokens))
	}
	tok := tokens[0]
	if tok.Power != 2 || tok.Toughness != 1 || !effectstest.HasEffectiveKeyword(t, g, tok.InstanceID, "haste") {
		t.Errorf("token is %d/%d haste=%v, want a 2/1 with haste", tok.Power, tok.Toughness, effectstest.HasEffectiveKeyword(t, g, tok.InstanceID, "haste"))
	}
	trekB15ToEndStep(t, g)
	if g.Battlefield.Contains(tok.InstanceID) {
		t.Error("the Klingon Warrior was not sacrificed at the end step")
	}
}

func TestTrekB25LRellSacrificesAnotherCreatureOrArtifactToSurveil(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := trekB25Seats(g)
	effectstest.SeedLibrary(me, "Top")
	lrell := effectstest.B12Push(g, me.ID, "L'Rell, Fierce Visionary", "Legendary Creature — Klingon Warrior", trekB25LRellOracle, 1, 4)
	land := effectstest.PushLandFor(g, me.ID, "Mountain", "Basic Land — Mountain")
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPool(me, 2)
	if err := g.ActivateCatalogAbility(me.ID, lrell, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{lrell}}); err == nil {
		t.Fatal("L'Rell sacrificed herself for \"another\"")
	}
	if err := g.ActivateCatalogAbility(me.ID, lrell, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{land}}); err == nil {
		t.Fatal("a land paid the creature-or-artifact sacrifice")
	}
	if err := g.ActivateCatalogAbility(me.ID, lrell, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{rock}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not sacrificed")
	}
	effectstest.PassPriorityAroundTable(t, g)
	sc := effectstest.SurveilChoiceFor(g, me.ID)
	if sc == nil || len(sc.ScryCards) != 1 {
		t.Fatalf("want a surveil 1 prompt, got %+v", sc)
	}
}

// --- Mirror Universe Intrusion -------------------------------------------------------

func TestTrekB25MirrorUniverseIntrusionCopiesOneCreatureWithMenace(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := trekB25Seats(g)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	if _, err := effectstest.CastWithOptionalCosts(t, g, "Mirror Universe Intrusion", "Sorcery", trekB25MirrorOracle,
		[]game.TargetRef{effectstest.CardRef(mine)}, nil, nil); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("err = %v, want ErrIllegalTarget for your own creature", err)
	}
	if _, err := effectstest.CastWithOptionalCosts(t, g, "Mirror Universe Intrusion", "Sorcery", trekB25MirrorOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}, nil, nil); err == nil {
		t.Fatal("an unkicked cast accepted an opponent as its target")
	}
	if _, err := effectstest.CastWithOptionalCosts(t, g, "Mirror Universe Intrusion", "Sorcery", trekB25MirrorOracle,
		[]game.TargetRef{effectstest.CardRef(bear)}, nil, nil); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	copies := trekB15Tokens(g, me.ID, "Their Bear")
	if len(copies) != 1 {
		t.Fatalf("%d copies, want 1", len(copies))
	}
	if !effectstest.HasEffectiveKeyword(t, g, copies[0].InstanceID, "menace") {
		t.Error("the copy has no menace")
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "menace") {
		t.Error("the original gained menace")
	}
}

func TestTrekB25MirrorUniverseIntrusionKickedCopiesEachCreatureThatOpponentControls(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := trekB25Seats(g)
	third := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	effectstest.B12Creature(g, third.ID, "Elsewhere", "Creature — Elf", 1, 1)
	if _, err := effectstest.CastWithOptionalCosts(t, g, "Mirror Universe Intrusion", "Sorcery", trekB25MirrorOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}, []int{0}, nil); err == nil {
		t.Fatal("a kicked cast accepted its caster as the target opponent")
	}
	if _, err := effectstest.CastWithOptionalCosts(t, g, "Mirror Universe Intrusion", "Sorcery", trekB25MirrorOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}, []int{0}, nil); err != nil {
		t.Fatalf("kicked cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, name := range []string{"Bear", "Ogre"} {
		copies := trekB15Tokens(g, me.ID, name)
		if len(copies) != 1 {
			t.Errorf("%d copies of %s, want 1", len(copies), name)
			continue
		}
		if !effectstest.HasEffectiveKeyword(t, g, copies[0].InstanceID, "menace") {
			t.Errorf("the %s copy has no menace", name)
		}
	}
	if n := len(trekB15Tokens(g, me.ID, "Rock")) + len(trekB15Tokens(g, me.ID, "Elsewhere")); n != 0 {
		t.Errorf("%d copies of a noncreature or another player's creature", n)
	}
}

// --- Nog, Ambitious Cadet -------------------------------------------------------------

func TestTrekB25NogClimbsFromRogueToScoutToOfficer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := trekB25Seats(g)
	nog := effectstest.B12Push(g, me.ID, "Nog, Ambitious Cadet", "Legendary Creature — Ferengi Rogue", trekB25NogOracle, 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, nog, "menace") {
		t.Error("Nog has no menace")
	}
	effectstest.AdvanceToMain(t, g)
	activate := func(idx, mana int) {
		t.Helper()
		effectstest.FillPool(me, mana)
		if err := g.ActivateCatalogAbility(me.ID, nog, idx, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("activate %d: %v", idx, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	// The Officer ability does nothing to a Rogue.
	activate(1, 3)
	if c := effectstest.EffectiveCardForTest(t, g, nog); c.HasSubtype("Officer") || !c.HasSubtype("Rogue") {
		t.Fatal("the Officer ability changed a Rogue")
	}
	activate(0, 1)
	c := effectstest.EffectiveCardForTest(t, g, nog)
	if !c.HasSubtype("Scout") || !c.HasSubtype("Ferengi") || c.HasSubtype("Rogue") {
		t.Errorf("Nog's types after the Scout ability: %v", c.Effective().Subtypes)
	}
	if p, tough := effectstest.EffectivePower(t, g, nog), effectstest.EffectiveToughness(t, g, nog); p != 3 || tough != 4 {
		t.Errorf("Nog is %d/%d, want 3/4", p, tough)
	}
	// The Scout ability on a Scout does nothing more.
	activate(0, 1)
	activate(1, 3)
	c = effectstest.EffectiveCardForTest(t, g, nog)
	if !c.HasSubtype("Officer") || c.HasSubtype("Scout") {
		t.Errorf("Nog's types after the Officer ability: %v", c.Effective().Subtypes)
	}
	if p, tough := effectstest.EffectivePower(t, g, nog), effectstest.EffectiveToughness(t, g, nog); p != 3 || tough != 4 {
		t.Errorf("the Officer is %d/%d, want the Scout's 3/4", p, tough)
	}
}

func TestTrekB25NogScoutSurveilsWhenItBecomesTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := trekB25Seats(g)
	effectstest.SeedLibrary(me, "One", "Two", "Three")
	nog := effectstest.B12Push(g, me.ID, "Nog, Ambitious Cadet", "Legendary Creature — Ferengi Rogue", trekB25NogOracle, 2, 2)
	// A Rogue that attacks surveils nothing.
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPool(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, nog, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(nog, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	sc := effectstest.SurveilChoiceFor(g, me.ID)
	if sc == nil || len(sc.ScryCards) != 2 {
		t.Fatalf("want a surveil 2 prompt when the Scout attacked, got %+v", sc)
	}
}

func TestTrekB25NogRogueHasNoTappedTrigger(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := trekB25Seats(g)
	effectstest.SeedLibrary(me, "One", "Two")
	nog := effectstest.B12Push(g, me.ID, "Nog, Ambitious Cadet", "Legendary Creature — Ferengi Rogue", trekB25NogOracle, 2, 2)
	effectstest.TrekB18AAttack(t, g, opp.ID, nog)
	if sc := effectstest.SurveilChoiceFor(g, me.ID); sc != nil {
		t.Error("a Rogue Nog surveilled when it attacked")
	}
}

func TestTrekB25NogOfficerPlaysGraveyardLandsAndAnExtraLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := trekB25Seats(g)
	nog := effectstest.B12Push(g, me.ID, "Nog, Ambitious Cadet", "Legendary Creature — Ferengi Rogue", trekB25NogOracle, 2, 2)
	land := effectstest.PushGraveyardCardTyped(me, "Dead Forest", "Basic Land — Forest")
	effectstest.AdvanceToMain(t, g)
	if effectstest.GrantedPermissionOn(g, me.ID, land, game.ZoneGraveyard).Granted() {
		t.Fatal("a Rogue Nog opened graveyard land plays")
	}
	effectstest.FillPool(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, nog, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate the Scout ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.GrantedPermissionOn(g, me.ID, land, game.ZoneGraveyard).Granted() {
		t.Fatal("a Scout Nog opened graveyard land plays")
	}
	effectstest.FillPool(me, 3)
	if err := g.ActivateCatalogAbility(me.ID, nog, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate the Officer ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.GrantedPermissionOn(g, me.ID, land, game.ZoneGraveyard).Granted() {
		t.Fatal("the Officer did not open graveyard land plays")
	}
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{FromZone: string(game.ZoneGraveyard)}); err != nil {
		t.Fatalf("play the land from the graveyard: %v", err)
	}
	if !g.Battlefield.Contains(land) {
		t.Error("the graveyard land never reached the battlefield")
	}
	if err := effectstest.B02cPlayLandRaw(t, g, "Island"); err != nil {
		t.Fatalf("the Officer's additional land play was refused: %v", err)
	}
	if err := effectstest.B02cPlayLandRaw(t, g, "Swamp"); err == nil {
		t.Error("a third land in one turn was allowed")
	}
	bolt := effectstest.PushGraveyardCardTyped(me, "Some Sorcery", "Sorcery")
	if effectstest.GrantedPermissionOn(g, me.ID, bolt, game.ZoneGraveyard).Granted() {
		t.Error("the Officer opened graveyard spell casts")
	}
}

// --- Phlox, Unorthodox Doctor --------------------------------------------------------

func TestTrekB25PhloxDrawsForACreatureWithACounterThatConnects(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := trekB25Seats(g)
	effectstest.SeedLibrary(me, "A", "B", "C")
	phlox := effectstest.B12Push(g, me.ID, "Phlox, Unorthodox Doctor", "Legendary Creature — Denobulan Doctor", trekB25PhloxOracle, 3, 4)
	marked := effectstest.B12Creature(g, me.ID, "Marked", "Creature — Bear", 2, 2)
	plain := effectstest.B12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(marked, game.CounterPlusOne, 1) })
	if !effectstest.HasEffectiveKeyword(t, g, phlox, "trample") {
		t.Error("Phlox has no trample")
	}
	hand := effectstest.HandSize(me)
	effectstest.TrekB25ToCombatDamage(t, g, opp.ID, marked, plain, phlox)
	if got := effectstest.HandSize(me) - hand; got != 1 {
		t.Errorf("drew %d, want 1 (only the creature with a counter)", got)
	}
}

// --- Sarek, Federation Ambassador ----------------------------------------------------

func TestTrekB25SarekDrawsOnePerFiveTypes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := trekB25Seats(g)
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	sarek := effectstest.B12Push(g, me.ID, "Sarek, Federation Ambassador", "Legendary Creature — Vulcan Advisor", trekB25SarekOracle, 0, 5)
	effectstest.B12Creature(g, me.ID, "Crew", "Creature — Human Soldier Pilot", 1, 1)
	effectstest.B12Creature(g, me.ID, "Drone", "Artifact Creature — Borg Construct Golem Wall Ogre", 1, 1)
	hand := effectstest.HandSize(me)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	// Vulcan, Advisor, Human, Soldier, Pilot: five types; the Borg adds none.
	if got := effectstest.HandSize(me) - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
	if n := effectstest.CountersOn(g, sarek, game.CounterPlusOne); n != 0 {
		t.Errorf("Sarek has %d +1/+1 counters after one draw", n)
	}
}

func TestTrekB25SarekPutsCountersAfterThreeDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := trekB25Seats(g)
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E")
	sarek := effectstest.B12Push(g, me.ID, "Sarek, Federation Ambassador", "Legendary Creature — Vulcan Advisor", trekB25SarekOracle, 0, 5)
	crew := effectstest.B12Creature(g, me.ID, "Crew", "Creature — Human Soldier Pilot Elf Goblin", 1, 1)
	effectstest.B12Creature(g, me.ID, "Mob", "Creature — Ogre Giant Troll Orc Dwarf Wizard Cleric Rogue Warrior", 1, 1)
	hand := effectstest.HandSize(me)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandSize(me) - hand; got != 3 {
		t.Fatalf("drew %d, want 3 for 16 types", got)
	}
	for _, id := range []uuid.UUID{sarek, crew} {
		if n := effectstest.CountersOn(g, id, game.CounterPlusOne); n != 2 {
			t.Errorf("%d +1/+1 counters, want 2", n)
		}
	}
}

func TestTrekB25SarekCountsOnlyCardsActuallyDrawn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := trekB25Seats(g)
	g.WithWriteLock(func() {
		for me.Library.Size() > 0 {
			me.Library.PopTop()
		}
	})
	effectstest.SeedLibrary(me, "Only")
	sarek := effectstest.B12Push(g, me.ID, "Sarek, Federation Ambassador", "Legendary Creature — Vulcan Advisor", trekB25SarekOracle, 0, 5)
	effectstest.B12Creature(g, me.ID, "Crew", "Creature — Human Soldier Pilot Elf Goblin", 1, 1)
	effectstest.B12Creature(g, me.ID, "Mob", "Creature — Ogre Giant Troll Orc Dwarf Wizard Cleric Rogue Warrior", 1, 1)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, sarek, game.CounterPlusOne); n != 0 {
		t.Errorf("Sarek has %d +1/+1 counters after drawing one card of three", n)
	}
}
