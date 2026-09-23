package game

import (
	"go/ast"
	"go/parser"
	gotoken "go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// token_abilities_test.go — ADR 0083 (#1248) at the engine seam: a
// token that is NOT a copy carries a triggered or a static ability of
// its own, found through the same accessors a printed card's are.
//
// Every test here stubs the catalog, because the claim is about the
// DISPATCH and not about any one card: `game.CatalogKey` falls back to
// `Card.TokenKey`, so the trigger harvest, the LTB harvest and the
// layer pass all answer for a token the moment its def has something
// in the slot. The card-level half — a real Reef Worm, a real Nesting
// Dragon — lives in cards/effects/token_abilities_cards_test.go.

const testTokenSlug = "test-token"

// tokenOnBattlefield puts a token with `slug`'s catalog key onto the
// battlefield and fires the entry event, so the layer listener stamps
// its CR 613.7 timestamp.
func tokenOnBattlefield(g *Game, controller uuid.UUID, name string, power, toughness int, slug string) uuid.UUID {
	c := NewCard(name, controller)
	c.TypeLine = "Token Creature — Test"
	c.Controller, c.Owner = controller, controller
	c.Power, c.Toughness = power, toughness
	c.TokenKey = TokenKey(slug)
	return pushTypedTestCard(g, c)
}

// TestATokenTriggerFiresAndResolves is the whole seam in one run: a
// token with a printed "when this token dies" makes its controller
// gain life, and the ability reaches the stack by the ordinary
// harvest rather than by anything token-shaped.
func TestATokenTriggerFiresAndResolves(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := tokenOnBattlefield(g, me.ID, "Pest", 1, 1, testTokenSlug)

	withCatalogTriggers(t, func(key string) []TriggeredAbility {
		if key != TokenKey(testTokenSlug) {
			return nil
		}
		return []TriggeredAbility{{
			Watches: []EventKind{EventLTB},
			AppliesTo: func(ev Event, source *Card, _ Characteristic, _ *Game) bool {
				return ev.CardID == source.InstanceID
			},
			Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
				controller := source.Controller
				return &StackItem{
					Kind:         StackItemTriggered,
					SourceCardID: source.InstanceID,
					Controller:   controller,
					Owner:        source.Owner,
					Label:        "Pest — you gain 1 life",
					Effect: func(g *Game, _ *StackItem) error {
						g.playerByIDLocked(controller).Life++
						return nil
					},
				}
			},
		}}
	})

	lifeBefore := me.Life
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
	if len(g.PendingTriggers) != 1 {
		t.Fatalf("PendingTriggers = %d, want the token's own dies-trigger", len(g.PendingTriggers))
	}
	if got := g.PendingTriggers[0].Label; got != "Pest — you gain 1 life" {
		t.Errorf("queued %q", got)
	}

	g.WithWriteLock(func() { g.drainPendingTriggersAPNAPLocked() })
	resolveTop(t, g)
	if me.Life != lifeBefore+1 {
		t.Errorf("life %d -> %d, want +1 from the token's trigger", lifeBefore, me.Life)
	}
}

// TestATokenDiesTriggerSurvivesTheCeaseToExistSweep is CR 111.7 and
// CR 704.5d in the right order: the token really reaches the
// graveyard, the harvest reads it there, and only the NEXT state-based
// check removes it. The identity the harvest keys on is the token key
// restored from lastKnownTriggerIdentity (CR 603.10), which is the
// field ADR 0083 decision 4 added.
func TestATokenDiesTriggerSurvivesTheCeaseToExistSweep(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := tokenOnBattlefield(g, me.ID, "Dragon Egg", 0, 2, testTokenSlug)

	fired := 0
	withCatalogTriggers(t, func(key string) []TriggeredAbility {
		if key != TokenKey(testTokenSlug) {
			return nil
		}
		return []TriggeredAbility{{
			Watches: []EventKind{EventLTB},
			AppliesTo: func(ev Event, source *Card, _ Characteristic, _ *Game) bool {
				return ev.CardID == source.InstanceID
			},
			Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
				fired++
				return &StackItem{
					Kind:         StackItemTriggered,
					SourceCardID: source.InstanceID,
					Controller:   source.Controller,
					Owner:        source.Owner,
					Label:        "hatch",
				}
			},
		}}
	})

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
	if fired != 1 {
		t.Fatalf("the dies-trigger fired %d times, want once", fired)
	}
	// The token is in the graveyard until the sweep runs, and gone
	// afterwards — CR 111.7 then CR 704.5d.
	if !me.Graveyard.Contains(id) {
		t.Fatal("the token never reached the graveyard, so nothing could have read it there")
	}
	g.WithWriteLock(func() { g.tokenCeaseToExistSBALocked() })
	if me.Graveyard.Contains(id) {
		t.Error("CR 704.5d did not remove the token from the graveyard")
	}
}

// TestADyingTokensTriggerIsReadOffTheIdentityItHadOnTheBattlefield is
// CR 603.10 for a token, and the reason `triggerIdentityLKI` grew a
// `TokenKey` (ADR 0083 decision 4).
//
// A token that a copy effect had turned into a DIFFERENT token dies.
// The battlefield exit restores its printed self on the way out, so
// the card sitting in the graveyard carries its OWN key again — but
// the trigger belongs to what it was while it was still there. The
// oracle ID has been snapshotted for this for as long as the snapshot
// has existed; without its sibling a token's identity is silently the
// one exception.
func TestADyingTokensTriggerIsReadOffTheIdentityItHadOnTheBattlefield(t *testing.T) {
	const wasCopying = "copied-token"
	g := newActiveGame(t)
	me := g.Seats[0]

	fired := ""
	withCatalogTriggers(t, func(key string) []TriggeredAbility {
		switch key {
		case TokenKey(testTokenSlug), TokenKey(wasCopying):
		default:
			return nil
		}
		return []TriggeredAbility{{
			Watches: []EventKind{EventLTB},
			AppliesTo: func(ev Event, source *Card, _ Characteristic, _ *Game) bool {
				return ev.CardID == source.InstanceID
			},
			Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
				fired = key
				return &StackItem{
					Kind:         StackItemTriggered,
					SourceCardID: source.InstanceID,
					Controller:   source.Controller,
					Owner:        source.Owner,
					Label:        key,
				}
			},
		}}
	})

	// A Pest that is currently a copy of a Dragon Egg.
	pest := NewCard("Pest", me.ID)
	pest.TypeLine = "Token Creature — Pest"
	pest.Controller, pest.Owner = me.ID, me.ID
	pest.TokenKey = TokenKey(testTokenSlug)
	egg := Card{Name: "Dragon Egg", TypeLine: "Token Creature — Dragon", TokenKey: TokenKey(wasCopying)}
	pest.applyCopy(CopiableValuesOf(egg), egg)
	id := pushTypedTestCard(g, pest)

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})

	if fired != TokenKey(wasCopying) {
		t.Errorf("the trigger fired for %q, want %q — CR 603.10 reads the identity the "+
			"permanent had while it was still on the battlefield", fired, TokenKey(wasCopying))
	}
}

// TestATokenStaticAppliesAndStopsWhenTheTokenLeaves is the static half
// (CR 611.2b): the token's anthem reaches another creature through the
// ordinary layer pass, and goes the moment the token does. The layer
// pass has always walked the whole battlefield; what it could not do
// was find a token's entry.
func TestATokenStaticAppliesAndStopsWhenTheTokenLeaves(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	token := tokenOnBattlefield(g, me.ID, "Honour Guard", 2, 2, testTokenSlug)

	bear := NewCard("Grizzly Bears", me.ID)
	bear.TypeLine = "Creature — Bear"
	bear.OracleID = "oracle-bears"
	bear.Controller, bear.Owner = me.ID, me.ID
	bear.Power, bear.Toughness = 2, 2
	bearID := pushTypedTestCard(g, bear)

	withStaticAbilities(t, func(key string) []StaticAbility {
		if key != TokenKey(testTokenSlug) {
			return nil
		}
		return []StaticAbility{{
			Layer:    Layer7PT,
			SubLayer: SubLayer7C_Modify,
			AppliesTo: func(target *Card, _ *Game, source *Card) bool {
				return target.InstanceID != source.InstanceID &&
					target.Controller == source.Controller
			},
			Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
				c.Power++
				c.Toughness++
			},
		}}
	})

	if got := layeredBattlefieldCard(t, g, bearID).Effective(); got.Power != 3 || got.Toughness != 3 {
		t.Fatalf("the Bear is %d/%d under the token's anthem, want 3/3", got.Power, got.Toughness)
	}

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(token); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
	if got := layeredBattlefieldCard(t, g, bearID).Effective(); got.Power != 2 || got.Toughness != 2 {
		t.Errorf("the Bear is %d/%d after the token left, want its printed 2/2 back", got.Power, got.Toughness)
	}
}

// TestACopyOfATokenCarriesItsTriggerAndStatic is CR 707.2: the token
// key is a copiable value, so a copy of a token that prints an ability
// prints the same one. Nothing in the copy machinery knows about
// tokens; PrintedValues.TokenKey travelling is the whole of it.
func TestACopyOfATokenCarriesItsTriggerAndStatic(t *testing.T) {
	withCatalogTriggers(t, func(key string) []TriggeredAbility {
		if key != TokenKey(testTokenSlug) {
			return nil
		}
		return []TriggeredAbility{{Watches: []EventKind{EventLTB}}}
	})
	withStaticAbilities(t, func(key string) []StaticAbility {
		if key != TokenKey(testTokenSlug) {
			return nil
		}
		return []StaticAbility{{Layer: Layer7PT, SubLayer: SubLayer7C_Modify}}
	})

	original := Card{
		InstanceID: uuid.New(),
		Name:       "Pest",
		TypeLine:   "Token Creature — Pest",
		TokenKey:   TokenKey(testTokenSlug),
	}
	clone := Card{InstanceID: uuid.New(), Name: "Clone"}
	clone.applyCopy(CopiableValuesOf(original), original)

	if got := len(TriggersForCard(clone)); got != 1 {
		t.Errorf("a copy of the Pest has %d triggered abilities, want the Pest's one", got)
	}
	if got := len(StaticAbilitiesForCard(clone)); got != 1 {
		t.Errorf("a copy of the Pest has %d static abilities, want the Pest's one", got)
	}

	// The other direction, the one CR 707.2 is usually quoted for: a
	// token that copies a printed card stops printing the token's
	// text, because the oracle ID wins in CatalogKey.
	printed := Card{InstanceID: uuid.New(), Name: "Llanowar Elves", OracleID: "oracle-llanowar"}
	tok := original
	tok.InstanceID = uuid.New()
	tok.applyCopy(CopiableValuesOf(printed), printed)
	if got := len(TriggersForCard(tok)); got != 0 {
		t.Errorf("a token copying Llanowar Elves kept %d of the Pest's triggers", got)
	}
}

// TestUndoAcrossATokenTrigger — the trigger is queued inside the
// destruction, so an undo has to unqueue it and a replay has to queue
// exactly one again.
func TestUndoAcrossATokenTrigger(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := tokenOnBattlefield(g, me.ID, "Pest", 1, 1, testTokenSlug)

	withCatalogTriggers(t, func(key string) []TriggeredAbility {
		if key != TokenKey(testTokenSlug) {
			return nil
		}
		return []TriggeredAbility{{
			Watches: []EventKind{EventLTB},
			AppliesTo: func(ev Event, source *Card, _ Characteristic, _ *Game) bool {
				return ev.CardID == source.InstanceID
			},
			Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
				return &StackItem{
					Kind:         StackItemTriggered,
					SourceCardID: source.InstanceID,
					Controller:   source.Controller,
					Owner:        source.Owner,
					Label:        "Pest — you gain 1 life",
				}
			},
		}}
	})

	before := g.Clone()
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
	if len(g.PendingTriggers) != 1 {
		t.Fatalf("setup: PendingTriggers = %d, want 1", len(g.PendingTriggers))
	}

	g.WithWriteLock(func() { g.RestoreFrom(before) })
	if len(g.PendingTriggers) != 0 {
		t.Fatalf("the trigger survived the undo: %d pending", len(g.PendingTriggers))
	}
	if !g.Battlefield.Contains(id) {
		t.Fatal("the token did not come back")
	}

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatalf("replay DestroyPermanentForEffect: %v", err)
		}
	})
	if len(g.PendingTriggers) != 1 {
		t.Errorf("after the replay: %d pending triggers, want exactly one", len(g.PendingTriggers))
	}
}

// TestTokenTextIsReadFromTheCatalog is the wire's source of truth
// (ADR 0083 decision 6): derived on every read, empty for a printed
// card, and silent for a token that copies one.
func TestTokenTextIsReadFromTheCatalog(t *testing.T) {
	const text = "When this token dies, you gain 1 life."
	prev := CatalogTokenText
	CatalogTokenText = func(key string) string {
		if key != TokenKey(testTokenSlug) {
			return ""
		}
		return text
	}
	t.Cleanup(func() { CatalogTokenText = prev })

	tok := Card{Name: "Pest", TokenKey: TokenKey(testTokenSlug)}
	if got := TokenTextForCard(tok); got != text {
		t.Errorf("token text = %q, want %q", got, text)
	}
	if got := TokenTextForCard(Card{OracleID: "oracle-bears"}); got != "" {
		t.Errorf("a printed card answered %q", got)
	}
	// CR 707.2 again: the oracle ID wins, so a token copy of a
	// printed card renders that card and not the token's text.
	copyOfACard := tok
	copyOfACard.OracleID = "oracle-llanowar"
	if got := TokenTextForCard(copyOfACard); got != "" {
		t.Errorf("a token copying a printed card still answered %q", got)
	}
}

// TestNoBattlefieldWalkGatesOnAnEmptyOracleID is ADR 0083 decision 3
// as a standing rule rather than nine one-off fixes.
//
// Before it, nine statics-gathering walks skipped a permanent whose
// OracleID was empty, which is EVERY TOKEN by construction — so a
// token could hold a catalog key with a block rule, a cast
// permission or "you have no maximum hand size" in it and be walked
// straight past. The question those walks mean to ask is "does this
// object have a catalog entry", and the empty KEY is how that is
// spelled (CatalogKey answers "" for an uncatalogued card and for a
// CR 708.2a face-down permanent alike).
//
// A source scan rather than nine behaviour tests, because the failure
// mode is a TENTH walk written next sprint with the old idiom copied
// from its neighbour.
//
// oracleIDEmptinessIsTheRightQuestion are the places where an oracle
// ID really is what is being asked about, each with the reason. The
// test fails on a set difference in either direction: a new gate is a
// regression, and an entry here that stops matching is a stale
// exemption to delete.
var oracleIDEmptinessIsTheRightQuestion = map[string]string{
	"effect_hooks.go:CatalogKey": "the token-key fallback itself — a composite FACE key " +
		"(\"<oracle>#1\") needs a real oracle ID, and a token has no second face",
	"mutations.go:printedIdentityOf":    "asks whether a real Scryfall PRINTING stands behind the card, not whether the catalog has an entry",
	"mutations.go:fromScryfallPrinting": "same question, and this is the function that names it",
	"snapshot.go:spellSpecRederivable":  "a stack item's oracle ID, gated on StackItemSpell — a token is not a spell",
	"snapshot.go:restoreStackItem":      "the mode-spec twin of spellSpecRederivable, same gate and same reason",
}

func TestNoBattlefieldWalkGatesOnAnEmptyOracleID(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil || len(files) == 0 {
		t.Fatalf("glob: %v (%d files)", err, len(files))
	}
	fset := gotoken.NewFileSet()
	seen := map[string]bool{}
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		af, err := parser.ParseFile(fset, f, src, 0)
		if err != nil {
			t.Fatalf("%s: %v", f, err)
		}
		// Walk declaration by declaration so a hit can be named by
		// the function it is in — a line number would go stale on the
		// next edit above it.
		for _, decl := range af.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			site := f + ":" + fn.Name.Name
			ast.Inspect(fn, func(n ast.Node) bool {
				bin, ok := n.(*ast.BinaryExpr)
				if !ok || (bin.Op != gotoken.EQL && bin.Op != gotoken.NEQ) {
					return true
				}
				sel, ok := bin.X.(*ast.SelectorExpr)
				if !ok || sel.Sel.Name != "OracleID" {
					return true
				}
				lit, ok := bin.Y.(*ast.BasicLit)
				if !ok || lit.Kind != gotoken.STRING || lit.Value != `""` {
					return true
				}
				seen[site] = true
				if _, allowed := oracleIDEmptinessIsTheRightQuestion[site]; allowed {
					return true
				}
				t.Errorf("%s:%d (%s) gates on an empty OracleID, which skips every token. "+
					"Ask the catalog KEY instead — `key := CatalogAbilityKey(*c); if key == \"\"` "+
					"(ADR 0083 decision 3). If an oracle ID really is the question, add the site "+
					"to oracleIDEmptinessIsTheRightQuestion with the reason.",
					f, fset.Position(bin.Pos()).Line, fn.Name.Name)
				return true
			})
		}
	}
	for site := range oracleIDEmptinessIsTheRightQuestion {
		if !seen[site] {
			t.Errorf("%s no longer gates on an empty OracleID: drop it from "+
				"oracleIDEmptinessIsTheRightQuestion", site)
		}
	}
}
