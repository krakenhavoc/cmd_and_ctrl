package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	oldB449Scourge   = "1e02f371-5dce-415a-8563-1f93c35401d5"
	oldB449Blindness = "705531b3-77cc-4d1f-851b-45e03c029991"
	oldB449Scarecrow = "87b61c1f-18e3-446a-a732-cff40adc5e05"
	oldB449Subject   = "53175678-1f5c-4889-9086-66381b3ddc3a"
	oldB449Neonate   = "7319eec8-48ba-4238-bce1-450abcae73a6"
	oldB449Cannibals = "973504d6-e2e4-49c4-86cb-69f31d66388b"
	oldB449Spare     = "557cc9a7-c8cc-4823-b108-2dfe648c789d"
	oldB449Terrors   = "6d720de8-b7b7-478f-ab1a-3487ce76aa00"
	oldB449Blood     = "7f1a9248-8e7b-46a8-942b-ca18c7da8c4a"
	oldB449Witch     = "67466d56-9f28-4bbe-ae54-6fd834581e77"
	oldB449Scholar   = "8662c9be-2244-4ea3-bf80-31a117f40a59"
	oldB449Sturm     = "b3045ee5-232b-4e63-8ea8-b5e46b49e71a"
)

func TestOldB449CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB449Scourge: "Scourge of Geier Reach", oldB449Blindness: "Hysterical Blindness",
		oldB449Scarecrow: "One-Eyed Scarecrow", oldB449Subject: "Ludevic's Test Subject",
		oldB449Subject + "#1": "Ludevic's Abomination", oldB449Neonate: "Bloodcrazed Neonate",
		oldB449Cannibals: "Village Cannibals", oldB449Spare: "Spare from Evil",
		oldB449Terrors: "Night Terrors", oldB449Blood: "Traitorous Blood",
		oldB449Witch: "Bitterheart Witch", oldB449Scholar: "Civilized Scholar",
		oldB449Scholar + "#1": "Homicidal Brute", oldB449Sturm: "Sturmgeist",
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

func oldB449Size(g *game.Game, id uuid.UUID) (int, int) {
	c := effectstest.ApaLive(g, id)
	return c.PowerForComparison(), c.CurrentToughness()
}

func TestOldB449ScourgeCountsOnlyOpponentsCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	s := effectstest.B12Push(g, me.ID, "Scourge of Geier Reach", "Creature — Elemental", oldB449Scourge, 3, 3)
	effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	if p, tough := oldB449Size(g, s); p != 3 || tough != 3 {
		t.Fatalf("alone it is %d/%d, want 3/3 (your own creatures do not count)", p, tough)
	}
	effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, g.Seats[2].ID, "Theirs too", "Creature — Bear", 2, 2)
	effectstest.B12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	if p, tough := oldB449Size(g, s); p != 5 || tough != 5 {
		t.Fatalf("with two opposing creatures it is %d/%d, want 5/5 (a land is not a creature)", p, tough)
	}
}

func TestOldB449HystericalBlindnessShrinksOnlyOpponentsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 5, 5)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 5, 5)
	effectstest.CastCatalogSpell(t, g, "Hysterical Blindness", "Instant", oldB449Blindness, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := oldB449Size(g, theirs); p != 1 || tough != 5 {
		t.Errorf("opposing creature is %d/%d, want 1/5", p, tough)
	}
	if p, _ := oldB449Size(g, mine); p != 5 {
		t.Errorf("your creature has power %d, want 5", p)
	}
	effectstest.AdvancePastCleanupForTest(t, g)
	if p, _ := oldB449Size(g, theirs); p != 5 {
		t.Errorf("the effect lasted past the turn: power %d", p)
	}
}

func TestOldB449ScarecrowShrinksOnlyOpposingFliers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "One-Eyed Scarecrow", "Artifact Creature — Scarecrow", oldB449Scarecrow, 2, 3)
	flier := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Bird", TypeLine: "Creature — Bird", Power: 2, Toughness: 2, Keywords: []string{"flying"}})
	walker := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	myFlier := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Mine", TypeLine: "Creature — Bird", Power: 2, Toughness: 2, Keywords: []string{"flying"}})
	if p, tough := oldB449Size(g, flier); p != 1 || tough != 2 {
		t.Errorf("opposing flier is %d/%d, want 1/2", p, tough)
	}
	if p, _ := oldB449Size(g, walker); p != 2 {
		t.Errorf("opposing ground creature has power %d, want 2", p)
	}
	if p, _ := oldB449Size(g, myFlier); p != 2 {
		t.Errorf("your own flier has power %d, want 2", p)
	}
}

func oldB449Row(name string, oracle string, front, back game.Face) func(uuid.UUID) game.Card {
	return func(owner uuid.UUID) game.Card {
		c := game.Card{
			InstanceID: uuid.New(), OracleID: oracle, Layout: game.LayoutTransform,
			Owner: owner, Controller: owner, Faces: []game.Face{front, back},
		}
		c.SetFace(0)
		return c
	}
}

func oldB449Name(g *game.Game, id uuid.UUID) string {
	c, ok := effectstest.BattlefieldCard(g, id)
	if !ok {
		return "(gone)"
	}
	return c.Name
}

func oldB449Activate(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID, mana string) error {
	t.Helper()
	effectstest.AdvanceToMainOf(t, g, g.Turn.ActiveSeat)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	if mana != "" {
		effectstest.DcMana(t, g, me.ID, mana)
	}
	err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true})
	if err == nil {
		effectstest.PassPriorityAroundTable(t, g)
	}
	return err
}

func TestOldB449TestSubjectHatchesOnTheFifthCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mk := oldB449Row("subject", oldB449Subject,
		game.Face{Name: "Ludevic's Test Subject", TypeLine: "Creature — Lizard Egg", ManaCost: "{1}{U}", Colors: []string{"U"}, Power: 0, Toughness: 3},
		game.Face{Name: "Ludevic's Abomination", TypeLine: "Creature — Lizard Horror", Colors: []string{"U"}, Power: 13, Toughness: 13})
	id := effectstest.PushBattlefieldCardWithTimestamp(g, mk(me.ID))
	if err := oldB449Activate(t, g, me, id, "{C}"); err == nil {
		t.Fatal("one mana paid for {1}{U}")
	}
	for i := 1; i <= 4; i++ {
		if err := oldB449Activate(t, g, me, id, "{U}{U}"); err != nil {
			t.Fatalf("activation %d: %v", i, err)
		}
		if got := effectstest.CountersOn(g, id, "hatchling"); got != i {
			t.Fatalf("after %d activations: %d hatchling counters", i, got)
		}
		if got := oldB449Name(g, id); got != "Ludevic's Test Subject" {
			t.Fatalf("hatched early after %d: %q", i, got)
		}
	}
	if err := oldB449Activate(t, g, me, id, "{U}{U}"); err != nil {
		t.Fatalf("fifth activation: %v", err)
	}
	if got := oldB449Name(g, id); got != "Ludevic's Abomination" {
		t.Fatalf("face = %q, want Ludevic's Abomination", got)
	}
	if got := effectstest.CountersOn(g, id, "hatchling"); got != 0 {
		t.Errorf("%d hatchling counters left, want all removed", got)
	}
	if p, tough := oldB449Size(g, id); p != 13 || tough != 13 {
		t.Errorf("Abomination is %d/%d", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, id, "trample") {
		t.Error("Abomination has no trample")
	}
}

func TestOldB449NeonateMustAttackAndGrowsOnDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	n := effectstest.B12Push(g, me.ID, "Bloodcrazed Neonate", "Creature — Vampire", oldB449Neonate, 2, 1)
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
	if err := g.PassPriority(); err == nil {
		t.Fatal("passing with the Neonate home succeeded, want the attack requirement to refuse it")
	}
	if err := g.DeclareAttacker(n, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, n, game.CounterPlusOne); got != 1 {
		t.Errorf("Neonate has %d +1/+1 counters, want 1", got)
	}
	if p, tough := oldB449Size(g, n); p != 3 || tough != 2 {
		t.Errorf("Neonate is %d/%d, want 3/2", p, tough)
	}
}

func TestOldB449CannibalsGrowOnlyWhenAnotherHumanDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	c := effectstest.B12Push(g, me.ID, "Village Cannibals", "Creature — Human", oldB449Cannibals, 2, 2)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.KillCreature(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, c, game.CounterPlusOne); got != 0 {
		t.Fatalf("a Bear dying gave %d counters", got)
	}
	human := effectstest.B12Creature(g, opp.ID, "Peasant", "Creature — Human Peasant", 1, 1)
	effectstest.KillCreature(t, g, opp.ID, human)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, c, game.CounterPlusOne); got != 1 {
		t.Fatalf("an opponent's Human dying gave %d counters, want 1", got)
	}
	effectstest.KillCreature(t, g, me.ID, c)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ZoneOfCard(g, c) == game.ZoneBattlefield {
		t.Fatal("the Cannibals did not die")
	}
}

func TestOldB449SpareFromEvilProtectsFromNonHumansOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Elf", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Elf", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Spare from Evil", "Instant", oldB449Spare, nil)
	effectstest.PassPriorityAroundTable(t, g)
	tok := game.ProtectionFromNonSubtypeCreatures("Human")
	has := func(id uuid.UUID) bool {
		c := effectstest.ApaLive(g, id)
		return game.HasKeyword(c, tok)
	}
	if !has(mine) {
		t.Error("your creature did not gain protection from non-Human creatures")
	}
	if has(theirs) {
		t.Error("an opponent's creature gained protection")
	}
	effectstest.AdvancePastCleanupForTest(t, g)
	if has(mine) {
		t.Error("the protection outlived the turn")
	}
}

func TestOldB449NightTerrorsExilesANonlandCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ids := effectstest.RevealHand(opp, effectstest.RhForest(), effectstest.RhBolt())
	effectstest.CastCatalogSpell(t, g, "Night Terrors", "Sorcery", oldB449Terrors, []game.TargetRef{effectstest.PbPlayer(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	var pick *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceRevealedHandPick {
			pick = c
		}
	}
	if pick == nil {
		t.Fatal("no revealed-hand pick was offered")
	}
	if err := g.ResolvePendingChoice(pick.ID, me.ID, []uuid.UUID{ids[1]}); err != nil {
		t.Fatalf("resolve: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(ids[1]) {
		t.Error("the nonland card was not exiled")
	}
	if !effectstest.HandHas(opp, ids[0]) {
		t.Error("the land left the hand: it was never a legal choice")
	}
}

func TestOldB449TraitorousBloodStealsUntapsAndGivesTrampleAndHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.CtrlPushCreature(g, opp.ID, "Bear")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == victim {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	effectstest.CastCatalogSpell(t, g, "Traitorous Blood", "Sorcery", oldB449Blood, []game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.ControllerOf(t, g, victim); got != me.ID {
		t.Fatalf("controller %s, want the caster", got)
	}
	if tapped, _ := effectstest.BattlefieldCardTapped(g, victim); tapped {
		t.Error("the creature stayed tapped")
	}
	for _, kw := range []string{"trample", "haste"} {
		if !effectstest.HasEffectiveKeyword(t, g, victim, kw) {
			t.Errorf("the creature did not gain %s", kw)
		}
	}
	effectstest.AdvancePastCleanupForTest(t, g)
	if got := effectstest.ControllerOf(t, g, victim); got != opp.ID {
		t.Error("the creature was not given back at cleanup")
	}
	if effectstest.HasEffectiveKeyword(t, g, victim, "trample") {
		t.Error("trample outlived the turn")
	}
}

func TestOldB449SturmgeistIsAsBigAsYourHandAndDrawsOnDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	s := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Sturmgeist", TypeLine: "Creature — Spirit", OracleID: oldB449Sturm,
		Owner: me.ID, Controller: me.ID,
	})
	size := me.Hand.Size()
	if p, tough := oldB449Size(g, s); p != size || tough != size {
		t.Fatalf("Sturmgeist is %d/%d with %d cards in hand", p, tough, size)
	}
	if !effectstest.HasEffectiveKeyword(t, g, s, "flying") {
		t.Error("no flying")
	}
	effectstest.DealCombatDamageToPlayer(g, s, g.Seats[1].ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != size+1 {
		t.Errorf("hand %d, want %d after the combat-damage draw", got, size+1)
	}
}

func TestOldB449WitchFetchesACurseOntoThePlayerOnDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	w := effectstest.B12Push(g, me.ID, "Bitterheart Witch", "Creature — Human Shaman", oldB449Witch, 1, 2)
	curse := effectstest.PushLibraryCard(me, game.Card{
		Name: "Curse of Test", TypeLine: "Enchantment — Aura Curse", Owner: me.ID, Controller: me.ID,
	})
	effectstest.KillCreature(t, g, me.ID, w)
	effectstest.AnswerTriggerTargets2(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.AnswerSearchByID(t, g, me.ID, curse)
	effectstest.AutoAnswerAll(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, curse)
	if !ok {
		t.Fatalf("the Curse is in zone %v, want the battlefield", effectstest.ZoneOfCard(g, curse))
	}
	if c.AttachedTo.Kind != game.TargetPlayer || c.AttachedTo.ID != opp.ID {
		t.Errorf("the Curse is attached to %+v, want the targeted player", c.AttachedTo)
	}
}

func TestOldB449ScholarLootsAndTurnsOverOnlyForACreatureCard(t *testing.T) {
	for _, tc := range []struct {
		name     string
		discard  game.Card
		wantFace string
	}{
		{"creature card", game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2}, "Homicidal Brute"},
		{"other card", game.Card{Name: "Bolt", TypeLine: "Instant", ManaCost: "{R}"}, "Civilized Scholar"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			mk := oldB449Row("scholar", oldB449Scholar,
				game.Face{Name: "Civilized Scholar", TypeLine: "Creature — Human Advisor", ManaCost: "{2}{U}", Colors: []string{"U"}, Power: 0, Toughness: 1},
				game.Face{Name: "Homicidal Brute", TypeLine: "Creature — Human Mutant", Colors: []string{"U"}, Power: 5, Toughness: 1})
			card := mk(me.ID)
			card.SummonedThisTurn = false
			id := effectstest.PushBattlefieldCardWithTimestamp(g, card)
			// Hand of one card and a library whose top card is the discard.
			g.WithWriteLock(func() {
				for len(me.Hand.Cards) > 0 {
					me.Hand.Cards = me.Hand.Cards[1:]
				}
			})
			tc.discard.Owner, tc.discard.Controller = me.ID, me.ID
			top := effectstest.PushLibraryCard(me, tc.discard)
			if err := oldB449Activate(t, g, me, id, ""); err != nil {
				t.Fatalf("activate: %v", err)
			}
			effectstest.AutoAnswerAll(t, g)
			effectstest.PassPriorityAroundTable(t, g)
			for _, c := range g.PendingChoices {
				if c != nil && c.Kind == game.PendingChoiceChooseCards && c.Chooser == me.ID {
					if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{top}); err != nil {
						t.Fatalf("discard: %v", err)
					}
					break
				}
			}
			effectstest.PassPriorityAroundTable(t, g)
			if !effectstest.InGraveyard(me, top) {
				t.Fatalf("the drawn card is in zone %v, want the graveyard (looted away)", effectstest.ZoneOfCard(g, top))
			}
			if got := oldB449Name(g, id); got != tc.wantFace {
				t.Errorf("face = %q, want %q", got, tc.wantFace)
			}
			if tapped, _ := effectstest.BattlefieldCardTapped(g, id); tc.wantFace == "Homicidal Brute" && tapped {
				t.Error("the Scholar did not untap before transforming")
			} else if tc.wantFace == "Civilized Scholar" && !tapped {
				t.Error("the Scholar untapped without a creature card")
			}
		})
	}
}

func TestOldB449BruteTapsAndTransformsAtEndStepOnlyIfItDidNotAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mk := oldB449Row("brute", oldB449Scholar,
		game.Face{Name: "Civilized Scholar", TypeLine: "Creature — Human Advisor", ManaCost: "{2}{U}", Colors: []string{"U"}, Power: 0, Toughness: 1},
		game.Face{Name: "Homicidal Brute", TypeLine: "Creature — Human Mutant", Colors: []string{"U"}, Power: 5, Toughness: 1})
	card := mk(me.ID)
	card.SetFace(1)
	id := effectstest.PushBattlefieldCardWithTimestamp(g, card)
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB449Name(g, id); got != "Civilized Scholar" {
		t.Errorf("face = %q, want Civilized Scholar", got)
	}
	if tapped, _ := effectstest.BattlefieldCardTapped(g, id); !tapped {
		t.Error("the Brute was not tapped")
	}
}

func TestOldB449BruteStaysIfItAttacked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mk := oldB449Row("brute", oldB449Scholar,
		game.Face{Name: "Civilized Scholar", TypeLine: "Creature — Human Advisor", ManaCost: "{2}{U}", Colors: []string{"U"}, Power: 0, Toughness: 1},
		game.Face{Name: "Homicidal Brute", TypeLine: "Creature — Human Mutant", Colors: []string{"U"}, Power: 5, Toughness: 1})
	card := mk(me.ID)
	card.SetFace(1)
	id := effectstest.PushBattlefieldCardWithTimestamp(g, card)
	effectstest.DeclareAttack(t, g, opp.ID, id)
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB449Name(g, id); got != "Homicidal Brute" {
		t.Errorf("face = %q, want Homicidal Brute (it attacked)", got)
	}
}
