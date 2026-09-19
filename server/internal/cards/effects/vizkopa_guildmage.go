package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vizkopa Guildmage — Creature — Human Wizard {W}{B}, 2/2 (EDHREC
// rank 4471):
//
//	"{1}{W}{B}: Target creature gains lifelink until end of turn.
//	 {1}{W}{B}: Whenever you gain life this turn, each opponent loses
//	 that much life."
//
// The two-mana guildmage that ends games in a lifegain deck: switch
// the second ability on, then gain a pile of life, and the whole
// table loses it. The first ability is the cheap way to make the
// second one fire — a lifelinked commander connecting for eight
// drains three opponents for eight each.
//
// DECLARED SIMPLIFICATION (weaker than printed): the SECOND ability
// is not implemented. What it creates is a TRIGGERED ability that
// exists for the rest of the turn and fires on an event ("whenever
// you gain life"), which is a floating trigger. The engine's two
// duration mechanisms are turn-scoped STATICS
// (RegisterScopedStaticForEffect, #279) and step-keyed DELAYED
// triggers (ScheduleDelayedTriggerForEffect, CR 603.7) — a static
// changes a characteristic, and a delayed trigger fires at a named
// step, so neither can express "watch an event kind until end of
// turn". Nothing is stronger than printed here: the Guildmage ships
// as its lifelink half only, which is strictly a subset of the card.
//
// The first ability is complete. Lifelink is a canonical keyword and
// the grant is turn-scoped, pinned to the targeted creature with its
// battlefield-entry stamp, so a creature flickered in response is a
// new object and is not granted anything (CR 400.7, CR 611.2c).
func init() {
	Register(Spec{
		OracleID:     "f19e7c5c-67fa-4ae4-89b8-afa0e08a6c48",
		Name:         "Vizkopa Guildmage",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Only the lifelink ability works. The second ability — \"whenever you gain life this turn, each opponent loses that much life\" — can't be activated.",
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{W}{B}: Target creature gains lifelink until end of turn.",
			Cost:    ManaCost("{1}{W}{B}"),
			Targets: TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						return GrantKeywordUntilEOT{
							Target:   t.ID,
							Keywords: []string{"lifelink"},
							Label:    "Vizkopa Guildmage — lifelink until end of turn",
						}.Apply(ctx)
					}
				}
				return nil
			},
		}},
	})
}
