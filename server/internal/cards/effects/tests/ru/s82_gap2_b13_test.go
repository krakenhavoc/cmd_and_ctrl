package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap2_b13_test.go — S82 GAP2 batch 13 (#5948, #6101).

const (
	gap2B13Bigger   = "8f351961-69e3-4637-b010-d13f19daca39"
	gap2B13Nalathni = "0887dc95-3e75-4f72-bf05-4067d3785376"
	gap2B13Vault    = "3f8e7a45-4c6e-4ee6-93d0-b7de9715ec97"
	gap2B13Devourer = "89215e6d-ebe7-4eb0-9999-11ee2e6eb1d2"
)

func TestGap2B13CardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		gap2B13Bigger:   "Bigger on the Inside",
		gap2B13Nalathni: "Nalathni Dragon",
		gap2B13Vault:    "Lim-Dûl's Vault",
		gap2B13Devourer: "Devourer of Destiny",
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

func gap2B13Pending(t *testing.T, g *game.Game, kind game.PendingChoiceKind, chooser uuid.UUID) *game.PendingChoice {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c.Kind == kind && c.Chooser == chooser {
			return c
		}
	}
	t.Fatalf("no %s prompt for %v: %+v", kind, chooser, g.PendingChoices)
	return nil
}

// gap2B13TopIDs is the top n cards of a library, top card first.
func gap2B13TopIDs(p *game.Player, n int) []uuid.UUID {
	cards := p.Library.Cards
	out := make([]uuid.UUID, 0, n)
	for i := len(cards) - 1; i >= 0 && len(out) < n; i-- {
		out = append(out, cards[i].InstanceID)
	}
	return out
}

// --- Nalathni Dragon ---

func TestGap2B13NalathniDragonFliesBandsAndPumps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dragon := effectstest.B12Push(g, me.ID, "Nalathni Dragon", "Creature — Dragon", gap2B13Nalathni, 1, 1)
	effectstest.AssertKeywords(t, g, dragon, "flying", game.KeywordBanding)
	for i := 1; i <= 3; i++ {
		effectstest.B06AddMana(me, "R")
		if err := g.ActivateCatalogAbility(me.ID, dragon, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activation %d: %v", i, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if p, _ := effectstest.EffectivePTOf(t, g, dragon); p != 4 {
		t.Errorf("power %d after three pumps, want 4", p)
	}
	for g.Turn.Step != game.StepEnd {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dragon) {
		t.Error("three activations do not sacrifice the Dragon")
	}
}

func TestGap2B13NalathniDragonSacrificedAfterFourActivations(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dragon := effectstest.B12Push(g, me.ID, "Nalathni Dragon", "Creature — Dragon", gap2B13Nalathni, 1, 1)
	for i := 1; i <= 4; i++ {
		effectstest.B06AddMana(me, "R")
		if err := g.ActivateCatalogAbility(me.ID, dragon, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activation %d: %v", i, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	for g.Turn.Step != game.StepEnd {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(dragon) {
		t.Error("the fourth activation sacrifices the Dragon at the next end step")
	}
}

func TestGap2B13NalathniDragonPumpNeedsRed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dragon := effectstest.B12Push(g, me.ID, "Nalathni Dragon", "Creature — Dragon", gap2B13Nalathni, 1, 1)
	effectstest.B06AddMana(me, "G")
	if err := g.ActivateCatalogAbility(me.ID, dragon, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("pumped with green mana")
	}
}

// --- Devourer of Destiny ---

func gap2B13DevourerCard() game.Card {
	return game.Card{Name: "Devourer of Destiny", TypeLine: "Creature — Eldrazi", ManaCost: "{5}{C}{C}", OracleID: gap2B13Devourer, Power: 6, Toughness: 6}
}

// Revealed by the starting player: the trigger resolves in their first
// upkeep, they keep one of the top four on top and the other three are
// exiled. The Devourer stays in hand.
func TestGap2B13DevourerRevealedLooksAtFourKeepsOneExilesTheRest(t *testing.T) {
	g := effectstest.OpeningTable(t)
	starter := g.Seats[g.StartingSeat]
	ids := effectstest.HandWith(g, g.StartingSeat, gap2B13DevourerCard(), effectstest.FillerCard("Spare"))
	effectstest.KeepAll(t, g)
	offer := gap2B13Pending(t, g, game.PendingChoiceConfirm, starter.ID)
	if err := g.ResolveConfirm(offer.ID, starter.ID, true); err != nil {
		t.Fatalf("reveal: %v", err)
	}
	if !starter.Hand.Contains(ids[0]) {
		t.Fatal("revealing the Devourer moved it")
	}
	for _, p := range g.Seats {
		c, _ := cardInHand(starter, ids[0])
		if !c.KnownBy[p.ID] {
			t.Errorf("%s did not see the revealed Devourer", p.Name)
		}
	}
	top := gap2B13TopIDs(starter, 4)
	size := starter.Library.Size()
	effectstest.PassPriorityAroundTable(t, g)
	pick := gap2B13Pending(t, g, game.PendingChoiceChooseCards, starter.ID)
	if len(pick.ChooseCards) != 4 {
		t.Fatalf("offered %d cards, want the top four", len(pick.ChooseCards))
	}
	if err := g.ResolveChooseCards(pick.ID, starter.ID, []uuid.UUID{top[2]}); err != nil {
		t.Fatalf("keep one: %v", err)
	}
	if got := gap2B13TopIDs(starter, 1); got[0] != top[2] {
		t.Error("the kept card is not on top of the library")
	}
	for _, id := range []uuid.UUID{top[0], top[1], top[3]} {
		if !effectstest.InExile(g, id) {
			t.Errorf("a card not kept was not exiled")
		}
	}
	if starter.Library.Size() != size-3 {
		t.Errorf("library %d, want %d", starter.Library.Size(), size-3)
	}
}

// "You may put one": keeping none exiles all four.
func TestGap2B13DevourerKeepingNoneExilesAllFour(t *testing.T) {
	g := effectstest.OpeningTable(t)
	starter := g.Seats[g.StartingSeat]
	effectstest.HandWith(g, g.StartingSeat, gap2B13DevourerCard())
	effectstest.KeepAll(t, g)
	if err := g.ResolveConfirm(gap2B13Pending(t, g, game.PendingChoiceConfirm, starter.ID).ID, starter.ID, true); err != nil {
		t.Fatal(err)
	}
	top := gap2B13TopIDs(starter, 4)
	effectstest.PassPriorityAroundTable(t, g)
	pick := gap2B13Pending(t, g, game.PendingChoiceChooseCards, starter.ID)
	if err := g.ResolveChooseCards(pick.ID, starter.ID, nil); err != nil {
		t.Fatal(err)
	}
	for _, id := range top {
		if !effectstest.InExile(g, id) {
			t.Error("a card was left when none was kept")
		}
	}
}

// Not revealed: nothing happens in the upkeep.
func TestGap2B13DevourerNotRevealedDoesNothing(t *testing.T) {
	g := effectstest.OpeningTable(t)
	starter := g.Seats[g.StartingSeat]
	effectstest.HandWith(g, g.StartingSeat, gap2B13DevourerCard())
	effectstest.KeepAll(t, g)
	if err := g.ResolveConfirm(gap2B13Pending(t, g, game.PendingChoiceConfirm, starter.ID).ID, starter.ID, false); err != nil {
		t.Fatal(err)
	}
	size := starter.Library.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 0 || starter.Library.Size() != size {
		t.Errorf("a hidden Devourer acted: prompts %+v, library %d", g.PendingChoices, starter.Library.Size())
	}
}

func TestGap2B13DevourerCastTriggerExilesAColoredPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	colored := effectstest.PushUginVictim(g, opp.ID, "A Green Bear", []string{"G"})
	colorless := effectstest.PushUginVictim(g, opp.ID, "A Colorless Golem", nil)
	id := uuid.New()
	c := gap2B13DevourerCard()
	c.InstanceID, c.Owner, c.Controller = id, me.ID, me.ID
	me.Hand.PushTop(c)
	effectstest.ToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if effectstest.HasID(p.PickTargetCards, colorless) {
		t.Error("a colorless permanent is not a legal target")
	}
	effectstest.PickCard(t, g, me.ID, colored)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, colored) {
		t.Error("the coloured permanent was not exiled")
	}
	if !g.Battlefield.Contains(id) {
		t.Error("the Devourer resolves after its cast trigger")
	}
}

// --- Bigger on the Inside ---

func gap2B13PutBigger(t *testing.T, g *game.Game, owner uuid.UUID) (uuid.UUID, int, string) {
	t.Helper()
	land := effectstest.GaPut(g, owner, "Forest", "Basic Land — Forest", "")
	effectstest.PushAuraOnLand(t, g, owner, "Bigger on the Inside", gap2B13Bigger, land)
	idx, ref := effectstest.GrantedActivatedIndex(t, g, land)
	if idx < 0 {
		t.Fatal("the enchanted land has no granted ability")
	}
	return land, idx, ref
}

func TestGap2B13BiggerOnTheInsideManaAndCascadeForYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedCheapLibrary(g, me, 6)
	land, idx, ref := gap2B13PutBigger(t, g, me.ID)
	if err := g.ActivateCatalogAbility(me.ID, land, idx, game.ActivateAbilityParams{
		Ref: ref, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := gap2B13Pending(t, g, game.PendingChoiceMana, me.ID)
	if err := g.ResolveManaChoice(pick.ID, me.ID, "G"); err != nil {
		t.Fatalf("pick: %v", err)
	}
	if got := me.ManaPool.Total(); got != 2 {
		t.Fatalf("pool %d, want 2", got)
	}
	effectstest.B06AddMana(me, "C", "C")
	effectstest.CastWithCost(t, g, "First Sorcery", "Sorcery", "{3}{G}", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountMayCastPrompts(g); got != 1 {
		t.Fatalf("the next spell got %d cascade offers, want 1", got)
	}
	effectstest.AnswerAllMayCast(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B06AddMana(me, "G", "C", "C", "C")
	effectstest.CastWithCost(t, g, "Second Sorcery", "Sorcery", "{3}{G}", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountMayCastPrompts(g); got != 0 {
		t.Errorf("the spell after it got %d cascade offers, want 0", got)
	}
}

// Aimed at an opponent: they pick the colour, the mana is theirs, the
// cascade promise is theirs, and your own next spell does not cascade.
func TestGap2B13BiggerOnTheInsideTargetingAnOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedCheapLibrary(g, me, 6)
	land, idx, ref := gap2B13PutBigger(t, g, me.ID)
	if err := g.ActivateCatalogAbility(me.ID, land, idx, game.ActivateAbilityParams{
		Ref: ref, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := gap2B13Pending(t, g, game.PendingChoiceMana, opp.ID)
	if err := g.ResolveManaChoice(pick.ID, opp.ID, "U"); err != nil {
		t.Fatalf("pick: %v", err)
	}
	if opp.ManaPool.Total() != 2 || me.ManaPool.Total() != 0 {
		t.Fatalf("pools: opponent %d, me %d", opp.ManaPool.Total(), me.ManaPool.Total())
	}
	waiting := 0
	for _, dt := range g.DelayedTriggers {
		if dt.Controller == opp.ID {
			waiting++
		}
	}
	if waiting != 1 {
		t.Errorf("%d cascade promises for the opponent, want 1", waiting)
	}
	effectstest.B06AddMana(me, "G", "C", "C", "C")
	effectstest.CastWithCost(t, g, "My Sorcery", "Sorcery", "{3}{G}", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountMayCastPrompts(g); got != 0 {
		t.Errorf("my spell got %d cascade offers, want 0", got)
	}
}

func TestGap2B13BiggerOnTheInsideEnchantsAnArtifactOrALandOnly(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		legal          bool
	}{
		{"a creature", "Creature — Bear", false},
		{"a land", "Basic Land — Forest", true},
		{"an artifact", "Artifact", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.ToMain(t, g)
			host := effectstest.GaPut(g, me.ID, "Host", tc.typeLine, "")
			id := uuid.New()
			me.Hand.PushTop(game.Card{InstanceID: id, Name: "Bigger on the Inside", TypeLine: "Enchantment — Aura", ManaCost: "{3}{R}{G}", OracleID: gap2B13Bigger, Owner: me.ID, Controller: me.ID})
			effectstest.B06AddMana(me, "R", "G", "C", "C", "C")
			err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: host}}})
			if (err == nil) != tc.legal {
				t.Errorf("enchanting %s: err %v, legal %v", tc.name, err, tc.legal)
			}
		})
	}
}

// --- Lim-Dûl's Vault ---

func gap2B13CastVault(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Lim-Dûl's Vault", TypeLine: "Instant", ManaCost: "{U}{B}", OracleID: gap2B13Vault, Owner: me.ID, Controller: me.ID})
	effectstest.B06AddMana(me, "U", "B")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func gap2B13Library(me *game.Player, n int) {
	me.Library.Cards = nil
	for i := 0; i < n; i++ {
		c := game.NewCard("Card", me.ID)
		me.Library.PushTop(c)
	}
}

// Pay once, then stop: the second five stay on top in the order chosen,
// the first five went to the bottom and then into the shuffle, and the
// payment cost 1 life.
func TestGap2B13LimDulsVaultPaysOnceThenKeepsTheSecondFiveOnTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	gap2B13Library(me, 15)
	first := gap2B13TopIDs(me, 5)
	second := gap2B13TopIDs(me, 10)[5:]
	life := me.Life
	gap2B13CastVault(t, g, me)

	look := gap2B13Pending(t, g, game.PendingChoiceLookAtTop, me.ID)
	if err := g.ResolveLookAtTop(look.ID, me.ID, first); err != nil {
		t.Fatalf("first look: %v", err)
	}
	ask := gap2B13Pending(t, g, game.PendingChoiceConfirm, me.ID)
	if err := g.ResolveConfirm(ask.ID, me.ID, true); err != nil {
		t.Fatalf("pay: %v", err)
	}
	if me.Life != life-1 {
		t.Fatalf("life %d, want %d", me.Life, life-1)
	}
	bottom := gap2B13Pending(t, g, game.PendingChoicePutInLibrary, me.ID)
	if err := g.ResolvePutInLibrary(bottom.ID, me.ID, first, nil); err != nil {
		t.Fatalf("bottom: %v", err)
	}
	look = gap2B13Pending(t, g, game.PendingChoiceLookAtTop, me.ID)
	order := []uuid.UUID{second[4], second[3], second[2], second[1], second[0]}
	if err := g.ResolveLookAtTop(look.ID, me.ID, order); err != nil {
		t.Fatalf("second look: %v", err)
	}
	ask = gap2B13Pending(t, g, game.PendingChoiceConfirm, me.ID)
	if err := g.ResolveConfirm(ask.ID, me.ID, false); err != nil {
		t.Fatalf("stop: %v", err)
	}
	if me.Life != life-1 {
		t.Errorf("stopping cost life: %d", me.Life)
	}
	if me.Library.Size() != 15 {
		t.Fatalf("library %d, want 15", me.Library.Size())
	}
	got := gap2B13TopIDs(me, 5)
	for i := range order {
		if got[i] != order[i] {
			t.Fatalf("top five %v, want %v in the order chosen", got, order)
		}
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("prompts left: %+v", g.PendingChoices)
	}
}

// Stopping at once costs nothing and keeps the first five on top.
func TestGap2B13LimDulsVaultStopAtOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	gap2B13Library(me, 12)
	first := gap2B13TopIDs(me, 5)
	life := me.Life
	gap2B13CastVault(t, g, me)
	look := gap2B13Pending(t, g, game.PendingChoiceLookAtTop, me.ID)
	if err := g.ResolveLookAtTop(look.ID, me.ID, first); err != nil {
		t.Fatal(err)
	}
	if err := g.ResolveConfirm(gap2B13Pending(t, g, game.PendingChoiceConfirm, me.ID).ID, me.ID, false); err != nil {
		t.Fatal(err)
	}
	if me.Life != life {
		t.Errorf("life %d, want %d", me.Life, life)
	}
	got := gap2B13TopIDs(me, 5)
	for i := range first {
		if got[i] != first[i] {
			t.Fatal("the first five are not on top in the order chosen")
		}
	}
}

// A player at 0 life can't pay 1 life (CR 119.4), so is not asked.
func TestGap2B13LimDulsVaultNotAskedWithoutLifeToPay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	gap2B13Library(me, 12)
	first := gap2B13TopIDs(me, 5)
	gap2B13CastVault(t, g, me)
	me.Life = 0
	look := gap2B13Pending(t, g, game.PendingChoiceLookAtTop, me.ID)
	if err := g.ResolveLookAtTop(look.ID, me.ID, first); err != nil {
		t.Fatal(err)
	}
	for _, c := range g.PendingChoices {
		if c.Kind == game.PendingChoiceConfirm && c.Chooser == me.ID {
			t.Fatal("asked to pay life the player does not have")
		}
	}
}

func TestGap2B13LimDulsVaultNeedsBlue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Lim-Dûl's Vault", TypeLine: "Instant", ManaCost: "{U}{B}", OracleID: gap2B13Vault, Owner: me.ID, Controller: me.ID})
	effectstest.B06AddMana(me, "B", "B")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast without blue mana")
	}
}

func cardInHand(p *game.Player, id uuid.UUID) (game.Card, bool) {
	for _, c := range p.Hand.Cards {
		if c.InstanceID == id {
			return c, true
		}
	}
	return game.Card{}, false
}
