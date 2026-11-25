package testss82m

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b243_cards_test.go — Amonkhet aftermath split cards, batch 243 (S82
// older sets, #5535). Every test plays the real catalog entry through the
// engine: the left half from the hand, the aftermath half from the graveyard.

const (
	oldB243StartFinish   = "4ae2750e-3650-42df-a539-ea4cb72fa136"
	oldB243SpringMind    = "6dd7e861-7ae2-49e6-a9ab-8fd761939e22"
	oldB243DestinedLead  = "7ebde396-6672-491a-a6ce-1de49b12379b"
	oldB243MouthFeed     = "980f6341-3c13-4405-8550-b838b228f788"
	oldB243OnwardVictory = "9a52c1ea-35ac-4028-8c1f-b7971c1df886"
	oldB243RagsRiches    = "c2c66419-c2ca-4259-afc6-eaec6d3bb897"
	oldB243ReduceRubble  = "cac993d5-28a1-48b6-9c2f-f216cb5f521d"
	oldB243FailureComply = "ed8184f8-f6a7-4f97-9f4e-bf72077af754"
	oldB243PrepareFight  = "fed7a6ad-68d8-4923-9ef0-619f73bfd6c2"
)

func TestOldB243EveryHalfIsRegisteredFull(t *testing.T) {
	for oracle, names := range map[string][2]string{
		oldB243StartFinish: {"Start", "Finish"}, oldB243SpringMind: {"Spring", "Mind"},
		oldB243DestinedLead: {"Destined", "Lead"}, oldB243MouthFeed: {"Mouth", "Feed"},
		oldB243OnwardVictory: {"Onward", "Victory"}, oldB243RagsRiches: {"Rags", "Riches"},
		oldB243ReduceRubble: {"Reduce", "Rubble"}, oldB243FailureComply: {"Failure", "Comply"},
		oldB243PrepareFight: {"Prepare", "Fight"},
	} {
		for i, key := range []string{oracle, oracle + "#1"} {
			spec, ok := effects.Lookup(key)
			if !ok {
				t.Errorf("%s is not in the catalog", names[i])
				continue
			}
			if spec.Name != names[i] || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
				t.Errorf("%s: name %q, completeness %v, %d caveats", names[i], spec.Name, spec.Completeness, len(spec.Caveats))
			}
		}
	}
}

// oldB243Split puts a free split card in `zone` ("hand" for the left half, "graveyard" for the
// aftermath half) and casts the given face.
func oldB243Split(t *testing.T, g *game.Game, oracle string, face int, targets []game.TargetRef, extra func(*game.CastSpellParams)) (uuid.UUID, error) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	card := game.Card{
		InstanceID: uuid.New(), OracleID: oracle, Layout: game.LayoutSplit, Owner: me.ID, Controller: me.ID,
		Faces: []game.Face{
			{Name: "Left", TypeLine: "Instant", ManaCost: "{0}"},
			{Name: "Right", TypeLine: "Instant", ManaCost: "{0}", OracleText: "Aftermath (Cast this spell only from your graveyard. Then exile it.)"},
		},
	}
	card.SetFace(0)
	if face == 0 {
		me.Hand.PushTop(card)
	} else {
		me.Graveyard.PushTop(card)
	}
	effectstest.AdvanceToMain(t, g)
	params := game.CastSpellParams{Face: face, Targets: targets, Strict: true}
	if face == 1 {
		params.FromZone = "graveyard"
	}
	if extra != nil {
		extra(&params)
	}
	return card.InstanceID, g.CastSpell(me.ID, card.InstanceID, params)
}

func oldB243Cast(t *testing.T, g *game.Game, oracle string, face int, targets []game.TargetRef) uuid.UUID {
	t.Helper()
	id, err := oldB243Split(t, g, oracle, face, targets, nil)
	if err != nil {
		t.Fatalf("cast face %d of %s: %v", face, oracle, err)
	}
	return id
}

func oldB243Tokens(g *game.Game, owner uuid.UUID, name string) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Controller == owner && effects.IsToken(c) && c.Name == name {
			out = append(out, c)
		}
	}
	return out
}

func TestOldB243StartMakesTwoVigilantWarriors(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	oldB243Cast(t, g, oldB243StartFinish, 0, nil)
	effectstest.PassPriorityAroundTable(t, g)
	w := oldB243Tokens(g, me.ID, "Warrior")
	if len(w) != 2 || w[0].Power != 1 || w[0].Toughness != 1 {
		t.Fatalf("Warriors = %d, want two 1/1", len(w))
	}
	if !effectstest.HasKeywordOf(g, w[0].InstanceID, "vigilance") {
		t.Error("the Warrior has no vigilance")
	}
}

func TestOldB243FinishSacrificesACreatureAndDestroysATargetThenExiles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	victim := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	if _, err := oldB243Split(t, g, oldB243StartFinish, 1, effectstest.CardTarget(victim), nil); err == nil {
		t.Fatal("Finish was cast without sacrificing a creature")
	}
	id, err := oldB243Split(t, g, oldB243StartFinish, 1, effectstest.CardTarget(victim), func(p *game.CastSpellParams) {
		p.SacrificeIDs = []uuid.UUID{fodder}
	})
	if err != nil {
		t.Fatalf("Finish: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) || g.Battlefield.Contains(victim) {
		t.Error("the sacrifice or the destroyed creature is still on the battlefield")
	}
	if !g.Exile.Contains(id) {
		t.Error("Finish was not exiled as it resolved (CR 702.127a)")
	}
}

func TestOldB243AftermathHalvesCannotBeCastFromHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	card := game.Card{
		InstanceID: uuid.New(), OracleID: oldB243SpringMind, Layout: game.LayoutSplit, Owner: me.ID, Controller: me.ID,
		Faces: []game.Face{
			{Name: "Spring", TypeLine: "Sorcery", ManaCost: "{0}"},
			{Name: "Mind", TypeLine: "Instant", ManaCost: "{0}", OracleText: "Aftermath (Cast this spell only from your graveyard. Then exile it.)"},
		},
	}
	card.SetFace(0)
	me.Hand.PushTop(card)
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1, Strict: true}); err == nil {
		t.Fatal("Mind was cast from the hand")
	}
}

func TestOldB243SpringFetchesATappedBasicAndMindDrawsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	forest := game.NewCard("Forest", me.ID)
	forest.TypeLine = "Basic Land — Forest"
	me.Library.PushTop(forest)
	oldB243Cast(t, g, oldB243SpringMind, 0, nil)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.SearchPromptFor(g, me.ID)
	if prompt == nil {
		t.Fatal("Spring did not offer a search")
	}
	if err := g.ResolveSearchLibrary(prompt.ID, me.ID, []uuid.UUID{forest.InstanceID}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(forest.InstanceID) || !effectstest.TappedNow(g, forest.InstanceID) {
		t.Error("the basic land is not on the battlefield tapped")
	}
	effectstest.SeedLibrary(me, "A", "B", "C")
	hand := me.Hand.Size()
	id := oldB243Cast(t, g, oldB243SpringMind, 1, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+2)
	}
	if !g.Exile.Contains(id) {
		t.Error("Mind was not exiled")
	}
}

func TestOldB243DestinedPumpsAndProtectsAndLeadResolves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	oldB243Cast(t, g, oldB243DestinedLead, 0, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 3 || tough != 2 {
		t.Errorf("Bear is %d/%d, want 3/2", p, tough)
	}
	if !effectstest.HasKeywordOf(g, bear, "indestructible") {
		t.Error("the Bear is not indestructible")
	}
	foe := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	id := oldB243Cast(t, g, oldB243DestinedLead, 1, effectstest.CardTarget(foe))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(id) {
		t.Error("Lead was not exiled as it resolved")
	}
}

func TestOldB243MouthMakesAHippoAndFeedCountsPowerThreeOrMore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oldB243Cast(t, g, oldB243MouthFeed, 0, nil)
	effectstest.PassPriorityAroundTable(t, g)
	h := oldB243Tokens(g, me.ID, "Hippo")
	if len(h) != 1 || h[0].Power != 3 || h[0].Toughness != 3 {
		t.Fatalf("Hippos = %d, want one 3/3", len(h))
	}
	effectstest.B12Creature(g, me.ID, "Small", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "Big", "Creature — Ogre", 5, 5)
	effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Ogre", 6, 6)
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	hand := me.Hand.Size()
	oldB243Cast(t, g, oldB243MouthFeed, 1, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d, want %d (the Hippo and the Big one; not the 2/2 or the opponent's)", me.Hand.Size(), hand+2)
	}
}

func TestOldB243OnwardAddsItsPowerAndVictoryGivesDoubleStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 3, 2)
	oldB243Cast(t, g, oldB243OnwardVictory, 0, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 6 || tough != 2 {
		t.Errorf("Bear is %d/%d, want 6/2", p, tough)
	}
	oldB243Cast(t, g, oldB243OnwardVictory, 1, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOf(g, bear, "double strike") {
		t.Error("the Bear has no double strike")
	}
}

func TestOldB243RagsShrinksEverythingAndRichesTakesOneFromEachOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	big := effectstest.B12Creature(g, me.ID, "Big", "Creature — Ogre", 4, 4)
	oldB243Cast(t, g, oldB243RagsRiches, 0, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) {
		t.Error("my 2/2 survived -2/-2")
	}
	if p, tough := effectstest.PtOf(t, g, big); p != 2 || tough != 2 {
		t.Errorf("Big is %d/%d, want 2/2", p, tough)
	}
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Bear", 3, 3)
	third := g.Seats[2]
	c := effectstest.B12Creature(g, third.ID, "C", "Creature — Bear", 1, 1)
	oldB243Cast(t, g, oldB243RagsRiches, 1, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, opp.ID, b)
	effectstest.AnswerChooseCards(t, g, third.ID, c)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{b, c} {
		if ctl := effectstest.ControllerOf(t, g, id); ctl != me.ID {
			t.Errorf("a chosen creature is controlled by %s, want me", ctl)
		}
	}
	if ctl := effectstest.ControllerOf(t, g, a); ctl != opp.ID {
		t.Error("an unchosen creature changed hands")
	}
}

func TestOldB243RubbleHoldsUpToThreeLandsThroughTheirNextUntap(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	land := effectstest.B12Permanent(g, opp.ID, "Their Island", "Basic Land — Island")
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(land) })
	oldB243Cast(t, g, oldB243ReduceRubble, 1, effectstest.CardTarget(land))
	effectstest.PassPriorityAroundTable(t, g)
	for i := 0; i < 40 && !(g.Turn.ActiveSeat == 1 && g.Turn.Step == game.StepUpkeep); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if g.Turn.ActiveSeat != 1 {
		t.Fatal("never reached the opponent's turn")
	}
	if !effectstest.TappedNow(g, land) {
		t.Error("the land untapped during its controller's next untap step")
	}
}

func TestOldB243ReduceCountersUnlessThreeIsPaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	spell := effectstest.CastCatalogSpell(t, g, "A Bear", "Creature — Bear", "", nil)
	oldB243Cast(t, g, oldB243ReduceRubble, 0, effectstest.CardTarget(spell))
	effectstest.PassUntilTaxed(t, g, me.ID)
	if c := effectstest.AnswerPayUnless(t, g, me.ID, false); c == nil {
		t.Fatal("no pay-unless prompt")
	}
	if g.Stack.Contains(spell) {
		t.Error("the spell was not countered")
	}
}

func TestOldB243FailureReturnsTheSpellToItsOwnersHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	spell := effectstest.CastCatalogSpell(t, g, "A Bear", "Creature — Bear", "", nil)
	oldB243Cast(t, g, oldB243FailureComply, 0, effectstest.CardTarget(spell))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Stack.Contains(spell) || !effectstest.InHandOf(me, spell) {
		t.Error("the spell did not return to its owner's hand")
	}
}

func TestOldB243ComplyBansTheNamedSpellForOpponentsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := effectstest.Top100Opponent(g)
	oldB243Cast(t, g, oldB243FailureComply, 1, nil)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceCardName, me.ID)
	if prompt == nil {
		t.Fatal("the caster was not asked for a card name")
	}
	if err := g.ResolveCardNameChoice(prompt.ID, me.ID, "Lightning Bolt"); err != nil {
		t.Fatalf("ResolveCardNameChoice: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	boltFor := func(p *game.Player, name string) error {
		id := uuid.New()
		g.WithWriteLock(func() {
			p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Instant", Owner: p.ID, Controller: p.ID})
		})
		return g.CastSpell(p.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{effectstest.PlayerTarget(me.ID)}})
	}
	if err := boltFor(opp, "Lightning Bolt"); err == nil {
		t.Error("an opponent cast the named spell")
	}
	if err := boltFor(opp, "Shock"); errors.Is(err, game.ErrCantCast) {
		t.Errorf("an opponent was barred from another name: %v", err)
	}
}

func TestOldB243PrepareUntapsPumpsAndGrantsLifelinkAndFightFights(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(bear) })
	oldB243Cast(t, g, oldB243PrepareFight, 0, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.TappedNow(g, bear) {
		t.Error("the Bear is still tapped")
	}
	if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("Bear is %d/%d, want 4/4", p, tough)
	}
	if !effectstest.HasKeywordOf(g, bear, "lifelink") {
		t.Error("the Bear has no lifelink")
	}
	foe := effectstest.B12Creature(g, opp.ID, "Goblin", "Creature — Goblin", 1, 3)
	oldB243Cast(t, g, oldB243PrepareFight, 1, []game.TargetRef{{Kind: game.TargetCard, ID: bear}, {Kind: game.TargetCard, ID: foe}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(foe) {
		t.Error("the Goblin survived 4 damage")
	}
}
