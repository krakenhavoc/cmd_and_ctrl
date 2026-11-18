package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chaotic Transformation — Sorcery {5}{R} (Dominaria United):
//
//	"Exile up to one target artifact, up to one target creature, up to one
//	 target enchantment, up to one target planeswalker, and/or up to one
//	 target land. For each permanent exiled this way, its controller reveals
//	 cards from the top of their library until they reveal a card that shares
//	 a card type with it, puts that card onto the battlefield, then shuffles."
//
// Five "up to one" clauses, each its own target (CR 115.3: one permanent can
// fill two clauses, as an artifact creature does for "artifact" and
// "creature", and is then exiled once). The legal picks are exiled together as
// one event (ExileCardsKeyedForEffect), and a keyed tail is told the ones that
// really reached exile ("exiled this way", CR 400.7).
//
// Then, for each of those permanents in turn, its controller as it last
// existed (CR 608.2h) reveals from the top of their library until a card that
// shares a card type with it as it last existed (an animated land is a
// creature too), puts that card onto the battlefield (CR 110.2a: they put it),
// then shuffles. Everything revealed but that card goes back by the shuffle,
// not to the bottom. A library with no such card is revealed in full and
// shuffled. Each permanent's put runs as a keyed tail that goes on to the next
// permanent, with the ones still to do carried as data, so an entry that
// pauses for a question is a restore point and the next one waits.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a754458d-55b5-44c2-98de-8bf26b82c1b1",
		Name:         "Chaotic Transformation",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetPermanent("up to one target artifact", Artifact()).WithCount(0, 1),
			TargetPermanent("up to one target creature", Creature()).WithCount(0, 1),
			TargetPermanent("up to one target enchantment", Enchantment()).WithCount(0, 1),
			TargetPermanent("up to one target planeswalker", Planeswalker()).WithCount(0, 1),
			TargetPermanent("up to one target land", Land()).WithCount(0, 1),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			var ids []uuid.UUID
			seen := map[uuid.UUID]bool{}
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard && !seen[t.ID] {
					seen[t.ID] = true
					ids = append(ids, t.ID)
				}
			}
			if len(ids) == 0 {
				return nil
			}
			return ctx.Game.ExileCardsKeyedForEffect(ids, ctx.Tail(chaoticTransformationExiled))
		},
	})
}

// chaoticTransformationExiled is the exile's tail: the permanents that
// reached exile, then each one's reveal. An on-disk key.
var chaoticTransformationExiled = CardTail("chaotic-transformation/exiled", func(ctx *Context, o game.TailOutcome) error {
	return chaoticTransformationNext(ctx, o.IDs)
})

// chaoticTransformationPut runs after one permanent's put: its controller
// shuffles, then the next permanent is taken. Carry is the permanents still to
// do, padded with uuid.Nil to chaoticTransformationMax. An on-disk key. It and
// chaoticTransformationNext refer to each other, so it is registered in init
// rather than by a package-level initializer (an initialization cycle).
var chaoticTransformationPut game.TailRef

func init() {
	chaoticTransformationPut = PutFromLibraryTailCarrying("chaotic-transformation/put-then-shuffle", chaoticTransformationMax,
		func(ctx *Context, res PutFromLibraryResult) error {
			if err := ctx.Game.ShuffleLibraryForEffect(res.Player); err != nil {
				return err
			}
			var rest []uuid.UUID
			for _, id := range res.Carry {
				if id != uuid.Nil {
					rest = append(rest, id)
				}
			}
			return chaoticTransformationNext(ctx, rest)
		})
}

// chaoticTransformationMax is the most permanents one cast can exile: one for
// each of the five clauses.
const chaoticTransformationMax = 5

// chaoticTransformationNext takes the first of `pending`: its controller
// reveals until a card shares a card type with it and puts that card onto the
// battlefield. The tail shuffles and goes on with the rest.
func chaoticTransformationNext(ctx *Context, pending []uuid.UUID) error {
	g := ctx.Game
	for len(pending) > 0 {
		id := pending[0]
		pending = pending[1:]
		info, ok := g.LastKnownPermanentForEffect(id)
		if !ok {
			continue
		}
		player := info.Controller
		types := append([]string(nil), info.Characteristic.Types...)
		run, hit := revealUntil(ctx, player, func(c game.Card) bool {
			for _, t := range types {
				if c.HasCardType(strings.ToLower(t)) {
					return true
				}
			}
			return false
		}, "Chaotic Transformation — reveal until a card that shares a card type")
		carry := make([]uuid.UUID, chaoticTransformationMax)
		copy(carry, pending)
		return g.PutByEnteringPlayerForEffect(func() error {
			return PutFromLibraryOntoBattlefield{
				Player:    player,
				Cards:     run,
				Match:     func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return hit != uuid.Nil && c.InstanceID == hit },
				All:       true,
				ThenTail:  chaoticTransformationPut,
				ThenCarry: carry,
			}.Apply(ctx)
		})
	}
	return nil
}
