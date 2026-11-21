package testsi

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// karn_living_legacy_test.go — Karn, Living Legacy (#5487) and the
// activated ability on an emblem it needs (#5868, CR 114.4, CR 113.6).

const karnLivingLegacyOracle = "09ab1278-811f-4b97-bd34-7493d6e081e9"

// karnLLTable seeds Karn with `loyalty` under seat 0 in seat 0's main
// phase.
func karnLLTable(t *testing.T, loyalty int) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	karn := effectstest.PushCatalogPermanent(g, me.ID, "Karn, Living Legacy", "Legendary Planeswalker — Karn", karnLivingLegacyOracle, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(karn, game.CounterLoyalty, loyalty) })
	effectstest.AdvanceToMain(t, g)
	return g, me, karn
}

// karnLLEmblem plays the −7 and returns the emblem's instance ID.
func karnLLEmblem(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g, me, karn := karnLLTable(t, 7)
	effectstest.B16Activate(t, g, me.ID, karn, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Emblems.Size() != 1 {
		t.Fatalf("one emblem: %d", me.Emblems.Size())
	}
	return g, me, me.Emblems.Cards[0].InstanceID
}

// karnLLArtifact puts an untapped artifact under `owner`.
func karnLLArtifact(g *game.Game, owner uuid.UUID, typeLine string, sick bool) uuid.UUID {
	return effectstest.PushCatalogPermanent(g, owner, "Ornithopter", typeLine, "", sick)
}

func karnLLPing(g *game.Game, player, emblem, artifact uuid.UUID, target game.TargetRef) error {
	return g.ActivateCatalogAbility(player, emblem, 0, game.ActivateAbilityParams{
		TapIDs:  []uuid.UUID{artifact},
		Targets: []game.TargetRef{target},
		Strict:  true,
	})
}

func TestKarnLivingLegacyIsFull(t *testing.T) {
	spec, ok := effects.Lookup(karnLivingLegacyOracle)
	if !ok || spec.Name != "Karn, Living Legacy" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("registered=%v %+v", ok, spec.Completeness)
	}
}

func TestKarnLivingLegacyPlusOneMakesATappedPowerstone(t *testing.T) {
	g, me, karn := karnLLTable(t, 4)
	effectstest.B16Activate(t, g, me.ID, karn, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	var stones []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Powerstone" && c.Controller == me.ID {
			stones = append(stones, c)
		}
	}
	if len(stones) != 1 || !stones[0].Tapped || !stones[0].IsArtifact() {
		t.Fatalf("one tapped Powerstone artifact: %+v", stones)
	}
	if effectstest.LoyaltyOf(g, karn) != 5 {
		t.Errorf("+1 loyalty: %d", effectstest.LoyaltyOf(g, karn))
	}
}

// The −1 pays X as it resolves (CR 608.2d), looks at that many, puts one
// into the hand and the rest on the bottom.
func TestKarnLivingLegacyMinusOneLooksAtTheAmountPaid(t *testing.T) {
	g, me, karn := karnLLTable(t, 4)
	effectstest.StockLibraryFor(g, me, 10)
	lib := me.Library.Size()
	effectstest.FillPool(me, 3)
	effectstest.B16Activate(t, g, me.ID, karn, 1, game.ActivateAbilityParams{})
	pa := effectstest.AnswerNumber(t, g, me.ID, 3)
	if pa.Max != 3 {
		t.Errorf("X ceiling = %d, want 3 (the mana available)", pa.Max)
	}
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	if n := me.ManaPool.Total(); n != 0 {
		t.Errorf("the 3 mana was paid: %d left", n)
	}
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil || len(pick.ChooseCards) != 3 || pick.ChooseMin != 1 || pick.ChooseMax != 1 {
		t.Fatalf("one of three: %+v", pick)
	}
	taken := pick.ChooseCards[1]
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{taken}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HandHasID(me, taken) {
		t.Error("the chosen card goes to the hand")
	}
	if me.Library.Size() != lib-1 {
		t.Errorf("library %d → %d", lib, me.Library.Size())
	}
	bottom := map[uuid.UUID]bool{}
	for _, c := range me.Library.Cards[:2] {
		bottom[c.InstanceID] = true
	}
	for _, id := range pick.ChooseCards {
		if id != taken && !bottom[id] {
			t.Errorf("%s was not put on the bottom", id)
		}
	}
	if effectstest.LoyaltyOf(g, karn) != 3 {
		t.Errorf("−1 loyalty: %d", effectstest.LoyaltyOf(g, karn))
	}
}

// Paying nothing looks at nothing.
func TestKarnLivingLegacyMinusOneWithNoManaDoesNothing(t *testing.T) {
	g, me, karn := karnLLTable(t, 4)
	effectstest.StockLibraryFor(g, me, 10)
	hand, lib := me.Hand.Size(), me.Library.Size()
	effectstest.B16Activate(t, g, me.ID, karn, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChooseCardsFor(g, me.ID) != nil {
		t.Error("nothing was paid, so there is nothing to look at")
	}
	if me.Hand.Size() != hand || me.Library.Size() != lib {
		t.Errorf("hand %d → %d, library %d → %d", hand, me.Hand.Size(), lib, me.Library.Size())
	}
}

// The emblem's ability: tap an untapped artifact you control, 1 damage
// to a player, from the emblem (CR 114.4).
func TestKarnLivingLegacyEmblemPingsAPlayer(t *testing.T) {
	g, me, emblem := karnLLEmblem(t)
	opp := g.Seats[1]
	art := karnLLArtifact(g, me.ID, "Artifact", false)
	life := opp.Life
	if err := karnLLPing(g, me.ID, emblem, art, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("activate the emblem: %v", err)
	}
	if !effectstest.B16Tapped(t, g, art) {
		t.Error("the artifact paid the cost")
	}
	var fromEmblem bool
	g.ReadSnapshot(func() {
		for _, it := range g.StackMeta {
			fromEmblem = fromEmblem || (it != nil && it.Kind == game.StackItemActivated && it.SourceCardID == emblem)
		}
	})
	if !fromEmblem {
		t.Fatal("the emblem is the ability's source")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("1 damage: %d → %d", life, opp.Life)
	}
	if me.Emblems.Size() != 1 {
		t.Error("the emblem stays")
	}
}

// A creature and a planeswalker; a summoning-sick artifact creature
// pays (tapping another permanent is not {T}, CR 302.6).
func TestKarnLivingLegacyEmblemPingsACreatureAndAPlaneswalker(t *testing.T) {
	g, me, emblem := karnLLEmblem(t)
	opp := g.Seats[1]
	bear := effectstest.B16Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	walker := effectstest.PushCatalogPermanent(g, opp.ID, "Some Walker", "Legendary Planeswalker — Test", "", false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(walker, game.CounterLoyalty, 3) })
	sick := karnLLArtifact(g, me.ID, "Artifact Creature — Thopter", true)
	plain := karnLLArtifact(g, me.ID, "Artifact", false)

	if err := karnLLPing(g, me.ID, emblem, sick, game.TargetRef{Kind: game.TargetCard, ID: bear}); err != nil {
		t.Fatalf("a summoning-sick artifact creature pays: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, ok := effectstest.BattlefieldCard(g, bear); !ok || c.DamageMarked != 1 {
		t.Errorf("1 damage on the Bear: %+v", c)
	}
	if err := karnLLPing(g, me.ID, emblem, plain, game.TargetRef{Kind: game.TargetCard, ID: walker}); err != nil {
		t.Fatalf("activate at the planeswalker: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LoyaltyOf(g, walker) != 2 {
		t.Errorf("1 damage to a planeswalker removes a loyalty counter: %d", effectstest.LoyaltyOf(g, walker))
	}
}

// A tapped artifact, a non-artifact and another player's artifact can't
// pay; nothing is paid on a refusal.
func TestKarnLivingLegacyEmblemRefusesWhatCantPay(t *testing.T) {
	g, me, emblem := karnLLEmblem(t)
	opp := g.Seats[1]
	target := game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}
	tapped := karnLLArtifact(g, me.ID, "Artifact", false)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(tapped) })
	land := effectstest.PushLandFor(g, me.ID, "Mountain", "Basic Land — Mountain")
	theirs := karnLLArtifact(g, opp.ID, "Artifact", false)
	life := opp.Life
	for name, id := range map[string]uuid.UUID{"tapped artifact": tapped, "land": land, "their artifact": theirs} {
		if err := karnLLPing(g, me.ID, emblem, id, target); err == nil {
			t.Errorf("%s paid the cost", name)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, emblem, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{target}, Strict: true}); err == nil {
		t.Error("no artifact named, and it activated")
	}
	if effectstest.B16Tapped(t, g, land) || effectstest.B16Tapped(t, g, theirs) {
		t.Error("a refused activation tapped something")
	}
	if g.Stack.Size() != 0 || opp.Life != life {
		t.Error("a refused activation reached the stack")
	}
}

// Only the emblem's owner activates it (CR 114.2: its owner is its
// controller).
func TestKarnLivingLegacyEmblemIsItsOwnersOnly(t *testing.T) {
	g, me, emblem := karnLLEmblem(t)
	opp := g.Seats[1]
	theirs := karnLLArtifact(g, opp.ID, "Artifact", false)
	err := karnLLPing(g, opp.ID, emblem, theirs, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID})
	if !errors.Is(err, game.ErrCardCallerMismatch) {
		t.Fatalf("an opponent activated my emblem: %v", err)
	}
	if effectstest.B16Tapped(t, g, theirs) {
		t.Error("the refusal tapped their artifact")
	}
}

// The ability is identified by its catalog row (ADR 0041 P9): a restore
// with it on the stack resolves it, and the restored emblem activates.
func TestKarnLivingLegacyEmblemSurvivesARestore(t *testing.T) {
	g, me, emblem := karnLLEmblem(t)
	opp := g.Seats[1]
	art := karnLLArtifact(g, me.ID, "Artifact", false)
	art2 := karnLLArtifact(g, me.ID, "Artifact", false)
	life := opp.Life
	if err := karnLLPing(g, me.ID, emblem, art, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	restored := effectstest.RestoreRoundTrip(t, g, true)
	ropp := restored.PlayerByIDForEffect(opp.ID)
	effectstest.PassPriorityAroundTable(t, restored)
	if ropp.Life != life-1 {
		t.Errorf("the restored ability resolved: %d → %d", life, ropp.Life)
	}
	if err := karnLLPing(restored, me.ID, emblem, art2, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("the restored emblem activates: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, restored)
	if ropp.Life != life-2 {
		t.Errorf("second ping: %d → %d", life, ropp.Life)
	}
}

// The wire carries the emblem's row to its owner alone, addressed as an
// ordinary activate_ability on the emblem (#5868).
func TestKarnLivingLegacyEmblemRowIsOnTheOwnersFrame(t *testing.T) {
	g, me, emblem := karnLLEmblem(t)
	opp := g.Seats[1]
	karnLLArtifact(g, me.ID, "Artifact", false)
	rows := func(viewer uuid.UUID) []protocol.ActivatedAbilityView {
		v := protocol.FilterViewFor(protocol.ViewOfGame(g), viewer.String())
		for _, s := range v.Seats {
			if s.ID != me.ID.String() {
				continue
			}
			for _, e := range s.Emblems {
				if e.InstanceID == emblem.String() {
					return e.ZoneAbilities
				}
			}
		}
		t.Fatal("the emblem is not on the wire")
		return nil
	}
	mine := rows(me.ID)
	if len(mine) != 1 || mine[0].Index != 0 {
		t.Fatalf("the owner's frame offers the one row: %+v", mine)
	}
	if got := rows(opp.ID); len(got) != 0 {
		t.Errorf("an opponent's frame carries the owner's rows: %+v", got)
	}
}
