package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// GAP2 batch 9 (S82, #5948): the cards that waited on the #6093 seam (new
// durations and timed cost reductions) and the #6095 seam (untap-step
// replacements and turn structure).

const (
	gap2B09BrazenCannonadeOracle = "b9212658-8406-489f-80d6-7af42df8956c"
	gap2B09LiaraOracle           = "fcc7ed01-ff64-4ffe-9b83-daa4132fd92d"
	gap2B09WillKenrithOracle     = "20dbfc9a-172e-4065-bfde-d40de3131787"
	gap2B09ConeOfColdOracle      = "56fd0752-9fb0-4c05-90e5-c3b268b57c6f"
	gap2B09LeechcraftOracle      = "a5d803bc-7fd4-4570-8cd6-cb138727a66e"
	gap2B09NinthDoctorOracle     = "c99f59c0-5cce-41a4-908f-d34e2c697419"
	gap2B09TsabosWebOracle       = "96a4e515-2b4e-4c2d-aa1a-963597484723"
	gap2B09PandoricaOracle       = "94d5c904-6504-4df2-b242-daf88a988475"
)

func TestS82Gap2B09CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gap2B09BrazenCannonadeOracle: "Brazen Cannonade",
		gap2B09LiaraOracle:           "Commander Liara Portyr",
		gap2B09WillKenrithOracle:     "Will Kenrith",
		gap2B09ConeOfColdOracle:      "Cone of Cold",
		gap2B09LeechcraftOracle:      "Bewitching Leechcraft",
		gap2B09NinthDoctorOracle:     "The Ninth Doctor",
		gap2B09TsabosWebOracle:       "Tsabo's Web",
		gap2B09PandoricaOracle:       "The Pandorica",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v", name, spec)
		}
	}
}

// gap2B09Library gives the player n filler cards and then `top` on top.
func gap2B09Library(p *game.Player, n int, top game.Card) uuid.UUID {
	p.Library.Cards = nil
	for i := 0; i < n; i++ {
		p.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", ManaCost: "{1}", Owner: p.ID, Controller: p.ID})
	}
	if top.InstanceID == uuid.Nil {
		top.InstanceID = uuid.New()
	}
	top.Owner, top.Controller = p.ID, p.ID
	p.Library.PushTop(top)
	return top.InstanceID
}

func gap2B09ZoneOf(g *game.Game, id uuid.UUID) game.ZoneKind {
	var out game.ZoneKind
	g.ReadSnapshot(func() {
		if z := g.FindCardZoneForEffect(id); z != nil {
			out = z.Kind
		}
	})
	return out
}

func gap2B09Tapped(g *game.Game, id uuid.UUID) bool {
	var out bool
	g.ReadSnapshot(func() {
		if c := effectstest.FindBattlefieldCardForTest(g, id); c != nil {
			out = c.Tapped
		}
	})
	return out
}

func gap2B09SetTapped(g *game.Game, id uuid.UUID, tapped bool) {
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, id).Tapped = tapped })
}

// gap2B09Price prices a spell for `caster` cast from `from`.
func gap2B09Price(t *testing.T, g *game.Game, caster uuid.UUID, typeLine, cost string, from game.ZoneKind) string {
	t.Helper()
	card := game.Card{InstanceID: uuid.New(), Name: "Priced Spell", TypeLine: typeLine, ManaCost: cost, Owner: caster, Controller: caster}
	base, err := game.ParseCost(cost)
	if err != nil {
		t.Fatalf("ParseCost: %v", err)
	}
	var out game.ParsedCost
	g.WithWriteLock(func() {
		var err error
		out, err = g.ApplyCostModifiersForEffect(base, game.CostQuery{Game: g, Card: card, Controller: caster, FromZone: from})
		if err != nil {
			t.Fatalf("ApplyCostModifiers: %v", err)
		}
	})
	return out.String()
}

// --- Brazen Cannonade ------------------------------------------------------

// Raid: after an attack, the postcombat main phase exiles the top card,
// which stays playable through the other players' turns and the start of
// your next turn, and stops being playable once that turn's combat ends.
func TestGap2BrazenCannonadeRaidExilesUntilEndOfCombatOnYourNextTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Brazen Cannonade", "Enchantment", gap2B09BrazenCannonadeOracle, false)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	top := gap2B09Library(me, 10, game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"})
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	effectstest.PassPriorityAroundTable(t, g)
	if gap2B09ZoneOf(g, top) != game.ZoneExile {
		t.Fatalf("the top card is in %q, want exile", gap2B09ZoneOf(g, top))
	}
	if !effectstest.B12ExiledPlayableBy(t, g, top, me.ID) {
		t.Fatal("the exiled card is not playable this turn")
	}
	effectstest.AdvanceToStepOf(t, g, 2, game.StepPrecombatMain)
	if !effectstest.B12ExiledPlayableBy(t, g, top, me.ID) {
		t.Fatal("the permission ended during another player's turn")
	}
	effectstest.AdvanceToStepOf(t, g, 0, game.StepPrecombatMain)
	if !effectstest.B12ExiledPlayableBy(t, g, top, me.ID) {
		t.Fatal("the permission ended before combat on your next turn")
	}
	effectstest.AdvanceToStepOf(t, g, 0, game.StepPostcombatMain)
	if effectstest.B12ExiledPlayableBy(t, g, top, me.ID) {
		t.Fatal("the card is still playable after combat on your next turn")
	}
}

// Without an attack the raid trigger does nothing.
func TestGap2BrazenCannonadeRaidNeedsAnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Brazen Cannonade", "Enchantment", gap2B09BrazenCannonadeOracle, false)
	effectstest.AdvanceToMain(t, g)
	top := gap2B09Library(me, 10, game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"})
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	effectstest.PassPriorityAroundTable(t, g)
	if gap2B09ZoneOf(g, top) != game.ZoneLibrary {
		t.Fatalf("the top card is in %q with no attack this turn, want the library", gap2B09ZoneOf(g, top))
	}
}

// An attacking creature you control dying deals 2 damage to each opponent.
func TestGap2BrazenCannonadeAttackerDiesDamagesEachOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Brazen Cannonade", "Enchantment", gap2B09BrazenCannonadeOracle, false)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	gap2B09Library(me, 10, game.Card{Name: "Filler", TypeLine: "Sorcery"})
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = p.Life
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, bear); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		want := 2
		if p.ID == me.ID {
			want = 0
		}
		if got := before[p.ID] - p.Life; got != want {
			t.Errorf("seat lost %d life, want %d", got, want)
		}
	}
}

// --- Commander Liara Portyr --------------------------------------------------

// Attacking two players: the top two cards are exiled and castable this
// turn, and spells cast from exile cost {2} less this turn.
func TestGap2LiaraPortyrExilesAndDiscountsPerPlayerAttacked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	liara := effectstest.PushCatalogPermanent(g, me.ID, "Commander Liara Portyr", "Legendary Creature — Human Soldier", gap2B09LiaraOracle, false)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	first, second := uuid.New(), uuid.New()
	gap2B09Library(me, 10, game.Card{InstanceID: second, Name: "Second", TypeLine: "Creature — Bear", ManaCost: "{3}{G}"})
	me.Library.PushTop(game.Card{InstanceID: first, Name: "First", TypeLine: "Sorcery", ManaCost: "{2}{R}", Owner: me.ID, Controller: me.ID})
	effectstest.AttackInDeclareAttackers(t, g, map[uuid.UUID]uuid.UUID{liara: opp.ID, bear: other.ID})
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{first, second} {
		if !effectstest.B12ExiledPlayableBy(t, g, id, me.ID) {
			t.Errorf("an exiled card is not castable this turn (zone %q)", gap2B09ZoneOf(g, id))
		}
	}
	if got := gap2B09Price(t, g, me.ID, "Creature — Bear", "{3}{G}", game.ZoneExile); got != "{1}{G}" {
		t.Errorf("a spell cast from exile costs %q, want {1}{G}", got)
	}
	if got := gap2B09Price(t, g, me.ID, "Creature — Bear", "{3}{G}", game.ZoneHand); got != "{3}{G}" {
		t.Errorf("a spell cast from hand costs %q, want it unreduced", got)
	}
	effectstest.AdvanceToStepOf(t, g, 1, game.StepUpkeep)
	if got := gap2B09Price(t, g, me.ID, "Creature — Bear", "{3}{G}", game.ZoneExile); got != "{3}{G}" {
		t.Errorf("the reduction outlived the turn: %q", got)
	}
	if effectstest.B12ExiledPlayableBy(t, g, first, me.ID) {
		t.Error("the exiled card is still castable after the turn")
	}
}

// --- Will Kenrith ------------------------------------------------------------

func TestGap2WillKenrithPlusTwoShrinksUntilYourNextTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	gap2B09Library(me, 20, game.Card{Name: "Filler", TypeLine: "Sorcery"})
	will := effectstest.GuffbWalker(g, me.ID, "Will Kenrith", "Legendary Planeswalker — Will", gap2B09WillKenrithOracle, 4)
	a := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Giant", 5, 5)
	b := effectstest.B12Creature(g, opp.ID, "Flier", "Creature — Bird", 4, 1)
	effectstest.B16Activate(t, g, me.ID, will, 0, game.ActivateAbilityParams{Targets: effectstest.CardTargets(a, b)})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if p, tough := effectstest.MkmB1PT(t, g, id); p != 0 || tough != 3 {
			t.Errorf("a target is %d/%d, want 0/3", p, tough)
		}
	}
	effectstest.AdvanceToStepOf(t, g, 2, game.StepUpkeep)
	if p, _ := effectstest.MkmB1PT(t, g, a); p != 0 {
		t.Error("the effect ended before Will's controller's next turn")
	}
	effectstest.AdvanceToStepOf(t, g, 0, game.StepUpkeep)
	if p, tough := effectstest.MkmB1PT(t, g, a); p != 5 || tough != 5 {
		t.Errorf("after your next turn began the target is %d/%d, want 5/5", p, tough)
	}
}

func TestGap2WillKenrithMinusTwoDrawsAndDiscountsThatPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	gap2B09Library(opp, 20, game.Card{Name: "Filler", TypeLine: "Sorcery"})
	will := effectstest.GuffbWalker(g, me.ID, "Will Kenrith", "Legendary Planeswalker — Will", gap2B09WillKenrithOracle, 4)
	hand := opp.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, will, 1, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Hand.Size() - hand; got != 2 {
		t.Errorf("the target drew %d, want 2", got)
	}
	for _, tc := range []struct {
		caster         uuid.UUID
		typeLine, want string
	}{
		{opp.ID, "Instant", "{1}{U}"},
		{opp.ID, "Sorcery", "{1}{U}"},
		{opp.ID, "Legendary Planeswalker — Test", "{1}{U}"},
		{opp.ID, "Creature — Bear", "{3}{U}"},
		{me.ID, "Instant", "{3}{U}"},
	} {
		if got := gap2B09Price(t, g, tc.caster, tc.typeLine, "{3}{U}", game.ZoneHand); got != tc.want {
			t.Errorf("%s: costs %q, want %q", tc.typeLine, got, tc.want)
		}
	}
	effectstest.AdvanceToStepOf(t, g, 0, game.StepUpkeep)
	if got := gap2B09Price(t, g, opp.ID, "Instant", "{3}{U}", game.ZoneHand); got != "{3}{U}" {
		t.Errorf("the reduction outlived your next turn beginning: %q", got)
	}
}

// −8: the target player's emblem copies their instants and sorceries.
func TestGap2WillKenrithEmblemCopiesInstantsAndSorceries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.Msc18Library(me, 20)
	effectstest.AdvanceToMain(t, g)
	will := effectstest.GuffbWalker(g, me.ID, "Will Kenrith", "Legendary Planeswalker — Will", gap2B09WillKenrithOracle, 8)
	effectstest.B16Activate(t, g, me.ID, will, 2, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(me.ID)}})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Emblems.Size() != 1 || opp.Emblems.Size() != 0 {
		t.Fatalf("emblems: me %d, opp %d; want 1 and 0", me.Emblems.Size(), opp.Emblems.Size())
	}
	hand := me.Hand.Size()
	effectstest.Msc18DivinationCast(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 4 {
		t.Errorf("drew %d, want 4: the Divination and its copy", got)
	}
}

// --- Cone of Cold ------------------------------------------------------------

func TestGap2ConeOfColdByTheRoll(t *testing.T) {
	for _, tc := range []struct {
		natural       int
		frozen, enter bool
	}{{5, false, false}, {15, true, false}, {20, true, true}} {
		cone := uuid.New()
		g := effectstest.D5681Table(t, cone, tc.natural)
		me, opp, _ := effectstest.BlbB1Seats(g)
		mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
		theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
		effectstest.D5681Spell(g, me, cone, "Cone of Cold", "Sorcery", "", gap2B09ConeOfColdOracle)
		if err := g.CastSpell(me.ID, cone, game.CastSpellParams{Strict: true}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !gap2B09Tapped(g, theirs) || gap2B09Tapped(g, mine) {
			t.Fatalf("roll %d: theirs tapped %v, mine tapped %v; want only theirs", tc.natural, gap2B09Tapped(g, theirs), gap2B09Tapped(g, mine))
		}
		// A creature entering under an opponent's control.
		late := uuid.New()
		g.WithWriteLock(func() {
			opp.Hand.PushTop(game.Card{InstanceID: late, Name: "Late", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
		})
		if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: opp.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, late); err != nil {
			t.Fatalf("MoveCardByID: %v", err)
		}
		if got := gap2B09Tapped(g, late); got != tc.enter {
			t.Errorf("roll %d: a creature entering under an opponent's control entered tapped %v, want %v", tc.natural, got, tc.enter)
		}
		effectstest.AdvanceToStepOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats), game.StepUpkeep)
		if got := gap2B09Tapped(g, theirs); got != tc.frozen {
			t.Errorf("roll %d: after its controller's untap step the creature is tapped %v, want %v", tc.natural, got, tc.frozen)
		}
	}
}

// --- Bewitching Leechcraft -----------------------------------------------------

func TestGap2BewitchingLeechcraftTapsAndUntapsOnlyForACounter(t *testing.T) {
	aura := uuid.New()
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp, _ := effectstest.BlbB1Seats(g)
	host := effectstest.B12Creature(g, opp.ID, "Host", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(host, game.CounterPlusOne, 1) })
	effectstest.D5681Spell(g, me, aura, "Bewitching Leechcraft", "Enchantment — Aura", "", gap2B09LeechcraftOracle)
	if err := g.CastSpell(me.ID, aura, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(host)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !gap2B09Tapped(g, host) {
		t.Fatal("the entry trigger did not tap the enchanted creature")
	}
	oppSeat := (g.Turn.ActiveSeat + 1) % len(g.Seats)
	effectstest.AdvanceToStepOf(t, g, oppSeat, game.StepUpkeep)
	if gap2B09Tapped(g, host) {
		t.Fatal("with a +1/+1 counter the creature did not untap")
	}
	var counters int
	g.ReadSnapshot(func() { counters = effectstest.FindBattlefieldCardForTest(g, host).Counters[game.CounterPlusOne] })
	if counters != 0 {
		t.Fatalf("the creature has %d +1/+1 counters, want 0", counters)
	}
	gap2B09SetTapped(g, host, true)
	effectstest.AdvanceToStepOf(t, g, (oppSeat+1)%len(g.Seats), game.StepUpkeep)
	effectstest.AdvanceToStepOf(t, g, oppSeat, game.StepUpkeep)
	if !gap2B09Tapped(g, host) {
		t.Error("with no counter the creature untapped during its controller's untap step")
	}
}

// --- The Ninth Doctor ----------------------------------------------------------

func TestGap2NinthDoctorUntapInYourUntapStepAddsAnUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	gap2B09Library(me, 20, game.Card{Name: "Filler", TypeLine: "Sorcery"})
	doc := effectstest.PushCatalogPermanent(g, me.ID, "The Ninth Doctor", "Legendary Creature — Time Lord Doctor", gap2B09NinthDoctorOracle, false)
	gap2B09SetTapped(g, doc, true)
	effectstest.AdvanceToStepOf(t, g, 1, game.StepUpkeep)
	effectstest.AdvanceToStepOf(t, g, 0, game.StepUpkeep)
	if gap2B09Tapped(g, doc) {
		t.Fatal("The Ninth Doctor did not untap")
	}
	effectstest.PassPriorityAroundTable(t, g)
	upcoming := g.UpcomingStepsForEffect()
	if len(upcoming) == 0 || upcoming[0].Step != game.StepUpkeep {
		t.Fatalf("the next step is %+v, want an additional upkeep", upcoming)
	}
	effectstest.AdvanceToStepOf(t, g, 0, game.StepDraw)
	if n := g.TurnTally.StepsBegun[game.StepUpkeep]; n != 2 {
		t.Errorf("%d upkeep steps began this turn, want 2", n)
	}
}

// Untapped going into the untap step, it does not become untapped, and an
// untap outside the untap step does not trigger it.
func TestGap2NinthDoctorOnlyAnUntapInYourUntapStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	gap2B09Library(me, 20, game.Card{Name: "Filler", TypeLine: "Sorcery"})
	doc := effectstest.PushCatalogPermanent(g, me.ID, "The Ninth Doctor", "Legendary Creature — Time Lord Doctor", gap2B09NinthDoctorOracle, false)
	effectstest.AdvanceToMain(t, g)
	gap2B09SetTapped(g, doc, true)
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(doc) })
	if n := effectstest.TriggersOnStackFrom(g, doc); n != 0 {
		t.Fatalf("an untap in the main phase put %d triggers on the stack", n)
	}
	effectstest.AdvanceToStepOf(t, g, 1, game.StepUpkeep)
	effectstest.AdvanceToStepOf(t, g, 0, game.StepUpkeep)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToStepOf(t, g, 0, game.StepDraw)
	if n := g.TurnTally.StepsBegun[game.StepUpkeep]; n != 1 {
		t.Errorf("%d upkeep steps began this turn, want 1", n)
	}
}

// --- Tsabo's Web ---------------------------------------------------------------

func TestGap2TsabosWebHoldsLandsWithNonManaAbilities(t *testing.T) {
	web := uuid.New()
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp, _ := effectstest.BlbB1Seats(g)
	gap2B09Library(me, 20, game.Card{Name: "Filler", TypeLine: "Sorcery"})
	wilds := effectstest.PushCatalogPermanent(g, opp.ID, "Evolving Wilds", "Land", effectstest.EvolvingWildsOracle, false)
	forest := effectstest.PushCatalogPermanent(g, opp.ID, "Forest", "Basic Land — Forest", "", false)
	gap2B09SetTapped(g, wilds, true)
	gap2B09SetTapped(g, forest, true)
	hand := me.Hand.Size()
	effectstest.D5681Spell(g, me, web, "Tsabo's Web", "Artifact", "", gap2B09TsabosWebOracle)
	if err := g.CastSpell(me.ID, web, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d on entry, want 1", got)
	}
	effectstest.AdvanceToStepOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats), game.StepUpkeep)
	if !gap2B09Tapped(g, wilds) {
		t.Error("a land with a non-mana activated ability untapped")
	}
	if gap2B09Tapped(g, forest) {
		t.Error("a land with only a mana ability stayed tapped")
	}
}

// --- The Pandorica -------------------------------------------------------------

func TestGap2PandoricaHoldsThePermanentOutUntilItUntaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gap2B09Library(me, 20, game.Card{Name: "Filler", TypeLine: "Sorcery"})
	box := effectstest.PushCatalogPermanent(g, me.ID, "The Pandorica", "Legendary Artifact", gap2B09PandoricaOracle, false)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	gap2B09SetTapped(g, bear, true)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "W", "W")
	effectstest.B16Activate(t, g, me.ID, box, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	var out bool
	g.ReadSnapshot(func() { out = g.IsPhasedOutForEffect(bear) })
	if !out {
		t.Fatal("the target did not phase out")
	}
	if g.Battlefield.Contains(bear) {
		t.Fatal("the phased-out target is still on the battlefield slice")
	}
	// Its controller's untap step does not phase it in: The Pandorica is
	// still tapped.
	effectstest.AdvanceToStepOf(t, g, 1, game.StepPrecombatMain)
	g.ReadSnapshot(func() { out = g.IsPhasedOutForEffect(bear) })
	if !out {
		t.Fatal("the target phased in while The Pandorica stayed tapped")
	}
	// Its controller may choose not to untap The Pandorica, and the
	// target stays out.
	choice := effectstest.AdvanceToUntapChoiceOf(t, g, 0)
	if err := g.ResolveUntapChoice(choice.ID, me.ID, nil); err != nil {
		t.Fatalf("ResolveUntapChoice: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.ReadSnapshot(func() { out = g.IsPhasedOutForEffect(bear) })
	if !out || !gap2B09Tapped(g, box) {
		t.Fatal("with The Pandorica left tapped the target phased in")
	}
	// The Pandorica untaps during its controller's untap step: the trigger
	// phases the target in.
	effectstest.AdvanceToStepOf(t, g, 1, game.StepUpkeep)
	choice = effectstest.AdvanceToUntapChoiceOf(t, g, 0)
	if err := g.ResolveUntapChoice(choice.ID, me.ID, []uuid.UUID{box}); err != nil {
		t.Fatalf("ResolveUntapChoice: %v", err)
	}
	effectstest.AdvanceToStepOf(t, g, 0, game.StepUpkeep)
	effectstest.PassPriorityAroundTable(t, g)
	g.ReadSnapshot(func() { out = g.IsPhasedOutForEffect(bear) })
	if out {
		t.Fatal("the target is still phased out after The Pandorica untapped")
	}
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the target is not back on the battlefield")
	}
	// It was untapped before it phased out (the ability untaps it first).
	if gap2B09Tapped(g, bear) {
		t.Error("the target came back tapped; the ability untaps it before it phases out")
	}
}

// The Pandorica leaving the battlefield phases the permanent in.
func TestGap2PandoricaLeavingPhasesItIn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gap2B09Library(me, 20, game.Card{Name: "Filler", TypeLine: "Sorcery"})
	box := effectstest.PushCatalogPermanent(g, me.ID, "The Pandorica", "Legendary Artifact", gap2B09PandoricaOracle, false)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "W", "W")
	effectstest.B16Activate(t, g, me.ID, box, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, box); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	var out bool
	g.ReadSnapshot(func() { out = g.IsPhasedOutForEffect(bear) })
	if out {
		t.Fatal("the target is still phased out after The Pandorica left the battlefield")
	}
}

// Sorcery speed, and "another" target: the ability refuses itself.
func TestGap2PandoricaRefusesItselfAndInstantSpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	box := effectstest.PushCatalogPermanent(g, me.ID, "The Pandorica", "Legendary Artifact", gap2B09PandoricaOracle, false)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "W", "W")
	if err := g.ActivateCatalogAbility(me.ID, box, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(box)}); err == nil {
		t.Error("The Pandorica targeted itself")
	}
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if err := g.ActivateCatalogAbility(me.ID, box, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)}); err == nil {
		t.Error("The Pandorica was activated outside a main phase")
	}
}
