package cardsto

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Twinning Staff — Artifact {3} (Commander 2020, reprinted in New Capenna
// Commander; #5918, #5908):
//
//	"If you would copy a spell one or more times, instead copy it that
//	 many times plus an additional time. You may choose new targets for
//	 the additional copy.
//	 {7}, {T}: Copy target instant or sorcery spell you control. You may
//	 choose new targets for the copy."
//
// The first ability is a replacement of the copy instruction (CR 614.1a,
// RepEventCopySpell, game/spell_copy_event.go), which every spell copy the
// engine makes goes through once per instruction: storm's copies, a
// replicate payment's, "copy that spell twice", a casualty or conspire
// copy, a Chain's "copy this spell", a copy made for each creature
// (CR 707.10e) and the Staff's own activated ability. "You" is the player
// who would copy, the copies' controller (CR 707.10). Two Staffs add two
// copies, each applying once (CR 614.5); the order the player would
// choose between them (CR 616.1) cannot change the count. The additional
// copy is always offered new targets (CR 707.10c), whatever the
// instruction says about its own copies.
//
// The second is Lithoform Engine's spell ability at {7}, and its own copy
// instruction is replaced too: with the Staff out, it makes two copies.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3ca8113b-f5b8-41a7-aae6-ebb32554dfe3",
		Name:         "Twinning Staff",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{copyASpellOneMoreTime("Twinning Staff")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{7}, {T}: Copy target instant or sorcery spell you control",
			Cost:    effects.Plus(effects.ManaCost("{7}"), effects.TapCost()),
			Targets: effects.TargetInstantOrSorcerySpell("target instant or sorcery spell you control", effects.YouControl()),
			Effect:  effects.CopyTargetedSpell,
		}},
	})
}

// copyASpellOneMoreTime is "If you would copy a spell one or more times,
// instead copy it that many times plus an additional time. You may choose
// new targets for the additional copy." on the copy instruction's window,
// for the permanent's controller. `name` is the card, for the label.
func copyASpellOneMoreTime(name string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventSpellCopied},
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
			return ev.Kind == game.RepEventCopySpell && ev.CopyCount > 0 && ev.CopyPlayer == src.Controller
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.CopyOneMoreWithNewTargets()
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label:    name + ": copy the spell an additional time",
		RowLabel: "If you would copy a spell one or more times, instead copy it that many times plus an additional time. You may choose new targets for the additional copy.",
	}
}
