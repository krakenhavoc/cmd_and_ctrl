package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// transform_cards_test.go — the two #1112 cards that prove ADR 0079's
// two verbs, end to end on real imported cards.
//
// # Why these go through deck.ToGameCard
//
// Same reason siege_transform_test.go gives, and it is sharper here.
// A transform card's whole point is `Faces[1]`, which lives on
// Scryfall's `card_faces[1]` and reaches game.Card only through the
// importer. castCatalogSpell builds a flat fixture with no faces at
// all, so a test written on one would exercise a CanTransform that
// always answers false and pass by doing nothing.

// transformRow builds a Scryfall record for a two-faced `transform`
// card in the shape deck.ToGameCard consumes: a null top-level mana
// cost, the joined type line, and the real data on the faces. Verbatim
// how all 401 transform oracle IDs appear in the bulk dump.
func transformRow(oracleID, front, frontType, frontCost, back, backType, power, toughness string, colors []string) cards.Card {
	return cards.Card{
		ID:       uuid.New(),
		Name:     front + " // " + back,
		Layout:   "transform",
		TypeLine: frontType + " // " + backType,
		OracleID: uuid.MustParse(oracleID),
		CardFaces: []cards.CardFace{
			{Name: front, TypeLine: frontType, ManaCost: frontCost, Colors: colors},
			{Name: back, TypeLine: backType, Power: power, Toughness: toughness, Colors: colors},
		},
	}
}

// importToBattlefield runs a row down the import road and puts the
// result onto the battlefield THROUGH the entry pipeline, so the ETB
// hook fires — which is what gives a Saga its CR 714.3 entry lore
// counter. Returns the permanent's instance ID.
func importToBattlefield(t *testing.T, g *game.Game, row cards.Card, p *game.Player) uuid.UUID {
	t.Helper()
	id := importToHand(row, p)
	var entered uuid.UUID
	var err error
	g.WithWriteLock(func() {
		entered, err = g.PutFromHandOntoBattlefieldForEffect(id, game.HandEntryOptions{})
	})
	if err != nil {
		t.Fatalf("put %q onto the battlefield: %v", row.Name, err)
	}
	if entered == uuid.Nil {
		t.Fatalf("%q never reached the battlefield", row.Name)
	}
	return entered
}

// importAndCast runs a row down the import road and CASTS it from the
// active seat's hand, the way castCatalogSpell does for a fixture.
//
// A Saga has to arrive this way rather than through
// importToBattlefield: chapter I triggers off the CR 714.3 entry lore
// counter, and with no spell on the stack the first priority pass
// advances the STEP instead of settling the trigger — which adds a
// second lore counter and leaves two chapters racing for one
// trigger-order prompt.
func importAndCast(t *testing.T, g *game.Game, row cards.Card, p *game.Player) uuid.UUID {
	t.Helper()
	id := importToHand(row, p)
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", row.Name, err)
	}
	return id
}

// pushArtifacts puts n plain artifacts onto a player's battlefield —
// what Storm the Vault's intervening if counts.
func pushArtifacts(g *game.Game, p *game.Player, n int) {
	for i := 0; i < n; i++ {
		pushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(),
			Name:       "Filler Relic",
			TypeLine:   "Artifact",
			Owner:      p.ID,
			Controller: p.ID,
		})
	}
}

func stormTheVaultRow() cards.Card {
	return transformRow(stormTheVaultOracleID,
		"Storm the Vault", "Legendary Enchantment", "{2}{U}{R}",
		"Vault of Catlacan", "Legendary Land", "", "", []string{"U", "R"})
}

func fableRow() cards.Card {
	return transformRow(fableOfTheMirrorBreakerOracleID,
		"Fable of the Mirror-Breaker", "Enchantment — Saga", "{2}{R}",
		"Reflection of Kiki-Jiki", "Enchantment Creature — Goblin Shaman",
		"2", "2", []string{"R"})
}

// assertFaceInvariant checks ADR 0034's invariant on every battlefield
// permanent: the flat printed fields are the active face's. game's own
// AssertFaceInvariant lives in a _test.go file and so is not importable
// from here, and the invariant is exactly what a transform can break.
func assertFaceInvariant(t *testing.T, g *game.Game) {
	t.Helper()
	for _, c := range g.Battlefield.Cards {
		if len(c.Faces) == 0 {
			if c.ActiveFace != 0 {
				t.Errorf("%q has no faces but ActiveFace = %d", c.Name, c.ActiveFace)
			}
			continue
		}
		if c.ActiveFace < 0 || c.ActiveFace >= len(c.Faces) {
			t.Errorf("%q ActiveFace = %d, out of range for %d faces", c.Name, c.ActiveFace, len(c.Faces))
			continue
		}
		f := c.Faces[c.ActiveFace]
		if c.Name != f.Name || c.TypeLine != f.TypeLine || c.ManaCost != f.ManaCost ||
			c.Power != f.Power || c.Toughness != f.Toughness {
			t.Errorf("%q desynchronised from Faces[%d] (%q %q) — something wrote "+
				"ActiveFace without going through SetFace", c.Name, c.ActiveFace, f.Name, f.TypeLine)
		}
	}
}

// --- Storm the Vault // Vault of Catlacan --------------------------

// TestStormTheVaultTransformsAtTheEndStep is the in-place verb's
// card-level proof: an enchantment becomes a land, in its slot,
// keeping its instance ID.
func TestStormTheVaultTransformsAtTheEndStep(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	id := importToBattlefield(t, g, stormTheVaultRow(), me)
	pushArtifacts(g, me, 5)

	advanceToEndStepOf(t, g, seat)
	passPriorityAroundTable(t, g)

	card := battlefieldCardFor(g, id)
	if card == nil {
		t.Fatal("Storm the Vault left the battlefield — a transform is not a zone change")
	}
	if card.ActiveFace != 1 {
		t.Fatalf("ActiveFace = %d, want the back face", card.ActiveFace)
	}
	if card.Name != "Vault of Catlacan" {
		t.Errorf("name = %q, want Vault of Catlacan", card.Name)
	}
	if !card.IsLand() {
		t.Errorf("type line = %q, want a Land", card.TypeLine)
	}
	assertFaceInvariant(t, g)
}

// TestStormTheVaultWithFourArtifactsDoesNotTrigger is CR 603.4's first
// check. Four artifacts is one short, and the ability must not even
// reach the stack — a card that triggered and then did nothing would
// still be observably different, because it would be answerable.
func TestStormTheVaultWithFourArtifactsDoesNotTrigger(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	id := importToBattlefield(t, g, stormTheVaultRow(), me)
	pushArtifacts(g, me, 4)

	advanceToEndStepOf(t, g, seat)
	if item := triggerOnStack(g, id); item != nil {
		t.Fatal("the end-step ability triggered with only four artifacts (CR 603.4)")
	}
	passPriorityAroundTable(t, g)

	if got := battlefieldCardFor(g, id).ActiveFace; got != 0 {
		t.Errorf("ActiveFace = %d, want the front face", got)
	}
}

// TestStormTheVaultLosesTheArtifactsInResponse is CR 603.4's SECOND
// check, and the one an implementation that only gated AppliesTo would
// fail. The trigger goes on the stack with five artifacts; by the time
// it resolves there are four, and it does nothing.
func TestStormTheVaultLosesTheArtifactsInResponse(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	id := importToBattlefield(t, g, stormTheVaultRow(), me)
	pushArtifacts(g, me, 5)

	advanceToEndStepOf(t, g, seat)
	if triggerOnStack(g, id) == nil {
		t.Fatal("the end-step ability did not trigger with five artifacts")
	}
	// Sacrifice one in response — the Treasure play the card invites.
	var victim uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Filler Relic" {
			victim = c.InstanceID
			break
		}
	}
	g.WithWriteLock(func() {
		if err := g.SacrificePermanentForEffect(victim); err != nil {
			t.Fatalf("sacrifice in response: %v", err)
		}
	})
	passPriorityAroundTable(t, g)

	if got := battlefieldCardFor(g, id).ActiveFace; got != 0 {
		t.Errorf("it transformed anyway — CR 603.4's second check did not run")
	}
}

// TestVaultOfCatlacanManaAbilitiesAreLiveAfterTheTransform is the real
// payoff of the "<oracle_id>#1" key: nobody wires the back face's
// abilities up, and they have to be there the instant the face flips.
func TestVaultOfCatlacanManaAbilitiesAreLiveAfterTheTransform(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	id := importToBattlefield(t, g, stormTheVaultRow(), me)
	pushArtifacts(g, me, 5)

	before := game.ManaAbilitiesForCard(*battlefieldCardFor(g, id))
	if len(before) != 0 {
		t.Fatalf("the FRONT face has %d mana abilities, want none", len(before))
	}

	advanceToEndStepOf(t, g, seat)
	passPriorityAroundTable(t, g)

	after := game.ManaAbilitiesForCard(*battlefieldCardFor(g, id))
	if len(after) != 2 {
		t.Fatalf("the back face has %d mana abilities, want 2", len(after))
	}
	// "{T}: Add {U} for each artifact you control" is a count read at
	// activation, and five artifacts is what is on the board.
	var scaled *game.ManaAbilityShape
	for i := range after {
		if after[i].ProducedFunc != nil {
			scaled = &after[i]
		}
	}
	if scaled == nil {
		t.Fatal("neither ability scales with the artifact count")
	}
	if got := scaled.ProducedFunc(g, me.ID, id); got != "{U}{U}{U}{U}{U}" {
		t.Errorf("produced = %q, want five {U} for five artifacts", got)
	}
}

// --- Fable of the Mirror-Breaker // Reflection of Kiki-Jiki ---------

// TestFableChapterOneMakesTheGoblinShaman pins the token's own attack
// trigger, which the card shipped WITHOUT until ADR 0083 (#1248): the
// assertion used to be that a non-copy token had no triggered
// abilities at all, because it had no catalog key to hang one off.
// The trigger itself is exercised end to end in
// token_abilities_cards_test.go.
func TestFableChapterOneMakesTheGoblinShaman(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	importAndCast(t, g, fableRow(), me)
	passPriorityAroundTable(t, g)

	if got := countBattlefieldByName(g, "Goblin Shaman"); got != 1 {
		t.Fatalf("Goblin Shamans after chapter I = %d, want 1", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Goblin Shaman" {
			continue
		}
		if c.Power != 2 || c.Toughness != 2 {
			t.Errorf("token is %d/%d, want 2/2", c.Power, c.Toughness)
		}
		if c.TokenKey != game.TokenKey("goblin-shaman") {
			t.Errorf("token key = %q, want the Goblin Shaman's — without it the "+
				"catalog cannot find its attack trigger", c.TokenKey)
		}
		if got := len(game.TriggersForCard(c)); got != 1 {
			t.Errorf("the token has %d triggered abilities, want its printed "+
				"\"whenever this creature attacks, create a Treasure token\"", got)
		}
	}
}

// TestFableChapterThreeExilesAndReturnsItTransformed is the second
// verb's card-level proof, and every assertion is chosen because an
// IN-PLACE transform would fail it.
func TestFableChapterThreeExilesAndReturnsItTransformed(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	saga := importAndCast(t, g, fableRow(), me)
	passPriorityAroundTable(t, g)

	// Chapter II queues a discard prompt; decline it and move on.
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	answerDiscard(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if got := loreCountersOn(g, saga); got != 2 {
		t.Fatalf("lore counters after chapter II = %d, want 2", got)
	}

	// Chapter III.
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)

	if battlefieldCardFor(g, saga) != nil {
		t.Fatal("the original object is still on the battlefield — chapter III " +
			"is an exile and a return, not an in-place transform")
	}
	if inGraveyardOf(g, me.ID, saga) {
		t.Fatal("the Saga was sacrificed (CR 714.4) instead of coming back")
	}
	var back *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Reflection of Kiki-Jiki" {
			back = &g.Battlefield.Cards[i]
		}
	}
	if back == nil {
		t.Fatal("Reflection of Kiki-Jiki never reached the battlefield")
	}
	if back.InstanceID == saga {
		t.Error("it kept its InstanceID — CR 400.7 makes the returning permanent a new object")
	}
	if back.ActiveFace != 1 {
		t.Errorf("ActiveFace = %d, want the back face", back.ActiveFace)
	}
	if back.Counters[game.CounterLore] != 0 {
		t.Errorf("it came back with %d lore counters", back.Counters[game.CounterLore])
	}
	if !back.IsCreature() || back.Power != 2 || back.Toughness != 2 {
		t.Errorf("came back as %q %d/%d, want a 2/2 creature",
			back.TypeLine, back.Power, back.Toughness)
	}
	if !back.SummonedThisTurn {
		t.Error("it is not summoning sick — the returning permanent entered this turn")
	}
	assertFaceInvariant(t, g)
}

// TestReflectionOfKikiJikiCopiesWithHaste is the back face's ability,
// which exists on the battlefield only because the "#1" key handed it
// over when the face flipped. The delayed sacrifice is asserted too:
// "sacrifice IT" is the token, and a misreading that sacrificed the
// Reflection would empty the board instead.
func TestReflectionOfKikiJikiCopiesWithHaste(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]

	// Put the back face straight onto the battlefield: chapter III's
	// route is covered above, and what is under test here is the
	// ability, not how the face got there.
	row := fableRow()
	reflection := importToHand(row, me)
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID == reflection {
				me.Hand.Cards[i].SetFace(1)
			}
		}
	})
	var entered uuid.UUID
	g.WithWriteLock(func() {
		var err error
		entered, err = g.PutFromHandOntoBattlefieldForEffect(reflection, game.HandEntryOptions{})
		if err != nil {
			t.Fatalf("put the back face onto the battlefield: %v", err)
		}
	})
	g.WithWriteLock(func() {
		// It entered this turn; the {T} needs it not to be sick.
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == entered {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})

	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Grizzly Bears",
		TypeLine:   "Creature — Bear",
		Owner:      me.ID, Controller: me.ID,
		Power: 2, Toughness: 2,
	})

	abilities := game.ActivatedAbilitiesForCard(*battlefieldCardFor(g, entered))
	if len(abilities) != 1 {
		t.Fatalf("the back face has %d activated abilities, want 1", len(abilities))
	}

	if err := g.ActivateCatalogAbility(me.ID, entered, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)

	var token *game.Card
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Name == "Grizzly Bears" && c.InstanceID != bear {
			token = c
		}
	}
	if token == nil {
		t.Fatal("no copy token was created")
	}
	if !game.HasKeyword(token, "haste") {
		t.Errorf("the copy does not have haste: %v", token.Keywords)
	}

	// The delayed trigger sacrifices the TOKEN at the next end step.
	tokenID := token.InstanceID
	advanceToEndStepOf(t, g, seat)
	passPriorityAroundTable(t, g)

	if battlefieldCardFor(g, tokenID) != nil {
		t.Error("the token survived the end step")
	}
	if battlefieldCardFor(g, entered) == nil {
		t.Error("Reflection of Kiki-Jiki sacrificed ITSELF — \"sacrifice it\" is the token")
	}
	if battlefieldCardFor(g, bear) == nil {
		t.Error("the copied creature was sacrificed")
	}
}
