package testsst

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// stickers_sheets_test.go — the ability stickers of the Unfinity sticker
// sheets (ADR 0208, #4307): every legal sheet is catalogued, says what each
// of its ability stickers gives, and the ability works when the sticker is
// on a permanent.

// stkSheetsNotCatalogued are the sheets with no entry, and why. A sheet
// with an ability sticker the engine cannot run is left out whole rather
// than shipped weaker than printed (AGENTS.md section 7): its stickers go on
// and do nothing, and the deck checker lists the sheet as not automated.
//
// Empty since Squid Fire Knight's dice replacement shipped (#4307).
var stkSheetsNotCatalogued = map[string]string{}

// TestStkEverySheetIsCatalogued: each of the 48 Commander-legal sheets has
// a Full catalog entry that names, per printed ability sticker, a keyword
// or a bundle that is registered.
func TestStkEverySheetIsCatalogued(t *testing.T) {
	if len(effectstest.StkPrintedSheets) != 48 {
		t.Fatalf("the sheet data has %d sheets, want 48", len(effectstest.StkPrintedSheets))
	}
	for _, p := range effectstest.StkPrintedSheets {
		spec, ok := effects.Lookup(p.Oracle)
		if why, left := stkSheetsNotCatalogued[p.Name]; left {
			if ok {
				t.Errorf("%s is catalogued now: drop it from stkSheetsNotCatalogued (%s)", p.Name, why)
			}
			continue
		}
		if !ok {
			t.Errorf("%s: no catalog entry", p.Name)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v, caveats %v", p.Name, spec.Completeness, spec.Caveats)
		}
		if spec.StickerSheet == nil {
			t.Errorf("%s: no StickerSheet", p.Name)
			continue
		}
		sheet := deck.StickerSheetOf(cards.Card{
			OracleID: uuid.MustParse(p.Oracle), Name: p.Name, TypeLine: effectstest.StickerSheetTypeLine, OracleText: p.Text,
		})
		abilities := 0
		for _, s := range sheet.Stickers {
			if s.Kind == game.StickerAbility {
				abilities++
			}
		}
		if len(spec.StickerSheet.Abilities) != abilities {
			t.Errorf("%s: the entry names %d ability stickers, the sheet prints %d", p.Name, len(spec.StickerSheet.Abilities), abilities)
		}
		for i := range spec.StickerSheet.Abilities {
			def, ok := game.CatalogStickerAbility(p.Oracle, i)
			if !ok {
				t.Errorf("%s ability sticker %d: not catalogued", p.Name, i)
				continue
			}
			if len(def.Keywords) == 0 && def.Grant == "" && def.GraveyardCastLife == 0 {
				t.Errorf("%s ability sticker %d gives nothing", p.Name, i)
			}
			for _, kw := range def.Keywords {
				if _, ok := game.CanonicalKeywords(kw); !ok {
					t.Errorf("%s ability sticker %d: %q is not a keyword the engine enforces", p.Name, i, kw)
				}
			}
		}
	}
}

func stkAbilityCount(g *game.Game, id uuid.UUID, kw string) int {
	n := 0
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		if c, ok := g.LookupCardForEffect(id); ok {
			for _, a := range c.Effective().Abilities {
				if a == kw {
					n++
				}
			}
		}
	})
	return n
}

// The sticker that is only keywords gives each of them.
func TestStkKeywordOnlyAbilityStickers(t *testing.T) {
	for _, tc := range []struct {
		sheet, sticker string
		keywords       []string
	}{
		{"Ancestral Hot Dog Minotaur", "Flying", []string{"flying"}},
		{"Contortionist Otter Storm", "Deathtouch, lifelink", []string{"deathtouch", "lifelink"}},
		{"Eldrazi Guacamole Tightrope", "Haste", []string{"haste"}},
		{"Happy Dead Squirrel", "Infect", []string{game.KeywordInfect}},
		{"Jetpack Death Seltzer", "Trample", []string{"trample"}},
		{"Misunderstood Trapeze Elf", "Hexproof", []string{"hexproof"}},
		{"Mystic Doom Sandwich", "Lifelink", []string{"lifelink"}},
		{"Narrow-Minded Baloney Fireworks", "Vigilance, reach", []string{"vigilance", "reach"}},
		{"Slimy Burrito Illusion", "Double strike", []string{"double strike"}},
		{"Space Fungus Snickerdoodle", "Skulk", []string{"skulk"}},
		{"Spooky Clown Mox", "Vigilance", []string{"vigilance"}},
		{"Trendy Circus Pirate", "Deathtouch", []string{"deathtouch"}},
		{"Unsanctioned Ancient Juggler", "Indestructible", []string{"indestructible"}},
		{"Urza's Dark Cannonball", "Shadow", []string{"shadow"}},
		{"Zombie Cheese Magician", "First strike", []string{"first strike"}},
	} {
		t.Run(tc.sheet, func(t *testing.T) {
			s := effectstest.NewStkScene(t, tc.sheet, tc.sticker)
			for _, kw := range tc.keywords {
				if !s.Has(kw) {
					t.Errorf("no %s from the sticker; abilities %v", kw, s.Card().Effective().Abilities)
				}
			}
		})
	}
}

// Prowess, prowess and exalted, exalted are two instances each (CR 113.2c).
func TestStkRepeatedKeywordStickersGiveTwoInstances(t *testing.T) {
	y := effectstest.NewStkScene(t, "Cursed Firebreathing Yogurt", "Prowess, prowess")
	if n := stkAbilityCount(y.G, y.Host, game.KeywordProwess); n != 2 {
		t.Errorf("prowess instances = %d, want 2", n)
	}
	c := effectstest.NewStkScene(t, "Urza's Dark Cannonball", "Exalted, exalted")
	if n := stkAbilityCount(c.G, c.Host, game.KeywordExalted); n != 2 {
		t.Errorf("exalted instances = %d, want 2", n)
	}
}

// Ancestral Hot Dog Minotaur: afflict 2.
func TestStkHotDogMinotaurAfflict(t *testing.T) {
	s := effectstest.NewStkScene(t, "Ancestral Hot Dog Minotaur", "Afflict 2")
	wall := effectstest.B12Creature(s.G, s.Opp.ID, "Wall", "Creature — Wall", 0, 4)
	effectstest.UnblockedAttackIntoBlockers(t, s.G, s.Host)
	life := s.Opp.Life
	if err := s.G.DeclareBlocker(wall, s.Host); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	if err := s.G.FinishBlocks(s.Opp.ID); err != nil {
		t.Fatalf("FinishBlocks: %v", err)
	}
	s.Settle()
	if got := life - s.Opp.Life; got != 2 {
		t.Errorf("the defender lost %d life to afflict 2", got)
	}
}

// Afflict does nothing when the creature is not blocked.
func TestStkHotDogMinotaurAfflictNeedsABlock(t *testing.T) {
	s := effectstest.NewStkScene(t, "Ancestral Hot Dog Minotaur", "Afflict 2")
	effectstest.B12Creature(s.G, s.Opp.ID, "Wall", "Creature — Wall", 0, 4)
	effectstest.UnblockedAttackIntoBlockers(t, s.G, s.Host)
	life := s.Opp.Life
	if err := s.G.FinishBlocks(s.Opp.ID); err != nil {
		t.Fatalf("FinishBlocks: %v", err)
	}
	s.Settle()
	if s.Opp.Life != life {
		t.Errorf("the defender lost life without blocking: %d to %d", life, s.Opp.Life)
	}
}

// Carnival Elephant Meteor: "Sacrifice this permanent: Draw two cards."
func TestStkCarnivalElephantSacrificeDraw(t *testing.T) {
	s := effectstest.NewStkScene(t, "Carnival Elephant Meteor", "Sacrifice this permanent")
	hand := s.Me.Hand.Size()
	s.Activate(0, "")
	if effects.OnBattlefield(s.G, s.Host) {
		t.Error("the permanent was not sacrificed")
	}
	if got := s.Me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d cards, want 2", got)
	}
}

// ... and "Whenever this creature attacks, proliferate."
func TestStkCarnivalElephantAttackProliferates(t *testing.T) {
	s := effectstest.NewStkScene(t, "Carnival Elephant Meteor", "proliferate")
	g := s.G
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(s.Host, game.CounterPlusOne, 1) })
	effectstest.DeclareAttack(t, g, s.Opp.ID, s.Host)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerProliferate(t, g, s.Host)
	if got := effectstest.CountersOn(g, s.Host, game.CounterPlusOne); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2 after proliferate", got)
	}
}

// Contortionist Otter Storm: "{T}: Target creature gains haste until end of turn."
func TestStkContortionistOtterGivesHaste(t *testing.T) {
	s := effectstest.NewStkScene(t, "Contortionist Otter Storm", "gains haste")
	other := effectstest.P2Push(s.G, s.Me.ID, "Other", "Creature — Bear", "", 1, 1)
	if effectstest.P2Keyword(s.G, other, "haste") {
		t.Fatal("setup: the other creature already has haste")
	}
	s.Activate(0, "", effectstest.P2Card(other))
	if !effectstest.P2Keyword(s.G, other, "haste") {
		t.Error("the target did not gain haste")
	}
}

// Cool Fluffy Loxodon: leaves-the-battlefield draw.
func TestStkCoolFluffyLoxodonDrawsWhenItLeaves(t *testing.T) {
	s := effectstest.NewStkSceneOn(t, "Cool Fluffy Loxodon", "leaves the battlefield", "Artifact", 0, 0)
	hand := s.Me.Hand.Size()
	s.DestroyHost()
	if got := s.Me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
}

// ... and the 13/13: a noncreature permanent becomes an Eldrazi creature in
// addition to its other types for the turn.
func TestStkCoolFluffyLoxodonBecomesA1313(t *testing.T) {
	s := effectstest.NewStkSceneOn(t, "Cool Fluffy Loxodon", "13/13", "Artifact", 0, 0)
	effectstest.CastAndResolveCreature(t, s.G, "Grizzly Bears", "Creature — Bear", "")
	s.Settle()
	ch := effectstest.EffectiveOf(t, s.G, s.Host)
	if ch.Power != 13 || ch.Toughness != 13 {
		t.Errorf("P/T = %d/%d, want 13/13", ch.Power, ch.Toughness)
	}
	hasType := func(types []string, want string) bool {
		for _, ty := range types {
			if strings.EqualFold(ty, want) {
				return true
			}
		}
		return false
	}
	if !hasType(ch.Types, "Creature") || !hasType(ch.Types, "Artifact") || !hasType(ch.Subtypes, "Eldrazi") {
		t.Errorf("types %v subtypes %v, want Artifact Creature — Eldrazi", ch.Types, ch.Subtypes)
	}
}

// Cursed Firebreathing Yogurt: "{2}, {T}: This permanent deals 2 damage to any target."
func TestStkCursedYogurtDeals2(t *testing.T) {
	s := effectstest.NewStkScene(t, "Cursed Firebreathing Yogurt", "deals 2 damage")
	life := s.Opp.Life
	s.Activate(0, "RR", game.TargetRef{Kind: game.TargetPlayer, ID: s.Opp.ID})
	if got := life - s.Opp.Life; got != 2 {
		t.Errorf("the opponent lost %d, want 2", got)
	}
}

// Deep-Fried Plague Myr: attack scries 1.
func TestStkPlagueMyrAttackScries(t *testing.T) {
	s := effectstest.NewStkScene(t, "Deep-Fried Plague Myr", "scry 1")
	effectstest.SeedLibrary(s.Me, "A", "B", "C")
	effectstest.DeclareAttack(t, s.G, s.Opp.ID, s.Host)
	effectstest.PassPriorityAroundTable(t, s.G)
	if effectstest.ScryChoiceFor(s.G, s.Me.ID) == nil {
		t.Error("the attack trigger asked no scry")
	}
}

// ... and the leaves-the-battlefield "you may destroy target artifact or enchantment".
func TestStkPlagueMyrLeavesDestroys(t *testing.T) {
	s := effectstest.NewStkScene(t, "Deep-Fried Plague Myr", "destroy target artifact or enchantment")
	rock := effectstest.P2Push(s.G, s.Opp.ID, "Rock", "Artifact", "", 0, 0)
	bear := effectstest.P2Push(s.G, s.Opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.B25Destroy(s.G, s.Host)
	p := effectstest.LatestPickTarget(s.G, s.Me.ID)
	if p == nil {
		t.Fatal("no target question")
	}
	if effectstest.HasID(p.PickTargetCards, bear) || !effectstest.HasID(p.PickTargetCards, rock) {
		t.Errorf("legal targets %v: want the artifact only", p.PickTargetCards)
	}
	if err := s.G.ResolvePickTargets(p.ID, s.Me.ID, []game.TargetRef{effectstest.P2Card(rock)}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	effectstest.AnswerLatestTriggerPrompt(t, s.G, s.Me.ID, true)
	s.Settle()
	if effects.OnBattlefield(s.G, rock) {
		t.Error("the artifact was not destroyed")
	}
}
