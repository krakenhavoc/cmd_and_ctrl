package testsex

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fra_prepare_b_test.go — Reality Fracture slice fra-prepare-b (#2795):
// the preparation cards that mill, threshold, rummage, wheel, or burn.

const (
	fraVoidExtrapolator   = "517f84b5-cf99-410d-96ea-910740231dac"
	fraTheorixMetamage    = "92a0aa44-0ef7-4e6f-8288-7c33067c5fd1"
	fraParadoxShaper      = "511951ed-fbff-4e44-9429-27f237496672"
	fraStingerquill       = "e8f755f7-ec93-4d2e-bffc-c719984fa13d"
	fraWhiplashWordsmith  = "e1ffb884-a89e-4e2a-9beb-8bd5838b6a94"
	fraPompousBattlemage  = "33e3793c-098f-4897-88f3-9f9ce9081e0f"
	fraVariableChaser     = "67c603cc-ad66-4a1c-8386-5901c9c01bfb"
	fraVigorbloomVanguard = "bd416426-d037-45e3-9647-816bf982cbea"
)

// fraPrepB builds a preparation card in the shape deck import gives.
func fraPrepB(owner uuid.UUID, oracle string, keywords []string, name, typeLine string, power, toughness int, spell, spellType string) game.Card {
	return effectstest.PreparationCard(owner, oracle, keywords,
		game.Face{Name: name, TypeLine: typeLine, ManaCost: "{1}", Power: power, Toughness: toughness},
		game.Face{Name: spell, TypeLine: spellType, ManaCost: "{1}"})
}

func fraBFillGraveyard(g *game.Game, p *game.Player, n int) {
	for i := 0; i < n; i++ {
		id := effectstest.PushGraveyardCardForTest(p, "Filler")
		g.WithWriteLock(func() {
			g.EmitEvent(game.Event{Kind: game.EventZoneMove, CardID: id, OldZone: game.ZoneHand, NewZone: game.ZoneGraveyard})
		})
	}
}

func TestVoidExtrapolatorThresholdAndOmitVariables(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	c := fraPrepB(me.ID, fraVoidExtrapolator, nil, "Void Extrapolator", "Creature — Aetherborn Warlock", 2, 2, "Omit Variables", "Sorcery")
	effectstest.CastFromHandAndResolve(t, g, me, c)
	if !g.IsPreparedForEffect(c.InstanceID) {
		t.Fatal("did not enter prepared")
	}
	me.Graveyard.Cards = nil
	fraBFillGraveyard(g, me, 6)
	if got := effectstest.EffectivePower(t, g, c.InstanceID); got != 2 {
		t.Errorf("power with 6 in graveyard = %d, want 2", got)
	}
	fraBFillGraveyard(g, me, 1)
	if p, tough := effectstest.EffectivePower(t, g, c.InstanceID), effectstest.EffectiveToughness(t, g, c.InstanceID); p != 3 || tough != 3 {
		t.Errorf("threshold P/T = %d/%d, want 3/3", p, tough)
	}
	lib, gy := me.Library.Size(), me.Graveyard.Size()
	effectstest.FraBCastCopy(t, g, me, "Omit Variables")
	if got := me.Library.Size(); got != lib-3 {
		t.Errorf("library %d, want %d after milling three", got, lib-3)
	}
	if got := me.Graveyard.Size(); got != gy+3 {
		t.Errorf("graveyard %d, want %d", got, gy+3)
	}
	if g.IsPreparedForEffect(c.InstanceID) {
		t.Error("still prepared after casting the copy")
	}
}

func TestTheorixMetamageThresholdGivesPowerAndFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	c := fraPrepB(me.ID, fraTheorixMetamage, nil, "Theorix Metamage", "Creature — Shade Wizard", 2, 3, "Omit Variables", "Sorcery")
	effectstest.PushBattlefieldCardWithTimestamp(g, c)
	me.Graveyard.Cards = nil
	fraBFillGraveyard(g, me, 6)
	if got := effectstest.EffectivePower(t, g, c.InstanceID); got != 2 {
		t.Errorf("power with 6 = %d, want 2", got)
	}
	if effectstest.EffectiveAbilitiesContain(t, g, c.InstanceID, "flying") {
		t.Error("flying below threshold")
	}
	fraBFillGraveyard(g, me, 1)
	if got := effectstest.EffectivePower(t, g, c.InstanceID); got != 3 {
		t.Errorf("power with 7 = %d, want 3", got)
	}
	if got := effectstest.EffectiveToughness(t, g, c.InstanceID); got != 3 {
		t.Errorf("toughness with 7 = %d, want 3 (+1/+0 only)", got)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, c.InstanceID, "flying") {
		t.Error("no flying at threshold")
	}
}

func TestParadoxShaperPreparesAtUpkeepAndTucksAGraveyardCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	c := fraPrepB(me.ID, fraParadoxShaper, nil, "Paradox Shaper", "Creature — Octopus Wizard", 1, 3, "Omit Variables", "Sorcery")
	effectstest.CastFromHandAndResolve(t, g, me, c)
	if g.IsPreparedForEffect(c.InstanceID) {
		t.Fatal("Paradox Shaper entered prepared; only its upkeep trigger prepares it")
	}

	// {2}: put target card from your graveyard on the bottom.
	victim := effectstest.PushGraveyardCardForTest(me, "Tucked")
	effectstest.FloatMana(t, g, me, "{C}{C}")
	lib := me.Library.Size()
	if err := g.ActivateCatalogAbility(me.ID, c.InstanceID, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(victim)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Library.Size(); got != lib+1 {
		t.Fatalf("library %d, want %d", got, lib+1)
	}
	if last := me.Library.Cards[0]; last.InstanceID != victim {
		t.Errorf("bottom card is %s, want the tucked card", last.Name)
	}

	// A card in an opponent's graveyard is not a legal target.
	theirs := effectstest.PushGraveyardCardForTest(g.Seats[1], "Theirs")
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, c.InstanceID, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(theirs)}); !errors.Is(err, game.ErrIllegalTarget) {
		t.Errorf("targeting an opponent's graveyard card: err = %v, want ErrIllegalTarget", err)
	}

	effectstest.AdvanceToUpkeepOf(t, g, 1) // another seat's upkeep: nothing happens
	if g.IsPreparedForEffect(c.InstanceID) {
		t.Fatal("prepared on an opponent's upkeep")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.IsPreparedForEffect(c.InstanceID) {
		t.Fatal("not prepared after its controller's upkeep")
	}
	if effectstest.CountPrepareCopies(g) != 1 {
		t.Errorf("%d copies in exile, want 1", effectstest.CountPrepareCopies(g))
	}
}

func TestStingerquillVoxmancerVerseHitsOnlyAnOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	c := fraPrepB(me.ID, fraStingerquill, nil, "Stingerquill Voxmancer", "Creature — Goblin Sorcerer", 1, 2, "Vicious Verse", "Sorcery")
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.CastFromHandAndResolve(t, g, me, c)
	if g.IsPreparedForEffect(c.InstanceID) {
		t.Fatal("entered prepared; only the upkeep trigger prepares it")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.IsPreparedForEffect(c.InstanceID) {
		t.Fatal("not prepared after upkeep")
	}
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	cp, _ := effectstest.PrepareCopyOf(g, "Vicious Verse")
	err := g.CastSpell(me.ID, cp.InstanceID, game.CastSpellParams{FromZone: string(game.ZoneExile),
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("targeting yourself: err = %v, want ErrIllegalTarget", err)
	}
	life := opp.Life
	effectstest.FraBCastCopy(t, g, me, "Vicious Verse", game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	if opp.Life != life-1 {
		t.Errorf("opponent life %d, want %d", opp.Life, life-1)
	}
}

func TestWhiplashWordsmithEntersPrepared(t *testing.T) {
	spec, ok := effects.Lookup(fraWhiplashWordsmith)
	if !ok || spec.Completeness != effects.CompletenessFull {
		t.Fatalf("spec = %+v ok=%v, want a complete card", spec, ok)
	}
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	c := fraPrepB(me.ID, fraWhiplashWordsmith, nil, "Whiplash Wordsmith", "Creature — Vampire Sorcerer", 3, 3, "Vicious Verse", "Sorcery")
	effectstest.CastFromHandAndResolve(t, g, me, c)
	if !g.IsPreparedForEffect(c.InstanceID) {
		t.Fatal("did not enter prepared")
	}
	life := opp.Life
	effectstest.FraBCastCopy(t, g, me, "Vicious Verse", game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	if opp.Life != life-1 {
		t.Errorf("opponent life %d, want %d", opp.Life, life-1)
	}
}

func TestPompousBattlemageHasProwessAndImprovisedActRummages(t *testing.T) {
	for _, discard := range []bool{true, false} {
		name := "declined"
		if discard {
			name = "discarded"
		}
		t.Run(name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
			c := fraPrepB(me.ID, fraPompousBattlemage, []string{"prowess"}, "Pompous Battlemage", "Creature — Goblin Sorcerer", 1, 1, "Improvised Act", "Sorcery")
			effectstest.CastFromHandAndResolve(t, g, me, c)
			if !effectstest.EffectiveAbilitiesContain(t, g, c.InstanceID, "prowess") {
				t.Error("no prowess")
			}
			hand, gy := me.Hand.Size(), me.Graveyard.Size()
			pitch := me.Hand.Cards[0].InstanceID
			cp, _ := effectstest.PrepareCopyOf(g, "Improvised Act")
			if err := g.CastSpell(me.ID, cp.InstanceID, game.CastSpellParams{FromZone: string(game.ZoneExile)}); err != nil {
				t.Fatal(err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if discard {
				effectstest.AnswerDiscard(t, g, me.ID, pitch)
			} else {
				effectstest.AnswerDiscard(t, g, me.ID)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if got := me.Hand.Size(); got != hand {
				t.Errorf("hand %d, want %d", got, hand)
			}
			want := gy
			if discard {
				want++
			}
			if got := me.Graveyard.Size(); got != want {
				t.Errorf("graveyard %d, want %d", got, want)
			}
		})
	}
}

func TestVariableChaserArcOfFortuneEachPlayerMayWheel(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	c := fraPrepB(me.ID, fraVariableChaser, []string{"flying", "prowess"}, "Variable Chaser", "Creature — Human Wizard", 2, 3, "Arc of Fortune", "Sorcery")
	effectstest.CastFromHandAndResolve(t, g, me, c)
	cp, _ := effectstest.PrepareCopyOf(g, "Arc of Fortune")
	if err := g.CastSpell(me.ID, cp.InstanceID, game.CastSpellParams{FromZone: string(game.ZoneExile)}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	oppHand := opp.Hand.Size()
	oppHandIDs := map[uuid.UUID]bool{}
	for _, h := range opp.Hand.Cards {
		oppHandIDs[h.InstanceID] = true
	}
	oldHand := map[uuid.UUID]bool{}
	for _, h := range me.Hand.Cards {
		oldHand[h.InstanceID] = true
	}
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.AnswerMayChoice(t, g, opp.ID, false)
	for _, p := range g.Seats[2:] {
		effectstest.AnswerMayChoice(t, g, p.ID, false)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != 7 {
		t.Errorf("caster hand = %d, want 7", got)
	}
	for _, h := range me.Hand.Cards {
		if oldHand[h.InstanceID] {
			t.Error("a card from the old hand is still in hand")
		}
	}
	if got := opp.Hand.Size(); got != oppHand {
		t.Errorf("declining opponent hand = %d, want %d", got, oppHand)
	}
	for _, h := range opp.Hand.Cards {
		if !oppHandIDs[h.InstanceID] {
			t.Error("declining opponent's hand changed")
		}
	}
}

func TestVigorbloomVanguardGrantsVigilanceToCountersAndSeedSutureAddsOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	c := fraPrepB(me.ID, fraVigorbloomVanguard, nil, "Vigorbloom Vanguard", "Creature — Troll Druid", 2, 2, "Seed Suture", "Sorcery")
	effectstest.CastFromHandAndResolve(t, g, me, c)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, g.Seats[1].ID, "Theirs", 2, 2)
	if effectstest.EffectiveAbilitiesContain(t, g, bear, "vigilance") {
		t.Fatal("vigilance without a counter")
	}
	life := me.Life
	effectstest.FraBCastCopy(t, g, me, "Seed Suture", game.TargetRef{Kind: game.TargetCard, ID: bear})
	if !effectstest.EffectiveAbilitiesContain(t, g, bear, "vigilance") {
		t.Error("the countered creature has no vigilance")
	}
	if me.Life != life+1 {
		t.Errorf("life %d, want %d", me.Life, life+1)
	}
	if effectstest.EffectiveAbilitiesContain(t, g, c.InstanceID, "vigilance") {
		t.Error("the Vanguard has vigilance without a counter on it")
	}
	if effectstest.EffectiveAbilitiesContain(t, g, theirs, "vigilance") {
		t.Error("an opponent's creature got vigilance")
	}
}
