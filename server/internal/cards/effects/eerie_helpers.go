package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// eerie_helpers.go — bodies the Duskmourn eerie creatures share
// (ADR 0103 PR 4). Append-only; the trigger itself is Eerie (rooms.go).

// thisCreatureUntilEOT is "this creature gets +P/+T and gains <keywords>
// until end of turn" — Dashing Bloodsucker, Cult Healer, Erratic
// Apparition, Infernal Phantom. Both halves read the source once, at
// resolution (CR 611.2c), and do nothing if it has left the battlefield.
func thisCreatureUntilEOT(label string, power, toughness int, keywords ...string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if !onBattlefield(g, item.SourceCardID) {
			return nil
		}
		ctx := NewContext(g, item)
		if err := (BoostUntilEOT{Target: item.SourceCardID, Power: power, Toughness: toughness, Label: label}).Apply(ctx); err != nil {
			return err
		}
		if len(keywords) == 0 {
			return nil
		}
		return GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: keywords, Label: label}.Apply(ctx)
	}
}

// eerieDiesWithPower is a "when this creature dies" trigger whose item
// carries the creature's last-known power (CR 603.10) in Params.Amount,
// for a body that deals that much.
func eerieDiesWithPower(label string, targets *game.TargetSpec, effect Effect) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches: []game.EventKind{game.EventLTB},
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			return cardDied(ev, source)
		},
		Targets: targets,
		Key:     label,
		Build: func(_ game.Event, source *game.Card, lki game.Characteristic, g *game.Game) *game.StackItem {
			item := game.NewTriggeredItem(source, label)
			item.Params.Amount = b13LastKnownPower(g, source.InstanceID, lki)
			return item
		},
		Effect: effect,
	}
}

// targetedPlayerOf is the first player a resolving item targets, or
// uuid.Nil when none is left legal.
func targetedPlayerOf(ctx *Context) uuid.UUID {
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			return t.ID
		}
	}
	return uuid.Nil
}
