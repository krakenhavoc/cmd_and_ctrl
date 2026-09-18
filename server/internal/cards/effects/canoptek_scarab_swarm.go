package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Canoptek Scarab Swarm — 1/1 Artifact Creature — Insect for {4}
// (EDHREC rank 4004):
//
//	"Flying
//	 Feeder Mandibles — When this creature enters, exile target
//	 player's graveyard. For each artifact or land card exiled this
//	 way, create a 1/1 colorless Insect artifact creature token with
//	 flying."
//
// Graveyard hate that leaves a board behind. Against a reanimator or
// an artifact deck it is a Bojuka Bog that pays two or three bodies
// for the privilege; against a deck with an empty yard it is a 1/1
// flier for four, which is the risk.
//
// It is in the batch because the token count is measured over what
// the effect ACTUALLY EXILED, not over what was in the graveyard when
// the trigger was announced. The two come apart whenever something
// leaves in between — a card returned in response, a card an
// opponent's own effect moved out, a card a replacement kept out of
// exile. Reading the count off a pre-move snapshot would over-pay;
// reading it off the graveyard afterwards would find it empty. So the
// clause counts the cards that reached exile.
//
// # "Artifact or land card"
//
// Read off the card as it sat in the graveyard — its printed types,
// which is all a card in a graveyard has. An artifact LAND counts
// once, not twice: it is one card, and the clause is a disjunction
// over cards rather than a tally of type matches. A creature card, an
// instant, a planeswalker: exiled, but paying nothing.
//
// # "Target player's" includes yourself
//
// Nothing restricts it to an opponent, and a self-targeting Swarm is
// a real line in a deck that has filled its own yard with artifacts
// and wants the bodies more than the recursion.
//
// The tokens are artifacts themselves, so they turn on the affinity
// and Metalcraft payoffs, and they fly.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5b7f6cc3-8d4e-42e5-a908-7ebe3bccaf1e",
		Name:            "Canoptek Scarab Swarm",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Canoptek Scarab Swarm — Feeder Mandibles: exile a graveyard, make an Insect per artifact or land",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetPlayer {
							continue
						}
						n, err := b38ExileGraveyardCountingArtifactsAndLands(ctx, t.ID)
						if err != nil {
							return err
						}
						if n <= 0 {
							return nil
						}
						return CreateToken{
							Controller: item.Controller,
							Template:   TokenCard("1/1 colorless Insect artifact with flying"),
							N:          n,
						}.Apply(ctx)
					}
					return nil
				}), TargetPlayer("target player")),
		},
	})
}

// b38ExileGraveyardCountingArtifactsAndLands exiles every card in
// `player`'s graveyard and returns how many of the cards that
// actually REACHED exile were artifact or land cards.
//
// Counting after the move is the point (see the file comment): a card
// that left in response, or that a replacement kept out of exile, must
// not pay. The type read is of the pre-move copy, because a card in
// exile is still findable but the pre-move copy is what "exiled this
// way" names, and a card that is both an artifact and a land counts
// once.
//
// Declared here rather than in batch38_helpers.go because it is one
// card's clause.
func b38ExileGraveyardCountingArtifactsAndLands(ctx *Context, player uuid.UUID) (int, error) {
	p := ctx.Game.PlayerByIDForEffect(player)
	if p == nil || p.Graveyard == nil {
		return 0, nil
	}
	type entry struct {
		id    uuid.UUID
		pays  bool
		exile bool
	}
	pile := make([]entry, 0, len(p.Graveyard.Cards))
	for _, c := range p.Graveyard.Cards {
		pile = append(pile, entry{id: c.InstanceID, pays: c.IsArtifact() || c.IsLand()})
	}
	for i := range pile {
		if err := (ExileTarget{Target: pile[i].id}).Apply(ctx); err != nil {
			return 0, err
		}
		if z := ctx.Game.FindCardZoneForEffect(pile[i].id); z != nil && z.Kind == game.ZoneExile {
			pile[i].exile = true
		}
	}
	n := 0
	for _, e := range pile {
		if e.exile && e.pays {
			n++
		}
	}
	return n, nil
}
