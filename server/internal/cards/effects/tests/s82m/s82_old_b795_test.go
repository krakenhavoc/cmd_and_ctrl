package testss82m

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b795_test.go — Mirage, batch 795 (#5633). Every test plays the real
// catalog entry through the engine.

const (
	oldB795Hunter   = "b96ccc2f-88e5-4522-924e-76fc76dc3b39"
	oldB795Numbers  = "b98177b5-599f-41b0-aa1e-4769d7005bbd"
	oldB795Dagger   = "bf034a20-e362-4f38-9bf9-857777c62598"
	oldB795Pendulum = "bf7c16db-01c9-41ed-bda6-c5326198ce77"
	oldB795Torch    = "c629feae-084c-452f-9c09-4b9716331498"
	oldB795Equity   = "c7986014-30e0-4999-ab95-b86620471419"
	oldB795Firewalk = "c9c0bb7e-8dfc-481c-80f7-b6652b63df39"
	oldB795Shadow   = "cdda63d7-e3ea-455c-b7d4-c6cfdda2703c"
	oldB795Torrent  = "f467ec85-ffe2-4963-82e0-509a7873fd63"
	oldB795Specter  = "fc638800-6aec-4a48-ad2c-c9a5ce93c236"

	oldB795CounterspellOracle = "cc187110-1148-4090-bbb8-e205694a39f5"
	oldB795ShockOracle        = "a9d288b8-cdc1-4e55-a0c9-d6edfc95e65d"
)

func TestOldB795EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB795Hunter: "Abyssal Hunter", oldB795Numbers: "Superior Numbers", oldB795Dagger: "Acidic Dagger",
		oldB795Pendulum: "Razor Pendulum", oldB795Torch: "Kaervek's Torch", oldB795Equity: "Mangara's Equity",
		oldB795Firewalk: "Suq'Ata Firewalker", oldB795Shadow: "Shadow Guildmage", oldB795Torrent: "Torrent of Lava",
		oldB795Specter: "Tainted Specter",
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

func oldB795Card(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

func oldB795Player(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetPlayer, ID: id}}
}

func oldB795Colored(g *game.Game, owner uuid.UUID, name string, p, tough int, colors []string, kws ...string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: name,
		TypeLine: "Creature — Test", Power: p, Toughness: tough, Colors: colors, Keywords: kws, Owner: owner, Controller: owner})
}

func oldB795Cast(t *testing.T, g *game.Game, name, typeLine, oracle string, x int, targets []game.TargetRef) {
	t.Helper()
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	c := game.NewCard(name, me.ID)
	c.TypeLine, c.OracleID = typeLine, oracle
	me.Hand.PushTop(c)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Targets: targets, XValue: x}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB795AbyssalHunterTapsItsTargetAndDealsItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	h := effectstest.BlbB1Push(g, me.ID, "Abyssal Hunter", "Creature — Human Assassin", oldB795Hunter, 3, 1)
	victim := oldB795Colored(g, opp.ID, "Bear", 2, 5, []string{"G"})
	effectstest.FloatMana(t, g, me, "{B}")
	effectstest.B16Activate(t, g, me.ID, h, 0, game.ActivateAbilityParams{Strict: true, Targets: oldB795Card(victim)})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, victim) {
		t.Error("the target is not tapped")
	}
	if got := effectstest.DamageOn(t, g, victim); got != 3 {
		t.Errorf("the target has %d damage, want 3 (the Hunter's power)", got)
	}
}

func TestOldB795SuperiorNumbersCountsTheExcess(t *testing.T) {
	for _, tc := range []struct {
		name         string
		mine, theirs int
		want         int
	}{{"three against one", 3, 1, 2}, {"even", 2, 2, 0}, {"outnumbered", 1, 3, 0}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.AdvanceToMain(t, g)
			for i := 0; i < tc.mine; i++ {
				oldB795Colored(g, me.ID, "Mine", 1, 1, []string{"G"})
			}
			var victim uuid.UUID
			for i := 0; i < tc.theirs; i++ {
				victim = oldB795Colored(g, opp.ID, "Theirs", 1, 9, []string{"R"})
			}
			if victim == uuid.Nil {
				t.Fatal("no victim")
			}
			oldB795Cast(t, g, "Superior Numbers", "Sorcery", oldB795Numbers, 0, []game.TargetRef{
				{Kind: game.TargetCard, ID: victim}, {Kind: game.TargetPlayer, ID: opp.ID}})
			if got := effectstest.DamageOn(t, g, victim); got != tc.want {
				t.Errorf("the target has %d damage, want %d", got, tc.want)
			}
		})
	}
}

func oldB795ActivateDagger(t *testing.T, g *game.Game, me *game.Player, dagger, target uuid.UUID) error {
	t.Helper()
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}")
	return g.ActivateCatalogAbility(me.ID, dagger, 0, game.ActivateAbilityParams{Strict: true, Targets: oldB795Card(target)})
}

func TestOldB795AcidicDaggerDestroysANonWallCreatureTheTargetHits(t *testing.T) {
	for _, tc := range []struct {
		name     string
		typeLine string
		dies     bool
	}{{"non-Wall", "Creature — Bear", true}, {"Wall", "Creature — Wall", false}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.AdvanceToMain(t, g)
			dagger := effectstest.PushCatalogPermanent(g, me.ID, "Acidic Dagger", "Artifact", oldB795Dagger, false)
			attacker := oldB795Colored(g, me.ID, "Attacker", 2, 5, []string{"R"})
			blocker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Blocker",
				TypeLine: tc.typeLine, Power: 0, Toughness: 9, Colors: []string{"W"}, Owner: opp.ID, Controller: opp.ID})
			if err := oldB795ActivateDagger(t, g, me, dagger, attacker); err != nil {
				t.Fatalf("activate: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.Pr7bAttack(t, g, opp.ID, attacker)
			effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{blocker: attacker})
			effectstest.Pr7bDamageStep(t, g)
			effectstest.PassPriorityAroundTable(t, g)
			if got := !g.Battlefield.Contains(blocker); got != tc.dies {
				t.Errorf("blocker destroyed = %v, want %v", got, tc.dies)
			}
		})
	}
}

func TestOldB795AcidicDaggerIsSacrificedWhenTheTargetLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	dagger := effectstest.PushCatalogPermanent(g, me.ID, "Acidic Dagger", "Artifact", oldB795Dagger, false)
	target := oldB795Colored(g, me.ID, "Target", 2, 2, []string{"R"})
	if err := oldB795ActivateDagger(t, g, me, dagger, target); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1Destroy(t, g, target)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(dagger) {
		t.Error("the Dagger was not sacrificed when its target left")
	}
}

func TestOldB795AcidicDaggerCannotBeActivatedOnceBlockersAreDeclared(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	dagger := effectstest.PushCatalogPermanent(g, me.ID, "Acidic Dagger", "Artifact", oldB795Dagger, false)
	attacker := oldB795Colored(g, me.ID, "Attacker", 2, 5, []string{"R"})
	blocker := oldB795Colored(g, opp.ID, "Blocker", 0, 9, []string{"W"})
	effectstest.Pr7bAttack(t, g, opp.ID, attacker)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{blocker: attacker})
	if err := oldB795ActivateDagger(t, g, me, dagger, attacker); err == nil {
		t.Error("activated after blockers were declared")
	}
}

func TestOldB795RazorPendulumHurtsAnyPlayerAtFiveOrLess(t *testing.T) {
	for _, tc := range []struct {
		name string
		life int
		want int
	}{{"five", 5, 2}, {"six", 6, 0}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.AdvanceToMain(t, g)
			effectstest.PushCatalogPermanent(g, me.ID, "Razor Pendulum", "Artifact", oldB795Pendulum, false)
			g.WithWriteLock(func() { opp.Life = tc.life })
			effectstest.AdvanceToEndStepOf(t, g, 1)
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.LifeOf(g, opp.ID); got != tc.life-tc.want {
				t.Errorf("life %d, want %d", got, tc.life-tc.want)
			}
		})
	}
}

func TestOldB795RazorPendulumHurtsItsOwnControllerToo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Razor Pendulum", "Artifact", oldB795Pendulum, false)
	g.WithWriteLock(func() { me.Life = 4 })
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, me.ID); got != 2 {
		t.Errorf("life %d, want 2", got)
	}
}

func TestOldB795KaerveksTorchDealsXDamageToAnyTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	before := effectstest.LifeOf(g, opp.ID)
	oldB795Cast(t, g, "Kaervek's Torch", "Sorcery", oldB795Torch, 4, oldB795Player(opp.ID))
	if got := effectstest.LifeOf(g, opp.ID); got != before-4 {
		t.Errorf("life %d, want %d", got, before-4)
	}
}

func TestOldB795KaerveksTorchTaxesSpellsThatTargetIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	torch := effectstest.CastCatalogSpell(t, g, "Kaervek's Torch", "Sorcery", oldB795Torch, oldB795Player(opp.ID))
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	cs := effectstest.HandCardFull(opp, "Counterspell", "Instant", "{U}{U}", oldB795CounterspellOracle, []string{"U"})
	at := game.TargetRef{Kind: game.TargetCard, ID: torch}
	effectstest.FloatForTest(g, opp, "UU")
	if err := g.CastSpell(opp.ID, cs, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{at}}); !errors.Is(err, game.ErrInsufficientMana) {
		t.Fatalf("Counterspell at the Torch for {U}{U} = %v, want ErrInsufficientMana", err)
	}
	effectstest.FloatForTest(g, opp, "CC")
	if err := g.CastSpell(opp.ID, cs, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{at}}); err != nil {
		t.Fatalf("Counterspell at the Torch for {2}{U}{U}: %v", err)
	}
	shock := effectstest.HandCardFull(opp, "Shock", "Instant", "{R}", oldB795ShockOracle, []string{"R"})
	effectstest.FloatForTest(g, opp, "R")
	if err := g.CastSpell(opp.ID, shock, game.CastSpellParams{Strict: true, Targets: oldB795Player(me.ID)}); err != nil {
		t.Errorf("a spell that does not target the Torch was taxed: %v", err)
	}
}

func oldB795PushEquity(g *game.Game, me *game.Player, color string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Mangara's Equity",
		TypeLine: "Enchantment", OracleID: oldB795Equity, ChosenColor: color, Owner: me.ID, Controller: me.ID})
}

func TestOldB795MangarasEquityReflectsDamageFromTheChosenColor(t *testing.T) {
	for _, tc := range []struct {
		name   string
		color  string
		toYou  bool
		wantBk int
	}{
		{"red creature hits you", "R", true, 3},
		{"black creature hits you when red is chosen", "B", true, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.AdvanceToMain(t, g)
			oldB795PushEquity(g, me, "R")
			dealer := oldB795Colored(g, opp.ID, "Dealer", 3, 9, []string{tc.color})
			effectstest.P2Ping(t, g, dealer, me.ID, 3)
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.DamageOn(t, g, dealer); got != tc.wantBk {
				t.Errorf("the creature has %d damage, want %d", got, tc.wantBk)
			}
		})
	}
}

func TestOldB795MangarasEquityGuardsYourWhiteCreaturesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	oldB795PushEquity(g, me, "B")
	dealer := oldB795Colored(g, opp.ID, "Dealer", 2, 9, []string{"B"})
	white := oldB795Colored(g, me.ID, "Knight", 2, 9, []string{"W"})
	green := oldB795Colored(g, me.ID, "Bear", 2, 9, []string{"G"})
	effectstest.P2Ping(t, g, dealer, green, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageOn(t, g, dealer); got != 0 {
		t.Fatalf("hurting a green creature reflected %d", got)
	}
	effectstest.P2Ping(t, g, dealer, white, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageOn(t, g, dealer); got != 2 {
		t.Errorf("hurting a white creature reflected %d, want 2", got)
	}
}

func TestOldB795SuqAtaFirewalkerPingsButCannotBeTargetedByRed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	f := effectstest.BlbB1Push(g, me.ID, "Suq'Ata Firewalker", "Creature — Human Wizard", oldB795Firewalk, 0, 1)
	before := effectstest.LifeOf(g, opp.ID)
	effectstest.B16Activate(t, g, me.ID, f, 0, game.ActivateAbilityParams{Strict: true, Targets: oldB795Player(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != before-1 {
		t.Errorf("life %d, want %d", got, before-1)
	}
	shock := effectstest.HandCardFull(me, "Shock", "Instant", "{R}", oldB795ShockOracle, []string{"R"})
	effectstest.FloatForTest(g, me, "R")
	if err := g.CastSpell(me.ID, shock, game.CastSpellParams{Strict: true, Targets: oldB795Card(f)}); err == nil {
		t.Error("a red spell targeted the Firewalker")
	}
}

func TestOldB795ShadowGuildmagePutsACreatureOnTopOfItsLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	m := effectstest.BlbB1Push(g, me.ID, "Shadow Guildmage", "Creature — Human Wizard", oldB795Shadow, 1, 1)
	bear := oldB795Colored(g, me.ID, "Bear", 2, 2, []string{"G"})
	effectstest.FloatMana(t, g, me, "{U}")
	effectstest.B16Activate(t, g, me.ID, m, 0, game.ActivateAbilityParams{Strict: true, Targets: oldB795Card(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Fatal("the Bear is still on the battlefield")
	}
	if top, err := me.Library.Top(); err != nil || top.InstanceID != bear {
		t.Error("the Bear is not on top of its owner's library")
	}
}

func TestOldB795ShadowGuildmagePingsAndHurtsYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	m := effectstest.BlbB1Push(g, me.ID, "Shadow Guildmage", "Creature — Human Wizard", oldB795Shadow, 1, 1)
	myLife, oppLife := effectstest.LifeOf(g, me.ID), effectstest.LifeOf(g, opp.ID)
	effectstest.FloatMana(t, g, me, "{R}")
	effectstest.B16Activate(t, g, me.ID, m, 1, game.ActivateAbilityParams{Strict: true, Targets: oldB795Player(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != oppLife-1 {
		t.Errorf("opponent life %d, want %d", got, oppLife-1)
	}
	if got := effectstest.LifeOf(g, me.ID); got != myLife-1 {
		t.Errorf("your life %d, want %d", got, myLife-1)
	}
}

func TestOldB795TorrentOfLavaHurtsGroundCreaturesAndEachCanShieldOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	shielded := oldB795Colored(g, me.ID, "Shielded", 2, 9, []string{"G"})
	bare := oldB795Colored(g, opp.ID, "Bare", 2, 9, []string{"G"})
	flyer := oldB795Colored(g, opp.ID, "Bird", 2, 9, []string{"W"}, "flying")
	c := game.NewCard("Torrent of Lava", me.ID)
	c.TypeLine, c.OracleID = "Sorcery", oldB795Torrent
	me.Hand.PushTop(c)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{XValue: 3}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	idx, ref := effectstest.GrantedActivatedIndex(t, g, shielded)
	if idx < 0 {
		t.Fatal("a creature has no granted ability while Torrent of Lava is on the stack")
	}
	effectstest.B16Activate(t, g, me.ID, shielded, idx, game.ActivateAbilityParams{Ref: ref, Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageOn(t, g, shielded); got != 2 {
		t.Errorf("the shielded creature has %d damage, want 2", got)
	}
	if got := effectstest.DamageOn(t, g, bare); got != 3 {
		t.Errorf("the bare creature has %d damage, want 3", got)
	}
	if got := effectstest.DamageOn(t, g, flyer); got != 0 {
		t.Errorf("the flyer has %d damage, want 0", got)
	}
	if idx, _ := effectstest.GrantedActivatedIndex(t, g, bare); idx >= 0 {
		t.Error("the grant outlived the spell")
	}
}

func TestOldB795TaintedSpecterPutOnTopStopsTheDiscardAndTheDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	s := effectstest.BlbB1Push(g, me.ID, "Tainted Specter", "Creature — Specter", oldB795Specter, 2, 2)
	opp.Hand.Cards = nil
	card := effectstest.PushCatalogHandCard(opp, "Filler", "Sorcery", "")
	oppLife := effectstest.LifeOf(g, opp.ID)
	effectstest.FloatMana(t, g, me, "{B}{B}{B}")
	effectstest.B16Activate(t, g, me.ID, s, 0, game.ActivateAbilityParams{Strict: true, Targets: oldB795Player(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AppAnswerWith(t, g, func(*game.PendingChoice) (bool, []uuid.UUID) { return true, []uuid.UUID{card} })
	if opp.Hand.Contains(card) {
		t.Error("the card is still in their hand")
	}
	if top, err := opp.Library.Top(); err != nil || top.InstanceID != card {
		t.Error("the card is not on top of their library")
	}
	if opp.Graveyard.Contains(card) {
		t.Error("the card was discarded")
	}
	if got := effectstest.LifeOf(g, opp.ID); got != oppLife {
		t.Errorf("life %d, want %d: nothing was discarded", got, oppLife)
	}
}

func TestOldB795TaintedSpecterDeclinedMakesThemDiscardAndDamagesEveryone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	s := effectstest.BlbB1Push(g, me.ID, "Tainted Specter", "Creature — Specter", oldB795Specter, 2, 2)
	bystander := oldB795Colored(g, opp.ID, "Bystander", 2, 5, []string{"G"})
	opp.Hand.Cards = nil
	card := effectstest.PushCatalogHandCard(opp, "Filler", "Sorcery", "")
	myLife, oppLife := effectstest.LifeOf(g, me.ID), effectstest.LifeOf(g, opp.ID)
	effectstest.FloatMana(t, g, me, "{B}{B}{B}")
	effectstest.B16Activate(t, g, me.ID, s, 0, game.ActivateAbilityParams{Strict: true, Targets: oldB795Player(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AppAnswerWith(t, g, func(*game.PendingChoice) (bool, []uuid.UUID) { return false, nil })
	effectstest.AnswerChooseCards(t, g, opp.ID, card)
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(card) {
		t.Fatal("the card was not discarded")
	}
	if got := effectstest.LifeOf(g, opp.ID); got != oppLife-1 {
		t.Errorf("their life %d, want %d", got, oppLife-1)
	}
	if got := effectstest.LifeOf(g, me.ID); got != myLife-1 {
		t.Errorf("your life %d, want %d", got, myLife-1)
	}
	if got := effectstest.DamageOn(t, g, bystander); got != 1 {
		t.Errorf("the bystander has %d damage, want 1", got)
	}
	if got := effectstest.DamageOn(t, g, s); got != 1 {
		t.Errorf("the Specter has %d damage, want 1", got)
	}
}

func TestOldB795TaintedSpecterIsSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	s := effectstest.BlbB1Push(g, me.ID, "Tainted Specter", "Creature — Specter", oldB795Specter, 2, 2)
	atk := oldB795Colored(g, me.ID, "Attacker", 1, 1, []string{"R"})
	effectstest.Pr7bAttack(t, g, opp.ID, atk)
	effectstest.FloatMana(t, g, me, "{B}{B}{B}")
	if err := g.ActivateCatalogAbility(me.ID, s, 0, game.ActivateAbilityParams{Strict: true, Targets: oldB795Player(opp.ID)}); err == nil {
		t.Error("activated outside a main phase")
	}
}
