package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b463_test.go — OLD batch 463 (#5575): New Phyrexia cards Xenograft,
// Chancellor of the Spires, Bludgeon Brawl, Psychic Surgery, Geosurge and
// Corrupted Resolve.

const (
	oldB463Xenograft = "76642e7a-3c8f-45b8-9cd7-dac8086bea87"
	oldB463Spires    = "e766a7ec-2761-453b-bfba-956de7b5e125"
	oldB463Brawl     = "a9aa3bc7-db5a-4792-a53a-40acb6ca4b20"
	oldB463Surgery   = "e5128230-3535-4f77-8b97-20d0791e7b1d"
	oldB463Geosurge  = "6ec2794f-db1a-4448-a75e-ed7aff565378"
	oldB463Resolve   = "87837a94-05db-48fe-b810-26d8f92cfc58"
)

func TestOldB463EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB463Xenograft: "Xenograft", oldB463Spires: "Chancellor of the Spires", oldB463Brawl: "Bludgeon Brawl",
		oldB463Surgery: "Psychic Surgery", oldB463Geosurge: "Geosurge", oldB463Resolve: "Corrupted Resolve",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v name=%q completeness=%s caveats=%v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB463XenograftMakesYourCreaturesTheChosenTypeToo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastAndPause(t, g, "Xenograft", "Enchantment", oldB463Xenograft)
	effectstest.AnswerCreatureType(t, g, me.ID, "Goblin")
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.LayeredCard(t, g, mine).HasSubtype("Goblin") {
		t.Error("my creature is not a Goblin")
	}
	if !effectstest.LayeredCard(t, g, mine).HasSubtype("Bear") {
		t.Error("my creature lost its own type")
	}
	if effectstest.LayeredCard(t, g, theirs).HasSubtype("Goblin") {
		t.Error("an opponent's creature became a Goblin")
	}
}

func TestOldB463BludgeonBrawlMakesArtifactsEquipmentWithTheirManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Bludgeon Brawl", "Enchantment", oldB463Brawl, false)
	push := func(owner uuid.UUID, name, typeLine, cost string) uuid.UUID {
		return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
			ManaCost: cost, Power: 1, Toughness: 1, Owner: owner, Controller: owner})
	}
	stone := push(me.ID, "Mind Stone", "Artifact", "{2}")
	free := push(me.ID, "Darksteel Relic", "Artifact", "")
	golem := push(me.ID, "Golem", "Artifact Creature — Golem", "{4}")
	theirs := push(opp.ID, "Coalition Relic", "Artifact", "{3}")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	isEquip := func(id uuid.UUID) bool { return effectstest.LayeredCard(t, g, id).HasSubtype("Equipment") }
	if !isEquip(stone) || !isEquip(free) || !isEquip(theirs) {
		t.Fatal("a noncreature artifact is not an Equipment")
	}
	if isEquip(golem) {
		t.Error("an artifact creature became an Equipment")
	}
	if idx, _ := effectstest.BeGrantedRow(t, g, golem, "Equip {X}"); idx >= 0 {
		t.Error("an artifact creature has an equip row")
	}
	if idx, _ := effectstest.BeGrantedRow(t, g, stone, "Equip {X}"); idx < 0 {
		t.Fatal("Mind Stone has no equip row")
	}
	g.WithWriteLock(func() { effectstest.FillPool(me, 2) })
	effectstest.BeEquip(t, g, me.ID, stone, bear, "Equip {X}")
	if left := me.ManaPool.Total(); left != 0 {
		t.Errorf("equip on a 2-mana-value artifact left %d mana of 2, want it to cost {2}", left)
	}
	if p := effectstest.EffectivePower(t, g, bear); p != 4 {
		t.Errorf("power %d with a 2-mana-value Equipment on a 2/2, want 4", p)
	}
}

func TestOldB463GeosurgeAddsSevenRedThatOnlyCastsArtifactsAndCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastCatalogSpell(t, g, "Geosurge", "Sorcery", oldB463Geosurge, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.ManaPool.CountColor("R"); got != 7 {
		t.Fatalf("%d {R} in the pool, want 7", got)
	}
	for _, tok := range me.ManaPool.Units() {
		want := game.ManaRestrictAnyType("Artifact", "Creature")
		var has, cast bool
		for _, r := range tok.Restrictions {
			has = has || r == want
			cast = cast || r == game.ManaRestrictCast
		}
		if !has || !cast {
			t.Errorf("a red mana carries %v, want cast-only with %q", tok.Restrictions, want)
		}
	}
}

func TestOldB463CorruptedResolveCountersOnlyAPoisonedPlayersSpell(t *testing.T) {
	for _, poison := range []int{0, 1} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		g.WithWriteLock(func() { me.Poison = poison })
		effectstest.ToMainForCost(t, g)
		victim := effectstest.CastCatalogSpell(t, g, "My Bear", "Creature — Bear", "", nil)
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
		effectstest.CastFromSeat(t, g, opp, "Corrupted Resolve", "Instant", oldB463Resolve, []game.TargetRef{{Kind: game.TargetCard, ID: victim}})
		effectstest.PassPriorityAroundTable(t, g)
		_, resolved := effectstest.BattlefieldCard(g, victim)
		if resolved == (poison > 0) {
			t.Errorf("poison %d: the spell resolved=%v", poison, resolved)
		}
	}
}

func oldB463Pending(t *testing.T, g *game.Game, kind game.PendingChoiceKind, chooser uuid.UUID) *game.PendingChoice {
	t.Helper()
	c := effectstest.LatestChoiceOfKindFor(g, kind, chooser)
	if c == nil {
		t.Fatalf("no %s prompt for %v: %+v", kind, chooser, g.PendingChoices)
	}
	return c
}

func oldB463Top(p *game.Player, n int) []uuid.UUID {
	var out []uuid.UUID
	for i := len(p.Library.Cards) - 1; i >= 0 && len(out) < n; i-- {
		out = append(out, p.Library.Cards[i].InstanceID)
	}
	return out
}

func TestOldB463PsychicSurgeryLooksAtTheShuffledLibraryExilesOneAndOrdersTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Psychic Surgery", "Enchantment", oldB463Surgery, false)
	g.WithWriteLock(func() {
		if err := g.ShuffleLibraryForEffect(opp.ID); err != nil {
			t.Fatalf("shuffle: %v", err)
		}
	})
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	top := oldB463Top(opp, 2)
	size := opp.Library.Size()
	effectstest.PassPriorityAroundTable(t, g)
	pick := oldB463Pending(t, g, game.PendingChoiceChooseCards, me.ID)
	if len(pick.ChooseCards) != 2 {
		t.Fatalf("offered %d cards, want the top two", len(pick.ChooseCards))
	}
	for _, id := range top {
		c, _ := effectstest.CardInZone(opp.Library, id)
		if !c.IsKnownTo(me.ID) || c.IsKnownTo(opp.ID) {
			t.Errorf("only the looker should know the cards: %v", c.KnownBy)
		}
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{top[1]}); err != nil {
		t.Fatalf("exile one: %v", err)
	}
	if !effectstest.InExile(g, top[1]) || opp.Library.Contains(top[1]) {
		t.Error("the chosen card was not exiled from the library")
	}
	if opp.Library.Size() != size-1 || oldB463Top(opp, 1)[0] != top[0] {
		t.Errorf("library %d (want %d), the other card should be on top", opp.Library.Size(), size-1)
	}
}

func TestOldB463PsychicSurgeryExilingNothingOrdersBothBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Psychic Surgery", "Enchantment", oldB463Surgery, false)
	g.WithWriteLock(func() { _ = g.ShuffleLibraryForEffect(opp.ID) })
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	top := oldB463Top(opp, 2)
	size := opp.Library.Size()
	effectstest.PassPriorityAroundTable(t, g)
	pick := oldB463Pending(t, g, game.PendingChoiceChooseCards, me.ID)
	if err := g.ResolveChooseCards(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("exile none: %v", err)
	}
	order := oldB463Pending(t, g, game.PendingChoicePutInLibrary, me.ID)
	if err := g.ResolvePutInLibrary(order.ID, me.ID, nil, []uuid.UUID{top[1], top[0]}); err != nil {
		t.Fatalf("order: %v", err)
	}
	if got := oldB463Top(opp, 2); got[0] != top[1] || got[1] != top[0] || opp.Library.Size() != size {
		t.Errorf("top two %v, want the order I chose %v with the library still %d cards", got, []uuid.UUID{top[1], top[0]}, size)
	}
}

func TestOldB463PsychicSurgeryIgnoresMyOwnShuffle(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Psychic Surgery", "Enchantment", oldB463Surgery, false)
	g.WithWriteLock(func() { _ = g.ShuffleLibraryForEffect(me.ID) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID) != nil || effectstest.LatestConfirmFor(g, me.ID) != nil {
		t.Error("Psychic Surgery reacted to my own shuffle")
	}
}

func oldB463SpiresCard() game.Card {
	return game.Card{Name: "Chancellor of the Spires", TypeLine: "Creature — Phyrexian Sphinx", ManaCost: "{4}{U}{U}{U}",
		OracleID: oldB463Spires, Power: 5, Toughness: 7}
}

func TestOldB463ChancellorRevealedMillsEachOpponentSevenAtTheFirstUpkeep(t *testing.T) {
	spec, _ := effects.Lookup(oldB463Spires)
	if spec.OpeningHand == nil || spec.OpeningHand.Reveal == nil || spec.OpeningHand.Reveal.At != game.RevealAtTheFirstUpkeep {
		t.Fatalf("opening-hand reveal at the first upkeep missing: %+v", spec.OpeningHand)
	}
	for _, reveal := range []bool{true, false} {
		g := effectstest.OpeningTable(t)
		starter := g.Seats[g.StartingSeat]
		ids := effectstest.HandWith(g, g.StartingSeat, oldB463SpiresCard(), effectstest.FillerCard("Spare"))
		effectstest.KeepAll(t, g)
		offer := oldB463Pending(t, g, game.PendingChoiceConfirm, starter.ID)
		if err := g.ResolveConfirm(offer.ID, starter.ID, reveal); err != nil {
			t.Fatalf("reveal: %v", err)
		}
		if !starter.Hand.Contains(ids[0]) {
			t.Fatal("revealing the Chancellor moved it")
		}
		before := map[uuid.UUID]int{}
		for _, p := range g.Seats {
			before[p.ID] = p.Library.Size()
		}
		effectstest.PassPriorityAroundTable(t, g)
		for _, p := range g.Seats {
			milled := before[p.ID] - p.Library.Size()
			want := 0
			if reveal && p.ID != starter.ID {
				want = 7
			}
			if milled != want {
				t.Errorf("reveal=%v: %s milled %d, want %d", reveal, p.Name, milled, want)
			}
		}
	}
}

func TestOldB463ChancellorEntersAndCastsAnOpponentsInstantForFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bolt := effectstest.GyCard(opp, "Lightning Bolt", "Instant", "{R}", effectstest.LightningBoltOracle)
	own := effectstest.GyCard(me, "Lightning Bolt", "Instant", "{R}", effectstest.LightningBoltOracle)
	_ = own
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	chancellor := effectstest.CastFromHandForTest(t, g, me, "Chancellor of the Spires", "Creature — Phyrexian Sphinx", "{4}{U}{U}{U}", oldB463Spires, game.CastSpellParams{})
	effectstest.PassPriorityUntilChoice(t, g)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePickTarget, me.ID) != nil {
		effectstest.AnswerFirstPickTarget(t, g)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.BattlefieldHas(g, chancellor) {
		t.Fatal("the Chancellor did not enter")
	}
	third := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	life := third.Life
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: game.AltCostKeyGrantedCost, Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: third.ID}},
	}); err != nil {
		t.Fatalf("casting the opponent's instant for free: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if third.Life != life-3 {
		t.Errorf("life %d → %d: the Bolt deals 3", life, third.Life)
	}
	if !opp.Graveyard.Contains(bolt) {
		t.Error("the spell goes back to its owner's graveyard")
	}
}
