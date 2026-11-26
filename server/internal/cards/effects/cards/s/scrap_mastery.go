package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Scrap Mastery — Sorcery {3}{R}{R} (Commander 2014):
//
//	"Each player exiles all artifact cards from their graveyard, then
//	 sacrifices all artifacts they control, then puts all cards they
//	 exiled this way onto the battlefield."
//
// Three simultaneous steps, each over every player at once: every artifact
// card in every graveyard is exiled together; then every artifact on the
// battlefield is sacrificed together (the artifacts just sacrificed went to
// graveyards and are not among the exiled cards); then the cards exiled in
// step one enter together, each under its owner's control (CR 400.3: a
// card goes to its owner's graveyard, so each player puts back their own).
// The exiled set is carried as plain data so every pause is a restore point.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "62dac700-49f7-4745-af67-37d2a81dcebc",
		Name:         "Scrap Mastery",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepArtifacts, How: game.SweepSacrifice, Partial: true}},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			var ids []uuid.UUID
			for _, p := range ctx.Game.Seats {
				if p == nil || p.Graveyard == nil {
					continue
				}
				for _, c := range p.Graveyard.Cards {
					if c.IsArtifact() {
						ids = append(ids, c.InstanceID)
					}
				}
			}
			if len(ids) == 0 {
				return scrapMasterySacrifice(ctx, nil)
			}
			return ctx.Game.ExileCardsKeyedForEffect(ids, ctx.ItemTail(scrapMasteryExiled))
		},
	})
}

var scrapMasteryExiled = effects.CardTail("then/scrap-mastery-exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	return scrapMasterySacrifice(ctx, o.IDs)
})

var scrapMasteryPut = effects.CardTail("then/scrap-mastery-put", func(ctx *effects.Context, o game.TailOutcome) error {
	return scrapMasteryReturn(ctx, o.Tail.Carry)
})

// scrapMasterySacrifice is step two: every artifact on the battlefield is
// sacrificed at once, then step three runs with the cards exiled in step one.
func scrapMasterySacrifice(ctx *effects.Context, exiled []uuid.UUID) error {
	var artifacts []uuid.UUID
	for _, c := range effects.MatchingBattlefield(ctx, effects.Artifact()) {
		artifacts = append(artifacts, c.InstanceID)
	}
	if len(artifacts) == 0 {
		return scrapMasteryReturn(ctx, exiled)
	}
	then := ctx.ItemTail(scrapMasteryPut)
	then.Carry = append([]uuid.UUID(nil), exiled...)
	return ctx.Game.SacrificeAllKeyedForEffect(ctx.Source(), artifacts, then)
}

// scrapMasteryReturn is step three: the exiled cards enter together under
// their owners' control (a card that has left exile since, or a token, is
// skipped).
func scrapMasteryReturn(ctx *effects.Context, exiled []uuid.UUID) error {
	if len(exiled) == 0 {
		return nil
	}
	return effects.ReturnFromExileTogether{Targets: exiled}.Apply(ctx)
}
