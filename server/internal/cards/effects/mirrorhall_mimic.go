package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mirrorhall Mimic // Ghastly Mimicry (#1855, ADR 0107 §4) — a disturb
// card whose back face is an Aura.
//
// Front face, Creature — Spirit {3}{U}, 0/0:
//
//	"You may have this creature enter as a copy of any creature on the
//	 battlefield, except it's a Spirit in addition to its other types.
//	 Disturb {3}{U}{U} (You may cast this card from your graveyard
//	 transformed for its disturb cost.)"
//
// Back face, Enchantment — Aura:
//
//	"Enchant creature
//	 At the beginning of your upkeep, create a token that's a copy of
//	 enchanted creature, except it's a Spirit in addition to its other
//	 types.
//	 If Ghastly Mimicry would be put into a graveyard from anywhere,
//	 exile it instead."
//
// The front face is Clone with Phantasmal Image's kind of except clause
// (EntersAsCopyOf, CR 707.9a): the copiable values gain the Spirit
// subtype, so the copy is a Spirit and still everything it copied. Not
// targeting, so hexproof does not stop it. Declined, or with nothing to
// copy, it is a 0/0 and dies (CR 704.5f) — into the graveyard, where it
// can be disturbed.
//
// The back face's token is the shared token copy (CreateTokenCopy) of
// the creature the Aura is on as the trigger resolves, with the same
// exception written onto the token's type line.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         mirrorhallMimicOracleID,
		Name:             "Mirrorhall Mimic",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Disturb("{3}{U}{U}")},
		Replacements: []game.ReplacementEffect{
			EntersAsCopyOf("Mirrorhall Mimic", anyCreatureOnBattlefield,
				func(_ *game.ReplacementEvent, v *game.PrintedValues, _ *game.Game, _ *game.Card) {
					v.AddSubtype("Spirit")
				}),
		},
	})
	Register(Spec{
		OracleID:     mirrorhallMimicOracleID + "#1",
		Name:         "Ghastly Mimicry",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Ghastly Mimicry — create a token copy of enchanted creature, a Spirit in addition", ghastlyMimicryUpkeep),
		},
		Replacements: []game.ReplacementEffect{DisturbedExile("Ghastly Mimicry")},
	})
}

const mirrorhallMimicOracleID = "5768fe50-a134-492c-a725-5ed02610c39f"

// ghastlyMimicryUpkeep copies the creature the Aura is attached to as
// the trigger resolves. An Aura that has left the battlefield is
// attached to nothing, and the trigger creates nothing.
func ghastlyMimicryUpkeep(g *game.Game, item *game.StackItem) error {
	aura, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok || !onBattlefield(g, item.SourceCardID) || aura.AttachedTo.Kind != game.TargetCard || aura.AttachedTo.ID == uuid.Nil {
		return nil
	}
	return CreateTokenCopy{
		Controller: item.Controller,
		Copy:       aura.AttachedTo.ID,
		N:          1,
		Except: func(t *game.Card) {
			t.TypeLine = withSubtypeInAddition(t.TypeLine, "Spirit")
		},
	}.Apply(NewContext(g, item))
}
