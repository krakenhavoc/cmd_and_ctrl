package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// defender_waivers.go — the card-facing half of ADR 0106's 2026-10-16
// amendment (#1936): "can attack … as though it didn't have defender"
// when the waiver names whom the creature may attack, or must leave the
// keyword in place. The engine half (the per-target check every
// declaration path shares) is game/defender_waivers.go.
//
//	Static: []game.StaticAbility{CanAttackPlayersWhoAttackedYouAsThoughNoDefender()}, // Weathered Sentinels
//	Effect: CanAttackThisTurnAsThoughNoDefender("Hightide Hermit — …"),             // Hightide Hermit
//
// The older cards that lift defender for every target (Stalked
// Researcher, Ghalta the Immovable) write it as losing the keyword, and
// keep doing so: for them the two differ only to a card that counts
// creatures with defender.

// CanAttackPlayersWhoAttackedYouAsThoughNoDefender is "This creature can
// attack players who attacked you during their last turn as though it
// didn't have defender" (Weathered Sentinels): a self-only layer-6
// static writing a DefenderWaiver the engine asks of each target. "You"
// is the creature's controller and "their last turn" is ADR 0108 §6's
// record, both read live at declaration. A planeswalker or battle is
// never a player it may attack this way.
func CanAttackPlayersWhoAttackedYouAsThoughNoDefender() game.StaticAbility {
	return game.StaticAbility{
		Label: "This creature can attack players who attacked you during their last turn as though it didn't have defender.",
		Layer: game.Layer6Ability,
		AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
			return target.IsCreature() && selfOnly(target, g, source)
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
			w := game.DefenderWaiver{OnlyPlayersWhoAttackedYou: true}
			if source != nil {
				w.Source, w.SourceName = source.InstanceID, source.Name
			}
			c.DefenderWaivers = append(c.DefenderWaivers, w)
		},
	}
}

// CanAttackThisTurnAsThoughNoDefender is the resolution of "This
// creature can attack this turn as though it didn't have defender"
// (Hightide Hermit): an until-end-of-turn ModWaiveDefender on the
// ability's source (CR 514.2 ends "this turn" effects with "until end of
// turn" ones). The keyword stays. Nothing happens if the source has left
// the battlefield: a new object is not "this creature" (CR 400.7).
func CanAttackThisTurnAsThoughNoDefender(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if !onBattlefield(g, item.SourceCardID) {
			return nil
		}
		return waiveDefenderUntilEOT(NewContext(g, item), item.SourceCardID, label)
	}
}

func waiveDefenderUntilEOT(ctx *Context, id uuid.UUID, label string) error {
	return untilEndOfTurn(ctx, id, nil, label, game.WaiveDefenderMod())
}
