package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dance of the Manse — Sorcery {X}{W}{U}:
//
//	"Return up to X target artifact and/or non-Aura enchantment cards each
//	 with mana value X or less from your graveyard to the battlefield. If
//	 X is 6 or more, those permanents are 4/4 creatures in addition to
//	 their other types."
//
// The target clause is Crashing Wave's "up to X" count (CountFromX, UpToX)
// bound to the announced X on the card's mana value (WithManaValueAtMostX,
// re-checked on resolution, CR 608.2b). The cards are returned together as
// one entry (CR 400.7), and when X is 6 or more they enter as 4/4
// creatures in addition to their other types (Abuelo's Awakening's entry
// overwrite, ADR 0210), for as long as they are those permanents.
//
// No simplification.
func init() {
	targets := effects.TargetCardInGraveyard("up to X target artifact and/or non-Aura enchantment cards each with mana value X or less from your graveyard",
		effects.YouOwn(), effects.Or(effects.Artifact(), effects.And(effects.Enchantment(), effects.Not(effects.Subtype("Aura"))))).WithManaValueAtMostX()
	targets.CountFromX = true
	targets.UpToX = true
	effects.Register(effects.Spec{
		OracleID:     "fdfac093-a5c0-47b4-9052-e0af517d83cf",
		Name:         "Dance of the Manse",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Targets:      targets,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			var ids []uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					ids = append(ids, t.ID)
				}
			}
			opts := game.ZoneEntryOptions{Controller: item.Controller}
			if ctx.X() >= 6 {
				opts.EntersAs = &game.EntersAs{
					AddTypes: []string{"Creature"},
					BasePT:   &game.EntersAsPT{Power: 4, Toughness: 4},
					Label:    "Dance of the Manse — a 4/4 creature in addition to its other types",
				}
			}
			return effects.PutCardsFromZoneTogetherWith(ctx, game.ZoneGraveyard, ids, opts, game.TailRef{}, 0)
		},
	})
}
