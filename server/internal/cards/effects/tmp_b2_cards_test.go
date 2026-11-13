package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tmp_b2_cards_test.go — Tempest, batch 2 (#5628). Each test plays the
// real catalog entry and asserts what its text says. The seeding helpers
// (tmpB1Push, tmpB1Shadow, tmpB1BuybackCases, tmpB1AssertBuyback) are
// batch 1's.

const (
	tmpB2LancerOracle        = "d2cbd71a-841b-4ff2-a794-247a030e0233"
	tmpB2HorrorOracle        = "07c20708-c305-46e7-a98e-c99f54aeafec"
	tmpB2SlayerOracle        = "c1d0537b-954f-4d3b-b9fe-1dd7ffaa1197"
	tmpB2EmissaryOracle      = "a8211eb1-332d-41c0-8086-4a18cd7a5832"
	tmpB2ShadowRiftOracle    = "66810ef7-ea3a-4afa-92f0-b21ec52cc8d0"
	tmpB2RealityAnchorOracle = "a7b4a519-d85f-40ae-bb5f-08bc1d301418"
	tmpB2EmbraceOracle       = "b19a7ae7-3af5-4aa6-80e1-8888054427de"
	tmpB2AnointOracle        = "fcb03c54-5b9a-428d-bbf6-ef8977770c8d"
	tmpB2JusticeOracle       = "c0d3dfc0-1ff2-4ac3-98eb-32eaa5b84c05"
	tmpB2TauntOracle         = "7ebec169-d941-460e-9e3c-ba54b5066b05"
	tmpB2MawcorOracle        = "4c0384d6-21cb-4645-b7a4-3af133e8a9a8"
	tmpB2RootwaterOracle     = "8e42ed82-8b56-46ad-a1ca-e29689a9d030"
	tmpB2DismissOracle       = "5c828a5e-10ae-4f63-86fd-2f160af43cc3"
	tmpB2EdictOracle         = "058917c1-21ab-488a-9f9c-591c55f3c596"
)

func TestTmpB2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		tmpB2LancerOracle:        "Soltari Lancer",
		tmpB2HorrorOracle:        "Dauthi Horror",
		tmpB2SlayerOracle:        "Dauthi Slayer",
		tmpB2EmissaryOracle:      "Soltari Emissary",
		tmpB2ShadowRiftOracle:    "Shadow Rift",
		tmpB2RealityAnchorOracle: "Reality Anchor",
		tmpB2EmbraceOracle:       "Dauthi Embrace",
		tmpB2AnointOracle:        "Anoint",
		tmpB2JusticeOracle:       "Evincar's Justice",
		tmpB2TauntOracle:         "Imps' Taunt",
		tmpB2MawcorOracle:        "Mawcor",
		tmpB2RootwaterOracle:     "Rootwater Hunter",
		tmpB2DismissOracle:       "Dismiss",
		tmpB2EdictOracle:         "Diabolic Edict",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- shadow creatures -----------------------------------------------------------

func TestTmpB2ShadowAttackersAreBlockedOnlyByShadow(t *testing.T) {
	for _, tc := range []struct{ name, oracle, typeLine string }{
		{"Soltari Lancer", tmpB2LancerOracle, "Creature — Soltari Knight"},
		{"Dauthi Horror", tmpB2HorrorOracle, "Creature — Dauthi Horror"},
		{"Dauthi Slayer", tmpB2SlayerOracle, "Creature — Dauthi Soldier"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, them := seatsForCombat(g)
			attacker := tmpB1Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 2, "B")
			shadow := tmpB1Shadow(g, them.ID, "Shadow blocker", "U")
			plain := tmpB1Push(g, them.ID, "Plain blocker", "Creature — Bear", "", 2, 2, "U")
			attackIntoBlocks(t, g, them.ID, attacker)
			if got := blockRefusal(t, g, shadow, attacker); got != "" {
				t.Errorf("a blocker with shadow was refused: %q", got)
			}
			if got := blockRefusal(t, g, plain, attacker); got != game.BlockReasonShadow {
				t.Errorf("a blocker without shadow: refusal %q, want %q", got, game.BlockReasonShadow)
			}
		})
	}
}

func TestTmpB2SoltariLancerHasFirstStrikeOnlyWhileAttacking(t *testing.T) {
	g := newCatalogGame(t)
	me, them := seatsForCombat(g)
	lancer := tmpB1Push(g, me.ID, "Soltari Lancer", "Creature — Soltari Knight", tmpB2LancerOracle, 2, 2, "W")
	if hasEffectiveKeyword(t, g, lancer, "first strike") {
		t.Fatal("Soltari Lancer has first strike before it attacks")
	}
	declareAttack(t, g, them.ID, lancer)
	if !hasEffectiveKeyword(t, g, lancer, "first strike") {
		t.Error("Soltari Lancer does not have first strike while attacking")
	}
}

func TestTmpB2DauthiHorrorCantBeBlockedByWhiteCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, them := seatsForCombat(g)
	horror := tmpB1Push(g, me.ID, "Dauthi Horror", "Creature — Dauthi Horror", tmpB2HorrorOracle, 2, 1, "B")
	white := tmpB1Shadow(g, them.ID, "White shadow blocker", "W")
	black := tmpB1Shadow(g, them.ID, "Black shadow blocker", "B")
	attackIntoBlocks(t, g, them.ID, horror)
	if got := blockRefusal(t, g, white, horror); got != game.BlockReasonCantBeBlockedBy {
		t.Errorf("a white creature with shadow: refusal %q, want %q", got, game.BlockReasonCantBeBlockedBy)
	}
	if got := blockRefusal(t, g, black, horror); got != "" {
		t.Errorf("a black creature with shadow was refused: %q", got)
	}
}

func TestTmpB2DauthiSlayerAttacksEachCombatIfAble(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, them := seatsForCombat(g)
	slayer := tmpB1Push(g, me.ID, "Dauthi Slayer", "Creature — Dauthi Soldier", tmpB2SlayerOracle, 2, 2, "B")
	advanceToDeclareAttackersOf(t, g, seat)
	var re *game.AttackRequirementError
	if err := g.PassPriority(); !errors.As(err, &re) || re.Attacker != slayer {
		t.Fatalf("pass with the Slayer at home = %v, want a refusal naming it", err)
	}
	if err := g.DeclareAttacker(slayer, them.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Errorf("pass with the Slayer attacking: %v", err)
	}
}

// --- shadow grants and losses ---------------------------------------------------

func TestTmpB2SoltariEmissaryGainsShadowUntilEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	emissary := tmpB1Push(g, me.ID, "Soltari Emissary", "Creature — Soltari Soldier", tmpB2EmissaryOracle, 2, 1, "W")
	if hasEffectiveKeyword(t, g, emissary, "shadow") {
		t.Fatal("Soltari Emissary has shadow before the activation")
	}
	dcMana(t, g, me.ID, "{W}")
	if err := g.ActivateCatalogAbility(me.ID, emissary, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, emissary, "shadow") {
		t.Fatal("Soltari Emissary does not have shadow after the activation")
	}
	if err := g.ActivateCatalogAbility(me.ID, emissary, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the ability was activated with an empty mana pool")
	}
	advanceTo(t, g, game.StepCleanup)
	if hasEffectiveKeyword(t, g, emissary, "shadow") {
		t.Error("Soltari Emissary still has shadow after the cleanup step")
	}
}

func TestTmpB2ShadowRiftGrantsShadowAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := seedCreature(g, "Grizzly Bears", me.ID)
	castCatalogSpell(t, g, "Shadow Rift", "Instant", tmpB2ShadowRiftOracle, cardTarget(bear))
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, bear, "shadow") {
		t.Error("the target did not gain shadow")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn", hand, got)
	}
}

func TestTmpB2RealityAnchorRemovesShadowAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, them := seatsForCombat(g)
	soldier := tmpB1Push(g, me.ID, "Soltari Foot Soldier", "Creature — Soltari Soldier", tmpB1FootSoldierOracle, 1, 1, "W")
	plain := tmpB1Push(g, them.ID, "Plain blocker", "Creature — Bear", "", 2, 2, "G")
	castCatalogSpell(t, g, "Reality Anchor", "Instant", tmpB2RealityAnchorOracle, cardTarget(soldier))
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if hasEffectiveKeyword(t, g, soldier, "shadow") {
		t.Fatal("the target still has shadow")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn", hand, got)
	}
	attackIntoBlocks(t, g, them.ID, soldier)
	if got := blockRefusal(t, g, plain, soldier); got != "" {
		t.Errorf("a creature without shadow was refused the block: %q", got)
	}
}

func TestTmpB2DauthiEmbraceGivesTargetShadow(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, them := seatsForCombat(g)
	embrace := tmpB1Push(g, me.ID, "Dauthi Embrace", "Enchantment", tmpB2EmbraceOracle, 0, 0, "B")
	theirs := seedCreature(g, "Their Bear", them.ID)
	dcMana(t, g, me.ID, "{B}{B}")
	if err := invB1Activate(g, me.ID, embrace, 0, cardTarget(theirs)...); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, theirs, "shadow") {
		t.Error("the target did not gain shadow")
	}
}

// --- buyback spells ---------------------------------------------------------------

func TestTmpB2AnointPreventsTheNextThreeDamage(t *testing.T) {
	for _, tc := range tmpB1BuybackCases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			caster := g.Seats[g.Turn.ActiveSeat]
			ox := tmpB1Push(g, caster.ID, "Ox", "Creature — Ox", "", 2, 5)
			id, err := castWithOptionalCosts(t, g, "Anoint", "Instant", tmpB2AnointOracle, cardTarget(ox), tc.optional, nil)
			if err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			passPriorityAroundTable(t, g)
			tmpB1AssertBuyback(t, caster, id, tc.wantHand)
			g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.New(), ox, 2) })
			if got := damageMarkedOn(g, ox); got != 0 {
				t.Fatalf("2 damage into a 3-point shield: marked %d, want 0", got)
			}
			g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.New(), ox, 2) })
			if got := damageMarkedOn(g, ox); got != 1 {
				t.Errorf("2 damage into the last point of the shield: marked %d, want 1", got)
			}
		})
	}
}

func TestTmpB2EvincarsJusticeDamagesEachCreatureAndEachPlayer(t *testing.T) {
	for _, tc := range tmpB1BuybackCases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			caster, opp := seatsForCombat(g)
			small := tmpB1Push(g, opp.ID, "Small", "Creature — Goblin", "", 2, 2)
			big := tmpB1Push(g, caster.ID, "Big", "Creature — Ox", "", 3, 3)
			life := map[uuid.UUID]int{}
			for _, p := range g.Seats {
				life[p.ID] = p.Life
			}
			id, err := castWithOptionalCosts(t, g, "Evincar's Justice", "Sorcery", tmpB2JusticeOracle, nil, tc.optional, nil)
			if err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			passPriorityAroundTable(t, g)
			for _, p := range g.Seats {
				if got := life[p.ID] - p.Life; got != 2 {
					t.Errorf("%s lost %d life, want 2", p.Name, got)
				}
			}
			if g.Battlefield.Contains(small) {
				t.Error("the 2/2 survived 2 damage")
			}
			if got := damageMarkedOn(g, big); got != 2 {
				t.Errorf("the 3/3 has %d damage marked, want 2", got)
			}
			tmpB1AssertBuyback(t, caster, id, tc.wantHand)
		})
	}
}

func TestTmpB2ImpsTauntMakesTheTargetAttack(t *testing.T) {
	for _, tc := range tmpB1BuybackCases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			seat := g.Turn.ActiveSeat
			caster := g.Seats[seat]
			bear := seedCreature(g, "Grizzly Bears", caster.ID)
			id, err := castWithOptionalCosts(t, g, "Imps' Taunt", "Instant", tmpB2TauntOracle, cardTarget(bear), tc.optional, nil)
			if err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			passPriorityAroundTable(t, g)
			tmpB1AssertBuyback(t, caster, id, tc.wantHand)
			advanceToDeclareAttackersOf(t, g, seat)
			var re *game.AttackRequirementError
			if err := g.PassPriority(); !errors.As(err, &re) || re.Attacker != bear {
				t.Errorf("pass with the target at home = %v, want a refusal naming it", err)
			}
		})
	}
}

// --- pingers ----------------------------------------------------------------------

func TestTmpB2PingersDealOneDamageToAnyTarget(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, typeLine string
		power, toughness       int
	}{
		{"Mawcor", tmpB2MawcorOracle, "Creature — Beast", 3, 3},
		{"Rootwater Hunter", tmpB2RootwaterOracle, "Creature — Merfolk", 1, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me, opp, _ := blbB1Seats(g)
			pinger := tmpB1Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness, "U")
			before := opp.Life
			if err := invB1Activate(g, me.ID, pinger, 0, playerTarget(opp.ID)); err != nil {
				t.Fatalf("activate: %v", err)
			}
			passPriorityAroundTable(t, g)
			if got := before - opp.Life; got != 1 {
				t.Errorf("the target lost %d life, want 1", got)
			}
			if !findCardAnywhere(t, g, pinger).Tapped {
				t.Errorf("%s did not tap", tc.name)
			}
			if err := invB1Activate(g, me.ID, pinger, 0, playerTarget(opp.ID)); err == nil {
				t.Error("the ability was activated again while tapped")
			}
		})
	}
}

func TestTmpB2MawcorHasFlying(t *testing.T) {
	g := newCatalogGame(t)
	id := tmpB1Push(g, g.Seats[0].ID, "Mawcor", "Creature — Beast", tmpB2MawcorOracle, 3, 3, "U")
	if !hasEffectiveKeyword(t, g, id, "flying") {
		t.Error("Mawcor does not have flying")
	}
}

// --- Dismiss and Diabolic Edict -----------------------------------------------------

func TestTmpB2DismissCountersASpellAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := batch01OpponentCasts(t, g, opp, "Lightning Bolt", lightningBoltOracle, "", b16TargetPlayer(me.ID))
	castCatalogSpell(t, g, "Dismiss", "Instant", tmpB2DismissOracle, b16TargetCard(bolt))
	hand := me.Hand.Size()
	life := me.Life
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(bolt) || me.Life != life {
		t.Fatal("the Bolt was not countered")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn", hand, got)
	}
}

func TestTmpB2DiabolicEdictMakesTheTargetSacrificeACreatureOfTheirChoice(t *testing.T) {
	g := newCatalogGame(t)
	_, opp, _ := blbB1Seats(g)
	a := blbB1Push(g, opp.ID, "Bear A", "Creature — Bear", "", 2, 2)
	b := blbB1Push(g, opp.ID, "Bear B", "Creature — Bear", "", 2, 2)
	castCatalogSpell(t, g, "Diabolic Edict", "Instant", tmpB2EdictOracle, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	if sacrificeChoiceFor(g, opp.ID) == nil {
		t.Fatal("the target player was not asked to sacrifice a creature")
	}
	answerSacrifice(t, g, opp.ID, b)
	if g.Battlefield.Contains(b) || !g.Battlefield.Contains(a) {
		t.Error("the chosen creature was not the one sacrificed")
	}
}
