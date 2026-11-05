package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// token_copy_end_step_grant_test.go — Electroduplicate and Kindle the Inner
// Flame: "Create a token that's a copy of target creature you control,
// except it has haste and 'At the beginning of the end step, sacrifice this
// token.'" CR 707.9a makes the quoted ability part of the token's copiable
// values, so it is the token's own trigger (not a CR 603.7 delayed one of
// the spell): a copy of the token has it and sacrifices itself, and it
// triggers at the beginning of every end step the token sees.

const cacklingCounterpartOracle = "9e2adca5-f39c-4a09-bcce-8238ebac2c4a"

type endStepGrantCard struct {
	name, typeLine, oracle, grant string
}

var endStepGrantCards = []endStepGrantCard{
	{"Electroduplicate", "Sorcery", electroduplicateOracle, electroduplicateGrant},
	{kindleName, kindleTypeLine, kindleOracle, kindleTheInnerFlameGrant},
}

// endStepGrantToken casts the card on a 4/4 the active seat controls and
// resolves it, returning the token's and the original's IDs.
func endStepGrantToken(t *testing.T, g *game.Game, c endStepGrantCard) (token, original uuid.UUID) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	original = pushVanillaCreature(g, me.ID, "Big Bear", 4, 4)
	castCatalogSpell(t, g, c.name, c.typeLine, c.oracle, []game.TargetRef{{Kind: game.TargetCard, ID: original}})
	passPriorityAroundTable(t, g)
	tokens, _ := copiesOnBattlefield(g, "Big Bear")
	if len(tokens) != 1 {
		t.Fatalf("%s made %d token copies, want 1", c.name, len(tokens))
	}
	tok := tokens[0]
	if !hasKeywordForTest(tok.Keywords, "haste") {
		t.Errorf("%s: the token has no haste: %v", c.name, tok.Keywords)
	}
	if !slices.Contains(tok.GrantedAbilities, game.GrantKey(c.grant)) {
		t.Fatalf("%s: the token's copiable granted abilities are %v, want %s", c.name, tok.GrantedAbilities, c.grant)
	}
	return tok.InstanceID, original
}

// resolveEndStepTriggers passes priority until the stack is empty,
// answering any trigger-order prompt in the order offered.
func resolveEndStepTriggers(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 8; i++ {
		passPriorityAroundTable(t, g)
		c := pendingOfKind(g, game.PendingChoiceTriggerOrder)
		if c == nil {
			return
		}
		if err := g.ResolveTriggerOrder(c.ID, c.Chooser, append([]uuid.UUID(nil), c.TriggerOrderIDs...)); err != nil {
			t.Fatalf("ResolveTriggerOrder: %v", err)
		}
	}
}

// The token is sacrificed by its own trigger at the end step, and a token
// copy of the token made by another copy effect (Cackling Counterpart) has
// the trigger too and is sacrificed beside it. The original stays.
func TestTokenCopyEndStepSacrificeIsCopiedWithTheToken(t *testing.T) {
	for _, c := range endStepGrantCards {
		t.Run(c.name, func(t *testing.T) {
			g := newCatalogGame(t)
			token, original := endStepGrantToken(t, g, c)

			castCatalogSpell(t, g, "Cackling Counterpart", "Instant", cacklingCounterpartOracle,
				[]game.TargetRef{{Kind: game.TargetCard, ID: token}})
			passPriorityAroundTable(t, g)
			var second uuid.UUID
			for _, tok := range func() []game.Card { ts, _ := copiesOnBattlefield(g, "Big Bear"); return ts }() {
				if tok.InstanceID != token {
					second = tok.InstanceID
					if !slices.Contains(tok.GrantedAbilities, game.GrantKey(c.grant)) {
						t.Fatalf("the copy of the token has granted abilities %v, want the copied %s", tok.GrantedAbilities, c.grant)
					}
					if !hasKeywordForTest(tok.Keywords, "haste") {
						t.Errorf("the copy of the token has no copied haste: %v", tok.Keywords)
					}
				}
			}
			if second == uuid.Nil {
				t.Fatal("Cackling Counterpart made no copy of the token")
			}

			advanceToEndStep(t, g)
			resolveEndStepTriggers(t, g)
			for what, id := range map[string]uuid.UUID{"the token": token, "the copy of the token": second} {
				if g.Battlefield.Contains(id) || !sacrificedInLog(g, id) {
					t.Errorf("%s was not sacrificed at the end step", what)
				}
			}
			if !g.Battlefield.Contains(original) {
				t.Error("the original creature was sacrificed")
			}
		})
	}
}

// A Clone entering as a copy of the token copies the trigger too (CR
// 707.9a) and is sacrificed at the end step, though it is not a token.
func TestTokenCopyEndStepSacrificeIsCopiedByAClone(t *testing.T) {
	for _, c := range endStepGrantCards {
		t.Run(c.name, func(t *testing.T) {
			g := newCatalogGame(t)
			token, _ := endStepGrantToken(t, g, c)

			cloneID := castCatalogSpell(t, g, "Clone", "Creature — Shapeshifter", oracleClone, nil)
			resolveWithCopyChoice(t, g, token)
			clone := copyBattlefieldCard(t, g, cloneID)
			if clone.Name != "Big Bear" || !slices.Contains(clone.GrantedAbilities, game.GrantKey(c.grant)) {
				t.Fatalf("the Clone is %q with granted abilities %v, want a copy of the token with %s", clone.Name, clone.GrantedAbilities, c.grant)
			}

			advanceToEndStep(t, g)
			resolveEndStepTriggers(t, g)
			if g.Battlefield.Contains(cloneID) || !sacrificedInLog(g, cloneID) {
				t.Error("the Clone of the token was not sacrificed by its copied trigger")
			}
		})
	}
}

// "At the beginning of the end step" is the token's ability, so it
// triggers at every end step, not only the next one: a token that has
// lost all its abilities until end of turn survives this end step, and
// is sacrificed at the next one (the next player's turn).
func TestTokenCopyEndStepSacrificeTriggersAtEachEndStep(t *testing.T) {
	for _, c := range endStepGrantCards {
		t.Run(c.name, func(t *testing.T) {
			g := newCatalogGame(t)
			token, _ := endStepGrantToken(t, g, c)
			g.WithWriteLock(func() {
				g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(token),
					[]game.Mod{game.LoseAllAbilitiesMod()}, g.UntilEndOfTurnDuration(), "loses all abilities")
			})

			advanceToEndStep(t, g)
			resolveEndStepTriggers(t, g)
			if !g.Battlefield.Contains(token) {
				t.Fatal("a token with no abilities was sacrificed — the trigger was not the token's own")
			}

			next := (g.Turn.ActiveSeat + 1) % len(g.Seats)
			advanceToEndStepOf(t, g, next)
			resolveEndStepTriggers(t, g)
			if g.Battlefield.Contains(token) || !sacrificedInLog(g, token) {
				t.Error("the token was not sacrificed at the next player's end step once it had its ability back")
			}
		})
	}
}

// Face Yourself's "except it has haste and 'At the beginning of the end
// step, if you don't control a planeswalker, sacrifice this creature.'" is
// the same CR 707.9a grant: a token copy of one of its tokens has the
// trigger too, and both are sacrificed with no planeswalker.
func TestFaceYourselfSacrificeIsCopiedWithTheToken(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rfaColoredCreature(g, opp.ID, "Their Bear", 2, 2, "G")
	castCatalogSpell(t, g, "Face Yourself", "Sorcery", rfaFaceYourself,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	var token uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Their Bear" && c.Controller == me.ID {
			token = c.InstanceID
			if !slices.Contains(c.GrantedAbilities, game.GrantKey(faceYourselfSacrifice)) {
				t.Fatalf("the token's copiable granted abilities are %v, want %s", c.GrantedAbilities, faceYourselfSacrifice)
			}
		}
	}
	if token == uuid.Nil {
		t.Fatal("no copy of their Bear under your control")
	}

	castCatalogSpell(t, g, "Cackling Counterpart", "Instant", cacklingCounterpartOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: token}})
	passPriorityAroundTable(t, g)
	var second uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Their Bear" && c.Controller == me.ID && c.InstanceID != token {
			second = c.InstanceID
			if !slices.Contains(c.GrantedAbilities, game.GrantKey(faceYourselfSacrifice)) {
				t.Fatalf("the copy of the token has granted abilities %v, want the copied %s", c.GrantedAbilities, faceYourselfSacrifice)
			}
		}
	}
	if second == uuid.Nil {
		t.Fatal("Cackling Counterpart made no copy of the token")
	}

	advanceToEndStep(t, g)
	resolveEndStepTriggers(t, g)
	for what, id := range map[string]uuid.UUID{"the token": token, "the copy of the token": second} {
		if g.Battlefield.Contains(id) {
			t.Errorf("%s was not sacrificed at the end step", what)
		}
	}
}
