package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// choreographed_sparks_test.go — #4688, ADR 0218. The second bullet's
// rider is gained as part of the copying process (CR 707.9a), so it is a
// copiable value of the copy and of the token the copy becomes (CR
// 608.3f): the token is sacrificed by its OWN trigger, which an
// ability-removal effect takes away and a Clone of the token copies.

const choreographedSparksOracle = "61495793-b0c7-40a5-ad3e-9df8e2b8096b"

// sparksCast casts Choreographed Sparks from the active seat with the
// given modes and targets, returning the cast error rather than failing.
func sparksCast(t *testing.T, g *game.Game, modes []int, targets []game.TargetRef) (uuid.UUID, error) {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Choreographed Sparks", TypeLine: "Instant", ManaCost: "{R}{R}",
		OracleID: choreographedSparksOracle, Owner: active.ID, Controller: active.ID,
	})
	advanceToMainOf(t, g, g.Turn.ActiveSeat)
	return id, g.CastSpell(active.ID, id, game.CastSpellParams{Modes: modes, Targets: targets})
}

func mustSparksCast(t *testing.T, g *game.Game, modes []int, targets []game.TargetRef) uuid.UUID {
	t.Helper()
	id, err := sparksCast(t, g, modes, targets)
	if err != nil {
		t.Fatalf("cast Choreographed Sparks: %v", err)
	}
	return id
}

// sparksBearToken casts a Grizzly Bears, copies it with the second bullet
// and resolves everything, returning the token's and the card's IDs.
func sparksBearToken(t *testing.T, g *game.Game) (token, card uuid.UUID) {
	t.Helper()
	spell := castCreatureSpellForCopy(t, g, "Grizzly Bears", "Creature — Bear", "oracle-bears", 2, 2)
	mustSparksCast(t, g, []int{1}, []game.TargetRef{{Kind: game.TargetCard, ID: spell}})
	passPriorityAroundTable(t, g)
	tokens, cards := copiesOnBattlefield(g, "Grizzly Bears")
	if len(tokens) != 1 || len(cards) != 1 {
		t.Fatalf("got %d tokens and %d cards, want one of each", len(tokens), len(cards))
	}
	return tokens[0].InstanceID, cards[0].InstanceID
}

func sparksOnBattlefield(g *game.Game, id uuid.UUID) bool {
	_, ok := findBattlefieldByID(g, id)
	return ok
}

func TestChoreographedSparksIsRegistered(t *testing.T) {
	spec, ok := Lookup(choreographedSparksOracle)
	if !ok || spec.Name != "Choreographed Sparks" {
		t.Fatalf("Choreographed Sparks is not registered (%v, %q)", ok, spec.Name)
	}
	if spec.Completeness != CompletenessFull || !spec.CantBeCopied {
		t.Errorf("completeness %v, can't be copied %v", spec.Completeness, spec.CantBeCopied)
	}
}

// The headline: the token has haste and is sacrificed at the next end
// step by its own trigger; the original card stays.
func TestChoreographedSparksTokenHasHasteAndIsSacrificedAtTheEndStep(t *testing.T) {
	g := newCatalogGame(t)
	token, card := sparksBearToken(t, g)
	if !effectiveAbilitiesContain(t, g, token, "haste") {
		t.Errorf("the token's abilities are %v, want haste", effectiveAbilities(t, g, token))
	}
	if effectiveAbilitiesContain(t, g, card, "haste") {
		t.Error("the ORIGINAL creature gained haste — the rider is the copy's")
	}
	tok, _ := findBattlefieldByID(g, token)
	if len(tok.GrantedAbilities) != 1 {
		t.Fatalf("the token's granted abilities are %v, want the sacrifice trigger", tok.GrantedAbilities)
	}

	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if sparksOnBattlefield(g, token) {
		t.Fatal("the token is still on the battlefield after the end step began")
	}
	if !sacrificedInLog(g, token) {
		t.Error("the token left without being sacrificed")
	}
	if !sparksOnBattlefield(g, card) {
		t.Error("the original Grizzly Bears was sacrificed too")
	}
}

// The rider is an ability of the token, not a delayed trigger of the
// spell: a token that has lost all its abilities is not sacrificed.
func TestChoreographedSparksTokenThatLosesItsAbilitiesIsNotSacrificed(t *testing.T) {
	g := newCatalogGame(t)
	token, _ := sparksBearToken(t, g)
	g.WithWriteLock(func() {
		g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(token),
			[]game.Mod{game.LoseAllAbilitiesMod()}, g.UntilEndOfTurnDuration(), "loses all abilities")
	})
	if effectiveAbilitiesContain(t, g, token, "haste") {
		t.Error("the token kept haste after losing all abilities")
	}
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if !sparksOnBattlefield(g, token) {
		t.Fatal("a token with no abilities was sacrificed — the rider was not the token's own ability")
	}
}

// CR 707.9a: haste and the trigger are copiable values of the token, so
// a Clone copying it has both and is sacrificed at the end step too.
func TestChoreographedSparksCloneOfTheTokenGetsHasteAndTheTrigger(t *testing.T) {
	g := newCatalogGame(t)
	token, card := sparksBearToken(t, g)

	cloneID := castCatalogSpell(t, g, "Clone", "Creature — Shapeshifter", oracleClone, nil)
	resolveWithCopyChoice(t, g, token)
	clone := copyBattlefieldCard(t, g, cloneID)
	if clone.Name != "Grizzly Bears" {
		t.Fatalf("the Clone is %q, want a copy of the token", clone.Name)
	}
	if clone.IsToken() {
		t.Error("the Clone became a token — copying a token does not copy the Token supertype")
	}
	if !effectiveAbilitiesContain(t, g, cloneID, "haste") {
		t.Errorf("the Clone's abilities are %v, want the copied haste", effectiveAbilities(t, g, cloneID))
	}
	if len(clone.GrantedAbilities) != 1 {
		t.Fatalf("the Clone's granted abilities are %v, want the copied sacrifice trigger", clone.GrantedAbilities)
	}

	advanceToEndStep(t, g)
	// Two triggers at once, the token's and the Clone's.
	for i := 0; i < 8; i++ {
		passPriorityAroundTable(t, g)
		c := pendingOfKind(g, game.PendingChoiceTriggerOrder)
		if c == nil {
			break
		}
		if err := g.ResolveTriggerOrder(c.ID, c.Chooser, append([]uuid.UUID(nil), c.TriggerOrderIDs...)); err != nil {
			t.Fatalf("ResolveTriggerOrder: %v", err)
		}
	}
	if sparksOnBattlefield(g, cloneID) || !sacrificedInLog(g, cloneID) {
		t.Error("the Clone of the token was not sacrificed by its copied trigger")
	}
	if sparksOnBattlefield(g, token) {
		t.Error("the token itself was not sacrificed")
	}
	if !sparksOnBattlefield(g, card) {
		t.Error("the original card was sacrificed")
	}
}

// The first bullet: copy an instant you control, with new targets.
func TestChoreographedSparksCopiesAnInstantWithNewTargets(t *testing.T) {
	g := newCatalogGame(t)
	me, victimA, victimB := g.Seats[0].ID, g.Seats[1].ID, g.Seats[2].ID
	bolt := castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victimA}})
	mustSparksCast(t, g, []int{0}, []game.TargetRef{{Kind: game.TargetCard, ID: bolt, Mode: 0}})

	for i := 0; i < 8 && latestPickTarget(g, me) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	prompt := latestPickTarget(g, me)
	if prompt == nil {
		t.Fatal("no re-target prompt for the copy")
	}
	if err := g.ResolvePickTarget(prompt.ID, me, game.TargetRef{Kind: game.TargetPlayer, ID: victimB}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, victimA); got != 37 {
		t.Errorf("original target life = %d, want 37", got)
	}
	if got := lifeOf(g, victimB); got != 37 {
		t.Errorf("re-targeted copy's target life = %d, want 37", got)
	}
}

// Both bullets at once, each with its own target group, resolving in
// printed order.
func TestChoreographedSparksBothModes(t *testing.T) {
	g := newCatalogGame(t)
	me, victimA := g.Seats[0].ID, g.Seats[1].ID
	bears := castCreatureSpellForCopy(t, g, "Grizzly Bears", "Creature — Bear", "oracle-bears", 2, 2)
	bolt := castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victimA}})
	mustSparksCast(t, g, []int{0, 1}, []game.TargetRef{
		{Kind: game.TargetCard, ID: bolt, Mode: 0},
		{Kind: game.TargetCard, ID: bears, Mode: 1},
	})
	for i := 0; i < 32 && !stackFullyEmpty(g); i++ {
		if p := latestPickTarget(g, me); p != nil {
			if err := g.ResolvePickTarget(p.ID, me, game.TargetRef{Kind: game.TargetPlayer, ID: victimA}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			continue
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if got := lifeOf(g, victimA); got != 34 {
		t.Errorf("life = %d, want 34 — the Bolt and its copy", got)
	}
	tokens, cards := copiesOnBattlefield(g, "Grizzly Bears")
	if len(tokens) != 1 || len(cards) != 1 {
		t.Fatalf("got %d tokens and %d cards, want one of each", len(tokens), len(cards))
	}
	if !effectiveAbilitiesContain(t, g, tokens[0].InstanceID, "haste") || len(tokens[0].GrantedAbilities) != 1 {
		t.Error("the creature copy did not gain haste and the trigger")
	}
}

// "This spell can't be copied": a Twincast aimed at Sparks makes nothing.
func TestChoreographedSparksCantBeCopied(t *testing.T) {
	g := newCatalogGame(t)
	bolt := castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}})
	sparks := mustSparksCast(t, g, []int{0}, []game.TargetRef{{Kind: game.TargetCard, ID: bolt, Mode: 0}})
	castCatalogSpell(t, g, "Twincast", "Instant", twincastOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: sparks}})
	// Twincast resolves first and copies nothing.
	for i := 0; i < len(g.Seats); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if n := stackCopies(g); n != 0 {
		t.Errorf("%d copies on the stack after Twincast resolved, want none", n)
	}
}

// The targets: "you control", instant-or-sorcery for the first bullet,
// creature for the second.
func TestChoreographedSparksRefusesIllegalTargets(t *testing.T) {
	t.Run("an opponent's spell", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		advanceToMainOf(t, g, g.Turn.ActiveSeat)
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
		bolt := castFor(t, g, opp, "Lightning Bolt", "Instant", lightningBoltOracle,
			game.CastSpellParams{Targets: rtPlayerRef(me.ID)})
		if _, err := sparksCast(t, g, []int{0}, []game.TargetRef{{Kind: game.TargetCard, ID: bolt, Mode: 0}}); err == nil {
			t.Error("Sparks copied an opponent's instant")
		}
	})
	t.Run("a creature spell for the first bullet", func(t *testing.T) {
		g := newCatalogGame(t)
		bears := castCreatureSpellForCopy(t, g, "Grizzly Bears", "Creature — Bear", "oracle-bears", 2, 2)
		if _, err := sparksCast(t, g, []int{0}, []game.TargetRef{{Kind: game.TargetCard, ID: bears, Mode: 0}}); err == nil {
			t.Error("the instant-or-sorcery bullet accepted a creature spell")
		}
	})
	t.Run("an instant for the second bullet", func(t *testing.T) {
		g := newCatalogGame(t)
		bolt := castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
			[]game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}})
		if _, err := sparksCast(t, g, []int{1}, []game.TargetRef{{Kind: game.TargetCard, ID: bolt}}); err == nil {
			t.Error("the creature bullet accepted an instant")
		}
	})
}

// The rider is data in the copy's copiable values, so it survives a
// capture → JSON → restore both while the copy is on the stack and once
// it is a token.
func TestChoreographedSparksRiderSurvivesARestore(t *testing.T) {
	t.Run("the copy on the stack", func(t *testing.T) {
		g := newCatalogGame(t)
		spell := castCreatureSpellForCopy(t, g, "Grizzly Bears", "Creature — Bear", "oracle-bears", 2, 2)
		mustSparksCast(t, g, []int{1}, []game.TargetRef{{Kind: game.TargetCard, ID: spell}})
		for i := 0; i < 8 && stackCopies(g) == 0; i++ {
			if err := g.PassPriority(); err != nil {
				t.Fatalf("PassPriority: %v", err)
			}
		}
		if stackCopies(g) != 1 {
			t.Fatal("setup: the copy is not on the stack")
		}
		g = restoreRoundTrip(t, g, true)
		passPriorityAroundTable(t, g)
		tokens, _ := copiesOnBattlefield(g, "Grizzly Bears")
		if len(tokens) != 1 {
			t.Fatalf("token copies = %d after the restore, want 1", len(tokens))
		}
		token := tokens[0].InstanceID
		if !effectiveAbilitiesContain(t, g, token, "haste") || len(tokens[0].GrantedAbilities) != 1 {
			t.Fatal("the restored copy's token lost haste or the trigger")
		}
		advanceToEndStep(t, g)
		passPriorityAroundTable(t, g)
		if sparksOnBattlefield(g, token) {
			t.Error("the restored copy's token was not sacrificed at the end step")
		}
	})
	t.Run("the token", func(t *testing.T) {
		g := newCatalogGame(t)
		token, _ := sparksBearToken(t, g)
		g = restoreRoundTrip(t, g, true)
		if !effectiveAbilitiesContain(t, g, token, "haste") {
			t.Error("the restored token lost haste")
		}
		advanceToEndStep(t, g)
		passPriorityAroundTable(t, g)
		if sparksOnBattlefield(g, token) {
			t.Error("the restored token was not sacrificed at the end step")
		}
	})
}
