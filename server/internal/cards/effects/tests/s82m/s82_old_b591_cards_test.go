package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b591_cards_test.go — Saviors of Kamigawa, batch 591 (#5601). Each
// test plays the real catalog entry and asserts what its text says.

const (
	b591ElderPineOracle   = "1658528b-df40-44fd-b3f6-eb042706c522"
	b591DeathknellOracle  = "4d324f8c-2139-4910-b3c3-731a2e820d6b"
	b591ToriiOracle       = "8b52c2ed-c1a1-4abc-a33b-223c3d0e229d"
	b591TendedOracle      = "e2770ebf-24e3-4787-956b-ca01aaa50a6f"
	b591MeishinOracle     = "3b8eea0b-6f8c-449c-9ab8-a22a035465d6"
	b591ReverenceOracle   = "60963292-2dee-4b1c-9c1c-c35cf8ea3496"
	b591CanopyOracle      = "d182ab7d-77bd-49d1-9001-9b4016350181"
	b591ErayoOracle       = "d83bb32d-f409-421a-9d93-bd7ea3240b47"
	b591SinkOracle        = "b1d43244-38b3-498d-8522-9b4011998a66"
	b591ChargeOracle      = "cd76e1a4-d6ff-419a-a881-1b77eb6ae581"
	b591BarrelOracle      = "d9ecb346-7745-4106-a3e1-c467a3b91ec7"
	b591PlowOracle        = "e04f151a-66bb-4bc3-a625-91cd23cb1b18"
	b591EssenceOracleSufx = "#1"
)

func TestOldB591EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		b591ElderPineOracle:                     "Elder Pine of Jukai",
		b591DeathknellOracle:                    "Deathknell Kami",
		b591ToriiOracle:                         "Torii Watchward",
		b591TendedOracle:                        "Kami of the Tended Garden",
		b591MeishinOracle:                       "Meishin, the Mind Cage",
		b591ReverenceOracle:                     "Reverence",
		b591CanopyOracle:                        "Dense Canopy",
		b591ErayoOracle:                         "Erayo, Soratami Ascendant",
		b591ErayoOracle + b591EssenceOracleSufx: "Erayo's Essence",
		b591SinkOracle:                          "Sink into Takenuma",
		b591ChargeOracle:                        "Charge Across the Araba",
		b591BarrelOracle:                        "Barrel Down Sokenzan",
		b591PlowOracle:                          "Plow Through Reito",
	}
	for oracle, name := range want {
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

func b591Seats(g *game.Game) (me, opp *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

// b591Library replaces a player's library with the cards, top first
// (PushLibraryCardForTest pushes to the bottom, so the first ends on top).
func b591Library(p *game.Player, cards ...game.Card) []uuid.UUID {
	p.Library.Cards = nil
	ids := make([]uuid.UUID, len(cards))
	for i := range cards {
		cards[i].Owner, cards[i].Controller = p.ID, p.ID
		ids[i] = effectstest.PushLibraryCardForTest(p, cards[i])
	}
	return ids
}

func b591Hand(p *game.Player, names ...string) {
	p.Hand.Cards = nil
	effectstest.FillHand(p, names...)
}

func b591Land(g *game.Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, Owner: owner, Controller: owner,
	})
}

func b591Cast(t *testing.T, g *game.Game, name, typeLine, oracle string, targets []game.TargetRef) uuid.UUID {
	t.Helper()
	return effectstest.CastCatalogSpell(t, g, name, typeLine, oracle, targets)
}

// b591AnswerSweep answers the open "return any number" pick with the first n
// candidates and passes priority until the stack settles.
func b591AnswerSweep(t *testing.T, g *game.Game, chooser uuid.UUID, n int) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceChooseCards && c.Chooser == chooser {
			if n > len(c.ChooseCards) {
				t.Fatalf("asked for %d of %d candidates", n, len(c.ChooseCards))
			}
			if err := g.ResolveChooseCards(c.ID, chooser, c.ChooseCards[:n]); err != nil {
				t.Fatalf("answer the sweep pick: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			return
		}
	}
	t.Fatal("no sweep pick is open")
}

func b591SweepOpen(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceChooseCards && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

func TestOldB591ElderPineTakesTheLandsAndBottomsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := b591Seats(g)
	effectstest.B12Push(g, me.ID, "Elder Pine of Jukai", "Creature — Spirit", b591ElderPineOracle, 2, 1)
	land := func(n string) game.Card { return game.Card{Name: n, TypeLine: "Basic Land — Forest"} }
	spell := func(n string) game.Card { return game.Card{Name: n, TypeLine: "Sorcery"} }
	ids := b591Library(me, land("L1"), spell("S1"), land("L2"), spell("Below"))
	// A Bear spell does nothing.
	b591Cast(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(ids[0]) {
		t.Fatal("a Bear spell triggered Elder Pine")
	}
	b591Cast(t, g, "A Spirit", "Creature — Spirit", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if ask := effectstest.PutInLibraryChoiceFor(g, me.ID); ask != nil {
		if err := g.ResolvePutInLibrary(ask.ID, me.ID, nil, ask.ScryCards); err != nil {
			t.Fatalf("order the bottom: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if !me.Hand.Contains(ids[0]) || !me.Hand.Contains(ids[2]) {
		t.Error("the land cards were not put into your hand")
	}
	if me.Hand.Contains(ids[1]) || me.Hand.Contains(ids[3]) {
		t.Error("a nonland card was put into your hand")
	}
	// The unrevealed fourth card is now on top; the revealed nonland card is on the bottom.
	if me.Library.Size() != 2 {
		t.Fatalf("the library holds %d cards, want 2", me.Library.Size())
	}
	if top := effectstest.LibraryTopNames(me, 2); top[0] != "Below" || top[1] != "S1" {
		t.Errorf("library %v, want the unrevealed card on top and the revealed nonland card below it", top)
	}
}

func TestOldB591SoulshiftersReturnASmallerSpirit(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		n                      int
	}{
		{"Torii Watchward", "Creature — Spirit", b591ToriiOracle, 4},
		{"Elder Pine of Jukai", "Creature — Spirit", b591ElderPineOracle, 2},
		{"Deathknell Kami", "Creature — Spirit", b591DeathknellOracle, 1},
		{"Kami of the Tended Garden", "Creature — Spirit", b591TendedOracle, 3},
	} {
		g := effectstest.NewCatalogGame(t)
		me, _ := b591Seats(g)
		kami := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
		fits, tooBig, bear := uuid.New(), uuid.New(), uuid.New()
		me.Graveyard.PushTop(game.Card{InstanceID: fits, Name: "Fits", TypeLine: "Creature — Spirit",
			ManaCost: effectstest.ChkB1GenericCost(tc.n), Owner: me.ID, Controller: me.ID})
		me.Graveyard.PushTop(game.Card{InstanceID: tooBig, Name: "Too Big", TypeLine: "Creature — Spirit",
			ManaCost: effectstest.ChkB1GenericCost(tc.n + 1), Owner: me.ID, Controller: me.ID})
		me.Graveyard.PushTop(game.Card{InstanceID: bear, Name: "Bear Card", TypeLine: "Creature — Bear",
			ManaCost: "{1}", Owner: me.ID, Controller: me.ID})
		effectstest.KillCreature(t, g, me.ID, kami)
		p := effectstest.PassUntilPickTarget(t, g, me.ID)
		if p == nil {
			t.Fatalf("%s: no soulshift target prompt", tc.name)
		}
		if !effectstest.HasID(p.PickTargetCards, fits) || effectstest.HasID(p.PickTargetCards, tooBig) || effectstest.HasID(p.PickTargetCards, bear) {
			t.Errorf("%s: soulshift offers %v, want the mana value %d Spirit", tc.name, p.PickTargetCards, tc.n)
		}
		effectstest.PickCard(t, g, me.ID, fits)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
		effectstest.PassPriorityAroundTable(t, g)
		if !me.Hand.Contains(fits) {
			t.Errorf("%s: soulshift did not return the Spirit card", tc.name)
		}
	}
}

func TestOldB591ToriiWatchwardHasVigilance(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.B12Push(g, g.Seats[0].ID, "Torii Watchward", "Creature — Spirit", b591ToriiOracle, 3, 3)
	if !effectstest.HasKeywordOnBattlefield(t, g, id, "vigilance") {
		t.Error("Torii Watchward lacks vigilance")
	}
}

func TestOldB591DeathknellKamiPumpsAndSacrificesItselfAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := b591Seats(g)
	kami := effectstest.PushCatalogPermanent(g, me.ID, "Deathknell Kami", "Creature — Spirit", b591DeathknellOracle, false)
	effectstest.AdvanceToMain(t, g)
	// The test permanent starts 1/1; the point is the +1/+1 and the sacrifice.
	before, _ := effectstest.EffectivePTOf(t, g, kami)
	effectstest.C2Pay(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, kami, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if after, _ := effectstest.EffectivePTOf(t, g, kami); after != before+1 {
		t.Errorf("power %d → %d, want +1", before, after)
	}
	if !g.Battlefield.Contains(kami) {
		t.Fatal("sacrificed at once")
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(kami) || !me.Graveyard.Contains(kami) {
		t.Error("Deathknell Kami was not sacrificed at the beginning of the end step")
	}
}

func TestOldB591KamiOfTheTendedGardenPaysGOrIsSacrificed(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, _ := b591Seats(g)
		seat := g.Turn.ActiveSeat
		kami := effectstest.PushCatalogPermanent(g, me.ID, "Kami of the Tended Garden", "Creature — Spirit", b591TendedOracle, false)
		effectstest.AdvanceToUpkeepOfSeat(t, g, seat)
		if pay {
			effectstest.C2Pay(t, g, me, "{G}")
		}
		for i := 0; i < 16 && !effectstest.HasPayUnlessFor(g, me.ID); i++ {
			if err := g.PassPriority(); err != nil {
				t.Fatalf("pass: %v", err)
			}
		}
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(kami) != pay {
			t.Errorf("pay=%v: the Kami is on the battlefield: %v", pay, g.Battlefield.Contains(kami))
		}
	}
}

func TestOldB591MeishinShrinksEveryCreatureByYourHandSize(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := b591Seats(g)
	effectstest.AdvanceToMain(t, g)
	b591Hand(me, "A", "B", "C")
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 5, 5)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Ogre", 6, 4)
	effectstest.PushCatalogPermanent(g, me.ID, "Meishin, the Mind Cage", "Legendary Enchantment", b591MeishinOracle, false)
	for _, tc := range []struct {
		id   uuid.UUID
		want int
	}{{mine, 2}, {theirs, 3}} {
		if p, tough := effectstest.EffectivePTOf(t, g, tc.id); p != tc.want || tough < 4 {
			t.Errorf("a creature has %d/%d with three cards in hand, want power %d", p, tough, tc.want)
		}
	}
	// It follows the hand live: a draw and nothing else.
	if err := g.DrawCard(me.ID); err != nil {
		t.Fatalf("draw: %v", err)
	}
	if p, _ := effectstest.EffectivePTOf(t, g, mine); p != 1 {
		t.Errorf("power %d after a draw (four cards in hand), want 1", p)
	}
	// An opponent's hand does not count.
	opp.Hand.Cards = nil
	effectstest.FillHand(opp, "X", "Y", "Z", "W", "V")
	if p, _ := effectstest.EffectivePTOf(t, g, mine); p != 1 {
		t.Errorf("power %d after the opponent's hand grew, want 1", p)
	}
}

func TestOldB591ReverenceKeepsSmallCreaturesFromAttackingYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, third := g.Seats[0], g.Seats[2]
	opp := g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Reverence", "Enchantment", b591ReverenceOracle, false)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	effectstest.AdvanceToDeclareAttackersOf(t, g, 1)
	if err := g.Clone().DeclareAttacker(bear, me.ID); err == nil {
		t.Error("a power 2 creature attacked the Reverence player")
	}
	if err := g.Clone().DeclareAttacker(ogre, me.ID); err != nil {
		t.Errorf("a power 3 creature could not attack the Reverence player: %v", err)
	}
	if err := g.Clone().DeclareAttacker(bear, third.ID); err != nil {
		t.Errorf("a power 2 creature could not attack another player: %v", err)
	}
}

func TestOldB591DenseCanopyLetsFliersBlockOnlyFliers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Dense Canopy", "Enchantment", b591CanopyOracle, 0, 0)
	flier := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird",
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID, Keywords: []string{"flying"},
	})
	ground := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	attackerFlier := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Hawk", TypeLine: "Creature — Bird",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID, Keywords: []string{"flying"},
	})
	attackerGround := effectstest.B12Creature(g, me.ID, "Ox", "Creature — Ox", 2, 2)
	refused := func(attacker, blocker uuid.UUID) bool {
		var r game.BlockRefusal
		g.WithWriteLock(func() {
			g.RecomputeLayersIfStaleLocked()
			var a, b *game.Card
			for i := range g.Battlefield.Cards {
				switch g.Battlefield.Cards[i].InstanceID {
				case attacker:
					a = &g.Battlefield.Cards[i]
				case blocker:
					b = &g.Battlefield.Cards[i]
				}
			}
			r = g.BlockPairRefusalLocked(a, b)
		})
		return r.Reason != ""
	}
	if !refused(attackerGround, flier) {
		t.Error("a creature with flying blocked a creature without flying")
	}
	if refused(attackerFlier, flier) {
		t.Error("a creature with flying could not block a flier")
	}
	if refused(attackerGround, ground) {
		t.Error("a ground creature could not block a ground creature")
	}
	// It grants no reach: a ground creature still can't block a flier.
	if !refused(attackerFlier, ground) {
		t.Error("a ground creature without reach blocked a flier")
	}
}

func b591ErayoRow() cards.Card {
	return effectstest.FlipRow(b591ErayoOracle, "{1}{U}", []string{"U"}, []string{"flying"},
		cards.CardFace{Name: "Erayo, Soratami Ascendant", TypeLine: "Legendary Creature — Moonfolk Monk",
			OracleText: "Flying\nWhenever the fourth spell of a turn is cast, flip Erayo.", Power: "1", Toughness: "1"},
		cards.CardFace{Name: "Erayo's Essence", TypeLine: "Legendary Enchantment",
			OracleText: "Whenever an opponent casts their first spell each turn, counter that spell."})
}

func TestOldB591ErayoFlipsOnTheFourthSpellAndCountersAnOpponentsFirst(t *testing.T) {
	g, me, opp := effectstest.FlipMain(t)
	erayo := effectstest.FlipEntered(t, g, b591ErayoRow(), me)
	effectstest.WantUnflipped(t, g, erayo, "Erayo, Soratami Ascendant")
	for i := 1; i <= 3; i++ {
		b591Cast(t, g, "Spell", "Sorcery", "", nil)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.WantUnflipped(t, g, erayo, "Erayo, Soratami Ascendant")
	}
	b591Cast(t, g, "Fourth", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.WantFlipped(t, g, erayo, "Erayo's Essence", []string{"U"}, 2)
	// A fifth spell does nothing more.
	b591Cast(t, g, "Fifth", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.WantFlipped(t, g, erayo, "Erayo's Essence", []string{"U"}, 2)
	// An opponent's first spell of a turn is countered, their second is not.
	effectstest.AangAdvanceToMain(t, g, 1)
	first := effectstest.CastCatalogSpell(t, g, "Opp First", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(first) || g.Battlefield.Contains(first) {
		t.Error("the opponent's first spell was not countered")
	}
	second := effectstest.CastCatalogSpell(t, g, "Opp Second", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(second) {
		t.Error("the opponent's second spell was countered")
	}
}

func TestOldB591SinkIntoTakenumaReturnsSwampsAndMakesThemDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := b591Seats(g)
	swamps := []uuid.UUID{
		b591Land(g, me.ID, "Swamp", "Basic Land — Swamp"),
		b591Land(g, me.ID, "Swamp", "Basic Land — Swamp"),
		b591Land(g, me.ID, "Swamp", "Basic Land — Swamp"),
	}
	island := b591Land(g, me.ID, "Island", "Basic Land — Island")
	b591Hand(opp, "A", "B", "C")
	b591Cast(t, g, "Sink into Takenuma", "Sorcery — Arcane", b591SinkOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	pick := b591SweepOpen(g, me.ID)
	if pick == nil {
		t.Fatal("no sweep pick")
	}
	for _, id := range pick.ChooseCards {
		if id == island {
			t.Fatal("the Island is offered")
		}
	}
	b591AnswerSweep(t, g, me.ID, 2)
	returned := 0
	for _, id := range swamps {
		if me.Hand.Contains(id) {
			returned++
		}
	}
	if returned != 2 {
		t.Fatalf("%d Swamps returned, want 2", returned)
	}
	// The target discards two cards of their own choice.
	d := effectstest.DiscardChoiceFor(g, opp.ID)
	if d == nil {
		t.Fatal("the target was not asked to discard")
	}
	if d.ChooseMin != 2 {
		t.Errorf("the target owes %d discards, want 2", d.ChooseMin)
	}
}

func TestOldB591SinkIntoTakenumaWithNoSwampsDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := b591Seats(g)
	b591Land(g, me.ID, "Island", "Basic Land — Island")
	b591Hand(opp, "A", "B")
	b591Cast(t, g, "Sink into Takenuma", "Sorcery — Arcane", b591SinkOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if b591SweepOpen(g, me.ID) != nil || opp.Hand.Size() != 2 {
		t.Error("a spell with no Swamp to return asked something or made a discard")
	}
}

func TestOldB591ChargeAcrossTheArabaPumpsByPlainsReturned(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := b591Seats(g)
	plains := []uuid.UUID{
		b591Land(g, me.ID, "Plains", "Basic Land — Plains"),
		b591Land(g, me.ID, "Plains", "Basic Land — Plains"),
		b591Land(g, me.ID, "Plains", "Basic Land — Plains"),
	}
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	b591Cast(t, g, "Charge Across the Araba", "Instant — Arcane", b591ChargeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	b591AnswerSweep(t, g, me.ID, 3)
	for _, id := range plains {
		if !me.Hand.Contains(id) {
			t.Error("a Plains was not returned")
		}
	}
	if p, tough := effectstest.EffectivePTOf(t, g, mine); p != 5 || tough != 5 {
		t.Errorf("your creature is %d/%d, want 5/5", p, tough)
	}
	if p, _ := effectstest.EffectivePTOf(t, g, theirs); p != 2 {
		t.Errorf("an opponent's creature has power %d, want 2", p)
	}
	effectstest.DcCleanup(g)
	if p, _ := effectstest.EffectivePTOf(t, g, mine); p != 2 {
		t.Errorf("the pump outlived the turn (%d)", p)
	}
}

func TestOldB591BarrelDownSokenzanDealsTwiceTheMountainsReturned(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := b591Seats(g)
	for i := 0; i < 2; i++ {
		b591Land(g, me.ID, "Mountain", "Basic Land — Mountain")
	}
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Ogre", 5, 5)
	b591Cast(t, g, "Barrel Down Sokenzan", "Instant — Arcane", b591BarrelOracle, []game.TargetRef{{Kind: game.TargetCard, ID: big}})
	effectstest.PassPriorityAroundTable(t, g)
	b591AnswerSweep(t, g, me.ID, 2)
	if got := effectstest.DamageMarkedOn(g, big); got != 4 {
		t.Errorf("the target has %d damage, want 4", got)
	}
}

func TestOldB591BarrelDownSokenzanMayReturnNone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := b591Seats(g)
	m := b591Land(g, me.ID, "Mountain", "Basic Land — Mountain")
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Ogre", 5, 5)
	b591Cast(t, g, "Barrel Down Sokenzan", "Instant — Arcane", b591BarrelOracle, []game.TargetRef{{Kind: game.TargetCard, ID: big}})
	effectstest.PassPriorityAroundTable(t, g)
	b591AnswerSweep(t, g, me.ID, 0)
	if !g.Battlefield.Contains(m) || effectstest.DamageMarkedOn(g, big) != 0 {
		t.Error("returning none moved a Mountain or dealt damage")
	}
}

func TestOldB591PlowThroughReitoPumpsTheTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := b591Seats(g)
	for i := 0; i < 2; i++ {
		b591Land(g, me.ID, "Plains", "Basic Land — Plains")
	}
	target := effectstest.B12Creature(g, me.ID, "Target", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	b591Cast(t, g, "Plow Through Reito", "Instant — Arcane", b591PlowOracle, []game.TargetRef{{Kind: game.TargetCard, ID: target}})
	effectstest.PassPriorityAroundTable(t, g)
	b591AnswerSweep(t, g, me.ID, 2)
	if p, tough := effectstest.EffectivePTOf(t, g, target); p != 4 || tough != 4 {
		t.Errorf("the target is %d/%d, want 4/4", p, tough)
	}
	if p, _ := effectstest.EffectivePTOf(t, g, other); p != 2 {
		t.Errorf("another creature has power %d, want 2", p)
	}
}
