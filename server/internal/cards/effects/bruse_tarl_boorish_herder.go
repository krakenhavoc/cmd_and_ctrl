package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bruse Tarl, Boorish Herder — Legendary Creature — Human Ally {2}{R}{W},
// 3/3: "Whenever Bruse Tarl enters or attacks, target creature you control
// gains double strike and lifelink until end of turn. Partner."
func init() {
	Register(Spec{
		OracleID:     "e39a1f6b-6b13-4a8b-ac0c-63ede4636812",
		Name:         "Bruse Tarl, Boorish Herder",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEntersOrAttacks("Bruse Tarl, Boorish Herder — target creature you control gains double strike and lifelink until end of turn",
				Do(OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
					return GrantKeywordUntilEOT{Target: id, Keywords: []string{"double strike", "lifelink"},
						Label: "Bruse Tarl — double strike and lifelink"}.Apply(ctx)
				}))),
				TargetCreature("target creature you control", YouControl())),
		},
	})
}
