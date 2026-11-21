package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_one_b02_helpers.go — Phyrexia: All Will Be One's Skullbombs and
// Sphere lands (S82, #5479).

// s82OneOilCounter is the oil counter's on-disk kind.
const S82OneOilCounter = "oil"

// s82OneSkullbomb is the Skullbomb cycle: a {1} artifact with "{1},
// Sacrifice this artifact: Draw a card." and a second, sorcery-speed
// "<cost>, Sacrifice this artifact: <effect>. Draw a card." `second` is
// that second ability with its Cost's mana part only; the sacrifice is
// added here. `apply` runs the effect on the ability's chosen target and
// the card is drawn after it, so an ability whose only target became
// illegal does nothing at all (CR 608.2b) and draws nothing.
func S82OneSkullbomb(oracleID, name, label string, manaCost string, targets *game.TargetSpec,
	purpose game.Purpose, apply func(ctx *Context, target uuid.UUID) error) Spec {
	return Spec{
		OracleID:     oracleID,
		Name:         name,
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{1}, Sacrifice this artifact: Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    Plus(ManaCost("{1}"), SacrificeThis()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
				},
			},
			{
				Label:        label,
				Purpose:      purpose,
				Cost:         Plus(ManaCost(manaCost), SacrificeThis()),
				Targets:      targets,
				SorcerySpeed: true,
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					if err := apply(ctx, id); err != nil {
						return err
					}
					return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
				},
			},
		},
	}
}

// s82OneSphereDrawLand is the five tapped Sphere lands: "This land
// enters tapped. {T}: Add <colour>. {1}<colour>, {T}, Sacrifice this
// land: Draw a card."
func S82OneSphereDrawLand(oracleID, name, colour string) Spec {
	return Spec{
		OracleID:     oracleID,
		Name:         name,
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{" + colour + "}",
			Label:    "Add {" + colour + "}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{1}{" + colour + "}, {T}, Sacrifice this land: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{1}{"+colour+"}"), TapCost(), SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
			},
		}},
	}
}

// s82OneYourArtifactOrCreature is "target artifact or creature you
// control".
func S82OneYourArtifactOrCreature() *game.TargetSpec {
	return TargetPermanent("target artifact or creature you control", And(Or(Artifact(), Creature()), YouControl()))
}

// s82OnePutOilOnTarget is "put an oil counter on target artifact or
// creature you control" as an activated ability's whole effect.
func S82OnePutOilOnTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return AddCounter{Target: id, Kind: S82OneOilCounter, N: 1}.Apply(ctx)
}
