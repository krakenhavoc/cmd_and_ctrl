package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Braids, Arisen Nightmare — Legendary Creature — Nightmare {1}{B}{B}, 3/3
// (EDHREC rank ~360):
//
//	"At the beginning of your end step, you may sacrifice an artifact,
//	 creature, enchantment, land, or planeswalker. If you do, each
//	 opponent may sacrifice a permanent of their choice that shares a
//	 card type with it. For each opponent who doesn't, that player
//	 loses 2 life and you draw a card."
//
// # The card #568's option_pick seam was waiting on
//
// Roadmap batch 02 (#295) filed Braids under "an opponent-choice
// sacrifice with a continuation" before #568 built
// game.PendingChoiceOptionPick, and the option_pick file comment names
// this card by name as one of the three it unblocks. This is that
// card, written now that the seam is closed.
//
// Three prompts chained in sequence, none of them new:
//
//  1. MayChoice — "you may sacrifice a permanent?" to the controller.
//     A decline ends the ability here, matching "you may ... If you
//     do, ..." (CR 603.3d has already confirmed the trigger itself;
//     this is the ability's own internal "if you do").
//  2. A plain battlefield ChooseCardsPrompt picks WHICH permanent, so
//     the rest of the ability knows what type to match against. Its
//     candidate set is filtered to the five printed types — nothing
//     lets Braids sacrifice a battle, the one permanent type the card
//     predates and does not name.
//  3. Once the sacrifice has actually happened (SacrificeAllThenForEffect's
//     `sacrificed` list, not the pick itself — CR 701.17a's own
//     "sacrifice is not replaceable" still opens a CR 903.9 window on
//     the MOVE for a sacrificed commander), each opponent in turn
//     order gets a PickOption: sacrifice a permanent that shares a
//     card type with what left, or take the printed consequence. The
//     consequence option is first and always legal, exactly as
//     Torment of Hailfire's "lose life" branch is, and an opponent
//     with no matching permanent is never asked at all — CR 608.2's
//     "as much as possible" enforced at build time, not at answer
//     time.
//
// # Card-type sharing
//
// The chosen permanent's types are read ONCE, via LookupCardForEffect
// before the sacrifice, and frozen into a small type-flag set —
// exactly the LKI discipline The Ozolith documents, needed here for
// the same reason: by the time an opponent answers, the permanent
// that left is gone and its type line cannot be re-read off the
// battlefield. An opponent's own answering permanent is checked with
// no restriction to the five printed types, because CR 608.2 asks
// only that it "share a card type with it" — which in practice is
// always one of those five, since that's the only kind Braids' own
// half can ever have been.
//
// # Sequencing
//
// Opponents are asked one at a time, in Context.Opponents() order,
// each queued by the previous one's answer — the same reason Torment
// of Hailfire's repetitions are a chain rather than one prompt fired
// per seat: a player who sacrifices their last matching permanent to
// an earlier question has a different (empty) option list by the time
// their own turn in the chain arrives, and only a strictly sequential
// ask observes that correctly.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e0445c80-fa53-4c3e-881e-940e9fce7f57",
		Name:         "Braids, Arisen Nightmare",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Braids, Arisen Nightmare — you may sacrifice a permanent", braidsArisenNightmareResolve),
		},
	})
}

// braidsCardTypeFlags is a snapshot of which of the six card types a
// permanent had, frozen at the moment it was chosen — before it
// leaves the battlefield and its own type line stops being
// answerable. Battle is tracked even though Braids can never offer it
// as ITS OWN half, because CR 608.2's "shares a card type" is
// evaluated on the departed permanent's actual types, not on the
// printed five-type list that only restricts the initial pick.
type braidsCardTypeFlags struct {
	artifact, creature, enchantment, land, planeswalker, battle bool
}

func braidsTypesOf(c game.Card) braidsCardTypeFlags {
	return braidsCardTypeFlags{
		artifact:     c.IsArtifact(),
		creature:     c.IsCreature(),
		enchantment:  c.IsEnchantment(),
		land:         c.IsLand(),
		planeswalker: c.IsPlaneswalker(),
		battle:       c.IsBattle(),
	}
}

// shares reports whether c has at least one card type in common with
// the frozen set.
func (t braidsCardTypeFlags) shares(c game.Card) bool {
	return (t.artifact && c.IsArtifact()) ||
		(t.creature && c.IsCreature()) ||
		(t.enchantment && c.IsEnchantment()) ||
		(t.land && c.IsLand()) ||
		(t.planeswalker && c.IsPlaneswalker()) ||
		(t.battle && c.IsBattle())
}

// braidsSacrificeCandidates lists the permanents `controller` may
// offer up: an artifact, creature, enchantment, land, or planeswalker
// — every permanent type Braids predates and does not name (battle)
// is excluded, which PermanentsControlledBy's "any permanent" would
// not do.
//
// Caller must hold g.mu.
func braidsSacrificeCandidates(g *game.Game, controller uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller != controller {
			continue
		}
		if c.IsArtifact() || c.IsCreature() || c.IsEnchantment() || c.IsLand() || c.IsPlaneswalker() {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// braidsMatchingPermanents lists the permanents `opponent` controls
// that share a card type with the frozen flag set.
//
// Caller must hold g.mu.
func braidsMatchingPermanents(g *game.Game, opponent uuid.UUID, types braidsCardTypeFlags) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller != opponent {
			continue
		}
		if types.shares(c) {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// braidsArisenNightmareResolve is the whole ability's Effect: the
// "you may sacrifice" offer, gated on there being anything legal to
// offer at all.
func braidsArisenNightmareResolve(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	controller := ctx.Controller()
	candidates := braidsSacrificeCandidates(g, controller)
	if len(candidates) == 0 {
		return nil
	}
	return MayChoice{
		Question: "Braids, Arisen Nightmare — sacrifice an artifact, creature, enchantment, land, or planeswalker?",
		OnYes: func(ctx *Context) error {
			return braidsChoosePermanentToSacrifice(ctx, controller, candidates)
		},
	}.Apply(ctx)
}

// braidsChoosePermanentToSacrifice is the second prompt: WHICH
// permanent, from the candidate set built at trigger time (safe to
// reuse — nothing else can act on the battlefield while this
// resolution's own prompts are open).
func braidsChoosePermanentToSacrifice(ctx *Context, controller uuid.UUID, candidates []uuid.UUID) error {
	item := ctx.Item
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    controller,
		FromPlayer: controller,
		Source:     ctx.Source(),
		Question:   "Braids, Arisen Nightmare — sacrifice which permanent?",
		Cards:      candidates,
		Min:        1,
		Max:        1,
		Zone:       game.ZoneBattlefield,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			if len(picked) == 0 {
				// The chooser left the game between the offer and the
				// pick (CR 800.4a) — nothing to sacrifice, so the "if
				// you do" clause never fires.
				return nil
			}
			var types braidsCardTypeFlags
			if c, ok := g.LookupCardForEffect(picked[0]); ok {
				types = braidsTypesOf(c)
			}
			return g.SacrificeAllThenForEffect(uuid.Nil, picked, func(g *game.Game, sacrificed []uuid.UUID) error {
				if len(sacrificed) == 0 {
					// The sacrifice did not land (a departed chooser
					// mid-CR-903.9-window would be the only real way);
					// "if you do" was not satisfied.
					return nil
				}
				fresh := NewContext(g, item)
				return braidsOpponentsStep(fresh, controller, types, fresh.Opponents())
			})
		},
	})
	return nil
}

// braidsOpponentsStep asks the next opponent still owed a question,
// one at a time, in the order Context.Opponents() returns — the same
// strictly-sequential chain Torment of Hailfire's repetitions use and
// for the same reason (#552): each opponent gets ONE prompt built out
// of an option_pick, and a card written as a chain of single-seat
// questions has no other shape to be than a sequence of them, asked
// and answered one at a time rather than batched.
//
// An opponent with no permanent sharing a card type is never asked —
// CR 608.2's "as much as possible" applied at build time — and simply
// takes the consequence.
func braidsOpponentsStep(ctx *Context, controller uuid.UUID, types braidsCardTypeFlags, remaining []uuid.UUID) error {
	for len(remaining) > 0 {
		victim := remaining[0]
		rest := remaining[1:]
		p := ctx.Game.PlayerByIDForEffect(victim)
		if p == nil || p.Eliminated {
			remaining = rest
			continue
		}
		matching := braidsMatchingPermanents(ctx.Game, victim, types)
		if len(matching) == 0 {
			if err := braidsApplyConsequence(ctx, controller, victim); err != nil {
				return err
			}
			remaining = rest
			continue
		}
		return PickOption{
			Player:   victim,
			Question: "Braids, Arisen Nightmare — sacrifice a permanent that shares a card type with it, or lose 2 life and let its controller draw a card?",
			Options: []game.ChoiceOption{
				{Label: "Don't sacrifice — lose 2 life; Braids's controller draws a card", LifeCost: 2},
				{Label: "Sacrifice a permanent that shares a card type with it"},
			},
			Then: func(ctx *Context, index int) error {
				if index == 1 {
					return SacrificeChoice{
						Player:     victim,
						Candidates: matching,
						Question:   "Braids, Arisen Nightmare — sacrifice a permanent that shares a card type with it",
						Then: func(ctx *Context) error {
							return braidsOpponentsStep(ctx, controller, types, rest)
						},
					}.Apply(ctx)
				}
				if err := braidsApplyConsequence(ctx, controller, victim); err != nil {
					return err
				}
				return braidsOpponentsStep(ctx, controller, types, rest)
			},
		}.Apply(ctx)
	}
	return nil
}

// braidsApplyConsequence is "that player loses 2 life and you draw a
// card" for one opponent who didn't sacrifice.
func braidsApplyConsequence(ctx *Context, controller, victim uuid.UUID) error {
	if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), victim, -2); err != nil {
		return err
	}
	return DrawCards{Player: controller, N: 1}.Apply(ctx)
}
