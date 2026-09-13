package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zurgo Helmsmasher — Legendary Creature — Orc Warrior, {2}{R}{W}{B}, 7/2:
//
//	"Haste"
//	"Zurgo Helmsmasher attacks each combat if able."
//	"During your turn, Zurgo Helmsmasher has indestructible."
//	"Whenever a creature dealt damage by Zurgo Helmsmasher this turn
//	 dies, put a +1/+1 counter on Zurgo Helmsmasher."
//
// The voltron commander that dares you to block it. Seven power on a
// four-mana commander is three hits to the 21-damage clock
// (CR 903.14a), and the two-toughness body that should make that
// suicidal is covered by conditional indestructible — but only on
// YOUR turn, which is the entire design.
//
// # The conditional static is the interesting part
//
// "During your turn" is a CR 613 continuous effect whose predicate
// reads the TURN, not the battlefield. `AppliesTo` runs on every
// recompute pass and re-answers the question each time, so the
// keyword genuinely blinks off the moment the turn passes: Zurgo
// attacks into a 5/5, survives, and then dies to a Shock on your
// opponent's upkeep. The layer engine bumps its version on turn
// change (layer_listener.go), so the recompute really does happen at
// the boundary rather than lazily.
//
// This is also the first card in the catalog whose static grants
// indestructible CONDITIONALLY, which is worth having: a predicate
// that can go false is the case a "grant once and forget"
// implementation of the keyword would have got wrong.
//
// # DECLARED SIMPLIFICATIONS
//
// Two of Zurgo's four lines are not enforced. Both are weaker than
// printed, never stronger.
//
//  1. "Attacks each combat if able" is a REQUIREMENT (CR 508.1d).
//     The engine has no attack-requirement machinery at all —
//     `DeclareAttacker` is entirely opt-in and nothing validates the
//     declared set against requirements — so Zurgo may legally sit
//     home. In a sandbox where players declare their own attacks
//     this is close to harmless: the requirement is on Zurgo's
//     controller, who wanted to attack anyway. It becomes real the
//     day the bot seat plays this card.
//
//  2. "Whenever a creature dealt damage by Zurgo this turn dies, put
//     a +1/+1 counter on Zurgo" needs per-creature damage-SOURCE
//     history. `Card.DamageMarked` is a bare integer and
//     `MarkedLethalByDeathtouch` is a bool; neither records WHO dealt
//     the damage, and nothing else in the engine does either. A
//     faithful version needs a damaged-by set on the card, cleared at
//     cleanup alongside DamageMarked — a small engine change, but one
//     with no second card asking for it yet, so it is not being
//     speculatively built here.
//
// Haste and the conditional indestructible are both fully live.
func init() {
	Register(Spec{
		OracleID:        "6c48d888-9f5d-43f4-adbd-61dbdba09260",
		Name:            "Zurgo Helmsmasher",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"Zurgo does not have to attack — \"attacks each combat if able\" is not enforced.", "The +1/+1 counter trigger never happens: the engine does not record which creature dealt a creature its damage.", "Indestructible saves a permanent from single-target removal and from lethal damage, but a board wipe (\"destroy all\") still destroys it."},
		PrintedKeywords: []string{"haste"},
		Static: []game.StaticAbility{{
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID &&
					isActivePlayer(g, source.Controller)
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Abilities = append(c.Abilities, "indestructible")
			},
		}},
	})
}
