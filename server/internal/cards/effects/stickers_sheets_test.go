package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
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
var stkSheetsNotCatalogued = map[string]string{
	"Squid Fire Knight": "its first ability sticker replaces a dice roll and lets a second player choose a roll to ignore after seeing the dice, " +
		"which every dice-rolling card would have to be able to wait for (they all read the rolls on the next line)",
}

// TestStkEverySheetIsCatalogued: each of the 48 Commander-legal sheets has
// a Full catalog entry that names, per printed ability sticker, a keyword
// or a bundle that is registered.
func TestStkEverySheetIsCatalogued(t *testing.T) {
	if len(stkPrintedSheets) != 48 {
		t.Fatalf("the sheet data has %d sheets, want 48", len(stkPrintedSheets))
	}
	for _, p := range stkPrintedSheets {
		spec, ok := Lookup(p.Oracle)
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
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v, caveats %v", p.Name, spec.Completeness, spec.Caveats)
		}
		if spec.StickerSheet == nil {
			t.Errorf("%s: no StickerSheet", p.Name)
			continue
		}
		sheet := deck.StickerSheetOf(cards.Card{
			OracleID: uuid.MustParse(p.Oracle), Name: p.Name, TypeLine: stickerSheetTypeLine, OracleText: p.Text,
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

// stkScene is a table with one sheet, a creature of the seat's and one
// ability sticker on it.
type stkScene struct {
	t    *testing.T
	g    *game.Game
	me   *game.Player
	opp  *game.Player
	host uuid.UUID
}

// newStkScene puts the ability sticker whose label contains `sticker` from
// `sheet` on a 2/2 Bear the seat owns.
func newStkScene(t *testing.T, sheet, sticker string) *stkScene {
	t.Helper()
	return newStkSceneOn(t, sheet, sticker, "Creature — Bear", 2, 2)
}

// newStkSceneOn is newStkScene with the host's type line and power and
// toughness chosen.
func newStkSceneOn(t *testing.T, sheet, sticker, typeLine string, power, toughness int) *stkScene {
	t.Helper()
	g, me := stkTableWith(t, 12, sheet)
	s := &stkScene{t: t, g: g, me: me, opp: g.Seats[1]}
	s.host = p2Push(g, me.ID, "Host", typeLine, "", power, toughness)
	stkGive(t, g, me, s.host, sticker)
	return s
}

func (s *stkScene) settle() { s.t.Helper(); passPriorityAroundTable(s.t, s.g) }

// attack declares the host as an attacker at the opponent and settles the
// attack triggers (stopping at a prompt).
func (s *stkScene) attack() {
	s.t.Helper()
	declareAttack(s.t, s.g, s.opp.ID, s.host)
	s.settle()
}

// activate activates the host's ability `index` with the mana named.
func (s *stkScene) activate(index int, mana string, targets ...game.TargetRef) {
	s.t.Helper()
	if err := p2Activate(s.t, s.g, s.me, s.host, index, mana, targets...); err != nil {
		s.t.Fatalf("activate %d: %v", index, err)
	}
	s.settle()
}

// destroyHost destroys the host and settles what that triggers.
func (s *stkScene) destroyHost() {
	s.t.Helper()
	b25Destroy(s.g, s.host)
	s.settle()
}

// pickTargets answers the open target question.
func (s *stkScene) pickTargets(targets ...game.TargetRef) {
	s.t.Helper()
	p := latestPickTarget(s.g, s.me.ID)
	if p == nil {
		s.t.Fatal("no target question is open")
	}
	if err := s.g.ResolvePickTargets(p.ID, s.me.ID, targets); err != nil {
		s.t.Fatalf("pick targets: %v", err)
	}
	s.settle()
}

func (s *stkScene) card() game.Card { s.t.Helper(); return stkCard(s.t, s.g, s.host) }

func (s *stkScene) has(kw string) bool { return p2Keyword(s.g, s.host, kw) }

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
			s := newStkScene(t, tc.sheet, tc.sticker)
			for _, kw := range tc.keywords {
				if !s.has(kw) {
					t.Errorf("no %s from the sticker; abilities %v", kw, s.card().Effective().Abilities)
				}
			}
		})
	}
}

// Prowess, prowess and exalted, exalted are two instances each (CR 113.2c).
func TestStkRepeatedKeywordStickersGiveTwoInstances(t *testing.T) {
	y := newStkScene(t, "Cursed Firebreathing Yogurt", "Prowess, prowess")
	if n := stkAbilityCount(y.g, y.host, game.KeywordProwess); n != 2 {
		t.Errorf("prowess instances = %d, want 2", n)
	}
	c := newStkScene(t, "Urza's Dark Cannonball", "Exalted, exalted")
	if n := stkAbilityCount(c.g, c.host, game.KeywordExalted); n != 2 {
		t.Errorf("exalted instances = %d, want 2", n)
	}
}

// Ancestral Hot Dog Minotaur: afflict 2.
func TestStkHotDogMinotaurAfflict(t *testing.T) {
	s := newStkScene(t, "Ancestral Hot Dog Minotaur", "Afflict 2")
	wall := b12Creature(s.g, s.opp.ID, "Wall", "Creature — Wall", 0, 4)
	unblockedAttackIntoBlockers(t, s.g, s.host)
	life := s.opp.Life
	if err := s.g.DeclareBlocker(wall, s.host); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	if err := s.g.FinishBlocks(s.opp.ID); err != nil {
		t.Fatalf("FinishBlocks: %v", err)
	}
	s.settle()
	if got := life - s.opp.Life; got != 2 {
		t.Errorf("the defender lost %d life to afflict 2", got)
	}
}

// Afflict does nothing when the creature is not blocked.
func TestStkHotDogMinotaurAfflictNeedsABlock(t *testing.T) {
	s := newStkScene(t, "Ancestral Hot Dog Minotaur", "Afflict 2")
	b12Creature(s.g, s.opp.ID, "Wall", "Creature — Wall", 0, 4)
	unblockedAttackIntoBlockers(t, s.g, s.host)
	life := s.opp.Life
	if err := s.g.FinishBlocks(s.opp.ID); err != nil {
		t.Fatalf("FinishBlocks: %v", err)
	}
	s.settle()
	if s.opp.Life != life {
		t.Errorf("the defender lost life without blocking: %d to %d", life, s.opp.Life)
	}
}

// Carnival Elephant Meteor: "Sacrifice this permanent: Draw two cards."
func TestStkCarnivalElephantSacrificeDraw(t *testing.T) {
	s := newStkScene(t, "Carnival Elephant Meteor", "Sacrifice this permanent")
	hand := s.me.Hand.Size()
	s.activate(0, "")
	if onBattlefield(s.g, s.host) {
		t.Error("the permanent was not sacrificed")
	}
	if got := s.me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d cards, want 2", got)
	}
}

// ... and "Whenever this creature attacks, proliferate."
func TestStkCarnivalElephantAttackProliferates(t *testing.T) {
	s := newStkScene(t, "Carnival Elephant Meteor", "proliferate")
	g := s.g
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(s.host, game.CounterPlusOne, 1) })
	declareAttack(t, g, s.opp.ID, s.host)
	passPriorityAroundTable(t, g)
	answerProliferate(t, g, s.host)
	if got := countersOn(g, s.host, game.CounterPlusOne); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2 after proliferate", got)
	}
}

// Contortionist Otter Storm: "{T}: Target creature gains haste until end of turn."
func TestStkContortionistOtterGivesHaste(t *testing.T) {
	s := newStkScene(t, "Contortionist Otter Storm", "gains haste")
	other := p2Push(s.g, s.me.ID, "Other", "Creature — Bear", "", 1, 1)
	if p2Keyword(s.g, other, "haste") {
		t.Fatal("setup: the other creature already has haste")
	}
	s.activate(0, "", p2Card(other))
	if !p2Keyword(s.g, other, "haste") {
		t.Error("the target did not gain haste")
	}
}

// Cool Fluffy Loxodon: leaves-the-battlefield draw.
func TestStkCoolFluffyLoxodonDrawsWhenItLeaves(t *testing.T) {
	s := newStkSceneOn(t, "Cool Fluffy Loxodon", "leaves the battlefield", "Artifact", 0, 0)
	hand := s.me.Hand.Size()
	s.destroyHost()
	if got := s.me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
}

// ... and the 13/13: a noncreature permanent becomes an Eldrazi creature in
// addition to its other types for the turn.
func TestStkCoolFluffyLoxodonBecomesA1313(t *testing.T) {
	s := newStkSceneOn(t, "Cool Fluffy Loxodon", "13/13", "Artifact", 0, 0)
	castAndResolveCreature(t, s.g, "Grizzly Bears", "Creature — Bear", "")
	s.settle()
	ch := effectiveOf(t, s.g, s.host)
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
	s := newStkScene(t, "Cursed Firebreathing Yogurt", "deals 2 damage")
	life := s.opp.Life
	s.activate(0, "RR", game.TargetRef{Kind: game.TargetPlayer, ID: s.opp.ID})
	if got := life - s.opp.Life; got != 2 {
		t.Errorf("the opponent lost %d, want 2", got)
	}
}

// Deep-Fried Plague Myr: attack scries 1.
func TestStkPlagueMyrAttackScries(t *testing.T) {
	s := newStkScene(t, "Deep-Fried Plague Myr", "scry 1")
	seedLibrary(s.me, "A", "B", "C")
	declareAttack(t, s.g, s.opp.ID, s.host)
	passPriorityAroundTable(t, s.g)
	if scryChoiceFor(s.g, s.me.ID) == nil {
		t.Error("the attack trigger asked no scry")
	}
}

// ... and the leaves-the-battlefield "you may destroy target artifact or enchantment".
func TestStkPlagueMyrLeavesDestroys(t *testing.T) {
	s := newStkScene(t, "Deep-Fried Plague Myr", "destroy target artifact or enchantment")
	rock := p2Push(s.g, s.opp.ID, "Rock", "Artifact", "", 0, 0)
	bear := p2Push(s.g, s.opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	b25Destroy(s.g, s.host)
	p := latestPickTarget(s.g, s.me.ID)
	if p == nil {
		t.Fatal("no target question")
	}
	if hasID(p.PickTargetCards, bear) || !hasID(p.PickTargetCards, rock) {
		t.Errorf("legal targets %v: want the artifact only", p.PickTargetCards)
	}
	if err := s.g.ResolvePickTargets(p.ID, s.me.ID, []game.TargetRef{p2Card(rock)}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	answerLatestTriggerPrompt(t, s.g, s.me.ID, true)
	s.settle()
	if onBattlefield(s.g, rock) {
		t.Error("the artifact was not destroyed")
	}
}
