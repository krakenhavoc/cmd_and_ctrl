package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b452_test.go — OLD batch 452 (#5573): Magic 2012 cards Autumn's
// Veil, Monomania, Scrambleverse, Scepter of Empires, Crumbling Colossus,
// Goblin Bangchuckers, Mind Unbound, Personal Sanctuary, Worldslayer,
// Timely Reinforcements, Sutured Ghoul and Griffin Rider.

const (
	oldB452Veil       = "03f5946b-e000-4b39-a49f-1c31668b51bd"
	oldB452Monomania  = "29c6935e-ef64-4298-af74-9b18eed93136"
	oldB452Scramble   = "35fe3c61-8eaa-465d-ac34-63efda72fef2"
	oldB452Scepter    = "377d9245-9a70-4e35-9ed1-308565fc1acc"
	oldB452Colossus   = "b7f72e54-7e3a-4720-be69-fc8c5eb4c190"
	oldB452Bangchuck  = "f2561cc9-3840-4bd7-ad8b-a85d35e47f71"
	oldB452Unbound    = "0f12aec9-5898-4261-bebd-1b545e0aea68"
	oldB452Sanctuary  = "20963cad-1c13-44a1-901a-bfc9930089e6"
	oldB452Worldslay  = "31be715f-9ad3-435c-9e52-e108fe96b969"
	oldB452Timely     = "64f4f0ed-e924-459e-86a4-8202fae6ef3e"
	oldB452Ghoul      = "8b4715b0-63d9-4ff8-a667-21a762518687"
	oldB452Griffin    = "af8aa11b-34be-42a3-a8a2-3cad38f29f66"
	oldB452BoltOracle = "4457ed35-7c10-48c8-9776-456485fdf070"
)

func TestOldB452EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB452Veil: "Autumn's Veil", oldB452Monomania: "Monomania", oldB452Scramble: "Scrambleverse",
		oldB452Scepter: "Scepter of Empires", oldB452Colossus: "Crumbling Colossus", oldB452Bangchuck: "Goblin Bangchuckers",
		oldB452Unbound: "Mind Unbound", oldB452Sanctuary: "Personal Sanctuary", oldB452Worldslay: "Worldslayer",
		oldB452Timely: "Timely Reinforcements", oldB452Ghoul: "Sutured Ghoul", oldB452Griffin: "Griffin Rider",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v name=%q completeness=%s caveats=%v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB452AutumnsVeilShieldsYourSpellsAndBarsBlueAndBlackSpellsFromYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Autumn's Veil", "Instant", oldB452Veil, nil)
	effectstest.CgResolveTop(t, g)
	if got := effectstest.CgShields(g, me); len(got) != 1 || got[0].SourceName != "Autumn's Veil" {
		t.Errorf("the counter shield on the panel: %+v", got)
	}
	// An opponent responds to a spell of mine: a blue spell can't target my
	// creature, a red one can.
	effectstest.CsPushSpell(g, effectstest.CsSpell{Card: effectstest.CsSorcery})
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	cast := func(colors []string) error {
		id := uuid.New()
		opp.Hand.PushTop(game.Card{InstanceID: id, Name: "Bolt", TypeLine: "Instant", OracleID: oldB452BoltOracle,
			Colors: colors, Owner: opp.ID, Controller: opp.ID})
		return g.CastSpell(opp.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(bear)})
	}
	if err := cast([]string{"U"}); err == nil {
		t.Error("a blue spell targeted my creature")
	}
	if err := cast([]string{"B"}); err == nil {
		t.Error("a black spell targeted my creature")
	}
	if err := cast([]string{"R"}); err != nil {
		t.Errorf("a red spell could not target my creature: %v", err)
	}
}

func TestOldB452MonomaniaLeavesTheTargetPlayerOneCardOfTheirChoice(t *testing.T) {
	for _, hand := range []int{4, 1, 0} {
		g := effectstest.NewCatalogGame(t)
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		g.WithWriteLock(func() { opp.Hand.Cards = nil })
		var ids []uuid.UUID
		for i := 0; i < hand; i++ {
			ids = append(ids, effectstest.HandCardFull(opp, "Card", "Sorcery", "{1}", "", nil))
		}
		effectstest.CastCatalogSpell(t, g, "Monomania", "Sorcery", oldB452Monomania, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
		effectstest.PassPriorityUntilChoice(t, g)
		if hand > 1 {
			effectstest.AnswerDiscard(t, g, opp.ID, ids[1], ids[2], ids[3])
		} else if effectstest.DiscardChoiceFor(g, opp.ID) != nil {
			t.Fatalf("hand %d: asked to discard", hand)
		}
		effectstest.PassPriorityAroundTable(t, g)
		want := min(hand, 1)
		if opp.Hand.Size() != want {
			t.Errorf("hand %d: %d cards left, want %d", hand, opp.Hand.Size(), want)
		}
		if hand == 4 && !opp.Hand.Contains(ids[0]) {
			t.Errorf("the card they chose to keep was discarded")
		}
	}
}

func TestOldB452PersonalSanctuaryPreventsDamageToYouOnlyOnYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Personal Sanctuary", "Enchantment", oldB452Sanctuary, false)
	life := me.Life
	effectstest.Pr7Hit(t, g, uuid.New(), me.ID, 5)
	if me.Life != life {
		t.Errorf("life %d on my turn, want %d", me.Life, life)
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.Pr7Hit(t, g, uuid.New(), me.ID, 3)
	if me.Life != life-3 {
		t.Errorf("life %d on an opponent's turn, want %d", me.Life, life-3)
	}
}

func TestOldB452MindUnboundDrawsOneMoreEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.PushCatalogPermanent(g, me.ID, "Mind Unbound", "Enchantment", oldB452Unbound, false)
	for i := 0; i < 20; i++ {
		effectstest.PushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Sorcery"})
	}
	start := me.Library.Size()
	// Cumulative cards drawn when each of my upkeep triggers has resolved:
	// 1 at the first, then 1 more for my draw step and 2 at the second.
	for i, cumulative := range []int{1, 4} {
		want := i + 1
		seat := g.Turn.ActiveSeat
		effectstest.IwNextUpkeep(t, g, nil)
		for g.Turn.ActiveSeat != seat {
			effectstest.IwNextUpkeep(t, g, nil)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.CountersOn(g, id, game.CounterLore); got != want {
			t.Fatalf("%d lore counters, want %d", got, want)
		}
		if drew := start - me.Library.Size(); drew != cumulative {
			t.Errorf("upkeep %d: %d cards drawn so far, want %d", want, drew, cumulative)
		}
	}
}

func TestOldB452WorldslayerDestroysEverythingElseOnCombatDamageToAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	slayer := effectstest.SeedEquipment(g, me.ID, "Worldslayer", oldB452Worldslay)
	effectstest.EquipTo(t, g, me.ID, slayer, bear)
	other := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	land := effectstest.SeedPermanent(g, opp.ID, "Their Island", "Basic Land — Island")
	effectstest.B27Damage(g, bear, opp.ID, 2) // noncombat
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.BattlefieldHas(g, other) {
		t.Fatal("noncombat damage triggered it")
	}
	effectstest.DealCombatDamageToPlayer(g, bear, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"the equipped creature": bear, "their creature": other, "their land": land} {
		if effectstest.BattlefieldHas(g, id) {
			t.Errorf("%s survived", name)
		}
	}
	if !effectstest.BattlefieldHas(g, slayer) {
		t.Error("the Equipment destroyed itself")
	}
}

func TestOldB452CrumblingColossusIsSacrificedAtEndOfCombatAfterAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	colossus := effectstest.B12Push(g, me.ID, "Crumbling Colossus", "Artifact Creature — Golem", oldB452Colossus, 7, 4)
	effectstest.AdvanceToDeclareAttackersOf(t, g, g.Turn.ActiveSeat)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.DeclareAttack(t, g, opp.ID, colossus)
	effectstest.PassPriorityAroundTable(t, g)
	for i := 0; i < 12 && g.Turn.Step != game.StepEndCombat && effectstest.BattlefieldHas(g, colossus); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldHas(g, colossus) {
		t.Error("the Colossus is still on the battlefield after combat")
	}
}

func TestOldB452GoblinBangchuckersFlipsToHurtTheTargetOrItself(t *testing.T) {
	sawWin, sawLoss := false, false
	for i := 0; i < 64 && !(sawWin && sawLoss); i++ {
		g := effectstest.NewCatalogGame(t)
		me, opp := effectstest.SeatsForCombat(g)
		effectstest.AdvanceToMain(t, g)
		goblin := effectstest.B12Push(g, me.ID, "Goblin Bangchuckers", "Creature — Goblin Warrior", oldB452Bangchuck, 2, 2)
		life := opp.Life
		effectstest.B16Activate(t, g, me.ID, goblin, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}})
		effectstest.PassPriorityUntilChoice(t, g)
		flip := effectstest.LatestChoiceOfKind(g, game.PendingChoiceCoinCall)
		if flip == nil {
			t.Fatal("no coin call")
		}
		if err := g.ResolveCoinCall(flip.ID, me.ID, "heads"); err != nil {
			t.Fatalf("coin: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		flips := effectstest.RandomEvents(g, game.EventFlipCoin)
		won := len(flips) > 0 && flips[len(flips)-1].Won
		switch {
		case won:
			sawWin = true
			if opp.Life != life-2 || !effectstest.BattlefieldHas(g, goblin) {
				t.Errorf("won: opponent life %d (want %d), goblin alive %v", opp.Life, life-2, effectstest.BattlefieldHas(g, goblin))
			}
		default:
			sawLoss = true
			if opp.Life != life || effectstest.BattlefieldHas(g, goblin) {
				t.Errorf("lost: opponent life %d (want %d), goblin alive %v (2 damage on a 2/2)", opp.Life, life, effectstest.BattlefieldHas(g, goblin))
			}
		}
	}
	if !sawWin || !sawLoss {
		t.Fatalf("never saw both outcomes (win %v, loss %v)", sawWin, sawLoss)
	}
}

func TestOldB452ScepterOfEmpiresDealsThreeWithCrownAndThrone(t *testing.T) {
	for _, tc := range []struct {
		name          string
		crown, throne bool
		want          int
	}{{"alone", false, false, 1}, {"crown only", true, false, 1}, {"both", true, true, 3}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := effectstest.SeatsForCombat(g)
			effectstest.AdvanceToMain(t, g)
			scepter := effectstest.B12Push(g, me.ID, "Scepter of Empires", "Artifact", oldB452Scepter, 0, 0)
			if tc.crown {
				effectstest.SeedPermanent(g, me.ID, "Crown of Empires", "Artifact")
			}
			if tc.throne {
				effectstest.SeedPermanent(g, me.ID, "Throne of Empires", "Artifact")
			}
			life := opp.Life
			effectstest.B16Activate(t, g, me.ID, scepter, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}})
			effectstest.PassPriorityAroundTable(t, g)
			if opp.Life != life-tc.want {
				t.Errorf("life %d, want %d", opp.Life, life-tc.want)
			}
		})
	}
}

func TestOldB452TimelyReinforcementsChecksLifeAndCreaturesSeparately(t *testing.T) {
	for _, tc := range []struct {
		name            string
		myLife          int
		myBears, theirs int
		wantLife        int
		wantSoldiers    int
	}{
		{"behind in both", 10, 0, 2, 16, 3},
		{"behind in life only", 10, 2, 1, 16, 0},
		{"behind in creatures only", 40, 0, 1, 40, 3},
		{"ahead in both", 40, 2, 1, 40, 0},
		{"tied creatures and life", 20, 1, 1, 20, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			effectstest.AdvanceToMain(t, g)
			for _, s := range g.Seats {
				s.Life = 20
			}
			me.Life = tc.myLife
			for i := 0; i < tc.myBears; i++ {
				effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
			}
			for i := 0; i < tc.theirs; i++ {
				effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
			}
			effectstest.CastCatalogSpell(t, g, "Timely Reinforcements", "Sorcery", oldB452Timely, nil)
			effectstest.PassPriorityAroundTable(t, g)
			soldiers := 0
			for _, c := range g.BattlefieldCardsForEffect() {
				if c.Controller == me.ID && c.Name == "Soldier" && c.IsToken() {
					soldiers++
				}
			}
			if me.Life != tc.wantLife || soldiers != tc.wantSoldiers {
				t.Errorf("life %d (want %d), Soldiers %d (want %d)", me.Life, tc.wantLife, soldiers, tc.wantSoldiers)
			}
		})
	}
}

func TestOldB452GriffinRiderGetsBigAndFliesWhileYouControlAGriffin(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rider := effectstest.B12Push(g, me.ID, "Griffin Rider", "Creature — Human Knight", oldB452Griffin, 1, 1)
	effectstest.B12Creature(g, opp.ID, "Their Griffin", "Creature — Griffin", 2, 2)
	if p := effectstest.EffectivePower(t, g, rider); p != 1 {
		t.Errorf("power %d with only an opponent's Griffin, want 1", p)
	}
	effectstest.B12Creature(g, me.ID, "My Bird", "Creature — Bird", 1, 1)
	if p := effectstest.EffectivePower(t, g, rider); p != 1 {
		t.Errorf("power %d with a non-Griffin, want 1", p)
	}
	effectstest.B12Creature(g, me.ID, "My Griffin", "Creature — Griffin", 3, 3)
	if p, tough := effectstest.EffectivePower(t, g, rider), effectstest.EffectiveToughness(t, g, rider); p != 4 || tough != 4 {
		t.Errorf("%d/%d with a Griffin, want 4/4", p, tough)
	}
	effectstest.AssertKeywords(t, g, rider, "flying")
}

func TestOldB452SuturedGhoulIsAsBigAsTheCardsItExiled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	small := effectstest.PushGraveyardCardForTest(me, "Small")
	big := uuid.New()
	g.WithWriteLock(func() {
		for i := range me.Graveyard.Cards {
			if me.Graveyard.Cards[i].InstanceID == small {
				me.Graveyard.Cards[i].TypeLine, me.Graveyard.Cards[i].Power, me.Graveyard.Cards[i].Toughness = "Creature — Bear", 2, 2
			}
		}
		me.Graveyard.PushTop(game.Card{InstanceID: big, Name: "Big", TypeLine: "Creature — Ox", Power: 4, Toughness: 5,
			Owner: me.ID, Controller: me.ID})
	})
	untouched := effectstest.PushGraveyardCardTyped(me, "Spell", "Sorcery")
	effectstest.AdvanceToMain(t, g)
	ghoul := effectstest.CastFromHandForTest(t, g, me, "Sutured Ghoul", "Creature — Zombie", "{4}{B}{B}{B}", oldB452Ghoul, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt for the creature cards to exile")
	}
	if len(pick.ChooseCards) != 2 || effectstest.HasID(pick.ChooseCards, untouched) {
		t.Errorf("offered %d cards, want the two creature cards only", len(pick.ChooseCards))
	}
	effectstest.AnswerChooseCards(t, g, me.ID, small, big)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, ghoul); !ok {
		t.Fatal("the Ghoul died")
	}
	if p, tough := effectstest.PtOf(t, g, ghoul); p != 6 || tough != 7 {
		t.Errorf("Sutured Ghoul is %d/%d, want 6/7", p, tough)
	}
}

func TestOldB452SuturedGhoulExilingNothingDiesAsAZeroZero(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushGraveyardCardTyped(me, "Bear", "Creature — Bear")
	effectstest.AdvanceToMain(t, g)
	ghoul := effectstest.CastFromHandForTest(t, g, me, "Sutured Ghoul", "Creature — Zombie", "{4}{B}{B}{B}", oldB452Ghoul, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if effectstest.BattlefieldHas(g, ghoul) {
		t.Error("a 0/0 Sutured Ghoul stayed on the battlefield")
	}
}

func TestOldB452ScrambleverseRedistributesNonlandPermanentsAndUntapsThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	var perms []uuid.UUID
	for i := 0; i < 16; i++ {
		perms = append(perms, effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bear",
			TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Tapped: true, Owner: me.ID, Controller: me.ID}))
	}
	land := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Forest",
		TypeLine: "Basic Land — Forest", Tapped: true, Owner: me.ID, Controller: me.ID})
	effectstest.CastCatalogSpell(t, g, "Scrambleverse", "Sorcery", oldB452Scramble, nil)
	effectstest.PassPriorityAroundTable(t, g)
	moved := 0
	for _, id := range perms {
		c, ok := effectstest.BattlefieldCard(g, id)
		if !ok {
			t.Fatal("a permanent left the battlefield")
		}
		if c.Tapped {
			t.Error("a permanent stayed tapped")
		}
		if c.Controller != me.ID {
			moved++
		}
	}
	if moved == 0 {
		t.Error("16 permanents and not one changed hands")
	}
	if c, _ := effectstest.BattlefieldCard(g, land); c.Controller != me.ID || !c.Tapped {
		t.Error("the land was touched")
	}
}
