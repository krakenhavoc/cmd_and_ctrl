package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cascade_test.go — S28 sub-PR 3. The engine's exile / bottom / prompt
// behaviour is pinned in game/cascade_test.go; these tests are about
// the DECLARATIONS: that the keyword reaches the stack from a spell,
// that "cascade, cascade" is two triggers and not one, and that the
// two granting cards hand it to the right spells.

const (
	BloodbraidElfOracle    = "3f0c9466-5ab9-4205-a84f-b4b27b5a678e"
	ShardlessAgentOracle   = "2afbaa9a-c171-4a8b-90f3-5250d8498356"
	MaelstromWandererOracl = "ad9b7fbc-61c8-43ee-a65c-99206fd1e4df"
	MaelstromNexusOracle   = "b1c346a4-fc9b-474a-9d51-cac168e363fa"
	ImotiOracle            = "eca32dcd-6845-433e-a631-ed1f0ee78f25"
)

// castWithCost seeds a spell with a real mana cost into the active
// seat's hand, walks to a main phase, and casts it.
func CastWithCost(t *testing.T, g *game.Game, name, typeLine, manaCost, oracle string) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: manaCost,
		OracleID: oracle, Owner: active.ID, Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// seedCheapLibrary replaces the active seat's library with a deck the
// cascade will hit: `n` cheap instants under a pile of lands.
func SeedCheapLibrary(g *game.Game, p *game.Player, cheap int) {
	p.Library.Cards = nil
	for i := 0; i < cheap; i++ {
		c := game.NewCard("Cheap Spell", p.ID)
		c.TypeLine = "Instant"
		c.ManaCost = "{U}"
		p.Library.PushTop(c)
	}
}

// isKeywordCastPrompt reports whether c is an offer to cast one card: a
// may_cast prompt, or the resolution_cast prompt cascade and discover
// (#4243), and rebound, suspend, madness and the other keyword and
// effect offers (#4250), ask. A keyword's resolution_cast names its
// keyword; a copy's does not.
func IsKeywordCastPrompt(c *game.PendingChoice) bool {
	if c == nil {
		return false
	}
	if c.Kind == game.PendingChoiceMayCast {
		return true
	}
	return c.Kind == game.PendingChoiceResolutionCast && c.MayCastKeyword != ""
}

// keywordCastPromptFor is chooser's open keyword cast, or nil.
func KeywordCastPromptFor(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if IsKeywordCastPrompt(c) && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

// offeredZone is the zone a resolution_cast's card is cast from: exile
// for most, the library for Ziatora's Envoy's top card.
func OfferedZone(g *game.Game, chooser, card uuid.UUID) game.ZoneKind {
	zone := game.ZoneExile
	g.WithWriteLock(func() {
		if _, z, ok := g.ResolutionCastCandidateLocked(chooser, card); ok {
			zone = z
		}
	})
	return zone
}

// keywordCastCard is the card a cascade or discover offer is about.
func KeywordCastCard(c *game.PendingChoice) uuid.UUID {
	if c.Kind == game.PendingChoiceResolutionCast {
		if len(c.DiscardOptions) == 0 {
			return uuid.Nil
		}
		return c.DiscardOptions[0]
	}
	return c.MayCastCard
}

// answerKeywordCast answers one such offer. Taking it casts the card
// during the resolution (a resolution_cast) or opens the free cast (a
// may_cast); declining is the keyword's own decline.
func AnswerKeywordCast(t *testing.T, g *game.Game, c *game.PendingChoice, apply bool) {
	t.Helper()
	if c.Kind == game.PendingChoiceResolutionCast {
		if !apply {
			if err := g.ResolvePendingChoice(c.ID, c.Chooser, nil); err != nil {
				t.Fatalf("declining the cast: %v", err)
			}
			return
		}
		card := KeywordCastCard(c)
		if err := g.CastForResolutionChoice(c.Chooser, c.ID, card, game.CastSpellParams{FromZone: string(OfferedZone(g, c.Chooser, card))}); err != nil {
			t.Fatalf("casting the offered card: %v", err)
		}
		return
	}
	if err := g.ResolveMayCast(c.ID, c.Chooser, apply); err != nil {
		t.Fatalf("ResolveMayCast: %v", err)
	}
}

// countMayCastPrompts reports how many cascade offers are waiting.
func CountMayCastPrompts(g *game.Game) int {
	n := 0
	for _, c := range g.PendingChoices {
		if IsKeywordCastPrompt(c) {
			n++
		}
	}
	return n
}

// answerAllMayCast answers every outstanding offer the same way,
// re-reading the queue each time because resolving one can queue the
// next (Maelstrom Wanderer's second cascade).
func AnswerAllMayCast(t *testing.T, g *game.Game, chooser uuid.UUID, apply bool) int {
	t.Helper()
	answered := 0
	for i := 0; i < 8; i++ {
		var found *game.PendingChoice
		for _, c := range g.PendingChoices {
			if IsKeywordCastPrompt(c) && c.Chooser == chooser {
				found = c
				break
			}
		}
		if found == nil {
			return answered
		}
		AnswerKeywordCast(t, g, found, apply)
		answered++
	}
	return answered
}
