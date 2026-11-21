package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Circle of Confinement — Enchantment {1}{W} (Innistrad: Crimson Vow):
//
//	"When this enchantment enters, exile target creature an opponent
//	 controls with mana value 3 or less until this enchantment leaves the
//	 battlefield.
//	 Whenever an opponent casts a Vampire spell with the same name as a
//	 card exiled with this enchantment, you gain 2 life."
//
// The exile is Banishing Light's until-this-leaves (CR 610.3), linked to the
// enters trigger (CR 607.2a) so the second ability can read the exiled
// card's name while it is still held. A token is exiled and gone, so it has
// no name to match. The cast trigger reads the spell as it is cast: an
// opponent's Vampire spell, by the text of this enchantment (CR 612.2), whose
// name is one of the exiled cards'. A second Circle's exiles are not this
// one's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d1cd7a93-4b83-4755-90bc-17cb64bfb131",
		Name:         "Circle of Confinement",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters(circleOfConfinementExileLabel, func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						return effects.ExileUntil{Target: t.ID, ThisLeaves: true, Linked: true,
							Label: "Circle of Confinement — the exiled card returns when Circle of Confinement leaves the battlefield"}.Apply(ctx)
					}
					return nil
				}),
				effects.TargetCreature("target creature an opponent controls with mana value 3 or less", effects.OpponentControls(), effects.ManaValueLE(3))),
			effects.On(game.EventCast, circleOfConfinementCastsTheSameName,
				"Circle of Confinement — you gain 2 life", effects.Do(effects.GainLife{Amount: 2})),
		},
	})
}

const circleOfConfinementExileLabel = "Circle of Confinement — exile target creature an opponent controls with mana value 3 or less until this enchantment leaves the battlefield"

func circleOfConfinementCastsTheSameName(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor == source.Controller {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || !spell.HasSubtype(source.TextSubtype("Vampire")) {
		return false
	}
	for _, c := range effects.VowB16ExiledWithCard(g, source, circleOfConfinementExileLabel) {
		if c.Name == spell.Name {
			return true
		}
	}
	return false
}
