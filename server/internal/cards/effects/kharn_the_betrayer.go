package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Khârn the Betrayer — Legendary Creature — Astartes Berserker {3}{R}, 5/1:
//
//	"Berzerker — Khârn the Betrayer attacks or blocks each combat if able.
//	 Sigil of Corruption — When you lose control of Khârn the Betrayer, draw two cards.
//	 The Betrayer — If damage would be dealt to Khârn the Betrayer, prevent that damage and an opponent of your choice gains control of it."
//
// The Betrayer is a prevention static with an additional effect (ADR 0108
// §8), once per recipient in a damage instance, and the additional effect
// asks a question as the damage is prevented (#2040, ADR 0108 amendment of
// 2026-10-16): the rulings say the control change "is part of the
// replacement effect and happens immediately, even if you are in the
// middle of resolving the spell or ability", and that a spell damaging it
// several times hands it to a new player each time. So the question is
// asked of the player who controls Khârn when the follow-up runs, which,
// for the second instance, is the player the first answer gave it to; the
// rest of the damage instruction, and the state-based actions after
// combat damage, wait for the answer. With one opponent there is no
// question. Damage that can't be prevented is dealt, and control still
// changes (CR 615.12).
//
// Berzerker is "attacks each combat if able" and "blocks each combat if
// able": a creature can only attack on its controller's turn and only
// block on another's, so the two requirements never meet.
//
// "You lose control" is both of the rulings' ways: another player gaining
// control of it (the player who lost it controls the trigger,
// WhenYouLoseControlOfThis) and it leaving the battlefield. The triggers
// go on the stack after the spell that moved it has finished, in APNAP
// order, as every trigger does (CR 603.3b; the ruling).
//
// No simplifications.
func init() {
	kharnOpponentChosen = OptionPickThen("option-pick/kharn-the-betrayer-opponent", kharnGainControl)
	Register(Spec{
		OracleID:     "3246eff5-8739-469f-a401-e5de927dd6bb",
		Name:         "Khârn the Betrayer",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			AttacksEachCombat(),
			BlocksEachCombat(),
		},
		Replacements: []game.ReplacementEffect{
			PreventDamageDealtTo(PreventionStatic{
				To:    ToThisCreature,
				Then:  kharnBetrayerBody,
				Label: "Khârn the Betrayer — prevent the damage and an opponent of your choice gains control of it",
			}),
		},
		Triggered: []game.TriggeredAbility{
			WhenYouLoseControlOfThis("Khârn the Betrayer — Sigil of Corruption: draw two cards", Do(DrawCards{N: 2})),
			WhenThisLeaves("Khârn the Betrayer — Sigil of Corruption: draw two cards", Do(DrawCards{N: 2})),
		},
	})
}

// kharnBetrayerBody is the additional effect: an opponent of the
// controller's choice gains control of it.
var kharnBetrayerBody = game.DelayedBody("kharn-the-betrayer/opponent-gains-control", kharnBetrayer)

// kharnOpponentChosen is the keyed answer to "which opponent". Assigned
// in init: the key is an on-disk identity, never renamed, never reused.
var kharnOpponentChosen game.OptionPickThen

func kharnBetrayer(g *game.Game, item *game.StackItem, _ game.EffectParams) error {
	info, ok := followUpThis(g, item)
	if !ok {
		return nil
	}
	// "Your choice": whoever controls it now, which is the player an
	// earlier answer in the same resolution handed it to.
	you := info.Controller
	var opponents []uuid.UUID
	for _, p := range g.Seats {
		if p != nil && !p.Eliminated && p.ID != you {
			opponents = append(opponents, p.ID)
		}
	}
	switch len(opponents) {
	case 0:
		return nil
	case 1:
		kharnTakeControl(g, item.SourceCardID, opponents[0])
		return nil
	}
	g.QueueChoosePlayerForEffect(game.ChoosePlayerPrompt{
		Chooser:  you,
		Among:    opponents,
		Source:   item.SourceCardID,
		Question: "Khârn the Betrayer — choose an opponent to gain control of it",
		ThenKey:  kharnOpponentChosen,
	})
	return nil
}

// kharnGainControl hands Khârn to the chosen opponent. Nobody chosen (the
// chooser left the game) changes nothing.
func kharnGainControl(ctx *Context, r game.OptionPicked) error {
	if r.Option == nil || r.Option.Player == uuid.Nil {
		return nil
	}
	kharnTakeControl(ctx.Game, r.Source, r.Option.Player)
	return nil
}

func kharnTakeControl(g *game.Game, kharn, to uuid.UUID) {
	if !g.GainControlForEffect(kharn, kharn, to, game.IndefiniteDuration(), "Khârn the Betrayer — an opponent gains control of it") {
		return
	}
	// "Happens immediately": the new controller is the controller for
	// the rest of the instruction, and for the next time it is damaged.
	g.RecomputeLayersIfStaleLocked()
}
