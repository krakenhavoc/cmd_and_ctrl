package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b578_test.go — Ravnica: City of Guilds, batch 578 (#5600). Every test
// plays the real catalog entry through the engine.

const (
	oldB578Goblin    = "58d39991-420b-4eca-967d-49ad3f9a4702"
	oldB578Ghosts    = "9243bd95-5467-490d-9bd2-4d5ce1ebc589"
	oldB578Conscript = "a2019c23-09b8-4b0f-8b91-d8137804cb95"
	oldB578Enforcer  = "a4c6f546-ad58-41ba-aadd-724958cd4611"
	oldB578Searing   = "a6342d2c-06e3-4443-9276-b49f4a06a4da"
	oldB578Fusillade = "b45cb1b8-091a-40d2-a53a-2fa95845f4fa"
	oldB578Bear      = "Creature — Bear"
)

func TestOldB578EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB578Goblin: "War-Torch Goblin", oldB578Ghosts: "Ghosts of the Innocent", oldB578Conscript: "Flash Conscription",
		oldB578Enforcer: "Sunhome Enforcer", oldB578Searing: "Searing Meditation", oldB578Fusillade: "Flame Fusillade",
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

func TestOldB578WarTorchGoblinSacrificesToHitABlockingCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	goblin := effectstest.B12Push(g, me.ID, "War-Torch Goblin", "Creature — Goblin Warrior", oldB578Goblin, 1, 1)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", oldB578Bear, 4, 4)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", oldB578Bear, 2, 2)
	idle := effectstest.B12Creature(g, opp.ID, "Idle", oldB578Bear, 2, 2)
	effectstest.Pr7aBlockedCombat(t, g, opp.ID, attacker, blocker)
	effectstest.FloatMana(t, g, me, "{R}")
	if err := g.ActivateCatalogAbility(me.ID, goblin, 0, game.ActivateAbilityParams{Strict: true,
		Targets: effectstest.CardTarget(idle)}); err == nil {
		t.Fatal("targeted a creature that is not blocking")
	}
	if err := g.ActivateCatalogAbility(me.ID, goblin, 0, game.ActivateAbilityParams{Strict: true,
		Targets: effectstest.CardTarget(blocker)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(goblin) {
		t.Error("the Goblin was not sacrificed")
	}
	if g.Battlefield.Contains(blocker) {
		t.Error("the blocking 2/2 survived 2 damage")
	}
	if !g.Battlefield.Contains(idle) {
		t.Error("the idle creature was hit")
	}
}

func TestOldB578GhostsOfTheInnocentHalvesDamageRoundedDownToEveryone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Ghosts of the Innocent", "Creature — Spirit", oldB578Ghosts, 4, 5)
	src := effectstest.B12Creature(g, opp.ID, "Source", oldB578Bear, 1, 1)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", oldB578Bear, 9, 9)
	for _, c := range []struct {
		hit, want int
	}{{5, 2}, {1, 0}, {4, 2}} {
		before := effectstest.LifeOf(g, opp.ID)
		g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, opp.ID, c.hit) })
		if got := before - effectstest.LifeOf(g, opp.ID); got != c.want {
			t.Errorf("%d damage to a player dealt %d, want %d", c.hit, got, c.want)
		}
		beforeMe := effectstest.LifeOf(g, me.ID)
		g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, me.ID, c.hit) })
		if got := beforeMe - effectstest.LifeOf(g, me.ID); got != c.want {
			t.Errorf("%d damage to the Ghosts' controller dealt %d, want %d", c.hit, got, c.want)
		}
	}
	effectstest.B27Damage(g, src, bear, 7)
	if got := effectstest.Pr6Marked(g, bear); got != 3 {
		t.Errorf("a creature took %d marked damage from 7, want 3", got)
	}
}

func TestOldB578SunhomeEnforcerGainsLifeEqualToItsCombatDamageAndFirebreathes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	enf := effectstest.B12Push(g, me.ID, "Sunhome Enforcer", "Creature — Giant Soldier", oldB578Enforcer, 2, 4)
	effectstest.FloatMana(t, g, me, "{C}{R}{C}{R}")
	for i := 0; i < 2; i++ {
		if err := g.ActivateCatalogAbility(me.ID, enf, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("pump %d: %v", i, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if p, _ := effectstest.PtOf(t, g, enf); p != 4 {
		t.Fatalf("power %d, want 4 after two pumps", p)
	}
	life, oppLife := effectstest.LifeOf(g, me.ID), effectstest.LifeOf(g, opp.ID)
	effectstest.Pr7bAttack(t, g, opp.ID, enf)
	effectstest.Pr7bDamageStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != oppLife-4 {
		t.Errorf("opponent life %d, want %d", got, oppLife-4)
	}
	if got := effectstest.LifeOf(g, me.ID); got != life+4 {
		t.Errorf("life %d, want %d (the 4 combat damage dealt)", got, life+4)
	}
}

func TestOldB578SunhomeEnforcerGainsLifeForDamageDealtToACreatureButNotNoncombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	enf := effectstest.B12Push(g, me.ID, "Sunhome Enforcer", "Creature — Giant Soldier", oldB578Enforcer, 2, 4)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", oldB578Bear, 1, 1)
	life := effectstest.LifeOf(g, me.ID)
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(enf, victim, 2) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, me.ID); got != life {
		t.Errorf("noncombat damage gained life: %d, want %d", got, life)
	}
	effectstest.EmitDamageEventForTest(g, enf, victim, 2, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, me.ID); got != life+2 {
		t.Errorf("combat damage to a creature: life %d, want %d", got, life+2)
	}
}

func oldB578Conscription(t *testing.T, g *game.Game, me *game.Player, target uuid.UUID, mana string) {
	t.Helper()
	spell := effectstest.HandCardFull(me, "Flash Conscription", "Instant", "{5}{R}", oldB578Conscript, []string{"R"})
	effectstest.FloatMana(t, g, me, mana)
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(target)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB578FlashConscriptionStealsUntapsAndGivesHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", oldB578Bear, 3, 3)
	effectstest.ArnTap(t, g, bear)
	oldB578Conscription(t, g, me, bear, "{R}{R}{R}{R}{R}{R}")
	got := effectstest.LayeredCard(t, g, bear)
	if got.Controller != me.ID {
		t.Fatalf("controller %s, want the caster", got.Controller)
	}
	if got.Tapped {
		t.Error("the creature is still tapped")
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("the creature has no haste")
	}
	// No {W} was spent: no lifegain ability.
	life := effectstest.LifeOf(g, me.ID)
	effectstest.DealCombatDamageToPlayer(g, bear, opp.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LifeOf(g, me.ID) != life {
		t.Errorf("life %d → %d without white mana spent", life, effectstest.LifeOf(g, me.ID))
	}
	// Control returns at the end of the turn.
	start := g.Turn.Seq
	for i := 0; i < 60 && g.Turn.Seq == start; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	if c := effectstest.LayeredCard(t, g, bear); c.Controller != opp.ID {
		t.Errorf("after the turn the controller is %s, want the owner", c.Controller)
	}
}

func TestOldB578FlashConscriptionWithWhiteGrantsTheLifegainTrigger(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", oldB578Bear, 3, 3)
	oldB578Conscription(t, g, me, bear, "{W}{R}{R}{R}{R}{R}")
	life := effectstest.LifeOf(g, me.ID)
	effectstest.DealCombatDamageToPlayer(g, bear, opp.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, me.ID); got != life+3 {
		t.Errorf("life %d, want %d: the caster gains the damage dealt", got, life+3)
	}
}

func TestOldB578SearingMeditationPaysTwoForTwoDamageWhenYouGainLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Searing Meditation", "Enchantment", oldB578Searing, 0, 0)
	for _, pay := range []bool{true, false} {
		before := effectstest.LifeOf(g, opp.ID)
		effectstest.FloatMana(t, g, me, "{C}{C}")
		g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 1) })
		effectstest.PassPriorityAroundTable(t, g)
		if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
			if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
				t.Fatalf("pick target: %v", err)
			}
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		want := before
		if pay {
			want = before - 2
		}
		if got := effectstest.LifeOf(g, opp.ID); got != want {
			t.Errorf("pay=%v: opponent life %d, want %d", pay, got, want)
		}
	}
}

func oldB578Ping(g *game.Game, who *game.Player, permanent uuid.UUID, target uuid.UUID) error {
	var err error
	for idx := 0; idx < 3; idx++ {
		if err = g.ActivateCatalogAbility(who.ID, permanent, idx, game.ActivateAbilityParams{Strict: true,
			Targets: effectstest.PlayerRefs(target)}); err == nil {
			return nil
		}
	}
	return err
}

func TestOldB578FlameFusilladeGivesEveryPermanentYouControlAPingThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	land := effectstest.B12Push(g, me.ID, "Plains", "Basic Land — Plains", "", 0, 0)
	creature := effectstest.B12Creature(g, me.ID, "Mine", oldB578Bear, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", oldB578Bear, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Flame Fusillade", "Sorcery", oldB578Fusillade, nil)
	effectstest.PassPriorityAroundTable(t, g)
	life := effectstest.LifeOf(g, opp.ID)
	for _, id := range []uuid.UUID{land, creature} {
		if err := oldB578Ping(g, me, id, opp.ID); err != nil {
			t.Fatalf("a permanent you control has no ping: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := effectstest.LifeOf(g, opp.ID); got != life-2 {
		t.Errorf("opponent life %d, want %d after two pings", got, life-2)
	}
	if err := oldB578Ping(g, me, land, opp.ID); err == nil {
		t.Error("the tapped land pinged a second time")
	}
	late := effectstest.B12Creature(g, me.ID, "Late", oldB578Bear, 2, 2)
	if err := oldB578Ping(g, me, late, opp.ID); err == nil {
		t.Error("a permanent that arrived after the spell resolved got the ping")
	}
	if err := oldB578Ping(g, opp, theirs, me.ID); err == nil {
		t.Error("an opponent's permanent got the ping")
	}
}
