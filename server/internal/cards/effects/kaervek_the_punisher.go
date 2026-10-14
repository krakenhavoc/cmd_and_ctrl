package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kaervek, the Punisher — Legendary Creature — Human Warlock {1}{B}{B},
// 3/3:
//
//	"Whenever you commit a crime, exile up to one target black card from
//	 your graveyard and copy it. You may cast the copy. If you do, you
//	 lose 2 life. (Targeting opponents, anything they control, and/or
//	 cards in their graveyards is a crime. Copies of permanent spells
//	 become tokens.)"
//
// The crime is the once-per-announcement CrimeCommitted annotation
// (YouCommittedCrime, CR 700.13, #2944): copies and new targets commit
// no new crime. Targeting a card in your own graveyard is not a crime,
// so Kaervek's own trigger never feeds itself.
//
// "Up to one target black card from your graveyard" is declared on the
// ability: with no black card there the trigger is not put on the
// stack, since it would do nothing (CR 603.3d), and the target is
// re-checked as it resolves (CR 608.2b).
//
// The exile goes through the CR 614 window, and only a card that really
// reached exile is copied, which is CopyAndCast (#3258, ADR 0066's
// 2026-10-13 card-copy amendment). The copy is offered at its printed
// mana cost ("You may cast the copy", not "without paying"), at instant
// speed, until Kaervek's controller next passes priority. A permanent
// copy resolves as a token (CR 707.10b); a copy not cast ceases to exist
// (CR 704.5e), and the original card stays in exile.
//
// "If you do, you lose 2 life" is the offer's FollowUp, which runs once,
// only when the copy is cast through this offer.
//
// No simplification.
var kaervekLosesTwo = game.RegisterCastFollowUp("kaervek-the-punisher-lose-2",
	func(g *game.Game, f game.CastFollowUp) error {
		return g.ChangePlayerLifeForEffect(f.Source, f.Player, -2)
	})

func init() {
	Register(Spec{
		OracleID:     "fa9a6083-75ab-41ae-b607-39d21d9a3d3a",
		Name:         "Kaervek, the Punisher",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventBecomesTarget,
					func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
						return YouCommittedCrime(ev, source)
					},
					"Kaervek, the Punisher — exile up to one black card from your graveyard and copy it; you may cast the copy",
					kaervekExileAndCopy),
				TargetCardInGraveyard("up to one target black card from your graveyard", YouOwn(), OfColor("B")).WithCount(0, 1)),
		},
	})
}

// kaervekExileAndCopy exiles the still-legal target and offers a cast of
// a copy of whatever really reached exile.
func kaervekExileAndCopy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			ids = append(ids, t.ID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	return g.ExileCardsThenForEffect(ids, func(g *game.Game, exiled []uuid.UUID) error {
		return CopyAndCast{Cards: exiled, FollowUp: kaervekLosesTwo}.Apply(NewContext(g, item))
	})
}
