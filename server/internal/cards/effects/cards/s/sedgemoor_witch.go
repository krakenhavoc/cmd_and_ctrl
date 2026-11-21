package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sedgemoor Witch — Creature — Human Warlock {2}{B}, 3/2:
//
//	"Menace
//	 Ward—Pay 3 life.
//	 Magecraft — Whenever you cast or copy an instant or sorcery spell,
//	 create a 1/1 black and green Pest creature token with "When this
//	 token dies, you gain 1 life.""
//
// The second life-cost ward after Refraction Elemental, and the reason
// effects/ward.go's life cost has two cards on it. Menace is a printed
// combat keyword; the ward is Ward(WardLife(3)); magecraft is the
// shared Magecraft trigger Archmage Emeritus and Storm-Kiln Artist use,
// making a Pest from the token table.
//
// "or COPY" rides EventSpellCopied (createSpellCopyLocked), so a copy
// of an instant or sorcery you copied makes a Pest too (CR 207.2c).
//
// The Pest's own "when this token dies, you gain 1 life" ships since
// ADR 0083 (#1248) — the same `token:pest` catalog template Beledros
// Witherbloom makes, because it is the same printed token. Until then
// this card declared a second caveat for it.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "25dce517-ac0a-4577-89ed-04296c7c4069",
		Name:            "Sedgemoor Witch",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardLife(3), "Sedgemoor Witch — ward, pay 3 life"),
			effects.Magecraft("Sedgemoor Witch — create a Pest (magecraft)", effects.Do(effects.CreateToken{Template: effects.PestToken(), N: 1})),
		},
	})
}
