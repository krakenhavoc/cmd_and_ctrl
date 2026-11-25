package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b100_cards_test.go — Jumpstart, S82 older sets batch 100 (#5510):
// Auras.

const (
	jmpFaceOfDivinity = "90b5ed05-7a67-4bd3-a8c4-2a4565d084fc"
	jmpEternalThirst  = "a1b66fbf-14a3-4582-8414-53f432f96dad"
	jmpVastwood       = "10172c40-7169-4098-9452-b35797d27065"
	jmpNarcolepsy     = "3da1412e-a72c-4da1-8848-37145ca6614a"
	jmpImplant        = "4fc0f58c-5f7f-4a32-8f14-cae696672beb"
	jmpDiadem         = "1e4de77b-27cb-4afd-8d4b-1cc89f2a6986"
	jmpApproach       = "7a1798d3-33fb-4092-a843-45a25f3b4e82"
)

func TestS82OldB100EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		jmpFaceOfDivinity: "Face of Divinity", jmpEternalThirst: "Eternal Thirst", jmpVastwood: "Vastwood Zendikon",
		jmpNarcolepsy: "Narcolepsy", jmpImplant: "Parasitic Implant", jmpDiadem: "Lightning Diadem",
		jmpApproach: "Death's Approach",
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

func jmp100Forest(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: owner, Controller: owner})
}

func TestS82OldB100FaceOfDivinityGivesFirstStrikeAndLifelinkOnlyWithAnotherAura(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Face of Divinity", effectstest.AuraTypeLine, jmpFaceOfDivinity, effectstest.CardRefs(host))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, host); p != 4 || tt != 4 {
		t.Fatalf("the enchanted creature is %d/%d, want 4/4", p, tt)
	}
	for _, kw := range []string{"first strike", "lifelink"} {
		if effectstest.HasEffectiveKeyword(t, g, host, kw) {
			t.Errorf("the creature has %s with only Face of Divinity on it", kw)
		}
	}
	other := effectstest.Cmdr2B19Attach(g, me.ID, host, "Other Aura", "")
	for _, kw := range []string{"first strike", "lifelink"} {
		if !effectstest.HasEffectiveKeyword(t, g, host, kw) {
			t.Errorf("the creature lacks %s with another Aura on it", kw)
		}
	}
	effectstest.KillCreature(t, g, me.ID, other)
	for _, kw := range []string{"first strike", "lifelink"} {
		if effectstest.HasEffectiveKeyword(t, g, host, kw) {
			t.Errorf("the creature still has %s after the other Aura left", kw)
		}
	}
}

func TestS82OldB100EternalThirstGrantsLifelinkAndCountersForOpponentsDeaths(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 1, 1)
	theirsToo := effectstest.B12Creature(g, g.Seats[2].ID, "Theirs too", "Creature — Bear", 1, 1)
	effectstest.CastCatalogSpell(t, g, "Eternal Thirst", effectstest.AuraTypeLine, jmpEternalThirst, effectstest.CardRefs(host))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, host, "lifelink") {
		t.Error("the enchanted creature has no lifelink")
	}
	effectstest.KillCreature(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B12Card(t, g, host).Counters[game.CounterPlusOne]; got != 0 {
		t.Errorf("your own creature dying gave %d counters", got)
	}
	effectstest.KillCreature(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.KillCreature(t, g, g.Seats[2].ID, theirsToo)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B12Card(t, g, host).Counters[game.CounterPlusOne]; got != 2 {
		t.Errorf("two opposing creatures died and the host has %d +1/+1 counters, want 2", got)
	}
}

func TestS82OldB100VastwoodZendikonAnimatesTheLandAndReturnsItWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := jmp100Forest(g, me.ID)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Vastwood Zendikon", effectstest.AuraTypeLine, jmpVastwood, effectstest.CardRefs(land))
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.B12Card(t, g, land)
	if !c.IsLand() || !c.IsCreature() || !c.HasSubtype("Elemental") {
		t.Fatalf("the enchanted land is %q, want a land Elemental creature", c.TypeLine)
	}
	if p, tt := effectstest.PtOf(t, g, land); p != 6 || tt != 4 {
		t.Errorf("the land is %d/%d, want 6/4", p, tt)
	}
	if cols := effectstest.EffectiveColors(t, g, land); len(cols) != 1 || cols[0] != "G" {
		t.Errorf("the land's colours are %v, want green", cols)
	}
	effectstest.KillCreature(t, g, me.ID, land)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(land) {
		t.Error("the dead land did not return to its owner's hand")
	}
}

func TestS82OldB100VastwoodZendikonEnchantsAnyPlayersLandButNotACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	theirs := jmp100Forest(g, opp.ID)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Vastwood Zendikon", effectstest.AuraTypeLine, jmpVastwood, effectstest.CardRefs(bear)); err == nil {
		t.Error("the Zendikon was cast on a creature")
	}
	effectstest.CastCatalogSpell(t, g, "Vastwood Zendikon", effectstest.AuraTypeLine, jmpVastwood, effectstest.CardRefs(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B12Card(t, g, theirs).IsCreature() {
		t.Error("an opponent's land was not animated")
	}
}

func TestS82OldB100NarcolepsyTapsTheEnchantedCreatureEveryUpkeepOnlyIfUntapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	host := effectstest.B12Creature(g, opp.ID, "Host", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Narcolepsy", effectstest.AuraTypeLine, jmpNarcolepsy, effectstest.CardRefs(host))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12Card(t, g, host).Tapped {
		t.Fatal("the creature was tapped before any upkeep")
	}
	// Seat 1's upkeep (not the Aura's controller's): "each upkeep".
	effectstest.AppToUpkeepOf(t, g, opp.Seat)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B12Card(t, g, host).Tapped {
		t.Error("the creature was not tapped at an opponent's upkeep")
	}
	_ = me
}

func TestS82OldB100NarcolepsyDoesNothingToATappedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[0]
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Bear", 2, 2)
	effectstest.Cmdr2B19Attach(g, opp.ID, host, "Narcolepsy", jmpNarcolepsy)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, host).Tapped = true })
	effectstest.AppToUpkeepOf(t, g, 1)
	if len(g.StackMeta) != 0 || len(g.PendingTriggers) != 0 {
		t.Errorf("a tapped creature put an ability on the stack (stack %d, pending %d)", len(g.StackMeta), len(g.PendingTriggers))
	}
}

func TestS82OldB100ParasiticImplantMakesTheCreaturesControllerSacrificeItAndYouGetAMyr(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	host := effectstest.B12Creature(g, opp.ID, "Host", "Creature — Bear", 2, 2)
	effectstest.Cmdr2B19Attach(g, me.ID, host, "Parasitic Implant", jmpImplant)
	effectstest.AppToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(host) {
		t.Error("the enchanted creature was not sacrificed")
	}
	myr := effectstest.FindBattlefieldByName(g, "Phyrexian Myr")
	if myr == uuid.Nil {
		t.Fatal("no Phyrexian Myr token")
	}
	c := effectstest.B12Card(t, g, myr)
	if c.Controller != me.ID || !c.IsArtifact() || !c.IsCreature() || !c.IsColorless() {
		t.Errorf("the token is %+v, want your colorless artifact creature", c)
	}
	if p, tt := effectstest.PtOf(t, g, myr); p != 1 || tt != 1 {
		t.Errorf("the Myr is %d/%d, want 1/1", p, tt)
	}
}

func TestS82OldB100ParasiticImplantStaysSilentOnOtherPlayersUpkeeps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	host := effectstest.B12Creature(g, opp.ID, "Host", "Creature — Bear", 2, 2)
	effectstest.Cmdr2B19Attach(g, me.ID, host, "Parasitic Implant", jmpImplant)
	effectstest.AppToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(host) {
		t.Error("the creature was sacrificed on someone else's upkeep")
	}
}

func TestS82OldB100LightningDiademPumpsAndShootsAnyTargetWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Bear", 2, 2)
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Lightning Diadem", effectstest.AuraTypeLine, jmpDiadem, effectstest.CardRefs(host))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		if err := g.ResolvePickTarget(effectstest.LatestPickTarget(g, me.ID).ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
			t.Fatalf("pick: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if p, tt := effectstest.PtOf(t, g, host); p != 4 || tt != 4 {
		t.Errorf("the enchanted creature is %d/%d, want 4/4", p, tt)
	}
	if opp.Life != life-2 {
		t.Errorf("opponent life %d, want %d", opp.Life, life-2)
	}
}

func TestS82OldB100DeathsApproachShrinksByTheControllersGraveyardCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := effectstest.B12Creature(g, opp.ID, "Host", "Creature — Ogre", 5, 5)
	effectstest.Cmdr2B19Attach(g, me.ID, host, "Death's Approach", jmpApproach)
	for _, tl := range []string{"Creature — Bear", "Artifact Creature — Golem", "Instant"} {
		opp.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Card", TypeLine: tl, Owner: opp.ID, Controller: opp.ID})
	}
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Mine", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if p, tt := effectstest.PtOf(t, g, host); p != 3 || tt != 3 {
		t.Errorf("the enchanted creature is %d/%d, want 3/3 (two creature cards in its controller's graveyard)", p, tt)
	}
	more := effectstest.B12Creature(g, opp.ID, "More", "Creature — Bear", 1, 1)
	effectstest.KillCreature(t, g, opp.ID, more)
	if p, _ := effectstest.PtOf(t, g, host); p != 2 {
		t.Errorf("after a third creature card the power is %d, want 2", p)
	}
}
