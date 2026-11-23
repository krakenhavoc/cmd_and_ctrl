package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Brine Comber // Brinebound Gift (#1855, ADR 0107 §4) — a disturb card
// whose back face is an Aura.
//
// Front face, Creature — Spirit {1}{W}{U}, 1/1:
//
//	"Whenever this creature enters or becomes the target of an Aura
//	 spell, create a 1/1 white Spirit creature token with flying.
//	 Disturb {W}{U} (You may cast this card from your graveyard
//	 transformed for its disturb cost.)"
//
// Back face, Enchantment — Aura:
//
//	"Enchant creature
//	 Whenever this Aura enters or enchanted creature becomes the target
//	 of an Aura spell, create a 1/1 white Spirit creature token with
//	 flying.
//	 If this Aura would be put into a graveyard from anywhere, exile it
//	 instead."
//
// Each face's ability is one trigger with two conditions. "Becomes the
// target of an Aura spell" is EventBecomesTarget for the creature, from
// a SPELL (not an ability) that is an Aura, read off the stack as the
// targets are chosen (CR 601.2c) — so the Spirit arrives above the Aura
// spell, whether or not that spell resolves. Disturbing Brinebound Gift
// onto a creature makes one Spirit as the Aura enters; the Gift's own
// targeting happened while it was a spell, before its ability existed.
//
// No simplification.
func init() {
	const label = "create a 1/1 white Spirit with flying"
	effects.Register(effects.Spec{
		OracleID:         BrineComberOracleID,
		Name:             "Brine Comber",
		Completeness:     effects.CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Disturb("{W}{U}")},
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventETB, game.EventBecomesTarget}, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return (ev.Kind == game.EventETB && ev.CardID == source.InstanceID) ||
					effects.TargetedByAnAuraSpell(ev, g, source.InstanceID)
			}, "Brine Comber — "+label, effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Spirit with flying"), N: 1})),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     BrineComberOracleID + "#1",
		Name:         "Brinebound Gift",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventETB, game.EventBecomesTarget}, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return (ev.Kind == game.EventETB && ev.CardID == source.InstanceID) ||
					(source.AttachedTo.Kind == game.TargetCard && effects.TargetedByAnAuraSpell(ev, g, source.AttachedTo.ID))
			}, "Brinebound Gift — "+label, effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Spirit with flying"), N: 1})),
		},
		Replacements: []game.ReplacementEffect{effects.DisturbedExile("Brinebound Gift")},
	})
}

const BrineComberOracleID = "8845ba0d-c2f4-49e4-b06e-54a06a8297e0"
