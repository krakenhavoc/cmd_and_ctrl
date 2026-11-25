package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b106_helpers.go — shared pieces of Core Set 2021 batch 106 (#5511).

// S82OldB106ShrinesYouControl is "the number of Shrines you control": the
// battlefield permanents `player` controls with the Shrine subtype, read
// through the layered view.
func S82OldB106ShrinesYouControl(g *game.Game, player uuid.UUID) int {
	return CountControlled(g, player, func(c game.Card) bool { return c.HasSubtype("Shrine") })
}

// S82OldB106NoncombatDamageToOpponentsThisTurn is "the total amount of
// noncombat damage dealt to your opponents this turn": every noncombat damage
// event this turn whose victim is a player other than `you`, from any source
// (Chandra's Incinerator). The tally records the amount actually dealt.
func S82OldB106NoncombatDamageToOpponentsThisTurn(g *game.Game, you uuid.UUID) int {
	n := 0
	for _, r := range g.TurnTally.NoncombatDamage {
		if r.Victim != uuid.Nil && r.Victim != you {
			n += r.Amount
		}
	}
	return n
}

// S82OldB106CreatureOrPlaneswalkerThatPlayerControls is the TargetsFrom
// clause "target creature or planeswalker that player controls" for a trigger
// about damage dealt to a player: "that player" is the player the triggering
// event names.
func S82OldB106CreatureOrPlaneswalkerThatPlayerControls(tc game.TriggerContext, _ *game.Card, g *game.Game) *game.TargetSpec {
	player := tc.Event.Target
	if player == uuid.Nil || g.PlayerByIDForEffect(player) == nil {
		return nil
	}
	return &game.TargetSpec{
		Mode: "permanent", Label: "target creature or planeswalker that player controls",
		Zones: []game.ZoneKind{game.ZoneBattlefield},
		CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
			return (c.IsCreature() || c.IsPlaneswalker()) && c.Controller == player
		},
		Min: 1, Max: 1,
	}
}
