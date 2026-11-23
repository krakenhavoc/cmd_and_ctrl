package testsa

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// answers_spells_protective_test.go — ADR 0009's amendment of 2026-10-30
// (#4734), slice 6's first batch: ADR 0142's #2877 sweep of the catalog's
// undeclared instants, protective kinds first. Each card declares on the
// card the union of what it can give (ADR 0142 sweep ruling 4):
// indestructible, regenerate and phasing out are survive, hexproof,
// shroud, protection and "can't be countered" are protect, a fog or a
// damage shield is prevent, +N/+N and +1/+1 counters are pump, a combat
// keyword or evasion is combat_grant, and creature tokens made at instant
// speed (amass, populate) are makes_blocker. None of them is modal,
// kicked or cast for an alternative cost that changes the spell, so the
// card's slot is the only one.
//
// Left for later batches: a card that also deals damage, destroys or
// exiles (its `remove` keeps today's windows, so declaring it changes no
// stop), and Fog, which internal/legal's tests keep as their undeclared
// cast.

// TestProtectiveInstantsDeclare pins every card of the batch.
func TestProtectiveInstantsDeclare(t *testing.T) {
	for _, r := range []struct {
		card string
		want game.Answers
	}{
		{"Angelic Intervention", game.AnswerProtect | game.AnswerPump},
		{"Arachnogenesis", game.AnswerPrevent | game.AnswerMakesBlocker},
		{"Azorius Ploy", game.AnswerPrevent},
		{"Blinding Fog", game.AnswerProtect | game.AnswerPrevent},
		{"Blossoming Defense", game.AnswerProtect | game.AnswerPump},
		{"Bombadil's Song", game.AnswerProtect | game.AnswerPump},
		{"Borg Incursion", game.AnswerProtect | game.AnswerMakesBlocker},
		{"Brace for Impact", game.AnswerPump | game.AnswerPrevent},
		{"Chain of Silence", game.AnswerPrevent},
		{"Chameleon Blur", game.AnswerPrevent},
		{"Change of Plans", game.AnswerSurvive | game.AnswerPump},
		{"Chase Inspiration", game.AnswerProtect | game.AnswerPump},
		{"Commencement of Festivities", game.AnswerPrevent},
		{"Darkness", game.AnswerPrevent},
		{"Defend the Hearth", game.AnswerPrevent},
		{"Dive Down", game.AnswerProtect | game.AnswerPump},
		{"Druid's Deliverance", game.AnswerPrevent | game.AnswerMakesBlocker},
		{"Eerie Interference", game.AnswerPrevent},
		{"Efflorescence", game.AnswerSurvive | game.AnswerPump | game.AnswerCombatGrant},
		{"Emergency Regeneration", game.AnswerProtect | game.AnswerSurvive},
		{"Endure", game.AnswerPrevent},
		{"Energy Arc", game.AnswerPrevent},
		{"Enshrouding Mist", game.AnswerPump | game.AnswerPrevent},
		{"Enter the Avatar State", game.AnswerProtect | game.AnswerCombatGrant},
		{"Ethereal Haze", game.AnswerPrevent},
		{"Feat of Resistance", game.AnswerProtect | game.AnswerPump},
		{"Feint", game.AnswerPrevent},
		{"Fighting Chance", game.AnswerPrevent},
		{"Flare of Faith", game.AnswerSurvive | game.AnswerPump},
		{"Fleeting Flight", game.AnswerPump | game.AnswerPrevent | game.AnswerCombatGrant},
		{"Fog of War", game.AnswerPrevent},
		{"Forfend", game.AnswerPrevent},
		{"Foxfire", game.AnswerPrevent},
		{"Galadhrim Ambush", game.AnswerPrevent | game.AnswerMakesBlocker},
		{"Gods Willing", game.AnswerProtect},
		{"Grand Crescendo", game.AnswerSurvive | game.AnswerMakesBlocker},
		{"Hallow", game.AnswerPrevent},
		{"Harmless Assault", game.AnswerPrevent},
		{"Healing Grace", game.AnswerPrevent},
		{"Heroic Intervention", game.AnswerProtect | game.AnswerSurvive},
		{"Hindervines", game.AnswerPrevent},
		{"Holy Day", game.AnswerPrevent},
		{"Horrid Vigor", game.AnswerSurvive | game.AnswerCombatGrant},
		{"Hunter's Ambush", game.AnswerPrevent},
		{"Indestructible Aura", game.AnswerPrevent},
		{"Inkshield", game.AnswerPrevent | game.AnswerMakesBlocker},
		{"Inspire Awe", game.AnswerPrevent},
		{"Lazotep Plating", game.AnswerProtect | game.AnswerPump | game.AnswerMakesBlocker},
		{"Leap of Faith", game.AnswerPrevent | game.AnswerCombatGrant},
		{"Lithomancer's Focus", game.AnswerPump | game.AnswerPrevent},
		{"Loran's Escape", game.AnswerProtect | game.AnswerSurvive},
		{"Luminesce", game.AnswerPrevent},
		{"Magic Damper", game.AnswerProtect | game.AnswerPump},
		{"Mending Hands", game.AnswerPrevent},
		{"Moogles' Valor", game.AnswerSurvive | game.AnswerMakesBlocker},
		{"Moonmist", game.AnswerPrevent},
		{"Octopus Form", game.AnswerProtect | game.AnswerPump},
		{"Orcish Medicine", game.AnswerSurvive | game.AnswerPump | game.AnswerCombatGrant | game.AnswerMakesBlocker},
		{"Overprotect", game.AnswerProtect | game.AnswerSurvive | game.AnswerPump | game.AnswerCombatGrant},
		{"Pay No Heed", game.AnswerPrevent},
		{"Ranger's Guile", game.AnswerProtect | game.AnswerPump},
		{"Redeem", game.AnswerPrevent},
		{"Repel the Abominable", game.AnswerPrevent},
		{"Restrain", game.AnswerPrevent},
		{"Riot Control", game.AnswerPrevent},
		{"Ripples of Potential", game.AnswerSurvive | game.AnswerPump},
		{"Rootborn Defenses", game.AnswerSurvive | game.AnswerMakesBlocker},
		{"Sacred Boon", game.AnswerPrevent},
		{"Safe Passage", game.AnswerPrevent},
		{"Samite Ministration", game.AnswerPrevent},
		{"Saved by the Shell", game.AnswerProtect | game.AnswerSurvive | game.AnswerPump | game.AnswerCombatGrant},
		{"Serene Sunset", game.AnswerPrevent},
		{"Sewers of Estark", game.AnswerPrevent | game.AnswerCombatGrant},
		{"Shielded Passage", game.AnswerPrevent},
		{"Shore Up", game.AnswerProtect | game.AnswerPump},
		{"Snakeskin Veil", game.AnswerProtect | game.AnswerPump},
		{"Soul Parry", game.AnswerPrevent},
		{"Stonewood Invocation", game.AnswerProtect | game.AnswerPump},
		{"Subdue", game.AnswerPump | game.AnswerPrevent},
		{"Surge of Salvation", game.AnswerProtect | game.AnswerPrevent},
		{"Tangle", game.AnswerPrevent},
		{"Tanglesap", game.AnswerPrevent},
		{"Telekinesis", game.AnswerPrevent},
		{"Temper", game.AnswerPump | game.AnswerPrevent},
		{"Terrifying Presence", game.AnswerPrevent},
		{"Test of Faith", game.AnswerPump | game.AnswerPrevent},
		{"Thwart the Enemy", game.AnswerPrevent},
		{"Unbreakable Formation", game.AnswerSurvive | game.AnswerPump | game.AnswerCombatGrant},
		{"Undergrowth", game.AnswerPrevent},
		{"Veil of Summer", game.AnswerProtect},
		{"Vine Snare", game.AnswerPrevent},
		{"Warning", game.AnswerPrevent},
		{"Water Wings", game.AnswerProtect | game.AnswerPump | game.AnswerCombatGrant},
		{"Winds of Qal Sisma", game.AnswerPrevent},
		{"Wrap in Vigor", game.AnswerSurvive},
	} {
		t.Run(r.card, func(t *testing.T) {
			spec := effectstest.SpecNamed(t, r.card)
			if got := spec.Purpose.Answers; got != r.want {
				t.Errorf("declares %s, want %s", got, r.want)
			}
			if spec.Modes != nil {
				t.Errorf("is modal; a mode's own declaration would replace the card's")
			}
		})
	}
}
