package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aether Charge — Enchantment {4}{R} (Onslaught):
//
//	"Whenever a Beast you control enters, you may have it deal 4 damage
//	 to target opponent or planeswalker."
//
// Warstorm Surge's trigger with a fixed amount and Final Strike's
// "target opponent or planeswalker": the entering Beast, not the
// enchantment, is the damage source, so its lifelink and deathtouch
// apply. The damage names the Beast by OBJECT, and a Beast that has
// left by resolution still deals it, with the keywords it had as it last
// existed. "Beast" is a word of the card's text (CR 612.2), so a
// changeling counts. The "you may" is asked as the trigger resolves (CR
// 603.5).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "15e83068-6253-4c65-8679-7295f3dc2075",
		Name:         "Aether Charge",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			Key:     "Aether Charge — the Beast deals 4 damage to target opponent or planeswalker",
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.IsCreature() && c.HasSubtype(source.TextSubtype("Beast"))
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{
				Question: "Aether Charge — have the Beast deal 4 damage to target opponent or planeswalker?",
			},
			Targets: targetOpponentOrPlaneswalker(),
			Effect:  aetherChargeBeastDealsFour,
		}},
	})
}

// aetherChargeBeastDealsFour is the trigger body: the entered Beast,
// live or by last-known information, deals 4 damage to the target.
func aetherChargeBeastDealsFour(g *game.Game, item *game.StackItem) error {
	if len(item.Targets) == 0 {
		return nil
	}
	ctx := NewContext(g, item)
	if _, ok := ctx.TriggeringPermanent(); !ok {
		return nil
	}
	ref := ctx.Trigger().Object.Ref()
	return DealDamage{SourceObject: &ref, Target: item.Targets[0].ID, Amount: 4}.Apply(ctx)
}
