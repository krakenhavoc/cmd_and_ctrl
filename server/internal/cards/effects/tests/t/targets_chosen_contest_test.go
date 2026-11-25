package testst

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// targets_chosen_contest_test.go — #6228: Psychic Battle's trigger
// (effects.TopCardContestChangesTargets) over the engine's
// EventTargetsChosen and CR 115.7a's "change the target or targets".
// The card itself ships with its OLD batch; these tests install the
// helper under a test key.

const contestTestKey = "test-6228-psychic-battle"

func withContestTrigger(t *testing.T) {
	t.Helper()
	prev := game.CatalogTriggers
	row := []game.TriggeredAbility{effects.TopCardContestChangesTargets("Psychic Battle")}
	game.CatalogTriggers = func(k string) []game.TriggeredAbility {
		if k == contestTestKey {
			return row
		}
		return prev(k)
	}
	t.Cleanup(func() { game.CatalogTriggers = prev })
}

// contestItems is every Psychic Battle trigger on the stack.
func contestItems(g *game.Game) []*game.StackItem {
	var out []*game.StackItem
	g.ReadSnapshot(func() {
		for _, it := range g.StackMeta {
			if it != nil && it.Kind == game.StackItemTriggered && strings.HasPrefix(it.Label, "Psychic Battle") {
				out = append(out, it)
			}
		}
	})
	return out
}

// contestTop puts a card of mana cost `cost` on top of p's library.
func contestTop(g *game.Game, p *game.Player, cost string) {
	g.WithWriteLock(func() {
		c := game.NewCard("Top "+cost, p.ID)
		c.TypeLine = "Sorcery"
		c.ManaCost = cost
		p.Library.PushTop(c)
	})
}

// contestSettle passes priority until the Psychic Battle trigger has
// resolved or a prompt is open.
func contestSettle(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 24 && len(contestItems(g)) > 0; i++ {
		if err := g.PassPriority(); err != nil {
			if errors.Is(err, game.ErrChoicePending) {
				return
			}
			t.Fatalf("PassPriority: %v", err)
		}
	}
}

func contestPrompt(g *game.Game) *game.PendingChoice {
	var out *game.PendingChoice
	g.ReadSnapshot(func() {
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceRetarget {
				out = c
			}
		}
	})
	return out
}

func contestTargets(g *game.Game, id uuid.UUID) []game.TargetRef {
	var out []game.TargetRef
	g.ReadSnapshot(func() {
		if it := g.StackMeta[id]; it != nil {
			out = append([]game.TargetRef(nil), it.Targets...)
		}
	})
	return out
}

// The acceptance: a Doom Blade cast under Psychic Battle triggers it;
// the opponent reveals the greatest mana value and redirects the Blade
// onto the caster's own creature, and the Blade kills that one. The
// redirect is a change of targets made by Psychic Battle, so it does
// not trigger Psychic Battle again.
func TestPsychicBattleHighestRevealChangesTheTarget(t *testing.T) {
	withContestTrigger(t)
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Psychic Battle", "Enchantment", contestTestKey, false)
	mine := effectstest.PushTypedCard(g, me.ID, "My Bear", "Creature — Bear", "{1}{G}")
	theirs := effectstest.PushTypedCard(g, opp.ID, "Their Bear", "Creature — Bear", "{1}{G}")
	for _, p := range g.Seats {
		contestTop(g, p, "{1}")
	}
	contestTop(g, opp, "{5}{U}")

	spell := effectstest.CastCatalogSpell(t, g, "Doom Blade", "Instant", effectstest.DoomBladeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	if n := len(contestItems(g)); n != 1 {
		t.Fatalf("%d Psychic Battle triggers after the cast, want 1", n)
	}
	contestSettle(t, g)
	p := contestPrompt(g)
	if p == nil || p.Chooser != opp.ID {
		t.Fatalf("prompt = %+v, want one for the opponent, who revealed the greatest mana value", p)
	}
	if p.PickTargetMin != 0 {
		t.Errorf("PickTargetMin = %d: the winner MAY change the target", p.PickTargetMin)
	}
	if err := g.ResolveRetarget(p.ID, opp.ID, []game.TargetRef{{Kind: game.TargetCard, ID: mine}}); err != nil {
		t.Fatalf("ResolveRetarget: %v", err)
	}
	if got := contestTargets(g, spell); len(got) != 1 || got[0].ID != mine {
		t.Fatalf("the Blade targets %+v, want my bear", got)
	}
	if n := len(contestItems(g)); n != 0 {
		t.Fatalf("Psychic Battle's own change triggered it %d more times", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || !g.Battlefield.Contains(theirs) {
		t.Error("the Blade should have destroyed the creature it was redirected to, and only that one")
	}
}

// A tie for greatest changes nothing and asks nobody.
func TestPsychicBattleTieLeavesTheTarget(t *testing.T) {
	withContestTrigger(t)
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Psychic Battle", "Enchantment", contestTestKey, false)
	effectstest.PushTypedCard(g, me.ID, "My Bear", "Creature — Bear", "{1}{G}")
	theirs := effectstest.PushTypedCard(g, opp.ID, "Their Bear", "Creature — Bear", "{1}{G}")
	for _, p := range g.Seats {
		contestTop(g, p, "{1}")
	}
	contestTop(g, me, "{3}")
	contestTop(g, opp, "{2}{R}")

	spell := effectstest.CastCatalogSpell(t, g, "Doom Blade", "Instant", effectstest.DoomBladeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	contestSettle(t, g)
	if p := contestPrompt(g); p != nil {
		t.Fatalf("a tie for greatest asked %v to change the target", p.Chooser)
	}
	if got := contestTargets(g, spell); len(got) != 1 || got[0].ID != theirs {
		t.Fatalf("the Blade targets %+v, want the original", got)
	}
}

// A change of targets made by another card is a player choosing targets
// (CR 115.7), so Psychic Battle triggers on it.
func TestPsychicBattleTriggersOnAnotherCardsRetarget(t *testing.T) {
	withContestTrigger(t)
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Psychic Battle", "Enchantment", contestTestKey, false)
	mine := effectstest.PushTypedCard(g, me.ID, "My Bear", "Creature — Bear", "{1}{G}")
	theirs := effectstest.PushTypedCard(g, opp.ID, "Their Bear", "Creature — Bear", "{1}{G}")
	for _, p := range g.Seats {
		contestTop(g, p, "{1}")
	}
	spell := effectstest.CastCatalogSpell(t, g, "Doom Blade", "Instant", effectstest.DoomBladeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	contestSettle(t, g) // a tie: no prompt
	if n := len(contestItems(g)); n != 0 {
		t.Fatalf("%d triggers left", n)
	}
	deflector := effectstest.PushTypedCard(g, opp.ID, "Deflector", "Enchantment", "{2}")
	g.WithWriteLock(func() {
		if err := g.OfferRetargetForEffect(game.RetargetOffer{
			ItemID: spell, Chooser: opp.ID, Policy: game.RetargetChangeOne, Source: deflector,
		}); err != nil {
			t.Fatalf("OfferRetargetForEffect: %v", err)
		}
	})
	p := contestPrompt(g)
	if p == nil {
		t.Fatal("no retarget prompt")
	}
	if err := g.ResolveRetarget(p.ID, opp.ID, []game.TargetRef{{Kind: game.TargetCard, ID: mine}}); err != nil {
		t.Fatalf("ResolveRetarget: %v", err)
	}
	if n := len(contestItems(g)); n != 1 {
		t.Fatalf("%d Psychic Battle triggers after another card changed the target, want 1", n)
	}
}
