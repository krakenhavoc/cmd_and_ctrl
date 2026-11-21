package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	s82BroForgingTheAnchorOracle  = "17c5b5c9-5c58-46dc-b0a7-a635b9997bb8"
	s82BroKaylasReconOracle       = "43bae4b3-57e8-4d9a-9d1b-9805b0e0fff5"
	s82BroSmeltingVatOracle       = "71d2a0fe-5426-4fa3-85a4-564c37184996"
	s82BroGlintRakerOracle        = "e0a95a8e-5867-481b-9e85-e2d5386362b8"
	s82BroHurkylOracle            = "957df9f4-73e9-4540-b838-52f65d0c62ee"
	s82BroSanwellOracle           = "4884ad6c-2290-44fc-ab9e-6e6b75b2d69a"
	s82BroHostileNegotiationsOrcl = "c5cbd852-c42f-4592-a0bb-a8bffa2dd2f4"
	s82BroTemporalAnchorOracle    = "168f2fbb-8cf7-43e0-8bb6-5e4cc3c1e363"
	s82BroQueenKaylaOracle        = "10c2428f-6e4a-4b92-9087-720242a5c25c"
	s82BroArmsRaceOracle          = "e801c5bd-68e6-4a74-bed3-addd18db3b2b"
)

func TestS82BroB15EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s82BroForgingTheAnchorOracle:  "Forging the Anchor",
		s82BroKaylasReconOracle:       "Kayla's Reconstruction",
		s82BroSmeltingVatOracle:       "Smelting Vat",
		s82BroGlintRakerOracle:        "Glint Raker",
		s82BroHurkylOracle:            "Hurkyl, Master Wizard",
		s82BroSanwellOracle:           "Sanwell, Avenger Ace",
		s82BroHostileNegotiationsOrcl: "Hostile Negotiations",
		s82BroTemporalAnchorOracle:    "The Temporal Anchor",
		s82BroQueenKaylaOracle:        "Queen Kayla bin-Kroog",
		s82BroArmsRaceOracle:          "Arms Race",
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

// s82BroLibrary replaces a library with the given cards, first on top.
func s82BroLibrary(p *game.Player, cards ...game.Card) []uuid.UUID {
	p.Library.Cards = nil
	ids := make([]uuid.UUID, len(cards))
	for i := len(cards) - 1; i >= 0; i-- {
		c := cards[i]
		c.InstanceID = uuid.New()
		c.Owner, c.Controller = p.ID, p.ID
		ids[i] = c.InstanceID
		p.Library.PushTop(c)
	}
	return ids
}

func s82BroCard(name, typeLine, cost string) game.Card {
	return game.Card{Name: name, TypeLine: typeLine, ManaCost: cost}
}

func s82BroIn(z *game.Zone, id uuid.UUID) bool {
	for _, c := range z.Cards {
		if c.InstanceID == id {
			return true
		}
	}
	return false
}

func TestS82BroB15ForgingTheAnchorTakesAnyNumberOfArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := s82BroLibrary(me,
		s82BroCard("Sol Ring", "Artifact", "{1}"),
		s82BroCard("Forest", "Basic Land — Forest", ""),
		s82BroCard("Bear Golem", "Artifact Creature — Golem", "{3}"),
		s82BroCard("Shock", "Instant", "{R}"),
		s82BroCard("Mind Stone", "Artifact", "{2}"),
		s82BroCard("Deep", "Sorcery", "{U}"), // sixth: not looked at
	)
	effectstest.CastCatalogSpell(t, g, "Forging the Anchor", "Sorcery", s82BroForgingTheAnchorOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, ids[0], ids[2], ids[4])
	effectstest.PassPriorityAroundTable(t, g)
	for _, i := range []int{0, 2, 4} {
		if !s82BroIn(me.Hand, ids[i]) {
			t.Errorf("artifact %d is not in hand", i)
		}
	}
	for _, i := range []int{1, 3} {
		if !s82BroIn(me.Library, ids[i]) {
			t.Errorf("rest card %d left the library", i)
		}
		if me.Library.Cards[0].InstanceID == ids[i] || me.Library.Cards[len(me.Library.Cards)-1].InstanceID == ids[i] {
			// the top is the sixth card; the rest are at the bottom
			if me.Library.Cards[len(me.Library.Cards)-1].InstanceID == ids[i] {
				t.Errorf("rest card %d is on top", i)
			}
		}
	}
	if top := me.Library.Cards[len(me.Library.Cards)-1].InstanceID; top != ids[5] {
		t.Errorf("the sixth card is not on top any more")
	}
	// Nothing outside the five is touched; a non-artifact is not offered.
}

func TestS82BroB15ForgingTheAnchorRefusesANonArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := s82BroLibrary(me,
		s82BroCard("Sol Ring", "Artifact", "{1}"),
		s82BroCard("Forest", "Basic Land — Forest", ""),
		s82BroCard("A", "Instant", ""), s82BroCard("B", "Instant", ""), s82BroCard("C", "Instant", ""),
	)
	effectstest.CastCatalogSpell(t, g, "Forging the Anchor", "Sorcery", s82BroForgingTheAnchorOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt")
	}
	if len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != ids[0] {
		t.Errorf("candidates %v, want only the artifact", pick.ChooseCards)
	}
	effectstest.AnswerChooseCards(t, g, me.ID) // declines
	effectstest.PassPriorityAroundTable(t, g)
	if s82BroIn(me.Hand, ids[0]) {
		t.Error("a declined artifact went to hand")
	}
	if !s82BroIn(me.Library, ids[0]) {
		t.Error("the declined artifact left the library")
	}
}

func TestS82BroB15KaylasReconstructionPutsUpToXSmallArtifactsAndCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := s82BroLibrary(me,
		s82BroCard("Cheap Bear", "Creature — Bear", "{1}{G}"),
		s82BroCard("Big Golem", "Artifact Creature — Golem", "{4}"),
		s82BroCard("Mind Stone", "Artifact", "{2}"),
		s82BroCard("Shock", "Instant", "{R}"),
		s82BroCard("Pup", "Creature — Dog", "{W}"),
		s82BroCard("A", "Sorcery", ""), s82BroCard("B", "Sorcery", ""), s82BroCard("Deep", "Sorcery", ""),
	)
	effectstest.PaCast(t, g, "Kayla's Reconstruction", "Sorcery", "{X}{W}{W}{W}", s82BroKaylasReconOracle, game.CastSpellParams{XValue: 2})
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt")
	}
	if pick.ChooseMax != 2 {
		t.Errorf("max %d, want X = 2", pick.ChooseMax)
	}
	for _, id := range pick.ChooseCards {
		if id == ids[1] || id == ids[3] {
			t.Errorf("offered an illegal card %v (mana value 4 or an instant)", id)
		}
	}
	if len(pick.ChooseCards) != 3 {
		t.Errorf("%d candidates, want 3 (Bear, Mind Stone, Pup)", len(pick.ChooseCards))
	}
	effectstest.AnswerChooseCards(t, g, me.ID, ids[0], ids[2])
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(ids[0]) || !g.Battlefield.Contains(ids[2]) {
		t.Error("the picked cards did not enter")
	}
	if !s82BroIn(me.Library, ids[4]) || !s82BroIn(me.Library, ids[1]) {
		t.Error("the rest did not stay in the library (bottom)")
	}
}

func TestS82BroB15KaylasReconstructionWithXZeroPutsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := s82BroLibrary(me, s82BroCard("Pup", "Creature — Dog", "{W}"), s82BroCard("B", "Sorcery", ""))
	top := s82BroCard("Top", "Sorcery", "")
	_ = top
	effectstest.PaCast(t, g, "Kayla's Reconstruction", "Sorcery", "{X}{W}{W}{W}", s82BroKaylasReconOracle, game.CastSpellParams{XValue: 0})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ids[0]) {
		t.Error("X=0 put a card onto the battlefield")
	}
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("X=0 asked for a pick")
	}
}

func TestS82BroB15SmeltingVatPutsTwoNoncreatureArtifactsUnderTheCeiling(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	vat := effectstest.B12Permanent(g, me.ID, "Smelting Vat", "Artifact")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == vat {
				g.Battlefield.Cards[i].OracleID = s82BroSmeltingVatOracle
			}
		}
	})
	sac := effectstest.PaPermanent(g, me.ID, "Three-Drop", "Artifact", "{3}", 0, 0)
	ids := s82BroLibrary(me,
		s82BroCard("Two A", "Artifact", "{2}"),
		s82BroCard("Two B", "Artifact", "{2}"),
		s82BroCard("One", "Artifact", "{1}"),
		s82BroCard("Golem", "Artifact Creature — Golem", "{1}"),
		s82BroCard("Four", "Artifact", "{4}"),
		s82BroCard("Bolt", "Instant", "{R}"),
		s82BroCard("Bear", "Creature — Bear", "{1}"),
		s82BroCard("Land", "Land", ""),
		s82BroCard("Ninth", "Sorcery", ""),
	)
	effectstest.AdvanceToMain(t, g)
	// {1}, {T}, sacrifice another artifact.
	if err := g.ActivateCatalogAbility(me.ID, vat, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{sac}}); err == nil {
		t.Fatal("activated with no mana to pay {1}")
	}
	if !g.Battlefield.Contains(sac) {
		t.Fatal("a refused activation sacrificed the artifact")
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, vat, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{sac}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(sac) {
		t.Error("the artifact was not sacrificed")
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick prompt")
	}
	if len(pick.ChooseCards) != 4 {
		t.Errorf("%d candidates, want the four noncreature artifacts", len(pick.ChooseCards))
	}
	// Two + Two = 4 > 3: refused.
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{ids[0], ids[1]}); err == nil {
		t.Fatal("two mana value 2 artifacts passed a ceiling of 3")
	}
	// Four alone is over 3.
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{ids[4]}); err == nil {
		t.Fatal("a mana value 4 artifact passed a ceiling of 3")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{ids[0], ids[2]}); err != nil {
		t.Fatalf("Two + One = 3 was refused: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(ids[0]) || !g.Battlefield.Contains(ids[2]) {
		t.Error("the picked artifacts did not enter")
	}
	for _, i := range []int{1, 3, 4, 5, 6, 7} {
		if g.Battlefield.Contains(ids[i]) || !s82BroIn(me.Library, ids[i]) {
			t.Errorf("rest card %d is not still in the library", i)
		}
	}
}

func TestS82BroB15GlintRakerGetsThePowerOfTheBiggestArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	raker := effectstest.B12Push(g, me.ID, "Glint Raker", "Creature — Drake", s82BroGlintRakerOracle, 1, 3)
	if p := effects.B31CurrentPowerOf(g, raker); p != 1 {
		t.Errorf("power %d with no artifacts, want 1", p)
	}
	effectstest.PaPermanent(g, me.ID, "Four", "Artifact", "{4}", 0, 0)
	effectstest.PaPermanent(g, me.ID, "Two", "Artifact", "{2}", 0, 0)
	if p := effects.B31CurrentPowerOf(g, raker); p != 5 {
		t.Errorf("power %d, want 1 + 4", p)
	}
	effectstest.PaPermanent(g, g.Seats[1].ID, "Opp Seven", "Artifact", "{7}", 0, 0)
	if p := effects.B31CurrentPowerOf(g, raker); p != 5 {
		t.Errorf("an opponent's artifact changed the power to %d", p)
	}
}

func TestS82BroB15GlintRakerRevealsThatManyAndTakesAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	raker := effectstest.B12Push(g, me.ID, "Glint Raker", "Creature — Drake", s82BroGlintRakerOracle, 1, 3)
	effectstest.PaPermanent(g, me.ID, "Two", "Artifact", "{2}", 0, 0) // power 3
	ids := s82BroLibrary(me,
		s82BroCard("Bolt", "Instant", "{R}"),
		s82BroCard("Sol Ring", "Artifact", "{1}"),
		s82BroCard("Mind Stone", "Artifact", "{2}"),
		s82BroCard("Bear", "Creature — Bear", "{1}"),
	)
	effectstest.AdvanceToMain(t, g)
	effectstest.DealCombatDamageToPlayer(g, raker, opp.ID, 3)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no artifact pick")
	}
	if len(pick.ChooseCards) != 2 {
		t.Errorf("%d artifact candidates among the three revealed, want 2", len(pick.ChooseCards))
	}
	effectstest.AnswerChooseCards(t, g, me.ID, ids[2])
	effectstest.PassPriorityAroundTable(t, g)
	if !s82BroIn(me.Hand, ids[2]) {
		t.Error("the chosen artifact is not in hand")
	}
	for _, i := range []int{0, 1} {
		if !s82BroIn(me.Graveyard, ids[i]) {
			t.Errorf("revealed card %d did not go to the graveyard", i)
		}
	}
	if !s82BroIn(me.Library, ids[3]) {
		t.Error("the fourth card, which was not revealed, left the library")
	}
}

// s82BroCastAndSettle casts a spell with the given type line from the active
// seat's hand and lets it resolve.
func s82BroCastAndSettle(t *testing.T, g *game.Game, name, typeLine string) {
	t.Helper()
	effectstest.CastCatalogSpell(t, g, name, typeLine, "", nil)
	effectstest.PassPriorityAroundTable(t, g)
}

func TestS82BroB15HurkylAsksOncePerNoncreatureSpellType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Hurkyl, Master Wizard", "Legendary Creature — Human Wizard Advisor", s82BroHurkylOracle, 2, 4)
	effectstest.AdvanceToMain(t, g)
	s82BroCastAndSettle(t, g, "Hot Rock", "Artifact")
	s82BroCastAndSettle(t, g, "Old Ward", "Enchantment")
	s82BroCastAndSettle(t, g, "A Bear", "Creature — Bear") // a creature spell adds nothing
	ids := s82BroLibrary(me,
		s82BroCard("Both", "Artifact Enchantment", "{2}"),
		s82BroCard("Only Rock", "Artifact", "{1}"),
		s82BroCard("Bolt", "Instant", "{R}"),
		s82BroCard("Golem", "Artifact Creature — Golem", "{3}"),
		s82BroCard("Aura", "Enchantment — Aura", "{1}"),
		s82BroCard("Sixth", "Artifact", "{1}"), // not revealed
	)
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt for the first type")
	}
	if len(pick.ChooseCards) != 3 { // Both, Only Rock, Golem
		t.Errorf("artifact candidates %v, want 3", pick.ChooseCards)
	}
	if pick.ChooseMax != 1 || pick.ChooseMin != 0 {
		t.Errorf("bounds %d..%d, want 0..1", pick.ChooseMin, pick.ChooseMax)
	}
	// The two-type card fills the artifact slot…
	effectstest.AnswerChooseCards(t, g, me.ID, ids[0])
	pick = effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt for the enchantment type")
	}
	// …and so cannot fill the enchantment slot too.
	for _, id := range pick.ChooseCards {
		if id == ids[0] {
			t.Error("the card already taken is offered again")
		}
	}
	if len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != ids[4] {
		t.Errorf("enchantment candidates %v, want just the Aura", pick.ChooseCards)
	}
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !s82BroIn(me.Hand, ids[0]) {
		t.Error("the picked card is not in hand")
	}
	if s82BroIn(me.Hand, ids[4]) {
		t.Error("a declined card went to hand")
	}
	for _, i := range []int{1, 2, 3, 4} {
		if !s82BroIn(me.Library, ids[i]) {
			t.Errorf("revealed card %d is not in the library", i)
		}
		if me.Library.Cards[len(me.Library.Cards)-1].InstanceID == ids[i] {
			t.Errorf("revealed card %d is on top; the rest goes to the bottom", i)
		}
	}
	if me.Library.Cards[len(me.Library.Cards)-1].InstanceID != ids[5] {
		t.Error("the unrevealed sixth card is not on top")
	}
}

func TestS82BroB15HurkylDoesNothingAfterOnlyACreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Hurkyl, Master Wizard", "Legendary Creature — Human Wizard Advisor", s82BroHurkylOracle, 2, 4)
	effectstest.AdvanceToMain(t, g)
	s82BroCastAndSettle(t, g, "A Bear", "Artifact Creature — Golem")
	ids := s82BroLibrary(me, s82BroCard("Rock", "Artifact", "{1}"), s82BroCard("B", "Instant", ""))
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("asked for a card after only a creature spell")
	}
	if me.Library.Cards[len(me.Library.Cards)-1].InstanceID != ids[0] {
		t.Error("the library was disturbed")
	}
}

func TestS82BroB15QueenKaylaWheelsAndReturnsOneOfEachSmallValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	queen := effectstest.B12Push(g, me.ID, "Queen Kayla bin-Kroog", "Legendary Creature — Human Noble", s82BroQueenKaylaOracle, 2, 3)
	effectstest.AdvanceToMain(t, g)
	me.Hand.Cards = nil
	mk := func(name, typeLine, cost string) uuid.UUID {
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Owner: me.ID, Controller: me.ID})
		return id
	}
	one := mk("One Rock", "Artifact", "{1}")
	oneB := mk("One Bear", "Creature — Bear", "{G}")
	two := mk("Two Golem", "Artifact Creature — Golem", "{2}")
	three := mk("Three Rock", "Artifact", "{3}")
	three2 := mk("Three Bolt", "Instant", "{2}{R}")
	four := mk("Four Rock", "Artifact", "{4}")
	s82BroLibrary(me, s82BroCard("D1", "Sorcery", ""), s82BroCard("D2", "Sorcery", ""), s82BroCard("D3", "Sorcery", ""),
		s82BroCard("D4", "Sorcery", ""), s82BroCard("D5", "Sorcery", ""), s82BroCard("D6", "Sorcery", ""), s82BroCard("D7", "Sorcery", ""))
	if err := g.ActivateCatalogAbility(me.ID, queen, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no mana to pay {4}")
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, queen, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != 6 {
		t.Errorf("hand %d, want 6 drawn", got)
	}
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no mana value 1 prompt")
	}
	if len(pick.ChooseCards) != 2 {
		t.Errorf("mana value 1 candidates %v, want the Rock and the Bear", pick.ChooseCards)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, one)
	pick = effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil || len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != two {
		t.Fatalf("mana value 2 prompt: %+v", pick)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, two)
	pick = effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil || len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != three {
		t.Fatalf("mana value 3 prompt (the Bolt is no artifact or creature): %+v", pick)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, three)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{one, two, three} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%v did not return to the battlefield", id)
		}
	}
	for _, id := range []uuid.UUID{oneB, three2, four} {
		if g.Battlefield.Contains(id) || !s82BroIn(me.Graveyard, id) {
			t.Errorf("%v should have stayed in the graveyard", id)
		}
	}
}

func TestS82BroB15QueenKaylaIsSorceryOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	queen := effectstest.B12Push(g, me.ID, "Queen Kayla bin-Kroog", "Legendary Creature — Human Noble", s82BroQueenKaylaOracle, 2, 3)
	effectstest.AdvanceToMain(t, g)
	// A spell on the stack makes it no longer a sorcery-speed window.
	effectstest.CastCatalogSpell(t, g, "Held", "Instant", "", nil)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, queen, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated with a spell on the stack")
	}
}

func TestS82BroB15ArmsRacePutsAnArtifactInWithHasteAndSacrificesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	race := effectstest.B12Permanent(g, me.ID, "Arms Race", "Enchantment")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == race {
				g.Battlefield.Cards[i].OracleID = s82BroArmsRaceOracle
			}
		}
	})
	effectstest.AdvanceToMain(t, g)
	golem := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: golem, Name: "Golem", TypeLine: "Artifact Creature — Golem", ManaCost: "{5}", Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID})
	bear := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: bear, Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{2}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	if err := g.ActivateCatalogAbility(me.ID, race, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no mana to pay {3}{R}")
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "R"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, race, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no put-from-hand prompt")
	}
	if len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != golem {
		t.Errorf("candidates %v, want only the artifact card", pick.ChooseCards)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, golem)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(golem) {
		t.Fatal("the artifact did not enter")
	}
	if !effectstest.HasEffectiveKeyword(t, g, golem, "haste") {
		t.Error("the artifact has no haste")
	}
	if !s82BroIn(me.Hand, bear) {
		t.Error("the non-artifact left the hand")
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(golem) || !s82BroIn(me.Graveyard, golem) {
		t.Error("the artifact was not sacrificed at the beginning of the end step")
	}
}

func TestS82BroB15ArmsRaceDecliningMakesNoSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	race := effectstest.B12Permanent(g, me.ID, "Arms Race", "Enchantment")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == race {
				g.Battlefield.Cards[i].OracleID = s82BroArmsRaceOracle
			}
		}
	})
	effectstest.AdvanceToMain(t, g)
	rock := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: rock, Name: "Rock", TypeLine: "Artifact", ManaCost: "{1}", Owner: me.ID, Controller: me.ID})
	me.ManaPool.AddMana(game.ManaToken{Color: "R"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, race, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID) // decline
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || !s82BroIn(me.Hand, rock) {
		t.Error("a declined put moved the artifact")
	}
}

// s82BroPendingCast is chooser's open resolution_cast prompt, or nil.
func s82BroPendingCast(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceResolutionCast && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

func s82BroSanwell(g *game.Game, me *game.Player) uuid.UUID {
	return effectstest.B12Push(g, me.ID, "Sanwell, Avenger Ace", "Legendary Creature — Human Pilot", s82BroSanwellOracle, 3, 6)
}

func TestS82BroB15SanwellPreventsDamageWhileAnArtifactCreatureAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	sanwell := s82BroSanwell(g, me)
	golem := effectstest.B12Push(g, me.ID, "Golem", "Artifact Creature — Golem", "", 2, 2)
	src := effectstest.Pr7Creature(g, opp.ID, "Shooter", 1, "R")
	effectstest.Pr6Damage(t, g, src, sanwell, 1)
	if got := effectstest.Pr6Marked(g, sanwell); got != 1 {
		t.Fatalf("damage with nothing attacking: marked %d, want 1", got)
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == golem {
				g.Battlefield.Cards[i].AttackingTarget = opp.ID
			}
		}
	})
	effectstest.Pr6Damage(t, g, src, sanwell, 5)
	if got := effectstest.Pr6Marked(g, sanwell); got != 1 {
		t.Errorf("with an artifact creature attacking: marked %d, want it still 1 (prevented)", got)
	}
}

func TestS82BroB15SanwellIgnoresAnAttackingNonArtifactCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	sanwell := effectstest.B12Push(g, me.ID, "Sanwell, Avenger Ace", "Legendary Creature — Human Pilot", s82BroSanwellOracle, 3, 4)
	bear := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bear {
				g.Battlefield.Cards[i].AttackingTarget = opp.ID
			}
		}
	})
	src := effectstest.Pr7Creature(g, opp.ID, "Shooter", 1, "R")
	effectstest.Pr6Damage(t, g, src, sanwell, 2)
	if got := effectstest.Pr6Marked(g, sanwell); got != 2 {
		t.Errorf("marked %d, want 2: a plain attacking creature prevents nothing", got)
	}
}

func TestS82BroB15SanwellExilesSixAndCastsAnArtifactCreatureOrVehicle(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	sanwell := s82BroSanwell(g, me)
	ids := s82BroLibrary(me,
		s82BroCard("Bolt", "Instant", "{R}"),
		s82BroCard("Golem", "Artifact Creature — Golem", ""),
		s82BroCard("Skiff", "Artifact — Vehicle", ""),
		s82BroCard("Bear", "Creature — Bear", ""),
		s82BroCard("Land", "Land", ""),
		s82BroCard("Rock", "Artifact", ""),
		s82BroCard("Seventh", "Sorcery", ""),
	)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(sanwell) })
	effectstest.PassPriorityAroundTable(t, g)
	prompt := s82BroPendingCast(g, me.ID)
	if prompt == nil {
		t.Fatal("no cast offer")
	}
	if len(prompt.DiscardOptions) != 2 {
		t.Errorf("offered %v, want the Golem and the Vehicle only", prompt.DiscardOptions)
	}
	for _, id := range ids[:6] {
		if !g.Exile.Contains(id) {
			t.Errorf("%v was not exiled", id)
		}
	}
	if g.Exile.Contains(ids[6]) {
		t.Error("the seventh card was exiled")
	}
	if err := g.CastForResolutionChoice(me.ID, prompt.ID, ids[1], game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("cast the Golem: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(ids[1]) {
		t.Error("the cast Golem did not resolve")
	}
	// The rest, including the uncast Vehicle, is on the bottom of the library.
	for _, i := range []int{0, 2, 3, 4, 5} {
		if g.Exile.Contains(ids[i]) || !s82BroIn(me.Library, ids[i]) {
			t.Errorf("card %d did not go to the library", i)
		}
		if me.Library.Cards[len(me.Library.Cards)-1].InstanceID == ids[i] {
			t.Errorf("card %d is on top, not the bottom", i)
		}
	}
}

func TestS82BroB15SanwellSecondCastIsNotOffered(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	sanwell := s82BroSanwell(g, me)
	ids := s82BroLibrary(me,
		s82BroCard("Golem A", "Artifact Creature — Golem", ""),
		s82BroCard("Golem B", "Artifact Creature — Golem", ""),
		s82BroCard("C", "Sorcery", ""), s82BroCard("D", "Sorcery", ""), s82BroCard("E", "Sorcery", ""), s82BroCard("F", "Sorcery", ""))
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(sanwell) })
	effectstest.PassPriorityAroundTable(t, g)
	prompt := s82BroPendingCast(g, me.ID)
	if prompt == nil {
		t.Fatal("no cast offer")
	}
	if err := g.CastForResolutionChoice(me.ID, prompt.ID, ids[0], game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("cast Golem A: %v", err)
	}
	if err := g.CastSpell(me.ID, ids[1], game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Error("a second spell could be cast from among them")
	}
}

func TestS82BroB15SanwellDecliningBottomsEverything(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	sanwell := s82BroSanwell(g, me)
	ids := s82BroLibrary(me,
		s82BroCard("Golem", "Artifact Creature — Golem", ""),
		s82BroCard("B", "Sorcery", ""), s82BroCard("C", "Sorcery", ""), s82BroCard("D", "Sorcery", ""), s82BroCard("E", "Sorcery", ""), s82BroCard("F", "Sorcery", ""))
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(sanwell) })
	effectstest.PassPriorityAroundTable(t, g)
	prompt := s82BroPendingCast(g, me.ID)
	if prompt == nil {
		t.Fatal("no cast offer")
	}
	if err := g.ResolvePendingChoice(prompt.ID, me.ID, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for i, id := range ids {
		if g.Exile.Contains(id) || !s82BroIn(me.Library, id) {
			t.Errorf("card %d is not back in the library", i)
		}
	}
}

func TestS82BroB15TemporalAnchorScriesAtUpkeepAndExilesWhatYouBottom(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	anchor := effectstest.B12Permanent(g, me.ID, "The Temporal Anchor", "Legendary Artifact")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == anchor {
				g.Battlefield.Cards[i].OracleID = s82BroTemporalAnchorOracle
			}
		}
	})
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	ids := s82BroLibrary(me,
		s82BroCard("Top", "Sorcery", ""), s82BroCard("Second", "Creature — Bear", ""),
		s82BroCard("Third", "Sorcery", ""), s82BroCard("Fourth", "Sorcery", ""),
		s82BroCard("Fifth", "Basic Land — Forest", ""), s82BroCard("Sixth", "Sorcery", ""))
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	scry := effectstest.ScryChoiceFor(g, me.ID)
	if scry == nil {
		t.Fatal("no scry 2 at the upkeep")
	}
	// Bottom both looked-at cards: two cards are exiled from the bottom.
	if err := g.ResolveScry(scry.ID, me.ID, []uuid.UUID{ids[0], ids[1]}, nil); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	exiled := 0
	for _, id := range ids {
		if g.Exile.Contains(id) {
			exiled++
		}
	}
	if exiled != 2 {
		t.Fatalf("%d cards exiled, want 2", exiled)
	}
	// The two bottomed cards were put on the bottom first, so the two at the
	// very bottom are exiled; the cards the scry kept on top are untouched.
	if !s82BroIn(me.Library, ids[2]) || !s82BroIn(me.Library, ids[3]) {
		t.Error("a kept card left the library")
	}
}

func TestS82BroB15HostileNegotiationsPilesTheSixAndTheOpponentChooses(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ids := s82BroLibrary(me,
		s82BroCard("A1", "Creature — Bear", "{1}"), s82BroCard("A2", "Sorcery", ""), s82BroCard("A3", "Land", ""),
		s82BroCard("B1", "Artifact", "{1}"), s82BroCard("B2", "Instant", ""), s82BroCard("B3", "Land", ""),
		s82BroCard("Stays", "Sorcery", ""),
	)
	pile1, pile2 := ids[:3], ids[3:6]
	life := effectstest.LifeOf(g, me.ID)
	handBefore := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Hostile Negotiations", "Instant", s82BroHostileNegotiationsOrcl, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range ids[:6] {
		if !effectstest.InExile(g, id) {
			t.Fatalf("card %v is not exiled", id)
		}
	}
	if effectstest.InExile(g, ids[6]) {
		t.Fatal("a seventh card was exiled")
	}
	// Both piles are face down, and I may look at each.
	for _, id := range ids[:6] {
		if !effectstest.KnownBy(g, id, me.ID) {
			t.Errorf("I cannot look at %v", id)
		}
		if effectstest.KnownBy(g, id, opp.ID) {
			t.Errorf("the opponent can see %v before any pile is turned up", id)
		}
	}
	// Turn the second pile face up.
	effectstest.AnswerOptionPick(t, g, me.ID, 1)
	for _, id := range pile2 {
		if !effectstest.KnownBy(g, id, opp.ID) {
			t.Errorf("the face-up pile's card %v is not public", id)
		}
	}
	for _, id := range pile1 {
		if effectstest.KnownBy(g, id, opp.ID) {
			t.Errorf("the face-down pile's card %v is visible to the opponent", id)
		}
	}
	effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
	pick := effectstest.LatestOptionPickFor(g, opp.ID)
	if pick == nil {
		t.Fatal("the opponent is not asked to choose a pile")
	}
	if len(pick.PickOptions) != 2 || len(pick.PickOptions[0].Cards) != 3 || len(pick.PickOptions[1].Cards) != 3 {
		t.Fatalf("options %+v, want two piles of three", pick.PickOptions)
	}
	// The opponent chooses the face-down pile (option index 1 in the
	// prompt's order: face-up first).
	effectstest.AnswerOptionPick(t, g, opp.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range pile1 {
		if !s82BroIn(me.Hand, id) {
			t.Errorf("the chosen pile's card %v is not in hand", id)
		}
	}
	for _, id := range pile2 {
		if !s82BroIn(me.Graveyard, id) {
			t.Errorf("the other pile's card %v is not in the graveyard", id)
		}
	}
	if got := me.Hand.Size() - handBefore; got != 3 {
		t.Errorf("hand grew by %d, want 3", got)
	}
	if got := effectstest.LifeOf(g, me.ID); got != life-3 {
		t.Errorf("life %d -> %d, want -3", life, got)
	}
}

func TestS82BroB15HostileNegotiationsFaceUpPileCanBeTheOneTaken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ids := s82BroLibrary(me,
		s82BroCard("A1", "Creature — Bear", "{1}"), s82BroCard("A2", "Sorcery", ""), s82BroCard("A3", "Land", ""),
		s82BroCard("B1", "Artifact", "{1}"), s82BroCard("B2", "Instant", ""), s82BroCard("B3", "Land", ""),
	)
	effectstest.CastCatalogSpell(t, g, "Hostile Negotiations", "Instant", s82BroHostileNegotiationsOrcl, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 0) // the first pile is turned face up
	effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
	effectstest.AnswerOptionPick(t, g, opp.ID, 0) // the face-up pile
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range ids[:3] {
		if !s82BroIn(me.Hand, id) {
			t.Errorf("%v is not in hand", id)
		}
	}
	for _, id := range ids[3:] {
		if !s82BroIn(me.Graveyard, id) {
			t.Errorf("%v is not in the graveyard", id)
		}
	}
}

func TestS82BroB15HostileNegotiationsWithAnEmptyLibraryStillCostsLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	me.Library.Cards = nil
	life := effectstest.LifeOf(g, me.ID)
	effectstest.CastCatalogSpell(t, g, "Hostile Negotiations", "Instant", s82BroHostileNegotiationsOrcl, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, me.ID); got != life-3 {
		t.Errorf("life %d -> %d, want -3", life, got)
	}
}
