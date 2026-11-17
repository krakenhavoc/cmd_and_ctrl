package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Unctus, Grand Metatect — Legendary Artifact Creature — Phyrexian
// Vedalken {1}{U}{U}, 2/4 (ONE):
//
//	"Other blue creatures you control have "Whenever this creature
//	 becomes tapped, draw a card, then discard a card."
//	 Other artifact creatures you control get +1/+1.
//	 {U/P}: Until end of turn, target creature you control becomes a blue
//	 artifact in addition to its other colors and types. Activate only as
//	 a sorcery."
//
// The first line is a layer 6 grant of a catalog bundle whose trigger
// is the HOST's own (ADR 0093: "this creature" is the recipient, and its
// controller loots), so it follows the recipient through control changes
// and a removal on the recipient takes it. Whether a creature is blue is
// read through the layered view, so the third ability's recolour feeds the
// first line. The anthem is a layer 7c +1/+1 on the other artifact
// creatures its controller controls (CR 613.4c); the third ability makes
// the target an artifact for the turn, which the anthem then sees. {U/P}
// is Phyrexian mana, paid with {U} or 2 life (CR 107.4f).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "54c18f09-5c5a-46e7-9f66-ff91bc12537c",
		Name:         "Unctus, Grand Metatect",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: unctusLootGrant,
			Triggered: []game.TriggeredAbility{
				OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, ThisBecameTapped,
					"Unctus, Grand Metatect — draw a card, then discard a card", b35LootOne),
			},
			Text: "Whenever this creature becomes tapped, draw a card, then discard a card.",
		}},
		Static: []game.StaticAbility{
			GrantAbilities(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID != source.InstanceID && target.Controller == source.Controller &&
					target.IsCreature() && target.HasColor("U")
			}, unctusLootGrant),
			{
				Layer:    game.Layer7PT,
				SubLayer: game.SubLayer7C_Modify,
				Label:    "Other artifact creatures you control get +1/+1.",
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.InstanceID != source.InstanceID && target.Controller == source.Controller &&
						target.IsCreature() && target.IsArtifact()
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power++
					c.Toughness++
				},
			},
		},
		Activated: []ActivatedAbility{{
			Label:        "{U/P}: Until end of turn, target creature you control becomes a blue artifact in addition to its other colors and types. Activate only as a sorcery.",
			Purpose:      game.Purpose{Answers: game.AnswerValue},
			Cost:         ManaCost("{U/P}"),
			Targets:      TargetCreature("target creature you control", YouControl()),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id := firstLegalCardTarget(ctx)
				if id == uuid.Nil || !onBattlefield(g, id) {
					return nil
				}
				return ScopedEffectFor{
					Target:   id,
					Mods:     []game.Mod{game.AddColorsMod("U"), game.AddTypesMod("Artifact")},
					Duration: g.UntilEndOfTurnDuration(),
					Label:    "Unctus, Grand Metatect — becomes a blue artifact until end of turn",
				}.Apply(ctx)
			},
		}},
	})
}

// unctusLootGrant is the on-disk key of Unctus's granted loot trigger.
const unctusLootGrant = "unctus-grand-metatect/loot-when-tapped"
