package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sealock Monster — Creature — Octopus {3}{U}{U}, 5/5:
//
//	"This creature can't attack unless defending player controls an
//	 Island.
//	 {5}{U}{U}: Monstrosity 3. (If this creature isn't monstrous, put
//	 three +1/+1 counters on it and it becomes monstrous.)
//	 When this creature becomes monstrous, target land becomes an Island
//	 in addition to its other types."
//
// The restriction is ADR 0107 §2's (#1879, CR 508.1c), with the defending
// player worked out per target (CR 508.5, 508.5a). The trigger is the
// card's own answer to it: aimed at an opponent's land, it gives that
// opponent an Island to be attacked through.
//
// "Becomes an Island in addition to its other types" adds the subtype: the
// land keeps its land types and rules text and gains the Island's mana
// ability (CR 205.1b, 305.7's last sentence, 305.6). No duration is
// printed, so the effect lasts as long as
// that land stays that object (CR 611.2a, 611.2c), the way The Legend of
// Kyoshi's chapter II does. It is a data record pinned to the land, so a
// table holding one is still a restore point.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "351c4f85-8792-4710-8cc9-d3e36657f6db",
		Name:         "Sealock Monster",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			CantAttackUnlessDefendingPlayerControls(QuerySubtype("Island")),
		},
		Activated: []ActivatedAbility{Monstrosity(ManaCost("{5}{U}{U}"), 3)},
		Triggered: []game.TriggeredAbility{sealockMonsterTrigger()},
	})
}

func sealockMonsterTrigger() game.TriggeredAbility {
	t := WhenBecomesMonstrous("Sealock Monster — target land becomes an Island in addition to its other types",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			land := FirstLegalBattlefieldTarget(ctx)
			if land == uuid.Nil {
				return nil
			}
			return ScopedEffectFor{
				Target:   land,
				Mods:     []game.Mod{game.AddSubtypesMod("Island")},
				Duration: g.PinnedTo(game.IndefiniteDuration(), land),
				Label:    "Sealock Monster — that land is an Island in addition to its other types",
			}.Apply(ctx)
		})
	t.Targets = TargetPermanent("target land", Land())
	return t
}
