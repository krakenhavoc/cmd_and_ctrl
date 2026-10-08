package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hauntwoods Shrieker — Creature — Beast Mutant {1}{G}{G}, 3/3:
//
//	"Whenever this creature attacks, manifest dread. (Look at the top
//	 two cards of your library. Put one onto the battlefield face down
//	 as a 2/2 creature and the other into your graveyard. Turn it face
//	 up any time for its mana cost if it's a creature card.)
//	 {1}{G}: Reveal target face-down permanent. If it's a creature
//	 card, you may turn it face up."
//
// The attack trigger is manifest dread (ADR 0082's 2026-10-07
// amendment). The activated ability is the first card to turn a
// permanent face up as an EFFECT (its second amendment, #2590): it
// targets any face-down permanent, an opponent's included, and
// reveals it (every seat becomes a knower; it stays face down unless
// the controller turns it up). The "if it's a creature card" test is
// the CARD underneath (PrintedIsCreature), not the 2/2 body, so a
// face-down Island is revealed and stays face down. "You may" is a
// yes/no asked of the Shrieker's controller after the reveal, and the
// turn is free.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "09c275bd-8323-485e-81f6-aff1f99d0d03",
		Name:         "Hauntwoods Shrieker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, Self, "Hauntwoods Shrieker — manifest dread", Do(ManifestDread{})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{G}: Reveal target face-down permanent. If it's a creature card, you may turn it face up.",
			Cost:    ManaCost("{1}{G}"),
			Targets: TargetPermanent("target face-down permanent", FaceDown()),
			Effect:  hauntwoodsShriekerRevealAndMaybeTurnUp,
		}},
	})
}

func hauntwoodsShriekerRevealAndMaybeTurnUp(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		id := t.ID
		g.RevealForEffect(game.RevealSpec{
			Player: ctx.Controller(),
			Source: ctx.Source(),
			Reason: "Hauntwoods Shrieker — reveal target face-down permanent",
			Cards:  []uuid.UUID{id},
		})
		c, ok := g.LookupCardForEffect(id)
		if !ok || !c.PrintedIsCreature() || !game.CanTurnFaceUpForEffect(c) {
			return nil
		}
		return MayChoice{
			Question: "Hauntwoods Shrieker — turn the revealed creature card face up?",
			OnYes: func(ctx *Context) error {
				return TurnFaceUp{Targets: []uuid.UUID{id}}.Apply(ctx)
			},
		}.Apply(ctx)
	}
	return nil
}
