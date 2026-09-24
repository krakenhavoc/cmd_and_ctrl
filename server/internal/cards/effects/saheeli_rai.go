package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Saheeli Rai — Legendary Planeswalker — Saheeli for {1}{U}{R},
// starting loyalty 3 (EDHREC rank 7871):
//
//	"+1: Scry 1. Saheeli Rai deals 1 damage to each opponent.
//	 −2: Create a token that's a copy of target artifact or creature
//	     you control, except it's an artifact in addition to its other
//	     types. That token gains haste. Exile it at the beginning of
//	     the next end step.
//	 −7: Search your library for up to three artifact cards with
//	     different names, put them onto the battlefield, then
//	     shuffle."
//
// All three abilities are wired in full, including the ultimate —
// which is the point worth recording, because it is the first
// ultimate in the catalog that is not an emblem. Saheeli's −7 is an
// ordinary tutor with a constraint, and the search primitive already
// carries both halves it needs: a limit that means "up to", so
// finding fewer (or none) is a legal answer, and a set-level
// Validate for "with different names" that no per-card predicate
// could express.
//
// The +1 puts the damage in the scry's continuation rather than
// after it. Scry only QUEUES a prompt — nothing moves until the
// player answers — so a second statement written after it would
// resolve first and print the wrong order on the log. The
// continuation runs even when the library was empty and there was
// nothing to look at, so an empty-library Saheeli still pings the
// table.
//
// The −2 is the interesting one. A token copy carries the original's
// oracle ID, so every ability the copied permanent has comes along
// for free; the "except it's an artifact in addition to its other
// types" clause is applied to the TEMPLATE before the token exists,
// which makes it a printed characteristic of a new object rather
// than a continuous effect on an existing one — and that is why it
// works here when Phyrexian Scriptures' version of the same sentence
// does not. Haste rides the entry options and the exile is a delayed
// trigger on the token's own instance, so a token that has already
// left by the end step is skipped rather than chased.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5f3fe679-aff1-41b3-8d75-c78c2c0636f0",
		Name:         "Saheeli Rai",
		Completeness: CompletenessFull,
		// Printed loyalty reaches the card through deck import
		// (ADR 0032 §1); this is the fallback for tokens, fixtures
		// and the dev spawner.
		StartingLoyalty: 3,
		Activated: []ActivatedAbility{
			{
				Label: "+1: Scry 1. Saheeli Rai deals 1 damage to each opponent.",
				Cost:  LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					// Plain values, not the item: the scry's
					// continuation runs after a player prompt, and
					// the stack item it was scheduled from is gone
					// by then.
					controller, source := item.Controller, item.SourceCardID
					return Scry{
						Player: controller,
						N:      1,
						Then: func(g *game.Game) error {
							return saheeliPingEachOpponent(g, controller, source)
						},
					}.Apply(NewContext(g, item))
				},
			},
			{
				Label:   "−2: Create a token that's a copy of target artifact or creature you control, except it's an artifact in addition to its other types. That token gains haste. Exile it at the beginning of the next end step.",
				Cost:    LoyaltyCost(-2),
				Targets: TargetPermanent("target artifact or creature you control", Or(Artifact(), Creature()), YouControl()),
				Effect:  saheeliHastyArtifactCopy,
			},
			{
				Label: "−7: Search your library for up to three artifact cards with different names, put them onto the battlefield, then shuffle.",
				Cost:  LoyaltyCost(-7),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return SearchLibrary{
						Player:    item.Controller,
						Predicate: saheeliIsArtifactCard,
						Dest:      game.ZoneBattlefield,
						Limit:     3,
						Shuffle:   true,
						Validate:  saheeliDifferentNames,
						Reason:    "Saheeli Rai — up to three artifact cards with different names, onto the battlefield",
					}.Apply(NewContext(g, item))
				},
			},
		},
	})
}

// saheeliPingEachOpponent is the second half of the +1, run from the
// scry's continuation. It takes plain IDs rather than a stack item
// because by the time a player has answered the scry prompt the
// ability that queued it has left the stack.
func saheeliPingEachOpponent(g *game.Game, controller, source uuid.UUID) error {
	for _, p := range g.Seats {
		if p == nil || p.Eliminated || p.ID == controller {
			continue
		}
		if err := g.DealDamageToPlayerForEffect(source, p.ID, 1); err != nil {
			return err
		}
	}
	return nil
}

// saheeliHastyArtifactCopy is the −2. Package-level so the ability's
// closure captures nothing.
func saheeliHastyArtifactCopy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	source := uuid.Nil
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			source = t.ID
			break
		}
	}
	if source == uuid.Nil {
		return nil
	}
	tmpl, ok := TokenCopyTemplate(g, source)
	if !ok {
		return nil
	}
	tmpl.TypeLine = saheeliArtifactInAddition(tmpl.TypeLine)
	made, err := g.CreateTokensForEffect(item.Controller, tmpl, 1, game.TokenEntryOptions{
		Keywords: []string{"haste"},
	})
	if err != nil {
		return err
	}
	if len(made) == 0 {
		return nil
	}
	return ScheduleDelayedTrigger{
		At:         game.StepEnd,
		Controller: item.Controller,
		Label:      "Saheeli Rai — exile the copy",
		Cards:      made,
		Body:       exileListedCardsBody,
	}.Apply(ctx)
}

// saheeliArtifactInAddition is "except it's an artifact in addition
// to its other types" (CR 707.9a applied to the copiable values).
// Adding rather than replacing is the difference from
// retypedTypeLine: a copy of a Bear is an Artifact Creature — Bear,
// still a Bear and still a creature.
//
// A permanent that is already an artifact is returned untouched, so
// copying a Sol Ring does not produce "Artifact Artifact".
func saheeliArtifactInAddition(printed string) string {
	super, types, sub := game.ParseTypeLine(printed)
	for _, t := range types {
		if t == "Artifact" {
			return printed
		}
	}
	head := make([]string, 0, len(super)+len(types)+1)
	head = append(head, super...)
	head = append(head, "Artifact")
	head = append(head, types...)
	line := strings.Join(head, " ")
	if len(sub) == 0 {
		return line
	}
	return line + " — " + strings.Join(sub, " ")
}

// saheeliIsArtifactCard is the −7's search predicate. Artifact
// CARDS, so an artifact creature qualifies and a creature that is
// not an artifact does not.
func saheeliIsArtifactCard(c game.Card) bool {
	return c.IsArtifact()
}

// saheeliDifferentNames is the −7's "with different names" clause,
// checked against the whole picked set rather than per card — which
// is what the Validate hook is for, and why Myriad Landscape needed
// it first.
func saheeliDifferentNames(chosen []game.Card) bool {
	seen := make(map[string]bool, len(chosen))
	for _, c := range chosen {
		if seen[c.Name] {
			return false
		}
		seen[c.Name] = true
	}
	return true
}
