package teststo

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pushArtifacts puts n plain artifacts onto a player's battlefield —
// what Storm the Vault's intervening if counts.
func pushArtifacts(g *game.Game, p *game.Player, n int) {
	for i := 0; i < n; i++ {
		effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(),
			Name:       "Filler Relic",
			TypeLine:   "Artifact",
			Owner:      p.ID,
			Controller: p.ID,
		})
	}
}

func fableRow() cards.Card {
	return effectstest.TransformRow(effects.FableOfTheMirrorBreakerOracleID,
		"Fable of the Mirror-Breaker", "Enchantment — Saga", "{2}{R}",
		"Reflection of Kiki-Jiki", "Enchantment Creature — Goblin Shaman",
		"2", "2", []string{"R"})
}

// --- Storm the Vault // Vault of Catlacan --------------------------

// TestStormTheVaultTransformsAtTheEndStep is the in-place verb's
// card-level proof: an enchantment becomes a land, in its slot,
// keeping its instance ID.
func TestStormTheVaultTransformsAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	id := effectstest.ImportToBattlefield(t, g, effectstest.StormTheVaultRow(), me)
	pushArtifacts(g, me, 5)

	effectstest.AdvanceToEndStepOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)

	card := effectstest.BattlefieldCardFor(g, id)
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
	effectstest.AssertFaceInvariant(t, g)
}

// TestStormTheVaultWithFourArtifactsDoesNotTrigger is CR 603.4's first
// check. Four artifacts is one short, and the ability must not even
// reach the stack — a card that triggered and then did nothing would
// still be observably different, because it would be answerable.
func TestStormTheVaultWithFourArtifactsDoesNotTrigger(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	id := effectstest.ImportToBattlefield(t, g, effectstest.StormTheVaultRow(), me)
	pushArtifacts(g, me, 4)

	effectstest.AdvanceToEndStepOf(t, g, seat)
	if item := effectstest.TriggerOnStack(g, id); item != nil {
		t.Fatal("the end-step ability triggered with only four artifacts (CR 603.4)")
	}
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.BattlefieldCardFor(g, id).ActiveFace; got != 0 {
		t.Errorf("ActiveFace = %d, want the front face", got)
	}
}

// TestStormTheVaultLosesTheArtifactsInResponse is CR 603.4's SECOND
// check, and the one an implementation that only gated AppliesTo would
// fail. The trigger goes on the stack with five artifacts; by the time
// it resolves there are four, and it does nothing.
func TestStormTheVaultLosesTheArtifactsInResponse(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	id := effectstest.ImportToBattlefield(t, g, effectstest.StormTheVaultRow(), me)
	pushArtifacts(g, me, 5)

	effectstest.AdvanceToEndStepOf(t, g, seat)
	if effectstest.TriggerOnStack(g, id) == nil {
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
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.BattlefieldCardFor(g, id).ActiveFace; got != 0 {
		t.Errorf("it transformed anyway — CR 603.4's second check did not run")
	}
}

// TestVaultOfCatlacanManaAbilitiesAreLiveAfterTheTransform is the real
// payoff of the "<oracle_id>#1" key: nobody wires the back face's
// abilities up, and they have to be there the instant the face flips.
func TestVaultOfCatlacanManaAbilitiesAreLiveAfterTheTransform(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	id := effectstest.ImportToBattlefield(t, g, effectstest.StormTheVaultRow(), me)
	pushArtifacts(g, me, 5)

	before := game.ManaAbilitiesForCard(*effectstest.BattlefieldCardFor(g, id))
	if len(before) != 0 {
		t.Fatalf("the FRONT face has %d mana abilities, want none", len(before))
	}

	effectstest.AdvanceToEndStepOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)

	after := game.ManaAbilitiesForCard(*effectstest.BattlefieldCardFor(g, id))
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
	if got := effectstest.ProducedOfShape(*scaled, g, me.ID, id); got != "{U}{U}{U}{U}{U}" {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.ImportAndCast(t, g, fableRow(), me)
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.CountBattlefieldByName(g, "Goblin Shaman"); got != 1 {
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
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	saga := effectstest.ImportAndCast(t, g, fableRow(), me)
	effectstest.PassPriorityAroundTable(t, g)

	// Chapter II queues a discard prompt; decline it and move on.
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LoreCountersOn(g, saga); got != 2 {
		t.Fatalf("lore counters after chapter II = %d, want 2", got)
	}

	// Chapter III.
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.BattlefieldCardFor(g, saga) != nil {
		t.Fatal("the original object is still on the battlefield — chapter III " +
			"is an exile and a return, not an in-place transform")
	}
	if effectstest.InGraveyardOf(g, me.ID, saga) {
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
	effectstest.AssertFaceInvariant(t, g)
}

// TestReflectionOfKikiJikiCopiesWithHaste is the back face's ability,
// which exists on the battlefield only because the "#1" key handed it
// over when the face flipped. The delayed sacrifice is asserted too:
// "sacrifice IT" is the token, and a misreading that sacrificed the
// Reflection would empty the board instead.
func TestReflectionOfKikiJikiCopiesWithHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]

	// Put the back face straight onto the battlefield: chapter III's
	// route is covered above, and what is under test here is the
	// ability, not how the face got there.
	row := fableRow()
	reflection := effectstest.ImportToHand(row, me)
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

	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Grizzly Bears",
		TypeLine:   "Creature — Bear",
		Owner:      me.ID, Controller: me.ID,
		Power: 2, Toughness: 2,
	})

	abilities := game.ActivatedAbilitiesForCard(*effectstest.BattlefieldCardFor(g, entered))
	if len(abilities) != 1 {
		t.Fatalf("the back face has %d activated abilities, want 1", len(abilities))
	}

	if err := g.ActivateCatalogAbility(me.ID, entered, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

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
	effectstest.AdvanceToEndStepOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.BattlefieldCardFor(g, tokenID) != nil {
		t.Error("the token survived the end step")
	}
	if effectstest.BattlefieldCardFor(g, entered) == nil {
		t.Error("Reflection of Kiki-Jiki sacrificed ITSELF — \"sacrifice it\" is the token")
	}
	if effectstest.BattlefieldCardFor(g, bear) == nil {
		t.Error("the copied creature was sacrificed")
	}
}

// TestStormTheVaultGoneBeforeItsTriggerResolvesTransformsNothing is
// CR 701.27a / CR 608.2b: the end-step trigger's condition still holds
// (five artifacts) but the Vault itself was destroyed in response, so
// "transform Storm the Vault" has no permanent to turn over. That is
// nothing happening, not an effect error (#4038).
func TestStormTheVaultGoneBeforeItsTriggerResolvesTransformsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	id := effectstest.ImportToBattlefield(t, g, effectstest.StormTheVaultRow(), me)
	pushArtifacts(g, me, 5)

	effectstest.AdvanceToEndStepOf(t, g, seat)
	if effectstest.TriggerOnStack(g, id) == nil {
		t.Fatal("the end-step ability did not trigger with five artifacts")
	}
	effectstest.B25Destroy(g, id)
	effectstest.PassPriorityAroundTable(t, g)

	if errs := effectstest.EffectErrors(g); len(errs) != 0 {
		t.Fatalf("effect errors: %v", errs)
	}
	if effectstest.BattlefieldCardFor(g, id) != nil {
		t.Error("the Vault is back on the battlefield")
	}
}

// TestTransformOfAPermanentThatLeftTheBattlefieldDoesNothing pins the
// targeted verb: a Transform aimed at an ID that is now in the
// graveyard returns nil (CR 701.27a transforms a permanent, CR 608.2b).
func TestTransformOfAPermanentThatLeftTheBattlefieldDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.ImportToBattlefield(t, g, effectstest.StormTheVaultRow(), me)
	effectstest.B25Destroy(g, id)

	var err error
	g.WithWriteLock(func() {
		err = effects.Transform{Target: id}.Apply(effects.NewContext(g, &game.StackItem{Controller: me.ID}))
	})
	if err != nil {
		t.Fatalf("Transform of a card in the graveyard: %v, want nil", err)
	}
}
