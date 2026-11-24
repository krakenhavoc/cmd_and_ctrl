package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	oldB27TraceOID     = "530140f1-e992-4b5e-882c-c15bacc328fa"
	oldB27SuperStateID = "ab2dfe99-bd77-44d7-92be-81643ea901f0"
	oldB27BreathOID    = "fdaa0faf-e7ee-48b3-9330-de0fc3792ad3"
	oldB27SoulScarOID  = "009240f3-b7f8-4cbe-a3e9-974da66fb62c"
	oldB27AishaOID     = "eae3e762-dacd-4bd2-923c-3abb5ceb729a"
	oldB27JessicaOID   = "eb947d87-fa06-438c-93d2-7d681468b2f1"
	oldB27HymnOID      = "4b977117-741c-4157-8fcf-719e475a8a4c"
	oldB27FilthOID     = "aa77980d-0434-40f0-990f-1c8c50610174"
	oldB27CameoOID     = "d2be289e-e560-405d-9728-d8a4ee9cbf56"
	oldB27Bear         = "Creature — Bear"
	oldB27Aura         = "Enchantment — Aura"
)

func TestOldB27CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB27TraceOID: "Trace of Abundance", oldB27SuperStateID: "Super State",
		oldB27BreathOID: "Breath of Fury", oldB27SoulScarOID: "Soul-Scar Mage",
		oldB27AishaOID: "Aisha of Sparks and Smoke", oldB27JessicaOID: "Reverend Mother Jessica",
		oldB27HymnOID: "Battle Hymn", oldB27FilthOID: "Rain of Filth", oldB27CameoOID: "Tigereye Cameo",
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

func oldB27Push(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, p, tgh int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: p, Toughness: tgh, Owner: owner, Controller: owner,
	})
}

func oldB27Attach(t *testing.T, g *game.Game, owner uuid.UUID, name, oracle string, host uuid.UUID) uuid.UUID {
	t.Helper()
	id := oldB27Push(g, owner, name, oldB27Aura, oracle, 0, 0)
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(id, game.TargetRef{Kind: game.TargetCard, ID: host}); err != nil {
			t.Fatalf("AttachForEffect: %v", err)
		}
	})
	return id
}

func oldB27Pool(p *game.Player, colors ...string) bool {
	got := effectstest.SortedPool(p)
	if len(got) != len(colors) {
		return false
	}
	for i := range got {
		if got[i] != colors[i] {
			return false
		}
	}
	return true
}

// --- Trace of Abundance ---------------------------------------------------------

func TestOldB27TraceOfAbundanceGivesShroudAndAnAdditionalManaOfAnyColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	forest := effectstest.B31Push(g, me.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	oldB27Attach(t, g, me.ID, "Trace of Abundance", oldB27TraceOID, forest)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if !effectstest.HasEffectiveKeyword(t, g, forest, "shroud") {
		t.Error("the enchanted land has no shroud")
	}
	src := effectstest.BattlefieldCardPtr(t, g, oldB27Push(g, opp.ID, "Their Bear", oldB27Bear, "", 2, 2))
	tgt := effectstest.BattlefieldCardPtr(t, g, forest)
	if game.CanBeTargetedBy(tgt, game.ZoneBattlefield, game.SourceObject(opp.ID, src)) ||
		game.CanBeTargetedBy(tgt, game.ZoneBattlefield, game.SourceObject(me.ID, src)) {
		t.Error("a land with shroud was targetable")
	}
	effectstest.TapForMana(t, g, me.ID, forest)
	pick := effectstest.PendingOfKind(g, game.PendingChoiceMana)
	if pick == nil {
		t.Fatal("no mana pick for the additional mana of any color")
	}
	if len(pick.ColorOptions) != 5 {
		t.Errorf("options = %v, want all five colours", pick.ColorOptions)
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "U"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if !oldB27Pool(me, "G", "U") {
		t.Errorf("pool = %v, want {G} and the chosen {U}", effectstest.SortedPool(me))
	}
}

// --- Tigereye Cameo -------------------------------------------------------------

func TestOldB27TigereyeCameoTapsForGreenOrWhite(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cameo := effectstest.B31Push(g, me.ID, "Tigereye Cameo", "Artifact", oldB27CameoOID, "", 0, 0)
	effectstest.TapForMana(t, g, me.ID, cameo)
	pick := effectstest.PendingOfKind(g, game.PendingChoiceMana)
	if pick == nil {
		t.Fatal("no colour pick for {G} or {W}")
	}
	if len(pick.ColorOptions) != 2 {
		t.Errorf("options = %v, want green and white", pick.ColorOptions)
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "W"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if !oldB27Pool(me, "W") {
		t.Errorf("pool = %v, want {W}", effectstest.SortedPool(me))
	}
	if effectstest.TappedOnBattlefield(t, g, cameo) == false {
		t.Error("the Cameo did not tap")
	}
}

// --- Battle Hymn ----------------------------------------------------------------

func TestOldB27BattleHymnAddsRedForEachCreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	for i := 0; i < 3; i++ {
		oldB27Push(g, me.ID, "Bear", oldB27Bear, "", 2, 2)
	}
	oldB27Push(g, opp.ID, "Their Bear", oldB27Bear, "", 2, 2)
	oldB27Push(g, me.ID, "Rock", "Artifact", "", 0, 0)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Battle Hymn", "Instant", oldB27HymnOID, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !oldB27Pool(me, "R", "R", "R") {
		t.Errorf("pool = %v, want three {R}", effectstest.SortedPool(me))
	}
}

// --- Rain of Filth --------------------------------------------------------------

func TestOldB27RainOfFilthLetsLandsBeSacrificedForBlackUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	swamp := effectstest.B31Push(g, me.ID, "Swamp", "Basic Land — Swamp", "", "", 0, 0)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Rain of Filth", "Instant", oldB27FilthOID, nil)
	effectstest.PassPriorityAroundTable(t, g)
	before := effectstest.PoolSize(g, me.ID)
	if err := g.ActivateManaAbility(me.ID, swamp, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("the granted sacrifice ability: %v", err)
	}
	if g.Battlefield.Contains(swamp) {
		t.Error("the land was not sacrificed")
	}
	if got := effectstest.PoolSize(g, me.ID) - before; got != 1 {
		t.Errorf("added %d mana, want 1", got)
	}
	if !me.Graveyard.Contains(swamp) {
		t.Error("the sacrificed land is not in the graveyard")
	}
}

func TestOldB27RainOfFilthDoesNotReachALandThatEntersLater(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Rain of Filth", "Instant", oldB27FilthOID, nil)
	effectstest.PassPriorityAroundTable(t, g)
	later := effectstest.B31Push(g, me.ID, "Swamp", "Basic Land — Swamp", "", "", 0, 0)
	if err := g.ActivateManaAbility(me.ID, later, 1, game.ManaAbilityParams{}); err == nil {
		t.Error("a land that entered after the spell resolved gained the ability")
	}
}

// --- Soul-Scar Mage -------------------------------------------------------------

func oldB27Bolt(t *testing.T, g *game.Game, target game.TargetRef) {
	t.Helper()
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{target})
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB27SoulScarMageTurnsNoncombatDamageToAnOpposingCreatureIntoCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	oldB27Push(g, me.ID, "Soul-Scar Mage", "Creature — Human Wizard", oldB27SoulScarOID, 1, 2)
	theirs := oldB27Push(g, opp.ID, "Their Giant", "Creature — Giant", "", 5, 5)
	mine := oldB27Push(g, me.ID, "My Giant", "Creature — Giant", "", 5, 5)
	effectstest.AdvanceToMain(t, g)
	oppLife := effectstest.LifeOf(g, opp.ID)
	oldB27Bolt(t, g, game.TargetRef{Kind: game.TargetCard, ID: theirs})
	c, _ := effectstest.BattlefieldCard(g, theirs)
	if c.DamageMarked != 0 || c.Counters[game.CounterMinusOne] != 3 {
		t.Errorf("their Giant has %d damage and %d -1/-1 counters, want 0 and 3", c.DamageMarked, c.Counters[game.CounterMinusOne])
	}
	// Counters, not damage: a second Bolt takes it to six counters on a 5/5.
	oldB27Bolt(t, g, game.TargetRef{Kind: game.TargetCard, ID: theirs})
	if g.Battlefield.Contains(theirs) {
		t.Error("six -1/-1 counters on a 5/5 left it alive")
	}
	oldB27Bolt(t, g, game.TargetRef{Kind: game.TargetCard, ID: mine})
	m, _ := effectstest.BattlefieldCard(g, mine)
	if m.DamageMarked != 3 || m.Counters[game.CounterMinusOne] != 0 {
		t.Errorf("my own Giant has %d damage and %d counters, want damage only", m.DamageMarked, m.Counters[game.CounterMinusOne])
	}
	oldB27Bolt(t, g, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	if got := oppLife - effectstest.LifeOf(g, opp.ID); got != 3 {
		t.Errorf("damage to a player lost %d life, want 3", got)
	}
}

func TestOldB27SoulScarMageOnlyCoversSourcesItsControllerControls(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	oldB27Push(g, opp.ID, "Soul-Scar Mage", "Creature — Human Wizard", oldB27SoulScarOID, 1, 2)
	mine := oldB27Push(g, me.ID, "My Giant", "Creature — Giant", "", 5, 5)
	effectstest.AdvanceToMain(t, g)
	oldB27Bolt(t, g, game.TargetRef{Kind: game.TargetCard, ID: mine})
	m, _ := effectstest.BattlefieldCard(g, mine)
	if m.DamageMarked != 3 || m.Counters[game.CounterMinusOne] != 0 {
		t.Errorf("an opponent's Mage changed my own spell's damage: %d damage, %d counters", m.DamageMarked, m.Counters[game.CounterMinusOne])
	}
}

// --- Super State ----------------------------------------------------------------

func TestOldB27SuperStateMakesA99FlyingFirstStrikeTramplingHaster(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := oldB27Push(g, me.ID, "Bear", oldB27Bear, "", 2, 2)
	oldB27Attach(t, g, me.ID, "Super State", oldB27SuperStateID, bear)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if p, tg := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 9 || tg != 9 {
		t.Errorf("the creature is %d/%d, want 9/9", p, tg)
	}
	for _, kw := range []string{"flying", "first strike", "trample", "haste"} {
		if !effectstest.HasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("the creature lacks %s", kw)
		}
	}
}

func TestOldB27SuperStateSplashesCombatDamageToEachOtherOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	var opps []*game.Player
	for _, p := range g.Seats {
		if p.ID != me.ID {
			opps = append(opps, p)
		}
	}
	bear := oldB27Push(g, me.ID, "Bear", oldB27Bear, "", 2, 2)
	oldB27Attach(t, g, me.ID, "Super State", oldB27SuperStateID, bear)
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = effectstest.LifeOf(g, p.ID)
	}
	effectstest.DealCombatDamageToPlayer(g, bear, opps[0].ID, 4)
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range opps[1:] {
		if got := lives[p.ID] - effectstest.LifeOf(g, p.ID); got != 4 {
			t.Errorf("other opponent %d lost %d, want 4", i, got)
		}
	}
	if got := effectstest.LifeOf(g, me.ID); got != lives[me.ID] {
		t.Errorf("my life changed to %d", got)
	}
	// The splash is not combat damage: it triggers nothing again.
	if got := lives[opps[0].ID] - effectstest.LifeOf(g, opps[0].ID); got != 0 {
		t.Errorf("the player first hit lost %d to the splash; the fake event deals none itself", got)
	}
}

// --- Breath of Fury -------------------------------------------------------------

func TestOldB27BreathOfFurySacrificesMovesAndGivesAnotherCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := oldB27Push(g, me.ID, "Bear", oldB27Bear, "", 2, 2)
	other := oldB27Push(g, me.ID, "Other Bear", oldB27Bear, "", 2, 2)
	tapped := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Tapped Bear", TypeLine: oldB27Bear, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID, Tapped: true,
	})
	aura := oldB27Attach(t, g, me.ID, "Breath of Fury", oldB27BreathOID, bear)
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.DealCombatDamageToPlayer(g, bear, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID); c != nil {
		if err := g.ResolveOwnPermanents(c.ID, me.ID, []uuid.UUID{other}); err != nil {
			t.Fatalf("choosing the creature: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the enchanted creature was not sacrificed")
	}
	if !effectstest.AttachedTo(t, g, aura, other) {
		t.Error("the Aura was not attached to the other creature")
	}
	if effectstest.TappedOnBattlefield(t, g, tapped) {
		t.Error("your creatures were not untapped")
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	if next := effectstest.StepOnce(t, g); next.Step != game.StepBeginCombat {
		t.Errorf("after combat: %+v, want an additional combat", next)
	}
}

func TestOldB27BreathOfFuryWithNoOtherCreatureDoesNotUntapOrAddACombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := oldB27Push(g, me.ID, "Bear", oldB27Bear, "", 2, 2)
	oldB27Attach(t, g, me.ID, "Breath of Fury", oldB27BreathOID, bear)
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.DealCombatDamageToPlayer(g, bear, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the enchanted creature was not sacrificed")
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	if next := effectstest.StepOnce(t, g); next.Step == game.StepBeginCombat {
		t.Error("an additional combat was granted with no creature to attach to")
	}
}

// --- Aisha of Sparks and Smoke --------------------------------------------------

func TestOldB27AishaGainsFirstStrikeForHybridMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	aisha := oldB27Push(g, me.ID, "Aisha of Sparks and Smoke", "Legendary Creature — Human Warrior", oldB27AishaOID, 4, 2)
	effectstest.AdvanceToMain(t, g)
	if effectstest.HasEffectiveKeyword(t, g, aisha, "first strike") {
		t.Fatal("Aisha starts with first strike")
	}
	effectstest.B06AddMana(me, "W")
	effectstest.B16Activate(t, g, me.ID, aisha, 0, game.ActivateAbilityParams{})
	if !effectstest.HasEffectiveKeyword(t, g, aisha, "first strike") {
		t.Error("{R/W} paid with white did not give first strike")
	}
}

func TestOldB27AishaOffersASorceryWithManaValueUpToTheDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	aisha := oldB27Push(g, me.ID, "Aisha of Sparks and Smoke", "Legendary Creature — Human Warrior", oldB27AishaOID, 4, 2)
	effectstest.AdvanceToMain(t, g)
	for _, c := range []struct{ name, typeLine, cost string }{
		{"Big Sorcery", "Sorcery", "{3}{R}"}, {"Instant", "Instant", "{R}"}, {"Fits", "Sorcery", "{1}{R}"},
	} {
		me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: c.name, TypeLine: c.typeLine, ManaCost: c.cost, Owner: me.ID, Controller: me.ID})
	}
	var fits uuid.UUID
	for _, c := range me.Hand.Cards {
		if c.Name == "Fits" {
			fits = c.InstanceID
		}
	}
	effectstest.DealCombatDamageToPlayer(g, aisha, opp.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.KeywordCastPromptFor(g, me.ID)
	if prompt == nil {
		t.Fatal("no free-cast offer")
	}
	if effectstest.KeywordCastCard(prompt) != fits {
		t.Errorf("offered %s, want only the mana value 2 sorcery", effectstest.KeywordCastCard(prompt))
	}
	effectstest.AnswerMayCastPrompt(t, g, me.ID, true)
	if !g.Stack.Contains(fits) && !me.Graveyard.Contains(fits) {
		t.Error("the sorcery was not cast")
	}
}

func TestOldB27AishaOffersNothingWhenNoSorceryFits(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	aisha := oldB27Push(g, me.ID, "Aisha of Sparks and Smoke", "Legendary Creature — Human Warrior", oldB27AishaOID, 4, 2)
	effectstest.AdvanceToMain(t, g)
	me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Big Sorcery", TypeLine: "Sorcery", ManaCost: "{3}{R}", Owner: me.ID, Controller: me.ID})
	effectstest.DealCombatDamageToPlayer(g, aisha, opp.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.KeywordCastPromptFor(g, me.ID) != nil {
		t.Error("a sorcery with mana value above the damage was offered")
	}
}

// --- Reverend Mother Jessica ----------------------------------------------------

func TestOldB27JessicaTapsAndGoadsACreatureWithLesserPowerWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	first := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	jessica := oldB27Push(g, me.ID, "Reverend Mother Jessica", "Legendary Creature — Human Noble Advisor", oldB27JessicaOID, 2, 2)
	small := oldB27Push(g, first.ID, "Small", oldB27Bear, "", 1, 1)
	big := oldB27Push(g, first.ID, "Big", oldB27Bear, "", 3, 3)
	equal := oldB27Push(g, first.ID, "Equal", oldB27Bear, "", 2, 2)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if kws := effectstest.HasEffectiveKeyword(t, g, jessica, game.KeywordProwess); !kws {
		t.Error("Jessica has no prowess")
	}
	effectstest.DeclareAttack(t, g, first.ID, jessica)
	for i := 0; i < 4; i++ {
		p := effectstest.LatestPickTarget(g, me.ID)
		if p == nil {
			break
		}
		for _, bad := range []uuid.UUID{big, equal} {
			if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
				t.Fatalf("a creature without lesser power (%s) was accepted", bad)
			}
		}
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: small}); err != nil {
			t.Fatalf("picking the smaller creature: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TappedOnBattlefield(t, g, small) {
		t.Error("the target was not tapped")
	}
	c, _ := effectstest.BattlefieldCard(g, small)
	if !c.IsGoadedBy(me.ID) {
		t.Error("the target was not goaded")
	}
	if effectstest.TappedOnBattlefield(t, g, big) || effectstest.TappedOnBattlefield(t, g, equal) {
		t.Error("a creature that wasn't targeted was tapped")
	}
}

// --- Wheel of Sun and Moon ------------------------------------------------------

const oldB27WheelOID = "6b6e4ee2-52e6-452b-af15-c90eab5fc746"

func oldB27Wheel(t *testing.T, g *game.Game, victim uuid.UUID) {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Wheel of Sun and Moon", oldB27Aura, oldB27WheelOID, []game.TargetRef{{Kind: game.TargetPlayer, ID: victim}})
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB27WheelOfSunAndMoonPutsTheEnchantedPlayersCardsOnTheBottomOfTheirLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp, other := g.Seats[1], g.Seats[2]
	theirs := oldB27Push(g, opp.ID, "Their Bear", oldB27Bear, "", 2, 2)
	sacrificed := oldB27Push(g, opp.ID, "Their Other Bear", oldB27Bear, "", 2, 2)
	third := oldB27Push(g, other.ID, "Third Bear", oldB27Bear, "", 2, 2)
	oldB27Wheel(t, g, opp.ID)
	size := opp.Library.Size()
	effectstest.KillCreature(t, g, opp.ID, theirs)
	if opp.Graveyard.Contains(theirs) {
		t.Fatal("the card went to the graveyard")
	}
	if got := opp.Library.Size() - size; got != 1 {
		t.Fatalf("the library grew by %d, want 1", got)
	}
	if bottom := opp.Library.Cards[0]; bottom.InstanceID != theirs {
		t.Errorf("the bottom card is %s, want the creature that died", bottom.Name)
	}
	revealed := false
	for _, ev := range g.Events {
		if ev.Kind == game.EventRevealCards && ev.CardID == theirs {
			revealed = true
		}
	}
	if !revealed {
		t.Error("the card was not revealed")
	}
	// A sacrifice is a different landing path and is replaced too.
	g.WithWriteLock(func() {
		if err := g.SacrificePermanentForEffect(sacrificed); err != nil {
			t.Fatalf("SacrificePermanentForEffect: %v", err)
		}
	})
	if opp.Graveyard.Contains(sacrificed) || opp.Library.Cards[0].InstanceID != sacrificed {
		t.Error("the sacrificed card was not put on the bottom of the library")
	}
	// Another player's card is untouched.
	effectstest.KillCreature(t, g, other.ID, third)
	if !other.Graveyard.Contains(third) {
		t.Error("a card of a player who isn't enchanted left the graveyard path")
	}
}
