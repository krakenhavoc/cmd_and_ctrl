package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Szarel, Genesis Shepherd — Legendary Creature — Insect Druid
// {2}{B}{R}{G}, 2/5 (Edge of Eternities Commander):
//
//	"Flying
//	 You may play lands from your graveyard.
//	 Whenever you sacrifice another nontoken permanent during your turn,
//	 put a number of +1/+1 counters equal to Szarel's power on up to one
//	 other target creature."
//
// Crucible of Worlds' standing land permission and a per-permanent
// sacrifice watcher (cmdrB9YouSacrificedAnotherNontoken): the event fires
// while the permanent is still on the battlefield, so a Szarel sacrificed
// with others still triggers for each of the others (the ruling). The
// power is read as the trigger resolves, last known if Szarel has left
// (CR 608.2h, Context.SourcePermanent); a power of zero or less puts
// nothing on. "Up to one other target creature" is a target clause with a
// minimum of zero that excludes the source.
//
// No simplification.
func init() {
	spec := Another(TargetCreature("up to one other target creature"))
	spec.Min = 0
	Register(Spec{
		OracleID:        "1f3c3365-7dc8-47e1-b304-2a09aa229292",
		Name:            "Szarel, Genesis Shepherd",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		CastPermissions: []game.CastPermission{{
			Zone:   game.ZoneGraveyard,
			Filter: game.PermissionFilter{LandsOnly: true},
			Label:  "Play a land from your graveyard (Szarel, Genesis Shepherd)",
		}},
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventSacrifice, func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
					return IsYourTurn(g, source.Controller) && cmdrB9YouSacrificedAnotherNontoken(ev, source, ch, g)
				}, "Szarel — put +1/+1 counters equal to its power on up to one other target creature",
					szarelCounters),
				spec),
		},
	})
}

// szarelCounters puts Szarel's power in +1/+1 counters on the target.
func szarelCounters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	info, ok := ctx.SourcePermanent()
	if !ok || info.Power <= 0 {
		return nil
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: info.Power}.Apply(ctx)
		}
	}
	return nil
}
