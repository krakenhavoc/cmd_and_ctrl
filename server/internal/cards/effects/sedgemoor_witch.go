package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
// instantOrSorceryCastByYou condition Archmage Emeritus and Storm-Kiln
// Artist already share, making a Pest from the token table.
//
// "or COPY" rides EventSpellCopied (createSpellCopyLocked), so a copy
// of an instant or sorcery you control makes a Pest too (CR 207.2c).
//
// The Pest's own "when this token dies, you gain 1 life" ships since
// ADR 0083 (#1248) — the same `token:pest` catalog template Beledros
// Witherbloom makes, because it is the same printed token. Until then
// this card declared a second caveat for it.
func init() {
	Register(Spec{
		OracleID:        "25dce517-ac0a-4577-89ed-04296c7c4069",
		Name:            "Sedgemoor Witch",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			Ward(WardLife(3), "Sedgemoor Witch — ward, pay 3 life"),
			OnAny(magecraftEvents, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return instantOrSorceryCastOrCopiedByYou(ev, source, g)
			}, "Sedgemoor Witch — create a Pest (magecraft)", Do(CreateToken{Template: PestToken(), N: 1})),
		},
	})
}
