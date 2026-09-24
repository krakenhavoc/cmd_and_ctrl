package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ninjutsu_test.go — the catalog half of #1227: ninjutsu (CR 702.49)
// on Ninja of the Deep Hours, Ingenious Infiltrator, Moonblade Shinobi
// and Prosperous Thief. The engine half — the unblocked-attacker read,
// the attacking entry and the paid-cost record — is in
// game/attacking_entry_test.go.

const (
	ninjaOfTheDeepHoursOracle = "1f3c2b00-0000-4ae1-9650-9553accac52e"
	ingeniousInfiltratorOracl = "354defd2-f63f-4a48-9fd4-2526b3878a84"
	moonbladeShinobiOracle    = "046e25ee-c96d-4cff-93be-f7e3379d713c"
	prosperousThiefOracle     = "6d03971a-8365-449a-8fbe-07b5b9bb42dc"
)

// pushNinjaToHand seeds a catalog card into a player's hand WITH a
// body — the zone every ninjutsu ability functions from (CR 702.49a),
// and the power that has to connect once the ninja is attacking.
// pushCatalogHandCard (cycling_test.go) is the P/T-free sibling.
func pushNinjaToHand(p *game.Player, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness,
		Owner: p.ID, Controller: p.ID,
		KnownBy: map[uuid.UUID]bool{p.ID: true},
	})
	return id
}

// ninjutsuBoard sets a combat up the way the keyword wants it: one
// attacker of mine swinging at seat 1, the named ninja in my hand, the
// cursor in the declare-blockers step and `mana` floating. Returns the
// ninja, the attacker and the two seats.
func ninjutsuBoard(t *testing.T, g *game.Game, name, oracle, mana string) (ninja, attacker uuid.UUID, me, opp *game.Player) {
	t.Helper()
	me, opp = g.Seats[0], g.Seats[1]
	attacker = b12Creature(g, me.ID, "Sneaky Rat", "Creature — Rat", 1, 1)
	ninja = pushNinjaToHand(me, name, "Creature — Human Ninja", oracle, 2, 2)
	declareAttack(t, g, opp.ID, attacker)
	advanceTo(t, g, game.StepDeclareBlockers)
	if mana != "" {
		if err := g.AddManaForEffect(me.ID, uuid.Nil, mana); err != nil {
			t.Fatalf("AddManaForEffect: %v", err)
		}
	}
	return ninja, attacker, me, opp
}

// attackEventsFor counts the EventAttack events naming `id` — the
// CR 506.3c assertion, since a permanent PUT onto the battlefield
// attacking was never declared and must announce nothing.
func attackEventsFor(g *game.Game, id uuid.UUID) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventAttack && ev.CardID == id {
			n++
		}
	}
	return n
}

// Every ninjutsu card declares the same hand ability with the same
// cost shape — the boot-time invariant, asserted here so a card file
// that hand-rolls the keyword instead of calling Ninjutsu() fails in
// this package rather than at server start.
func TestNinjutsuCardsDeclareAHandAbilityWithTheReturnCost(t *testing.T) {
	for _, tc := range []struct{ name, oracle, mana string }{
		{"Ninja of the Deep Hours", ninjaOfTheDeepHoursOracle, "{1}{U}"},
		{"Ingenious Infiltrator", ingeniousInfiltratorOracl, "{U}{B}"},
		{"Moonblade Shinobi", moonbladeShinobiOracle, "{2}{U}"},
		{"Prosperous Thief", prosperousThiefOracle, "{1}{U}"},
	} {
		abilities := game.ActivatedAbilitiesForCard(game.Card{OracleID: tc.oracle})
		if len(abilities) != 1 {
			t.Errorf("%s: %d activated abilities, want the one ninjutsu ability", tc.name, len(abilities))
			continue
		}
		ab := abilities[0]
		if ab.Cost.Mana != tc.mana {
			t.Errorf("%s: cost %q, want %q", tc.name, ab.Cost.Mana, tc.mana)
		}
		if ab.Cost.ReturnToHand.Empty() {
			t.Errorf("%s: no return-to-hand cost component", tc.name)
		} else if n := ab.Cost.ReturnToHand.Count; n != 1 {
			t.Errorf("%s: returns %d permanents, want 1", tc.name, n)
		}
		if !game.AbilityFunctionsFromZone(ab, game.ZoneHand) {
			t.Errorf("%s: the ability does not function from the hand (CR 702.49a)", tc.name)
		}
		if game.AbilityFunctionsFromZone(ab, game.ZoneBattlefield) {
			t.Errorf("%s: the ability is offered on the battlefield too", tc.name)
		}
		if ab.SorcerySpeed {
			t.Errorf("%s: sorcery-speed — ninjutsu is activated in the declare-blockers step", tc.name)
		}
		if why := game.AbilityNeedsPermanentSource(ab.Cost); why != "" {
			t.Errorf("%s: declares %s, which needs a permanent on the battlefield", tc.name, why)
		}
	}
}

// The headline. The unblocked attacker goes back to hand as the cost,
// and the ninja arrives from hand tapped and attacking the player the
// returned creature was attacking (CR 702.49a).
func TestNinjutsuReturnsTheAttackerAndTheNinjaArrivesAttacking(t *testing.T) {
	g := newCatalogGame(t)
	ninja, attacker, me, opp := ninjutsuBoard(t, g, "Ninja of the Deep Hours", ninjaOfTheDeepHoursOracle, "{U}{U}")

	if err := g.ActivateCatalogAbility(me.ID, ninja, 0, game.ActivateAbilityParams{
		ReturnIDs: []uuid.UUID{attacker},
	}); err != nil {
		t.Fatalf("activate ninjutsu: %v", err)
	}
	if !me.Hand.Contains(attacker) {
		t.Error("the unblocked attacker was not returned to hand — the cost is paid at announce")
	}
	if !me.Hand.Contains(ninja) {
		t.Error("the ninja left the hand at announce; it enters on RESOLUTION")
	}
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(ninja) {
		t.Fatal("the ninja is not on the battlefield")
	}
	c, ok := g.LookupCardForEffect(ninja)
	if !ok {
		t.Fatal("the ninja cannot be looked up")
	}
	if !c.Tapped {
		t.Error("the ninja entered untapped — ninjutsu puts it onto the battlefield TAPPED")
	}
	if c.AttackingTarget != opp.ID {
		t.Errorf("the ninja is attacking %s, want %s — the same player the returned creature was attacking (CR 702.49a)",
			c.AttackingTarget, opp.ID)
	}
}

// CR 506.3c: the ninja was never DECLARED as an attacker, so nothing
// that watches attack declarations may see it — no EventAttack, and
// the declaration's lock-in must not pick its AttackingTarget up as a
// staged declaration next time it runs.
func TestNinjutsuNinjaIsNeverDeclaredAsAnAttacker(t *testing.T) {
	g := newCatalogGame(t)
	ninja, attacker, me, _ := ninjutsuBoard(t, g, "Ninja of the Deep Hours", ninjaOfTheDeepHoursOracle, "{U}{U}")

	if err := g.ActivateCatalogAbility(me.ID, ninja, 0, game.ActivateAbilityParams{
		ReturnIDs: []uuid.UUID{attacker},
	}); err != nil {
		t.Fatalf("activate ninjutsu: %v", err)
	}
	// The resolution that put the ninja onto the battlefield ended in
	// a state-check pass, which is where commitAttackDeclarationLocked
	// runs — so the lock-in has already had its chance to announce it.
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(ninja) {
		t.Fatal("the ninja is not on the battlefield")
	}
	if n := attackEventsFor(g, ninja); n != 0 {
		t.Errorf("%d EventAttack for the ninja, want 0 (CR 506.3c)", n)
	}
	if n := attackEventsFor(g, attacker); n != 1 {
		t.Errorf("%d EventAttack for the creature that WAS declared, want 1", n)
	}
}

// The timing restriction, and it is the COST rather than a Condition:
// in the declare-attackers step no creature is an unblocked attacker
// yet (CR 509.1h), so nothing can pay.
func TestNinjutsuIsRefusedBeforeBlockersAreDeclared(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := b12Creature(g, me.ID, "Sneaky Rat", "Creature — Rat", 1, 1)
	ninja := pushNinjaToHand(me, "Ninja of the Deep Hours", "Creature — Human Ninja", ninjaOfTheDeepHoursOracle, 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{U}{U}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}

	err := g.ActivateCatalogAbility(me.ID, ninja, 0, game.ActivateAbilityParams{
		ReturnIDs: []uuid.UUID{attacker},
	})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("activate in declare_attackers: %v, want ErrIllegalTarget — nothing is unblocked yet", err)
	}
	if !g.Battlefield.Contains(attacker) {
		t.Error("the refused activation returned the attacker anyway")
	}
	if !me.Hand.Contains(ninja) {
		t.Error("the refused activation moved the ninja")
	}
}

// And a BLOCKED attacker cannot pay either, which is the clause's
// whole point: ninjutsu rewards getting through, not swinging.
func TestNinjutsuRefusesABlockedAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := b12Creature(g, me.ID, "Sneaky Rat", "Creature — Rat", 1, 1)
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	ninja := pushNinjaToHand(me, "Ninja of the Deep Hours", "Creature — Human Ninja", ninjaOfTheDeepHoursOracle, 2, 2)
	declareAttack(t, g, opp.ID, attacker)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	lockInBlocks(t, g)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{U}{U}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}

	err := g.ActivateCatalogAbility(me.ID, ninja, 0, game.ActivateAbilityParams{
		ReturnIDs: []uuid.UUID{attacker},
	})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("activate naming a blocked attacker: %v, want ErrIllegalTarget", err)
	}
}

// #1279: while the defending player is still DECLARING — they have a
// creature that could block and have not finished — the attacker is
// neither blocked nor unblocked, so nothing can pay. Before #1279 this
// was the window in which ninjutsu was available a beat early. Once the
// defender finishes with no block, it pays.
func TestNinjutsuWaitsForTheDefenderToFinishDeclaring(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := b12Creature(g, me.ID, "Sneaky Rat", "Creature — Rat", 1, 1)
	b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	ninja := pushNinjaToHand(me, "Ninja of the Deep Hours", "Creature — Human Ninja", ninjaOfTheDeepHoursOracle, 2, 2)
	declareAttack(t, g, opp.ID, attacker)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{U}{U}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if got := g.BlockDeclarationStatusOf(opp.ID); got != game.BlockDeclarationPending {
		t.Fatalf("setup: the defender should still be declaring, status %q", got)
	}

	err := g.ActivateCatalogAbility(me.ID, ninja, 0, game.ActivateAbilityParams{
		ReturnIDs: []uuid.UUID{attacker},
	})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("activate while the defender is declaring: %v, want ErrIllegalTarget", err)
	}
	if !g.Battlefield.Contains(attacker) {
		t.Fatal("the refused activation returned the attacker anyway")
	}

	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Fatalf("FinishBlocks: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, ninja, 0, game.ActivateAbilityParams{
		ReturnIDs: []uuid.UUID{attacker},
	}); err != nil {
		t.Fatalf("activate once the defender declared none: %v", err)
	}
	if !me.Hand.Contains(attacker) {
		t.Error("the unblocked attacker was not returned as the cost")
	}
}

// The ninja is unblocked when it arrives, so it connects in the combat
// damage step — and Ninja of the Deep Hours' own trigger draws.
func TestNinjaOfTheDeepHoursConnectsAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	ninja, attacker, me, opp := ninjutsuBoard(t, g, "Ninja of the Deep Hours", ninjaOfTheDeepHoursOracle, "{U}{U}")

	if err := g.ActivateCatalogAbility(me.ID, ninja, 0, game.ActivateAbilityParams{
		ReturnIDs: []uuid.UUID{attacker},
	}); err != nil {
		t.Fatalf("activate ninjutsu: %v", err)
	}
	passPriorityAroundTable(t, g)
	life, hand := opp.Life, me.Hand.Size()
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)

	if got := life - opp.Life; got != 2 {
		t.Errorf("defender lost %d life, want 2 — the ninja arrived unblocked", got)
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("hand changed by %d, want +1 (the ninja connected and drew)", got)
	}
}

// Ingenious Infiltrator's "whenever a Ninja you control deals combat
// damage to a player" fires off its OWN arrival — it is a Ninja.
func TestIngeniousInfiltratorDrawsOffItsOwnNinjutsu(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := b12Creature(g, me.ID, "Sneaky Rat", "Creature — Rat", 1, 1)
	ninja := pushNinjaToHand(me, "Ingenious Infiltrator", "Creature — Vedalken Ninja", ingeniousInfiltratorOracl, 2, 3)
	declareAttack(t, g, opp.ID, attacker)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{U}{B}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}

	if err := g.ActivateCatalogAbility(me.ID, ninja, 0, game.ActivateAbilityParams{
		ReturnIDs: []uuid.UUID{attacker},
	}); err != nil {
		t.Fatalf("activate ninjutsu: %v", err)
	}
	passPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)

	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("hand changed by %d, want +1 — a Ninja you control connected", got)
	}
}

// Moonblade Shinobi's payoff is a body: the Illusion arrives in the
// combat damage step, after blockers, so it is a flier for next turn.
func TestMoonbladeShinobiCreatesAnIllusionWhenItConnects(t *testing.T) {
	g := newCatalogGame(t)
	ninja, attacker, me, _ := ninjutsuBoard(t, g, "Moonblade Shinobi", moonbladeShinobiOracle, "{U}{U}{U}")

	if err := g.ActivateCatalogAbility(me.ID, ninja, 0, game.ActivateAbilityParams{
		ReturnIDs: []uuid.UUID{attacker},
	}); err != nil {
		t.Fatalf("activate ninjutsu: %v", err)
	}
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)

	if n := countTokensControlled(g, me.ID, "Illusion"); n != 1 {
		t.Errorf("%d Illusion tokens, want 1", n)
	}
}

// Prosperous Thief's trigger is a BATCH — "one or more Ninja or Rogue
// creatures you control deal combat damage to a player" is one trigger
// per player, however many of them got there.
func TestProsperousThiefMakesOneTreasurePerPlayerConnected(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rogue := b12Creature(g, me.ID, "Sneaky Rogue", "Creature — Human Rogue", 1, 1)
	decoy := b12Creature(g, me.ID, "Decoy", "Creature — Rat", 1, 1)
	ninja := pushNinjaToHand(me, "Prosperous Thief", "Creature — Human Ninja", prosperousThiefOracle, 3, 2)
	declareAttack(t, g, opp.ID, rogue, decoy)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{U}{U}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}

	// The decoy goes home to pay, so the Rogue and the Thief both
	// connect with the same player — one Treasure, not two.
	if err := g.ActivateCatalogAbility(me.ID, ninja, 0, game.ActivateAbilityParams{
		ReturnIDs: []uuid.UUID{decoy},
	}); err != nil {
		t.Fatalf("activate ninjutsu: %v", err)
	}
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)

	if n := countTokensControlled(g, me.ID, "Treasure"); n != 1 {
		t.Errorf("%d Treasure tokens, want 1 — the clause is one trigger per player connected with", n)
	}
}
