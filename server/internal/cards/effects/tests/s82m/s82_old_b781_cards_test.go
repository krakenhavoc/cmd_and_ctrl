package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b781_cards_test.go — Portal, batch 781 (S82, #5630). Each test plays
// the real catalog entry through the engine.

const (
	oldB781DreadChargeOracle = "6b7933b3-821a-4d64-9a10-af8a69fa008e"
	oldB781MemoriesOracle    = "95a2802a-2621-40c3-84f8-51e8aad7b6f0"
	oldB781CruelFateOracle   = "e199e183-e857-4dc7-87f4-36b32c4dac96"
	oldB781MindKnivesOracle  = "f9be5566-c3df-44cc-9de4-29510c8c245f"
	oldB781CommandOracle     = "fd57dfa5-f29d-4b79-8748-c34a73efb7f0"
)

func TestOldB781EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB781DreadChargeOracle: "Dread Charge", oldB781MemoriesOracle: "Ancestral Memories",
		oldB781CruelFateOracle: "Cruel Fate", oldB781MindKnivesOracle: "Mind Knives",
		oldB781CommandOracle: "Command of Unsummoning",
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

func TestOldB781DreadChargeLetsOnlyBlackCreaturesBlockBlackCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	black := effectstest.SfPush(g, me.ID, "Black Attacker", "Creature — Test", 2, []string{"B"})
	red := effectstest.SfPush(g, me.ID, "Red Attacker", "Creature — Test", 2, []string{"R"})
	blackBlocker := effectstest.SfPush(g, opp.ID, "Black Blocker", "Creature — Test", 2, []string{"B"})
	redBlocker := effectstest.SfPush(g, opp.ID, "Red Blocker", "Creature — Test", 2, []string{"R"})
	effectstest.CastCatalogSpell(t, g, "Dread Charge", "Sorcery", oldB781DreadChargeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	late := effectstest.SfPush(g, me.ID, "Late Black", "Creature — Test", 2, []string{"B"})
	card := func(id uuid.UUID) *game.Card { return effectstest.BattlefieldCardPtr(t, g, id) }
	g.ReadSnapshot(func() {
		if r := g.BlockPairRefusalLocked(card(black), card(redBlocker)); r.Legal() {
			t.Error("a red creature blocked a black creature")
		}
		if r := g.BlockPairRefusalLocked(card(black), card(blackBlocker)); !r.Legal() {
			t.Errorf("a black creature could not block: %q", r.Reason)
		}
		if r := g.BlockPairRefusalLocked(card(red), card(redBlocker)); !r.Legal() {
			t.Errorf("a red creature became unblockable: %q", r.Reason)
		}
		if r := g.BlockPairRefusalLocked(card(late), card(redBlocker)); !r.Legal() {
			t.Errorf("a creature that arrived later was affected: %q", r.Reason)
		}
	})
}

func TestOldB781AncestralMemoriesTakesTwoOfSevenAndBinsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := effectstest.SeedLibrary(me, "One", "Two", "Three", "Four", "Five", "Six", "Seven", "Eight")
	effectstest.CastCatalogSpell(t, g, "Ancestral Memories", "Sorcery", oldB781MemoriesOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil || len(pick.ChooseCards) != 7 || pick.ChooseMin != 2 || pick.ChooseMax != 2 {
		t.Fatalf("pick = %+v, want exactly two of the top seven", pick)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{pick.ChooseCards[0], pick.ChooseCards[6]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	inHand, inYard, inLib := 0, 0, 0
	for _, id := range ids {
		switch {
		case me.Hand.Contains(id):
			inHand++
		case me.Graveyard.Contains(id):
			inYard++
		case me.Library.Contains(id):
			inLib++
		}
	}
	if inHand != 2 || inYard != 5 || inLib != 1 {
		t.Errorf("hand %d, graveyard %d, library %d, want 2, 5 and 1", inHand, inYard, inLib)
	}
}

func TestOldB781CruelFateBinsOneAndReordersTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	var top []uuid.UUID
	libBefore := len(opp.Library.Cards)
	for _, n := range []string{"Six", "Five", "Four", "Three", "Two", "One"} {
		top = append(top, effectstest.SsCard(opp.Library, opp.ID, n, "Sorcery", 0, 0))
	}
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Cruel Fate", "Sorcery", oldB781CruelFateOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}); err == nil {
		t.Fatal("Cruel Fate targeted its own caster")
	}
	effectstest.CastCatalogSpell(t, g, "Cruel Fate", "Sorcery", oldB781CruelFateOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil || len(pick.ChooseCards) != 5 || pick.ChooseMin != 1 || pick.ChooseMax != 1 {
		t.Fatalf("pick = %+v, want exactly one of the top five", pick)
	}
	chosen := pick.ChooseCards[2]
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{chosen}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(chosen) || opp.Library.Contains(chosen) {
		t.Error("the chosen card is not in their graveyard")
	}
	if me.Graveyard.Contains(chosen) {
		t.Error("the card went to the caster's graveyard")
	}
	if len(opp.Graveyard.Cards) != 1 {
		t.Errorf("their graveyard holds %d cards, want 1", len(opp.Graveyard.Cards))
	}
	if got := len(opp.Library.Cards); got != libBefore+5 {
		t.Errorf("their library holds %d cards, want %d", got, libBefore+5)
	}
	for _, id := range top[1:] {
		if id != chosen && !opp.Library.Contains(id) {
			t.Error("a card of the look did not return to the library")
		}
	}
}

func TestOldB781MindKnivesDiscardsOneAtRandomFromTheTargetOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	handBefore := len(opp.Hand.Cards)
	for _, n := range []string{"A", "B", "C"} {
		opp.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: n, TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	}
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Mind Knives", "Sorcery", oldB781MindKnivesOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}); err == nil {
		t.Fatal("Mind Knives targeted its own caster")
	}
	effectstest.CastCatalogSpell(t, g, "Mind Knives", "Sorcery", oldB781MindKnivesOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if len(opp.Hand.Cards) != handBefore+2 || len(opp.Graveyard.Cards) != 1 {
		t.Errorf("hand %d, graveyard %d, want %d and 1", len(opp.Hand.Cards), len(opp.Graveyard.Cards), handBefore+2)
	}
}

func TestOldB781CommandOfUnsummoningBouncesTwoAttackersOnlyWhenAttacked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.SfPush(g, me.ID, "Attacker A", "Creature — Test", 3, []string{"R"})
	b := effectstest.SfPush(g, me.ID, "Attacker B", "Creature — Test", 3, []string{"R"})
	home := effectstest.SfPush(g, me.ID, "Home", "Creature — Test", 1, []string{"R"})
	cast := func(p *game.Player, ids ...uuid.UUID) error {
		effectstest.FloatForTest(g, p, "CCU")
		id := uuid.New()
		g.WithWriteLock(func() {
			p.Hand.PushTop(game.Card{InstanceID: id, Name: "Command of Unsummoning", TypeLine: "Instant", ManaCost: "{2}{U}",
				OracleID: oldB781CommandOracle, Owner: p.ID, Controller: p.ID})
		})
		return g.CastSpell(p.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardTargets(ids...)})
	}
	if err := cast(opp, a); err == nil {
		t.Fatal("cast outside the declare attackers step")
	}
	effectstest.SfAttackHold(t, g, opp.ID, a, b)
	if err := cast(opp, home); err == nil {
		t.Fatal("targeted a creature that isn't attacking")
	}
	if err := cast(g.Seats[2], a); err == nil {
		t.Fatal("cast by a player nobody attacked")
	}
	if err := cast(opp, a, b, home); err == nil {
		t.Fatal("targeted three creatures")
	}
	if err := cast(opp, a, b); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(a) || !me.Hand.Contains(b) || g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("the attackers were not returned to their owner's hand")
	}
}
