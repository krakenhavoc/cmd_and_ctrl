package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Seifer, Balamb Rival — Legendary Creature — Human Mercenary {2}{B}{R},
// 4/3 (Final Fantasy, #4462):
//
//	"First strike
//	 Whenever you attack a player, goad target creature that player
//	 controls.
//	 Whenever a creature attacking one of your opponents becomes blocked by
//	 two or more creatures, that attacking creature gains deathtouch until
//	 end of turn."
//
// "Whenever you attack a player" is once per player attacked in a
// declaration (OncePerBatchPerPlayer, Long-Range Sensor's shape), and an
// attack at a planeswalker or battle is not an attack at a player (CR
// 508.3e), so it does not trigger. The target is a creature the attacked
// player controls (ControlledByThatPlayer, Alela's clause), goaded by the
// engine's goad (CR 701.15, GoadTarget).
//
// The deathtouch trigger watches the once-per-attacker "becomes blocked"
// event (CR 509.1h), whichever player controls the attacker, and asks as
// it triggers whether the attacker is attacking one of Seifer's opponents
// and how many creatures block it (two or more). The grant is to that
// creature as it resolves and does not care whether its blockers are still
// there (the ruling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "417b9d9c-0a67-4554-a275-a6ae15a04bfb",
		Name:            "Seifer, Balamb Rival",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			OncePerBatchPerPlayer(Targeting(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b16YouAttackedAPlayer(ev, source, g)
			}, "Seifer, Balamb Rival — goad target creature that player controls", GoadTarget),
				ControlledByThatPlayer(TargetCreature("target creature that player controls")))),
			On(game.EventBecomesBlocked, seiferBlockedByTwoOrMore,
				"Seifer, Balamb Rival — that attacking creature gains deathtouch until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil || !onBattlefield(g, item.Trigger.Event.Target) {
						return nil
					}
					return GrantKeywordUntilEOT{
						Target:   item.Trigger.Event.Target,
						Keywords: []string{"deathtouch"},
						Label:    "Seifer, Balamb Rival — deathtouch until end of turn",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

// seiferBlockedByTwoOrMore is "a creature attacking one of your opponents
// becomes blocked by two or more creatures": the blocked attacker is
// attacking a player the source's controller has as an opponent, and at
// least two creatures block it.
func seiferBlockedByTwoOrMore(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventBecomesBlocked {
		return false
	}
	attacker, ok := g.LookupCardForEffect(ev.Target)
	if !ok || !attacker.IsCreature() || !seiferAttacksAnOpponentOf(g, attacker, source.Controller) {
		return false
	}
	return len(blockersOf(g, ev.Target)) >= 2
}

// seiferAttacksAnOpponentOf reports whether the attacker is attacking a
// player (not a planeswalker or battle) who is an opponent of `you`.
func seiferAttacksAnOpponentOf(g *game.Game, attacker game.Card, you uuid.UUID) bool {
	if attacker.AttackingTarget == uuid.Nil || attacker.AttackingTarget == you {
		return false
	}
	return g.PlayerByIDForEffect(attacker.AttackingTarget) != nil
}
