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
// Two sandbox simplifications, both weaker than printed, and both the
// same gaps other cards already declare:
//
//   - "or COPY" — the engine emits no event when a spell is copied
//     (CopySpellForEffect puts the copy on the stack without one), so
//     magecraft fires on casts only. Archmage Emeritus declares the
//     same gap.
//   - The Pest's "when this token dies, you gain 1 life" is not on the
//     token. A token template carries no triggered abilities and has no
//     oracle ID to key one on; Beledros Witherbloom declares the same
//     gap. Watching Pest deaths from the Witch instead would be wrong
//     the moment the Witch left the battlefield, so it is not faked.
func init() {
	Register(Spec{
		OracleID:     "25dce517-ac0a-4577-89ed-04296c7c4069",
		Name:         "Sedgemoor Witch",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Magecraft only fires on instants and sorceries you cast, not on copies of them.",
			"The Pest tokens don't gain you 1 life when they die.",
		},
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			Ward(WardLife(3), "Sedgemoor Witch — ward, pay 3 life"),
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return instantOrSorceryCastByYou(ev, source, g)
			}, "Sedgemoor Witch — create a Pest (magecraft)", Do(CreateToken{Template: TokenCard("1/1 black and green Pest"), N: 1})),
		},
	})
}
