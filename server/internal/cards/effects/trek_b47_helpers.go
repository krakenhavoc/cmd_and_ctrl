package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b47_helpers.go — shared pieces for Star Trek Commander batch 47
// (#4974).

// trekB47FaeriesEnterWithTypeCounters is Jean-Luc, Mon Capitaine's
// "Faeries you control enter with an additional +1/+1 counter on them for
// each creature type among non-Borg creatures you already control": a
// CR 614.12 entry replacement over every Faerie entering under the
// source's controller, counted as the Faerie enters
// (trekB4CreatureTypesAmongNonBorg over the battlefield, which the
// entering permanent is not yet on — "already control"). "Faerie" and
// "Borg" are words of the source's text (CR 612.2). The source itself is
// never in the subset it modifies (CR 614.12: an entry replacement comes
// from the entering permanent only when it affects only that permanent).
// A count of zero adds nothing. "Faeries" is the subtype, so a kindred
// Faerie permanent counts as well as a creature.
func trekB47FaeriesEnterWithTypeCounters(label string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventZoneMove},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield || src == nil || ev.CardID == src.InstanceID {
				return false
			}
			entering, ok := g.LookupCardForEffect(ev.CardID)
			return ok && entering.Controller == src.Controller && entering.HasSubtype(src.TextSubtype("Faerie"))
		},
		Replace: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error {
			if n := trekB4CreatureTypesAmongNonBorg(g, src.Controller, src.TextSubtype("Borg")); n > 0 {
				ev.AddCounterAtETB(game.CounterPlusOne, n)
			}
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: label,
	}
}

// trekB47OpponentDrewDuringYourTurn is Troi, Winging It's "whenever an
// opponent draws a card during your turn": one trigger per card an
// opponent draws while the source's controller is the active player.
func trekB47OpponentDrewDuringYourTurn(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
	return ByAnOpponent(ev, source, ch, g) && IsYourTurn(g, source.Controller)
}

// trekB47ArtifactCreaturesYouControl is "artifact creatures you control",
// read on post-layer types, the source included when it is one.
func trekB47ArtifactCreaturesYouControl(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.IsArtifact() && target.IsCreature() && target.Controller == source.Controller
}
