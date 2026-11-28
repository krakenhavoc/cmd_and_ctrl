package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b652_helpers.go — Invasion, S82 older sets batch 652 (#5616).

// OldB652OtherCreaturesOfColorGet is "Other <colour> creatures get +p/+t": a
// layer 7c static over every other creature of that colour, whoever controls
// it (Urborg Shambler). The colour is a word of the source's text (CR 612.2)
// and is read through the layers (CR 105.2).
func OldB652OtherCreaturesOfColorGet(color string, power, toughness int, label string) game.StaticAbility {
	return game.StaticAbility{
		Label:    label,
		Layer:    game.Layer7PT,
		SubLayer: game.SubLayer7C_Modify,
		AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.IsCreature() && target.InstanceID != source.InstanceID &&
				slices.Contains(target.EffectiveColors(), source.TextColor(color))
		},
		Apply: PumpBy(power, toughness),
	}
}

// OldB652PumpPerOpponentsPermanent is "This creature gets +p/+t for each
// <land type> your opponents control" (the Knights of Invasion): a layer 7c
// static whose size is the number of permanents with that subtype controlled by
// players other than the source's controller, counted as the layers stand. The
// subtype is a word of the source's text (CR 612.2).
func OldB652PumpPerOpponentsPermanent(subtype string, power, toughness int, label string) game.StaticAbility {
	return game.StaticAbility{
		Label:     label,
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7C_Modify,
		AppliesTo: SelfCreature,
		Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
			word := source.TextSubtype(subtype)
			n := 0
			for _, p := range g.Seats {
				if p == nil || p.ID == source.Controller || p.ID == uuid.Nil {
					continue
				}
				n += len(PermanentsControlledByMatching(g, p.ID, Subtype(word)))
			}
			c.Power += n * power
			c.Toughness += n * toughness
		},
	}
}

// OldB652LabeledKeywordGrant is KeywordGrant with the printed clause as its
// label: "This creature has <keyword> as long as <condition>" (Tek).
func OldB652LabeledKeywordGrant(label, keyword string, applies func(target *game.Card, g *game.Game, source *game.Card) bool) game.StaticAbility {
	grant := KeywordGrant(applies, keyword)
	grant.Label = label
	return grant
}

// OldB652PumpPerBlocker is "it gets +1/+1 until end of turn for each creature
// blocking it" on the trigger's own source (Sparring Golem): the blockers are
// counted as the trigger resolves, and a source that has left, or come back as
// a new object, gets nothing (CR 400.7).
func OldB652PumpPerBlocker(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if !SourceIsStillThisPermanent(g, item) {
			return nil
		}
		n := g.CreaturesBlockingForEffect(item.SourceCardID)
		return BoostUntilEOT{Target: item.SourceCardID, Power: n, Toughness: n, Label: label}.Apply(NewContext(g, item))
	}
}
