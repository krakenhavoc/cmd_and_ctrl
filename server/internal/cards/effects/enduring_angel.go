package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const enduringAngelOracleID = "93d1e21b-7326-4aad-9f6e-e0e391845a18"

// Enduring Angel — Creature — Angel {2}{W}{W}{W}, 3/3:
//
//	"Flying, double strike
//	 You have hexproof.
//	 If your life total would be reduced to 0 or less, instead transform
//	 this creature and your life total becomes 3. Then if this creature
//	 didn't transform this way, you lose the game."
//
// Angelic Enforcer — Creature — Angel, */*:
//
//	"Flying
//	 You have hexproof.
//	 Angelic Enforcer's power and toughness are each equal to your life
//	 total.
//	 Whenever this creature attacks, double your life total."
//
// The third line is the seam of #5678, and it turned out to be a
// replacement on the window every falling life total already runs (CR
// 614, game/life_instead.go): a loss, a set-life effect (CR 119.5), a
// life payment (CR 119.4) and the life damage costs (CR 120.3a) are all
// one RepEventLife event. See enduringAngelLifeInstead below.
//
// "You have hexproof" is the derived player keyword Aegis of the Gods
// uses, declared on both faces, so it is unbroken across the transform
// (CR 712.18: the same object). The back face's power and toughness are
// Serra Avatar's characteristic-defining ability, and "double your life
// total" gains or loses the difference (CR 701.10d) through the ordinary
// life change, so a gain doubler or a "can't gain life" sees it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        enduringAngelOracleID,
		Name:            "Enduring Angel",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "double strike"},
		PlayerKeywords:  []string{KeywordHexproof},
		Replacements:    []game.ReplacementEffect{enduringAngelLifeInstead()},
	})
	Register(Spec{
		OracleID:        enduringAngelOracleID + "#1",
		Name:            "Angelic Enforcer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		PlayerKeywords:  []string{KeywordHexproof},
		Static:          []game.StaticAbility{powerAndToughnessEqualToYourLifeTotal()},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Angelic Enforcer — double your life total",
				func(g *game.Game, item *game.StackItem) error {
					return doubleLifeTotal(g, item.SourceCardID, item.Controller)
				}),
		},
	})
}

// enduringAngelLifeTotal is the total "your life total becomes 3" sets.
const enduringAngelLifeTotal = 3

// enduringAngelLifeInstead is "If your life total would be reduced to 0
// or less, instead transform this creature and your life total becomes
// 3. Then if this creature didn't transform this way, you lose the
// game."
//
// It applies to a life event of its controller's that is a loss and
// would leave the total at 0 or less, read as the event stands when the
// window asks (a doubler applied first has made the loss larger). The
// body, in the printed order:
//
//   - transform this creature, when an instruction to transform it would
//     turn it over now: a copy on a single-faced card, or one under a
//     "can't transform" effect, stays as it is (CR 701.27c, CR 101.2);
//   - the event becomes "your life total becomes 3", a gain or a loss of
//     the difference (CR 119.5), left in the window so the caller's
//     "life lost this way" and a "whenever you lose life" trigger read
//     what really moved;
//   - if it did not transform, the event carries the loss of the game,
//     carried out once the life has landed (CR 104.3e), so a player who
//     can't lose the game is left at 3 life.
//
// The window offers one effect an event only once (CR 614.5), and the
// transformed permanent no longer has the ability, so a second Enduring
// Angel is untouched by an event the first one answered: the total is
// no longer going to 0.
//
// The auto-tapper previews this window without touching the game
// (game.ReplacementIsPreviewForEffect); a preview only rewrites the
// event.
func enduringAngelLifeInstead() game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventChangeLife},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if src == nil || src.Controller == uuid.Nil || ev.LifePlayer != src.Controller {
				return false
			}
			return g.LifeWouldBeReducedToZeroOrLessForEffect(ev)
		},
		Replace: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error {
			if src == nil || g.ReplacementIsPreviewForEffect(ev) {
				g.LifeTotalBecomesInsteadForEffect(ev, enduringAngelLifeTotal)
				return nil
			}
			transformed := g.CanTransformForEffect(src.InstanceID)
			if transformed {
				if err := g.TransformPermanentForEffect(src.InstanceID); err != nil {
					return err
				}
			}
			g.LifeTotalBecomesInsteadForEffect(ev, enduringAngelLifeTotal)
			if !transformed {
				ev.LifeThenLoses = src.InstanceID
			}
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: "Enduring Angel: transform it and your life total becomes 3 instead of 0 or less",
	}
}

// powerAndToughnessEqualToYourLifeTotal is the characteristic-defining
// ability "<This>'s power and toughness are each equal to your life
// total" (CR 604.3, layer 7a): Serra Avatar, Angelic Enforcer. "You" is
// the controller, or the owner where there is none, so it works in every
// zone. DependsOnLifeTotal drops the layer cache when a life total
// changes, so the answer is never one life change stale (#1117).
func powerAndToughnessEqualToYourLifeTotal() game.StaticAbility {
	return game.StaticAbility{
		Label:              "Power and toughness are each equal to your life total.",
		Layer:              game.Layer7PT,
		SubLayer:           game.SubLayer7A_CDA,
		AppliesTo:          selfOnly,
		DependsOnLifeTotal: true,
		Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
			n := yourLifeTotal(g, source)
			c.Power, c.Toughness = n, n
		},
	}
}
