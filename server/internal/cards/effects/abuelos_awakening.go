package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Abuelo's Awakening — Sorcery {X}{3}{W} (The Lost Caverns of Ixalan,
// #4470):
//
//	"Return target artifact or non-Aura enchantment card from your
//	 graveyard to the battlefield with X additional +1/+1 counters on it.
//	 It's a 1/1 Spirit creature with flying in addition to its other
//	 types."
//
// Excava's entry overwrite (ADR 0210): the card enters as the permanent,
// with the X counters as part of its entry (CR 614.1c, through the counter
// pipeline), and it is a 1/1 Spirit creature with flying in addition to its
// other types, for as long as it is that permanent. The X is the one the
// spell was cast with, and 0 puts on none. The target is re-checked as the
// spell resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2b22cf44-7b03-4919-9812-6796bf193484",
		Name:         "Abuelo's Awakening",
		Completeness: CompletenessFull,
		XMatters:     true,
		Targets: TargetCardInGraveyard("target artifact or non-Aura enchantment card from your graveyard",
			YouOwn(), Or(Artifact(), And(Enchantment(), Not(Subtype("Aura"))))),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			opts := game.ZoneEntryOptions{
				Controller: item.Controller,
				EntersAs: &game.EntersAs{
					AddTypes:    []string{"Creature"},
					AddSubtypes: ctx.TextSubtypes("Spirit"),
					Keywords:    []string{"flying"},
					BasePT:      &game.EntersAsPT{Power: 1, Toughness: 1},
					Label:       "Abuelo's Awakening — a 1/1 Spirit creature with flying in addition to its other types",
				},
			}
			if x := ctx.X(); x > 0 {
				opts.EntersWithCounters = map[string]int{game.CounterPlusOne: x}
			}
			return putCardsFromZoneTogetherWith(ctx, game.ZoneGraveyard, []uuid.UUID{t.ID}, opts, game.TailRef{}, 0)
		},
	})
}
