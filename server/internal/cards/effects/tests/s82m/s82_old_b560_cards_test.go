package testss82m

import (
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b560_cards_test.go — Coldsnap pump and combat tricks, S82 OLD batch
// 560, played through the real catalog entries.

const (
	oldB560Stromgald   = "059ed13a-857c-420a-a127-5f720010b9fd"
	oldB560Centaur     = "2b548f9a-c160-427d-9bb4-1fe940f8ab28"
	oldB560Disciple    = "3a09c951-9861-4e95-acf1-5ce1a3778493"
	oldB560Lovisa      = "720879dc-7593-41c0-ab3d-936069f26930"
	oldB560Spitter     = "76ed34d2-e09e-4b53-98bd-d96b367fd315"
	oldB560Faerie      = "7a25ccf3-0fe1-447c-bdf4-57ec31b19a0f"
	oldB560Outrider    = "7b8559a7-e11f-430d-8a7a-785a7005c4c4"
	oldB560SoundCall   = "8f0ab2fb-297a-4824-8d2c-0120eec09e5e"
	oldB560WhiteShield = "96ef0bd9-4abb-49af-8744-0b2ad595a393"
	oldB560VoidMaw     = "ad2a1b07-58f5-44c2-92e0-464ea2d45e7e"
	oldB560StoneSpirit = "b36c66d3-24ae-4291-8f8b-8716e354975e"
	oldB560Rimehorn    = "b760b1db-764b-4b40-8a84-691b5c9c1dbc"
	oldB560BullAurochs = "edd5b228-f19f-4b27-8b92-75545389da14"
	oldB560Snowcrusher = "f20aa69a-e2eb-4103-9684-12390ef07c6c"
)

func TestOldB560CardsAreFullWithNoCaveats(t *testing.T) {
	for id, name := range map[string]string{
		oldB560Stromgald: "Stromgald Crusader", oldB560Centaur: "Boreal Centaur",
		oldB560Disciple: "Disciple of Tevesh Szat", oldB560Lovisa: "Lovisa Coldeyes",
		oldB560Spitter: "Steam Spitter", oldB560Faerie: "Diamond Faerie",
		oldB560Outrider: "Kjeldoran Outrider", oldB560SoundCall: "Sound the Call",
		oldB560WhiteShield: "White Shield Crusader", oldB560VoidMaw: "Void Maw",
		oldB560StoneSpirit: "Greater Stone Spirit", oldB560Rimehorn: "Rimehorn Aurochs",
		oldB560BullAurochs: "Bull Aurochs", oldB560Snowcrusher: "Phyrexian Snowcrusher",
	} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v as %q (%v, caveats %v)", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB560Snow(color string) game.ManaToken {
	return game.ManaToken{Color: color, Source: uuid.New(), SourceKinds: game.ManaSourceLand | game.ManaSourceSnow}
}

func oldB560Pump(t *testing.T, g *game.Game, owner uuid.UUID, name, typeLine, oracle string, p, tough int) uuid.UUID {
	t.Helper()
	id := effectstest.B12Push(g, owner, name, typeLine, oracle, p, tough)
	effectstest.AdvanceToMain(t, g)
	return id
}

func TestOldB560CrusadersFlyAndPumpWithTheirOwnColour(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, typeLine, symbol, protection string
	}{
		{"Stromgald Crusader", oldB560Stromgald, "Creature — Zombie Knight", "B", "protection from white"},
		{"White Shield Crusader", oldB560WhiteShield, "Creature — Human Knight", "W", "protection from black"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			c := oldB560Pump(t, g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 1)
			if !effectstest.HasEffectiveKeyword(t, g, c, tc.protection) {
				t.Errorf("no %s", tc.protection)
			}
			effectstest.B06AddMana(me, tc.symbol)
			effectstest.B16Activate(t, g, me.ID, c, 0, game.ActivateAbilityParams{})
			effectstest.PassPriorityAroundTable(t, g)
			if !effectstest.HasEffectiveKeyword(t, g, c, "flying") {
				t.Error("the first ability did not grant flying")
			}
			effectstest.B06AddMana(me, tc.symbol, tc.symbol)
			effectstest.B16Activate(t, g, me.ID, c, 1, game.ActivateAbilityParams{})
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.EffectivePower(t, g, c); got != 3 {
				t.Errorf("power = %d, want 3", got)
			}
		})
	}
}

func TestOldB560BorealCentaurNeedsSnowManaOnceEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	c := oldB560Pump(t, g, me.ID, "Boreal Centaur", "Snow Creature — Centaur Warrior", oldB560Centaur, 2, 2)
	effectstest.B06AddMana(me, "G")
	if err := g.ActivateCatalogAbility(me.ID, c, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("ordinary mana paid {S}")
	}
	me.ManaPool.AddMana(oldB560Snow("G"), oldB560Snow("G"))
	effectstest.B16Activate(t, g, me.ID, c, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePower(t, g, c), effectstest.EffectiveToughness(t, g, c); p != 3 || tough != 3 {
		t.Errorf("Centaur %d/%d, want 3/3", p, tough)
	}
	if err := g.ActivateCatalogAbility(me.ID, c, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated twice in one turn")
	}
}

func TestOldB560DiscipleShrinksAndThenSacrificesForSix(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	d := effectstest.B21Push(g, me.ID, "Disciple of Tevesh Szat", "Creature — Human Cleric", oldB560Disciple, 3, 1, "B")
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 2, 1)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Bear", 5, 6)
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, d, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(small)})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("-1/-1 did not kill the X/1")
	}
	effectstest.UntapForTest(g, d)
	effectstest.B06AddMana(me, "B", "B", "B", "B", "B", "B")
	effectstest.B16Activate(t, g, me.ID, d, 1, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(big)})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(d) {
		t.Error("the Disciple was not sacrificed")
	}
	if g.Battlefield.Contains(big) {
		t.Error("-6/-6 did not kill the 5/6")
	}
}

func TestOldB560LovisaPumpsBarbariansWarriorsAndBerserkersOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lovisa := effectstest.B12Push(g, me.ID, "Lovisa Coldeyes", "Legendary Creature — Human", oldB560Lovisa, 3, 3)
	war := effectstest.B12Creature(g, opp.ID, "Warrior", "Creature — Human Warrior", 2, 2)
	bar := effectstest.B12Creature(g, me.ID, "Barbarian", "Creature — Human Barbarian", 2, 2)
	ber := effectstest.B12Creature(g, me.ID, "Berserker", "Creature — Human Berserker", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	for _, id := range []uuid.UUID{war, bar, ber} {
		if p, tough := effectstest.EffectivePower(t, g, id), effectstest.EffectiveToughness(t, g, id); p != 4 || tough != 4 {
			t.Errorf("a Barbarian, Warrior or Berserker is %d/%d, want 4/4", p, tough)
		}
		if !effectstest.HasEffectiveKeyword(t, g, id, "haste") {
			t.Error("no haste")
		}
	}
	if effectstest.EffectivePower(t, g, bear) != 2 || effectstest.HasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("a Bear was pumped")
	}
	if effectstest.EffectivePower(t, g, lovisa) != 3 {
		t.Error("Lovisa (a Human) was pumped")
	}
}

func TestOldB560SpitterOutriderAndSnowcrusherRows(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	spit := effectstest.B12Push(g, me.ID, "Steam Spitter", "Creature — Spider", oldB560Spitter, 1, 5)
	out := effectstest.B12Push(g, me.ID, "Kjeldoran Outrider", "Creature — Human Soldier", oldB560Outrider, 2, 2)
	crush := effectstest.B12Push(g, me.ID, "Phyrexian Snowcrusher", "Snow Artifact Creature — Phyrexian Juggernaut", oldB560Snowcrusher, 6, 5)
	effectstest.AdvanceToMain(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, spit, "reach") {
		t.Error("Steam Spitter has no reach")
	}
	effectstest.B06AddMana(me, "R", "W")
	effectstest.B16Activate(t, g, me.ID, spit, 0, game.ActivateAbilityParams{})
	effectstest.B16Activate(t, g, me.ID, out, 0, game.ActivateAbilityParams{})
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, oldB560Snow("G"))
	effectstest.B16Activate(t, g, me.ID, crush, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, spit); got != 2 {
		t.Errorf("Steam Spitter power = %d, want 2", got)
	}
	if got := effectstest.EffectiveToughness(t, g, out); got != 3 {
		t.Errorf("Outrider toughness = %d, want 3", got)
	}
	if got := effectstest.EffectivePower(t, g, crush); got != 7 {
		t.Errorf("Snowcrusher power = %d, want 7", got)
	}
}

func TestOldB560SnowcrusherAttacksEachCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.TypedCreature(g, me.ID, "Phyrexian Snowcrusher", "Snow Artifact Creature — Phyrexian Juggernaut", oldB560Snowcrusher, 6, 5, "haste")
	if !effectstest.FaMustAttack(t, g) {
		t.Error("the Snowcrusher is not required to attack")
	}
}

func TestOldB560DiamondFaeriePumpsOnlySnowCreaturesYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	faerie := effectstest.B12Push(g, me.ID, "Diamond Faerie", "Snow Creature — Faerie", oldB560Faerie, 3, 3)
	snow := effectstest.B12Creature(g, me.ID, "Snow Bear", "Snow Creature — Bear", 2, 2)
	plain := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Snow Bear", "Snow Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, oldB560Snow("U"))
	effectstest.B16Activate(t, g, me.ID, faerie, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{faerie, snow} {
		if got := effectstest.EffectivePower(t, g, id); got < 3 {
			t.Errorf("a snow creature you control was not pumped (power %d)", got)
		}
	}
	if effectstest.EffectivePower(t, g, plain) != 2 || effectstest.EffectivePower(t, g, theirs) != 2 {
		t.Error("a nonsnow or opposing creature was pumped")
	}
}

func TestOldB560SoundTheCallsWolfGrowsWithCopiesInGraveyards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Sound the Call", "Sorcery", oldB560SoundCall, nil)
	effectstest.PassPriorityAroundTable(t, g)
	var wolf uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Wolf" && c.IsToken() {
			wolf = c.InstanceID
		}
	}
	if wolf == uuid.Nil {
		t.Fatal("no Wolf token")
	}
	// The spell itself is now in its owner's graveyard: one card named Sound the Call.
	if p, tough := effectstest.EffectivePower(t, g, wolf), effectstest.EffectiveToughness(t, g, wolf); p != 2 || tough != 2 {
		t.Errorf("Wolf %d/%d with one copy in a graveyard, want 2/2", p, tough)
	}
	// A second copy resolves into its owner's graveyard: now two cards named
	// Sound the Call lie in graveyards, and the first Wolf is 3/3.
	effectstest.CastCatalogSpell(t, g, "Sound the Call", "Sorcery", oldB560SoundCall, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePower(t, g, wolf), effectstest.EffectiveToughness(t, g, wolf); p != 3 || tough != 3 {
		t.Errorf("Wolf %d/%d with two copies in graveyards, want 3/3", p, tough)
	}
}

func TestOldB560VoidMawExilesOtherCreaturesAndEatsThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	maw := effectstest.B12Push(g, me.ID, "Void Maw", "Creature — Horror", oldB560VoidMaw, 4, 5)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.AvB18Destroy(t, g, theirs)
	effectstest.AvB18Destroy(t, g, mine)
	if opp.Graveyard.Contains(theirs) || me.Graveyard.Contains(mine) {
		t.Fatal("another creature went to a graveyard")
	}
	if z, _ := effectstest.CardWhere(g, opp, theirs); z != "exile" {
		t.Errorf("their creature is in %s, want exile", z)
	}
	if z, _ := effectstest.CardWhere(g, me, mine); z != "exile" {
		t.Errorf("my creature is in %s, want exile", z)
	}
	other := uuid.New()
	g.WithWriteLock(func() {
		g.Exile.PushTop(game.Card{InstanceID: other, Name: "Unlinked", TypeLine: "Creature — Bear", Owner: opp.ID, Controller: opp.ID})
	})
	if err := g.ActivateCatalogAbility(me.ID, maw, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{other}}); err == nil {
		t.Fatal("a card not exiled with the Maw paid the cost")
	}
	effectstest.B16Activate(t, g, me.ID, maw, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{theirs}})
	if !opp.Graveyard.Contains(theirs) {
		t.Error("the paid card did not go to its owner's graveyard")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePower(t, g, maw), effectstest.EffectiveToughness(t, g, maw); p != 6 || tough != 7 {
		t.Errorf("Maw %d/%d, want 6/7", p, tough)
	}
	// The Maw itself dies normally.
	effectstest.AvB18Destroy(t, g, maw)
	if !me.Graveyard.Contains(maw) {
		t.Error("the Maw was exiled instead of dying")
	}
}

func TestOldB560GreaterStoneSpiritGrantsToughnessAndAPumpRow(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	spirit := effectstest.B12Push(g, me.ID, "Greater Stone Spirit", "Creature — Elemental Spirit", oldB560StoneSpirit, 4, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "R", "R", "R", "R")
	effectstest.B16Activate(t, g, me.ID, spirit, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectiveToughness(t, g, bear); got != 4 {
		t.Errorf("toughness = %d, want 4", got)
	}
	idx, ref := effectstest.GrantedActivatedIndex(t, g, bear)
	if idx < 0 {
		t.Fatal("the target has no granted pump row")
	}
	effectstest.B16Activate(t, g, me.ID, bear, idx, game.ActivateAbilityParams{Ref: ref})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, bear); got != 3 {
		t.Errorf("power after the granted {R} pump = %d, want 3", got)
	}
	effectstest.AdvancePastCleanupForTest(t, g)
	if idx, _ := effectstest.GrantedActivatedIndex(t, g, bear); idx >= 0 {
		t.Error("the granted row outlived the turn")
	}
}

func TestOldB560AurochsGetPlusOnePerOtherAttackingAurochs(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.Seats1684(g)
	bull := effectstest.TypedCreature(g, me.ID, "Bull Aurochs", "Creature — Aurochs", oldB560BullAurochs, 2, 1, "haste")
	rime := effectstest.TypedCreature(g, me.ID, "Rimehorn Aurochs", "Snow Creature — Aurochs", oldB560Rimehorn, 3, 3, "haste")
	effectstest.AttackAndTrigger(t, g, bull, rime)
	effectstest.AnswerTriggerOrderInOfferedOrder(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, bull); got != 3 {
		t.Errorf("Bull Aurochs power = %d, want 3", got)
	}
	if got := effectstest.EffectivePower(t, g, rime); got != 4 {
		t.Errorf("Rimehorn Aurochs power = %d, want 4", got)
	}
}

func TestOldB560RimehornForcesABlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Seats1684(g)
	rime := effectstest.TypedCreature(g, me.ID, "Rimehorn Aurochs", "Snow Creature — Aurochs", oldB560Rimehorn, 3, 3)
	attacker := effectstest.TypedCreature(g, me.ID, "Attacker", "Creature — Bear", "", 2, 2, "haste")
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, oldB560Snow("G"))
	effectstest.B16Activate(t, g, me.ID, rime, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(blocker, attacker)})
	effectstest.PassPriorityAroundTable(t, g)
	if recs := effectstest.BlocksAttackerRecords(g); len(recs) != 1 || recs[0].Objects[0].ID != attacker {
		t.Fatalf("records = %+v, want one naming the attacker", recs)
	}
	effectstest.AttackAndTrigger(t, g, attacker)
	effectstest.OnToBlockers(t, g)
	if owed := effectstest.MustBlockNow(g); owed[blocker] != attacker {
		t.Errorf("must_block = %v, want the chosen creature to owe a block of the attacker", owed)
	}
}

func TestOldB560GreaterStoneSpiritCantBeBlockedByFliers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Seats1684(g)
	spirit := effectstest.TypedCreature(g, me.ID, "Greater Stone Spirit", "Creature — Elemental Spirit", oldB560StoneSpirit, 4, 4, "haste")
	flier := effectstest.TypedCreature(g, opp.ID, "Flier", "Creature — Bird", "", 1, 1, "flying")
	ground := effectstest.B12Creature(g, opp.ID, "Ground", "Creature — Bear", 2, 2)
	effectstest.AttackAndTrigger(t, g, spirit)
	effectstest.OnToBlockers(t, g)
	if err := effectstest.ReqBlock(t, g, spirit, flier); err == nil {
		t.Error("a creature with flying blocked the Spirit")
	}
	if err := effectstest.ReqBlock(t, g, spirit, ground); err != nil {
		t.Errorf("a creature without flying could not block: %v", err)
	}
}
