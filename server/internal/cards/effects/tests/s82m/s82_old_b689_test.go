package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b689_test.go — Mercadian Masques cards, S82 OLD batch 689 (#5619),
// played through the real catalog entries.

const (
	o689RobberFly    = "0d44d4f3-9713-4d18-a985-9e7ff0df85d3"
	o689Nautilus     = "0e1267d4-0155-4503-829a-03264c8d14a0"
	o689Atlas        = "5bae9ffb-99b3-4feb-8aed-265d874fd68b"
	o689Horn         = "72d3b499-dbff-4631-b6e3-0a292b39d527"
	o689Voucher      = "8496a8f1-f505-403a-8007-c35aedb90e2d"
	o689Heir         = "8685ebfd-d04d-47f5-a919-e118db1e2276"
	o689Djinn        = "8ccb3bca-73da-450d-a9d0-61d3e2bd05a5"
	o689Hall         = "8da9f975-519d-455a-80c1-2b468cad79a1"
	o689Summons      = "910ff092-7c9d-49d3-a6df-497683e45bbd"
	o689Pathfinder   = "ac00d3db-d9e4-4fc3-a4ea-dd1187f406bb"
	o689CustomsDepot = "de64eab5-9d7a-4408-8358-47d30a14ce40"
	o689LandGrant    = "e36f4649-934d-4e8a-bd5f-39f164ea6937"
	o689Bifurcate    = "f21ec455-aaf6-4842-aea2-69a7fd99b80f"
)

func TestS82OldB689CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		o689RobberFly: "Robber Fly", o689Nautilus: "Chambered Nautilus", o689Atlas: "Mercadian Atlas",
		o689Horn: "Horn of Plenty", o689Voucher: "Credit Voucher", o689Heir: "Saprazzan Heir",
		o689Djinn: "Indentured Djinn", o689Hall: "Assembly Hall", o689Summons: "Cateran Summons",
		o689Pathfinder: "Silverglade Pathfinder", o689CustomsDepot: "Customs Depot", o689LandGrant: "Land Grant",
		o689Bifurcate: "Bifurcate",
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

// o689BlockedAttacker has `attacker` (seat 0's) attack seat 1 and be blocked
// by a flying creature, then settles to the first prompt.
func o689BlockedAttacker(t *testing.T, g *game.Game, attacker uuid.UUID) {
	t.Helper()
	opp := g.Seats[1]
	blocker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Flier", TypeLine: "Creature — Bird", Power: 0, Toughness: 9,
		Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, attacker); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.InitSettle(t, g)
}

func TestS82OldB689RobberFlyMakesTheDefenderDiscardTheHandAndDrawThatMany(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fly := effectstest.PushCatalogPermanent(g, me.ID, "Robber Fly", "Creature — Insect", o689RobberFly, false)
	effectstest.D2dSetHand(opp, 3)
	var old []uuid.UUID
	for _, c := range opp.Hand.Cards {
		old = append(old, c.InstanceID)
	}
	effectstest.SeedLibrary(opp, "A", "B", "C", "D", "E")
	mine := me.Hand.Size()
	o689BlockedAttacker(t, g, fly)
	if opp.Hand.Size() != 3 {
		t.Fatalf("defender's hand = %d, want 3 (discarded 3, drew 3)", opp.Hand.Size())
	}
	for _, id := range old {
		if opp.Hand.Contains(id) {
			t.Errorf("an original card %s is still in the defender's hand", id)
		}
		if !opp.Graveyard.Contains(id) {
			t.Errorf("an original card %s did not reach the graveyard", id)
		}
	}
	if me.Hand.Size() != mine {
		t.Errorf("the attacker's controller's hand changed: %d -> %d", mine, me.Hand.Size())
	}
}

func TestS82OldB689RobberFlyWithAnEmptyHandDrawsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fly := effectstest.PushCatalogPermanent(g, me.ID, "Robber Fly", "Creature — Insect", o689RobberFly, false)
	opp.Hand.Cards = nil
	effectstest.SeedLibrary(opp, "A", "B")
	o689BlockedAttacker(t, g, fly)
	if opp.Hand.Size() != 0 {
		t.Errorf("an empty hand drew %d cards", opp.Hand.Size())
	}
}

func TestS82OldB689NautilusAndHeirDrawWhenBlockedIfYouWish(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, typeLine string
		draws                  int
	}{
		{"Chambered Nautilus", o689Nautilus, "Creature — Nautilus Beast", 1},
		{"Saprazzan Heir", o689Heir, "Creature — Merfolk", 3},
	} {
		for _, yes := range []bool{true, false} {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			c := effectstest.PushCatalogPermanent(g, me.ID, tc.name, tc.typeLine, tc.oracle, false)
			me.Hand.Cards = nil
			effectstest.SeedLibrary(me, "A", "B", "C", "D")
			o689BlockedAttacker(t, g, c)
			effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, yes)
			effectstest.PassPriorityAroundTable(t, g)
			want := 0
			if yes {
				want = tc.draws
			}
			if me.Hand.Size() != want {
				t.Errorf("%s (yes=%v): hand = %d, want %d", tc.name, yes, me.Hand.Size(), want)
			}
		}
	}
}

func TestS82OldB689NautilusDoesNotTriggerUnblocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	c := effectstest.PushCatalogPermanent(g, me.ID, "Chambered Nautilus", "Creature — Nautilus Beast", o689Nautilus, false)
	me.Hand.Cards = nil
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(c, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	effectstest.LockInBlocks(t, g)
	effectstest.InitSettle(t, g)
	if len(g.PendingChoices) != 0 || me.Hand.Size() != 0 {
		t.Errorf("an unblocked Nautilus triggered: %d prompts, hand %d", len(g.PendingChoices), me.Hand.Size())
	}
}

func TestS82OldB689IndenturedDjinnLetsEachOtherPlayerDrawUpToThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for _, s := range g.Seats {
		s.Hand.Cards = nil
		effectstest.SeedLibrary(s, "A", "B", "C", "D", "E")
	}
	effectstest.AdvanceToMain(t, g)
	djinn := effectstest.HandCardOf(g, me, "Indentured Djinn", "Creature — Djinn", "", o689Djinn)
	if err := g.CastSpell(me.ID, djinn, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for i, seat := range []int{1, 2, 3} {
		effectstest.AnswerOptionPick(t, g, g.Seats[seat].ID, []int{3, 0, 2}[i])
	}
	for seat, want := range map[int]int{0: 0, 1: 3, 2: 0, 3: 2} {
		if got := g.Seats[seat].Hand.Size(); got != want {
			t.Errorf("seat %d hand = %d, want %d", seat, got, want)
		}
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, djinn, "flying") {
		t.Error("the Djinn lacks flying")
	}
}

func TestS82OldB689CateranSummonsFindsAMercenary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Wolf", TypeLine: "Creature — Wolf"},
		game.Card{Name: "Hired Hand", TypeLine: "Creature — Human Mercenary"})
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "B")
	id := effectstest.HandCardOf(g, me, "Cateran Summons", "Sorcery", "{B}", o689Summons)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, ids[1])
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[1]) {
		t.Error("the Mercenary is not in hand")
	}
	if me.Hand.Contains(ids[0]) {
		t.Error("a non-Mercenary was fetched")
	}
}

func TestS82OldB689SilvergladePathfinderPaysDiscardAndFetchesATappedBasic(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	pf := effectstest.PushCatalogPermanent(g, me.ID, "Silverglade Pathfinder", "Creature — Dryad Spellshaper", o689Pathfinder, false)
	discard := effectstest.HandCardOf(g, me, "Spare", "Creature — Bear", "", "")
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "CG")
	if err := g.ActivateCatalogAbility(me.ID, pf, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{discard}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, ids[1])
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(discard) {
		t.Error("the discard cost was not paid")
	}
	if !effectstest.TappedOnBattlefield(t, g, ids[1]) {
		t.Error("the fetched land did not enter tapped")
	}
	if !effectstest.TappedOnBattlefield(t, g, pf) {
		t.Error("the Pathfinder was not tapped as a cost")
	}
}

func TestS82OldB689LandGrantRevealsHandOnlyWithNoLandInHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	me.Hand.Cards = nil
	forestInHand := effectstest.HandCardOf(g, me, "Forest", "Basic Land — Forest", "", "")
	grant := effectstest.HandCardOf(g, me, "Land Grant", "Sorcery", "{1}{G}", o689LandGrant)
	params := game.CastSpellParams{AlternativeCost: "land-grant", Strict: true}
	if err := g.CastSpell(me.ID, grant, params); err == nil {
		t.Fatal("Land Grant was cast for free with a land card in hand")
	}
	g.WithWriteLock(func() { me.Hand.Remove(forestInHand) })
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Island", TypeLine: "Basic Land — Island"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	if err := g.CastSpell(me.ID, grant, params); err != nil {
		t.Fatalf("free cast with no land in hand: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, ids[1])
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[1]) {
		t.Error("the Forest is not in hand")
	}
}

func TestS82OldB689LandGrantCanBePaidForNormally(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.HandCardOf(g, me, "Forest", "Basic Land — Forest", "", "")
	effectstest.FloatForTest(g, me, "CG")
	grant := effectstest.HandCardOf(g, me, "Land Grant", "Sorcery", "{1}{G}", o689LandGrant)
	ids := effectstest.SeedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	if err := g.CastSpell(me.ID, grant, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("paid cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, ids[0])
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[0]) {
		t.Error("the Forest is not in hand")
	}
}

func TestS82OldB689BifurcatePutsAPermanentWithTheTargetsNameOntoTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	target := effectstest.B12Push(g, opp.ID, "Grizzly Bears", "Creature — Bear", "", 2, 2)
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear", Power: 2, Toughness: 2},
		game.Card{Name: "Wolf", TypeLine: "Creature — Wolf"})
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "CCCG")
	id := effectstest.HandCardOf(g, me, "Bifurcate", "Sorcery", "{3}{G}", o689Bifurcate)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, ids[0])
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldCardForTest(g, ids[0]) == nil {
		t.Error("the Grizzly Bears did not enter the battlefield")
	}
	if effectstest.FindBattlefieldCardForTest(g, ids[1]) != nil {
		t.Error("a differently named card entered")
	}
}

func TestS82OldB689BifurcateRefusesATokenTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tok := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: "Token Creature — Bear", Power: 2, Toughness: 2,
		Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "CCCG")
	id := effectstest.HandCardOf(g, me, "Bifurcate", "Sorcery", "{3}{G}", o689Bifurcate)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: tok}}}); err == nil {
		t.Error("Bifurcate targeted a token")
	}
}

func TestS82OldB689CustomsDepotLootsForOneWhenYouCastACreature(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.PushCatalogPermanent(g, me.ID, "Customs Depot", "Enchantment", o689CustomsDepot, false)
		me.Hand.Cards = nil
		effectstest.SeedLibrary(me, "A", "B", "C")
		effectstest.AdvanceToMain(t, g)
		effectstest.FloatForTest(g, me, "CC")
		bear := effectstest.HandCardOf(g, me, "Bear", "Creature — Bear", "{1}", "")
		if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		if pay {
			effectstest.AnswerChooseCardsOrDiscard(t, g, me)
		}
		effectstest.PassPriorityAroundTable(t, g)
		want := 0
		if pay {
			want = 0 // drew one, discarded one
		}
		if me.Hand.Size() != want {
			t.Errorf("pay=%v: hand = %d, want %d", pay, me.Hand.Size(), want)
		}
		if got := len(me.Graveyard.Cards); pay && got != 1 {
			t.Errorf("graveyard = %d, want 1 discarded card", got)
		}
	}
}

func TestS82OldB689CustomsDepotIgnoresNoncreatureSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Customs Depot", "Enchantment", o689CustomsDepot, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "C")
	spell := effectstest.HandCardOf(g, me, "Cantrip", "Sorcery", "{1}", "")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("a noncreature spell asked for the payment")
	}
}

func TestS82OldB689MercadianAtlasDrawsOnlyWhenNoLandWasPlayed(t *testing.T) {
	for _, played := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.PushCatalogPermanent(g, me.ID, "Mercadian Atlas", "Artifact — Book", o689Atlas, false)
		me.Hand.Cards = nil
		effectstest.SeedLibrary(me, "A", "B", "C")
		effectstest.AdvanceToMain(t, g)
		if played {
			g.WithWriteLock(func() {
				if g.LandsPlayedThisTurn == nil {
					g.LandsPlayedThisTurn = map[uuid.UUID]int{}
				}
				g.LandsPlayedThisTurn[me.ID] = 1
			})
		}
		effectstest.AdvanceTo(t, g, game.StepEnd)
		effectstest.PassPriorityUntilChoice(t, g)
		if played {
			if me.Hand.Size() != 0 || len(g.PendingChoices) != 0 {
				t.Errorf("a land was played: hand %d, prompts %d", me.Hand.Size(), len(g.PendingChoices))
			}
			continue
		}
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
		effectstest.PassPriorityAroundTable(t, g)
		if me.Hand.Size() != 1 {
			t.Errorf("no land played: hand = %d, want 1", me.Hand.Size())
		}
	}
}

func TestS82OldB689HornOfPlentyDrawsAtTheNextEndStepForAPayingCaster(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.PushCatalogPermanent(g, me.ID, "Horn of Plenty", "Artifact", o689Horn, false)
		me.Hand.Cards = nil
		effectstest.SeedLibrary(me, "A", "B", "C")
		effectstest.AdvanceToMain(t, g)
		effectstest.FloatForTest(g, me, "CC")
		spell := effectstest.HandCardOf(g, me, "Cantrip", "Sorcery", "{1}", "")
		if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		if me.Hand.Size() != 0 {
			t.Fatalf("drew immediately: hand = %d", me.Hand.Size())
		}
		effectstest.AdvanceTo(t, g, game.StepEnd)
		effectstest.PassPriorityAroundTable(t, g)
		want := 0
		if pay {
			want = 1
		}
		if me.Hand.Size() != want {
			t.Errorf("pay=%v: hand after the end step = %d, want %d", pay, me.Hand.Size(), want)
		}
	}
}

func TestS82OldB689CreditVoucherShufflesChosenCardsAndDrawsThatMany(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	v := effectstest.PushCatalogPermanent(g, me.ID, "Credit Voucher", "Artifact", o689Voucher, false)
	effectstest.D2dSetHand(me, 3)
	pick := []uuid.UUID{me.Hand.Cards[0].InstanceID, me.Hand.Cards[1].InstanceID}
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E")
	libBefore := me.Library.Size()
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "CC")
	if err := g.ActivateCatalogAbility(me.ID, v, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if effectstest.FindBattlefieldCardForTest(g, v) != nil {
		t.Error("the Voucher was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, pick...)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != 3 {
		t.Errorf("hand = %d, want 3 (1 kept + 2 drawn)", me.Hand.Size())
	}
	if me.Library.Size() != libBefore {
		t.Errorf("library = %d, want %d (2 shuffled in, 2 drawn)", me.Library.Size(), libBefore)
	}
}

func TestS82OldB689CreditVoucherMayChooseNone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	v := effectstest.PushCatalogPermanent(g, me.ID, "Credit Voucher", "Artifact", o689Voucher, false)
	effectstest.D2dSetHand(me, 2)
	effectstest.SeedLibrary(me, "A", "B")
	libBefore := me.Library.Size()
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "CC")
	if err := g.ActivateCatalogAbility(me.ID, v, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != 2 || me.Library.Size() != libBefore {
		t.Errorf("hand %d, library %d: choosing nothing changed zones", me.Hand.Size(), me.Library.Size())
	}
}

func TestS82OldB689AssemblyHallRevealsACreatureAndFindsAnotherCopy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hall := effectstest.PushCatalogPermanent(g, me.ID, "Assembly Hall", "Artifact", o689Hall, false)
	me.Hand.Cards = nil
	wolf := effectstest.HandCardOf(g, me, "Wolf", "Creature — Wolf", "", "")
	effectstest.HandCardOf(g, me, "Shock", "Instant", "", "")
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Wolf", TypeLine: "Creature — Wolf"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "CCCC")
	if err := g.ActivateCatalogAbility(me.ID, hall, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, wolf)
	effectstest.AnswerSearchByID(t, g, me.ID, ids[0])
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[0]) || !me.Hand.Contains(wolf) {
		t.Error("both Wolves should be in hand")
	}
	if me.Hand.Contains(ids[1]) {
		t.Error("a differently named card was found")
	}
}

func TestS82OldB689AssemblyHallWithNoCreatureInHandDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hall := effectstest.PushCatalogPermanent(g, me.ID, "Assembly Hall", "Artifact", o689Hall, false)
	me.Hand.Cards = nil
	effectstest.HandCardOf(g, me, "Shock", "Instant", "", "")
	effectstest.SeedSearchLibrary(me, game.Card{Name: "Wolf", TypeLine: "Creature — Wolf"})
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "CCCC")
	if err := g.ActivateCatalogAbility(me.ID, hall, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 0 || me.Hand.Size() != 1 {
		t.Errorf("no creature to reveal, yet %d prompts and hand %d", len(g.PendingChoices), me.Hand.Size())
	}
}
