package testsds

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	cardse "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/e"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// electroduplicateBearToken casts Electroduplicate on a 4/4 Big Bear and
// resolves it, returning the token's and the bear's IDs.
func electroduplicateBearToken(t *testing.T, g *game.Game) (token, bear uuid.UUID) {
	t.Helper()
	bear = effectstest.PushVanillaCreature(g, g.Seats[0].ID, "Big Bear", 4, 4)
	effectstest.CastCatalogSpell(t, g, "Electroduplicate", "Sorcery", effectstest.ElectroduplicateOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: bear},
	})
	effectstest.PassPriorityAroundTable(t, g)
	tokens, _ := effectstest.CopiesOnBattlefield(g, "Big Bear")
	if len(tokens) != 1 {
		t.Fatalf("%d token copies of the bear, want 1", len(tokens))
	}
	return tokens[0].InstanceID, bear
}

func TestElectroduplicateMakesAHastyCopyThatSacrificesAtEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	token, bear := electroduplicateBearToken(t, g)
	tok, _ := effectstest.FindBattlefieldByID(g, token)
	if tok.Power != 4 || tok.Toughness != 4 {
		t.Errorf("token P/T = %d/%d, want 4/4", tok.Power, tok.Toughness)
	}
	if !effectstest.HasKeywordForTest(tok.Keywords, "haste") {
		t.Errorf("the token should have haste: %v", tok.Keywords)
	}
	// CR 707.9a: the quoted trigger is a copiable value of the token.
	if want := game.GrantKey(cardse.ElectroduplicateGrant); !slices.Equal(tok.GrantedAbilities, []string{want}) {
		t.Errorf("the token's granted abilities are %v, want [%s]", tok.GrantedAbilities, want)
	}

	// Advance to the end step; the token's own trigger sacrifices it.
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OnBattlefieldByID(g, token) || !effectstest.SacrificedInLog(g, token) {
		t.Errorf("the token should be sacrificed at the end step")
	}
	if !effectstest.OnBattlefieldByID(g, bear) {
		t.Error("the original bear was sacrificed too")
	}
}

// The sacrifice is the token's own ability, not a delayed trigger of the
// spell: a token that has lost all its abilities stays.
func TestElectroduplicateTokenThatLosesItsAbilitiesIsNotSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	token, _ := electroduplicateBearToken(t, g)
	g.WithWriteLock(func() {
		g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(token),
			[]game.Mod{game.LoseAllAbilitiesMod()}, g.UntilEndOfTurnDuration(), "loses all abilities")
	})
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.OnBattlefieldByID(g, token) {
		t.Fatal("a token with no abilities was sacrificed: the sacrifice was not the token's own ability")
	}
}

// #4950, CR 707.9a: a Clone copying the token copies haste and the
// trigger, so it is sacrificed at the end step too.
func TestElectroduplicateCloneOfTheTokenIsSacrificedAtEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	token, bear := electroduplicateBearToken(t, g)

	cloneID := effectstest.CastCatalogSpell(t, g, "Clone", "Creature — Shapeshifter", effectstest.OracleClone, nil)
	effectstest.ResolveWithCopyChoice(t, g, token)
	clone := effectstest.CopyBattlefieldCard(t, g, cloneID)
	if clone.Name != "Big Bear" || clone.IsToken() {
		t.Fatalf("the Clone is %q (token %v), want a nontoken copy of the token", clone.Name, clone.IsToken())
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, cloneID, "haste") {
		t.Errorf("the Clone's abilities are %v, want the copied haste", effectstest.EffectiveAbilities(t, g, cloneID))
	}
	if len(clone.GrantedAbilities) != 1 {
		t.Fatalf("the Clone's granted abilities are %v, want the copied sacrifice trigger", clone.GrantedAbilities)
	}

	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassThroughEndStepTriggers(t, g)
	if effectstest.OnBattlefieldByID(g, cloneID) || !effectstest.SacrificedInLog(g, cloneID) {
		t.Error("the Clone of the token was not sacrificed by its copied trigger")
	}
	if effectstest.OnBattlefieldByID(g, token) {
		t.Error("the token itself was not sacrificed")
	}
	if !effectstest.OnBattlefieldByID(g, bear) {
		t.Error("the original bear was sacrificed")
	}
}

// #4950: a TOKEN copy of the token, made by a copy effect with no
// exception of its own (Cackling Counterpart), copies the trigger too
// (TokenCopyTemplate carries the token's grants) and is sacrificed.
func TestElectroduplicateTokenCopyOfTheTokenIsSacrificedAtEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	token, bear := electroduplicateBearToken(t, g)

	effectstest.CastCatalogSpell(t, g, "Cackling Counterpart", "Instant", effectstest.CacklingCounterpartOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: token},
	})
	effectstest.PassPriorityAroundTable(t, g)
	tokens, _ := effectstest.CopiesOnBattlefield(g, "Big Bear")
	if len(tokens) != 2 {
		t.Fatalf("%d Big Bear tokens, want 2", len(tokens))
	}
	var second game.Card
	for _, c := range tokens {
		if c.InstanceID != token {
			second = c
		}
	}
	if !effectstest.HasKeywordForTest(second.Keywords, "haste") {
		t.Errorf("the copy of the token should have the copied haste: %v", second.Keywords)
	}
	if len(second.GrantedAbilities) != 1 {
		t.Fatalf("the copy of the token has granted abilities %v, want the copied sacrifice trigger", second.GrantedAbilities)
	}

	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassThroughEndStepTriggers(t, g)
	if effectstest.OnBattlefieldByID(g, second.InstanceID) || !effectstest.SacrificedInLog(g, second.InstanceID) {
		t.Error("the token copy of the token was not sacrificed by its copied trigger")
	}
	if effectstest.OnBattlefieldByID(g, token) {
		t.Error("the Electroduplicate token was not sacrificed")
	}
	if !effectstest.OnBattlefieldByID(g, bear) {
		t.Error("the original bear was sacrificed")
	}
}

// The trigger is data in the token's copiable values (a grant name the
// catalog resolves), so it survives a capture → JSON → restore: with the
// token on the battlefield before the end step, and with its trigger on
// the stack.
func TestElectroduplicateTokenSurvivesARestore(t *testing.T) {
	t.Run("the token before the end step", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		token, _ := electroduplicateBearToken(t, g)
		g = effectstest.RestoreRoundTrip(t, g, true)
		tok, ok := effectstest.FindBattlefieldByID(g, token)
		if !ok || len(tok.GrantedAbilities) != 1 || !effectstest.EffectiveAbilitiesContain(t, g, token, "haste") {
			t.Fatalf("the restored token lost haste or the trigger: %+v", tok.GrantedAbilities)
		}
		effectstest.AdvanceTo(t, g, game.StepEnd)
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.OnBattlefieldByID(g, token) || !effectstest.SacrificedInLog(g, token) {
			t.Error("the restored token was not sacrificed at the end step")
		}
	})
	t.Run("the trigger on the stack", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		token, _ := electroduplicateBearToken(t, g)
		effectstest.AdvanceTo(t, g, game.StepEnd)
		if effectstest.TriggerOnStack(g, token) == nil {
			t.Fatal("setup: the token's sacrifice trigger is not on the stack at the end step")
		}
		g = effectstest.RestoreRoundTrip(t, g, true)
		if effectstest.TriggerOnStack(g, token) == nil {
			t.Fatal("the restore dropped the sacrifice trigger")
		}
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.OnBattlefieldByID(g, token) || !effectstest.SacrificedInLog(g, token) {
			t.Error("the restored trigger did not sacrifice the token")
		}
	})
}

func TestElectroduplicateFlashback(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Big Bear", 4, 4)
	id := effectstest.SeedGraveyardCard(t, g, "Electroduplicate", "Sorcery", effectstest.ElectroduplicateOracle)

	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "flashback",
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.B43TokensNamed(g, me.ID, "Big Bear"); got != 1 {
		t.Fatalf("token copies = %d, want 1", got)
	}
	if !g.Exile.Contains(id) {
		t.Errorf("the flashed-back sorcery should be exiled")
	}
}
