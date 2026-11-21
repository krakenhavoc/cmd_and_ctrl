package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lluwen, Exchange Student // Pest Friend — a preparation card (CR 722, ADR
// 0090), Secrets of Strixhaven (#4456):
//
//	Lluwen, Exchange Student {2}{B}{G}, Legendary Creature — Elf Druid 3/4:
//	"Lluwen enters prepared. (While it's prepared, you may cast a copy of its
//	 spell. Doing so unprepares it.)
//	 Exile a creature card from your graveyard: Lluwen becomes prepared.
//	 Activate only as a sorcery."
//	Pest Friend {B/G}, Sorcery:
//	"Create a 1/1 black and green Pest creature token with "Whenever this
//	 token attacks, you gain 1 life.""
//
// "Enters prepared" is SelfEntersPrepared. The activated ability's cost is an
// exile of a creature card from your own graveyard (Ellie and Alan's cost,
// chosen as the ability is announced), at sorcery speed; becoming prepared
// does nothing if Lluwen is already prepared (CR 722.3a) or has left. Pest
// Friend makes the Secrets of Strixhaven Pest (PestAttackToken: life when it
// attacks, not the Strixhaven Pest's life when it dies).
//
// No simplification.
func init() {
	const id = "cb125cd4-c7c2-415d-b802-bdcfc2454403"
	effects.Register(effects.Spec{
		OracleID:     id,
		Name:         "Lluwen, Exchange Student",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersPrepared()},
		Activated: []effects.ActivatedAbility{{
			Label:        "Exile a creature card from your graveyard: Lluwen becomes prepared. Activate only as a sorcery.",
			Cost:         effects.ExileFromGraveyard(1, "a creature card", effects.MatchCreature),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.BecomePrepared{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			},
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     id + "#1",
		Name:         "Pest Friend",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{Template: effects.PestAttackToken(), N: 1}.Apply(ctx)
		},
	})
}
