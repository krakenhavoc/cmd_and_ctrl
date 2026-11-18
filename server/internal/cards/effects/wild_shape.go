package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wild Shape — Instant {G} (Adventures in the Forgotten Realms, #5499):
//
//	"Choose one. Until end of turn, target creature you control has that
//	 base power and toughness, becomes that creature type, and gains that
//	 ability.
//	 • 1/3 Turtle with hexproof.
//	 • 1/5 Spider with reach.
//	 • 3/3 Elephant with trample."
//
// One until-end-of-turn record on the target with three mods at one
// timestamp, the way Glamer Gifter writes its base size: base power and
// toughness at layer 7b (pumps and counters still apply on top), the
// creature type replacing its others at layer 4 (CR 205.1a; its card types
// stay), and the keyword at layer 6.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "99f4a22a-ba33-43f0-bc58-9df8388706cd",
		Name:         "Wild Shape",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			wildShapeMode("1/3 Turtle with hexproof.", 1, 3, "Turtle", "hexproof"),
			wildShapeMode("1/5 Spider with reach.", 1, 5, "Spider", "reach"),
			wildShapeMode("3/3 Elephant with trample.", 3, 3, "Elephant", "trample"),
		),
	})
}

// wildShapeMode is one bullet: the target creature you control has the
// base power and toughness, becomes the creature type and gains the
// keyword until end of turn.
func wildShapeMode(label string, power, toughness int, creatureType, keyword string) game.ModeOption {
	return ModeWithPurpose(ModeDoing(label,
		TargetCreature("target creature you control", YouControl()),
		rfCreatureBModeTargetDoes(func(ctx *Context, id uuid.UUID) error {
			return untilEndOfTurn(ctx, id, nil, "Wild Shape — "+label,
				game.SetBasePowerMod(power), game.SetBaseToughnessMod(toughness),
				game.SetCreatureTypesMod(creatureType), game.AddKeywordsMod(keyword))
		})), ForTargets(BenefitUntilEOT(0, 0, 0, keyword)))
}
