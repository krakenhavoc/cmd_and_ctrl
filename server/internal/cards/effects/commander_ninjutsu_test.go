package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// commander_ninjutsu_test.go — #1278: commander ninjutsu (CR 702.49c)
// on Yuriko, the Tiger's Shadow. Plain ninjutsu is ninjutsu_test.go;
// the enumerator half is legal/ninjutsu_test.go and the wire half
// protocol/ninjutsu_view_test.go.

const yurikoOracle = "a7043fbd-1dfd-42cf-be4b-cc343d0949e5"

// yurikoCard is Yuriko with the body and the designation the tests
// need, for whichever zone the caller pushes her into.
func yurikoCard(owner uuid.UUID) game.Card {
	return game.Card{
		InstanceID: uuid.New(),
		Name:       "Yuriko, the Tiger's Shadow",
		TypeLine:   "Legendary Creature — Human Ninja",
		OracleID:   yurikoOracle,
		ManaCost:   "{1}{U}{B}",
		Power:      1, Toughness: 3,
		Owner: owner, Controller: owner,
		IsCommander: true,
	}
}

// yurikoInCommandZone seeds Yuriko into her owner's command zone as a
// commander that has already been cast `casts` times, so a tax that
// wrongly applied would have something to charge.
func yurikoInCommandZone(p *game.Player, casts int) uuid.UUID {
	c := yurikoCard(p.ID)
	p.Command.PushTop(c)
	if casts > 0 {
		p.CommanderCasts[c.InstanceID] = casts
	}
	return c.InstanceID
}

// yurikoBoard is ninjutsuBoard with Yuriko in the command zone: one
// attacker of mine swinging at `defender`, the cursor in declare
// blockers and exactly {U}{B} floating.
func yurikoBoard(t *testing.T, g *game.Game, defender func(me, opp *game.Player) uuid.UUID) (yuriko, attacker uuid.UUID, me, opp *game.Player) {
	t.Helper()
	me, opp = g.Seats[0], g.Seats[1]
	attacker = b12Creature(g, me.ID, "Sneaky Rat", "Creature — Rat", 1, 1)
	yuriko = yurikoInCommandZone(me, 2)
	declareAttack(t, g, defender(me, opp), attacker)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{U}{B}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	return yuriko, attacker, me, opp
}

func atThePlayer(_, opp *game.Player) uuid.UUID { return opp.ID }

// The keyword's shape: the ninjutsu cost, functioning from the hand
// AND the command zone and nowhere else (CR 702.49c, CR 113.6).
func TestCommanderNinjutsuFunctionsFromHandAndCommandZone(t *testing.T) {
	abilities := game.ActivatedAbilitiesForCard(game.Card{OracleID: yurikoOracle})
	if len(abilities) != 1 {
		t.Fatalf("%d activated abilities, want the one commander ninjutsu ability", len(abilities))
	}
	ab := abilities[0]
	if ab.Cost.Mana != "{U}{B}" {
		t.Errorf("cost %q, want {U}{B}", ab.Cost.Mana)
	}
	if ab.Cost.ReturnToHand.Empty() || ab.Cost.ReturnToHand.Count != 1 {
		t.Error("no return-an-unblocked-attacker component")
	}
	for zone, want := range map[game.ZoneKind]bool{
		game.ZoneHand:        true,
		game.ZoneCommand:     true,
		game.ZoneBattlefield: false,
		game.ZoneGraveyard:   false,
	} {
		if got := game.AbilityFunctionsFromZone(ab, zone); got != want {
			t.Errorf("functions from %s = %v, want %v", zone, got, want)
		}
	}
	// And plain ninjutsu did not grow the zone with it.
	plain := game.ActivatedAbilitiesForCard(game.Card{OracleID: ninjaOfTheDeepHoursOracle})[0]
	if game.AbilityFunctionsFromZone(plain, game.ZoneCommand) {
		t.Error("plain ninjutsu functions from the command zone — only CR 702.49c says it may")
	}
}

// The headline. From the command zone, in declare blockers, with an
// unblocked attacker: the attacker goes home, Yuriko arrives tapped
// and attacking the same player, and the commander tax neither applies
// nor grows — she was PUT, not cast (CR 903.8).
func TestCommanderNinjutsuFromTheCommandZoneEntersAttackingUntaxed(t *testing.T) {
	g := newCatalogGame(t)
	yuriko, attacker, me, opp := yurikoBoard(t, g, atThePlayer)

	// Exactly {U}{B} is floating and she has been cast twice, so a
	// {4} tax would make the activation unpayable.
	if err := g.ActivateCatalogAbility(me.ID, yuriko, 0, game.ActivateAbilityParams{
		ReturnIDs: []uuid.UUID{attacker},
	}); err != nil {
		t.Fatalf("activate commander ninjutsu from the command zone: %v", err)
	}
	if !me.Hand.Contains(attacker) {
		t.Error("the unblocked attacker was not returned to hand")
	}
	if !me.Command.Contains(yuriko) {
		t.Error("Yuriko left the command zone at announce; she enters on RESOLUTION")
	}
	passPriorityAroundTable(t, g)

	if me.Command.Contains(yuriko) {
		t.Fatal("Yuriko is still in the command zone")
	}
	c, ok := g.LookupCardForEffect(yuriko)
	if !ok || !g.Battlefield.Contains(yuriko) {
		t.Fatal("Yuriko is not on the battlefield")
	}
	if !c.Tapped {
		t.Error("Yuriko entered untapped — CR 702.49c puts her onto the battlefield TAPPED")
	}
	if c.AttackingTarget != opp.ID {
		t.Errorf("Yuriko is attacking %s, want %s (the returned creature's defender)", c.AttackingTarget, opp.ID)
	}
	if c.Controller != me.ID {
		t.Errorf("Yuriko entered under %s, want her owner %s", c.Controller, me.ID)
	}
	if !c.IsCommander {
		t.Error("Yuriko lost the commander designation on a non-cast exit")
	}
	if got := me.CommanderCasts[yuriko]; got != 2 {
		t.Errorf("CommanderCasts = %d, want 2 unchanged — putting a commander is not casting it (CR 903.8)", got)
	}
	if n := attackEventsFor(g, yuriko); n != 0 {
		t.Errorf("%d EventAttack for Yuriko, want 0 (CR 506.3c)", n)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("%d pending choices after the entry — CR 903.9 is about leaving for a library, hand, graveyard or exile, not arriving on the battlefield",
			len(g.PendingChoices))
	}
}

// CR 702.49c: "the same player OR PLANESWALKER" — the defender comes
// off the returned creature, so a swing at a planeswalker lands her on
// the planeswalker.
func TestCommanderNinjutsuEntersAttackingTheSamePlaneswalker(t *testing.T) {
	g := newCatalogGame(t)
	var walker uuid.UUID
	yuriko, attacker, me, _ := yurikoBoard(t, g, func(_, opp *game.Player) uuid.UUID {
		walker = pushWalkerForTest(g, opp.ID, "Their Walker", "", 4)
		return walker
	})

	if err := g.ActivateCatalogAbility(me.ID, yuriko, 0, game.ActivateAbilityParams{
		ReturnIDs: []uuid.UUID{attacker},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)

	c, ok := g.LookupCardForEffect(yuriko)
	if !ok || !g.Battlefield.Contains(yuriko) {
		t.Fatal("Yuriko is not on the battlefield")
	}
	if c.AttackingTarget != walker {
		t.Errorf("Yuriko is attacking %s, want the planeswalker %s", c.AttackingTarget, walker)
	}
}

// The timing restriction from the command zone is the same COST it is
// from the hand: in declare attackers nothing is an unblocked attacker
// yet (CR 509.1h), so the activation is refused and nothing moves.
func TestCommanderNinjutsuIsRefusedBeforeBlockersAreDeclared(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := b12Creature(g, me.ID, "Sneaky Rat", "Creature — Rat", 1, 1)
	yuriko := yurikoInCommandZone(me, 0)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{U}{B}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}

	err := g.ActivateCatalogAbility(me.ID, yuriko, 0, game.ActivateAbilityParams{
		ReturnIDs: []uuid.UUID{attacker},
	})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("activate in declare_attackers: %v, want ErrIllegalTarget — nothing is unblocked yet", err)
	}
	if !g.Battlefield.Contains(attacker) {
		t.Error("the refused activation returned the attacker anyway")
	}
	if !me.Command.Contains(yuriko) {
		t.Error("the refused activation moved Yuriko")
	}
}

// Another player cannot activate YOUR commander's ability: the command
// zone's "you" is the card's owner (CR 108.4), exactly as the hand's.
func TestCommanderNinjutsuIsTheOwnersAlone(t *testing.T) {
	g := newCatalogGame(t)
	yuriko, attacker, _, opp := yurikoBoard(t, g, atThePlayer)
	if err := g.AddManaForEffect(opp.ID, uuid.Nil, "{U}{B}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	err := g.ActivateCatalogAbility(opp.ID, yuriko, 0, game.ActivateAbilityParams{
		ReturnIDs: []uuid.UUID{attacker},
	})
	if !errors.Is(err, game.ErrCardCallerMismatch) {
		t.Fatalf("an opponent activating my commander: %v, want ErrCardCallerMismatch", err)
	}
}

// The hand half of CR 702.49c is plain ninjutsu, and still works: a
// commander that went to hand rather than the command zone can be
// ninjutsu'd from there.
func TestCommanderNinjutsuFromTheHand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := b12Creature(g, me.ID, "Sneaky Rat", "Creature — Rat", 1, 1)
	card := yurikoCard(me.ID)
	me.Hand.PushTop(card)
	yuriko := card.InstanceID
	declareAttack(t, g, opp.ID, attacker)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{U}{B}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}

	if err := g.ActivateCatalogAbility(me.ID, yuriko, 0, game.ActivateAbilityParams{
		ReturnIDs: []uuid.UUID{attacker},
	}); err != nil {
		t.Fatalf("activate from hand: %v", err)
	}
	passPriorityAroundTable(t, g)

	c, ok := g.LookupCardForEffect(yuriko)
	if !ok || !g.Battlefield.Contains(yuriko) {
		t.Fatal("Yuriko is not on the battlefield")
	}
	if c.AttackingTarget != opp.ID {
		t.Errorf("Yuriko is attacking %s, want %s", c.AttackingTarget, opp.ID)
	}
}

// CR 400.7 on the way back: a commander discarded in response to her
// own hand ninjutsu takes CR 903.9's command zone — a zone the ability
// ALSO names, with the same instance ID. She is a new object there and
// the ability has lost her; it must not put her onto the battlefield.
func TestCommanderNinjutsuLosesACommanderThatMovedInResponse(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := b12Creature(g, me.ID, "Sneaky Rat", "Creature — Rat", 1, 1)
	card := yurikoCard(me.ID)
	me.Hand.PushTop(card)
	yuriko := card.InstanceID
	declareAttack(t, g, opp.ID, attacker)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{U}{B}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, yuriko, 0, game.ActivateAbilityParams{
		ReturnIDs: []uuid.UUID{attacker},
	}); err != nil {
		t.Fatalf("activate from hand: %v", err)
	}
	// The response: out of the hand and into the command zone, as a
	// discard whose owner took CR 903.9's offer would leave her.
	g.WithWriteLock(func() {
		if _, err := game.MoveCard(me.Hand, me.Command, yuriko); err != nil {
			t.Fatalf("MoveCard: %v", err)
		}
	})
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(yuriko) {
		t.Fatal("Yuriko entered the battlefield although she changed zones in response (CR 400.7)")
	}
	if !me.Command.Contains(yuriko) {
		t.Error("Yuriko left the command zone")
	}
}

// The designation rides her onto the battlefield: she deals COMMANDER
// damage when she connects (CR 903.10a), and her own trigger fires off
// it — she is a Ninja — revealing the top card into hand and draining
// every opponent for its mana value.
func TestYurikoConnectsForCommanderDamageAndFlips(t *testing.T) {
	g := newCatalogGame(t)
	yuriko, attacker, me, opp := yurikoBoard(t, g, atThePlayer)
	top := uuid.New()
	me.Library.PushTop(game.Card{
		InstanceID: top, Name: "Three-Drop", TypeLine: "Sorcery", ManaCost: "{2}{B}",
		Owner: me.ID, Controller: me.ID,
	})

	if err := g.ActivateCatalogAbility(me.ID, yuriko, 0, game.ActivateAbilityParams{
		ReturnIDs: []uuid.UUID{attacker},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)

	if got := opp.CommanderDamage[yuriko]; got != 1 {
		t.Errorf("commander damage from Yuriko = %d, want 1 — the designation survives a non-cast exit", got)
	}
	if !me.Hand.Contains(top) {
		t.Error("the revealed card is not in hand")
	}
	for _, p := range g.Seats {
		want := 3
		switch p.ID {
		case me.ID:
			want = 0
		case opp.ID:
			want = 1 + 3 // combat damage, then the flip
		}
		if got := lives[p.ID] - p.Life; got != want {
			t.Errorf("seat %s lost %d life, want %d", p.Name, got, want)
		}
	}
}

// And the trigger is per Ninja, so a Ninja that is NOT Yuriko sets it
// off too — a land flip costs nobody anything.
func TestYurikoFlipsForAnotherNinjaAndALandDrainsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushBattlefieldCardWithTimestamp(g, yurikoCard(me.ID))
	ninja := b12Creature(g, me.ID, "Some Ninja", "Creature — Human Ninja", 2, 2)
	land := uuid.New()
	me.Library.PushTop(game.Card{
		InstanceID: land, Name: "Swamp", TypeLine: "Basic Land — Swamp",
		Owner: me.ID, Controller: me.ID,
	})
	declareAttack(t, g, opp.ID, ninja)
	life := opp.Life
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)

	if !me.Hand.Contains(land) {
		t.Error("the revealed land is not in hand")
	}
	if got := life - opp.Life; got != 2 {
		t.Errorf("defender lost %d life, want 2 — combat only; a land's mana value is 0", got)
	}
}
