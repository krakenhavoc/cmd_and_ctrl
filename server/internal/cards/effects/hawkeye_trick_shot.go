package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hawkeye, Trick Shot — {3}{R} Legendary Creature — Human Archer Hero 3/4
// (Marvel Super Heroes Commander, #4455):
//
//	"Reach, first strike
//	 Whenever Hawkeye or another Hero you control enters, it deals damage
//	 equal to the number of Heroes you control to any target."
//
// Warstorm Surge's shape: the entering Hero, not Hawkeye, is the damage
// source (named by object, so its keywords and last-known information
// apply), and the amount is the number of Heroes its controller controls as
// the trigger RESOLVES, so a Hero that arrived in response is counted.
//
// No simplification.
func init() {
	const label = "Hawkeye, Trick Shot — the Hero deals damage equal to the number of Heroes you control to any target"
	t := On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		c, ok := enteredUnderYourControl(ev, source, g, false)
		// "Hero" is a word of the source's text (CR 612.2).
		return ok && c.HasSubtype(source.TextSubtype("Hero"))
	}, label, msc12HeroDamage)
	t.Targets = TargetAny()
	Register(Spec{
		OracleID:        "ead2aee3-4f95-4ff1-8be9-ea9619e2386e",
		Name:            "Hawkeye, Trick Shot",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "first strike"},
		Triggered:       []game.TriggeredAbility{t},
	})
}

// msc12HeroDamage is the trigger body: the entered Hero deals damage equal
// to the Heroes its controller controls now.
func msc12HeroDamage(g *game.Game, item *game.StackItem) error {
	if len(item.Targets) == 0 {
		return nil
	}
	ctx := NewContext(g, item)
	if ctx.Trigger().Object == nil {
		return nil
	}
	hero := "Hero"
	if src, ok := g.LookupCardForEffect(item.SourceCardID); ok {
		// "Hero" is a word of Hawkeye's text (CR 612.2).
		hero = src.TextSubtype("Hero")
	}
	heroes := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.HasSubtype(hero) {
			heroes++
		}
	}
	ref := ctx.Trigger().Object.Ref()
	return DealDamage{SourceObject: &ref, Target: item.Targets[0].ID, Amount: heroes}.Apply(ctx)
}
