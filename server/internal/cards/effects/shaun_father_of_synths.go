package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shaun, Father of Synths — Legendary Creature — Human Scientist {3}{U}{R},
// 3/4 (Fallout, #5948):
//
//	"Whenever you attack, you may create a tapped and attacking token that's a
//	 copy of target attacking legendary creature you control other than Shaun,
//	 except it's not legendary and it's a Synth artifact creature in addition
//	 to its other types.
//	 When Shaun leaves the battlefield, exile all Synth tokens you control."
//
// "Whenever you attack" is once per declaration (OncePerBatch). The target,
// an attacking legendary creature its controller controls other than Shaun, is
// chosen as the trigger goes on the stack and re-checked as it resolves; the
// "may" is asked then. The token takes the creature's copiable values (CR 707.2)
// with the exception written into its type line (no Legendary supertype, plus
// Artifact and the Synth subtype, CR 707.9b), and enters tapped and attacking a
// player, planeswalker or battle its controller chooses, never declared as an
// attacker (CR 508.4). The leave trigger exiles every Synth token its controller
// controls as it resolves, including ones made by another Shaun.
//
// No simplification.
func init() {
	attack := OncePerBatch(Optional(Targeting(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
		return attackDeclaredByYou(ev, source.Controller)
	}, "Shaun, Father of Synths — create a tapped and attacking Synth copy of target attacking legendary creature", shaunCopy),
		Another(TargetCreature("target attacking legendary creature you control other than Shaun", AttackingCreature(), Legendary(), YouControl()))),
		"Shaun, Father of Synths — create a tapped and attacking token copy of the target?"))
	Register(Spec{
		OracleID:     "0bd11d6c-b707-4b37-b49c-a3274b2a5a81",
		Name:         "Shaun, Father of Synths",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			attack,
			WhenThisLeaves("Shaun, Father of Synths — exile all Synth tokens you control", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				var ids []uuid.UUID
				for _, c := range MatchingBattlefield(ctx, And(YouControl(), IsTokenPredicate(), HasSubtype("Synth"))) {
					ids = append(ids, c.InstanceID)
				}
				for _, id := range ids {
					if err := (ExileTarget{Target: id}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}),
		},
	})
}

// shaunCopy makes the Synth copy of the chosen attacker.
func shaunCopy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	tmpl, ok := TokenCopyTemplate(g, id)
	if !ok {
		return nil
	}
	tmpl.TypeLine = shaunSynthTypeLine(tmpl.TypeLine)
	return CreateTokensAttackingYourChoice{
		Controller: item.Controller,
		Template:   copiedToken(tmpl),
		Copy:       true,
		N:          1,
		Tapped:     true,
		Prefer:     triggerAttackTarget(item),
		Label:      "Shaun, Father of Synths",
	}.Apply(ctx)
}

// shaunSynthTypeLine is the copy's type line with the exception applied: not
// legendary, an artifact creature, and a Synth in addition to its other creature
// types.
func shaunSynthTypeLine(line string) string {
	line = strings.Replace(line, "Legendary ", "", 1)
	left, right, found := strings.Cut(line, " — ")
	if !strings.Contains(left, "Artifact") {
		left = strings.Replace(left, "Creature", "Artifact Creature", 1)
	}
	if !found {
		return left + " — Synth"
	}
	if !strings.Contains(" "+right+" ", " Synth ") {
		right += " Synth"
	}
	return left + " — " + right
}
