package testscm

import (
	"errors"
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// commander_color_cards_test.go — CR 903.4b (#5948): The Prismatic Piper,
// Faceless One and Clara Oswald, driven through the real pregame: decks
// seated, the game started, every hand kept.

const (
	ccPiperOracle    = "b2aeee90-7b8d-47a6-9167-cc21f7fbfd43"
	ccFacelessOracle = "ef220170-c01a-4ca5-b887-f248c50227ee"
	ccClaraOracle    = "542eae01-058d-4959-92a8-49d6b1e33ecd"
)

// ccCommander is a CR 903.4b commander as the deck importer stamps it:
// colourless, with an empty imported identity.
func ccCommander(name, oracle string) game.Card {
	c := game.NewCard(name, uuid.Nil)
	c.TypeLine = "Legendary Creature — Shapeshifter"
	c.ManaCost = "{5}"
	c.Power, c.Toughness = 3, 3
	c.OracleID = oracle
	c.ScryfallID = uuid.NewString()
	c.ColorIdentity = []string{}
	c.IsCommander = true
	return c
}

// ccSpell is a deck card of the given colour identity (none for a
// colourless one).
func ccSpell(name string, identity ...string) game.Card {
	c := game.NewCard(name, uuid.Nil)
	c.TypeLine = "Artifact"
	c.ManaCost = "{2}"
	c.ScryfallID = uuid.NewString()
	c.OracleID = uuid.NewString()
	c.ColorIdentity = identity
	return c
}

// ccTable seats two players, seat 0 with the given commander and deck,
// seat 1 with a plain colourless deck, starts the game and returns it with
// the mulligan window still open.
func ccTable(t *testing.T, commander game.Card, deck []game.Card) (*game.Game, uuid.UUID) {
	t.Helper()
	g := game.NewGame()
	commander.InstanceID = uuid.New()
	seat0 := append([]game.Card{commander}, deck...)
	if _, err := g.AddPlayer("A", seat0); err != nil {
		t.Fatalf("AddPlayer A: %v", err)
	}
	other := make([]game.Card, 20)
	for i := range other {
		other[i] = ccSpell("Filler")
	}
	if _, err := g.AddPlayer("B", other); err != nil {
		t.Fatalf("AddPlayer B: %v", err)
	}
	if err := g.Start(rand.New(rand.NewPCG(5, 9))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	return g, commander.InstanceID
}

func ccKeepAll(t *testing.T, g *game.Game) {
	t.Helper()
	for i := range g.Seats {
		if err := g.KeepHand(g.Seats[(g.StartingSeat+i)%len(g.Seats)].ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
}

func ccDeck(n int, identity ...string) []game.Card {
	out := make([]game.Card, n)
	for i := range out {
		out[i] = ccSpell("Deck card", identity...)
	}
	return out
}

func ccColorPrompts(g *game.Game) []*game.PendingChoice {
	var out []*game.PendingChoice
	g.ReadSnapshot(func() {
		for _, c := range g.PendingChoices {
			if c.Kind == game.PendingChoiceColor {
				out = append(out, c)
			}
		}
	})
	return out
}

// ccCard finds a card by ID in the owner's zones or on the battlefield.
func ccCard(g *game.Game, owner, id uuid.UUID) (game.Card, bool) {
	var out game.Card
	var ok bool
	g.ReadSnapshot(func() {
		for _, p := range g.Seats {
			if p.ID != owner {
				continue
			}
			for _, z := range []*game.Zone{p.Command, p.Hand, p.Library, p.Graveyard, g.Battlefield, g.Exile} {
				for i := range z.Cards {
					if z.Cards[i].InstanceID == id {
						out, ok = z.Cards[i], true
						return
					}
				}
			}
		}
	})
	return out, ok
}

func ccIdentity(g *game.Game, player uuid.UUID) []string {
	var out []string
	g.ReadSnapshot(func() { out, _ = g.CommanderIdentityColorsForEffect(player) })
	return out
}

// A deck whose other cards are red chose red at deck construction (CR
// 903.4b): the Piper is red as the game begins, nobody is asked, and the
// choice is revealed in the history.
func TestPrismaticPiperOfAMonoRedDeckIsRedWithoutBeingAsked(t *testing.T) {
	g, piper := ccTable(t, ccCommander("The Prismatic Piper", ccPiperOracle), ccDeck(20, "R"))
	me := g.Seats[0].ID
	if c, _ := ccCard(g, me, piper); !c.IsColorless() {
		t.Fatalf("the Piper is %v before the mulligan window closed, want colourless", c.EffectiveColors())
	}
	from := len(g.Events)
	ccKeepAll(t, g)

	if got := ccColorPrompts(g); len(got) != 0 {
		t.Fatalf("asked %d colour prompts; the deck already chose red", len(got))
	}
	c, ok := ccCard(g, me, piper)
	if !ok || c.CommanderColor != "R" || !slices.Equal(c.EffectiveColors(), []string{"R"}) {
		t.Fatalf("Piper: CommanderColor %q, colours %v; want red", c.CommanderColor, c.EffectiveColors())
	}
	if got := ccIdentity(g, me); !slices.Equal(got, []string{"R"}) {
		t.Errorf("commander identity = %v, want [R] (CR 903.4b: the choice affects the identity)", got)
	}
	revealed := false
	for _, ev := range g.Events[from:] {
		if ev.Kind == game.EventColorChosen && ev.CardID == piper && ev.Label == "R" && ev.Actor == me {
			revealed = true
		}
	}
	if !revealed {
		t.Error("the chosen colour was not revealed in the history")
	}
	if err := g.PassPriority(); errors.Is(err, game.ErrChoicePending) {
		t.Error("the table is parked with nothing to answer")
	}
}

// A colourless deck shows no choice, so the owner is asked, with all five
// colours, and the table waits on the answer. The colour holds in every
// zone the commander goes to.
func TestPrismaticPiperOfAColourlessDeckAsksItsOwnerAndKeepsTheColour(t *testing.T) {
	g, piper := ccTable(t, ccCommander("The Prismatic Piper", ccPiperOracle), ccDeck(20))
	me, them := g.Seats[0].ID, g.Seats[1].ID
	ccKeepAll(t, g)

	prompts := ccColorPrompts(g)
	if len(prompts) != 1 {
		t.Fatalf("colour prompts = %d, want 1", len(prompts))
	}
	p := prompts[0]
	if p.Chooser != me || p.Source != piper || !slices.Equal(p.ColorOptions, game.AllColors) {
		t.Fatalf("prompt: chooser %s source %s options %v; want seat 0, the Piper, all five", p.Chooser, p.Source, p.ColorOptions)
	}
	if err := g.PassPriority(); !errors.Is(err, game.ErrChoicePending) {
		t.Errorf("PassPriority with the prompt open: err = %v, want ErrChoicePending", err)
	}
	if err := g.ResolveColorChoice(p.ID, them, "G"); !errors.Is(err, game.ErrNotTheChooser) {
		t.Errorf("the other seat answered: err = %v", err)
	}
	if err := g.ResolveColorChoice(p.ID, me, "G"); err != nil {
		t.Fatalf("answer green: %v", err)
	}
	c, _ := ccCard(g, me, piper)
	if c.CommanderColor != "G" || !c.HasColor("G") || len(c.EffectiveColors()) != 1 {
		t.Fatalf("Piper in the command zone: %q %v, want green", c.CommanderColor, c.EffectiveColors())
	}
	if got := ccIdentity(g, me); !slices.Equal(got, []string{"G"}) {
		t.Errorf("commander identity = %v, want [G]", got)
	}

	if err := g.PassPriority(); errors.Is(err, game.ErrChoicePending) {
		t.Error("the table is still parked after the answer")
	}

	// CR 903.4b: "even as the commander changes zones". (A move to the
	// graveyard or exile opens CR 903.9a's command-zone question, which
	// this test leaves unanswered.)
	for _, dst := range []game.ZoneKind{game.ZoneBattlefield, game.ZoneGraveyard, game.ZoneExile} {
		src, _ := ccZoneOf(g, me, piper)
		if err := g.MoveCardByID(ccRef(src, me), ccRef(dst, me), piper); err != nil {
			t.Fatalf("move to %s: %v", dst, err)
		}
		c, ok := ccCard(g, me, piper)
		var colors []string
		g.ReadSnapshot(func() {
			if dst == game.ZoneBattlefield {
				colors = g.Battlefield.Cards[len(g.Battlefield.Cards)-1].Effective().Colors
			} else {
				colors = c.EffectiveColors()
			}
		})
		if !ok || !slices.Equal(colors, []string{"G"}) {
			t.Errorf("in %s the Piper is %v, want green", dst, colors)
		}
	}
}

// ccRef is a zone ref: a shared zone has no owner.
func ccRef(kind game.ZoneKind, owner uuid.UUID) game.ZoneRef {
	if kind == game.ZoneBattlefield || kind == game.ZoneExile {
		return game.ZoneRef{Kind: kind}
	}
	return game.ZoneRef{Kind: kind, Owner: owner}
}

func ccZoneOf(g *game.Game, owner, id uuid.UUID) (game.ZoneKind, bool) {
	var kind game.ZoneKind
	var ok bool
	g.ReadSnapshot(func() {
		p := g.Seats[0]
		for _, s := range g.Seats {
			if s.ID == owner {
				p = s
			}
		}
		for k, z := range map[game.ZoneKind]*game.Zone{
			game.ZoneCommand: p.Command, game.ZoneHand: p.Hand, game.ZoneLibrary: p.Library,
			game.ZoneGraveyard: p.Graveyard, game.ZoneBattlefield: g.Battlefield, game.ZoneExile: g.Exile,
		} {
			if z.Contains(id) {
				kind, ok = k, true
			}
		}
	})
	return kind, ok
}

// "If The Prismatic Piper is your commander": a Piper that is not a
// commander asks nothing and is colourless.
func TestPrismaticPiperThatIsNotACommanderIsColourless(t *testing.T) {
	notCommander := ccCommander("The Prismatic Piper", ccPiperOracle)
	notCommander.IsCommander = false
	notCommander.InstanceID = uuid.New()
	deck := append(ccDeck(19, "U"), notCommander)
	g, _ := ccTable(t, ccSpell("Not a commander"), deck)
	ccKeepAll(t, g)
	if got := ccColorPrompts(g); len(got) != 0 {
		t.Fatalf("asked %d colour prompts with no CR 903.4b commander", len(got))
	}
	c, ok := ccCard(g, g.Seats[0].ID, notCommander.InstanceID)
	if !ok || c.CommanderColor != "" || !c.IsColorless() {
		t.Errorf("a non-commander Piper: %q %v, want colourless", c.CommanderColor, c.EffectiveColors())
	}
}

// Faceless One reads its deck the same way: a mono-white deck makes it
// white.
func TestFacelessOneOfAMonoWhiteDeckIsWhite(t *testing.T) {
	g, id := ccTable(t, ccCommander("Faceless One", ccFacelessOracle), ccDeck(20, "W"))
	ccKeepAll(t, g)
	if got := ccColorPrompts(g); len(got) != 0 {
		t.Fatalf("asked %d prompts", len(got))
	}
	c, _ := ccCard(g, g.Seats[0].ID, id)
	if !slices.Equal(c.EffectiveColors(), []string{"W"}) {
		t.Errorf("Faceless One is %v, want white", c.EffectiveColors())
	}
}

// With a partner, the deck's colours outside the partner's identity are
// the choice: Clara with a blue Doctor in a blue-black deck is black.
func TestClaraOswaldWithADoctorTakesTheColourOutsideTheDoctors(t *testing.T) {
	doctor := game.NewCard("A Doctor", uuid.Nil)
	doctor.TypeLine = "Legendary Creature — Time Lord Doctor"
	doctor.ManaCost = "{2}{U}"
	doctor.ColorIdentity = []string{"U"}
	doctor.ScryfallID, doctor.OracleID = uuid.NewString(), uuid.NewString()
	doctor.IsCommander = true
	deck := append([]game.Card{doctor}, ccDeck(10, "U")...)
	deck = append(deck, ccDeck(10, "B")...)
	g, clara := ccTable(t, ccCommander("Clara Oswald", ccClaraOracle), deck)
	ccKeepAll(t, g)
	if got := ccColorPrompts(g); len(got) != 0 {
		t.Fatalf("asked %d prompts", len(got))
	}
	c, _ := ccCard(g, g.Seats[0].ID, clara)
	if c.CommanderColor != "B" {
		t.Errorf("Clara chose %q, want B", c.CommanderColor)
	}
	if got := ccIdentity(g, g.Seats[0].ID); !slices.Equal(got, []string{"B", "U"}) && !slices.Equal(got, []string{"U", "B"}) {
		t.Errorf("combined identity = %v, want blue and black", got)
	}
}

// A deck that needs no colour outside the Doctor's offers every colour.
func TestClaraOswaldWithADoctorInItsOwnColourIsAsked(t *testing.T) {
	doctor := game.NewCard("A Doctor", uuid.Nil)
	doctor.TypeLine = "Legendary Creature — Time Lord Doctor"
	doctor.ManaCost = "{2}{U}"
	doctor.ColorIdentity = []string{"U"}
	doctor.ScryfallID, doctor.OracleID = uuid.NewString(), uuid.NewString()
	doctor.IsCommander = true
	deck := append([]game.Card{doctor}, ccDeck(20, "U")...)
	g, clara := ccTable(t, ccCommander("Clara Oswald", ccClaraOracle), deck)
	ccKeepAll(t, g)
	prompts := ccColorPrompts(g)
	if len(prompts) != 1 || prompts[0].Source != clara || len(prompts[0].ColorOptions) != 5 {
		t.Fatalf("prompts %+v, want one for Clara with all five colours", prompts)
	}
	if err := g.ResolveColorChoice(prompts[0].ID, g.Seats[0].ID, "U"); err != nil {
		t.Fatalf("answer: %v", err)
	}
	c, _ := ccCard(g, g.Seats[0].ID, clara)
	if !slices.Equal(c.EffectiveColors(), []string{"U"}) {
		t.Errorf("Clara is %v, want blue", c.EffectiveColors())
	}
}

// "If a triggered ability of a Doctor you control triggers, that ability
// triggers an additional time." A Young Pyromancer that is a Doctor makes
// two Elementals per instant with Clara out; one that is not makes one.
func TestClaraOswaldDoublesTheTriggersOfADoctorYouControl(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		want           int
	}{
		{"a Doctor", "Creature — Time Lord Doctor", 2},
		{"not a Doctor", "Creature — Human Shaman", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.PushCatalogPermanent(g, me.ID, "Clara Oswald", "Legendary Creature — Human Advisor", ccClaraOracle, false)
			effectstest.PushCatalogPermanent(g, me.ID, "Young Pyromancer", tc.typeLine, effectstest.B07YoungPyromancerOracle, false)
			effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
				[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.Gl1bPyromancerTokens(g, me.ID); got != tc.want {
				t.Errorf("Elementals = %d, want %d", got, tc.want)
			}
		})
	}
}

// "A Doctor you control": an opponent's Doctor is not doubled.
func TestClaraOswaldDoesNotDoubleAnOpponentsDoctor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Clara Oswald", "Legendary Creature — Human Advisor", ccClaraOracle, false)
	effectstest.PushCatalogPermanent(g, opp.ID, "Young Pyromancer", "Creature — Time Lord Doctor", effectstest.B07YoungPyromancerOracle, false)
	g.WithWriteLock(func() { g.Turn.ActiveSeat = 1 })
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.Gl1bPyromancerTokens(g, opp.ID); got != 1 {
		t.Errorf("opponent Elementals = %d, want 1", got)
	}
}
