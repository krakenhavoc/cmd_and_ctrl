package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Draugr Necromancer — Snow Creature — Zombie Cleric {3}{B}, 4/4
// (Kaldheim; OLD b066, #6161):
//
//	"If a nontoken creature an opponent controls would die, exile that
//	 card with an ice counter on it instead.
//	 You may cast spells from among cards in exile your opponents own
//	 with ice counters on them, and you may spend mana from snow sources
//	 as though it were mana of any color to cast those spells."
//
// The replacement is Kalitas's (OpponentNontokenCreatureWouldDie) with
// Dauthi Voidwalker's ice-counter stamp (ReplacementEvent.ExileCounter),
// which lands only if the card really reaches exile. The cast is a static
// permission over exiled cards an opponent owns with an ice counter
// (Tinybones' filter), for spells only, at the spell's normal timing, and
// its spending clause is CastPermission.AnyColorFromSnow (CR 609.4b): each
// coloured symbol also takes any mana a snow source produced, while other
// mana pays as printed. Any card may carry an ice counter; the permission
// reads the counter, not who put it there, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ccdf3399-b836-47ab-802d-af5a9c24d759",
		Name:         "Draugr Necromancer",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches:   []game.EventKind{game.EventZoneMove},
			AppliesTo: effects.OpponentNontokenCreatureWouldDie,
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.NewZone = game.ZoneExile
				ev.NewZoneOwner = uuid.Nil
				ev.ExileCounter = draugrIceCounter
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Draugr Necromancer: exile with an ice counter instead of dying",
		}},
		CastPermissions: []game.CastPermission{{
			Zone: game.ZoneExile,
			Filter: game.PermissionFilter{
				NotOwnedByHolder: true,
				WithCounter:      draugrIceCounter,
			},
			CastOnly:         true,
			AnyColorFromSnow: true,
			Label:            "Cast a spell an opponent owns with an ice counter from exile (Draugr Necromancer)",
		}},
	})
}

const draugrIceCounter = "ice"
