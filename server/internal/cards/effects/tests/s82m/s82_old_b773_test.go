package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b773_test.go — S82 OLD batch 773 (Weatherlight: activated
// abilities, #5629), each card driven through the real activation path.

const (
	oldb773ZombieScavengers = "8c5298da-6759-444e-8870-6c96d226c940"
	oldb773ErtaisFamiliar   = "711cd80d-c58d-4c2a-87db-d2e4c75f60d3"
	oldb773BosiumStrip      = "2fa6f1b1-00af-433f-b8e1-36db99cd9bba"
	oldb773OrcishSettlers   = "8b8d4ae0-0d18-4f06-a883-0edff494d06b"
	oldb773XanthicStatue    = "a8de92c4-84cb-40e7-8e34-428380726ed4"
	oldb773Alms             = "efd46a6e-a901-4224-8a29-d53a4dc7750c"
	oldb773MasterOfArms     = "37e90cab-d904-4387-9d18-62f125c38c46"
	oldb773ChimericSphere   = "b4aae26a-44a4-4ef0-9a1d-14a07a0865a3"
	oldb773LlanowarDruid    = "b8cadeee-36b7-4a61-91ff-7d6c7c780464"
)

func TestOldB773CardsAreFullWithNoCaveats(t *testing.T) {
	for _, oracle := range []string{oldb773ZombieScavengers, oldb773ErtaisFamiliar, oldb773BosiumStrip, oldb773OrcishSettlers,
		oldb773XanthicStatue, oldb773Alms, oldb773MasterOfArms, oldb773ChimericSphere, oldb773LlanowarDruid} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %v with caveats %v, want full with none", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldb773Main(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	return g, me, opp
}

func oldb773Activate(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID, row int, params game.ActivateAbilityParams) error {
	t.Helper()
	params.Strict = true
	err := g.ActivateCatalogAbility(me.ID, id, row, params)
	if err == nil {
		effectstest.PassPriorityAroundTable(t, g)
	}
	return err
}

func oldb773Graveyard(p *game.Player, typeLines ...string) []uuid.UUID {
	var ids []uuid.UUID
	for _, tl := range typeLines {
		id := uuid.New()
		p.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Card " + tl, TypeLine: tl, Owner: p.ID, Controller: p.ID})
		ids = append(ids, id)
	}
	return ids
}

func TestOldB773ZombieScavengersExilesOnlyTheTopCreatureCard(t *testing.T) {
	g, me, _ := oldb773Main(t)
	zombie := effectstest.B12Push(g, me.ID, "Zombie Scavengers", "Creature — Zombie", oldb773ZombieScavengers, 3, 1)
	ids := oldb773Graveyard(me, "Creature — Bear", "Creature — Elf", "Sorcery")
	low, top, sorcery := ids[0], ids[1], ids[2]
	if err := oldb773Activate(t, g, me, zombie, 0, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{low}}); err == nil {
		t.Fatal("a creature card below the top creature card paid the cost")
	}
	if err := oldb773Activate(t, g, me, zombie, 0, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{sorcery}}); err == nil {
		t.Fatal("a sorcery paid a creature clause")
	}
	if err := oldb773Activate(t, g, me, zombie, 0, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{top}}); err != nil {
		t.Fatal(err)
	}
	if !g.Exile.Contains(top) || !me.Graveyard.Contains(low) {
		t.Fatal("the top creature card was not the one exiled")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(zombie) })
	if !g.Battlefield.Contains(zombie) || !effectstest.IsTapped(g, zombie) {
		t.Error("the regeneration shield did not save and tap the Zombie")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(zombie) })
	if g.Battlefield.Contains(zombie) {
		t.Error("one shield saved the Zombie twice")
	}
}

func TestOldB773ZombieScavengersNeedsACreatureCardInTheGraveyard(t *testing.T) {
	g, me, _ := oldb773Main(t)
	zombie := effectstest.B12Push(g, me.ID, "Zombie Scavengers", "Creature — Zombie", oldb773ZombieScavengers, 3, 1)
	ids := oldb773Graveyard(me, "Sorcery", "Instant")
	if err := oldb773Activate(t, g, me, zombie, 0, game.ActivateAbilityParams{ExileIDs: ids[:1]}); err == nil {
		t.Fatal("paid with no creature card in the graveyard")
	}
}

func TestOldB773ErtaisFamiliarMillsWhenItPhasesOutOrLeaves(t *testing.T) {
	g, me, _ := oldb773Main(t)
	for i := 0; i < 8; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	fam := effectstest.B12Push(g, me.ID, "Ertai's Familiar", "Creature — Illusion", oldb773ErtaisFamiliar, 2, 2)
	lib := len(me.Library.Cards)
	g.WithWriteLock(func() { _ = g.PhaseOutForEffect(uuid.Nil, fam) })
	effectstest.PassPriorityAroundTable(t, g)
	if !g.IsPhasedOutForEffect(fam) {
		t.Fatal("the Familiar did not phase out")
	}
	if got := lib - len(me.Library.Cards); got != 3 {
		t.Fatalf("phasing out milled %d cards, want 3", got)
	}
	g.WithWriteLock(func() { _ = g.PhaseInForEffect(fam) })
	effectstest.PassPriorityAroundTable(t, g)
	lib = len(me.Library.Cards)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(fam) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := lib - len(me.Library.Cards); got != 3 {
		t.Fatalf("leaving the battlefield milled %d cards, want 3", got)
	}
}

func TestOldB773ErtaisFamiliarCantPhaseOutAfterItsAbility(t *testing.T) {
	g, me, _ := oldb773Main(t)
	for i := 0; i < 8; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	fam := effectstest.B12Push(g, me.ID, "Ertai's Familiar", "Creature — Illusion", oldb773ErtaisFamiliar, 2, 2)
	effectstest.FillPoolColored(me, "U", 1)
	if err := oldb773Activate(t, g, me, fam, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	if !g.CantPhaseOutForEffect(fam) {
		t.Fatal("the Familiar can still phase out")
	}
	lib := len(me.Library.Cards)
	g.WithWriteLock(func() { _ = g.PhaseOutForEffect(uuid.Nil, fam) })
	effectstest.PassPriorityAroundTable(t, g)
	if g.IsPhasedOutForEffect(fam) || len(me.Library.Cards) != lib {
		t.Error("a Familiar that can't phase out phased out (or milled)")
	}
	// The effect lasts until the controller's next upkeep.
	effectstest.AdvanceTo(t, g, game.StepEnd)
	if !g.CantPhaseOutForEffect(fam) {
		t.Error("the effect ended before the controller's next upkeep")
	}
}

func TestOldB773MasterOfArmsTapsOnlyABlockerOfItself(t *testing.T) {
	g, me, opp := oldb773Main(t)
	master := effectstest.B12Push(g, me.ID, "Master of Arms", "Creature — Human Soldier", oldb773MasterOfArms, 2, 2)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 5)
	idle := effectstest.B12Creature(g, opp.ID, "Idle Wall", "Creature — Wall", 0, 5)
	effectstest.DeclareAttack(t, g, opp.ID, master)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, master); err != nil {
		t.Fatal(err)
	}
	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.FillPoolColored(me, "W", 1)
	effectstest.FillPool(me, 1)
	if err := oldb773Activate(t, g, me, master, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: idle}}}); err == nil {
		t.Fatal("tapped a creature that is not blocking the Master")
	}
	if err := oldb773Activate(t, g, me, master, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: wall}}}); err != nil {
		t.Fatal(err)
	}
	if !effectstest.IsTapped(g, wall) {
		t.Error("the blocker did not tap")
	}
}

func TestOldB773LlanowarDruidUntapsEveryForestWhoeverControlsIt(t *testing.T) {
	g, me, opp := oldb773Main(t)
	druid := effectstest.B12Push(g, me.ID, "Llanowar Druid", "Creature — Elf Druid", oldb773LlanowarDruid, 1, 2)
	mine := effectstest.B12Push(g, me.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	theirs := effectstest.B12Push(g, opp.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	island := effectstest.B12Push(g, me.ID, "Island", "Basic Land — Island", "", 0, 0)
	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{mine, theirs, island} {
			effectstest.FindBattlefieldCardForTest(g, id).Tapped = true
		}
	})
	if err := oldb773Activate(t, g, me, druid, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	if g.Battlefield.Contains(druid) {
		t.Error("the Druid was not sacrificed")
	}
	if effectstest.IsTapped(g, mine) || effectstest.IsTapped(g, theirs) {
		t.Error("a Forest stayed tapped")
	}
	if !effectstest.IsTapped(g, island) {
		t.Error("an Island untapped")
	}
}

func TestOldB773OrcishSettlersDestroysExactlyXLands(t *testing.T) {
	g, me, opp := oldb773Main(t)
	settlers := effectstest.B12Push(g, me.ID, "Orcish Settlers", "Creature — Orc", oldb773OrcishSettlers, 1, 1)
	a := effectstest.B12Push(g, opp.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	b := effectstest.B12Push(g, opp.ID, "Island", "Basic Land — Island", "", 0, 0)
	c := effectstest.B12Push(g, opp.ID, "Swamp", "Basic Land — Swamp", "", 0, 0)
	bear := effectstest.PushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	land := func(ids ...uuid.UUID) []game.TargetRef {
		var refs []game.TargetRef
		for _, id := range ids {
			refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: id})
		}
		return refs
	}
	effectstest.FillPoolColored(me, "R", 1)
	effectstest.FillPool(me, 4)
	if err := oldb773Activate(t, g, me, settlers, 0, game.ActivateAbilityParams{XValue: 2, Targets: land(a)}); err == nil {
		t.Fatal("one target accepted for X = 2")
	}
	if err := oldb773Activate(t, g, me, settlers, 0, game.ActivateAbilityParams{XValue: 2, Targets: land(a, bear)}); err == nil {
		t.Fatal("a creature accepted as a land target")
	}
	if err := oldb773Activate(t, g, me, settlers, 0, game.ActivateAbilityParams{XValue: 2, Targets: land(a, b)}); err != nil {
		t.Fatal(err)
	}
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) || !g.Battlefield.Contains(c) {
		t.Error("the wrong lands were destroyed")
	}
	if g.Battlefield.Contains(settlers) {
		t.Error("the Settlers were not sacrificed")
	}
}

func TestOldB773XanthicStatueBecomesAnEightEightGolemWithTrample(t *testing.T) {
	g, me, _ := oldb773Main(t)
	statue := effectstest.PushCatalogPermanent(g, me.ID, "Xanthic Statue", "Artifact", oldb773XanthicStatue, false)
	effectstest.FillPool(me, 5)
	if err := oldb773Activate(t, g, me, statue, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	c := effectstest.EffectiveCardForTest(t, g, statue)
	if !c.IsCreature() || !c.IsArtifact() || !c.HasSubtype("Golem") {
		t.Errorf("creature %v artifact %v Golem %v", c.IsCreature(), c.IsArtifact(), c.HasSubtype("Golem"))
	}
	if p, tough := effectstest.EffectivePower(t, g, statue), effectstest.EffectiveToughness(t, g, statue); p != 8 || tough != 8 {
		t.Errorf("Statue is %d/%d, want 8/8", p, tough)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, statue, "trample") {
		t.Error("the Statue lacks trample")
	}
	g.WithWriteLock(func() { g.ClearEndOfTurnScopedStaticsLocked() })
	if effectstest.EffectiveCardForTest(t, g, statue).IsCreature() {
		t.Error("the Statue stayed a creature past end of turn")
	}
}

func TestOldB773ChimericSphereLaterActivationWins(t *testing.T) {
	g, me, _ := oldb773Main(t)
	sphere := effectstest.PushCatalogPermanent(g, me.ID, "Chimeric Sphere", "Artifact", oldb773ChimericSphere, false)
	effectstest.FillPool(me, 6)
	if err := oldb773Activate(t, g, me, sphere, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	if p, tough := effectstest.EffectivePower(t, g, sphere), effectstest.EffectiveToughness(t, g, sphere); p != 2 || tough != 1 {
		t.Errorf("first form is %d/%d, want 2/1", p, tough)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, sphere, "flying") || !effectstest.EffectiveCardForTest(t, g, sphere).HasSubtype("Construct") {
		t.Error("the 2/1 form lacks flying or the Construct type")
	}
	if err := oldb773Activate(t, g, me, sphere, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	if p, tough := effectstest.EffectivePower(t, g, sphere), effectstest.EffectiveToughness(t, g, sphere); p != 3 || tough != 2 {
		t.Errorf("second form is %d/%d, want 3/2", p, tough)
	}
	if effectstest.EffectiveAbilitiesContain(t, g, sphere, "flying") {
		t.Error("the 3/2 form kept flying")
	}
	if err := oldb773Activate(t, g, me, sphere, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	if p, tough := effectstest.EffectivePower(t, g, sphere), effectstest.EffectiveToughness(t, g, sphere); p != 2 || tough != 1 {
		t.Errorf("re-activated first form is %d/%d, want 2/1", p, tough)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, sphere, "flying") {
		t.Error("the later flying grant did not give flying back")
	}
}

func TestOldB773AlmsPreventsOnePointAndExilesTheTopCard(t *testing.T) {
	g, me, _ := oldb773Main(t)
	alms := effectstest.PushCatalogPermanent(g, me.ID, "Alms", "Enchantment", oldb773Alms, false)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 3, 5)
	ids := oldb773Graveyard(me, "Creature — Bear", "Sorcery")
	effectstest.FillPool(me, 1)
	target := []game.TargetRef{{Kind: game.TargetCard, ID: bear}}
	if err := oldb773Activate(t, g, me, alms, 0, game.ActivateAbilityParams{Targets: target, ExileIDs: ids[:1]}); err == nil {
		t.Fatal("a card below the top paid the cost")
	}
	if err := oldb773Activate(t, g, me, alms, 0, game.ActivateAbilityParams{Targets: target, ExileIDs: ids[1:]}); err != nil {
		t.Fatal(err)
	}
	if !g.Exile.Contains(ids[1]) || !me.Graveyard.Contains(ids[0]) {
		t.Error("the top card was not the one exiled")
	}
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.Nil, bear, 3) })
	var dmg int
	g.ReadSnapshot(func() { dmg = effectstest.FindBattlefieldCardForTest(g, bear).DamageMarked })
	if dmg != 2 {
		t.Errorf("the Bear has %d damage, want 2 (one prevented)", dmg)
	}
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.Nil, bear, 1) })
	g.ReadSnapshot(func() { dmg = effectstest.FindBattlefieldCardForTest(g, bear).DamageMarked })
	if dmg != 3 {
		t.Errorf("the shield prevented a second point: damage %d, want 3", dmg)
	}
}

func TestOldB773BosiumStripCastsOnlyFromTheTopAndExiles(t *testing.T) {
	g, me, opp := oldb773Main(t)
	strip := effectstest.PushCatalogPermanent(g, me.ID, "Bösium Strip", "Artifact", oldb773BosiumStrip, false)
	bolt := func() uuid.UUID {
		id := uuid.New()
		me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Lightning Bolt", TypeLine: "Instant", OracleID: effectstest.LightningBoltOracle,
			ManaCost: "{R}", Owner: me.ID, Controller: me.ID})
		return id
	}
	lower := bolt()
	creature := oldb773Graveyard(me, "Creature — Bear")[0]
	upper := bolt()
	cast := func(id uuid.UUID) error {
		return g.CastSpell(me.ID, id, game.CastSpellParams{
			Strict: true, FromZone: string(game.ZoneGraveyard), AlternativeCost: game.AltCostKeyGrantedCost,
			Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
		})
	}
	effectstest.FillPoolColored(me, "R", 3)
	if err := cast(upper); err == nil {
		t.Fatal("cast from the graveyard before the ability was activated")
	}
	effectstest.FillPool(me, 3)
	if err := oldb773Activate(t, g, me, strip, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	if err := cast(lower); err == nil {
		t.Fatal("a Bolt below the top of the graveyard was cast")
	}
	if err := cast(upper); err != nil {
		t.Fatalf("the top Bolt was not castable: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(upper) || me.Graveyard.Contains(upper) {
		t.Error("the spell cast from the graveyard did not go to exile")
	}
	if err := cast(lower); err == nil {
		t.Fatal("a Bolt under a creature card was cast")
	}
	if _, err := me.Graveyard.Remove(creature); err != nil {
		t.Fatal(err)
	}
	if err := cast(lower); err != nil {
		t.Fatalf("a Bolt that became the top card was not castable: %v", err)
	}
}
