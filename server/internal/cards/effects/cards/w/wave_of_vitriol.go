package cardsw

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wave of Vitriol — Sorcery {5}{G}{G} (Commander 2014):
//
//	"Each player sacrifices all artifacts, enchantments, and nonbasic
//	 lands they control. For each land sacrificed this way, its controller
//	 may search their library for a basic land card and put it onto the
//	 battlefield tapped. Then each player who searched their library this
//	 way shuffles."
//
// Everything is sacrificed at once; the lands among the sacrificed set are
// counted as they last existed (CR 603.10a: a Song of the Dryads permanent
// was a land, a creature-land is one) and credited to the controller they
// had. Each player with N sacrificed lands may search for up to N basic
// lands, tapped, and shuffles only if they searched.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5b8b1f03-9e2e-40f2-97e0-c9fb8d780d56",
		Name:         "Wave of Vitriol",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepAllPermanents, How: game.SweepSacrifice, Partial: true}},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			var ids []uuid.UUID
			for _, c := range effects.MatchingBattlefield(ctx, effects.Or(effects.Artifact(), effects.Enchantment(), effects.NonbasicLand())) {
				ids = append(ids, c.InstanceID)
			}
			if len(ids) == 0 {
				return nil
			}
			return ctx.Game.SacrificeAllKeyedForEffect(ctx.Source(), ids, ctx.ItemTail(waveOfVitriolSacrificed))
		},
	})
}

var waveOfVitriolSacrificed = effects.CardTail("then/wave-of-vitriol-sacrificed", func(ctx *effects.Context, o game.TailOutcome) error {
	lands := map[uuid.UUID]int{}
	for _, id := range o.IDs {
		info, ok := ctx.Game.LastKnownPermanentForEffect(id)
		if !ok {
			continue
		}
		for _, t := range info.Characteristic.Types {
			if strings.EqualFold(t, "land") {
				lands[info.Controller]++
				break
			}
		}
	}
	for _, seat := range effects.SeatsFromController(ctx.Game, ctx.Controller()) {
		n := lands[seat]
		if n == 0 {
			continue
		}
		if err := (effects.SearchLibrary{
			Player:        seat,
			Predicate:     effects.IsBasicLand,
			Dest:          game.ZoneBattlefield,
			Limit:         n,
			UpTo:          true,
			Optional:      true,
			Shuffle:       true,
			TappedOnEntry: true,
			Source:        ctx.Source(),
			Reason:        "Wave of Vitriol — a basic land card for each land sacrificed",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
