package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     id,
		Name:         "Lluwen, Exchange Student",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersPrepared()},
		Activated: []ActivatedAbility{{
			Label:        "Exile a creature card from your graveyard: Lluwen becomes prepared. Activate only as a sorcery.",
			Cost:         ExileFromGraveyard(1, "a creature card", MatchCreature),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				return BecomePrepared{Target: item.SourceCardID}.Apply(NewContext(g, item))
			},
		}},
	})
	Register(Spec{
		OracleID:     id + "#1",
		Name:         "Pest Friend",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{Template: PestAttackToken(), N: 1}.Apply(ctx)
		},
	})
}
