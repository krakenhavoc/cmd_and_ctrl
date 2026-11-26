package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b423_cards_test.go — Return to Ravnica, life and detain (S82 OLD batch
// 423, #5566). Each test plays the real catalog entry.

const (
	oldB423SearchWarrantOracle = "27744033-5cb6-4d26-ad76-f4b4f1c0466c"
	oldB423TabletOracle        = "71ce7034-1400-4e7c-9ef3-37baed669faf"
	oldB423AfflictionOracle    = "9e3f5d0b-103f-40e2-beb5-26b1fcb136f3"
	oldB423RixMaadiOracle      = "cb88ada9-d3bf-4f0a-817a-c41f0ea9cc78"
	oldB423CrocodileOracle     = "fef48e69-ab4d-4364-be41-7560440afc18"
	oldB423ArchonOracle        = "27ecb717-6251-4066-b57e-7070e9e44bbf"
	oldB423PrahvOracle         = "5912d4b7-d18d-4619-a96b-d2615200971f"
	oldB423InjunctionOracle    = "6f6bb2dd-c888-43bd-b9ef-8bd451ab669b"
	oldB423MartialLawOracle    = "dfcf9f38-eb26-4033-b7c0-45ed29417aa5"
	oldB423SpiritOracle        = "e67e3cf1-6da9-4b52-a945-47bf54f69ffa"
)

func TestOldB423EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB423SearchWarrantOracle: "Search Warrant",
		oldB423TabletOracle:        "Tablet of the Guilds",
		oldB423AfflictionOracle:    "Shrieking Affliction",
		oldB423RixMaadiOracle:      "Rix Maadi Guildmage",
		oldB423CrocodileOracle:     "Stonefare Crocodile",
		oldB423ArchonOracle:        "Archon of the Triumvirate",
		oldB423PrahvOracle:         "New Prahv Guildmage",
		oldB423InjunctionOracle:    "Inaction Injunction",
		oldB423MartialLawOracle:    "Martial Law",
		oldB423SpiritOracle:        "Soulsworn Spirit",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB423Hand(g *game.Game, p *game.Player, n int) {
	g.WithWriteLock(func() {
		p.Hand.Cards = nil
		for i := 0; i < n; i++ {
			p.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Card", TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
		}
	})
}

func TestOldB423SearchWarrantGainsLifeEqualToTheirHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	oldB423Hand(g, opp, 3)
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Search Warrant", "Sorcery", oldB423SearchWarrantOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
	for _, c := range opp.Hand.Cards {
		if !c.IsKnownTo(me.ID) {
			t.Error("a card in their hand was not revealed to me")
		}
	}
}

func TestOldB423SearchWarrantEmptyHandGainsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	oldB423Hand(g, opp, 0)
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Search Warrant", "Sorcery", oldB423SearchWarrantOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("life %d, want %d", me.Life, life)
	}
}

func oldB423TabletTable(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPool(me, 2)
	effectstest.CastCatalogSpell(t, g, "Tablet of the Guilds", "Artifact", oldB423TabletOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	first := effectstest.PendingOfKind(g, game.PendingChoiceColor)
	if first == nil {
		t.Fatal("no first color prompt")
	}
	effectstest.AnswerColor(t, g, me.ID, "W")
	second := effectstest.PendingOfKind(g, game.PendingChoiceColor)
	if second == nil {
		t.Fatal("no second color prompt")
	}
	for _, o := range second.ColorOptions {
		if o == "W" {
			t.Errorf("the second prompt offers the first color again: %v", second.ColorOptions)
		}
	}
	effectstest.AnswerColor(t, g, me.ID, "U")
	var tablet uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Tablet of the Guilds" {
			tablet = c.InstanceID
		}
	}
	if got := g.ChosenColorOf(tablet); got != "WU" {
		t.Fatalf("chosen colors = %q, want WU", got)
	}
	return g, me, tablet
}

func oldB423CastColored(t *testing.T, g *game.Game, me *game.Player, colors []string, cost string) {
	t.Helper()
	id := uuid.New()
	me.ManaPool = game.ManaPool{}
	for _, c := range colors {
		me.ManaPool.AddMana(game.ManaToken{Color: c})
	}
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Spell", TypeLine: "Sorcery", ManaCost: cost, Colors: colors, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB423TabletGainsOneLifePerChosenColorTheSpellIs(t *testing.T) {
	g, me, _ := oldB423TabletTable(t)
	for _, tc := range []struct {
		name   string
		colors []string
		cost   string
		want   int
	}{
		{"white", []string{"W"}, "{W}", 1},
		{"white and blue", []string{"W", "U"}, "{W}{U}", 2},
		{"white and red", []string{"W", "R"}, "{W}{R}", 1},
		{"red", []string{"R"}, "{R}", 0},
	} {
		life := me.Life
		oldB423CastColored(t, g, me, tc.colors, tc.cost)
		if got := me.Life - life; got != tc.want {
			t.Errorf("%s spell: gained %d, want %d", tc.name, got, tc.want)
		}
	}
}

func TestOldB423TabletIgnoresAnOpponentsSpells(t *testing.T) {
	g, me, _ := oldB423TabletTable(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	life := me.Life
	id := uuid.New()
	g.WithWriteLock(func() {
		opp.Hand.PushTop(game.Card{InstanceID: id, Name: "Instant", TypeLine: "Instant", ManaCost: "{W}", Colors: []string{"W"}, Owner: opp.ID, Controller: opp.ID})
	})
	opp.ManaPool.AddMana(game.ManaToken{Color: "W"})
	if err := g.CastSpell(opp.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Skipf("opponent could not cast out of turn here: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("life changed to %d from %d on an opponent's spell", me.Life, life)
	}
}

func TestOldB423ShriekingAfflictionDrainsAnOpponentWithOneOrFewerCards(t *testing.T) {
	for _, tc := range []struct {
		hand int
		lose int
	}{{0, 3}, {1, 3}, {2, 0}} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.PushPermanent(g, me.ID, game.Card{Name: "Shrieking Affliction", OracleID: oldB423AfflictionOracle, TypeLine: "Enchantment"})
		oldB423Hand(g, opp, tc.hand)
		life, mine := opp.Life, me.Life
		effectstest.BeginUpkeep(t, g, opp.ID)
		if got := life - opp.Life; got != tc.lose {
			t.Errorf("hand %d: lost %d, want %d", tc.hand, got, tc.lose)
		}
		if me.Life != mine {
			t.Errorf("hand %d: my life changed", tc.hand)
		}
	}
}

func TestOldB423ShriekingAfflictionIgnoresYourOwnUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushPermanent(g, me.ID, game.Card{Name: "Shrieking Affliction", OracleID: oldB423AfflictionOracle, TypeLine: "Enchantment"})
	oldB423Hand(g, me, 0)
	life := me.Life
	effectstest.BeginUpkeep(t, g, me.ID)
	if me.Life != life {
		t.Errorf("I lost life on my own upkeep: %d -> %d", life, me.Life)
	}
}

func TestOldB423RixMaadiGuildmageShrinksABlockingCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	guild := effectstest.B12Push(g, me.ID, "Rix Maadi Guildmage", "Creature — Human Shaman", oldB423RixMaadiOracle, 2, 2)
	idle := effectstest.B12Creature(g, opp.ID, "Idle", "Creature — Bear", 2, 2)
	err := effectstest.RavB2Activate(t, g, me.ID, guild, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(idle)}, "B", "R")
	if err == nil {
		t.Error("a creature that is not blocking was accepted")
	}
}

func TestOldB423RixMaadiGuildmageDrainsAPlayerWhoLostLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	guild := effectstest.B12Push(g, me.ID, "Rix Maadi Guildmage", "Creature — Human Shaman", oldB423RixMaadiOracle, 2, 2)
	if err := effectstest.RavB2Activate(t, g, me.ID, guild, 1, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}, "B", "R"); err == nil {
		t.Fatal("a player who has not lost life was accepted")
	}
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -1); err != nil {
			t.Fatalf("lose life: %v", err)
		}
	})
	life := opp.Life
	effectstest.RavB2MustActivate(t, g, me.ID, guild, 1, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}, "B", "R")
	if opp.Life != life-1 {
		t.Errorf("life %d, want %d", opp.Life, life-1)
	}
}

func TestOldB423RixMaadiGuildmageShrinksABlocker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	guild := effectstest.B12Push(g, opp.ID, "Rix Maadi Guildmage", "Creature — Human Shaman", oldB423RixMaadiOracle, 2, 2)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 3, 3)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.RavB2MustActivate(t, g, opp.ID, guild, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(blocker)}, "B", "R")
	if p, tough := effectstest.PtOf(t, g, blocker); p != 2 || tough != 2 {
		t.Errorf("the blocker is %d/%d, want 2/2", p, tough)
	}
}

func TestOldB423StonefareCrocodileGainsLifelink(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	croc := effectstest.B12Push(g, me.ID, "Stonefare Crocodile", "Creature — Crocodile", oldB423CrocodileOracle, 3, 2)
	if effectstest.HasEffectiveKeyword(t, g, croc, "lifelink") {
		t.Fatal("it has lifelink before the ability")
	}
	effectstest.RavB2MustActivate(t, g, me.ID, croc, 0, game.ActivateAbilityParams{}, "C", "C", "B")
	if !effectstest.HasEffectiveKeyword(t, g, croc, "lifelink") {
		t.Error("no lifelink after the ability")
	}
}

func TestOldB423NewPrahvGuildmageGrantsFlyingAndDetains(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	guild := effectstest.B12Push(g, me.ID, "New Prahv Guildmage", "Creature — Human Wizard", oldB423PrahvOracle, 2, 2)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	land := effectstest.B12Permanent(g, opp.ID, "Their Forest", "Basic Land — Forest")
	effectstest.RavB2MustActivate(t, g, me.ID, guild, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(bear)}, "W", "U")
	if !effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the target did not gain flying")
	}
	if err := effectstest.RavB2Activate(t, g, me.ID, guild, 1, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(land)}, "C", "C", "C", "W", "U"); err == nil {
		t.Error("a land was accepted as the nonland target")
	}
	effectstest.RavB2MustActivate(t, g, me.ID, guild, 1, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(rock)}, "C", "C", "C", "W", "U")
	if got := effectstest.RestrictionsOf(t, g, rock); got&game.CantAttackOrBlock != game.CantAttackOrBlock {
		t.Errorf("the artifact is not detained: %v", got.Names())
	}
}

func TestOldB423NewPrahvGuildmageCannotDetainYourOwnPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	guild := effectstest.B12Push(g, me.ID, "New Prahv Guildmage", "Creature — Human Wizard", oldB423PrahvOracle, 2, 2)
	mine := effectstest.B12Permanent(g, me.ID, "My Forest", "Basic Land — Forest")
	if err := effectstest.RavB2Activate(t, g, me.ID, guild, 1, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(mine)}, "C", "C", "C", "W", "U"); err == nil {
		t.Error("my own permanent was accepted")
	}
}

func TestOldB423InactionInjunctionDetainsAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Inaction Injunction", "Sorcery", oldB423InjunctionOracle, effectstest.B16TargetCard(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.RestrictionsOf(t, g, theirs); got&game.CantAttackOrBlock != game.CantAttackOrBlock {
		t.Errorf("the creature is not detained: %v", got.Names())
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("hand changed by %d, want +1", got)
	}
}

func TestOldB423InactionInjunctionRefusesYourOwnCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	mine := effectstest.B12Creature(g, g.Seats[g.Turn.ActiveSeat].ID, "My Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Inaction Injunction", "Sorcery", oldB423InjunctionOracle, effectstest.B16TargetCard(mine)); err == nil {
		t.Error("my own creature was accepted")
	}
}

func TestOldB423MartialLawDetainsAtYourUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushPermanent(g, me.ID, game.Card{Name: "Martial Law", OracleID: oldB423MartialLawOracle, TypeLine: "Enchantment"})
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.BeginUpkeep(t, g, me.ID)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.AnswerPickTarget(t, g, theirs)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := effectstest.RestrictionsOf(t, g, theirs); got&game.CantAttackOrBlock != game.CantAttackOrBlock {
		t.Errorf("the creature is not detained: %v", got.Names())
	}
}

func TestOldB423MartialLawIgnoresAnOpponentsUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushPermanent(g, me.ID, game.Card{Name: "Martial Law", OracleID: oldB423MartialLawOracle, TypeLine: "Enchantment"})
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.BeginUpkeep(t, g, opp.ID)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("a target prompt on an opponent's upkeep")
	}
	if got := effectstest.RestrictionsOf(t, g, theirs); got&game.CantAttackOrBlock != 0 {
		t.Errorf("the creature was detained on their own upkeep: %v", got.Names())
	}
}

func TestOldB423SoulswornSpiritIsUnblockableAndDetainsOnEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	spirit := effectstest.CastHoldCreature(t, g, "Soulsworn Spirit", "Creature — Spirit", oldB423SpiritOracle, 2, 1)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.AnswerPickTarget(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.RestrictionsOf(t, g, theirs); got&game.CantAttackOrBlock != game.CantAttackOrBlock {
		t.Errorf("the creature is not detained: %v", got.Names())
	}
	if got := effectstest.RestrictionsOf(t, g, spirit); got&game.CantBeBlocked == 0 {
		t.Errorf("the Spirit is blockable: %v", got.Names())
	}
}

func TestOldB423ArchonOfTheTriumvirateDetainsUpToTwoOnAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	archon := effectstest.B12Push(g, me.ID, "Archon of the Triumvirate", "Creature — Archon", oldB423ArchonOracle, 4, 5)
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Permanent(g, opp.ID, "B", "Artifact")
	c := effectstest.B12Creature(g, opp.ID, "C", "Creature — Bear", 2, 2)
	land := effectstest.B12Permanent(g, opp.ID, "Land", "Basic Land — Forest")
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(archon, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil || pick.PickTargetMax != 2 {
		t.Fatalf("up to two targets: %+v", pick)
	}
	if effectstest.HasID(pick.PickTargetCards, land) || effectstest.HasID(pick.PickTargetCards, mine) {
		t.Error("a land or my own permanent was offered")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, effectstest.CardRefs(a, b, c)); err == nil {
		t.Fatal("three targets were accepted")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, effectstest.CardRefs(a, b)); err != nil {
		t.Fatalf("two targets: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if got := effectstest.RestrictionsOf(t, g, id); got&game.CantAttackOrBlock != game.CantAttackOrBlock {
			t.Errorf("%v is not detained: %v", id, got.Names())
		}
	}
	if got := effectstest.RestrictionsOf(t, g, c); got&game.CantAttackOrBlock != 0 {
		t.Errorf("the third was detained: %v", got.Names())
	}
}
