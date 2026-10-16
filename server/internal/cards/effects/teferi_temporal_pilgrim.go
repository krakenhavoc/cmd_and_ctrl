package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Teferi, Temporal Pilgrim — Legendary Planeswalker — Teferi for
// {3}{U}{U}, starting loyalty 4 (EDHREC rank 3342):
//
//	"Whenever you draw a card, put a loyalty counter on Teferi.
//	 0: Draw a card.
//	 −2: Create a 2/2 blue Spirit creature token with vigilance and
//	     'Whenever you draw a card, put a +1/+1 counter on this
//	     token.'
//	 −12: Target opponent chooses a permanent they control and
//	      returns it to its owner's hand. Then they shuffle each
//	      nonland permanent they control into its owner's library."
//
// The static-looking first line is a triggered ability and it is
// complete: every draw its controller makes — their draw step, their
// Rhystic Study, or Teferi's own [0] — puts a loyalty counter on.
// The counter is placed by an EFFECT rather than paid as a cost, so
// Doubling Season and Vorinclex do apply to it, which is the
// opposite of what happens to the [0] itself (ADR 0032 §8).
//
// The [0] is the reason AbilityCost.Loyalty is a POINTER: zero is a
// real printed cost that spends the turn's once-per-turn window, and
// an int could not tell it from "not a loyalty ability at all". This
// is the first card in the catalog to print one, so it is the first
// to exercise that distinction end to end.
//
// The −2's Spirit is a catalog token template (ADR 0083, #3861):
// `token:spirit-temporal-pilgrim`, whose draw trigger is found through
// game.CatalogKey's token-key fallback exactly as a printed card's is,
// so the token grows on each of its controller's draws.
//
// The −12 is complete (#3861). The target opponent picks one permanent
// they control (a choose_cards prompt over their own board, the
// opponent as chooser), it returns to its owner's hand, and only when
// that move has settled — a commander's owner may be asked about the
// command zone (CR 903.9) — is every nonland permanent the opponent
// still controls put into its owner's library as one simultaneous
// move, after which each of those libraries is shuffled. The shuffle
// follows the instruction rather than what landed: a commander that
// went to the command zone instead still leaves its owner shuffling,
// as Chaos Warp's does. A land the opponent picked is bounced; lands
// they did not pick stay.
func init() {
	Register(Spec{
		OracleID:     "5f0fa785-37ab-46a8-9ec0-6f8b29576d93",
		Name:         "Teferi, Temporal Pilgrim",
		Completeness: CompletenessFull,
		// Printed loyalty reaches the card through deck import
		// (ADR 0032 §1); this is the fallback for tokens, fixtures
		// and the dev spawner.
		StartingLoyalty: 4,
		Triggered: []game.TriggeredAbility{
			WheneverYouDraw("Teferi, Temporal Pilgrim — put a loyalty counter on Teferi", teferiPilgrimLoyaltyCounter),
		},
		Activated: []ActivatedAbility{
			{
				Label: "[0]: Draw a card.",
				Cost:  LoyaltyCost(0),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
				},
			},
			{
				Label: "−2: Create a 2/2 blue Spirit creature token with vigilance and \"Whenever you draw a card, put a +1/+1 counter on this token.\"",
				Cost:  LoyaltyCost(-2),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateToken{
						Controller: item.Controller,
						Template:   teferiPilgrimSpiritToken(),
						N:          1,
					}.Apply(NewContext(g, item))
				},
			},
			{
				Label:   "−12: Target opponent chooses a permanent they control and returns it to its owner's hand. Then they shuffle each nonland permanent they control into its owner's library.",
				Cost:    LoyaltyCost(-12),
				Targets: TargetPlayer("target opponent", Opponent()),
				Effect:  teferiPilgrimMinusTwelve,
			},
		},
	})
}

// teferiPilgrimLoyaltyCounter is the body of the draw trigger.
// Package-level so the triggered item captures nothing.
//
// The source may have left the battlefield between the trigger going
// on the stack and resolving — a draw in response to the −2, and a
// Teferi at 2 loyalty who paid it — and AddCounter does not gate on
// zone, so the guard is this file's job.
func teferiPilgrimLoyaltyCounter(g *game.Game, item *game.StackItem) error {
	if !b09SourceStillOnBattlefield(g, item) {
		return nil
	}
	return AddCounter{
		Target: item.SourceCardID,
		Kind:   game.CounterLoyalty,
		N:      1,
	}.Apply(NewContext(g, item))
}

// teferiPilgrimSpiritToken is the −2's 2/2 blue Spirit with vigilance
// and "Whenever you draw a card, put a +1/+1 counter on this token."
func teferiPilgrimSpiritToken() game.Card { return tokenFromCatalog(printedTeferiPilgrimSpiritToken) }

// printedTeferiPilgrimSpiritToken is that Spirit as PRINTED, ability
// included. The slug takes the card's name as its qualifier, because
// "spirit" alone names many different tokens.
func printedTeferiPilgrimSpiritToken() tokenTemplate {
	return tokenTemplate{
		Slug: "spirit-temporal-pilgrim",
		Card: game.Card{
			Name:      "Spirit",
			TypeLine:  "Token Creature — Spirit",
			Power:     2,
			Toughness: 2,
			Colors:    []string{"U"},
			Keywords:  []string{"vigilance"},
		},
		Triggered: []game.TriggeredAbility{
			WheneverYouDraw("Spirit — put a +1/+1 counter on it", putCounterOnSelf),
		},
		Text: "Whenever you draw a card, put a +1/+1 counter on this token.",
	}
}

// teferiPilgrimMinusTwelve asks the target opponent for the permanent
// they return. An opponent who controls no permanent is not asked, and
// has nothing nonland to shuffle away either.
func teferiPilgrimMinusTwelve(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	opp := uuid.Nil
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			opp = t.ID
			break
		}
	}
	if opp == uuid.Nil {
		return nil
	}
	perms := permanentsControlledByMatching(g, opp, nil)
	if len(perms) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    opp,
		FromPlayer: opp,
		Source:     item.SourceCardID,
		Question:   "Teferi, Temporal Pilgrim — choose a permanent you control to return to its owner's hand",
		Cards:      perms,
		Min:        1,
		Max:        1,
		Zone:       game.ZoneBattlefield,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			return teferiPilgrimReturnThenShuffle(g, opp, picked)
		},
	})
	return nil
}

// teferiPilgrimReturnThenShuffle returns the pick to its owner's hand
// and, once that has settled, shuffles the rest away. A pick the
// opponent no longer controls is not returned — "a permanent they
// control" — but the shuffle still happens.
func teferiPilgrimReturnThenShuffle(g *game.Game, opp uuid.UUID, picked []uuid.UUID) error {
	if len(picked) == 0 {
		return teferiPilgrimShuffleNonland(g, opp)
	}
	if c, ok := g.LookupCardForEffect(picked[0]); !ok || c.Controller != opp {
		return teferiPilgrimShuffleNonland(g, opp)
	}
	return g.BounceToHandThenForEffect(picked[0], func(g *game.Game, _ bool) error {
		return teferiPilgrimShuffleNonland(g, opp)
	})
}

// teferiPilgrimShuffleNonland puts every nonland permanent `opp`
// controls into its owner's library as one move, then shuffles each
// of those owners' libraries, in seat order.
func teferiPilgrimShuffleNonland(g *game.Game, opp uuid.UUID) error {
	ids := permanentsControlledByMatching(g, opp, Not(Land()))
	if len(ids) == 0 {
		return nil
	}
	owners := map[uuid.UUID]bool{}
	for _, id := range ids {
		if c, ok := g.LookupCardForEffect(id); ok {
			owners[c.Owner] = true
		}
	}
	return g.TuckCardsToLibraryThenForEffect(ids, game.TuckOptions{}, func(g *game.Game, _ []uuid.UUID) error {
		for _, p := range g.Seats {
			if p == nil || !owners[p.ID] {
				continue
			}
			if err := g.ShuffleLibraryForEffect(p.ID); err != nil {
				return err
			}
		}
		return nil
	})
}
