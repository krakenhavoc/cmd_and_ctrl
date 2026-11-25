package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Finest Hour — {2}{G}{W}{U} Enchantment:
//
//	"Exalted (Whenever a creature you control attacks alone, that creature
//	 gets +1/+1 until end of turn.)
//	 Whenever a creature you control attacks alone, if it's the first combat
//	 phase of the turn, untap that creature. After this phase, there is an
//	 additional combat phase."
//
// Exalted is the printed keyword, whose trigger the engine derives (CR
// 702.83a). The second ability is Raiyuu's lone-attacker trigger with
// Karlach's intervening "if it's the first combat phase of the turn" (CR
// 603.4): read as the attack is declared and again as the trigger resolves,
// so the added combat's own lone attacker adds nothing. "Alone" is the whole
// declaration (CR 506.5). The creature untaps if it is still on the
// battlefield; the added combat comes with no main phase.
func init() {
	const label = "Finest Hour — untap that creature; additional combat phase"
	effects.Register(effects.Spec{
		OracleID:        "05678253-a40f-4c11-8ac4-335b3a02532e",
		Name:            "Finest Hour",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordExalted},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !effects.AttackDeclaredByYou(ev, source.Controller) || !game.AttackedAlone(g) || !effects.IsFirstCombatPhase(g) {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.Controller == source.Controller
			}, label, finestHourUntapAndCombat),
		},
	})
}

// finestHourUntapAndCombat is the trigger at resolution: the intervening if
// again, then the lone attacker untaps and another combat follows.
func finestHourUntapAndCombat(g *game.Game, item *game.StackItem) error {
	if !effects.IsFirstCombatPhase(g) {
		return nil
	}
	ctx := effects.NewContext(g, item)
	if attacker, ok := effects.NeoB09LoneAttacker(g, item); ok {
		if err := (effects.UntapTarget{Target: attacker.InstanceID}).Apply(ctx); err != nil {
			return err
		}
	}
	return effects.ExtraCombatAfterThisPhase().Apply(ctx)
}
