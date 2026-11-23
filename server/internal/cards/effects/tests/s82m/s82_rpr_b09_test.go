package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_rpr_b09_test.go — S82 reprint pass batch 9 (#5908). Every test plays
// the real catalog card through the engine.

const (
	rprB09Shattered   = "e339c6cb-c75d-4dd4-8c1a-78f844983242"
	rprB09Sword       = "f1c4269a-1d99-4f3b-8107-a005df3468d5"
	rprB09Silverwing  = "ba968a1e-c946-4cbe-9727-160c73cde3f4"
	rprB09Slurrk      = "a0892be5-3272-431e-bbed-fc6520b259a7"
	rprB09Swordmaster = "11ad6e75-9bbd-40ea-b04a-e6a61db3bcda"
	rprB09Soul        = "add500e1-8eeb-45d7-a10d-2fae2cc69233"
	rprB09Swindle     = "81fa8e6a-5fb8-43de-802f-df5b985c6a7d"
	rprB09Struggle    = "0b1f2522-960c-405f-9622-4cadbc7ba545"
	rprB09Elenora     = "8f6a3dcf-32e0-48c7-9265-0e298df4a5e4"
	rprB09Assembly    = "007c110d-4364-41e5-a8bc-6ac05ecbbd00"
	rprB09Triskelion  = "74f67dcf-5afb-45aa-8d4b-3cdb23f6f2a1"
)

func TestRprB09EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		rprB09Shattered: "Shattered Angel", rprB09Sword: "Sigiled Sword of Valeron",
		rprB09Silverwing: "Silverwing Squadron", rprB09Slurrk: "Slurrk, All-Ingesting",
		rprB09Swordmaster: "Smitten Swordmaster", rprB09Swordmaster + "#1": "Curry Favor",
		rprB09Soul: "Soul of New Phyrexia", rprB09Swindle: "Spell Swindle",
		rprB09Struggle: "Struggle", rprB09Struggle + "#1": "Survive",
		rprB09Elenora: "Syr Elenora, the Discerning", rprB09Assembly: "Thopter Assembly",
		rprB09Triskelion: "Triskelion",
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

func TestRprB09TriskelionEntersWithThreeCountersAndShoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tri := effectstest.CastCreatureCard(t, g, "Triskelion", "Artifact Creature — Construct", rprB09Triskelion, 1, 1)
	if got := effectstest.CounterOn(g, tri, game.CounterPlusOne); got != 3 {
		t.Fatalf("counters = %d, want 3", got)
	}
	if p := effectstest.EffectivePowerWithCounters(t, g, tri); p != 4 {
		t.Errorf("power %d, want 4", p)
	}
	before := effectstest.LifeOf(g, opp.ID)
	for i := 1; i <= 3; i++ {
		if err := g.ActivateCatalogAbility(me.ID, tri, 0, game.ActivateAbilityParams{
			Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}, Strict: true,
		}); err != nil {
			t.Fatalf("activation %d: %v", i, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := effectstest.LifeOf(g, opp.ID); got != before-3 {
		t.Errorf("opponent life %d, want %d", got, before-3)
	}
	// A 1/1 with no counters left is still on the battlefield, and can't pay.
	if err := g.ActivateCatalogAbility(me.ID, tri, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}, Strict: true,
	}); err == nil {
		t.Error("activated with no counter to remove")
	}
}

func TestRprB09ShatteredAngelGainsOnlyForOpponentsLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B43Catalog(g, me.ID, "Shattered Angel", "Creature — Phyrexian Angel", rprB09Shattered, 3, 3)
	before := effectstest.LifeOf(g, me.ID)
	effectstest.B12PlayFromHand(t, g, "Forest", "Basic Land — Forest", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OwedMays(g, me.ID) != 0 || effectstest.LifeOf(g, me.ID) != before {
		t.Fatal("your own land triggered the Angel")
	}
	for _, yes := range []bool{false, true} {
		effectstest.AdvanceToNextSeatsTurn(t, g)
		effectstest.B12PlayFromHand(t, g, "Swamp", "Basic Land — Swamp", "", game.CastSpellParams{})
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.OwedMays(g, me.ID) == 0 {
			t.Fatal("an opponent's land must offer the life")
		}
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, yes)
		effectstest.PassPriorityAroundTable(t, g)
		if yes {
			before += 3
		}
		if got := effectstest.LifeOf(g, me.ID); got != before {
			t.Errorf("answer %v: life %d, want %d", yes, got, before)
		}
	}
}

func TestRprB09SyrElenoraPowerDrawAndTax(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	hand := effectstest.HandSize(me)
	el := effectstest.CastCreatureCard(t, g, "Syr Elenora, the Discerning", "Legendary Creature — Human Knight", rprB09Elenora, 0, 4)
	if effectstest.HandSize(me) != hand+1 {
		t.Errorf("hand %d, want %d (the draw)", effectstest.HandSize(me), hand+1)
	}
	if p := effectstest.EffectivePowerWithCounters(t, g, el); p != effectstest.HandSize(me) {
		t.Errorf("power %d, want hand size %d", p, effectstest.HandSize(me))
	}
	if err := g.DrawCard(me.ID); err != nil {
		t.Fatal(err)
	}
	if p := effectstest.EffectivePowerWithCounters(t, g, el); p != effectstest.HandSize(me) {
		t.Errorf("power %d after gaining a card, want %d", p, effectstest.HandSize(me))
	}

	// The tax: an opponent's Lightning Bolt at Elenora costs {2} more.
	other := effectstest.B12Creature(g, me.ID, "Bystander", "Creature — Bear", 2, 2)
	cast := func(target uuid.UUID, mana ...string) error {
		id := uuid.New()
		opp.Hand.PushTop(game.Card{InstanceID: id, Name: "Lightning Bolt", TypeLine: "Instant", OracleID: effectstest.LightningBoltOracle,
			ManaCost: "{R}", Owner: opp.ID, Controller: opp.ID})
		for _, c := range mana {
			opp.ManaPool.AddMana(game.ManaToken{Color: c})
		}
		return g.CastSpell(opp.ID, id, game.CastSpellParams{Targets: effectstest.B16TargetCard(target), Strict: true})
	}
	if err := cast(el, "R"); err == nil {
		t.Error("cast a spell at Syr Elenora for its printed cost")
	}
	opp.ManaPool = game.ManaPool{}
	if err := cast(el, "R", "C", "C"); err != nil {
		t.Errorf("could not pay {R}+{2} at Syr Elenora: %v", err)
	}
	opp.ManaPool = game.ManaPool{}
	if err := cast(other, "R"); err != nil {
		t.Errorf("a spell at another creature was taxed: %v", err)
	}
}

func TestRprB09SpellSwindleCountersAndMakesTreasuresEqualToManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spell := effectstest.Batch01OpponentCasts(t, g, opp, "Big Bolt", effectstest.LightningBoltOracle, "{2}{R}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	effectstest.CastCatalogSpell(t, g, "Spell Swindle", "Instant", rprB09Swindle, []game.TargetRef{{Kind: game.TargetCard, ID: spell}})
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(spell) {
		t.Fatal("the spell was not countered")
	}
	if n := effectstest.CountOnBattlefield(g, "Treasure", me.ID); n != 3 {
		t.Errorf("%d Treasures, want 3", n)
	}
}

func TestRprB09SilverwingSquadronSizeAndKnightsPerOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sq := effectstest.B43Catalog(g, me.ID, "Silverwing Squadron", "Creature — Human Knight", rprB09Silverwing, 0, 0)
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	g.RunStateChecksForTest()
	if p, tt := effectstest.EffectivePower(t, g, sq), effectstest.EffectiveToughness(t, g, sq); p != 2 || tt != 2 {
		t.Fatalf("size %d/%d, want 2/2 (two creatures you control)", p, tt)
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.AttackWith(t, g, opp.ID, sq)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountOnBattlefield(g, "Knight", me.ID); n != 3 {
		t.Errorf("%d Knight tokens, want 3 (three opponents)", n)
	}
}

func TestRprB09SigiledSwordGrantsAndMakesAnAttackingKnight(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sword := effectstest.B43Catalog(g, me.ID, "Sigiled Sword of Valeron", "Artifact — Equipment", rprB09Sword, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.EquipTo(t, g, me.ID, sword, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, bear); p != 4 {
		t.Errorf("power %d, want 4", p)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "vigilance") || !effectstest.AvB5HasSubtype(t, g, bear, "Knight") {
		t.Error("equipped creature lacks vigilance or the Knight type")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassAndSettleAttackSplit(t, g, me.ID)
	n := 0
	attacking := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Knight" && c.Controller == me.ID && effects.IsToken(c) {
			n++
			attacking = c.AttackingTarget == opp.ID
		}
	}
	if n != 1 || !attacking {
		t.Errorf("%d Knight tokens (attacking %v), want one attacking %v", n, attacking, opp.ID)
	}
}

func TestRprB09ThopterAssemblyReturnsOnlyWithNoOtherThopter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	asm := effectstest.B43Catalog(g, me.ID, "Thopter Assembly", "Artifact Creature — Thopter", rprB09Assembly, 5, 5)
	other := effectstest.B12Creature(g, me.ID, "Other Thopter", "Artifact Creature — Thopter", 1, 1)
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(asm) {
		t.Fatal("returned to hand with another Thopter around")
	}
	effectstest.KillCreature(t, g, me.ID, other)
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(asm) || !me.Hand.Contains(asm) {
		t.Error("did not return to hand")
	}
	if n := effectstest.CountOnBattlefield(g, "Thopter", me.ID); n != 5 {
		t.Errorf("%d Thopter tokens, want 5", n)
	}
}

func TestRprB09SlurrkGrowsEveryCounteredCreatureWhenOneDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	slurrk := effectstest.CastCreatureCard(t, g, "Slurrk, All-Ingesting", "Legendary Creature — Ooze", rprB09Slurrk, 0, 0)
	if got := effectstest.CounterOn(g, slurrk, game.CounterPlusOne); got != 5 {
		t.Fatalf("Slurrk entered with %d counters, want 5", got)
	}
	a := effectstest.B12Creature(g, me.ID, "Counter Bear", "Creature — Bear", 2, 2)
	bare := effectstest.B12Creature(g, me.ID, "Bare Bear", "Creature — Bear", 2, 2)
	dying := effectstest.B12Creature(g, me.ID, "Dying Bear", "Creature — Bear", 2, 2)
	effectstest.SetCounters(g, a, map[string]int{game.CounterPlusOne: 1})
	effectstest.SetCounters(g, dying, map[string]int{game.CounterPlusOne: 1})
	effectstest.KillCreature(t, g, me.ID, dying)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterOn(g, slurrk, game.CounterPlusOne) != 6 || effectstest.CounterOn(g, a, game.CounterPlusOne) != 2 {
		t.Errorf("Slurrk %d, bear %d; want 6 and 2", effectstest.CounterOn(g, slurrk, game.CounterPlusOne), effectstest.CounterOn(g, a, game.CounterPlusOne))
	}
	if effectstest.CounterOn(g, bare, game.CounterPlusOne) != 0 {
		t.Error("a creature with no counter got one")
	}
	// No counter on the dead creature: nothing happens.
	plain := effectstest.B12Creature(g, me.ID, "Plain", "Creature — Bear", 1, 1)
	effectstest.KillCreature(t, g, me.ID, plain)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterOn(g, slurrk, game.CounterPlusOne) != 6 {
		t.Error("a death with no counter still triggered Slurrk")
	}
	// Slurrk's own death counts (last-known counters), and a rival's creature doesn't.
	opp := g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.SetCounters(g, theirs, map[string]int{game.CounterPlusOne: 1})
	effectstest.KillCreature(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterOn(g, slurrk, game.CounterPlusOne) != 6 {
		t.Error("an opponent's creature dying triggered Slurrk")
	}
	effectstest.KillCreature(t, g, me.ID, slurrk)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterOn(g, a, game.CounterPlusOne) != 3 {
		t.Errorf("Slurrk's own death left the bear at %d, want 3", effectstest.CounterOn(g, a, game.CounterPlusOne))
	}
}

func TestRprB09CurryFavorCountsKnights(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Creature(g, me.ID, "Knight A", "Creature — Human Knight", 2, 2)
	effectstest.B12Creature(g, me.ID, "Knight B", "Creature — Knight", 2, 2)
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	lives := effectstest.LifeOfOpponents(g)
	mine := effectstest.LifeOf(g, me.ID)
	effectstest.CastCatalogSpell(t, g, "Curry Favor", "Sorcery — Adventure", rprB09Swordmaster+"#1", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, me.ID); got != mine+2 {
		t.Errorf("life %d, want %d", got, mine+2)
	}
	for i, l := range effectstest.LifeOfOpponents(g) {
		if l != lives[i]-2 {
			t.Errorf("opponent %d at %d, want %d", i, l, lives[i]-2)
		}
	}
}

func TestRprB09SoulOfNewPhyrexiaGivesIndestructibleFromBattlefieldAndGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	soul := effectstest.B43Catalog(g, me.ID, "Soul of New Phyrexia", "Artifact Creature — Phyrexian Avatar", rprB09Soul, 6, 6)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, g.Seats[1].ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	for i := 0; i < 5; i++ {
		me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	}
	if err := g.ActivateCatalogAbility(me.ID, soul, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "indestructible") || !effectstest.HasEffectiveKeyword(t, g, soul, "indestructible") {
		t.Error("your permanents did not gain indestructible")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "indestructible") {
		t.Error("an opponent's permanent gained indestructible")
	}
	// Not without the mana.
	if err := g.ActivateCatalogAbility(me.ID, soul, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated for no mana")
	}
}

func TestRprB09SoulOfNewPhyrexiaGraveyardAbilityExilesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Soul of New Phyrexia", TypeLine: "Artifact Creature — Phyrexian Avatar",
		OracleID: rprB09Soul, Owner: me.ID, Controller: me.ID})
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	for i := 0; i < 5; i++ {
		me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Contains(id) {
		t.Error("the card is still in the graveyard")
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "indestructible") {
		t.Error("the Bear did not gain indestructible")
	}
}

func TestRprB09StruggleDealsDamageEqualToLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 3; i++ {
		effectstest.B12Permanent(g, me.ID, "Mountain", "Basic Land — Mountain")
	}
	effectstest.B12Permanent(g, opp.ID, "Their Forest", "Basic Land — Forest")
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 2, 3)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Bear", 2, 4)
	effectstest.CastCatalogSpell(t, g, "Struggle", "Instant", rprB09Struggle, effectstest.B16TargetCard(small))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("3 damage should kill a 3-toughness creature")
	}
	effectstest.CastCatalogSpell(t, g, "Struggle", "Instant", rprB09Struggle, effectstest.B16TargetCard(big))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(big) {
		t.Error("3 damage killed a 4-toughness creature (opponent lands counted?)")
	}
}

func TestRprB09SurviveShufflesEveryGraveyardIntoItsLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.Batch01GraveyardCard(me, "Mine", "Creature — Bear")
	b := effectstest.Batch01GraveyardCard(opp, "Theirs", "Creature — Bear")
	libMe, libOpp := me.Library.Size(), opp.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Survive", "Sorcery", rprB09Struggle+"#1", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Contains(a) || opp.Graveyard.Contains(b) {
		t.Error("a graveyard card stayed put")
	}
	if me.Library.Size() != libMe+1 || opp.Library.Size() != libOpp+1 {
		t.Errorf("libraries %d/%d, want %d/%d", me.Library.Size(), opp.Library.Size(), libMe+1, libOpp+1)
	}
}
